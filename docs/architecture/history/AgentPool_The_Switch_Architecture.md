# AgentPool — The Switch Architecture

> **A local subscription-aware continuity gateway for native AI agents.**
>
> **One agent. One logical session. One account at a time. Swap accounts only at safe model-request seams.**

**Status:** recommended architecture / implementation blueprint  
**Research cut:** 2026-10-07  
**Supersedes:** the heavier “durable agent execution broker” interpretation for the primary Codex path  
**Primary objective:** make a pool of independently authenticated, user-authorized AI subscriptions look like one continuous source of compute to software such as TradeAgent, T3 Code, or a human-launched coding agent—without reimplementing the coding agent itself.

---

# 0. Executive Verdict

The original AgentPool idea is right.

The original architecture is **too large for the actual job**.

The crucial simplification is this:

> **Do not rotate coding-agent processes when you can rotate the upstream subscription credential beneath a process that remains alive.**

For Codex, the clean architecture is therefore **not**:

```text
agent process A
    ↓ quota exhausted
kill / checkpoint / replace
    ↓
agent process B
    ↓
recover thread / workspace / tools / context
```

It is:

```text
same Codex process
same Codex thread
same tools
same workspace
same orchestration
same model request protocol
        │
        ▼
   AgentPool gateway
        │
        ├── account A  ← current
        ├── account B
        ├── account C
        └── account D

A approaches/exceeds usable capacity
        ↓
next safe inference request uses B
        ↓
Codex itself never changed
```

That distinction deletes an enormous amount of complexity.

AgentPool should **not** own the agent’s job lifecycle, workspace lifecycle, transcript, tool state, subagents, checkpoints, or task orchestration when the native agent already owns those things correctly.

AgentPool should own only the part the native agent cannot own:

> **Which authorized subscription account should serve the next model request?**

Everything else is secondary.

---

# 1. The Product in One Sentence

> **AgentPool is a local credential-routing gateway that exposes one stable model endpoint backed by a pool of independently authenticated subscription accounts, keeps requests sticky to an account while it is healthy, drains accounts before or at capacity, and switches the next safe request to another eligible account without restarting the calling agent.**

That is the product.

Not an AI framework.  
Not an agent orchestrator.  
Not a new coding agent.  
Not a distributed scheduler.  
Not a transcript store.  
Not a generic “LLM platform.”

It is a **switch**.

---

# 2. The Mental Model

Think of AgentPool like a fuel selector feeding an engine.

The coding agent is the engine.

The accounts are fuel tanks.

AgentPool does not rebuild the engine when one tank runs low. It selects another usable tank at a boundary where switching is safe.

```text
                        ┌──────────────────────────┐
                        │      CODING AGENT        │
                        │                          │
                        │ thread / context         │
                        │ tools / MCP              │
                        │ workspace                │
                        │ permissions              │
                        │ subagents                │
                        │ orchestration            │
                        └────────────┬─────────────┘
                                     │
                              model requests
                                     │
                                     ▼
                        ┌──────────────────────────┐
                        │        AGENTPOOL         │
                        │                          │
                        │  choose authorized acct │
                        │  observe capacity       │
                        │  sticky route           │
                        │  drain / cooldown       │
                        │  safe bootstrap retry   │
                        └────────────┬─────────────┘
                                     │
                   ┌─────────────────┼─────────────────┐
                   │                 │                 │
                   ▼                 ▼                 ▼
              Account A         Account B         Account C
              READY             READY             COOLDOWN
                   │                 │
                   └──────────┬──────┘
                              ▼
                       provider service
```

The **agent state remains above AgentPool**.

The **credential choice remains inside AgentPool**.

That boundary is what makes the system small.

---

# 3. Why This Is Better Than Process Rotation

The original design treated account exhaustion as an execution-failover problem.

For Codex, it can usually be treated as a **transport-routing problem** instead.

That changes almost everything.

| Problem | Process-rotation architecture | Switch architecture |
|---|---|---|
| Preserve Codex thread | must resume/reconstruct | thread never moved |
| Preserve tools/MCP | supervisor concern | native agent already owns it |
| Preserve workspace | leases/checkpoints needed | process keeps same workspace |
| Preserve permissions | translate/replay | native agent keeps them |
| Preserve subagents | potentially difficult | unchanged |
| Change account | start another runtime | change upstream credential |
| Avoid duplicate tool effects | complex turn replay logic | never replay after output escapes |
| Crash recovery | durable job machinery | client retries; gateway restores tiny routing state |
| Cross-account continuity | native-session compatibility problem | request routing problem |
| Runtime upgrades | broker must understand agent lifecycle | native agent can upgrade independently |
| T3 integration | ACP execution facade | point Codex transport at AgentPool |
| Core domain objects | many | four or five |

The result is not merely “less code.”

It removes entire **classes of correctness problems**.

---

# 4. The Architectural Rule That Keeps It Simple

> **Move the switching boundary as far downward as possible.**

Use this order:

```text
1. credential switch beneath a live agent      ← preferred
2. provider transport switch                   ← acceptable
3. runtime/account process switch              ← fallback only
4. semantic handoff to another agent/provider  ← last resort
```

AgentPool’s core should implement **#1** extremely well.

Do not burden the core with #3 and #4 merely because some future provider may require them.

Provider-specific fallback mechanisms can be added later behind adapters if they become necessary.

---

# 5. What Current Evidence Supports

## 5.1 OpenAI explicitly supports multiple saved ChatGPT registrations

OpenAI’s current Sign in with ChatGPT documentation describes applications managing multiple ChatGPT account registrations, keeping every registration’s credentials distinct, and switching among saved registrations.

This means the multi-account concept does not require inventing a credential-file swapping trick.

The correct abstraction is:

```text
Account A
  ├── validated registration identity
  ├── its own OAuth client registration
  ├── its own rotating refresh token
  └── its own access token

Account B
  ├── validated registration identity
  ├── its own OAuth client registration
  ├── its own rotating refresh token
  └── its own access token
```

Never combine credentials across registrations.

Token refresh is serialized per account because refresh tokens rotate.

## 5.2 Codex can use a custom model provider / base URL

Current Codex supports user-defined model providers with a configurable `base_url`, authentication settings, headers, and Responses wire protocol.

That is the seam AgentPool needs.

The Codex agent can continue owning its native session while its model traffic is sent to:

```text
http://127.0.0.1:<agentpool-port>/v1
```

instead of directly to the upstream service.

## 5.3 `codex-acp` already supports custom gateway authentication

The official/open ACP adapter around Codex App Server supports client-provided custom gateway configuration.

That makes a T3 path possible without teaching AgentPool how to become a full coding agent:

```text
T3 Code
   │ ACP
   ▼
codex-acp
   │
   ▼
Codex App Server
   │ Responses
   ▼
AgentPool
   │
   ▼
OpenAI subscription account pool
```

## 5.4 Open-source routers prove the routing pattern

