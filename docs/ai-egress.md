# Remote-AI egress: what leaves the network, and the three gates (U-013)

probectl's default AI engine is **air-gapped** (the deterministic builtin; a
loopback Ollama/vLLM is equally local). Configuring a **remote** model
endpoint (`openai` / `anthropic`, or `ollama` pointed at a non-loopback host)
changes the sovereignty story: per RCA question, data leaves the operator's
network. This page is the disclosure — it is written to be attached to a DPA.

## Exactly what is sent, per remote call

One HTTPS POST to the configured endpoint containing:

- the user's **question text**, verbatim¹;
- per evidence item (max `PROBECTL_AI_MAX_EVIDENCE`, default 50): its ID,
  **plane** (network/bgp/flow/device/ebpf/incident…), **severity**,
  **title**, **summary**, and timestamp¹;
- the system prompt (static probectl text) and model name.

Never sent: credentials/tokens (the API key authenticates the call itself and
is stored as an S41 secret reference), raw telemetry rows, packet payloads,
database contents, or anything outside the caller's tenant + RBAC scope (the
evidence is gathered tenant-first through the S23 engine).

¹ After the C8 redaction pass (see `internal/ai/redact.go`): secrets are
ALWAYS masked; IPs and free-text PII (emails, phone numbers, MACs —
AIRCA-002) by default; hostnames and operator custom patterns
(`PROBECTL_AI_REDACT_PATTERNS`) per policy — before anything leaves.

**One gate, three doors (Sprint 20 — AIRCA-001/005):** the same
`EgressGate` (consent + redaction + audit) covers the remote RCA model,
**MCP tool results** (the MCP caller is an external AI client; see
`docs/mcp.md`), and the **test-authoring** model. The audit event carries
`surface = rca | mcp | author`. No AI call path exists outside the gate —
a static test (`TestNoAIClientOutsideGate`) enforces it.

**Processing by the model provider is governed by YOUR agreement with that
provider** — probectl sets no retention terms on the remote side. DPA inputs:
processor = the model provider; data categories = the list above; transfer
trigger = each `/v1/ai/*` query by a consenting tenant; safeguards = TLS
(hardened client), redaction pass, per-tenant consent, audit trail.

## The three gates

1. **Operator acknowledgment (boot-time, fail-closed).** A remote endpoint
   refuses to start until the operator sets
   `PROBECTL_AI_EGRESS_ACK=yes-send-tenant-data-to-the-remote-model` —
   the off-network flow must be a deliberate decision, never a default.
2. **Per-tenant consent (call-time, default deny).**
   `tenant_governance.ai_remote_egress` (provider governance API/console)
   defaults to **false**; the analyzer refuses remote synthesis for a
   non-consenting tenant (`ErrEgressDenied`). The builtin/loopback path is
   exempt and keeps working for everyone.
3. **Audit (every call).** Each remote call appends `ai.remote_egress` to the
   tenant's tamper-evident audit stream: endpoint, model, evidence count and
   the **data categories (planes)** that left — never the content itself.

## Turning it on

```sh
PROBECTL_AI_MODEL_PROVIDER=anthropic
PROBECTL_AI_MODEL_ENDPOINT=https://api.anthropic.com
PROBECTL_AI_MODEL_TOKEN=vault:ai/anthropic#key
PROBECTL_AI_EGRESS_ACK=yes-send-tenant-data-to-the-remote-model
# then, per tenant that may use it:
# PUT /provider/v1/tenants/{id}/governance  {"ai_remote_egress": true, ...}
```


## Remote-provider resilience (Sprint 21 — AIRCA-004)

A slow or down provider degrades RCA gracefully, never takes it down. The
configured-model path is wrapped (`internal/ai.ResilientModel`) with:

- **circuit breaker** (`internal/breaker`, U-078): 3 consecutive failures
  open the circuit for 30s — calls short-circuit instead of stacking
  timeouts;
- **timeout**: the configured model timeout (`PROBECTL_AI_MODEL_TIMEOUT`)
  is ctx-enforced at the wrapper regardless of client settings;
- **response cache**: identical question+evidence within 10 minutes is
  answered without a provider round-trip (256 entries). Evidence IDs are
  session-random (U-037), so entries key on CONTENT and cached citations
  are remapped positionally onto the current session's IDs — citation
  grounding still validates them;
- **graceful degradation**: on breaker-open/timeout/error the air-gapped
  BUILTIN answers instead — the answer carries `degraded: true` and the
  root cause is prefixed `PARTIAL RESULT — remote model unavailable (…)`,
  with the builtin's own grounded citations (RED-005 holds while degraded).

The consent gate is unchanged: a remote-CONFIGURED deployment requires
tenant consent before any synthesis, including cache hits and fallback
answers (strict reading: the configuration, not the individual packet,
is the consent subject). The builtin default path is never wrapped.
