# AgentPool — Robust Local Subscription-Aware Agent Execution Broker

**Status:** implementation architecture / researched design baseline  
**Research cut:** 2026-10-06  
**Primary goal:** turn any number of independently authenticated, user-authorized AI agent accounts into durable local execution pools without reimplementing the agents themselves.  
**Design priority:** correctness and recoverability first; upgradeability second; convenience third.

---

## 0. Final Definition

> **AgentPool is a local execution broker that turns independently authenticated AI-agent accounts into logical pools of durable execution capacity. Applications and humans request an execution capability, model, or pool; AgentPool selects an eligible account, starts the provider's supported machine-facing agent runtime, supervises it, observes usage/capacity where the provider exposes it, moves unavailable accounts out of admission, and continues the same durable job through another compatible account when continuity can be proven safe. When a provider boundary cannot preserve native state safely, AgentPool uses an explicit handoff or waits rather than guessing.**

AgentPool is deliberately **not** an LLM proxy, prompt platform, model API, AI framework, credential-sharing service, or replacement for Codex, Claude Code, Google Antigravity, T3 Code, TradeAgent, or AI Concierge.

The system owns one thing: **reliable local execution continuity across a changing set of authenticated agent capacity.**

---

# 1. What the Research Changes

The original concept is valid, but the implementation should not be built around a universal `provider -> slot -> session` assumption. Current providers already expose materially different machine interfaces, authentication boundaries, session stores, quota surfaces, and resume behavior.

The architecture therefore uses one rule above all others:

> **Provider names never imply capabilities. Capabilities are discovered, version-gated, conformance-tested, and attached to a concrete runtime/account configuration.**

This matters because the current 2026 provider landscape is already heterogeneous:

| Area | OpenAI Codex | Anthropic Claude Code | Google subscription agent |
|---|---|---|---|
| Preferred machine interface | `codex app-server` JSON over stdio | Claude Code structured `stream-json` / Agent SDK path | Official Google Antigravity ACP server |
| Subscription-backed programmatic path | Official Sign in with ChatGPT flow | `claude -p` / Agent SDK currently draw from subscription usage | Antigravity replaces Gemini CLI for individual consumer subscription access |
| Multiple account isolation | Separate OAuth registrations; OpenAI explicitly documents managing multiple saved ChatGPT accounts | Separate `CLAUDE_CONFIG_DIR` per account | Separate Antigravity provider/auth profile per account |
| Same-account native resume | Yes, via native thread/session mechanisms | Yes, via Claude session resume | Yes when the ACP runtime advertises resume |
| Cross-account native resume | Possible for Codex when the session domain is shared and the exact runtime passes conformance; T3 Code already demonstrates the pattern | Do **not** assume: auth/config isolation also isolates local session history | Do **not** assume |
| Structured quota visibility | Strongest: account rate-limit RPC, multi-window observations, account association | Partial/provider-version-dependent; structured rate-limit events exist but must be treated cautiously | Current official ACP surface does not reliably expose full plan quota |
| Model catalog means entitlement? | No; a successful inference is authoritative | Do not assume catalog alone proves billing/model path | Account catalog is useful; actual turn remains authoritative |

Two current facts are especially important:

1. OpenAI's open-source Sign in with ChatGPT documentation explicitly describes applications managing **multiple ChatGPT account registrations**, keeping each registration's credentials separate, and switching among saved accounts. It also provides an official Codex app-server integration for using a ChatGPT-plan OAuth token.
2. Google moved individual Google AI Pro/Ultra/free coding-agent subscription traffic away from Gemini CLI to Antigravity in June 2026. A design that hard-codes `gemini` as the Google consumer-subscription adapter is already obsolete.

AgentPool must therefore be built to survive exactly this kind of provider evolution.

---

# 2. Product Boundaries

## 2.1 AgentPool owns

- account/principal registrations and references to their authentication material;
- provider runtime discovery and compatibility validation;
- provider process lifecycle;
- logical pools;
- deterministic slot eligibility and routing;
- concurrency admission and leases;
- capacity observations and cooldown/admission decisions;
- durable jobs, attempts, native-session metadata, and turn boundaries;
- process crash/restart reconciliation;
- continuity decisions;
- structured execution events;
- local API and ACP compatibility surfaces;
- safe waiting and wake-up when no eligible capacity exists;
- provider/runtime upgrade compatibility information;
- optional generic handoff packaging when native continuation is impossible.

## 2.2 The host application owns

- why work should happen;
- task decomposition;
- domain-specific hierarchy and management;
- business/domain policy;
- semantic verification requirements;
- project-specific handoffs when it can produce a better one;
- domain-specific reconciliation of external side effects;
- decisions about whether cross-provider substitution is semantically acceptable.

For TradeAgent this means its hierarchy, JEV/context-cleansing logic, managerial decisions, trading/domain logic, and verification policy remain entirely outside AgentPool.

## 2.3 Native agents own

- model interaction;
- provider-specific context construction;
- provider-native session semantics;
- built-in tools and subagents;
- provider-specific MCP/tool behavior;
- provider authentication flows when AgentPool is configured to delegate auth;
- provider-specific safety and permission semantics.

## 2.4 Explicit non-goals

AgentPool does not become:

- an OpenAI-compatible inference proxy;
- a generic LLM router;
- a new prompt protocol;
- a distributed cluster scheduler;
- a transcript database by default;
- a cloud account service;
- a team identity system;
- an implementation of Codex/Claude/Antigravity internals;
- a system that claims arbitrary external side effects can be replayed exactly once.

---

# 3. Architectural North Star

The desired external experience stays extremely small:

```bash
agentpool run codex
```

or:

```text
execute(
    pool = "codex",
    workspace = "/projects/tradeagent",
    task = "Implement and verify X",
    continuity = "strict-automatic"
)
```

or, for an ACP client such as T3 Code:

```bash
agentpool acp --pool codex
```

Everything below that boundary is AgentPool's responsibility:

```text
Which paid account?
Which runtime version?
Which auth profile?
Which native session domain?
Is this model actually available to this account?
Is included subscription usage currently available?
Is paid overage allowed?
Is another account compatible with this native session?
Did the previous turn reach a safe continuation boundary?
Can we resume natively, hand off, or only wait?
Did the daemon crash while a provider process was alive?
Is the workspace already leased by another job?
Did a provider update change protocol behavior?
Can this event be replayed to a reconnecting client?
```

---

# 4. Two Planes, One Core

AgentPool should expose two different external surfaces rather than forcing every client through one protocol.

```text
                  HUMAN CLI
                     │
                     │
TradeAgent ───────┐  │  ┌────── AI Concierge
                  ▼  ▼  ▼
          ┌─────────────────────┐
          │  CONTROL PLANE API  │
          │  durable jobs       │
          │  pools / slots      │
          │  events / admin     │
          └─────────┬───────────┘
                    │
             ┌──────▼──────┐
             │ AGENTPOOL   │
             │   DAEMON    │
             └──────┬──────┘
                    │
          ┌─────────┴──────────┐
          │                    │
          ▼                    ▼
  provider drivers      ACP compatibility
                               facade
                                 │
                                 ▼
                        T3 / Zed / ACP client
```

## 4.1 Control plane

The control plane is AgentPool's authoritative API for durable execution.

Default transport:

- macOS/Linux: Unix domain socket with owner-only permissions;
- Windows: named pipe with current-user ACL;
- optional loopback HTTP only as an explicit later compatibility mode.

The wire protocol should be versioned JSON-RPC 2.0 or an equivalently simple typed request/response protocol. JSON-RPC is preferred because AgentPool already has to handle JSON-RPC-like provider protocols and it keeps the local interface inspectable.

The control plane exposes durable concepts ACP does not need to know about: pools, attempts, leases, capacity, account setup, runtime management, diagnostics, and event replay.

## 4.2 Agent compatibility plane

`agentpool acp --pool <pool>` is a lightweight ACP agent facade. It connects to the daemon and maps one logical AgentPool pool to one ACP-visible agent.

This is important for T3 Code because current T3 provider settings support generic ACP agents launched from a local command. AgentPool therefore does not need a T3-specific fork or custom T3 provider implementation.

The ACP facade should:

- support stable ACP v1 first;
- negotiate protocol versions rather than assume them;
- add ACP v2 only behind a separate compatibility gate while the upstream Rust SDK still marks v2 as unstable/draft;
- advertise model/mode/thought settings through ACP config options where supported;
- forward permission requests and structured tool/session updates;
- map the external ACP session to one durable AgentPool job;
- keep the external session identity stable while internal attempts/accounts can change.

---

# 5. Implementation Language and Runtime Shape

## 5.1 Use Rust for the core daemon

For this project, Rust is the better long-term implementation choice than a Node/TypeScript daemon.

Reasons are architectural, not aesthetic:

- AgentPool is a long-lived local supervisor whose hardest bugs are process ownership, cancellation, concurrency, file permissions, crash recovery, and state-machine correctness.
- The official ACP Rust SDK is mature enough for stable ACP v1 and currently includes subprocess/process-group handling; v2 can remain separately gated.
- Rust makes one self-contained native binary practical on macOS, Linux, and Windows.
- Windows Job Objects, Unix process groups, filesystem permission checks, atomic rename, and native IPC are all first-class implementation concerns.
- SQLite has an extremely mature Rust binding (`rusqlite`) and the state volume is small enough that no async database framework is required.
- Codex app-server and Claude structured output are line-oriented JSON protocols and do not require a JavaScript runtime.

The project should remain one repository and one shipped product. Rust modules/crates are internal boundaries, **not** a dynamic plugin framework.

