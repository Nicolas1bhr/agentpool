# AgentPool — The Switch
## v5.0.1 · OpenCode-native inference gateway · Provider-only boundary clarification and implementation specification

> **TradeAgent owns the work. OpenCode executes the agents. AgentPool supplies the inference.**
>
> **One stable logical provider to OpenCode; multiple independently configured and certified physical inference routes behind AgentPool.**
>
> **AgentPool never owns an agent, its task, its session, its tools, its files, its memory, its handoff, or its trading permissions.**
>
> **To TradeAgent, OpenCode remains the agent runtime; to OpenCode, `agentpool` is only a configured model provider. AgentPool is never a bridge that runs OpenCode.**

| Document control | Value |
|---|---|
| Status | **Implementation specification — not a statement that the complete system has been built or certified** |
| Architecture version | **5.0.1** (clarification-only revision of v5.0.0), research baseline **2026-10-08** |
| Decision | AgentPool is the **mandatory inference data plane** for OpenCode installations managed by TradeAgent; it is **only an OpenCode model provider at the OpenCode interface**, not a headless-agent bridge |
| Clarification scope | Explicit provider-vs-runtime contract, host integration boundaries and acceptance gates; **no changed product scope, routing policy or connector promise** |
| Consumer | OpenCode CLI/TUI/headless/embedded runtime, under TradeAgent or directly under one authorized local owner |
| First runtime target | **Pin one specific OpenCode V2 release**; retain a separate verified V1 compatibility profile only if needed |
| Core implementation | Small local-first **Go service**, SQLite metadata, OS credential vault, HTTP inference transport and local owner-only control IPC |
| Extension | Certified protocol adapters and modular API-key, OAuth, subscription-backed, enterprise, and local credential connectors; no speculative universal compatibility |
| Supersedes | AgentPool Architecture Paper; Architecture Final; The Switch; v2; v3; v4 **on the location of the provider boundary and all resulting interfaces** |
| Existing artifact | `AgentPool_OpenCode2_Manual_Switch_Source_2026-10-08.zip` is a **narrow API-key pass-through prototype**, not complete v5 functionality |

**Normative language:** `MUST` = blocking requirement; `SHOULD` = default unless a documented test-backed exception; `MAY` = optional. `DESIGN` = not yet implemented; `DOCUMENTED` = a vendor-published interface; `PROVEN` = a passing, pinned conformance test; `UNKNOWN` is **never** treated as `PROVEN`. The requirement is to build the smallest system that preserves all MUST invariants—not to accumulate components.

---

# 0. Product definition and non-negotiable boundary

## 0.0 The implementation contract — read before writing code (NORMATIVE)

**The user-facing contract is deliberately simple:** in TradeAgent, an OpenCode worker configured with `agentpool/<alias>` MUST still behave like an ordinary OpenCode worker configured with an eligible direct provider/model. TradeAgent creates and controls **OpenCode**, not AgentPool. OpenCode carries out its normal agent loop, session/context management and permitted tool calls. **The only replacement inside OpenCode is its model-inference provider configuration/transport.** AgentPool receives individual model requests and returns compatible model results; it never runs an OpenCode worker or takes possession of an agent's work.

"As if OpenCode were running DeepSeek directly" describes **the unchanged host-to-agent control surface and agent lifecycle**, not a promise that DeepSeek, Claude, GPT or another upstream model has identical capabilities, answers, cost, context, tool semantics or provider-owned state. The selected logical model remains explicit (`agentpool/<model-alias>`), truthful and capability-checked. **OpenCode knows it selected a provider named `agentpool`; the physical credential/account behind that provider is abstracted, not the model's identity.**

### 0.0.1 Exactly four permitted communication relationships

| Caller → recipient | Interface / what passes | Ownership and prohibition |
|---|---|---|
| **TradeAgent → OpenCode** | Normal version-pinned OpenCode host API/CLI/SDK: create, prompt, observe, cancel, resume and manage workers | **TradeAgent controls the agent; OpenCode executes it.** This is the same interface TradeAgent would use if OpenCode talked to a provider directly. |
| **OpenCode → AgentPool** | Custom model-provider HTTP contract: model request, streaming result, protocol-level errors/cancellation | **AgentPool is a provider endpoint, not an agent API.** No job IDs, task queue, worker startup, memory handoff or tool execution API is introduced here. |
| **AgentPool → upstream model endpoint** | One selected, admitted, certified model request per dispatch attempt and its response | **AgentPool controls account/credential selection and inference transport only.** No calls to OpenCode's agent-control API. |
| **Trusted TradeAgent host / human owner ↔ AgentPool control IPC** | Separately authenticated **out-of-band** grants, credential registration, manual selection, route/usage status, policy and diagnostics | **Management plane only.** This does not carry agent prompts or act as an alternative host-to-agent runtime. An untrusted OpenCode agent may not invoke owner controls. |

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

