package flow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Emitter receives normalized record batches from the collector. The bus
// emitter (emit.go) is the production implementation; tests capture in memory.
type Emitter interface {
	Emit(ctx context.Context, recs []Record) error
}

// Stats are the collector's monotonic counters (probectl observes probectl).
type Stats struct {
	Packets        atomic.Uint64
	Records        atomic.Uint64
	DecodeErrors   atomic.Uint64
	TemplateMisses atomic.Uint64
	QueueDrops     atomic.Uint64
	EmitErrors     atomic.Uint64
	// DroppedRecords counts records LOST after emit retries were exhausted
	// (CORRECT-001) — distinct from EmitErrors (failed flush attempts). Telemetry
	// loss is never silent: it rides the stats snapshot + the periodic stats log.
	DroppedRecords atomic.Uint64
}

// StatsSnapshot is a point-in-time copy for logging/tests.
type StatsSnapshot struct {
	Packets, Records, DecodeErrors, TemplateMisses, QueueDrops, EmitErrors, DroppedRecords uint64
}

// Collector binds the configured UDP listeners, decodes datagrams into
// records, and emits size/time-bounded batches.
//
// Security posture (CLAUDE.md §7 guardrail 12): NetFlow/IPFIX/sFlow are UDP
// export protocols with no transport security of their own, so every datagram
// is treated as untrusted input — decoders are bounds-checked and template
// state is TTL'd and size-capped. Deploy the collector adjacent to exporters
// (management network); records become trusted only by the agent's own tenant
// binding, never by anything the datagram claims.
type Collector struct {
	cfg  *Config
	emit Emitter
	log  *slog.Logger
	dec  *Decoder

	queue chan Record
	stats Stats

	// emit resilience (CORRECT-001): a bounded local retry absorbs transient bus
	// blips before a batch is dropped. The flow agent emits TO the bus, so the
	// bus is the failing dependency — there is no separate bus DLQ to route to
	// (it would hit the same outage); local retry + a dropped-records counter is
	// the honest contract (D4 second clause).
	emitMaxRetries int
	emitRetryBase  time.Duration
	sleep          func(context.Context, time.Duration) // injectable for tests

	mu    sync.Mutex
	conns map[string]net.PacketConn // protocol name -> bound socket
	done  chan struct{}
}

// New validates cfg and builds a collector.
func New(cfg *Config, em Emitter, log *slog.Logger) (*Collector, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if em == nil {
		return nil, errors.New("flow: emitter is required")
	}
	if log == nil {
		log = slog.Default()
	}
	c := &Collector{
		cfg:            cfg,
		emit:           em,
		log:            log,
		dec:            NewDecoder(cfg.TemplateTTL, cfg.MaxTemplates),
		queue:          make(chan Record, cfg.QueueSize),
		emitMaxRetries: 2,
		emitRetryBase:  50 * time.Millisecond,
		conns:          make(map[string]net.PacketConn),
		done:           make(chan struct{}),
	}
	c.sleep = c.defaultSleep
	return c, nil
}

// Start binds the enabled listeners and launches the read + flush loops. It
// returns once everything is listening (tests can then query LocalAddr).
func (c *Collector) Start(ctx context.Context) error {
	type listener struct {
		name string
		l    ListenerConfig
	}
	var ls []listener
	if c.cfg.NetFlow.Enabled {
		ls = append(ls, listener{"netflow", c.cfg.NetFlow})
	}
	if c.cfg.IPFIX.Enabled {
		ls = append(ls, listener{"ipfix", c.cfg.IPFIX})
	}
	if c.cfg.SFlow.Enabled {
		ls = append(ls, listener{"sflow", c.cfg.SFlow})
	}
	for _, l := range ls {
		conn, err := net.ListenPacket("udp", l.l.Listen)
		if err != nil {
			c.Close()
			return fmt.Errorf("flow: listen %s on %q: %w", l.name, l.l.Listen, err)
		}
		if uc, ok := conn.(*net.UDPConn); ok && c.cfg.ReadBufferBytes > 0 {
			// Best-effort: a larger kernel buffer is the first high-volume
			// defense (burst absorption); failure is logged, not fatal.
			if err := uc.SetReadBuffer(c.cfg.ReadBufferBytes); err != nil {
				c.log.Warn("flow: set read buffer failed", "listener", l.name, "error", err.Error())
			}
		}
		c.mu.Lock()
		c.conns[l.name] = conn
		c.mu.Unlock()
		workers := c.cfg.Workers
		if workers < 1 {
			workers = 1
		}
		for i := 0; i < workers; i++ {
			go c.readLoop(ctx, l.name, conn)
		}
		c.log.Info("flow: listening", "protocol", l.name, "addr", conn.LocalAddr().String(),
			"workers", workers, "tenant", c.cfg.TenantID)
	}
	go c.flushLoop(ctx)
	return nil
}