Suggested workspace:

```text
agentpool/
├── crates/
│   ├── agentpool-core/          # domain model, state transitions, policies
│   ├── agentpool-storage/       # SQLite and migrations
│   ├── agentpool-runtime/       # processes, IPC, runtime installations
│   ├── agentpool-protocol/      # control-plane wire types
│   ├── agentpool-provider/      # driver traits + normalized provider events
│   ├── provider-codex/
│   ├── provider-claude/
│   ├── provider-acp/            # generic inbound ACP provider driver
│   ├── agentpool-acp/           # outward ACP facade
│   ├── agentpool-daemon/
│   └── agentpool-cli/
├── compatibility/               # tested runtime manifests / quirks
├── fixtures/                    # redacted protocol fixtures
├── tests/
└── docs/
```

A TypeScript SDK can later be generated for TradeAgent/Concierge from the control-plane schema. It should not require the daemon itself to be written in TypeScript.

---

# 6. Core Domain Model

The original five-object model is too compressed for durable failover. The minimum robust model is:

```text
ProviderDriver
    │
    ├── RuntimeInstallation
    │
    ├── Principal ── AuthProfile
    │
    ├── ExecutionProfile
    │
    └── SessionDomain
             │
             ▼
            Slot
             │
             ▼
            Pool
             │
             ▼
             Job
             │
        ┌────┴────┐
        ▼         ▼
     Attempt   Attempt ...
        │
   NativeSession
        │
       Turn
```

Supporting durability objects:

```text
Lease
QuotaObservation
WorkspaceLease
Checkpoint / HandoffEnvelope
Event
RuntimeLease
IdempotencyRecord
```

## 6.1 ProviderDriver

A built-in integration implementation, for example:

```text
codex-app-server
claude-stream-json
acp-generic
```

A driver is protocol logic. It is not an account.

## 6.2 RuntimeInstallation

One exact executable/distribution version and fingerprint.

```text
runtime_id
provider_driver
version
binary_path
binary_hash
source             # managed | external
installed_at
compatibility_state
capability_snapshot
```

No running process should ever silently move from one runtime version to another.

## 6.3 Principal

One independently authenticated user/account/workspace identity.

```text
principal_id
provider
provider_subject_or_account_id
workspace_id?          # where applicable
email_hint?            # display only, never identity
plan_hint?
created_at
last_verified_at
```

Email is never a primary key. OpenAI explicitly warns that separate account registrations can have the same email; identity and registration must remain distinct.

## 6.4 AuthProfile

How a principal is authenticated and where credentials live.

Examples:

```text
OpenAI SIWC OAuth registration
Claude CLAUDE_CONFIG_DIR profile
Antigravity ACP profile
```

The SQLite database stores only metadata and a secret-store reference when AgentPool itself must own a token.

## 6.5 ExecutionProfile

The deterministic behavior configuration expected when a slot runs:

```text
model policy
reasoning/effort policy
permission policy
MCP set
provider settings
hooks/plugins policy
allowed workspace roots
environment allowlist
```

It has a stable content hash. Continuity can require two slots to have the same execution-profile hash.

## 6.6 SessionDomain

The provider-native namespace in which sessions/history live.

This object is the key correction to the earlier architecture. Account identity and session storage are not universally the same boundary.

Examples:

- multiple Codex principals may intentionally share one Codex session domain while injecting different OAuth credentials into app-server processes;
- a Claude profile's `CLAUDE_CONFIG_DIR` contains credentials **and** session history, so separate accounts normally have separate session domains;
- an Antigravity ACP profile should be treated as its own session domain unless the provider explicitly proves cross-profile resume.

Multiple slots may reference the same session domain only when the driver/runtime's conformance profile permits it.

## 6.7 Slot

A schedulable execution identity:

```text
Slot =
  ProviderDriver
+ RuntimeInstallation
+ Principal
+ AuthProfile
+ ExecutionProfile
+ SessionDomain
+ BillingPolicy
```

A slot is configuration/data. There is no hard-coded number of slots.

## 6.8 Pool

A named policy over candidate slots.

Examples:

```text
codex
claude
google
best-coder
fast
research
```

A pool contains eligibility rules, routing strategy, continuity policy, billing policy, and optional provider ordering.

## 6.9 Job

The durable provider-independent execution identity exposed to clients.

One ACP session normally maps to one job. A TradeAgent builder task normally maps to one job.

A job survives:

- caller disconnect;
- provider process restart;
- account capacity exhaustion;
- daemon restart;
- compatible account rotation;
- optional provider handoff.

## 6.10 Attempt

One assignment of a job to one slot/runtime/native session.

Attempts are essential because a job may move across slots while retaining one stable job ID.

## 6.11 NativeSession

Provider-owned conversation/thread state:

```text
provider
session_domain_id
native_session_id
runtime_id
created_by_principal
resume_constraints
```

The ID is namespaced. A native session ID alone is never globally meaningful.

## 6.12 Turn

A single accepted unit of user/host input through a provider-defined idle/completion boundary.

Turn state is where replay safety is determined.

## 6.13 Lease

A fencing token granting authority to mutate a job or occupy a slot/workspace.

Every replacement attempt increments the lease generation. An old worker whose lease generation is stale can no longer commit authoritative state.

---

# 7. Capability Model: Never Guess From Provider Name

Every usable runtime produces a `RuntimeCapabilityProfile`.

Conceptually:

```yaml
protocol:
  kind: codex-app-server
  structured_events: true
  explicit_turn_acceptance: true
  explicit_turn_completion: true

sessions:
  native_resume: true
  cross_principal_resume: conditional
  resume_requires_same_domain: true

identity:
  observable: strong
  billing_mode_observable: strong

capacity:
  snapshot: multi_window
  reset_time: sometimes
  account_association: true

permissions:
  structured: true

models:
  catalog: true
  catalog_is_entitlement: false

process:
  deterministic_shutdown: tested
```

Capabilities have evidence:

```text
PROTOCOL_NEGOTIATED
OFFICIAL_DOCUMENTATION
CONFORMANCE_TEST
KNOWN_GOOD_MANIFEST
USER_OVERRIDE
```

A dangerous capability such as cross-principal native resume must require stronger evidence than a display capability.

### Compatibility admission order

```text
1. exact runtime fingerprint known-bad?       -> BLOCK
2. exact/range known-good?                    -> use declared + probed capabilities
3. unknown version?                           -> run non-destructive conformance suite
4. required capability passes?                -> admit as VERIFIED_LOCAL
5. optional capability fails?                 -> degrade only that feature
6. lifecycle/auth safety cannot be proven?    -> QUARANTINE/BLOCK
```

This is the center of AgentPool's upgradeability strategy.

---

# 8. Provider Integration — OpenAI Codex

## 8.1 Preferred authentication: Sign in with ChatGPT for OSS

OpenAI now documents an open-source OAuth flow specifically for local/self-hosted applications using ChatGPT-plan usage.

The design should use it directly rather than rotating `auth.json` files.

For each ChatGPT registration AgentPool stores:

```text
validated subject identity
issued client_id
workspace binding where applicable
stable local account label
granted scopes
token expiry metadata
protected token secret reference
```

OpenAI's current documentation explicitly requires separate credential records for multiple saved account registrations and warns not to combine one registration's client ID with another registration's tokens.

Token refreshes for one registration must be serialized because refresh tokens rotate.

## 8.2 Preferred execution interface: Codex app-server

Launch:

```text
codex app-server --listen stdio://
```

with the ChatGPT-plan provider configuration and only the selected account's access token in the child environment.

The adapter drives:

```text
initialize
initialized
thread/start
thread/resume
turn/start
structured notifications
account/read
account/rateLimits/read
model/list
```

A native thread ID is stored durably as soon as it is returned.

`turn/completed` is authoritative for successful completion; failed/interrupted status is not silently promoted to success.

## 8.3 Session-domain strategy

Preferred AgentPool mode:

- auth lives outside Codex state through OAuth token injection;
- a managed Codex session domain owns local native thread/history state;
- multiple paid ChatGPT principals may target the same session domain only after the exact runtime passes AgentPool's cross-principal-resume conformance test.

This is cleaner than cloning or symlinking an entire `CODEX_HOME` because provider releases can add new files at any time. T3 Code currently demonstrates that shared Codex session state plus account-specific authentication can continue the same threads, but its shadow-home implementation has also encountered breakage when Codex added new state files. AgentPool should avoid making filesystem mirroring the primary abstraction.

A native-login compatibility backend may be provided later for users who do not want AgentPool-managed OAuth, but it is secondary to the supported token-injection path.

## 8.4 Capacity

Codex currently exposes a structured `account/rateLimits/read` response with:

- an `ordinaryUsageAllowed` value;
- one historical/default bucket;
- optional multiple buckets keyed by backend limit ID;
- window utilization/reset timestamps where supplied;
- optional credit/reset-credit information;
- associated account ID where supplied.

Important rule from OpenAI's own schema:

> if `ordinaryUsageAllowed` is unavailable/null, clients must not infer recovery merely from percentages or reset times.

AgentPool therefore stores rate limits as observations, not as a single `cooldown_until` fact.

The model catalog is also **not** an entitlement proof. A successful inference turn is authoritative for model access.

## 8.5 Billing policy

Default:

```text
included_subscription_only
```

AgentPool must never silently consume credits/overage merely to preserve uptime. Credits are an explicit per-pool/per-slot policy:

```text
included_subscription_only
allow_provider_credits
```

---

# 9. Provider Integration — Anthropic Claude Code

## 9.1 Current supported machine path

