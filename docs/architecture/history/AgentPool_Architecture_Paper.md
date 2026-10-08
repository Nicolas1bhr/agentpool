# AgentPool — Local Subscription-Aware AI Agent Switcher

**Status:** Concept / architecture paper  
**Purpose:** Define a deliberately small local program that turns multiple authenticated AI subscriptions and terminal-agent installations into reliable, reusable execution pools for humans, developer tools, and autonomous software.

---

## 1. Executive Summary

AgentPool is a small, local-first **AI agent execution broker**.

Its job is not to replace Codex, Claude Code, Gemini CLI, T3 Code, or the model providers themselves. Its job is to sit between software that needs an AI agent and the native terminal agents already installed and authenticated on the machine.

The user may have any number of subscriptions or authenticated accounts:

- OpenAI / Codex accounts
- Anthropic / Claude accounts
- Google / Gemini accounts
- future compatible terminal-agent providers

Each authenticated account becomes a **slot**. Slots are grouped into logical **pools**. Software asks AgentPool for a capability or pool rather than a particular account.

AgentPool selects a healthy slot, launches or resumes the native agent, watches its execution, detects capacity/rate-limit failures, moves unavailable slots into cooldown, and automatically continues through another eligible slot whenever safe.

The intended experience is:

```text
Software asks for "codex"
        │
        ▼
     AgentPool
        │
        ├── OpenAI slot 1 ── exhausted
        ├── OpenAI slot 2 ── selected automatically
        └── OpenAI slot 3 ── reserve
        │
        ▼
native Codex process
```

The calling software does not need to know which subscription is executing the work.

This same interface should work for an interactive developer using a terminal or T3, and for software such as **TradeAgent**, **AI Concierge**, or another autonomous system that needs to launch terminal agents programmatically.

The design principle is:

> **Pool subscriptions and terminal-agent capacity without reimplementing the agents themselves.**

---

## 2. The Problem

Modern coding and autonomous-agent workflows increasingly use subscription-backed terminal agents rather than raw model API calls.

A heavy user may legitimately have several paid subscriptions because one subscription does not provide enough interactive capacity for long-running work. Today, however, that capacity is fragmented.

A typical workflow looks like:

```text
Long-running job
      │
      ▼
Subscription A
      │
   limit hit
      │
      ▼
job interrupted
      │
      ▼
human logs out
      │
      ▼
human logs into Subscription B
      │
      ▼
tries to resume work
```

This has several problems:

1. Capacity exists but is not pooled.
2. Authentication switching requires human intervention.
3. Long-running agents can stop merely because one account reached a temporary limit.
4. Applications become coupled to specific provider accounts.
5. Every project tends to reinvent provider selection and recovery.
6. Multi-provider capacity cannot be expressed through one simple execution contract.
7. Account rotation can become unsafe if implemented by blindly replacing credential files.
8. Cross-provider continuation is fundamentally different from same-provider account rotation and must not be treated as identical.

The problem is therefore not merely authentication.

It is **execution continuity across independently authenticated agent capacity**.

---

## 3. Product Definition

AgentPool is a local daemon and CLI that provides:

- unlimited logical account slots;
- multiple provider adapters;
- logical execution pools;
- automatic slot selection;
- capacity and cooldown tracking;
- automatic same-provider failover where supported;
- safe job waiting when all eligible slots are unavailable;
- automatic wake-up and retry after capacity returns;
- durable job/checkpoint metadata;
- a stable programmatic interface for external software;
- optional cross-provider fallback through explicit handoff/checkpointing;
- native terminal-agent execution rather than model-protocol emulation.

AgentPool should remain small enough to audit and understand.

It is infrastructure, not an AI platform.

---

## 4. What AgentPool Is Not

AgentPool is **not**:

- a new LLM API;
- an OpenAI-compatible proxy;
- a replacement for Codex;
- a replacement for Claude Code;
- a replacement for Gemini CLI;
- a prompt-management platform;
- a model host;
- a billing system;
- a web SaaS;
- a general multi-agent framework;
- a new context-management system;
- a replacement for T3;
- a mechanism for copying authentication tokens between accounts;
- a system that pretends Claude and Codex sessions are interchangeable.

