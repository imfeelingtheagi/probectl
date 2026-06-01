package l7

import (
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestDNSQueryResponse(t *testing.T) {
	p := newDNSParser()
	q := new(dns.Msg)
	q.SetQuestion("example.com.", dns.TypeA)
	q.Id = 0x1234
	qb, err := q.Pack()
	if err != nil {
		t.Fatal(err)
	}
	r := new(dns.Msg)
	r.SetReply(q)
	r.Rcode = dns.RcodeNameError // NXDOMAIN
	rb, err := r.Pack()
	if err != nil {
		t.Fatal(err)
	}

	t0 := time.Unix(5, 0)
	if c := p.OnData(DataEvent{Kind: Request, Time: t0, Payload: qb}); len(c) != 0 {
		t.Fatalf("query emitted %d calls, want 0", len(c))
	}
	calls := p.OnData(DataEvent{Kind: Response, Time: t0.Add(3 * time.Millisecond), Payload: rb})
	if len(calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(calls))
	}
	c := calls[0]
	if c.Protocol != ProtoDNS || c.Method != "A" || c.Resource != "example.com." || c.Status != "NXDOMAIN" || !c.Error {
		t.Errorf("call = %+v", c)
	}
	if c.Latency != 3*time.Millisecond {
		t.Errorf("latency = %v", c.Latency)
	}
}
