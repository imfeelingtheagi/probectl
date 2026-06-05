# Runbook: tenant offboarding (export → erase → attest)

The clean-offboarding procedure (S-T5). Export and verifiable deletion are
**core** — a compliance right in every edition. The provider-console trigger
is the MSP convenience layer on the same engine.

## 0. Before you start

- Confirm the request's authority (tenant admin or contract owner).
- Know your **backup TTL**: live-store deletion is attested; snapshots
  expire per YOUR backup policy. Set `PROBECTL_BACKUP_RETENTION_NOTE`
  (e.g. "nightly snapshots, 14-day TTL, region X") so every attestation
  states it verbatim. **An attestation without a backup story is
  incomplete in spirit — be explicit.**

## 1. Export (portability)

Tenant self-service: **Admin → Data lifecycle → Export my data**, or
`GET /v1/lifecycle/export` (permission `lifecycle.export`). The bundle is a
tar.gz: `manifest.json` (counts + object inventory + format notes),
`postgres/<table>.jsonl` (every tenant-owned row), `flows.jsonl`. TSDB
series export rides the Prometheus-compatible API (federation/PromQL) — the
manifest says so. Hand the bundle to the customer BEFORE erasing.

## 2. Suspend, then offboard (provider console)

Suspend stops the tenant's users; offboard frees the licensed band slot and
(S-T2) tears down a siloed/hybrid tenant's **containers** (schema /
ClickHouse database). Pooled rows still exist at this point — that is what
step 3 erases.

## 3. Erase (irreversible, verifiable)

Provider console (admin): **tenant row → Erase**, or
`POST /provider/v1/tenants/{id}/erase` with `{"confirm":"<tenant-slug>"}` —
or tenant-side `POST /v1/lifecycle/erase` (permission `lifecycle.erase`).
The engine, per store:

| Store | Mechanism | Verification |
|---|---|---|
| Postgres (pooled or silo-routed) | per-table `DELETE` **under the tenant's own scope** (RLS + silo routing — cannot touch another tenant), multi-pass for FK ordering | per-table `count(*) == 0` in-scope |
| Provider rows about the tenant (usage, quotas, branding, break-glass, retention) | provider-role scoped deletes | per-table count == 0 |
| ClickHouse flows | pooled: synchronous lightweight delete (`mutations_sync=2`); siloed: `DROP DATABASE` | post-delete count == 0 |
| Object store | `DeletePrefix` on `tenant/<id>/` and `silo/<id>/` | post-delete list empty |
| Tenant keys (S-T6, byok-licensed) | **crypto-shred**: every key version's wrapped KEK nulled, state `destroyed` — any `tk1:` ciphertext (including in backups) is permanently unreadable; destroyed chains refuse re-keying | versions-destroyed count on the attestation's `tenant_keys` line; unlicensed deployments record "no per-tenant keyring installed" |
| TSDB | memory mode: in-place series delete. **Prometheus mode: MANUAL STEP** — run the admin `delete_series` API for `{tenant_id="<id>"}` or let retention expire them; the attestation marks this store incomplete until then | per mode |

The tenant registry row is **tombstoned** (`status=deleted`) — the
attestation keeps a referent; the row holds no telemetry.

## 4. The attestation

The engine returns (and the provider audit chain records, with the report's
SHA-256) the deletion report: per-store deleted counts + verified-zero
flags + your backup-TTL statement + `complete`. **If `complete=false`, the
notes name exactly what remains** (e.g. the prometheus manual step) — finish
it and re-run erase (idempotent). Hand the attestation JSON to the customer;
the SHA-256 on the tamper-evident provider chain is their proof it was not
edited after the fact.

## 5. After

- Backups: the attested deletion covers LIVE stores. Your snapshots expire
  per the stated TTL — do not restore an erased tenant's backup except
  under legal hold.
- Per-domain TLS certs (S-T4 custom domains): remove the ingress cert/DNS.
- The agents' mTLS identities are dead with the registry rows; decommission
  the tenant's agent hosts.
