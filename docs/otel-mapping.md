# Canonical signal → OpenTelemetry mapping

netctl's signal schemas are modeled on OpenTelemetry **resource** and **network
semantic conventions from their first emission** (S6), so the OTLP layer (S22)
*exposes* signals as OTLP rather than remapping a divergent model. This file is
the canonical mapping; a CI conformance test (`internal/otel`) enforces that the
code never invents an attribute name where an OTel convention exists.

## Result (`netctl.result.v1.Result`)

| proto field             | OTel attribute / role            | notes                                   |
| ----------------------- | -------------------------------- | --------------------------------------- |
| `tenant_id`             | resource: `netctl.tenant.id`     | outermost scope (F50); netctl namespace |
| `agent_id`              | resource: `netctl.agent.id`      | producing agent; netctl namespace       |
| `canary_type`           | `netctl.canary.type`             | icmp/tcp/udp/http/dns/… (netctl)        |
| `server_address`        | `server.address`                 | the probed target                       |
| `server_port`           | `server.port`                    | omitted when 0                          |
| `network_transport`     | `network.transport`              | tcp / udp / icmp                        |
| `network_protocol_name` | `network.protocol.name`          | http / dns / …                          |
| `start_time_unix_nano`  | span/metric start timestamp      | OTel nanosecond epoch                   |
| `duration_nano`         | duration                         | nanoseconds                             |
| `success`               | outcome                          | → `netctl_probe_success` (1/0)          |
| `error_message`         | `error.message` (when failing)   |                                         |
| `metrics{}`             | metric data points               | name → value (see TSDB below)           |
| `attributes{}`          | additional OTel-convention attrs | canary-supplied (`network.*`, `server.*`, `client.*`) |

There is no standard OTel tenancy attribute, so tenant/agent identity uses the
`netctl.*` namespace; everything else follows the OTel specification.

## TSDB metric/label schema (Prometheus / VictoriaMetrics)

The consumer (`internal/tsdb`) turns each Result into time series:

- `netctl_probe_success` — gauge, 1 on success / 0 on failure.
- `netctl_probe_duration_seconds` — gauge, the probe duration.
- `netctl_probe_<metric>` — one gauge per entry in `metrics{}` (the metric key is
  sanitized to a valid Prometheus name, e.g. `rtt.avg.ms` → `rtt_avg_ms`).

**Labels** (cardinality-bounded on purpose): `tenant_id`, `agent_id`,
`canary_type`, `server_address`. `tenant_id` is a label (pooled mode); siloed mode
uses per-tenant series. High-cardinality per-hop/per-target detail belongs in
ClickHouse, not as metric labels (CLAUDE.md / S6 watch-out).
