// SPDX-License-Identifier: LicenseRef-probectl-TBD

package canary

import (
	"bytes"
	"testing"
)

func TestA2AFrameRoundTrip(t *testing.T) {
	token := []byte("abcdefgh")
	req := encodeA2AReq(token, 9, 111)
	if len(req) != a2aReqLen {
		t.Fatalf("request length = %d, want %d", len(req), a2aReqLen)
	}
	reply := makeA2AReply(req, 222, 333)
	if len(reply) != a2aReplyLen {
		t.Fatalf("reply length = %d, want %d", len(reply), a2aReplyLen)
	}
	rep, ok := parseA2AReply(reply)
	if !ok || !bytes.Equal(rep.token, token) || rep.seq != 9 || rep.t1 != 111 || rep.t2 != 222 || rep.t3 != 333 {
		t.Fatalf("parsed %+v ok=%v", rep, ok)
	}
	if _, ok := parseA2AReply(reply[:10]); ok {
		t.Error("short buffer should not parse")
	}
}
