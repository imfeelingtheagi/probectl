# Pooled-tenant residency & at-rest encryption: verification (U-042)

**Task:** A2 (unified diligence register). **Finding verified:** U-042 [Medium, theirs-adopted,
"not independently verified in code by coordinator"] — *"Pooled tenants have no data-residency
confinement; at-rest encryption delegated to operator"* (T:R15 → T:COMPLY-006, T:COMPLY-004).
**Scope:** investigation only; no behavior changes.

## Verdict

**U-042 is verified — both arms of the claim are accurate**, with one important nuance the
register should record: the codebase does not merely *lack* pooled residency, it **structurally
refuses** it (a residency value on a pooled tenant is a provisioning error), and the silo-only
scope of residency is already honestly documented in `docs/isolation.md`. The gaps are (a) no
outward-facing residency policy statement, (b) no operator guidance for keeping a pooled
deployment region-bound (including the multi-region/HA caveat), and (c) the at-rest duty for
bulk telemetry stores is delegated to the operator **without any doc saying so**.

Register update: U-042 basis "theirs-adopted (pending verification)" → **verified**; severity
Medium stands.

## What residency confinement exists today

### Vocabulary (core, all editions)

- `tenants.isolation_model` (`pooled` default; `pooled|siloed|hybrid` CHECK) and
  `tenants.residency` (default `''`) — `migrations/0025_tenant_isolation.sql:12–24`. The
  migration comment scopes intent: residency is an "operator-facing residency/data-plane name";
  silo mechanics live in `ee/silo` behind the `siloed_isolation` feature.

### Enforcement (siloed/hybrid only, ClickHouse only)

- A residency name maps to a **data plane** — a named ClickHouse endpoint:
  `ee/silo/silo.go:43–51` (`DataPlane`), parsed from `PROBECTL_DATAPLANES`
  (`silo.go:53–79`; `internal/config/config.go:296–300`).
- Provisioning validates the name (`ee/silo/provisioner.go:47–49` `ValidResidency`) and creates
  the tenant's ClickHouse database **on** that plane; reads/writes route there
  (`ee/silo/router.go`, `docs/isolation.md:43`).
- **Pooled tenants are refused residency outright** — `ee/provider/service.go:360–362`:
  `"provider: residency targeting requires a siloed or hybrid tenant"`. The provider console
  hides the field for pooled (`web/src/test/provider-console.test.tsx:204–207` pins this).
