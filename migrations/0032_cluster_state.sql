-- 0032: Multi-region / active-active HA cluster state (S-EE2, F33).
--
-- cluster_state is a GLOBAL singleton (not tenant-owned): it records the
-- current PostgreSQL writer's promotion EPOCH and region. The epoch is the
-- split-brain fence — it increases by one on every promotion (the region
-- failover runbook calls cluster_promote() AFTER promoting the standby), and
-- it is carried to the replicas by streaming replication. A stale ex-primary
-- left behind by a partition keeps the OLD epoch, so the control plane fences
-- writes to it (a primary on a superseded epoch is never written to).
--
-- It is global infrastructure state, so it is never tenant-scoped and never
-- copied into a per-tenant silo schema (added to the silo deny list in core).
--
-- Idempotent + expand-only (CLAUDE.md §6).

CREATE TABLE IF NOT EXISTS cluster_state (
    id            boolean     PRIMARY KEY DEFAULT true,
    writer_region text        NOT NULL DEFAULT '',
    writer_epoch  bigint      NOT NULL DEFAULT 0,
    promoted_at   timestamptz NOT NULL DEFAULT now(),
    promoted_by   text        NOT NULL DEFAULT '',
    CONSTRAINT cluster_state_singleton CHECK (id)
);

-- Seed the single row (epoch 0). Repeated runs leave it untouched.
INSERT INTO cluster_state (id, writer_region, writer_epoch)
    VALUES (true, '', 0)
ON CONFLICT (id) DO NOTHING;

-- cluster_promote(region) is the authoritative promotion step: it bumps the
-- epoch and records the new writer region + actor. Run it on the NEWLY
-- PROMOTED primary as the final failover step (see docs/runbooks/region-
-- failover.md). The monotonic epoch is what fences the old primary.
CREATE OR REPLACE FUNCTION cluster_promote(p_region text, p_actor text DEFAULT '')
RETURNS bigint
LANGUAGE sql
AS $$
    UPDATE cluster_state
       SET writer_region = p_region,
           writer_epoch  = writer_epoch + 1,
           promoted_at   = now(),
           promoted_by   = p_actor
     WHERE id
    RETURNING writer_epoch;
$$;

-- The app role reads cluster_state (the fencing probe) from any endpoint.
-- Promotion is an operator/automation action via the provider role.
GRANT SELECT ON cluster_state TO probectl_app;
GRANT SELECT, UPDATE ON cluster_state TO probectl_provider;
GRANT EXECUTE ON FUNCTION cluster_promote(text, text) TO probectl_provider;
