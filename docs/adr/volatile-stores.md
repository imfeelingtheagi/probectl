# ADR: rebuild-on-restart for the in-process derived stores (U-047)

- **Status:** Accepted — 2026-06-07
- **Context finding:** U-047 (O:ARCH-007) — topology, threat detections, and
  alert state are in-process and lost on restart.

## Decision

Three in-process stores are **derived views of a durable source**, not
systems of record. For them we **formally adopt rebuild-on-restart** rather
than adding persistence — and prove the cold-start behavior with tests. One
piece of genuinely non-derivable state (alert silences/acks) is the
exception and is called out as a tracked gap, not silently dropped.

| Store | What it holds | Durable source it re-derives from | Restart behavior |
|---|---|---|---|
| `internal/topology` (`MemoryStore` / `IndexedStore`) | the tenant-scoped temporal graph | the bus stream (`probectl.ebpf.flows` etc.) + the path/flow stores it observes; the `TopologyConsumer` re-subscribes at boot | **rebuilds** as new observations arrive; cold start is an empty graph that refills within the stream/retention window |
| `internal/threat` (`DetectionStore`) | recent NDR/posture detections per tenant, bounded | the detection stream; the **durable trail is the incident record + SIEM export** (already persisted) | **rebuilds** from the stream; the forensic copy already left via incidents/SIEM, so nothing forensic is lost |
| `internal/alert` (`Engine` firing state) | per-series firing/pending state | re-derived on the **next evaluation** against the metric source | firing state **re-derives** automatically; see the exception below |

## Why rebuild rather than persist

These are **caches of a stream**, not the system of record. The
authoritative data already lives durably: flows/paths in ClickHouse, metric
series in the TSDB, detections' forensic copy in incidents + the SIEM
export, audit in the tamper-evident chain. Persisting a *second* copy of a
derived view would add write-path cost, a schema to migrate, and a
consistency problem (the snapshot vs. the live stream) for state that is
correct again within one evaluation/observation cycle. Rebuild-on-restart
is the simpler, correct choice — and it is now a **decided, tested**
property instead of an implicit one.

Cross-tenant isolation is unaffected: each store is keyed by tenant and a
cold start cannot surface another tenant's data (it surfaces *nothing* until
re-derived).

## The exception: alert silences & acknowledgements

Silences and acks are **operator inputs**, not derivable from any stream —
re-deriving firing state cannot reconstruct "operator X silenced this until
3pm." Today they are in-process and **lost on restart** (the firing alert
simply reappears un-silenced — fail-safe: louder, not quieter). This is the
one real durability gap. It is **documented in code** (`internal/alert`)
and tracked as a follow-up: silences/acks should ride a small
Postgres-backed table the way alert *rules* already do. Until then,
operators re-apply silences after a control-plane restart.

## Consequences

- A control-plane restart shows a briefly empty topology/detection view and
  reappearing (un-silenced) firing alerts — expected and documented.
- No new migration, write path, or snapshot-consistency surface for derived
  state.
- The cold-start contract is enforced by tests (`internal/topology`,
  `internal/alert`): a fresh store is empty and correct, and re-derives from
  its inputs.

## Tests proving cold start

- `internal/topology`: a new `MemoryStore` returns an empty snapshot for any
  tenant (no stale data, no panic) and rebuilds the graph as observations
  replay.
- `internal/alert`: a new `Engine` has no active alerts and no silences/acks
  at construction, and re-derives firing state from the metric source on the
  first evaluation after a "restart" (fresh engine).
