# AgentPool — The Switch
## v5.2.0 · Proactive inference capacity, safe recovery, and optional JEV operational advisor · Implementation specification

> **TradeAgent owns the work. OpenCode executes the agents. AgentPool supplies the inference.**
>
> **One stable logical provider to OpenCode; an extensible pool of independently authenticated and certified API-key, subscription-backed, enterprise and local routes behind AgentPool.**
>
> **Automatic eligibility-based route selection, proactive capacity admission, and safe multi-agent concurrency are core product requirements. Manual selection is an override, not the normal operating mechanism.**
>
> **JEV can advise the AgentPool route scorer asynchronously; deterministic code remains the sole authority for capacity arithmetic, policy, billing, credentials, transport and retry safety.**
>
> **AgentPool never owns an agent, its task, its session, its tools, its files, its memory, its handoff, or its trading permissions.**
>
> **To TradeAgent, OpenCode remains the agent runtime; to OpenCode, `agentpool` is only a configured model provider. AgentPool is never a bridge that runs OpenCode.**

| Document control | Value |
|---|---|
| Status | **Implementation specification — not a statement that the complete system has been built or certified** |
| Architecture version | **5.2.0** (preserves v5.1 provider-only automatic pooling; adds proactive capacity, end-to-end retry contract, optional JEV risk advisory), research baseline **2026-10-08** |
| Decision | AgentPool is the **mandatory inference data plane** for OpenCode installations managed by TradeAgent; it is **only an OpenCode model provider at the OpenCode interface**, not a headless-agent bridge |
| Revision scope | Preserve **all** v5.1 provider-only and automatic-pool requirements; add pre-dispatch capacity avoidance, fast local route decisions, scoped reservations and a nonblocking optional JEV-based risk advisor, plus pinned OpenCode model-turn retry testing |
| Consumer | OpenCode CLI/TUI/headless/embedded runtime, under TradeAgent or directly under one authorized local owner |
| First runtime target | **Pin one specific OpenCode V2 release**; retain a separate verified V1 compatibility profile only if needed |
| Core implementation | Small local-first **Go service**, SQLite metadata, OS credential vault, HTTP inference transport and local owner-only control IPC; **in-process** capacity tracker and optional bounded JEV advisory worker, never a separate agent service |
| Credential pool | Multiple independent registrations per provider and credential mode, with **no arbitrary product-level hard cap**; resource/operational limits still apply; certification and capacity identity remain per route |
| Extension | Certified protocol adapters and modular API-key, OAuth, subscription-backed, enterprise, and local credential connectors; replaceable optional risk-scoring provider (JEV initially), no speculative universal compatibility or assumed trained forecast |
| Supersedes | AgentPool Architecture Paper; Architecture Final; The Switch; v2; v3; v4 **on the location of the provider boundary and all resulting interfaces** |
| Existing artifact | `AgentPool_OpenCode2_Manual_Switch_Source_2026-10-08.zip` is a **narrow API-key pass-through prototype**, not complete v5 functionality |

**Normative language:** `MUST` = blocking requirement; `SHOULD` = default unless a documented test-backed exception; `MAY` = optional. `DESIGN` = not yet implemented; `DOCUMENTED` = a vendor-published interface; `PROVEN` = a passing, pinned conformance test; `UNKNOWN` is **never** treated as `PROVEN`. The requirement is to build the smallest system that preserves all MUST invariants—not to accumulate components.

---

# 0. Product definition and non-negotiable boundary

## 0.0 The implementation contract — read before writing code (NORMATIVE)

**The user-facing contract is deliberately simple:** in TradeAgent, an OpenCode worker configured with `agentpool/<alias>` MUST still behave like an ordinary OpenCode worker configured with an eligible direct provider/model; **AgentPool MUST automatically choose the appropriate certified physical account/route for each new inference request according to that alias, request capabilities, permissions, continuity and capacity.** TradeAgent creates and controls **OpenCode**, not AgentPool. OpenCode carries out its normal agent loop, session/context management and permitted tool calls. **The only replacement inside OpenCode is its model-inference provider configuration/transport.** AgentPool receives individual model requests and returns compatible model results; it never runs an OpenCode worker or takes possession of an agent's work.

"As if OpenCode were running DeepSeek directly" describes **the unchanged host-to-agent control surface and agent lifecycle**, not a promise that DeepSeek, Claude, GPT or another upstream model has identical capabilities, answers, cost, context, tool semantics or provider-owned state. The selected logical model remains explicit (`agentpool/<model-alias>`), truthful and capability-checked. **OpenCode knows it selected a provider named `agentpool`; the physical credential/account behind that provider is abstracted, not the model's identity.**

### 0.0.1 Exactly four permitted communication relationships

| Caller → recipient | Interface / what passes | Ownership and prohibition |
|---|---|---|
| **TradeAgent → OpenCode** | Normal version-pinned OpenCode host API/CLI/SDK: create, prompt, observe, cancel, resume and manage workers | **TradeAgent controls the agent; OpenCode executes it.** This is the same interface TradeAgent would use if OpenCode talked to a provider directly. |
| **OpenCode → AgentPool** | Custom model-provider HTTP contract: model request, streaming result, protocol-level errors/cancellation | **AgentPool is a provider endpoint, not an agent API.** No job IDs, task queue, worker startup, memory handoff or tool execution API is introduced here. |
| **AgentPool → upstream model endpoint** | One selected, admitted, certified model request per dispatch attempt and its response | **AgentPool controls account/credential selection and inference transport only.** No calls to OpenCode's agent-control API. |
| **Trusted TradeAgent host / human owner ↔ AgentPool control IPC** | Separately authenticated **out-of-band** grants, credential registration, policy constraints, optional manual pin/override, route/usage status and diagnostics | **Management plane only.** This does not carry agent prompts or act as an alternative host-to-agent runtime. An untrusted OpenCode agent may not invoke owner controls. |

The fourth relationship is important: **"provider only" describes AgentPool's role in agent execution, not an absence of security, owner-control or telemetry APIs.** Those APIs are ancillary to inference access and MUST remain separate from the provider HTTP surface. TradeAgent MAY explicitly consume route/usage provenance; this does not grant AgentPool control over work or require the agent loop to understand physical accounts.

### 0.0.2 The equivalence test that prevents architecture drift

```text
DIRECT PROVIDER                           AGENTPOOL PROVIDER
TradeAgent                                TradeAgent
  └── OpenCode worker                       └── SAME OpenCode worker/host API
        ├── sessions, tools, context              ├── SAME sessions, tools, context
        └── model provider → direct API           └── model provider → agentpool/<alias>
                                                      └── AgentPool → approved upstream

The changed boundary is the model-provider endpoint only.
NO extra AgentPool-controlled OpenCode process; NO replacement agent loop.
```

The two configurations are equivalent **at the TradeAgent↔OpenCode control boundary** once integrated. They are not automatically equivalent at the upstream model, authentication, cost or provider-state boundary. The normal model-stream protocol MUST remain faithful to OpenCode's selected package and enabled capabilities. TradeAgent may need its **own** Codex→OpenCode runtime migration first; this architecture does **not** assert that the current Codex-based worker is already OpenCode or that this migration is zero-code-change.

**Non-negotiable anti-patterns:** reject any proposal in which (a) TradeAgent sends jobs to AgentPool and AgentPool launches OpenCode; (b) AgentPool exposes `runAgent`, `startWorker`, `sendPrompt`, `executeTool`, `resumeSession`, or handoff operations; (c) OpenCode embeds physical API keys and independently selects provider accounts; (d) a secondary model request bypasses AgentPool; (e) AgentPool silently substitutes an incompatible logical model, provider state or billing route; or (f) AgentPool pretends to transfer a currently streaming reply onto another account.

**Automatic routing is the default:** AgentPool chooses among all **eligible registered routes** for each *new inference request*. A human may optionally pin or override a route, but normal load selection, capacity-aware preference and eligible next-request fallback MUST NOT require a button press. Neither automatic nor manual switching replays an ongoing model call, controls an agent or migrates hidden provider state. A physical-route change can preserve an OpenCode session only after its exact same-effective-model, protocol, tool and state-continuity conditions pass; otherwise it fails explicitly and **TradeAgent/OpenCode**, not AgentPool, chooses a checkpoint or new-session recovery path.

### 0.0.3 Minimal observable acceptance criteria (all mandatory)

1. A version-pinned real OpenCode session works through `agentpool/<real-verified-alias>` with its ordinary agent control operations, streamed model output, multi-turn context and permitted tool calls. TradeAgent's worker lifecycle code does **not** call an AgentPool agent-execution interface.
2. The *same* OpenCode session/tool scenario is tested with a direct certified provider and with AgentPool using the **same physical model/feature contract**; differences in underlying billing, authentication and provider-state constraints are documented, not concealed.
3. In a managed worker's effective configuration and network trace, every model-generation call (including auxiliary models/subagents) reaches AgentPool or is explicitly disabled; no undeclared native provider escapes.
4. A real **automatic** selection A→B across two separately registered, proven-equivalent routes changes only later eligible requests as A reaches capacity/cooldown or becomes inadmissible; no operator action is required. The earlier in-flight route stays frozen.
5. An owner-initiated manual pin/override affects later eligible requests only; explicitly disabling the override returns to automatic selection.
6. An OpenCode request for alias/model X is never silently redirected to unrelated model family Y; a cross-model choice is explicit in the caller's authorized alias/host policy.
7. Registering many principals (including separately certified API-key and subscription connector types) does not require changing or recompiling the core router; practical resource limits are bounded and tested.
8. AgentPool's local admin/grant/usage integration remains outside the model HTTP protocol and cannot be mistaken for an OpenCode host API; an agent prompt cannot elevate those controls.

**Reviewer decision rule:** if a proposed change requires AgentPool to know which task an agent is doing, run a worker, store chat history as agent memory, execute tools or implement a second job lifecycle, **reject it as outside v5**. The gateway may retain minimal opaque correlation and provider-bound state *references* required for authorized routing/accounting, but never owns or reconstructs TradeAgent/OpenCode agent state.

### 0.0.4 Mandatory automatic-pool product contract (NORMATIVE)

**AgentPool is a pool, not merely a manually operated switch.** The system MUST support a dynamically extensible registry of multiple independent credentials, subscriptions, accounts, projects, service principals and local inference deployments. Each registration has its own protected authentication state and zero or more certified physical routes; each capacity owner has its own verified/shared allowance scope. The design MUST impose no arbitrary fixed number of registrations, routes or principals in product logic, schema, configuration or routing algorithms. Bounded RAM, concurrency, descriptor, operating-system and provider limits still require explicit resource controls and scale tests. **"Technically unlimited" means extensible without a coded small fixed ceiling, NOT infinite concurrent requests or a guarantee of unlimited provider capacity.**

**On every new OpenCode inference request:** (1) identify the explicit logical model alias and demanded capabilities; (2) authenticate the client and apply its grant, provider/data/billing policy; (3) enumerate the pool routes that are certified to serve **that exact alias**; (4) filter by credential validity, continuity affinity, scoped quota observations/cooldowns, concurrency and budget; (5) automatically select an eligible route with a deterministic sticky/fair strategy; (6) freeze that route for the full request and stream its answer; (7) record physical provenance and update observations. When the preferred route is unavailable or cooldown-blocked, the next not-yet-dispatched request SHOULD automatically use another eligible equivalent route without asking the operator. If none qualifies, fail explicitly.

**Different and same-provider registrations are both first-class:** 30 separately registered credentials for the same compatible model may all belong to one logical alias; separately certified subscriptions for multiple providers can coexist in the same AgentPool installation. If OpenCode explicitly asks for `agentpool/<claude-alias>`, only approved routes capable of that alias are candidates. If it requests `agentpool/<gemini-alias>`, a separate eligible set is evaluated. AgentPool does **not** infer that changing the physical provider or model family is safe merely because both answer text prompts. Any cross-family selection requires an explicit, supported logical-model choice or a separately certified equivalence contract; it may require TradeAgent-managed checkpoint/handoff.

**Subscription-backed inference is a mandatory connector *category* in the architecture, not a promise that every commercial subscription can be used as an inference endpoint.** Each connector must implement a provider-appropriate authorization/inference protocol and pass independent technical and operational certification. Automatic rotation occurs only among authorized registered routes and **never silently changes billing mode or account/data authority**. "Available" or "quota remaining" is evidence-based and may be unknown; capacity signals are scoped to the provider-defined owner, not assumed per key.

**Example contract:** OpenCode sends request 1 for a particular authorized coding model; AgentPool chooses certified Route A. Route A subsequently becomes capacity-blocked. When OpenCode sends request 2 for that same model, AgentPool selects certified Route B automatically if it is compatible, within budget, and carries no nonportable provider-side state. The existing response from A was never spliced, request 1 was not blindly replayed, and TradeAgent was never asked to choose accounts. If Route B is unsuitable, AgentPool returns a precise reason instead of making an unapproved substitution.

### 0.0.5 Speed, proactive capacity and JEV dual-role contract (NORMATIVE)

**Product principle:** foreseeable capacity depletion MUST be handled as a *pre-dispatch scheduling condition* when usable evidence exists; unexpected transport or capacity failures follow explicit classified recovery. **No added remote inference, credential refresh, quota probe, JEV prediction call, global drain, or expensive full-registry scan is permitted on the ordinary inference admission path.** These steps are precomputed, refreshed from normal traffic, or scheduled off-path. The fast path authenticates, checks the immutable capability/grant/billing constraints, consults a local indexed route snapshot, atomically reserves capacity and dispatches.

**JEV is explicitly allowed inside AgentPool as an optional advisory decision primitive, not as a second agent and not a replacement for deterministic logic.** It can evaluate **sanitized operational summaries** (provider-error classifications, failures relative to peer routes, intermittent flapping, qualitative route deterioration, anomaly categories) and return typed probabilities/scores. These suggestions may become bounded, expiring *soft ranking penalties* ONLY after measured live/offline calibration demonstrates benefit over deterministic baselines; they can never grant access, fabricate quota, change a model/credential/billing class, suppress a hard cooldown, select a forbidden route, authorize retries, execute a tool, or block the hot path. **When JEV is disabled, delayed, down, uncertain, costly or uncalibrated, baseline routing is unchanged.**

**Two separate JEV consumers, one shared conceptual primitive:** TradeAgent's JEV assesses agent-context and work-quality degradation for **manager-mediated** checkpoint/handoff. AgentPool's JEV evaluates sanitized infrastructure *risk only* and produces optional routing advisory. The two deployments share no task ownership or hidden session memory. AgentPool may publish deterministic route-health/continuity telemetry to TradeAgent; it **does not** request a JEV agent reset or create a handoff. If both use TypeSafe, isolate their credentials, model versioning, privacy scopes, budgets and failure modes; never require the AgentPool predictor to route its *own* TypeSafe inference through the gateway (avoid recursive dependency).

**Evidence limits:** a published JEV `noul` probability or `choice` confidence is **not** automatically a calibrated probability that a provider account will hit a rate limit during a specified horizon. A provider capacity forecast requires labeled provider-specific outcome data and benchmarked calibration; numeric quota arithmetic, timestamps and reservations belong to code. JEV's new operational predictor is **DESIGN**, not an already trained/certified feature. Source authority: TypeSafe [API primitives](https://docs.typesafe.ai/api), [JEV 1.13 limitations](https://docs.typesafe.ai/model-jaggedness/jev-1.13), [confidence](https://docs.typesafe.ai/confidence).


## 0.1 What it is

**AgentPool is a standalone inference service that OpenCode sees as one configured model provider named `agentpool`.** Every managed model request produced by an OpenCode agent crosses AgentPool. AgentPool authenticates the local caller, checks model, credential, billing and data-access controls, **automatically selects an eligible physical upstream route from its registered pool on every new request**, forwards or translates the inference request through a certified adapter, and streams the result back in the exact downstream protocol that OpenCode expects.

TradeAgent continues running its multi-agent organization using OpenCode without choosing the underlying credential or physical route as part of executing each agent step. Its **trusted host integration still provisions a scoped AgentPool grant and may read audit/usage/model provenance out of band**; these are infrastructure functions, not a replacement for the OpenCode host API. Conversely, AgentPool never sees a TradeAgent assignment as something to plan, execute or manage. It handles **individual inference requests**, not agent jobs.

The desired outside appearance is a **stable OpenCode model provider** and, within the guarantees below, uninterrupted **agent-level work** across ordinary model request boundaries—not literally one unbroken provider connection or a magically transferable reasoning state.

## 0.2 The authoritative topology

```text
  HUMAN / OPERATING APPLICATION
       |                                  Other personal clients (optional)
       v                                                     |
  TRADEAGENT (.NET application)                               |
  - chiefs, managers, JEV, roles, budgets, permissions        |
  - research, tools, worktrees, tasks, durable evidence       |
  - owns session lifecycle, checkpoints, handoffs            |
       |                                                     |
       | creates / controls OpenCode work                     |
       v                                                     |
  OPENCODE (CLI/TUI, headless server, or certified SDK)       |
  - executes the agent loop and its tool calls                |
  - owns prompts, messages, compaction, native sessions       |
  - selects logical model: agentpool/<model-alias>            |
  - ALL managed inference requests to AgentPool               |
       |                                                     |
       +-------------------------------+---------------------+
                                       |
                           HTTPS/HTTP loopback model requests
                                       |
                                       v
                         AGENTPOOL (standalone local service)
  +-------------------------------------------------------------------+
  |  PRIVATE CONTROL PLANE           INFERENCE DATA PLANE              |
  |  - register credential           - authenticate local client      |
  |  - inspect/select permitted      - validate alias & capabilities  |
  |    physical routes               - resolve eligible route        |
  |  - budgets/usage/cooldown         - reserve/commit attempt         |
  |  - issue workload grants          - stream protocol-correct result|
  |  - pause/revoke/diagnose          - classify usage and failures    |
  +-------------------------------------------------------------------+
            |                     |                    |            |
            v                     v                    v            v
      OpenAI API            Anthropic API        Google API      Local LLM
      API/connector modes   API/connector modes  API/connector  compatible API
                                                modes

       Provider connectors are integrated and certified independently.
```

**Hard routing invariant:** for each managed OpenCode worker, every model-generation request—including ancillary model calls such as summarization, titles, compaction and native subagents—MUST either reach AgentPool or be **disabled/explicitly out of scope**. A hidden call directly to an installed provider is a boundary failure. Operating-system egress controls and tested effective provider policy enforce this; merely choosing `agentpool` as the default model is insufficient.