Anthropic's current June 2026 help guidance states that, for now, Claude Agent SDK, `claude -p`, and third-party application usage still draw from the user's Claude subscription usage limits.

AgentPool should use a structured Claude process, not scrape an interactive terminal:

```text
claude -p
  --input-format stream-json
  --output-format stream-json
```

with the supported session/resume flags as required.

The adapter consumes structured messages for:

- session identity;
- assistant output;
- tool activity;
- subagent activity where exposed;
- permission requests;
- result/termination;
- usage/rate-limit observations where emitted.

## 9.2 Account isolation

Each paid Claude account receives an absolute, canonical `CLAUDE_CONFIG_DIR`.

Official Claude Code documentation currently states that this directory contains settings, credentials, session history, and plugins and is the supported mechanism for running multiple accounts side-by-side.

That fact has an important consequence:

```text
separate Claude account
        usually means
separate native session domain
```

Therefore AgentPool does **not** assume that a Claude native session can move from account A's config directory to account B's config directory.

Cross-account Claude continuation is a handoff/new-session operation unless a future runtime explicitly proves otherwise.

## 9.3 Billing-path guard

Claude has several possible auth/billing paths. Before launching a subscription-backed slot, AgentPool must scrub or explicitly set the launch environment so ambient values cannot silently override the intended subscription.

At minimum, subscription slots reject/scrub conflicting values such as:

```text
ANTHROPIC_API_KEY
ANTHROPIC_AUTH_TOKEN
unexpected ANTHROPIC_BASE_URL
cloud-provider routing variables
```

unless those values are intentionally part of the slot's declared auth mode.

Recent Claude Code issue reports show why this is necessary: headless/auth-status behavior has had versions where displayed account state did not match the credential actually used for inference. AgentPool therefore combines environment isolation, version conformance, and identity/billing checks; it never treats `auth status` alone as proof.

## 9.4 Capacity semantics

Claude capacity observation is weaker than Codex and must be modeled accordingly.

Structured `rate_limit_event` data can expose reset information and limit type on current versions, but generic HTTP 429/provider throttling must not be classified as subscription exhaustion automatically.

Classification order:

```text
explicit subscription usage event  -> capacity observation
explicit reset timestamp           -> cooldown observation
transient provider throttle        -> bounded retry/backoff
unknown 429                         -> UNKNOWN, not account exhaustion
```

## 9.5 Extra usage

As with OpenAI, extra paid usage/overage is an explicit policy, never an automatic surprise.

---

# 10. Provider Integration — Google Antigravity

## 10.1 Do not build consumer subscription routing around Gemini CLI

Google announced that on June 18, 2026, Gemini CLI and Gemini Code Assist stopped serving individual Google AI Pro/Ultra and free-individual coding-agent requests. Consumer subscription execution moved to Antigravity CLI / Antigravity 2.0.

AgentPool's Google subscription adapter should therefore be named around the actual product boundary, for example:

```text
google-antigravity
```

not a frozen `gemini` assumption.

## 10.2 Preferred machine interface: official Antigravity ACP server

The official ACP Registry contains Google's Antigravity ACP agent distributed from Google infrastructure. AgentPool should consume this as an ACP **client** through its generic inbound ACP driver.

Benefits:

- no terminal scraping;
- negotiated capabilities;
- provider-owned authentication;
- standard session/prompt/cancel/permission semantics;
- native session resume when advertised;
- easier future replacement if Google updates the ACP distribution.

## 10.3 Account isolation

Use one AgentPool auth/profile root per Google account and one slot per independently authenticated principal. The installed Antigravity executable can be shared; credentials/configuration cannot.

T3 Code already implements this pattern and additionally removes ambient Google credentials to prevent a provider instance from silently using another account or billing project. AgentPool should do the same.

## 10.4 Capacity

Current Antigravity ACP does not reliably expose the complete plan quota to ACP clients. Therefore:

```text
capacity = UNKNOWN
```

is a normal healthy state.

AgentPool learns authoritative exhaustion from the provider's structured failure/retry information when it occurs. It must not screen-scrape an unrelated Antigravity CLI session whose login may be different from the ACP profile.

A future ACP quota capability can be adopted without changing the scheduler because quota is already modeled as optional observations.


---

# 11. Generic ACP Provider Driver

A major future-proofing decision is to make ACP a first-class **inbound** provider protocol as well as an outward compatibility protocol.

```text
              AgentPool core
                   │
        ┌──────────┴──────────┐
        │                     │
 provider-specific       generic ACP client
    adapters                  │
        │                     ▼
 Codex / Claude       Antigravity / future ACP agents
```

The generic ACP driver should:

- negotiate the highest mutually supported protocol version;
- support stable v1 by default;
- preserve unknown extensions/fields where forwarding is safe;
- treat omitted capability markers as unsupported;
- expose ACP session resume only when negotiated;
- normalize ACP permission requests, config options, model options, usage updates, tool updates, and stop reasons into AgentPool's internal event model;
- keep provider-specific enrichers outside the generic ACP transport.

This gives future providers an inexpensive path into AgentPool: if a provider ships a correct ACP agent, only a small capability/identity/quota enrichment layer may be required instead of a bespoke execution adapter.

---

# 12. Authentication and Principal Isolation

Authentication bugs are catastrophic because the software may appear to work while consuming the wrong account.

AgentPool therefore treats account identity as an execution invariant.

## 12.1 Required principal properties

Every slot has:

```text
expected_principal_id
expected_auth_mode
expected_billing_mode
expected_workspace/account binding where applicable
```

Before admission, the driver obtains the strongest available identity evidence and records:

```text
observed_principal
observed_plan_hint
observed_auth_mode
observed_at
evidence_strength
```

If the provider can strongly prove that the actual principal differs from the slot's expected principal:

```text
slot -> PRINCIPAL_MISMATCH -> QUARANTINED
```

No automatic fallback to “whatever account works” is allowed.

## 12.2 SecretStore abstraction

SQLite never stores raw refresh/access tokens.

```text
SecretStore
├── macOS Keychain
├── Windows credential protection/store
├── Linux Secret Service when available
└── explicit owner-only file fallback for headless systems
```

The database stores an opaque secret reference.

OpenAI's own OSS flow permits owner-only protected files; AgentPool can support that fallback, but a platform credential store is preferable where reliable.

Provider-native secrets owned by Claude or Antigravity remain provider-native. AgentPool does not copy them into its own database.

## 12.3 Environment construction is allowlist-first

A child does not inherit the daemon's entire environment.

Instead:

```text
base OS-safe environment
+ explicit path/runtime variables
+ slot-declared variables
+ required credential injection
- provider-conflicting billing variables
- unrelated secrets
```

This prevents an exported API key, cloud credential, or custom base URL from silently changing the billing identity of a subscription slot.

## 12.4 Canonical paths only

All provider homes, workspaces, binaries, sockets, and managed runtime paths are canonical absolute paths.

No persisted path contains shell `~` expansion or depends on the caller's current working directory.

## 12.5 Auth changes are serialized

Login, logout, token refresh, profile replacement, and sign-out are exclusive slot operations.

During auth mutation:

```text
admission closed
new attempts blocked
active sessions handled according to driver policy
credential mutation completes
identity reverified
admission reopened
```

This prevents a helper process from retaining one account while the database thinks another account is active.

---

# 13. Capacity Is an Observation, Not a Boolean

A single `READY / COOLDOWN` flag is not expressive enough.

Each slot has four independent dimensions:

```text
authentication state
runtime/provider health
allocation state
capacity state
```

Example:

```text
auth:        VALID
runtime:     HEALTHY
allocation:  IDLE
capacity:    UNKNOWN
```

is perfectly usable.

## 13.1 QuotaObservation

```yaml
QuotaObservation:
  source: structured_snapshot | structured_error | retry_after | inferred
  scope: principal | app | model | provider | unknown
  limit_id: optional
  model_scope: optional
  status: available | exhausted | unknown
  used_percent: optional
  remaining: optional
  reset_at_utc: optional
  retry_after_ms: optional
  overage_available: optional
  observed_at_utc: required
  confidence: authoritative | strong | weak
  raw_schema_version: optional
```

Multiple observations can coexist.

This handles:

- several-hour rolling windows;
- weekly windows;
- app-specific usage limits;
- provider/model buckets;
- credits/overage;
- provider throttling;
- providers that expose no quota snapshot at all.

## 13.2 Dispatch is authoritative

External clients can use the same account outside AgentPool. Consequently:

> **Capacity snapshots are hints; the actual provider response at dispatch is authoritative.**

The scheduler must always tolerate an account becoming unavailable immediately after selection.

## 13.3 Wake-up behavior

When a trustworthy reset time exists:

```text
WAITING_FOR_CAPACITY
      │
      ├── durable wake_at = earliest eligible reset
      │
      ▼
reservation / probe with jitter
      │
      ├── available -> dispatch
      └── still unavailable -> new observation / wait
```

When no reset time exists, use bounded exponential backoff with jitter and a maximum probe cadence.

Never hammer every account at once.

---

# 14. Routing

Routing should remain deterministic and inspectable.

No machine-learning scheduler is needed.

## 14.1 Three-stage routing

### Stage A — eligibility

Reject any slot that fails a hard requirement:

```text
disabled
quarantined
auth invalid
principal mismatch
runtime unsupported
required capability missing
requested model known unavailable
included usage authoritatively exhausted
billing policy incompatible
slot concurrency lease unavailable
workspace policy incompatible
execution-profile incompatibility
```

### Stage B — continuity class

For an existing job, prefer:

```text
1. same live attempt / process
2. same native session + same principal
3. same native session + another compatible principal
4. provider handoff at safe boundary
5. cross-provider handoff if pool policy allows it
6. wait
```

Continuity outranks load balancing.

### Stage C — deterministic score

Among equivalent candidates, consider:

```text
known usable capacity
lower capacity pressure
recent failure penalty
active load
pool/provider priority
queue age / fairness
stable slot ID as final tie-break
```

The same state produces the same routing decision.

## 14.2 Sticky execution

Do not rotate accounts merely because another slot has slightly more capacity.

Once a job is running successfully, stay on the current slot until one of these occurs:

- current slot is exhausted;
- auth/runtime becomes invalid;
- host explicitly requests a change;
- policy requires rebalancing;
- current model becomes unavailable.

This minimizes continuity risk.

## 14.3 Reservation before probe/start

The scheduler obtains a short-lived reservation before performing an expensive authenticated probe or launch. This prevents a reset event from waking dozens of jobs that all race for the same newly available account.

## 14.4 Fairness

Even v1 needs minimal fairness because TradeAgent, Concierge, T3, and humans may share the same pool.

Each job has:

```text
priority_class
submitted_at
client_id
```

Default ordering:

```text
interactive > normal > background
```

with queue aging so background work cannot starve indefinitely.

Optional per-client concurrency limits prevent one autonomous host from consuming every slot.

---

# 15. Billing Policy Is First-Class

A “working” provider request is not enough if it silently changes from subscription usage to metered billing.

Every slot has one declared billing mode:

```text
SUBSCRIPTION_INCLUDED_ONLY
SUBSCRIPTION_PLUS_EXPLICIT_CREDITS
API_METERED
CUSTOM
```

Pools can restrict accepted modes.

For the product described here, the default pool is:

```text
SUBSCRIPTION_INCLUDED_ONLY
```

The following transitions are forbidden unless explicitly configured:

```text
subscription -> API key
subscription -> usage credits
authenticated account A -> ambient account B
personal subscription -> cloud project billing
```

If the provider cannot prove its billing path sufficiently for a runtime version, the slot can remain available only if its conformance policy allows that uncertainty. For strict pools it is blocked.

---

# 16. Job, Attempt, and Turn State Machines

## 16.1 Job state

```text
CREATED
   │
   ▼
QUEUED
   │
   ▼
DISPATCHING
   │
   ▼
RUNNING ───────────────► COMPLETED
   │
   ├── user/input ─────► WAITING_FOR_INPUT ──► RUNNING
   │
   ├── no capacity ────► WAITING_FOR_CAPACITY ─► DISPATCHING
   │
   ├── recoverable ────► RECOVERING ──────────► DISPATCHING
   │
   ├── unsafe unknown ─► RECOVERY_BLOCKED
   │
   └── cancel ─────────► CANCELLING ───────────► CANCELLED

unrecoverable failure ─► FAILED
```

`RECOVERY_BLOCKED` is not a failure. It means AgentPool refuses to guess about a potentially unsafe continuation.

## 16.2 Attempt state

```text
RESERVED
STARTING
RUNNING
STOPPING
ENDED
ORPHANED
```

Outcome is recorded separately:

```text
SUCCESS
CAPACITY_EXHAUSTED
AUTH_LOST
RUNTIME_FAILURE
PROVIDER_FAILURE
CANCELLED
INTERRUPTED
OUTCOME_UNKNOWN
```

## 16.3 Turn state

```text
PREPARED
ACCEPTED
RUNNING
IDLE/COMPLETED
FAILED
INTERRUPTED
OUTCOME_UNKNOWN
```

The distinction between `PREPARED` and `ACCEPTED` is critical for idempotency: if a provider never accepted the turn, retrying the same input is usually safe. Once accepted, AgentPool cannot assume the provider performed no work.

---

# 17. Execution Leases and Fencing

A durable local agent supervisor must defend against duplicate workers after crashes and races.

Every authoritative job attempt owns a monotonic lease generation:

```text
job-42 generation 17 -> attempt A
```

If recovery creates a replacement:

```text
job-42 generation 18 -> attempt B
```

Attempt A is permanently fenced out even if its process later emits output.

The same pattern applies to slot occupancy and exclusive workspace leases.

## 17.1 Invariant

> **At most one lease generation may authoritatively mutate a job at a time.**

The database transaction that supersedes an attempt must atomically:

1. revoke the old lease;
2. create the new lease generation;
3. create the replacement attempt;
4. append the corresponding durable event.

No provider process can self-declare ownership.

---

# 18. Workspace Safety

Account concurrency and workspace concurrency are different problems.

Two separate slots can still destroy the same checkout if they execute simultaneously.

AgentPool therefore requires a `WorkspaceLease` whenever the job can mutate files.

```text
workspace_identity
canonical_root
mode: exclusive | shared-readonly | host-managed
lease_generation
job_id
```

## 18.1 Default

For coding jobs:

```text
one exclusive writer per canonical worktree
```

## 18.2 Host-managed isolation

Sophisticated hosts such as TradeAgent may create their own worktrees/isolated workspaces. They declare:

```text
workspace_mode = host_managed_isolated
workspace_identity = stable UUID
```

AgentPool then protects the declared identity but does not invent Git policy.

## 18.3 Optional AgentPool worktrees

A later helper may create Git worktrees for interactive/simple clients. It is not required for the first MVP and must not turn the core broker into a source-control orchestrator.

---

# 19. Continuity: The Most Important Logic

Continuity is not classified merely as “same provider” or “different provider.”

AgentPool computes a concrete `ContinuityPlan` between the current attempt and a target slot.

Possible levels:

```text
REUSE_LIVE_PROCESS
NATIVE_RESUME_SAME_PRINCIPAL
NATIVE_RESUME_CROSS_PRINCIPAL
HANDOFF_AT_SAFE_BOUNDARY
RECONCILE_THEN_HANDOFF
WAIT_FOR_ORIGINAL_CAPACITY
UNSAFE
```

## 19.1 Continuity decision inputs

```text
source driver/runtime
source principal
source session domain
native session metadata
last durable turn state
provider failure classification
in-flight tool/action state
target driver/runtime
 target principal
 target session domain
execution-profile compatibility
host continuity policy
runtime conformance evidence
```

## 19.2 Safe replay rule

> **AgentPool never blindly replays an accepted turn.**

If the request failed before provider acceptance, rerouting the same turn is permitted.

If it was accepted, AgentPool must either:

- resume the provider's native session/turn using a proven mechanism;
- begin a new continuation turn from a confirmed safe boundary;
- run explicit reconciliation according to host/provider policy;
- or wait/block.

## 19.3 Tool and side-effect tracking

The normalized event stream records tool lifecycle where the provider exposes it:

```text
TOOL_STARTED
TOOL_COMPLETED
TOOL_FAILED
```

If execution disappears while a non-idempotent external action may be in flight, the result is not guessed.

Example:

```text
provider starts deploy
remote deployment succeeds
connection dies before acknowledgement
```

No local repository inspection can prove whether the deployment happened.

Strict mode becomes:

```text
RECOVERY_BLOCKED / OUTCOME_UNKNOWN
```

until a host/provider reconciliation strategy resolves it.

That is not reduced reliability. It is the only reliable behavior.

## 19.4 Strict automatic continuity

The recommended default is:

```text
continuity = strict-automatic
```

It automatically rotates whenever a proven safe continuity path exists and otherwise waits/blocks instead of risking duplicate effects.

A later opt-in mode can allow:

```text
continuity = reconcile-best-effort
```

for workloads whose host provides idempotency/reconciliation hooks.

---

# 20. Codex Cross-Account Continuation

Codex is the first provider where AgentPool should attempt true account-pool continuity.

The design target is:

```text
job-42
  │
  ├── attempt 1
  │     principal = OpenAI-A
  │     thread = T
  │
  │   included capacity exhausted
  │
  └── attempt 2
        principal = OpenAI-B
        thread = T
```

This feature is enabled only when all of these hold:

```text
same Codex session domain
compatible execution profile
compatible model request
exact runtime passes cross-principal-resume conformance
source turn is at a resumable boundary
OpenAI-B has eligible included usage
no conflicting active process owns thread T
```

If any condition fails, the scheduler chooses a handoff or waits.

This feature must have its own conformance test and can be disabled independently from ordinary Codex execution.

---

# 21. Claude and Antigravity Cross-Account Continuation

For the initial architecture:

```text
Claude account A -> Claude account B
```

is a provider-preserving **handoff**, not native resume.

Likewise:

```text
Antigravity account A -> Antigravity account B
```

is a handoff unless the exact runtime later advertises and passes a verified cross-profile resume capability.

The architecture loses nothing by being conservative here because the capability model can be upgraded later without changing the Job/Attempt model.

---

# 22. Cross-Provider Handoff

Cross-provider continuation is explicitly a new attempt with a provider-independent handoff envelope.

It is not represented as native session resume.

```text
Codex
  │
  ▼
HandoffEnvelope
  │
  ▼
Claude / Antigravity / future agent
```

## 22.1 HandoffEnvelope

```yaml
job_id: job-42
objective: "Implement X and verify Y"

workspace:
  identity: "..."
  root: "/projects/tradeagent"
  repository_head: "abc123"
  branch: "feature/x"
  dirty_state_fingerprint: "..."

execution:
  source_provider: codex
  source_model: "..."
  source_native_session: "..."
  last_confirmed_turn: "..."
  provider_event_cursor: 381

progress:
  host_checkpoint: <opaque or structured>
  agent_summary: <optional, provenance-marked>

observations:
  completed_tool_actions: [...]
  unresolved_actions: [...]
  tests_seen: [...]

continuation:
  instructions: "Inspect current state before changing anything..."
  provenance:
    - field: host_checkpoint
      source: HOST_ASSERTED
    - field: repository_head
      source: OBSERVED
```

