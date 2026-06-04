package flowstore

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Memory is the in-process Store: the lightweight-mode backend, the test
// double, and the reference implementation the ClickHouse SQL must agree with.
// Rows are bounded (FIFO eviction) so an unconsumed dev deployment cannot grow
// without limit.
type Memory struct {
	mu   sync.Mutex
	rows []Row
	max  int
}

// NewMemory builds a Memory store bounded to ~1M rows.
func NewMemory() *Memory {
	return &Memory{max: 1 << 20}
}

// Insert appends rows, evicting oldest beyond the bound.
func (m *Memory) Insert(_ context.Context, rows []Row) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = append(m.rows, rows...)
	if over := len(m.rows) - m.max; over > 0 {
		m.rows = append([]Row(nil), m.rows[over:]...)
	}
	return nil
}

// Len reports stored rows (tests).
func (m *Memory) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.rows)
}

// inWindow snapshots the tenant's rows inside the query window — the tenant
// filter is applied before anything else (CLAUDE.md §6).
func (m *Memory) inWindow(tenant string, from, to time.Time) []Row {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Row, 0, 256)
	for _, r := range m.rows {
		if r.TenantID != tenant {
			continue
		}
		if r.TS.Before(from) || r.TS.After(to) {
			continue
		}
		out = append(out, r)
	}
	return out
}

// TopTalkers aggregates the window by the requested key.
func (m *Memory) TopTalkers(_ context.Context, q TopQuery) ([]TopRow, error) {
	if err := q.normalize(); err != nil {
		return nil, err
	}
	type agg struct {
		detail      string
		bytes, pkts uint64
		flows       uint64
	}
	groups := make(map[string]*agg)
	for _, r := range m.inWindow(q.TenantID, q.Now.Add(-q.Window), q.Now) {
		var key, detail string
		switch q.By {
		case BySrc:
			key = r.SrcAddr
		case ByDst:
			key = r.DstAddr
		case ByPair:
			key, detail = r.SrcAddr, r.DstAddr
		case BySrcASN:
			if r.SrcASN == 0 {
				continue
			}
			key, detail = strconv.FormatUint(uint64(r.SrcASN), 10), r.SrcASName
		case ByDstASN:
			if r.DstASN == 0 {
				continue
			}
			key, detail = strconv.FormatUint(uint64(r.DstASN), 10), r.DstASName
		}
		if key == "" {
			continue
		}
		gk := key + "\x00" + detail
		g, ok := groups[gk]
		if !ok {
			g = &agg{detail: detail}
			groups[gk] = g
		}
		g.bytes += r.BytesScaled
		g.pkts += r.PacketsScaled
		g.flows++
	}
	out := make([]TopRow, 0, len(groups))
	for gk, g := range groups {
		key := gk
		if i := indexByte(gk, 0); i >= 0 {
			key = gk[:i]
		}
		out = append(out, TopRow{Key: key, Detail: g.detail, Bytes: g.bytes, Packets: g.pkts, Flows: g.flows})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Bytes != out[j].Bytes {
			return out[i].Bytes > out[j].Bytes
		}
		return out[i].Key < out[j].Key
	})
	if len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, nil
}

// Capacity buckets the window into per-(exporter, iface) throughput points.
func (m *Memory) Capacity(_ context.Context, q CapacityQuery) ([]CapacityPoint, error) {
	if err := q.normalize(); err != nil {
		return nil, err
	}
	return m.capacitySeries(q), nil
}

func (m *Memory) capacitySeries(q CapacityQuery) []CapacityPoint {
	type key struct {
		exporter string
		iface    uint32
		bucket   int64
	}
	type agg struct{ bytes, pkts uint64 }
	bucketSecs := int64(q.Bucket / time.Second)
	groups := make(map[key]*agg)
	for _, r := range m.inWindow(q.TenantID, q.Now.Add(-q.Window), q.Now) {
		if q.Exporter != "" && r.Exporter != q.Exporter {
			continue
		}
		iface := r.InIf
		if q.Direction == "out" {
			iface = r.OutIf
		}
		k := key{r.Exporter, iface, r.TS.Unix() / bucketSecs * bucketSecs}
		g, ok := groups[k]
		if !ok {
			g = &agg{}
			groups[k] = g
		}
		g.bytes += r.BytesScaled
		g.pkts += r.PacketsScaled
	}
	out := make([]CapacityPoint, 0, len(groups))
	for k, g := range groups {
		out = append(out, CapacityPoint{
			TS:       time.Unix(k.bucket, 0).UTC(),
			Exporter: k.exporter,
			Iface:    k.iface,
			Bps:      float64(g.bytes) * 8 / float64(bucketSecs),
			Pps:      float64(g.pkts) / float64(bucketSecs),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].TS.Equal(out[j].TS) {
			return out[i].TS.Before(out[j].TS)
		}
		if out[i].Exporter != out[j].Exporter {
			return out[i].Exporter < out[j].Exporter
		}
		return out[i].Iface < out[j].Iface
	})
	return out
}

// Anomalies runs the shared detector over the capacity series.
func (m *Memory) Anomalies(_ context.Context, q AnomalyQuery) ([]Anomaly, error) {
	if err := q.normalize(); err != nil {
		return nil, err
	}
	return detectAnomalies(m.capacitySeries(q.capacityQuery()), q), nil
}

// Close is a no-op.
func (m *Memory) Close() error { return nil }

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
