# AgentPool: Feasibility, Continuity, and the Limits of Inference Capacity Pooling

## An independent, evidence-based technical research paper

**Research date:** 8 October 2026  
**Architecture examined:** AgentPool *The Switch*, principally v5.2.0, with eight preceding supplied design documents  
**Research type:** Open-source landscape analysis; public-source technical verification; architecture falsification; implementation-readiness assessment  
**Status:** Research and design evaluation — **not** a production certification, independent security audit, legal opinion, completed integration test, or audited financial model  
**Intended audience:** Founders, investors, senior engineers, architecture reviewers, and technical due-diligence teams

---

## Abstract

A growing class of autonomous software uses coding-agent runtimes to perform extended sequences of model requests, tool calls, and decisions. Its productive execution can be disrupted when any one upstream inference identity is rate-limited, unavailable, or otherwise ineligible. AgentPool proposes a small, independently deployable inference gateway beneath such runtimes. The runtime continues to own its agent loop, session history, tools, and task lifecycle; the gateway selects an eligible, compatible, authorized upstream inference route for **each new request**, observes capacity, and relays responses. Its defining promise is *logical agent-session continuity across suitable request boundaries*, not seamless transfer of an active model stream or hidden provider state.

This paper tests the proposal against vendor documentation, public source code, issue trackers, and community reports. It finds that its basic mechanisms are **technically feasible and substantially precedented**. OpenCode exposes a configurable custom-provider interface; independent open-source projects implement multi-credential rotation and local HTTP inference proxies; and established gateways already implement routing, health filtering, cooldowns, and streaming. The critical unproven parts are the composition of those mechanisms into a verified AgentPool deployment, the portability of provider-specific continuation state, safe client-side retry behavior, concurrency and crash correctness, and the security of its credential boundary.

A separate limitation is decisive for the business thesis. **Technical access to consumer-subscription inference is not equivalent to permission to aggregate subscription capacity.** As of the research date, OpenAI's Sign in with ChatGPT Terms explicitly restrict subscription-plan pooling and account rotation; Anthropic separately restricts third-party intermediation of consumer Claude credentials. Engineering success cannot eliminate a contractual restriction. Accordingly, AgentPool should be presented as an infrastructure design for **appropriately authorized API, enterprise, local, and individually approved connector capacity**, with further subscription integrations requiring provider-specific evidence and authorization.

Finally, the proposed JEV-based operational advisor is plausible as an optional classifier of ambiguous infrastructure conditions, but no measured evidence shows that it improves routing beyond deterministic observations and ordinary health heuristics. JEV should remain off the critical inference path and outside the definition of a successful core release.

**Principal conclusion:** AgentPool warrants a focused engineering validation program. The architecture is plausible; the generalized continuity claims, consumer-subscription economics, and competitive differentiation are **not yet demonstrated**. The correct near-term objective is a reproducible, live OpenCode-to-AgentPool-to-two-authorized-routes proof, with strict stream, context, accounting, and concurrency acceptance criteria.

**Keywords:** inference gateway; agent runtime; OpenCode; request-level routing; account selection; context portability; streaming; model provider; capacity admission; rate limits; JEV; reliability; technical due diligence.

---

# 1. Research question and thesis

The question is not simply whether one application can use more than one model API key. That capability has been available for years. The harder question is whether a deliberately narrow gateway can maintain a **stable inference interface for long-running agent software** while safely selecting among multiple independently admitted identities, model deployments, and credential modes — without becoming a second agent orchestrator or concealing changes that affect correctness, confidentiality, or cost.

The evaluated AgentPool proposal asserts five essential product properties:

1. **Provider-only placement.** OpenCode performs model calls and agent work; AgentPool supplies inference transport, route selection, authorization, and bounded operational accounting. TradeAgent or another host remains responsible for its agents' objectives, external effects, and job recovery.
2. **Automatic pool selection.** For each newly admitted compatible request, AgentPool selects an authorized physical route without requiring an operator to press a switch button. A manual pin exists as an override, not the ordinary routing mechanism.
3. **Credential-mode extensibility.** API keys, service identities, enterprise capacity, locally hosted endpoints, and eligible subscription-backed integrations are represented by separate connectors with verified capabilities, not forced into a universal OAuth abstraction.
4. **Conservative continuity.** Upstream route changes occur at safely established request boundaries. An already streaming response is not silently spliced into a different account. Nonportable provider-side state produces an explicit stop or client-managed recovery.
5. **Small and deterministic control path.** Routing uses scoped, local evidence and bounded reservations. An optional asynchronous JEV advisor may influence *soft preference* after validation, but never permissions, quota arithmetic, financial controls, or retry authorization.

The central hypothesis can be stated in falsifiable form:

> For a specified OpenCode build, protocol, model, and set of authorized equivalent upstream routes, AgentPool can reselect the physical inference identity between consecutive model requests within the **same OpenCode-owned session**, preserve supported context and tool semantics, and add acceptably little overhead — without taking ownership of the agent's execution lifecycle.

A stronger proposition — that the agent will *always* remain uninterrupted across arbitrary upstream failures and arbitrary subscription identities — is **not** part of the defensible hypothesis.

## 1.1 The normative topology

```text
Human / application / TradeAgent
           |
           | normal agent-control API
           v
       OpenCode
   agent loop | tools | sessions | compaction | model request construction
           |
           | configured model provider: agentpool/<explicit-alias>
           v
       AgentPool
   local caller auth | grant check | eligible route selection
   capacity/cooldown | request forwarding | streaming | scoped observations
           |
           | ONE immutable admitted upstream route per dispatched request
           v
   Authorized API / enterprise / local / separately approved connector
```

An owner-only management interface is permitted alongside the inference endpoint for registration, grants, route inspection, and diagnostics. It does not turn AgentPool into an agent-control API. The meaning of "standalone" is therefore **an independently operated service**, not a service that launches OpenCode agents.

## 1.2 Scope exclusions

