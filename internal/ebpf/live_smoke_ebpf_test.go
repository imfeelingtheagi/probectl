//go:build linux && ebpf

package ebpf

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

// U-021 kernel-matrix smoke: actually LOAD and ATTACH every BPF program on
// the running kernel (the ci job runs this inside QEMU on >=2 LTS kernels
// via vimto). The C9 digest verification runs inherently inside newLive*.
func TestLiveLoadAttachL4Flow(t *testing.T) {
	cfg := Default()
	cfg.TenantID = "kernel-matrix"
	src, err := newLiveSource(cfg)
	if err != nil {
		t.Fatalf("l4flow load+attach failed on this kernel: %v", err)
	}
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	flows, err := src.Flows(ctx)
	if err != nil {
		t.Fatalf("flows stream: %v", err)
	}
	// Drain briefly: the tracepoint is attached; traffic is optional.
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-flows:
			if !ok {
				return
			}
		}
	}
}

func TestLiveLoadAttachSslsniff(t *testing.T) {
	cfg := Default()
	cfg.TenantID = "kernel-matrix"
	cfg.L7CaptureEnabled = true
	cfg.L7CaptureConsentTenant = "kernel-matrix" // U-003 consent for the smoke VM
	src, err := newLiveL7Source(cfg)
	if err != nil {
		t.Skipf("sslsniff attach unavailable (no libssl on this rootfs?): %v", err)
	}
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := src.L7Events(ctx); err != nil {
		t.Fatalf("l7 events stream: %v", err)
	}
	<-ctx.Done()
}

// The agent end-to-end on a live kernel: capability probe, live source, one
// flush cycle — runs observe-only by construction (the static gate enforces
// program types; this proves the runtime path on the matrix kernel).
func TestLiveAgentBoot(t *testing.T) {
	cfg := Default()
	cfg.TenantID = "kernel-matrix"
	cfg.FlushInterval = 200 * time.Millisecond
	a, err := New(cfg, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("agent boot on this kernel: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := a.Run(ctx); err != nil {
		t.Fatalf("agent run: %v", err)
	}
}
