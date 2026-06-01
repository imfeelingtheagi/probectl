package l7

// Tracker parses one connection: it detects the protocol from the first request
// bytes (plus the destination-port hint) and delegates to the matching parser.
type Tracker struct {
	dstPort uint32
	proto   string
	parser  Parser
}

// NewTracker returns a Tracker for a connection to dstPort.
func NewTracker(dstPort uint32) *Tracker { return &Tracker{dstPort: dstPort} }

// OnData feeds one plaintext chunk and returns any completed calls.
func (t *Tracker) OnData(d DataEvent) []Call {
	if t.parser == nil {
		if d.Kind != Request {
			return nil // need request bytes to detect; a stray early response is dropped
		}
		t.proto = Detect(d.Payload, t.dstPort)
		t.parser = parserFor(t.proto)
		if t.parser == nil {
			return nil
		}
	}
	return t.parser.OnData(d)
}

// Flush returns any calls still buffered (e.g. on connection close).
func (t *Tracker) Flush() []Call {
	if t.parser == nil {
		return nil
	}
	return t.parser.Flush()
}

// Protocol returns the detected protocol ("" until detected).
func (t *Tracker) Protocol() string { return t.proto }

func parserFor(proto string) Parser {
	switch proto {
	case ProtoHTTP1:
		return newHTTP1Parser()
	case ProtoHTTP2, ProtoGRPC:
		return newHTTP2Parser()
	case ProtoDNS:
		return newDNSParser()
	case ProtoKafka:
		return newKafkaParser()
	default:
		return nil
	}
}

// Manager multiplexes many connections, keyed by a connection id from the
// capture layer (e.g. a socket cookie / fd identity).
type Manager struct {
	conns map[uint64]*Tracker
}

// NewManager returns an empty Manager.
func NewManager() *Manager { return &Manager{conns: map[uint64]*Tracker{}} }

// OnData routes a data event for connID (to dstPort) to its tracker.
func (m *Manager) OnData(connID uint64, dstPort uint32, d DataEvent) []Call {
	t := m.conns[connID]
	if t == nil {
		t = NewTracker(dstPort)
		m.conns[connID] = t
	}
	return t.OnData(d)
}

// Close flushes and removes a connection's tracker.
func (m *Manager) Close(connID uint64) []Call {
	t := m.conns[connID]
	if t == nil {
		return nil
	}
	delete(m.conns, connID)
	return t.Flush()
}

// Len returns the number of tracked connections.
func (m *Manager) Len() int { return len(m.conns) }