This paper does not propose that AgentPool should own agent schedules, worktrees, durable task graphs, tools, trading permissions, JEV-driven agent handoffs, prompt memory, or model training. It does not assume arbitrary model families are interchangeable, that all provider subscriptions are suitable for proxy access, or that a local API facade guarantees production-grade financial accounting.

---

# 2. Method, evidence classes, and limitations

The review cross-checked nine supplied AgentPool architecture files against current public evidence. The documents are **design evidence**: they establish intended behavior and internal consistency, not successful deployed behavior. Outside evidence was grouped as follows.

| Class | Meaning | What it can establish | What it cannot establish |
|---|---|---|---|
| **A — Vendor contract/documentation** | Official API, terms, or runtime documentation | A documented interface or stated permission boundary | That an AgentPool integration works on a specific pinned build |
| **B — Implemented open-source mechanism** | Public source containing relevant router, proxy, streaming, or credential logic | The mechanism exists in code and can be examined | Its uptime, test pass rate, legality, or suitability for AgentPool |
| **C — Reported field failure / maintainer test** | GitHub issue, PR, reproducer, maintainer benchmark | A credible specific failure mode or design tradeoff | Failure prevalence, current incidence, or a universal provider rule |
| **D — Independent integration proof** | Pinned artifact, real endpoint, reproducible traces and adversarial tests | Evidence that AgentPool fulfills a named capability | Generalizing beyond tested runtime/provider/credential mode |
| **E — Product-specific permission** | Applicable binding terms or written provider authorization | Whether a credential/operating mode may be used for a specified purpose | Correctness or reliability of the implementation |

**Classes A–C were reviewed in this study; class D was not produced for AgentPool; class E was evaluated from published conditions but not from any private commercial agreement.** No paid-account experimentation, actual AgentPool latency measurement, production incident analysis, executed `go test`/`npm test`, or legal counsel review was performed.

Sources are labeled **[V]** for vendor/official docs, **[O]** for open-source projects, and **[F]** for field reports; direct links and specific qualification notes appear in the reference section. Most GitHub links point to mutable default branches. These are suitable for discovery, **not** reproducibility; the engineering team must freeze commit IDs and package versions in its future test evidence.

**Important interpretation rule:** where a repository claims an ability in its README, this paper describes it as *project-documented* unless source-level mechanisms were also checked. Even code inspection remains weaker than a pinned end-to-end live test. GitHub issue reports are used as failure-mode evidence, not as statistics about reliability.

---

# 3. Results from the open-source landscape

## 3.1 OpenCode supplies the necessary provider seam

OpenCode V2 documents a `providers` configuration containing an explicit runtime `package`, `settings.baseURL`, and per-provider model catalog. It includes native and compatible provider packages for OpenAI Responses, Anthropic, Google, and several other transports. It also allows endpoint overrides for a compatible proxy. This independently supports the **architectural placement** of AgentPool beneath the OpenCode agent loop [V1].

The conclusion should be stated precisely. A valid OpenCode provider configuration proves that the runtime is designed to send certain model requests to a configurable endpoint. It does **not** automatically prove (a) every internal compaction/title/auxiliary/subagent request follows that endpoint; (b) a particular AgentPool stream carries tools and reasoning faithfully; (c) OpenCode preserves every provider-specific continuation artifact; or (d) the host program can migrate from a different coding-agent runtime without changes. The required proof is therefore not "a custom `baseURL` worked" but "the pinned OpenCode worker behaved equivalently through AgentPool on a complete, observable multi-turn fixture."

OpenCode V2 also documents timeout and retry behavior, including a plugin retry hook whose decisions can be adjusted. Crucially, its documentation says that the hook does **not** expose how a following attempt is internally performed [V2]. Thus, a retry hook is a mechanism worth testing, not proof that a repeated attempt is safe after ambiguous output or tool activity.

## 3.2 `codex-multi-auth`: direct evidence of local request-level rotation

The public `ndycode/codex-multi-auth` project is the closest conceptual antecedent to the original Switch proposal. Its README documents multiple local ChatGPT sign-ins, runtime selection and an optional loopback proxy for official Codex sessions [O1]. More importantly, its current TypeScript source exposes an HTTP server and request handler, request-body limits, bearer-authenticated routing logic, account selection/commit handling, and forwarding through selected routes [O2].

This source inspection supports a substantive conclusion: **request-level credential choice below a running native agent is not a novel, unimplemented idea**. A real codebase attempts it today. The source also makes clear how much caution it requires: routing mutexes, concurrent selection state, request normalization, client isolation, response termination, bounded retries, and account-state persistence are all implementation concerns.

The project is **not** evidence that OpenAI authorizes unrestricted ChatGPT-plan pooling. Its own README cautions that it is an independent personal-development tool and recommends official API infrastructure for commercial/production use [O1]. Nor does its existence prove full portability of hidden reasoning and all server-side continuation mechanisms.

## 3.3 `oc-codex-multi-auth`: an integrated alternative

`ndycode/oc-codex-multi-auth` provides another model: instead of a standalone inference service, account management and routing are incorporated into an OpenCode plugin. Its documentation describes an OpenCode V2 adapter, an account pool, credential refresh, routing, and diagnostic commands [O3].

This is an important competitive and architectural challenge. If the only intended consumer were OpenCode, a plugin may be simpler than an independently authenticated daemon with its own management API and state database. **AgentPool must justify the service boundary** through concrete benefits such as multiple authorized clients sharing a consistent routing policy, custody of upstream credentials outside worker environments, per-client admission, auditable telemetry, and replaceability of the agent runtime. Those are reasonable benefits, but their *incremental value* has not been measured against the plugin alternative.

## 3.4 CLIProxyAPI: the closest Go service analogue

