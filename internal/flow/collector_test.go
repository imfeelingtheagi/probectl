package flow

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"
)

// captureEmitter records emitted batches for assertions.
type captureEmitter struct {
	mu   sync.Mutex
	recs []Record
}

func (c *captureEmitter) Emit(_ context.Context, recs []Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recs = append(c.recs, recs...)
	return nil
}

func (c *captureEmitter) snapshot() []Record {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Record(nil), c.recs...)
}

func testConfig() *Config {
	cfg := Default()
	cfg.TenantID = "t-acme"
	cfg.AgentID = "agent-1"
	cfg.NetFlow = ListenerConfig{Enabled: true, Listen: "127.0.0.1:0"}
	cfg.IPFIX = ListenerConfig{Enabled: false}
	cfg.SFlow = ListenerConfig{Enabled: true, Listen: "127.0.0.1:0"}
	cfg.FlushInterval = 30 * time.Millisecond
	cfg.BatchSize = 8
	return cfg
}

// TestCollectorEndToEnd sends real datagrams (v5 + v9 template/data + sFlow)
// over UDP and asserts tenant-bound records reach the emitter.
func TestCollectorEndToEnd(t *testing.T) {
	em := &captureEmitter{}
	c, err := New(testConfig(), em, slog.Default())
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer c.Close()

	nfAddr, sfAddr := c.LocalAddr("netflow"), c.LocalAddr("sflow")
	if nfAddr == "" || sfAddr == "" {
		t.Fatalf("listeners not bound: nf=%q sf=%q", nfAddr, sfAddr)
	}
	if c.LocalAddr("ipfix") != "" {
		t.Fatal("ipfix listener bound although disabled")
	}

	send := func(addr string, pkt []byte) {
		conn, err := net.Dial("udp", addr)
		if err != nil {
			t.Fatalf("dial %s: %v", addr, err)
		}
		defer conn.Close()
		if _, err := conn.Write(pkt); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	unix := uint32(time.Now().Unix())
	// v5 + v9 share the netflow socket (version-sniffed).
	send(nfAddr, buildNF5(1000, unix, 0, []nf5rec{{
		src: [4]byte{10, 0, 0, 1}, dst: [4]byte{10, 0, 0, 2}, pkts: 1, bytes: 64, proto: 6, sport: 1, dport: 2,
	}}))
	send(nfAddr, buildNF9Template(1000, unix, 7, 260, nf9V4Fields))
	send(nfAddr, buildNF9Data(1000, unix, 7, 260, [][]byte{
		nf9V4Row([4]byte{10, 0, 0, 3}, [4]byte{10, 0, 0, 4}, 5, 6, 17, 128, 2, 0, 0),
	}))
	hdr := buildEthIPv4TCP(0, [4]byte{10, 0, 0, 5}, [4]byte{10, 0, 0, 6}, 80, 1024, 0x10, 6)
	send(sfAddr, buildSFlowRaw(64, 1, 2, hdr, false, false))

	deadline := time.Now().Add(3 * time.Second)
	for {
		recs := em.snapshot()
		if len(recs) >= 3 {
			byProto := map[string]int{}
			for _, r := range recs {
				byProto[r.Protocol]++
				if r.TenantID != "t-acme" || r.AgentID != "agent-1" {
					t.Fatalf("record not tenant-bound: %+v", r)
				}
				if r.Exporter != "127.0.0.1" {
					t.Fatalf("exporter = %q, want 127.0.0.1", r.Exporter)
				}
			}
			if byProto[ProtoNetFlow5] < 1 || byProto[ProtoNetFlow9] < 1 || byProto[ProtoSFlow5] < 1 {
				t.Fatalf("missing protocols in %v", byProto)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out; got %d records, stats %+v", len(recs), c.StatsSnapshot())
		}
		time.Sleep(10 * time.Millisecond)
	}

	s := c.StatsSnapshot()
	if s.Packets < 4 || s.Records < 3 {
		t.Errorf("stats = %+v", s)
	}
	// Garbage must be counted, not crash anything.
	send(nfAddr, []byte{0xDE, 0xAD})
	deadline = time.Now().Add(2 * time.Second)
	for c.StatsSnapshot().DecodeErrors == 0 {
		if time.Now().After(deadline) {
			t.Fatal("decode error not counted for garbage datagram")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestCollectorValidation: bad configs and nil emitters are refused.
func TestCollectorValidation(t *testing.T) {
	cfg := testConfig()
	if _, err := New(cfg, nil, nil); err == nil {
		t.Error("nil emitter accepted")
	}
	cfg.TenantID = ""
	if _, err := New(cfg, &captureEmitter{}, nil); err == nil {
		t.Error("missing tenant accepted")
	}
	cfg = testConfig()
	cfg.NetFlow.Enabled, cfg.IPFIX.Enabled, cfg.SFlow.Enabled = false, false, false
	if _, err := New(cfg, &captureEmitter{}, nil); err == nil {
		t.Error("no listeners accepted")
	}
}
