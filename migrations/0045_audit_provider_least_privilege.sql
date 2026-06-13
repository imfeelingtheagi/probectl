-- 0045_audit_provider_least_privilege.sql — TENANT-005: tighten the provider
-- role's audit_events capability to least privilege.
--
-- 0029 granted the provider role a `FOR ALL USING(true)` policy plus
-- `GRANT SELECT, DELETE`. The erase engine only ever needs to (a) DELETE one
-- tenant's append-only chain and (b) COUNT-VERIFY the remainder for that SAME
-- tenant. The `FOR ALL USING(true)` SELECT path was an unconstrained
-- cross-tenant read capability — a future provider code path (or a compromise
-- of the provider connection) could enumerate EVERY tenant's tamper-evident
-- audit log, contradicting "provider operators get no implicit read access to
-- tenant telemetry" (guardrail 7.1/7.7).
--
-- Replace the blanket policy with two least-privilege policies:
--   * provider_audit_erase   FOR DELETE USING(true)  — the erase mutation
--     (the provider deletes by an explicit `WHERE tenant_id = $1`).
--   * provider_audit_verify  FOR SELECT — GUC-scoped EXACTLY like the
--     app-role audit_select policy: the provider can read audit_events only
--     for the tenant currently in `probectl.tenant_id`. With no GUC set (the
--     default provider transaction) it reads NOTHING — fail closed. The erase
--     engine sets the GUC for the verify read of the tenant being erased.
--
-- Net effect: the provider can no longer SELECT another tenant's audit rows
-- outside the scoped, in-progress erase. Break-glass tenant reads continue to
-- run through the tenant (app) role + GUC, never the provider role.
--
-- Idempotent + expand-only (CLAUDE.md §6).

DROP POLICY IF EXISTS provider_lifecycle_erase ON audit_events;

DROP POLICY IF EXISTS provider_audit_erase ON audit_events;
CREATE POLICY provider_audit_erase ON audit_events
  FOR DELETE TO probectl_provider USING (true);

DROP POLICY IF EXISTS provider_audit_verify ON audit_events;
CREATE POLICY provider_audit_verify ON audit_events
  FOR SELECT TO probectl_provider
  USING (tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid);

-- The grant stays SELECT, DELETE — the policies now constrain SELECT to the
-- GUC-scoped tenant rather than every tenant.
GRANT SELECT, DELETE ON audit_events TO probectl_provider;
