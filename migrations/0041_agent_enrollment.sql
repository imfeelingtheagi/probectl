-- 0041_agent_enrollment.sql
-- Sprint 11 (WIRE-002/RED-002/TENANT-103/ARCH-004): agent enrollment + SVID
-- issuance state. ADR: docs/adr/agent-enrollment.md.
--
--   agent_enroll_tokens  one-time, tenant-scoped join tokens (HASH only —
--                        a DB read cannot mint an enrollment). Pre-tenant
--                        lookup by hash, so RLS follows the 0040 mcp_tokens
--                        pattern: unrestricted with NO tenant context (the
--                        consume path), tenant-confined when one is set.
--   agent_identities     every issued SVID (serial, spiffe, expiry) — the
--                        issuance provenance behind the Sprint 4 binding and
--                        the data source for Sprint 12 revocation.
--   agent_ca             the deployment's agent CA hierarchy: root CERT only
--                        (its key is exported once for offline custody and
--                        never stored); intermediate cert + key SEALED via
--                        internal/tenantcrypto. Deployment-global (no
--                        tenant_id — like schema metadata).

CREATE TABLE IF NOT EXISTS agent_enroll_tokens (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  agent_id      text,                    -- optional: pin the enrolling agent's id
  name          text NOT NULL DEFAULT '',
  token_hash    bytea NOT NULL UNIQUE,   -- crypto.Hash of the secret; never the secret
  created_by    text NOT NULL DEFAULT '',
  created_at    timestamptz NOT NULL DEFAULT now(),
  expires_at    timestamptz NOT NULL,
  used_at       timestamptz,             -- single-use: consumed atomically
  used_by_agent text,
  revoked_at    timestamptz
);
CREATE INDEX IF NOT EXISTS agent_enroll_tokens_tenant_idx ON agent_enroll_tokens (tenant_id);

ALTER TABLE agent_enroll_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_enroll_tokens FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON agent_enroll_tokens;
CREATE POLICY tenant_isolation ON agent_enroll_tokens
  USING (
    NULLIF(current_setting('probectl.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid
  )
  WITH CHECK (
    NULLIF(current_setting('probectl.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid
  );
GRANT SELECT, INSERT, UPDATE ON agent_enroll_tokens TO probectl_app;

CREATE TABLE IF NOT EXISTS agent_identities (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  agent_id     text NOT NULL,
  spiffe_id    text NOT NULL,
  serial       text NOT NULL UNIQUE,     -- hex; feeds the revocation list (Sprint 12)
  not_after    timestamptz NOT NULL,
  issued_at    timestamptz NOT NULL DEFAULT now(),
  rotated_from text                      -- previous serial when this is a rotation
);
CREATE INDEX IF NOT EXISTS agent_identities_tenant_agent_idx ON agent_identities (tenant_id, agent_id);

ALTER TABLE agent_identities ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_identities FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON agent_identities;
CREATE POLICY tenant_isolation ON agent_identities
  USING (
    NULLIF(current_setting('probectl.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid
  )
  WITH CHECK (
    NULLIF(current_setting('probectl.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid
  );
GRANT SELECT, INSERT ON agent_identities TO probectl_app;

CREATE TABLE IF NOT EXISTS agent_ca (
  kind       text PRIMARY KEY,           -- 'root' | 'intermediate'
  cert_pem   text NOT NULL,
  key_sealed text,                       -- tenantcrypto-sealed (intermediate only; root key NEVER stored)
  created_at timestamptz NOT NULL DEFAULT now()
);
GRANT SELECT, INSERT, UPDATE ON agent_ca TO probectl_app;
