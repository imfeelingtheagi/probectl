# probectl — Remediation Sprint Plan

**Source of truth:** `outputs/00-INVESTMENT-COMMITTEE-MEMO.html` + `outputs/INDEX.md` (16-audit diligence run @ commit `f856f25`).
**Goal:** close **every actionable finding** (post-triage: **~124 of 136** — 12 struck with evidence, see `SPRINT-PLAN-TRIAGE.md` §2) and drive technical readiness from **49 → ~92–95**. Licensing epics are intentionally **last** (Phase L).
**Format:** one drop-in **Claude Code** prompt per sprint (the fenced `text` block). Paste a sprint's block into Claude Code pointed at the `probectl` repo, let it implement, then run the **Verify** commands and merge once the **CI gate** is green.

> **AMENDED (post-triage, 2026-06-07).** This plan was produced by a blind audit instance with no access to `CLAUDE.md` or the PRD. It has been triaged against the repo at HEAD and the design contract — full verdicts + evidence in **`SPRINT-PLAN-TRIAGE.md`**. Net: **12 findings struck with citations** (marked ~~STRUCK~~ below), **~24 re-scoped** to their real residual, **5 design conflicts resolved by founder decision** (U-044 StreamConfig ADR stands — harden, don't remove; U-086 NetworkPolicy documented-holes default stands — add a strict profile; isolation-test harness pattern stands; U-047 rebuild-on-restart ADR stands — persist silences only; rca-eval flips to blocking now that the baseline exists). All three big builds proceed: S7 (CH driver), S11 (enrollment — **design ADR + threat-model delta reviewed before code**), S22 (OTLP traces+logs — resolves PRD v1.0 §5.1-2 as "build"). A new **Sprint 0** records the strikes and lands the one-line guardrail fix immediately.

---

## How to use this doc
- Work **top to bottom** — sprints are ordered for maximum readiness gain *and* dependency-correctness. Each sprint lists `Depends on:` so you can parallelize safely if you have more than one engineer.
- Each sprint closes a small, adjacent cluster of finding IDs (you can trace any ID back to its audit report via `INDEX.md`).
- Every sprint ends with **runnable Verify commands** and a **CI gate** so the fix can't regress. Verify steps assume a standard dev box (Go 1.26 toolchain, `clang`/LLVM for eBPF, Docker + compose, and ClickHouse/Kafka/Postgres for integration) — where a step needs a kernel or live datastore it is marked `[needs infra]`.
- **Definition of done for a sprint:** all its finding IDs closed, Verify green, CI gate added/passing, and a one-line entry added to `CHANGELOG.md` referencing the IDs.

## Strategy & ordering rationale
1. **Phase 0 (Foundation)** makes CI a *real* merge gate and cleans the repo first, so every later fix is enforced rather than aspirational.
2. **Phase 1** clears the two **non-licensing Criticals** (dev-auth bypass, cross-tenant ingest) — the holdback blockers that also unblock a clean red-team retest.
3. **Phases 2–9** harden the highest-weight value/risk drivers in dependency order: ClickHouse/SEC → agent identity/transport (WIRE) → **scale/reliability (the single biggest absolute lift, 10→~90)** → eBPF agent → AI/RCA → architecture → supply chain → tests.
4. **Phases 10–12** finish ops, compliance (non-licensing), and docs truthfulness.
5. **Phase L (Licensing) runs LAST** per your instruction. It clears the final legal/IP Critical and **lifts the score cap**.

### The score-cap mechanic (read this)
Any **single unresolved Critical caps the reconciled headline score at ≤ 49**. There are 4 distinct Criticals:

| Critical | Source IDs | Cleared by | Lifts cap when |
|---|---|---|---|
| Cross-tenant ingest | TENANT-101, WIRE-001 | **Sprint 4** | merged |
| Dev-mode auth bypass | RED-001, SEC-001 | **Sprint 3** | merged |
| Legal/IP not transferable | LICENSE-001/-002, DATAROOM-001/-002 | **Sprints 32–33** (Phase L, last) | merged |
| No compliance attestation | COMPLY-001 | **Sprint 29** (evidence) + external SOC 2 audit | attestation received |

**Consequence:** component/domain scores climb steadily through Phases 1–12, but the **headline reconciled score stays ≤ 49 until Phase L is merged** (and fully lifts only once the SOC 2 attestation lands — an external, ~6–12-month track you should kick off early, see Parallel Tracks). This is *why* licensing is safe to do last: it's the final cap-release, not a prerequisite for the engineering work.

### Projected readiness trajectory
| Milestone | Non-licensing weighted (uncapped) | Headline (capped) |
|---|---|---|
| Today | 58.9 | **49** |
| After Phase 1 (S6) | ~63 | 49 (2 Criticals remain) |
| After Phase 4 / scale (S17) | ~82 | 49 |
| After Phase 9 (S25) | ~92 | 49 |
| After Phase 12 (S30) | ~94 | 49 (legal + compliance Criticals remain) |
| **After Phase L (S35) + SOC 2 attestation** | ~94 | **~92–95 (cap lifted)** |

> **Re-baseline note (post-triage):** several domain scores (TEST, SUPPLY, AIRCA, OPS) were scored against findings that do not exist at HEAD (e.g. `-race` already runs in three Make targets; vimto is pinned; FIPS wording is already honest). Expect the re-baselined uncapped score to start in the **mid-60s**, not 58.9. The cap mechanics are unchanged.

> Realistic ceiling note: COMPLY/SOC 2 (`COMPLY-001`) cannot be fully *closed* by code — its Type II attestation is an external auditor deliverable. The ~92–95 headline assumes the attestation is received; until then the compliance Critical holds the cap even after all code is merged.

---

## Sprint overview (recommended order)

| # | Sprint | Phase | Closes (IDs) | Size | Lever |
|---|---|---|---|---|---|
| 0 | **Triage record + immediate guardrail fix** | 0 Foundation | records the 12 strikes; EBPF-006 (one-line denylist add); known-risks register; U-register ID map | ~0.5d | honest tracking |
| 1 | CI becomes a real gate + repo hygiene | 0 Foundation | TEST-002, SUPPLY-007, CODE-007, CODE-003 (gitignore-assert only), CODE-006, TEST-008, DATAROOM-005, DATAROOM-011 | ~2d | enables all |
| 2 | Control-plane coverage gate + commit signing + safe error handling | 0 Foundation | TEST-001, CODE-001, CODE-004, CODE-002 | ~3d | TEST/CODE |
| 3 | **Kill dev-mode auth bypass** [CRIT] | 1 Criticals | RED-001, SEC-001 | ~2d | SEC/cap |
| 4 | **Server-side tenant binding on ingest** [CRIT] | 1 Criticals | TENANT-101, WIRE-001, RED-006, TENANT-107 (~~RED-004 struck — fixed at 93529b0~~) | ~5d | TENANT/cap |
| 5 | DB-level tenant isolation backstop | 1 Criticals | TENANT-102, TENANT-104, TENANT-106 | ~3d | TENANT |
| 6 | Cross-tenant test suite (expanded; harness pattern kept) | 1 Criticals | TENANT-105 | ~2d | TENANT |
| 7 | ClickHouse driver + parameterized queries | 2 SEC/Arch | ARCH-002, SEC-005, TENANT-108 | ~4d | ARCH/SEC |
| 8 | At-rest telemetry encryption by default | 2 SEC | SEC-002, COMPLY-004 | ~3d | SEC |
| 9 | Auth & session hardening | 2 SEC | SEC-003, SEC-004, SEC-009, RED-007 (verify) (~~RED-009 struck — LogValue allowlist~~) | ~2d | SEC |
| 10 | SSRF, info-leak & shipped-credential cleanup | 2 SEC | SEC-006, SEC-007, SEC-008, OPS-003, OPS-010 | ~2d | SEC |
| 11 | Agent enrollment & SVID issuance | 3 WIRE | WIRE-002, RED-002, TENANT-103, ARCH-004 | ~6d | WIRE |
| 12 | Cert revocation + transport TLS + replay protection | 3 WIRE | WIRE-003, WIRE-004, WIRE-005, WIRE-006, WIRE-007, RED-008 | ~5d | WIRE |
| 13 | Harden StreamConfig stub (explicit deny; U-044 ADR stands) | 3 WIRE | ARCH-003 | ~1d | ARCH/SEC |
| 14 | Ingest throughput & backpressure | 4 SCALE | SCALE-001, SCALE-009, SCALE-012, SCALE-013 (~~SCALE-008 struck — DLQ shipped (D2); one verify-test stays~~) | ~5d | SCALE |
| 15 | Cardinality control & rate limiting | 4 SCALE | SCALE-003, SCALE-004, SCALE-005, SCALE-007, SCALE-011 | ~5d | SCALE |
| 16 | Storage durability: path TTL, OTLP consumer, TSDB index, silences | 4 SCALE | SCALE-006 (path tables only), SCALE-010, SCALE-014, ARCH-005 (silences/acks only — U-047 ADR stands) | ~4d | SCALE/ARCH |
| 17 | Scale validation RUN (harness exists — E1/U-005) + missing benchmarks [BLOCKER B5] | 4 SCALE | SCALE-002, SCALE-015, DOCS-001, DOCS-006 | ~3d + iron | SCALE |
| 18 | eBPF TLS-capture scoping + kernel-side redaction | 5 EBPF | EBPF-001, EBPF-002, RED-003 | ~5d | EBPF |
| 19 | Agent integrity (signature) + capability/denylist/limits + CI matrix | 5 EBPF | EBPF-003, EBPF-004, EBPF-005, EBPF-006, EBPF-007, EBPF-008 | ~5d | EBPF |
| 20 | MCP/AI egress controls, redaction & audit | 6 AIRCA | AIRCA-001, AIRCA-002, AIRCA-003, AIRCA-005, RED-005 | ~4d | AIRCA |
| 21 | RCA resilience + blocking eval gate | 6 AIRCA | AIRCA-004, TEST-003 (~~AIRCA-006, DOCS-003 struck — analyzer correctly labeled + locatable~~) | ~2d | AIRCA |
| 22 | OTLP traces+logs + OTel Collector path | 7 ARCH | ARCH-001, ARCH-006, DOCS-005 (claim wording folds in here) | ~5d | ARCH |
| 23 | Pin remaining inputs + analyzer lockfile | 8 SUPPLY | SUPPLY-001, SUPPLY-002, SUPPLY-003, SUPPLY-006 (~~SUPPLY-004 struck — vimto@v0.4.0 pinned; SUPPLY-005 struck — cutoff artifact (keep 10-line toolchain doc)~~) | ~2d | SUPPLY |
| 24 | Test depth: integration home, coverage floors, e2e on PRs, parallel | 9 TEST | TEST-004, TEST-005, TEST-006, TEST-007, TEST-009 (~~TEST-010 struck — -race in Makefile:116/121/136~~) | ~4d | TEST |
| 25 | Green-build capstone: build + `-race` + govulncheck + eBPF load [static-only closure] | 9 TEST | (verifies S14–S19) | ~3d | methodology |
| 26 | k8s/Helm hardening: probes, strict NetPolicy profile, /metrics, backup jobs | 10 OPS | OPS-001, OPS-004 (strict profile — U-086 default stands), OPS-005, OPS-009 (~~OPS-006 struck — migrations run at boot, U-046~~) | ~3d | OPS |
| 27 | Backup encryption + DR drill + air-gap IdP | 10 OPS | OPS-002, OPS-007, OPS-008, COMPLY-002, DATAROOM-008, DOCS-007 | ~5d | OPS |
| 28 | Data residency enforcement + audit-trail coverage + PII-min doc | 11 COMPLY | COMPLY-003, COMPLY-005, COMPLY-006, DATAROOM-009 | ~4d | COMPLY |
| 29 | Data-room package + external-engagement kickoff | 11 COMPLY | DATAROOM-004, DATAROOM-006, DATAROOM-007, DATAROOM-010, DATAROOM-012, DATAROOM-013, COMPLY-001 | ~4d | DATAROOM |
| 30 | Sovereignty-caveats doc sweep | 12 DOCS | DOCS-008 (~~DOCS-004 struck — README:140 already honest; DOCS-005 → S22~~) | ~1d | DOCS |
| 31 | (Optional) Billing scaffold cleanup + export round-trip | 12 CODE | CODE-005 (re-scoped — metering exists: internal/usage + ee/billing) | ~2d | CODE |
| 32 | **Commit OSS + EE license texts + SPDX headers** [CRIT] | L Licensing | LICENSE-001, LICENSE-002, LICENSE-003, DATAROOM-001, DOCS-002 | ~3d | LICENSE/cap |
| 33 | **IP chain of title: CLA/DCO + assignments** [CRIT] | L Licensing | LICENSE-004, DATAROOM-002 | ~2d | LICENSE/cap |
| 34 | Third-party license inventory + NOTICE + eBPF GPL boundary | L Licensing | DATAROOM-003, LICENSE-005, LICENSE-006 | ~3d | LICENSE |
| 35 | Trademark strategy & registration | L Licensing | LICENSE-007 | ~2d (legal) | LICENSE |

_Rough roll-up (post-triage): **~110 eng-days nominal** — at the demonstrated agent-driven cadence, realistically **4–6 weeks of sessions**. The calendar critical path is external: counsel (Phase L), the SOC 2 auditor, the pen test, and reference hardware for S17/S27 (excluded from the day count)._

---

# Phase 0 — Foundation & guardrails

> Do these first. They make CI a real merge gate and clean the repo so every later fix is enforced and auditable.

## Sprint 0 — Triage record + immediate guardrail fix  `[NEW — post-triage]`
**Closes:** records the 12 strikes; lands EBPF-006 immediately · **Depends on:** none · **Size:** ~0.5d

```text
You are working in the probectl Go monorepo. Goal: record the post-triage state so every later sprint tracks honestly, and land the one guardrail fix too cheap to wait for its sprint.

Tasks:
1. EBPF-006 NOW (one line + re-audit): add "bpf_probe_write_user" to the forbidden-helper list in internal/ebpf/observeonly_test.go (line ~31), and re-audit the list against the current bpf-helpers(7) set for any other state-mutating helpers (bpf_probe_write_user is the memory-corrupting one). This is a CLAUDE.md §7.8 observe-only guardrail hardening.
2. Known-risks register: create docs/diligence/known-risks.md seeded from SPRINT-PLAN-TRIAGE.md §2-§4 (residual risks, owners, status). This pre-pays Sprint 29 task 1.
3. Register reconciliation map: create docs/diligence/finding-id-map.md mapping this plan's IDs to the closed U-register where they overlap (SCALE-002≈U-005, OPS-007≈U-053, DOCS-006≈U-051, ARCH-003≈U-044, OPS-009≈U-030-followup, TENANT-105≈U-025, WIRE-003≈U-038, RED-006≈U-090, SCALE-003≈U-026, SCALE-008≈U-040, OPS-006≈U-046, ARCH-005≈U-047, DOCS-004≈U-019, TEST-003≈U-049, SCALE-015≈U-051, COMPLY-003≈U-042) so the two registers never double-count.
4. Commit SPRINT-PLAN-TRIAGE.md + this amended plan alongside the audit outputs.

Acceptance:
- observe-only test forbids bpf_probe_write_user and passes; known-risks + ID map committed.

Verify:
- grep -n "bpf_probe_write_user" internal/ebpf/observeonly_test.go ; expect present in forbidden list
- go test ./internal/ebpf/ -run ObserveOnly -v
- ls docs/diligence/known-risks.md docs/diligence/finding-id-map.md

CI gate: none new (the observe-only test is already required).
```

## Sprint 1 — CI becomes a real gate + repo hygiene
**Closes:** TEST-002 (H), SUPPLY-007 (L), CODE-007 (L), CODE-003 (M), CODE-006 (L), TEST-008 (L), DATAROOM-005 (H), DATAROOM-011 (L) · **Depends on:** none · **Size:** ~3d

```text
You are working in the probectl Go monorepo. Goal: make CI an enforced merge gate, prove branch protection from inside the repo, and remove repo-hygiene risks.

Tasks:
1. Branch protection as code (TEST-002, SUPPLY-007, CODE-007): add a committed, version-controlled policy that is the source of truth for required checks. Add a GitHub ruleset file under .github/ (e.g. .github/rulesets/main.json) describing required status checks (build, test, lint, govulncheck, trivy, license-gate), required PR review, and dismiss-stale-approvals. Add a CODEOWNERS file mapping each top-level dir (internal/, cmd/, ee/, deploy/, migrations/) to owners. Add a CI job `verify-branch-protection` that calls the GitHub API and fails if live settings drift from the committed ruleset.
2. Remove the committed test private key (CODE-006): delete internal/auth/testdata/oidc_test_key.pem and replace it with a key generated at test-setup time (t.TempDir + crypto/rsa GenerateKey in a TestMain or helper). Add a gitleaks/trufflehog CI step that fails on committed private keys.
3. CODE-003 (TRIAGED — gitignore-assert only): `git ls-files` contains NO tracked binaries or coverage files at HEAD (verified in triage) — the 36MB binary and coverage.out are untracked local artifacts. Just assert .gitignore covers `probectl-control`, `*.out`, `bin/`, and built binaries, and add a one-line note to docs/dev/repo-hygiene.md. Do NOT do BFG/history surgery.
4. Coverage receipt (TEST-008, DATAROOM-005) — AMENDED: have CI write coverage.out + coverage-summary.txt and upload them as retained workflow artifacts, plus a PR-comment summary. Do NOT have a bot push to main — that would violate the branch ruleset installed by task 1 of this very sprint. docs/quality/coverage.md documents where the artifacts live.
5. Publish CI/scan artifacts for the data room (DATAROOM-011): ensure govulncheck, trivy, and test logs are uploaded as retained workflow artifacts and linked from docs/quality/.

Acceptance:
- A PR that fails build/test/lint/govulncheck/trivy cannot be merged (verify-branch-protection passes and required checks are enforced).
- No private keys or large binaries are tracked; secret-scan CI step is green.
- docs/quality/coverage.md documents the artifact location; receipts retained per-run.

Verify:
- grep -rn "BEGIN .*PRIVATE KEY" $(git ls-files) ; expect no matches
- git ls-files | grep -E "probectl-control$|coverage.out$" ; expect empty
- gitleaks detect --no-banner ; expect 0 leaks
- Open a draft PR that breaks a test; confirm merge is blocked.

CI gate: new jobs `verify-branch-protection` and `secret-scan` are required checks in .github/workflows/ci.yml.
```

## Sprint 2 — Control-plane coverage gate + commit signing + safe error handling
**Closes:** TEST-001 (H), CODE-001 (H), CODE-004 (M), CODE-002 (H) · **Depends on:** Sprint 1 · **Size:** ~3d

```text
You are working in the probectl Go monorepo. Goal: close the control-plane test-coverage blind spot, enforce commit signing going forward, and fix silently-discarded errors.

Tasks:
1. Cover internal/control (TEST-001): add table-driven tests for every HTTP handler in internal/control/ — auth middleware (session + dev rejection), tenant-scoped routes, error paths. Wire internal/control into the coverage gate with a floor (start at the current measured %, ratchet up). Fail CI if internal/control coverage drops.
2. Commit signing going forward (CODE-001, CODE-004): require signed commits via the branch ruleset from Sprint 1 (require_signatures). Document the GPG/SSH-signing setup in CONTRIBUTING.md. Add a CI check that rejects unsigned commits on PR branches. (Historical unsigned commits are noted as accepted prior state in docs/quality/provenance.md.)
3. Fix silent JSON unmarshal discards (CODE-002) — confirmed sites from triage: internal/store/abac.go:27,30 (an ABAC policy that fails to decode silently becomes empty — with deny-override semantics that is fail-OPEN; fix first), internal/audit/worm.go:250, internal/store/users.go:32, internal/store/changes.go:29. Make each return/log+fail-closed; sweep for any others (`grep -rn "_ = json.Unmarshal"`). Add tests proving malformed input is rejected, not silently treated as empty.

Acceptance:
- internal/control has a non-trivial coverage floor enforced in CI; new handlers require tests.
- Unsigned commits are rejected on PRs; CONTRIBUTING.md documents signing.
- No JSON decode error is discarded on auth/audit/remediation paths.

Verify:
- go test ./internal/control/... -cover
- grep -rn "_ = json.Unmarshal\|json.Unmarshal(" internal | grep -v _test ; review each for error handling
- go test ./... -run Unmarshal -v
- Push an unsigned commit to a test branch; confirm the signing check fails.

CI gate: coverage floor for internal/control added to the cover-gate job; `require-signed-commits` is a required check.
```

---

# Phase 1 — Clear the non-licensing Criticals

> These two Criticals are holdback blockers and gate a clean red-team retest. Merging Sprints 3 and 4 removes 2 of the 4 Criticals.

## Sprint 3 — Kill the dev-mode authentication bypass  `[CRITICAL]`
**Closes:** RED-001 (C), SEC-001 (H) · **Depends on:** Sprint 1 · **Size:** ~2d · **Clears:** CRIT-AUTHBYPASS

```text
You are working in the probectl control plane. A dev auth mode grants a trusted-header principal ALL permissions; today it is selectable at runtime and gated only by a log warning.

Evidence to start from:
- internal/config/config.go:501 — AuthMode = enum("PROBECTL_AUTH_MODE","session","dev","session")
- internal/config/config.go:138 — comment: "dev" = trusted-header dev principal with all permissions
- cmd/probectl-control/main.go:129 — `if cfg.AuthMode == "dev"` branch

Tasks:
1. Make dev auth physically impossible in a release build. Put the dev-auth code path behind a Go build tag (e.g. //go:build devauth) so it is NOT compiled into release binaries at all. Release/production build targets in the Makefile and release.yml must build without that tag.
2. Defense in depth at runtime: if AuthMode=="dev" is somehow set, refuse to start unless BOTH a non-production build tag is present AND an explicit PROBECTL_DEV_AUTH_ACK=i-understand env is set; otherwise fatal-exit (not just warn). Never allow dev auth when bound to a non-loopback interface.
3. Emit a loud startup audit event + metric when dev auth is active.
4. Tests: prove a release-tagged build rejects/omits dev mode; prove session mode is the only path in default builds.

Acceptance:
- A binary built with the release target has no dev-auth code path (verified by symbol/string absence).
- Default config = session (fail-closed); dev auth requires build tag + explicit ack + loopback.
- Cross-tenant/red-team retest of this vector fails to bypass auth.

Verify (AMENDED — the original strings-grep checked a source COMMENT, which never reaches a binary and passes vacuously):
- go build -tags="" -o /tmp/pc ./cmd/probectl-control && ! go tool nm /tmp/pc | grep -qi "devauth\|devPrincipal" && echo SYMBOLS_OK
- go test ./internal/auth/... ./cmd/probectl-control/... -run "Auth|Dev" -v
- PROBECTL_AUTH_MODE=dev /tmp/pc ; expect fatal exit in a release build (behavioral check — this is the authoritative one)

CI gate: a `no-devauth-in-release` job greps the release binary for dev-auth strings/symbols and fails if present.
```

## Sprint 4 — Server-side tenant binding on ingest  `[CRITICAL]`
**Closes:** TENANT-101 (C), WIRE-001 (H), RED-006 (M), TENANT-107 (L) · ~~RED-004 struck — fixed at commit 93529b0 (stampTenant overwrites in place; FuzzOTLPPayload regression-guards it)~~ · **Depends on:** Sprint 3 · **Size:** ~5d · **Clears:** CRIT-XTENANT

```text
You are working in the probectl telemetry ingestion path. Today the flow/device/eBPF ingest trusts a client-asserted tenant_id: the agent stamps records with its own configured tenant and the server accepts it, enabling cross-tenant injection on bus-published planes.

Evidence to start from:
- internal/flow/collector.go:207 — recs[i].TenantID = c.cfg.TenantID (agent-side stamp)
- internal/flow/config.go:29-32,128,151 — TenantID from agent config / PROBECTL_FLOW_TENANT
- WIRE-001 / TENANT-101 in outputs/WIRE.html, outputs/TENANT.html
- RED-006 (publish-path silo routing), TENANT-107 (flow/device/eBPF emitters ignore siloed bus namespaces — confirmed: no BusNamespace handling in any emitter)
- NOTE: RED-004 is already fixed (93529b0) — keep its fuzz corpus as the regression guard; do not re-implement

Tasks:
1. Re-stamp tenant server-side from the authenticated agent identity (SPIFFE SVID / enrollment claim), NOT from the payload. On every ingest entrypoint (flow, device, eBPF, OTLP), derive tenant_id from the verified mTLS/SVID principal and OVERWRITE any client-supplied tenant_id; reject if the authenticated identity has no tenant binding.
2. Siloed emitter routing (TENANT-107): make the flow/device/eBPF emitters resolve and publish to the tenant's siloed bus namespace when the tenant is siloed/hybrid; a siloed tenant's records must never ride the shared topics. (Same code path as task 1 — do it once here, S6 adds the tests.)
3. Fix silo routing fail-open (RED-006): if siloed tenant bus-namespace routing fails, FAIL CLOSED (drop + alert), never silently degrade to the shared lane.
4. Add cross-tenant injection tests: a client presenting tenant A's identity but payload tenant B must be re-stamped to A (or rejected), and must never land in B's partition/table.

Acceptance:
- No ingest path accepts a payload-supplied tenant_id as authoritative; tenant derives from authenticated identity.
- Silo routing failure is fail-closed.
- Red-team cross-tenant injection test (payload tenant != identity tenant) is blocked and covered by CI.

Verify:
- go test ./internal/flow/... ./internal/device/... ./internal/ebpf/... ./internal/pipeline/... -run "Tenant|Stamp|Inject" -v
- go test -fuzz=FuzzStampTenant -fuzztime=60s ./internal/pipeline/...  [if fuzz target present]
- [needs infra] integration: publish a record with mismatched tenant; assert it appears only under the identity's tenant in ClickHouse.

CI gate: cross-tenant ingest tests are required and run un-tagged (see Sprint 6).
```

## Sprint 5 — DB-level tenant isolation backstop
**Closes:** TENANT-102 (H), TENANT-104 (M), TENANT-106 (M) · **Depends on:** Sprint 4 · **Size:** ~3d

```text
You are working in probectl's datastore layer (Postgres + ClickHouse). Add database-level isolation so an application bug cannot cross tenants, and stop at-rest encryption from silently degrading.

Evidence to start from:
- TENANT-102: ClickHouse telemetry service account has no DB-level isolation backstop
- TENANT-104: Postgres RLS bypassable via the raw pool if the DB login role is privileged; no startup assertion
- TENANT-106: at-rest tenant encryption silently degrades to plaintext when unconfigured

Tasks:
1. Postgres RLS startup assertion (TENANT-104 — TRIAGED: the role posture already exists — migrations/0007_app_role.sql creates probectl_app NOLOGIN NOSUPERUSER NOBYPASSRLS; what's missing is the boot-time self-check): add a startup self-check that queries pg_roles/current_setting and FATALs if the connected role can bypass RLS, is superuser, or if FORCE ROW LEVEL SECURITY is missing on any tenant table. Add a migration test asserting rolbypassrls=false.
2. ClickHouse service-account scoping (TENANT-102 — TRIAGED: row policies already exist — internal/store/flowstore/clickhouse.go:229-230 creates probectl_tenant_isolation per table; the residual is that the SERVICE user is policy-exempt by design, `USING 1 TO serviceUser`): narrow the service account's blast radius — either per-request `SET` + a policy bound to that setting, or split read/write users so the query path cannot read cross-tenant even if the app is compromised. Document the resulting threat model in docs/security/tenant-isolation.md, including what the exemption still permits.
3. At-rest encryption fail-closed (TENANT-106): if tenant at-rest encryption is enabled in config but no key is resolvable, FATAL at startup instead of writing plaintext. Add a config validation + test.

Acceptance:
- App cannot start with an RLS-bypassing DB role or unforced RLS.
- ClickHouse reads are constrained by a tenant row policy; a cross-tenant query returns nothing.
- Misconfigured at-rest encryption fails closed, never silent plaintext.

Verify:
- go test ./internal/store/... ./internal/tenancy/... -run "RLS|Isolation|Encrypt" -v
- [needs infra] psql: SELECT rolbypassrls,rolsuper FROM pg_roles WHERE rolname=current_user; expect f,f
- [needs infra] CH: issue a query for tenant B while scoped to tenant A; expect 0 rows.

CI gate: migration test asserting non-bypassrls app role; isolation tests required.
```

## Sprint 6 — Cross-tenant test suite (expanded; harness pattern kept)
**Closes:** TENANT-105 (M), TENANT-107 (L) · **Depends on:** Sprint 4, Sprint 5 · **Size:** ~3d

```text
You are working in probectl's test suite. Expand the cross-tenant suite to cover the new ingest-path guarantees end to end.

FOUNDER DECISION (triage): the build-tag/skip-harness pattern STAYS — these tests need live Postgres/ClickHouse, and un-gating them would break `go test ./...` on DB-less dev boxes. The `make test-isolation` target already exists (Makefile:121, runs with -race) and the `cross-tenant-isolation` CI job is already required. The work is the MISSING CASES, not the plumbing.

Evidence to start from:
- TENANT-105: the suite predates Sprint 4/5 — no ingest-path injection coverage, no service-account-scoping coverage

Tasks:
1. Ingest-path injection cases (from Sprint 4): for EACH ingest surface (flow, device, eBPF, OTLP), a client presenting tenant A's identity with payload tenant B must be re-stamped to A (or rejected) and must never land in B's partition/table — asserted against real PG + CH.
2. Query-path cases (from Sprint 5): with the CH service-account scoping in place, assert a cross-tenant query returns nothing through every store reader.
3. Siloed-namespace cases (from Sprint 4 / TENANT-107): a siloed tenant's records appear ONLY on its namespaced topics; shared topics never carry them.

Acceptance:
- The isolation suite covers ingest → bus → store → query end to end, including the S4 injection vectors; runs in the existing required CI job.

Verify:
- make test-isolation
- go test -tags integration ./... -run "CrossTenant|Isolation|Silo|Inject" -v   [needs infra]

CI gate: the existing `cross-tenant-isolation` required job picks up the new cases (no new plumbing).
```

---

# Phase 2 — Control-plane security hardening (SEC, weight 10)

## Sprint 7 — ClickHouse driver + parameterized queries
**Closes:** ARCH-002 (H), SEC-005 (L), TENANT-108 (L) · **Depends on:** Sprint 5 · **Size:** ~4d

```text
You are working in probectl's ClickHouse access layer. Today ClickHouse is reached via the raw HTTP API with queries built by string concatenation + manual escaping (no driver, no bind parameters) — an injection and maintainability risk that recurs across audits.

Evidence to start from:
- ARCH-002: ClickHouse accessed via raw HTTP API — no driver, no parameterized queries
- SEC-005 / TENANT-108: CH telemetry queries built by manual string escaping

Tasks:
1. Adopt the official ClickHouse Go driver (clickhouse-go/v2) OR the native protocol client; replace raw-HTTP query construction in internal/store (pathstore, tsdb, flow/device telemetry readers).
2. Convert ALL CH queries to parameterized/bound queries; eliminate manual escaping. Tenant scoping must be a bound parameter, not string-interpolated.
3. Add a lint/CI check (e.g. a small AST or grep gate) that fails if a CH query is built with fmt.Sprintf/string concatenation.
4. Tests: query-builder unit tests + injection attempt tests (a tenant_id like `x' OR '1'='1` must be safely bound).

Acceptance:
- No raw-HTTP, string-concatenated CH queries remain; all use bound parameters.
- Injection attempt tests pass (no escape).
- CI fails on reintroduced string-built CH SQL.

Verify:
- grep -rn "Sprintf(" internal/store | grep -i "select\|insert\|where" ; expect none for CH paths
- go test ./internal/store/... -run "Query|Inject|Param" -v

CI gate: `no-stringbuilt-sql` check added to lint job.
```

## Sprint 8 — At-rest telemetry encryption by default
**Closes:** SEC-002 (M), COMPLY-004 (M) · **Depends on:** Sprint 5 · **Size:** ~3d

```text
You are working in probectl deployment + crypto config. The compose/raw-binary default is keyless plaintext-passthrough at rest; the operator's duty to encrypt bulk telemetry stores is undocumented.

Evidence to start from:
- SEC-002: keyless plaintext-passthrough at rest is the default in compose/raw-binary recipe
- COMPLY-004: bulk telemetry at-rest encryption delegated to operator without documented duty

Tasks:
1. Change the shipped default so at-rest encryption for tenant telemetry is ON, backed by a generated-or-required key (fail closed per TENANT-106). Provide a documented KMS/file-key option.
2. For the storage layers probectl does not encrypt itself (ClickHouse/Postgres/object store volumes), add an explicit operator obligation to docs/hardening.md + a preflight check that warns (or fails in `--strict`) if it detects an unencrypted volume/driver.
3. Update deploy/compose/.env.example and Helm values with encryption settings ON by default.

Acceptance:
- Default deployment encrypts probectl-managed at-rest telemetry; no silent plaintext.
- docs/hardening.md documents the operator's storage-encryption duties; preflight check exists.

Verify:
- go test ./internal/crypto/... ./internal/tenantcrypto/... -run "AtRest|Encrypt" -v
- grep -n "encrypt" deploy/compose/.env.example deploy/helm/*/values.yaml
- run the preflight against an unencrypted fixture; expect warning/strict-fail.

CI gate: config-validation test asserting encryption-on default.
```

## Sprint 9 — Auth & session hardening
**Closes:** SEC-003 (M), SEC-004 (L), SEC-009 (L), RED-007 (M, verify-first) · ~~RED-009 struck — Config.LogValue is a slog allowlist (internal/config/config.go:733); startup logs pass through it~~ · **Depends on:** Sprint 3 · **Size:** ~2d

```text
You are working in probectl auth/session/transport. Close brute-force, OIDC, cookie, MCP-stdio, and secret-logging gaps.

Evidence to start from:
- SEC-003: no brute-force throttle/lockout on provider operator login (highest-privilege)
- SEC-004: OIDC nonce generated and sent but never validated on callback
- SEC-009: session-cookie Secure flag tied to app's own TLS, not deployment edge
- RED-007: MCP stdio — TRIAGED: internal/ai/mcp/stdio.go:13-16 documents caller-side token auth ("the binary authenticates the token before calling this"); verify the binary entry actually enforces it

Tasks:
1. Throttle + lockout on provider/operator login (SEC-003): add per-account + per-IP rate limiting with exponential backoff and temporary lockout; audit-log lockouts. Reuse internal/breaker if suitable.
2. Validate OIDC nonce on callback (SEC-004): compare the returned nonce to the stored value; reject mismatch. Add a test.
3. Cookie Secure (SEC-009): set Secure (and SameSite, HttpOnly) based on deployment-edge config (e.g. PROBECTL_PUBLIC_TLS=true), not only the app's own listener TLS.
4. MCP stdio auth (RED-007 — verify-first): confirm the binary entry that calls ServeStdio actually authenticates the token as the doc contract states; add the missing enforcement if absent, and a test either way. Document the local-trust model in docs/mcp.md.

Acceptance:
- Repeated bad logins are throttled/locked and audited (NOTE: the tenant login path already has authLimiter + lockout (U-024, internal/control/authlimit.go) — the gap is the PROVIDER/operator login in ee/provider/handler.go, which does its own authn outside the limiter; reuse internal/auth.Limiter there); OIDC nonce enforced on callback (confirmed missing: nonce generated at internal/control/auth.go:282, sent via AuthCodeURL, never stored or compared); cookies Secure at the edge (confirmed: auth.NewManager ties Secure to cfg.TLSEnabled() — the app's own listener, wrong behind a TLS-terminating ingress); MCP stdio token path verified+tested.

Verify:
- go test ./internal/auth/... ./internal/scim/... -run "Throttle|Lockout|Nonce|Cookie" -v
- go test ./... -run "MCPAuth" -v
- grep -rn "Secure:" internal/control internal/auth | head

CI gate: auth-hardening tests required.
```

## Sprint 10 — SSRF, info-leak & shipped-credential cleanup
**Closes:** SEC-006 (L), SEC-007 (L), SEC-008 (L), OPS-003 (M), OPS-010 (L) · **Depends on:** Sprint 1 · **Size:** ~2d

```text
You are working in probectl's SSRF guard, error surfaces, and shipped deploy recipes. Close the residual low-severity hardening gaps.

Evidence to start from:
- SEC-006: default-weak datastore password in the shipped (non-dev) compose recipe
- SEC-007: SSRF guard does not deny 0.0.0.0/8 (only the exact unspecified address)
- SEC-008: SCIM store errors echo internal text; /openapi.json and /version unauthenticated info
- OPS-003: compose backup services embed plaintext dev credentials
- OPS-010: sslmode=disable in CI integration fixtures

Tasks:
1. SSRF (SEC-007): extend the SSRF denylist to the full 0.0.0.0/8 range (and re-confirm link-local 169.254/16, metadata 169.254.169.254, loopback, RFC1918, ULA). Add table tests for each blocked range.
2. Shipped secrets (SEC-006, OPS-003): remove default/weak passwords from non-dev compose; require operator-provided secrets (fail to start if unset); move backup-service creds to env/secret refs. Keep dev recipe clearly separated and labeled.
3. Info leak (SEC-008): return generic errors for SCIM store failures (log details server-side only); gate /openapi.json and /version behind auth or strip version detail in production.
4. CI fixtures (OPS-010): change sslmode=disable to require/verify-full in integration fixtures (use a test CA), so CI exercises TLS like production.

Acceptance:
- SSRF guard blocks 0.0.0.0/8; no weak default secrets ship; SCIM errors are generic; CI uses TLS to Postgres.

Verify:
- go test ./internal/... -run "SSRF|Guard" -v
- grep -rn "sslmode=disable" . ; expect only documented dev-only spots (or none)
- grep -rniE "password.*(changeme|postgres|admin|probectl)" deploy/compose ; expect none in non-dev recipe

CI gate: SSRF range tests required; secret-scan (Sprint 1) covers shipped creds.
```

---

# Phase 3 — Agent identity, enrollment & transport (WIRE, weight 11)

## Sprint 11 — Agent enrollment & SVID issuance
**Closes:** WIRE-002 (H), RED-002 (H), TENANT-103 (M), ARCH-004 (M) · **Depends on:** Sprint 4 · **Size:** ~6d

```text
You are working on probectl's agent trust root. Agent enrollment / SPIFFE SVID issuance is absent from the repo (explicitly deferred, S-EE1): the trust root and bootstrap are unauthenticated/operator-manual, which undermines the server-side tenant binding from Sprint 4.

Evidence to start from:
- WIRE-002 / RED-002: agent enrollment + SVID/CA issuance absent; bootstrap unauthenticated
- TENANT-103: telemetry isolation trust root depends on out-of-repo SVID issuance/enrollment
- ARCH-004: SPIFFE SVID issuance stub — agent cert lifecycle is operator-manual

Tasks:
1. Implement an enrollment service: a join-token (or cloud-IID/OIDC) bootstrap where an agent presents a one-time, tenant-scoped enrollment token and receives a short-lived SVID bound to its tenant + agent identity. Store issued identities; bind tenant claim into the SVID.
2. Implement SVID issuance + automatic rotation (short TTL) with a documented CA hierarchy (root → intermediate → leaf). Make the tenant binding from Sprint 4 derive from this SVID.
3. Provide an enrollment CLI/flow in cmd/probectl-agent and docs/agent/enrollment.md.
4. Tests: enrollment happy-path, token replay rejection, wrong-tenant rejection, rotation.

Acceptance:
- An agent cannot ingest without a valid enrollment-issued, tenant-bound SVID; tokens are single-use; SVIDs rotate.
- Sprint 4's server-side tenant binding now reads a real, repo-managed identity.

Verify:
- go test ./internal/agent/... ./internal/agenttransport/... ./internal/tenancy/... -run "Enroll|SVID|Rotate|Bootstrap" -v
- [needs infra] end-to-end: enroll an agent, confirm SVID issued with tenant claim, confirm ingest works only after enrollment.

CI gate: enrollment + identity tests required.
```

## Sprint 12 — Cert revocation + transport TLS + replay protection
**Closes:** WIRE-003 (H), WIRE-004 (M), WIRE-005 (L), WIRE-006 (L), WIRE-007 (L), RED-008 (L) · **Depends on:** Sprint 11 · **Size:** ~5d

```text
You are working on probectl's transport security. Wire up revocation, enforce TLS everywhere, unify TLS config, add replay protection, and constrain ca_file.

Evidence to start from:
- WIRE-003 — TRIAGED: revocation IS wired into the handshake (G4/U-038: agenttransport/server.go:43 uses crypto.ServerMTLSConfigRevocable; refused-handshake tests exist). The residual: NOTHING FEEDS the list — no registry/operator path calls Server.RevocationList()
- WIRE-004: control HTTP API serves plaintext on :8080 by default; TLS optional/unenforced (internal/control/server.go:413-415)
- WIRE-005: OTLP & MCP receivers use weaker TLS than the rest
- WIRE-006: no application-layer replay/freshness protection on ingestion
- WIRE-007: TLS 1.2 floor (not 1.3) across the fleet
- RED-008: ca_file parameter reads arbitrary agent filesystem paths

Tasks:
1. Feed the revocation list (WIRE-003 residual): the handshake check exists — add the operator path: a `probectl-control revoke-agent <tenant> <agent>` command + an admin API that resolves the agent's cert serial/SPIFFE id from the registry and pushes it into Server.RevocationList() (Replace/RevokeSerial), persisted so restarts keep revocations. Test: revoke via the CLI, assert the next handshake is refused.
2. Enforce TLS (WIRE-004): make TLS mandatory for the control API; remove/guard the plaintext ListenAndServe so plaintext requires an explicit, loud, non-default opt-in (or only loopback). Default = TLS.
3. Unify TLS config (WIRE-005, WIRE-007): one hardened tls.Config (TLS 1.3 floor where feasible, AEAD-only cipher suites, curve preferences) reused by ALL receivers incl. OTLP + MCP.
4. Replay/freshness (WIRE-006): add a signed timestamp/nonce + bounded window on ingestion; reject stale/replayed batches.
5. Constrain ca_file (RED-008): restrict ca_file to an allowlisted dir; reject path traversal.

Acceptance:
- Revoked agents are refused; control API is TLS-only by default; all receivers share the hardened TLS config; replayed batches rejected; ca_file constrained.

Verify:
- go test ./internal/agenttransport/... ./internal/control/... -run "Revoke|TLS|Replay|Freshness|CAFile" -v
- grep -n "ListenAndServe\b" internal/control/server.go ; confirm guarded/non-default
- [needs infra] connect a revoked cert; expect refusal.

CI gate: transport-security tests required; a check that no receiver builds a bespoke weak tls.Config.
```

## Sprint 13 — Harden the StreamConfig stub (U-044 ADR stands)
**Closes:** ARCH-003 (H) · **Depends on:** Sprint 1 · **Size:** ~1d

```text
You are working in probectl's agent proto/control surface. FOUNDER DECISION (triage): the U-044 ADR stands — StreamConfig stays in the schema for wire compatibility (docs/adr/config-push.md: de-document, don't implement; removal would need the blocking buf-breaking exception process for zero capability gain). Close the audit's attack-surface concern WITHIN that decision.

Evidence to start from:
- proto/probectl/agent/v1/agent.proto:23-30 — StreamConfig documented as UNIMPLEMENTED STUB (U-044)
- internal/agenttransport/service.go — the stub currently sends one empty epoch-0 frame and HOLDS THE STREAM OPEN
- The agent has no config-apply path at all (it never processes StreamConfig payloads), so "agent-RCE surface" overstates it — but the held-open stream is pointless surface.

Tasks:
1. Server: replace the hold-open stub with an immediate explicit deny — return codes.Unimplemented ("config push is not a capability; see docs/adr/config-push.md (U-044)") without sending any frame. No proto change.
2. Agent: delete any client-side StreamConfig call/stub usage so no code path even initiates the stream.
3. Tests: (a) the server returns Unimplemented; (b) static assertion that the agent binary has no StreamConfig client invocation; (c) the existing proto conformance test still passes (RPC remains in the schema).
4. ADR addendum: one paragraph in docs/adr/config-push.md recording this hardening + the blind-audit finding it answers (ARCH-003).

Acceptance:
- The RPC remains wire-compatible; the server explicitly denies; the agent never calls it; ADR updated.

Verify:
- grep -rn "StreamConfig" internal/agent cmd/probectl-agent ; expect no client invocation
- go test ./internal/agenttransport/... -run "StreamConfig" -v
- buf lint && go build ./...   (NO buf breaking change should be flagged)

CI gate: Unimplemented-deny test required; buf breaking stays green (no schema change).
```

---

# Phase 4 — Telemetry scale & reliability (SCALE, the biggest absolute lift: 10 → ~90)

> SCALE is the weakest domain. These four sprints rebuild the data-plane hot paths and then *prove* the SLOs (clearing blocker B5). Sprint 17 is the validation gate.

## Sprint 14 — Ingest throughput & backpressure
**Closes:** SCALE-001 (H), SCALE-009 (M), SCALE-012 (M), SCALE-013 (M-L) · ~~SCALE-008 struck — retry+DLQ shipped in D2/U-040 (internal/pipeline/retry_dlq_test.go); one verify-test remains~~ · **Depends on:** Sprint 4 · **Size:** ~5d

```text
You are working in probectl's ingestion consumer / pipeline. The consume path is single-threaded per topic with a synchronous remote-write, telemetry is dropped on transient store failure, and messages are re-decoded many times.

Evidence to start from:
- SCALE-001: ingest consume path single-threaded per topic with synchronous remote-write
- SCALE-009: path store inserts are per-path-run (NOTE: hops/links are already batched per path via JSONEachRow slices, pathstore/clickhouse.go:158,173 — the residual is a cross-path batching window)
- SCALE-012: per-heartbeat Postgres UPDATE scales linearly with fleet, unbatched
- SCALE-013: fan-out re-decode multiplier — each message independently unmarshaled by many consumers

Tasks:
1. Parallelize the consumer (SCALE-001): partitioned worker pool per topic with bounded concurrency + ordered-by-key where required; decouple decode from remote-write via a buffered channel.
2. Decode once, fan out (SCALE-013): unmarshal each message a single time into a shared immutable struct passed to downstream consumers.
3. Batching (SCALE-009, SCALE-012): batch path-store INSERTs and heartbeat UPDATEs (size + time window) using CH async/batch insert and Postgres COPY/multi-row upsert.
4. Verify-test (SCALE-008 residual): retry+DLQ already exists in the pipeline (D2) — add ONE test proving the device-plane write path routes through it on transient store failure (the flow path is already covered); keep the drop metric ~0 assertion.
5. Add benchmarks for the consumer hot path (feeds Sprint 17).

Acceptance:
- Consumer scales with cores/partitions; no per-record synchronous remote-write on the hot path; inserts/updates are batched; transient failures retry/DLQ instead of dropping.

Verify:
- go test ./internal/pipeline/... ./internal/bus/... ./internal/store/... -run "Consume|Batch|Backpressure|Retry|DLQ" -v
- go test -bench=BenchmarkIngest -benchmem ./internal/pipeline/...
- [needs infra] kill the store mid-load; assert zero dropped records (all retried/DLQ'd).

CI gate: ingest benchmarks run in nightly; a drop-rate test is a required PR check.
```

## Sprint 15 — Cardinality control & rate limiting
**Closes:** SCALE-003 (H), SCALE-004 (M), SCALE-005 (H), SCALE-007 (M), SCALE-011 (M) · **Depends on:** Sprint 14 · **Size:** ~5d

```text
You are working in probectl's cardinality + fairness layer. Limiters are per-process/in-memory/never-evicted, per-tenant rate limiting is OFF by default (fail-open), the device plane has no cap, tenant-keyed partitioning hot-spots, and enrichment is synchronous on the hot path.

Evidence to start from:
- SCALE-003: cardinality limiter & fairness gate are per-process, in-memory, never-evicted (unbounded growth, no cross-replica view)
- SCALE-004: per-tenant ingest rate limiting OFF by default (fail-open) — internal/fairness/fairness.go (rate<=0 = unlimited)
- SCALE-005: device-telemetry plane (SNMP/gNMI) ingested with NO cardinality cap and NO rate limit
- SCALE-007: tenant-keyed bus partitioning creates a hot-partition ceiling for one large tenant
- SCALE-011: synchronous per-record ASN/geo enrichment on the flow hot path

Tasks:
1. Bound + evict the EXISTING limiter (SCALE-003 — TRIAGED: per-tenant cardinality caps shipped in D3/U-026, internal/pipeline/cardinality.go with rejection at consumer.go:207; the residual is unbounded key growth and per-replica state): give the existing structure an LRU/TTL bound; cross-replica sharing (Redis/CH counters) is OPTIONAL — add only if the multi-replica deployment profile needs it, and note the trade-off (a new stateful dependency vs slightly-over-limit tolerance). Emit limiter metrics either way.
2. Rate limiting defaults (SCALE-004 confirmed — fairness.go rate<=0 = unlimited; SCALE-005 verify-first — the pipeline cardinality cap at consumer.go:207 likely already covers device-plane series): ship sane per-tenant default limits (not unlimited) for ALL planes; make "unlimited" an explicit opt-in. VERIFY device-plane coverage by the existing cap before adding a second one.
3. Hot-partition mitigation (SCALE-007): add sub-partitioning (tenant_id + hash bucket) so a single large tenant spreads across partitions; document the trade-off.
4. Async enrichment (SCALE-011): move ASN/geo enrichment off the synchronous hot path (worker pool + cache); degrade gracefully if enrichment lags.

Acceptance:
- Limiters are bounded + shared across replicas; all planes have default rate + cardinality caps; large tenants don't hot-spot a single partition; enrichment is async.

Verify:
- go test ./internal/fairness/... ./internal/pipeline/... -run "Cardinality|RateLimit|Evict|Partition|Enrich" -v
- go test -bench=BenchmarkEnrich -benchmem ./internal/pipeline/...
- load a single tenant past the cap; assert throttling + bounded memory.

CI gate: cardinality/rate-limit default tests required; memory-bound assertion in CI.
```

## Sprint 16 — Storage durability: path-table TTL, OTLP consumer, TSDB index, silences
**Closes:** SCALE-006 (H — path tables only; flow TTL already exists, flowstore/clickhouse.go:108), SCALE-010 (H), SCALE-014 (L), ARCH-005 (M — silences/acks only; U-047 ADR stands) · **Depends on:** Sprint 7, Sprint 14 · **Size:** ~4d

```text
You are working in probectl's storage + OTLP plane. Path/traceroute CH tables have no TTL, the OTLP topic has no consumer, the default in-memory TSDB query is O(n), and several stores are lost on restart.

Evidence to start from:
- SCALE-006: path/traceroute CH tables partition by tenant_id only with NO TTL (unbounded growth)
- SCALE-010: OTLP ingestion plane publishes to topic probectl.otlp.metrics with NO consumer (internal/bus/bus.go:26) — external OTLP silently dropped
- SCALE-014: default in-memory TSDB query is an O(n) linear scan (the path the CLI hits)
- ARCH-005: in-memory stores (topology, detections, snapshots) lost on restart

Tasks:
1. Path-table TTL (SCALE-006 residual): the flow tables already get `MODIFY TTL ... DELETE` from per-tenant retention at boot (U-046 ledger) — apply the same idempotent boot-applied TTL + (tenant_id, time-bucket) partitioning to the PATH/traceroute tables only.
2. Wire the OTLP consumer (SCALE-010): implement the consumer for probectl.otlp.metrics (and traces/logs once Sprint 22 lands) so externally-ingested OTLP is actually stored; add a test that an OTLP push is queryable.
3. TSDB query (SCALE-014): replace the O(n) scan with an indexed/bucketed structure (or route the CLI to the CH-backed store); add a benchmark.
4. Silences/acks persistence (ARCH-005 — FOUNDER DECISION: the rebuild-on-restart ADR stands, docs/adr/volatile-stores.md (U-047): topology/detections are caches of durable streams with cold-start tests; re-litigating that is out of scope): persist ONLY alert silences/acks — the ADR's own documented exception (operator inputs, not derivable). Store them like alert rules; add the restart-survival test the ADR calls for.

Acceptance:
- Telemetry tables have retention/TTL; OTLP data is consumed + queryable; TSDB query is sub-linear; key stores survive restart.

Verify:
- go test ./internal/store/... ./internal/otel/... ./internal/topology/... -run "TTL|Retention|OTLP|TSDB|Persist|Restart" -v
- grep -rn "OTLPMetricsTopic" internal ; confirm a consumer references it
- [needs infra] push OTLP metrics; query them back; restart control plane; confirm topology persists.

CI gate: OTLP round-trip test + retention migration test required.
```

## Sprint 17 — Scale validation RUN (harness exists) + missing benchmarks  `[BLOCKER B5]`
**Closes:** SCALE-002 (H, BLOCKER), SCALE-015 (L), DOCS-001 (H), DOCS-006 (M) · **Depends on:** Sprint 14, 15, 16 · **Size:** ~5d

```text
You are working on probectl's scale validation. Headline scale SLOs are PROVISIONAL — no full-scale (L/XL) run has ever executed. TRIAGED: the load harness ALREADY EXISTS (E1/U-005: `make load-test TIER=S|M|L|XL`, S-tier smoke runs in CI on every pass, scale-gate materiality floor enforced since D12). This sprint is the RUN + the missing benchmarks, not a build.

Evidence to start from:
- SCALE-002 (BLOCKER B5): headline scale SLOs provisional; no L/XL run executed
- SCALE-015: only 2 benchmarks; none cover the control-plane ingest consumer, TSDB, or path store
- DOCS-001: scale SLOs explicitly provisional — L/XL result tables empty
- DOCS-006: agent overhead numbers are fixture/userspace-only — live ring-buffer path unvalidated

Tasks:
1. Use the existing harness (`make load-test TIER=L|XL`); extend its profiles only where a plane is missing from the drive set.
2. Add the missing benchmarks (SCALE-015 — confirmed: only 2 benchmark files in internal/ today): ingest consumer, TSDB query, path store write.
3. Run the L/XL profile on reference hardware [needs infra]; capture throughput, p50/p99 latency, drop rate, memory/cardinality behavior; fill the empty SLO tables in docs/ with REAL numbers + the run environment.
4. Validate live agent overhead (DOCS-006): measure the real eBPF ring-buffer path (not the fixture) and publish overhead numbers.
5. Add a `make scale-gate` target + a nightly/manual CI job that runs the M profile as a regression guard and fails if SLOs regress beyond a threshold.

Acceptance:
- SLO tables in docs are filled from a real L/XL run with the environment documented; benchmarks cover the hot paths; a scale regression gate exists.

Verify:
- go test -bench=. -benchmem ./internal/pipeline/... ./internal/store/...
- make scale-gate   [needs infra for L/XL; M profile runnable in CI]
- grep -rn "PROVISIONAL\|TBD" docs | grep -i scale ; expect none remain

CI gate: `scale-gate` (M profile) runs in nightly; SLO doc has no PROVISIONAL placeholders.
```

---

# Phase 5 — eBPF & privileged agent (EBPF, weight 14; already 80 — close the rest)

## Sprint 18 — eBPF TLS-capture scoping + kernel-side redaction
**Closes:** EBPF-001 (M), EBPF-002 (M), RED-003 (H) · **Depends on:** Sprint 11 · **Size:** ~5d

```text
You are working in probectl's eBPF L7/TLS capture (internal/ebpf, bpf/sslsniff.bpf.c). Today sslsniff captures TLS plaintext host-wide (all libssl processes) and redaction is userspace/post-capture, so full plaintext (URLs/PII) transits the kernel ring for ALL processes.

Evidence to start from:
- EBPF-001 / RED-003: sslsniff captures TLS plaintext host-wide; no process-scope filter
- EBPF-002: redaction is userspace/post-capture; full plaintext transits kernel ring
- internal/ebpf/l7source.go (TLS-uprobe source), bpf/sslsniff.bpf.c

Tasks:
1. Process-scope the capture (EBPF-001, RED-003): add an allowlist (cgroup id / PID / container label / binary path) so sslsniff only attaches to / emits for explicitly opted-in workloads, not every libssl process on the host. Default = capture disabled (consistent with the existing double-consent gate).
2. Move redaction earlier (EBPF-002): perform field/URL/PII redaction (or length-only / hashed capture) in the eBPF program or at the earliest userspace boundary, before plaintext is buffered/forwarded, so unredacted plaintext does not transit the full ring for non-targeted data.
3. Tests: process-scope filter unit tests; a redaction test proving sensitive fields never reach the forwarder.

Acceptance:
- TLS capture only applies to opted-in processes; non-targeted process plaintext is never captured; redaction happens before plaintext leaves the earliest boundary.

Verify:
- go test ./internal/ebpf/... -run "Scope|Filter|Redact|Consent" -v
- clang -target bpf -O2 -g -c bpf/sslsniff.bpf.c -o /tmp/sslsniff.o   [needs clang]
- [needs infra/kernel] load + attach with an allowlist; confirm a non-allowlisted process produces no events.

CI gate: eBPF unit tests + (in the kernel CI job) an allowlist attach test.
```

## Sprint 19 — Agent integrity + capability probe + CI matrix (denylist landed in S0)
**Closes:** EBPF-003 (M), EBPF-004 (M), EBPF-005 (L), EBPF-006 (L), EBPF-007 (L), EBPF-008 (L) · **Depends on:** Sprint 18 · **Size:** ~5d

```text
You are working on probectl's agent integrity, capabilities, and eBPF CI. Replace build-time self-checksum with a real signature, tighten capability probing + the observe-only denylist, add resource caps, and broaden the kernel test matrix.

Evidence to start from:
- EBPF-003: object integrity is a build-time self-checksum, not a signature (agent binary not signed/verified)
- EBPF-004: shipped default is a fixture replayer; the ~231-LOC live path is CI-tested but not the default
- EBPF-005: capability probe omits CAP_PERFMON (reports ready then fails at attach)
- EBPF-006: observe-only denylist omits bpf_probe_write_user
- EBPF-007: systemd unit lacks CPU/memory caps (Helm chart has them)
- EBPF-008: CI load-matrix narrow (2 kernels, x86 only); arm64 + hardened distros not load-tested

Tasks:
1. Load-time object signatures (EBPF-003 — TRIAGED: release artifacts are ALREADY cosign-signed (C6/U-067) and BPF objects are digest-verified before any kernel load (C9/U-014, VerifyObjectDigest)): the residual is load-time SIGNATURE verification if operator-supplied BPF objects are ever to be supported — otherwise document the embedded-digest model as the trust boundary and close. Decide which; if signatures, verify cosign sigs at load and refuse unsigned objects.
2. Shipped-image default (EBPF-004 — TRIAGED: fixture-vs-live is a BUILD-TAG fact, not a runtime flag — `make ebpf-agent` builds -tags ebpf with the live CO-RE loader): assert the SHIPPED agent image is the live (-tags ebpf) build via an image-build test/CI check, and document fixture mode as dev/test-only.
3. Capability probe (EBPF-005 — confirmed: capability_linux.go:33 checks CapBPF only): add CAP_PERFMON (needed for the L7 uprobes on >=5.8) and CAP_NET_ADMIN-if-needed to the readiness probe so the agent fails fast with a clear reason instead of failing at attach.
4. Observe-only denylist (EBPF-006): DONE IN SPRINT 0 (one-liner). Confirm bpf_probe_write_user is in observeonly_test.go's forbidden list and the re-audit note is recorded; nothing further here.
5. systemd resource caps (EBPF-007): add CPUQuota/MemoryMax (and tasks/IO limits) to the systemd unit to match the Helm chart.
6. Broaden CI matrix (EBPF-008): add arm64 and at least one hardened distro (lockdown/secureboot) kernel to the eBPF load matrix.

Acceptance:
- Agent refuses unsigned BPF objects; live path is default (or explicitly justified); probe covers all required caps; denylist includes bpf_probe_write_user; systemd has resource caps; CI loads on x86_64 + arm64 + a hardened kernel.

Verify:
- go test ./internal/ebpf/... ./cmd/probectl-ebpf-agent/... -run "Sign|Verify|Cap|Denylist|ObserveOnly" -v
- grep -n "bpf_probe_write_user" $(git ls-files internal/ebpf) ; confirm present in denylist
- grep -n "CPUQuota\|MemoryMax" deploy/**/ *.service
- [needs infra] kernel CI matrix shows x86_64 + arm64 + hardened green.

CI gate: signature-verify test + observe-only denylist test required; kernel matrix expanded.
```

---

# Phase 6 — AI / RCA (AIRCA, weight 9; already 85)

## Sprint 20 — MCP/AI egress controls, redaction & audit
**Closes:** AIRCA-001 (M), AIRCA-002 (M), AIRCA-003 (M), AIRCA-005 (L), RED-005 (M) · **Depends on:** Sprint 4 · **Size:** ~4d

```text
You are working in probectl's AI/RCA + MCP layer (internal/ai, the MCP server, analyzer integration). The MCP server can egress raw tenant telemetry to external AI clients without the same controls as the main RCA path; pre-egress redaction is IP-focused and best-effort; MCP read-tool calls aren't audited; the AI test-authoring path bypasses the egress-consent gate; and an uncited root_cause string can pass injection unblocked.

Evidence to start from:
- AIRCA-001: MCP server egresses raw tenant telemetry to external AI clients with none of the main-path controls
- AIRCA-002: pre-egress redaction is IP-focused/best-effort; hostnames + free-text PII pass
- AIRCA-003: MCP read-tool calls not audited
- AIRCA-005: AI test-authoring egress bypasses the per-tenant egress-consent gate
- RED-005: uncited root_cause string passes injection unblocked by citation-integrity (one path)

Tasks:
1. Unify egress controls (AIRCA-001, AIRCA-005): route ALL external-AI egress (RCA, MCP, test-authoring) through ONE gate that enforces the per-tenant egress consent, provider selection (local Ollama default), and redaction. No path may bypass it.
2. Extend the EXISTING redactor (AIRCA-002 — TRIAGED: internal/ai/redact.go:12-30 already masks IPs by default, hostnames per policy, and secrets/bearer values ALWAYS (C8)): add email + free-text-PII patterns and configurable custom patterns to it. Do not rewrite it; add tests with realistic telemetry.
3. Audit MCP reads (AIRCA-003): log every MCP read-tool call (who/tenant/what telemetry/when) to the audit trail.
4. Citation integrity (RED-005): require the root_cause output to carry validated citations; reject/flag uncited claims on every path (close the one bypass).

Acceptance:
- Every external-AI egress goes through the consent+redaction+audit gate; redaction covers hostnames/PII; MCP reads are audited; uncited RCA output is blocked.

Verify:
- go test ./internal/ai/... -run "Egress|Consent|Redact|Audit|Citation" -v
- grep -rn "openai\|anthropic\|http" internal/ai analyzer | grep -iv test ; confirm all behind the gate
- attempt an MCP egress without consent; expect denial + audit entry.

CI gate: egress-gate + redaction tests required; a check that no AI client call exists outside the gate package.
```

## Sprint 21 — RCA resilience + blocking eval gate
**Closes:** AIRCA-004 (L), TEST-003 (M) · ~~AIRCA-006 + DOCS-003 struck — analyzer/ is labeled the BGP analyzer everywhere (pyproject description, CLAUDE.md §3/§5, docs/architecture.md) and its 11 .py modules run in the test-python CI job at 94% coverage~~ · **Depends on:** Sprint 20 · **Size:** ~2d

```text
You are working in probectl's AI/RCA reliability. Add resilience to the remote-model path and make the RCA quality eval blocking.

Evidence to start from:
- AIRCA-004: no caching or circuit-breaker on the remote-model path; a slow provider degrades RCA
- TEST-003: RCA quality eval is non-blocking (continue-on-error: true) — FOUNDER DECISION: flip it now; U-049 said "non-blocking initially" and the baseline exists (accuracy 0.91 / precision 0.92)

Tasks:
1. Resilience (AIRCA-004): wrap the remote-model client with internal/breaker (already in-repo, U-078), timeouts, and a response cache; on provider failure, degrade gracefully (fall back to the builtin air-gapped model with a clearly-marked partial-result banner).
2. Blocking eval (TEST-003): remove continue-on-error from the rca-eval job; floor: answer_accuracy >= 0.85 AND citation_precision >= 0.85 (comfortably below the 0.91/0.92 baseline, catches regressions).

Acceptance:
- Remote-model path has breaker+timeout+cache and degrades to the builtin; rca-eval is a blocking gate with the stated floor.

Verify:
- go test ./internal/ai/... -run "Breaker|Cache|Timeout|Fallback" -v
- grep -n "continue-on-error" .github/workflows/ci.yml ; expect not on rca-eval

CI gate: RCA quality eval is required (no continue-on-error); analyzer pytest runs in CI.
```

---

# Phase 7 — Architecture completeness (ARCH, weight 8)

## Sprint 22 — OTLP traces + logs + OTel Collector path
**Closes:** ARCH-001 (H), ARCH-006 (L), DOCS-005 (M — claim wording lands here, not in S30) · **Depends on:** Sprint 16 · **Size:** ~5d · **Resolves PRD v1.0 §5.1-2 as BUILD (founder decision)**

```text
You are working in probectl's OTLP receiver + OTel integration. Today OTLP traces and logs are absent from the receiver (metrics-only), and there's no upstream OTel Collector integration, limiting ecosystem reuse.

Evidence to start from:
- ARCH-001: OTLP traces and logs absent from the OTLP receiver
- ARCH-006: no upstream OTel Collector integration — custom pipeline limits ecosystem reuse

Tasks:
1. Implement OTLP traces + logs (ARCH-001): add receiver support + storage for traces and logs (not just metrics), wiring to the OTLP consumer from Sprint 16; add round-trip tests for all three signals.
2. OTel Collector path (ARCH-006): document + support deployment behind a standard OTel Collector (provide a collector config / exporter), or expose probectl as a standard OTLP endpoint a Collector can export to; add an example to deploy/ and docs/.
3. Update the "OpenTelemetry-native" claims to match actual signal coverage (coordinate with DOCS-005 in Sprint 30).

Acceptance:
- OTLP metrics, traces, and logs are all received + stored + queryable; a Collector-based deployment is documented + exercised.

Verify:
- go test ./internal/otel/... -run "Traces|Logs|Metrics|OTLP" -v
- [needs infra] send OTLP traces+logs via an OTel Collector; query them back.

CI gate: OTLP three-signal round-trip test required.
```

---

# Phase 8 — Supply chain & build provenance (SUPPLY, weight 5)

## Sprint 23 — Pin remaining inputs + analyzer lockfile
**Closes:** SUPPLY-001 (H), SUPPLY-002 (M), SUPPLY-003 (M), SUPPLY-006 (L) · ~~SUPPLY-004 struck — vimto is pinned (`go install lmb.io/vimto@v0.4.0`, sumdb-verified); SUPPLY-005 struck — auditor knowledge-cutoff artifact (Go 1.26 is the current official toolchain); keep the 10-line provenance doc as a courtesy~~ · **Depends on:** Sprint 1 · **Size:** ~2d

```text
You are working on probectl's build/supply-chain config. Pin all mutable inputs, add lockfiles, and document/justify the forward-dated Go toolchain.

Evidence to start from:
- SUPPLY-001: shipped compose defaults to :latest for probectl-control (deploy/compose/probectl.yml:54,66)
- SUPPLY-002: ruff/black installed in CI without version pins
- SUPPLY-003: Python analyzer has no lockfile (pyproject.toml uses >= floors)
- SUPPLY-004: lmb.io/vimto installed at CI runtime via go install @vX without pin/verify
- SUPPLY-005: forward-dated go 1.26 / toolchain go1.26.4 with stdlib-CVE patch comments — provenance unverifiable
- SUPPLY-006: pyyaml installed without version pin in helm-gate CI step

Tasks:
1. Pin the compose image (SUPPLY-001): default to a digest-pinned image (ghcr.io/...@sha256:...) not :latest; add Dependabot/renovate for docker + actions; document the pinning policy.
2. Pin Python tooling (SUPPLY-002, SUPPLY-006): pin ruff/black/pyyaml to exact versions in CI (or a requirements-dev.txt with hashes).
3. Lockfile for analyzer (SUPPLY-003): generate a locked, hash-pinned dependency set (uv/pip-tools) and add a CI check that the lock matches pyproject.
4. Toolchain provenance doc (SUPPLY-005 courtesy — the finding itself is struck): 10 lines in docs/build/toolchain.md stating the toolchain is the official go1.26.x release, sumdb-verified, pinned via go.mod toolchain directive.

Acceptance:
- No :latest or unpinned tool installs anywhere in CI/compose; analyzer has a hash-locked dependency set; Go toolchain origin is documented + verified.

Verify:
- grep -rn ":latest" deploy ; expect none (except clearly-labeled local dev)
- grep -rn "go install .*@v\|pip install \|uv pip" .github/workflows ; confirm all pinned
- python -m piptools compile --generate-hashes (or uv lock) leaves no diff

CI gate: a `supply-pins` check fails on :latest, unpinned go install, or unpinned pip install.
```

---

# Phase 9 — Test depth + green-build capstone (TEST, weight 4; static-only closure)

## Sprint 24 — Test depth: integration home, coverage floors, e2e on PRs, parallel
**Closes:** TEST-004 (M, re-scoped), TEST-005 (M), TEST-006 (M), TEST-007 (L), TEST-009 (L) · ~~TEST-010 struck — `-race` already runs in `make test` (Makefile:116), the isolation suite (:121), and coverage (:136)~~ · **Depends on:** Sprint 2 · **Size:** ~4d

```text
You are working on probectl's test maturity. Make integration real, gate e2e on PRs, raise coverage floors on sensitive packages, parallelize, and run with the race detector.

Evidence to start from:
- TEST-004: integration test module (test/integration/) is a one-line placeholder
- TEST-005: very low coverage floors for high-sensitivity packages (tenantlife 25%, ee/silo, ...)
- TEST-006: full-stack e2e is nightly-only — not a required PR gate
- TEST-007: zero t.Parallel() across 320 test files — suite entirely serial
- TEST-009: store/pathstore (41.6%) and store/tsdb (49.2%) under-covered
- TEST-010: cover-gate does not use -race

Tasks:
1. Integration home (TEST-004 — TRIAGED: real integration suites EXIST — the store integration suite, test/e2e black-box (U-054, nightly), and compose-backed CI jobs; the 16-line test/integration/smoke_test.go is just not where they live): either build test/integration/ out as the cross-plane ingest→store→query home, or document the actual layout in test/README.md and point the gate at the real suites. Don't duplicate coverage that exists.
2. Raise floors (TEST-005, TEST-009): set + ratchet coverage floors for tenantlife, ee/silo, store/pathstore, store/tsdb (and other tenant/auth-critical packages); add the missing tests.
3. e2e on PRs (TEST-006): promote the full-stack e2e from nightly-only to a required (possibly smaller/faster) PR gate.
4. Parallelize (TEST-007): add t.Parallel() to independent tests; ensure no shared-state flakes.
Acceptance:
- Integration layout documented or consolidated; sensitive-package coverage floors enforced; e2e gates PRs; tests parallelized where hot.

Verify:
- go test ./test/integration/... -tags=integration -v   [needs infra]
- go test ./internal/tenantlife/... ./internal/store/... -cover ; confirm floors

CI gate: integration + e2e (PR profile) + `-race` cover-gate are required checks; floors enforced.
```

## Sprint 25 — Green-build capstone: verify-all umbrella + blocking scanners  `[static-only closure]`
**Closes:** verifies/locks-in Sprints 14–19 (no new IDs) · **Depends on:** Sprints 14–24 · **Size:** ~3d

```text
You are closing the diligence "STATIC-ONLY" methodology gap. TRIAGED: most of these already run as standing required CI gates (build/vet/lint, -race tests, kernel-matrix eBPF load, govulncheck+trivy in dependency-scan + the scheduled security-scan). The real content is the umbrella target + making the scanners explicitly blocking + archived receipts.

Tasks:
1. Prove the build is green end-to-end: go build ./... ; go vet ./... ; golangci-lint run — fix anything that surfaces.
2. Prove tests pass with the race detector: go test -race ./... (and the integration/e2e profiles from Sprint 24).
3. Run vulnerability + supply scans for real and gate on them: govulncheck ./... ; trivy fs . ; ensure CI BLOCKS on findings (not advisory).
4. Compile + load the eBPF programs (closing the audit's clang gap): clang -target bpf compile bpf/*.bpf.c, then load/attach under the CI kernel matrix from Sprint 19; assert the observe-only denylist + signature verification hold at load time.
5. Capture all of the above outputs as committed/retained CI artifacts (extends DATAROOM-005/-011) so future diligence has executed receipts, not static-only caveats.

Acceptance:
- A single `make verify-all` (build, vet, lint, test -race, govulncheck, trivy, eBPF compile+load) passes locally and in CI and is a required gate; receipts are archived.

Verify:
- make verify-all
- govulncheck ./... ; echo "exit=$?"  (must be 0 to merge)
- [needs clang+kernel] eBPF compile+load job green in CI.

CI gate: `verify-all` is the umbrella required check; govulncheck/trivy are blocking (no continue-on-error).
```

---

# Phase 10 — Deployment & fleet ops (OPS, weight 3)

## Sprint 26 — k8s/Helm hardening: probes, strict NetPolicy profile, /metrics, backup jobs
**Closes:** OPS-001 (H), OPS-004 (M — strict profile; U-086 default stands), OPS-005 (M), OPS-009 (L) · ~~OPS-006 struck — CH migrations run at boot via the U-046 versioned ledger (flowstore/clickhouse.go:79-108 + internal/store/chmigrate); an init-container would duplicate the boot path~~ · **Depends on:** Sprint 1 · **Size:** ~3d

```text
You are working in probectl's Helm charts + k8s manifests (deploy/). Close day-2 ops gaps.

Evidence to start from:
- OPS-001: agent DaemonSet has no liveness/readiness probes
- OPS-004: NetworkPolicy ships allow-all egress + any-pod-to-API ingress
- OPS-005: control plane has no Prometheus /metrics scrape endpoint
- OPS-006: ClickHouse migrations have no Helm init-container
- OPS-009: ClickHouse backup CronJobs not in the Helm chart

Tasks:
1. Probes (OPS-001): add liveness + readiness probes to the agent DaemonSet (and control Deployment) wired to real health endpoints; readiness reflects enrollment + datastore connectivity.
2. NetworkPolicy strict profile (OPS-004 — FOUNDER DECISION: the U-086 default stands — default-on with two DOCUMENTED holes, because default-deny ingressFrom locks out unknown ingress controllers): add deploy/helm/probectl/values-strict.yaml with full default-deny (named ingress-controller selector REQUIRED, datastore/scrape egress allow-list, no holes) and reference it from docs/hardening.md as the regulated-profile recommendation. The default profile keeps its documented holes.
3. /metrics (OPS-005): expose a Prometheus /metrics endpoint on the control plane (+ ServiceMonitor); cover the self-metrics the audit said require none today.
4. Backup CronJobs (OPS-009 — already a tracked follow-up from U-030, deploy/backup/README.md): fold the PG + CH backup CronJobs into the chart behind `backup.enabled`.

Acceptance:
- Agent + control have probes; values-strict.yaml renders full default-deny; /metrics scrapeable; backups are chart-managed behind backup.enabled.

Verify:
- helm template deploy/helm/probectl | grep -E "livenessProbe|readinessProbe|NetworkPolicy|/metrics|initContainers|CronJob"
- helm lint deploy/helm/*
- [needs infra] kubeconform/kube-linter on rendered manifests; deploy to kind and curl /metrics.

CI gate: `helm-lint` + `kubeconform` over rendered charts required; a test asserting probes + NetworkPolicy presence.
```

## Sprint 27 — Backup encryption + DR drill + air-gap IdP
**Closes:** OPS-002 (H), OPS-007 (M), OPS-008 (L), COMPLY-002 (H), DATAROOM-008 (M), DOCS-007 (M) · **Depends on:** Sprint 8, Sprint 26 · **Size:** ~5d

```text
You are working on probectl's backup/DR + air-gap story. Encrypt backups, validate DR with a real drill, support self-hosted IdP, and make erasure cover backups.

Evidence to start from:
- OPS-002: backup artifacts are unencrypted — tenant telemetry written to disk in plaintext
- OPS-007 / DATAROOM-008: RTO/RPO targets PROVISIONAL — no representative-infra drill
- OPS-008: air-gap deployment requires external OIDC IdP — no self-hosted path documented
- COMPLY-002: erasure attestation does not cover backups/snapshots
- DOCS-007: multi-region HA implemented but RTO/RPO unvalidated, runbooks incomplete

Tasks:
1. Encrypt backups (OPS-002): encrypt backup artifacts at rest (reuse the at-rest key mgmt from Sprint 8); never write tenant telemetry to disk in plaintext. Test restore from an encrypted backup.
2. Erasure covers backups (COMPLY-002): implement (or document with enforcement) tenant erasure propagation to backups/snapshots within a bounded window; provide an attestation that includes backups. Add a verifiable test/report.
3. DR drill (OPS-007, DATAROOM-008, DOCS-007): run a representative multi-region failover drill [needs infra]; record real RTO/RPO; fill the PROVISIONAL tables; complete the runbooks in docs/runbooks/.
4. Self-hosted IdP (OPS-008): document + support an air-gapped, self-hosted OIDC IdP path (e.g. Dex/Keycloak) so air-gap deployments don't need an external IdP.

Acceptance:
- Backups encrypted + restore-tested; erasure provably covers backups; real RTO/RPO recorded; air-gap IdP path documented + exercised.

Verify:
- go test ./internal/... -run "Backup|Restore|Erasure|Retention" -v
- grep -rn "PROVISIONAL" docs | grep -iE "rto|rpo|dr" ; expect none
- [needs infra] restore from encrypted backup; run failover drill.

CI gate: backup encrypt/restore + erasure-covers-backups tests required.
```

---

# Phase 11 — Compliance surfaces & data room (non-licensing) (COMPLY/DATAROOM)

## Sprint 28 — Data residency enforcement + audit-trail coverage + PII-min doc
**Closes:** COMPLY-003 (H), COMPLY-005 (L), COMPLY-006 (L), DATAROOM-009 (M) · **Depends on:** Sprint 5 · **Size:** ~4d

```text
You are working on probectl's compliance controls. Make data residency enforced (not just a label), close audit-trail gaps, and publish the data-handling/PII-minimization documentation.

Evidence to start from:
- COMPLY-003: pooled-tenant data residency has no enforcement; PROBECTL_RESIDENCY is a reporting label only; docs/residency.md absent
- COMPLY-005: SCIM provisioning + break-glass actions not verified to be in the audit trail
- COMPLY-006: no outward-facing compliance/residency policy page
- DATAROOM-009: no consolidated telemetry data-handling / PII-minimization document

Tasks:
1. Enforce residency (COMPLY-003 — scoped per design): SILOED/HYBRID tenants get real enforcement — their data planes are per-tenant, so route storage + AI egress by region and reject cross-region writes for region-pinned tenants. POOLED tenants keep the honest U-042 posture: they inherit the deployment's region — surface that inherited region explicitly in the API/UI and document that per-row residency in a pooled store is not a thing (don't pretend). Write docs/residency.md covering both.
2. Audit-trail coverage test (COMPLY-005 — TRIAGED: SCIM provisioning IS audited (internal/control/scim.go:20 auditSCIM) and break-glass has its own separately-chained stream (U-012); the residual is a TEST): add coverage tests asserting both event families land in their chains; fix anything the test finds.
3. PII-minimization doc (DATAROOM-009, COMPLY-006): write docs/compliance/data-handling.md (what telemetry is collected, PII classification incl. IPs-as-PII, minimization/redaction, retention) and an outward-facing compliance/residency policy page.

Acceptance:
- Residency is enforced for region-pinned tenants + documented; SCIM + break-glass are audited (tested); data-handling/PII doc + policy page exist.

Verify:
- go test ./internal/compliance/... ./internal/audit/... ./internal/scim/... -run "Residency|Audit|BreakGlass|SCIM" -v
- ls docs/residency.md docs/compliance/data-handling.md
- region-pin a tenant; attempt a cross-region write; expect rejection.

CI gate: residency-enforcement + audit-coverage tests required.
```

## Sprint 29 — Data-room package + external-engagement kickoff
**Closes:** DATAROOM-004 (H), DATAROOM-006 (M), DATAROOM-007 (M), DATAROOM-010 (M), DATAROOM-012 (L), DATAROOM-013 (L), COMPLY-001 (C, evidence portion) · **Depends on:** Sprint 25, Sprint 28 · **Size:** ~4d (code/doc) + external calendar time

```text
You are assembling probectl's acquisition data room and kicking off the external engagements that can't be done in code. Produce every committed artifact a buyer's GRC/security team expects; flag the external deliverables.

Evidence to start from:
- DATAROOM-004: no external pen test / independent security assessment
- DATAROOM-006 / COMPLY-001: SOC 2 mapping is a self-declared draft skeleton — no auditor engagement (this is the 4th Critical; its closure needs an external auditor)
- DATAROOM-007: DPA + subprocessor list are unfinalized drafts
- DATAROOM-010: no standalone known-risks register
- DATAROOM-012: no customer/deployment references
- DATAROOM-013: ADR coverage sparse (2 ADRs for 115k LOC)

Tasks (code/doc-actionable now):
1. Risk register (DATAROOM-010): create docs/diligence/known-risks.md consolidating residual risks (seed it from this remediation plan + audit reports) with owners + status.
2. ADRs (DATAROOM-013): backfill ADRs for the major decisions touched in Phases 1–10 (tenant binding, enrollment/SVID, ClickHouse driver, scale architecture, eBPF scoping, StreamConfig removal, persistence). Put them in docs/adr/.
3. DPA + subprocessor list (DATAROOM-007): finalize docs/legal/DPA.md + docs/legal/subprocessors.md (note: legal review required to execute).
4. SOC 2 evidence package (DATAROOM-006/COMPLY-001): turn the skeleton into a real control-to-evidence mapping (cite the code/tests/CI now implemented) ready to hand an auditor.

Tasks (external — kick off now, track to completion):
5. Engage a third-party pen test (DATAROOM-004) and a SOC 2 readiness/Type I→II auditor (COMPLY-001). These have calendar lead time — start them at the BEGINNING of the remediation program in parallel (see Parallel Tracks), not here.
6. Customer references (DATAROOM-012): commercial task — collect deployment references.

Acceptance:
- known-risks register, backfilled ADRs, finalized DPA/subprocessor drafts, and an auditor-ready SOC 2 evidence mapping are committed. Pen test + SOC 2 audit + references are in flight with owners + dates.

Verify:
- ls docs/diligence/known-risks.md docs/adr/ docs/legal/DPA.md docs/legal/subprocessors.md
- review the SOC 2 control-to-evidence mapping references real, merged controls.

CI gate: a docs-presence check for the data-room artifacts (non-blocking informational gate).
Note: COMPLY-001's Critical does not fully clear until the external SOC 2 attestation is received.
```

---

# Phase 12 — Docs truthfulness & optional commercial

## Sprint 30 — Sovereignty-caveats doc sweep
**Closes:** DOCS-008 (L) · ~~DOCS-004 struck — README:140 already states CMVP cert #5247 is the MODULE's and "probectl itself holds no product-level certificate" (the U-019/B10 honest-claims pass); DOCS-005 folded into Sprint 22~~ · **Depends on:** Sprint 17, Sprint 22 · **Size:** ~1d

```text
You are documenting precisely which optional probectl features egress data and thus affect the sovereignty claim.

Evidence to start from:
- DOCS-008: sovereignty claim accurate for default config; optional threat-intel/outage feeds + remote AI (and MCP external clients, post-S20) egress data

Tasks:
1. One page (docs/sovereignty-caveats.md, linked from README): the default config is fully sovereign (no phone-home, air-gapped AI); the EXPLICIT opt-ins that egress are: open-data/threat-intel/outage feed fetches (read-only, outbound), remote AI model endpoints (consent-gated, redacted — U-013/C8), MCP serving external AI clients (consent-gated post-S20), and SIEM/ITSM/webhook integrations the operator configures. State what leaves in each case and where the gate/audit lives.
2. Cross-link from docs/hardening.md and the air-gap install docs.

Acceptance:
- A buyer can answer "what leaves the network and when" from one page; every egress names its gate.

Verify:
- ls docs/sovereignty-caveats.md && grep -rn "sovereignty-caveats" README.md docs/hardening.md

CI gate: a docs-claims lint (grep for known overclaim phrases) flags for review.
```

## Sprint 31 — (Optional) Billing scaffold cleanup + export round-trip
**Closes:** CODE-005 (M) · **Depends on:** Sprint 2 · **Size:** ~6d · **Optional / commercial**

```text
You are cleaning up the billing scaffold. TRIAGED: commercial metering EXISTS — internal/usage is the per-tenant metering seam (S-T3: agents/tests/ingest/AI-calls) and ee/billing does the usage export. The finding's "0 functions" refers to internal/billing, a 6-line dead doc.go scaffold the metering never ended up living in.

Evidence to start from:
- CODE-005: internal/billing is a dead scaffold; the real metering is internal/usage + ee/billing

Tasks:
1. Delete internal/billing (or repoint it as the documented home of the usage seam if the package name is worth keeping — pick one, note it in the commit).
2. Add an invoice-export round-trip test: meter usage for a tenant → export via ee/billing → assert the export matches the metered values (accuracy + tenant attribution).

Acceptance:
- No dead scaffold; metering accuracy covered by a round-trip test.

Verify:
- go test ./internal/usage/... ./ee/billing/... -run "Meter|Usage|Billing|Export" -v

CI gate: billing tests required (once implemented).
```

---

# Phase L — Licensing, IP & legal readiness  (RUN LAST)

> Per your instruction, all licensing work is last. Merging Sprints 32–33 clears the final **Legal/IP Critical** and releases the score cap (the compliance Critical releases when the external SOC 2 attestation lands). These are mostly engineering-mechanical but **gated on legal/founder decisions** — start the *decisions* early (see Parallel Tracks) even though the *commits* land here.

## Sprint 32 — Commit OSS + EE license texts + SPDX headers  `[CRITICAL]`
**Closes:** LICENSE-001 (C), LICENSE-002 (C), LICENSE-003 (H), DATAROOM-001 (C), DOCS-002 (H) · **Depends on:** legal decision (license choice) · **Size:** ~3d · **Clears:** half of CRIT-LEGAL-IP

```text
You are finalizing probectl's licensing. The root LICENSE file contains only "TBD", every ee/ file carries a commercial-license PLACEHOLDER, and ~819 Go files have no SPDX/copyright headers.

Evidence to start from:
- LICENSE-001 / DATAROOM-001 / DOCS-002: root LICENSE = literal "TBD"
- LICENSE-002: all ee/ files carry a commercial license placeholder (no binding EE license)
- LICENSE-003: zero SPDX-License-Identifier/copyright headers on ~819 Go files

PREREQUISITE (human, do first): founder + counsel choose (a) the OSS license for the open core (e.g. Apache-2.0 / AGPL-3.0 / BSL-1.1) and (b) the EE commercial license text. This sprint COMMITS that decision; it does not make it.

Tasks:
1. Replace root LICENSE with the chosen full OSS license text. Add a LICENSE-EE (or ee/LICENSE) with the finalized commercial terms; replace the placeholder in every ee/ file.
2. Add SPDX-License-Identifier + copyright headers to all source files (Go, TS, Python, C, proto) via a scripted header tool; OSS files get the OSS identifier, ee/ files get the commercial identifier. Add a CI check (e.g. addlicense/reuse) that fails on missing/incorrect headers.
3. Update README + docs to state the licensing model accurately (open core + EE); remove the source-available "not OSS yet" badge once real.

Acceptance:
- Valid OSS license text at root; binding EE license on all ee/ files; every source file has a correct SPDX header; CI enforces headers; README licensing is accurate.

Verify:
- head -5 LICENSE ; grep -c "TBD" LICENSE ; expect 0
- reuse lint   (or addlicense -check ./...) ; expect pass
- grep -rL "SPDX-License-Identifier" $(git ls-files '*.go') ; expect empty

CI gate: `license-headers` (reuse/addlicense) is a required check.
```

## Sprint 33 — IP chain of title: CLA/DCO + contributor assignments  `[CRITICAL]`
**Closes:** LICENSE-004 (H), DATAROOM-002 (C) · **Depends on:** Sprint 32 · **Size:** ~2d · **Clears:** rest of CRIT-LEGAL-IP

```text
You are establishing probectl's IP chain of title. There is no CLA/DCO, and contributor IP ownership for the 8 non-founder commits is unverified.

Evidence to start from:
- DATAROOM-002: no CLA, DCO, or IP-assignment — IP chain unproven for all contributors
- LICENSE-004: no CLA/DCO; 8 non-founder commits' IP ownership unverified

Tasks:
1. Adopt a DCO (Developer Certificate of Origin) and/or CLA: add CONTRIBUTING.md DCO sign-off requirement + a DCO bot/CI check requiring Signed-off-by on every commit going forward.
2. Retroactively secure assignments for the existing non-founder contributors (netctl, imfeelingtheagi, and any others): obtain signed CLAs/assignment for their commits; record in docs/legal/ip-assignments/ (note: human/legal task — produce the templates + tracking, collect signatures).
3. Add a CI gate that future PRs must be DCO-signed.

Acceptance:
- DCO/CLA enforced on new commits; signed assignments collected for all prior non-founder contributors; chain of title documented.

Verify:
- grep -n "Signed-off-by" .github/ CONTRIBUTING.md ; DCO check present
- ls docs/legal/ip-assignments/ ; one record per non-founder contributor
- git log --format='%an' | sort -u ; cross-check each has an assignment.

CI gate: `dco-check` required on all PRs.
```

## Sprint 34 — Third-party license inventory + NOTICE + eBPF GPL boundary
**Closes:** DATAROOM-003 (H), LICENSE-005 (M), LICENSE-006 (L) · **Depends on:** Sprint 32 · **Size:** ~3d

```text
You are producing probectl's third-party license inventory and resolving specific dependency-license risks, including the eBPF GPL boundary.

Evidence to start from:
- DATAROOM-003: no committed third-party license inventory — GPL contamination risk unassessed
- LICENSE-005: MaxMind GeoLite2 CC BY-SA commercial-use terms must be resolved before provider/MSP redistribution
- LICENSE-006: vendored gNMI Apache-2.0 code lacks a NOTICE file

Tasks:
1. Generate a committed dependency-license inventory for Go (go-licenses), npm (license-checker), and Python (pip-licenses); commit to docs/legal/third-party-licenses.md + a machine-readable SBOM-linked list. Add a CI gate (go-licenses check) that fails on disallowed licenses (e.g. unexpected GPL/AGPL in the OSS-distributed binary).
2. eBPF GPL boundary: document the boundary — the kernel-side BPF programs (bpf/*.bpf.c, GPL-licensed for GPL-only helpers) vs the userspace loader (your chosen license). Confirm no GPL symbol leaks into userspace linkage; write docs/legal/ebpf-gpl-boundary.md with the analysis + the SEC("license") declarations.
3. MaxMind GeoLite2 (LICENSE-005): resolve CC BY-SA commercial/redistribution terms — either switch to a redistributable data source, require operators to supply their own license/key, or obtain a commercial MaxMind license; document the decision.
4. Add the NOTICE file for vendored gNMI (LICENSE-006) and any other vendored third-party code — including the vendored libbpf BPF headers (`internal/ebpf/bpf/headers/`, LGPL-2.1 OR BSD-2-Clause; provenance in their `VENDOR.md`); centralize attributions. NOTE: `gen_third_party.sh`/`NOTICE` only enumerate Go modules, so vendored non-Go source (gNMI, libbpf headers) must be added separately.

Acceptance:
- Committed third-party license inventory + CI license gate; documented eBPF GPL boundary with no contamination; MaxMind terms resolved; NOTICE file present.

Verify:
- go-licenses report ./... > /tmp/lic.txt ; grep -iE "GPL|AGPL" /tmp/lic.txt ; review each
- ls NOTICE docs/legal/third-party-licenses.md docs/legal/ebpf-gpl-boundary.md
- grep -rn "SEC(\"license\")" internal/ebpf/bpf ; confirm declarations match the boundary doc.

CI gate: `license-inventory` (go-licenses/pip-licenses allowlist) is a required check.
```

## Sprint 35 — Trademark strategy & registration
**Closes:** LICENSE-007 (L) · **Depends on:** Sprint 32 · **Size:** ~2d (mostly legal) · **External/legal**

```text
You are addressing probectl's trademark exposure. The "probectl" name + the editions fence rely on trademark, but there's no registration/strategy.

Evidence to start from:
- LICENSE-007: probectl name + editions fence rely on trademark; no registration or strategy

Tasks (mostly legal/founder, minimal code):
1. File for trademark registration on the "probectl" name + logo in the relevant jurisdictions (counsel-led).
2. Add a TRADEMARK.md trademark-usage policy (how the community + downstreams may use the name; how the OSS/EE editions fence interacts with the mark).
3. Document the strategy in docs/legal/trademark.md.

Acceptance:
- Trademark application filed; TRADEMARK.md usage policy committed; strategy documented.

Verify:
- ls TRADEMARK.md docs/legal/trademark.md
- (legal) registration receipt tracked in docs/legal/.

CI gate: docs-presence check (informational).
```

---

# Parallel tracks (start at program kickoff — long calendar lead times)

These are **not Claude Code sprints** but gate the headline score; start them on **day 1** even though their commits/closure land later:

| Track | Why early | Gates | Lands |
|---|---|---|---|
| **SOC 2 readiness → Type I → Type II** | 6–12+ mo auditor calendar; clears COMPLY-001 (the 4th Critical / cap) | B6 / CRIT-COMPLIANCE | post-program |
| **External penetration test** | scheduling + retest cycle; DATAROOM-004 | data-room / buyer security review | mid-program |
| **License + EE terms decision (counsel)** | unblocks Sprints 32–34 | B1 / CRIT-LEGAL-IP | before Phase L |
| **Contributor IP assignments (counsel)** | collecting signatures takes time; unblocks Sprint 33 | B2 / CRIT-LEGAL-IP | before/at Phase L |
| **Trademark filing (counsel)** | filing + examination; Sprint 35 | LICENSE-007 | post-filing |

---

# Appendix A — Finding → Sprint coverage matrix

Every actionable finding (Critical/High/Medium/Low) maps to exactly one sprint. Info-severity findings are positive/observational (no remediation). External-only items are flagged.

| Audit | Findings → Sprint |
|---|---|
| TENANT | 101→S4 · 102→S5 (re-scoped) · 103→S11 · 104→S5 (re-scoped) · 105→S6 (re-scoped) · 106→S5 · 107→S4 · 108→S7 |
| RED | 001→S3 · 002→S11 · 003→S18 · 004→~~STRUCK (fixed 93529b0)~~ · 005→S20 · 006→S4 · 007→S9 (verify) · 008→S12 · 009→~~STRUCK (LogValue allowlist)~~ |
| WIRE | 001→S4 · 002→S11 · 003→S12 · 004→S12 · 005→S12 · 006→S12 · 007→S12 |
| SEC | 001→S3 · 002→S8 · 003→S9 · 004→S9 · 005→S7 · 006→S10 · 007→S10 · 008→S10 · 009→S9 |
| SCALE | 001→S14 · 002→S17 (run; harness exists) · 003→S15 (re-scoped) · 004→S15 · 005→S15 (verify) · 006→S16 (path only) · 007→S15 · 008→~~STRUCK (D2 DLQ; verify-test in S14)~~ · 009→S14 (re-scoped) · 010→S16 · 011→S15 · 012→S14 · 013→S14 · 014→S16 · 015→S17 |
| EBPF | 001→S18 · 002→S18 · 003→S19 (re-scoped) · 004→S19 (re-scoped) · 005→S19 · 006→**S0** (done immediately) · 007→S19 (verify) · 008→S19 |
| AIRCA | 001→S20 · 002→S20 (extend existing redactor) · 003→S20 · 004→S21 · 005→S20 · 006→~~STRUCK (correctly labeled)~~ |
| ARCH | 001→S22 · 002→S7 · 003→S13 (harden, not remove — U-044 ADR) · 004→S11 · 005→S16 (silences only — U-047 ADR) · 006→S22 |
| SUPPLY | 001→S23 · 002→S23 · 003→S23 · 004→~~STRUCK (pinned @v0.4.0)~~ · 005→~~STRUCK (cutoff artifact; courtesy doc in S23)~~ · 006→S23 · 007→S1 |
| TEST | 001→S2 · 002→S1 · 003→S21 (blocking now) · 004→S24 (re-scoped) · 005→S24 · 006→S24 · 007→S24 · 008→S1 · 009→S24 · 010→~~STRUCK (-race in Makefile)~~ |
| CODE | 001→S2 · 002→S2 (sites cited) · 003→S1 (gitignore-assert only) · 004→S2 · 005→S31 (re-scoped — metering exists) · 006→S1 · 007→S1 |
| OPS | 001→S26 · 002→S27 · 003→S10 · 004→S26 (strict profile) · 005→S26 · 006→~~STRUCK (boot migrations, U-046)~~ · 007→S27 (iron) · 008→S27 · 009→S26 · 010→S10 |
| DOCS | 001→S17 · 002→S32 · 003→~~STRUCK (locatable + tested)~~ · 004→~~STRUCK (README:140 honest)~~ · 005→S22 · 006→S17 · 007→S27 · 008→S30 |
| COMPLY | 001→S29 (+ external SOC 2) · 002→S27 · 003→S28 · 004→S8 · 005→S28 · 006→S28 |
| LICENSE | 001→S32 · 002→S32 · 003→S32 · 004→S33 · 005→S34 · 006→S34 · 007→S35 |
| DATAROOM | 001→S32 · 002→S33 · 003→S34 · 004→S29 (external) · 005→S1 · 006→S29 (external) · 007→S29 · 008→S27 · 009→S28 · 010→S29 · 011→S1 · 012→S29 (commercial) · 013→S29 |

# Appendix B — Suggested execution cadence
- **Sprint 0 first** (triage record + EBPF-006 one-liner), then numeric order.
- **Solo / Claude Code, one sprint at a time:** follow the numeric order. Stop after S6 to confirm 2 Criticals cleared and red-team retest is clean before investing in scale. S11 requires its design ADR + threat-model delta REVIEWED before code (founder gate).
- **Small team (2–3):** run by phase. After Phase 0, parallelize: one engineer on Phase 1 (Criticals), one on Phase 2 (SEC). Don't start Phase 3 (enrollment) before S4 lands. Phase 4 (scale) can run alongside Phases 5–8 once S4/S5 are in.
- **Kick off all Parallel Tracks on day 1.**
- **Phase L only after** the non-licensing engineering work is merged (your stated preference) — it's the final cap-release.

# Appendix C — Per-sprint Definition of Done (apply to every sprint)
1. All listed finding IDs implemented + closed. 2. Verify commands pass locally. 3. New CI gate added and green (and required). 4. Tests added/updated; coverage floor respected. 5. `CHANGELOG.md` entry references the closed IDs. 6. Docs updated where the sprint touched behavior/claims. 7. PR is DCO/signed (after S2/S33) and reviewed per CODEOWNERS.

