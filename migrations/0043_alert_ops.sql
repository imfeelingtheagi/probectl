-- 0043_alert_ops.sql
-- Sprint 16 (ARCH-005, scoped per docs/adr/volatile-stores.md): persist ONLY
-- alert silences/acks — operator inputs, not derivable from streams (the
-- ADR's documented exception). Firing state itself stays engine-derived.

CREATE TABLE IF NOT EXISTS alert_ops (
  tenant_id      uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  fingerprint    text NOT NULL,
  rule_id        text NOT NULL DEFAULT '',
  silenced_until timestamptz,
  acked_by       text NOT NULL DEFAULT '',
  acked_at       timestamptz,
  updated_at     timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, fingerprint)
);

ALTER TABLE alert_ops ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_ops FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON alert_ops;
CREATE POLICY tenant_isolation ON alert_ops
  USING (tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('probectl.tenant_id', true), '')::uuid);
GRANT SELECT, INSERT, UPDATE, DELETE ON alert_ops TO probectl_app;
