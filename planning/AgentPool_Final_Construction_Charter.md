# AgentPool — Final Construction Charter

**Candidate v2 · Adversarially revised · NOT approved · No repository or production changes authorized by this document**

> **Build the finished product. Preserve the vision. Prove it works.**
>
> **TradeAgent owns the work. OpenCode executes the agents. AgentPool supplies inference.**

This is the **one construction handoff**, not a replacement architecture, another planning program, or permission to declare a prototype finished. It governs *what must be delivered, how to operate autonomously without redefining it, and what evidence permits the word FINISHED*. The governing contracts own the detailed mechanics.

## 1. The destination

```text
MacBook / Windows desktop / other approved clients       TradeAgent-managed hosts
             OpenCode TUI/headless                           OpenCode workers
                            \                              /
                             authenticated HTTPS inference
                                           |
                        ONE CENTRAL AGENTPOOL — approved KVM 4
                    one provider identity · one route selector
                     one Go authority · one SQLite state core
                    separate, secured remote owner dashboard
                      logical environments A / B / C / ... / N
                                           |
                       certified and permitted physical routes
              API keys · authorized accounts · enterprise · local
```

**Everyday experience.** The owner opens the dashboard, adds a genuinely supported connection using its permitted authentication flow, and sees it placed in the earliest compatible logical environment. They create an AgentPool inference key, copy a version-correct configuration into an independent OpenCode installation, and use an explicitly named model. AgentPool selects the best *eligible equivalent* physical connection for each new request without manual account changes. Owner administration, grants, registration health, actual eligibility, budgets, diagnostics, and recovery are available without SSH for routine use. No upstream key is copied into a client.

**Non-negotiable boundaries:**

1. **Inference provider only.** No agents, sessions, prompts-as-jobs, work queues, tool execution, context management, handoffs, TradeAgent control plane, or OpenCode launcher inside AgentPool. The only changed OpenCode boundary is its model-provider connection.
2. **One central authority.** Independent clients use ordinary authenticated public HTTPS; administration uses separately authenticated HTTPS. No required VPN, sidecar, per-device gateway, managed-client fleet, second selector, or second business-state authority. A single host is a deliberate availability trade-off, not HA.
3. **Independent registrations, logical slots.** Every registration has its own identity, credential generation, lifecycle, policy, and evidenced capacity/billing ownership. First-compatible placement is atomic and ordered as Slots v2 specifies. No arbitrary coded cap; bound actual RAM, active work, spend, and concurrency. Slots are records, not runtimes or quota buckets.
4. **Automatic routing, truthful equivalence.** Route only among certified and authorized implementations of the *requested exact alias/profile*: model/revision/variant, required protocol, capabilities, data/billing policy, and continuation class. Freeze the selected route for each physical attempt. Later eligible requests may use equivalent B when A is blocked; never silently substitute a different model, splice an in-flight stream, migrate nonportable provider state, or replay an ambiguously sent request.
5. **Faithful native OpenCode.** Prove TUI, headless, streaming, reasoning, tool calls, cancellations, auxiliary/title/compaction/subagent paths, provider errors, and observed native retries for each required pinned client/model/protocol profile. For managed hosts, every enabled model-generation path reaches AgentPool or is explicitly disabled **by the host**; AgentPool cannot secure arbitrary independent machines' unrelated egress.
6. **Secrets, authority, money.** Upstream secrets remain server-side. Browser-owner, consumer-inference, delegated-host (if required), and local recovery privileges cannot promote one another. Enforce parent/child grants and shared real capacity/billing scopes without multiplication. Strict accounting is durable *under its declared assumptions*; unknown provider spend, outcomes, and quota remain unknown. No universal external invoice guarantee.
7. **Small, maintainable implementation.** Prefer one confined Go service, one SQLite transactional authority, one proven recoverable secret-custody design, and the approved existing TLS ingress. No Redis, additional broker, per-slot executor, new proxy, network overlay, or distributed HA by default. Extra isolation or processes require a reproduced necessity, measured net benefit, and narrow authorization. Protect unrelated VPS services.
8. **Permitted connectors, not aspirational buttons.** API, authorized account/subscription, enterprise, and local are extensible categories; each *required concrete* mode must use documented, currently permitted flows and pass exact real-client certification. Never impersonate official clients, extract their stored credentials, evade quotas, or call an unauthorized subscription a working route. Optional JEV is asynchronous advice only; routing works deterministically without it.