Keeping these non-goals explicit is important.

The project becomes much harder and less reliable if it tries to own model protocols, tool calling, MCP semantics, reasoning formats, provider authentication, or agent internals.

Those responsibilities stay with the native provider agents.

---

## 5. Core Mental Model

AgentPool has five important concepts:

```text
PROVIDER
   │
   └── SLOT
         │
         └── POOL
               │
               └── JOB
                     │
                     └── SESSION
```

### Provider

An adapter for a native execution system.

Examples:

```text
codex
claude
gemini
```

### Slot

One independently authenticated provider account/environment.

Examples:

```text
codex/openai-01
codex/openai-02
codex/openai-03

claude/anthropic-01
claude/anthropic-02

gemini/google-01
```

There is no architectural slot limit.

### Pool

A logical collection of eligible slots/capabilities.

Examples:

```text
codex
claude
gemini
best-coder
research
fast
```

The caller normally requests a pool rather than a slot.

### Job

The provider-independent description of work AgentPool is responsible for keeping alive.

### Session

The provider-native execution/session associated with a job at a particular point in time.

---

## 6. Unlimited Slots

Slots should be configuration/data, not hard-coded program structure.

For example:

```yaml
providers:
  codex:
    slots:
      - id: openai-01
      - id: openai-02
      - id: openai-03
      - id: openai-04

  claude:
    slots:
      - id: anthropic-01
      - id: anthropic-02

  gemini:
    slots:
      - id: google-01
```

Adding the fifth OpenAI subscription should require no code change.

Conceptually:

```bash
agentpool add codex
agentpool add codex
agentpool add claude
agentpool add gemini
```

Each provider adapter controls the appropriate official/native authentication process and isolated environment.

AgentPool stores references and metadata around authentication environments. It should avoid owning provider passwords or inventing authentication mechanisms.

---

## 7. Pools Hide Accounts

Applications should almost never request:

```text
openai-02
```

They should request:

```text
codex
```

or:

```text
best-coder
```

Example:

```yaml
pools:
  codex:
    providers:
      - codex
    slots: "*"
    strategy: capacity-aware

  claude:
    providers:
      - claude
    slots: "*"
    strategy: capacity-aware

  best-coder:
    providers:
      - codex
      - claude
      - gemini
    strategy: quality-first
```

This creates a stable abstraction.

A project can gain or lose subscriptions without changing its integration.

---

## 8. Automatic Same-Provider Rotation

This is the first and most important capability.

Suppose:

```text
CODEX POOL

openai-01    READY
openai-02    READY
openai-03    READY
```

AgentPool assigns a job:

```text
JOB-42
  │
  ▼
openai-01
```

If the slot becomes capacity-limited:

```text
openai-01
    │
    ▼
CAPACITY_EXHAUSTED
    │
    ▼
COOLDOWN until T
```

AgentPool evaluates the remaining pool:

```text
openai-01    COOLDOWN
openai-02    READY       ← next
openai-03    READY
```

If the provider supports safe session continuation across authenticated slots, AgentPool resumes the job through `openai-02`.

No human selection is required.

---

## 9. All Slots Exhausted

Running out of the entire pool is not a fatal job state.

Example:

```text
openai-01    cooldown until 20:31
openai-02    cooldown until 21:12
openai-03    cooldown until 20:08
```

AgentPool calculates:

```text
next possible capacity = 20:08
```

The job transitions to:

```text
WAITING_FOR_CAPACITY
```

Its durable state remains intact.

At or around the earliest reset:

```text
wake
 │
 ▼
probe eligible slot
 │
 ├── available ──► resume
 │
 └── unavailable ──► update cooldown and wait
```

The desired invariant is:

> **Temporary subscription exhaustion pauses work; it does not destroy work.**

---

## 10. Slot State Machine

A small explicit state machine is preferable to implicit heuristics.