- Even for silos, the pin covers **only ClickHouse flow data**. Explicitly NOT pinned:
  Postgres control/config state, the TSDB, the object store, bus brokers —
  `ee/silo/silo.go:45–47` and `docs/isolation.md:53–66` ("residency claims must be real …
  **Do not sell residency beyond this list**").

### Pooled tenants: disclosure only, no confinement

- `PROBECTL_RESIDENCY` (`internal/config/config.go:40,385`) is a deployment-level
  "default data-residency region (governance)" label. Its only consumers are reporting
  surfaces: the cluster topology view (`internal/cluster/cluster.go:62`,
  `cmd/probectl-control/main.go:447`), the redacted config dump (`config.go:641`), the
  composed governance view (`internal/govern/govern.go:123–125` — "recorded for the unified
  view only"; `ee/provider/governance.go:75`), and the tenant lifecycle endpoint
  (`internal/control/lifecycleapi.go:97–101`, empty → omitted for pooled). **Nothing enforces
  it.** `migrations/0033_governance.sql:3–6` states the design: "enforcement of those stays
  with their owners" — and for pooled residency there is no owner.
- A pooled tenant's data therefore lives wherever the deployment's shared stores live: pooled
  Postgres rows under RLS, shared ClickHouse tables (`tenant_id` in partition/order key),
  deployment TSDB (tenant-labeled series), shared object store under a tenant prefix
  (`internal/objectstore/objectstore.go:5–8`).
- **Multi-region/HA caveat (S-EE2):** `internal/cluster` is region-aware *replication*, not
  residency — an active-active deployment replicates the shared Postgres state across regions
  by design (`internal/cluster/cluster.go:1–9`). For pooled tenants, enabling multi-region HA
  actively moves their control/config data across regions. A pooled "residency" stance is only
  as true as the deployment being single-region — currently stated nowhere.

## Where at-rest encryption is enforced vs assumed

| Data | At-rest encryption | Where |
|---|---|---|
| Sensitive tenant-owned **values** (alert-channel secrets, integration tokens; "set grows by consumer") | **Enforced in app** when configured: envelope sealing via the one core seam | `internal/tenantcrypto` (self-describing `dv1:` deployment KEK / `tk1:` per-tenant); KEK = `PROBECTL_ENVELOPE_KEY` (`docs/configuration.md:40`, `deploy/compose/probectl.yml:66`, `deploy/compose/.env.example:9`); per-tenant keys/BYOK = `ee/tenantkeys` (`byok` feature) with a no-fallback fail-safe (`ee/tenantkeys/tenantkeys.go:22–25`, `docs/byok.md`) |
| Secret references | resolved at use, cached encrypted, 5-minute lease | `internal/config/resolve.go` (S41), `docs/configuration.md:1200` |
| **Bulk telemetry** — ClickHouse flows, Postgres rows, TSDB series, object-store blobs | **Not encrypted by probectl; assumed from the operator's storage layer** | No SSE/TDE/KMS code path exists in any store client (`internal/store/flowstore/`, `internal/store/tsdb/`, `internal/objectstore/` — filesystem + memory backends only, S3/MinIO "slots in" later, `objectstore.go:1–8`); transit TLS is supported, rest is not addressed |
| Documentation of that operator duty | **Missing** | `docs/hardening.md:137–145` covers crypto/secrets (envelope KEK, BYOK) only; no doc (hardening, install, isolation, provider-plane) instructs operators to encrypt the CH/PG/TSDB/object volumes (LUKS/EBS/StorageClass/SSE) or ties it to compliance claims |

So: "at-rest encryption delegated to operator" is **true** for everything except the
app-sealed sensitive values — and the delegation is currently implicit.

## Remediation options

**Option B+C below is the recommended pair** (both Small; together they meet the U-042
acceptance — "Residency policy doc + region-pinning for pooled mode, **or** documented
silo-only residency guarantee"). Option A is the roadmap item if MSP demand requires pooled
residency.

### A. Region-pinned pooled cells (engineering; M–L)

Extend the data-plane concept from per-silo ClickHouse to **pooled cells**: deploy one pooled
stack (PG + CH + TSDB + object + bus) per region, register cells like data planes
(`PROBECTL_DATAPLANES` pattern), assign pooled tenants to a cell at provisioning (reuse
`tenants.residency` + `ValidResidency`), and route per tenant in the store layer (the
`ee/silo/router.go` seam generalizes). Real residency for pooled mode, priced as Provider-tier;
significant store-routing and migration work; interacts with S-EE2 (replication must become
cell-local).

### B. Documented silo-only residency guarantee (docs/product; S) — recommended now

Promote what the code already enforces into the **contract**: "Residency requires
siloed/hybrid isolation; pooled tenants inherit the deployment's location." Concretely: a
`docs/residency.md` policy page (sales/compliance-facing) built from `docs/isolation.md:53–66`'s
pinned/not-pinned list incl. the ClickHouse-only scope and the S-EE2 roadmap; provider-console
copy stating pooled = deployment region; keep the `service.go:360–362` refusal as the tested
invariant (it is the guarantee).

### C. Documented operator duty for at-rest + pooled-deployment residency (docs; S) — recommended now

Add a "Data residency & at-rest encryption" section to `docs/hardening.md` (checklist) and
`docs/install.md`: (1) bulk-store at-rest encryption is the operator's layer — volume/SSE/TDE
guidance per store (LUKS/dm-crypt, encrypted EBS/StorageClass in Helm values, MinIO/S3 SSE
when that backend lands), with `PROBECTL_ENVELOPE_KEY` + BYOK clarified as covering sensitive
*values*, not bulk telemetry; (2) a pooled deployment claiming a residency region must be
deployed single-region — multi-region HA (S-EE2) replicates shared state across regions; (3)
set `PROBECTL_RESIDENCY` so the governance view discloses the deployment region to tenants.

## Follow-ups (out of scope here)

Register: mark U-042 verified (this doc); spawn the Option B/C doc tasks (the actual policy
pages are deliverables beyond this investigation). If Option A is ever scheduled, fold in the
silo bus/object/TSDB pinning gaps (`docs/isolation.md:61–66`) so "residency" converges to one
meaning across modes.