Projects such as CLIProxyAPI already demonstrate several useful operational patterns:

- OAuth-backed multi-account routing;
- round-robin / weighted routing;
- session affinity;
- immediate failover away from unavailable credentials;
- quota-aware cooldown;
- “bootstrap retry” before any streamed response bytes escape;
- early draining near quota exhaustion.

AgentPool should learn from these patterns without copying their broader proxy scope.

## 5.5 Real bugs tell us what *not* to do

Current ecosystem failures are especially useful:

- T3 has encountered active-writer races while replacing Codex execution instances.
- T3’s shadow-home strategy has broken when Codex added new top-level state files.
- OpenCode has encountered duplicate concurrent LLM streams in one session.
- Proxy projects have misclassified quota as credential-wide even when limits were bucket/model-specific.
- cooldown implementations have been tripped by machine suspend/resume behavior.

The Switch Architecture avoids most of these failure modes by refusing to own agent state in the first place.

---

# 6. North-Star User Experience

A human should be able to do this:

```bash
agentpool login openai
agentpool login openai
agentpool login openai

agentpool run codex
```

and then forget which account is currently serving the agent.

TradeAgent should be able to launch a native Codex worker configured once:

```text
model provider = AgentPool
pool           = codex
```

and run normally.

T3 Code should be able to launch a local ACP provider backed by Codex and AgentPool:

```bash
agentpool acp codex
```

From all three callers’ perspectives:

```text
there is one usable Codex service
```

Internally:

```text
account A → account B → account C → account A after reset → ...
```

No client should need to know the rotation happened unless it asks for diagnostics.

---

# 7. Final Architecture

## 7.1 One binary, two tiny surfaces

Ship one executable:

```text
agentpool
```

It contains:

```text
┌─────────────────────────────────────────────────┐
│ AgentPool                                       │
│                                                 │
│  DATA PLANE                                     │
│  loopback HTTP                                  │
│  Responses-compatible upstream gateway          │
│                                                 │
│  CONTROL PLANE                                  │
│  local Unix socket / Windows named pipe         │
│  login / accounts / pools / status / doctor     │
│                                                 │
│  CORE                                           │
│  auth + routing + quota + tiny persistence      │
└─────────────────────────────────────────────────┘
```

No microservices.

No internal plugin bus.

No distributed coordinator.

No separate scheduler service.

No message broker.

No external database.

## 7.2 Data plane

Bind only to loopback by default:

```text
127.0.0.1:<port>
```

Expose the minimum provider-compatible surface needed by supported native agents.

For Codex v1:

```text
POST /v1/responses
```

Optionally expose model/catalog compatibility endpoints only when a client genuinely requires them.

The data plane does four things:

```text
authenticate local caller
identify requested pool/model/session hint
select one eligible upstream account
proxy the request + response stream
```

That is all.

## 7.3 Control plane

Use owner-only local IPC:

- Unix domain socket on macOS/Linux;
- named pipe with current-user ACL on Windows.

The control plane is for:

```text
login
logout
account list
pool configuration
status
quota observations
diagnostics
controlled shutdown
```

Do not expose administration over the ordinary model HTTP endpoint.

---

# 8. The Entire Core Domain Model

The primary path needs only these concepts.

## 8.1 Account

One independently authenticated subscription registration.

```yaml
Account:
  id: openai-01
  provider: openai
  principal_id: "provider-issued stable identity"
  auth_ref: "secret-store reference"
  enabled: true
  weight: 1
```

An account is never identified only by email.

## 8.2 Pool

A named set of accounts plus routing policy.

```yaml
Pool:
  id: codex
  provider: openai
  members:
    - openai-01
    - openai-02
    - openai-03
  routing: sticky-round-robin
  billing: subscription-included-only
```

## 8.3 Binding

An optional affinity from a logical caller/session to one current account.

```yaml
Binding:
  pool: codex
  key: "opaque-client-or-session-key"
  account: openai-02
  generation: 17
  last_used_at: "..."
```

Bindings improve cache locality and reduce unnecessary account hopping.

They are **not required for correctness**.

If a supported client cannot expose a stable session key, AgentPool can route each request independently.

## 8.4 CapacityObservation

A provider-derived fact about one account and one specific limit scope.

```yaml
CapacityObservation:
  account: openai-02
  limit_id: "optional provider bucket id"
  model_scope: "optional model scope"
  status: available | exhausted | unknown
  used_percent: 94.2
  reset_at: "2026-10-07T21:00:00Z"
  confidence: authoritative | strong | weak
  observed_at: "..."
```

Do **not** collapse all capacity into one boolean per account.

## 8.5 RequestAttempt

This can remain ephemeral in memory.

It represents one attempt to service one incoming model request with one account.

```text
request
  ├── attempt account A → quota rejection before output
  └── attempt account B → stream begins → committed
```

There is no durable `Job` object in the gateway core.

There is no durable `Attempt` state machine for the coding agent.

There is no AgentPool-owned `Turn` object.

Those belong to the native agent.

---

# 9. The Four Account States

Keep account admission brutally simple.

```text
READY
DRAINING
COOLDOWN
DISABLED
```

## READY

Eligible for new bindings and requests.

## DRAINING

Still healthy, but AgentPool intends to stop assigning fresh work because a relevant capacity bucket is near exhaustion.

Existing in-flight requests are allowed to finish.

A sticky binding may move on its **next** request.

## COOLDOWN

Authoritative capacity information says this account should not receive the relevant request until a known reset—or until a later successful probe/observation makes it eligible again.

Cooldown must be scoped to the relevant quota/model bucket where possible.

## DISABLED

Explicitly disabled by the user or blocked for authentication/policy reasons.

Do not invent fifteen health states.

Separate diagnostic fields can explain *why* an account is not usable.

---

# 10. The Safe Switching Boundary

This is the most important rule in the system.

> **AgentPool may transparently retry another account only while no model output has escaped to the caller.**

This gives an extremely clean safety boundary.

## 10.1 Safe

```text
client sends request
        ↓
AgentPool sends upstream using account A
        ↓
provider rejects A for relevant capacity/auth reason
        ↓
AgentPool has emitted zero response bytes/events to client
        ↓
AgentPool retries same inference using account B
```

For Codex, the agent has not yet seen a model response and therefore has not acted on it.

No client-side tool call has been triggered by a response AgentPool discarded.

## 10.2 Unsafe to hide

```text
account A begins valid response
        ↓
AgentPool forwards first semantic event to Codex
        ↓
stream later breaks
```

At this point AgentPool **does not replay the request invisibly**.

It passes the failure upstream to the native agent.

The native agent owns recovery from a partially observed model turn because only it understands its session semantics.

This one rule replaces pages of complex replay/handoff reasoning.

---

# 11. Bootstrap Retry

A high-quality gateway should distinguish:

```text
pre-output failure
```

from:

```text
mid-stream failure
```