## 2. Authority and source-of-truth lock

At construction kickoff, inspect actual `main` HEAD, outstanding branches, repository state, and the host **read-only**. Pin the commit and SHA/section identities of these governing sources in the release evidence:

1. `docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md` — inference, aliases, protocol, capacity, accounting, streaming, retry, JEV.
2. `planning/AgentPool_Environment_Slots_and_Multi_Login_Implementation_Contract_v2.md` — identity, environments, atomic placement, credential lifecycle, EN/RT/SEC/PERF proofs.
3. `planning/AgentPool_Central_Deployment_and_Execution_Contract_v3.md` — **currently a proposal**; its approved exact amendments govern central hosting, custody, remote access, recovery, commissioning.
4. `docs/PROVIDER_POLICY.md` — one-owner scope, permission, connector certification.
5. `docs/ORCHESTRATION-STANDARD.md`, `CLAUDE.md`, `ROADMAP.md`, `planning/README.md` — build discipline and inherited P/G evidence gates.

**First authority gate:** obtain explicit owner approval of the central direction and the *exact* v3 parent amendments. Reconcile contradictions in v5.2, Slots v2, README, `CLAUDE.md`, roadmap, and planning index before relying on the amended contract. Change only conflicting clauses and reference the replacement authority; preserve every other protection. A later Markdown date or this charter alone **does not** override parent rules. Preserve a single clause-to-test/evidence mapping rather than rewriting the technical specs into more master documents.

If HEAD changes after the audit, reconcile the actual diff and release matrix before continuing. Tests and evidence refer to an exact commit, binary, client, credential mode, model and environment; previous green evidence does not silently transfer.

## 3. Freeze the one release profile before builders improvise it

The owner and orchestrator must fill and explicitly approve **one short, machine-readable release manifest** (part of the release evidence, not another design essay). At minimum it names:

- **Required outcomes:** first permitted real API/model/profile; **two separately registered, genuinely certified equivalent eligible routes** for automatic A→B demonstration; exactly which authorized subscription/account, enterprise and/or local modes are *required* for the intended finished product; explicitly deferred modes with owner acceptance. A future architecture category does **not** by itself certify a specific commercial subscription.
- **Clients:** exact pinned OpenCode V2 version/package/protocol, concrete supported macOS/Windows/other targets, TUI/headless obligations, auxiliary calls, retry controls, and required TradeAgent integration path. Unattended TradeAgent use is required if selected; *delegated child-grant minting* is not automatically required unless the profile specifies it.
- **Operations:** actual host inventory and one selected native/container packaging, public origins/TLS, listener isolation, owner login/MFA and recovery, credential custody/root-key recovery, managed client grant lifetime/rotation, live smoke test accounts, backup destination, restore/cutover authority.
- **Risk envelope:** permitted read-only access, approved test credentials, maximum aggregate live test spend, allowed infrastructure mutations and deployment/release authority, measured co-host baseline, predeclared numerical performance/footprint/stream and backup/RPO/RTO limits. Never invent grant or spending authorization.
- **Proof contract:** applicable exact EN/RT/SEC/PERF/CD3/v5.2 test IDs, measurable pass/fail thresholds and evidence locations, with explicit `NOT_APPLICABLE` reasons only for genuinely excluded profiles. No test is relaxed after its first failed measurement.

Builders may research and run synthetic fixtures before everything is known; they may not decide missing owner preferences, invent credentials, deploy publicly, or silently convert required capabilities to optional ones. Batch the owner's decisions and grant *bounded, explicit, stage-specific authority* up front where possible so autonomous work is not needlessly interrupted. Production changes and billable calls remain impossible until actually authorized. A prohibited required connector is an **external blocker**, never a license to circumvent terms.

## 4. Construct in four movements