**Operator switching means selecting an eligible *next inference request* route.** It is not replaying the ongoing model call, controlling an agent or migrating hidden provider state. A physical-route change can preserve an OpenCode session only after its exact same-effective-model, protocol, tool and state-continuity conditions pass; otherwise it fails explicitly and **TradeAgent/OpenCode**, not AgentPool, chooses a checkpoint or new-session recovery path.

### 0.0.3 Minimal observable acceptance criteria (all mandatory)

1. A version-pinned real OpenCode session works through `agentpool/<real-verified-alias>` with its ordinary agent control operations, streamed model output, multi-turn context and permitted tool calls. TradeAgent's worker lifecycle code does **not** call an AgentPool agent-execution interface.
2. The *same* OpenCode session/tool scenario is tested with a direct certified provider and with AgentPool using the **same physical model/feature contract**; differences in underlying billing, authentication and provider-state constraints are documented, not concealed.
3. In a managed worker's effective configuration and network trace, every model-generation call (including auxiliary models/subagents) reaches AgentPool or is explicitly disabled; no undeclared native provider escapes.
4. A real manual selection A→B changes only later **eligible** inference requests. The earlier in-flight route remains frozen; unsupported portability must reject rather than corrupt the OpenCode session.
5. AgentPool's local admin/grant/usage integration remains outside the model HTTP protocol and cannot be mistaken for an OpenCode host API; an agent prompt cannot elevate those controls.

**Reviewer decision rule:** if a proposed change requires AgentPool to know which task an agent is doing, run a worker, store chat history as agent memory, execute tools or implement a second job lifecycle, **reject it as outside v5**. The gateway may retain minimal opaque correlation and provider-bound state *references* required for authorized routing/accounting, but never owns or reconstructs TradeAgent/OpenCode agent state.

## 0.1 What it is

**AgentPool is a standalone inference service that OpenCode sees as one configured model provider named `agentpool`.** Every managed model request produced by an OpenCode agent crosses AgentPool. AgentPool authenticates the local caller, checks model, credential, billing and data-access controls, chooses an eligible physical upstream route, forwards or translates the inference request through a certified adapter, and streams the result back in the exact downstream protocol that OpenCode expects.

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
| Human/manual physical route selection | May present UI | Should not manage credentials | **Owns owner-only action** | Defines permissions |

**Explicit exclusion:** AgentPool does not launch, supervise or recover OpenCode processes, determine which agent should work next, schedule TradeAgent departments, call tools, handle broker credentials, own JEV, rewrite system prompts, replace memory, or manage a task queue. A diagnostic smoke-test command may invoke a **synthetic inference request**; that does not confer agent-execution responsibility.

## 0.4 Seven engineering invariants

1. **Only inference crosses AgentPool's model-provider data plane.** Separate authenticated control/telemetry operations exist for grants, routing, usage and credentials; they are **not agent-execution APIs**. No job state or tool-execution ownership migrates into the gateway.
2. **All managed inference crosses AgentPool.** No accidental native-provider fallback, including internal OpenCode model calls.
3. **One request, one upstream commitment.** No splitting a live stream across accounts or providers.
4. **No implicit model identity changes.** A stable alias may move only between **certified, authorized, semantically equivalent** physical routes. Model-family changes require a visible logical-model/session decision.
5. **No implicit billing changes.** A subscription or free local path must not quietly become a metered API path.
6. **Credential-mode isolation.** API keys, account-based integrations, OAuth/subscription connectors and local models have distinct credential, capacity, billing and failure semantics; never collapse them into one generic token type.
7. **Fail closed, surface truthful state.** Unknown compatibility, unclear financial exposure, ambiguous retry safety, revoked grants and partial streams never become silent successes.

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
| Manual backend selection | Choose a compatible physical route for future inference requests without transferring live streams | **DESIGN; integration proof required** |

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

