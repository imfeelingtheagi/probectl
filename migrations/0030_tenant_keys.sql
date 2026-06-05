-- 0030_tenant_keys.sql — S-T6 (MT/EE): per-tenant key isolation / BYOK.
--
-- tenant_keys: the per-tenant KEK chain. Managed mode stores the tenant KEK
-- WRAPPED under the deployment master (never plaintext); BYOK mode stores
-- only a secret REFERENCE (S41: vault:/aws:/azure:/gcp:/cyberark:) — the key
-- material itself never touches probectl storage. Versions implement
-- rotation: the active version seals new data; retired versions stay
-- decrypt-only; destroyed versions render their ciphertexts permanently
-- unreadable (cryptographic offboarding).
--
-- Tenant-RLS'd (a tenant manages its own keys via security.keys) + the
-- explicit provider policy (the erase engine destroys keys at
-- crypto-offboarding). NEVER copied into silo schemas (key material is
-- control-plane security state).
--
-- Idempotent + expand-only (CLAUDE.md §6).

INSERT INTO permissions (key, description) VALUES
    ('security.keys', 'Manage the tenant''s at-rest encryption keys (rotate, BYOK)')
ON CONFLICT (key) DO NOTHING;

INSERT INTO role_permissions (tenant_id, role_id, permission_key)
    SELECT r.tenant_id, r.id, 'security.keys'
    FROM roles r
    WHERE r.tenant_id = '00000000-0000-0000-0000-000000000001'
      AND r.slug = 'admin'
ON CONFLICT (role_id, permission_key) DO NOTHING;

CREATE TABLE IF NOT EXISTS tenant_keys (
    tenant_id    uuid        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    version      integer     NOT NULL,
    mode         text        NOT NULL DEFAULT 'managed'
                   CHECK (mode IN ('managed', 'byok')),
    state        text        NOT NULL DEFAULT 'active'
                   CHECK (state IN ('active', 'retired', 'destroyed')),
    wrapped_kek  bytea,                -- managed: KEK sealed under the master
    byok_ref     text NOT NULL DEFAULT '', -- byok: the S41 secret reference
    created_at   timestamptz NOT NULL DEFAULT now(),
    retired_at   timestamptz,
    destroyed_at timestamptz,
    destroyed_by text        NOT NULL DEFAULT '',
    PRIMARY KEY (tenant_id, version)
);
-- One active version per tenant.
CREATE UNIQUE INDEX IF NOT EXISTS tenant_keys_active_idx
    ON tenant_keys (tenant_id) WHERE state = 'active';

ALTER TABLE tenant_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_keys FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_keys;
CREATE POLICY tenant_isolation ON tenant_keys
  USING (tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid);
DROP POLICY IF EXISTS provider_keys ON tenant_keys;
CREATE POLICY provider_keys ON tenant_keys
  FOR ALL TO probectl_provider USING (true) WITH CHECK (true);

GRANT SELECT, INSERT, UPDATE ON tenant_keys TO probectl_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_keys TO probectl_provider;
