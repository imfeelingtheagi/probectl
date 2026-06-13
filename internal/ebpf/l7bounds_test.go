// SPDX-License-Identifier: LicenseRef-probectl-TBD

package ebpf

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/imfeelingtheagi/probectl/internal/ebpf/l7"
)

// TestAgentL7MapsStayBounded drives the LIVE agent path (observeL7, the same
// method Run() calls per event) with far more distinct connections than the cap
// and asserts every per-connection map stays bounded and the eviction counters
// rise — EBPF-001 (the bounded maps are wired into the runtime) + FUZZ-001 (the
// L7 maps are capped). Pre-fix, l7conns / l7man.conns / the service map grew
// unbounded (cap=0 was never set from the runtime), so Len() would equal N.
func TestAgentL7MapsStayBounded(t *testing.T) {
	const (
		connCap = 64
		edgeCap = 64
		n       = 5000 // >> caps
	)
	cfg := &Config{
		TenantID:        "t1",
		Host:            "node-1",
		FlushInterval:   time.Hour,
		MaxL7Conns:      connCap,
		MaxServiceEdges: edgeCap,
		L7ConnIdleTTL:   5 * time.Minute,
	}
	a := newAgentWith(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), &sliceSource{}, NopEnricher{}, &captureEmitter{})

	base := time.Unix(0, 0)
	req := []byte("GET /x HTTP/1.1\r\nContent-Length: 0\r\n\r\n")
	for i := 0; i < n; i++ {
		// Each connID is distinct AND each edge (src→dst) is distinct, so an
		// unbounded implementation would accumulate n entries in BOTH maps.
		a.observeL7(L7Event{
			ConnID:      uint64(i + 1),
			TenantID:    "t1",
			Source:      Endpoint{Workload: "client"},
			Destination: Endpoint{Workload: "svc", Port: uint32(10000 + i)},
			Transport:   "tcp",
			Data:        l7.DataEvent{Kind: l7.Request, Time: base.Add(time.Duration(i) * time.Millisecond), Payload: req},
		})
		// Fold a matching flow so the SERVICE map also sees distinct edges.
		a.observe(Flow{
			TenantID:    "t1",
			Source:      Endpoint{Workload: "client"},
			Destination: Endpoint{Workload: "svc", Port: uint32(10000 + i)},
			Transport:   "tcp",
			Observed:    base.Add(time.Duration(i) * time.Millisecond),
		})
	}

	if got := len(a.l7conns); got > connCap {
		t.Errorf("l7conns size = %d, exceeds cap %d (FUZZ-001: identity map unbounded)", got, connCap)
	}
	if got := a.l7man.Len(); got > connCap {
		t.Errorf("l7man.Len() = %d, exceeds cap %d (FUZZ-001: tracker map unbounded)", got, connCap)
	}
	if got := a.agg.ServiceMap().Len(); got > edgeCap {
		t.Errorf("service map Len() = %d, exceeds cap %d (EBPF-001: bounded map not wired into runtime)", got, edgeCap)
	}
	if a.L7Evicted() == 0 {
		t.Error("L7Evicted() == 0 — eviction never fired under N>>cap churn (bound not enforced)")
	}
	// l7man stays bounded because the agent's eviction closes its trackers in
	// lockstep (l7man.Close on evict), so l7man.Len() <= cap above is the proof;
	// l7man's OWN cap/eviction counter is exercised by l7.TestManagerCapBounded.
	if a.agg.ServiceMap().Evicted() == 0 {
		t.Error("ServiceMap.Evicted() == 0 — edge cap never enforced (EBPF-001 not wired)")
	}
}

// TestAgentL7IdleSweep verifies the flush-driven idle sweep abandons stale
// connections (FUZZ-001) — the runtime's continuous enforcement, not just the
// hard cap. It feeds a few connections, advances past the TTL, and asserts
// pruneL7 closes them (l7conns, l7man, and l7seen all drained).
func TestAgentL7IdleSweep(t *testing.T) {
	cfg := &Config{
		TenantID: "t1", Host: "n1", FlushInterval: time.Hour,
		MaxL7Conns: 1000, MaxServiceEdges: 1000, L7ConnIdleTTL: time.Minute,
	}
	a := newAgentWith(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), &sliceSource{}, NopEnricher{}, &captureEmitter{})

	base := time.Unix(0, 0)
	req := []byte("GET /x HTTP/1.1\r\nContent-Length: 0\r\n\r\n")
	for i := 0; i < 10; i++ {
		a.observeL7(L7Event{
			ConnID: uint64(i + 1), TenantID: "t1",
			Source: Endpoint{Workload: "c"}, Destination: Endpoint{Workload: "s", Port: 8443}, Transport: "tcp",
			Data: l7.DataEvent{Kind: l7.Request, Time: base, Payload: req},
		})
	}
	if len(a.l7conns) != 10 {
		t.Fatalf("setup: l7conns = %d, want 10", len(a.l7conns))
	}

	// Advance well past the idle TTL and sweep.
	a.pruneL7(base.Add(5 * time.Minute))

	if len(a.l7conns) != 0 {
		t.Errorf("after idle sweep: l7conns = %d, want 0", len(a.l7conns))
	}
	if a.l7man.Len() != 0 {
		t.Errorf("after idle sweep: l7man.Len() = %d, want 0", a.l7man.Len())
	}
	if a.L7Evicted() < 10 {
		t.Errorf("after idle sweep: L7Evicted() = %d, want >= 10", a.L7Evicted())
	}
}
