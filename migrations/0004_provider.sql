-- 0004_provider.sql
-- Provider / MSP management plane (F51-foundation) — a privilege domain DISTINCT
-- from tenant data. Provider operators are NOT tenant users: they live in their
-- own table and never get implicit read access to tenant telemetry. Accessing a
-- tenant's data requires an explicit, time-bounded, tenant-consented, separately
-- audited break-glass grant (CLAUDE.md §7 guardrails 1, 7). These tables are
-- global (not tenant-owned), so they carry no tenant_id and no tenant RLS.

CREATE TABLE IF NOT EXISTS provider_operators (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email      text NOT NULL UNIQUE,
    name       text NOT NULL DEFAULT '',
    status     text NOT NULL DEFAULT 'active'
                 CHECK (status IN ('active', 'suspended', 'disabled')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS break_glass_grants (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id uuid NOT NULL REFERENCES provider_operators(id) ON DELETE CASCADE,
    tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reason      text NOT NULL,
    scope       text NOT NULL DEFAULT 'read' CHECK (scope IN ('read', 'write')),
    granted_by  text NOT NULL,
    granted_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz,
    revoked_by  text,
    CHECK (expires_at > granted_at)
);
CREATE INDEX IF NOT EXISTS break_glass_operator_idx ON break_glass_grants (operator_id);
CREATE INDEX IF NOT EXISTS break_glass_tenant_idx ON break_glass_grants (tenant_id);
-- Active grants: not revoked and not expired (filtered in queries by now()).
CREATE INDEX IF NOT EXISTS break_glass_active_idx ON break_glass_grants (tenant_id, expires_at)
    WHERE revoked_at IS NULL;
