# High availability and the in-memory view constraint (ARCH-003)

This page explains, in plain terms, which parts of probectl are safe to run with
more than one control-plane replica today, and which are not yet — so you can
make an informed scaling choice while the durable view-serving layer is being
built.

## The one-sentence version

Ingest and anything backed by a real datastore is coherent at any replica count;
a handful of read views that still live in each replica's RAM are not yet
coherent across replicas, so if you rely on those views being consistent, run a
single replica until the durable view layer lands.

## Why this exists

The control plane is stateless for the paths that matter most. Agent results,
flow batches, and device metrics are written straight through to the TSDB
(Prometheus/VictoriaMetrics) and ClickHouse. Any replica answering a query reads
the same shared store, so those answers are identical no matter which replica
you hit. You can scale those horizontally with no caveat.

A few features, however, build their serving state by consuming the bus into an
in-process structure and answering queries from that RAM copy:

- topology (`/v1/topology`) — the live adjacency graph,
- latest-result view (`/v1/results/latest`),
- threat detections (`/v1/threat/detections`),
- TLS/cert posture (`/v1/tls/posture`),
- endpoint/DEM views.

Each replica only consumes the slice of bus traffic that its own consumer group
members received. With more than one replica, replica A may have seen an edge or
detection that replica B has not, so the same query can return different answers
depending on which replica a load balancer happens to route you to. Nothing is
lost or cross-tenant — the data is still tenant-scoped and durable upstream —
but the *view* is not coherent across replicas.

## What to do today

| Deployment goal | Safe replica count |
|---|---|
| Ingest throughput / API for TSDB+ClickHouse-backed queries | any (scale freely) |
| Consistent topology / latest-result / threat / TLS / endpoint views | **1** (until the durable view layer lands) |

If you need both high ingest throughput and consistent in-RAM views right now,
the pragmatic split is: keep `replicaCount: 1` for the control plane and scale
the data tier (Kafka, ClickHouse, the TSDB) independently — that tier is where
the volume actually is.

## What's coming (ARCH-003)

The fix is to stop serving those views from per-replica RAM: either elect a
single leader to own the view consumers and have other replicas proxy to it, or
move the derived views into a shared store that every replica reads. Once that
ships, this constraint is lifted and the medium/large reference values run their
documented replica counts with coherent views. Track it under ARCH-003.