```text
UNKNOWN
   │
   ▼
PROBING
   │
   ├──────────────► READY
   │                  │
   │                  ▼
   │                 BUSY
   │                  │
   │                  ├── success ─────► READY
   │                  │
   │                  └── limit ───────► COOLDOWN
   │                                       │
   │                                       ▼
   └───────────────────────────────────► PROBING

Other terminal states:

AUTH_REQUIRED
DISABLED
BROKEN
```

Useful slot metadata:

```text
slot_id
provider
auth_environment
state
supported_models
last_probe_at
last_success_at
cooldown_until
failure_count
current_job_id
capacity_metadata
```

---

## 11. Job State Machine

Jobs also need explicit states.

```text
QUEUED
  │
  ▼
STARTING
  │
  ▼
RUNNING
  │
  ├── completed ─────────► COMPLETED
  │
  ├── user stop ─────────► CANCELLED
  │
  ├── recoverable error ─► RECOVERING
  │                           │
  │                           ▼
  │                        RUNNING
  │
  └── no capacity ────────► WAITING_FOR_CAPACITY
                              │
                              ▼
                           RECOVERING
```

An unrecoverable provider/process/integration failure can transition to:

```text
FAILED
```

Capacity exhaustion by itself should not.

---

## 12. The Hard Part: Safe Failover

Automatic rotation must not simply rerun the previous command.

Imagine an agent turn:

```text
1. edit database migration
2. run migration
3. push commit
4. provider hits limit before returning final text
```

If AgentPool blindly repeats the entire turn on another account, side effects may happen twice.

Therefore AgentPool needs to distinguish:

```text
provider capacity failure
```

from:

```text
safe replay boundary
```

The safest initial design is conservative.

AgentPool records:

- native session identifier where available;
- process lifecycle;
- last completed agent event;
- repository/worktree state;
- checkpoint timestamp;
- provider failure classification.

When continuation cannot be proven safe, the replacement agent should be instructed to **inspect current state and continue**, not blindly replay all previous actions.

The governing principle is:

> **Resume from observed state, never assume the failed turn performed no side effects.**

---

## 13. Same-Provider vs Cross-Provider Continuation

These must be separate mechanisms.

### Same-provider slot failover

Example:

```text
Codex account A
       │
       ▼
Codex account B
```

If the native provider/session architecture permits it, the same native session can potentially continue.

This is the preferred failover because it preserves maximum context fidelity.

### Cross-provider fallback

Example:

```text
Codex
  │
  ▼
Claude
```

This cannot be represented as native session continuation.

Claude does not inherit Codex's hidden/native conversation state simply because AgentPool wants it to.

Cross-provider continuation therefore requires a **handoff**.

```text
Codex job
   │
   ▼
checkpoint
   │
   ▼
provider-independent resume package
   │
   ▼
Claude bootstrap
   │
   ▼
continue from repository + checkpoint
```

This distinction should remain visible throughout the architecture.

---

## 14. Durable Provider-Independent Checkpoint

AgentPool should maintain a small durable record for each job.

It does not need to duplicate the entire agent transcript.

A checkpoint can contain:

```yaml
job:
  id: job-42

  objective:
    "Implement X and verify Y."

  workspace:
    repository: /projects/tradeagent
    working_directory: /projects/tradeagent
    branch: feature/x

  execution:
    preferred_pool: codex
    current_provider: codex
    current_slot: openai-02

  progress:
    summary: >
      Implemented A and B. C remains.
    completed:
      - A
      - B
    pending:
      - C
      - tests

  state:
    files_touched:
      - src/a.ts
      - src/b.ts
    last_known_commit: abc123
    dirty_worktree: true

  decisions:
    - "Use architecture X because Y."

  continuation:
    instructions: >
      Inspect current repository state before changing anything.
      Continue with C. Do not repeat completed migration.
```

For sophisticated host applications, the host may supply richer checkpoint/handoff information itself.

AgentPool should allow that without requiring it.

---

# 15. Terminal Agents Are First-Class

This is critical.

