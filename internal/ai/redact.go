package ai

import (
	"fmt"
	"hash/fnv"
	"net/netip"
	"regexp"
	"strings"
)

// Pre-egress redaction (U-013, C8): before a prompt leaves the network to a
// REMOTE model, IPs (configurable), hostnames (per policy) and obvious
// secrets/tokens (always) are masked. Masking is deterministic per value —
// the same IP becomes the same token within and across calls — so the model
// can still correlate evidence without ever seeing the value. The local
// paths (builtin model, loopback Ollama/vLLM) are never redacted.

// RedactionPolicy selects what is masked before remote egress. Secrets are
// ALWAYS masked for a remote model regardless of policy.
type RedactionPolicy struct {
	MaskIPs       bool
	MaskHostnames bool
}

// DefaultRedaction is the remote-model default: IPs masked, hostnames kept
// (they are usually the subject of the question), secrets always masked.
var DefaultRedaction = RedactionPolicy{MaskIPs: true, MaskHostnames: false}

var (
	// Always-on secret shapes: bearer/authorization values, key=value
	// credentials, AWS access key IDs, PEM blocks.
	reBearer = regexp.MustCompile(`(?i)\b(bearer|authorization:)\s+[A-Za-z0-9._~+/=-]{8,}`)
	reKV     = regexp.MustCompile(`(?i)\b((?:api[_-]?key|access[_-]?key|secret|token|password|passwd|pwd)\s*[=:]\s*)\S+`)
	reAKIA   = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	rePEM    = regexp.MustCompile(`-----BEGIN [A-Z0-9 ]+-----[\s\S]*?-----END [A-Z0-9 ]+-----`)

	reIPv4 = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}(?:/\d{1,2})?\b`)
	// IPv6 candidates: two-plus colons over hex groups (covers :: compression
	// and zones); each candidate is then VALIDATED with netip.ParseAddr, so
	// times ("12:30") and C++ scope operators never match.
	reIPv6Candidate = regexp.MustCompile(`(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}(?:%[0-9a-zA-Z]+)?(?:/\d{1,3})?|::[fF]{4}:(?:\d{1,3}\.){3}\d{1,3}`)

	reHostname = regexp.MustCompile(`\b(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z]{2,})\b`)
)

// redactText applies the policy to one string.
func redactText(s string, pol RedactionPolicy) string {
	// Secrets first (always), so an IP inside a token is gone either way.
	s = rePEM.ReplaceAllString(s, "[pem:redacted]")
	s = reBearer.ReplaceAllStringFunc(s, func(m string) string { return mask("secret", m) })
	s = reKV.ReplaceAllStringFunc(s, func(m string) string {
		kv := reKV.FindStringSubmatch(m)
		return kv[1] + mask("secret", m)
	})
	s = reAKIA.ReplaceAllStringFunc(s, func(m string) string { return mask("secret", m) })

	if pol.MaskIPs {
		s = reIPv6Candidate.ReplaceAllStringFunc(s, func(m string) string {
			bare := strings.TrimSuffix(m, "/"+lastSlashPart(m))
			if _, err := netip.ParseAddr(strings.TrimSpace(bare)); err != nil {
				return m // not a real v6 address (e.g. a time, "::" prose)
			}
			return mask("ip", m)
		})
		s = reIPv4.ReplaceAllStringFunc(s, func(m string) string { return mask("ip", m) })
	}
	if pol.MaskHostnames {
		s = reHostname.ReplaceAllStringFunc(s, func(m string) string { return mask("host", m) })
	}
	return s
}

// lastSlashPart returns what follows the final '/' (the CIDR suffix), or ""
// when there is none.
func lastSlashPart(s string) string {
	if i := strings.LastIndexByte(s, '/'); i >= 0 {
		return s[i+1:]
	}
	return ""
}

// mask renders a stable token for value: same value, same token — correlation
// survives, the value does not.
func mask(class, value string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(value))
	return fmt.Sprintf("[%s:%08x]", class, h.Sum32())
}

// redactSynthesisInput deep-copies the input with the policy applied to the
// question and every evidence title/summary. The caller's evidence is never
// mutated (the local pipeline keeps the raw values for citation display).
func redactSynthesisInput(in SynthesisInput, pol RedactionPolicy) SynthesisInput {
	out := SynthesisInput{Question: redactText(in.Question, pol)}
	out.Evidence = make([]Evidence, len(in.Evidence))
	for i, e := range in.Evidence {
		e.Title = redactText(e.Title, pol)
		e.Summary = redactText(e.Summary, pol)
		out.Evidence[i] = e
	}
	return out
}
