package flowstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// tenant_id leads both the partition and the ORDER BY so tenant-scoped reads
// prune at the storage layer (CLAUDE.md §4, §6); the day component bounds part
// sizes at NetFlow volumes and makes the retention TTL cheap to apply. The
// LowCardinality columns keep the high-volume dictionary small.
const createFlows = `CREATE TABLE IF NOT EXISTS probectl_flows (
  tenant_id String, agent_id String, exporter String, obs_domain UInt32,
  protocol LowCardinality(String),
  ts DateTime64(3), start_ts DateTime64(3),
  src_addr String, dst_addr String, src_port UInt16, dst_port UInt16,
  transport LowCardinality(String), net_type LowCardinality(String),
  in_if UInt32, out_if UInt32, vlan UInt16, tos UInt8, tcp_flags UInt8, next_hop String,
  bytes UInt64, packets UInt64, sampling UInt64, bytes_scaled UInt64, packets_scaled UInt64,
  src_asn UInt32, src_as_name String, src_country LowCardinality(String),
  dst_asn UInt32, dst_as_name String, dst_country LowCardinality(String)
) ENGINE = MergeTree
PARTITION BY (tenant_id, toYYYYMMDD(ts))
ORDER BY (tenant_id, ts, exporter, src_addr, dst_addr)`

// ClickHouse persists flows over the ClickHouse HTTP interface (pathstore
// pattern: zero driver dependencies; https URL = TLS in transit).
type ClickHouse struct {
	base   string
	client *http.Client
}

// NewClickHouse connects, ensures the schema, and (when retentionDays > 0)
// applies the delete-TTL — idempotently, so repeated starts are safe.
func NewClickHouse(rawURL string, retentionDays int) (*ClickHouse, error) {
	c := &ClickHouse{base: strings.TrimRight(rawURL, "/"), client: &http.Client{Timeout: 30 * time.Second}}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := c.exec(ctx, createFlows, nil); err != nil {
		return nil, fmt.Errorf("flowstore: create table: %w", err)
	}
	if retentionDays > 0 {
		ttl := fmt.Sprintf("ALTER TABLE probectl_flows MODIFY TTL toDateTime(ts) + INTERVAL %d DAY DELETE", retentionDays)
		if err := c.exec(ctx, ttl, nil); err != nil {
			return nil, fmt.Errorf("flowstore: apply retention TTL: %w", err)
		}
	}
	return c, nil
}

// chRow is the JSONEachRow insert shape (times rendered as ClickHouse strings).
type chRow struct {
	Row
	TSStr    string `json:"ts"`
	StartStr string `json:"start_ts"`
}

// Insert streams rows as one JSONEachRow batch (the high-volume path: one HTTP
// request per collector flush, not per record).
func (c *ClickHouse) Insert(ctx context.Context, rows []Row) error {
	if len(rows) == 0 {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i := range rows {
		r := chRow{Row: rows[i],
			TSStr:    rows[i].TS.UTC().Format("2006-01-02 15:04:05.000"),
			StartStr: rows[i].StartTS.UTC().Format("2006-01-02 15:04:05.000")}
		if err := enc.Encode(r); err != nil {
			return fmt.Errorf("flowstore: encode row: %w", err)
		}
	}
	return c.exec(ctx, "INSERT INTO probectl_flows FORMAT JSONEachRow", &buf)
}

// topSQL builds the top-talkers query (exported via a test for the tenant
// guard: the WHERE must lead with tenant_id).
func topSQL(q TopQuery) string {
	var key, detail, extra string
	switch q.By {
	case BySrc:
		key, detail = "src_addr", "''"
	case ByDst:
		key, detail = "dst_addr", "''"
	case ByPair:
		key, detail = "src_addr", "dst_addr"
	case BySrcASN:
		key, detail, extra = "toString(src_asn)", "any(src_as_name)", " AND src_asn != 0"
	case ByDstASN:
		key, detail, extra = "toString(dst_asn)", "any(dst_as_name)", " AND dst_asn != 0"
	}
	groupBy := "k"
	if q.By == ByPair {
		groupBy = "k, d"
	}
	return fmt.Sprintf(
		`SELECT %s AS k, %s AS d, sum(bytes_scaled) AS b, sum(packets_scaled) AS p, count() AS f `+
			`FROM probectl_flows WHERE tenant_id=%s AND ts >= %s AND ts <= %s%s `+
			`GROUP BY %s ORDER BY b DESC, k ASC LIMIT %d`,
		key, detail, chStr(q.TenantID), chTime(q.Now.Add(-q.Window)), chTime(q.Now), extra, groupBy, q.Limit)
}

// TopTalkers runs the aggregation in ClickHouse.
func (c *ClickHouse) TopTalkers(ctx context.Context, q TopQuery) ([]TopRow, error) {
	if err := q.normalize(); err != nil {
		return nil, err
	}
	rows, err := c.query(ctx, topSQL(q))
	if err != nil {
		return nil, err
	}
	out := make([]TopRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, TopRow{
			Key:     chToString(r["k"]),
			Detail:  chToString(r["d"]),
			Bytes:   uint64(chToFloat(r["b"])),
			Packets: uint64(chToFloat(r["p"])),
			Flows:   uint64(chToFloat(r["f"])),
		})
	}
	return out, nil
}

