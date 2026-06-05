# Per-tenant key isolation / BYOK (S-T6, F56)

probectl can encrypt each tenant's sensitive at-rest values under that
tenant's **own key**, making tenants cryptographically separable and
offboarding a **key-destruction event**. This is the cryptographic complement
to siloed storage (S-T2): silos separate *where* data lives; per-tenant keys
separate *who can ever read it*.

**Edition:** `ee/` — unlocked by the `byok` feature (Enterprise tier; also
consumed by MSP crypto-offboarding). Unlicensed deployments keep the
deployment-wide envelope (below) and the `/v1/security/keys` surface stays
hidden (404).

## The sealing model

All sensitive tenant-owned values (today: alert-channel secrets — webhook
HMACs, integration tokens; the set grows by consumer, not by mechanism) pass
through one core seam, `internal/tenantcrypto`. Stored values are
**self-describing**:

| Stored prefix | Sealed by | Notes |
|---|---|---|
| *(none)* | nothing (legacy/dev plaintext) | read as-is; re-sealed on next write |
| `dv1:` | the **deployment envelope** — one master KEK (`PROBECTL_ENVELOPE_KEY`) | the default whenever the envelope key is configured |
| `tk1:<version>:` | the **per-tenant keyring** (this feature) | tenant + version bound into the AEAD |

Reads dispatch on the prefix, so enabling per-tenant keys **never migrates
data**: existing `dv1:` rows keep decrypting (the deployment sealer stays
registered as an opener), and new writes seal under the tenant key —
decrypt-on-read compatibility is structural.

**The fail-safe rule (non-negotiable):** once a value is sealed under a
tenant key, opening it requires *that* key. An unavailable, unresolvable, or
destroyed key is an **error** — never a silent fallback to the deployment
master or any shared key. A recognized sealed prefix with no installed sealer
also refuses to pass through as plaintext.

## Key hierarchy

```
managed mode:  deployment master KEK (PROBECTL_ENVELOPE_KEY)
                 └─ wraps each tenant's KEK (random 32B, stored wrapped in tenant_keys)
                      └─ encrypts that tenant's values (AES-GCM via internal/crypto)

byok mode:     customer's secret manager (Vault / CyberArk / cloud KMS, via S41 refs)
                 └─ holds the tenant KEK (base64, exactly 32 bytes); probectl stores
                    ONLY the reference (e.g. vault:kv/tenants/acme#kek) and resolves
                    it at use time — the material is never persisted by probectl
```

A tenant's first seal **auto-provisions** a managed v1 key. The AAD binds
`tenant:<id>:<caller-context>:v<version>` into every ciphertext, so a sealed
blob cannot be replayed into another tenant's row even by direct DB writes.

## Rotation (no downtime)

`POST /v1/security/keys/rotate` (`security.keys` permission, admin-seeded):

- `{"mode": "managed"}` — mint a new wrapped KEK as version N+1.
- `{"mode": "byok", "byok_ref": "vault:kv/tenants/acme#kek"}` — activate BYOK.

The previous version is **retired, not destroyed**: new data seals under the
new version immediately; existing data keeps decrypting under its recorded
version. No re-encryption pass, no downtime. (Bulk rewrap of old rows is
deliberately deferred — values re-seal naturally on their next write.)

**The BYOK lockout guard:** a `byok_ref` is probe-resolved **before**
activation. A dead reference is rejected — you cannot rotate yourself into a
key probectl cannot reach. After activation, however, **the customer owns the
lock**: if you revoke probectl's access to the reference (or delete the
secret), your sealed data becomes unreadable and sealing fails — that is the
feature, not a bug. There is no recovery path through probectl. Document your
secret-manager's own backup/escrow policy accordingly.

## Cryptographic offboarding

Tenant erasure (S-T5) destroys the tenant's entire key chain **before** the
attestation is sealed: every version's wrapped KEK is nulled and the state
set `destroyed` (the `tenant_keys` rows survive as evidence, material-free).
Any `tk1:` ciphertext that ever leaves a backup window is permanently
unreadable. The S-T5 attestation carries a `tenant_keys` store line recording
how many versions were crypto-shredded; deployments without the byok feature
record "no per-tenant keyring installed" — honest, never implied.

Destroyed chains refuse re-keying: a destroyed tenant cannot silently get a
fresh v1 by writing new data.

## Surfaces

- `GET /v1/security/keys` — the chain (version/mode/state/timestamps). Key
  **material never crosses this API** in either direction.
- `POST /v1/security/keys/rotate` — managed rotation or BYOK activation.
- Tenant **Admin → Encryption keys** card: chain state, managed rotation,
  BYOK activation. Hidden entirely when unlicensed.
- Rotations are audited (`security.key_rotate`, tenant stream).

## Configuration

| Key | Meaning |
|---|---|
| `PROBECTL_ENVELOPE_KEY` | base64 32-byte deployment master KEK. Required for `dv1` sealing and (as the wrap root) for managed tenant keys. A licensed byok deployment **must** set it. |
| `PROBECTL_ENVELOPE_KEY_ID` | key identifier recorded in `dv1` values (default `dev`). |
| S41 secret backends | BYOK references resolve through the same resolver as every other secret reference (`vault:`, `cyberark:`, `awskms:` …). |

## Operational notes

- **KEK cache:** unwrapped/resolved KEKs are cached in memory for 30 seconds
  (rotation and destroy purge the tenant's cache immediately). A BYOK
  revocation is therefore effective within seconds, not sessions.
- **Key-store outage:** Postgres unavailability for `tenant_keys` fails
  seal/open with an error (fail safe). Telemetry ingestion is unaffected —
  only sealed-value reads/writes (e.g. alert-channel secret use) degrade.
- **What is NOT per-tenant-keyed:** bulk telemetry stores (ClickHouse/TSDB)
  rely on S-T2 isolation + S-T5 verifiable deletion; per-tenant keys cover
  the *sensitive-value* class. Extending coverage is a consumer-by-consumer
  decision through the same seam.
