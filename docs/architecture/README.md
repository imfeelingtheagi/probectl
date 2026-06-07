# docs/architecture/ — the authoritative spec set (U-088)

| Document | Role | Status |
|---|---|---|
| [`../architecture.md`](../architecture.md) | System architecture — the live, kept-current view | authoritative, in-repo |
| [`PRD-v0.5.md`](PRD-v0.5.md) | Product requirements (v0.5, architecture-review pass) | **imported verbatim 2026-06-07** from the founder's working folder; updates land here, not outside |
| `../../CLAUDE.md` | Engineering contract (guardrails, conventions, editions) | authoritative, in-repo |

Provenance + caveats for the imported PRD:

- Imported **verbatim** — no content edits at import time. If this file and
  `CLAUDE.md` conflict, flag it (CLAUDE.md §0 rule) rather than silently
  resolving.
- The PRD may reference founder-private working artifacts (sprint plans,
  marketing drafts, cost models). Those references are **unverifiable from
  this repo** — treat them as context, not as committed claims. Everything
  load-bearing for diligence is in-repo (docs/, the register remediations).
- Numeric SLO targets remain **PROVISIONAL pending sign-off**
  (`docs/scale-gate.md`); positioning claims follow the scoped wording in
  `docs/otlp.md` (U-020) and `README.md`.

Keeping it current: spec changes ship in the same PR as the code that
implements them (CLAUDE.md §8 — docs updated with every change); PRD
revisions bump the file version here.