## 0.3 Ownership matrix

| Concern | TradeAgent | OpenCode | AgentPool | Upstream provider |
|---|---|---|---|---|
| Agent organization, objectives, evaluation, trading controls | **Owns** | No | Never | No |
| Job/attempt lifecycle, durable checkpoints, JEV escalation | **Owns** | Executes requested work | Never | No |
| CLI/TUI/headless agent execution, tool calls, agent prompts | Controls and restricts | **Owns** | Never | No |
| Session messages, model-facing context, compaction | Institutional memory/evidence is authoritative | **Owns native runtime context** | Does not persist it | May retain provider-specific state |
| Logical model alias and allowed choices | Sets work grant and budgets | Sends alias and request | Validates alias | No |
| Physical provider, credential, billing path, quota evidence | Policy constraints | Must not see upstream secret | **Owns routing** | Authoritative on allowance and charges |
| Model stream transport, strict wire correctness | Receives outcome | Generates request/consumes response | **Owns relay/adapter** | Generates tokens and events |
| Tool side effects, filesystem writes, broker orders | **Authorizes/reconciles** | Executes only granted tools | Never executes them | No |
| Automatic physical-route selection / optional owner override | Supplies workload constraints | Sends requested alias, not an upstream account | **Owns automatic pool selection; owner may pin/override** | Defines permissions |

**Explicit exclusion:** AgentPool does not launch, supervise or recover OpenCode processes, determine which agent should work next, schedule TradeAgent departments, call tools, handle broker credentials, own **TradeAgent's JEV agent-context supervisor**, rewrite system prompts, replace memory, or manage a task queue. An optional in-process JEV-powered **infrastructure risk advisor** is permitted solely within the boundaries of Sections 0.0.5 and 8.5. A diagnostic smoke-test command may invoke a **synthetic inference request**; that does not confer agent-execution responsibility.

## 0.4 Ten engineering invariants

1. **Only inference crosses AgentPool's model-provider data plane.** Separate authenticated control/telemetry operations exist for grants, routing, usage and credentials; they are **not agent-execution APIs**. No job state or tool-execution ownership migrates into the gateway.
2. **All managed inference crosses AgentPool.** No accidental native-provider fallback, including internal OpenCode model calls.
3. **One request, one upstream commitment.** No splitting a live stream across accounts or providers.
4. **No implicit model identity changes.** A stable alias may move only between **certified, authorized, semantically equivalent** physical routes. Model-family changes require a visible logical-model/session decision.
5. **No implicit billing changes.** A subscription or free local path must not quietly become a metered API path.
6. **Credential-mode isolation.** API keys, account-based integrations, OAuth/subscription connectors and local models have distinct credential, capacity, billing and failure semantics; never collapse them into one generic token type.
7. **Fail closed, surface truthful state.** Unknown compatibility, unclear financial exposure, ambiguous retry safety, revoked grants and partial streams never become silent successes.
8. **Pool-first automation.** For each new request, AgentPool automatically selects among all eligible, certified, permitted registered routes; route exhaustion/cooldown causes next-request reevaluation without human intervention. Manual pinning is an override. No fixed registration count is hard-coded.
9. **Proactive by evidence, never by invented quota.** Ordinary requests use bounded local route snapshots, verified provider observations, scoped reservations and conservative headroom; estimates never replace hard provider limits.
10. **JEV advisory and nonblocking.** No probabilistic result grants authority or enters the mandatory critical path; disable it without changing core correctness.

## 0.5 Capabilities we promise—and those we do not

| Requirement | Intended property | Status today |
|---|---|---|
| OpenCode sees `agentpool` as a custom provider | Supported by published OpenCode custom-provider architecture | **DOCUMENTED; pin/test exact installed build** |
| OpenCode works in interactive and headless modes | Both produce model calls through configured provider | **Integration test required** |
| Forward an OpenAI-compatible API-key route | Small, certifiable transport pass-through | **Prototype exists, not release certified** |
| Change eligible credential at next safe request | Possible without recreating TradeAgent's logical job | **Protocol/continuity tests required** |
| All Claude/Google/local requests through AgentPool | Possible via individually certified compatible endpoints/adapters | **Not automatically proven** |
| Universal transparent cross-model swaps | Not a valid unconditional promise | **REJECTED** |
| Mid-token cross-account or cross-model stream splicing | Not a valid promise | **REJECTED** |
| Account- or subscription-backed provider connector | Pluggable provider-native authentication, refresh, rate observation and inference transport | **DESIGN; certify each integration** |
| Automatic multi-route selection | Automatically choose a certified eligible account/credential for each new request, respect affinity, cooldown, quotas and budget | **CORE DESIGN; mandatory multi-route acceptance proof** |
| Extensible mixed-mode credential pool | Register many independent API keys, provider accounts, subscription connectors and local routes without reworking the allocator | **CORE DESIGN; each connector still certified individually** |
| Manual backend selection | Optional owner pin/override for future inference requests without transferring live streams | **DESIGN; integration proof required** |

---

# 1. Design decision: the model provider sits BELOW OpenCode

## 1.1 The correction to v4

**Old v4 placement:** TradeAgent → OpenCode headless bridge → *optional* AgentPool, with Anthropic/Google/local direct from OpenCode and sometimes a separate app-owned API harness. This allowed inference to bypass AgentPool and required coordination across two routing locations.

**New v5 placement:** TradeAgent → OpenCode agent runtime → **AgentPool mandatory logical provider** → selected upstream. AgentPool's **only product role** is inference access, including authorization, selection, accounting and transport. OpenCode is not an AgentPool dependency in the process model: the gateway speaks a certified inference protocol and can also serve other eligible clients.

This change is not a cosmetic diagram update. It relocates credential and provider policy enforcement into the single inference boundary and removes the previous `AgentPool → OpenCode worker bridge` control relationship.

## 1.2 Reasons this architecture is preferable

- **Smaller core:** no duplicate agent supervisor, session database, model agent loop, tool router, history or task scheduler in AgentPool.
- **Real separation:** TradeAgent can evolve its organization without changing the inference service; AgentPool can gain providers without changing TradeAgent.
- **One credential custody boundary:** upstream provider secrets live in AgentPool, not in OpenCode worker environments.
- **Common behavior:** OpenCode TUI/headless and authorized other clients can share documented gateway behavior, while preserving distinct local grants.
- **Easier observability:** model attempts have a single routing, rate and billing decision point.
- **Replaceability:** OpenCode can later be replaced by another inference-protocol-compatible agent client without rewriting AgentPool's authority or storage.

**Trade-off:** a universal single-provider facade makes AgentPool responsible for **the inference protocol surface**, not just switching the Authorization header. Tools, reasoning, multimodal payloads, compaction, prompt caching, streaming and error semantics differ by provider. The right answer is **capability-gated adapters**, not a universal adapter that pretends everything is OpenAI-compatible.

## 1.3 The narrowest useful MVP

**MVP-A (transport proof):** OpenCode → authenticated AgentPool → **one API-key route** → exact same model/protocol response back. Complete a real multi-turn session with tool execution and streamed output. No rotation required.

**MVP-B (automatic multi-route pool):** at least **two independently registered and certified same-effective-model routes** and a true pool selector. Automatically choose eligible routes per new OpenCode request; prove A→B without owner input when A enters a known cooldown, is disabled, or otherwise ceases to qualify; preserve route pinning for in-flight streams, context-affinity requirements, usage attribution and safe failure handling. Optional owner pin/override is tested as a control, **not the default path**. Prove native OpenCode session continuation only under live end-to-end continuity tests.

**MVP-C (mixed-provider and credential-mode breadth):** one additional certified backend/connector type, then further API-key, account-based, subscription-backed, enterprise and local connectors after each passes its authentication, tool/stream/reasoning, billing and continuity test suite. Local models are good early candidates when they implement the required compatible API reliably. Multiple registrations of any certified connector type use the **same core auto-selector**; no per-provider routing fork.

**Only advanced routing optimization is deferred.** P1 proves transparent inference transport; P2 builds the registry/selector foundation; **P4 is the automatic A→B route-switching release gate**. Human selection is not a substitute for this proof. Forecasting, sophisticated capacity optimization and speculative proactive prewarming may follow later, without altering the core automatic eligibility selector.

---

# 2. OpenCode client/provider contract

## 2.1 The two separate interfaces

```text
 TradeAgent  -- host-to-agent control --> OpenCode session/SDK/server
 OpenCode   -- agent-to-model request ---> AgentPool inference HTTP endpoint
 AgentPool  -- model-to-provider API ------> approved physical backend
```

The first interface is **OpenCode's application API**. TradeAgent decides how to create, direct, cancel or resume its OpenCode agents. The second interface is **the provider API shape**. AgentPool implements **only the second interface for model inference**; its separate owner IPC performs infrastructure authorization, selection and telemetry, **never OpenCode session control**. A model-provider URL is not an OpenCode session server URL. The host's normal session and tool-event APIs remain OpenCode's, even when the model request physically transits AgentPool. **Do not implement an AgentPool-hosted OpenCode API wrapper, an OpenCode child process in AgentPool, or a duplicate agent loop.**

This separation also means AgentPool cannot make a model call 'continue working' if OpenCode stops executing its agent loop; it can only keep successfully serving model requests from whichever OpenCode session is active.

## 2.2 Exact V2 reference config — Chat Completions profile

The following is an **illustrative native V2 configuration** to verify against the **pinned deployed build**; do not ship a fake model ID or assume unsupported features. `AGENTPOOL_CLIENT_TOKEN` is an AgentPool-issued **local audience-limited gateway credential**, **not** the upstream provider key. The entry can be kept in a private generated OpenCode config home used by TradeAgent or merged into the owner's interactive setup.

```jsonc
// opencode.jsonc — EXAMPLE, native OpenCode V2 shape
{
  "$schema": "https://opencode.ai/config.json",
  "model": "agentpool/coding-compatible",
  "providers": {
    "agentpool": {
      "name": "AgentPool",
      "package": "@opencode/ai/providers/openai-compatible",
      "env": ["AGENTPOOL_CLIENT_TOKEN"],
      "settings": {
        "baseURL": "http://127.0.0.1:43819/v1",
        "transport": "http"
      },
      "models": {
        "coding-compatible": {
          "name": "AgentPool • verified coding model",
          "modelID": "coding-compatible"
        }
      }
    }
  }
}
```

The `coding-compatible` alias is **not** an actual model entitlement. In deployment it MUST be replaced by a truthful, published/locally configured **catalog entry** whose model identifier, capabilities, context/output limits and selected package are verified against the real backend. To avoid falsely advertising one model as a different one, a sensible deployed alias includes family/revision/variant (for example `openai-gpt-<real-revision>-tools`), but no example here promises availability of a given paid model.

**Important V2 facts:** native V2 uses `providers`, `package`, `settings.baseURL`, and a `models` map. Model-level `package` overrides are documented. V1 uses `provider`, `npm`, `options.baseURL` and AI SDK packages. These must not be mixed without a verified migration path; a V2 native package must implement the corresponding native provider interface. See sources S1–S3.

## 2.3 Exact V2 reference config — Responses profile

For a verified Responses-capable downstream endpoint use the documented compatible Responses package, **not** the chat-completions package with a hopeful URL substitution:

```jsonc
// Replace the complete provider block above when exercising a Responses-only profile.
{
  "$schema": "https://opencode.ai/config.json",
  "model": "agentpool/responses-compatible",
  "providers": {
    "agentpool": {
      "name": "AgentPool",
      "package": "@opencode/ai/providers/openai-compatible/responses",
      "env": ["AGENTPOOL_CLIENT_TOKEN"],
      "settings": {"baseURL": "http://127.0.0.1:43819/v1", "transport": "http"},
      "models": {
        "responses-compatible": {
          "name": "AgentPool • verified Responses model",
          "modelID": "responses-compatible"
        }
      }
    }
  }
}
```

A combined production catalog MAY include both Chat and Responses model entries through V2's **per-model package override** once an executable conformance test establishes that behavior; the two separate blocks above are the **safer smoke-test baselines**. Do not use an AI SDK Chat Completions transport against `/responses`, or a native provider package with a synthetic endpoint that implements only part of its expected schema.

## 2.4 V1 compatibility, if genuinely required

V1 documents a custom provider `provider.agentpool.npm = "@ai-sdk/openai-compatible"` (Chat Completions) or `@ai-sdk/openai` (Responses) with `options.baseURL`. A separate versioned template may be maintained. The `opencode` executable name is **not** reliable proof of a V1 runtime: OpenCode's current V2 migration documents changes to installation and executable naming. Inspect **actual version, command, runtime package and API behavior** rather than assuming `opencode2` or `opencode` is always one major version.

## 2.5 Headless operation and personal TUI

- **TUI:** the user launches OpenCode and selects `agentpool/<alias>`. Every model request reaches local AgentPool. Upstream secrets never enter the TUI configuration.
- **Headless:** TradeAgent launches or embeds OpenCode using its own worker-lifecycle interfaces; the generated effective OpenCode configuration points at the same local AgentPool model endpoint.
- **Client authentication:** TradeAgent requests a short-lived scoped gateway grant. An owner-driven TUI can use a distinct, revocable personal grant. Neither is an upstream API key.
- **Model/small-model/compaction/subagent cases:** every enabled OpenCode operation that generates tokens must resolve to admitted AgentPool model entries. Internal defaults may differ from the primary model; detect and test them.
- **No silent bypass:** runtime policy, fully isolated config discovery, outbound network policy, and local gateway telemetry must independently confirm that nothing reaches providers directly.

## 2.6 Trust boundary of an OpenCode process

OpenCode can read prompts, worktree data and tool results. Config files alone are not sufficient to sandbox it: it may also load global/project settings, plugins, skills and environment variables. **TradeAgent** owns OS-level worker containment and tool permission, not AgentPool. Initially one contained runtime per security/data-authority equivalence class (conservatively one worker per task attempt) is safest. Grant only the worktree, tools, local AgentPool address, and narrowly scoped gateway bearer the job requires. Never inject upstream provider keys into worker environment or mount AgentPool's secret vault into the worker.

A Windows Job Object is **process-lifecycle isolation, not by itself a security sandbox**. Windows restricted tokens/AppContainer or another independently certified isolation mechanism are required where workers may execute unsafe tools. On macOS/Linux use actual process/filesystem/network isolation, not just a working directory. Neither this document nor AgentPool authorizes uncontained OpenCode tool execution against live trading systems.

---

# 3. The public inference contract

## 3.1 Stable downstream endpoint

AgentPool MUST expose only the certified local inference surfaces actually needed by the installed OpenCode package. The reference prefixes are:

```text
http://127.0.0.1:43819/v1/chat/completions    POST (certified Chat profile)
http://127.0.0.1:43819/v1/responses           POST (certified Responses profile)
```

Optional `/v1/models` is **advisory discovery**, never authority to access new models or proof of entitlement. No generic `/*` proxy endpoint. Reject other methods, unexpected upgrade/WebSocket requests, unknown paths, redirects, unsupported content types and unapproved API variants. Bind to `127.0.0.1` (and optionally separately to `::1` with strict validation) only. Ports are deployment parameters; `43819` is example-only.

For each admitted HTTP request AgentPool MUST:

1. Authenticate the gateway client and workload grant **before** reading a large body or contacting upstream.
2. Bound request size, time, complexity, concurrent admissions and stream buffering.
3. Parse and validate the requested `model` alias, protocol and declared capabilities.
4. Resolve the authorized exact physical route and reserve its concurrency/cost exposure.
5. Freeze route identity, credential epoch and request content for **this one attempt**.
6. Rewrite the alias into the exact upstream model ID **only within a route manifest**; never obey client-supplied `base_url`, upstream `Authorization`, provider IDs or billing endpoints.
7. Dispatch over pinned TLS or trusted local transport, with a strict host/path/redirect policy.
8. Return downstream headers, status and stream in the **correct client-facing protocol**, honoring cancellation/backpressure.
9. Finalize with truthful usage, incomplete/ambiguous accounting and reason-coded outcome.

A client may send opaque prompt content; the gateway does not need to store it. A protocol adapter may inspect/transform the payload **ephemerally**, with no persistence by default. All transformations must be certified and lossless for the admitted capability set or fail closed.

## 3.2 Logical model aliases are contracts—not vague labels

```text
LogicalModelAlias {
  id                     // e.g. coding-compatible (inside agentpool/...)
  downstream_protocol    // chat_completions | responses
  nominal_model_identity // provider, family, version/snapshot, variant
  known_context_limit
  known_output_limit
  capabilities           // tools, vision, images, reasoning, parallel tools, etc.
  admissible_routes[]    // independently certified exact upstream profiles
  continuation_class     // stateless-same-model | provider-state-bound | none
  permitted_billing_modes[]
  cost_policy            // hard cap and reservations
  policy_revision
}
```

An alias can have several credentials only when they lead to **the same effective model contract** or a specifically certified equivalence group. Same display name, similar score or access to 'GPT' is not proof of equivalence. Two providers exposing similarly named models may differ in model snapshot, system prompts, reasoning behavior, tools, caching and safety terms. In strict or reproducibility-sensitive TradeAgent roles: **pin exact physical model and variant per attempt**; do not change them invisibly.

For different model families, expose **different aliases**: `agentpool/<real-openai-model>`, `agentpool/<real-claude-model>`, `agentpool/<real-gemini-model>`, `agentpool/<real-local-model>`. TradeAgent or a user makes an explicit visible choice to change the logical model and, if appropriate, checkpoints the task. Provider change without model change and model change without task change are separate operations.

## 3.3 Protocol compatibility levels

| Level | Gateway operation | Allowed when |
|---|---|---|
| **L0: byte/protocol pass-through** | Replace auth + upstream model ID, preserve protocol/events with bounded framing | Exact downstream/upstream protocol and every used capability are compatible |
| **L1: narrow event normalization** | Rewrite documented model/usage/error metadata; preserve tool/reasoning semantics | Fixture and live conformance prove correct streaming/termination |
| **L2: provider-native translation** | Convert entire message/tool/reasoning stream between independently defined APIs | Full capability matrix, reversibility where needed, failure behavior proven |
| **L3: unsupported** | Reject and show compatibility reason | Feature, state, auth, terms or actual transport is not proven |