**MVP-B (manual selection):** multiple **certified same-protocol routes** representing a verified same physical model/capability class; owner-initiated future-request selection, usage attribution and safe failure handling. Preserve one native OpenCode session during a boundary switch only if an end-to-end test proves it.

**MVP-C (provider breadth):** one additional certified backend adapter, then each next backend only on passing its specific authentication, tool/stream/reasoning and continuity test suite. Local models are good early candidates when they implement the required compatible API reliably. Include API-key and account-based connectors as independent extension targets.

**Route automation is deferred from the basic MVP.** First certify the invariant that a human-selected route change affects only subsequent requests. Keep the selector interface extensible for future capacity-aware routing strategies without restructuring the transport, credential vault or OpenCode integration.

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
| Multiple registered credentials/accounts | Native route registry | Verify independent custody, distinct principals, capacity-owner identity and operator selection |

**Design stance:** subscription-based, account-based, API-key and local inference are all legitimate *architectural connector categories*. They are not all technically identical, and no single generic connector is presumed to work for every provider. Provider-specific deployment authorization is assessed separately from the product's routing architecture; it does not remove connector slots or hard-code a provider-specific limitation into the core design.

The compatibility registry records `UNCONFIGURED | PROBING | CERTIFIED | DISABLED` for a specific `provider + credential_mode + endpoint + protocol_profile + app_identity`. Certification is a technical statement established by live tests, never a guess based on the presence of a credential. The owner/workload grant independently determines which certified routes a caller may use.

## 4.3 Subscription and account-based connector boundary

Account-based providers can expose browser authorization, OAuth, CLI-associated login, or other documented credential flows. An AgentPool connector owns its **own isolated credential registration and refresh lifecycle**; it must not assume an API key can be substituted for an OAuth token or that one provider's endpoint speaks another provider's protocol.

For each connector, implement only the applicable building blocks:

- Provider-appropriate interactive login and credential vault integration, where supported by its integration.
- Authenticated inference transport; token refresh, expiration, logout and revocation.
- Model catalog discovery and capability/cost/capacity observations.
- Multiple locally registered principals with separate route identity, quotas and audit history.
- Human-controlled selection of the next route, with active streams pinned to their original route.
- Explicit account/session affinity when a provider returns opaque response IDs, reasoning state or other nonportable context.

Keep authentication modules replaceable. **Absence of a finished subscription connector is an implementation backlog item, not an architectural prohibition.** Do not claim that a hypothetical connector already works until its end-to-end proof passes.

## 4.4 Credential lifecycle

- Register credentials only through the local owner interface. Do not accept raw provider secrets from OpenCode model requests or TradeAgent prompts.
- Store secrets in macOS Keychain / Windows Credential Manager or equivalent OS-backed secret store. SQLite contains an **opaque reference**, never the token.
- Keep provider access/refresh tokens separate from AgentPool's short-lived local client bearers; encrypt/protect at rest and keep logging redacted.
- Refresh only through a **documented provider flow**, single-flight per registration; bound refresh versions with `credential_epoch`, never overwrite a newer token with a stale response.
- Revoke/disconnect: stop new admissions, invalidate grants/routes and safely settle/abort in-flight attempts by policy. Never reauthorize automatically using an unrelated account.
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

# 5. Routing behavior: predictable now, extensible later

## 5.1 Three explicit routing modes

| Mode | Behavior | Admission |
|---|---|---|
| `fixed` | One chosen physical route until operator/app policy changes it | **Default v1** |
| `manual_next_request` | Owner explicitly selects a compatible certified route for future requests; current streams remain pinned | **MVP-B** |
| `adaptive_opt_in` | Future route selector considers health, observed capacity, cost and affinity at safe request boundaries | **Future phase; disabled in MVP** |

The route switch is a request-boundary operation. It validates owner control, model equivalence, workload grants and billing class; it cannot migrate an active stream, bypass caller permissions, or change a model family invisibly.

## 5.2 Selection algorithm

1. Validate client grant, alias/model, provider mode and exact request shape.
2. Build the eligible set from certified routes covered by the active workload grant.
3. Respect a session/attempt **affinity** where continuation is provider-state-bound or where cache locality strongly favors staying put.
4. Apply explicit pinned route/owner choice, where configured.
5. If a future opt-in selector is enabled, evaluate route health, capacity-owner cooldown, reserved concurrency, billing class and deterministic fairness; preserve explicit user selection and model-affinity requirements.
6. Freeze the chosen credential/model/billing identity with a unique `route_epoch` for the request.
7. Return one of `ADMITTED`, `REJECTED`, or `WAIT_FOR_USER/RESET` with a typed reason; do not silently change what the caller asked for.

