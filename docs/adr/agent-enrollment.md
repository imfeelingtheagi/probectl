# ADR: Agent enrollment & SVID issuance (Sprint 11 — WIRE-002/RED-002/TENANT-103/ARCH-004)

**Status:** ACCEPTED (founder-approved 2026-06-07; mint surface: admin API + CLI)
**Decision drivers:** the Sprint 4 server-side tenant binding verifies agents
against a registry whose entries are created from the mTLS certificate's
SPIFFE identity — but nothing in the repo *issues* those certificates. The
trust root is operator-manual (`gen-cert` + hand-distribution), so the
strongest isolation chain in the product rests on an undocumented manual step.

## What exists already (this ADR builds on, not around)

- mTLS transport with SPIFFE URI identity, verified per connection
  (`internal/agenttransport`, `crypto.ServerMTLSConfigRevocable`) and a
  handshake-checked revocation list (U-038).
- `crypto.AgentSPIFFEID(tenant, agent)` — the identity scheme
  (`spiffe://probectl/tenant/<tenant>/agent/<agent>`), trust-domain-pinned.
- `crypto.CA` — ECDSA P-256 CA + SPIFFE-SAN leaf issuance (dev/test today).
- `crypto.ClientMTLSConfigRotating` — the agent client already hot-reloads
  cert material; rotation needs an issuer, not a transport change.
- Registration (F50): tenant/agent are read from the VERIFIED cert, never the
  request.

## Decision 1 — Bootstrap: single-use, tenant-scoped join tokens

An operator (admin RBAC, audited) mints an **enrollment token** scoped to a
tenant (and optionally a fixed agent id): 32 random bytes via
`internal/crypto`, shown once, **stored only as a hash** (the session-token
pattern). The agent presents it exactly once; the row is consumed atomically
(`UPDATE ... WHERE used_at IS NULL` — replay = no row = refusal). Tokens
expire (default 1h) and are revocable before use.

*Why not cloud-IID/OIDC attestation now:* probectl's primary deployments are
sovereign/air-gapped — a token works everywhere a cloud identity document
does not exist. The enrollment endpoint is a seam: an `attestor` field on the
request leaves room for `aws-iid`/`gcp`/`oidc` attestors later without
changing the issued identity shape.

## Decision 2 — CA hierarchy: repo-managed root → intermediate → leaf

- **Agent root CA** (10y, P-256): generated once by
  `probectl-control agent-ca init`; signs ONLY intermediates. The root KEY is
  printed/exported at init for offline custody and is NOT required at
  runtime. (`MaxPathLen=1`.)
- **Issuing intermediate** (1y default): held by the control plane, **sealed
  at rest through `internal/tenantcrypto`** (the deployment envelope / BYOK —
  same posture as every other secret). Signs leaves only (`MaxPathLenZero`).
- **Leaf SVIDs** (default TTL **24h**): issued from a **CSR** — the private
  key is generated on the agent and never leaves it. The SVID carries the
  SPIFFE URI SAN binding `tenant + agent`, client-auth EKU only.
- The CA **bundle** (root + intermediate) is what transports trust; agents
  receive it at enrollment and on every rotation (intermediate roll-over =
  agents pick up the new bundle on next rotation).
- trustctl (the sibling cert-lifecycle product) can later REPLACE the issuing
  intermediate; the enrollment/rotation API is the integration seam.

## Decision 3 — Issuance flows

**Enroll (pre-identity, HTTPS):** `POST /enroll/agent` on the control API (off /v1 — the /v1 surface is the RBAC'd session API; this is a bootstrap surface like /auth)
— server-auth TLS only (the agent has no cert yet; HTTPS-by-default recipes
make this safe), authenticated by the join token. Request: token, agent CSR,
hostname/version. Server: consume token → derive `tenant` from the TOKEN
(never the request) → assign/verify agent id → sign leaf → record issued
identity (serial, SPIFFE id, expiry) in the registry → return leaf + CA
bundle. The agent is thereby registered (the Sprint 4 binding sees it).

**Rotate (identified, mTLS):** the agent calls rotation over its EXISTING
mTLS channel before expiry (at ~2/3 TTL). The server authenticates the
CURRENT SVID, requires the CSR's identity to EQUAL the proven identity
(tenant+agent — no privilege change on rotation), checks the revocation list,
issues a fresh leaf, records it. `ClientMTLSConfigRotating` hot-swaps the
files; connections re-handshake naturally.

**Agent CLI:** `probectl-agent enroll --server https://control:8443
--token <jt> --dir /var/lib/probectl-agent/identity [--ca-pin <sha256>]`
writes key/cert/bundle (0600) and exits; the runtime config points at those
paths and the runtime rotates them automatically. `--ca-pin` (printed when
the token is minted) authenticates the SERVER on first contact in
self-signed/quickstart deployments — trust-on-first-use is refused when a pin
is provided and mismatched.

## Threat-model delta (what changes, honestly)

| Threat | Before | After |
|---|---|---|
| Forged tenant claim on the bus (S4 residual) | possible with another tenant's REGISTERED agent id over pooled bus creds | agent ids exist only via enrollment; identity is cryptographic end-to-end — the S4 registry rows now have an issuance provenance |
| Token theft in transit/at rest | n/a (no tokens) | single-use + short expiry + hash-at-rest + tenant-scoped: a stolen UNUSED token is a bounded, audited window; a USED one is inert; minting and consumption are both audited |
| Control-plane DB read | cert material was operator-managed, outside the DB | intermediate CA key sealed via tenantcrypto (envelope/BYOK); a DB read without the KEK yields ciphertext; token hashes are one-way |
| Stolen agent key | manual revocation of a manually-tracked cert | 24h leaf TTL bounds exposure; serial recorded at issuance feeds the EXISTING handshake revocation list (operator path lands in Sprint 12) |
| Rogue/compromised control plane | already total (it terminates ingest) | unchanged — the issuer is the control plane; root custody offline limits blast radius to intermediate lifetime |
| Enrollment endpoint abuse | n/a | unauthenticated callers can only burn CPU: every request requires a valid unconsumed token before any signing; rate-limited like login surfaces |
| Wrong-tenant enrollment | operator error distributes a cert with the wrong SPIFFE | tenant comes ONLY from the token; an agent cannot request a tenant |

**Residuals after this sprint (stated):** (1) revocation *feeding* is Sprint
12 (the check exists; this sprint records serials so 12 has data); (2) no
workload attestation beyond token possession at first boot — cloud/OIDC
attestors are the documented extension seam; (3) the bus credential itself
(Kafka ACLs) is unchanged — the S4 consumer-side verification plus
cryptographic issuance is the compensating pair until siloed lanes.

## Out of scope (deliberate)

External CA integration (trustctl seam documented), cloud attestors, CRL/OCSP
distribution (the in-process revocation list is the mechanism; Sprint 12
wires its operator path), per-probe identities.