**Ship L0 first.** Do not call an API OpenAI-compatible merely because it accepts a basic text `POST`. Chat Completions and Responses are different contracts. Native Anthropic Messages, Gemini, Azure variations and local inference may need L1/L2 adapters. A `stream=true` text demo does not establish tool calling, parallel tool calls, images, usage, compaction or model-state continuity.

## 3.4 Capability negotiation and fail-closed checks

Each provider adapter declares a versioned, tested capability matrix for `text_input`, `text_output`, `streaming`, `system/developer`, `tools`, `parallel_tools`, `structured_output`, `vision`, `audio`, `reasoning_summary`, `opaque_reasoning`, `prompt_caching`, `provider_state`, `resume`, `cancel`, `usage`, `max_context`, `max_output`, `error_shape`, and the exact event schema revision.

The **intersection** of OpenCode model requirements, alias capabilities and route capabilities determines admission. Unknown property means unsupported for that call; the gateway cannot silently drop tools, degrade an image to text, remove reasoning signatures, ignore a JSON schema, swap roles, or truncate context. An unsupported-but-optional feature can be disabled in a separately selected **explicit downgraded alias** after the host acknowledges the change.

## 3.5 Translation boundaries

- **OpenAI API:** the exact endpoint, model and API project are certified; the request's selected protocol must match the adapter.
- **Anthropic API:** native Messages streaming/tool/extended-thinking semantics must be translated or served through a properly certified compatible endpoint. Distinct API-key and account-based integrations use separate authentication adapters and protocol certificates.
- **Google API/Vertex:** distinct auth, model, tool and error semantics; separate adapters or certified compatibility surfaces. Cloud project, region and billing class are explicit.
- **Local models:** only the configured loopback process/model/deployment is trusted. Compatible endpoint behavior is **tested**, not inferred from Ollama/vLLM/LM Studio branding. Local inference can fail through GPU exhaustion or context mismatch even with no monetary cost.
- **Other providers:** an adapter is just an implementation of an interface and proof suite; the core router remains unchanged.

Keep provider-specific behavior in `internal/adapters/<provider>` or a separately versioned adapter package; no `if provider == ...` spread across the allocator. A mature external library MAY implement an adapter **after** license, supply-chain, privacy, semantic and version-lock reviews; it may not quietly introduce a second policy router or agent executor.

---

# 4. Credential custody, integration modes and route identity

## 4.1 The identity model

Never collapse these different identities:

```text
LocalOperator          person allowed to register/manage routes
ClientApplication      TradeAgent, personal TUI, or another admitted client
WorkloadGrant          short-lived permission for a specific caller and purpose
ProviderPrincipal      authenticated account / org / workspace / project
CredentialRegistration protected secret reference + credential epoch
CredentialMode         api_key / oauth / subscription_connector / enterprise / local
BillingRoute           API project, subscription connector, or local model
CapacityOwner          provider-defined shared allowance/rate-limit bucket
PhysicalRoute          provider + principal + billing + credential + model + endpoint
ProtocolProfile        downstream + upstream protocol, adapter, capability revision
```

Credential registrations, capacity owners and physical routes are separate objects. Multiple credentials may share a provider-defined allowance, and one account may expose several models or billing routes. This is a **capacity-accounting and identity-modeling requirement**, not a reason to erase multi-account or subscription-backed connectors from the product design.

## 4.2 Authentication modes are modular first-class connectors

| Mode | Product design | Integration acceptance |
|---|---|---|
| Provider API key | Native support | Verify API endpoint, scoped key, billing and model access |
| Local model/service | Native support | Verify local endpoint, protocol, model and resource limits |
| Enterprise/service account | Provider adapter | Verify identity, token lifecycle, tenant/project scoping and billing |
| OAuth/account-based integration | Provider adapter | Verify interactive authorization, delegated scopes, refresh/revocation and inference transport |
| Subscription-backed integration | Provider adapter | Verify applicable connection flow, subscription capacity signals, request transport, streaming and account isolation |
| Multiple registered credentials/accounts | Extensible native route registry; no coded small fixed cap | Verify independent custody, distinct principals/capacity owners, and automatic per-request selection across all eligible routes |

**Design stance:** subscription-based, account-based, API-key and local inference are all legitimate *architectural connector categories*. They are not all technically identical, and no single generic connector is presumed to work for every provider. Provider-specific deployment authorization is assessed separately from the product's routing architecture; it does not remove connector slots or hard-code a provider-specific limitation into the core design.

The compatibility registry records `UNCONFIGURED | PROBING | CERTIFIED | DISABLED` for a specific `provider + credential_mode + endpoint + protocol_profile + app_identity`. Certification is a technical statement established by live tests, never a guess based on the presence of a credential. The owner/workload grant independently determines which certified routes a caller may use.

## 4.3 Subscription and account-based connector boundary

Account-based providers can expose browser authorization, OAuth, CLI-associated login, or other documented credential flows. An AgentPool connector owns its **own isolated credential registration and refresh lifecycle**; it must not assume an API key can be substituted for an OAuth token or that one provider's endpoint speaks another provider's protocol.

For each connector, implement only the applicable building blocks:

- Provider-appropriate interactive login and credential vault integration, where supported by its integration.
- Authenticated inference transport; token refresh, expiration, logout and revocation.
- Model catalog discovery and capability/cost/capacity observations.
- Multiple locally registered principals with separate route identity, quotas and audit history.
- Automatic next-request route selection across every certified eligible registered principal, with optional human pin/override; active streams remain pinned to their original route.
- Explicit account/session affinity when a provider returns opaque response IDs, reasoning state or other nonportable context.

Keep authentication modules replaceable. **Absence of a finished subscription connector is an implementation backlog item, not an architectural prohibition.** Do not claim that a hypothetical connector already works until its end-to-end proof passes.

## 4.4 Credential lifecycle

- Register credentials only through the local owner interface. Do not accept raw provider secrets from OpenCode model requests or TradeAgent prompts.
- Store secrets in macOS Keychain / Windows Credential Manager or equivalent OS-backed secret store. SQLite contains an **opaque reference**, never the token.
- Keep provider access/refresh tokens separate from AgentPool's short-lived local client bearers; encrypt/protect at rest and keep logging redacted.
- Refresh only through a **documented provider flow**, single-flight per registration; bound refresh versions with `credential_epoch`, never overwrite a newer token with a stale response.
- Revoke/disconnect: stop new admissions for affected registrations, invalidate grants/routes as appropriate and safely settle/abort in-flight attempts by policy. The automatic selector MAY use another **already registered, certified, explicitly authorized** compatible route for a subsequent new request, but MUST NOT recreate a revoked credential, authenticate an unregistered account or expand the workload grant.
- Credentials are obtained through a connector-defined authentication flow, stored in an OS vault, never exposed to model prompts, and never taken from another application's private authentication storage without a designated integration.
- Owner authorization, workload grants, integration configuration and provider-specific deployment requirements are distinct concerns; the gateway must not infer one from another.

## 4.5 Route eligibility predicate

A physical route is eligible if and only if **all** apply:

```text
route.enabled
AND connector.certification == CERTIFIED AND connector.mode_covers_this_request
AND principal matches registration and allowed client application
AND route billing class is inside grant and alias policy
AND data classification, provider/region/project restrictions pass
AND exact upstream model + snapshot + variant is certified for alias
AND protocol/capability compatibility is PROVEN for this request
AND live credential/refresh epoch is valid and allowed
AND per-capacity-owner cooldown/concurrency policy allows admission
AND the upstream reports no active capacity block for this route
AND reservation fits the host financial/resource hard cap
AND workload grant is valid, bound to caller and unexpired
AND route transport and destination are allowlisted
AND context/provider-state continuity requirement is satisfied
```

If credential validity, caller authority, billing exposure, identity or compatibility is UNKNOWN, deny. Unknown remaining capacity MAY allow one budgeted first use when the connector is certified and every other admission condition holds. A provider `429` updates capacity observations and yields a typed outcome; it does not silently change model identity or billing mode.

---

# 5. Routing behavior: automatic pool first, safe boundaries always

## 5.1 Three explicit routing modes

| Mode | Behavior | Admission |
|---|---|---|
| `auto_eligible` | **Default production behavior.** For every new request, choose among all certified, permitted, compatible routes in the registered pool using continuity affinity, capacity/cooldown, concurrency, budget and deterministic fairness. Re-evaluate automatically when the preferred route is no longer eligible. | **CORE MVP-B / production default** |
| `manual_pin` | Owner explicitly pins one compatible certified physical route for future requests; if it becomes ineligible, **fail visibly** rather than disregarding the pin. Clearing the pin returns to `auto_eligible`. Active streams stay on their original route. | **Operator override and diagnostic mode** |
| `fixed` | A workload or administrator deliberately constrains admission to one physical route, with no fallback. | **Explicit restricted profile, never the ordinary default** |

Automatic routing is **not** the same as automatic *replay*. If a selected upstream returns an early `429`, a partial stream or a timeout after transmission, AgentPool records the truthful failed/ambiguous outcome and applies scoped cooldown; it MUST NOT simply resend that **same dispatched request** on another route. When OpenCode later issues a distinct authorized inference request, AgentPool re-evaluates the eligible set and may automatically choose another route. OpenCode's own retry behavior must be understood and tested separately.

The pool MAY contain many routes to the same model across different accounts/API keys, and many *different models and connector types* simultaneously. **Selection is scoped to the alias requested by OpenCode.** The router cannot silently substitute an unrelated model to avoid capacity limits, or substitute a metered billing path for a non-metered path without explicit grant. Billing allowances and permissions are verified before selection, not after.

## 5.2 Selection algorithm (normative semantics, local fast path)

1. Authenticate the caller and validate the specific requested logical alias, protocol, capabilities, data classification and grant envelope. OpenCode chooses the logical model; **AgentPool chooses its physical account/route**.
2. Enumerate all registered physical routes admitted for **that alias**. Do not restrict candidates to one hardcoded key, account, connector mode or provider.
3. Reject routes lacking a certified exact transport/model/capability profile, valid connector/credential, permitted billing class, appropriate data authority, budget allowance or transport authorization.
4. Exclude routes on observed capacity-owner cooldown, with exhausted concurrency, or known unavailability. Where authoritative request/token/quota signals exist, estimate request-specific demand and reserved in-flight demand, require configurable evidence-based headroom, and proactively skip capacity-insufficient routes *before* dispatch. Unknown capacity is handled by the certified connector's conservative first-use policy, **not** an invented remaining balance.
5. Enforce provider-bound state affinity. For stateless portable requests, prefer a healthy sticky route for cache locality if eligible, but don't let ordinary sticky preference prevent a necessary compatible automatic switch. For nonportable state, deny instead of silently moving accounts.
6. Apply an explicit `manual_pin`/`fixed` constraint when selected; otherwise use a **deterministic fair automatic selector** over the eligible set (e.g. per-capacity-owner least-reserved/round-robin with stable tie-breakers). The optional, calibrated, fresh JEV advisory may *only* alter soft ordering inside this already-eligible set, within audited policy bounds. Honor priority/weight and no-starvation requirements for multiple clients.
7. Atomically reserve capacity/cost, freeze route/model/account/credential/billing identity plus `route_epoch` for one inference attempt, and dispatch only to that immutable route.
8. Attribute results and capacity observations to the actual principal and shared capacity owner. Future requests may choose another eligible route without notifying or blocking TradeAgent's agent runtime. Return `NO_ELIGIBLE_ROUTE` or a more precise typed reason when the safe eligible set is empty.

**Pseudo-code (DESIGN; interfaces and exact state types are finalized in code):**

```go
func Select(req InferenceRequest, grant Grant) Decision {
    if !ValidGrant(grant, req.Client) { return Deny("GRANT_INVALID") }
    alias := Catalog.Lookup(req.Model)
    if !alias.IsCertifiedFor(req.Protocol, req.Capabilities) { return Deny("MODEL_UNCERTIFIED") }

    routes := Routes.ForAlias(alias.ID).Filter(func(r Route) bool {
        return Authorized(r, grant, alias, req) &&
               ProtocolProven(r, alias, req) &&
               BillingAdmitted(r, grant) &&
               ContextPortableOrSticky(r, req) &&
               LocalPreDispatchCapacityFits(r, req) && WithinBudget(r, req, grant)
    })
    if routes.Empty() { return Deny(ExplainNoRoute(req, grant, alias)) }

    // AdvisoryScore may only reorder authorized candidates; nil means baseline.
    chosen := ChoosePinnedOrAutomatic(routes, req.Affinity, grant.Mode, FreshAdvisoryOrNil(req))
    return ReserveOneAndFreeze(chosen, req.RequestID, grant)
}
```

**Pool-size invariant:** the registry and selection implementation MUST be cardinality-independent in semantics (one route, two, 100 or more). Admission may use indexed lookup and bounded work, not an unbounded scan of every global account for every request. There must be no compile-time `MAX_ACCOUNTS=2`-style product cap; deployable resource budgets and performance requirements are explicit. Credentials that share a real upstream capacity owner do **not** constitute independent quota pools.

This allocator selects **requests** only. It cannot read TradeAgent's role queue or choose which agent should execute a job.

## 5.3 Request boundaries versus stream boundaries

```text
Time ──────────────────────────────────────────────────────────>
OpenCode session S:  prompt ── request #1 ── tools ── request #2 ── tools ...
                              |                     |
AgentPool route:           API route A           API route B (automatic next-request choice)
                           [stream #1]           [stream #2]
                              |                     |
                          end/abort             end/abort

No operation moves the middle of [stream #1] into route B.
```

A safe selection change updates the **route used by later admissions**. It cannot change already dispatched calls. This normally does **not** require pausing unrelated in-flight streams. Where connector session state or server invariants prevent overlap, a **scoped drain barrier** must finish/cancel affected streams first. The prototype's global admission lock/drain is a sound simple v1 option for one trusted profile, but a production multi-workload service SHOULD drain only the affected route scope and preserve unrelated work.

## 5.4 Same-session continuation has preconditions

Keeping the same OpenCode process and session ID is not enough. Switching a physical route without restarting OpenCode is transparent only when: the same effective model is still served; context is fully reconstructible from OpenCode-owned messages; no provider-specific continuation ID/cached hidden reasoning ties the next request to route A; the exact tool/reasoning format is supported; and the next request actually reaches route B. Each condition requires conformance evidence. When any fails, AgentPool returns `PROVIDER_STATE_BOUND` or `MODEL_SWITCH_REQUIRED` and **TradeAgent/OpenCode** must select a new model or resume from an approved checkpoint. AgentPool does not perform the handoff.

## 5.5 Honest errors and status

Use stable typed reasons such as:

```text
UNAUTHORIZED_CLIENT       GRANT_EXPIRED       GRANT_REVOKED
POLICY_BLOCKED            BILLING_NOT_ALLOWED  HUMAN_APPROVAL_REQUIRED
MODEL_UNKNOWN             MODEL_INCOMPATIBLE    MODEL_SWITCH_REQUIRED
PROTOCOL_UNSUPPORTED      CAPABILITY_MISMATCH   PROVIDER_STATE_BOUND
CREDENTIAL_EXPIRED        CREDENTIAL_REVOKED    CAPACITY_COOLDOWN
CAPACITY_UNKNOWN          BUDGET_EXCEEDED       NO_ELIGIBLE_ROUTE
TOO_MANY_REQUESTS         LOCAL_OVERLOADED      DEADLINE_EXCEEDED
UPSTREAM_4XX              UPSTREAM_5XX         UPSTREAM_RATE_LIMIT
UPSTREAM_STREAM_TRUNCATED PARTIAL_OUTPUT       OUTCOME_AMBIGUOUS
CLIENT_CANCELLED          SERVICE_SHUTDOWN     INTERNAL_ERROR
```

Keep sensitive route identities hidden from unprivileged callers. Log redacted physical route IDs for owner diagnostics. Don't encode secret provider keys, complete prompts, reasoning or tool payloads in errors.

## 5.6 Predictable capacity is ordinary scheduling, not recovery (NORMATIVE)

**Fast path:** authenticate → verify grant/alias/capabilities/billing/context → indexed candidate snapshot → remove exact hard blocks → compare estimated request demand plus local outstanding reservations against provider-evidenced remaining allowance and headroom → choose sticky/fair route → atomic capacity-owner reservation → freeze attempt → send. Route computation must not call external provider endpoints or JEV. Cache snapshot reads must obey version/epoch fencing; reservation is still transactional/atomic. Response transport must remain streamed with bounded memory.

**Telemetry update path:** process response headers, usage and errors *asynchronously from the request scheduling perspective*, but persist attempt/eligibility events before returning a final disposition where durability is required. Maintain source-specific observations (`limit_scope`, `capacity_owner`, `reset`, `remaining`, `last_seen`, `expires`, `confidence`, rate dimension). `Retry-After` is a lower bound when valid, not a prediction of exact availability. Different quotas (RPM, input/output TPM, RPD, monetary hard cap, GPU slots, burst windows) are tracked separately and **all** hard dimensions must pass. Provider projects/workspaces may share limits even with multiple keys; when identity is unverified do not treat them as independent.

**Estimation split:** exact arithmetic and time-window tracking use deterministic code. Predicted request-token demand may use simple upper bounds derived from request parameters and calibrated per-model historical buckets; actual tokenization/hidden reasoning may differ. Reserve conservatively, reconcile usage without double-releasing, and explicitly degrade into bounded admission when provider data is absent. Do not pretend any route can guarantee successful service because a dashboard percentage is below 100%.

**Route scheduling states** are derived views, not substitute authorization: `READY` (may be considered), `SATURATED` (insufficient local slot/known headroom for *this request*), `COOLDOWN` (provider-evidenced scoped delay), `DISABLED` (administrative/auth/certification hard stop). Capacity evidence quality is separately `KNOWN | ESTIMATED | UNKNOWN | STALE`, per dimension; there is no fabricated universal total balance. A route can be ready for a small request but saturated for a large one. No state transition resurrects revoked credentials or clears provider `Retry-After` early.

**Parallel requests:** a top-level OpenCode worker gets its own scoped grant; a worker does not reserve an entire account for its lifetime. Separate requests may use the same account concurrently subject to provider capacity-owner limits. A switch affecting route A cannot pause unrelated active streams on route B. Fairness applies across workloads, not merely account round-robin; bound queue length, queue residence time, concurrency and client retry amplification. Keep provider connections warm where safe; refresh credentials off the hot path with strict epoch fencing. Protect large pools with alias-based indexing, not O(total_registered_accounts) work on each admission.