**Pseudo-code (DESIGN):**

```go
func Select(req InferenceRequest, grant Grant) Decision {
    if !ValidGrant(grant, req.Client) { return Deny("GRANT_INVALID") }
    alias := Catalog.Lookup(req.Model)
    if !alias.IsCertifiedFor(req.Protocol, req.Capabilities) { return Deny("MODEL_UNCERTIFIED") }

    routes := Routes.Filter(func(r Route) bool {
        return Authorized(r, grant, alias, req) &&
               ProtocolProven(r, alias, req) &&
               BillingAdmitted(r, grant) &&
               ContextPortable(r, req) &&
               HealthyCapacity(r) && WithinBudget(r, req, grant)
    })
    if routes.Empty() { return Deny(ExplainNoRoute(req, grant, alias)) }

    chosen := StickyOrExplicitOrFair(routes, req.Affinity, grant.Mode)
    return ReserveOneAndFreeze(chosen, req.RequestID, grant)
}
```

This allocator selects **requests** only. It cannot read TradeAgent's role queue or choose which agent should execute a job.

## 5.3 Request boundaries versus stream boundaries

```text
Time ──────────────────────────────────────────────────────────>
OpenCode session S:  prompt ── request #1 ── tools ── request #2 ── tools ...
                              |                     |
AgentPool route:           API route A           API route B (after selection)
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

## 7.4 Reproducibility and JEV

JEV's future probabilistic, non-text-generating monitoring and manager-led handoff process belongs wholly to TradeAgent. An AgentPool rate-limit event may be **reported** to TradeAgent as infrastructure telemetry, but AgentPool does not instruct JEV, summarize an agent's mind, start replacement agents or decide strategic task continuation. TradeAgent's manager/handoff worker handles checkpoints if OpenCode cannot continue. Physical-model provenance must be available for later audit of AI-assisted research; do not mix different models invisibly in a scientific result or trading decision.

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
| Request retry count inside AgentPool | **0** | Raise only per separately certified unambiguous case |
| Unbounded provider monetary fallback | **Never** | Explicit grant required for any metered path |

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
- For future opt-in routing, use sticky preference and deterministic route selection at **new** request boundaries; when none are eligible return `NO_ELIGIBLE_ROUTE` with an actionable reason.
- Avoid a high-frequency credential/limit-probing loop; use official read-only quota/headers where available and owner-approved probes only when charged work is budgeted.
- Model local backend availability separately from remote provider quota; GPU/OOM cooldowns are not subscription limits.

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
- **Selection:** owner-authorized control action increments route-selection generation; new requests atomically observe that generation. In-flight attempts retain immutable old generation.
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
agentpool providers                              # registered adapters and certification
agentpool routes                                 # configured/certified routes and admission reasons
agentpool models                                 # logical catalog / compatibility matrix
agentpool credential add                         # owner-only, documented credential flow
agentpool credential revoke <registration-id>     # stop new work and invalidate grant
agentpool select <alias> <route-id>               # manual_next_request selection
agentpool pause <alias-or-route>                   # deny new admissions
agentpool resume <alias-or-route>                  # only if route still certified and enabled
agentpool grants list                             # masked scope and expiry
agentpool audit --since <cursor>                  # metadata-only events
```

These are **proposed commands**, not an assertion that the uploaded prototype already implements them. `select` MUST display old/new **provider, model, billing class, principal hint, scope, effect on current streams and admission decision**, and require interactive confirmation (or a separately authenticated owner control API call). Confirmation guards against accidental route or billing changes. Changing a non-equivalent model through `select` is **rejected**; use an explicit OpenCode/TradeAgent logical-model change instead.

## 11.2 Control IPC contract

Versioned owner/host API prefix: `ap.v1`. Management calls are not exposed on the inference HTTP port.

