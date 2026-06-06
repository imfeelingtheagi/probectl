package pipeline

import (
	"sort"
	"strings"
	"sync"

	"github.com/imfeelingtheagi/probectl/internal/store/tsdb"
)

// Cardinality caps (U-017): Result.Metrics is agent-supplied, so a single
// misbehaving (or hostile) agent could mint unbounded unique series and blow
// up every store downstream. Ingest now tracks ACTIVE series identities per
// (tenant, agent) and per tenant; a NEW identity past the cap is rejected
// (dropped + counted, per-series — known identities keep flowing), so one
// agent's explosion can never become another tenant's problem (the fairness
// stance, S-T7).

// Default caps: generous for real probes (a canary emits a handful of
// metrics), hard walls for an explosion.
const (
	DefaultMaxSeriesPerAgent  = 1000
	DefaultMaxSeriesPerTenant = 50000
)

// CardinalityLimiter admits series identities under per-agent and per-tenant
// caps. It is safe for concurrent use.
type CardinalityLimiter struct {
	perAgent  int
	perTenant int

	mu      sync.Mutex
	tenants map[string]*tenantSeries
	dropped uint64 // total rejected series (never silent)
}

type tenantSeries struct {
	all     map[string]struct{}            // tenant-wide identities
	byAgent map[string]map[string]struct{} // agent -> identities
	dropped uint64
}

// NewCardinalityLimiter builds a limiter; non-positive caps use the defaults.
func NewCardinalityLimiter(perAgent, perTenant int) *CardinalityLimiter {
	if perAgent <= 0 {
		perAgent = DefaultMaxSeriesPerAgent
	}
	if perTenant <= 0 {
		perTenant = DefaultMaxSeriesPerTenant
	}
	return &CardinalityLimiter{perAgent: perAgent, perTenant: perTenant, tenants: map[string]*tenantSeries{}}
}

// seriesIdentity is the cardinality key: metric name + every label pair.
func seriesIdentity(s tsdb.Series) string {
	keys := make([]string, 0, len(s.Labels))
	for k := range s.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(s.Metric)
	for _, k := range keys {
		b.WriteByte(0)
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(s.Labels[k])
	}
	return b.String()
}

// Filter returns the admitted subset of series for (tenant, agent) and the
// number rejected by the caps. Known identities always pass (steady-state
// telemetry keeps flowing at the cap); only NEW identities are gated.
func (l *CardinalityLimiter) Filter(tenant, agent string, series []tsdb.Series) ([]tsdb.Series, int) {
	if l == nil {
		return series, 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	ts := l.tenants[tenant]
	if ts == nil {
		ts = &tenantSeries{all: map[string]struct{}{}, byAgent: map[string]map[string]struct{}{}}
		l.tenants[tenant] = ts
	}
	ag := ts.byAgent[agent]
	if ag == nil {
		ag = map[string]struct{}{}
		ts.byAgent[agent] = ag
	}

	admitted := series[:0]
	droppedHere := 0
	for _, s := range series {
		id := seriesIdentity(s)
		if _, known := ag[id]; known {
			admitted = append(admitted, s)
			continue
		}
		if len(ag) >= l.perAgent || len(ts.all) >= l.perTenant {
			droppedHere++
			continue
		}
		ag[id] = struct{}{}
		ts.all[id] = struct{}{}
		admitted = append(admitted, s)
	}
	if droppedHere > 0 {
		l.dropped += uint64(droppedHere)
		ts.dropped += uint64(droppedHere)
	}
	return admitted, droppedHere
}

// CardinalityStats reports the rejection counters.
type CardinalityStats struct {
	Dropped       uint64
	TenantDropped map[string]uint64
}

// Stats snapshots the counters (per-tenant drops included, for fairness
// visibility).
func (l *CardinalityLimiter) Stats() CardinalityStats {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := CardinalityStats{Dropped: l.dropped, TenantDropped: map[string]uint64{}}
	for t, ts := range l.tenants {
		if ts.dropped > 0 {
			out.TenantDropped[t] = ts.dropped
		}
	}
	return out
}