## 22.2 Semantic ownership

When a sophisticated host provides a checkpoint, AgentPool stores and transports it without pretending to understand the host's semantics.

AgentPool's own generic summary is optional fallback material and is always provenance-marked. It never silently converts model assertions into facts.

---

# 23. Process Supervision and Crash Recovery

A durable Job does not mean the underlying provider subprocess magically survives every crash.

## 23.1 Process record

For every spawned provider process record:

```text
pid
process_start_identity/birth marker
process_group or Windows job object
binary canonical path
runtime_id
attempt_id
lease_generation
started_at
stdio/protocol kind
last_heartbeat/event time
```

PID alone is insufficient because PIDs are reused.

## 23.2 Process trees

Provider agents can spawn shells, language servers, test processes, MCP servers, browsers, and subagents.

Cancellation/cleanup must operate on the process group/job object:

```text
structured cancel
  -> grace interval
  -> terminate process group
  -> bounded wait
  -> force kill process group
  -> verify/report survivors
```

On Unix the ACP Rust SDK already provides useful process-group behavior for ACP child agents; non-ACP drivers need equivalent supervision.

## 23.3 Daemon restart reconciliation

At startup:

```text
1. open/recover SQLite
2. acquire singleton daemon lock
3. inspect non-terminal jobs/attempts
4. validate recorded process identities
5. revoke stale execution leases
6. determine whether each native session can be resumed safely
7. kill or quarantine orphan provider processes when ownership cannot be recovered
8. create replacement attempt only after fencing old authority
9. append recovery events
10. resume admission
```

For stdio-driven providers, a daemon crash normally destroys the control channel even if the OS process remains momentarily alive. AgentPool should not pretend it can reconstruct a lost stdio stream. It should terminate/reap the orphan and recover through provider-native session state at the last proven boundary.

If the daemon died during an ambiguous side effect, the job enters `RECOVERY_BLOCKED` rather than replaying.

---

# 24. Cancellation Semantics

Cancellation races with recovery and provider failover.

Therefore cancellation is a durable state transition, not just `kill(pid)`.

```text
transaction:
  job -> CANCELLING
  revoke active execution lease
  prevent new reservations
  emit JOB_CANCELLING
```

Then terminate active attempts/process groups.

Only after the adapter reports/observes termination does the job become:

```text
CANCELLED
```

If an external action's outcome is unknown, that uncertainty remains recorded even though the job is cancelled.

A stale recovery worker cannot restart a job after cancellation because its lease generation is fenced.

---

# 25. Permissions and Human Input

Provider agents can require:

- tool approval;
- user choice;
- authentication;
- permission escalation;
- interactive confirmation.

The normalized runtime therefore distinguishes:

```text
WAITING_FOR_INPUT
WAITING_FOR_PERMISSION
```

from ordinary `RUNNING`.

## 25.1 PermissionPolicy

Host requests a normalized policy such as:

```text
supervised
autonomous-safe
full-access
custom
```

Each driver maps that policy to native semantics.

If the provider cannot faithfully represent a required policy, the adapter returns:

```text
POLICY_UNREPRESENTABLE
```

It does not silently pick something “close enough.”

For ACP clients, permission requests are forwarded to the ACP client when appropriate. For autonomous jobs, the configured host policy answers only requests that policy explicitly allows.

---

# 26. Model and Capability Selection

Model selection is not a plain string equality check.

A request carries:

```yaml
model:
  requested: "gpt-..."
  policy: exact | prefer | provider_default

required_capabilities:
  - shell
  - mcp
  - image_input
  - native_resume
```

The adapter returns the effective model and observed capabilities.

If `policy = exact` and the provider uses a different effective model:

```text
attempt -> FAILED_CONFIGURATION / MODEL_MISMATCH
```

rather than silently continuing.

For ACP clients, model choices are advertised through session config options where the negotiated ACP version/client supports them. This allows `agentpool acp --pool codex` to expose models without inventing a proprietary T3 extension.

The model catalog is refreshed per account because entitlements can differ.

---

# 27. Control-Plane API

A deliberately small v1 API is enough.

## 27.1 Handshake

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "system.hello",
  "params": {
    "protocol_major": 1,
    "protocol_minor": 0,
    "client": {"name": "tradeagent", "version": "..."}
  }
}
```

The daemon responds with its version and supported capabilities.

Major incompatibility fails explicitly.

## 27.2 Job creation

```json
{
  "method": "job.create",
  "params": {
    "idempotency_key": "tradeagent/unit-381/builder-1",
    "pool": "codex",
    "workspace": "/projects/tradeagent-wt-381",
    "task": "Implement unit 381 and run required verification",
    "model": {"requested": "gpt-5.6-sol", "policy": "prefer"},
    "continuity": "strict-automatic",
    "billing": "subscription-included-only",
    "priority": "normal"
  }
}
```

Response:

```json
{
  "job_id": "job-913",
  "revision": 1,
  "state": "queued"
}
```

Repeating the same `idempotency_key` returns the same logical job rather than creating a duplicate.

## 27.3 Essential methods

```text
system.hello
system.status
job.create
job.get
job.cancel
job.submit_turn
job.respond_input
job.attach_checkpoint
job.list
pool.list
pool.get
slot.list
slot.get
slot.enable
slot.disable
auth.begin
auth.complete
auth.logout
runtime.list
runtime.doctor
events.subscribe
events.read
```

Administrative mutation methods require the local client's appropriate capability grant.

## 27.4 Event replay

Every durable job event has:

```text
event_id
job_id
job_sequence
attempt_id?
turn_id?
timestamp_utc
kind
schema_version
payload
```

Clients reconnect with:

```text
events.subscribe(job_id, after_sequence = 124)
```

Delivery is **at least once**. Clients deduplicate by sequence/event ID.

---

# 28. Event Model

Important normalized events include:

```text
JOB_CREATED
JOB_QUEUED
ATTEMPT_RESERVED
ATTEMPT_STARTED
PRINCIPAL_SELECTED
NATIVE_SESSION_CREATED
NATIVE_SESSION_RESUMED
TURN_PREPARED
TURN_ACCEPTED
TURN_RUNNING
AGENT_MESSAGE
TOOL_STARTED
TOOL_COMPLETED
TOOL_FAILED
PERMISSION_REQUIRED
INPUT_REQUIRED
QUOTA_OBSERVED
CAPACITY_EXHAUSTED
ATTEMPT_ENDED
CONTINUITY_PLAN_SELECTED
JOB_WAITING_FOR_CAPACITY
JOB_RECOVERY_BLOCKED
JOB_RESUMED
CHECKPOINT_UPDATED
JOB_COMPLETED
JOB_FAILED
JOB_CANCELLED
```

Only state-changing/operationally meaningful events must be persisted durably.

Token-by-token model deltas do not belong in the primary SQLite event journal. They may be streamed live and optionally stored in bounded diagnostic traces.

---

# 29. SQLite Storage

SQLite remains the correct database for a single-machine daemon.

Use a bundled/tested SQLite build through `rusqlite` so behavior does not vary with an old system SQLite.

Recommended startup pragmas:

```text
journal_mode = WAL
synchronous = FULL
foreign_keys = ON
busy_timeout = bounded value
```

`FULL` is intentionally conservative: AgentPool writes little data compared with a normal application database, and durable routing/lease decisions matter more than maximizing transaction throughput.

WAL is appropriate because all authoritative state is local to one host; AgentPool must reject placing the database on a network filesystem.

## 29.1 Tables

```text
schema_migrations
meta

provider_runtimes
runtime_capabilities
runtime_leases

principals
auth_profiles
execution_profiles
session_domains
slots

pools
pool_members

jobs
attempts
execution_leases
native_sessions
turns
workspace_leases

quota_observations
checkpoints