**Speed SLO proposal (target, not a verified benchmark):** measure local gateway-added admission latency excluding model time and socket setup at p50/p95/p99, with 1/2/100/1,000 registered routes and 1/20/100 concurrent workers. The release team MUST set host-hardware-specific budgets after capturing baseline traces; do not ship a claim of single-digit milliseconds without measurement. Enforce **zero extra upstream quota/model-probe network round trips per ordinary inference request**, and **zero JEV wait on the request path**. Load-test 429 storms, reset-boundary contention and 100-to-1 capacity sharing.

**Source-backed provider caveats:** OpenAI documents request/token windows, `x-ratelimit-*` observations, `Retry-After` and spend-cap `429`; Anthropic documents organization/workspace and request/input/output rate-limit dimensions and exposes rate-limit headers; Gemini API evaluates RPM/TPM/RPD per project rather than per key. Their scopes vary by connector/mode/version; source links and exact observed behavior are certified per adapter in Section 18.


---

# 6. Inference transport, streaming, retry and recovery

## 6.1 Dispatch state machine (normative)

```text
 RECEIVED
    | validate client, model, grant, protocol, size, budget
    v
 ADMITTED (one route reserved, no upstream transmission yet)
    | snapshot credential/model/route and attempt ID
    v
 CONNECTING
    | send/attempt to send upstream request
    v
 SENT_UNCOMMITTED ------------------------------+
    |                                            |
    | observe first valid upstream response     | ambiguity/failure
    v                                            v
 RESPONSE_READY                              UNCERTAIN
    | validate downstream status/headers         (possibly billed)
    v                                            |
 COMMITTED (first downstream byte/event)         |
    |                                            |
    +--> STREAMING --> COMPLETED_VALID           |
                |                                |
                +----> PARTIAL/FAILED -----------+
                +----> CANCELLED -----------------+

 Invalid request/policy: REJECTED (no upstream transmission)
 Rejection/early 429: FAILED_PRE_COMMIT, never claim that no charge occurred
```

**Commit points are distinct:** (a) the **upstream dispatch point** at which an inference *may have been billed*, and (b) the **downstream commitment point** after which OpenCode *may have observed output*. The second is not equivalent to the first. Record both; absence of a received response is not proof that the provider did no work. The service MUST not mark `SENT_UNCOMMITTED` as definitively unbilled after a timeout, crash or lost connection.

State transitions must be monotonic and persistent enough that restart never turns an ambiguous dispatched attempt into a clean retry. An attempt's effective credential/billing/model/route identity cannot change after `CONNECTING`.

## 6.2 Failure/retry matrix

| Scenario | Gateway action | Transparent retry? |
|---|---|---|
| Invalid client/model/policy/budget | Reject locally before dispatch | **No** |
| TCP connection fails before **provable** transmission | Return typed failure; optional retry only if exact pre-send proof and policy permits | **Off in v1** |
| Upstream responds explicit 429 before downstream stream | Preserve 429/retry evidence; cooldown scoped owner; without attempting a mid-request route change | **No in v1** |
| Upstream 401/403/entitlement failure | Disable/recheck affected registration | **No** |
| Upstream connection drops after request body sent but before first event | `OUTCOME_AMBIGUOUS`; charge unknown | **No automatic replay** |
| Broken SSE after visible text/tool delta | Terminate with detectable failure; `PARTIAL_OUTPUT` | **Never stitch** |
| Client cancels during generation | Close upstream promptly and record uncertain usage | **No** |
| AgentTool execution succeeded but subsequent model request fails | Return failure to OpenCode/TradeAgent; preserve tool effect | **Never replay tools in gateway** |
| AgentPool restarts with dispatched records | Reconcile/mark uncertain; do not resend | **No** |
| Route becomes unavailable between *completed* requests | Select only a contractually permitted, compatible new route for a **new request** | **Only mode-gated** |
| Cross-model fallback necessary | Emit `MODEL_SWITCH_REQUIRED`; TradeAgent decides checkpoint/model change | **Not transparent** |

The gateway MUST NOT have a generic "retry 429 on next account" branch, nor retry on `5xx` by default if the provider may have processed the request. A successful downstream stream with a missing terminal frame is **not** a successful completion. A client may independently retry an HTTP/model call: gateway idempotency keys deduplicate **its admission bookkeeping** but do not force an upstream provider to process a request exactly once. Client retry behavior must therefore be measured and constrained end-to-end.

**Known OpenCode consideration:** current V2 provider docs describe built-in retries for certain timeouts. The release team MUST capture actual retry behavior of its pinned OpenCode build with a fake provider and budget accordingly. A nominal gateway `no retry` policy does not prevent OpenCode from sending another request. Where native replay is uncontrollable or duplicates can cause unacceptable cost/effect ambiguity, that runtime/profile is **not admitted for unattended execution**.

## 6.3 Streaming contract — Chat Completions

For `POST /v1/chat/completions`:

- Preserve HTTP status semantics, `Content-Type`, valid SSE `data:` framing, UTF-8, event ordering, chunk IDs, role deltas, tool-call IDs, `index`, function names and arguments, `finish_reason`, optional usage and `[DONE]` terminal marker where required by the downstream contract.
- Support arbitrary TCP chunk boundaries: a single `data:` event may span reads or contain multiple framed lines; never equate one network read with one SSE event.
- Bound individual event size, cumulative header/body size, request time, idle time, decompression, and concurrent socket count; propagate backpressure, not unbounded buffering.
- Do not decode/re-encode an L0 byte stream if that alters content. If inspecting it for terminal status, use an independent bounded parser while forwarding semantics unchanged.
- If translating an upstream protocol, produce valid, ordered **downstream** SSE with tests for interleaved tool deltas and completion/error sequencing.
- Never invent output tokens or append a fabricated successful `[DONE]` to hide upstream truncation.

## 6.4 Streaming contract — Responses

For `POST /v1/responses`:

- Preserve the selected Responses schema and SSE event lifecycle, including response creation, content and tool/item deltas, any declared reasoning metadata, terminal completion/incomplete/failure semantics, and correct usage fields.
- Do not mix Chat Completions `choices[].delta` payloads with Responses `response.output_*` events.
- Do not flatten function-call items into text, discard reasoning signatures, replace unknown item types with fake text, or rewrite `previous_response_id`/stateful continuation references into unrelated provider state.
- If upstream sends a provider-owned response ID, record its provider/credential binding **only for routing correctness** if absolutely necessary; the ID cannot be used with another physical route absent proof. Do not make AgentPool an archive of transcript state.
- Streaming completion MUST be verified by the adapter against the exact event schema. On incomplete upstream output, terminate and return truthful failure or abort transport, subject to the downstream protocol's ability to signal errors.

## 6.5 Headers, HTTP errors and cancellation

- Inbound OpenCode authentication is an **AgentPool local bearer** or an equivalently scoped transport credential; upstream Authorization headers are injected **only** from AgentPool's secret store.
- Do not forward inbound cookies, arbitrary upstream header overrides, referer/proxy headers or caller-controlled endpoint names.
- Use strict allowlists for request headers and outbound response headers; strip `Set-Cookie`, hop-by-hop headers, untrusted CORS, sensitive upstream identifiers and unsafe redirects.
- Map 401/403, 408, 429, 5xx and malformed-body errors to a precise, documented downstream schema. If the selected client cannot parse the chosen error envelope, **the route is incompatible** until this is fixed.
- Treat `Retry-After` as advisory, bounded evidence tied to its real capacity owner/window; do not fabricate a reset date or global token counter.
- Propagate cancellation from OpenCode to the upstream request and close streams promptly; cancellation **does not prove** upstream computation or charges stopped.
- Do not issue a second upstream request to infer why the first failed unless an owner-approved budgeted diagnostic explicitly allows it.

## 6.6 Provider errors after downstream commitment

HTTP status has already been sent once the client starts receiving a `200` stream. AgentPool cannot change that status to `502` mid-stream. When an adapter's selected protocol supports a valid terminal error event, emit that exact event and close. Otherwise abort the connection, record `PARTIAL_OUTPUT`, and prove that the pinned OpenCode client classifies the aborted stream as **failure**, not as a successful partial answer. It is forbidden to silently normalize a truncated response into completion.

## 6.7 Time, sleep/wake, deadlines and idempotency

- Maintain monotonic in-process deadlines for request wall time; persist absolute cooldown observations with source, provenance and expiry, validating against clock skew/suspend on restart.
- A restart invalidates all in-memory reservations and replays durable `DISPATCHED` records into `UNCERTAIN` unless provider-side confirmation proves otherwise.
- A stale worker generation/route selection epoch cannot modify a current route or settle another attempt.
- Duplicate client idempotency keys with identical body hash return the existing disposition **without re-dispatch** while retained; a key reused with a different hash is rejected. This is bookkeeping idempotency, not upstream exactly-once.
- On `SIGTERM`/Windows service stop: close admission, give in-flight streams a configured bounded grace period, cancel remaining streams, persist uncertainty and exit; never create a fake successful terminal event during shutdown.

## 6.8 Safe OpenCode model-turn recovery contract (DESIGN; opt-in after conformance)

**Objective:** a known or confidently predicted depleted route is excluded *before dispatch*; recovery handles the unexpected case only. **Default internal AgentPool replay remains zero.** A definitively rejected, pre-output provider `429` is recorded and its verified capacity scope cooled down. If the pinned OpenCode runtime exposes a verified retry policy hook and the adapter proves an eligible response was rejected **before downstream commitment**, an authorized model-turn retry MAY cause OpenCode to send a **new** request that AgentPool routes to another allowed equivalent account. This is a *new OpenCode attempt*, never gateway-side splicing or silent replay.

**Admission guard for eligible retry:** (1) upstream classification is explicit and from a certified adapter, (2) no model/tool deltas were delivered, (3) the continuation class permits route movement or retry stays pinned, (4) workload's cumulative retry/time/budget ceilings permit it, (5) duplicate tool effects are impossible or externally reconciled, (6) OpenCode's actual retry hook/error mapping is conformance-tested against an exact pinned runtime, and (7) retry honors provider reset information and avoids immediate repeated use of cooled-down scopes. Provider charging may remain uncertain despite early rejection—ledger records that uncertainty; never promise zero bill.

**Nonretryable by default:** ambiguous sent-but-unanswered timeouts, partial SSE, provider-state-bound continuation, credential revocation, account entitlement/billing failure, unsupported model/adapter, exhausted financial budget, or any error where exact client replay behavior has not been tested. `Retry-After` is not ignored or converted into a short arbitrary sleep. One bounded end-to-end retry budget (client/SDK/gateway) replaces nested retry loops; the pinned OpenCode default max attempts is a hard limit. If the client cannot guarantee a safe retry, return an explicit failure to OpenCode/TradeAgent. TradeAgent owns checkpoint/handoff/reconciliation.

