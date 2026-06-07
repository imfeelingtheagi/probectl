# Data Processing Addendum — template (U-065)

> **DRAFT FOR COUNSEL REVIEW — not legal advice, not an executed agreement.**
> Prepared to accelerate enterprise procurement; counsel owns final text.
> The factual basis (architecture, data categories, egress gates) is
> code-enforced and cited; the legal language is a starting skeleton.

## 0. The unusual (and favorable) starting point

probectl is **operator-hosted**: the customer runs every component inside
their own infrastructure, and the software **never phones home**
(CLAUDE.md §7.2 — no telemetry beacons; license verification is offline
local math). In the default deployment the vendor **processes no customer
personal data at all** — there is nothing to flow to the vendor. This DPA
therefore has two parts: Part A (the default, near-empty processing
reality) and Part B (the **optional** remote-AI egress the customer may
elect, the only feature that sends tenant data off-network — and then to a
*customer-chosen* model provider, not to the vendor).

## Part A — default deployment (vendor processes nothing)

1. **Roles.** Customer/operator: controller (and processor for its own
   tenants, where the customer is an MSP). Vendor: **neither controller
   nor processor of telemetry** — supplier of software only.
2. **Vendor-side processing.** None by default. Optional, customer-initiated
   exchanges only: support bundles (sanitized by `internal/support`;
   customer reviews before sending) and vulnerability reports per
   SECURITY.md. [Counsel: a lightweight processing clause may still be
   wanted for support artifacts.]
3. **Data categories handled by the SOFTWARE inside the customer's
   environment** (for the customer's own records/ROPA): network telemetry —
   flow 5-tuples, probe results, paths, device/BGP events, L7 *metadata*
   (only under the double-keyed consent of U-003; bodies redacted by
   default); operational identities — usernames/emails of the customer's
   operators (SSO), agent/host identifiers; audit records. Multi-tenant
   isolation is storage-layer-enforced and CI-gated
   (`docs/security/threat-model.md` B1).
4. **Security measures (Annex II equivalent).** TLS on every channel; mTLS
   agent transport; RLS/row-policy tenant isolation; envelope encryption of
   secrets; tamper-evident audit + signed WORM export; backups documented
   (`docs/ops/backup-restore.md`) and operated by the customer. Evidence:
   `docs/compliance/soc2-mapping.md`.
5. **Data subject rights / deletion.** Tenant erasure is built in and
   attested store-by-store (U-027, `internal/tenantlife`); export likewise.
   The customer executes both — the vendor holds no copy.
6. **International transfers.** None by the vendor by default (no data
   reaches the vendor). Customer-side transfers are the customer's own.
7. **Subprocessors.** None by default — see
   [subprocessors.md](subprocessors.md).

## Part B — optional remote-AI egress annex (only if the customer enables it)

Factual basis: `docs/ai-egress.md` (C7) — code-enforced, audited.

1. **Trigger.** Disabled by default (air-gapped builtin model). Activation
   requires the operator's explicit boot acknowledgment
   (`PROBECTL_AI_EGRESS_ACK=yes-send-tenant-data-to-the-remote-model`) AND
   per-tenant default-deny consent.
2. **Processor.** The **customer-selected model provider** (e.g. Anthropic,
   OpenAI, Azure, Bedrock) under the **customer's own agreement with that
   provider** — the vendor is not in the data path and sets no retention
   terms. [Counsel: this annex obligates the *customer* to maintain a DPA
   with their provider; it does not create a vendor subprocessor.]
3. **Categories transferred.** Redacted RCA evidence summaries (C8
   redaction: IPs and detected secrets masked, hostnames per policy) —
   never credentials, raw telemetry rows, packet payloads, or another
   tenant's data (tenant-first scoping). Each call's outbound categories
   are recorded in the tenant's audit stream.
4. **Safeguards.** TLS via the hardened client; redaction pass; per-tenant
   consent; per-call audit (`ai.remote_egress` events).
5. **Local alternative.** Ollama/vLLM in-network endpoints provide the same
   capability with zero egress; recommended for regulated tenants.

## Annex I — parties & processing details
[Counsel completes: parties, duration, subject matter; categories per A.3/B.3.]

## Annex III — subprocessor list
Incorporate [subprocessors.md](subprocessors.md) by reference.