For a Responses/SSE request:

1. accept the complete request body;
2. select account A;
3. open upstream request;
4. buffer only the response bootstrap;
5. classify an immediate structured capacity/auth rejection;
6. if retryable and zero downstream semantic output has escaped, select B;
7. once a valid response event is committed downstream, disable transparent failover for that request;
8. stream normally.

This is the same essential pattern used by mature routing proxies: **retry before first byte, not after**.

### Request-body replay

Most model requests are finite JSON bodies.

AgentPool should make retryability deterministic:

```text
small/normal body → retain in memory until request commits
large body        → spool to owner-only temporary file
```

Never depend on a one-shot caller socket if transparent bootstrap retry is enabled.

The temporary body copy is deleted immediately after request completion/failure.

No prompt is persisted in the database.

---

# 12. Routing Policy

The router should be deterministic enough to reason about and simple enough to test exhaustively.

## 12.1 Candidate filtering

An account is eligible when:

```text
enabled
AND auth usable
AND provider/account policy permits request
AND requested billing path is permitted
AND relevant model/bucket is not authoritatively exhausted
AND account is not globally quarantined for credential mismatch
```

Unknown capacity is **not** the same as unavailable.

## 12.2 Selection order

For a request with a binding key:

```text
1. current bound account if READY and eligible
2. another READY account
3. eligible account with UNKNOWN capacity
4. otherwise: no capacity
```

Among equivalent accounts:

```text
stable weighted round-robin
```

No scoring engine is needed.

No ML scheduler is needed.

No “optimizer” is needed.

## 12.3 Stickiness

Once a session/client is on an account, keep it there while the account is healthy.

Benefits:

- fewer auth switches;
- better upstream cache locality;
- easier debugging;
- less routing churn;
- predictable quota consumption.

Stickiness is an optimization, not a correctness requirement.

## 12.4 Binding generation

Every rebinding increments a monotonic generation:

```text
session S generation 41 → account A
session S generation 42 → account B
```

A stale concurrent task is never allowed to overwrite generation 42 with an older routing decision.

This is the one fencing primitive worth keeping from the larger architecture.

---

# 13. Aggressive Rotation Without Being Reckless

The user goal is not merely “recover after 429.”

It is to make exhaustion nearly invisible.

Do that with **preparation**, not speculative process replacement.

## 13.1 Prewarm

When an account approaches a configured threshold:

```text
refresh/validate next candidate's token
load current capacity snapshot if non-consuming
make sure it is admission-ready
```

Do **not** consume an inference merely to prove the account works unless explicitly running a deep diagnostic.

## 13.2 Drain

If a trustworthy relevant bucket crosses a configured drain threshold:

```text
account A: READY → DRAINING
```

Then:

```text
current in-flight request finishes on A
next request from bound session uses B
```

No current stream is cut merely to save the last few percent.

## 13.3 Hard exhaustion

If an authoritative provider response says A cannot serve the request:

```text
A relevant bucket → COOLDOWN
```

The request may bootstrap-retry B if nothing has been emitted downstream.

## 13.4 Thresholds are policy, not truth

Do not hard-code “95% always means exhausted.”

Recommended configuration:

```yaml
rotation:
  prewarm_remaining_percent: 10
  drain_remaining_percent: 5
```

but only apply percentage-based draining when:

- the provider exposes a trustworthy relevant usage bucket;
- the bucket applies to the requested model/plan path;
- `ordinaryUsageAllowed` or equivalent semantics do not contradict it.

Otherwise stay sticky until an authoritative denial or reset observation.

---

# 14. Capacity Model: Small but Correct

The dangerous simplification would be:

```text
account A = 92% used
```

That is often not enough.

Capacity can be scoped by:

```text
provider
application
account
workspace
model
limit bucket
rolling window
weekly window
```

So AgentPool stores observations keyed as narrowly as the provider exposes them.

Conceptually:

```text
(account, limit_id?, model_scope?) → observation
```

Routing asks:

> “Is there an authoritative observation that makes this account ineligible for **this request**?”

It does not ask:

> “Is this account globally dead?”

unless the provider actually says so.

---

# 15. Reset Times and Machine Sleep

Reset timestamps are wall-clock facts.

Persist them as absolute UTC instants:

```text
2026-10-07T21:00:00Z
```

Do not persist a monotonic “two hours from now” timer as the truth.

Why:

```text
machine sleeps for 3 hours
        ↓
provider reset passes while asleep
        ↓
AgentPool wakes
        ↓
absolute timestamp is already in the past
        ↓
account can be reconsidered immediately
```

A small jitter can prevent every waiting request from probing at exactly the same millisecond.

---

# 16. Billing Is a Hard Boundary

The default pool policy is:

```text
subscription-included-only
```

AgentPool must never preserve continuity by silently changing the economic path.

Forbidden unless explicitly configured:

```text
subscription → API key
subscription → metered credits
authenticated principal A → ambient principal B
personal subscription → cloud project billing
model X → cheaper/different model Y
```

Continuity means:

```text
same requested service semantics
through another explicitly authorized eligible account
```

not “anything that returns a response.”

---

# 17. OpenAI / Codex: The Primary Implementation

Codex should be AgentPool v1.

It is the cleanest fit for the actual product idea.

## 17.1 Authentication

Use OpenAI’s supported Sign in with ChatGPT flow for eligible open-source/local applications.

Each registration is independent.

Persist only metadata in AgentPool’s database:

```text
account id
provider subject / account identity
registration/client metadata
secret-store reference
token expiry metadata
last verified time
```

Raw refresh/access tokens belong in the platform secret store or an owner-only fallback specifically permitted by the supported flow.

### Refresh invariant

One refresh lock per account:

```text
account A refresh in progress
        ↓
all other A refreshers await same result
```

Never race rotating refresh tokens.

## 17.2 Upstream request

The data plane removes/ignores any client-supplied upstream credential and attaches exactly the credential selected by AgentPool.

```text
client auth         → proves permission to use local AgentPool
upstream account    → chosen by AgentPool
```

These are separate trust domains.

## 17.3 Codex provider configuration

Supported launchers configure Codex to use a custom provider whose base URL points to AgentPool.

Conceptually:

```toml
model_provider = "agentpool"

[model_providers.agentpool]
base_url = "http://127.0.0.1:PORT/v1"
wire_api = "responses"
requires_openai_auth = false
```

The exact configuration syntax should be generated from the currently installed Codex version rather than scattered across user dotfiles.

## 17.4 Session hint

AgentPool’s launcher may inject:

```text
X-AgentPool-Session: <opaque random id>
```

through supported provider headers.

This improves stickiness.

Correctness does not depend on it.

If Codex itself exposes a stable request/session identifier, AgentPool may use it **only after version-gated conformance**, not as an undocumented permanent contract.

## 17.5 Quota observation

Use structured account/rate-limit information when available.

Important behavior:

