// SPDX-License-Identifier: LicenseRef-probectl-TBD

package l7

import (
	"encoding/binary"
	"time"

	"github.com/twmb/franz-go/pkg/kmsg"
)

// kafkaParser parses the Kafka wire protocol at the header level: it matches a
// response to its request by correlation id and reports the API name (Produce,
// Fetch, …) as the method plus the latency. Per-partition error codes live deep
// in version-specific response bodies and are out of scope for S21 (documented).
type kafkaParser struct {
	reqBuf, respBuf []byte
	pending         map[int32]kafkaReq
}

type kafkaReq struct {
	api   string
	start time.Time
	bytes uint64
}

func newKafkaParser() *kafkaParser { return &kafkaParser{pending: map[int32]kafkaReq{}} }

func (p *kafkaParser) OnData(d DataEvent) []Call {
	if d.Kind == Request {
		p.reqBuf = append(p.reqBuf, d.Payload...)
		for {
			msg, rest, ok := scanKafkaMessage(p.reqBuf)
			if !ok {
				break
			}
			p.reqBuf = rest
			// [int32 size][int16 api_key][int16 api_version][int32 correlation_id]…
			if len(msg) >= 12 {
				apiKey := int16(binary.BigEndian.Uint16(msg[4:6]))
				corr := int32(binary.BigEndian.Uint32(msg[8:12]))
				p.pending[corr] = kafkaReq{api: kmsg.NameForKey(apiKey), start: d.Time, bytes: uint64(len(msg))}
			}
		}
		return nil
	}

	var calls []Call
	p.respBuf = append(p.respBuf, d.Payload...)
	for {
		msg, rest, ok := scanKafkaMessage(p.respBuf)
		if !ok {
			break
		}
		p.respBuf = rest
		// [int32 size][int32 correlation_id]…
		if len(msg) >= 8 {
			corr := int32(binary.BigEndian.Uint32(msg[4:8]))
			if req, ok := p.pending[corr]; ok {
				delete(p.pending, corr)
				calls = append(calls, Call{
					Protocol:  ProtoKafka,
					Method:    req.api,
					Start:     req.start,
					Latency:   d.Time.Sub(req.start),
					ReqBytes:  req.bytes,
					RespBytes: uint64(len(msg)),
				})
			}
		}
	}
	return calls
}

func (p *kafkaParser) Flush() []Call { return nil }

// scanKafkaMessage returns one complete length-prefixed message ([int32 size]
// [payload]), or ok=false if not fully captured. The returned slice includes the
// 4-byte size prefix.
func scanKafkaMessage(buf []byte) (msg, rest []byte, ok bool) {
	if len(buf) < 4 {
		return nil, buf, false
	}
	size := int(binary.BigEndian.Uint32(buf[0:4]))
	total := 4 + size
	if size < 0 || size > 100<<20 || total > len(buf) { // 100 MiB sanity cap
		return nil, buf, false
	}
	return buf[:total], buf[total:], true
}