events
idempotency_keys
```

Raw credential material is absent.

## 29.2 Transaction rule

Every externally visible state transition and its event are committed in the same SQLite transaction.

Clients never receive:

```text
JOB_RUNNING
```

unless the authoritative state supporting that event is already durable.

## 29.3 Event compaction

Durable semantic events remain inspectable, but old high-volume diagnostic material is rotated separately. Periodic job snapshots can accelerate recovery without deleting the authoritative sequence needed for audit/replay policy.

---

# 30. Runtime Installation and Upgrade System

Upgradeability is a core subsystem, not “we'll update dependencies later.”

## 30.1 Two runtime modes

### Managed

AgentPool installs provider runtime artifacts under:

```text
~/.agentpool/runtimes/<driver>/<version-or-fingerprint>/
```

Each installation is immutable.

### External

The user supplies a canonical binary path. AgentPool never overwrites it and treats version changes as externally managed.

## 30.2 Side-by-side upgrades

For managed runtimes:

```text
install new version beside old
verify archive/hash/signature metadata where available
run conformance suite
mark eligible
atomically move "current for new attempts" pointer
keep old runtime while any runtime lease exists
remove old runtime only after all leases release
```

A running agent never has its executable replaced underneath it.

## 30.3 Rollback

The previous known-good runtime remains available until the new runtime has passed conformance and a configurable burn-in period.

If new attempts begin failing a provider-specific invariant:

```text
new runtime -> QUARANTINED
current pointer -> previous known-good
existing healthy processes on old/new immutable binaries continue according to policy
```

## 30.4 Compatibility manifest

AgentPool ships a data-driven manifest containing:

```text
driver
runtime/version ranges
known good
known bad
required quirks
capability overrides
minimum conformance suite version
```

The manifest is not hard-coded through scattered `if provider_version` statements.

A later signed manifest update channel can refresh compatibility knowledge independently of a full AgentPool binary release.

## 30.5 Forward-compatible parsing

Provider schemas evolve quickly. The parser follows this rule:

```text
known discriminator + malformed required fields -> error
unknown future enum/field with safe fallback    -> preserve/ignore safely
unknown lifecycle semantic required for safety  -> degrade/block capability
```

New plan names must not crash a provider merely because a closed enum in AgentPool is stale.

---

# 31. `agentpool doctor`

Diagnostics are a first-class product surface.

```bash
agentpool doctor
agentpool doctor --provider codex
agentpool doctor --slot openai-02
agentpool doctor --deep
```

Normal doctor is non-destructive and should not:

- open login browsers;
- consume model quota;
- start MCP servers unnecessarily;
- mutate provider settings;
- create throwaway sessions merely to prove existence.

It checks layered capability levels:

```text
binary present
binary fingerprint/version
protocol initializes
required protocol methods/capabilities
profile paths valid
credential metadata available
principal evidence matches
model catalog readable where non-consuming
capacity endpoint readable where non-consuming
runtime compatibility manifest
filesystem permissions
workspace roots
SQLite integrity
stale process/lease records
```

`--deep` may perform an explicit low-cost inference/conformance test after warning the user that it can consume plan usage.

The result is stored against the exact runtime/profile fingerprint so repeated safe probes are unnecessary.

---

# 32. Provider Failure Classification

`classifyFailure(): FailureReason` is insufficient. The normalized result needs recovery semantics.

```yaml
FailureClassification:
  reason: RATE_LIMIT | AUTH | NETWORK | PROVIDER_OVERLOAD | ...
  scope: principal | model | provider | runtime | turn | unknown
  retryability: immediate | backoff | after_reset | reauth | never | unknown
  retry_after: optional
  continuity_safety: safe_before_accept | safe_boundary | ambiguous | unsafe
  confidence: authoritative | strong | weak
```

Examples:

```text
subscription limit reached
  scope = principal/app
  retryability = after_reset or another eligible principal

provider overload
  scope = provider
  retryability = bounded backoff

network interruption before turn acceptance
  scope = attempt
  retryability = reroute-safe

auth refresh permanently invalid
  scope = principal
  retryability = reauth

unsupported runtime protocol
  scope = runtime
  retryability = never until upgrade/downgrade

unknown failure after tool start
  scope = turn
  continuity_safety = ambiguous
```

Unknown errors never trigger automatic “capacity rotation” merely because they contain `429` or the word `limit`.

---

# 33. Quarantine

A repeatedly unsafe slot/runtime should stop cycling through admission.

Quarantine triggers can include:

```text
principal mismatch
credential/billing ambiguity
protocol corruption
unsupported schema/lifecycle change
repeated unexplained process crashes
session-domain inconsistency
failed conformance invariant
known-bad runtime manifest entry
```

Quarantine is sticky until:

- the user fixes/re-authenticates;
- the runtime changes;
- the compatibility manifest explicitly resolves it;
- or an explicit diagnostic action clears it after successful verification.

---

# 34. Observability Without Building a Secret Archive

AgentPool must explain every important routing decision.

Human-readable log example:

```text
18:01:04 job-42 created pool=codex
18:01:04 selected slot=openai-01 continuity=native_same_principal
18:01:05 attempt=91 started runtime=codex/0.x fingerprint=...
19:43:17 quota exhausted slot=openai-01 limit=codex reset=22:00
19:43:17 continuity candidate openai-02 = native_cross_principal
19:43:18 lease generation 12 -> 13
19:43:18 native thread resumed on openai-02
21:10:42 job-42 completed
```

But logs must not persist everything the agent sees.

Default durable logging:

```text
metadata + state transitions + redacted diagnostics
```

Raw provider transcript/tool payload retention:

```text
off by default
bounded if enabled
access-controlled
rotated
redacted where possible
```

Secrets can leak through shell output, `.env` files, MCP responses, command arguments, and provider diagnostics even if OAuth tokens themselves are perfectly redacted.

---

# 35. Local Security Model

The daemon is effectively a privileged code-execution broker for the current OS user. Treat its API accordingly.

## 35.1 IPC

Default:

```text
Unix socket mode 0600
Windows named pipe restricted to current user
```

No unauthenticated TCP listener.

## 35.2 Client capabilities

A local client registration can be granted scopes such as:

```text
jobs:read
jobs:write
slots:read
slots:admin
auth:admin
runtimes:admin
```

TradeAgent does not need permission to export OAuth credentials merely because it can submit a job.

## 35.3 Workspace authorization

Clients are restricted to configured workspace roots.

Before dispatch:

```text
canonicalize path
resolve symlinks/reparse points
verify target is inside an allowed root
verify workspace policy
```

A malicious local client cannot turn:

```text
workspace = ~/.ssh
```

into arbitrary agent access unless the user explicitly granted that root.

## 35.4 Process invocation

Always spawn:

```text
binary_path + argv[]
shell = false
```

Never interpolate task text, model names, paths, or session IDs into a shell command string.

## 35.5 Binary identity

Managed binaries are launched by canonical stored path. External binaries are resolved/canonicalized during configuration/probe rather than trusting a mutable `PATH` at every execution.

---

# 36. T3 Code Integration

Current T3 architecture makes this much cleaner than the original paper assumed.

T3 supports generic ACP providers from the registry or a local command. Therefore AgentPool can expose each logical pool directly:

```text
AgentPool / Codex
AgentPool / Claude
AgentPool / Google
AgentPool / Best Coder
```

Conceptually the local T3 provider is:

```text
source: local
command: agentpool
args: ["acp", "--pool", "codex"]
```

No T3 source patch is required for the basic integration.

The ACP facade advertises model/config choices through ACP config options, allowing a compatible client model picker to change model selection without exposing account slots.

T3 sees one stable ACP session while AgentPool may internally move:

```text
OpenAI account 1
       ↓
OpenAI account 2
```

The slot switch appears as execution metadata/events, not as a new T3 conversation.

---

# 37. TradeAgent Integration

TradeAgent should use the native control-plane API, not ACP, because it benefits from durable job/lease/capacity semantics.

```text
TradeAgent manager
      │
      │ job.create(pool="codex", ...)
      ▼
AgentPool
      │
      ├── slot selection
      ├── runtime process
      ├── capacity rotation
      ├── wait/resume
      └── durable event stream
```

TradeAgent retains:

```text
managerial hierarchy
task decomposition
verification policy
context-cleansing / JEV logic
semantic checkpoints
project-specific recovery
```

AgentPool retains:

```text
execution attempt ownership
accounts/capacity
provider lifecycle
native sessions
safe continuity
```

TradeAgent can attach an opaque semantic checkpoint to a job. AgentPool can include it in any later handoff without taking ownership of TradeAgent's reasoning architecture.

---

# 38. AI Concierge Integration

AI Concierge uses the same control API but can request capability-oriented pools:

```text
pool = best-coder
required_capabilities = [shell, repository_write]
continuity = strict-automatic
```

The Concierge does not care whether execution uses:

```text
Codex account A -> Codex account B
```

or, when explicitly allowed:

```text
Codex -> handoff -> Claude
```

It observes one durable Job.

---

# 39. CLI Surface

The CLI should remain small and unsurprising.

```bash
agentpool status
agentpool pool list
agentpool slot list
agentpool slot add codex
agentpool slot add claude
agentpool slot login <slot>
agentpool slot disable <slot>

agentpool run codex
agentpool run codex --model <model>
agentpool submit codex --workspace . --task-file task.md
agentpool attach job-42
agentpool detach
agentpool job job-42
agentpool cancel job-42

agentpool doctor
agentpool runtime list
agentpool runtime verify <runtime>

agentpool acp --pool codex
```

`run` is interactive. `submit` is durable/autonomous. They are different execution modes even though they share the same core Job machinery.

---

# 40. Configuration Shape

Configuration is declarative; secrets are referenced, not embedded.

```yaml
pools:
  codex:
    candidates:
      - slot: openai-pro-1
      - slot: openai-pro-2
      - slot: openai-pro-3
    routing: capacity-aware-sticky
    continuity: strict-automatic
    billing: subscription-included-only

  best-coder:
    candidates:
      - slot: openai-pro-1
      - slot: openai-pro-2
      - slot: claude-max-1
      - slot: google-ultra-1
    routing: ordered-capability-aware
    continuity:
      same_session: automatic
      handoff: allowed
      cross_provider: allowed
    billing: subscription-included-only

slots:
  openai-pro-1:
    driver: codex-app-server
    runtime: codex-managed
    principal: openai-principal-1
    auth: openai-siwc-1
    execution_profile: coding-default
    session_domain: codex-shared

  openai-pro-2:
    driver: codex-app-server
    runtime: codex-managed
    principal: openai-principal-2
    auth: openai-siwc-2
    execution_profile: coding-default
    session_domain: codex-shared

  claude-max-1:
    driver: claude-stream-json
    runtime: claude-system
    principal: claude-principal-1
    auth: claude-profile-1
    execution_profile: coding-default
    session_domain: claude-profile-1
