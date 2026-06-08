# SOC 2 control mapping — skeleton (U-065)

> **DRAFT — control-mapping STARTED, not an audit.** Maps Trust Services
> Criteria (Security/Common Criteria) to probectl's existing technical
> controls with code/CI evidence per control, plus honest organizational
> gaps (solo-founder reality). An auditor scopes the actual examination;
> note that in operator-hosted deployments many operational criteria are
> the CUSTOMER's controls — column 4 says whose burden each one is.

| TSC | Control as implemented | Evidence (code / CI / doc) | Owner |
|---|---|---|---|
| CC1.x Control environment | Engineering guardrails are written, versioned, and CI-enforced (CLAUDE.md §7); single-maintainer review discipline via conventional commits + branch protection | CLAUDE.md; `docs/ops/branch-protection.md` | Vendor (org maturity gap noted below) |
| CC2.x Communication | Security policy + reporting channel; per-deployment `security.txt`; IR plan with notification flow | `SECURITY.md`; `docs/security/incident-response.md` | Vendor + Customer |
| CC3.x Risk assessment | Threat model versioned in-repo, evidence-linked, gap-honest; unified diligence register drives remediation | `docs/security/threat-model.md` | Vendor |
| CC4.x Monitoring of controls | Controls run as CI gates on every push: isolation suite, helm hardening, action pins, crypto-import guard, coverage floors, scheduled vuln scans | `.github/workflows/ci.yml`, `security-scan.yml`; `scripts/check_*.sh` | Vendor |
| CC5.x Control activities | Fail-closed defaults policy (auth, TLS, consent gates) | U-001, U-010, C7/U-003 implementations | Vendor |
| CC6.1 Logical access | RBAC/ABAC per route; tenant-first enforcement incl. AI/MCP; SSO (OIDC) + SCIM | `internal/control/v1.go` route permissions; `internal/auth`, `internal/scim` | Customer operates; vendor supplies |
| CC6.1 (tenant isolation) | RLS at storage layer + CH row policies + query-proxy refusal of unscoped reads; **cross-tenant CI gate on every push** | `cross-tenant-isolation` job; U-025/U-026 | Vendor (mechanism) |
| CC6.2/6.3 Provisioning & removal | Tenant lifecycle with attested, store-by-store erasure; SCIM deprovisioning | `internal/tenantlife` (U-027) | Customer operates |
| CC6.6 Boundary protection | TLS on every listener; mTLS+SPIFFE agent transport with trust-domain pin; bus TLS with plaintext refusal; OTLP authenticated+tenant-scoped | guardrail §7.4/§7.12; U-010/U-011; `docs/otlp.md` | Vendor (defaults) |
| CC6.7 Transmission/at-rest | Hardened TLS client policy (U-036); envelope encryption for secrets; BYOK/per-tenant keys (EE); at-rest posture documented honestly | `internal/crypto`; `docs/audit/residency-findings.md` (U-042) | Shared |
| CC6.8 Malicious software | Signed releases (cosign keyless) + SBOM; digest-pinned images; SHA-pinned actions; BPF object digests verified pre-load | C6/C11/U-007/U-014; `docs/ops/verify-artifacts.md` | Vendor |
| CC7.1 Vulnerability mgmt | scheduled scans (govulncheck/trivy per `security-scan.yml`) + on every PR; SHA/digest/hash-pinned deps enforced by CI pin gates; fuzz smoke on parsers | U-069/C12 | Vendor |
| CC7.2/7.3 Anomaly detection & incident handling | probectl-observes-probectl metrics; severity matrix, SLAs, evidence preservation | `docs/security/incident-response.md` (U-066) | Shared |
| CC7.4/7.5 Incident recovery | Backup/restore tooling with a **CI-executed drill**; timed failover drill; DR runbook | U-030 `docs/ops/backup-restore.md`; U-053 `docs/ops/dr.md` | Customer operates; vendor proves mechanism |
| CC8.1 Change management | Conventional commits; release refuses red CI; expand/contract migration gate; buf breaking gate | U-083; `migration-gate`, `proto` jobs | Vendor |
| CC9.x Risk mitigation / vendors | No-subprocessor default; customer-elected AI providers disclosed | `subprocessors.md`; `docs/ai-egress.md` | Shared |
| A1.x Availability | HA/multi-region model + drills; noisy-neighbor fairness + SLO gate | `docs/multi-region.md`; U-053/U-055 | Customer operates |
| C1.x Confidentiality | Tenant isolation (above); L7 capture consent + redaction; AI egress triple gate + C8 redaction | U-003/C13; C7/C8/D9 | Vendor (defaults) |
| P-series Privacy | Out of scope for this skeleton; DPA template covers data categories | `dpa-template.md` | Counsel |

## Honest organizational gaps (an auditor will ask)

- **Separation of duties:** solo founder — compensating controls are
  branch protection, CI gates that the author cannot bypass without trace,
  and the signed/WORM audit trail. Formal SoD arrives with headcount.
- **Formal policies** (HR, vendor mgmt, BCP as org documents) — not yet
  drafted; the technical halves above exist.
- **Type II evidence period** — controls are young; a readiness assessment
  should precede any audit window.

## Next steps

1. Counsel/auditor review of this mapping.
2. CAIQ self-assessment generated from this table (same evidence).
3. Pick an audit window after the org gaps close.