- multiple limit buckets may coexist;
- account association should be checked where exposed;
- reset times are observations;
- a null/unknown ordinary-usage flag must not be guessed into availability;
- successful dispatch remains authoritative.

## 17.6 Model access

A model catalog is not proof that a particular account can execute a request.

Do not preemptively rewrite model names.

The actual request is authoritative.

---

# 18. T3 Code Integration

T3 should not need an AgentPool fork.

Preferred shape:

```text
T3 Code
   │
   │ ACP
   ▼
codex-acp
   │
   │ native Codex app-server interaction
   ▼
Codex
   │
   │ Responses provider configured to AgentPool
   ▼
AgentPool
   │
   ▼
subscription account pool
```

The current Codex ACP adapter already has a custom-gateway mechanism, which is exactly the seam needed.

## 18.1 `agentpool acp codex`

This command should be a **thin launcher**, not a second implementation of ACP.

Its job:

```text
resolve AgentPool local endpoint
create local gateway auth token/session hint
construct current codex-acp gateway configuration
exec codex-acp
forward stdio unchanged
```

Where practical, use the upstream adapter instead of vendoring its protocol logic.

## 18.2 What happens during account rotation

Nothing happens at T3’s ACP layer.

Nothing happens to the Codex thread.

Only the next Responses request is authenticated upstream by a different AgentPool account.

That is the desired invisibility.

---

# 19. TradeAgent Integration

TradeAgent should treat AgentPool as infrastructure, not as another manager in its hierarchy.

TradeAgent owns:

```text
agent hierarchy
task decomposition
JEV/context-cleansing logic
worktrees/workspaces
verification
retries at semantic task level
tool policy
checkpoints
long-running orchestration
```

AgentPool owns:

```text
subscription account selection
quota/cooldown observations
auth refresh
safe pre-output retry
local transport security
```

Recommended setup:

```text
TradeAgent worker
   │
   ▼
Codex native runtime
   │
   ▼
AgentPool pool: codex
```

TradeAgent can ask AgentPool for diagnostics, but does not need to participate in account rotation.

### Critical consequence

AgentPool does **not** need to understand TradeAgent’s units, managers, JEV events, repository state, or handoff format.

That boundary should remain permanent.

---

# 20. Human CLI Integration

`agentpool run codex` is only a convenience launcher.

It should:

1. ensure daemon is running;
2. create an ephemeral local access token;
3. generate a session hint;
4. invoke the installed Codex runtime with a temporary/in-memory AgentPool provider override;
5. leave the user’s normal Codex configuration untouched where possible;
6. exit with the native Codex exit status.

AgentPool does not wrap every Codex terminal byte or emulate the TUI.

It launches the real thing.

---

# 21. Claude: Do Not Force the Codex Architecture Onto It

The current Claude subscription path is different.

Claude Code supports separate account/config roots, and current subscription-backed machine usage exists through supported Claude tooling. But current gateway/base-URL and OAuth behavior has enough provider/version-specific edge cases that AgentPool should **not pretend the Codex data-plane trick is universally portable**.

Therefore:

```text
v1: Codex switch path
v2: Claude adapter after dedicated conformance research
```

## 21.1 Claude strategy

Try the simplest supported transport seam first.

If current Claude runtime can be proven to route subscription-backed inference through a safe custom endpoint without corrupting OAuth/billing semantics:

```text
use gateway mode
```

If not:

```text
use a Claude-specific runtime/profile adapter
```

That fallback remains outside the core router.

## 21.2 No universal session assumption

Separate `CLAUDE_CONFIG_DIR` roots isolate credentials and local session history.

Do not claim native cross-account resume unless the exact runtime proves it.

If account replacement requires a native-agent handoff, that is a Claude adapter concern—not a reason to turn the entire AgentPool core back into an execution broker.

---

# 22. Google Consumer Antigravity: Disabled by Policy

The previous architecture should be corrected here.

Google’s current Antigravity FAQ states that using third-party software/tools/services to access Antigravity violates its Terms of Service and can lead to suspension or termination. Google directs users who want third-party coding agents with Gemini toward permitted Enterprise/API paths.

Therefore:

```text
consumer Antigravity subscription rotation = NOT SUPPORTED
```

until Google explicitly provides a permitted integration path.

AgentPool must never turn “technically possible” into “supported.”

Future Google support can be added through:

```text
explicitly permitted API path
Gemini Enterprise path
or a future officially authorized third-party subscription mechanism
```

without changing the core architecture.

---

# 23. Tiny Persistence Model

SQLite is still useful, but the schema becomes small.

Recommended tables:

```text
accounts
pools
pool_members
bindings
capacity_observations
schema_migrations
```

Optional:

```text
audit_events
```

That is enough for the core product.

## 23.1 What is deliberately absent

No:

```text
jobs
attempts
turns
native_sessions
workspace_leases
execution_leases
checkpoints
handoff_envelopes
runtime_leases
full event-sourced agent history
```

Those belonged to a system that owned agent execution.

This system does not.

## 23.2 SQLite settings

For a local single-writer daemon:

```text
journal_mode = WAL
foreign_keys = ON
busy_timeout = bounded
synchronous = FULL
```

The data volume is tiny, so durability is more valuable than squeezing maximum write throughput.

## 23.3 Bounded observations

Quota/history data is operational telemetry, not a forever log.

Retain a bounded recent history or periodically compact it.

Do not recreate the “gigabytes of event table” failure mode seen in long-running agent tooling.

---

# 24. What Must Be Durable

AgentPool restart should preserve:

```text
configured accounts
pool membership
secret references
known cooldown/reset instants
sticky bindings if useful
schema/config version
```

It does **not** need to preserve an agent turn.

If AgentPool crashes mid-request:

```text
TCP/SSE stream dies
        ↓
native caller sees transport failure
        ↓
AgentPool user service restarts
        ↓
caller uses its own normal retry/session recovery
```

This is a deliberate boundary.

Trying to hide a gateway crash after partial output would reintroduce unsafe replay semantics.

---

# 25. Concurrency Model

Keep concurrency rules explicit.

## 25.1 One daemon

Acquire an owner-only singleton lock.

One AgentPool daemon owns the local account database and refresh state.

## 25.2 Many requests

Different requests may use the same account concurrently if provider policy permits it.

AgentPool is not required to serialize all inference per account.

## 25.3 Serialized credential mutation

Per account:

```text
login/logout/refresh/registration mutation = exclusive
```

## 25.4 Binding update fencing

Binding updates use compare-and-swap semantics:

```text
UPDATE bindings
SET account_id=?, generation=generation+1
WHERE key=? AND generation=?
```

A stale router decision cannot resurrect an old account binding.

## 25.5 No duplicate upstream owner

One request-attempt object owns one upstream stream at a time.

When bootstrap failover occurs:

```text
cancel/close old upstream attempt completely
then start next attempt
```

Never leave two candidate streams racing to become authoritative.

---

