# Self-hosted / air-gapped OIDC IdP (OPS-008)

probectl authenticates operators via **OIDC** (`AUTH_MODE=session`,
`PROBECTL_OIDC_ISSUER` + client id/secret + redirect URL). Nothing in
probectl requires a *cloud* IdP — any standards-compliant OIDC provider
works, including one you run **inside the air-gap**. This removes the last
external dependency from a sovereign deployment: telemetry never leaves the
network (CLAUDE.md §7.2) and now neither does the login flow.

## The contract

probectl is a plain OIDC **relying party**. The IdP must:

- expose a discovery document at `${issuer}/.well-known/openid-configuration`
  reachable from the control plane (in-cluster DNS is fine);
- issue ID tokens (the `openid` scope; `email`/`groups` if you map roles);
- honor the `nonce` (probectl validates it on callback — SEC-004);
- redirect back to `${PROBECTL_OIDC_REDIRECT_URL}` over HTTPS.

Group/role mapping uses the same SCIM/ABAC path as any IdP (`docs/` identity
sections) — the self-hosted IdP is not a special case in probectl.

## Reference: Dex (smallest air-gap footprint)

[Dex](https://dexidp.io) is a tiny OIDC provider with a static-password
connector — no external directory needed, ideal for an air-gapped install.
Run it in-cluster and point probectl at it:

```yaml
# dex-config.yaml (ConfigMap) — issuer is the in-cluster service URL
issuer: https://dex.probectl.svc.cluster.local:5556/dex
storage:
  type: kubernetes        # or sqlite3 on a PVC for a single replica
  config: { inCluster: true }
web:
  https: 0.0.0.0:5556
  tlsCert: /etc/dex/tls/tls.crt
  tlsKey: /etc/dex/tls/tls.key
staticClients:
  - id: probectl
    name: probectl
    secret: "${DEX_PROBECTL_CLIENT_SECRET}"   # = PROBECTL_OIDC_CLIENT_SECRET
    redirectURIs:
      - https://probectl.example/auth/callback # = PROBECTL_OIDC_REDIRECT_URL
enablePasswordDB: true
# staticPasswords: bootstrap an admin; thereafter wire an in-network LDAP if you have one
```

probectl side (Helm values or env):

```sh
PROBECTL_AUTH_MODE=session
PROBECTL_OIDC_ISSUER=https://dex.probectl.svc.cluster.local:5556/dex
PROBECTL_OIDC_CLIENT_ID=probectl
PROBECTL_OIDC_CLIENT_SECRET=...        # the Dex staticClient secret (a secret ref)
PROBECTL_OIDC_REDIRECT_URL=https://probectl.example/auth/callback
```

The Dex image is digest-pinned in your registry mirror like every other
air-gapped image (the air-gapped bundle, `docs/hardening.md`).

## Reference: Keycloak (full-feature)

For larger orgs already running **Keycloak**, create a realm + a
confidential `probectl` client (standard flow, the redirect URI above) and
point `PROBECTL_OIDC_ISSUER` at
`https://keycloak.internal/realms/<realm>`. Keycloak's discovery, nonce, and
group claims satisfy the contract above unchanged. Run it on an in-network
host with its own datastore; nothing crosses the air-gap.

## Trust & TLS

The control plane validates the IdP's TLS certificate (guardrail 12, never
disabled). For an internal CA, mount your CA bundle so the control plane
trusts the IdP's cert — the same trust store the rest of probectl uses for
outbound TLS. A self-signed IdP cert from a private CA is fine **as long as
that CA is in the trust store** — probectl never skips verification.

## Exercised

The OIDC relying-party path (discovery, nonce validation, callback, role
mapping) is covered by the auth suite. The IdP itself is operator-run; the
air-gap exercise — stand up Dex in a disconnected cluster and complete a
login — is the deployment-time `[needs infra]` half of OPS-008, scripted by
the values above.
