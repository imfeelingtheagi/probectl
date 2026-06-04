-- 0021_flow_permissions.sql — S38 (F17): the flow-analytics read permission.
--
-- Flow records themselves live in ClickHouse (internal/store/flowstore), not
-- Postgres — this migration only seeds the RBAC permission gating the
-- /v1/flows/* views (the tenant boundary is enforced first, then this).
-- Idempotent + expand-only (CLAUDE.md §6).

INSERT INTO permissions (key, description) VALUES
    ('flow.read', 'Read flow analytics: top-talkers, capacity, anomaly views')
ON CONFLICT (key) DO NOTHING;

-- Grant flow.read to the default tenant's seeded read roles (admin/editor/viewer);
-- other tenants are seeded at provisioning.
INSERT INTO role_permissions (tenant_id, role_id, permission_key)
    SELECT r.tenant_id, r.id, 'flow.read'
    FROM roles r
    WHERE r.tenant_id = '00000000-0000-0000-0000-000000000001'
      AND r.slug IN ('admin', 'editor', 'viewer')
ON CONFLICT (role_id, permission_key) DO NOTHING;