`router-for-me/CLIProxyAPI` exposes compatible inference endpoints and advertises multiple credential/account selection modes across model/provider families. Its public Go selector and cooldown code demonstrate the relevant class of infrastructure: candidate eligibility, failures, resets, streaming, and multi-route selection [O4]. It is especially relevant because its implementation language and local-service shape are close to the AgentPool v5.2 proposal.

The existence of CLIProxyAPI weakens claims that a generic Go multi-provider proxy would constitute new technology or a defensible moat. It strengthens the claim that such a system is technically achievable. The project's issue tracker also supplies unusually concrete negative evidence about cooldowns and quota scope, discussed below.

CLIProxyAPI's inclusion of consumer OAuth integrations is not contractual evidence that AgentPool may offer them. Conversely, its API-key and compatible endpoint architecture remains a useful implementation reference independently of the legal status of particular connector modes.

## 3.5 LiteLLM and Portkey: routing is an established infrastructure category

LiteLLM and Portkey already expose broad AI gateway capabilities, including provider abstractions, deployment selection, compatibility surfaces, usage accounting, and failover [O5, O6]. These projects show that model aliases, healthy-deployment preference, bounded retries, and streaming relay belong to a mature design space.

The relevant conclusion is not to duplicate those platforms wholesale. AgentPool deliberately aims at a **narrow local inference switch** rather than a general enterprise gateway, prompt-observability SaaS, billing platform, or policy suite. Its attractiveness depends on whether that smaller trust and correctness boundary produces a materially simpler, more reliable tool for agent workloads.

## 3.6 Competitive conclusion

| Alternative | Strength in the evidence | Remaining gap relative to the AgentPool thesis |
|---|---|---|
| `codex-multi-auth` | Working code for local credential routing beneath Codex | Narrow native-Codex orientation; not proof of all-provider continuity or commercial subscription entitlement |
| `oc-codex-multi-auth` | Native OpenCode plugin integration | Does not by itself establish the value of a runtime-independent shared local gateway |
| CLIProxyAPI | Go proxy, multiple protocols/accounts, routing/cooldowns | Different scope and guarantees; compatibility is per actual request/protocol |
| LiteLLM | Mature multi-deployment routing and observability | Much broader platform surface; may be excessive for a strictly local switch |
| Portkey Gateway | Broad gateway ecosystem and guardrails | Similar general-purpose scope tradeoff |
| **AgentPool (proposed)** | Tight provider-only split; explicit safe boundaries; permission/continuity requirements | **Not yet proven in running integration; moat and economics not established** |

The market is **not empty**. The plausible differentiation is *verified, trustworthy continuity for many simultaneous native agents operating against legitimately provisioned capacity*, not invention of account rotation.

---

# 4. Field evidence: failure modes that constrain the design

Public issues make the evaluation more precise than a review based solely on vendor promises.

## 4.1 Cooldown clocks can lose capacity after suspend/resume

CLIProxyAPI issue **#5590** documents a reported macOS scenario in which a quota cooldown outlived the upstream reset after host sleep; its author included a deterministic clock-representation regression and identified Go monotonic-versus-wall-clock behavior. The issue was subsequently **closed and labeled Fixed**, so it should not be represented as an unresolved defect in current CLIProxyAPI [F1].

**Engineering consequence:** provider-advertised absolute reset times must be interpreted consistently across suspend/resume; state restored after a restart cannot trust elapsed-clock assumptions blindly. An expired cooldown is an opportunity to re-admit a route subject to current evidence, not a promise of quota restoration.

## 4.2 Quota errors can be model-scoped, not identity-global

Issue **#5619** reports that a Codex-model quota exhaustion could disable all models for the same credential, including others with available capacity; the report supplied selector code paths and concrete request logs. It is **closed and labeled Fixed** as checked, so it is historical negative evidence rather than a current universal claim [F2].

**Engineering consequence:** error classifications must distinguish the exact observed scope (model, route, app, credential, capacity owner, project, or unknown). An error body lacking an explicit scope is not a license to invent account-global exhaustion. A conservative, bounded recheck of another model may be better than globally suppressing all routes under the credential.

## 4.3 Cached quota recovery is not always the provider's true state

Issue **#5611** concerns an account staying cooled down after upstream capacity was restored [F3]. This illustrates why predictions and previously observed reset windows must have provenance and expiry, and why an operator should be able to identify stale observations without flushing unrelated healthy states.

**Engineering consequence:** use TTL-tagged observations, source-specific precedence, and a bounded revalidation policy rather than treating old health classifications as perpetual truth.

## 4.4 Multi-client access needs client isolation

Issue **#3467** highlights a possible failure in which one local proxy client could drive excessive traffic across pooled upstream identities [F4]. This is not an incidence estimate, but it exposes a relevant threat model: once a gateway is shared by multiple agents, a single worker can consume finite capacity that another relies on.

**Engineering consequence:** owner-approved per-client grants, concurrency limits, budgets where applicable, resource ceilings, and a well-defined rejection path are core correctness mechanisms, not extraneous enterprise features.

## 4.5 A successful HTTP connection need not mean a completed model turn

OpenCode issue **#21173** describes a Responses request routed through a custom proxy whose output or tool execution could stall without completing cleanly [F5]. Issue **#23767** describes Anthropic provider-specific tool fields rejected by certain proxy configurations [F6]. Both are reports concerning particular runtimes and may not reproduce against a current V2 build.

**Engineering consequence:** wire conformance includes event ordering, terminal stream markers, error events after headers, tool-call payloads, cancellation propagation, and behavior on a chunk stall. Passing a plain-text prompt or receiving HTTP 200 is not sufficient.

## 4.6 Reports are design evidence, not quantitative reliability statistics

Issue reports may be fixed, duplicate, incomplete, configuration-specific, or erroneous. They support a **test inventory**, not a probability of failure for AgentPool. An investor presentation must not turn a dozen public bug reports into a numerical defect rate. Instead, take each report's concrete reproduction steps, classify the invariant it threatens, and create a version-pinned regression fixture.

---

# 5. Continuity: what a safe switch does and does not preserve

