# AgentPool — Pre-Construction Planning Workspace

**Phase:** PLANNING ONLY — no implementation has been authorized by creating this folder.  
**Project:** [AgentPool](../README.md)  
**Established:** 9 October 2026

This is the dedicated place to **finish and validate the design before implementation begins**. We use it to consolidate the contract, explicitly resolve architectural choices, design reproducible experiments, record evidence, and prepare bounded implementation milestones. A design document or a successful synthetic test is **not** proof that a provider connector works in production.

## Start here

**Proposed final construction handoff (candidate v2; not approved or construction authorization):**
[AgentPool — Final Construction Charter](AgentPool_Final_Construction_Charter.md)

This charter consolidates the intended destination, autonomous construction boundaries, ten release outcomes, and evidence required for a finished product. It does not supersede the governing architecture contracts or authorize infrastructure changes.

**Proposed central deployment addendum (v3; parent amendments require approval):**
[AgentPool Central Deployment & Execution Contract v3](AgentPool_Central_Deployment_and_Execution_Contract_v3.md)


**Current design addendum:**
[AgentPool Environment Slots & Multi-Login — Implementation Contract v2.0](AgentPool_Environment_Slots_and_Multi_Login_Implementation_Contract_v2.md)

This is the starting contract for logical environment slots, multi-login enrollment, credentials, route topology, and their verification. It preserves the product boundary:

> **TradeAgent owns the work. OpenCode executes the agents. AgentPool supplies inference.**

It replaces the *v1 environment-slot addendum within its stated scope*, **not** the rest of the existing architecture.

**Parent and governing documents:**

- [AgentPool v5.2 architecture](../docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md) — parent authority for inference behavior, routing, model identity, grants, capacity, streaming, retries, and JEV.
- [Provider and connector policy](../docs/PROVIDER_POLICY.md) — authorization and certification requirements; each connector and credential mode is assessed individually.
- [Existing roadmap](../ROADMAP.md) — current milestone baseline, to be reconciled with a future approved build plan rather than silently overwritten.

If a planning document conflicts with a governing requirement, **record the exact conflicting clauses and decide explicitly before building**. Do not resolve conflicts through silent omission or by weakening the product vision.

## How planning proceeds

1. **Lock the product invariant.** Every proposed optimization must preserve the A/B/C/N logical slot behavior, independent credential registrations, central automatic inference selection, OpenCode compatibility, and the no-agent-execution boundary.
2. **Identify unknowns precisely.** Record an uncertainty as a falsifiable question, not as an assumption or reason to shrink the product. The v2 contract provides experiments **U01–U20** in §16.
3. **Specify the experiment before execution.** Define the exact pinned versions, host OS, permitted credentials, fixture/real-upstream mode, observed outputs, PASS/FAIL criteria, and the implementation decision following either result.
4. **Record outcomes with evidence.** Distinguish DESIGN, UNKNOWN, TESTED-SYNTHETIC, and PROVEN-ON-PINNED-REAL-CLIENT. Never relabel synthetic success as provider authorization or full OpenCode conformance.
5. **Prepare implementation slices.** Each future task must name its scope, dependencies, acceptance cases, failure behavior, rollback plan, and measurable performance/correctness gate.
6. **Get an explicit construction go-ahead.** Until then, this folder collects plans and evidence; adding a document does not authorize changes to runtime, CI, provider integrations, or security posture.

## Pre-construction readiness checklist

- [ ] Review v2 alongside v5.2 and record any conflicting clauses and their resolutions.
- [ ] Select and pin the initial OpenCode build, model-provider package, and downstream protocol for the first real thin slice.
- [ ] Choose the initial supported provider/credential mode and capture its authorization and capability evidence.
- [ ] Establish the experiment/evidence format and owners for applicable U01–U20 questions.
- [ ] Finalize the logical-slot, enrollment, vault, routing, accounting, and failure contracts without unresolved ambiguities.
- [ ] Define benchmark baselines and objective pass/fail release targets, including security and resource-consumption gates.
- [ ] Produce a dependency-ordered implementation plan, with reviewable PR boundaries and rollback procedures.
- [ ] Obtain explicit approval to leave the planning phase.

## Where future planning material belongs

Keep the **index in this README** current as material is added. Create subfolders only when they have real contents; suggested categories are `decisions/` (accepted/rejected architectural decisions), `experiments/` (protocol probes, results and sanitized evidence), and `build-plan/` (approved task breakdowns and test gates). Do not copy large canonical documents into several places where they can silently diverge.

**Current state:** The Slots v2 implementation contract, proposed Central Deployment v3 contract, and proposed Final Construction Charter v2 are filed here. They do not imply an owner-approved construction release, completed experiments, permission to change infrastructure, or a running implementation.
