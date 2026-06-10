# Using the AI: first question → your own model → your own tools

This is the task-ordered walkthrough of probectl's AI surface. The reference
depth lives in [`ai-rca.md`](ai-rca.md), [`ai-query.md`](ai-query.md),
[`ai-egress.md`](ai-egress.md), and [`mcp.md`](mcp.md) — this page just walks
you through *doing* it, in the order an operator actually does.

**The honesty contract first**, because it explains everything below: the
assistant answers with **citations to real signals you are allowed to see**, it
runs **air-gapped by default** (an in-process deterministic synthesizer — no
language model is contacted until you configure one), and it **never acts** —
it explains, and at most *proposes* (a human approves). Tenant first, then
RBAC, on every path: the model never sees data the asking user couldn't query
directly.

## 1. Ask your first question (zero setup)

The builtin engine works the moment the control plane is up — nothing to
enable. In the UI it's **Ask (AI)** in the nav (`/ask`). Over the API:

```sh
curl -sS --cacert ./certs/ca.crt -H 'Content-Type: application/json' \
  -d '{"question": "why is checkout slow?", "subject": "checkout"}' \
  https://127.0.0.1:8443/v1/ai/ask
```

The body takes `question` (required, 1–2000 chars) and an optional `subject`
to focus the evidence search. (On the from-source dev rig from
[`getting-started.md`](getting-started.md) this works with plain curl; a
production deployment needs your session/bearer auth like any `/v1` call.)

## 2. Read the answer like an operator

The response is a cited verdict, not prose:

- `root_cause` — the one-line probable cause; `root_cause_citations` are the
  validated citations grounding *that headline claim specifically*, and
  `root_cause_grounded: false` means the model's own headline was rejected
  for citing nothing real (the claim is replaced — never shown ungrounded).
- `findings` + `evidence` — every supporting claim links to an underlying
  signal (an incident, a change event); follow the IDs into the UI.
- `confidence` — scored, not vibes; `insufficient_evidence: true` is the
  assistant saying "I don't know" — a *feature*: it refuses to invent a story
  when the planes are quiet.
- `degraded: true` — your remote model was down and the air-gapped builtin
  answered instead (the platform never lets a dead AI provider take down RCA).

What feeds it today: **incidents** (already correlated across planes) and
**change events** — the two evidence sources wired in the shipped engine
([`ai-rca.md`](ai-rca.md) is explicit about what's live vs. seams).

## 3. Upgrade the prose: a model on your own hardware

The builtin ranks and cites deterministically; a local model writes nicer
narrative — and a **loopback** model changes nothing about sovereignty: no
acknowledgment, no consent, no egress.

```sh
# Ollama, same host:
ollama pull llama3.1
PROBECTL_AI_MODEL_PROVIDER=ollama \
PROBECTL_AI_MODEL_ENDPOINT=http://127.0.0.1:11434 \
PROBECTL_AI_MODEL_NAME=llama3.1 \
  ./bin/probectl-control
```

vLLM is the same idea through the `openai` adapter (there is deliberately no
`vllm` provider — vLLM speaks the OpenAI-compatible API on `:8000`). Both
recipes, plus OpenAI/Anthropic/Azure, are in
[`ai-rca.md` → Copy-paste recipes](ai-rca.md#copy-paste-recipes).

## 4. A remote model is a decision, not a config value

Pointing at a non-loopback endpoint means tenant evidence leaves your network,
so two gates have to open — an **operator acknowledgment** at boot
(`PROBECTL_AI_EGRESS_ACK=yes-send-tenant-data-to-the-remote-model`) and a
**per-tenant consent bit** (`tenant_governance.ai_remote_egress`,
default-deny) — and every remote call lands in the tenant's tamper-evident
audit stream as `ai.remote_egress` (categories of what was sent, never the
content). Until both gates are open, calls fail with a denial message that
says exactly this. The full runbook for both editions is
[`ai-egress.md` → Turning it on](ai-egress.md#turning-it-on).

## 5. Hand the map to your own AI (MCP)

The same tenant-then-RBAC boundary is exposed as an **MCP server**, so Claude
Desktop, Claude Code, or any MCP client can query *your* network as tools:
mint a token (`probectl-control mcp-token --user <uuid>`), drop the stdio
config into the client, and eight tools appear — six read-only queries,
`explain_degradation` (rides the same egress gate), and `propose_remediation`
(proposal-only; a human approves in probectl). The filled-in Claude Desktop
config and the TLS HTTP-bridge variant are in [`mcp.md`](mcp.md).

```json
{
  "mcpServers": {
    "probectl": {
      "command": "/usr/local/bin/probectl-control",
      "args": ["mcp-stdio"],
      "env": {
        "PROBECTL_MCP_TOKEN": "<the value mcp-token printed>",
        "PROBECTL_DATABASE_URL": "postgres://probectl:probectl@localhost:5432/probectl?sslmode=disable"
      }
    }
  }
}
```

## 6. What it deliberately will not do

No autonomous actions (proposals only, human-gated). No cross-tenant reads —
structurally, not by prompt. No silent egress — the gates above, audited. No
answer persistence unless you opt in (`PROBECTL_AI_PERSIST_ANSWERS`, default
off). No hallucinated citations — ungrounded claims are replaced, and "I
don't know" is a first-class answer.

## See also

[`ai-authoring.md`](ai-authoring.md) — describe a test in plain English, get a
validated config proposal. [`ai-query.md`](ai-query.md) — the semantic query
layer underneath all of this and its two-level scoping.