The word *continuity* carries several different technical meanings. They must not be conflated.

| Layer | Ownership | Safe change across authorized physical routes? |
|---|---|---|
| Host task identity and orchestration | TradeAgent / calling application | Usually retained because AgentPool does not own them |
| OpenCode process/session ID | OpenCode | Can remain unchanged if the runtime tolerates the new request route |
| Client-visible messages and tool history | OpenCode | Often reconstructible; must verify actual request construction |
| Model capability and semantic identity | OpenCode alias + verified provider route | Only for individually certified equivalent routes |
| Provider-issued response/conversation IDs | Upstream provider | **Unknown until tested**; often tied to a state domain |
| Encrypted reasoning and other opaque items | Provider-specific protocol | **Not assumed portable** among accounts, projects, regions, endpoints or models |
| Prompt cache | Provider and deployment | May be lost after switching; performance/cost consequences are possible |
| Already-emitted stream chunks | Current upstream request/client | **Not transferable by merely changing credentials** |
| Tool side effects | OpenCode / host environment | Cannot be safely replayed by the gateway |
| Billing and entitlement | Selected provider principal/project | Must be explicitly permitted and attributed to the selected route |

OpenAI's Responses guidance distinguishes client-supplied history, server-side continuation via response identifiers, and stateless encrypted reasoning workflows [V3, V4]. Community reports about encrypted-content verification across inference backends show why a generic portability claim would be premature [F7]. None of these sources prove that every encrypted item is nonportable; they establish that portability requires **the exact state domain and protocol** to be validated.

The narrow defensible target is:

> A completed request may use route A; after its client-visible outcome is settled, a *subsequent compatible request* may use route B in the same OpenCode session without restarting the agent, provided the request does not depend on unportable upstream state.

This is considerably more achievable than "the model never disconnects". It also avoids pretending that a streaming response from A could be seamlessly completed by B after partial output.

## 5.1 Failure and retry semantics

An AgentPool dispatch should freeze its route/identity for the duration of that attempt. When an upstream error occurs:

- **Before upstream dispatch:** a new route may be chosen safely because no request has yet been sent.
- **Definitive pre-output rejection:** AgentPool may record the failure, cool the correct scope, and return a typed outcome. A separate, verified OpenCode-owned retry can then be admitted as a **new request** through B. Whether the client can safely replay is an **empirical** question.
- **Sent, but response ambiguous:** the gateway must not assume the request failed; billing or provider-side work may have happened.
- **Partial stream / model tokens / tool instructions emitted:** do not transparently replay or splice. Surface the failure and allow the runtime/host to decide recovery.
- **Opaque state bound to A:** reject migration even if B is otherwise healthy, unless the specific provider state is demonstrably portable.

An OpenCode retry hook is not by itself enough to prove the second case [V2]. The benchmark must instrument native retry behavior, cumulative attempt budgets, tool effects, finalization, and SDK-internal retries.

---

# 6. Capacity allocation and concurrency: minimum sufficient engineering

It is important to separate **automatic selection** from **perfect prediction**. A gateway can do useful proactive routing with documented quota headers, hard cooldowns, recent errors, local in-flight reservations, and stable preference. It cannot infer an exact universal inventory of "remaining tokens" from incomplete subscription interfaces.

Google's Gemini API explicitly applies several quota dimensions **per project rather than per API key**, and quota dimensions vary by model [V5]. Therefore a pool of 100 API keys may be only one effective capacity owner. Similar ambiguity can arise with subscriptions that share entitlements across applications, workspaces, or registrations.

The minimal design should use four distinct concepts:

- **Registration:** an authenticated credential or connection definition, never assumed to represent an independent quota by itself.
- **Physical route:** a verified destination + effective model/protocol/credential/billing identity.
- **Capacity owner and scope:** the upstream quota namespace, which may be shared by several routes and may have separate per-model/per-window counters.
- **Inference attempt:** one immutable selected route and its request/response/failure provenance.

A bounded deterministic admission path is:

```text
1. Authenticate local caller; check workload grant and logical alias.
2. Find indexed routes certified for the exact requested model/protocol/capability.
3. Reject disallowed model, data authority, credential, billing, or state domain.
4. Exclude known hard cooldowns and insufficient known headroom.
5. Account for live local reservations and concurrency at the correct owner/scope.
6. Select by a simple deterministic sticky/fair strategy.
7. Atomically freeze the route and reserve only what the local design can justify.
8. Stream the response; finalize once; update evidence without inventing usage.
```

When upstream capacity is **unknown**, the gateway can perform one bounded permitted attempt, if other admission conditions are satisfied. It cannot promise the attempt will succeed or treat uncertainty as infinite available quota.

A **small local ledger** is defensible for race and crash correctness. However, the v5.2 draft's SQLite schema is best treated as a *logical model*, not an already complete atomic-reservation implementation. A production contract still needs precise durable reservation/settlement state, handling for sent-but-unanswered attempts, epoch fencing for revoked credentials, and prevention of double release after crash/restart.

## 6.1 Fast path and future proofing

The immediate request path should not need a remote quota check, credential refresh, JEV model call, expensive global registry traversal, or a gateway-wide stream drain. Scheduled credential maintenance and telemetry can populate small local snapshots. The architecture should be **unbounded in registration count as a data model**, while explicitly bounded in memory, descriptors, queue depth, concurrent dispatch and permitted spend.

The proposition "technically unlimited accounts" should mean **no arbitrary `MAX_ACCOUNTS=2` product restriction**, *not* unlimited throughput, zero overhead, or unlimited independent provider quota.

---

# 7. Commercial and policy constraints

This section is about the documented operating conditions, not a legal judgment about all possible private agreements or future provider programs.

## 7.1 OpenAI Sign in with ChatGPT (SIWC)