```

The YAML is user configuration; authoritative runtime state remains in SQLite.

---

# 41. Testing Strategy

AgentPool should be tested more like a database/process supervisor than like a normal CLI utility.

## 41.1 Deterministic fake provider

Build a first-party fake provider capable of scripting:

```text
auth success/failure
principal mismatch
quota exhaustion before turn acceptance
quota exhaustion after acceptance
transient 429
provider overload
structured reset time
no reset time
process crash
stdout corruption
missing completion event
duplicate events
out-of-order notifications
token refresh request
session resume success/rejection
tool starts then connection disappears
tool completes then process disappears
runtime version changes capability
permission request
human-input request
```

No core recovery behavior should require burning real subscription quota to test.

## 41.2 Property/state-machine tests

Continuously assert invariants such as:

1. At most one authoritative execution lease exists per job.
2. At most one exclusive AgentPool workspace writer exists per workspace identity.
3. A cancelled job never becomes running again without an explicit new job.
4. A stale lease generation can never commit a state transition.
5. An accepted ambiguous turn is never blindly replayed in strict mode.
6. Unknown failures are never automatically reclassified as capacity exhaustion.
7. A slot never launches with another slot's credential reference.
8. Billing mode never changes without policy permission.
9. Cross-principal native resume occurs only when the exact runtime capability permits it.
10. Every emitted durable event corresponds to committed database state.
11. Every job event sequence is monotonic.
12. Runtime removal never deletes a binary with a live runtime lease.

## 41.3 Crash/fault injection

Kill the daemon at every important boundary:

```text
before DB transition
immediately after DB transition
before process spawn
after process spawn before PID record
after turn prepared
immediately after provider acceptance
during tool execution
before completion event persistence
during capacity update
during cancellation
during auth refresh
```

Restart and assert the exact expected durable state.

## 41.4 Provider protocol fixtures

Store redacted golden fixtures for supported versions of:

```text
Codex app-server
Claude stream-json
ACP agents
```

Fixtures include unknown future fields and enums to prove tolerant parsing.

## 41.5 Real-provider conformance suite

For each supported runtime version:

```text
initialize
identity/auth mode
model catalog
one safe inference
native session creation
native session resume
cancel
permission roundtrip where supported
quota read where supported
process termination
restart/resume
```

For multi-account Codex specifically:

```text
create thread under account A
complete turn
terminate process
launch app-server under account B in same session domain
resume exact thread
continue turn
verify account B is the active principal
verify thread history continuity
```

This test is what grants the runtime capability:

```text
native_resume_cross_principal = true
```

not a provider-name assumption.

## 41.6 Known-bad regression tests

Every real provider breakage discovered in production becomes a fixture/test before the workaround is considered complete.

Examples worth explicitly covering from current ecosystem history:

- new Codex plan enum should not break parsing;
- new Codex home files should not break authentication/session isolation;
- Claude auth metadata may disagree with actual credential path;
- ambient API keys must not override a subscription slot;
- an Antigravity process/profile must not inherit ambient Google credentials;
- provider runtime packaging can change or leak temporary directories;
- missing quota telemetry must degrade to `UNKNOWN`, not failure.

---

# 42. MVP: What to Build First

The MVP should remain extremely focused, but its foundations must be final-quality.

## 42.1 MVP proof

> **A durable Codex job can use a pool of independently authenticated ChatGPT accounts, continue through another eligible account when the current account's included usage becomes unavailable and the native thread boundary is proven safe, wait durably when no eligible account exists, and survive AgentPool restart without duplicate authoritative execution.**

## 42.2 Required MVP components

### Core

- Rust daemon and CLI;
- local Unix-socket / Windows-pipe control plane;
- SQLite WAL storage and migrations;
- Job / Attempt / Turn / Lease model;
- idempotent job creation;
- event journal and replay;
- workspace leases;
- deterministic scheduler;
- capacity observations;
- crash recovery;
- process-group/job-object supervision;
- quarantine and diagnostics.

### Codex

- OpenAI Sign in with ChatGPT OSS registration flow;
- multiple separate saved principals;
- protected token storage;
- serialized token refresh;
- Codex app-server structured client;
- account identity verification;
- account rate-limit observations;
- thread start/resume;
- cross-principal same-thread conformance gate;
- exact/prioritized model policy;
- included-only billing policy.

### Client surfaces

- `agentpool run`;
- `agentpool submit`;
- status/job/event CLI;
- control API;
- `agentpool acp --pool codex` using stable ACP v1.

### Tests

- deterministic fake provider;
- crash injection;
- core property tests;
- Codex protocol fixtures;
- real two-account Codex conformance scenario.

## 42.3 Explicitly not in MVP

- Claude execution;
- Antigravity execution;
- cross-provider handoff;
- automatic Git worktree creation;
- GUI;
- remote networking;
- distributed scheduling;
- dynamic third-party binary plugin ABI;
- speculative quota prediction;
- cost optimization engine.

These are designed for, not implemented prematurely.

---

# 43. Phase 2 — Claude and Generic ACP

After Codex pool continuity is trustworthy:

1. add `claude-stream-json` driver;
2. add generic inbound ACP driver;
3. add official Antigravity ACP as the first generic ACP-backed subscription agent;
4. add profile isolation and provider-specific billing guards;
5. expose all pools through the same control API and ACP facade;
6. validate provider-specific capacity classification;
7. keep cross-account Claude/Antigravity continuity as handoff/wait until a native-resume capability is proven.

---

# 44. Phase 3 — Provider-Independent Handoff

Implement `HandoffEnvelope` and host checkpoint hooks.

Order:

```text
same native session
  > provider-preserving safe handoff
  > cross-provider handoff
  > wait
```

TradeAgent and other sophisticated hosts can provide semantic checkpoints. AgentPool's generic fallback remains evidence/provenance-aware.

---

# 45. Phase 4 — Operational Polish

Only after the execution core is proven:

- managed provider runtime installers where valuable;
- signed compatibility-manifest refresh;
- automatic safe rollback to known-good runtime;
- richer capacity dashboards;
- optional AgentPool GUI;
- remote control through an authenticated transport without moving execution off the owning machine;
- optional advanced fair scheduling;
- optional generic ACP Registry onboarding for new providers.

---

# 46. Non-Negotiable Reliability Invariants

These are testable product requirements, not aspirations.

### R1 — One authoritative attempt

A job has at most one current execution lease generation.

### R2 — One protected writer

An exclusive workspace identity has at most one AgentPool write lease.

### R3 — No credential crossover

A slot never runs with another slot's auth material or undeclared ambient billing credentials.

### R4 — Principal verification

A strongly detected principal mismatch blocks execution.

### R5 — No blind replay

AgentPool never automatically repeats an accepted ambiguous turn in strict mode.

### R6 — Capacity is recoverable

Confirmed subscription-capacity exhaustion is not a terminal job failure while a safe eligible continuation or future reset exists.

### R7 — Exhaustion is durable

When every eligible slot is unavailable, the job waits durably.

### R8 — Recovery is automatic when safe

Waiting jobs are reconsidered when capacity observations expire/reset or a slot becomes eligible.

### R9 — Unknown is not exhausted

An unknown/provider/network failure is never relabeled as quota exhaustion without evidence.

### R10 — Billing path does not silently change

Included subscription execution never falls through to metered API/credits unless policy explicitly allows it.

### R11 — Runtime version is immutable per process

A live process keeps the exact executable version/fingerprint with which it started.

### R12 — Capability is evidence-backed

Cross-principal resume, quota introspection, model entitlement, and similar advanced features are enabled only by negotiated/tested capability evidence.

### R13 — Durable before visible

A durable event is emitted only after the corresponding state transaction commits.

### R14 — Event replay is deterministic

Clients can reconnect after sequence N and recover every later durable event at least once.

### R15 — Cancellation is fenced

Once a job's cancellation revokes its execution lease, no stale attempt can bring it back to running.

### R16 — Provider update cannot silently break the core

Unknown runtime changes either pass conformance, degrade optional capabilities, or quarantine. They do not become assumed-safe behavior.

### R17 — Account count is data

Adding another compatible paid account requires configuration/authentication, not application code changes.

### R18 — Hosts remain independent

TradeAgent, Concierge, T3, scripts, and humans consume the same execution primitive without embedding account-rotation logic.

### R19 — Provider specifics stay at the boundary

Core scheduling and job state do not contain Codex/Claude/Antigravity authentication branches.

### R20 — Local-first

Normal operation requires no AgentPool cloud service.

---

# 47. Acceptance Scenarios

## A. Normal Codex execution

```text
job created
-> slot A reserved
-> principal verified
-> app-server started
-> native thread created
-> turn accepted
-> turn completed
-> job completed
```

## B. Codex account A reaches included limit at a safe boundary

```text
job-42 / thread T / account A
-> authoritative capacity event
-> account A becomes ineligible for new work
-> continuity planner verifies account B + same session domain + runtime capability
-> execution lease generation increments
-> app-server starts with account B token
-> thread T resumes
-> job-42 remains the same job
```

## C. All Codex accounts unavailable

```text
A exhausted until 22:00
B exhausted until 21:10
C capacity unknown / probe unavailable
-> C receives bounded eligibility attempt if appropriate
-> no safe slot
-> job WAITING_FOR_CAPACITY
-> durable wake scheduled near 21:10 with jitter
-> retry/resume
```

## D. Provider process dies before turn acceptance

```text
turn PREPARED
process dies before ACCEPTED
-> same input may be dispatched to replacement attempt
```

## E. Process dies after unknown external action

```text
tool started
acknowledgement not observed
process/daemon disappears
-> outcome unknown
-> strict mode RECOVERY_BLOCKED
-> no blind replay
```

## F. Daemon crashes

```text
DB already contains job/attempt/lease/last durable event
-> daemon restarts
-> stale process ownership reconciled
-> old lease fenced
-> native session resumed only at proven boundary
-> duplicate authoritative attempt impossible
```

## G. Provider update introduces unknown plan enum

```text
parser receives new plan string
-> preserved as unknown/display value
-> runtime does not crash merely because enum is new
-> safety-critical capabilities remain independently probed
```

## H. Claude slot contains ambient API key

```text
slot expects subscription
-> launch environment builder detects conflicting billing variable
-> variable removed or configuration rejected
-> no silent API billing
```

## I. Google quota telemetry absent

```text
Antigravity slot healthy
capacity snapshot = UNKNOWN
-> slot remains eligible
-> actual turn succeeds: observation improves
or
-> provider reports quota failure: capacity becomes exhausted/retry observation
```

---

# 48. The Architecture in One Diagram

```text
                                ┌─────────────────────────────┐
                                │          CLIENTS            │
                                │ TradeAgent / Concierge / CLI│
                                └──────────────┬──────────────┘
                                               │
                                      Control plane RPC
                                               │
                                               ▼
