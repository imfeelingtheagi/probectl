# Remediation Sprint Plan — Triage & Amendments

**Reviews:** `REMEDIATION-SPRINT-PLAN.md` (blind-instance audit @ `f856f25`, 136 findings, 35 sprints)
**Against:** the repo at HEAD + `CLAUDE.md` (guardrails §7, conventions §6) + `probectl-PRD-v1.0.md` (delivered-state contract + ADRs)
**Method:** evidence-grade — a finding is only struck or re-scoped with a concrete citation (file:line, commit, CI job, or Make target). Anything ambiguous stays in the plan as a *verify-first* task.
**Decisions locked by the founder before this triage:**
1. **StreamConfig (ARCH-003/S13):** keep the U-044 ADR (wire-compatible stub) — harden it, don't remove it.
2. **NetworkPolicy (OPS-004/S26):** keep the U-086 documented-holes default — add a `values-strict.yaml` full-deny profile.
3. **Big builds:** do all three — S7 (CH driver), S11 (enrollment/SVID, **ADR-first**), S22 (OTLP traces+logs — this resolves PRD v1.0 §5.1-2 as "build").
4. **Triage bar:** evidence-grade strikes only.

---

## 1. Verdict summary

The blind audit is **structurally excellent and ~70% substantively right**. It found one genuinely new Critical the first (U-) register missed. But it ran **static-only against a repo it had no design context for**, so it also re-litigates settled ADRs, misses merged work, and contains a few tool artifacts.

| Verdict | Count | Meaning |
|---|---|---|
| **Confirmed** | ~78 | Real at HEAD; execute as planned (sometimes with sharpened scope) |
| **Partial / re-scope** | ~24 | The mechanism exists; only the named residual gap is real work |
| **Stale / wrong — struck** | 12 | Contradicted by cited code/commits/CI at HEAD |
| **Design-conflict — resolved** | 5 | Audit asks reverse deliberate decisions; resolved per the founder's calls above |
| **External / counsel / evidence-run** | ~17 | Not closable by code (SOC 2, pen test, license texts, L/XL iron, references) |

**Net effect on the plan:** 0 sprints dropped outright, **2 sprints effectively dissolve** (S30 loses 2 of 3 items; S31 shrinks ~4×), **14 sprints get scope edits**, the rest stand. Revised engineering effort: **~110 eng-days nominal**, which at the demonstrated agent-driven cadence is realistically **4–6 weeks of sessions**; the calendar critical path remains the external tracks (auditor, pen test, counsel, reference hardware), unchanged from PRD v1.0 §5.

**The single most important finding stands confirmed:** TENANT-101/WIRE-001 — `internal/flow/collector.go:207` stamps `TenantID` from agent *config*, not authenticated identity, on every bus-published plane. Sprint 4 is correctly the first real merge.

---

## 2. Struck findings (the evidence)

These come OUT of the plan. Each strike is citable to a future auditor.