# 26. Security Model

AgentPool has access to valuable subscription credentials. Treat that as the highest-risk part of the product.

## 26.1 Secrets

Preferred:

```text
macOS Keychain
Windows Credential Manager / protected store
Linux Secret Service
```

Fallback only when necessary:

```text
owner-only encrypted/protected local secret file
```

SQLite stores opaque references, not raw refresh tokens.

## 26.2 Local data-plane authentication

Do not assume every process on localhost should be allowed to consume the user’s subscriptions.

Supported launchers receive a high-entropy ephemeral/local bearer secret.

Example:

```text
Authorization: Bearer <local-agentpool-token>
```

The gateway strips this before sending upstream.

## 26.3 Environment hygiene

AgentPool launchers construct the environment deliberately.

Do not inherit unrelated provider secrets such as ambient API keys when those could silently change billing paths.

## 26.4 Loopback only

Remote exposure is **not** an MVP feature.

No `0.0.0.0` listener by default.

Remote/team use is a different security product and should not leak into the local switch.

## 26.5 No prompt logging by default

Operational logs may include:

```text
request id
account id
model
latency
status
quota classification
bytes
```

but not prompt/response bodies by default.

---

# 27. Failure Classification

The router needs only enough classification to decide:

```text
retry another account now?
wait/cooldown this account?
pass the error to caller?
```

A compact result is enough:

```yaml
Failure:
  class: capacity | auth | provider | network | request | unknown
  scope: account | model_bucket | provider | request | unknown
  retry_pre_output: true | false
  cooldown_until: optional
  confidence: authoritative | strong | weak
```

## Rules

### Capacity

Authoritative account/model capacity denial before output:

```text
mark relevant scope unavailable
bootstrap retry another account
```

### Auth

Expired access token with valid refresh path:

```text
refresh once under account lock
retry same account before output
```

Permanently invalid registration:

```text
disable/quarantine account
try another account if safe
```

### Provider overload

Do not rotate every account as if their subscriptions are exhausted.

Apply provider-level bounded backoff and surface failure according to retry policy.

### Unknown 429

Do not automatically classify every 429 as subscription exhaustion.

Use structured error semantics where available.

### Mid-stream failure

No transparent cross-account replay.

Pass it to caller.

---

# 28. “No Capacity” Behavior

If every eligible account is unavailable:

```text
AgentPool does not invent compute
```

Return a structured local error containing only safe operational data:

```yaml
error: agentpool_no_capacity
pool: codex
next_known_reset: "2026-10-07T21:00:00Z"
retryable: true
```

The native agent/client decides whether to wait/retry.

Optionally, the AgentPool launcher can provide a convenience wait mode for humans:

```bash
agentpool run codex --wait-for-capacity
```

but the core data plane should not hold arbitrary requests open for hours.

---

# 29. Admin CLI

Keep the command surface understandable from memory.

```text
agentpool start
agentpool stop
agentpool status

agentpool login openai
agentpool accounts
agentpool disable <account>
agentpool enable <account>
agentpool logout <account>

agentpool pools
agentpool pool show codex

agentpool run codex
agentpool acp codex

agentpool doctor
agentpool doctor --deep
```

Optional debugging:

```text
agentpool route explain --pool codex --model <model>
agentpool quota
agentpool logs
```

Do not expose every internal field as a CLI subcommand.

---

# 30. Configuration

Prefer one human-readable config file plus the SQLite runtime state.

Example:

```yaml
version: 1

server:
  listen: 127.0.0.1:0

pools:
  codex:
    provider: openai
    billing: subscription-included-only
    routing: sticky-round-robin
    accounts:
      - openai-01
      - openai-02
      - openai-03

rotation:
  bootstrap_retries: 2
  prewarm_remaining_percent: 10
  drain_remaining_percent: 5

security:
  require_local_bearer: true
  log_bodies: false
```

Rules:

- reject unknown dangerous configuration rather than silently guessing;
- validate account IDs and pool membership at startup;
- percentages may be disabled per provider if telemetry is not trustworthy;
- configuration changes are atomic;
- never store secrets inline in the YAML.

---

# 31. Implementation Language

The architecture no longer needs the complexity that originally made a large Rust workspace attractive.

For this **gateway-first** design, the recommended implementation is:

> **one Go module, one binary.**

Why Go fits this specific problem:

- excellent standard-library HTTP server/client primitives;
- straightforward streaming and cancellation with `context.Context`;
- simple single-binary distribution;
- low ceremony for a small long-running daemon;
- mature OAuth/HTTP ecosystem;
- easy bounded worker/concurrency patterns;
- very good fit for reverse-proxy/gateway code;
- current open-source multi-account proxy projects prove the workload is natural in Go.

This is not a claim that Go is “better than Rust.”

It is a claim that **this smaller architecture does not need Rust’s additional systems complexity to be robust**.

If the implementation team is materially stronger in Rust, Rust remains an excellent choice. The architecture should not depend on language-specific tricks.

---

# 32. Repository Shape

Do not start with eleven crates/packages representing speculative abstractions.

A deliberately boring repository is better:

```text
agentpool/
├── cmd/
│   └── agentpool/
│       └── main.go
├── internal/
│   ├── account/
│   ├── auth/
│   │   └── openai/
│   ├── gateway/
│   │   └── responses/
│   ├── routing/
│   ├── quota/
│   ├── store/
│   ├── control/
│   ├── launcher/
│   └── diagnostics/
├── migrations/
├── fixtures/
├── testdata/
├── docs/
├── go.mod
└── README.md
```

There is no plugin ABI in v1.

Providers are built-in packages until there is real evidence that dynamic plugins are needed.

---

# 33. Provider Adapter Contract

Even with a small core, provider-specific semantics should not leak everywhere.

A minimal internal provider interface is enough:

```go
type Provider interface {
    Authenticate(...)
    Refresh(...)
    PrepareUpstreamRequest(...)
    ClassifyResponse(...)
    ReadCapacity(...)
    VerifyPrincipal(...)
}
```

The gateway owns transport lifecycle.

The provider adapter owns only facts that differ by provider.

Do not add methods for sessions, tools, workspaces, permissions, checkpoints, or subagents to this interface.

If a future provider requires runtime-level switching, implement a **separate adapter mode**, not ten speculative hooks in this contract.

---

# 34. Compatibility and Upgrades

Do not build a giant runtime manager in v1.

The core compatibility problem is much smaller:

```text
Can the installed/native client still speak the model transport AgentPool expects?
Can the provider auth/quota semantics still be classified correctly?
```

## 34.1 Conformance probes

At startup/version change, non-destructive tests may verify:

```text
AgentPool gateway reachable
local bearer accepted
OpenAI account identity obtainable
capacity endpoint/schema parseable where available
Codex custom-provider config accepted
one synthetic protocol fixture still parses
```

## 34.2 Deep test

`agentpool doctor --deep` may consume a tiny model request after explicitly warning the user.

