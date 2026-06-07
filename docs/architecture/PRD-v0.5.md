# probectl — Product Requirements Document (v0.5, architecture-review / UX-foundation pass)

| | |
|---|---|
| **Product** | probectl |
| **Owner** | Shankar (solo founder) |
| **Status** | Draft v0.5 — architecture-review / UX-foundation pass (building in private) |
| **Last Updated** | May 31, 2026 |
| **License** | Source-available, specific license TBD (deferred; MSP resale will require a commercial-licensing path — §4.6/§10.3) |
| **Lineage** | Evolution + rebrand of "NetVantage" into the `ctl` family (sibling to trustctl) |

> **One-sentence vision:** probectl is a self-hosted, source-available, **multi-tenant** network observability platform that unifies active/synthetic testing, BGP/routing intelligence, flow analytics, device telemetry, and eBPF host visibility into a single OpenTelemetry-native control plane — with an AI assistant that performs root-cause analysis across all planes, native security/threat detection, change-aware topology, cost and SLO intelligence, enriched by maximal public internet + threat-intel data — deployable as a sovereign single-tenant install *or* operated by an MSP/provider that resells hard-isolated, white-labeled tenants, with your telemetry never leaving the operator's network.

> **Version history.** v0.1 — core five-plane platform + open data. v0.2 — F500 enterprise-readiness layer (identity/RBAC/audit/FIPS/SIEM/governance/scale/lifecycle). v0.3 — differentiation moat (security/threat layer, change intelligence + topology graph, FinOps, SLO engine, segmentation validation, agentic remediation + AI authoring). **v0.4 (this) — multi-tenancy / MSP-resale pass:** tenancy is promoted to a **ground-up, first-class architectural property** (it is no longer a deferred Phase-4 "managed offering"). Adds a hard tenant-isolation foundation, a provider/MSP management plane, selectable pooled/siloed/hybrid isolation models, per-tenant metering & billing export, white-label/per-tenant branding, per-tenant key isolation/BYOK, per-tenant data export & verifiable deletion, and tenant fairness/noisy-neighbor isolation — new requirements **F50–F57**, a new MSP persona (**P7**), and a new **Epic U**. The single-tenant sovereign deployment is preserved as the one-tenant degenerate case. **v0.5 (this) — architecture-review / UX-foundation pass:** a pre-build architecture & product review drove six structural changes (no new feature scope, but sequencing + a UX foundation that protect best-in-class delivery): (1) a **frontend foundation & design-system** is now established early (M2) as the basis for every UI surface — themeable tokens, component library, app shell + command palette, WCAG 2.2 AA — making **F54 a foundation/full split** (tokens foundational; per-tenant branding becomes a token override, not a per-screen retrofit); (2) the **topology foundation is sequenced before the AI layer** (M9, not M10 — the prior order was a dependency inversion); (3) **OTel semantic conventions are designed into the result/event schema from M2** so the P2 OTLP layer *exposes* rather than retrofits; (4) a **load/perf harness + scale baseline is left-shifted to M6** and an **eBPF feasibility spike precedes M7**; (5) the **hero visuals and the AI surface are design-led, iterated surfaces**, not one-shot; (6) **milestone integration gates** verify cross-plane correlation at each milestone boundary.

> **F500-viability principle (carried):** "F500-viable" means the architecture does not *preclude* enterprise (expensive-to-retrofit foundations designed in now) plus a credible roadmap to the enterprise feature set — much of it in an **Enterprise Edition** track, with operational scale and 24×7 support supplied by the eventual acquirer.

> **Tenancy principle [v0.4]:** **One codebase, two operating modes.** (1) **Sovereign single-tenant** — a regulated/air-gapped org self-hosts; the deployment *is* the tenant boundary; the provider plane is transparent/disabled. (2) **Multi-tenant / provider** — an MSP (or internal platform team) self-hosts once and resells/serves many hard-isolated tenants. Both run the same binaries. Tenancy is the **outermost scope and security boundary** on every record, agent, query, metric, event, and object — designed in from M1, never retrofitted. "Sovereign" is preserved in both modes: in provider mode, telemetry lives with the *operator the customer chose* (a regional/sector/sovereign MSP), never in a probectl-the-vendor public SaaS.

---

## 1. Executive Summary

### 1.1 Product Vision

**One-sentence description.** A self-hosted, source-available, **multi-tenant** "ThousandEyes-plus" — the entire ThousandEyes software layer (canaries, path visualization, BGP, internet-health) plus four more observability planes, a native security/threat layer, change-aware topology, and cost/SLO intelligence — unified, OTel-native, AI-driven, deployable sovereign-single-tenant *or* MSP-multi-tenant, and built to the trust/scale bar a Fortune 500 requires.

