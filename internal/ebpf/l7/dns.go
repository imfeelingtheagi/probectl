package l7

import (
	"time"

	"github.com/miekg/dns"
)

// dnsParser parses DNS (UDP payloads, or TCP with a 2-byte length prefix). It
// matches a response to its query by the DNS message id and reports the qtype as
// the method, the qname as the resource, and the rcode as the status.
type dnsParser struct {
	pending map[uint16]dnsReq
}

type dnsReq struct {
	qname string
	qtype string
	start time.Time
	bytes uint64
}

func newDNSParser() *dnsParser { return &dnsParser{pending: map[uint16]dnsReq{}} }

func (p *dnsParser) OnData(d DataEvent) []Call {
	msg, ok := unpackDNS(d.Payload)
	if !ok {
		return nil
	}
	// Use the QR bit, not the capture direction, to classify (robust for UDP).
	if !msg.Response {
		if len(msg.Question) > 0 {
			q := msg.Question[0]
			p.pending[msg.Id] = dnsReq{
				qname: q.Name,
				qtype: dns.TypeToString[q.Qtype],
				start: d.Time,
				bytes: uint64(len(d.Payload)),
			}
		}
		return nil
	}

	req, ok := p.pending[msg.Id]
	if !ok {
		return nil
	}
	delete(p.pending, msg.Id)
	return []Call{{
		Protocol:  ProtoDNS,
		Method:    req.qtype,
		Resource:  req.qname,
		Status:    dns.RcodeToString[msg.Rcode],
		Error:     msg.Rcode != dns.RcodeSuccess,
		Start:     req.start,
		Latency:   d.Time.Sub(req.start),
		ReqBytes:  req.bytes,
		RespBytes: uint64(len(d.Payload)),
	}}
}

func (p *dnsParser) Flush() []Call { return nil }

func unpackDNS(b []byte) (*dns.Msg, bool) {
	try := func(data []byte) (*dns.Msg, bool) {
		m := new(dns.Msg)
		if err := m.Unpack(data); err != nil {
			return nil, false
		}
		return m, true
	}
	if m, ok := try(b); ok {
		return m, true
	}
	if len(b) > 2 { // TCP framing: 2-byte length prefix
		if m, ok := try(b[2:]); ok {
			return m, true
		}
	}
	return nil, false
}
