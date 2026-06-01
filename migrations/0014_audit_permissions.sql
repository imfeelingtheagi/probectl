-- 0014_audit_permissions.sql
-- Launch hardening (S19, F23): RBAC permission keys for reading and exporting the
-- tamper-evident audit trail. The audit log is sensitive, so these are granted to
-- the admin role only (NOT to viewer/editor) — adding them here, after the S13
-- role seed, means viewer's "%.read" grant does not retroactively include them.
-- Idempotent.

INSERT INTO permissions (key, description) VALUES
    ('audit.read',   'Read the tenant audit trail'),
    ('audit.export', 'Export / stream the tenant audit trail to a SIEM')
ON CONFLICT (key) DO NOTHING;

-- Grant the new audit permissions to the default tenant's admin role only.
INSERT INTO role_permissions (tenant_id, role_id, permission_key)
    SELECT r.tenant_id, r.id, p.key
    FROM roles r
    CROSS JOIN (VALUES ('audit.read'), ('audit.export')) AS p(key)
    WHERE r.tenant_id = '00000000-0000-0000-0000-000000000001' AND r.slug = 'admin'
ON CONFLICT (role_id, permission_key) DO NOTHING;