| ID | Audit claim | Evidence it's wrong at HEAD |
|---|---|---|
| RED-004 | stampTenant edge case unfixed | Fixed in commit `93529b0`: `internal/otel/otlp/receiver.go` `stampTenant` overwrites the empty-valued attribute in place; `FuzzOTLPPayload` asserts via the pipeline's own `ResourceTenant` reader |
| RED-009 | config secrets potentially logged at startup | `internal/config/config.go:733` — `Config.LogValue()` is a slog allowlist (guardrail 6); startup logs `cfg` through it; DB URL passes through `redactURL` |
| TEST-010 | cover-gate doesn't use `-race` | `Makefile:116` (`go test -race` in `make test`), `:121` (isolation suite `-race`), `:136` (coverage profile `-race`) |
| SUPPLY-004 | vimto installed without pin/verify | `.github/workflows/ci.yml` kernel-matrix: `CGO_ENABLED=0 go install lmb.io/vimto@v0.4.0` — exact pin, sumdb-checksum-verified by `go install` |
| SUPPLY-005 | "forward-dated go 1.26" provenance unverifiable | Auditor knowledge-cutoff artifact — Go 1.26 is the current official toolchain (June 2026), checksummed via the Go toolchain sumdb. *Keep a 10-line `docs/build/toolchain.md` as a courtesy; the finding itself is invalid* |
| DOCS-003 | analyzer .py sources not locatable | `analyzer/probectl_analyzer/*.py` — 11 modules; pytest runs them in the `test-python` CI job at 94% coverage (85% floor) |
| DOCS-004 | FIPS headline misleads (module ≠ product) | `README.md:140` already states: CMVP cert **#5247** for the module, "probectl itself holds no product-level certificate" (the U-019/B10 honest-claims pass) |
| AIRCA-006 | data-room mislabels analyzer/ as the AI RCA layer | `analyzer/pyproject.toml` description = "probectl BGP analyzer"; `CLAUDE.md` §3/§5, `docs/architecture.md`, PRD all label it BGP; the AI layer is `internal/ai` |
| CODE-003 | committed binary blobs (36MB probectl-control, coverage.out) | `git ls-files` contains no `bin/`, `probectl-control`, or `*.out` entries — untracked local artifacts, not repo contents. *Keep only the `.gitignore` assertion + a repo-hygiene doc line* |
| OPS-006 | ClickHouse migrations lack a Helm init-container | Migrations run **at boot** via the U-046 versioned ledger: `internal/store/flowstore/clickhouse.go:79-108` (idempotent `MODIFY TTL`, ledger-recorded schema), `internal/store/chmigrate`. An init-container would duplicate the boot path |
| SCALE-008 | flow/device pipelines drop telemetry, no retry/DLQ | `internal/pipeline/retry_dlq_test.go` + consumer retry/DLQ shipped in D2 (U-040). *Residual: one verify-first test that the device plane's writes route through it* |
| SCALE-006 (flow half) | telemetry tables have no TTL | `flowstore/clickhouse.go:108` applies `ALTER TABLE … MODIFY TTL … DELETE` from per-tenant retention at boot. *The **path/traceroute** tables genuinely lack TTL — that half stands* |

Also corrected in place: **Sprint 6's "add `make test-isolation`"** — it already exists (`Makefile:121`, runs with `-race`; the `cross-tenant-isolation` CI job is required). The sprint's real content is the *new cases*, not the target.

---

## 3. Re-scoped findings (mechanism exists; only the residual is work)

| ID | What already exists (citation) | The real residual |
|---|---|---|
| TENANT-102 | CH row policies shipped in D5: `flowstore/clickhouse.go:229-230` — per-tenant policy + an exempt service user | The **service account** is policy-exempt by design (`USING 1 TO serviceUser`). Residual: scope service-account reads (per-request `SET` + policy, or split read/write users) — honest defense-in-depth increment |
| TENANT-104 | App role is `NOSUPERUSER NOBYPASSRLS` from `migrations/0007_app_role.sql:14` | Residual: the **startup self-assertion** (query `pg_roles`, FATAL if bypassrls/unforced RLS) — cheap, do it |
| TENANT-105 | Isolation suite exists, CI-required (`cross-tenant-isolation` job), `make test-isolation` with `-race` | Residual: add **ingest-path injection** cases (post-S4) + query-path cases (post-S5). **Keep the skip-harness/tag pattern** — "un-gate from plain `go test ./...`" would break DB-less dev boxes (founder decision) |
| TENANT-106 | Keyless-dev passthrough is documented design (`internal/tenantcrypto/tenantcrypto.go:5`) | Residual: **fail-fatal when encryption is configured but the key is unresolvable**; loud boot warning in keyless mode |
| WIRE-003 | Revocation IS wired into the handshake (G4/U-038): `agenttransport/server.go:43` `ServerMTLSConfigRevocable`; refusal tests exist | Residual: nothing **feeds** the list — add the registry/operator path (`probectl-control revoke-agent` → `RevocationList().Replace`) |
| RED-006 | The silo **router** fails closed with a 1×TTL stale cap (U-090, `ee/silo/router.go`) | Residual: the **emit/publish** path (= TENANT-107): flow/device/eBPF emitters have no `BusNamespace` handling — siloed tenants' records ride shared topics. Real; fix in S4/S6 |
| RED-007 | stdio MCP documents caller-side token auth: `internal/ai/mcp/stdio.go:13-16` ("the binary authenticates the token before calling this") | Residual: **verify + test** the binary entry actually enforces it; document the local-trust model |
| SCALE-003 | Per-tenant cardinality caps exist (D3/U-026): `internal/pipeline/cardinality.go`, rejection at `consumer.go:207` | Residual: **bounded/evicting** limiter state + (optional) cross-replica sharing. Don't rebuild what exists |
| SCALE-005 | The pipeline cardinality cap applies to consumed series (`consumer.go:207`) | Residual: *verify* device-plane series flow through it + ship **non-unlimited default rate limits** (SCALE-004 confirmed: `fairness.go` rate<=0 = unlimited) |
| SCALE-009 | Path inserts are already per-path batched (`JSONEachRow` over hop/link slices, `pathstore/clickhouse.go:158,173`) | Residual: cross-path batching window — worthwhile, smaller than written |
| EBPF-003 | Release artifacts are cosign-signed (C6/U-067); BPF objects digest-verified at load (C9/U-014) | Residual: load-time **signature** (not just embedded digest) on BPF objects if you want operator-supplied objects; otherwise document the digest model as the boundary |
| EBPF-004 | Fixture-vs-live is a **build-tag** fact (`make ebpf-agent` builds `-tags ebpf`, `Makefile:100`) | Residual: assert the **shipped agent image** is the live build + document fixture as dev/test mode |
| AIRCA-001 | MCP is tenant-first + RBAC-scoped (S25); HTTP transport TLS+token | Residual: route MCP responses under the **per-tenant egress-consent flag** (the client *is* an external AI) — config + check, not a rebuild |
| AIRCA-002 | Redaction already masks IPs (default), hostnames (policy), secrets/bearer **always** (`internal/ai/redact.go:12-30`, C8) | Residual: add **email/free-text-PII patterns** + custom-pattern config |
| COMPLY-005 | SCIM provisioning IS audited: `internal/control/scim.go:20-21` `auditSCIM`; break-glass has its own stream (U-012) | Residual: a **coverage test** asserting both event families land in the chains |
| CODE-005 | Metering exists: `internal/usage` (S-T3 seam) + `ee/billing` export | Residual: `internal/billing` is a 6-line dead scaffold — **delete it or repoint** the usage seam; not a 6-day build |
| TEST-004 | Real integration suites exist (store integration suite, `test/e2e`, compose-backed CI jobs) | Residual: `test/integration/` itself is a 16-line smoke — either build it out as the cross-plane home or document the layout and point the gate at the real suites |
| ARCH-005 | Rebuild-on-restart is a **decided ADR** (`docs/adr/volatile-stores.md`, U-047) with cold-start tests | Residual: persist **alert silences/acks** — the ADR's own documented exception. Everything else in S16.4 is the ADR re-litigated; strike it |