| Method | Authorization | Side effect |
|---|---|---|
| `hello` / `health` | Authenticated peer | Negotiate protocol, report capabilities |
| `credential.list/add/revoke` | Owner | Protected credential lifecycle |
| `provider.list/probe` | Owner (probe may cost) | Adapter availability/certification |
| `model.list` | Authorized host | Filtered aliases, limits, capability profiles |
| `route.inspect` | Authorized owner/host | Redacted eligible set/reason codes |
| `selection.preview/commit` | Owner and permission gate | Explicit future-request route choice |
| `grant.issue/renew/revoke` | Owner or delegated host | Scoped capability token and budget |
| `usage.read` | Authorized owner/host | Attempt and cost metadata for its scope |
| `events.read` | Authorized owner/host | Durable redacted events after cursor |
| `pause/resume/shutdown` | Owner | Route/service controls |
| `doctor` | Owner | Config/upgrade/security/release checks |

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
  default_route_mode: fixed
  inference_retry_count: 0
  allow_remote_client: false

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
    credential_vault_ref: REPLACE_WITH_OS_VAULT_REFERENCE
    principal_ref: REPLACE_WITH_VERIFIED_PROJECT_ID
    billing_class: metered_api
    capacity_owner_ref: REPLACE_WITH_VERIFIED_CAPACITY_OWNER
    enabled: true

routes:
  - id: primary-responses
    registration: primary-api-project
    upstream_model_id: REPLACE_WITH_REAL_SUPPORTED_MODEL_ID
    model_revision: REPLACE_WITH_VERIFIED_MODEL_REVISION
    protocol: responses
    proof_ref: REPLACE_WITH_PASSING_CERTIFICATION
    enabled: true

aliases:
  - id: responses-compatible
    downstream_protocol: responses
    allowed_routes: [primary-responses]
    routing_mode: fixed
    allowed_billing_classes: [metered_api]
    # Tool, context and output capabilities come from certified manifest.

clients:
  - id: personal-opencode
    allowed_aliases: [responses-compatible]
    billing_classes: [metered_api]
    max_concurrency: 1
    requires_explicit_budget: true
