-- 0044_auth_provider_rls.sql
-- TENANT-104 follow-up: the boot-time isolation posture check
-- (internal/tenancy.AssertIsolationPosture) FATALs the control plane unless
-- EVERY table carrying a tenant_id column has FORCE ROW LEVEL SECURITY. Three
-- tables predate that rule and were created without RLS because they are read
-- OUTSIDE any tenant context, via the raw pool:
--   * sessions      — looked up BY TOKEN at login; the token hash IS the tenant
--                     selector, so the lookup runs before a tenant is known.
--   * scim_tokens   — same shape: SCIM requests authenticate by token hash.
--   * break_glass_grants — provider-plane; operators create/list/revoke grants
--                     across tenants (internal/store/provider.go), with no
--                     tenant GUC set.
--
-- Apply the SAME policy migration 0040 used for mcp_tokens (the audit there
-- noted it is "the same shape as sessions"): the policy is FAIL-CLOSED when a
-- tenant context is set (the probectl.tenant_id GUC) — any in-tenant query gets
-- storage-layer isolation for free — and UNRESTRICTED when the GUC is UNSET,
-- which is exactly today's behavior and is what the pre-tenant auth lookups and
-- the provider-plane break-glass paths require. Because "GUC unset => allow"
-- is evaluated by the policy itself (not by a role privilege), this works for
-- ANY login role — a superuser/BYPASSRLS connection (dev/CI) or a hardened
-- non-superuser member role (sovereign deployments) alike. No new roles and no
-- application changes are needed; the existing table grants are unchanged.
--
-- Defense-in-depth, strictly additive vs today's no-RLS. Idempotent +
-- backward-compatible (CLAUDE.md §6).

DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['sessions', 'scim_tokens', 'break_glass_grants']
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
        EXECUTE format($pol$
            CREATE POLICY tenant_isolation ON %I
              USING (
                NULLIF(current_setting('probectl.tenant_id', true), '') IS NULL
                OR tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid
              )
              WITH CHECK (
                NULLIF(current_setting('probectl.tenant_id', true), '') IS NULL
                OR tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid
              )
        $pol$, t);
    END LOOP;
END $$;
