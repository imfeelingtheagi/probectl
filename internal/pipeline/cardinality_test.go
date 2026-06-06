package pipeline

import (
	"fmt"
	"testing"

	"github.com/imfeelingtheagi/probectl/internal/store/tsdb"
)

func seriesFor(metric string) []tsdb.Series {
	return []tsdb.Series{{Metric: metric, Labels: map[string]string{"tenant_id": "x"}, Value: 1}}
}

// U-017: one agent flooding unique series hits its cap (rejected + counted);
// a DIFFERENT tenant is completely unaffected.
func TestCardinalityCapFloodIsolatesTenants(t *testing.T) {
	l := NewCardinalityLimiter(10, 100)

	totalAdmitted, totalDropped := 0, 0
	for i := 0; i < 50; i++ {
		adm, dropped := l.Filter("tenant-flood", "agent-1", seriesFor(fmt.Sprintf("m_%d", i)))
		totalAdmitted += len(adm)
		totalDropped += dropped
	}
	if totalAdmitted != 10 {
		t.Fatalf("admitted = %d, want exactly the per-agent cap (10)", totalAdmitted)
	}
	if totalDropped != 40 {
		t.Fatalf("dropped = %d, want 40", totalDropped)
	}
	st := l.Stats()
	if st.Dropped != 40 || st.TenantDropped["tenant-flood"] != 40 {
		t.Fatalf("stats = %+v, want the drops attributed to the flooder", st)
	}

	// The quiet tenant admits freely — the flood is not its problem.
	adm, dropped := l.Filter("tenant-quiet", "agent-9", seriesFor("steady_metric"))
	if len(adm) != 1 || dropped != 0 {
		t.Fatalf("quiet tenant impacted: admitted=%d dropped=%d", len(adm), dropped)
	}
	if st := l.Stats(); st.TenantDropped["tenant-quiet"] != 0 {
		t.Fatalf("quiet tenant has drops: %+v", st)
	}
}

// Known identities keep flowing at the cap — steady-state telemetry is never
// starved; only NEW identities are gated.
func TestCardinalityKnownSeriesKeepFlowing(t *testing.T) {
	l := NewCardinalityLimiter(1, 10)
	if adm, d := l.Filter("t", "a", seriesFor("known")); len(adm) != 1 || d != 0 {
		t.Fatalf("first series rejected: %d/%d", len(adm), d)
	}
	if _, d := l.Filter("t", "a", seriesFor("overflow")); d != 1 {
		t.Fatal("cap did not reject the new identity")
	}
	for i := 0; i < 5; i++ {
		if adm, d := l.Filter("t", "a", seriesFor("known")); len(adm) != 1 || d != 0 {
			t.Fatalf("known identity starved at the cap (iteration %d)", i)
		}
	}
}

// The per-tenant wall holds across many agents.
func TestCardinalityPerTenantWall(t *testing.T) {
	l := NewCardinalityLimiter(1000, 5)
	dropped := 0
	for agent := 0; agent < 10; agent++ {
		_, d := l.Filter("t", fmt.Sprintf("a%d", agent), seriesFor(fmt.Sprintf("m%d", agent)))
		dropped += d
	}
	if dropped != 5 {
		t.Fatalf("dropped = %d, want 5 (the tenant wall)", dropped)
	}
}

// Distinct label VALUES are distinct identities (the explosion vector).
func TestCardinalityIdentityIncludesLabels(t *testing.T) {
	l := NewCardinalityLimiter(2, 10)
	a := []tsdb.Series{{Metric: "m", Labels: map[string]string{"server_address": "a"}}}
	b := []tsdb.Series{{Metric: "m", Labels: map[string]string{"server_address": "b"}}}
	c := []tsdb.Series{{Metric: "m", Labels: map[string]string{"server_address": "c"}}}
	l.Filter("t", "ag", a)
	l.Filter("t", "ag", b)
	if _, d := l.Filter("t", "ag", c); d != 1 {
		t.Fatal("label-value explosion not capped")
	}
}
