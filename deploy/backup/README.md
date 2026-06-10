# Scheduled backups

Cron examples for the two durable stores — PostgreSQL (control-plane state)
and ClickHouse (high-cardinality events). The scripts they wrap
(`scripts/backup_*.sh` / `scripts/restore_*.sh`), the restore procedure,
RTO/RPO expectations, and the recovery drill are in
[`docs/ops/backup-restore.md`](../../docs/ops/backup-restore.md). The restore
path is exercised, not asserted: CI's `backup-drill` job runs a full
seed → backup → wipe → restore → verify cycle (`make backup-restore-drill`) on
every pass.

**Backups contain tenant data.** Store them encrypted at rest on an
access-controlled volume/bucket inside the operator's own infrastructure —
telemetry, including its backups, never leaves the operator's network (a
[non-negotiable](../../CONTRIBUTING.md#non-negotiables)).

## Compose (host cron)

`compose-backup.yml` defines one-shot backup services as an overlay on the
[dev/test stack](../compose/dev.yml). It carries that stack's fixed dev
credentials, so adapt the credentials (or use the Kubernetes paths below) for
anything beyond it. Schedule the services from the host's crontab:

```cron
# Nightly at 02:00/02:15 — keep PG and CH staggered.
0 2 * * *  cd /opt/probectl && docker compose -f deploy/compose/dev.yml -f deploy/backup/compose-backup.yml run --rm pg-backup
15 2 * * * cd /opt/probectl && docker compose -f deploy/compose/dev.yml -f deploy/backup/compose-backup.yml run --rm ch-backup
```

Postgres dumps (plus a `.sha256`) land in the `backups` volume. ClickHouse's
`BACKUP` statement runs **server-side**, so its archives land on the ClickHouse
container's own backups disk — the `chbackups` volume configured by
[`clickhouse-backups.xml`](../compose/clickhouse-backups.xml). Copy artifacts
off-box (the restore scripts take the off-box file) and prune to your
retention.

**ClickHouse backups disk must be writable by the clickhouse user (uid
101).** A freshly created volume mounts root-owned: the dev/compose scripts
fix this with a best-effort root `chmod 1777 /backups`; in Kubernetes set
the **ClickHouse server pod's** `securityContext.fsGroup: 101` (or pre-chown
the PVC) so the `BACKUP`/`RESTORE` statements — which write server-side —
can create their files and lock.

## Kubernetes

Two supported paths:

- **Helm-managed** — set `backup.enabled=true` on the `probectl` chart. It
  renders Postgres + ClickHouse CronJobs from the same digest-pinned images,
  envelope-encrypts the Postgres dump in-pipe (plaintext never touches the
  backups volume), and is wired by `backup.credentialsSecret` plus a backups
  PVC (`backup.persistence.*`). Off by default; the strict profile enables it.
  See [`deploy/helm/`](../helm/README.md).
- **Standalone manifests** — `k8s-cronjob-postgres.yaml` and
  `k8s-cronjob-clickhouse.yaml` (images digest-pinned to the same versions the
  compose stack runs) for clusters that don't use the chart: adjust the
  namespace, the `probectl-backups` PVC, and the `probectl-db-credentials`
  secret to your deployment, then `kubectl apply`.