**Target user and use case.** Two coupled audiences:
1. **End users** — Platform/SRE and network engineers at mid-to-large and Fortune 500 organizations, especially regulated, sovereign, or air-gapped environments (finance, healthcare, government, defense, critical infrastructure), who need ThousandEyes-class visibility but cannot or will not ship telemetry to a public SaaS. Use case: "Is it the network, the app, the route, the config change, or a threat? Why is X slow for office Y, and is anything malicious?" answered in one tool, on their (or their provider's) own infrastructure.
2. **Operators / resellers [v0.4]** — **MSPs and managed-network/observability providers** (and large internal platform teams running probectl as an internal multi-tenant service) who want to **self-host once and serve/resell many hard-isolated, white-labeled customer tenants**, billing per usage, onboarding a new tenant in minutes, under their own brand — including regional and sovereign-cloud providers who can offer a self-hosted alternative to the hyperscale-SaaS incumbents.

**Key differentiators (the "no-brainer complete package").**
1. **Self-hosted / data-sovereign** — fully on-prem, air-gap capable; data residency by construction; no third-party data-processor in the vendor-risk path. In provider mode, data stays with the customer's chosen MSP, not a public SaaS.
2. **Multi-tenant by construction [v0.4]** — hard tenant isolation as the outermost boundary; a provider/MSP management plane; pooled/siloed/hybrid isolation; per-tenant metering, billing, branding, keys, and verifiable deletion — so a partner can self-host-and-resell. The sovereign single-tenant install is the one-tenant case of the same system.
3. **Five planes, one pane** — active/synthetic + routing/BGP + flow + device + eBPF host, instead of stitching ThousandEyes + Kentik + LibreNMS + Hubble.
4. **OpenTelemetry-native** — network signals correlate with application traces/logs; standards-aligned and future-proof.
5. **AI-native RCA + agentic** — natural-language query, automated cross-plane root-cause, AI test authoring, and guarded remediation, via a built-in assistant *and* a first-class MCP server, with a pluggable model adapter (cloud or local LLM for sovereignty); tenant-scoped.
6. **Security-native** — TLS/cert posture + NDR-lite threat detection fed by open threat-intel, on the traffic probectl already inspects (the trustctl-adjacent unfair advantage).
7. **Change-aware** — a live, telemetry-fed topology graph plus change-to-incident correlation that answers "what changed."
8. **Maximal open-data + threat-intel ingestion** — free internet-wide and threat context without owning a fleet.
9. **Enterprise trust by design** — identity, RBAC/ABAC, audit, FIPS-mode crypto, SIEM/ITSM integration, governance, cost, and SLO/business-impact built in.

**Success definition.** Not near-term revenue. Success = open-source adoption (stars, deployments, agents, contributors), parity with the ThousandEyes software layer, the differentiated AI/OTel/sovereign/security/change/cost story, **MSP/provider adoption (tenants served, resellers onboarded)**, demonstrated F500-readiness, and inbound strategic interest. Build-to-acquire with an open-core flywheel; the multi-tenant **provider plane is also a natural commercial-tier monetization lever and an MSP go-to-market channel**.

### 1.2 Strategic Alignment

**Business objectives.** Build a defensible, acquirable network-observability asset at the intersection of funded trends — OpenTelemetry standardization, AI-driven observability/AIOps, data-sovereignty-driven on-prem demand — *and* widen the addressable acquirer set by being security-native (NDR-adjacent), FinOps-aware, SLO-aware, **and MSP/provider-ready (multi-tenant, white-label, metered)**. The multi-tenant capability adds both a **new acquirer archetype (MSP/RMM/PSA platforms)** and a **distribution channel (MSPs reselling probectl)**. Mirror the trustctl playbook in an adjacent, larger market.

**User problems.** The "green dashboards, unhappy users" seam — fragmented, single-plane, mostly-SaaS tools where no one sees the chain end to end, no one knows *what changed*, and security/cost/SLO live in yet more separate tools — while sovereignty + procurement gating lock regulated F500s out of incumbents entirely. **For MSPs [v0.4]:** the multi-tenant network-monitoring tools they can resell (Auvik, Domotz, LogicMonitor, RMM-bundled monitors) are SaaS, shallow on the five planes, not sovereign, and not AI-native — and offer no self-hosted/white-label/sovereign option for the MSP's regulated or cost-sensitive clients.

**Market opportunity and timing.** Broad observability markets are multi-billion-dollar with double-digit CAGRs (analyst definitions vary; ~$2.9B–$28B 2024–2025 base depending on scope; directional). Adjacent expansions probectl now touches — NDR/network security, AIOps, FinOps, DEM, **and the MSP/managed-services channel** — each add buyers. The defensible wedge remains the underserved intersection: self-hosted, **multi-tenant**, ThousandEyes-class visibility, unified, AI/security/change/cost-aware. Timing: OTel is de-facto standard; eBPF is the default data plane; sovereignty pushes on-prem; AI-in-observability and AI-SRE are the fastest-growing slices; and **MSPs are actively seeking sovereign/white-label alternatives** to SaaS monitoring they cannot tailor or run in-region.

**Competitive advantage.** Incumbents are SaaS-first and single-plane (ThousandEyes/Catchpoint, Kentik, LibreNMS/Zabbix, Hubble/Pixie/Coroot/groundcover, Datadog/Grafana/Dynatrace). The closest AI-for-network player (Selector AI) is SaaS and AIOps-broad; the self-hosted-eBPF players (Coroot/groundcover) are K8s-scoped and not internet-aware; Kentik is paid SaaS. **MSP-oriented monitors [v0.4]** (Auvik, Domotz, LogicMonitor, Obkio, RMM/PSA bundles from ConnectWise/N-able/Kaseya/Datto) are multi-tenant but SaaS-only, network-shallow, not sovereign, and not AI/OTel-native. probectl's edge is the *combination* — **sovereign + multi-tenant/white-label** + unified + OTel-native + AI/agentic + security-native + change-aware + cost/SLO-aware + open-data-fed — with UX engineered to beat the clunky incumbents.

### 1.3 Resource Requirements

**Effort.** Multi-phase, solo-feasible MVP. Phase 1 = one person; Phase 2 = +1–2 contributors; Phase 3 + Enterprise Edition = contributors, partnership, or funding (and increasingly the acquirer). **Tenancy is sequenced to stay solo-feasible [v0.4]:** the *ground-up* part (tenant as outermost scope in the data model + context propagation + pooled logical isolation) is M1 schema-and-discipline work, not a new product surface; the heavier provider surfaces (operator console, siloed/hybrid isolation, metering/billing, white-label, per-tenant BYOK) are Phase 2 with contributor help. The v0.3 differentiators remain sequenced cheap-substrate-first.

**Timeline (indicative).** Phase 1 (MVP / ThousandEyes core + enterprise + **tenancy foundations**) ~0–6 mo; Phase 2 (eBPF + OTel + AI/MCP + security/TLS + change-intelligence + AI-authoring + first enterprise features + **provider plane / isolation models / metering / white-label**) ~6–12 mo; Phase 3 (flow + device + NDR + FinOps + SLO + segmentation, full package) ~12–18 mo; Enterprise Edition + (optional first-party) managed offering 18+ mo. Flexes around newborn-care priorities; Phase 1 sized solo-achievable.

**Team and skills.** Founder (Go, networking, PKI/security/CISSP, product). Future: Go/eBPF engineer, frontend/design contributor, BGP/data-engineering contributor, detection-engineering, FinOps/SLO, **and a multi-tenancy/billing-integration contributor**; security/compliance + support functions via Enterprise Edition or acquirer.

**Budget.** Minimal cash burn (self-hosted infra; open data + most threat-intel feeds free; Florida LLC). Primary cost is founder time; opportunistic contractor spend in Phase 2+; certification/audit spend deferred to Enterprise Edition / acquirer. **MSP resale requires a commercial-licensing path** (the one piece that turns the open core into reseller-grade terms — flagged, deferred).

### 1.4 Differentiation Moat (the acquirer story)

probectl's defensibility is not any single feature; it is the **compounding intersection** of capabilities that no competitor combines, several of which are *cheap for probectl and expensive for everyone else* because probectl already inspects the underlying traffic — now extended with a **multi-tenant/MSP dimension** that adds both buyers and a channel.

**The base moat (v0.1–v0.2).** The union of five observability planes, self-hosted/sovereign, OTel-native, AI-native, open-data-fed, and enterprise-ready. No competitor occupies this intersection: ThousandEyes/Cisco can't be self-hosted; Coroot/groundcover aren't internet-aware; Kentik is paid SaaS; Selector is SaaS AIOps; the OSS observability crowd (SigNoz/OpenObserve/Coroot) does app/infra, not network active-testing/path/BGP/internet measurement.

**The v0.3 moat extensions — each widens the acquirer set:**
- **Security-native (NDR + TLS/cert posture).** probectl already sees flow + eBPF + DNS + BGP + TLS — the exact substrate self-hosted NDR vendors (ExtraHop, Corelight, Darktrace, Cisco Secure Network Analytics) charge a premium for. *Roughly doubles the addressable acquirer set to include security players; the one thing only a CISSP/PKI founder can credibly ship.*
- **Change-aware (live topology graph + change-to-incident correlation).** White space between static digital-twin vendors (Forward Networks, IP Fabric) and observability tools; it is what makes probectl's AI RCA *win* against the AI-SRE wave.
- **Cost-aware (FinOps).** Cloud egress + cross-AZ/region traffic is a large *unmonitored* F500 cost; surfacing dollars drags in a FinOps/finance buyer.
- **SLO + business-impact.** Network SLOs/error-budgets are rare; speaking in services, BUs, and error budgets makes probectl executive-grade.
- **Compliance-as-observability (segmentation validation).** Proving segmentation (PCI, zero-trust) from eBPF/flow is a sharp regulated-F500 hook and trustctl-adjacent.

**The v0.4 moat extension — multi-tenancy / MSP [new]:**
- **Sovereign multi-tenancy a SaaS incumbent structurally cannot match, and an MSP-channel a single-tenant tool cannot serve.** Because tenancy is built in (not a hosted-only feature), probectl is the platform an **MSP can self-host and resell under its own brand** — including in a **sovereign or in-region footprint** for clients who reject hyperscale SaaS. This adds a whole acquirer archetype (**MSP platforms / RMM / PSA / managed-monitoring vendors** — ConnectWise, N-able, Kaseya, Datto/Auvik-class, LogicMonitor-class) *and* a distribution channel (resellers), *and* a commercial-tier monetization lever (the provider/management plane as a paid edition). It compounds the rest: every plane, detection, and SLO an MSP resells is multiplied across its tenant base.

**Why it compounds.** Each plane and extension feeds the AI layer and topology graph, so every addition makes RCA, detection, and correlation better — a data-and-integration flywheel a single-plane competitor cannot replicate, a sovereign one a SaaS incumbent structurally cannot, **and a multi-tenant/white-label one the MSP channel actively pulls into more accounts**.

---

## 2. Problem Statement & Opportunity

### 2.1 Problem Definition

**Core pain.** Modern outages live in the seams between layers, and the cause may be a route change, an ISP brownout, a path asymmetry, a DNS regression, a saturated link, a kernel-level reset, **a config change nobody correlated, or an active threat** — while the teams who could see each are in different tools, and security, cost, and SLO live in still more. For Fortune 500s, sovereignty/air-gap rules lock out SaaS incumbents, and procurement/security gating blocks anything lacking enterprise identity/audit/compliance posture. **For MSPs [v0.4]:** they must monitor *many clients* without cross-client leakage, bill per client, brand the experience as their own, and ideally offer a sovereign/in-region option — but the multi-tenant monitors available to them are SaaS-only, network-shallow, and not self-hostable.

**Quantified impact.** MTTR is dominated by triage — *which* layer broke, *what* changed, *is it malicious* — not the fix. AI-assisted incident response is reported to cut MTTR substantially; the bottleneck is cross-layer + change + security correlation. At F500 scale, major-incident minutes carry direct revenue and regulatory cost, and unmonitored cloud egress carries direct dollar cost. **For MSPs, tenant onboarding time, per-tenant cost-to-serve, and the blast radius of a cross-tenant leak are the economics that make or break the service.**

**Evidence.** (a) The eBPF community's "green dashboards, suffering users" refrain. (b) Analysts cite OTel, AIOps/automated RCA, and sovereignty as demand catalysts; the AI-SRE field is converging on "what changed" and auto-remediation. (c) Tool sprawl spans observability + NDR + FinOps + SLO. (d) Sovereignty/compliance excludes regulated verticals from SaaS incumbents. (e) **The MSP/MSSP market runs on multi-tenant, white-label, per-seat/per-usage tooling, and is actively seeking sovereign and AI-native options their SaaS vendors don't provide.**

### 2.2 Opportunity Analysis

**Market.** Adjacent/superset markets are large and fast-growing (see §1.2). The v0.3 extensions add buyers across NDR, AIOps, FinOps, and DEM; **the v0.4 multi-tenancy adds the MSP/managed-services channel and its platform acquirers.** The defensible wedge remains the underserved intersection — self-hosted, multi-tenant, ThousandEyes-class visibility, unified, AI/security/change/cost-aware.

**Segment.** Fortune 500 and regulated/sovereignty-constrained enterprises (finance, healthcare, public sector, defense, telecom, critical infrastructure); **MSPs / managed-network & managed-security providers, and regional/sovereign-cloud providers reselling a managed probectl**; plus the self-hosting/OSS operations community (flywheel).

**Business impact.** Revenue intentionally deferred. Asset value is strategic and now broader: a parity-plus, sovereign, **multi-tenant**, AI-native, security-native, enterprise-ready platform is acquirable by network *and* security *and* observability *and* **MSP-platform** incumbents. Open-core monetization (Enterprise Edition, **the provider/management plane**, optional managed) is a later option; **MSP resale is an explicit channel.**

**Competitive gap.** The full intersection (§1.4) exists in no single product. Integration + sovereignty + the security/change/cost extensions + **resaleable multi-tenancy** is the moat.

### 2.3 Success Criteria

**Primary.** OSS adoption + ThousandEyes-software-layer parity coverage + demonstrated F500-readiness gates + **early MSP/provider adoption (resellers onboarded, tenants served, zero cross-tenant incidents).**

**Secondary.** Time-to-first-insight; AI-RCA/NL/authoring usage; MCP connections; threat detections actioned; cert-expiry incidents prevented; egress cost surfaced/saved; SLOs tracked; segmentation policies validated; **tenant-onboarding time; per-tenant cost-to-serve; white-label deployments; metered-usage accuracy**; contributor velocity; design-partner logos (esp. regulated/air-gapped F500 **and reference MSPs**).

**Expected behavior changes.** Teams retire 2–4 separate tools; first triage question answered by the AI assistant in seconds; security/procurement approve because data stays in-house; FinOps/SLO conversations move into probectl; **MSPs stand up a branded, multi-tenant managed-observability service on probectl and onboard clients in minutes instead of provisioning per-client stacks.**

**Anticipated outcomes.** A credible OSS project with a self-sustaining contributor base; a defensible "complete package" narrative across observability + security + cost + SLO + **multi-tenant/MSP**; demonstrable F500-readiness; **a nascent reseller channel**; and inbound strategic/acquisition interest from a *widened* acquirer set.

---

## 3. User Requirements & Stories

### 3.1 Primary Personas

**P1 — Priya, Platform/SRE Lead (regulated F500).** Owns reliability across network + app; can't send telemetry to SaaS. Goal: one self-hosted pane that says "network or app, and why," fast, with SLOs and change context. Success: MTTR down, fewer bridges, audits pass.

**P2 — Nikhil, Network Engineer.** Lives in path viz and BGP. Goal: ThousandEyes-grade path + BGP, self-hosted, fed by public collectors, with topology and change awareness. Success: detects/explains routing/path incidents before users complain.

**P3 — Sana, Security/SOC Analyst.** Watches BGP hijacks/leaks, anomalous egress/exfil, DNS anomalies, TLS/cert posture, threat-intel hits. Goal: network-threat signals correlated and exported to the SIEM, on-prem. Success: faster detection of network-layer security events without a separate NDR.

**P4 — Devon, Self-hoster / OSS Platform builder (flywheel).** Cost-sensitive, OSS-first, opinionated about UX. Goal: a free, beautiful, self-hostable tool that does what the expensive incumbents do. Success: deploys in minutes, contributes plugins/detections, evangelizes. *(Note: the MSP/reseller motion that previously lived in this persona is now P7.)*

**P5 — Maria, Enterprise Security / Compliance & Architecture (F500 gatekeeper).** Runs vendor-risk + architecture + procurement. Cares about SSO/SCIM, RBAC/audit, FIPS, SBOM/CVE posture, residency, SIEM, segmentation evidence, TLS posture, **and — when a provider is involved — tenant-isolation guarantees and per-tenant key control.** Goal: approve without exceptions. Success: passes the security questionnaire and architecture review first-pass.

**P6 — Raj, FinOps / Cloud-Cost Owner (F500).** Owns cloud spend efficiency. Cares about egress and cross-AZ/region traffic cost. Goal: see which services/flows drive network cost and act. Success: surfaces and reduces unmonitored egress spend.

**P7 — Lena, MSP Founder / Managed-Service Platform Owner (reseller). [new in v0.4]** Runs a managed network-observability (and/or managed-security) service for many SMB-to-enterprise clients; may target a regulated or in-region niche. Cares about **hard tenant isolation** (a leak across clients is existential), **fast tenant onboarding**, a **single operator console across all clients**, **per-tenant metering/billing** that feeds her PSA/billing, **white-labeling** under her brand (custom domain, logo, themed notifications), **per-tenant data residency / keys** for regulated clients, **noisy-neighbor protection**, and **commercial licensing she can legally resell on**. Goal: launch and scale a branded, profitable, sovereign-capable managed service on probectl. Success: onboards a new client tenant in minutes, bills accurately per usage, never leaks across tenants, and resells under her own brand — optionally self-hosted in-region.

**Secondary.** Acquirer integration teams; data-residency/privacy officers; OTel-centric platform teams; IT/SRE leadership owning MTTR + tool-consolidation budgets; **an MSP's own end-customer tenant-admins (covered by P1/P5 within their tenant via delegated admin).**

### 3.2 User Journey Mapping

**Current state (enterprise).** Alert → app dashboard green → NetOps pinged → ThousandEyes for path/BGP → Kentik for flow → SSH to device → maybe Hubble → *separately*, security checks NDR, FinOps checks a cost tool, nobody correlates the config change — 45+ minutes across many UIs and teams; sovereignty blocks some tools; new tools add weeks of procurement.

**Current state (MSP) [v0.4].** To add a client, the MSP either spins up a *separate per-client stack* (operationally expensive) or onboards them into a SaaS monitor that is network-shallow, can't be branded deeply, can't run in-region, and commingles data in ways regulated clients reject — and still leaves the MSP stitching multiple single-plane tools per client.

**Future state (probectl).** *Enterprise:* Alert → probectl single pane → unified incident timeline overlays all planes plus change and threat for the affected target/segment → ask the assistant "why is Salesforce slow for the Dallas office, did anything change, and is it malicious?" → cross-plane RCA with evidence and a *proposed guarded remediation* → engineer confirms. *MSP [v0.4]:* Lena opens the **provider console**, provisions a new **tenant** in minutes (pooled or siloed), points the client's agents at it, applies her **branding**, and the tenant's team logs in (SSO, scoped to *their* tenant only) to a fully isolated probectl; **usage meters per tenant** flow to her billing; she watches **fleet health across all tenants** from one operator view, while no tenant can see — or affect — another.

**Key touchpoints.** Agent install/onboarding; template + AI test authoring; path-viz + topology-graph hero views; the unified incident/change/threat timeline; the AI NL-query + remediation surface; MCP integration; alerts → on-call/ITSM; dashboards; security/threat view; cost/FinOps view; SLO/error-budget dashboards; compliance/segmentation evidence; enterprise admin console; **the provider/MSP management console (tenant lifecycle, fleet-across-tenants, metering/billing, branding).**

**Pain points / opportunity areas.** Time-to-first-insight, cross-plane + change + threat correlation, sovereignty, UX quality, enterprise approvability, cost/SLO visibility, **and (MSP) tenant-onboarding speed, cross-tenant isolation assurance, per-tenant cost-to-serve, and white-label fidelity.**

### 3.3 Core User Stories

*(Epics A–N carried from v0.2; epics O–T from v0.3; new Epic U added in v0.4.)*

**Epic A — Synthetic/Canary Monitoring (ThousandEyes parity).** US-A1 agent-to-server tests; US-A2 agent-to-agent; US-A3 HTTP + DNS; US-A4 (P2) browser/transaction synthetic.

**Epic B — Path Visualization (hero).** US-B1 interactive ECMP/MPLS-aware path viz; loss localized to hop/link with timing.

**Epic C — BGP / Routing Intelligence.** US-C1 BGP monitoring via public collectors (RouteViews + RIS), origin-change/hijack/leak, RPKI, correlation, alerting.

**Epic D — eBPF Host & L7 (P2).** US-D1 CNI-agnostic eBPF agent, L3–L7 flows + live service map, emits OTel.

**Epic E — Flow Analytics (P3).** US-E1 NetFlow/sFlow/IPFIX, top-talkers/capacity/anomalies.

**Epic F — Device / Telemetry (P3).** US-F1 SNMP + gNMI/OpenConfig device/interface health.

**Epic G — OTel-Native Unification (P2).** US-G1 OTLP ingest/export, OTel resource + semantic conventions, OBI.

**Epic H — AI RCA & MCP (P2).** US-H1 NL "why is X slow for Y?" with cross-plane evidence, pluggable model adapter (local option); US-H2 RBAC-scoped MCP server. *(All AI/MCP calls are tenant-scoped first, then RBAC-scoped — §5.1/§7.1.)*

**Epic I — Open-Data Enrichment (P1, marquee).** US-I1 ASN/geo/IXP, public BGP context, outage signals, optional RIPE Atlas scheduling.

**Epic J — Alerting, Dashboards, Incidents (P1).** US-J1 alert rules + unified cross-plane incident timeline.

**Epic K — Enterprise Identity & Access (P1 foundation → P2).** US-K1 SSO + SCIM; US-K2 fine-grained RBAC/ABAC + delegated admin scoped to **Tenant →** Orgs/Teams/Projects.

**Epic L — Enterprise Audit, Compliance & Governance (P1 foundation → P2/P3).** US-L1 tamper-evident audit → SIEM (tenant-scoped); US-L2 IP/PII handling, retention/erasure, residency.

**Epic M — Enterprise Operations & Lifecycle (P1 foundation → P2/P3).** US-M1 zero-downtime upgrades + DR; US-M2 agent fleet management; US-M3 FIPS-validated crypto.

**Epic N — Enterprise Integrations (P2/P3).** US-N1 on-call (PagerDuty/Opsgenie/Slack/Teams), ITSM (ServiceNow/Jira), CMDB, Grafana datasource, Prometheus federation, secrets (Vault/CyberArk + trustctl identities).

**Epic O — Security & Threat (P2 → P3).**
- **US-O1 (TLS/cert observability, P2):** cert expiry/chain/issuer/SAN; TLS version + cipher + (optional) JA3/JA3S; CT-log correlation; deprecated-protocol flags; one-click trustctl handoff; alerting.
- **US-O2 (NDR-lite threat detection, P3):** DNS exfil/DGA, beaconing, egress anomalies, Tor/known-bad-ASN, lateral-movement from the service map; confidence-scored; tunable + suppression; detection-as-code; SIEM export.
- **US-O3 (threat-intel enrichment, P2):** ingest open feeds; per-record provenance; graceful degradation; AUP/terms tracked.

**Epic P — Change Intelligence & Topology (P2 → P3).**
- **US-P1 (change-to-incident correlation, P2):** ingest config/route/BGP/IaC/deploy change events; change timeline; auto-surface candidate changes; feed AI RCA.
- **US-P2 (live topology graph, P2 foundation → P3 full):** graph from path/flow/eBPF/device; versioned; queryable; powers RCA + what-if.

**Epic Q — Cost / FinOps (P3).** US-Q1 egress + cross-AZ/region traffic → dollars; per-service/flow attribution; chatty-service detection; trends/alerts.

**Epic R — SLO & Business Impact (P3).** US-R1 OpenSLO-compatible SLOs on network signals; error budgets + burn-rate; service/BU mapping; exec dashboards.

**Epic S — Compliance & Segmentation (P3).** US-S1 declare expected segmentation; validate against eBPF/flow; flag violations; exportable PCI/NIST/zero-trust evidence.

**Epic T — Agentic Remediation & AI Authoring (P2 → P3/EE).**
- **US-T1 (guarded remediation, P3/EE):** proposed actions; human-gated by default; RBAC + **tenant**-scoped; dry-run/simulation; blast-radius limits; fully audited.
- **US-T2 (AI test authoring + auto-discovery, P2):** NL → canary config; auto-discover services/prefixes/endpoints and propose tests.

**Epic U — Multi-Tenancy & MSP Enablement (P1 foundation → P2). [new in v0.4]**
- **US-U1 (hard tenant isolation, P1 — foundational):** As Lena/Maria, I want every record, agent, query, metric, event, and object scoped to a tenant so that no tenant can ever read or infer another's data. *Acceptance:* `tenant_id` is the outermost scope on all entities; tenant context is propagated through API, bus, storage, and the AI/query layer; isolation is enforced **at the storage+query layer as defense-in-depth above RBAC** (not by application code alone); automated tests prove a tenant-scoped caller (incl. the AI/MCP layer) cannot retrieve cross-tenant data; cross-tenant access is impossible by construction in pooled mode.
- **US-U2 (provider / MSP management plane, P1 foundation → P2):** As Lena, I want a cross-tenant operator console to run my managed service. *Acceptance:* provision / onboard / suspend / offboard tenants; per-tenant configuration & lifecycle; fleet health and inventory **across all tenants** in one view; **provider operators are a distinct privilege domain** — managing a tenant does **not** grant read access to its telemetry; any access to tenant data requires explicit, time-bounded, tenant-consented, **fully audited break-glass elevation**.
- **US-U3 (isolation models — pooled / siloed / hybrid, P1 pooled → P2 siloed/hybrid):** As Lena/Maria, I want to choose how strongly each tenant is isolated. *Acceptance:* **pooled** (shared stores; logical `tenant_id` isolation) by default; **siloed** (per-tenant DB schema/instance, ClickHouse database, bus topic namespace, object-store prefix/bucket) for high-assurance tenants; **hybrid** (shared control plane, isolated data planes); selectable per-deployment **and per-tenant**; per-tenant data-residency targeting.
- **US-U4 (per-tenant metering, billing & quotas, P2):** As Lena, I want to meter and bill each tenant. *Acceptance:* meter agents, tests, ingest volume, stored events, AI calls, and retention per tenant; a usage API + export feed for the MSP's billing/PSA; per-tenant quotas and rate limits; showback/chargeback views.
- **US-U5 (white-label / per-tenant branding, P2):** As Lena, I want to resell under my brand. *Acceptance:* per-provider and per-tenant branding — logo, color theme, custom domain/subdomain, branded email notifications and login; tenant-scoped theming so the end customer sees the MSP's brand, not probectl's.
- **US-U6 (per-tenant lifecycle: export, residency & verifiable deletion, P2; + per-tenant keys P2/EE):** As Lena/Maria, I want clean tenant offboarding and key control. *Acceptance:* tenant-scoped data export (portability); **verifiable full deletion across all stores** (Postgres/ClickHouse/TSDB/object + documented backup policy); offboarding runbook; per-tenant retention/erasure (ties to F34); optional **per-tenant envelope-encryption keys / BYOK** (ties to F31/F56) so a tenant's data is cryptographically separable.

### 3.4 Story Format

As a [persona], I want [capability] so that [benefit]; Given/When/Then; acceptance criteria measurable and testable (see §4 / §8).

---

## 4. Functional Requirements

### 4.1 Core Features — Must Have (Phase 1 MVP: "ThousandEyes core + enterprise + tenancy foundations")

- **F1.** Canary agent (single Go binary, multi-arch, compiled-in plugins, store-and-forward). *(Tenant-bound: an agent belongs to exactly one tenant — F50.)*
- **F2.** Network tests (agent-to-server + agent-to-agent; ICMP/TCP/UDP; loss/latency/jitter; continuous monitoring).
- **F3.** Path visualization (ECMP/MPLS-aware; per-hop/link; merged topology; the hero UX).
- **F4.** HTTP server tests. **F5.** DNS tests.
- **F6.** BGP / routing monitoring (RouteViews + RIS; origin-change/hijack/leak; RPKI; correlation).
- **F7.** Open-data enrichment (ASN/geo/IXP, public BGP context, outage signals, optional RIPE Atlas). *(Open-data ingestion is shared infrastructure; enrichment is applied per-tenant — F50/F57.)*
- **F8.** Alerting (threshold + dynamic; channels; webhooks). **F9.** Dashboards + unified incident timeline.
- **F10.** Control plane + versioned REST API + gRPC (agents) + CLI/TUI. *(Tenant-aware: every request carries/derives tenant context — F50.)*
- **F22.** Enterprise identity foundation (SSO; RBAC-ready role model). *(Roles are scoped within a tenant; see F50/F51 for the provider-operator domain.)*
- **F23.** Audit foundation (immutable, tamper-evident; export hooks). *(Audit is tenant-scoped; provider-plane actions and break-glass are separately audited — F51.)*
- **F24.** Multi-org data model foundation (**Tenant →** Orgs → Teams → Projects). *(In v0.4 the hierarchy gains Tenant as the outermost level — see F50.)*
- **F50. Tenancy & hard-isolation foundation. [new — ground-up, P1]** `tenant_id` as the **outermost scope and security boundary** on every record, agent, query, metric, event, and object; tenant context propagated through API, bus, storage, and the AI/query layer; **pooled isolation enforced at the storage+query layer (defense-in-depth, above RBAC)**; cross-tenant access impossible by construction; the single-tenant sovereign install is the one-tenant case. *Designed in from M1 — the expensive-to-retrofit core of v0.4.*
- **F51. Provider / MSP management plane (foundation). [new — P1 foundation → P2]** Tenant entity + lifecycle (provision / suspend / offboard) and a **provider-operator privilege domain distinct from tenant data access**; minimal tenant provisioning in P1, full operator console in P2 (US-U2). Managing a tenant never silently grants read access to its telemetry; tenant-data access requires audited, consented, time-bounded **break-glass**.
- **F52. Tenant isolation models — pooled / siloed / hybrid. [new — P1 pooled → P2 siloed/hybrid]** Pooled (shared stores, logical isolation) in P1; siloed (per-tenant DB schema/ClickHouse db/bus namespace/object prefix) and hybrid in P2; selectable per-deployment **and per-tenant**; per-tenant residency.

### 4.2 Should Have — Phase 2 ("the wedge" + first enterprise features + first differentiators + the provider surfaces)

- **F11.** eBPF host/L7 agent (CNI-agnostic; L3–L7; service map; emits OTel).
- **F12.** OTel-native data model (OTLP in/out; semantic conventions; OBI).
- **F13.** AI RCA + NL query (unified semantic query; cited evidence; pluggable model adapter + local option). *(Tenant boundary enforced before RBAC at the query layer.)*
- **F14.** MCP server (tenant- + RBAC-scoped tools; trustctl-consistent).
- **F15.** Browser/transaction synthetic. **F16.** Endpoint agent (DEM signals).
- **F25.** Enterprise identity full (SCIM 2.0; ABAC + custom roles; AD/LDAP/Entra/Okta; MFA; delegated admin within tenant).
- **F26.** SIEM integration (audit + security/incident export; syslog/CEF/OTLP; tenant-scoped routing).
- **F27.** On-call & ITSM integration (PagerDuty/Opsgenie/Slack/Teams; ServiceNow/Jira).
- **F28.** Zero-downtime lifecycle (rolling upgrades; backward-compatible migrations; rollback; staged fleet rollouts).
- **F29.** IaC & GitOps (Terraform modules; config-as-code; Helm hardening — incl. provider/multi-tenant reference values).
- **F36. TLS/certificate observability.** Cert expiry/chain/issuer/SAN; TLS version/cipher/(JA3); CT-log correlation; deprecated-protocol flags; trustctl remediation handoff; alerting.
- **F38. Threat-intel open-data enrichment.** Score flows/connections against open feeds; provenance; AUP-tracked.
- **F39. Change intelligence + change-to-incident correlation.** Ingest config/route/BGP/IaC/deploy change events; change timeline; surface candidate changes per incident; feed AI RCA.
- **F40. Live topology graph.** *(Foundation in P2 → full in P3.)* Telemetry-fed, versioned, queryable network/topology model powering RCA + what-if.
- **F45. AI test authoring + auto-discovery.** NL → canary config; auto-discover services/prefixes/endpoints and propose tests.
- **F46. Last-mile / WiFi / ISP endpoint diagnostics.** *(Extends F16; P2/P3.)*
- **F51 (full). Provider / MSP management console. [new — P2]** Full cross-tenant operator console (US-U2): tenant inventory & lifecycle, fleet-across-tenants health, configuration, and audited break-glass.
- **F53. Per-tenant metering, usage & billing export. [new — P2]** Meter agents/tests/ingest/events/AI-calls/retention per tenant; usage API + export to the MSP's billing/PSA; per-tenant quotas & rate limits; showback/chargeback (US-U4).
- **F54. White-label / per-tenant branding. [P2 — foundation/full split, v0.5]** *Foundation:* the **themeable design-token system is built in the frontend design system at M2** so no UI hardcodes design values (F54-foundation). *Full:* per-provider and per-tenant logo/colors/custom domain/branded notifications and login are then a runtime **override of those tokens — not a per-screen retrofit** (F54-full; US-U5).
- **F55. Per-tenant data lifecycle — export, residency & verifiable deletion. [new — P2]** Tenant-scoped export; verifiable full deletion across all stores; offboarding runbook; per-tenant retention/erasure (US-U6; ties to F34).

### 4.3 Could Have — Phase 3 (full five planes + deeper enterprise + heavier differentiators)

- **F17.** Flow analytics (NetFlow/IPFIX/sFlow; top-talkers/capacity/anomaly).
- **F18.** Device / streaming telemetry (SNMP + gNMI/OpenConfig).
- **F19.** Collective internet-outage view (open-data + customer vantage points).
- **F20.** RUM convergence. **F21.** Voice/RTP tests.
- **F30.** CMDB + Grafana datasource + Prometheus federation.
- **F31.** Secrets integration (Vault/CyberArk/cloud KMS + trustctl identities). *(Also the backend for per-tenant keys — F56.)*
- **F37. NDR-lite threat detection engine.** DNS exfil/DGA, beaconing, egress anomalies, Tor/bad-ASN, lateral-movement; confidence-scored; tunable + suppression; detection-as-code; SIEM export.
- **F41. Network/cloud egress cost (FinOps) observability.** Flow + cloud egress + public pricing → per-service/flow cost; cross-AZ/region detection; trends/alerts.
- **F42. SLO + business-impact engine (OpenSLO-compatible).** SLIs from network planes; error budgets + burn-rate; service/BU mapping; exec dashboards; OpenSLO import/export.
- **F43. Continuous compliance / segmentation validation.** Declare expected segmentation; validate against eBPF/flow; flag violations; exportable PCI/NIST/zero-trust evidence.
- **F47. Network chaos / fault injection.** Inject latency/loss/partition; validate observability/SLOs catch it.
- **F48. Carbon / power observability.** Kepler-style power/carbon per workload/network.
- **F57. Tenant fairness / noisy-neighbor isolation. [new — P2/P3]** Per-tenant quotas, rate limits, ingest backpressure isolation, and query-cost guards so one tenant cannot degrade others in pooled mode (US-U4-adjacent; hardened as scale grows).

### 4.4 Enterprise Edition Track (Phase 2+ → ongoing)

- **F32.** FIPS-mode crypto (FIPS 140-3 validated module path; STIG/CIS hardening guides).
- **F33.** Multi-region / HA at scale (active-active control plane; RPO/RTO). *(Per-tenant residency interacts with multi-region — §5.1.)*
- **F34.** Advanced governance (data classification, retention/erasure, redaction, residency, BYOK/HYOK, key rotation). *(Now also per-tenant — F56.)*
- **F35.** Supportability (diagnostics + support bundles; health/readiness; self-monitoring; EE/acquirer-operated support, SLAs, PS). *(Support bundles are tenant-scoped and secret-stripped.)*
- **F44. Agentic remediation (guarded).** *(P3/EE.)* Propose + (on approval) execute safe network actions; human-gated default; RBAC + tenant-scoped; dry-run/simulation; blast-radius limits; fully audited.
- **F49. Plugin / detection marketplace + detection-as-code.** *(Phase 4.)* Community canary plugins, dashboards, and network detection rules (Sigma-style); signed/verified; flywheel.
- **F56. Per-tenant key isolation / BYOK. [new — P2/EE]** Per-tenant envelope-encryption keys; optional per-tenant BYOK/HYOK + rotation (builds on F31/F34); the cryptographic complement to data isolation, enabling siloed/regulated tenants and clean cryptographic offboarding.

> **Edition placement note [v0.4]:** the **provider/management plane (F51), white-label (F54), and per-tenant metering/billing (F53)** are the natural **commercial/enterprise-tier** features (classic open-core monetization) and are also what an MSP must license to resell. The **open core** remains the five planes + path/BGP + TLS posture + threat-intel enrichment + AI RCA/NL + topology graph + **single-tenant and the tenancy/isolation primitives (F50/F52-pooled)**. Final boundary tied to the license decision (§4.6/§10.3).

### 4.5 Won't Have — This Iteration

- **probectl-the-vendor operating its own first-party public multi-tenant SaaS — deferred. [changed in v0.4]** *Important distinction:* probectl now **provides** multi-tenancy as a first-class capability so **partners/MSPs can self-host and resell** it; what remains out of scope this iteration is **probectl itself running a hosted public SaaS** (an optional later business decision, not an architectural gap).
- **Full APM / distributed-tracing replacement** — OTel-native + correlate, not replace.
- **SIEM / log-analytics platform** — integrate (F26), don't replace. **Full NDR/IDS suite or signature IPS** — observability-with-network-threat-detection, not a packet-inspection IPS or SIEM.
- **Owning a global cloud-agent / BGP-peering fleet** — not replicated (§10).
- **FedRAMP authorization (now)** — building blocks present so it's achievable later.
- **24×7 support desk staffed by the founder** — supportability is a requirement (F35); the desk is EE/acquirer-provided. *(For MSP mode, tier-1 support to end customers is the MSP's responsibility — that's their business.)*
- **Autonomous (un-gated) remediation** — human-gated/observe-only by default (F44).

### 4.6 Feature Prioritization Logic

MoSCoW + impact/effort, sequenced to maximize: (1) **Acquisition signal + channel** — Phase 1 = visible ThousandEyes parity + enterprise + **tenancy foundations**; Phase 2 = AI/OTel/eBPF differentiation + cheap-substrate differentiators + first enterprise features + **the provider surfaces that unlock MSP resale**; Phase 3 = full planes + heavier differentiators that widen the acquirer set. (2) **Solo-founder feasibility** — Phase 1 solo (tenancy foundation is schema-and-discipline, not a new surface); Phase 2 +contributors. (3) **Open-source + channel flywheel** — front-load self-hostable, star-driving features and the single-tenant/tenancy primitives; gate the provider-plane/white-label/metering as the commercial lever MSPs license.

> **Open vs Enterprise boundary** (deferred, tied to license): enterprise-tier ≈ SSO/SCIM, advanced RBAC/ABAC, audit-to-SIEM, multi-region HA, FIPS, advanced NDR + agentic remediation, **and the provider/management plane + white-label + per-tenant metering/billing (F51/F53/F54) + per-tenant BYOK (F56)**; open core ≈ five planes, path/BGP, TLS posture, threat-intel enrichment, AI RCA/NL, topology graph, **single-tenant + tenancy/isolation primitives (F50, F52-pooled)**. **MSP resale additionally requires a commercial license** (§10.3).

---

## 5. Technical Requirements

### 5.1 Architecture Specifications

**Languages.** Go (agents + control plane; single binary; trustctl-consistent). Python for the BGP analyzer. eBPF via `cilium/ebpf` or libbpf. **Open consideration:** Rust for BGP/path/detection hot paths. **Crypto abstraction (F500):** all crypto behind an interface so a FIPS 140-3 validated module compiles in for FIPS mode — designed in from M1.

**Tenancy architecture [v0.4 — pervasive].** **Tenant is the outermost scope and security boundary**, above Org → Team → Project. Properties:
- **Identity & context propagation.** Every API request, gRPC stream, bus message, storage operation, metric, object, and AI/MCP query carries or derives a `tenant_id`; the control plane resolves tenant context at the edge and propagates it through all layers. Agents are **bound to a single tenant** at registration.
- **Isolation enforcement.** Pooled mode enforces tenant scoping **at the storage + query layer (defense-in-depth, above RBAC)** — not by application code alone — so a missing application check cannot leak data. Cross-tenant access is impossible by construction. The **unified semantic query layer enforces the tenant boundary *first*, then RBAC** (it is the AI/MCP security boundary, now two-level).
- **Isolation models (F52).** **Pooled** — shared Postgres (row-level `tenant_id` + enforced scoping/RLS), shared ClickHouse (tenant_id partition/order key), shared metrics (tenant label or per-tenant series), shared object store (per-tenant prefix), shared bus (tenant-tagged/partitioned topics). **Siloed** — per-tenant Postgres schema or instance, per-tenant ClickHouse database, per-tenant bus topic namespace, per-tenant object bucket/prefix; strongest isolation for regulated tenants. **Hybrid** — shared control plane, isolated data planes. Selectable per-deployment **and per-tenant**.
- **Provider/management plane (F51).** A distinct service/privilege domain for cross-tenant operations (tenant lifecycle, fleet-across-tenants, metering, branding). **Provider operators are not tenant users**: managing a tenant never grants read access to its telemetry; tenant-data access is via explicit, time-bounded, tenant-consented, **separately audited break-glass**.
- **Keys (F56).** Per-tenant envelope keys (optional BYOK/HYOK) via the secrets backend (F31), so siloed/regulated tenants are cryptographically separable and offboarding can be a key-destruction event.
- **Fairness (F57).** Per-tenant quotas, rate limits, ingest backpressure isolation, and query-cost guards prevent noisy neighbors in pooled mode.
- **Residency & multi-region (F33 interaction).** Per-tenant data-residency targeting; in multi-region, a tenant can be pinned to a region/data plane.

**Control plane.** Stateless, **tenant-aware** Go API server; PostgreSQL for durable state (**tenants**, tests, agents, policies, incidents, orgs/teams/projects, roles, audit, SLOs, segmentation policies, detection rules, **tenant branding, quotas, metering**); horizontal scaling; active-active multi-region (EE).

**Agent ↔ control-plane transport.** gRPC bidirectional streaming; mTLS with SPIFFE-style identity (trustctl synergy); **agent identity encodes its tenant**.

**Result/event bus.** Kafka default (topics `probectl.<type>.results`/`.events`; Protobuf; JSON dev fallback); **tenant-tagged and partitioned in pooled mode, namespaced per tenant in siloed mode**. Lightweight mode (<5 agents): NATS / Redis Streams / direct-to-TSDB.

**Storage.** Metrics: Prometheus/VictoriaMetrics (**tenant label or per-tenant series**). High-cardinality (flow, eBPF flows, path/BGP/threat/change events, cost records): ClickHouse (**tenant_id in partition/order key, pooled; per-tenant database, siloed**). Topology graph: graph-over-Postgres/ClickHouse → dedicated engine at scale (**tenant-scoped**). Object/blob: filesystem / S3-compatible / MinIO (**per-tenant prefix/bucket**).

**Agent fleet model.** Enterprise + Endpoint (P2) + eBPF (P2) agents; fleet management (F28) — staged rollouts, version-skew, remote config, IaC/GitOps; **fleet inventory is tenant-scoped, and the provider plane aggregates fleet health across tenants (F51)**. No global cloud-agent fleet; internet-wide perspective from open data + optional RIPE Atlas federation.

**Path engine.** ECMP-aware (Paris/Dublin-traceroute); ICMP/TCP-SYN/UDP; MPLS detection.

**eBPF subsystem (P2).** CNI-agnostic, observability-only (Retina model); kernel-compat matrix; emits OTel; service-map; also the substrate for L7/TLS observation, threat detection, and segmentation validation. *(eBPF data is tenant-scoped like all other signals.)*

**Flow subsystem (P3).** GoFlow2/Akvorado-style collectors → ClickHouse; also feeds threat detection, FinOps, and segmentation validation.

**Device subsystem (P3).** SNMP + gNMI/OpenConfig streaming.

**OTel pipeline (P2).** OTLP receivers/exporters + OBI. **The result/event schema is modeled on OTel resource + network semantic conventions from M2 (not retrofitted)** — so the P2 OTLP layer *exposes* signals as OTLP rather than remapping a divergent model; **tenant carried as a resource attribute**. **OTLP endpoints are TLS with authenticated, tenant-scoped receivers.**

**Security/threat subsystem (P2 → P3).** TLS/cert observer; threat-intel ingestion (shared feeds, per-tenant scoring); detection engine (confidence-scored; detection-as-code; suppression); detections → SIEM (F26, tenant-routed) + the incident timeline.

**Change & topology subsystem (P2 → P3).** Change-event ingestion; change timeline + incident correlation; live topology graph (versioned; tenant-scoped); powers RCA + what-if.

**Cost / FinOps subsystem (P3).** Correlate flow + cloud egress + public pricing; per-service/flow attribution; cross-AZ/region detection. *(Per-tenant attribution doubles as MSP cost-to-serve insight.)*

**SLO subsystem (P3).** OpenSLO-compatible SLI/SLO over network signals; error-budget + burn-rate; service/BU mapping.

**AI layer (P2+).** Unified semantic query/correlation over all stores + topology → built-in NL-query/RCA UI, test-authoring, guarded remediation (F44), the MCP server, and REST/gRPC/CLI; pluggable model adapter; local-model path for sovereignty; **all AI + MCP calls enforce the tenant boundary first, then caller RBAC**; remediation is human-gated/observe-only by default with dry-run + blast-radius limits + audit.

**Identity & tenancy (F500 + MSP).** **Tenant →** Orgs → Teams → Projects; RBAC + custom roles + ABAC; SCIM-driven lifecycle; SSO (SAML/OIDC), **per-tenant IdP** supported; delegated admin within tenant; **provider-operator domain separate (F51)**; designed in from M1.

**Frontend & design system (P1 foundation) [v0.5].** A themeable **design-token system** (no hardcoded design values), a documented component library, an app shell with a **command palette** + keyboard-first model, and a **WCAG 2.2 AA** baseline are built **early (M2) as the foundation for every UI surface** — so the hero visuals, the AI surface, and all later screens build on a system rather than inventing one, and **white-label (F54) is a per-tenant token override, not a per-screen retrofit**. **Served HTTPS-by-default with HSTS, a CSP, and Secure+HttpOnly+SameSite session cookies.** (See §6.)

**Packaging & deployment.** Single-binary agent (compiled-in plugins); multi-arch Docker; Helm (K8s/OpenShift) **with single-tenant and multi-tenant/provider reference values**; docker-compose (small/all-in-one); air-gapped bundle; Terraform modules + GitOps config-as-code. **Deployment topologies (§5.4):** sovereign single-tenant; MSP pooled; MSP siloed/hybrid. **The shipped Helm + compose are HTTPS-by-default (TLS-terminating ingress, HSTS; no plaintext API exposure).**

### 5.2 API Requirements

- **REST (OpenAPI 3.1, versioned + deprecation policy + LTS).** CRUD for tests/agents/dashboards/alerts/incidents/orgs/roles/SLOs/segmentation-policies/detection-rules; query endpoints (threats, changes, topology, costs); **all tenant-scoped**. **Provider API [v0.4]:** tenant lifecycle (create/suspend/offboard), per-tenant config, branding, quotas, **usage/metering export**, and break-glass — under the provider-operator privilege domain.
- **gRPC** (agent ↔ control plane; tenant-bound identity).
- **MCP server.** Tenant- + RBAC-scoped tools: `get_path`, `get_bgp_events`, `query_flows`, `list_tests`, `correlate_incident`, `explain_degradation`, `get_tls_posture`, `get_threats`, `get_changes`, `query_topology`, `query_costs`, `get_slo_status`, `validate_segmentation`, `propose_remediation` (proposal only; execution gated). *(All scoped to the caller's tenant first; the HTTP/SSE transport is TLS + authenticated when network-exposed.)*
- **Webhooks + integrations.** Alert/incident lifecycle; inbound change-event webhooks; on-call/ITSM/SIEM/CMDB connectors; **per-tenant routing**; **usage/billing export to PSA/billing systems (F53)**. **Inbound webhooks verify the sender's signature (HMAC) over TLS and are authenticated + tenant-scoped; events are treated as untrusted.**
- **OTLP** ingest/export (P2; tenant as resource attribute; **TLS endpoints; authenticated, tenant-scoped receivers**).
- **AuthN/Z.** **All API/UI traffic is HTTPS (HSTS; TLS 1.2+; Secure+HttpOnly+SameSite cookies).** SSO (per-tenant IdP); SCIM; **tenant boundary** + RBAC + ABAC; scoped keys/service accounts; rate limiting (per-tenant); structured errors.
- **CLI/TUI.** Web-parity; tenant-context aware; provider subcommands for operators.

### 5.3 Data Requirements — including Open-Data + Threat-Intel Ingestion

**Core data model (entities).** **Tenant (the outermost scope) [v0.4]**; Provider/ProviderOperator + break-glass grant [v0.4]; TenantBranding [v0.4]; TenantQuota + MeteringRecord/UsageRecord [v0.4]; (per-tenant key reference [v0.4]). Test/Canary; Result; Agent (**tenant-bound**); Path; BGPEvent; Flow (P3); DeviceMetric (P3); eBPFFlow/ServiceEdge (P2); Incident (cross-plane, change + threat refs, timeline, RCA, evidence); Alert/Rule; Dashboard/Widget; Policy; AuditEvent (immutable; **tenant-scoped + a separate provider/break-glass audit stream**); Org/Team/Project; Role/Permission (RBAC+ABAC); User/ServiceAccount (SCIM); OTel resource/signal mapping; OpenDataSource (type, cadence, creds, health, AUP/provenance). From v0.3: TLSObservation/Certificate; ThreatSignal/Detection + DetectionRule; ThreatIntelSource; ChangeEvent; TopologyNode/TopologyEdge (versioned); CostRecord/EgressMetric; SLO/SLI/ErrorBudget; SegmentationPolicy/ValidationResult; RemediationAction/Runbook. **`tenant_id` is a mandatory attribute on every tenant-scoped entity.**

**Open-Data + Threat-Intel Ingestion Layer.** *(Shared infrastructure across tenants; enrichment/scoring applied per-tenant; provenance + AUP tracked.)*

*Internet measurement / routing (v0.1):* RouteViews; RIPE RIS + RIS Live; RIPEstat; RIPE Atlas; CAIDA (BGPStream/Ark); PeeringDB; RPKI (Routinator/rpki-client); IRR (RADB); MaxMind GeoLite2; Team Cymru IP-to-ASN; RIR delegated-stats; IODA; Cloudflare Radar; OONI; M-Lab; OpenINTEL; Certificate Transparency logs (crt.sh); public looking glasses.

*Threat intelligence (v0.3):* Spamhaus DROP/EDROP; abuse.ch — URLhaus, Feodo Tracker, ThreatFox, SSLBL; FireHOL; Tor exit-node list; GreyNoise (community tier); CINS Army list; Emerging Threats open rules. *(Several have non-commercial or attribution terms — tracked in `ThreatIntelSource`/`OpenDataSource` AUP metadata; **especially relevant once an MSP resells commercially — §10.3**.)*

*Cost (v0.3):* cloud provider egress/transfer pricing (public pricing data/APIs).

Design: pluggable ingestion normalizes each source; cached; **degrade gracefully**; per-source AUP + provenance tracked; **shared once, scoped per tenant**.

### 5.4 Performance & Scale Specifications (F500 + MSP magnitude)

**Agent footprint.** Retina-class efficiency; fast cold start; single binary.

**Reference architectures (now spanning sovereign and provider topologies) [v0.4]:**
- **S** — homelab/compose, single-tenant, ≤~25 agents.
- **M** — HA K8s, single-tenant or **small pooled multi-tenant** (handful of tenants), hundreds of agents.
- **L** — multi-node, sharded ClickHouse, **pooled multi-tenant (many tenants)** or siloed for a few large tenants; thousands of agents.
- **XL/global** — multi-region active-active, **MSP-scale pooled + siloed mix, tens to hundreds of tenants**, tens of thousands of agents; high-volume flow/eBPF/threat/cost data; per-tenant residency.

**Targets (load-test gated):** sub-second interactive/query latency on typical windows **within a tenant**; fast topology-graph + path-viz render; ingestion + detection sized per tier; **per-tenant isolation under multi-tenant load (no cross-tenant performance bleed — F57)**; numeric SLOs set/validated in P2/P3. Note: threat detection, cost correlation, the topology graph, and **multi-tenant fan-out** add cardinality/compute — sized into L/XL and gated by load tests including a **noisy-neighbor scenario**.

---

## 6. User Experience Requirements

> Product principle: UX must *crush* the incumbents — without sacrificing the enterprise admin, the security/cost/SLO surfaces, **or the provider/MSP operator experience**.

### 6.1 Design Principles

- Single pane, all planes + security + cost + SLO + change. TTFI < ~10 min (aided by AI authoring + auto-discovery). Sub-second, keyboard-first, dark-native, command palette.
- **Built on a design-system foundation [v0.5]:** a themeable token system + component library + accessible app shell, established early (M2) — the basis for every surface and the mechanism for white-label; **no UI hardcodes design values**.
- **AI as a primary surface** (NL query → RCA → proposed remediation) — **a design-led, iterated surface**. **Correlation timeline as the incident home**, overlaying change + threat alongside the five planes.
- **Two hero visuals** — the interactive path visualization *and* the live topology graph — are **design-led, iterated surfaces** (spike → iterate, not one-shot); they carry the "crush the incumbents" bar.
- Enterprise admin, security/threat, cost/FinOps, SLO/error-budget, and compliance/segmentation evidence are clean first-class surfaces.
- **Provider/MSP operator surface [v0.4]:** a clean cross-tenant console (tenant list/health, onboarding wizard, per-tenant config, branding, metering/billing) — distinct from any single tenant's UI, and visually separable so operators never confuse tenant context.
- **White-label fidelity [v0.4]:** the end-customer tenant UI fully reflects the MSP's brand (logo, theme, domain, notification branding); probectl branding is replaceable per tenant.
- Opinionated defaults (templates, alerts, dashboards, starter detections, starter SLOs).

### 6.2 Interface Requirements

- Responsive web UI; embeddable widgets; CLI + TUI (web-parity).
- Enterprise admin console (users/roles/orgs **within a tenant**, SCIM, audit search/export, integration + detection-rule config).
- **Provider console [v0.4]:** tenant lifecycle, fleet-across-tenants, metering/billing, branding, break-glass (audited).
- IA (tenant context): Targets/Tests → Path/Topology → Incidents/Timeline → Security → Cost → SLOs → Ask (AI) → Dashboards → Compliance → Admin/Settings. **Provider context:** Tenants → Fleet → Usage/Billing → Branding → Operators/Break-glass → Settings.
- Clear, actionable error/empty states; explicit, always-visible **tenant indicator** to prevent context confusion.

### 6.3 Usability & Accessibility Criteria

- Onboarding task-completion ≥ 90% without docs; TTFI measured/trended; AI-answer usefulness rated; path-viz + topology comprehension validated; **tenant-onboarding time measured for the provider flow.**
- **Accessibility:** WCAG 2.2 AA + published **VPAT**.

---

## 7. Non-Functional Requirements

### 7.1 Security

- Self-hosted, data-sovereign by default; air-gapped install; no phone-home telemetry by default. *(Provider mode: data stays with the operating MSP, never a probectl public SaaS.)*
- **Tenant isolation is the outermost security boundary [v0.4].** Cross-tenant data leakage is treated as a catastrophic failure class. **Defense-in-depth:** storage-layer scoping (RLS/partitioning or physical siloing) **plus** query-layer enforcement **plus** automated cross-tenant isolation tests in CI; the AI/MCP query layer enforces **tenant first, then RBAC**. **Provider operators cannot read tenant telemetry** without time-bounded, tenant-consented, separately-audited **break-glass**. Optional **per-tenant keys/BYOK (F56)** make tenants cryptographically separable.
- AuthN/Z: SSO (SAML/OIDC; **per-tenant IdP**); SCIM 2.0; AD/LDAP/Entra/Okta; **tenant boundary** + RBAC + custom roles + ABAC; service accounts; MFA; session policies; separation of duties (incl. operator-vs-tenant separation).
- In transit: **TLS on every network listener — the API/UI are HTTPS-by-default (HSTS; Secure+HttpOnly+SameSite cookies; CSP), OTLP and MCP are TLS, datastore/bus connections use TLS in transit, and outbound fetches validate certificates**; **mTLS** (SPIFFE-style; trustctl synergy; tenant-bound agent identity) for the agent transport. At rest: envelope encryption; KMS/HSM; **per-tenant keys + BYOK/HYOK + rotation (F56, EE)**. FIPS 140-3 mode (EE/regulated).
- **Inbound ingestion is authenticated and untrusted [v0.5].** Every inbound surface — the API, **OTLP receivers, and webhooks** — is **authenticated and tenant-scoped**, **verifies the sender's signature where the sender signs (Git/CI webhook HMAC)**, and treats received data as **untrusted input**. When TLS, a credential, or a required signature is missing, **fail closed**.
- **AI sovereignty + safety:** local-model path; AI + MCP enforce tenant + caller RBAC; remediation human-gated/observe-only by default, with dry-run, blast-radius limits, and full audit.
- **Threat-detection hygiene:** confidence scoring, tunable thresholds, suppression, detection-as-code; detections are signals/integrations, not an un-tuned auto-blocking IPS.
- Vulnerability management: dependency + image scanning in CI; CVE patch SLAs; periodic pen-testing (**including a cross-tenant isolation/penetration focus**); responsible disclosure + security.txt; SLSA-aligned provenance; SBOM; signed releases.
- Hardening: secure defaults; no default creds; CIS/STIG guides (EE/regulated). Audit: immutable, tamper-evident; **tenant-scoped + a separate provider/break-glass audit trail**; SIEM export.

### 7.2 Performance

See §5.4. Sub-second interactions within a tenant; bounded agent overhead; detection/cost/graph workloads sized into L/XL tiers; **per-tenant fairness (F57) validated under multi-tenant load**; numeric SLOs validated in P2/P3.

### 7.3 Reliability & DR

- Control-plane HA; active-active multi-region (EE). Agent store-and-forward (no silent data loss). Bus durability or local buffering.
- Zero-downtime upgrades (rolling; backward-compatible idempotent migrations; rollback; version-skew). Backup/restore + tested DR runbook; defined RPO/RTO; **per-tenant backup/restore and offboarding (verifiable deletion — F55)**. Self-monitoring + health/readiness endpoints.

### 7.4 Scalability

- Horizontal control plane; sharded/partitioned TSDB + ClickHouse; topology graph scaled via dedicated engine at XL. Open-data/threat-intel ingestion scales independently + degrades gracefully. Plugin/detection model grows capabilities without core changes. **Multi-tenant scale: pooled scaling across many tenants + siloed scaling for large/regulated tenants; per-tenant quotas (F57); per-tenant residency in multi-region (F33).**

### 7.5 Compliance & Certifications

- Posture mapping to SOC 2 / ISO 27001 / NIST (+ sector frameworks). Residency by construction (documented procurement claim; **now also per-tenant**). SIEM export of audit + security/threat events.
- **Compliance-as-observability:** segmentation validation (F43) → exportable PCI/NIST/zero-trust evidence; TLS/cert posture (F36) feeds cryptographic-control evidence.
- **Tenant-isolation as a compliance artifact [v0.4]:** documented isolation model, break-glass controls, and per-tenant key/erasure capability support an MSP's own client-facing compliance and an enterprise tenant's vendor-risk review of the MSP.
- Certification path (EE/acquirer-supported): SOC 2 Type II (managed) + ISO 27001 (org); FIPS 140-3 (product); VPAT; FedRAMP achievable-but-deferred.
- Procurement readiness: SIG / CAIQ; DPA / MSA / EULA templates (**incl. MSP/reseller terms — §10.3**); data-flow + architecture diagrams (single-tenant and multi-tenant).

### 7.6 Data Governance & Privacy

- PII handling (IPs are PII): classification, configurable retention + erasure, redaction/masking, documented lawful basis; **per-tenant retention/erasure and verifiable deletion (F55)**. Residency controls (**per-tenant**). Threat-intel + open-data provenance tracked per source/record. Retention/tiering + downsampling + archival. **In provider mode, the MSP is typically the data processor and its tenants the controllers — DPA templates reflect this (§10.3).**

---

## 8. Success Metrics & Analytics

> Revenue is intentionally **not** a near-term KPI. Metrics emphasize adoption, parity, differentiation, F500-readiness, **MSP/provider traction**, and acquisition signal.

### 8.1 Key Performance Indicators

- **Adoption:** stars; self-hosted deployments; active agents; monitored targets; downloads; community detection rules/plugins.
- **MSP / provider [new in v0.4]:** resellers/providers running probectl; **tenants served; tenant-onboarding time; per-tenant cost-to-serve; white-label deployments; metered-usage accuracy; zero cross-tenant incidents** (a hard gate).
- **Engagement:** TTFI; weekly active UI users; AI NL-query/RCA/authoring invocations; MCP connections.
- **Differentiation usage:** TLS/cert issues caught (and trustctl handoffs); threat detections actioned; change-correlated incidents; topology queries; egress cost surfaced/saved; SLOs tracked; segmentation validated; guarded remediations proposed/approved.
- **Parity & coverage:** % ThousandEyes software-layer surface shipped; plane-coverage (1→5); AI/OTel demo readiness.
- **F500-readiness gates:** enterprise foundations + features shipped; reference-architecture scale validated (L/XL, **incl. multi-tenant noisy-neighbor**); SIG/CAIQ readiness; VPAT published.
- **Strategic / acquisition signal:** design-partner logos (regulated/air-gapped F500 + **reference MSPs**); inbound from networking, security, observability, **and MSP-platform** players; PR/conference traction.

### 8.2 Analytics Implementation

- Sovereignty-respecting telemetry (opt-in, self-hostable; no phone-home). Public adoption via download proxy + optional opt-in heartbeats. Internal self-monitoring (**per-tenant + provider-aggregate**). AI-answer + detection feedback capture to improve quality.

### 8.3 Success Measurement

- Baselines per phase; targets per phase (Phase 1 = OSS launch + parity + foundations **incl. tenancy**; Phase 2 = AI/OTel/eBPF + TLS/threat-intel/change/authoring + first enterprise features + **provider plane / isolation models / metering / white-label + first reference MSP**; Phase 3 = full planes + NDR/FinOps/SLO/segmentation + L/XL scale incl. multi-tenant). Per-milestone retro vs requirement IDs.

---

## 9. Implementation Plan

### 9.1 Development Phases

- **Phase 1 — MVP / "ThousandEyes core + enterprise + tenancy foundations" (~0–6 mo, solo).** F1–F10 + F22–F24 + **F50 (tenancy/isolation foundation), F51-foundation (tenant entity/lifecycle + provider-operator domain), F52-pooled**.
- **Phase 2 — Wedge + first enterprise features + cheap-substrate differentiators + provider surfaces (~6–12 mo, +1–2 contributors).** F11–F16 + F25–F29 + F36/F38/F39/F40-foundation/F45/F46 + **F51-full (provider console), F52 siloed/hybrid, F53 (metering/billing), F54 (white-label), F55 (lifecycle/export/deletion), F56 (per-tenant keys), F57 (fairness)**.
- **Phase 3 — Full five planes + heavier differentiators (~12–18 mo).** F17–F21 + F30–F31 + F37/F40-full/F41/F42/F43/F47/F48.
- **Enterprise Edition track (Phase 2+ → ongoing).** F32–F35 + F44; certs; support/SLA/PS. *(F51/F53/F54/F56 are the MSP-facing commercial-tier features.)*
- **Phase 4 — (Optional) first-party managed offering + scale hardening + F49 (marketplace).** Revisit license + open/enterprise boundary + reseller terms.

### 9.2 Resource Allocation

- Phase 1: founder. Phase 2: + Go/eBPF engineer, frontend/design, AI/RCA + detection-engineering + AI-authoring, **+ a multi-tenancy/billing-integration contributor**. Phase 3 + EE: + BGP/flow/data-engineering, FinOps/SLO; security/compliance + support via EE/acquirer. OSS contributors front-loaded; certification/audit spend deferred.

### 9.3 Timeline & Milestones (indicative; requirement-ID-linked)

- **M1 — Scaffolding + agent skeleton + tenancy foundation:** control plane + Postgres (**tenants/**orgs/roles/audit schema), gRPC agent reg (mTLS, **tenant-bound**), single-binary agent, CI (dep/image scan **+ cross-tenant isolation test harness**), multi-arch Docker, crypto abstraction, **pooled tenant scoping + tenant context propagation (F50/F52-pooled), tenant entity + minimal provisioning + provider-operator domain (F51-foundation)**. (F1, F10, F22–F24, F50, F51-foundation, F52-pooled)
- **M2 — Network tests + results pipeline (tenant-scoped, OTel-convention schema) + frontend foundation & design system.** (F2, F8, F9, F54-foundation)
- **M3 — Path visualization (hero — design-led).** (F3)
- **M4 — DNS + HTTP tests.** (F4, F5)
- **M5 — BGP monitoring + open-data enrichment (shared ingest, per-tenant scoping).** (F6, F7)
- **M6 — Alerting + incident timeline + SSO + audit + load/perf harness & scale baseline + OSS launch.** → **Phase 1 GA.** (F8, F9, F22, F23) *(Ships single-tenant-ready with the tenancy foundation present; scale baseline captured here, full L/XL at M14.)*
- **M7 — eBPF feasibility spike → eBPF agent + service map.** (F11)
- **M8 — OTLP exposure + OBI (semantic conventions already in the M2 schema).** (F12)
- **M9 — Topology foundation (moved earlier so RCA is topology-grounded) → AI RCA + NL + tenant/RBAC-scoped MCP + AI authoring.** (F40-foundation, F13, F14, F45)
- **M10 — Security & change pass:** TLS/cert + trustctl handoff; threat-intel; change-to-incident. (F36, F38, F39)
- **M11 — Enterprise features + lifecycle + provider surfaces:** SCIM/ABAC; SIEM; on-call/ITSM; zero-downtime lifecycle; IaC; last-mile endpoint; **provider console (F51-full), siloed/hybrid isolation (F52), per-tenant metering/billing (F53), white-label (F54), per-tenant lifecycle/export/deletion (F55), per-tenant keys (F56), fairness (F57)**. → **Phase 2.** (F25–F29, F46, F51-full, F52, F53, F54, F55, F56, F57)
- **M12 — Flow + device + observability-stack integrations.** (F17, F18, F30, F31)
- **M13 — Heavier differentiators:** NDR; full topology; FinOps; SLO; segmentation; outage; RUM; voice. (F37, F40-full, F41, F42, F43, F19, F20, F21)
- **M14 — Flourishes + scale validation:** chaos; carbon; **L/XL load-test gate incl. multi-tenant noisy-neighbor scenario**. → **Phase 3.** (F47, F48)
- **EE milestones (parallel/after):** FIPS + STIG/CIS (F32); multi-region HA + RPO/RTO (F33; **per-tenant residency**); advanced governance/BYOK (F34); supportability (F35); guarded remediation (F44). **Phase 4:** marketplace (F49) + (optional) first-party managed offering.

(Founder note: cadence flexes around newborn-care; Phase 1 solo-sized; the tenancy *foundation* is schema-and-discipline at M1, the heavier provider surfaces are Phase 2 with contributors.)

*(Integration gates [v0.5]: each milestone closes with a **cross-plane correlation test** — inject a known multi-plane fault, then assert it surfaces as **one correlated, fully tenant-scoped incident** with cross-plane evidence — run in addition to per-sprint checks. The gate grows with the planes available (synthetic↔path at M2–M3; +routing at M5; unified timeline at M6; +eBPF at M7; topology-grounded cited RCA at M9; +change at M10). A red gate is a stop-ship for that milestone.)*

---

## 10. Risk Assessment & Mitigation

### 10.1 Technical Risks

- **Cross-tenant data leakage (catastrophic). [new — v0.4, highest-severity]** A single leak destroys MSP trust and is a reportable security incident. → Tenancy designed in from M1 (F50); **defense-in-depth** (storage-layer scoping/RLS or physical siloing **+** query-layer enforcement **+** mandatory cross-tenant isolation tests in CI); tenant boundary enforced **before** RBAC in the AI/MCP query layer; optional per-tenant keys (F56); isolation-focused pen-testing; siloed model available for the highest-assurance tenants.
- **Tenancy retrofit risk. [new]** Bolting tenancy on later would be a rewrite. → Promoted to a **ground-up** requirement (F50) at M1; `tenant_id` mandatory on every entity; the single-tenant install is just one tenant, so there is no separate code path to drift.
- **Noisy-neighbor / one tenant degrading others (pooled). [new]** → Per-tenant quotas, rate limits, backpressure isolation, query-cost guards (F57); siloed option; noisy-neighbor load-test gate (M14).
- **Provider plane as a high-value attack surface. [new]** A compromised operator could reach many tenants. → Operator domain is separate from tenant data; **no implicit data access**; audited, consented, time-bounded break-glass; strong operator MFA/SoD; per-tenant keys limit blast radius.
- **Scope (five planes + enterprise + differentiators + tenancy/provider).** *High.* → Strict phasing; tenancy foundation is cheap-at-M1 schema work, provider surfaces are P2; v0.3 differentiators cheap-substrate-first; hard "Won't Have."
- **eBPF kernel/CNI/version fragmentation.** → CNI-agnostic, observability-only; OBI/OTel alignment; kernel matrix; **a feasibility spike precedes the agent build (M7)**.
- **Path-viz + topology accuracy/scale.** → Paris/Dublin; versioned graph; graph engine at XL; load-test gate.
- **Sequencing & UX-foundation risks (identified in the v0.5 architecture review; now mitigated). [new — v0.5]** Three risks that would have compounded if left: (a) a **topology→RCA dependency inversion** (RCA needs the topology graph) → the topology foundation was moved ahead of the AI layer (M9, was M10); (b) **scale validated too late** (only M14) → a load/perf harness + scale baseline left-shifted to M6 (full L/XL still gated at M14); (c) **UI/UX + white-label treated as emergent** → a frontend foundation & design system established at M2 (themeable tokens, component library, app shell, WCAG 2.2 AA), making the hero/AI surfaces design-led and white-label a token override. Cross-plane correlation is verified by **milestone integration gates**, and **OTel drift** is prevented by designing semantic conventions into the M2 schema.
- **Integration-surface exposure (API / OTLP / webhooks / datastores). [new — v0.5]** A network-*security* product must not itself expose unauthenticated or plaintext surfaces. → A **transport-security guardrail** mandates TLS on every listener, **HTTPS-by-default deploys (HSTS)**, **authenticated + signature-verified + untrusted ingestion** (OTLP, webhooks, API), **datastore/bus TLS in transit**, and **cert-validated outbound fetches** — operationalized in S1/S3 (HTTPS), S22 (OTLP), S29 (webhook HMAC), and the deploy sprints; **fail closed** when a secure channel, credential, or signature is missing.
- **High-cardinality storage cost (flow/eBPF/threat/change/cost, ×tenants).** → ClickHouse + retention tiers + sampling; tenant partitioning; lightweight mode for small users.
- **Threat-detection false positives.** → confidence scoring, tunable thresholds, suppression, detection-as-code, SIEM hand-off.
- **Cloud cost-data dependency.** → start with egress volume + public pricing; degrade gracefully.
- **Retrofitting enterprise/security/tenancy into a non-ready architecture.** → identity/multi-org/**tenancy**/audit/crypto-abstraction designed in from M1; eBPF/flow as shared substrate.
- **F500 + MSP-scale validation.** → S–XL reference architectures (single-tenant + pooled + siloed) + load tests (incl. noisy-neighbor) as a release gate.

### 10.2 Open-Data & Threat-Intel Risks

- **AUP/terms for commercial *and reseller* use. [updated — v0.4]** RouteViews/RIS are open; RIPE Atlas is credit-based; some threat feeds restrict commercial redistribution. **MSP resale is commercial use across many customers — this makes the AUP matrix a gating item before any reseller goes live, not just before a first-party managed offering.** → Track per-source terms + provenance in `OpenDataSource`/`ThreatIntelSource`; resolve commercial/reseller redistribution terms before enabling provider mode commercially; does not block private development or single-tenant OSS use.
- **Rate limits / credits / freshness / coverage.** → cache; host RIPE Atlas probes; multi-source redundancy; **shared-ingest-once across tenants** to avoid per-tenant rate-limit multiplication; never depend on a single external source for core function.

### 10.3 Compliance, Procurement & Commercial Risks

- **License must support resale. [updated — v0.4]** *Material.* A source-available core is fine for self-hosting, but **MSPs reselling a managed service need explicit commercial-licensing terms** (and likely IP-indemnification). → Define a commercial/reseller license tier alongside the source-available core; the **provider plane/white-label/metering (F51/F53/F54)** is the natural paid edition; decision deferred but now **on the critical path to enabling the MSP channel**.
- **MSP data-processor / DPA posture. [new]** In provider mode the MSP is typically the data processor and its tenants the controllers. → Provide DPA/MSA/reseller-agreement templates and per-tenant residency/erasure/key controls so MSPs can meet their own client obligations.
- **Certification/audit effort & cost.** → deferred to EE/acquirer; building blocks present so certs (incl. tenant-isolation attestations) are achievable without re-architecture.
- **Support/SLA expectations.** → supportability built in (F35); for MSP mode, tier-1 to end customers is the MSP's responsibility; EE/acquirer provides operator-facing support/SLA.
- **"Becoming a security product" dilutes focus.** → position as observability-with-network-threat-signal, not a full NDR/IPS/SIEM; integrate with the SOC stack; lead with the trustctl-adjacent TLS/threat-intel story.

### 10.4 Competitive / Business Risks

- **Well-funded incumbents (Cisco/ThousandEyes/Splunk/Isovalent, Datadog, Kentik, Grafana, Catchpoint, Selector AI).** → compete on the intersection they can't serve: self-hosted/sovereign, unified, AI/agentic, security-native, change-aware, cost/SLO-aware, **and multi-tenant/white-label for MSPs**. (The v0.3 + v0.4 extensions widen the acquirer set across security, observability, FinOps, **and MSP-platform** players.)
- **MSP-platform incumbents (Auvik, Domotz, LogicMonitor, ConnectWise/N-able/Kaseya/Datto). [new]** → they are SaaS-only, network-shallow, not sovereign, not AI/OTel-native; probectl is the **self-hostable, five-plane, AI-native, sovereign-capable, white-label** platform an MSP can resell — including in-region. Risk: they add depth/AI; mitigated by probectl's plane breadth + sovereignty + open-core flywheel.
- **AI-SRE wave commoditizing RCA.** → probectl's RCA is grounded in network ground-truth across five planes + topology + change + threat, self-hosted, local-model-capable, **and tenant-scoped**.
- **Vantage-fleet gap vs ThousandEyes.** → customers/MSPs bring vantage points; public collectors + optional RIPE Atlas; not out-scaled.
- **Solo-founder bandwidth (+ newborn).** → realistic phasing; tenancy foundation cheap-at-M1; provider surfaces P2 with contributors; OSS leverage.
- **Acquisition-thesis risk.** → the v0.3 moat (security + change + cost/SLO) **plus the v0.4 multi-tenant/MSP dimension** materially widen the acquirer set and add a distribution channel.

### 10.5 Mitigation Summary

Phase discipline + a hard "Won't Have" list de-risk scope; enterprise + security + **tenancy** foundations designed in from M1 de-risk retrofits; **defense-in-depth tenant isolation + mandatory cross-tenant tests + audited break-glass + optional per-tenant keys de-risk the catastrophic leakage class**; shared eBPF/flow substrate makes security/segmentation cheap; multi-source open-data/threat-intel with graceful degradation + shared-ingest-once de-risk external dependencies and per-tenant rate limits; confidence-scored, tunable, detection-as-code de-risk false positives; human-gated/observe-only remediation de-risks agentic action; per-tenant quotas/fairness de-risk noisy neighbors; reference architectures (single-tenant + pooled + siloed) + load-test gates de-risk F500/MSP scale; a commercial/reseller license tier de-risks the MSP channel; and the self-hosted/sovereign/AI-native/security-native/change-aware/**multi-tenant** positioning de-risks competition, procurement, and the acquisition thesis.

---

## PRD Template (condensed)

### 1. Executive Summary
- **Product:** probectl · **Owner:** Shankar · **Status:** Draft v0.4 (private; F500-hardened + differentiation moat + multi-tenancy/MSP) · **Updated:** May 31, 2026
- **Vision:** Self-hosted, source-available, OTel-native, **multi-tenant** unified network observability across five planes + native security, change-aware topology, and cost/SLO intelligence; AI/agentic; enterprise-ready; deployable sovereign-single-tenant *or* MSP-multi-tenant — data never leaves the operator's network.
- **Success Metrics:** OSS adoption + ThousandEyes parity + differentiation usage + **MSP/provider traction (tenants, resellers, zero cross-tenant incidents)** + F500-readiness gates + acquisition signal.

### 2. Problem & Opportunity
- **Problem:** Fragmented, single-plane, mostly-SaaS tooling can't see the seam, what changed, or threats; security/cost/SLO live in separate tools; sovereignty + procurement lock F500s out; **MSPs lack a sovereign, deep, white-label, multi-tenant network platform to resell.**
- **Opportunity:** Underserved intersection of OTel + AI/agentic + sovereignty + security + change + cost/SLO + **resaleable multi-tenancy**; widened acquirer set + an MSP channel.
- **Solution:** One self-hosted, enterprise-ready, **multi-tenant** pane unifying five planes + security + change/topology + cost + SLO; OTel-native; AI-driven; open-data + threat-intel-fed; sovereign single-tenant *or* MSP-multi-tenant on one codebase.

### 3. User Requirements
- **Primary Users:** Platform/SRE + network engineers (regulated/F500); security/SOC; FinOps owners; enterprise security/compliance gatekeepers; **MSP/provider operators (resellers)**; OSS self-hosters.
- **Key Use Cases:** "Network/app/route/change/threat, and where?"; path/BGP/topology incident detection; cross-plane RCA + guarded remediation; sovereign NDR-lite + TLS posture; egress cost; network SLOs; segmentation evidence; **stand up + resell a branded, hard-isolated, metered multi-tenant managed service.**
- **Success Criteria:** Faster triage/MTTR; fewer tools; audits + procurement pass; threats caught; cost cut; SLOs met; **MSPs onboard tenants in minutes, bill per usage, never leak across tenants, resell under their brand.**

### 4. Product Requirements
- **Must (P1):** five-plane core foundations (F1–F10) + enterprise foundations (F22–F24) + **tenancy foundations (F50, F51-foundation, F52-pooled)**.
- **Should (P2):** eBPF; OTel; AI RCA/NL/MCP; synthetic/endpoint; SCIM/SIEM/on-call/lifecycle/IaC; TLS/threat-intel/change/topology foundation/AI authoring/last-mile; **provider console (F51), siloed/hybrid (F52), metering/billing (F53), white-label (F54), per-tenant lifecycle/deletion (F55), per-tenant keys (F56), fairness (F57)**.
- **Could (P3):** flow; device; outage/RUM/voice; CMDB/Grafana/Prometheus; secrets; NDR; full topology; FinOps; SLO; segmentation; chaos; carbon.
- **Enterprise Edition:** FIPS; multi-region HA; advanced governance/BYOK; supportability; guarded remediation; (Phase 4) marketplace. *(F51/F53/F54/F56 = MSP commercial-tier.)*
- **Won't (now):** **probectl-operated first-party public SaaS** (capability shipped for partners; vendor-run hosting deferred); APM/tracing replacement; SIEM/IPS replacement; global fleet; FedRAMP authorization; founder-staffed 24×7 desk; un-gated autonomous remediation.

### 5. Technical Specifications
- **Architecture:** Go agents (**tenant-bound**) + stateless **tenant-aware** Go control plane + Postgres (**tenants**/orgs/roles/audit/SLOs/policies/detections/branding/quotas/metering); gRPC (mTLS/SPIFFE); Kafka bus (tenant-tagged/namespaced; + lightweight mode); Prometheus + ClickHouse (**tenant-partitioned**) + topology graph; eBPF (Retina-style, shared substrate); OTLP (tenant resource attr); pluggable AI adapter + tenant/RBAC-scoped MCP; **tenant boundary enforced at storage+query layer (defense-in-depth) above RBAC**; provider/management plane (distinct operator domain + audited break-glass); pooled/siloed/hybrid isolation; per-tenant metering/branding/keys/deletion; crypto abstraction for FIPS; multi-org + RBAC/ABAC + SCIM/SSO (per-tenant IdP); versioned API + LTS; fleet mgmt; IaC/GitOps; air-gapped bundle.
- **Dependencies:** public open-data + open threat-intel feeds + cloud pricing data; OpenTelemetry; optional local LLM. *(Reseller commercial-use AUP — §10.2/§10.3.)*
- **Performance:** low agent overhead; sub-second UI within tenant; S→XL reference architectures (single-tenant + pooled + siloed); scale + detection/cost/graph + **multi-tenant noisy-neighbor** validated by load tests.

### 6. Success Metrics
- **Primary:** OSS adoption + ThousandEyes parity + F500-readiness gates + **MSP/provider traction**.
- **Secondary:** TTFI; AI usage; MCP; differentiation usage; **tenant-onboarding time; per-tenant cost-to-serve; white-label deployments; zero cross-tenant incidents**; contributors; design-partner logos (+ reference MSPs).
- **Timeline:** reviewed per milestone (M1–M14 + EE).

---

## Quality Checklist

- ✅ Problem clearly defined with evidence (seam/"green dashboards"; "what changed"; threats/cost/SLO sprawl; sovereignty + procurement gating; **MSP multi-tenant/white-label/sovereign gap**).
- ✅ Solution aligned to user needs (sovereign, **multi-tenant**, unified, AI/agentic, security-native, change-aware, cost/SLO-aware, enterprise-ready) and the build-to-acquire goal + MSP channel.
- ✅ Requirements specific and measurable (F1–F57, US-* incl. epics O–U, acceptance criteria, IDs).
- ✅ Acceptance criteria testable (per story; DoD via §8; **incl. cross-tenant isolation tests**).
- ✅ Technical feasibility validated (NetVantage architecture + proven OSS building blocks; enterprise + security + **tenancy** foundations designed-in; shared eBPF/flow substrate).
- ✅ Success metrics defined and trackable (adoption/parity/differentiation/**MSP traction**/F500-readiness/acquisition).
- ✅ Risks identified with mitigation (technical incl. **cross-tenant leakage as highest-severity**, open-data/threat-intel incl. reseller AUP, compliance/procurement/commercial incl. reseller licensing + DPA posture, competitive incl. MSP-platform incumbents).
- ✅ F500-viability addressed (enterprise identity/RBAC/audit/FIPS/SIEM/governance/scale/lifecycle).
- ✅ Differentiation moat explicit (§1.4) for the acquirer story (incl. **the v0.4 multi-tenant/MSP extension**).
- ⬜ Stakeholder alignment — solo founder; revisit license + open/enterprise boundary **+ reseller licensing**, acquirer thesis, **tenant-isolation default (pooled/siloed/hybrid)**, Phase-3 scope, numeric SLOs, remediation-gating policy, threat-intel/reseller AUP, and **provider-plane edition placement** as open items.

---

### Open items for the next pass (pre-sprint decisions)
1. **License + open-vs-enterprise boundary + reseller terms [updated]** — now on the critical path to the MSP channel: define a commercial/reseller license alongside the source-available core; place the provider plane/white-label/metering in the commercial tier; IP-indemnification.
2. **Tenant-isolation default & model [new]** — confirm pooled as the default with siloed/hybrid available; define which controls are required for siloed/regulated tenants (per-tenant keys, per-tenant residency); confirm RLS vs physical-siloing strategy per store.
3. **Provider-plane edition placement [new]** — confirm the provider/management plane, white-label, and per-tenant metering/billing are enterprise/commercial-tier (vs open core).
4. **First-party SaaS? [new]** — confirm that this iteration ships multi-tenancy *for partners/MSPs to operate*, and that probectl-the-vendor running a hosted public SaaS remains deferred/optional.
5. **Billing/metering integration targets [new]** — which PSA/billing systems to export usage to first (e.g., common MSP PSA/billing platforms); metering granularity (agents/tests/ingest/events/AI-calls/retention).
6. **White-label scope [new]** — confirm per-tenant branding scope (logo/theme/custom domain/email/login); whether per-provider master branding is also needed.
7. **Acquirer thesis** — confirm primary archetype (observability vs security vs sovereign-infra vs FinOps vs **MSP-platform**); the v0.3 + v0.4 moat widens this.
8. **Open-data + threat-intel + reseller AUP matrix** — formalize per-source terms before enabling provider mode commercially.
9. **Rust vs Go** for BGP/path/detection hot paths.
10. **Phase-3 scope** — keep/cut RUM (F20), voice (F21), chaos (F47), carbon (F48).
11. **Numeric SLOs** — concrete latency/availability/throughput targets per reference tier (now incl. multi-tenant).
12. **FIPS scope** — confirm FIPS 140-3 mode required (assumed yes).
13. **Remediation-gating policy** — confirm observe-only/human-gated default for F44; approval + blast-radius model.
14. **Branding** — confirm `probectl` (vs NetVantage) across repo/org/domain.