It validates end-to-end:

```text
Codex/custom client → AgentPool → selected account → provider → stream
```

## 34.3 Version drift

Keep quirks in one compatibility table/data structure.

Do not scatter:

```text
if version == x
```

through routing code.

---

# 35. Observability

The default status view should answer the questions the user actually cares about.

Example:

```text
AgentPool  healthy
Pool       codex

ACCOUNT      STATE       RELEVANT USAGE     RESET        ACTIVE
openai-01    DRAINING    96%                21:00        2
openai-02    READY       42%                21:00        5
openai-03    READY       unknown            —            1
openai-04    COOLDOWN    exhausted          20:14        0

Requests     18 active
Rotations    7 today
Failures     0 unsafe retries
```

Operational metrics worth exposing:

```text
requests_total
requests_active
bootstrap_retries_total
account_rotations_total
capacity_rejections_total
auth_refresh_total
midstream_failures_total
routing_latency
upstream_latency
```

The most important invariant metric is:

```text
post-output-transparent-retries_total = 0
```

forever.

---

# 36. Testing Strategy

The smaller system should be tested harder, not more casually.

## 36.1 Deterministic fake upstream

Build a fake Responses server that can script:

```text
success
429 before headers
structured quota error before first SSE event
auth failure
reset timestamp
provider overload
network close before output
network close after one event
slow stream
malformed event
unknown future fields
```

## 36.2 Router property tests

Properties:

```text
disabled accounts are never selected
relevant cooldown accounts are never selected
unrelated model-bucket cooldown does not globally disable account
healthy sticky binding remains sticky
rebind increments generation
stale generation cannot overwrite newer binding
selection terminates even when all accounts fail
no account is retried forever
```

## 36.3 Stream safety tests

The critical suite:

```text
failure before downstream output → retry permitted
failure after downstream output  → retry forbidden
first account quota reject        → second account receives identical request body
old upstream stream is closed     → before next becomes authoritative
```

## 36.4 Auth tests

```text
separate registration credentials never cross
refresh is single-flight per account
rotating refresh token is committed atomically
principal mismatch disables account
ambient API key cannot override subscription path
```

## 36.5 Sleep/reset test

```text
persist reset_at
simulate wall clock passing while process stopped/asleep
restart
account becomes reconsiderable immediately
```

## 36.6 Real two-account Codex test

This is the MVP proof:

```text
1. authenticate account A and B
2. launch one Codex instance through AgentPool
3. complete several turns through A
4. force/simulate A becoming unavailable before a new inference response begins
5. next request is transparently serviced through B
6. same Codex process remains alive
7. same Codex thread continues
8. tools/workspace remain unchanged
9. verify selected upstream principal is B
10. verify no duplicate downstream model stream occurred
```

That test matters more than a thousand architecture diagrams.

---

# 37. MVP Definition

The MVP is now dramatically smaller and much closer to the actual vision.

## 37.1 MVP proof

> **One live Codex session can run through AgentPool using at least two independently authenticated ChatGPT accounts; when the current account becomes unusable before a new model response has been exposed, the same Codex process continues the next request through another eligible account without the user or host rebuilding the agent session.**

## 37.2 Required

### Core

- one daemon / one binary;
- loopback Responses-compatible data plane;
- owner-only local control plane;
- SQLite tiny state store;
- account/pool/binding/capacity model;
- sticky weighted round-robin;
- DRAINING / COOLDOWN behavior;
- safe bootstrap retry;
- per-account serialized token refresh;
- binding generation fencing;
- bounded operational logging;
- automatic user-service restart where supported.

### OpenAI

- supported Sign in with ChatGPT account registration;
- multiple independent saved registrations;
- principal verification;
- protected secret storage;
- structured capacity observations where available;
- relevant bucket/model scoping;
- subscription-included-only policy.

### Codex

- generated custom-provider configuration;
- `agentpool run codex`;
- compatibility test against current Codex transport;
- real two-account continuity test.

### T3

- thin `agentpool acp codex` launcher or documented codex-acp custom-gateway configuration;
- no T3 fork.

## 37.3 Explicitly not MVP

- Claude rotation;
- Google consumer Antigravity;
- cross-provider handoff;
- agent job database;
- workspace leasing;
- prompt/transcript persistence;
- generic ACP server implementation;
- runtime package manager;
- distributed scheduling;
- remote networking;
- GUI;
- plugin ABI;
- semantic checkpointing;
- automatic Git worktrees;
- cost optimizer;
- speculative AI scheduler.

---

# 38. Build Order

Build in the order that proves the core hypothesis fastest without sacrificing correctness.

```text
1. fake Responses upstream + streaming test harness
2. local gateway with zero account logic
3. account/pool store
4. deterministic router
5. bootstrap retry boundary
6. quota/cooldown/drain model
7. OpenAI supported authentication
8. secret store + serialized refresh
9. real upstream Responses request
10. Codex custom-provider launcher
11. real two-account continuity test
12. daemon restart + tiny state recovery
13. T3 / codex-acp launcher
14. doctor / diagnostics
15. hardening + packaging
```

Do not start with Claude.

Do not start with a GUI.

Do not start with a provider plugin system.

Do not start with a durable agent job scheduler.

First prove that **the same living Codex can switch subscription accounts underneath itself**.

---

# 39. What Was Deleted From the Previous Architecture

This is the most important simplification audit.

| Previous concept | New decision | Why |
|---|---|---|
| AgentPool durable `Job` | delete from core | native agent/host owns the job |
| `Attempt` per agent assignment | delete from core | only request attempts matter |
| AgentPool `Turn` state | delete | native agent owns turns |
| `NativeSession` database | delete | native agent owns session |
| `SessionDomain` abstraction | delete from gateway core | no native session migration in primary path |
| workspace leases | delete | TradeAgent/native client owns workspace concurrency |
| process-tree supervision | delete from core | launcher is not execution owner |
| provider runtime manager | defer | native runtime can be externally managed |
| durable event journal | replace with bounded audit/metrics | gateway has little semantic state |
| cross-provider handoff envelope | defer | not needed to switch Codex accounts |
| continuity planner | replace with one rule | retry only before downstream output |
| seven-level continuity classes | delete | transport seam makes them unnecessary |
| agent permission normalization | delete | native agent owns permissions |
| tool lifecycle tracking | delete | native agent owns tools |
| checkpoint orchestration | delete | host owns checkpoints |
| ACP compatibility plane | replace with thin launcher | use existing codex-acp |
| generic inbound ACP driver | defer | no evidence core needs it |
| managed runtime side-by-side upgrade | defer | not required for switch MVP |
| execution fencing | reduce | only binding generation / upstream owner fencing remains |
| large provider capability matrix | reduce | only transport/auth/quota capabilities matter |
| process crash reconciliation | delete from core | caller observes transport failure and retries |

This is the difference between building a small piece of infrastructure and accidentally building a second agent platform.

