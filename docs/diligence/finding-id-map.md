# Finding-ID reconciliation map

Two diligence registers exist:

1. **U-register** (`U-001`…`U-094`) — the first 16-audit run; remediated across
   waves A–G; closed findings reference their U-IDs in commit messages.
2. **Second-audit register** (`TENANT-1xx`, `RED-xxx`, `WIRE-xxx`, `SEC-xxx`,
   `SCALE-xxx`, `EBPF-xxx`, `AIRCA-xxx`, `ARCH-xxx`, `SUPPLY-xxx`, `TEST-xxx`,
   `CODE-xxx`, `OPS-xxx`, `DOCS-xxx`, `COMPLY-xxx`, `LICENSE-xxx`,
   `DATAROOM-xxx`) — the blind-instance audit @ `f856f25`, executed via the
   amended `REMEDIATION-SPRINT-PLAN.md`.

This map keeps the two from double-counting: where a second-audit finding
overlaps a U-item, progress is reported ONCE, under the second-audit ID, with
the U-ID as lineage. Verdicts come from `SPRINT-PLAN-TRIAGE.md`.

## Direct overlaps (same underlying item)

| 2nd-audit ID | U-register | Relationship |
|---|---|---|
| SCALE-002 | U-005 | Same item: L/XL run never executed. The U-005 harness (`make load-test TIER=…`) IS the Sprint 17 tool — only the run remains |
| OPS-007 | U-053 | Same item: RTO/RPO provisional pending a representative drill (CI drill exists) |
| DOCS-006 / SCALE-015 | U-051 | Same item: agent overhead measured fixture-side; live ring-buffer row pending (bench suite exists) |
| ARCH-003 | U-044 | Same surface: StreamConfig stub. U-044 decided keep+de-document (ADR); Sprint 13 hardens within that decision |
| OPS-009 | U-030 follow-up | Same item: backup CronJobs into the Helm chart (noted in `deploy/backup/README.md`) |
| TENANT-105 | U-025 | Same suite: the cross-tenant isolation tests (incl. ClickHouse). Second audit asks for ingest-path cases the U-025 suite predates |
| WIRE-003 | U-038 | U-038 wired revocation into the handshake; WIRE-003's residual is the registry/operator feed |
| RED-006 | U-090 | U-090 made the silo *router* fail closed (1×TTL stale cap); RED-006's residual is the emit/publish path |
| SCALE-003 | U-026 | U-026 shipped the per-tenant cardinality caps; residual is bounding/evicting the limiter state |
| SCALE-008 | U-040 | **Struck**: U-040 shipped retry+DLQ (`internal/pipeline/retry_dlq_test.go`); one device-plane verify-test remains in Sprint 14 |
| OPS-006 | U-046 | **Struck**: CH migrations run at boot via the U-046 versioned ledger; no init-container needed |
| ARCH-005 | U-047 | Decided by ADR (`docs/adr/volatile-stores.md`): rebuild-on-restart stands; only the silences/acks exception persists (Sprint 16) |
| DOCS-004 | U-019 | **Struck**: the honest-claims pass (B10) already states module-cert-not-product (README:140, CMVP #5247) |
| TEST-003 | U-049 | U-049 shipped the eval as non-blocking-initially BY DESIGN; Sprint 21 flips it to blocking now that the baseline exists (0.91/0.92) |
| COMPLY-003 | U-042 | U-042 documented pooled residency honestly as a label; COMPLY-003 adds enforcement for siloed/hybrid + explicit disclosure for pooled |

## Partial overlaps (U-item covers part; the 2nd-audit residual is real)

| 2nd-audit ID | U-register lineage | What the U-item covered / what remains |
|---|---|---|
| RED-001 / SEC-001 | U-001 | U-001 made the auth default fail-closed (`session`); the dev code path still ships in release binaries → Sprint 3 compiles it out |
| RED-004 | U-082 | **Struck**: the stampTenant edge case was *found by* the U-082 fuzz target and *fixed* in the same commit (93529b0) |
| SEC-003 | U-024 | U-024 added login throttle+lockout for tenant SSO; the provider/operator login path is outside it → Sprint 9 |
| AIRCA-002 | U-013 (C8) | C8 redaction masks IPs/hostnames/secrets; emails + free-text PII patterns remain → Sprint 20 extends it |
| EBPF-003 | U-014 + U-067 | U-014 digest-verifies BPF objects at load; U-067 (C6) cosign-signs release artifacts; load-time *signatures* on objects remain a decision → Sprint 19 |
| EBPF-006 | U-017 lineage | The observe-only gate existed; `bpf_probe_write_user` was missing from its denylist → **closed in Sprint 0** (this commit) |
| TENANT-102 | U-025 (D5) | D5 shipped CH row policies; the service-account exemption is the residual → Sprint 5 |
| SUPPLY-002/006 | U-059/U-060 | G13 pinned the Go toolchain/codegen/lint; the *Python* CI installs (ruff/black/pyyaml) were missed → Sprint 23 |
| DOCS-008 | U-006 + U-013 | Feed degradation + AI egress consent exist; the consolidated "what egresses when" page remains → Sprint 30 |

## Struck outright (no U-lineage needed — see SPRINT-PLAN-TRIAGE.md §2)

RED-009 (LogValue allowlist) · TEST-010 (`-race` in Makefile:116/121/136) ·
SUPPLY-004 (vimto pinned @v0.4.0) · SUPPLY-005 (toolchain cutoff artifact) ·
DOCS-003 / AIRCA-006 (analyzer correctly labeled + tested) · CODE-003
(no tracked binaries).

## Reporting rule

Sprint commits reference the **second-audit IDs**; where lineage exists, the
commit body may add "(lineage: U-xxx)". Progress dashboards count the
second-audit register only — the U-register is closed history.