```

**Do not blindly use the above placeholders in a working deployment.** In particular, project IDs/capacity-owner scope must come from authoritative information; when unknown the corresponding cross-account routing mode cannot activate. Additional Anthropic/Google/local routes are separate adapter/registration/manifest records and may be added without moving any agent logic into AgentPool.

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
8. Run a permitted same-effective-model credential/route change between completed inference turns and test for continued context/tools without duplicate effects.
9. Run a stateful provider-response-ID/opaque-reasoning test and **require sticky route or explicit continuity failure** when state is nonportable.
10. Run strict egress monitoring and prove no managed model call leaks to another upstream provider endpoint.

Only when those pass should multi-provider normalization, future adaptive selection and complex quota observation be constructed. Mock tests alone cannot prove actual provider billing, live opaque-context portability, authentication flow reliability or production OS isolation.

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
2. Manual `select` races with an in-flight stream; first retains A, subsequent receives B, never mixed.
3. Two simultaneous `select` calls have deterministic generation/conflict semantics.
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

---

# 15. Practical build sequence — complete thin slice first

| Phase | Implement | Exit gate |
|---|---|---|
| **P0 — Freeze & measure** | Pin actual OpenCode V2 release/provider package, enumerate target API-key and account-based connector modes, identify current TradeAgent integration seam | Real config and captured HTTP trace; written connector capability profile |
| **P1 — Gateway proof** | One authenticated loopback endpoint, fake upstream, single exact protocol, real OpenCode TUI/headless fixture | Full multi-turn stream with tools; no bypass; no hidden fake success |
| **P2 — Minimal core** | Go service, schema, vault, local IPC, one alias, one ordinary API key, request/stream state machine, budget grant | Security/stream fault tests pass; no upstream token in OpenCode |
| **P3 — TradeAgent connection** | Separately migrate Codex→OpenCode within TradeAgent if needed; scoped per-worker gateway tokens, private OpenCode config, metadata correlation, strict egress and tool boundaries | Real TradeAgent job uses **the same native OpenCode host/session/tool interfaces** in both direct-provider and AgentPool-provider fixtures; no AgentPool agent-execution layer |
| **P4 — Manual route switching** | Same-model physical-route allowlist, explicit owner selection and immutable request-epoch rules, realistic provider state proof | Request A→B at safe boundary passes; opaque state rejects when necessary |
| **P5 — New protocol/provider** | One additional adapter, explicit capabilities, full stream/tool/error/usage fixtures | That backend's route matrix earns `PROVEN`; no universal guarantee implied |
| **P6 — Production hardening** | Cross-platform vault/IPC/sandbox proofs, signed releases, recovery/soak, connector auditing, finances/usage reconciliation | All release blocking gates pass on the target OS |
| **P7 — Optional adaptive selection** | Capacity-aware opt-in routing with independently certified connectors and operator-defined route constraints | Transparent continuity proof, account/credential isolation and complete fault tests |

**Ordering decision:** P1 is more important than an allocator. If OpenCode cannot run a realistic agent through a custom AgentPool provider with correct tools, events, reasoning and timeout behavior, further routing sophistication is wasted. Implement subscription-backed connectors as isolated provider adapters, with their own authentication/transport proofs, while using the already certified API-key adapter as a working baseline.

## 15.1 Definition of done for every PR

- Explicit contract and tests accompany behavior changes.
- At least one deterministic negative/edge-case test for every safety rule touched.
- Before/after model/protocol/credential authorization is fully traceable.
- Error behavior and financial uncertainty are tested, not assumed.
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
8. An authentication-mode and compatibility feature matrix showing what can actually be selected.
9. A release runbook: credential setup, route disabled/429, manual future-route selection, broken stream, restart, stop, rollback.
10. A signed/dated release scorecard and unresolved issues register.

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
             certified route manifest; future request scoped selection.

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
6. Who may select a permitted physical account/API project serving an alias? **AgentPool**, under operator/grant/mode constraints.
7. Can an in-progress stream switch? **No.**
8. If model/context state cannot transfer, who handles handoff? **OpenCode/TradeAgent**, not AgentPool.
9. How are subscription-backed and API-key routes integrated without conflating their authentication, capacity or billing semantics? **Separate certified connector modes.**
10. What is the evidence that this exact route works? **Pinned live conformance result**, not a provider marketing label.

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

**Release policy:** start with one fully functioning API route. Publish only capabilities that have passed. An incomplete but honest v1 is superior to a featureful gateway that drops tools, loses reasoning, retries billable requests or misrepresents upstream authentication and billing behavior.

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

**User-supplied architectural sources retained:** `AgentPool_Architecture_Paper(3).md`, `AgentPool_Architecture_Final(4).md`, `AgentPool_The_Switch_Architecture(3).md`, `AgentPool_The_Switch_Architecture_v2(3).md`, `AgentPool_The_Switch_Architecture_v3_Implementation_Spec(2).md`, `AgentPool_v4_OpenCode_Final_Implementation_Spec.md`, and `AgentPool_OpenCode2_Manual_Switch_Source_2026-10-08.zip`. Earlier documents are historical architectural evidence; **v5's corrected inference-only boundary takes precedence when a prior diagram/contract contradicts it**.

### Evidence grades

1. **Exact installed OpenCode/provider connector source and binary**, for runtime and credential behavior.
2. **Real version-pinned integration and fault-injection traces**, for compatibility, refresh and continuity.
3. **Official upstream API/protocol documentation**, for target transport design.
4. **Measured provider capacity, account isolation and billing behavior**, for operational correctness.
5. **Unit tests with synthetic provider and fixtures**, for local invariant correctness.
6. **Earlier architectural narratives**, for product intent and constraints.

A lower grade cannot override a higher grade. In particular: an elegant diagram does not prove an authentication connector works; correct fake SSE does not prove real opaque-context portability; and an existing source archive does not prove deployed production readiness.

---

# 19. Implementation authorization — the one-sentence decision

> **Build AgentPool as a compact, standalone, provider-neutral inference gateway. To TradeAgent, an OpenCode worker using AgentPool remains an ordinary OpenCode worker managed through the same native OpenCode host interface as a worker using a direct provider. To OpenCode, `agentpool` is only a configured model provider with an exact, certified inference protocol. Keep all agent execution and all TradeAgent task/memory/tool/trading orchestration above that boundary. Route each admitted model request through one certified API-key, account-based, subscription-backed, enterprise or local connector, stream its response faithfully, and change physical routes only at safe request boundaries. All AgentPool administration/provisioning/telemetry is separate infrastructure control, never an agent-runtime API.**

**Do not build:** an AgentPool agent scheduler, an AgentPool/OpenCode headless worker bridge, a second TradeAgent manager, provider-native secrets in OpenCode, an untested generic subscription-auth implementation, a universal fake OpenAI-compatibility translator, hidden model/billing changes, replay of ambiguous upstream requests, or live-token stream splicing.

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

*End of v5.0.1 clarification of v5. The document defines the product and its gates; implementation is complete only when the scorecard is PROVEN. No v6 scope or new agent runtime was introduced.*