---

# 40. What Was *Not* Deleted

Simplification must not become fragility.

These survive because they directly protect the switch:

```text
separate account identities
protected credentials
serialized refresh
principal verification
subscription-only billing guard
model/bucket-aware capacity
absolute reset timestamps
sticky routing
prewarm/drain behavior
safe pre-output retry
one authoritative upstream stream
binding generation fencing
bounded persistence
provider-specific error classification
conformance tests
diagnostics
```

This is the Toyota principle:

> fewer moving parts, but the remaining moving parts are over-engineered where failure would be expensive.

---

# 41. Non-Negotiable Invariants

These should be executable tests.

## I1 — One account registration, one credential set

Credentials from different registrations are never mixed.

## I2 — The requested billing path does not silently change

Subscription-included-only means subscription-included-only.

## I3 — AgentPool never transparently retries after downstream semantic output begins

Zero exceptions in strict mode.

## I4 — Exactly one upstream attempt is authoritative for a request

Failover closes/fences the previous attempt before replacement becomes authoritative.

## I5 — Unknown quota is not exhaustion

Lack of telemetry does not disable a healthy account.

## I6 — Exhaustion scope is preserved

Model/bucket-specific exhaustion does not become global account exhaustion without evidence.

## I7 — Provider overload is not account exhaustion

Do not burn through the whole pool because the provider itself is unhealthy.

## I8 — Reset instants survive sleep/restart correctly

Persist absolute wall-clock timestamps.

## I9 — Stale routing decisions cannot revert a newer binding

Monotonic generation / compare-and-swap.

## I10 — No account switch requires a Codex process restart in the primary path

If this invariant cannot be demonstrated in the MVP, the core architectural hypothesis has failed and must be revisited.

## I11 — AgentPool does not own tool/workspace semantics

No gradual scope creep back into agent orchestration.

## I12 — Technically possible but provider-prohibited access is not supported

Policy compatibility is part of correctness.

---

# 42. Acceptance Scenarios

## Scenario A — Normal sticky operation

```text
session S → account A
request 1 → A
request 2 → A
request 3 → A
```

Expected:

```text
no needless rotation
```

## Scenario B — Early drain

```text
A reaches trusted drain threshold
current request is already streaming
```

Expected:

```text
current request finishes on A
A becomes DRAINING
next request for S binds B
```

## Scenario C — Hard exhaustion before output

```text
A rejects request before downstream output
```

Expected:

```text
A relevant capacity scope enters COOLDOWN
B receives identical request
client sees only B's valid response
```

## Scenario D — Stream breaks after output

```text
A emitted valid semantic output
stream dies
```

Expected:

```text
no hidden retry on B
client receives stream failure
```

## Scenario E — All accounts unavailable

Expected:

```text
structured no-capacity error
nearest trustworthy reset if known
no spin loop
no API-key fallback
```

## Scenario F — Machine sleeps past reset

Expected:

```text
on wake/restart, expired cooldown is reconsidered immediately
```

## Scenario G — Account B refresh races

Expected:

```text
one refresh transaction
all waiters share result
no refresh-token reuse race
```

## Scenario H — Provider-wide 503

Expected:

```text
bounded provider backoff
not A→B→C→D quota rotation storm
```

## Scenario I — Model X exhausted, model Y usable

Expected:

```text
X requests avoid affected account
Y requests may still use it if provider evidence says Y is independent
```

## Scenario J — AgentPool crashes mid-stream

Expected:

```text
caller observes transport failure
AgentPool restarts
no attempt to fabricate continuation from partial stream
```

## Scenario K — T3 user switches nothing manually

Expected:

```text
same ACP/Codex session keeps running
AgentPool rotations remain invisible
```

---

# 43. Phase 2 and Beyond

Only expand after the Codex switch is boringly reliable.

## Phase 2 — Claude

Research the exact current supported seam.

Goal order:

```text
transport-level switch if provably supported
otherwise provider-specific runtime/profile adapter
```

Keep provider-specific complexity isolated.

## Phase 3 — Other permitted providers

A new provider earns support only if there is a legitimate authenticated machine interface and a clear billing/identity model.

Prefer:

```text
model transport gateway
```

before:

```text
runtime replacement
```

## Phase 4 — Optional higher-level continuity

Only if a real customer/project needs cross-provider agent continuation, introduce a separate component such as:

```text
agentpool-handoff
```

or leave that job entirely to TradeAgent.

Do not bake it back into the credential router.

---

# 44. Decision Table for Future Providers

When adding a provider, answer these questions in order.

| Question | Yes | No |
|---|---|---|
| Is third-party use permitted for this auth/subscription path? | continue | do not support |
| Can the native agent point model traffic at a custom endpoint? | use gateway mode | continue |
| Can AgentPool inject the selected account without changing agent session state? | use switch architecture | continue |
| Can a supported native runtime be swapped/resumed across accounts? | provider adapter may supervise | continue |
| Can safe handoff be proven at agent boundary? | optional higher layer | wait / unsupported |

This prevents provider oddities from contaminating the common path.

---

# 45. Operational Philosophy

AgentPool should feel uneventful.

Good behavior looks like:

```text
A healthy → keep using A
A nearing relevant limit → prepare B
A drains → next request uses B
A resets → becomes eligible again
```

Not:

```text
constant account hopping
constant probing
constant process spawning
constant checkpoint generation
constant filesystem mirroring
```

A reliable switch is mostly idle.

---

# 46. Why This Architecture Matches the Original Vision Better

The stated vision is that external software should experience:

```text
one continuous stream of compute
```

while AgentPool secretly manages:

```text
many subscription accounts
```

The most faithful implementation is therefore the one where the *external agent itself remains continuous*.

The heavier broker made continuity something AgentPool had to reconstruct.

The Switch Architecture makes continuity something AgentPool **does not disturb**.

That is a stronger design.

---

# 47. Final Product Boundary

AgentPool owns exactly this pipeline:

```text
incoming model request
        ↓
local caller authorization
        ↓
pool selection
        ↓
eligible account selection
        ↓
selected upstream credential
        ↓
provider request
        ↓
capacity/error classification
        ↓
optional pre-output retry
        ↓
transparent response stream
```

Everything above the incoming model request belongs to the coding agent / host.

Everything below the provider response belongs to the provider.

That is the line AgentPool should defend ruthlessly.

---

# 48. The “Volkswagen/Toyota Engine” Version

If the entire design has to fit on one screen, it is this:

```text
                 NATIVE AGENT
         (Codex / T3 / TradeAgent)
                      │
                      │ model request
                      ▼
              ┌───────────────┐
              │   AgentPool   │
              │               │
              │  1. sticky A  │
              │  2. check A   │
              │  3. send      │
              │  4. if A is   │
              │     rejected  │
              │     before    │
              │     output: B │
              │  5. stream    │
              └───────┬───────┘
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
          A           B           C

RULES
─────
• stay sticky while healthy
• prepare next account before needed
• switch only between safe model requests
• retry only before output escapes
• never change billing/model silently
• never pretend unknown quota is exhausted
• never move agent state if you do not have to
```