// capacitySQL buckets throughput per exporter/interface in ClickHouse.
func capacitySQL(q CapacityQuery) string {
	iface := "in_if"
	if q.Direction == "out" {
		iface = "out_if"
	}
	secs := int64(q.Bucket / time.Second)
	exporterFilter := ""
	if q.Exporter != "" {
		exporterFilter = " AND exporter=" + chStr(q.Exporter)
	}
	return fmt.Sprintf(
		`SELECT exporter, %s AS iface, toStartOfInterval(ts, INTERVAL %d second) AS t, `+
			`sum(bytes_scaled)*8/%d AS bps, sum(packets_scaled)/%d AS pps `+
			`FROM probectl_flows WHERE tenant_id=%s AND ts >= %s AND ts <= %s%s `+
			`GROUP BY exporter, iface, t ORDER BY t, exporter, iface`,
		iface, secs, secs, secs, chStr(q.TenantID), chTime(q.Now.Add(-q.Window)), chTime(q.Now), exporterFilter)
}

// Capacity runs the bucket aggregation in ClickHouse.
func (c *ClickHouse) Capacity(ctx context.Context, q CapacityQuery) ([]CapacityPoint, error) {
	if err := q.normalize(); err != nil {
		return nil, err
	}
	rows, err := c.query(ctx, capacitySQL(q))
	if err != nil {
		return nil, err
	}
	out := make([]CapacityPoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, CapacityPoint{
			TS:       chParseTime(chToString(r["t"])),
			Exporter: chToString(r["exporter"]),
			Iface:    uint32(chToFloat(r["iface"])),
			Bps:      chToFloat(r["bps"]),
			Pps:      chToFloat(r["pps"]),
		})
	}
	return out, nil
}

// Anomalies fetches the capacity series and applies the shared detector, so
// ClickHouse and Memory flag identically.
func (c *ClickHouse) Anomalies(ctx context.Context, q AnomalyQuery) ([]Anomaly, error) {
	if err := q.normalize(); err != nil {
		return nil, err
	}
	series, err := c.Capacity(ctx, q.capacityQuery())
	if err != nil {
		return nil, err
	}
	return detectAnomalies(series, q), nil
}

// Close is a no-op (the HTTP client needs no teardown).
func (c *ClickHouse) Close() error { return nil }

// --- HTTP plumbing (pathstore pattern) ---------------------------------------

func (c *ClickHouse) query(ctx context.Context, sql string) ([]map[string]any, error) {
	u := c.base + "/?query=" + url.QueryEscape(sql+" FORMAT JSONEachRow")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("flowstore: clickhouse query: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("flowstore: clickhouse query status %d: %s", resp.StatusCode, body)
	}
	var rows []map[string]any
	for _, line := range bytes.Split(body, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, fmt.Errorf("flowstore: decode row: %w", err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (c *ClickHouse) exec(ctx context.Context, query string, body io.Reader) error {
	u := c.base + "/?query=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, body)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("flowstore: clickhouse request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("flowstore: clickhouse status %d: %s", resp.StatusCode, b)
	}
	return nil
}

// chStr renders a ClickHouse string literal with the necessary escaping.
func chStr(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s) + "'"
}

// chTime renders an absolute DateTime64 literal.
func chTime(t time.Time) string {
	return "toDateTime64(" + chStr(t.UTC().Format("2006-01-02 15:04:05.000")) + ", 3)"
}

// chParseTime parses ClickHouse DateTime / DateTime64 strings.
func chParseTime(s string) time.Time {
	for _, layout := range []string{"2006-01-02 15:04:05.000", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

func chToString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func chToFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case string:
		var f float64
		_, _ = fmt.Sscanf(n, "%g", &f)
		return f
	}
	return 0
}