┌───────────────┐                    ┌────────────────────────────┐
│ T3 / ACP IDEs │── ACP stdio ─────►│     agentpool acp shim     │
└───────────────┘                    └──────────────┬─────────────┘
                                                   │
                                                   ▼
                                    ┌────────────────────────────┐
                                    │      AGENTPOOL DAEMON      │
                                    │                            │
                                    │ jobs / attempts / leases   │
                                    │ scheduler / continuity     │
                                    │ quota observations         │
                                    │ process supervision        │
                                    │ runtime compatibility      │
                                    │ event journal              │
                                    └───────────┬────────────────┘
                                                │
                                             SQLite
                                                │
                 ┌──────────────────────────────┼─────────────────────────────┐
                 │                              │                             │
                 ▼                              ▼                             ▼
       ┌───────────────────┐          ┌───────────────────┐        ┌────────────────────┐
       │ Codex app-server  │          │ Claude stream-json│        │ Generic ACP client │
       └─────────┬─────────┘          └─────────┬─────────┘        └──────────┬─────────┘
                 │                              │                             │
       ┌─────────┴──────────┐        ┌──────────┴──────────┐                  ▼
       │ SIWC principals   │        │ CLAUDE_CONFIG_DIRs │        Google Antigravity ACP
       │ account A/B/C/... │        │ account A/B/C/...  │        future ACP providers
       └───────────────────┘        └─────────────────────┘
```

---

# 49. Why This Design Stays Upgradeable

The project is future-proof not because it predicts provider behavior, but because it **refuses to encode provider behavior in the wrong layer**.

When OpenAI changes a rate-limit field:

```text
Codex adapter / compatibility manifest changes
core scheduler does not
```

When Anthropic changes auth storage:

```text
Claude auth/profile adapter changes
Job/Attempt/Lease model does not
```

When Google ships a new ACP version:

```text
generic ACP driver negotiates/adapts
pool and job API do not
```

When T3 moves from ACP v1 to v2:

```text
ACP facade gains negotiated v2 surface
AgentPool control plane does not
```

When a fourth provider appears:

```text
new built-in driver
or generic ACP registration
```

When a provider stops supporting cross-account native resume:

```text
capability flips false
continuity planner selects handoff/wait
```

Nothing above the capability boundary needs to be rewritten.

---

# 50. What “Perfect” Means Here

AgentPool cannot make external provider services infallible, and it cannot make arbitrary irreversible external actions mathematically exactly-once when the external system supplies no idempotency/reconciliation primitive.

The robust definition is stronger and more useful:

> **AgentPool must never invent certainty. It automatically continues whenever the provider/runtime gives it a provably safe path; when certainty is unavailable, it preserves the job and refuses the unsafe action.**

That produces a system that can run unattended without the dangerous failure mode of “availability at any cost.”

For the normal capacity event the project exists to solve, the path should be invisible to the host:

```text
paid account A exhausted
        ↓
eligible paid account B selected
        ↓
native session resumed when capability allows
        ↓
same AgentPool job continues
```

For a provider where native cross-account state is impossible:

```text
account A exhausted
        ↓
safe-boundary checkpoint / handoff
        ↓
account B or provider B
        ↓
same AgentPool job continues
```

And only when the state is genuinely ambiguous:

```text
preserve + block
```

instead of silently doing the wrong thing.

---

# 51. Recommended Build Order

The highest-quality sequence is:

```text
1. domain model + invariants
2. SQLite transitions/event journal
3. fake provider + fault injector
4. leases/workspace leases/process supervisor
5. deterministic scheduler/capacity model
6. control-plane API
7. Codex SIWC account manager
8. Codex app-server driver
9. two-account native-continuity conformance
10. daemon crash/restart recovery
11. ACP outward facade
12. real T3 integration
13. hardening / known-bad compatibility manifest
14. Claude driver
15. generic inbound ACP + Antigravity
16. cross-provider handoff
```

Do **not** start by implementing three providers. If the first six layers are wrong, every adapter will have to compensate differently and the architecture will decay immediately.

---

# 52. Source and Evidence Baseline

This architecture was researched against current public documentation and current implementation evidence on 2026-10-06. Provider behavior changes quickly; these sources establish the design baseline, while AgentPool's runtime conformance system is responsible for detecting drift after implementation.

## OpenAI / Codex — primary sources

- OpenAI, **Sign in with ChatGPT — open-source plan usage overview**  
  https://developers.openai.com/siwc/token-sharing-open-source
- OpenAI, **Registration and sign-in**  
  https://developers.openai.com/siwc/token-sharing-open-source/sign-in
- OpenAI, **Accounts and sessions** — explicitly documents multiple saved ChatGPT account registrations and switching accounts  
  https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions
- OpenAI, **Codex app-server — Sign in with ChatGPT**  
  https://developers.openai.com/siwc/token-sharing-open-source/codex-app-server
- OpenAI, **Errors and recovery**  
  https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery
- OpenAI, **Preview limitations**  
  https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations
- OpenAI Help, **Using your ChatGPT plan in other apps and sites**  
  https://help.openai.com/en/articles/20001542-using-your-chatgpt-plan-in-other-apps-and-sites
- OpenAI Help, **Using Codex with your ChatGPT plan**  
  https://help.openai.com/en/articles/11369540-using-codex-with-your-chatgpt-plan
- OpenAI Codex source, current app-server account/rate-limit protocol schemas  
  https://github.com/openai/codex/tree/main/codex-rs/app-server-protocol

## Anthropic / Claude Code — primary sources

- Anthropic Help, **Use the Claude Agent SDK with your Claude plan** — current June 2026 update states Agent SDK, `claude -p`, and third-party app usage still draw from subscription usage  
  https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan
- Claude Code docs, **Environment variables** — `CLAUDE_CONFIG_DIR`, API-key precedence, multiple-account isolation  
  https://code.claude.com/docs/en/env-vars
- Claude Code docs, CLI/structured execution references  
  https://code.claude.com/docs/en/cli-reference
- Claude Code repository/issues are used only as implementation-risk evidence for auth/status and rate-limit edge cases, not as normative protocol guarantees  
  https://github.com/anthropics/claude-code

## Google / Antigravity — primary sources

- Google Developers Blog, **Transitioning Gemini CLI to Antigravity CLI**  
  https://developers.googleblog.com/an-important-update-transitioning-gemini-cli-to-antigravity-cli/
- Official ACP Registry entry for Google Antigravity  
  https://github.com/agentclientprotocol/registry/tree/main/antigravity-acp
- Google Antigravity CLI repository  
  https://github.com/google-antigravity/antigravity-cli

## ACP — primary sources

- Agent Client Protocol repository and protocol documentation  
  https://github.com/agentclientprotocol/agent-client-protocol
- ACP Rust SDK  
  https://github.com/agentclientprotocol/rust-sdk
- Current Rust SDK documentation  
  https://docs.rs/agent-client-protocol/latest/agent_client_protocol/

As of the research cut, the Rust SDK exposes stable ACP v1 by default while protocol v2 remains separately gated behind an unstable feature. AgentPool should therefore dual-structure the ACP boundary and negotiate rather than prematurely dropping v1.

## T3 Code — current implementation evidence

T3 is not treated as a provider specification; it is valuable evidence that several required integration patterns already work in a real multi-provider agent harness.

- T3 provider architecture constraints  
  https://github.com/pingdotgg/t3code/blob/main/docs/internals/providers.md
- T3 Codex multi-account/shared-session setup  
  https://github.com/pingdotgg/t3code/blob/main/docs/user/providers-codex.md
- T3 Claude separate-account setup  
  https://github.com/pingdotgg/t3code/blob/main/docs/user/providers-claude.md
- T3 Antigravity provider/account/runtime behavior  
  https://github.com/pingdotgg/t3code/blob/main/docs/user/providers-antigravity.md
- T3 provider settings, including generic local ACP provider support  
  https://github.com/pingdotgg/t3code/blob/main/packages/contracts/src/settings.ts

## Storage/runtime references

- SQLite, **Write-Ahead Logging**  
  https://sqlite.org/wal.html
- Rust `rusqlite` current documentation  
  https://docs.rs/rusqlite/latest/rusqlite/

---

# 53. Scope Anchor

Every proposed feature should be judged against this sentence:

> **AgentPool turns separately authenticated agent accounts into reliable local execution pools, preserving one durable job across account/runtime capacity changes whenever continuity is provably safe, while keeping provider behavior behind replaceable capability-driven adapters.**

If a feature does not strengthen account isolation, execution continuity, durability, provider compatibility, observability, or safe upgradeability around that definition, it probably belongs somewhere else.