---

## 4. Design-conflict resolutions (founder-decided)

| Conflict | Audit ask | Resolution |
|---|---|---|
| **S13 / ARCH-003** StreamConfig | Remove the RPC (buf-breaking) | **Keep U-044 ADR.** Amend S13 to: (a) server returns an explicit `codes.Unimplemented` deny instead of holding the stream open, (b) delete the agent-side client stub usage, (c) test asserting no config-apply path exists in the agent, (d) ADR addendum. ~1d, zero wire break |
| **S26 / OPS-004** NetworkPolicy | Default-deny everything | **Keep U-086 documented-holes default.** Add `values-strict.yaml` (default-deny ingress+egress, named ingress-controller selector required) + reference it from `docs/hardening.md`. Mark OPS-004 as resolved-by-profile |
| **S6 / TENANT-105** isolation tests | Un-gate from plain `go test` | **Keep the harness/tag pattern** (tests need live PG/CH). The CI job is already required; expand cases instead |
| **S16.4 / ARCH-005** persistence | Persist all in-memory stores | **Keep U-047 ADR** (rebuild-on-restart). Only silences/acks persist (the ADR's exception) |
| **S21 / TEST-003** rca-eval | Make blocking | **Accept now** — U-049 said "non-blocking initially"; the baseline exists (0.91/0.92). Flip with floor `accuracy ≥ 0.85 ∧ precision ≥ 0.85` |

---

## 5. The amended sprint list (apply these edits to REMEDIATION-SPRINT-PLAN.md)

Verdict key: **KEEP** (as written) · **MOD** (edit scope) · **LITE** (shrinks materially) · external unchanged.

| # | Sprint | Verdict | Edits to make |
|---|---|---|---|
| 0 | **NEW — Triage record** | ADD | Commit this document; add `docs/diligence/known-risks.md` seeded from §2–§4 (pre-pays S29 task 1); reconcile finding IDs against the closed U-register in a two-column map so two trackers never drift |
| 1 | CI gate + hygiene | MOD | Drop CODE-003 BFG/history work (struck — keep only `.gitignore` assertions). Change "CI bot commits coverage badge to main" → **artifact + PR-comment only** (a bot pushing to a protected main contradicts the same sprint's ruleset). Keep: rulesets-as-code, drift check, CODEOWNERS, gitleaks, CODE-006 key removal |
| 2 | Coverage + signing + errors | MOD | Keep `internal/control` floor + CODE-002 (confirmed: `_ = json.Unmarshal` on `store/abac.go:27,30`, `audit/worm.go:250`, `store/users.go:32`, `store/changes.go:29` — ABAC silent-empty is the one that matters: fail closed). Commit **signing** stays ruleset-side; DCO arrives in S33 — note the overlap so it's configured once |
| 3 | Dev-auth kill `[CRIT]` | KEEP | Fix the broken verify: `strings`-grepping a *source comment* proves nothing (comments don't compile). Replace with: release build + `go tool nm` symbol absence + a boot test that `PROBECTL_AUTH_MODE=dev` fatals in a release binary |
| 4 | Tenant binding on ingest `[CRIT]` | KEEP | **First real merge.** Drop RED-004 from Closes (struck). Confirmed core: `flow/collector.go:207` + device/eBPF config-stamping. Add TENANT-107 emitter-namespace routing here (it's the same code path) or in S6 — once, not twice |
| 5 | DB-level backstop | MOD | TENANT-104 → startup `rolbypassrls`/FORCE-RLS assertion only (role posture already in `0007_app_role.sql`). TENANT-102 → service-account scoping increment (policies exist, `flowstore:229`). TENANT-106 → fail-closed-when-configured + keyless boot warning |
| 6 | Cross-tenant suite | MOD | Keep harness/tags (founder decision). Add: ingest-injection cases (payload tenant ≠ identity tenant across all four ingest surfaces), query-path cases, silo-namespace emitter tests. Strike "add make test-isolation" (exists, `Makefile:121`) |
| 7 | CH driver + params | KEEP | Confirmed (`chStr` manual escaping, raw HTTP). Carry the existing breaker (U-078) through the driver swap; keep `BreakerStats`. Add the `no-stringbuilt-sql` lint |
| 8 | At-rest encryption default | MOD | Coordinate with S5's TENANT-106 (same seam — do the fail-closed logic once). Keep: encryption-on shipped default, operator-duty doc + preflight |
| 9 | Auth/session hardening | MOD | Drop RED-009 (struck — `LogValue` allowlist). Keep: SEC-003 (confirmed — `ee/provider/handler.go` does its own operator authn outside `authLimiter`; reuse `internal/auth.Limiter`), SEC-004 (confirmed — nonce sent at `control/auth.go:282`, never stored/compared on callback), SEC-009 (confirmed — `auth.NewManager(..., cfg.TLSEnabled())` ties Secure to own listener; add edge-TLS config), RED-007 → verify+test stdio token path |
| 10 | SSRF + info-leak + creds | KEEP | All confirmed: `ssrf.go:67` blocks only the exact unspecified addr (add 0.0.0.0/8); `probectl.yml:26,79` default password `probectl` in the non-dev recipe; `PGPASSWORD: probectl` hardcoded in `compose-backup.yml:10`; `/version` + `/openapi.json` unauthenticated (`server.go:307-308`); CI `sslmode=disable` |
| 11 | Enrollment + SVID `[ADR-first]` | KEEP | Founder decision: **design ADR + threat-model delta reviewed before code** (new trust root; CLAUDE.md §9). Then implement as written. The plan's join-token design is sound |
| 12 | Revocation + TLS + replay | MOD | WIRE-003 → registry feed + `revoke-agent` CLI only (handshake already wired, G4). Keep: TLS-mandatory control API (confirmed plaintext fallback `server.go:415`), unified `tls.Config` for OTLP/MCP, TLS 1.3 floor (Go-only fleet — safe), replay window, ca_file allowlist |
| 13 | StreamConfig | REWRITE | Per §4: explicit deny + remove client stub + no-config-path test + ADR addendum. No proto removal, no buf exception |
| 14 | Ingest throughput | MOD | Drop SCALE-008 from Closes (struck; add one verify-test for device-plane DLQ routing). Keep: per-topic worker pool (confirmed: concurrency is per-*topic*, serial within), decode-once fan-out (verify multiplicity first), batched path inserts (note per-path batching exists) + heartbeat batching |
| 15 | Cardinality + rate limits | MOD | SCALE-003 → bound/evict the existing limiter (don't rebuild); SCALE-004 confirmed (ship non-unlimited defaults, unlimited = explicit opt-in); SCALE-005 verify-first; SCALE-007 sub-partitioning keep; SCALE-011 async-enrichment keep after verifying it's actually sync |
| 16 | Storage durability | MOD | SCALE-006 → **pathstore TTL only** (flow TTL exists, `flowstore:108`); SCALE-010 keep (confirmed: `probectl.otlp.metrics` has zero consumers — wire one or delete the topic; decide against S22's consumer); SCALE-014 keep (linear scans in `tsdb/memory.go`); ARCH-005 → **silences/acks persistence only** per U-047 ADR |
| 17 | Scale validation `[B5]` | MOD | **The harness already exists** (E1/U-005: `make load-test TIER=L\|XL`, S-tier smoke in CI; scale-gate floor in CI from D12). Re-scope to: add the 3 missing benchmarks (consumer/TSDB/pathstore), **run L/XL on reference hardware** `[needs infra]`, fill `docs/scale-gate.md` + `docs/agent-overhead.md` live row (U-051), flip PROVISIONAL labels. This is PRD v1.0 §5.2, not a build |
| 18 | sslsniff scoping | KEEP | Confirmed (no pid/cgroup filter in `sslsniff.bpf.c`; consent gate is binary, not per-process). The most important eBPF item. Kernel-side or earliest-boundary redaction as written |
| 19 | Agent integrity + caps | MOD | EBPF-006 is a **one-line confirmed gap**: `bpf_probe_write_user` missing from the forbidden list in `observeonly_test.go:31` — add it (+ `bpf_probe_write_user` re-audit) immediately, don't wait for the sprint. EBPF-005 confirmed (probe checks `CapBPF` only, `capability_linux.go:33` — add PERFMON). EBPF-003 → per §3. EBPF-004 → shipped-image assertion. EBPF-007 verify-first. EBPF-008 arm64 matrix keep |
| 20 | AI egress unification | MOD | Keep: single egress gate (AIRCA-005 confirmed — `internal/ai/author` has no consent/egress references), MCP read audit (AIRCA-003 confirmed — no audit calls in `internal/ai/mcp`), MCP-under-consent (AIRCA-001 per §3), RED-005 root-cause citation check. AIRCA-002 → **add email/PII patterns** to the existing redactor (don't rewrite it — cite `redact.go:12-30`) |
| 21 | RCA resilience + eval gate | LITE | Drop AIRCA-006 + DOCS-003 (struck). Keep: breaker+cache on the remote-model path (reuse `internal/breaker`), TEST-003 flip rca-eval to blocking with the 0.85/0.85 floor |
| 22 | OTLP traces + logs | KEEP | Founder decision (resolves PRD §5.1-2 = build). Fold in DOCS-005 (claim wording) + ARCH-006 (collector example config). Coordinate the consumer with S16/SCALE-010 |
| 23 | Supply pins | LITE | Drop SUPPLY-004 + SUPPLY-005 as findings (struck; keep the 10-line toolchain provenance doc). Keep: SUPPLY-001 (confirmed `:latest` default, `probectl.yml:54`), SUPPLY-002/-006 (confirmed unpinned `pip install ruff black` ci.yml:51, `pyyaml` :223), SUPPLY-003 (uv lock for analyzer) |
| 24 | Test depth | MOD | Drop TEST-010 (struck — `-race` everywhere in Makefile). TEST-004 per §3. Keep: floors for `tenantlife`/`ee/silo`/`pathstore`/`tsdb` (TEST-005/009), PR-profile e2e (TEST-006), `t.Parallel()` for the slowest packages (TEST-007 — confirmed zero usages) |
| 25 | Green-build capstone | LITE | Most steps are already standing CI gates. Real content: `make verify-all` umbrella target + making govulncheck/trivy explicitly blocking + archived receipts. ~1d, keep as the phase-closer |
| 26 | k8s/Helm ops | MOD | Drop OPS-006 (struck — boot migrations). OPS-004 → `values-strict.yaml` per §4. Keep: agent DaemonSet probes (confirmed absent), `/metrics` + ServiceMonitor (confirmed absent), backup CronJobs into chart (already on the roadmap — U-030 follow-up) |
| 27 | Backup crypto + DR + IdP | KEEP | All confirmed real: backups unencrypted (U-030 scripts write plain dumps), erasure-not-covering-backups is a documented gap (COMPLY-002), DR rep-drill = the U-053 evidence run `[needs infra]`, air-gap IdP path (Dex example + values) |
| 28 | Residency + audit coverage | MOD | COMPLY-003 keep — but scope per design: **siloed/hybrid tenants get enforcement** (region-pinned data planes exist); **pooled tenants get the documented inherited-region disclosure** (U-042 honesty stands — don't pretend pooled rows have per-row residency). COMPLY-005 → coverage test only (`auditSCIM` exists, `scim.go:20`) |
| 29 | Data room + external kickoff | KEEP | Seed `known-risks.md` from THIS triage (§2–§4 are the residual-risk register). ADR backfill: enrollment (S11), CH driver (S7), scale architecture (S14-16), sslsniff scoping (S18), tenant-binding (S4) |
| 30 | Docs truthfulness | LITE | Drop DOCS-004 (struck — README:140 already honest). DOCS-005 folds into S22. Remaining: DOCS-008 sovereignty-caveat sweep (one page in `docs/` listing exactly which optional features egress: threat-intel/outage feeds, remote AI, and now MCP) |
| 31 | Billing MVP | LITE | Re-scope per CODE-005 evidence: metering exists (`internal/usage` + `ee/billing`); the work is deleting/repointing the dead `internal/billing` scaffold + an invoice-export round-trip test. ~1–2d, not 6 |
| 32–35 | Phase L (license/IP/TM) | KEEP | Unchanged — counsel-gated, last per the standing instruction. Matches PRD v1.0 §5.3. Start the counsel *decisions* now (parallel track), commit at the end |

**Revised order stays the plan's order** with one swap: do the **EBPF-006 one-liner** (denylist add) immediately as part of Sprint 0 — it's a guardrail-§7.8 hardening with a 5-minute diff; no reason to let it wait for S19.

---

## 6. Cross-cutting corrections to the plan document itself

1. **Add the missing register reconciliation.** The plan never references the closed U-001…U-094 register. Add a section mapping overlapping IDs (e.g. SCALE-002≈U-005, OPS-007≈U-053, DOCS-006≈U-051, ARCH-003≈U-044, OPS-009≈U-030-followup, TENANT-105≈U-025) so progress reports don't double-count.
2. **Fix Sprint 3's verify command.** `strings /tmp/pc | grep "trusted-header dev principal"` greps for a *source comment* — comments never reach the binary, so it passes vacuously today. Use symbol analysis + a behavioral boot test.
3. **Fix Sprint 1's self-contradiction.** A CI bot committing `docs/quality/coverage.md` to `main` violates the branch ruleset the same sprint installs. Receipts should be workflow artifacts + a PR-comment bot.
4. **Re-state the score trajectory honestly.** With 12 strikes and ~24 re-scopes, "Today: 58.9 uncapped" is understated — several domain scores (TEST, SUPPLY, AIRCA, OPS) were scored against findings that don't exist at HEAD. Expect the re-baselined uncapped score to start in the mid-60s. The cap mechanics (49 until Phase L + SOC 2) are correct and unchanged.
5. **Mark the three evidence-run sprints clearly.** S17 (L/XL), S27 (DR drill), and the agent-overhead row aren't code sprints — they need reference hardware and map 1:1 to PRD v1.0 §5.2. Don't let them block code phases; schedule the iron in parallel.
6. **Sprint sizes assume human pace.** Keep the day-sizes as planning units, but the realistic wall-clock at the demonstrated cadence is 4–6 weeks for Phases 0–12, with counsel/auditor/iron as the long poles.

---

## 7. Recommended execution sequence (post-amendment)

1. **Sprint 0** (this triage committed + known-risks register + EBPF-006 one-liner + ID reconciliation map)
2. **S1 → S2 → S3 → S4** (foundation, then both non-licensing Criticals; S4 is the merge that matters)
3. **S5 → S6** (DB backstop + the expanded isolation suite proving S4)
4. **S7–S10** (security hardening; S7 is the big refactor — protect it with the injection tests first)
5. **S11-ADR → S11 → S12 → S13** (trust root; ADR reviewed before code)
6. **S14 → S15 → S16** (scale rebuild) → **S17** `[iron]` in parallel with →
7. **S18 → S19** (eBPF) → **S20 → S21** (AI) → **S22** (OTLP signals) → **S23 → S24 → S25** (supply/test/capstone)
8. **S26 → S27 → S28 → S29 → S30 → S31-lite** (ops/compliance/docs)
9. **Phase L** (S32–S35) when counsel returns texts — kick off the counsel decisions, SOC 2 engagement, and pen test **now** (the plan's Parallel Tracks table is right).

---

*Prepared against HEAD with the CLAUDE.md guardrails and PRD v1.0 as the design contract. Every strike in §2 carries a citation an external reviewer can check; everything ambiguous stayed in the plan as verify-first.*