OpenAI's official **Sign in with ChatGPT Terms**, dated **29 September 2026**, restrict plan usage to the connected application, require user control and consent for background use, and prohibit pooling ChatGPT-plan usage or rotating accounts to bypass limits [V6]. These provisions apply independently of the question of whether the technical OAuth registration/refresh mechanism supports multiple accounts.

The terms also require that eligible users be able to use their plan through an SIWC app without having to pay the app maker or upgrade to an app paid tier [V6]. Thus both the *routing mode* and the proposed *commercial distribution model* need review before claiming compliant SIWC-based revenue.

A separate OpenAI SIWC DevKit repository carries a **noncommercial license** for OpenAI-authored code and documentation [V7]. That license question is distinct from the service's plan-use restrictions. A public GitHub repository is not automatically unrestricted source for a commercial service.

**Implication:** Neither "one owner bought all the accounts" nor "the switch happens manually" establishes a universal exception to the applicable terms. Consumer-subscription pooling must not be treated as a cleared feature based only on technical access. A different, independently authorized API arrangement remains a distinct operating mode.

## 7.2 Anthropic Claude/Claude Code

Anthropic's current Claude Code legal and compliance documentation draws a clear distinction. End users can authenticate with their own subscriptions to use the **unmodified native Claude Code application** under applicable terms. Developers offering third-party products that use Claude capabilities should use authorized API or supported cloud credentials; the published guidance prohibits third-party applications from intermediating Free/Pro/Max plan credentials on users' behalf [V8].

**Implication:** The design cannot infer that because Claude Code runs natively under a subscription, AgentPool may extract or redirect that subscription's authentication into its own model-provider interface. Those are different products and authorization flows. Anthropic API and approved enterprise/cloud routes remain viable connector targets subject to their own terms.

## 7.3 Open-source success is not entitlement evidence

The presence of community tools using consumer account credentials is evidence of **technical implementability** and potential user demand. It is not proof of official permission, commercial stability, support, or indemnification. In investor documentation, the technical feasibility claim and the authorization claim must be separate line items with different supporting evidence.

## 7.4 A permission-correct product definition

A future-proof formulation is:

> AgentPool provides automatic request-boundary inference selection across independently registered, explicitly authorized, compatible capacity sources; credentials, data authority, model capabilities, quota scopes, and financial exposure remain isolated and attributable.

This allows substantial engineering value through API-key pooling, multiple paid API projects where genuinely independent, enterprise deployments, local hardware, and separately authorized provider programs. It **does not promise** that every consumer subscription can become a generic model API.

---

# 8. JEV as an optional operational advisor

The TypeSafe API documents a non-text-first typed decision surface, including `noul` probability judgments, `choice` over alternatives, and `score` [V9]. Its published JEV 1.13 limitations advise against relying on it for precise arithmetic or date/time comparisons [V10]. This is consistent with v5.2's proposal to keep deterministic control over quotas, reset times, reservations, permissions, and billing.

A plausible use is the qualitative assessment of **sanitized infrastructure summaries**: whether a pattern of failures resembles a temporary provider-wide issue, isolated credential trouble, or insufficient evidence. That is a hypothesis, **not an existing validated operational predictor**. A probabilistic answer to a vague semantic question is not automatically a calibrated probability of rate-limit failure during a particular horizon.

The proper integration is optional, off-path, and reversible:

1. The deterministic selector establishes the **eligible set** and remains fully functional without JEV.
2. A small asynchronous advisory module receives only permitted, sanitized provider-health classifications; no prompts, raw provider errors, user files, credentials or agent memory.
3. JEV returns an expiring **soft penalty** affecting ranking only among already authorized, compatible eligible routes.
4. It cannot authorize an account, enlarge quota, override a hard cooldown or retry restriction, change billing/model identities, or cancel an in-flight stream.
5. It is first tested in shadow mode against deterministic baselines and held-out provider events. Enable influence only on measurable improvement in the selected business metric without unacceptable false route avoidance, cost, unfairness, or latency.

In particular, this AgentPool JEV advisor must not take over the distinct TradeAgent JEV responsibility of detecting *agent-context degradation* and escalating to managers for checkpoint/handoff. Those are different domains with different evidence and authority.

**Research conclusion:** JEV is an appropriate **future experiment**, not a necessary component of a viable initial AgentPool.

---

# 9. Claim assessment: what can responsibly be said

Status definitions: **Documented** = an upstream interface is officially specified; **Implemented elsewhere** = open-source code or project mechanism exists; **Reasonable design** = credible engineering inference; **Unverified** = needs AgentPool live tests; **Restricted** = materially constrained by published provider rules; **Rejected** = too strong or contradicted by evidence.

