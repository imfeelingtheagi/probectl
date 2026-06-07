# Subprocessor disclosure (U-065)

> **DRAFT FOR COUNSEL REVIEW.** Factual statements below are code-enforced
> and cited.

## Default deployment: none

probectl is operator-hosted and never phones home (CLAUDE.md §7.2): no
telemetry, analytics, crash reporting, license callbacks, or managed
infrastructure touches customer data. **The vendor engages no
subprocessors for customer personal data, because no customer personal
data reaches the vendor.**

| Subprocessor | Purpose | Data | Location |
|---|---|---|---|
| — | — | — | — |

## Customer-elected processors (not vendor subprocessors)

If the customer enables optional remote-AI egress (`docs/ai-egress.md`,
default off, triple-gated), the customer transmits redacted RCA evidence
directly to the **model provider the customer chose and contracted with**:

| Provider (customer-selected) | Engaged by | Data | Gate |
|---|---|---|---|
| Anthropic / OpenAI / Azure OpenAI / AWS Bedrock — or none (local Ollama/vLLM) | the customer, under the customer's own agreement | C8-redacted evidence summaries, per consenting tenant | boot ack + per-tenant consent + per-call audit |

The vendor is not a party to that flow and receives nothing from it.

## Adjacent (non-personal-data) suppliers, for completeness

Source hosting + CI (GitHub) and the artifact registry (ghcr.io) handle
**code and signed release artifacts**, not customer data. Optional
open-data feeds (RouteViews, RIPE, CT logs, …) are *inbound* public data,
fetched read-only by the customer's deployment (guardrail §7.10).

## Change notice

Any future true subprocessor requires updating this file in-repo (it is
versioned) and notice per the DPA's change-notification clause.