**OpenCode evidence:** [V2 plugin documentation](https://opencode.ai/v2/docs/build/plugins) describes `ctx.session.hook('retry', event => ...)` with `event.error.status`, attempt, and mutable `event.decision`. The document **does not prove** that a `429` hook alone detects whether any output was already committed, or prove how replay handles provider response IDs/tool calls. Do not ship optimistic retry code from the example alone; trace native behavior and verify fixture integrity first. OpenAI's official rate-limit guide warns of SDK retries and retry amplification.


---

# 7. Context, reasoning, caching and physical model continuity

## 7.1 Three memory domains

| Domain | Authoritative owner | Gateway obligations |
|---|---|---|
| Native OpenCode messages, prompts, context and compaction | **OpenCode** | Do not store/author/compact; preserve the exact admitted inference contract |
| TradeAgent task state, manager instructions, JEV handoffs, institutional evidence and policy | **TradeAgent** | Do not manage or mutate; return accurate route/usage/provenance metadata |
| Provider-native response IDs, context caches, hidden reasoning and session stores | **Specific provider route** | Never assume portability; bind state to correct provider/principal/model if passed through |

A new API key serving the same provider/model does not guarantee the same **provider-side state**. Shared prompt prefix caches, conversation IDs, encrypted reasoning blobs, persistent Responses resources, provider-native compaction, server-side tools and auth-scoped artifacts may be attached to the original account/project or otherwise nonportable.

## 7.2 Safe continuation classes

**C0 — Self-contained request.** OpenCode submits everything needed for the inference; no opaque provider-side references. A switch to an independently certified equivalent route MAY be safe at the next request if policy permits and stream/tool behavior is compatible.

**C1 — Provider-bound state.** Input includes a stateful response ID, encrypted reasoning or a cache key whose ownership/portability is unknown. **Sticky-route only.** If unavailable, return a typed continuity error and let OpenCode/TradeAgent rebuild from supported plain-context/checkpoint information.

**C2 — Explicit change of model family.** Not transparently equivalent even if both models are capable. TradeAgent must explicitly allow and record a new logical model choice; important/reproducible work SHOULD checkpoint and open a new attempt. The gateway never rewrites alias A into model B secretly.

**C3 — Unsupported or mixed state.** Fail closed; no lossy recreation of messages, tool calls or reasoning. A human or TradeAgent manager handles recovery.

## 7.3 Prompt caching rules

Cache keys and pricing are provider-specific. Do not claim cached tokens or transferable context just because OpenCode continues its session. For capacity and cost purposes record `cache_read`, `cache_write`, `uncached_input`, `output`, `reasoning`, and unknown categories **only when real upstream usage data supports them**; mark missing values `unknown` rather than zero. A route switch may destroy cache locality and increase billable input even when logical history is unchanged. Default to sticky routing for a compatible live session and only optimize for cost if observed/policy permitted.

## 7.4 Reproducibility and TradeAgent JEV

JEV's probabilistic, non-text-generating **agent-context monitoring and manager-led handoff** process belongs wholly to TradeAgent. This does not prohibit **a separate JEV-powered operational advisory inside AgentPool**, strictly limited to redacted routing-health signals defined in Section 8.5. An AgentPool rate-limit event may be **reported** to TradeAgent as infrastructure telemetry, but AgentPool does not instruct JEV, summarize an agent's mind, start replacement agents or decide strategic task continuation. TradeAgent's manager/handoff worker handles checkpoints if OpenCode cannot continue. Physical-model provenance must be available for later audit of AI-assisted research; do not mix different models invisibly in a scientific result or trading decision.

---

# 8. Capacity, budgets, admission and accounting

## 8.1 Capacity is scoped, not an invented token balance

Key rate-limit/capacity observations by the strongest verifiable provider-specific scope:

```text
(provider, capacity_owner, API project/workspace/subscription product,
 model or quota bucket, region, client/app if applicable,
 window identifier/reset observation)
```

A registration may have separate request and token limits; distinct registrations can share the same organization allowance. Track values with `observed_at`, `expires_at`, `source`, `confidence`, `limit_scope`, `remaining_if_provided`, `reset_if_provided` and `cooldown_until`. No assumptions that an account has a fixed replenishing quota, that a paid subscription equals an API rate limit, or that a 429 proves another route has capacity. Unknown values remain unknown.

## 8.2 Conservative reservation and limits

Before dispatch reserve one concurrency slot and an estimated worst-case billable exposure based on admitted model, request length, `max_output_tokens`/equivalent, per-route unit price and host cap. Reject if the envelope cannot be bounded or a required price is unknown under a strict financial cap. Financial budgets are a **control**, not a guarantee that API fees can never be charged after network ambiguity. Account for known provider retries from OpenCode, tool-initiated additional model calls, compaction and background/model utility calls.

Suggested initial resource defaults are *configuration examples*, not source-backed performance guarantees:

| Control | Starting value | Policy |
|---|---:|---|
| Max request body | 8 MiB | Reject larger with `413`, tune for certified multimodal cases |
| Concurrent upstream streams | 8 per daemon | Bound globally **and** per capacity owner; tune by tests |
| Admission waiting queue | 32 | Bounded/fair; reject or apply backpressure |
| Connect deadline | 10 s | Narrower than total inference deadline |
| First-response/header deadline | 120 s | Adjust for verified reasoning models |
| Idle stream deadline | 300 s | Must not prematurely abort intentionally long model reasoning |
| Max model attempt runtime | 20 min | Client-cancellation and provider limits may be shorter |
| Manual future-route change | At next request | Scope drain only if route profile requires |
| Request retry count inside AgentPool | **0** | Keep zero by default; any safe model-turn retry is owned by the pinned OpenCode runtime, tested under Section 6.8 |
| Unbounded provider monetary fallback | **Never** | Explicit grant required for any metered path |
| External quota check per ordinary request | **0** | Consume cached/observed capacity state; no remote probe in hot path |
| JEV advisory wait per ordinary request | **0** | Optional asynchronous snapshot only; failure returns to baseline |

Every limit is enforced *in code* with a test, not only shown in the CLI. Model-specific context and output limits are sourced from verified manifests; never copy arbitrary example values into production admission.

## 8.3 Usage ledger

Every attempt produces a durable, redacted record:

```text
AttemptLedger {
  attempt_id, request_id_hash, client_id, grant_id, model_alias,
  physical_route_id, provider_principal_ref, capacity_owner_ref,
  credential_epoch, route_epoch, upstream_model_id, protocol_profile,
  billing_class, request_started_at, dispatched_at, completed_at,
  upstream_status, completion_state, usage_source, usage_tokens_by_type,
  price_catalog_revision, estimated_cost, reconciled_cost,
  billing_uncertain, failure_reason, policy_epoch
}
```

Never expose the real upstream credential to TradeAgent. Expose only enough provenance for risk review: **effective physical model/version, provider, billing class, observed usage/cost, compatibility class, and whether continuation was certified**. Physical account identifiers can be redacted hashes according to authorization.

Where the provider supports reliable usage/billing reconciliation, ingest it later and resolve `UNCERTAIN` records. There is no guarantee that a streamed response includes usage, or that stream cancellation prevents a charge.

## 8.4 Fairness, contention and health

- Maintain route-scoped and capacity-owner-scoped semaphores, not an unbounded global goroutine queue.
- Use fair scheduling among granted clients within configured weights; do not let a noisy TUI exhaust a critical TradeAgent service grant.
- In default `auto_eligible` mode, use sticky preference and deterministic, fair route selection at **every new** request boundary; when none are eligible return `NO_ELIGIBLE_ROUTE` with an actionable reason. No manual selection is required for normal operation.
- Avoid a high-frequency credential/limit-probing loop; use official read-only quota/headers where available and owner-approved probes only when charged work is budgeted.
- Model local backend availability separately from remote provider quota; GPU/OOM cooldowns are not subscription limits.
- Optimize against provider *observed* capacity using pre-dispatch reservations and request-specific headroom; JEV advisory is never consulted synchronously and cannot change eligibility (Sections 5.6 and 8.5).

## 8.5 Optional JEV route-risk advisory: operational, asynchronous and bounded (NORMATIVE boundaries; DESIGN implementation)

### 8.5.1 Why JEV belongs here — and where it does not

A TypeSafe System One/JEV model accepts `state` plus typed `Choice`, `Score`, and `Noul` questions and returns bounded probability information. That makes it plausible for **semantic judgments** over ambiguous or unfamiliar operational evidence, e.g. 'Does this sanitized error pattern resemble short-lived provider saturation rather than a revoked credential?' or 'Does the sequence resemble broad provider degradation rather than a single principal failure?'. It is **not** suitable for exact remaining-token arithmetic, time comparisons, reserving capacity, permission checks, API protocol translation, or inferring a quota where the provider publishes none. The route should work *without* JEV before enhancement.

**Recommended implementation:** one optional small **in-process advisory worker** (bounded mailbox, configurable wakeup on meaningful events or coarse cadence) reads **sanitized aggregate** health summaries from the same telemetry that deterministic admission uses. It calls a single independently configured typed-decision provider adapter, subject to strict timeout, cost budget, data policy and circuit breaker. It writes small TTL-tagged observations into an immutable copy-on-write advisory snapshot. Request admission reads that snapshot without blocking; if stale/unavailable, selection proceeds with deterministic baseline. Do not run a Jev decision request per inference request or per account; coalesce by verified capacity owner/provider cluster and by event window, with enough separate keys to preserve meaningful scope.

### 8.5.2 Example inputs and questions (ILLUSTRATIVE, NOT CALIBRATED)

```json
{
  "state": {
    "provider_family": "provider_example",
    "scope_kind": "capacity_owner",
    "outcome_pattern": "intermittent pre-output saturation rejections across peer routes",
    "time_window_class": "recent",
    "known_provider_reset_observed": false,
    "hard_quota_class": "unknown",
    "provider_errors": ["capacity_rate_limited", "transient_upstream_unavailable"],
    "sanitized_notes": "status classifications only; no upstream raw error text"
  },
  "model": "jev-1.13.0",
  "questions": {
    "systemic_degradation": {
      "type": "noul",
      "instructions": "Does this sanitized pattern resemble a broader transient provider service degradation, rather than an isolated credential failure?"
    },
    "failure_pattern": {
      "type": "choice",
      "instructions": "Which single observed operational pattern best matches the supplied state?",
      "criteria": {
        "transient_capacity": "Repeated temporary quota/saturation rejections, without evidence of revocation",
        "provider_instability": "Intermittent provider-side failures across several independently permitted routes",
        "single_registration": "Failures concentrated on one credential or principal",
        "insufficient_evidence": "State is too sparse or contradictory to assign one pattern"
      }
    }
  }
}
```

This demonstrates **request shape**, not available live credentials, an accurate model prediction, or a working production adapter. Real implementation pins a tested model revision and typed schema, builds state from deterministic **classifications**, never transmits raw prompts, source files, credentials, private full error messages or per-agent content to the external advisor. JEV receives only metadata for the allowed provider scope. Avoid presenting absolute times or exact quantities as information JEV must compute.

### 8.5.3 Advisory output contract

```text
RouteRiskAdvisory {
  advisor_id, version, rubric_revision, dataset_revision,
  scope_key_hash, snapshot_epoch,
  evaluated_at, expires_at,
  typed_answers, confidence_or_uncertainty,
  measured_calibration_band,
  shadow_or_enabled,
  bounded_soft_penalty,     // always nonnegative and within code-defined cap
  provenance_redacted,
}
```

**Hard boundary:** deterministic predicate constructs the *eligible* set. The advisor cannot add a route to this set, remove a hard cooldown, change `Retry-After`, override manual pin, introduce another billing class, downgrade context/model compatibility, lift resource bounds, alter credentials, change the request, approve retry, or terminate an in-flight stream. It can only change relative *preference* within the eligible set and within explicit fairness/stickiness bounds. Prefer a *penalty only* (never bonus above baseline) with minimum evidence, TTL, epoch, origin and reason. A suspicious classification may create an owner-visible diagnostic, but may **not** silently disable a route absent separate deterministic evidence or owner action. With no prediction JEV has exactly zero influence.

### 8.5.4 Model quality and anti-flapping policy

1. Start **disabled**. Record a fixed offline evaluation set of sanitized failure/health traces and the exact future labels from provider outcomes. Compare JEV against deterministic `Retry-After`, capacity reservation, EWMA failure-rate, and per-owner cooldown baselines before any traffic influence.
2. Shadow only: evaluate prediction Brier score/log loss, expected calibration error, precision/recall for provider-wide incident, per-provider and per-model breakdown, false-positive route avoidance, false-negative failure, wasted viable capacity, p95/p99 added dispatch latency and total advisor spend. Compare at matched workloads; use held-out future periods, not mixed windows leaking outcomes.
3. Promote only when materially better on pre-registered targets, no meaningful fairness/cost regression, privacy passes and adequate confidence under **domain-specific** calibration. JEV's general confidence is not a validated probability of provider failure. Keep a no-advisor control cohort and permit immediate rollback to deterministic baseline.
4. Bound penalties, expiry and update cadence; use hysteresis/minimum residence to prevent candidate oscillation. Hard provider cooldown and reset rules always prevail. An unavailable model, malformed output, stale advisor, noisy signal, provider-side JEV rate limit, cost cap or advisory timeout must not change the hot path.
5. Treat Jev model or rubric revision as a new uncalibrated artifact; revert to shadow until revalidated. Model dependency independence: the advisor failure must never exhaust or circularly depend on the inference routes it is advising.

**Promotion gate:** the optional advisor is **not** required for AgentPool to meet its core automatic-pooling release definition. It may be released only after measurable, repeatable marginal value against simple deterministic alternatives.

## 8.6 High-speed capacity bookkeeping and correctness rules

- Each authoritative limiter dimension uses **provider-defined capacity owner and scope**. Examples include project, workspace, organization, model family, request tokens, input/output tokens, regional allowance, or local GPU concurrency; sharing is not guessed from key count. Different constraints may overlap.
- Update authoritative provider observations from ordinary response headers, usage reports and documented read-only quota APIs, when available. Distinguish *request observed headers* from *inferred local usage*. No request-path `/limits` call. Missing data stays unknown; stale observations do not become an unearned grant.
- Reserve capacity at the **capacity-owner + limiter-dimension** level with atomic compare-and-reserve for concurrent threads, and keep a distinct dispatch/billing ledger. An upper-bound estimate can be wrong; treat estimates as safety headroom signals, not provider guarantees.
- For active request cancellation or crash, release only locally controlled concurrency and record billable/token uncertainty; do not roll back spent remote usage because the stream disappeared.
- Capacity owner can share multiple physical routes. When one route sees a 429 that provider documentation ties to a shared owner, cool down *that owner/dimension*, not just the API key; when scope is unknown, apply a cautious scoped block without asserting global exhaustion.
- Separate hot path from observation refresh and advisory scoring. Response metadata updates must be bounded and backpressure-safe, admission counters must remain correct even if noncritical logging is delayed. No unbounded goroutines, queues, scans or per-request subprocesses.
- Use request deadlines, fair bounded queues and policy priorities; never let a slow, retrying background agent starve a high-priority workload. A system starting 100 agents does not imply 100 active model streams; only simultaneous admitted requests consume upstream slots.


---

# 9. Threat model, service exposure and security

## 9.1 Assets and adversaries

Assets: upstream API keys and subscription-connector credentials; work/prompt content in transit; authorization grants; financial budget caps; model and billing provenance; user and TradeAgent project data; process configuration. Threats: another local user/process, hostile repository configuration, malicious OpenCode plugin/skill, compromised worker, untrusted prompt/tool result, confused-deputy client, accidental LAN exposure, malicious upstream redirect, stale control message, concurrent operator action, logs leaking secrets, compromised update and model alias spoofing.

## 9.2 Separate the data and control planes

**Data plane:** local loopback HTTP, token authenticated, exact inference path allowlist and per-grant model policy. No account listing, route selection, token export, admin action or unrestricted proxy path on that interface.

**Control plane:** owner-only Unix socket (macOS/Linux; private `0700` parent, `0600` socket, checked peer identity) or owner-only Windows named pipe with verified SID and restrictive ACL. Use versioned JSON-RPC 2.0 or an equally strict typed protocol. Caller identity, per-method authorization, maximum message length and monotonic policy epoch are mandatory.

`localhost` alone is not an access-control guarantee. Any same-user untrusted program may reach loopback; the gateway bearer and scoped control IPC are still required. Browser-origin requests are rejected; don't enable permissive CORS. No LAN/WAN listener in v1. Proxy headers never affect trusted peer identity.

## 9.3 Outbound egress

- Pinned scheme/hostname/path for each upstream registration; HTTPS with certificate verification for remote APIs. Local plaintext allowed only to explicitly trusted loopback endpoint.
- Reject redirects; do not follow user-controlled URLs, `Host`, `X-Forwarded-*`, absolute-form proxy requests or arbitrary destination overrides.
- Disable ambient HTTP_PROXY/HTTPS_PROXY when it would subvert endpoint restrictions, or use a tightly controlled reviewed proxy policy.
- Parse addresses against rebinding, localhost aliases, IPv4/IPv6 discrepancies, credentials in URLs and local-file/socket schemes.
- Upstream key lives only in the AgentPool process and its protected OS vault; never write to `opencode.json`, OpenCode auth.json, prompts or environment inherited by a general-purpose worker.

## 9.4 Privacy and data minimization

- Do not persist prompt text, source files, tool outputs, reasoning tokens, provider response bodies or raw OAuth tokens.
- **Advisor privacy:** TypeSafe/JEV receives only redacted aggregate operational summaries (not user prompts, raw provider error bodies, source paths, secrets, account identifiers or per-user work metadata). External decision calls are opt-in with provider allowlisting, reviewable retention/terms, separate scoped credentials and a spending cap. Treat potentially adversarial upstream error strings as untrusted: classify/strip in code before using as adviser state.
- Prefer structured metrics with hashed attempt/route IDs; strip auth headers, cookies, user code, stack traces containing credentials and raw provider errors from ordinary logs.
- Optional debug traces require **local explicit opt-in**, bounded retention, secure file permissions and visible privacy warning. Replay fixtures use synthetic data by default.
- Provider-side retention and data processing are controlled by the effective upstream route; respect TradeAgent project classification and geographic restrictions before admitting a call.
- Encrypt backups and never include decrypted secret-vault material by default.

## 9.5 Security is independently enforceable

TradeAgent's own containment, file permission, tool allowlist and trading risk kernel remain in force. A malicious agent prompt cannot call AgentPool's owner control API or obtain another project grant. A separate project with the same human operator is not automatically permitted to share its prompt data or subscription principal with the current worker. No model inference request is allowed onto protected live order execution paths; broker commands stay behind TradeAgent's independently validated tool/risk/reconciliation controls.

## 9.6 Kill switches

Owner can disable one registration, one provider, one alias, one client grant, or all outward inference. Emergency revoke must immediately stop *new* admissions and attempt best-effort cancellation of affected streams, recording uncertainty. A disabled route cannot be reactivated by automatic fallback, config reload, token refresh or a stale worker.

---

# 10. Data model, persistence and recovery

## 10.1 Small single-service design

Choose one Go executable, one SQLite database with WAL and schema migrations, OS-backed secret store, a small HTTP listener, and a private control socket. Do not add Kafka, Redis, Kubernetes, Kubernetes operators, an agent scheduler, distributed queues, gRPC mesh, vector database, RAG database or arbitrary plugin host in the MVP. A few bounded in-memory maps and semaphores are sufficient for active reservations; immutable route/policy metadata and attempt disposition are durable.

The core provider adapters should share **one** interface:

```go
// DESIGN contract, illustrative types. Exact Go definitions are versioned in code.
type Adapter interface {
    Capabilities() CapabilityManifest
    Validate(req RequestEnvelope, route PhysicalRoute) error
    BuildUpstream(req RequestEnvelope, route PhysicalRoute) (UpstreamRequest, error)
    Stream(ctx context.Context, upstream UpstreamResponse, sink EventSink) (Usage, Outcome, error)
    Classify(err error, status int) FailureClass
}
```

The allocator depends on the **interface and certified manifest**, not upstream vendor logic. The adapter cannot choose a different credential, billing class or endpoint after the allocator has frozen the route.

## 10.2 Minimal logical schema

All primary keys are opaque random IDs except monotonic generation/fence counters. Dates are UTC with a clearly documented clock model.

```sql
-- Logical migration 0001; adjust field types and constraints in actual migration.
CREATE TABLE providers (
  id TEXT PRIMARY KEY,
  enabled INTEGER NOT NULL,
  manifest_revision TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE registrations (
  id TEXT PRIMARY KEY,
  provider_id TEXT NOT NULL REFERENCES providers(id),
  principal_ref TEXT NOT NULL,
  credential_vault_ref TEXT NOT NULL,
  credential_epoch INTEGER NOT NULL DEFAULT 1,
  billing_class TEXT NOT NULL,
  capacity_owner_ref TEXT NOT NULL,
  enabled INTEGER NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE connector_certifications (
  id TEXT PRIMARY KEY,
  registration_id TEXT NOT NULL REFERENCES registrations(id),
  app_identity TEXT NOT NULL,
  credential_mode TEXT NOT NULL,
  protocol_profile TEXT NOT NULL,
  state TEXT NOT NULL CHECK(state IN ('UNCONFIGURED','PROBING','CERTIFIED','DISABLED')),
  adapter_revision TEXT NOT NULL,
  test_evidence_ref TEXT,
  expires_at TEXT,
  updated_at TEXT NOT NULL
);
CREATE TABLE physical_routes (
  id TEXT PRIMARY KEY,
  registration_id TEXT NOT NULL REFERENCES registrations(id),
  upstream_endpoint_ref TEXT NOT NULL,
  upstream_model_id TEXT NOT NULL,
  model_revision TEXT NOT NULL,
  adapter_id TEXT NOT NULL,
  adapter_revision TEXT NOT NULL,
  capability_manifest_hash TEXT NOT NULL,
  enabled INTEGER NOT NULL,
  generation INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE model_aliases (
  id TEXT PRIMARY KEY,
  protocol_profile TEXT NOT NULL,
  model_identity TEXT NOT NULL,
  capability_manifest_hash TEXT NOT NULL,
  policy_epoch INTEGER NOT NULL,
  enabled INTEGER NOT NULL
);
CREATE TABLE alias_route_allowlist (
  alias_id TEXT NOT NULL REFERENCES model_aliases(id),
  route_id TEXT NOT NULL REFERENCES physical_routes(id),
  proof_ref TEXT NOT NULL,
  PRIMARY KEY(alias_id, route_id)
);
CREATE TABLE client_grants (
  id TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  workload_ref_hash TEXT NOT NULL,
  bearer_hash TEXT NOT NULL,
  grant_policy_json TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  revoked_at TEXT,
  policy_epoch INTEGER NOT NULL
);
CREATE TABLE route_selection (
  scope_ref TEXT PRIMARY KEY,
  chosen_route_id TEXT REFERENCES physical_routes(id),
  selection_mode TEXT NOT NULL,
  generation INTEGER NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE capacity_observations (
  id TEXT PRIMARY KEY,
  capacity_owner_ref TEXT NOT NULL,
  scope_json TEXT NOT NULL,
  source TEXT NOT NULL,
  observed_at TEXT NOT NULL,
  expires_at TEXT,
  remaining_json TEXT,
  cooldown_until TEXT,
  confidence TEXT NOT NULL
);
-- Optional advisory table is metadata-only and independently expirable;
-- do not store JEV query raw state or model responses with sensitive identifiers.
CREATE TABLE route_advisory_snapshots (
  scope_key_hash TEXT PRIMARY KEY,
  risk_model_revision TEXT NOT NULL,
  rubric_revision TEXT NOT NULL,
  observed_snapshot_epoch INTEGER NOT NULL,
  expires_at TEXT NOT NULL,
  bounded_penalty REAL NOT NULL CHECK(bounded_penalty >= 0.0),
  shadow_mode INTEGER NOT NULL,
  evidence_ref TEXT
);
CREATE TABLE inference_attempts (
  id TEXT PRIMARY KEY,
  client_grant_id TEXT NOT NULL REFERENCES client_grants(id),
  request_hash TEXT NOT NULL,
  model_alias TEXT NOT NULL,
  route_id TEXT NOT NULL REFERENCES physical_routes(id),
  route_generation INTEGER NOT NULL,
  credential_epoch INTEGER NOT NULL,
  state TEXT NOT NULL,
  dispatched_at TEXT,
  committed_at TEXT,
  finalized_at TEXT,
  usage_json TEXT,
  billing_uncertain INTEGER NOT NULL DEFAULT 1,
  failure_code TEXT,
  policy_epoch INTEGER NOT NULL
);
CREATE TABLE audit_events (
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at TEXT NOT NULL,
  action TEXT NOT NULL,
  actor_ref TEXT NOT NULL,
  target_ref TEXT,
  result TEXT NOT NULL,
  details_redacted_json TEXT
);
```

**Production migration requirements:** enforce uniqueness/foreign keys/indexes, versioned schemas and check constraints; store no raw private bearer, prompt or secret; hash bearers with a suitable resistant lookup scheme and compare safely; bind `request_hash` to idempotency key; enforce transactional grant/attempt/counter updates; persist selection generations atomically; detect incompatible schema versions and stop rather than resetting the database. The core schema is intentionally compact enough to review.

## 10.3 Atomic transactions

- **Grant issuance:** verify caller/role/parent constraints, reserve budget envelope and mint random scoped token in one audited operation.
- **Selection:** owner-authorized control action increments route-selection generation; new requests atomically observe that generation. In-flight attempts retain immutable old generation. An advisory penalty is a soft, ephemeral snapshot; it does not mutate the route selection generation, eligibility, granted privileges or active attempts.
- **Dispatch:** create attempt + reserve concurrency/budget and freeze route BEFORE sending. Mark `dispatched_at` durably at the appropriate pre-send boundary; conservative ambiguity survives crash.
- **Completion:** settle observed usage/cost and release held capacity once; duplicate finalizers cannot double-release.
- **Revocation:** atomically revoke a grant/registration and increment the policy epoch; stale renewals/requests cannot reopen it.

Protect against SQLite busy contention with bounded, tested retries only for **local database operations**, never conflating it with provider inference retries. Choose backups through SQLite's supported consistent backup mechanism; test restore and schema rollback.

## 10.4 Crash states

| Crash location | Restart interpretation | Safety response |
|---|---|---|
| Before admission committed | No dispatched attempt | Reject/reissue with new authorized request |
| After reservation but before provable send | Reservation may be stale | Reconcile local resource only; no outgoing replay |
| After upstream send, before any reply | `OUTCOME_AMBIGUOUS` | Usage exposure unknown; no automatic replay |
| During downstream stream | `PARTIAL_OUTPUT` or `OUTCOME_AMBIGUOUS` | OpenCode/host must detect interrupted turn |
| After upstream completed, before ledger finalized | Provider charge may exist | Reconcile from durable observed data; never re-send |
| During manual route change | Old or new committed generation, never half-selection | Verify atomic epoch and restore consistent route |
| During credential refresh | Previous or newer credential epoch | Single-flight reconcile/reconnect; never resurrect revoked key |

---

# 11. Owner control surface and host-facing observability

## 11.1 CLI semantics (DESIGN)

```bash
agentpool serve                                  # local daemon
agentpool status                                 # redacted health / route capacity
agentpool doctor                                 # pinned OpenCode/protocol/security diagnosis
agentpool capacity                               # redacted, source-scoped authoritative observations
agentpool advisor status                         # disabled/shadow/enabled, health, model/rubric version
agentpool advisor disable                        # owner-only, immediate return to baseline
agentpool providers                              # registered adapters and certification
agentpool routes                                 # configured/certified routes and admission reasons
agentpool models                                 # logical catalog / compatibility matrix
agentpool credential add                         # owner-only, documented credential flow
agentpool credential revoke <registration-id>     # stop new work and invalidate grant
agentpool select <alias> <route-id>               # owner manual_pin override (next requests)
agentpool auto <alias>                            # clear override; resume automatic routing
agentpool pool inspect <alias>                    # show all eligible/ineligible routes and reasons
agentpool pause <alias-or-route>                   # deny new admissions
agentpool resume <alias-or-route>                  # only if route still certified and enabled
agentpool grants list                             # masked scope and expiry
agentpool audit --since <cursor>                  # metadata-only events
```

These are **proposed commands**, not an assertion that the uploaded prototype already implements them. Automatic per-request selection requires **no operator call**. `select` is only an optional manual pin and MUST display old/new **provider, model, billing class, principal hint, scope, effect on current streams and admission decision**, and require interactive confirmation (or a separately authenticated owner control API call). `auto` clears the pin through owner-only IPC. Confirmation guards against accidental route or billing changes. Changing a non-equivalent model through `select` is **rejected**; use an explicit OpenCode/TradeAgent logical-model change instead.

## 11.2 Control IPC contract

Versioned owner/host API prefix: `ap.v1`. Management calls are not exposed on the inference HTTP port.

| Method | Authorization | Side effect |
|---|---|---|
| `hello` / `health` | Authenticated peer | Negotiate protocol, report capabilities |
| `credential.list/add/revoke` | Owner | Protected credential lifecycle |
| `provider.list/probe` | Owner (probe may cost) | Adapter availability/certification |
| `model.list` | Authorized host | Filtered aliases, limits, capability profiles |
| `route.inspect` | Authorized owner/host | Redacted eligible set/reason codes |
| `selection.preview/commit` | Owner and permission gate | Optional manual pin/unpin of future-request route; automatic routing remains default |
| `pool.inspect` | Authorized owner/host | Explain eligible/ineligible routes and capacity/certification reasons without exposing secrets |
| `grant.issue/renew/revoke` | Owner or delegated host | Scoped capability token and budget |
| `usage.read` | Authorized owner/host | Attempt and cost metadata for its scope |
| `events.read` | Authorized owner/host | Durable redacted events after cursor |
| `pause/resume/shutdown` | Owner | Route/service controls |
| `doctor` | Owner | Config/upgrade/security/release checks |
| `capacity.inspect` | Owner or authorized host | Read scoped capacity evidence/last update and uncertainty |
| `advisor.status/disable` | Owner | Advisory state, calibration/version/last-error; disable to restore baseline |

For each method, implement committed JSON Schema, versioning, authentication, authorization matrix, idempotency semantics, timeout, error code and side-effect test. Do not invent a client API that grants global credential-switch authority to untrusted OpenCode agent prompts.

## 11.3 Inference API authentication

Every OpenCode worker gets **one short-lived, scoped, random gateway credential** with audience `agentpool.inference`, allowed logical aliases, max budget/time/concurrency, app/workload identity and expiry. The bearer is bound to the caller's OS/process isolation where technically possible, and never accepted on owner control IPC. On request, the server looks up a non-plaintext representation of that bearer, checks effective grant and policy epoch, and rejects any model/route/billing outside its envelope.

**No trusted security metadata from raw model headers.** An agent could edit its own HTTP payload; `X-TradeAgent-Role`, `X-Account`, `X-Route` and similar caller fields are hints only and cannot enlarge authority. Authorized host control calls mint grants **before** running OpenCode.

## 11.4 Necessary information returned to TradeAgent

Since AgentPool is only a provider, the OpenCode model response remains a **normal inference stream**. The host may separately query `usage.read` and subscribe to metadata events to learn:

```text
request_id_hash / workload_ref
logical_model_alias
actual_provider + upstream_model_revision
physical_route_pseudonym + billing_class
capability_profile + policy_epoch
attempt_state + failure_class
observed_tokens/cost, uncertainty and capacity cooldown
```

Do not require TradeAgent's tool loop to parse private AgentPool JSON inside model text. Do not expose an unrelated app's routing metadata to a different app.

---
# 12. TradeAgent integration — deliberately minimal

## 12.1 No AgentPool host-to-agent bridge

The `OpenCodeRuntimeAdapter`, headless worker process management, event subscription, agent instructions, worktrees, tool grants, JEV monitoring, task retries and durable handoff remain inside **TradeAgent**. **Migrating TradeAgent from its existing Codex runtime to OpenCode is a separate TradeAgent-owned task.** After that migration, AgentPool does not add a second agent runner: changing an eligible direct provider to `agentpool/<alias>` affects the worker's model provider settings and its scoped infrastructure grant, **not who owns the job, which OpenCode host API executes it, or how OpenCode calls tools**. AgentPool's integration consists of only:

1. **Provisioning:** TradeAgent's trusted host requests a scoped model-access grant for a worker from AgentPool control IPC.
2. **Configuration:** TradeAgent gives that worker a private OpenCode configuration whose permitted model/provider references are `agentpool/<alias>` and whose local inference bearer is the worker's grant.
3. **Execution:** TradeAgent runs the OpenCode agent via its chosen actual versioned CLI/server/SDK API. OpenCode talks directly to AgentPool on every model step.
4. **Telemetry:** TradeAgent correlates provider attempts/usage/physical-model provenance through AgentPool's redacted metadata interface and its own OpenCode event stream.
5. **Recovery:** TradeAgent decides what to do with stopped agents, failed inference, tool side effects, checkpoints or model changes. AgentPool **does not** recover agents.

```text
TradeAgent               OpenCode                  AgentPool            Provider
    |                       |                         |                  |
    |-- issue model grant -------------------------->|                  |
    |<----------------- scoped local token ----------|                  |
    |-- create native session -->|                   |                  |
    |-- prompt/run ------------->|                   |                  |
    |                       |-- model request ------->|                  |
    |                       |                    auth + route            |
    |                       |                         |--- inference --->|
    |                       |                         |<-- model SSE ----|
    |                       |<-------- model SSE -----|                  |
    |                       |-- run permitted tools   |                  |
    |<-- native events ---------|                     |                  |
    |                       |-- next model request -->|                  |
    |                       |                  choose next eligible        |
    |                       |                  route if needed          |
    |                       |                         |--- inference --->|
    |<--- usage metadata ----------------------------|                  |
```

There is **no** `AgentPool.StartWorker`, `AgentPool.SessionPrompt`, `AgentPool.RunTools`, `AgentPool.Handoff`, or `AgentPool.Trade` operation. If a code review introduces one, it is architecture drift unless it is an independently justified product change.

## 12.2 App-owned alternative runtime paths

Earlier TradeAgent code/specs describe both an application-owned restricted API harness and a CLI worker. Whether to **retain** that restricted harness is a **TradeAgent architecture decision**, not a requirement to make AgentPool an agent executor. Under the desired **single managed inference boundary**, any retained app-owned model harness SHOULD point to AgentPool as its provider instead of independently maintaining upstream credentials/selection. But no pre-existing code path may be silently rewired or granted more tools/authority during this migration. The protected broker/risk kernel must remain application-owned under either runtime.

## 12.3 Safe OpenCode model changes

When the current logical alias is incompatible with a new model family, TradeAgent may make an explicit OpenCode model-selection call through its pinned native host API, but MUST preserve task authority, tool state and outcome provenance. If native same-session model switch is unreliable, it must checkpoint, start a fresh native session and resume with a manager-approved handoff. These are TradeAgent operations. AgentPool only refuses false equivalence and exposes the new alias on the inference side.

## 12.4 Personal CLI mode

AgentPool remains useful without TradeAgent. A user can run OpenCode interactively with `agentpool/<alias>`, use the same safe gateway and select **certified** physical API and account-based routes through AgentPool's owner CLI. The personal-mode grant does not inherit TradeAgent's project data authority or financial budgets; it has a separate local purpose and permission scope. No personal TUI call can spend a TradeAgent API budget merely because the same device/user owns both environments.

## 12.5 Independent JEV deployments: operational telemetry, not a new control bridge

```text
TradeAgent JEV                         AgentPool JEV (optional)
  evaluates context quality             evaluates SANITIZED route-health patterns
  signals manager                       produces expiring soft risk advice
  manager owns handoff                   router owns deterministic eligibility
  TradeAgent owns agent memory           AgentPool owns no agent/session data

TradeAgent  <- read-only redacted route/cost/continuity events -- AgentPool
```

Both may use TypeSafe's typed decision interface, but their contracts, grants, data scope, rubrics, time horizons and thresholds are separate. AgentPool does **not** send model prompts or agent reasoning to its JEV. When `PROVIDER_STATE_BOUND` or prolonged `NO_ELIGIBLE_ROUTE` prevents safe continuation, AgentPool emits normal typed status/telemetry; **TradeAgent's own manager** decides whether to wait, choose a new logical model, or invoke its existing JEV-monitored checkpoint/handoff process. No gateway-to-JEV instruction may directly stop, resume or relaunch an OpenCode worker.


---

# 13. Repository structure and engineering rules

## 13.1 Production-oriented compact repository

```text
agentpool/
├── cmd/agentpool/                 # one binary: serve, select, status, doctor, etc.
├── internal/
│   ├── gateway/                   # HTTP admission, request isolation, stream relay
│   ├── catalog/                   # logical aliases, limits and capabilities
│   ├── routing/                   # route eligibility, affinity, modes, epochs
│   ├── capacity/                  # scoped limiter observations, reservations, headroom
│   ├── advisory/                  # OPTIONAL async typed JEV decisions and calibration
│   ├── adapters/
│   │   ├── openai_chat/           # certified Chat Completions L0/L1
│   │   ├── openai_responses/      # certified Responses L0/L1
│   │   ├── anthropic/             # deferred until certified
│   │   ├── google/                # deferred until certified
│   │   └── local_compatible/      # model-specific certification
│   ├── credential/                # vault references, refresh, revocation
│   ├── policy/                    # service permissions and authority
│   ├── budget/                    # reservations, usage, capacity ownership
│   ├── storage/                   # SQLite migrations and durable events
│   ├── control/                   # owner IPC JSON-RPC
│   └── telemetry/                 # safe structured status and audit metadata
├── specs/                         # JSON Schemas and normative state machines
├── examples/                      # version-specific OpenCode/AgentPool configs
├── fixtures/                      # replayable synthetic protocol traces
├── tests/                         # unit, integration, fault, live-gated, platform
├── compatibility/                 # pinned, hash-identified test proofs
├── docs/                          # runbook, install, security, provider policy
├── go.mod
└── README.md
```

Only **one** implementation of request admission, identity, attempt-state transitions and policy evaluation. Do not duplicate them inside each adapter. The directory listing describes target modules: minimal MVP files MAY be merged where a strong single abstraction is clearer. Tests and schemas MUST be versioned alongside the component they certify.

## 13.2 Engineering defaults

- Prefer standard-library Go for the HTTP relay and core service; add a dependency only for a demonstrable engineering advantage after security/version review. SQLite driver and OS vault bindings are deliberate dependencies.
- Keep all provider configuration immutable during each request. Config reload creates a new epoch; old in-flight routes retain their frozen snapshot.
- Hot-path admission uses zero external probing and zero JEV wait; bound local capacity-owner reservation contention, always test races and stale epochs.
- TypeSafe model adapter, if present, is behind one optional typed `RiskAdvisor` interface; disable/revert without changing the baseline router and without a new daemon/agent execution dependency.
- Use explicit `context.Context` cancellation, bounded concurrency, bounded request and SSE parsing, no uncontrolled goroutine spawning, and typed no-secret errors.
- Go `race`, `vet`, fuzz and property tests are part of CI, not optional after deployment.
- Use stable opaque IDs, strict enum decoders, `CHECK`/unique constraints, migrations and typed schemas. Reject unknown security-significant config fields or revisions.
- Build tests against real captured *synthetic* OpenCode wire shape, not an invented hand-written OpenAI API request.
- Keep the documentation truthful: mark anything not run on a real installed OpenCode build `UNVERIFIED` even if Go fake-server unit tests pass.
- Sign/version released binaries and route manifests where available; pin checksums of OpenCode runtime, provider package, gateway adapter and test fixtures.

## 13.3 Route manifest (DESIGN example)

This is an **AgentPool-side illustrative YAML manifest**, not a currently implemented parser. It intentionally stores vault references, not credential strings:

```yaml
schema_version: 1
service:
  mode: local-owner
  inference_bind: 127.0.0.1:43819
  control: owner-private-ipc
  default_route_mode: auto_eligible
  inference_retry_count: 0
  allow_remote_client: false
  capacity_admission: conservative_pre_dispatch
  quota_probe_on_request: false
  advisor:
    mode: disabled  # disabled | shadow | calibrated_soft_penalty
    provider: typesafe_jev  # optional, NOT core hot path
    max_inflight: 1
    max_queue: 32
    request_timeout_ms: 1000  # illustrative; advisory only
    # credentials stored separately in OS vault; never inline

providers:
  - id: openai-api
    adapter: openai-responses
    adapter_revision: REPLACE_WITH_CERTIFIED_REVISION
    connector_certification: REPLACE_WITH_CERTIFICATION_ID
    # Provider destination is administrative config, never client request data.
    endpoint: https://api.openai.com/v1

registrations:
  - id: primary-api-project
    provider: openai-api
    credential_vault_ref: REPLACE_WITH_OS_VAULT_REFERENCE_A
    principal_ref: REPLACE_WITH_VERIFIED_PROJECT_ID_A
    billing_class: metered_api
    capacity_owner_ref: REPLACE_WITH_VERIFIED_CAPACITY_OWNER_A
    enabled: true
  - id: secondary-api-project
    provider: openai-api
    credential_vault_ref: REPLACE_WITH_OS_VAULT_REFERENCE_B
    principal_ref: REPLACE_WITH_VERIFIED_PROJECT_ID_B
    billing_class: metered_api
    capacity_owner_ref: REPLACE_WITH_VERIFIED_CAPACITY_OWNER_B
    enabled: true

routes:
  - id: primary-responses
    registration: primary-api-project
    upstream_model_id: REPLACE_WITH_REAL_SUPPORTED_MODEL_ID
    model_revision: REPLACE_WITH_VERIFIED_MODEL_REVISION
    protocol: responses
    proof_ref: REPLACE_WITH_PASSING_CERTIFICATION_A
    enabled: true
  - id: secondary-responses
    registration: secondary-api-project
    upstream_model_id: REPLACE_WITH_SAME_VERIFIED_MODEL_ID
    model_revision: REPLACE_WITH_SAME_VERIFIED_MODEL_REVISION
    protocol: responses
    proof_ref: REPLACE_WITH_PASSING_CERTIFICATION_B
    enabled: true

aliases:
  - id: responses-compatible
    downstream_protocol: responses
    # Two independently certified routes; selector chooses automatically.
    allowed_routes: [primary-responses, secondary-responses]
    routing_mode: auto_eligible
    allowed_billing_classes: [metered_api]
    # Tool, context and output capabilities come from certified manifest.
    # These two capacity owners are NOT independent unless verified upstream.

clients:
  - id: personal-opencode
    allowed_aliases: [responses-compatible]
    billing_classes: [metered_api]
    max_concurrency: 1
    requires_explicit_budget: true
```

**Do not blindly use the above placeholders in a working deployment.** The two records are a **schematic demonstration of automatic two-route selection**, not credentials or proof that their capacity owners are actually independent; project IDs/capacity-owner scope must come from authoritative information, and when unknown the corresponding cross-account quota-independence assumption cannot activate. A route can still be registered and examined while not admitted. Additional Anthropic/Google/local **API-key or certified account/subscription-backed routes** are separate adapter/registration/manifest records and may be added without moving any agent logic into AgentPool. They appear under their own truthful aliases unless the exact cross-route equivalence has been independently proven.

---

# 14. Conformance harness and blocking acceptance gates

## 14.1 The feasibility proof comes BEFORE the complete router

Build a fake compatible provider that records each received request and deterministically emits scripted events. Point a real, exact-version OpenCode client at a skeletal AgentPool listener. Run the following **before** designing provider rotation heuristics:

1. An interactive TUI prompt reaches `agentpool`, not a direct provider.
2. A headless prompt reaches the **same** AgentPool contract.
3. A multi-turn session including `read`/permitted tool action and follow-up model call completes with valid streamed tool events.
4. The built-in small/utility model, compaction, summarization and any permitted subagents also traverse AgentPool—or are explicitly disabled, with a test.
5. Fake-provider 429/401/5xx/timeout/broken-SSE/cancel cases visibly fail in OpenCode with correct typed outcome and **no silent success**.
6. Verify whether the selected OpenCode version performs automatic model retries and whether retries repeat full request/tool messages; fix or prohibit unsafe unattended paths.
7. Capture the exact native model provider package, endpoint, headers, stream event shape, settings and configuration merge result.
8. Run an **unattended automatic A→B switch** across independently registered same-effective-model credentials between completed inference turns after Route A is made ineligible; test continued context/tools and absence of duplicate effects. Also test manual pin/unpin as an exception.
9. Run a stateful provider-response-ID/opaque-reasoning test and **require sticky route or explicit continuity failure** when state is nonportable.
10. Run strict egress monitoring and prove no managed model call leaks to another upstream provider endpoint.

Only after those proofs should new provider-protocol translation, complex capacity forecasting and other advanced optimizations be constructed. **Basic automatic selection across two independently certified equivalent routes is required in MVP-B**, not deferred. Mock tests alone cannot prove actual provider billing, live opaque-context portability, authentication flow reliability or production OS isolation.

## 14.2 Core property/invariant suite

**Provider-parity test (required before claiming transparent integration):** run the **same** approved OpenCode host/agent fixture (session creation → model output → permitted tool call → subsequent model request → completion/cancellation) with (A) one exact upstream model connected directly and (B) the identical upstream model exposed via AgentPool. Prove both use OpenCode's native worker/session and tool-event interfaces; any required gateway-specific work is confined to client credentials, model-provider configuration and optional separately scoped infrastructure telemetry. Differences in network/usage metadata are expected; changes to agent control flow or unproven model semantics are not acceptable.

| ID | Release-blocking assertion |
|---|---|
| INV-01 | AgentPool owns no OpenCode job/session/tool/task lifecycle API |
| INV-02 | Every managed OpenCode inference path uses AgentPool or is disabled |
| INV-03 | No upstream network call before client, alias, grant, permission and budget admission |
| INV-04 | Client cannot choose upstream URL, model behind alias, provider principal, billing class or token |
| INV-05 | Alias never silently changes effective model identity/variant or billing mode |
| INV-06 | Exactly one frozen physical route per upstream dispatch attempt |
| INV-07 | Selection epoch changes affect future requests, not already streaming responses |
| INV-08 | In-flight streams never splice provider accounts or model streams |
| INV-09 | Invalid/expired credentials or uncertified connectors cannot become eligible via fallback |
| INV-10 | Operator-selected account/credential changes apply only to later requests, with isolated vault records and auditable route epochs |
| INV-11 | Provider principal/capacity owner sharing cannot be counted as independent quota |
| INV-12 | Every accepted route has a passing exact-protocol capability proof |
| INV-13 | Unknown context/opaque state cannot be stripped or routed to an unproven new principal |
| INV-14 | A truncated streamed answer cannot be reported as a successful completed turn |
| INV-15 | Gateway does not replay ambiguous provider requests or tool calls |
| INV-16 | Crash recovery never treats dispatched uncertainty as zero usage/cost |
| INV-17 | Grant/budget/revocation cannot be enlarged by prompt-controlled metadata |
| INV-18 | Secret material never appears in logs, trace headers, OpenCode config or upstream errors |
| INV-19 | Control methods cannot be invoked on inference HTTP endpoint |
| INV-20 | Physical endpoint destination cannot be supplied by caller or redirected |
| INV-21 | Unknown protocol fields/enums affecting security/correctness are rejected |
| INV-22 | A denied route remains denied after config reload, refresh, restart and fallback |
| INV-23 | All tool/broker/live-trading effects remain solely under TradeAgent's authority |
| INV-24 | One credential refresh cannot roll back a newer credential epoch |
| INV-25 | Multi-client local scopes cannot read each other's route/account/usage/grant data |
| INV-26 | Inactive, dead or revoked service grants cannot authorize new requests |
| INV-27 | SSE backpressure never permits unbounded memory growth |
| INV-28 | OS sleep, resume, clock jumps and ambiguous cooldowns do not create accidental permission |
| INV-29 | Deployment model catalog matches actual advertised capabilities and account entitlements |
| INV-30 | Any version/config/transport change invalidates affected certification until retested |
| INV-31 | An OpenCode worker with `agentpool/<alias>` is controlled through the **same native OpenCode host/CLI/SDK interface** used with a direct provider; AgentPool exposes no agent-control API |
| INV-32 | Physical-route selection, budget policy and credential custody are absent from OpenCode agent work/host lifecycle code; separately authenticated provisioning and telemetry are allowed |
| INV-33 | Direct-vs-AgentPool comparative multi-turn/tool fixture shows provider-only substitution, with no duplicate agent runtime or AgentPool-managed process |
| INV-34 | Owner/grant/usage control IPC is separate from inference transport; the phrase **provider only** cannot be used to remove necessary management-plane security |
| INV-35 | Existing TradeAgent Codex→OpenCode runtime migration is tracked and tested as TradeAgent-owned, never implemented as an AgentPool OpenCode-bridge workaround |
| INV-36 | Every admitted new model request automatically selects a currently eligible certified physical route for the requested alias; no operator action is needed for normal selection |
| INV-37 | When selected Route A is known ineligible, a subsequent distinct request automatically chooses certified compatible Route B within authorization/budget constraints; never replays a dispatched request |
| INV-38 | Multiple API keys, subscription-backed registrations, principals and connector types coexist in the registry and feed one common selector; no coded small fixed registration ceiling |
| INV-39 | Requested model identity, required capabilities, billing grant and data authority constrain automatic routing; different-model requests select different alias-specific pools |
| INV-40 | Explicit manual pin is respected and does not silently fail over; clearing it restores default automatic route selection atomically |
| INV-41 | A shared provider capacity owner is counted once even when multiple API keys are registered; capacity-exhaustion selection observes scoped cooldowns |
| INV-42 | Pool-size tests at 1, 2, 100 and 1,000 synthetic registered routes preserve routing semantics, resource bounds and reasonable indexed lookup |

## 14.3 Streaming fixtures (realistic cases)

- Every cut position of a valid SSE event; split `data:` across UTF-8 multibyte boundaries and chunk boundaries.
- Multiple events in one network read; CRLF/LF mixed; comments; blank lines; unknown extension events.
- Tool calls with `function.name` delivered once, late, partial or absent; `arguments` assembled across chunks; two calls interleaved; IDs reused/duplicated.
- Chat `finish_reason` followed or not followed by optional usage and terminal `[DONE]`; invalid duplicate completion.
- Responses `response.created`, deltas, reasoning-associated items, tool-call items, `response.completed`, `response.failed`, `response.incomplete` and truncation.
- Provider sends HTTP 200 with JSON error instead of SSE; 429 with Retry-After; 401 after expired credential; 503 before first byte.
- Headers delivered, then connection drops; stream begins then idle timeout; malicious endless event; gzip/unsupported compression; invalid UTF-8.
- Prompt size limits; response item over max size; schema mismatch; malformed JSON; correct cancellation and provider socket shutdown.
- Multimodal message/tool payload where backend lacks image support; structured-output constraints and parallel tool capability mismatch.
- No usage, delayed usage, conflicting usage, cancel before usage, cache usage inconsistent with model metadata.
- Native provider state token bound to route A and then replayed to B: must be rejected unless separately certified.

## 14.4 Routing and concurrency race tests

1. 100 concurrent requests contending for 1 route slot have no over-admission or starvation.
2. Automatic A→B eligibility change races with an in-flight stream; first retains A, subsequent distinct request selects B, never mixed. Test manual pin/unpin races separately.
3. Two simultaneous manual `select`/`auto` control calls have deterministic generation/conflict semantics; default automatic allocator uses committed route policy.
4. Revoke grant while stream is active: new requests denied; in-flight handling matches configured policy; ledger remains consistent.
5. Provider registration revoked during refresh: refresh cannot resurrect it.
6. Restart after `SENT_UNCOMMITTED`: no silent dispatch replay and usage marked uncertain.
7. Two keys with shared capacity owner: shared cooldown and concurrency are applied once, not counted twice.
8. Unsupported cross-model route hidden in the allowlist: eligibility excludes it even when all good routes are exhausted.
9. Explicit no-metered-fallback grant rejects API route even when subscription/local path unavailable.
10. Secret rotation while in-flight requests run: old attempt uses frozen safe credential snapshot; later use only valid new epoch.
11. Agent A project attempt cannot call for Agent B project model or usage.
12. Background compaction/unexpected utility-model request is intercepted or blocked.
13. Direct native provider egress blocked even when a hostile project config adds another provider.
14. Caller floods before authentication: bounded connections/body size and CPU work.
15. Repeated idempotency key with changed request hash is refused.
16. Register 100 independent credentials and 1,000 physical routes across multiple aliases and modes; the same selector still admits only eligible routes without a fixed configured ceiling.
17. Two keys for one alias: disable/exhaust A and verify the next **new** request chooses B with no button press or model-family substitution.
18. Different requested aliases route to disjoint certified pools; a Claude alias cannot silently use an unrelated Gemini/DeepSeek model.
19. `manual_pin` prevents unattended fallback to a different route; `auto` unpin re-enables automatic selection.
20. API-key and subscription-backed connector records coexist; only actually certified and authorized subscription routes are admitted.

## 14.5 Live-gated provider certification

A live route is `PROVEN` only after an approval/proof record includes:

```text
Provider / endpoint / authentication type / billing mode / principal scope
Connector certification evidence and adapter revision
Exact OpenCode binary hash, version, provider package + options hash
AgentPool service hash, adapter revision, alias capability manifest hash
Tested exact upstream model ID/snapshot/variant
Operating system, transport type, network/sandbox profile
Successful long multi-turn context and tool-use trace (sanitized)
Native compaction and provider-state continuity results
429, cancellation, partial stream, failed authorization and timeout results
Model usage / estimated versus reconciled cost observations
Observed client retry behavior and complete egress verification
Approval decision: PASS / FAIL / CONDITIONALLY_RESTRICTED
Dates, reviewers, reproducible test fixture/checksum
```

Provider API access, OpenCode published docs and a correct configuration file prove **possibility**, not end-to-end compatibility. Explicit `FAIL` for one route does not prevent a separately supported route from passing.

## 14.6 TradeAgent application gates

- Existing role grants, broker restrictions, risk engine and evidence acceptance never depend on trusting text produced by a provider.
- TradeAgent can observe and cancel a running OpenCode worker without calling AgentPool's owner control plane.
- TradeAgent separates logical job/attempt ID from OpenCode session ID and AgentPool inference attempt IDs.
- JEV/manager handoff creates a durable, validated checkpoint independent of the native session when continuation is unsafe.
- Tool side effects are reconciled before rerunning a failed worker; partial streamed commands do not become accepted work.
- Auth/billing/model switches are reflected in provenance and cost reports, never hidden from serious analysis or backtesting.
- A live-money armed mode cannot enable an unsandboxed CLI agent or bypass the app's broker/risk kernel.

## 14.7 Security and operations tests

- Loopback-only bind; IPv4/IPv6 and localhost edge cases; no LAN/unauthorized browser access.
- Peer-user identity/ACL for Unix control socket or Windows named pipe, including outsider-process tests.
- Secret-vault migration and revocation; redacted logs; no token in process command arguments.
- Hostile repository config/plugin tries to change provider or exfiltrate via direct network; fails independently of OpenCode policy.
- TLS certificate failure, DNS rebinding, proxy env injection, upstream redirect, URL path escape, SSRF.
- CPU/RAM/FD/socket exhaustion and rate limits; slowloris request; oversized event; client disconnect.
- SIGHUP/config reload, SIGTERM/service stop, sleep/resume, SQLite busy/corruption, disk full, power loss and backup/restore.
- Rolling upgrade/rollback of protocol version, model catalog, provider adapter and database migration.
- Windows service lifecycle and Mac/Linux launchd/systemd lifecycle independently verified before release on those platforms.

## 14.8 Proactive capacity, JEV and retry acceptance matrix (NEW)

**Core (blocking before v5.2 production readiness; RET cases must prove disabled-by-default safety, and any positive automatic OpenCode retry needs separate live certification):**

| Case | Test and mandatory observation |
|---|---|
| CAP-01 | If A has provider-evidenced insufficient request/token headroom and compatible B can serve the model, first attempt goes **directly to B**; no avoidable A 429 |
| CAP-02 | Simultaneous workers atomically reserve per shared capacity owner, even when 100 API keys point at the same real quota; no over-allocation created by local races |
| CAP-03 | RPM, input TPM, output TPM, daily/project and money limits are separately modeled as supported, with provenance/expiry and honest UNKNOWN |
| CAP-04 | Two different physical keys in one capacity owner cannot increase counted quota; Gemini-style same-project keys are tested as shared where applicable |
| CAP-05 | Normal selection makes zero quota-probe, authentication-refresh or JEV network round trips; no global stream drain |
| CAP-06 | One route enters COOLDOWN while other eligible routes and already streaming agents keep working; no cross-client lock or priority starvation |
| CAP-07 | At 1/2/100/1,000 routes and 1/20/100 concurrent workers, record p50/p95/p99 local overhead, queue delay, memory growth and starvation with fixed seed/replay |
| CAP-08 | Stale rate-limit snapshot, sleep/wake, inconsistent reset headers, clock skew and leaked reservations fail safely; expired/revoked credentials never reopen |
| CAP-09 | A large estimated request is avoided when capacity fits only a small one; same route may be READY for small, SATURATED for large |
| CAP-10 | Known exhausting account does not create a retry storm or secondary queue avalanche; all-route exhaustion emits actionable bounded failure/wait |
| RET-01 | Pinned OpenCode hook sees explicit early 429; if exact replay/tool/state/budget proof passes, one new client-owned attempt uses B; otherwise no automatic retry |
| RET-02 | Partial stream, ambiguous sent timeout, tool side effect, provider-bound response ID, billing hard-cap and unknown retry metadata cannot be silently replayed |
| RET-03 | Provider `Retry-After`, client max retry and cumulative grant budget remain enforced, including SDK-internal retries |

**Optional JEV promotion (blocking only if `calibrated_soft_penalty` is enabled):**

| Case | Test and mandatory observation |
|---|---|
| JEV-01 | At missing/advisory timeout/model 429/offline/malformed response, baseline deterministic routing remains correct with no added hot-path wait |
| JEV-02 | Prompt, context, customer source, raw provider error body, account name, credential and IP are absent from all advisor requests/traces |
| JEV-03 | Advisory can ONLY penalize the priority among **already eligible** routes; 1,000 adversarial high-confidence results cannot bypass authorization, cooldown, billing, alias, context or owner pin |
| JEV-04 | Every advisory contains model/rubric/schema revision, snapshot epoch, TTL, calibration band and source; stale/low-confidence/uncalibrated results have zero influence |
| JEV-05 | Shadow-mode held-out future events and a deterministic comparator support improved utility, Brier/log loss and calibrated false-positive/negative trade-offs; no nontrivial fairness, capacity, cost or latency regression |
| JEV-06 | Choice option reorder, ambiguous language, contradictory error messages and numeric/date traps tested against provider-described jaggedness; do not perform quota arithmetic with JEV |
| JEV-07 | Advisor disabling/hot rollback instantly restores deterministic ranking, does not cancel upstream streams, reauthorize credentials or require TradeAgent changes |
| JEV-08 | Advisor provider is authenticated with independent bounded credential/egress/finance policy, cannot recursively depend on the very routes it evaluates, and is not exposed to OpenCode prompts |
| JEV-09 | Cross-model/version/rubric changes invalidate learned advisory activation until shadow-mode recertification |

**Operator proof:** a single timeline trace must show (a) a known-capacity A excluded before dispatch, (b) B selected by deterministic code, (c) optional JEV observation produced out of band without request delay, and (d) a genuinely unexpected early 429 handled according to the independently proven OpenCode retry contract. A demo showing only account-switching after a visible failure does not satisfy CAP-01.


---

# 15. Practical build sequence — complete thin slice first

| Phase | Implement | Exit gate |
|---|---|---|
| **P0 — Freeze & measure** | Pin actual OpenCode V2 release/provider package, enumerate target API-key and account-based connector modes, identify current TradeAgent integration seam | Real config and captured HTTP trace; written connector capability profile |
| **P1 — Gateway proof** | One authenticated loopback endpoint, fake upstream, single exact protocol, real OpenCode TUI/headless fixture | Full multi-turn stream with tools; no bypass; no hidden fake success |
| **P2 — Minimal core** | Go service, **multi-registration schema and automatic selector interface**, vault, local IPC, one alias, one ordinary API key, request/stream state machine, budget grant | Security/stream fault tests pass; no upstream token in OpenCode; single-route case uses same future multi-route selector |
| **P3 — TradeAgent connection** | Separately migrate Codex→OpenCode within TradeAgent if needed; scoped per-worker gateway tokens, private OpenCode config, metadata correlation, strict egress and tool boundaries | Real TradeAgent job uses **the same native OpenCode host/session/tool interfaces** in both direct-provider and AgentPool-provider fixtures; no AgentPool agent-execution layer |
| **P4 — Automatic multi-route switching and proactive admission** | Extensible indexed route registry; two certified same-model routes; `auto_eligible`, per-owner multi-dimensional rate evidence/reservations/headroom, optional pin/override, immutable epochs and provider-state checks | **Without operator action**, known-insufficient A is excluded *before* new request dispatch and B is selected; opaque state rejects and unrelated streams continue |
| **P5 — New protocol/provider** | One additional adapter, explicit capabilities, full stream/tool/error/usage fixtures | That backend's route matrix earns `PROVEN`; no universal guarantee implied |
| **P6 — Production hardening** | Cross-platform vault/IPC/sandbox proofs, signed releases, recovery/soak, connector auditing, finances/usage reconciliation; measure multi-agent hot-path p50/p95/p99, quota-owner races and OpenCode retry behavior | All release blocking gates and CAP/RET tests pass on target OS; no hidden cost/latency from predictor |
| **P7 — Optional JEV risk advisor (separate promotion)** | Asynchronous TypeSafe/JEV typed risk observations on redacted provider-health summaries; shadow/calibrate/opt-in bounded soft penalty with deterministic fallback | JEV-01..09 and held-out calibration/utility benchmark pass; no mandatory hot-path inference, no changes to eligibility or agent ownership |

**Ordering decision:** P1 is more important than an allocator, but **P4 automatic multi-route selection AND proactive evidence-based capacity admission are core product release gates**; the optional JEV advisor must not be required to ship them. If OpenCode cannot run a realistic agent through a custom AgentPool provider with correct tools, events, reasoning and timeout behavior, further routing sophistication is wasted. Implement subscription-backed connectors as isolated provider adapters, with their own authentication/transport proofs, while using the already certified API-key adapter as a working baseline.

## 15.1 Definition of done for every PR

- Explicit contract and tests accompany behavior changes.
- At least one deterministic negative/edge-case test for every safety rule touched.
- Before/after model/protocol/credential authorization is fully traceable.
- Error behavior and financial uncertainty are tested, not assumed.
- Where capacity code changes, include pre-dispatch proof, shared-owner concurrent reservation, stale observation and zero-external-hot-path checks.
- Any advisor change preserves deterministic allow/deny behavior and carries a shadow/evaluation plan and redacted-output privacy proof.
- No unreviewed new dependency, network target, persistence of user prompt, or authority expansion.
- Compatibility manifest and operator documentation updated in the same PR.
- No claim of `PROVEN` without pinned real-client evidence.

## 15.2 Milestone deliverables

1. A standalone binary and service install script, plus uninstall/revoke procedure.
2. A versioned example OpenCode V2 configuration for each certified transport profile.
3. A test-only fake inference provider with deterministic SSE/tool/retry fixtures.
4. A private-IPC control schema and tiny owner CLI.
5. A complete typed route/alias/connector-certification manifest and usage ledger.
6. An integration smoke test that launches **real OpenCode** in both interactive and headless modes.
7. A documented, minimal TradeAgent worker provisioning snippet, **not** another agent supervisor inside AgentPool.
8. An authentication-mode and compatibility feature matrix showing what the **automatic selector** may choose for each alias, and why other routes are excluded.
9. A release runbook: pool registration, **automatic next-request failover** after route disabled/429, optional manual pin/unpin, broken stream, restart, stop, rollback.
10. A signed/dated release scorecard and unresolved issues register.
11. A proactive admission benchmark and capacity-owner race suite, with calibrated model-specific request sizing and no hot-path quota probes.
12. An optional isolated JEV advisor reference interface, disabled-by-default example, redacted fixture, and promotion/rollback report (P7 only).

---

# 16. Migration from the supplied architectures and Go starter

## 16.1 Preserve the validated engineering ideas

The v2/v3/v4 documents contributed useful hard requirements that remain binding:

- Fail-closed provider admission, explicit connector capability/credential states, and independently versioned admission rules.
- Identity distinction: model, credential principal, capacity owner, app, data authority, billing route.
- Request/stream commitment state machine, no ambiguous automatic replay, correct cancellation, and SSE fixtures.
- OS secret storage, local owner-only admin IPC, bearer-scoped inference and no raw prompts/secrets in routine logs.
- Safe next-request switch barrier/affinity, audited route generations and truthful capacity/cost accounting.
- SQLite durability, fenced grants, migration/restart reconciliation, bounded concurrency and timeouts.
- Version-pinned exact OpenCode API and source-backed release gates.
- TradeAgent as sole owner of institutional memory, JEV/manager handoffs, durable jobs and trading authority.

**What is removed or reassigned:** the previous v4 architecture's optional transport gateway, OpenCode worker bridge **inside AgentPool**, separate model-provider routing inside OpenCode, and any `AgentPool` obligation to create/restart/recover native OpenCode sessions. TradeAgent's own OpenCode control adapter remains a TradeAgent component.

## 16.2 Existing ZIP starter: reuse, do not overclaim

The supplied `AgentPool_OpenCode2_Manual_Switch_Source_2026-10-08.zip` contains a Go HTTP pass-through starter and deterministic gateway tests, plus an OpenCode V2 sample and a TypeScript TradeAgent bridge example. Its included status file says:

- It currently targets **only compatible OpenAI-format `/v1/chat/completions` and `/v1/responses`**, with all routes constrained to one fixed upstream endpoint.
- It has a manual API-profile selection barrier; account-based and subscription connectors are not yet implemented in that starter.
- Its OpenCode V2 wiring is an **example**, not a real pinned-client certification; Claude/Google/local models are delegated natively to OpenCode, which **does not meet the new mandatory AgentPool boundary**.
- It does not claim production-ready OS vaults, Windows IPC, full multi-client grants, quota/usage accounting or durable multi-provider capability validation.

Treat it as **seed code**, not as proof that the new v5 is already running. Refactor incrementally:

```text
Keep/refine: auth, loopback, strict endpoint/path, rate/resource bounds,
             relay, deterministic SSE tests, local owner control.

Change:       hardcoded one-upstream model/profile assumptions -> per-alias
             certified multi-registration route manifest; automatic
             future-request scoped selection.

Add:          explicit app/workload grants, real OS credential vault,
             privacy-safe usage ledger, connector capability records,
             strict OpenCode egress/profile proof.

Extend later: Anthropic/Google/local adapter certification and honest
             cross-provider protocol translation (not untested magic).

Remove:       implicit 'Claude/Google bypass AgentPool' from the managed
             OpenCode configuration and any AgentPool headless-agent control.
```

**Unverified status:** this document inspected the supplied source archive and its self-reported test inventory; it does **not** claim a complete live OpenCode/TradeAgent acceptance run or live cloud billing tests.

## 16.3 Zero-drift review questions

Every implementation review MUST be able to answer, with source and a test:

1. Where is the unique model inference boundary? **AgentPool.**
2. Who chose the job and owns its durable outcome? **TradeAgent.**
3. Who generates the next model request and executes tools? **OpenCode.**
4. Who has the upstream secret? **AgentPool only.**
5. Who may change logical model family? **User/TradeAgent through OpenCode**, with policy and provenance.
6. Who selects a permitted physical account/API project serving an alias? **AgentPool automatically for each new request**, within the alias/grant/continuity/budget constraints; an owner may optionally pin/override.
7. Can an in-progress stream switch? **No.**
8. If model/context state cannot transfer, who handles handoff? **OpenCode/TradeAgent**, not AgentPool.
9. How are subscription-backed and API-key routes integrated without conflating their authentication, capacity or billing semantics? **Separate certified connector modes.**
10. What is the evidence that this exact route works? **Pinned live conformance result**, not a provider marketing label.
11. Can known capacity exhaustion be avoided *before* dispatch? **Yes when supported by evidence, using local scoped reservations/headroom; otherwise conservatively UNKNOWN.**
12. Is JEV needed for correct routing? **No.** It is an optional asynchronous, calibrated soft risk advisor; arithmetic and authority stay in code.
13. Who decides if an unsuccessful model turn retries? **Only a certified OpenCode client-side model-turn retry policy, never an AgentPool agent/replay loop.**

---

# 17. Final release scorecard

No production release if a mandatory cell is UNKNOWN or FAIL. API-key, local, enterprise and subscription-backed operating modes each require their own authentication, protocol and operational proof matrix.

| Gate | Required evidence | Current document status |
|---|---|---|
| **A. Correct architecture** | Code/API inspection: no AgentPool job executor; mandatory inference provider | **DESIGN** |
| **B. Real OpenCode downstream contract** | Pinned V2 build and TUI/headless full tool/stream trace | **UNVERIFIED** |
| **C. Single-route API proof** | Real API-key route test + fake transport suite | **UNVERIFIED for v5** |
| **D. Correct retry and incomplete-stream semantics** | Fault injection, timeout/replay observation | **UNVERIFIED** |
| **E. Connector capability and billing** | Auth mode certification, owner grant, capacity accounting and spending controls | **Route-specific, UNVERIFIED** |
| **F. Credential boundary** | OS vault, scoped gateway grant, no secret leakage | **UNVERIFIED** |
| **G. No native provider bypass** | Effective policy + process/network egress proof | **UNVERIFIED** |
| **H. Model and context continuity** | Same route/certified A→B, compaction, opaque state tests | **UNVERIFIED** |
| **I. Durable crash/recovery** | Restart, ledger reconciliation and epoch fencing | **UNVERIFIED** |
| **J. TradeAgent safety** | Actual app integration + tool and broker isolation | **UNVERIFIED** |
| **K. OS/platform release** | Mac/Windows test matrix as applicable, installer/rollback | **UNVERIFIED** |
| **L. Adapter expansion** | Exact new provider/model capability matrix | **Deferred per provider** |
| **M. Provider-only host parity** | Direct-provider versus AgentPool comparative real OpenCode worker, session, tool and control fixture; no gateway agent API | **UNVERIFIED** |
| **N. Core automatic multi-route pool** | Independent A/B, pre-dispatch capacity avoidance on known ineligibility, correct alias/billing/state affinity, pin/unpin, 100/1,000 route scale fixture | **UNVERIFIED; required before claiming proactive pooled switching** |
| **O. Proactive multi-agent admission and safe retries** | CAP-01..10, RET-01..03, real OpenCode error-hook proof, fair contention, zero extra hot-path provider/JEV calls | **DESIGN / UNVERIFIED; core CAP tests mandatory, retries require separate certification** |
| **P. Optional JEV operational advisor** | JEV-01..09, TypeSafe API pinned, shadow performance vs deterministic baseline, privacy/rollback | **DESIGN / NOT TRAINED OR CERTIFIED; not core release blocker while disabled** |

**Release policy:** start with one fully functioning API route as a **transport proof**, then require **tested proactive admission and automatic A→B multi-route selection** before calling AgentPool's core pooled-switching product complete. Optional JEV activation is gated on measurable marginal value; it must never be the reason the basic gateway is unshippable. Publish only capabilities that have passed. An incomplete but honest v1 is superior to a featureful gateway that drops tools, loses reasoning, retries billable requests or misrepresents upstream authentication and billing behavior.

---

# 18. Sources and evidence hierarchy

**Research date:** 2026-10-08. Vendor documentation and protocol behavior can change; the team MUST snapshot/pin the exact API/SDK revision, binary version and relevant source commit for each acceptance run. Links here document interface behavior, not proof that any specific connector has been configured.

| ID | Primary source | Supports / constrains |
|---|---|---|
| **S1** | [OpenCode V2 — Providers](https://opencode.ai/v2/docs/providers) | Native V2 `providers`, custom gateway package, `settings.baseURL`, explicit models, native-compatible Responses package, model package override |
| **S2** | [OpenCode V2 — Migrate from V1](https://opencode.ai/v2/docs/migrate-v1) | V1/V2 schema and server/client changes; actual executable naming/version must be verified |
| **S3** | [OpenCode V1 — Providers](https://opencode.ai/docs/providers) | `provider`, `npm`, `options.baseURL`, Chat/Responses AI SDK distinctions |
| **S4** | [OpenCode V2 — Policies](https://opencode.ai/v2/docs/policies) | Effective provider restrictions, including `provider.use`; runtime policy is only one defense layer |
| **S5** | [OpenCode V2 — Embedded SDK](https://opencode.ai/v2/docs/build/sdk) | Host-to-agent API lives outside AgentPool; current implementation must be pinned |
| **S6** | [OpenCode V2 — Plugins](https://opencode.ai/v2/docs/build/plugins) | Model/session controls and plugin hooks; plugins are not a substitute for enforced gateway routing |
| **S7** | [OpenCode V2 — Config](https://opencode.ai/v2/docs/config) | Config precedence and runtime provider/model configuration |
| **S8** | [OpenCode V2 — API](https://opencode.ai/v2/docs/api) | Native headless host API: TradeAgent's interface, **not** AgentPool's inference API |
| **S9** | [OpenCode source repository](https://github.com/anomalyco/opencode) | Pin real provider package behavior and client retries to a revision |
| **S10** | [Google — Gemini function calling](https://ai.google.dev/gemini-api/docs/function-calling) | Provider-specific tool-stream behavior, not automatic OpenAI protocol equivalence |
| **S11** | [OpenCode V1/V2 migration warning](https://opencode.ai/v2/docs/migrate-v1) | Version drift/compatibility admission and separate SDK/client contracts |
| **S12** | [OpenAI API — Rate limits](https://developers.openai.com/api/docs/guides/rate-limits) | Request/token limit dimensions, response headers, `Retry-After`, cap behavior, SDK retry amplification and limits |
| **S13** | [Anthropic API — Rate limits](https://platform.claude.com/docs/en/api/rate-limits) | Organization/workspace scopes; request/input/output token dimensions, reset/remaining headers and spend cap |
| **S14** | [Gemini API — Rate limits](https://ai.google.dev/gemini-api/docs/rate-limits) | Per-project not per-API-key quotas, RPM/TPM/RPD and provider-dependent hard limits |
| **S15** | [TypeSafe AI — Introduction](https://docs.typesafe.ai/introduction) | Jev typed decisions, Choice/Score/Noul, parallel questions, code-controlled composition |
| **S16** | [TypeSafe AI — API reference](https://docs.typesafe.ai/api) | `systemone` request/response types and model versioning; not a native text-generating inference provider |
| **S17** | [TypeSafe AI — Jev 1.13 jaggedness](https://docs.typesafe.ai/model-jaggedness/jev-1.13) | Exact arithmetic/date and adversarial-state limitations; necessity of deterministic capacity calculations |
| **S18** | [TypeSafe AI — Confidence](https://docs.typesafe.ai/confidence) | Distribution confidence is not verified real-world failure probability; use domain thresholds and validation |
| **S19** | [OpenCode V2 — Plugins, retry policy](https://opencode.ai/v2/docs/build/plugins) | Retry hook exists, but attempt replay, commitment and tool safety need specific runtime conformance tests |
| **S20** | [TypeSafe AI — Official launch, 2026-09-15](https://typesafe.ai/blog/introducing-system-one-models-and-jev) | Published model positioning and vendor-reported timing; not proof of AgentPool risk-prediction quality |

**User-supplied architectural sources retained:** `AgentPool_Architecture_Paper(3).md`, `AgentPool_Architecture_Final(4).md`, `AgentPool_The_Switch_Architecture(3).md`, `AgentPool_The_Switch_Architecture_v2(3).md`, `AgentPool_The_Switch_Architecture_v3_Implementation_Spec(2).md`, `AgentPool_v4_OpenCode_Final_Implementation_Spec.md`, and `AgentPool_OpenCode2_Manual_Switch_Source_2026-10-08.zip`. Earlier documents are historical architectural evidence; **v5's corrected inference-only boundary takes precedence when a prior diagram/contract contradicts it**.

### Evidence grades

1. **Exact installed OpenCode/provider connector source and binary**, for runtime and credential behavior.
   **For JEV**, exact TypeSafe model/API revision, question rubric, privacy-safe state schema and measured calibration evidence are separately required.
2. **Real version-pinned integration and fault-injection traces**, for compatibility, refresh and continuity.
3. **Official upstream API/protocol documentation**, for target transport design.
4. **Measured provider capacity, account isolation and billing behavior**, for operational correctness.
5. **Unit tests with synthetic provider and fixtures**, for local invariant correctness.
6. **Earlier architectural narratives**, for product intent and constraints.

**Research interpretation (2026-10-08):** official docs prove interface options and rate-limit dimensions; they do not prove JEV can forecast API quotas, know all subscription balances, improve dispatch outcome, avoid provider-side billing ambiguity, or safely replay a failed model turn. Those are explicit **tests/experimental scope** here. New release versions must rerun source and on-device conformance checks.

A lower grade cannot override a higher grade. In particular: an elegant diagram does not prove an authentication connector works; correct fake SSE does not prove real opaque-context portability; and an existing source archive does not prove deployed production readiness.

---

# 19. Implementation authorization — the one-sentence decision

> **Build AgentPool as a compact, standalone, provider-neutral inference gateway with an extensible pool of independently registered API keys, account-based connections, subscription-backed connectors, enterprise routes and local models. To TradeAgent, an OpenCode worker using AgentPool remains an ordinary OpenCode worker managed through the same native OpenCode host interface as a worker using a direct provider. To OpenCode, `agentpool` is only a configured model provider with an exact, certified inference protocol. On every new inference request, AgentPool MUST automatically select a certified, authorized, compatible physical route for the logical model and capabilities OpenCode requested, accounting for affinity, **proactive evidence-based scoped capacity headroom and atomic reservations**, cooldown, cost and permissions. When a selected route is unavailable, a subsequent new request MUST be eligible for automatic re-selection without a human action; optional manual pinning must not be required for normal operation. Keep all agent execution and all TradeAgent task/memory/tool/trading orchestration above that boundary. Freeze one selected upstream account for each dispatched request, stream its result faithfully, never splice an ongoing response or silently swap model family/billing class, and fail visibly when no safe route exists. All AgentPool administration/provisioning/telemetry is separate infrastructure control, never an agent-runtime API. Predictable provider limits are a fast, pre-dispatch scheduling problem; unexpected safe-to-retry failures may be reattempted only via a separately certified **OpenCode-owned** model-turn recovery policy. An optional **asynchronous** JEV operational advisor may affect bounded soft route preference after calibration but never access, quota arithmetic, billing, compatibility, retries, in-flight streams, or TradeAgent agent/handoff control. The service MUST work correctly and quickly with JEV entirely absent.**

**Do not build:** an AgentPool agent scheduler, an AgentPool/OpenCode headless worker bridge, a second TradeAgent manager, provider-native secrets in OpenCode, an untested generic subscription-auth implementation, a universal fake OpenAI-compatibility translator, hidden model/billing changes, replay of ambiguous upstream requests, live-token stream splicing, a JEV API call per inference admission, a probabilistic authorization decision, or an imaginary universal subscription quota counter.

**First proof to demonstrate to the team:**

```text
TradeAgent job -> OpenCode (real TUI/headless agent)
                -> agentpool/<verified-model>
                -> AgentPool (scoped client grant, certified API-key route)
                -> upstream model API
                -> AgentPool SSE stream
                -> OpenCode consumes model answer and executes tool
                -> TradeAgent receives completed, audited work.
```

*End of AgentPool v5.2.0. This document preserves v5.1 automatic inference pooling and the provider-only boundary; it adds proactive deterministic capacity admission, certified model-turn recovery, and strictly optional JEV-powered route-risk advisory. Design is not proof of implementation.*