AgentPool is not designed only for a human typing:

```bash
agentpool run codex
```

Its native execution unit is a **terminal agent process**.

Conceptually:

```text
AgentPool
   │
   ├── spawn Codex CLI
   ├── spawn Claude CLI
   └── spawn Gemini CLI
```

It owns the orchestration around the process while leaving actual AI-agent behavior to the native CLI.

A provider adapter is responsible for operations such as:

```text
authenticate
probe
start
resume
interrupt
classify failure
extract capacity information
```

A conceptual interface:

```ts
interface ProviderAdapter {
  authenticate(): Promise<Slot>;

  probe(slot: Slot): Promise<SlotHealth>;

  start(job: Job, slot: Slot): Promise<NativeSession>;

  resume(
    job: Job,
    session: NativeSession,
    slot: Slot
  ): Promise<NativeSession>;

  interrupt(session: NativeSession): Promise<void>;

  classifyFailure(
    event: ProviderEvent
  ): FailureReason;
}
```

This makes terminal agents interchangeable **at the orchestration boundary**, without pretending their internals are identical.

---

# 16. Human CLI

The simplest user interface should remain excellent.

Examples:

```bash
agentpool status
```

```text
CODEX
  openai-01      READY
  openai-02      BUSY        job-42
  openai-03      COOLDOWN    38m

CLAUDE
  anthropic-01   READY
  anthropic-02   READY

GEMINI
  google-01      READY
```

Run:

```bash
agentpool run codex
```

Choose a model/capability:

```bash
agentpool run codex --model gpt-5.6
```

Resume:

```bash
agentpool resume job-42
```

Inspect:

```bash
agentpool job job-42
```

The normal user should not have to reason about slots.

---

# 17. Programmatic Interface

This is what makes AgentPool useful to our other software.

AgentPool should expose a small stable local interface.

Possible transports:

1. local Unix socket;
2. local HTTP endpoint bound to loopback;
3. CLI with structured JSON I/O;
4. later, ACP if useful for compatible clients.

The semantic API matters more than the transport.

A host should be able to request:

```json
{
  "pool": "codex",
  "model": "gpt-5.6",
  "workspace": "/projects/tradeagent",
  "task": "Investigate and implement X",
  "continuity": "automatic"
}
```

AgentPool responds with a durable job identity:

```json
{
  "job_id": "job-42",
  "state": "running"
}
```

The host can subscribe to or poll structured events:

```text
JOB_STARTED
AGENT_EVENT
SLOT_CHANGED
CAPACITY_WAIT
JOB_RESUMED
CHECKPOINT_UPDATED
JOB_COMPLETED
JOB_FAILED
```

The important abstraction is:

> The host owns the task. AgentPool owns execution continuity.

---

# 18. Plugging Into TradeAgent

TradeAgent should not need code such as:

```text
try OpenAI account A
if rate limited:
    try OpenAI account B
if unavailable:
    try Claude
...
```

That logic belongs below TradeAgent.

Instead:

```text
TradeAgent
    │
    │ request terminal agent
    ▼
AgentPool
    │
    ▼
"codex" pool
    │
    ├── openai-01
    ├── openai-02
    └── openai-03
```

TradeAgent can say:

```text
Run this engineering/research task in workspace X using pool codex.
```

AgentPool returns:

```text
job-42
```

TradeAgent can then observe the job.

If OpenAI account 1 becomes exhausted:

```text
TradeAgent
    │
    │ sees same job-42
    │
    ▼
AgentPool
    │
    ├── old slot unavailable
    └── new slot selected
```

TradeAgent does **not** need to reorganize its own hierarchy merely because a subscription reached a rate limit.

This is a clean separation of responsibilities.

### TradeAgent owns

- agent hierarchy;
- managerial decisions;
- task decomposition;
- verification policy;
- domain logic;
- project-specific handoffs;
- its own context-cleansing/reset logic;
- deciding what work should happen.

### AgentPool owns

- provider authentication environments;
- available execution capacity;
- slot health;
- account rotation;
- provider process lifecycle;
- cooldowns;
- safe execution recovery;
- waiting for capacity;
- provider selection according to requested policy.