// Run starts the collector and blocks until ctx is canceled.
func (c *Collector) Run(ctx context.Context) error {
	if err := c.Start(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	c.Close()
	return nil
}

// Close shuts the sockets down; in-flight batches flush on the next tick.
func (c *Collector) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for name, conn := range c.conns {
		_ = conn.Close()
		delete(c.conns, name)
	}
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}

// LocalAddr reports the bound address for a protocol listener ("" when not
// bound) — primarily for tests that listen on port 0.
func (c *Collector) LocalAddr(protocol string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if conn, ok := c.conns[protocol]; ok {
		return conn.LocalAddr().String()
	}
	return ""
}

// StatsSnapshot returns a copy of the counters.
func (c *Collector) StatsSnapshot() StatsSnapshot {
	return StatsSnapshot{
		Packets:        c.stats.Packets.Load(),
		Records:        c.stats.Records.Load(),
		DecodeErrors:   c.stats.DecodeErrors.Load(),
		TemplateMisses: c.stats.TemplateMisses.Load(),
		QueueDrops:     c.stats.QueueDrops.Load(),
		EmitErrors:     c.stats.EmitErrors.Load(),
		DroppedRecords: c.stats.DroppedRecords.Load(),
	}
}

// readLoop reads datagrams from one socket, decodes, and enqueues records.
// Queue overflow drops (counted) rather than blocking the socket — at NetFlow
// volumes, back-pressure on a UDP reader only moves the drop into the kernel.
func (c *Collector) readLoop(ctx context.Context, name string, conn net.PacketConn) {
	buf := make([]byte, 65535)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-c.done:
				return
			default:
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			c.stats.DecodeErrors.Add(1)
			continue
		}
		c.stats.Packets.Add(1)
		exporter := exporterHost(addr)
		recs, misses, derr := c.dec.Decode(buf[:n], exporter, time.Now())
		if misses > 0 {
			c.stats.TemplateMisses.Add(uint64(misses))
		}
		if derr != nil {
			c.stats.DecodeErrors.Add(1)
			c.log.Debug("flow: decode failed", "listener", name, "exporter", exporter, "error", derr.Error())
		}
		for i := range recs {
			recs[i].TenantID = c.cfg.TenantID
			recs[i].AgentID = c.cfg.AgentID
			select {
			case c.queue <- recs[i]:
			default:
				c.stats.QueueDrops.Add(1)
			}
		}
	}
}

// flushLoop drains the queue into size/time-bounded batches for the emitter,
// and logs a stats line each interval (probectl observes probectl).
func (c *Collector) flushLoop(ctx context.Context) {
	batch := make([]Record, 0, c.cfg.BatchSize)
	ticker := time.NewTicker(c.cfg.FlushInterval)
	statsTicker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	defer statsTicker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		c.flushBatch(ctx, batch)
		batch = make([]Record, 0, c.cfg.BatchSize)
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-c.done:
			flush()
			return
		case rec := <-c.queue:
			batch = append(batch, rec)
			if len(batch) >= c.cfg.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-statsTicker.C:
			s := c.StatsSnapshot()
			c.log.Info("flow: collector stats",
				"packets", s.Packets, "records", s.Records, "decode_errors", s.DecodeErrors,
				"template_misses", s.TemplateMisses, "queue_drops", s.QueueDrops,
				"emit_errors", s.EmitErrors, "dropped_records", s.DroppedRecords,
				"templates", c.dec.TemplateCount())
		}
	}
}

// flushBatch emits one batch with a bounded local retry; on exhaustion the
// batch is dropped and COUNTED (DroppedRecords) + logged — never silently lost
// (CORRECT-001). Success counts the records.
func (c *Collector) flushBatch(ctx context.Context, batch []Record) {
	if err := c.emitWithRetry(ctx, batch); err != nil {
		c.stats.EmitErrors.Add(1)
		c.stats.DroppedRecords.Add(uint64(len(batch)))
		c.log.Error("flow: emit failed after retries — batch dropped (telemetry loss)",
			"records", len(batch), "attempts", c.emitMaxRetries+1, "error", err.Error(),
			"dropped_records_total", c.stats.DroppedRecords.Load())
		return
	}
	c.stats.Records.Add(uint64(len(batch)))
}

// emitWithRetry attempts the emit up to 1+emitMaxRetries times with jittered
// exponential backoff. A canceled context stops retrying immediately.
func (c *Collector) emitWithRetry(ctx context.Context, batch []Record) error {
	var err error
	for attempt := 0; ; attempt++ {
		if err = c.emit.Emit(ctx, batch); err == nil {
			return nil
		}
		if attempt >= c.emitMaxRetries || ctx.Err() != nil {
			return err
		}
		backoff := c.emitRetryBase << attempt
		c.sleep(ctx, backoff+time.Duration(rand.Int64N(int64(backoff)/2+1)))
	}
}

// defaultSleep waits dur or until ctx is done (overridable in tests).
func (c *Collector) defaultSleep(ctx context.Context, dur time.Duration) {
	t := time.NewTimer(dur)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// exporterHost extracts the exporter IP (no port) from the datagram source.
func exporterHost(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	return host
}
