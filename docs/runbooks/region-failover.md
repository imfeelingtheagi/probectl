# Runbook: region failover

Promote a standby and move the PostgreSQL writer to a new region with no
split-brain and bounded data loss. See `docs/multi-region.md` for the model.

**Roles:** a DB operator (Postgres failover) + a platform operator (verify).
**Pre-req:** streaming replication healthy; the writer endpoint is a DNS name /
proxy you can re-point; `cluster_state` migrated (0032).

## 0. Decide: is a failover warranted?

A failover is for a **primary region loss** (the primary is down or isolated),
not transient lag. Check `/readyz` across regions and
`probectl_cluster_replica_lag_seconds`. probectl will already be fencing writes
(503 `writer_unavailable`) if the primary is unreachable — reads and telemetry
keep flowing meanwhile.

## 1. Confirm the standby is caught up (RPO)

- `sync` replication: the synchronous standby has all committed data → RPO 0.
- `async`: check the standby's replay lag at the moment of loss
  (`pg_last_xact_replay_timestamp()`). Data written within the lag window may be
  lost — record it. Do **not** promote a badly-lagged standby if a fresher one
  exists.

## 2. Promote the standby (DB operator)

Use your Postgres failover tooling (`pg_ctl promote`, Patroni `switchover`/
`failover`, or the managed-DB failover action). After promotion the new primary
is **writable** and on a **new timeline**.

## 3. Stamp the promotion epoch (the split-brain fence)

On the **newly promoted primary**, run:

```sql
SELECT cluster_promote('<new-writer-region>', '<operator>');
```

This bumps `cluster_state.writer_epoch` and records the new writer region. The
new epoch replicates to the other standbys. **This is the fence:** the old
primary keeps the lower epoch, so probectl will refuse to write to it even if
the endpoint briefly points back. Skipping this step risks a split-brain write —
**do not skip it.**

## 4. Re-point the writer endpoint

Move `PROBECTL_DATABASE_URL`'s DNS name / proxy to the promoted primary (managed
DB: usually automatic; Patroni/HAProxy: tracks the leader; manual DNS: update +
wait for TTL). Replicas (`PROBECTL_DATABASE_READ_URL`) in each region keep
pointing at their local node.

## 5. Verify (platform operator) — the RTO check

probectl re-probes every 5s and resumes writes automatically — **no restart
needed**. Confirm on a replica in the surviving region:

```
GET /readyz  →  cluster.writes_usable: true
              cluster.writer.region:  <new-writer-region>
              cluster.highest_epoch:  <bumped value>
```

A mutating request (e.g. saving config) should now succeed instead of 503.
Time from primary loss to `writes_usable:true` is the realised **RTO** —
compare against `PROBECTL_RTO_SECONDS` (provisional ≤ 60 s).

## 6. Rebuild the old region as a standby

Bring the former primary back as a **standby** of the new primary (re-clone or
`pg_rewind` onto the new timeline). It rejoins on the current epoch. Until it
does, probectl correctly fences it (lower epoch / in recovery).

## Watch-outs

- **Never** promote two standbys for the same cluster — only one primary.
  `cluster_promote` makes the winner unambiguous (highest epoch); a second
  promotion that does not also win the endpoint is fenced.
- **Residency:** do not fail a residency-restricted tenant's data into a region
  its policy forbids. Strict tenants run **siloed** (S-T2) with region-pinned
  stores rather than global replication.
- **Backups:** a failover is not a backup. Keep the documented backup/PITR
  policy (`PROBECTL_BACKUP_RETENTION_NOTE`) regardless.