**I — Ground and unlock.** Inspect actual repository and deployment read-only; resolve parent conflicts; pin the real client/package and its effective merged configuration; verify the first permitted connector and the second real equivalent route; freeze the manifest and benchmarks; inventory access, recovery, host constraints, legal/financial dependencies. Record one actionable `BLOCKED` fact for each missing owner/external prerequisite. No production mutation before approval.

**II — Make one real model call perfect.** Build the smallest *private* vertical slice: pinned real OpenCode TUI/headless → authenticated AgentPool → one actually permitted upstream, preserving exact streaming/tool/error behavior. Compare with the same direct physical model. Capture downstream retry attempts and failure ambiguity. Use the final single-selector interface even with one route. Then implement the smallest durable registration/custody, policy, grant, accounting and request/stream state cores supporting that slice. The cheap route is native-compatible forwarding; add translation only for proven required profiles.

**III — Make pooling an owner-operable product.** Complete atomic A/B/C/N placement and credential lifecycle; secure remote first-run and usable dashboard; certified required connectors; immutable indexed route catalog, atomic capacity/budget admission, automatic equivalent A→B on subsequent requests, accurate route/grant diagnostics; copy-ready and verified independent client configuration; unattended TradeAgent→OpenCode where required. Exercise owner journeys, races, revocation, independent concurrent users and failed-enrollment recovery. No dead controls, false status, invisible client prerequisites or dummy integrations sold as finished.

**IV — Commission and hand over.** Deploy the single approved production profile with isolated backend/private admin, existing ingress and co-host protections; run real pinned client tests through actual DNS/TLS, negative/race/crash/security/soak tests and measured performance; rehearse root-key loss, backup/restore, old-host fencing, drain, update/migration and rollback; repair every release-blocking defect and repeat affected proofs. Supply reproducible install, operation, recovery and release evidence. Verify the **running deployed product** after final code/config changes—not only a merge, mock, staging environment or previous build.

These are outcome phases, **not four giant builder tickets**. Order small independent changes by real dependency. Gate shared mutable resources and avoid colliding integration/CI runs.

## 5. Ten immutable release outcomes

The exact owner-approved release profile is `FINISHED` **only if each applicable result is PASS on the shipped build**:

| ID | Required result | Minimum distinguishing proof |
|---|---|---|
| **R1 — Installed** | Exactly one supported **production** AgentPool instance on approved KVM 4; working HTTPS ingress and confined backend; co-hosts unaffected. | Service restart, DNS/TLS/IPv4/IPv6, listener permissions and co-host measurements on real host. Disposable fixture/staging instances are permitted; not alternate production authorities. |
| **R2 — Owner-operable** | Secure bootstrap, login/recovery, real dashboard for adding permitted API/account modes, environments, accounts/routes/models, budgets/grants, diagnostics and maintenance. | Fresh-owner end-to-end UI journeys, denied/disabled actions explained; no routine SSH/developer-only operations or dead required controls. |
| **R3 — Real enrollment** | Every **owner-required** concrete connector authenticates legally, independently and follows Slots v2 first-compatible atomic placement and lifecycle. | Real credential/identity/refresh/revoke, duplicate/cancel/callback/race/crash, eligibility and provider-terms evidence. |
| **R4 — Faithful clients** | Native pinned OpenCode TUI and headless on all required independent client platforms; required unattended TradeAgent integration; no upstream secret on client. | Same-model direct-versus-pool multi-turn streamed reasoning/tool/cancel/error/auxiliary traces; host-owned retry/bypass proofs. |
| **R5 — Automatic pool** | Two independently registered, certified equivalent physical routes; A blocked and later request uses B without human intervention; in-flight A remains pinned. | Real provider traces, independently evidenced capacity-owner scope, continuity-affinity and incompatible-state refusal. |
| **R6 — Financial/failure safety** | Correct grant hierarchy, durable/observational budget modes, shared capacity, reservations, stream failure classification, revocation and refresh under races/crashes. | Negative and concurrency tests, strict holds before sending, honest ambiguous outcomes and OpenCode retry interaction. |
| **R7 — Security** | Owner/inference boundaries, credential confidentiality, egress allowlisting, admin browser security, dependency/licensing/security release review and abuse bounds. | Threat-model tests, isolation from co-hosts, secret/diff scans, no known release-blocking vulnerability or mandatory uncertified connector. |
| **R8 — Recoverability** | Consistent protected off-host backup, separate root recovery, ordinary restart, failed migration/rollback, whole-host restoration and exclusive cutover. | Actual restore rehearsal proving old tokens/old budgets do not regain authority; blocked reopening when former host cannot be fenced. |
| **R9 — Efficiency** | Predeclared p50/p95/p99 added latency, admission/throughput, stream memory, idle footprint, resource caps and acceptable VPS co-host impact. | Paired direct/gateway tests with pinned fixtures, realistic and worst-case concurrency, 1–1,000 dormant slots, reproducible host load. |
| **R10 — Transferability** | Finished installed product handed over with source/build identity, approved support matrix, full sanitized evidence, installation, client setup, admin, update and recovery instructions. | Fresh competent operator can enroll, issue/revoke key, run real client inference, diagnose an ineligible route and rehearse recovery without original builders. |