That boundary prevents AgentPool from becoming a second TradeAgent orchestrator.

---

# 19. Plugging Into AI Concierge

The same abstraction applies to AI Concierge.

```text
AI Concierge
      │
      ├── needs coding agent
      ├── needs research agent
      └── needs terminal-capable model
              │
              ▼
          AgentPool
```

AI Concierge might request:

```text
pool = best-coder
```

or:

```text
pool = claude
```

It receives a durable execution job.

The Concierge does not care whether the actual execution path becomes:

```text
Claude account 1
```

or:

```text
Claude account 1
      ↓ capacity exhausted
Claude account 2
```

That complexity stays local to AgentPool.

This is particularly valuable because the Concierge can remain provider-neutral while still benefiting from subscription-backed terminal agents.

---

# 20. Plugging Into Any Future Software

The intended generic relationship is:

```text
┌───────────────────────────────────────────┐
│                APPLICATION                │
│                                           │
│ TradeAgent / AI Concierge / IDE / script │
└─────────────────────┬─────────────────────┘
                      │
              execution contract
                      │
                      ▼
┌───────────────────────────────────────────┐
│                 AGENTPOOL                 │
│                                           │
│ routing                                   │
│ slot health                               │
│ failover                                  │
│ cooldown                                  │
│ recovery                                  │
│ durable jobs                              │
└─────────────────────┬─────────────────────┘
                      │
                 adapters
                      │
       ┌──────────────┼──────────────┐
       ▼              ▼              ▼
     CODEX          CLAUDE         GEMINI
       │              │              │
    slots...        slots...       slots...
```

The application should therefore depend on an **execution capability**, not a vendor account.

---

# 21. T3 Integration

T3 should eventually see AgentPool as one or several logical providers:

```text
AgentPool / Codex
AgentPool / Claude
AgentPool / Gemini
AgentPool / Best Coder
```

Instead of exposing:

```text
Codex Account 1
Codex Account 2
Codex Account 3
Claude Account 1
Claude Account 2
...
```

From T3's perspective:

```text
T3
 │
 ▼
AgentPool / Codex
 │
 ▼
automatic slot selection
```

This is exactly the experience AgentPool exists to provide.

An ACP-compatible interface may eventually be a clean integration path, provided it does not unnecessarily expand the first implementation.

---

# 22. Routing

V1 routing should be intentionally boring.

A pool can use:

```text
first-ready
round-robin
capacity-aware
```

The default should probably be **capacity-aware with deterministic tie-breaking**.

The scheduler considers:

```text
slot enabled?
authenticated?
provider healthy?
requested model supported?
cooldown expired?
already busy?
known remaining capacity?
recent failures?
```

It returns the highest-ranked eligible slot.

The router should be deterministic enough to debug.

No machine learning is required.

---

# 23. Cross-Provider Policies

Cross-provider pools can be useful, but they should be explicit.

Example:

```yaml
pools:
  best-coder:
    candidates:
      - provider: codex
        model: gpt-5.6
      - provider: claude
        model: opus
      - provider: gemini
        model: pro

    failover:
      same_provider: automatic
      cross_provider: checkpoint
```

This says:

1. rotate between compatible slots transparently;
2. if the entire provider is unavailable, create/verify a handoff;
3. bootstrap another provider from durable state.

A host may disable cross-provider fallback:

```yaml
cross_provider: never
```

That is important for jobs where model continuity matters more than uptime.

---

# 24. Concurrency

Multiple projects may use AgentPool simultaneously.

Example:

```text
TradeAgent ───── job-41 ─────┐
                             │
AI Concierge ─── job-42 ─────┼──► AgentPool
                             │
T3 ───────────── job-43 ─────┘
```

AgentPool might allocate:

```text
job-41 → openai-01
job-42 → anthropic-01
job-43 → openai-02
```

A slot should normally have explicit concurrency semantics.

For v1, the safest default is:

> one managed foreground agent session per slot unless the adapter explicitly supports more.

This avoids accidental capacity contention and authentication/session races.

---

# 25. Isolation

Every account must have an isolated authentication environment.

Conceptually:

```text
~/.agentpool/
    providers/
        codex/
            openai-01/
            openai-02/
            openai-03/

        claude/
            anthropic-01/
            anthropic-02/

        gemini/
            google-01/
```

The exact filesystem strategy depends on each provider.

The invariant is:

> Authentication for slot A must never be implemented by overwriting the live credentials of slot B.

Provider adapters may use provider-supported home/config overrides, shadow homes, environment variables, or another safe isolation mechanism.

---

# 26. Security Model

AgentPool runs locally by default.

V1 should:

- bind programmatic endpoints only to loopback/local IPC;
- never expose provider credentials through its API;
- never log access/refresh tokens;
- use restrictive filesystem permissions for AgentPool state;
- delegate authentication to official/native provider flows;
- redact sensitive environment variables from logs;
- avoid copying credential files unless a provider explicitly requires and safely supports it;
- make destructive job cancellation explicit;
- treat host-supplied shell/workspace permissions separately from provider authentication.

AgentPool should not become a credential vault unless there is no alternative.

---

# 27. Observability

Automatic infrastructure must explain what it did.

A useful event log might read:

```text
18:01:04  job-42 created
18:01:04  pool=codex model=gpt-5.6
18:01:04  selected openai-01
18:01:05  native session started
19:43:17  capacity limit detected
19:43:17  openai-01 cooldown until 22:00
19:43:17  selected openai-02
19:43:18  native session resumed
21:10:42  job completed
```

This is far more useful than silently switching accounts.

External software should be able to consume the same structured events.

---

# 28. Storage

SQLite is enough.

Potential tables:

```text
providers
slots
pools
pool_members
jobs
sessions
checkpoints
slot_health
events
```

SQLite provides:

- durability;
- transactions;
- concurrency appropriate to a local daemon;
- inspectability;
- simple backups;
- no external infrastructure.

No Redis, PostgreSQL, Kafka, or cloud service should be required.

---

# 29. Suggested Process Architecture

```text
                    agentpool CLI
                         │
                         ▼
                ┌────────────────┐
                │ local daemon   │
                │                │
                │ scheduler      │
                │ router         │
                │ job manager    │
                │ event stream   │
                └───────┬────────┘
                        │
                     SQLite
                        │
          ┌─────────────┼─────────────┐
          ▼             ▼             ▼
     Codex adapter  Claude adapter  Gemini adapter
          │             │             │
          ▼             ▼             ▼
      native CLI     native CLI     native CLI
```

The daemon is valuable because long-running jobs must survive the terminal/UI that started them.

A T3 window closing should not necessarily kill a managed autonomous job.

---

# 30. Minimal Repository Shape

Keep the repository compact.

```text
agentpool/
│
├── src/
│   ├── cli/
│   ├── daemon/
│   ├── core/
│   │   ├── jobs
│   │   ├── pools
│   │   ├── routing
│   │   └── recovery
│   │
│   ├── providers/
│   │   ├── codex
│   │   ├── claude
│   │   └── gemini
│   │
│   ├── runtime/
│   │   ├── process
│   │   ├── events
│   │   └── checkpoint
│   │
│   ├── api/
│   └── storage/
│
├── tests/
├── docs/
└── README.md
```

Avoid premature plugin frameworks.

A small TypeScript interface plus three adapters is sufficient until reality proves otherwise.

---

# 31. MVP

The first MVP should prove only the most valuable invariant:

> A long-running Codex job can survive exhaustion of one authenticated OpenAI slot without human intervention.

MVP scope:

### Required

- local daemon;
- CLI;
- SQLite;
- arbitrary number of Codex slots;
- isolated authentication;
- slot probing;
- Codex pool;
- deterministic routing;
- capacity/rate-limit classification;
- cooldown tracking;
- automatic account rotation;
- automatic waiting when all slots are exhausted;
- automatic retry after reset;
- durable jobs;
- structured logs/events;
- safe process recovery;
- basic programmatic execution API.

