-- 0028_tenant_lifecycle.sql — S-T5 (MT): per-tenant lifecycle (CORE).
--
-- Export + verifiable deletion are a COMPLIANCE RIGHT, not a commercial
-- feature (ratified editions decision): the permissions and the retention
-- table are core schema. lifecycle.export downloads the tenant's data;
-- lifecycle.erase sets retention and runs the irreversible full erasure —
-- seeded to ADMINS ONLY (big hammers carry explicit intent).
--
-- tenant_retention: per-tenant erasure controls (NULL = deployment default).
-- Tenant-RLS'd (a tenant manages its own policy) + the explicit provider
-- policy (the sweeper + provider console read across tenants).
--
-- Idempotent + expand-only (CLAUDE.md §6).

INSERT INTO permissions (key, description) VALUES
    ('lifecycle.export', 'Export the tenant''s data (portability bundle)'),
    ('lifecycle.erase',  'Set retention/erasure policy and run verifiable tenant erasure')
ON CONFLICT (key) DO NOTHING;

INSERT INTO role_permissions (tenant_id, role_id, permission_key)
    SELECT r.tenant_id, r.id, p.key
    FROM roles r, (VALUES ('lifecycle.export'), ('lifecycle.erase')) AS p(key)
    WHERE r.tenant_id = '00000000-0000-0000-0000-000000000001'
      AND r.slug = 'admin'
ON CONFLICT (role_id, permission_key) DO NOTHING;

CREATE TABLE IF NOT EXISTS tenant_retention (
    tenant_id           uuid        PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    flow_retention_days integer,
    updated_at          timestamptz NOT NULL DEFAULT now(),
    updated_by          text        NOT NULL DEFAULT '',
    CHECK (flow_retention_days IS NULL OR flow_retention_days >= 1)
);

ALTER TABLE tenant_retention ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_retention FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenant_retention;
CREATE POLICY tenant_isolation ON tenant_retention
  USING (tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid);
DROP POLICY IF EXISTS provider_retention ON tenant_retention;
CREATE POLICY provider_retention ON tenant_retention
  FOR ALL TO probectl_provider USING (true) WITH CHECK (true);

GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_retention TO probectl_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_retention TO probectl_provider;
