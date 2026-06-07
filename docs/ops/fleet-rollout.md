# Staged fleet rollout (U-031)

How a fleet of probectl agents moves to a new version: in waves, from
signed artifacts, with the registry verifying every wave and any failure
halting the train.

## The model — and what is deliberately absent

**There is no agent self-update channel.** An agent never fetches or execs
new code; update authority stays with the operator's orchestrator (Helm /
`install.sh` / your config management), exactly like any other workload
(preserved strength ST-04 — a self-update channel is a fleet-wide RCE
primitive). The control plane's role is to **plan** waves from the agent
registry and **verify** each wave back out of it — never to push bits.

The engine is `internal/agent/rollout.go` (the seam the CLI/console wire
into): deterministic waves from the lifecycle cohorts (canary ≈5% → early
≈20% → main), fixed at plan time by a stable hash of the agent id, agents
already on target excluded. Planning **fails closed** on three things: an
artifact without a recorded signature verification (C6), a target outside
the N/N-1 version-skew window against the control plane (the skew gate
stays green — `internal/lifecycle.Policy.Check`), and an empty/up-to-date
fleet.

## Operator flow

**0. Verify the artifact (C6) — and record it.** Per
[`verify-artifacts.md`](verify-artifacts.md):

```sh
cosign verify ghcr.io/imfeelingtheagi/probectl-ebpf-agent@sha256:<digest> \
  --certificate-identity-regexp 'github.com/imfeelingtheagi/probectl' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

The plan requires the exact **digest**, the verification **method**, and
**who verified** — an unattested artifact refuses to plan. VM binaries:
`cosign verify-blob` with the release checksums per the same doc.

**1. Plan.** Snapshot the fleet from the registry (`GET /v1/agents`) and
plan: waves render as e.g. `canary[3]=pending early[11]=pending
main[46]=pending`. The wave membership (exact agent ids) is the
orchestrator's worklist.

**2. Advance one wave.** `Advance` releases exactly one wave (never two,
never out of order) and starts its verify window (default 15m). Apply it
with your orchestrator, by digest:

- **Kubernetes** (the U-016 chart): `helm upgrade probectl-agent
  deploy/helm/probectl-agent --reuse-values --set
  image.tag="<version>@sha256:<digest>"` — scope waves to node sets via
  `nodeSelector`/separate releases per ring.
- **VMs**: `sudo deploy/agent/install.sh ./probectl-ebpf-agent-<version>`
  on the wave's hosts (after `cosign verify-blob`).

**3. Verify from the registry — the agents themselves are the evidence.**
Every wave member must re-register on the **target version** with a
**fresh heartbeat** (≤5m stale). All good → the wave completes and the next
can be advanced. Stragglers inside the window: keep waiting, re-verify.

**4. Halt-on-error.** Past the window, ANY straggler — still on the old
version, reporting nothing, or vanished from the registry ("upgraded then
went dark") — **halts the entire rollout**, naming the agents. A halted
rollout exposes no current wave, refuses Advance and Verify, and never
resumes on its own.

**5. Resume is explicit.** After remediation (roll the node back, replace
it, fix the artifact), `Resume` with a written remediation note returns the
failed wave to applying with a fresh window. That note is the audit trail
of what went wrong mid-rollout.

## Properties worth relying on

| Property | Where it is enforced |
|---|---|
| Signed artifacts only | plan refuses without digest+method+verifier (C6); deploys are by digest |
| No self-update | nothing in the agent fetches code; orchestrator-only (ST-04) |
| Skew gate stays green | plan refuses targets outside N/N-1 vs the control plane |
| Deterministic waves | stable-hash cohorts, fixed at plan time, sorted membership |
| No overlap / no skipping | Advance refuses while a wave is unverified |
| Halt-on-error | registry-verified; stragglers/dark agents past the window freeze the train |
| Mid-rollout safety | N/N-1 means old + new agents coexist on the bus throughout |

Rollback is the same machine pointed backwards: plan a rollout to the
previous (still-signed) version — the skew window that lets waves coexist
forwards lets them coexist backwards.
