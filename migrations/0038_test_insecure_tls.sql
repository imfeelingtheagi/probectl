-- 0038: insecure-TLS canary override permission (U-040, C10).
--
-- insecure_skip_verify=true on an HTTP canary disables certificate failure
-- on the probe (the chain is still captured for S27 posture). Setting it is
-- deny-by-default: gated on this admin-seeded permission and recorded
-- explicitly in the test.create/test.update audit entries.
--
-- Idempotent + expand-only (CLAUDE.md §6).

INSERT INTO permissions (key, description) VALUES
    ('test.insecure_tls', 'Allow a test to disable TLS certificate verification (insecure_skip_verify; audited)')
ON CONFLICT (key) DO NOTHING;

INSERT INTO role_permissions (tenant_id, role_id, permission_key)
    SELECT r.tenant_id, r.id, 'test.insecure_tls'
    FROM roles r
    WHERE r.slug = 'admin' AND r.is_system
ON CONFLICT (role_id, permission_key) DO NOTHING;
