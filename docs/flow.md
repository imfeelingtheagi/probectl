# Flow analytics (S38, F17) — NetFlow / IPFIX / sFlow

The flow plane is probectl's passive traffic-analytics layer: network devices
export flow records, `probectl-flow-agent` decodes them into one normalized,
tenant-bound schema, and the control plane enriches and persists them to
ClickHouse, serving the top-talkers / capacity / anomaly views over `/v1/flows/*`.

```mermaid
flowchart LR
  R[routers / switches<br/>NetFlow v5/v9 · IPFIX · sFlow v5] -- UDP --> C[probectl-flow-agent<br/>decode · templates · sampling]
  C -- "probectl.flow.events (FlowBatch, tenant-keyed)" --> B[(bus)]
  B --> P[control plane FlowConsumer<br/>ASN/geo enrich (opt-in, S15)]
  P --> S[(ClickHouse probectl_flows<br/>tenant-first partition + ORDER BY)]
  S --> API["/v1/flows/top · /v1/flows/capacity · /v1/flows/anomalies"]
```

## Wire protocols

| Protocol   | Port (default) | Notes                                                                |
| ---------- | -------------- | -------------------------------------------------------------------- |
| NetFlow v5 | `:2055`        | fixed 48-byte records; header sampling interval; v4 only by design   |
| NetFlow v9 | `:2055`        | template-based (RFC 3954); shares the socket with v5 (version-sniffed) |
| IPFIX      | `:4739`        | RFC 7011: enterprise fields + variable-length fields skipped by length |
| sFlow v5   | `:6343`        | raw-packet-header samples parsed Eth/802.1Q/IPv4/IPv6/TCP/UDP        |

**Templates (v9/IPFIX).** Data sets arriving before their template are counted
as template misses and dropped — exporters re-send templates on their refresh
cycle, so the gap self-heals. Template state is per (exporter, observation
domain), TTL'd (default 30m) and size-capped (default 4096) so a hostile or
misbehaving exporter cannot grow collector memory.

**Sampling.** Records keep the RAW exported counters plus the sampling rate —
from the v5 header, v9/IPFIX options templates (IE 34 / 50 / 305), an inline
element, or the sFlow sample itself — and carry pre-scaled estimates
(`bytes_scaled = bytes x rate`) that all analytics use. Unsampled = rate 1.

## Security posture

Flow export protocols are **plaintext UDP with no authentication — by protocol
design** (the industry reality GoFlow2/Kentik/Akvorado all share). probectl's
posture (CLAUDE.md §7 guardrail 12):

- every datagram is **untrusted input**: decoders are pure and bounds-checked,
  record counts and template state are capped, malformed input is counted and
  dropped — never a panic;
- the tenant on every record comes from the **collector's own tenant binding**
  (its config/identity), never from anything the datagram claims;
- deploy the collector **adjacent to its exporters** (management network /
  same site) so flow datagrams never cross untrusted segments;
- everything downstream of the collector rides the standard authenticated
  paths (bus → control plane → ClickHouse over HTTP(S)).

## Enrichment (S15)

The control-plane consumer fills `source/destination.as.number`, the AS organization
name, and the country code via the opendata enricher — **opt-in**
(`PROBECTL_FLOW_ENRICH_ASN=true`), because Team Cymru lookups are outbound DNS
and probectl never phones home by default. Device-asserted AS numbers (NetFlow
v5/v9/IPFIX export them) always pass through and are never overridden;
enrichment only fills blanks, is cached, and degrades gracefully.

## Storage

ClickHouse table `probectl_flows` (created idempotently by the flow store):

- `PARTITION BY (tenant_id, toYYYYMMDD(ts))` and
  `ORDER BY (tenant_id, ts, exporter, src_addr, dst_addr)` — the tenant leads
  both, so tenant-scoped reads prune at the storage layer and a day's parts
  stay bounded at NetFlow volumes;
- `PROBECTL_FLOW_RETENTION_DAYS=N` applies `TTL toDateTime(ts) + INTERVAL N DAY DELETE`;
- `memory` mode (default) serves the same Store contract for the lightweight
  deploy and is the reference implementation the SQL must agree with.

## Query API (tenant-scoped, `flow.read`)

```text
GET /v1/flows/top?by=src|dst|pair|src_asn|dst_asn&window=1h&limit=10
GET /v1/flows/capacity?exporter=&direction=in|out&window=1h&bucket=3m
GET /v1/flows/anomalies?window=1h&bucket=3m&k=3&min_bps=1000
```

Top-talkers aggregates sampling-corrected bytes/packets/flow-counts by the
requested key. Capacity buckets per-(exporter, interface) throughput into
bps/pps. Anomalies compares each interface's latest bucket against the mean +
`k`·stddev of its own preceding buckets (shared detector across both store
backends), flagging departures above `min_bps`.

## Operations

- The collector logs a stats line every 60s: packets, records, decode errors,
  template misses, queue drops, emit errors, cached templates.
- High-volume tuning: raise `PROBECTL_FLOW_READ_BUFFER_BYTES` (kernel burst
  absorption) and `PROBECTL_FLOW_WORKERS` first; queue overflow drops are
  counted rather than back-pressuring the socket (back-pressure on UDP just
  moves the drop into the kernel).
- Decoder throughput is gated in CI (`TestHighVolumeDecode`, 50k records/s
  floor; measured ~7M records/s on a dev laptop).

## Example

```bash
# collector (near the devices)
PROBECTL_FLOW_TENANT=t-acme PROBECTL_FLOW_BUS_MODE=kafka \
PROBECTL_FLOW_BUS_BROKERS=localhost:9092 ./bin/probectl-flow-agent

# point a device at it (Cisco-style)
# flow exporter EXP destination <collector-ip> transport udp 2055

# ask the API
curl -s "https://localhost:8443/v1/flows/top?by=src_asn&window=15m&limit=5"
```

See `deploy/agent/probectl-flow-agent.example.yml` for the YAML form of every
key, and `docs/configuration.md` for the full reference.