That is the machine to build.

---

# 49. Implementation Checklist

## Foundation

- [ ] single daemon process
- [ ] singleton lock
- [ ] loopback data plane
- [ ] owner-only control plane
- [ ] local bearer authentication
- [ ] SQLite migrations
- [ ] secret-store abstraction

## Routing

- [ ] accounts
- [ ] pools
- [ ] sticky binding
- [ ] stable weighted round-robin
- [ ] binding generation CAS
- [ ] READY/DRAINING/COOLDOWN/DISABLED
- [ ] relevant-bucket capacity filtering
- [ ] absolute reset timestamps

## Streaming

- [ ] finite request replay source
- [ ] response bootstrap buffer
- [ ] structured pre-output error classification
- [ ] bounded retry count
- [ ] close old upstream before replacement
- [ ] irreversible “downstream committed” flag
- [ ] zero retries after commit

## OpenAI

- [ ] supported SIWC account registration
- [ ] separate registration metadata
- [ ] token secret storage
- [ ] single-flight refresh
- [ ] principal verification
- [ ] rate-limit observation parser
- [ ] subscription-only guard

## Codex

- [ ] generated custom-provider config
- [ ] current Responses compatibility
- [ ] `agentpool run codex`
- [ ] two-account end-to-end continuity fixture

## T3

- [ ] codex-acp custom gateway path
- [ ] `agentpool acp codex`
- [ ] same-thread smoke test

## Hardening

- [ ] daemon restart tests
- [ ] machine sleep/reset tests
- [ ] provider 5xx storm tests
- [ ] malformed SSE tests
- [ ] huge request spool test
- [ ] secret leakage test
- [ ] ambient API-key contamination test
- [ ] bounded telemetry retention
- [ ] doctor command

---

# 50. Research Notes and Sources

This design intentionally distinguishes **normative provider documentation** from **implementation/community evidence**.

Provider behavior changes quickly. Every production adapter must therefore retain conformance tests instead of assuming these sources remain true forever.

## 50.1 OpenAI / Codex — primary

- OpenAI — Sign in with ChatGPT, open-source plan usage overview  
  <https://developers.openai.com/siwc/token-sharing-open-source>
- OpenAI — Accounts and sessions  
  <https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions>
- OpenAI — Registration and sign-in  
  <https://developers.openai.com/siwc/token-sharing-open-source/sign-in>
- OpenAI — Codex app-server integration  
  <https://developers.openai.com/siwc/token-sharing-open-source/codex-app-server>
- OpenAI — Errors and recovery  
  <https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery>
- OpenAI — Preview limitations  
  <https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations>
- OpenAI Codex source — current model-provider/custom base URL implementation  
  <https://github.com/openai/codex>

### Design conclusions drawn

- independent account registrations must remain separate;
- token refresh must be serialized per registration;
- Codex can be pointed at a custom Responses-compatible provider;
- structured rate-limit data is useful but must be treated as scoped observations;
- a model catalog is not equivalent to entitlement proof.

## 50.2 Agent Client Protocol / Codex ACP — primary/open source

- Agent Client Protocol  
  <https://agentclientprotocol.com/>
- ACP Rust SDK / protocol repositories  
  <https://github.com/agentclientprotocol>
- Codex ACP adapter  
  <https://github.com/zed-industries/codex-acp>

### Design conclusions drawn

- ACP already has native session/resume concepts;
- AgentPool does not need to invent another agent-session protocol;
- the Codex ACP adapter’s custom gateway configuration provides a narrow T3 integration seam.

## 50.3 T3 Code — implementation evidence

- T3 Code repository  
  <https://github.com/pingdotgg/t3code>

Relevant evidence reviewed includes:

- multi-account Codex behavior;
- shared Codex session state vs separate account authentication;
- local ACP provider configuration;
- active-writer race reports;
- shadow-home breakage when Codex added new state files.

### Design conclusion drawn

Filesystem mirroring and runtime replacement are fragile compared with switching beneath a stable Codex process.

## 50.4 Multi-account routing projects — implementation evidence

- CLIProxyAPI  
  <https://github.com/router-for-me/CLIProxyAPI>
- CLIProxyAPI Plus  
  <https://github.com/router-for-me/CLIProxyAPIPlus>
- codexctl and other small Codex account switchers reviewed through GitHub search
- Claude multi-account router projects reviewed through GitHub search

Useful patterns observed:

```text
OAuth account pools
sticky/session-affinity routing
round-robin / weighted selection
quota cooldown
early disabling/draining near limits
bootstrap retries before first response byte
```

Useful failures observed:

```text
credential-wide cooldown applied to narrower quota buckets
stale cooldown after machine sleep
quota-limited credentials remaining selectable
prompt-cache churn from needless account switching
```

## 50.5 Anthropic / Claude — primary + implementation evidence

- Anthropic Help — current Claude Agent SDK / `claude -p` subscription usage guidance  
  <https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan>
- Claude Code documentation  
  <https://code.claude.com/docs>
- Claude Code repository/issues  
  <https://github.com/anthropics/claude-code>

### Design conclusion drawn

Claude should receive a dedicated adapter only after the current supported transport/auth path is conformance-tested; do not force Codex assumptions onto it.

## 50.6 Google Antigravity — primary

- Google Antigravity FAQ / documentation  
  <https://antigravity.google/docs/faq>

### Critical policy conclusion

Current Google guidance says third-party software/tools/services must not access Antigravity through the consumer login path. AgentPool therefore does not support consumer Antigravity subscription rotation unless that policy changes.

## 50.7 OpenCode — implementation evidence

- OpenCode repository/issues  
  <https://github.com/anomalyco/opencode>

Useful failure modes reviewed:

- duplicate overlapping LLM streams;
- long-running event/database growth.

### Design conclusion drawn

Keep one authoritative upstream stream per request and keep AgentPool’s persisted operational history bounded.

---

# 51. Final Recommendation

Build **AgentPool v1 as a Codex-focused local subscription switch**, not as a universal durable execution broker.

The priority stack should be:

```text
1. correctness of account identity
2. correctness of safe pre-output failover
3. correctness of quota scope
4. invisible Codex continuity
5. operational simplicity
6. only then additional providers
```

The defining implementation test is not:

> “Can AgentPool restart a complicated agent job on another account?”

It is:

> **“Can one living Codex thread keep working while AgentPool changes the authorized subscription account underneath its next inference request, with no duplicate response and no billing surprise?”**

If that works reliably, AgentPool has achieved the original vision with dramatically less machinery.

And because it owns less, it has less to break when Codex, T3, provider schemas, tools, workspaces, and orchestration evolve around it.

That is the architecture I would build.