`FINISHED` cannot be inferred from milestone counts. Every R1–R10 result must be backed by exact version-bound evidence. `BLOCKED` names an external unmet owner/provider/host dependency; `INCOMPLETE` names unfinished or failing engineering work. Neither is `FINISHED`. Optional JEV, experimental WebSocket, optional isolation tiers and unselected connector modes do not become release blockers by accident; a required capability cannot be silently relabeled deferred.

## 6. Autonomous execution without losing the project

**Engineering autonomy:** builders decide implementation internals, maintained dependencies, modular seams, UI styling within journeys, indexing, tests, safe refactors and repair strategies. Do not ask permission for routine compliant engineering. Do not reopen an already-settled architecture merely because another implementation is familiar.

**Reserved owner decisions:** product boundary or mandatory profile changes; a new persistent authority/service or required client component; weakened certification, continuity, protection or exact-model guarantees; material hosting change; new billable exposure, credentials, DNS, production/security mutations beyond granted bounds; final production release unless already explicitly authorized. Neither a builder nor an owner can waive provider terms.

**Make the huge governing library cheap to use.** The orchestrator resolves authority and contradictions once against pinned sources. Each small builder brief names its exact governing clauses, invariants, acceptance IDs, dependencies and negative cases; builders consult the original referenced sections as needed. Do **not** demand that every fresh agent re-read hundreds of pages or invent a competing interpretation. Retain one canonical requirement → test → evidence record, not duplicate planning tracks and review paperwork.

**Avoid the illusion of immortal autonomous sessions.** Follow the repository orchestration standard for worktrees, independent builders/fixers, reviews, red-first protection tests, full gates, verified landing and durable handoffs. But its session-scoped crons, usage-window resets and transient agent IDs are **not** proof that execution resumes after its host/session disappears. Before unattended construction, verify the actual persistence/resume mechanism, available machine/permissions, CI, network, credential/test and deployment access. If there is no durable supervisor, say so; use the supported bounded session workflow and its git-backed checkpoint. Never report expected future work as work performed.

**When a proof fails:** identify the exact violated invariant; reproduce and capture it; repair the smallest compliant cause; rerun affected dependent proofs. An unsupported connector does not justify an account-spoofing workaround; a failed protocol does not justify silent model substitution; a slow transaction does not justify losing strict durability. If a true external prerequisite remains unavailable, leave an actionable blocker with its minimum owner action and continue safe independent work. Fixes are judged against the full verified product, not a reduced test fixture.

## 7. The only acceptable terminal report

Give a compact release certificate with: (a) immutable shipped commit/binary/dependency and production config identities; (b) approved manifest naming exact client versions/models/connectors/permissions and exclusions; (c) R1–R10 PASS with precise synthetic **and** real evidence locations, negative tests and running production smoke; (d) owner/admin/bootstrap and recovery instructions without disclosing secrets; (e) known residual risks and external constraints; and (f) a truthful final verdict. A changed binary, schema, client, credential flow or production config invalidates dependent proof until rerun.

> **The goal is not an impressive plan, a green CI run or a technically interesting router. The goal is a finished, independently usable, secure central model-inference product—working for real OpenCode clients, with automatic certified pooling, and understandable without its builders.**