| Claim | Assessment | Appropriate formulation |
|---|---|---|
| OpenCode can call a custom inference provider | **Documented** [V1] | The extension seam exists; pin and test the exact client and package |
| A provider endpoint can sit below the agent loop | **Documented / reasonable** [V1] | A separation supported by the runtime model, not an integrated AgentPool result |
| Multiple credentials can be selected per request | **Implemented elsewhere** [O1–O4] | Proven engineering pattern, connector and authorization specific |
| One logical alias can map to several API deployments | **Implemented elsewhere** [O4–O6] | Only equivalent, certified routes qualify |
| Automatic pre-dispatch avoidance can work | **Supported conditionally** [O1, O4, F1–F3] | Useful with correctly scoped and sufficiently fresh evidence |
| Every predictable 429 can be eliminated | **Rejected** | Unknown/hidden capacity and races prevent this guarantee |
| There can be arbitrarily many registration records | **Reasonable design** | Dynamic registry without a fixed product ceiling; bounded resources |
| Concurrent workers can share a gateway | **Implemented in category; AgentPool unverified** | Requires per-client controls, races/fairness and load tests |
| Running streams can be spliced to a new account | **Rejected as general guarantee** | Next-request switch only; partial-stream error surfaces honestly |
| Retained OpenCode session ID guarantees state portability | **Rejected** [V3, V4, F7] | Session lifetime and model/provider state are distinct |
| Encrypted reasoning migrates across all accounts | **Unverified** [V3, V4, F7] | Test exact protocol and state domain or disallow movement |
| OpenCode retry hook makes failover safe automatically | **Unverified** [V2] | Hook exists; internal retry semantics still need conformance |
| All auxiliary inference follows a single setting | **Unverified** | Instrument compaction, titles, subagents and hidden model calls |
| All providers can be translated through one facade | **Rejected as universal guarantee** [V1, F6] | Add protocol/model connectors incrementally with conformance |
| Each API key provides an independent quota | **Rejected** [V5] | Track verified shared capacity owner/project/workspace scope |
| Consumer paid account rotation is automatically permitted | **Restricted** [V6, V8] | Permission and technical function require separate verification |
| OAuth login implies perpetual credential validity | **Rejected** | Refresh and revocation are operational failure states |
| JEV provides typed probabilities | **Documented** [V9] | Probability interface exists |
| JEV predicts upcoming quota exhaustion accurately | **Unverified** [V9, V10] | Requires labeled provider-specific calibration |
| JEV improves reliability beyond simple heuristics | **Unverified** | Requires matched-workload and held-out evaluation |
| A small local Go + SQLite implementation is feasible | **Reasonable design** [O4] | Sound direction; AgentPool correctness/latency not yet demonstrated |
| Account switching is a novel patent-like technical idea | **Rejected as investor claim** [O1–O6] | Defensibility must come from measurable product execution |
| AgentPool v5.2 is implementation-certified | **Rejected** | Its own acceptance matrix marks major gates unverified |
| Subscription routing is impossible in every product mode | **Rejected as overgeneralization** | API, enterprise, local, and individually approved programs differ |
| A focused validation program is worth conducting | **Supported recommendation** | Conditional engineering GO with explicit stop criteria |

---

# 10. Architecture consequences: preserve the idea, reduce speculative machinery

The strongest part of v5.2 is its **correct ownership boundary**, not its quantity of subsystems. A lean implementation must keep the guarantees that external evidence shows matter while avoiding a second scheduling platform.

## 10.1 Requirements that survive diligence

- A stable OpenCode model-provider integration, with version-pinned protocol support.
- Explicit model alias and capability checks; no silent cross-family or billing substitution.
- A dynamic registry of independently authenticated routes; authorization and real quota owner kept distinct.
- Fast deterministic selection and bounded concurrency using local observations.
- One physical upstream commitment per inference attempt.
- Exact stream/error/cancellation handling, including ambiguous response and post-header errors.
- Correct scoping and expiry of rate-limit evidence, including after suspend/restart.
- Safe no-capacity outcomes and appropriately bounded client retries.
- Credential custody and operator controls isolated from untrusted agent/plugin inputs.
- Minimal audit and crash state sufficient to prevent double release, stale authorization, and false billing certainty.

## 10.2 Features that should remain staged

- Cross-provider protocol translation beyond one proven native/compatible route type.
- Subscription-backed OAuth connector variants not yet independently authorized or certified.
- Sophisticated forecasts of hidden token usage or probabilistic route deterioration.
- A production JEV predictive subsystem, its training/calibration pipeline, and per-provider models.
- A universal compatibility matrix before the first two-route switching demonstration.
- Distributed scheduling, transcript storage, agent process lifecycle, semantic task resumption and trading-domain management — **outside AgentPool entirely**.

## 10.3 Build-versus-reuse decision

Before investing in a ground-up gateway, compare a thin implementation against adopting or extracting reusable pieces from CLIProxyAPI and other permissively licensed gateway software. Apply a concrete decision framework: exact protocol fit; security isolation; difficulty of removing unwanted features; upstream maintenance exposure; licensing and supply chain; minimal state footprint; testability; portability; and lifetime maintenance cost. A bespoke Go service is justifiable only if it yields a measurably clearer, safer, more maintainable boundary than adapting an existing engine.

**No evidence yet establishes that custom-building AgentPool is less expensive than reusing existing software.** This must be evaluated, not assumed.

---

# 11. Falsifiable validation plan

The proposed tests are **exit gates**, not claims that tests have already passed. Every result must name the exact OpenCode build, model/provider package, route versions, endpoint and credential mode, OS, test code revision, and sanitized trace bundle.

### Gate 0 — Authorization, source and licensing

**Objective:** select an operating mode whose technical and commercial use is clear. Identify two independently admitted API/enterprise/local routes for the same effective model. Capture their applicable terms, billing owner, allowance scope, and model capabilities. Freeze third-party code licenses and commits. **Fail** if a required mode relies on unspecified consumer-subscription intermediation.

### Gate 1 — Single-route agent parity

**Objective:** run one genuine OpenCode headless and/or TUI agent through a directly supported upstream and through AgentPool, using the same logical model and a multi-turn fixture. Observe tool calls, messages, reasoning items where exposed, compaction, cancellations, final stream events and errors. **Pass** only on defined equivalence of the supported protocol and tool semantics, not byte-for-byte identical stochastic model outputs.

### Gate 2 — Automatic A→B without an agent restart

**Objective:** start one persistent OpenCode session; finish a request through A; make A ineligible using provider-evidenced or controlled test capacity state; send the next request through B without manual selection or runtime restart. Verify context, account attribution, model identity and stream completion. Repeat with injected nonportable state; **pass** only if unsafe switching is rejected explicitly.

### Gate 3 — Error and retry adversarial suite

**Objective:** inject definitive pre-output 429, per-model/project cooldowns, 401/403, 5xx, `Retry-After`, timeouts after send, incomplete SSE, tool call emission, cancellation, and malformed provider events. **Pass** only when outcomes remain typed and accurate, no attempt is duplicated without a proven safe client policy, and the correct capacity scope is modified. Include upstream errors represented *inside* streams after HTTP headers.

### Gate 4 — Multi-agent concurrency