### Explicitly postponed

- Claude;
- Gemini;
- cross-provider failover;
- T3-native integration;
- ACP;
- GUI;
- remote networking;
- teams;
- cloud synchronization;
- advanced scheduling;
- cost optimization;
- distributed execution.

This keeps the first proof exceptionally focused.

---

# 32. Phase 2 — Multi-Provider

After Codex rotation is proven reliable:

```text
ProviderAdapter
      │
      ├── CodexAdapter     ✓
      ├── ClaudeAdapter
      └── GeminiAdapter
```

Each new adapter must prove:

1. isolated authentication;
2. health probing;
3. model discovery/selection;
4. process launch;
5. native resume behavior;
6. rate-limit detection;
7. reset/cooldown extraction where possible;
8. interruption behavior;
9. crash recovery;
10. session-state boundaries.

We should not assume one provider behaves like another.

---

# 33. Phase 3 — Software Integration

Once provider execution is stable:

```text
TradeAgent ───────┐
AI Concierge ─────┼──► AgentPool local API
T3 ───────────────┤
scripts ──────────┘
```

This is where AgentPool becomes shared infrastructure rather than a personal shell convenience.

The API should remain deliberately smaller than the internal implementation.

---

# 34. Phase 4 — Cross-Provider Continuity

Only after the native adapters are trustworthy should AgentPool implement:

```text
Codex pool exhausted
        │
        ▼
checkpoint
        │
        ▼
Claude
```

This phase requires explicit handoff semantics and should never masquerade as native session continuity.

For sophisticated host systems such as TradeAgent, the host may generate the best handoff itself and give AgentPool a continuation package.

AgentPool should support both:

```text
generic AgentPool checkpoint
```

and:

```text
host-authored checkpoint
```

---

# 35. Reliability Invariants

The implementation should be judged against explicit invariants.

### R1 — No credential swapping

One slot never becomes another by overwriting a shared live credential file.

### R2 — Capacity failure is recoverable

A provider limit does not mark a job failed while another eligible slot exists.

### R3 — Exhaustion is pausable

If no capacity exists, the job waits durably rather than disappearing.

### R4 — Automatic recovery

Waiting jobs resume when eligible capacity returns.

### R5 — No blind replay

AgentPool never assumes an interrupted agent turn had zero side effects.

### R6 — Observable routing

Every slot selection and failover is recorded.

### R7 — Host independence

TradeAgent, AI Concierge, T3, and scripts use the same execution abstraction.

### R8 — Provider independence

Core scheduling code does not contain Codex-specific authentication logic.

### R9 — Local-first

No external AgentPool service is required for normal operation.

### R10 — Small core

Features that belong to native agents or host applications are not duplicated inside AgentPool.

---

# 36. Example End-to-End Scenario

TradeAgent needs a builder agent.

```text
TradeAgent
    │
    │ POST job
    ▼
AgentPool
```

Request:

```text
pool: codex
workspace: ~/TradeAgent
task: implement unit X
```

AgentPool creates:

```text
job-913
```

Scheduler sees:

```text
openai-01 READY
openai-02 READY
openai-03 READY
```

and chooses:

```text
openai-01
```

Codex works for two hours.

Then:

```text
openai-01 → CAPACITY_EXHAUSTED
```

AgentPool records the event and current state.

It marks:

```text
openai-01 → COOLDOWN
```

It selects:

```text
openai-02
```

If native Codex continuation is safe:

```text
resume native session
```

Otherwise:

```text
inspect/checkpoint state
resume safely
```

TradeAgent still sees:

```text
job-913 RUNNING
```

No managerial agent needs to care that the underlying subscription changed.

Later all three OpenAI slots become exhausted:

```text
job-913 → WAITING_FOR_CAPACITY
```

TradeAgent sees that state if it cares, but no intervention is required.

At the first valid reset:

```text
AgentPool → probe → resume
```

Eventually:

```text
job-913 → COMPLETED
```

