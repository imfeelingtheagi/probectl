# Voice/RTP tests (S47c, F21)

The `voice` canary scores VoIP path quality the way a phone would
experience it: it streams REAL RTP packets (version-2 header, codec payload
type, codec cadence and payload size) to a reflecting target and computes
**MOS** (ITU-T G.107 E-model), **RFC 3550 jitter**, and **packet loss** from
the echoes.

## Creating a test

```json
POST /v1/tests
{"name": "voip to pbx", "type": "voice", "target": "pbx.acme.example:5004",
 "interval_seconds": 60, "timeout_seconds": 3,
 "params": {"codec": "g711", "duration_seconds": "3", "dscp": "46"}}
```

| Param | Default | Meaning |
|---|---|---|
| `codec` | `g711` | codec preset: `g711` (PCMU, 50 pps, 160 B) or `g729` (50 pps, 20 B) |
| `duration_seconds` | `3` | simulated call length (1–10 s; 50 packets/s) |
| `dscp` | `46` | DSCP marking (EF by convention for voice) |

The target must **echo UDP datagrams** (a probectl agent responder or any
UDP echo service) — a non-echoing target reports as 100% loss and an
explicit *"voice path unmeasurable"* failure, never a fabricated score.

## The model (and its honesty)

This is the **simplified, transport-only E-model** — stated on every result
(`voice.model = itu-t-g107-e-model-simplified`), because the S47c watch-out
is exactly that MOS models get over-trusted:

- `R = 93.2 − Id − Ie,eff` (default-parameter R0; no advantage factor, no
  simultaneous-impairment terms).
- **Delay** `Id = 0.024·d + 0.11·(d−177.3)·H(d−177.3)` — the G.107 delay
  curve with its 177.3 ms knee. One-way delay `d` is **estimated** as
  RTT/2 + codec delay + a modeled jitter buffer (2× jitter, 40–120 ms),
  and the result says so (`voice.one_way_estimate`).
- **Loss** `Ie,eff = Ie + (95 − Ie)·Ppl/(Ppl + Bpl)` — G.113 codec
  parameters (g711: Ie 0 / Bpl 25.1; g729: Ie 11 / Bpl 19), random-loss
  assumption.
- **MOS** from R via the G.107 Annex-B cubic (1.0 floor, 4.5 ceiling).
- **Jitter** is the RFC 3550 §6.4.1 interarrival estimator over received
  packets (`J += (|D| − J)/16`), reported as `voice.jitter.ms`.

Narrowband codecs only: the 93.2 baseline IS the narrowband model;
wideband (G.722/Opus) needs the wideband E-model variant — out of scope,
documented rather than approximated.

## Result fields (the contract)

Metrics: `voice.mos`, `voice.r_factor`, `voice.jitter.ms`,
`voice.one_way.ms`, `voice.loss.pct`, plus the standard latency family
(`rtt.min/avg/max/stddev.ms`, `loss.ratio`, `packets.sent/received`).
Attributes: `voice.codec`, `voice.model`, `voice.method`
(`rtp-udp-echo-reflection`), `voice.jitter_buffer_ms`,
`voice.one_way_estimate`.

Results render in the Targets & Tests result views (S-FE5) with MOS up
front (≥4.0 good / ≥3.6 fair / below = poor) and the model named in the
detail — a computed score is never presented as a measured listening test.

Out of scope by design: SIP call control (no signaling — RTP transport
quality only), wideband codecs, and owning reflector infrastructure
(agent-to-agent reflection rides the existing a2a responder).