**Objective:** validate 1, 20, and 100 workers and 1, 2, 100, and 1,000 registered routes in repeatable synthetic and real-limited scenarios. Measure admission p50/p95/p99, first-token overhead, memory/FD use, fairness, cross-client starvation and shared-owner reservation correctness. **Pass** against hardware-specific thresholds established from direct-provider baselines, never a fabricated universal millisecond number.

### Gate 5 — Lifecycle and security

**Objective:** kill/restart during each dispatch phase; force SQLite busy/full/corrupt; sleep/wake across quota reset; revoke a credential during refresh; inject malicious repo/plugin configuration; attempt raw-secret leakage, direct-provider egress bypass and unauthorized local API access. **Pass** only with bounded failures, no privilege escalation, no double-released reservations, and honest unknown billing/usage state.

### Gate 6 — Additional protocol

**Objective:** certify one independently permitted Anthropic/Google/local route after the first protocol is proven. **Pass** only on the new protocol's tools, special fields, reasoning/state, streaming, usage, and errors. Multi-provider catalog presence is not a substitute.

### Gate 7 — JEV research, separately optional

**Objective:** evaluate advisory-only JEV against deterministic `Retry-After`, capacity/cooldown, EWMA or simple fault-rate baselines on a held-out time-separated dataset. Metrics include false-positive avoidance, preventable 429s, useful capacity rejected, downstream interruption rate, extra spend and latency. **No production routing influence** without calibrated incremental benefit.

## 11.1 Go/no-go conditions

A successful first milestone is **not** a universal subscription broker. It is a real OpenCode agent that maintains its supported session behavior through an authenticated, small AgentPool service while automatically moving later requests between two authorized compatible inference routes — with conclusive evidence of correct failures.

Pause or redesign if: the pinned OpenCode provider API cannot support essential tools; opaque state prevents practical continuity for the selected effective model; the proxy causes significant unsolved stream stalls; upstream credential modes lack required authorization; atomic concurrency behavior proves unnecessarily complex for the expected workloads; or the custom service fails to justify its complexity compared with an existing plugin/gateway.

---

# 12. Investor interpretation and unresolved economic questions

The most credible investor proposition is **reliable access orchestration across legitimate compute sources**, not the creation of unlimited discounted inference from arbitrarily many consumer subscriptions. Existing projects substantially undermine any narrative of fundamental first-mover advantage in credential rotation.

The following remain unknown and should be measured separately from engineering feasibility:

- **Target customer and buyer:** Is AgentPool sold to autonomous-software developers, internal multi-agent teams, infrastructure operators, or power users? Those segments have different procurement and security demands.
- **Willingness to pay:** Does a standalone agent-oriented inference router solve a sufficiently costly interruption, trust, or integration problem compared with existing free gateways?
- **Comparative reliability:** Does it reduce preventable agent interruptions and recovery time at a fixed inference spend relative to direct-provider OpenCode, LiteLLM, CLIProxyAPI or a plugin?
- **Cost economics:** What are real inference prices, approved subscription entitlements if any, server and support costs, local installation burden, and incremental benefits? No generic cost-saving percentage is supportable today.
- **Competitive defensibility:** Can verified continuity fixtures, least-privilege credential isolation, trustworthy diagnostics and operational simplicity form a sustainable advantage, or will gateways incorporate the same guarantees?
- **Provider dependence:** How much revenue or user value would disappear if all consumer-subscription proxy integrations were unavailable?

A sensible investor story is therefore a **risk-reducing validation plan with explicitly bounded claims**. The engineering bar is feasible but nontrivial; the commercial bar is unproven. Those should never be combined into a single confidence score.

---

# 13. Discussion, threats to validity, and conclusion

## 13.1 What the evidence proves

Current official OpenCode documentation exposes the desired provider boundary. Public sources demonstrate local credential selection, route health/cooldown handling, compatible API relay, and streaming. Established gateways validate the overall category. Issue reports show that the apparently small switching problem becomes difficult at precise quota scopes, stale deadlines, midstream errors, and shared-client contention. These observations support the **choice of architecture and the need for strict conformance testing**.

## 13.2 What remains unproven

No independently reproduced AgentPool deployment, pinned OpenCode multi-turn equivalence trace, authorized two-route session-continuity demonstration, security review, production fault-injection result, provider connector entitlement proof, JEV gain study, or cost benchmark was produced as part of this paper. Those are material unknowns, not footnotes. Claims of practical certainty based solely on documentation would be premature.

## 13.3 Threats to validity

- Public projects can advertise capabilities not covered by robust independent tests.
- Issue trackers disproportionately surface problems and do not reveal incident denominators.
- Mutable default branches and rapidly changing V2 packages reduce reproducibility.
- Provider conditions and product surfaces can change after 8 October 2026.
- Technical permission varies by product, account, billing identity, deployment and negotiated terms.
- A benchmark on stateless text requests may fail to predict behavior for tools, compaction, multimodal, caching, or opaque reasoning.
- Concurrency figures are meaningless without hardware, request distribution, output sizes, actual limit scopes, and failure injection.
- A working competing project does not demonstrate a defensible market or willingness to pay.

## 13.4 Final assessment

**Engineering judgment: CONDITIONAL GO.** Preserve the v5.2 provider-only AgentPool design and its automatic selection requirement. Build the smallest two-route, authorized, testable implementation rather than expanding toward an agent broker or universal protocol platform. Treat JEV as optional future research. Require exact empirical conformance and operating-mode authorization before making continuity or subscription-backed claims.

**Commercial judgment: NOT YET INVESTMENT-PROVEN.** The category is real and technically precedented, but product-specific performance, distribution, competitive advantage, financial benefits, and the core consumer-subscription premise need independent validation. Public terms currently prohibit important forms of subscription-plan intermediation/aggregation. That obstacle cannot be removed by calling the operation a manual switch or by noting that each account was paid for.

The central design proposition remains worth testing:

> **TradeAgent owns the work. OpenCode owns the agent. AgentPool owns only the safe, authorized selection and delivery of inference.**

