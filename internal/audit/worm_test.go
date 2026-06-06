package audit

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/imfeelingtheagi/probectl/internal/objectstore"
)

// chainedEvents builds a synthetic, correctly-chained provider stream.
func chainedEvents(n int) []Event {
	out := make([]Event, n)
	prev := genesis
	for i := range out {
		h := "h" + string(rune('0'+i%10)) + "-" + strings.Repeat("x", i%3+1)
		out[i] = Event{Seq: int64(i + 1), Actor: "op", Action: "a", PrevHash: prev, Hash: h}
		prev = h
	}
	return out
}

func sourceOf(events []Event) WormSource {
	return func(_ context.Context, afterSeq int64, limit int) ([]Event, error) {
		var page []Event
		for _, ev := range events {
			if ev.Seq > afterSeq && len(page) < limit {
				page = append(page, ev)
			}
		}
		return page, nil
	}
}

func testLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// U-041: export → verify round-trips; incremental exports build separate
// signed segments and the cross-segment chain verifies end to end.
func TestWormExportAndChainVerify(t *testing.T) {
	all := chainedEvents(7)
	store := objectstore.NewMemory()
	ctx := context.Background()

	// First export sees only the first 4 events.
	w, err := NewWormExporter(sourceOf(all[:4]), store, nil, nil, testLog())
	if err != nil {
		t.Fatal(err)
	}
	if n, err := w.ExportOnce(ctx); err != nil || n != 4 {
		t.Fatalf("first export: n=%d err=%v", n, err)
	}
	// Second export picks up the rest from the derived cursor.
	w.source = sourceOf(all)
	if n, err := w.ExportOnce(ctx); err != nil || n != 3 {
		t.Fatalf("second export: n=%d err=%v", n, err)
	}
	// Idempotent when nothing is new.
	if n, err := w.ExportOnce(ctx); err != nil || n != 0 {
		t.Fatalf("noop export: n=%d err=%v", n, err)
	}
	if err := w.VerifyWORMChain(ctx); err != nil {
		t.Fatalf("verify: %v", err)
	}
	// The public key is published next to the segments.
	if _, err := store.Get(ctx, "worm/audit/provider/signing.pub"); err != nil {
		t.Fatal("verification key not published")
	}
	keys, _ := store.List(ctx, "worm/audit/provider/segment-")
	if len(keys) != 4 { // 2 segments + 2 signatures
		t.Fatalf("objects = %v", keys)
	}
}

// Tampering with an exported segment breaks its signature — detected.
func TestWormTamperedSegmentFailsVerification(t *testing.T) {
	store := objectstore.NewMemory()
	ctx := context.Background()
	w, _ := NewWormExporter(sourceOf(chainedEvents(3)), store, nil, nil, testLog())
	if _, err := w.ExportOnce(ctx); err != nil {
		t.Fatal(err)
	}
	keys, _ := store.List(ctx, "worm/audit/provider/segment-")
	var segKey string
	for _, k := range keys {
		if !strings.HasSuffix(k, ".sig") {
			segKey = k
		}
	}
	obj, _ := store.Get(ctx, segKey)
	tampered := []byte(strings.Replace(string(obj.Data), `"actor":"op"`, `"actor":"evil"`, 1))
	_ = store.Put(ctx, segKey, "application/json", tampered)

	err := w.VerifyWORMChain(ctx)
	if err == nil || !strings.Contains(err.Error(), "signature INVALID") {
		t.Fatalf("tampered segment passed: %v", err)
	}
}

// A purge in the source (events vanish before export) surfaces as a seq gap.
func TestWormDetectsPurgeGap(t *testing.T) {
	store := objectstore.NewMemory()
	ctx := context.Background()
	all := chainedEvents(6)
	purged := append(append([]Event{}, all[:2]...), all[4:]...) // 3 and 4 are gone

	w, _ := NewWormExporter(sourceOf(purged), store, nil, nil, testLog())
	if _, err := w.ExportOnce(ctx); err != nil {
		t.Fatal(err)
	}
	err := w.VerifyWORMChain(ctx)
	if err == nil || !strings.Contains(err.Error(), "GAP") {
		t.Fatalf("purged events not detected: %v", err)
	}
}