To TradeAgent, the entire execution was one durable job.

That is the product.

---

# 37. Example AI Concierge Scenario

AI Concierge receives a task requiring substantial repository work.

It requests:

```text
pool: best-coder
continuity: automatic
```

Policy currently prefers:

```text
Codex
```

AgentPool selects an available Codex slot.

If individual OpenAI slots become exhausted, same-provider rotation happens automatically.

If the entire Codex pool becomes unavailable and the pool permits cross-provider fallback:

```text
AgentPool
    │
    ▼
create/obtain handoff
    │
    ▼
select Claude
    │
    ▼
bootstrap Claude from workspace + checkpoint
```

AI Concierge receives the same logical job stream.

It can react to provider changes if desired, but does not need to orchestrate them.

---

# 38. Why This Should Be a Separate Program

It may be tempting to implement this directly inside TradeAgent.

That would be a mistake.

Subscription capacity is not trading-domain logic.

Likewise it does not belong specifically to AI Concierge or T3.

The same problem exists everywhere:

```text
TradeAgent
AI Concierge
T3
terminal
CI-like autonomous workers
future applications
```

Therefore:

```text
            shared infrastructure

                 AgentPool
                    ▲
          ┌─────────┼─────────┐
          │         │         │
      TradeAgent  Concierge   T3
```

This also means improvements to provider reliability immediately benefit every project.

---

# 39. What Makes AgentPool Valuable

The project is deceptively small.

Its value does not come from implementing AI.

It comes from turning fragmented subscription capacity into a dependable local execution primitive.

Without AgentPool:

```text
"I have three Codex subscriptions."
```

With AgentPool:

```text
"I have a Codex pool."
```

Without AgentPool:

```text
"Claude account 1 is exhausted."
```

With AgentPool:

```text
"Claude capacity is temporarily reduced."
```

Without AgentPool:

```text
"My autonomous job died because an account hit a limit."
```

With AgentPool:

```text
"The job is still running."
```

That is the conceptual transformation.

---

# 40. Design Philosophy

AgentPool should remain:

**Small.**  
It solves execution capacity and continuity, not every AI infrastructure problem.

**Native.**  
It uses the providers' real terminal agents rather than reimplementing them.

**Durable.**  
Long-running jobs survive temporary capacity loss.

**Provider-aware but provider-independent.**  
Adapters understand provider details; the core does not.

**Transparent.**  
Users can always see which slot is executing and why a switch occurred.

**Safe.**  
Failover never assumes that interrupted tool execution was side-effect-free.

**Composable.**  
Humans, T3, TradeAgent, AI Concierge, and future software all consume the same primitive.

**Local-first.**  
The machine already contains the authenticated agents and workspaces. AgentPool coordinates them locally.

---

# 41. North-Star Interface

Ultimately, all of the complexity should collapse into something this simple:

```bash
agentpool run codex
```

for a human,

and:

```text
execute(
    pool = "codex",
    workspace = "...",
    task = "...",
    continuity = "automatic"
)
```

for software.

Everything below that boundary—

```text
Which account?
How much capacity remains?
Did it hit a limit?
When does it reset?
Can another account continue?
Should the job wait?
Has capacity returned?
Which native CLI should run?
How is its authentication isolated?
How is the session resumed safely?
```

—is AgentPool's problem.

Everything above it can simply ask for an AI agent and expect the job to remain alive.

---

# 42. Final Definition

> **AgentPool is a small, local-first execution broker that turns any number of independently authenticated AI terminal-agent subscriptions into logical pools of durable agent capacity. It automatically selects healthy accounts, rotates around capacity limits, waits and resumes when capacity returns, and exposes one stable execution interface to humans and software such as T3, TradeAgent, AI Concierge, and future autonomous systems. It preserves native provider agents instead of replacing them, and uses explicit checkpoints whenever execution must cross a boundary that cannot safely preserve native session state.**

That sentence should remain the project's scope anchor.

If a proposed feature does not strengthen that definition, it probably does not belong in AgentPool.