The project will be credible to a skeptical engineering reviewer — and therefore to a serious investor — when it demonstrates that proposition under controlled failures with less complexity than the available alternatives.

---

# References and research trail

**Vendor and primary specification sources**

- **[V1]** OpenCode, *Providers* (V2). Custom provider catalogs, runtime packages, endpoint overrides, timeout and retry statements. https://opencode.ai/v2/docs/providers
- **[V2]** OpenCode, *Build Plugins* (V2). Session model hooks and retry-decision API, including its non-disclosure of internal replay details. https://opencode.ai/v2/docs/build/plugins
- **[V3]** OpenAI, *Conversation state*. Client/server conversation continuation and response-state mechanics. https://developers.openai.com/api/docs/guides/conversation-state
- **[V4]** OpenAI, *Reasoning models*. Reasoning items and stateless/continued inference semantics. https://developers.openai.com/api/docs/guides/reasoning
- **[V5]** Google, *Gemini API rate limits*. RPM/TPM/RPD and project-level rather than API-key-level limits. https://ai.google.dev/gemini-api/docs/rate-limits
- **[V6]** OpenAI, *Sign in with ChatGPT Terms*, 29 September 2026. Connected-application limits, consent, no-charge access, and account-rotation/pooling prohibitions. https://openai.com/policies/sign-in-with-chatgpt-terms/
- **[V7]** OpenAI, *Sign in with ChatGPT DevKit* and accompanying Noncommercial License v1.0. https://github.com/openai/sign-in-with-chatgpt-devkit ; https://github.com/openai/sign-in-with-chatgpt-devkit/blob/main/LICENSE
- **[V8]** Anthropic, *Claude Code legal and compliance*. Authentication modes, consumer subscriptions, unmodified native runtime, and third-party product restrictions. https://code.claude.com/docs/en/legal-and-compliance
- **[V9]** TypeSafe, *API reference*. Typed `noul`, `choice`, `score` and state inputs. https://docs.typesafe.ai/api
- **[V10]** TypeSafe, *JEV 1.13 jaggedness*. Numerical and time-comparison limitations. https://docs.typesafe.ai/model-jaggedness/jev-1.13

**Open-source primary research**

- **[O1]** `ndycode/codex-multi-auth`, repository, README, MIT license and usage caveat. https://github.com/ndycode/codex-multi-auth
- **[O2]** `codex-multi-auth`, live request-level proxy source (mutable `main` at research date). https://raw.githubusercontent.com/ndycode/codex-multi-auth/main/lib/runtime-rotation-proxy.ts
- **[O3]** `ndycode/oc-codex-multi-auth`, OpenCode plugin and V2 adapter documentation. https://github.com/ndycode/oc-codex-multi-auth
- **[O4]** `router-for-me/CLIProxyAPI`, Go inference proxy and selector. https://github.com/router-for-me/CLIProxyAPI ; https://github.com/router-for-me/CLIProxyAPI/blob/main/sdk/cliproxy/auth/selector.go
- **[O5]** BerriAI, LiteLLM AI gateway and router. https://github.com/BerriAI/litellm
- **[O6]** Portkey AI, Open Source Gateway. https://github.com/Portkey-AI/gateway
- **[O7]** OpenCode *OAuth Marathon* discussion (feature proposal, not unconditional release proof). https://github.com/anomalyco/opencode/issues/8591

**Field reports and adversarial test sources**

- **[F1]** CLIProxyAPI #5590, sleep/resume quota deadline defect; **closed / Fixed** when checked. https://github.com/router-for-me/CLIProxyAPI/issues/5590
- **[F2]** CLIProxyAPI #5619, model-bucket quota misclassified at credential scope; **closed / Fixed** when checked. https://github.com/router-for-me/CLIProxyAPI/issues/5619
- **[F3]** CLIProxyAPI #5611, stale quota cooldown recovery. https://github.com/router-for-me/CLIProxyAPI/issues/5611
- **[F4]** CLIProxyAPI #3467, per-client proxy capacity/admission complaint. https://github.com/router-for-me/CLIProxyAPI/issues/3467
- **[F5]** OpenCode #21173, custom Responses proxy stream/turn stalls. https://github.com/anomalyco/opencode/issues/21173
- **[F6]** OpenCode #23767, Anthropic provider tool-field incompatibility through proxies. https://github.com/anomalyco/opencode/issues/23767
- **[F7]** OpenAI Developer Community, encrypted reasoning verification portability report. https://community.openai.com/t/how-stateless-is-reasoning-encrypted-content/1366697
- **[F8]** LiteLLM discussion #15910, HTTP and streaming error propagation. https://github.com/BerriAI/litellm/discussions/15910
- **[F9]** OpenAI Codex #32173, possible stale account binding. https://github.com/openai/codex/issues/32173
- **[F10]** OpenAI Codex #31162, OAuth refresh/revocation. https://github.com/openai/codex/issues/31162

**Provided design baseline**

- `AgentPool_Architecture_Paper(4).md`
- `AgentPool_Architecture_Final(5).md`
- `AgentPool_The_Switch_Architecture(4).md`
- `AgentPool_The_Switch_Architecture_v2(4).md`
- `AgentPool_The_Switch_Architecture_v3_Implementation_Spec(3).md`
- `AgentPool_v5_OpenCode_Inference_Gateway_Implementation_Spec(2).md`
- `AgentPool_v5_Clarified_Provider_Contract(1).md`
- `AgentPool_v5.1_Automatic_Inference_Pool_Implementation_Spec(1).md`
- `AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec(1).md`

**Evidence caution:** These nine files establish intended scope and release criteria. They are not empirical proof of a working AgentPool binary. Any mentions within them of unprovided prior versions or prototype archives remain unverified by this study. Reproduction requires pinned artifacts, real permissioned credentials, deterministic fixtures, measured traces and independent review.

*End of independent research paper — 8 October 2026.*
