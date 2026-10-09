# AgentPool — Environment Slots & Multi-Login
## v2.0 · Lean Runtime, Deterministic Enrollment & Certified Inference · Implementation Contract

**Document date:** 9 October 2026  
**Document status:** **NORMATIVE DESIGN — NOT IMPLEMENTED OR PROVEN**  
**Target project:** [`Nicolas1bhr/agentpool`](https://github.com/Nicolas1bhr/agentpool)  
**Parent architecture:** [`AgentPool v5.2 — Proactive Routing & JEV Implementation Specification`](https://github.com/Nicolas1bhr/agentpool/blob/main/docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md)  
**Replaces:** `AgentPool_Environment_Slots_and_Multi_Login_Implementation_Contract_v1.md`, as to its slot enrollment, placement, worker topology, registration lifecycle, and associated build order  
**Based on:** `AgentPool_Environment_Slots_Multi_Login_v2_Lean_Runtime_Audit_and_Contract.md` and its 32 adversarial findings  
**Audience:** Build agents, reviewers, maintainers, test engineers, and the product owner  
**Repository changes performed by this document:** None  
**Canonical authority:** This file is the v2 contract **only for its specified addendum scope**; it does not secretly supersede unrelated requirements in v5.2.

> **THE PRODUCT IN ONE SENTENCE.** AgentPool exposes one model-inference provider to OpenCode while holding any number of logically grouped, separately registered, provider-authorized inference connections; an owner adds API or supported subscription connections from a local dashboard, AgentPool automatically places each registration into the earliest compatible environment with a vacant position for that integration (creating another logical environment only if necessary), and one central route selector automatically chooses an eligible, certified physical connection for each new inference request without ever executing an agent.

> **THE BOUNDARY THAT MUST NEVER MOVE.** **TradeAgent owns the work. OpenCode executes the agents. AgentPool supplies inference.** TradeAgent and OpenCode are not children of AgentPool. Environment slots do not execute agents. AgentPool does not run prompts as jobs, execute tools, own agent memory, orchestrate handoffs, control trading, or synthesize model output.

> **THE V2 CORRECTION.** **Environment slot = persistent logical placement metadata.** It is NOT inherently a process, container, sandbox, browser session, independent quota bucket, HTTP connection, worker, or permission to access a subscription. A registered credential is independently managed; an inference route is independently eligible; an optional isolated executor exists only where its connector's proven requirements demand it. The common API transport runs in-process in the single Go gateway.

> **HOW TO HANDLE AN UNKNOWN.** Never guess, silently bypass, weaken a contract, or redefine the product. Execute the exact proof procedure in §16. Record evidence. Apply its stated PASS/FAIL decision. If the necessary provider mode cannot be certified, leave that **connector mode** unavailable; the generic multi-login product, logical slot model, and other connectors continue unchanged.

---

# 0. Interpretation and authority

## 0.1 Normative language

- **MUST/MUST NOT:** non-negotiable for the stated feature/milestone; a violation blocks that milestone.
- **SHOULD:** default engineering rule. Deviation requires owner-reviewed evidence of improvement and passing unchanged product tests.
- **MAY:** optional, does not weaken MUST requirements.
- **DESIGN:** specified behavior not yet implemented.
- **PROVEN:** passed a real, version-pinned conformance test with evidence; synthetic tests alone do not prove vendor compatibility.
- **SUPPORTED:** technically functional **and** permitted for the specific connector, identity type, flow, and use.
- **UNKNOWN:** lacking evidence; not proof of support, authorization, identity, independent capacity, or equivalence.
- **INELIGIBLE:** the route is explicitly not selectable for the considered request; it may still be a valid saved registration.
- **NO-GO:** do not ship/enable the affected mode until its test gate passes. NO-GO for one connector is not NO-GO for the entire architecture.

When a test discovers a mismatch, **fix the implementation to the invariant** or submit a separately reviewed contract amendment. Builders cannot convert a failing test into a revised expectation in the same PR.

## 0.2 Precedence and fixed boundaries

1. Laws, applicable provider authorization, owner security constraints, and the verified downstream/upstream protocol govern whether any given connection can be activated.
2. AgentPool v5.2 remains authoritative for model aliases, grants, central route selection, budgets, policy, context continuation, stream correctness, automatic eligible next-request switching, and the optional JEV advisory.
3. This v2 governs logical environment topology, enrollment-time placement, registration-to-slot relationship, default transport implementation, related states and schema, testing and build order. Its explicit replacements of v1 are recorded in §19.
4. Examples are synthetic until the connector passes the certification gate. Names such as Codex, Google, Anthropic, OpenAI, or Claude in occupancy diagrams **do not grant an entitlement or imply that a subscription can serve third-party inference**.
5. A real contradiction with the parent is a **documented blocking conflict**. Record both clauses, the proposed minimal fix, affected acceptance tests, and owner decision. Never assume this addendum authorizes changing the inference boundary.

## 0.3 What the owner will experience (locked)

The owner opens AgentPool's authenticated **local dashboard** and chooses **Add a subscription** or **Add an API**. They choose one offered *connector integration and auth mode*, complete its officially supported authentication flow, and AgentPool automatically places the resulting registration:

- Try existing **compatible, usable** logical environments in stable ascending creation sequence.
- A position is vacant when no current registration/committing reservation occupies the exact integration key and all declared placement constraints permit it.
- Reuse the earliest available position.
- Create the next logical environment if no compatible existing position exists.
- Show the **actual committed slot**, verified or unknown principal, protected registration status, and independently assessed route eligibility.
- Never create an OpenCode agent, provider CLI agent, browser-agent process, background model task, or per-slot worker as a side effect of adding an account.
- Existing OpenCode configuration remains unchanged when accounts are added or disconnected. Its logical model catalog may be refreshed or explicitly regenerated only by an authorized management operation, not by passing physical keys to OpenCode.

Example occupancy (synthetic integration labels):

| Logical environment | `codex-example` | `google-example` | `anthropic-example` |
|---|---:|---:|---:|
| Slot A | occupied | occupied | vacant |
| Slot B | occupied | vacant | occupied |
| Slot C | occupied | occupied | occupied |

- Add `anthropic-example` → **Slot A**, no D.
- Add `google-example` → **Slot B**, no D.
- Add `codex-example` → **Slot D** created, because A/B/C already have it.
- A cancelled or duplicate enrollment does **not** needlessly create D.
- Slot IDs are durable opaque IDs; A/B/C are friendly display labels only.

**No arbitrary fixed MAX_SLOTS.** N is bounded only by configured resource constraints, real connector/provider limits, security, and physical hardware. “Technically unlimited” never means infinite live processes, model calls, spend, or independent quotas.

## 0.4 Claims we expressly do not make

- That any particular consumer plan authorizes third-party inference.
- That separate logins, keys, slots, billing aliases, or browser profiles imply separate upstream rate budgets.
- That two same-named models are behaviorally, protocol-, tool-, state-, or billing-equivalent.
- That an in-process default gives per-account OS compromise containment.
- That a retry after ambiguous send is safe merely because another account is available.
- That zero extra latency, zero idle RAM, or infinite throughput can be guaranteed.
- That all provider features can be translated into a single universal HTTP protocol.
- That this document contains measurements or certification that do not yet exist.

---

# 1. Core architecture — definitive implementation choice

## 1.1 Topology

```text
TRADING / CODING WORK (NO CHANGE)

TradeAgent ──normal host/session/tool controls──> OpenCode (TUI/headless)
                                                    │
                                                    │ one logical model provider:
                                                    │ agentpool/<explicit-model-alias>
                                                    ▼
┌──────────────────────── AGENTPOOL GO DAEMON ────────────────────────────┐
│ INFERENCE DATA PLANE (loopback, gateway-grant authenticated)           │
│   Parser + grant/policy → immutable indexed candidates → selector      │
│        → scoped capacity/strict-budget reservation → frozen attempt    │
│        → certified adapter / reusable upstream transport → stream      │
│                                                                        │
│ OWNER CONTROL PLANE (separate listener/origin, privileged login)        │
│   Dashboard → provisional enrollment → verified identity              │
│        → deterministic atomic logical placement → route publication    │
│                                                                        │
│ PRIVATE STATE / SERVICES                                                │
│   SQLite WAL metadata + durable accounting where required             │
│   Secret custody (OS vault or separately certified secure mode)        │
│   Registration/route/identity/capacity catalog                          │
│   Bounded asynchronous refresh/health/telemetry                        │
│   Optional isolated connector executor ONLY IF PROVEN NEEDED          │
│                                                                        │
│ LOGICAL REGISTRY (NO PER-SLOT PROCESSES BY DEFAULT)                     │
│   Slot A : integration X → registration X1; Y → Y1                    │
│   Slot B : integration X → registration X2; Z → Z1                    │
│   Slot N : created as needed; no idle worker or allocated network      │
└────────────────────────────────────────────────────────────────────────┘
                                 │
                      approved model inference
                                 ▼
                 Provider-native API / local model endpoint
```

**Mandatory separation:** enrollment-time slot placement does not select a model route; inference-time routing does not allocate a slot, log in, or manipulate provider identities. They operate on independent indexes and state machines.

## 1.2 Implementation tiers

| Tier | When used | Process count | Security statement |
|---|---|---|---|
| **T0 — Default** | Certifiable API/native local transports whose libraries can safely execute in Go core | One AgentPool daemon regardless of logical slot count | Registration-scoped state and secret custody, **not** compromise isolation from core |
| **T1 — Dedicated connector executor** | A specific connector demonstrably needs process isolation, incompatible runtime, or protected library | Only bounded executors actually required by that connector | Isolation strength depends on independently tested OS permissions |
| **T2 — Hardened credential compartment** | Owner's threat model requires one compromised executor not read another registration's secret | Restricted OS identities / brokered capabilities, number determined by security boundary | Must prove least privilege per OS; more expensive than T0 |

**T0 is the reference implementation.** Neither Tier 1 nor Tier 2 introduces an agent runtime or a second global selector. Optional executors are keyed to the smallest necessary *connector or secret-security equivalence class*, **not** the friendly A/B/C slot label. They receive a frozen, narrow dispatch capability. They cannot browse projects, call TradeAgent/OpenCode, run tools, rewrite task instructions or manage global credentials.

**Tier-selection rule:** If a permitted connector functions and passes security, protocol, reliability, and performance tests at T0, implement T0. Promote that connector to T1/T2 **only** on documented, reproduced incompatibility or independently justified threat model; compare costs and attach proof. If an OS cannot prove the requested T2 isolation, mark T2 unsupported on that OS and do not advertise or silently emulate it with file directories/child PIDs.

## 1.3 Responsibilities and forbidden dependencies

| Component | Owns | MUST NOT own |
|---|---|---|
| TradeAgent | Agent task, policies, scheduling, checkpoints, trading authority, OpenCode workers | Physical upstream credential selection per inference call |
| OpenCode | Agent loop, session, prompt construction, tools, compaction, model calls/retries | Stored upstream AgentPool credentials or slot-worker control |
| AgentPool gateway | Authentication of callers, request admission, eligible physical routing, upstream transport, streaming, budgets | Agent tasks, tool execution, memory/handoff, model-generated decisions |
| Slot allocator | Owner-initiated placement of committed registrations | Inference dispatch or capacity-based model selection |
| Logical slot | Durable grouping/occupancy and friendly UI label | OS process readiness, independent quota or model state |
| Connector adapter | One certified auth/inference protocol and route capability | Global model selector or general-purpose execution |
| Credential broker | Restricted retrieval/refresh/rotation/revoke of approved secrets | Inference route policy or arbitrary source-code execution |
| Optional JEV advisor | Asynchronous soft health-risk ranking evidence, after calibration | Authorization, quotas, budget enforcement, retries or hot-path blocking |

## 1.4 Why this is simpler — without claiming magic performance

Dormant logical slots require compact registry rows, not one process, IPC endpoint, health poll, socket, OAuth profile, timer, refresh goroutine or model client each. Compatible outbound connections may share a vetted `http.Transport` by endpoint and trust profile while credentials are injected **per frozen request**. This reduces possible per-slot overhead. It does **not** guarantee better performance until §17 comparisons demonstrate it; T0 necessarily centralizes a larger trusted computing base.

---

# 2. Exact object model and identity relationships

Every object below exists for a reason. Do not collapse distinct object IDs into one overloaded “account/slot/provider” string.

| Object | Exact role | Relationship / invariant |
|---|---|---|
| `IntegrationKey` | Manifest-owned identity for product + credential mechanism and relevant scope, e.g., `vendor.product.authmode` | Stable, immutable within version; not just vendor brand |
| `EnvironmentSlot` | Persistent logical grouping, compatibility class and immutable sequence | Many distinct integration keys; **no implied runtime** |
| `SlotOccupancy` | The current one position for `(slot_id, integration_key)` | **At most one committed or committing position per pair**; history elsewhere |
| `EnrollmentAttempt` | One owner-approved authentication transaction, idempotent | Provisional; does not automatically reserve a permanent slot |
| `CredentialRegistration` | Independent secret handle, credential generation and owner choice | Exactly one live slot occupancy; may expose zero or more routes |
| `VerifiedPrincipal` | Actual provider-defined account/tenant/org/project/subject where identity can be proven | Unknown when no safe means of validation; *not* automatically capacity owner |
| `CapacityOwner` | Provider-defined actual rate/allowance scope and limit dimensions | May span registrations and/or endpoints; independence needs evidence |
| `PhysicalRoute` | Specific registration, model+variant, protocol, endpoint, billing identity and credential mode | Zero or more per registration; eligibility derived, not automatic |
| `ModelAlias` | OpenCode-visible truthful, versioned model contract | Maps only to certified routes with the requested equivalence/continuation semantics |
| `WorkloadGrant` | Short-lived scoped gateway authority for one OpenCode client/workload | Never equals owner/admin authorization or upstream credential |
| `TransportProfile` | Endpoint/trust/proxy/protocol/HTTP pool configuration | Share only when the **whole** transport security profile matches |
| `InferenceAttempt` | One admitted, frozen dispatch and its outcome evidence | One physical route per attempt; **no midstream splice** |
| `ContinuationAffinity` | Minimal routing reference to nonportable provider state | Pins an authorized same-route/state class; does not own chat memory |
| `ConnectorCertification` | Pinned evidence of permitted auth/inference and protocol behavior | Required before route enters eligible set |

## 2.1 Integration-key selection (deterministic)

An `IntegrationKey` is declared in the connector manifest. It includes exactly the dimensions that distinguish independent **installation/auth positions** and cannot be guessed by stringifying a model/vendor name. Different API, enterprise, project-scoped or approved subscription mechanisms MAY have separate keys. Do not conflate “Google account”, “Gemini API key”, “Google Vertex service account”, and “Gemini consumer subscription.”

If a provider supports multiple projects under a single credential or a registration exposes many models, keep *one registration* with many routes rather than duplicate the same credential to force artificial slot positions. If two distinct keys refer to a shared real rate-limit owner, bind them to the same proven `CapacityOwner` record.

Manifest contract (logical, not committed Go syntax):

```yaml
integration_key: vendor.product.auth-mode
manifest_version: 1
credential_modes: [api_key]           # only individually approved modes
placement:
  profile: default-local
  exclusive: false
  incompatible_with: []
  requires_provisional_namespace: false
execution:
  tier: T0                         # default; T1/T2 require justification
  protocol_profiles: [chat-v1]    # actual identifiers decided by certification
policy:
  allowed_inference_modes: []      # empty means NOT ELIGIBLE until certified
  certification_record: null
identity:
  verification: optional         # may be required, optional, or impossible
  duplicate_scope: null
capacity:
  ownership_evidence: unknown
```

An unknown or empty inference certification MUST block route eligibility. A UI entry is not a transport license.

## 2.2 Slot compatibility

A usable existing slot is eligible for integration `I` iff ALL hold:

1. slot state is `ACTIVE` and owner/group matches;
2. its immutable compatibility profile supports `I`'s declared execution/credential constraints;
3. no committed/committing occupancy for `(slot, I)` exists;
4. no existing integration declares mutual exclusion with `I`, and `I` declares no mutual exclusion with existing integrations;
5. OS/runtime restrictions are met;
6. no owner quarantine or active migration blocks placement.

First-fit means **lowest immutable slot sequence among eligible compatible vacancies**, not “choose visually first even if unsafe.” A slot can host different integration keys without meaning it owns or can access all their secrets. A disconnected, expired, rate-limited or suspended registration still occupies its position until explicitly removed or atomically replaced through reauthorization.

**Slot IDs are never reused.** Slot friendly labels MAY be renamed without affecting IDs; to avoid confusion, default labels retain monotonic sequence (A/B/C… or Slot 1/2/3…). “N+1” means a new durable record and new sequence, not a new process.

---

# 3. Enrollment transaction — authenticate provisionally, commit placement late

## 3.1 Exact owner journey

1. **Authorize owner.** Control plane authenticates the owner and applies CSRF/Origin/Host/session/rate protections *before* creating any attempt.
2. **Select mode.** Dashboard shows only supported, currently permitted, manifest-defined integration/auth combinations. Unsupported ones may be visible as informational entries but their **Connect** action is disabled and reason explicit.
3. **Create attempt.** Persist `EnrollmentAttempt` with an owner-scoped unique idempotency key, integration key, expected connector revision, randomized nonce, expiry and one-time auth transaction binding. Return an informational first-fit slot **preview**; preview is not a reservation or promise.
4. **Authenticate safely.** API-key modes receive owner-supplied secret through protected control flow. OAuth uses the actual provider-authorized flow and permitted client identity; bind PKCE/state/issuer/redirect and owner session. Device flow only if that connector actually permits it. Never scrape a CLI credential, browser cookies, passwords or pretend to be the vendor's client.
5. **Verify what can be verified.** Validate issuer/audience/scopes/identity and entitlements when disclosed; identify tenant/project/workspace where meaningful. An API key with no safe principal endpoint is `PRINCIPAL_UNKNOWN`; that does not mean broken registration or independent quota.
6. **Stage secret.** Write under an attempt-scoped, expiring opaque secret handle to OS-backed (or separately certified) custody. The database stores only the handle and nonsecret metadata.
7. **Commit atomically.** Under one **short** SQLite write transaction: fence against duplicate/expired/cancelled attempts and concurrent identity conflicts; rerun first-compatible-vacancy; create a logical slot if necessary; insert one occupancy and registration; mark attempt committed; append audit event.
8. **Finalize asynchronously and idempotently.** Promote the staged vault handle (or retain a validated durable handle), reconcile any crash boundary, publish refreshed routing snapshots, queue permitted route certification/health checks.
9. **Show fact, not fiction.** Return actual committed slot and `REGISTERED` state; show each inferred route independently as `UNVERIFIED`, `VALIDATING`, `ELIGIBLE`, or precise ineligible reason.

## 3.2 Two-phase commit rationale

The external provider login, OS key store, and SQLite do **not** offer a shared atomic transaction. Do **not** keep a SQLite transaction open while awaiting a browser, provider, vault prompt, or network response. Stage the credential, commit a durable opaque reference, and reconcile at startup or failure. A staged handle must be safe from accidental GC while its committed reference is being finalized.

Default: do **not** create or reserve a permanent slot until the authentication result is accepted and the verified identity conflict policy can be resolved. An authorized connector needing its own isolated pre-auth profile may create a bounded `ProvisionalNamespace(enrollment_id)`; that ephemeral namespace does **not** count as a logical environment. Transferring/binding the resulting credential is allowed only when the connector's auth protocol supports it without copying nonportable external session state.

## 3.3 Duplicate identities — explicit semantics

- **Same verified principal + integration + scope, same type of credential:** default action is **reauthorize the original registration**, preserving its slot and id. Never silently fabricate another independently metered account.
- **Owner intentionally adds a distinct authorized API credential to the same principal/project:** MAY create a second registration in the next valid slot only when connector policy supports multiple keys; bind it to the same proven shared capacity owner unless provider evidence says otherwise. Display “same principal/shared or unknown capacity,” not “new quota.”
- **Identity UNKNOWN:** do not invent an equality or independence conclusion. Preserve a nonsecret credential fingerprint/HMAC for exact-secret duplicate detection where the secret format permits; never log or expose raw/fingerprint values, and do not equate distinct secrets with distinct quotas. Unknown-principal enrollment is allowed only if connector certification tolerates it.
- **Two simultaneous verified duplicates:** the uniqueness/reconciliation decision is made transactionally by the database; one succeeds or one becomes a reauthorization/explicit-duplicate flow. Neither overwrites the other's token.
- **Wrong browser account selected:** show the *returned verified identity* and require owner confirmation before linking/replacing. A browser cookie never establishes identity isolation.
- **Reauthorization changes principal:** refuse in-place replacement; require explicit new registration flow. Existing active stream rules remain governed by revocation semantics.

## 3.4 Commit pseudocode (normative semantics)

```text
CommitEnrollment(owner, attemptID, presentedGeneration, authResult):
  verify owner session and single-use transaction binding
  verify applicable connector permission/certification for AUTH mode
  verify authResult and stage secret under attempt-specific protected handle

  BEGIN IMMEDIATE                         // no provider/network call within TX
    A := SELECT attempt WHERE id=attemptID AND owner=owner
    if A.state == COMMITTED: return stored immutable commit result
    require A.generation == presentedGeneration
    require A.state == VERIFIED and not cancelled/expired
    if verifiedPrincipal collides:
      if owner decision missing: mark A AWAITING_OWNER_DECISION;
                                COMMIT; return NEED_OWNER_DECISION
      if choice == REAUTHORIZE_EXISTING:
         require same verified principal + matching integration/scope
         UPDATE existing registration using credential-generation CAS
         KEEP its slot occupancy, slot ID and registration ID unchanged
         UPDATE A to COMMITTED with existing registration and slot result
         INSERT audit + routing-catalog-outbox event
         COMMIT; finalize secret handle; return existing result
      if choice != DISTINCT_CREDENTIAL_ALLOWED_BY_CONNECTOR:
         ROLLBACK; reject
      # A legitimate distinct key may share principal/capacity; never
      # assume another quota merely because a new registration is added.
    slots := ACTIVE compatible slots ordered by immutable sequence
    S := first slots where (S,I) has no current occupancy
    if S missing:
       S := INSERT logical slot with durable monotonic sequence
    INSERT current occupancy(S.id, I, newRegistrationID)
       with UNIQUE(slot_id, integration_key)
    INSERT new CredentialRegistration with custody handle + epoch
    UPDATE A to COMMITTED, store canonical result, increment generation
    INSERT audit and routing-catalog-outbox event
  COMMIT

  idempotently mark secret handle permanent / reconcile if interrupted
  publish new immutable route/registration snapshot from outbox
  schedule cert/health checks off inference hot path
  return committed slot and registration, never assert ELIGIBLE by inference
```

**Failure invariants:** all concurrency safety depends on actual SQL constraints, not application `if` checks; idempotency replay returns identical recorded result after crash; the generation comparison prevents delayed OAuth callbacks from reviving a cancelled/expired transaction. Duplicate identity constraints use only verified canonical identity fields and the connector's defined scope. Historical attempts remain auditable outside the current occupancy relation.

## 3.5 Enrollment state transition table

| Current state | Event | Next state | Allowed side effect |
|---|---|---|---|
| `CREATED` | start auth | `AUTHORIZING` | Start approved provider flow |
| `AUTHORIZING` | valid success | `VERIFIED` | Stage protected credential; confirm identity |
| `AUTHORIZING` | cancel | `CANCELLED` | Expire auth nonce; safely remove staged secrets |
| `CREATED/AUTHORIZING` | expiry | `EXPIRED` | Same cleanup; callback rejected |
| `VERIFIED` | duplicate requiring choice | `AWAITING_OWNER_DECISION` | No slot commit |
| `AWAITING_OWNER_DECISION` | owner confirms supported choice | `VERIFIED` | Bind explicit decision |
| `VERIFIED` | commit succeeds | `COMMITTED` | Allocate occupancy, bind secret handle, audit |
| Nonterminal | unrecoverable validation failure | `FAILED` | Error code + cleanup |
| Any terminal state | repeated callback/cancel | unchanged | Return recorded status / reject replay |

All transitions are fenced by stored generation and attempt nonce. A stuck provisional state has a finite expiry; cleanup is idempotent. Default TTL must be a configurable finite value and tested at its boundary; use the provider's advertised auth lifetime when shorter. Never infer wall-clock expiry as a sufficient cross-process lock: transactional generation is decisive.

---

# 4. Slot occupancy, deletion, relocation and lifecycle

## 4.1 Slot state

`ACTIVE` (placement eligible) → `QUARANTINED` (not eligible for new placements) → `ACTIVE` or `RETIRED`. A retired slot cannot be resurrected under a reused ID. Quarantining a slot **does not automatically revoke all its credentials or stop their streams** unless the reason is a security/security-policy action explicitly requiring route suspension. Placement and inference controls are distinct.

## 4.2 Occupancy state

A committed registration's integration position is `OCCUPIED` independent of auth expiry, cooldown and health. Before commit there is normally no slot occupancy; only a brief transactional commit may reserve an occupancy. Explicit **disconnect** affects credentials/routing but does not necessarily free the slot; **remove registration** frees the current occupancy after safe revocation/drain rules. Audit history remains.

## 4.3 Delete / reconnect / remove

- **Reconnect:** same registration ID and slot if verified identity matches, new credential generation; maintain historical use linkage.
- **Disable:** block new admissions, retain protected credential and occupancy; already frozen attempts follow chosen drain/abort semantics.
- **Disconnect:** invalidate or revoke credential according to supported upstream API and owner choice; block new admissions; keep occupancy until removal if the owner may reconnect.
- **Remove registration:** mark ineligible; deny new admissions; drain/abort active attempts according to explicit owner action; unlink secret handle, release occupancy once references are gone, retain tombstone and audit. A vault deletion error must be surfaced and retried without silently claiming deletion.
- **Remove slot:** only if empty, no pending commit, no in-flight binding, no audit-referenced active registration, and no migration. Preserve immutable slot identity in history; do not shift other slots to fill visual gaps.
- **Relocate registration:** **not part of v2 MVP**. Never move it silently to compact a display. A future relocation requires a separately approved atomic slot-binding migration and proof against active attempts, vault and external nonportable auth state.

Rate-limit cooldown is **not** a vacancy and **never** triggers registration/slot creation on the inference path.

---
# 5. Registration, credential custody and refresh

## 5.1 Registration state (independent of slot)

```text
STAGED → REGISTERED → VALIDATING → READY
                         │          │
                         └→ VALIDATION_FAILED
                                    │
READY → REFRESHING → READY           │
  │          └→ REAUTH_REQUIRED ←────┘
  ├→ DISABLED
  ├→ REAUTH_REQUIRED
  └→ REVOKING → REVOKED
```

`READY` means a specific credential is accessible and its requisite auth validation has passed; it **does not** prove that every model route it might serve is protocol-certified, available, permitted, within budget or presently selected. Route eligibility is derived separately. `REFRESHING` need not block admissions if an unexpired old credential is valid and the connector's refresh contract permits concurrent use; an expired credential always blocks.

## 5.2 Credential custody (mandatory)

1. Secret values never go into SQLite, browser local storage, OpenCode configuration, process command-line arguments, telemetry, audit logs, crash reports, test snapshots, GitHub, or exception messages.
2. The database stores an **opaque secret reference**, registration ID, credential generation, and nonsecret metadata. File paths and vault item IDs are never accepted from inference clients as authority.
3. Default to a proven OS key store compatible with the host's attended/unattended lifecycle. **Headless use must be separately exercised**: a vault that requires a desktop unlock after reboot does not satisfy unattended readiness.
4. If the target host lacks a usable unattended OS vault, offer **only a separately reviewed encrypted-at-rest deployment mode** with explicit key provisioning, filesystem permissions, rotation/backup plan, and startup-unlock guarantees. Plaintext fallbacks are forbidden.
5. Gate all secret read operations by internal registration ID, credential epoch, caller role and connector scope. In T0, the daemon is a trusted process with access to permitted credentials: this is **logical access control, not OS compromise isolation**.
6. Resolve an authorized secret for a frozen attempt without an OAuth interaction or token refresh on the inference hot path. If credential cannot be retrieved safely, fail the route's admission.
7. Keep secret lifetimes in memory as short as practical; avoid unnecessary copies; recognize that Go garbage collection is not a guarantee of zeroization. For the stronger compromise boundary, use T2 instead of asserting impossible memory isolation.
8. Never blindly inherit external environment variables from OpenCode or TradeAgent into AgentPool. A connector must receive only the specific credentials it needs.
9. Disallow redirects that would move credentials to an unapproved hostname or protocol. Every actual upstream endpoint and TLS trust profile comes from the certified connector manifest and owner-approved configuration.
10. Audit **operations and identity references**, never secret bytes or raw tokens.

## 5.3 Refresh and rotation (single-flight, fenced)

- Lock refresh **per registration ID**, not per slot/provider globally; concurrent uses of the same registration share one refresh result. No unbounded registration-spanning mutex.
- Each credential has monotonically increasing `credential_generation`. A late refresh or OAuth callback cannot overwrite a newer credential generation.
- If the provider rotates refresh tokens, persist access/refresh token changes as one consistent secret bundle using the store's available atomic replace operation; where not atomic, implement a recoverable two-version handoff and test both crash boundaries.
- Refresh before expiration using bounded, jittered, prioritized scheduling **off the inference path**. If the provider rejects or revokes the credential, block new admissions for its routes and mark `REAUTH_REQUIRED` with actual evidence.
- A refreshed credential may replace auth material for *future requests* on the same route; a frozen attempt remains bound to its auth/route epoch and may finish if still authorized, or abort if immediate revoke is required. Do not silently swap an in-flight bearer unless the certified transport explicitly supports that semantic.
- Provider-wide or tenant-wide revocation propagates to other registrations only when credible evidence establishes the common identity/scope.
- Do not start one timer or goroutine per dormant slot. One bounded scheduler handles registrations needing refresh/validation.

## 5.4 Credential reconciliation after crashes

**Design contract:** Vault and SQLite are not a single transaction. On daemon startup and periodically off path:

1. Enumerate *metadata* of expiring staged vault handles and SQLite enrollment states; do not dump secret material.
2. If a staged handle belongs to `COMMITTED` registration and the handle is recoverable, finalize it idempotently.
3. If a staged handle belongs to terminal `CANCELLED/FAILED/EXPIRED`, revoke if supported and delete after an explicitly bounded grace that cannot race with final commit.
4. If SQLite references a missing or unreadable handle, set registration `REAUTH_REQUIRED`/`CUSTODY_UNAVAILABLE` and disable routes. Never generate a replacement secret or silently consider it READY.
5. If a vault item is marked for deletion but still referenced by an in-flight frozen attempt, delay deletion or abort by explicit security policy. Use durable ref counts/tombstones where needed; do not rely solely on in-memory counts across crashes.
6. Append a redacted reconciliation event; duplicate runs result in the same state.

## 5.5 Secret backup/recovery

Metadata backup **does not equal credential backup**. Clearly identify whether OS keychain entries migrate to another host. Restoration on a new host should default to needing reauthorization unless a separately certified encrypted export/import scheme is deliberately implemented. Document which secrets, refresh tokens and registration identities are recoverable. Test vault-loss, database-only restore, keychain-only restore, and host migration without inventing recoverable credentials.

---

# 6. Connector authorization and protocol certification

## 6.1 One connector mode is one independently certified product

The UI's **Subscription** and **API** tabs are organization labels. A connector is enabled for inference only for a specific combination of:

`provider + product + credential/auth mode + caller/use model + exact inference endpoint/protocol + entitlement + model/capability profile + terms version`.

Permission to sign in is not permission to send arbitrary provider inference from an intermediary. A first-party OAuth client flow is not automatically an allowed third-party client flow. Different paid accounts are not inherently new capacity owners. A provider may prohibit pooling even when each account belongs to the same person.

**Mandatory provider-policy rule:** Follow `docs/PROVIDER_POLICY.md`: one owner's own permitted accounts, no credential resale/shared ownership, no spoofing first-party apps, no extraction of private CLI tokens, no web-UI automation where unsupported, and no evasion of rate/usage caps. If any operating mode violates current provider requirements, disable that **mode**, preserve its generic schema slot, and record a `POLICY_UNSUPPORTED` decision.

## 6.2 Connector manifest (required fields)

```text
ConnectorManifest:
  integration_key; manifest_version; upstream_vendor; product; auth_mode
  application_identity; documented_auth_flow; reviewed_terms_url_and_section
  terms_review_date; allowed_use_context; required_scopes
  principal_verification_method_or_UNKNOWN
  capacity_scope_method_or_UNKNOWN; duplicate_identity_scope
  endpoint_allowlist; redirect_policy; tls_trust_profile
  allowed_model_ids_and_variants; upstream_protocols
  request_and_response_capabilities; continuation_class
  billing_class; pricing_evidence_or_UNKNOWN
  placement_profile; coexistence_incompatibilities
  execution_tier; supported_OSes; secret_store_requirements
  cert_artifact_path; cert_revision; tested_OpenCode_revision
  rollout_status: DISABLED | EXPERIMENTAL | ELIGIBLE | SUSPENDED
```

Unknown fields are **not silently replaced with defaults that grant more access**. Identity and capacity may legitimately remain unknown, with explicit conservative handling. Unknown authorization, endpoint/protocol correctness, model equivalence, data authority, or credential transport is **not** eligible.

## 6.3 Certification levels and release outcomes

| Level | Evidence | Routing allowed? |
|---|---|---|
| `C0_DOCUMENTED` | Source describes what a flow claims to do | **No** |
| `C1_FAKE_FIXTURES` | Local synthetic protocol/credential tests | **No real route** |
| `C2_LIVE_CONNECTOR` | Authorized real upstream auth + multi-turn streamed tool fixture passed | Only exact proven feature/profile when other gates pass |
| `C3_OPEN_CODE_E2E` | Exact pinned OpenCode TUI/headless/tool/compaction/error integration passed | Production candidacy for that pinned package/profile |
| `C4_HARDENED` | OS isolation, stress, crash, security and policy release checks passed | Production-ready within recorded platform/scope |

`C2` does not imply `C3`; a basic text completion does not certify reasoning, tools, vision, caching, stateful continuation or WebSocket. An unsupported feature is disabled for the route, and any request that requires it is rejected; it is **never silently dropped**.

## 6.4 Canonical equivalence rule

A model alias contains its nominal provider/model family, effective model ID/revision, capability profile, protocol/adapter version, output/context limits, billing mode set, data policy and continuation class. **Only routes whose effective contract is explicitly certified as compatible may be candidates**. A different upstream account serving the *same verified effective model* can be another physical route. A different provider or merely similar model name cannot be substituted automatically unless the parent v5.2 equivalence policy explicitly authorizes it and conformance proves it.

For reproducibility-sensitive tasks, retain physical model revision and response provenance. Cross-model escalation or deliberate change belongs to TradeAgent/OpenCode policy, not to opportunistic AgentPool fallback.

## 6.5 Certification record format

Each real connector PR MUST include `compatibility/<integration-key>/<auth-mode>/<pinned-revision>/certification.md` (or a stable equivalent path) containing:

- official terms/developer URLs, section, review date, intended owner and operating context;
- exact auth client/app identity and auth grant type, redirect/device flow details;
- verified principal/tenant/project fields or honest `UNKNOWN`;
- provider-defined capacity/rate/billing-owner dimensions or honest `UNKNOWN`;
- permitted endpoint host/path/protocol and redirect/TLS controls;
- exact model names/revisions/variants and supported capability matrix;
- OpenCode executable/build/commit, native provider package, effective config, transport option and plugin/hook versions;
- success cases, failure-injection logs, sanitized wire evidence, fixture IDs and test hashes;
- known limits/explicit unsupported features, policy reviewer, release decision and expiry/review trigger.

No certification record → **no route eligibility**, even if an experimental call appeared to work.

---

# 7. One central inference selector — normal fast path

## 7.1 Frozen request execution

```text
A. Authenticate OpenCode gateway grant before large-body reading or upstream calls.
B. Validate endpoint/method/size/content type, alias, declared feature needs,
   client grant, data permissions and requested protocol.
C. Load immutable indexed candidate snapshot for this alias and protocol.
D. Apply hard eligibility (model, scope, billing, continuation, policy,
   credential status, readiness, budget, capacity owner, concurrency).
E. Pick one eligible physical route centrally using deterministic sticky/fair policy.
F. Atomically reserve the actual capacity scope and required strict budget.
G. Recheck route/credential/config generations at the dispatch boundary;
   release/rescore if invalid BEFORE sending; bounded admission attempts.
H. Freeze route ID, registration ID, credential epoch, actual model,
   endpoint/TLS profile, continuation, grant and billing scope.
I. Dispatch through its certified adapter/reused transport; stream to OpenCode.
J. Finalize outcome, accounting, reservations and structured health evidence.
```

One attempt has **exactly one frozen physical route**. Existing streams are not reselected, spliced, or transparently replayed. Another eligible route is selected automatically for a **later not-yet-dispatched inference** if the preferred route becomes ineligible. No owner button, OpenCode restart or agent reset is required when the model/continuation contract permits this.

**Admission linearization (race rule):** a route's final admission fence and an owner revoke/disable transaction must be ordered by a bounded shared registration-generation guard. If revoke/disable wins before the admission fence, the attempt MUST NOT dispatch. If admission wins first, the attempt is considered already admitted; revoke then applies the documented `DRAIN` or `ABORT` policy, including cancelling a prepared-but-unsent attempt when `ABORT` is requested. Releasing the guard before network I/O is required; no global lock is held while streaming. This definition, not an impossible zero-time gap between checking a flag and writing a socket, is the concurrency contract tested in U12/RT-14.

## 7.2 Route index and snapshot publication

- The selector reads an **immutable, atomically published candidate snapshot** keyed at minimum by `(logical_model_alias, downstream_protocol)` and prefiltered by certified capabilities. Dynamic grant-specific policy, capacity, cost and credential epoch are rechecked per request; do not materialize every combination of grants/models.
- Snapshot rebuilding occurs on registration/connector/route/model/permission/credential/health/catalog changes, **not** by scanning SQLite for every inference. Maintain bounded incremental invalidation or coalesced rebuilds and monotonic catalog generations.
- `Eligible` is a **derived admission predicate** over the snapshot plus current dynamic constraints, not a static green database flag. Staleness is blocked by final generation/reservation checks before sending.
- Do not hold a global selection mutex through provider I/O, SSE streams, vault refresh, or SQLite accounting operations. Capacity reservation is scoped to the real rate owner/dimension; unrelated routes proceed.
- Deterministic tie-breaking makes race tests reproducible. Affinity prefers a known-good route when compatible with context and cache locality, while bounded fairness prevents avoidable starvation of peers.
- An optional JEV score can only apply a bounded expiring soft penalty to an already eligible route, asynchronously. In absence of calibrated JEV evidence, the baseline selector is unchanged.

## 7.3 No costly activities in the hot path

A normal admitted model request MUST NOT execute:

- slot search/placement or SQLite global registration scan;
- interactive OAuth, browser login, manual token copying, CLI invocation;
- external quota probing or synchronous connector discovery;
- a model call to JEV or to any other advisory model;
- arbitrary subprocess/container startup or per-slot worker discovery;
- unbounded global queue wait, refresh storm or broad health sweep;
- credentials read from untrusted OpenCode environment variables;
- retries whose send/side-effect state is ambiguous.

**Necessary correctness exception:** a strict durable spend ledger may perform a short, bounded, pre-dispatch durable reservation. This is a conscious reliability cost, not permission for unrelated synchronous DB work.

## 7.4 Capacity scope, forecasts and observations

Capacity owners may be account, organization, project, model, deployment, region, pricing tier, time-window, or another provider-defined dimension. Store **measured/reportable scopes**, not a guessed “tokens remaining” integer per slot. Maintain separate `reported`, `observed`, `estimated` and `UNKNOWN` evidence, each timestamped and TTL-bound.

A 429 affecting `account X / model Y` must not automatically cool down every route by vendor name. Conversely, if A and B truly share a project-scoped allowance, marking only A limited is incorrect. When the exact scope is unknown, use a documented conservative inference, cap concurrency, and let live error evidence refine it; do not assert quota independence.

Provider `Retry-After`, reset timestamps and other capacity headers are parsed and bounded by connector-specific tested rules, with clock-skew and stale-data treatment. Unknown capacity does not equal zero *and* does not equal infinite: hard concurrency/budget limits still apply; monitoring can degrade priority without fabricating available quota.

## 7.5 Hard budgets — two declared modes, never silently mixed

`STRICT_DURABLE` means a **crash-resilient spending authorization fence**: before upstream send, durably reserve a conservative maximum authorized exposure for the attempted request, based on verified model pricing/billing dimensions and enforced request/output bounds. After outcome, reconcile known actual billable use; on crash or ambiguous outcome, retain a conservative hold until safe reconciliation/owner resolution. Use transactional atomic ledger semantics. A successful request cannot exceed authorized local exposure under the *specified accounting assumptions*; provider billing discrepancies are recorded, not concealed.

`FAST_OBSERVATIONAL` means an in-memory/async best-effort budget with request/concurrency ceilings and eventual cost telemetry, **without** a guarantee of a crash-proof monetary maximum. It is allowed only when the owner explicitly chooses not to request a strict hard spend cap. The UI and grant state the selected budget mode unambiguously. Neither mode authorizes an upstream billing method the workload grant forbids.

If a connector cannot establish a conservative maximum exposure (unknown rates, uncapped output or unbounded side charges), `STRICT_DURABLE` **rejects that dispatch or requires a sufficiently conservative owner-approved bound**; never assume an unknown maximum is zero. A strict ledger is permitted to be slower; benchmark it separately. Durable reservations after crash are not simply released because the daemon forgot the actual outcome.

## 7.6 Outbound HTTP connection reuse without secret mixing

- Reuse `http.Transport` instances only by complete certified `TransportProfile`: approved destination host/port, proxy choice, TLS roots/pinning where applicable, HTTP version, client certificate / security identity, dial policy, and isolation requirements.
- Never share a mutable `Authorization` default header or a request body across registrations. Construct per-request headers from the frozen registration; prevent cross-request header leakage through pooling and concurrency tests.
- Avoid redirect-follow behavior that transmits credentials to a new host; connector host/path allowlist is authoritative.
- Apply bounded idle-connection limits, keepalive, request deadlines, streaming backpressure, cancellation and file-descriptor accounting. Do not preopen a connection per dormant slot.
- A connector using client-cert or authenticated proxy credentials that change transport identity **must not** share its transport across incompatible principals.
- Route changes within a single frozen stream are forbidden even if a pooled socket closes.

---

# 8. Streaming, state continuity, retry and partial failure

## 8.1 Protocol support is profile-specific

Ship the narrowest documented, version-pinned OpenCode-facing HTTP profile first (e.g., Chat Completions **or** Responses with the proper native package). Do **not** silently mix the two: they differ in stream events, tool-call structures, reasoning/references and finalization semantics. Add other protocols only after their own test matrix passes. A package label reading “OpenAI compatible” alone is not a certification.

OpenCode's documented V2 custom-provider configuration uses `providers`, `package`, `settings.baseURL`, and a per-model map. `modelID` may differ from the OpenCode-visible alias. The exact generated config MUST be tested against the pinned build, and its upstream model names must be truthful. `AGENTPOOL_CLIENT_TOKEN` represents a scoped **local gateway grant**, never an upstream provider secret.

**Reference (synthetic alias; verify package/build before deployment):**

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "model": "agentpool/verified-example-model",
  "providers": {
    "agentpool": {
      "name": "AgentPool",
      "env": ["AGENTPOOL_CLIENT_TOKEN"],
      "package": "@opencode/ai/providers/openai-compatible",
      "settings": {
        "baseURL": "http://127.0.0.1:43819/v1",
        "transport": "http"
      },
      "models": {
        "verified-example-model": {
          "name": "AgentPool (certified exact model)",
          "modelID": "verified-example-model"
        }
      }
    }
  }
}
```

`verified-example-model` is **not an actual entitlement or production alias**. The shipped model must be replaced with a certified real model/capability entry. For a Responses-capable route, use the **separately tested Responses package and endpoint**, not the Chat package with an opportunistic path change. Verify auxiliary calls (`primary`, `compaction`, `title`, `generate`), tool calls and subagent configuration so no intended inference escapes AgentPool.

## 8.2 Continuation classes

- `STATELESS_SAME_MODEL`: OpenCode sends complete necessary context; switching at a request boundary among certified equivalent routes is allowed.
- `PROVIDER_STATE_BOUND`: a provider response ID, reasoning token, stored previous response, WebSocket prefix or other opaque backend object binds subsequent continuation to a verified compatible physical route/state domain. Route selector must honor the binding or return `PROVIDER_STATE_BOUND`.
- `NONPORTABLE`: same-session cross-route continuation not demonstrated; do not switch until TradeAgent/OpenCode explicitly restarts from a portable checkpoint or contractually safe prompt state.

Caching, hidden reasoning, tool-call IDs, provider-side stored context and session-level WebSockets require **independent proofs**. Do not treat the fact that OpenCode has a session ID as proof that the provider state is portable.

## 8.3 Attempt lifecycle and send ambiguity

```text
RECEIVED → AUTHORIZED → ADMITTED → RESERVED → PREPARING
                                            │
                        PREPARING → NOT_SENT_FAILED
                                            │
                                            ▼
                                     SENDING → MAY_HAVE_SENT
                                                    │
                                           RESPONSE_HEADERS
                                                    │
                                          STREAMING → COMPLETE
                                             │              
                                   PARTIAL_ERROR / CLIENT_CANCEL
                                                    
SENDING or MAY_HAVE_SENT → UNKNOWN_OUTCOME on ambiguous network failure
```

**Critical distinction:** `NOT_SENT_FAILED` is legal only when the implementation can demonstrate no upstream request bytes were sent and the connector assures no side effect. Once writing has begun—or send state is uncertain—the attempt is `MAY_HAVE_SENT`. A timeout, connection reset, or missing response is **not proof the upstream did nothing**.

- Select B instead of A **before send** if A is blocked or its reservation/epoch invalidates.
- After a partial or ambiguous send, **AgentPool does not automatically replay** on A or B. Return the exact verified failure classification and provenance; preserve unknown billable exposure.
- If OpenCode independently retries the logical model step, that is an **OpenCode-side decision**. AgentPool must not claim end-to-end at-most-once semantics without a verified host-level retry/identity contract.
- If an upstream API offers a documented idempotency mechanism, certify its precise semantics before using it. Do not invent generic idempotency by adding a random header.

## 8.4 Host-side retry contract (mandatory proof)

OpenCode V2 documentation states provider timeout retries and exposes an explicit plugin `retry` hook (which can veto retries). It also exposes request-kind and HTTP/experimental WebSocket hooks. **This evidence does not prove a particular pinned OpenCode build will prevent duplicate billable model requests on every ambiguous failure.** Therefore:

1. For the pinned build, capture transport-attempt traces from OpenCode direct-to-fixture and via AgentPool for early failure, timeout **before send**, timeout **after send**, partial stream, malformed SSE and 429.
2. Record whether OpenCode retries, how many attempts it schedules, whether the message/prompt/tool state remains identical, and whether it can be configured or controlled via its documented hook.
3. If the default retries an ambiguous failure, install a **narrow, owned OpenCode-side retry policy/plugin** that suppresses such retries when the error classification can be conveyed and parsed with verified semantics; this plugin belongs to the OpenCode host configuration, **not** AgentPool agent control.
4. If reliable classification cannot be conveyed through the pinned interface, choose a **conservative no-automatic-retry** policy for affected failure classes. If even that cannot be enforced, do not certify the impacted provider/transport profile as preserving nonduplicate semantics; expose a documented NO-GO and advance unrelated features.
5. Do not secretly compensate by having AgentPool manage sessions, replay model turns, or launch a new OpenCode worker.

An owner-approved manual retry with explicit duplicate/spend risk acknowledgement is different from an invisible gateway replay and remains outside the AgentPool automatic selection path.

## 8.5 Streaming invariants

- Pass through exact accepted protocol events, preserving ordering, content, role, tools, reasoning signatures, usage, finish/error semantics, and termination markers. A proposed adapter normalization must pass native conformance fixtures before eligibility.
- Small bounded buffers; do not accumulate the entire upstream response before delivering it. Flush frames correctly; avoid coalescing away time-sensitive chunks.
- Respect downstream cancellation and upstream context cancellation, but distinguish *client cancelled* from *provider definitely stopped billing*.
- Enforce maximum request-body size, frame size, header size, bounded stream lifetime/idle intervals, active stream count and read/write deadlines. Make limits configurable and report them accurately.
- A downstream client stall must not create unbounded buffers, goroutines or retained requests; apply bounded backpressure, then controlled timeout/cancel if limits are reached.
- After downstream response commitment, do not attempt to transform an upstream error into a fresh success or new HTTP status that cannot be sent. Emit only protocol-valid terminal error semantics or close with a correctly classified partial failure.
- Never parse provider output as commands to run locally; AgentPool does not execute returned tool calls.

## 8.6 WebSocket decision

**MVP:** Pin OpenCode's provider `settings.transport` to `http` for an exact profile that has passed HTTP conformance. Treat WebSocket support as **separate subsequent certification**, not a hidden requirement to create a new socket router in v2. OpenCode V2 supports WebSocket reuse for some built-in models, and its experimental hooks may affect routing/auth/session behavior; those details MUST be tested on the pinned exact build. A later WebSocket gateway profile must prove connection lifetime, affinity, replay, credential refresh and handoff safety. If the pinned package cannot be safely forced to HTTP, that package is NO-GO until its effective behavior is understood and certified. This is a protocol-specific gate, not a retreat from AgentPool's central inference boundary.

## 8.7 Typed failures

Expose stable **internal** error reasons and a specifically certified downstream mapping, not arbitrary success-like payloads. Minimum internal reasons:

`UNAUTHENTICATED_CALLER`, `GRANT_DENIED`, `UNKNOWN_ALIAS`, `MODEL_CAPABILITY_MISMATCH`, `CONNECTOR_NOT_PERMITTED`, `ROUTE_UNCERTIFIED`, `NO_ELIGIBLE_ROUTE`, `CAPACITY_BLOCKED`, `BUDGET_BLOCKED`, `PRINCIPAL_UNKNOWN` (diagnostic; not always fatal), `CREDENTIAL_UNAVAILABLE`, `REAUTH_REQUIRED`, `PROVIDER_STATE_BOUND`, `PROVIDER_REJECTED`, `UPSTREAM_RATE_LIMIT`, `UPSTREAM_TRANSIENT`, `UPSTREAM_PERMANENT`, `NOT_SENT`, `UNKNOWN_UPSTREAM_OUTCOME`, `PARTIAL_STREAM`, `CLIENT_CANCELLED`, `UPSTREAM_PROTOCOL_ERROR`, `SECURITY_POLICY_DENIED`.

Each error mapping must state: request sent? model identity? retry-safe? how much stream committed? budget hold outcome? correct HTTP/Chat/Responses shape for OpenCode? The client must never interpret an error sentinel as actual model text. Do not expose unredacted upstream provider errors, headers or credentials to untrusted clients.

---
# 9. Database and state implementation — additive, transaction-safe

## 9.1 Persistence strategy

Reference persistence: **SQLite WAL** for one local AgentPool daemon, normalized durable control records and strict spend ledger (if enabled); in-memory immutable routing snapshots and scoped atomics for frequently read routing state; protected external secret custody. SQLite permits simultaneous readers but **serializes writes**. Keep writes short, perform network/UI/vault waits outside write transactions, and benchmark strict-budget contention separately.

**Migration authority:** v5.2 defines conceptual `providers`, `registrations`, `connector_certifications`, `physical_routes`, grants, budgets and ledgers. Builders MUST inspect actual repository migrations before generating DDL. If an already-deployed table exists, evolve it additively and transactionally; do not blindly paste conflicting `CREATE TABLE` or duplicate parent objects. The following is the v2 **logical canonical schema**, not permission to replace deployed parent data.

## 9.2 Canonical schema contract

```sql
-- SQLITE EXAMPLE. IDs are opaque immutable UUID/ULID-like values, not labels.
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS environment_slots (
  id                   TEXT PRIMARY KEY,
  slot_sequence        INTEGER NOT NULL UNIQUE CHECK (slot_sequence > 0),
  owner_id             TEXT NOT NULL,
  compatibility_profile TEXT NOT NULL,
  state                TEXT NOT NULL CHECK (state IN ('ACTIVE','QUARANTINED','RETIRED')),
  generation           INTEGER NOT NULL DEFAULT 1,
  created_at           TEXT NOT NULL,
  updated_at           TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS enrollment_attempts (
  id                   TEXT PRIMARY KEY,
  owner_id             TEXT NOT NULL,
  owner_request_key    TEXT NOT NULL,
  integration_key      TEXT NOT NULL,
  connector_revision   TEXT NOT NULL,
  state                TEXT NOT NULL CHECK (state IN (
                        'CREATED','AUTHORIZING','VERIFIED',
                        'AWAITING_OWNER_DECISION','COMMITTED',
                        'FAILED','CANCELLED','EXPIRED')),
  generation           INTEGER NOT NULL DEFAULT 1,
  auth_nonce_digest    TEXT NOT NULL,
  expires_at           TEXT NOT NULL,
  staged_secret_ref    TEXT,
  result_registration_id TEXT,
  result_slot_id       TEXT,
  failure_code         TEXT,
  created_at           TEXT NOT NULL,
  updated_at           TEXT NOT NULL,
  UNIQUE (owner_id, owner_request_key)
);

-- One row for each CURRENT OCCUPANCY. Deleting it frees a logical position.
-- Historical records live in registration/audit/event tables.
CREATE TABLE IF NOT EXISTS slot_occupancies (
  slot_id              TEXT NOT NULL REFERENCES environment_slots(id),
  integration_key      TEXT NOT NULL,
  registration_id      TEXT NOT NULL UNIQUE,
  generation           INTEGER NOT NULL DEFAULT 1,
  created_at           TEXT NOT NULL,
  PRIMARY KEY (slot_id, integration_key)
);

CREATE TABLE IF NOT EXISTS credential_registrations (
  id                   TEXT PRIMARY KEY,
  owner_id             TEXT NOT NULL,
  integration_key      TEXT NOT NULL,
  principal_id         TEXT,
  principal_verification TEXT NOT NULL CHECK (
                          principal_verification IN ('VERIFIED','UNKNOWN')),
  secret_ref           TEXT NOT NULL,
  credential_generation INTEGER NOT NULL DEFAULT 1,
  state                TEXT NOT NULL,
  connector_revision   TEXT NOT NULL,
  created_at           TEXT NOT NULL,
  updated_at           TEXT NOT NULL
);

-- Never point route key fields at slot IDs as a substitute for registration,
-- principal, protocol, billing or capacity owner.
CREATE TABLE IF NOT EXISTS slot_catalog_outbox (
  event_id             TEXT PRIMARY KEY,
  catalog_generation   INTEGER NOT NULL,
  affected_registration_id TEXT,
  event_kind           TEXT NOT NULL,
  created_at           TEXT NOT NULL,
  applied_at           TEXT
);

CREATE INDEX IF NOT EXISTS ix_slots_owner_state_seq
  ON environment_slots(owner_id, state, slot_sequence);
CREATE INDEX IF NOT EXISTS ix_enroll_owner_state_expiry
  ON enrollment_attempts(owner_id, state, expires_at);
CREATE INDEX IF NOT EXISTS ix_occupancies_integr_key
  ON slot_occupancies(integration_key, slot_id);
CREATE INDEX IF NOT EXISTS ix_registrations_owner_state
  ON credential_registrations(owner_id, state);
CREATE UNIQUE INDEX IF NOT EXISTS ux_outbox_generation_event
  ON slot_catalog_outbox(catalog_generation, event_id);
```

**Add real foreign keys/constraints in finalized migrations:** Where the repository permits creation-order-independent migration, add `slot_occupancies.registration_id REFERENCES credential_registrations(id)` with a deferrable transaction strategy or choose inserts in correct order. Enforce that occupancy and registration share the same `owner_id` and `integration_key` with checked transaction code and/or database constraints/triggers. The example schema is intentionally **not a production migration to paste unchanged**: its job is to fix names, relationships and uniqueness for the implementation plan, not assume a pre-existing parent schema.

**Active record vs history:** committed registrations, tombstones, audit entries, attempts and event generations are preserved after occupancy deletion. No uniqueness constraint on historical `(slot_id, integration_key)` should prevent future reuse. Unique verified-principal scope must be connector-defined and prevent unapproved duplicates; do not create a fake global unique constraint on a nullable/unverifiable principal.

## 9.3 Slot sequence allocation

Within the same short `BEGIN IMMEDIATE` commit transaction, issue an insertion using a durable monotonic sequence source. Options: an `AUTOINCREMENT`-backed allocator table or an atomic next-sequence row. **Never** use unguarded `SELECT MAX(slot_sequence)+1` outside a write transaction; never renumber existing IDs on delete/restore. Test 100 simultaneous commits and daemon crash after committing a new slot but before responding to the UI.

## 9.4 Outbox and snapshot recovery

The same transaction that commits a registration/route/grant modification MUST append a catalog-outbox or equivalent versioned invalidation event. A background publisher rebuilds affected immutable snapshots and atomically installs them. If AgentPool crashes between SQLite commit and snapshot publication, restart deterministically rebuilds or replays until in-memory state matches the authoritative durable generation. A registration must **not** become inference-eligible while secret custody finalization/certification is incomplete, even if a snapshot contains its metadata.

## 9.5 Hard-limit ledger discipline

Store strict-budget grant balances, reservations, uncertain amounts, and final spend in the durable ledger specified by v5.2. **Do not duplicate a second budget counter inside slots.** Reserve under one transaction that guarantees no competing caller can overbook that strict grant/capacity billing scope. Use bounded busy-timeout / `SQLITE_BUSY` treatment and explicit `BUDGET_BLOCKED` vs infrastructure-contention errors; do not hold locks while talking to providers. Crash recovery must preserve ambiguous reservations until authoritative reconciliation or conservative expiry rule documented by the connector and approved by the owner.

---

# 10. Control interface, UI and local security

## 10.1 Surfaces

The inference server and owner control server are **different local listeners/origins and different authentication authorities**. The dashboard is owner-only infrastructure management. Inference endpoints remain browser-unfriendly, grant-authenticated and subject to v5.2's strict browser `Origin` rejection. Do **not** weaken the inference interface to enable the dashboard. Owner UI uses its own authenticated control endpoints.

**HTTP route names in this section are illustrative design names; concrete API paths must be frozen with matching OpenAPI/API tests in G2/G3.** This keeps paths changeable until their pinned implementations are validated without leaving permissions/behavior ambiguous.

Proposed surface separation:

```text
127.0.0.1:<data-port>     /v1/chat/completions OR /v1/responses
                           /healthz   (minimal, nonsecret)
                           [no owner credential mutations]

127.0.0.1:<control-port>  /admin/session/*
                           /admin/integrations/catalog
                           /admin/enrollments/*
                           /admin/slots/*
                           /admin/registrations/*
                           /admin/routes/diagnostics
                           /admin/audit/*
                           [owner-admin authentication required]
```

## 10.2 Browser protections

- A loopback-bound server is **not** automatically private from malicious pages. Authenticate owner sessions; use anti-CSRF tokens, `SameSite` cookies where appropriate, strict `Host`/`Origin` and CSRF validation, anti-DNS-rebinding behavior, session expiry, CSP and proper escape/sanitization of provider-returned labels/errors.
- Sensitive owner writes MUST reject untrusted content from model prompts, repository files, arbitrary tools and inference headers. A workload grant cannot be promoted to an admin session.
- OAuth callback authorization is based on one-time transaction state/PKCE/issuer binding and owner enrollment identity; it is **not** a blanket way to bypass admin restrictions. Native loopback redirect details must follow the connector's official registration and OAuth requirements.
- Secrets supplied by owner through the UI are transmitted only over the approved local control channel and never echoed, saved to web storage, exposed in browser errors, or placed in long-lived form state. Display only safe masks and verified principal attributes.
- Remote administration is **out of scope by default**. If owner needs remote dashboard, submit a separate explicit authenticated TLS/network exposure/security design and tests, not a casual bind to `0.0.0.0`.

## 10.3 Minimum dashboard screens (implementation acceptance)

| Screen | Required content/actions | Misleading behavior forbidden |
|---|---|---|
| **Overview** | Logical slots, total registrations, eligible actual routes, disabled modes, shared capacity scopes, recent errors, active streams | Treat all registered accounts as usable |
| **Add connection** | Two groupings: Subscription/API; select permitted integration/auth mode; scope/terms/billing notice; authenticate; verified identity confirmation | Generic login button pretending all subscriptions work |
| **Environments** | Slot A/B/... with integration occupied/vacant positions, owner-masked principal, committed-vs-preview distinction | Process/sandbox icons claiming each slot has one |
| **Accounts & routes** | Registration ID, integration key, identity confidence, secret/refresh status, capacity owner evidence, exact available models, reasons routes blocked | Treat cooldown as empty slot or duplicate key as new capacity |
| **Diagnostics** | Per-route eligibility predicate with safe reasons, last evidence times, test certificate revision, admission/stream outcome codes | Plaintext prompts/tokens or invented quota values |
| **Administration** | Disable, reconnect, disconnect, remove, rotate secret where supported, inspect audit, strict budget mode and controlled drain/abort | Deleting active credentials without effects explained |

After successful enrollment display **“Registered in Slot A — route validation pending”** until all required route tests/entitlement checks pass; only then **“Eligible for [exact tested model/profile]”**. A login may remain registered while all routes are ineligible. That is intentional and truthful.

## 10.4 UI ↔ data-plane independence

Web UI availability cannot be required for a healthy inference call. UI rebuilds, dashboard tab changes, an abandoned OAuth flow, or enrollment of an unrelated integration cannot interrupt a frozen stream. Conversely, an OpenCode request cannot create slots or mutate account registration through HTTP paths or clever model-generated text.

---

# 11. Failure isolation and operational recovery

## 11.1 Scope failure to the actual dependency

| Failure | Required blast radius / response |
|---|---|
| Credential for Google-like registration X revoked | Deny new X routes; do not touch Anthropic-like registration Y in same slot |
| One endpoint has 5xx | Degrade only affected endpoint/deployment or proven shared scope; retry semantics follow §8 |
| 429 for shared project | Update that project's scoped capacity observation across all proven connected registrations |
| Transient vault outage | Deny admissions that cannot safely resolve secret; no interactive hot-path unlock |
| Enrollment/browser callback late | No committed slot or current registration changed |
| Optional T1 executor crashes | Only routes dependent on that executor become unavailable; supervisor restarts with bounded backoff |
| In-process adapter panic | Recover only at carefully audited gateway boundary, classify the attempt, quarantine suspect adapter/route; recognize full daemon crash remains possible |
| Daemon killed midstream | All affected current streams end without fake completion; restart marks unsettled durable attempts UNKNOWN_OUTCOME when ledger semantics require |
| DB locked/disk full | Preserve already admitted frozen streams where safe; block commits/strict reservations requiring durability; no writes falsely reported successful |
| Provider auth policy no longer permits route | Immediately block new admissions for that connector mode; drain/abort existing according to owner/legal risk requirements |
| OpenCode accidentally bypasses AgentPool | Host integration test fails; apply host egress/config remediation, not a gateway-only patch |
| Two legitimate API keys share billing identity | Capacity/budget linkage shared per actual identity; slots remain independent visual placements |

## 11.2 Shutdown, upgrade and restore

- Graceful shutdown rejects new inference admissions, drains active streams until an explicitly configured bounded deadline, then cancels remaining streams truthfully. Do not send duplicate work to a replacement instance.
- Rolling upgrade of the **one daemon** is not a magical zero-downtime operation. Document connection interruption unless a future separately approved handoff mechanism proves otherwise. Never add cluster/Redis/Kubernetes solely to hide this fact.
- Schema migration runs against backed-up durable data and has forward/backward compatibility and restore tests. Preserve secret handle identity; failed migration should leave original metadata recoverable.
- Startup loads compact metadata, reconciles secrets and strict budget holds, rebuilds immutable catalog snapshot, then marks routes individually eligible when policy, credentials and technical readiness are proven. Do not mark every registration READY merely because the daemon listens.
- Scheduled validation and refresh use bounded concurrency, jitter and priority, not a periodic loop per dormant slot. Suspend expensive probes when unnecessary and preserve a clear `UNKNOWN`/stale health state.

## 11.3 Observability — evidence, not privacy loss

Structured redacted events should include nonsecret request/correlation ID, registration/route ID, alias, credential and route generations, capacity owner scope ID, attempt state, auth mode, error code, duration buckets, usage estimate vs observed, policy revision, and test/cert revision. They must **not** include raw prompts, response bodies, tool arguments, credentials, bearer headers or provider secrets by default. Any diagnostic payload capture needs separate explicit security review and expiration.

Route health observations are per route/capacity owner, not per friendly slot. Historical audit is append-only/tamper-evident to the extent needed for local owner operations; do not claim cryptographic tamper-proof logs without actually implementing/verifying them.

---

# 12. Security requirements and explicit threat model

## 12.1 Trust boundaries

- **Untrusted:** OpenCode's model-generated text/tool output, repository files, malicious local web pages, provider-returned labels/errors, external HTTP responses and arbitrary client headers.
- **Trusted but constrained:** authenticated owner control session, AgentPool daemon under the local owner's OS context, certified connector code, its explicitly allowed OS secret store, pinned OpenCode gateway grants.
- **OS-compromise exposure in T0:** A malicious module executing inside the AgentPool process can potentially reach other credentials accessible to that process. **T0 MUST NOT advertise strong per-account compromise isolation.**
- **T1/T2 additional guarantees:** only after an actual OS security-domain demonstration. Directory permissions under the same unrestricted user and local IPC “mutual authentication” using shared secrets are insufficient as proof against a process running with the same owner privileges.

## 12.2 Hardening checklist

1. Loopback default binding for data and UI; no unauthenticated remote endpoints.
2. Gateway-grant tokens limited by audience, model, workload, cost and expiry; no admin actions.
3. Control-plane admin sessions authenticated, CSRF/Host/Origin/rebinding protected.
4. No arbitrary forward proxy, caller-defined upstream URLs or redirects.
5. OS vault or separately certified encrypted unattended custody; no plaintext SQLite.
6. No raw upstream secrets in OpenCode, environment dumps, crash files, shell args or logs.
7. Connector package supply-chain and signing/provenance reviewed; locked versions and least privilege.
8. Strict request/response size and concurrency limits, bounded streaming buffers and per-connector egress rules.
9. Owner/admin action authorizes registration, disable, reconnect, delete; prompts cannot invoke them.
10. Explicit model/provider/billing/data policy checks before dispatch; unknown connector permission blocks.
11. No automatic ambiguous retries, unsafe cross-account provider state, or hidden model substitution.
12. Cross-OS claims tested separately; Linux capability is not a macOS/Windows proof.
13. Documentation explains local single-daemon compromise risk in T0 and optional T2 cost.
14. All credential/certificate changes invalidate future route admissions through generation checks.

## 12.3 Threat-driven choice of T1/T2

A move to T1/T2 must answer: attacker capability, asset, privilege boundary, exploit route, what secret access is denied, OS mechanism used, proof procedure, memory/latency cost, and fallback if unavailable. Acceptable reasons include proven native-library instability, incompatible runtime state, or explicit requirement that a compromised connector not steal unrelated credentials. **“Every visual slot should feel isolated” is not a sufficient reason.** If strong separation is required but unavailable, fail the requested hardened deployment mode clearly, not the entire logical-slot architecture.

---

# 13. Explicit non-goals (anti-scope-creep)

AgentPool v2 MUST NOT:

- execute TradeAgent jobs or manage OpenCode workers, tools, files, sessions, prompts-as-tasks, trading decisions or agent memory;
- use a second model agent or orchestration layer to perform provider-side inference;
- spawn one Codex/Claude/Gemini CLI per registration or invoke CLI headless requests pretending they are raw model transports;
- implement account farming, browser cookie/session-token theft, third-party client impersonation, a quota-evasion mechanism, or disallowed subscription credential intermediation;
- sell or redistribute other users' credentials; the supported product is one owner using their own permitted connections;
- promise every paid subscription/API-key registration is an independently metered allowance;
- silently substitute models/variants/billing/data authorities for lack of capacity;
- allocate environments because a request failed, a model was called, or a route entered cooldown;
- require Redis, containers, Kubernetes, a global job scheduler or independent slot workers without evidence;
- put interactive login, JEV, remote quota probes, or extra inference on the ordinary dispatch path;
- hide uncertainty by fake success responses, inaccurate completion events, stale credential status, or guessed budgets;
- turn a typed technical limitation into an unauthorized workaround; the correct response is a documented experiment and specific supported/unsupported decision;
- change parent v5.2's unrelated architecture as a side effect of “cleaning up” this addendum.

---
# 14. Milestone plan — exact build order, no architectural wandering

**This is an implementation order, not a claim these gates are complete.** Do not move the slot UI ahead of a functioning gateway merely because the UI is easier to demonstrate. Every gate concludes with a recorded PASS or NO-GO and a traceable test artifact. Builders MAY parallelize synthetic code/tests that do not create architectural divergence, but they MUST NOT merge later phases under a false claim that earlier real integration gates passed.

| Gate | Required implementation | Observable exit gate | Equivalent v5.2 milestone |
|---|---|---|---|
| **G0 — Pin reality** | Inspect repo, pin exact OpenCode V2 executable/package/config, inspect its request/retry/auxiliary behavior, define first allowed API connector and actual model | Direct OpenCode→certified provider trace plus precise package/config/transport manifest; no spec conflict | P0 |
| **G1 — One real inference lane** | One minimal authenticated loopback AgentPool endpoint, one permitted API credential, one truthful alias, exact protocol and bounded streaming; one selector function even with one route | Pinned OpenCode TUI and headless multi-turn fixture with streamed tool calls succeeds through AgentPool; no extra agent surface | P1, early P2 |
| **G2 — Durable registration + safe gateway** | Registration/schema, secret custody, capacity/grants, strict-mode reservation (if enabled), central indexed router, secret refresh if connector needs it | Security/fault tests, no upstream secret exposed to OpenCode, correct restart semantics | P2 |
| **G3 — Logical slots + enrollment** | Provisional auth, verified principal/unknown handling, A/B/C/N allocator, SQLite atomic commit, owner UI and diagnostics | Full EN suite including concurrent commit, duplicate/cancel/crash; **no process per slot** | v2 addendum to P2 |
| **G4 — Managed OpenCode host** | Scoped grant/config provisioning for TradeAgent workers, all auxiliary models, egress restrictions and retries verified | Identical TradeAgent↔OpenCode host API and tool tests against direct/gateway; no model-bypass route | P3 |
| **G5 — Automatic same-model switching** | Two independent eligible registrations, true capacity-owner model, snapshots, affinity, route freezing, proactive admission | In a real uninterrupted pinned OpenCode session A→B on **later** compatible request, no manual button, no lost tools/opaque-state violations | P4 |
| **G6 — Connector/protocol breadth** | Additional certified API/enterprise/local/authorized subscription profile; native adapter only as required | Per-mode policy, transport, streams/tools/errors certification; unsupported modes remain disabled | P5 |
| **G7 — Operating-system and performance hardening** | Linux/macOS/Windows (only targeted OSes), vault headless proof, hostile-page security, fault injection, scale/load benchmarks, rollback | Exact tested OS matrix and release SLO gate; no false strong isolation claim | P6 |
| **G8 — Conditional T1/T2** | Isolated executor or hardened broker only where a specific connector/threat model requires it | Actual OS denial/compromise demonstration + measured overhead; central selector unchanged | Conditional P6 extension |
| **G9 — Optional JEV advisory** | Shadow-mode sanitized risk signal, held-out calibration, opt-in bounded penalties | Demonstrated gain vs deterministic baseline, no hard-eligibility authority, no hot-path inference | P7 |

**Mandatory sequencing:** Complete G0 before real-provider claims; complete G1 before investing in a general worker/executor system. G3 may be prototyped with synthetic fixtures while G1 is in progress, but no claim of production multi-login is made until real connector and gateway gates both pass. G8 is **not** a prerequisite to T0 deployment and is not performed merely because slots exist. G9 never blocks release of the deterministic router.

**PR definition of done:** code + passing tests + pinned dependency versions + a sanitized evidence bundle + updated actual support matrix + the precise parent/v2 clauses satisfied + README/dashboard changes + upgrade/rollback note. A checkbox checked without the evidence file does not count as PASS.

---

# 15. Test harness and evidence format (required before all experiments)

## 15.1 Four distinct test environments

- **H1 — Local deterministic upstream fixture.** Tiny controlled HTTP service with Chat or Responses fixtures, tool events, fragmented SSE, pauses, 429/401/5xx, connection resets, intentional post-send ambiguity, malformed frames, slow clients, cross-credential authorization capture. It must log **nonsecret** received attempt counts and a hash of sanitized request identity. No external billing.
- **H2 — In-process/unit property suite.** Fake credential broker, fake capacity owner, deterministic clock, SQLite temporary database, race/concurrency harness, and executable route manifest fixtures; no provider terms claims.
- **H3 — Real pinned OpenCode.** Exactly recorded executable version/build/commit, provider package/runtime version, isolated effective config, TUI/headless trace, plugin hooks and retry behavior. Direct-to-fixture control and AgentPool-to-fixture experimental runs use the same session/tool scenario and payload class.
- **H4 — Permitted real upstream account.** Owner-approved, documented connector mode; narrow low-cost live conformance with sanitizer; confirm real model ID/capability/protocol without storing prompt secrets. Do not run live against prohibited subscription modes.

**Each result must identify H1/H2/H3/H4.** Synthetic PASS does not become live PROVEN. All load tests must identify fixture vs real provider; timing conclusions about gateway overhead should use H1, not fluctuating upstream traffic.

## 15.2 Required evidence bundle per test

```text
evidence/<gate>/<test-id>/<run-id>/
  manifest.json           # UTC time, git SHA, OS/kernel, CPU/RAM, runtime versions
  test-command.txt        # exact invocation, redacted argument values
  test-result.json        # PASS/FAIL/SKIP with reason; test assertions and counts
  trace-redacted.jsonl    # IDs, status, epochs, routing/attempt state, NO prompts/tokens
  observations.md         # what was observed vs what was only assumed
  artifact-hashes.txt     # hashes for included fixtures and binaries where applicable
  environment.md          # feature flags, grant/strict-budget mode, effective config
  remedy.md               # required if FAIL/UNKNOWN, exact next experiment and owner
```

The tool/CI running the test MUST store the report and link its committed CI artifact. A source URL checked for connector policy also records its access date, policy/version section and direct reason for authorization decision. Never include real credentials or private prompts in the evidence bundle. Real accounts are exercised only with permitted controlled fixtures and bounded spend.

## 15.3 Standard procedure for every uncertainty

1. Write the falsifiable claim, scope, pinned versions and **expected observation**.
2. Prepare control/baseline and injected variant with the same input.
3. Execute with explicit start/end and record attempt count, time, state transitions and final client observation.
4. Check both positive behavior and a malicious/negative case.
5. Mark `PASS`, `FAIL`, `NOT_RUN`, or `INCONCLUSIVE` without euphemisms.
6. Apply the **decision matrix in §16**. Do not silently replace it with a speculative design.
7. Attach evidence; if blocked by unavailable provider authorization, mark only that connector mode `POLICY_UNSUPPORTED` and continue other approved connectors.

---

# 16. Technical uncertainty → exact experiment → deterministic decision

This section is the heart of the “no technical limitations become ambiguous product requirements” rule. **Do the experiment before implementing a workaround.** Each experiment is a release gate for the affected feature only; other independently certifiable paths stay on schedule.

## U01 — What exactly does the pinned OpenCode build send?

**Setup:** Pin executable SHA/build, V2 native provider package, `opencode.jsonc`, one synthetic model alias, transport `http`. Use H3→H1 directly, then H3→AgentPool→H1. Instrument request method/path, content type, request kind and every stream event; redact messages.

**Run:** identical TUI and headless 3-turn coding fixture: plain text, tool call/tool result, and follow-up; then compaction/title/generate if enabled by that runtime. Repeat across server restart and OpenCode resumption.

**PASS:** request paths and schemas fully characterized, advertised package works in both paths with semantically equivalent tool events, and all intended generation calls traverse AgentPool. **FAIL decision:** choose the exact documented native package/protocol matching observed requests; update only its tested adapter/config; do not implement a generic guess-based proxy. If no viable package exists, that OpenCode build/profile is NO-GO while other builds may be evaluated.

## U02 — What does OpenCode retry and what can the host veto?

**Setup:** H1 records a unique upstream attempt per receipt. H3 uses a documented `retry` hook configured first to observe, then to veto a selected class. H1 injects failure (a) before any body received, (b) after full request received but before headers, (c) after first valid stream event, (d) 429, (e) header timeout, (f) chunk timeout.

**Run:** for each case, compare OpenCode direct fixture and through AgentPool; capture attempts by same logical turn, hook events, status, errors, tool effects and client transcript.

**PASS:** every retry class is known and ambiguous cases can be suppressed by supported host policy; AgentPool adds no replay. **FAIL decision:** default to host-level conservative no-auto-retry for affected classes. If not enforceable, mark the profile unsafe for guaranteed nonduplicate billing/tool behavior; do not build an AgentPool model-turn recovery engine.

## U03 — Is provider-bound state portable between registered credentials?

**Setup:** Two permitted live or fully faithful isolated upstream credentials A/B serving exactly one certified model. Construct a provider-state-bearing result (response ID, reasoning signature, cached context or native stored response) plus an equivalent full-context stateless control.

**Run:** A request1 → continue request2 on A; A request1 → continue request2 on B. Repeat with and without explicit portable context and tools.

**PASS (stateless class):** B accepts continuation with verified equivalent semantics and no forbidden state leakage. **FAIL/UNKNOWN decision:** classify `PROVIDER_STATE_BOUND` or `NONPORTABLE`, pin affinity or reject switch, and leave portable-context restart to OpenCode/TradeAgent. Never infer portability from two text-only calls.

## U04 — Are credentials independently owned by distinct real capacity buckets?

**Setup:** Use official provider metadata/headers/docs, two registered keys known to be in same or different project when test accounts allow, and H1 for synthetic shared-scope behavior.

**Run:** map principal/org/project/region/model/limit dimensions; inject a 429 for owner X; confirm all and only bound routes see scoped cooldown. Verify `UNKNOWN` without inference of independence.

**PASS:** capacity scoping matches evidence; uncertain scopes remain conservative and labeled. **FAIL decision:** link verified shared buckets or mark `UNKNOWN`, reduce admission confidence and retest. No new logical slot is a quota fix.

## U05 — Can identity be verified for the chosen API-key or OAuth mode?

**Setup:** Use only officially documented identity/entitlement endpoint if any, and auth scopes permitted for the connector. Supply valid A, valid B, duplicate A, revoked key, wrong browser account and one deliberately unverifiable key class.

**Run:** compare returned issuer/subject/project, scopes, canonical identity and auth result; verify duplicates and explicit owner confirmation path.

**PASS:** authenticated principals are correctly bound; unverifiable API keys are `PRINCIPAL_UNKNOWN` and cannot falsely inflate capacity. **FAIL decision:** do not invent an identity endpoint or run a billable probe without owner authorization; keep specific `UNKNOWN` policy or disable that mode if identity proof is essential to its security.

## U06 — Is a subscription connector actually authorized to supply inference?

**Setup:** For each proposed subscription mode, collect current official provider terms/developer docs on third-party auth client identity, inference endpoint, account usage, pooling, and usage-limit constraints. Keep a signed human-reviewed connector authorization record.

**Run:** first **documentary permission** review; only if PASS, run minimal official auth+stream tests with that permitted client. No account/browser token extraction and no spoofing.

**PASS:** flow and use are explicitly supportable and tested, for the exact connector scope. **FAIL/UNKNOWN decision:** mode `POLICY_UNSUPPORTED` / `DISABLED`, explanatory UI, no automatic routing; API/enterprise/local alternatives stay in product. An available browser login is not sufficient.

## U07 — Can the target OS vault operate unattended?

**Setup:** Fresh macOS/Linux/Windows target environment as applicable, locked/unlocked, logged-in/logged-out, reboot, service user, secret generation and rotation. Exercise the actual selected OS key store or approved encrypted custody fallback.

**Run:** register secret, close UI/user session, reboot, start daemon through intended service manager, do one authenticated non-billable credential read, rotate, simulate vault outage, restore.

**PASS:** reads/rotation work unattended without leaking secret and blocked/unavailable states fail closed without hot-path prompts. **FAIL decision:** choose and certify an explicit unattended secure custody mode for that OS, or mark unattended deployment unsupported. Never fall back to plaintext DB or OpenCode environment secrets.

## U08 — Does T0 really contain connector failure independently?

**Setup:** Two fake independent registrations in the same slot; one adapter fixture that times out, panics, returns malformed data or exhausts its allotted request resources.

**Run:** keep B streaming while A faults; monitor B's p95 latency, cancellation and registry eligibility. Repeat under race detector/sanitizers and daemon restart.

**PASS:** ordinary local failures remain scoped and buffers bounded. **FAIL decision:** fix T0 containment; for genuine native-library crash/process-global incompatibility, propose T1 for **that connector**, attach reproducer and measured executor cost. Do not attach an executor to every slot automatically.

## U09 — Does T2 provide real OS compromise isolation?

**Setup:** Two credentials under separate proposed hardened boundaries. Run a malicious test process with exactly the privileges available to the lower-trust connector; target the other credential store, control IPC and runtime memory as permitted by safe test setup.

**Run:** attempt reads, calls and privilege escalation from A into B, across macOS/Linux/Windows independently. Verify IPC registration/epoch binding, filesystem, OS identity and outbound restrictions.

**PASS:** OS-level denial proven for scope specified in threat model. **FAIL decision:** mark hardened mode unavailable on that OS, or redesign OS security boundary. Same-UID child PIDs/directories are not accepted as proof.

## U10 — Are HTTP pools safe to share?

**Setup:** H1 endpoint tracks authorization and synthetic principal IDs. Generate 20–100 concurrent alternating requests from A/B through the same allowed `TransportProfile`; include a redirect host, TLS config mismatch and proxy identity mismatch.

**Run:** assert each upstream request carries **only** its own frozen registration secret, redirects never leak it, incompatible transport profiles do not coalesce, cancellation and keepalive reuse are bounded.

**PASS:** zero cross-credential contamination across all concurrency trials; connection reuse produces measured benefit or is harmless. **FAIL decision:** fix per-request auth injection; disable pooling only for affected profile pending proof.

## U11 — Can strict budgets survive crashes without phantom refunds?

**Setup:** SQLite strict ledger, two competing callers, fixed approved budget and conservative upper-bound request cost; H1 holds requests after receiving them.

**Run:** kill daemon at each phase: before reservation, after reservation/before send, after upstream receipt/before headers, midstream, after final usage/before ledger finalize. Restart and attempt more work against same grant.

**PASS:** no double-reservation, no invented refund of possibly billable work, no grant overspend under declared exposure assumptions; ambiguous holds remain until reconciliation. **FAIL decision:** repair transactional ledger and replay recovery; strict budget mode NO-GO until passing. Never weaken to async ledger while retaining hard-cap label.

## U12 — How do route snapshots behave under live edits and races?

**Setup:** H2 repeatedly publishes new immutable registry generations while 100 callers admit requests; interleave revoke, disable, owner budget/grant change and new registration.

**Run:** inject edit immediately before selection, after selection/before send, after send, during stream. Trace frozen generation and final eligibility decision.

**PASS:** no request dispatched after a blocking policy change has taken effect at the authoritative admission fence; already-sent streams follow documented drain/abort and are never rerouted. **FAIL decision:** correct final generation validation/atomic reserve ordering; do not fix by serializing all streams under one global mutex.

## U13 — How much memory do dormant slots consume?

**Setup:** H2 load 1/10/100/1,000 synthetic logical slots and equivalent unactivated registrations (and separately 0 vs a fixed number of active routes). Keep same daemon target and idle window.

**Run:** capture process count, threads/goroutines, RSS/heap, descriptors, open sockets, periodic CPU and background wake-ups over a fixed observation window; compare against zero-slot baseline.

**PASS:** process count unchanged by logical slot growth; no per-slot timers, sessions or idle network connections. Metadata memory grows boundedly with stored records. **FAIL decision:** remove accidental per-slot runtime allocation; a hardcoded small slot cap is not an acceptable fix.

## U14 — Does streamed backpressure remain bounded?

**Setup:** H1 sends thousands of bounded valid SSE frames and then large responses. H3/H1 clients read normally, at 1 byte/sec, and pause for 30 seconds; test downstream cancellation and provider cancellation acknowledgements.

**Run:** measure RSS, goroutine count, producer backpressure, buffer high-water mark, cancel propagation, final client error and unknown billable exposure.

**PASS:** bounded memory/work and controlled timeout/cancel, no fake complete SSE sequence. **FAIL decision:** fix streaming adapter/backpressure; do not buffer an entire response or silently drop frames.

## U15 — How expensive is the real gateway?

**Setup:** H1 simulates identical deterministic upstream response and controlled delays for direct and AgentPool runs on the same dedicated host, warm transport and frozen model; separate fast-budget and strict-budget modes.

**Run:** 1, 20, 100 simultaneous client levels with at least 10,000 measured requests per configuration after warmup (or a documented statistical equivalent for long streams). Collect p50/p95/p99 admission time, extra time-to-first-byte, total throughput, CPU, RSS, allocations, goroutines, active sockets, stream jitter and cancellation tail. Run ≥3 repeated sessions and record distributions, not one lucky run.

**PASS rule:** no unbounded scaling, correctness tests remain green and performance meets the explicitly ratified project SLO defined in §17. **FAIL decision:** profile CPU/allocations/locks, apply smallest fix and rerun controls. Never assert “3x faster” without this test.

## U16 — Can WebSocket be disabled and later added safely?

**Setup:** H3 pinned V2 provider profile with `settings.transport: "http"` and experimental WebSocket hooks as documented; local fixture records actual connection protocols.

**Run:** prove no WebSocket traffic in certified HTTP profile; exercise HTTP tool/compaction/timeout behavior. Separately (later) test a WebSocket profile including request2 on same socket, credential rotation, route A→B boundary, provider state and reconnect.

**PASS HTTP:** emitted traffic stays on certified HTTP transport. **FAIL:** choose a different supported package/build whose HTTP behavior is provable; don't pretend a WebSocket was routed by an HTTP-only gateway. **PASS WS (later):** complete request/session/affinity proof. Otherwise leave WS disabled.

## U17 — Can registration and vault two-phase commit recover every crash cut?

**Setup:** Fault-inject after staged vault create, verified callback, before SQLite write, after occupancy commit, after outbox write, before secret finalize and during GC. Run a concurrent same-key enrollment.

**Run:** restart after each cut, execute reconciler twice, verify slots/occupancies/secrets and audit state; fire delayed callback using old generation.

**PASS:** either exactly one committed protected registration or clearly failed/expired attempt; no ghost permanent vacancy, credential overwrite, leaked orphan, resurrected callback or double slot sequence. **FAIL decision:** fix idempotent reconciliation/transaction fencing before enabling UI.

## U18 — Can the local browser attack management?

**Setup:** Host hostile local HTML in a distinct origin; attempt CSRF POST, DNS-rebinding host, forged Origin/Host, no-cookie writes, stale admin session, cookie replay and malicious provider display text.

**Run:** try add/disable/revoke/rotate/delete and OAuth callback mismatch, then attempt same mutation with only an OpenCode workload grant.

**PASS:** every unauthorized state mutation rejected, no secret revealed, legitimate owner workflow still functions. **FAIL decision:** control-plane release NO-GO until fixed.

## U19 — Can a repo/tool prompt bypass AgentPool?

**Setup:** Isolated TradeAgent-owned OpenCode process with project config/plugins/environment intentionally trying to select another provider, use an alternate small model, title/compaction call or direct upstream credential.

**Run:** inspect effective config and host network egress for all primary and auxiliary model calls, plus subagents; test a deliberate disallowed direct-model call.

**PASS:** every intended model generation goes through AgentPool or is explicitly denied/disabled; no upstream credentials in untrusted worker env. **FAIL decision:** repair TradeAgent's host container/config/egress gate, not by giving AgentPool tool authority.

## U20 — Does connector policy stay current?

**Setup:** For each real connector release, review and record provider primary-source docs/terms, specific account type, authorization client, use, limits and model endpoint; compare to last accepted certification.

**Run:** simulate terms status changing to prohibited, an expired policy record, a revoked permission and a modified connector version.

**PASS:** affected routes leave eligible catalog before new admissions; UI explains `POLICY_UNSUPPORTED` or revalidation; other connectors unaffected. **FAIL decision:** implement cert invalidation/release gate, not a global ban on the subscription *category*.

---
# 17. Performance release targets and benchmarking discipline

## 17.1 What “faster, better, less resource-intensive” means

**Scope of measurement:** added AgentPool overhead against the *same endpoint/model*, not provider token-generation speed. Use H1 to remove model/network variance. Test warmed connections and a cold-start profile separately; **never mix them into one blended number**. Run representative CPU contention plus idle target hardware. No speculative speed multiplier is an accepted benchmark.

**Initial G7 engineering targets — prospective, not observed or universal guarantees:**

| Metric | Initial target | Method / caveat |
|---|---|---|
| Local gateway admission, `FAST_OBSERVATIONAL` | p95 **≤ 5 ms** at 20 concurrent clients | H1, warm config, no upstream model delay, exact target host pinned at G0 |
| Incremental time to first streamed event, fast mode | p95 **≤ 15 ms** at 20 concurrent clients | Same deterministic H1 events direct vs gateway, ≥3 runs |
| Incremental first-event delay in `STRICT_DURABLE` | p95 **≤ 30 ms** at 20 concurrent clients | Includes required SQLite budget reservation; do **not** silently disable durable mode |
| Scaling check | No unbounded memory/goroutine/descriptor increase after teardown at 1/20/100 concurrent callers | Record high-water and post-idle baselines, investigate retained resources |
| Dormant logical slots | **Zero additional live worker processes and zero per-slot open upstream connections** for 1→1,000 slots | H2/H1; metadata bytes may grow with N; no invented constant-zero memory claim |
| Credential integrity | **Zero** authorization bleed, wrong-key upstream requests or plaintext secret leak | Repeated concurrency and hostile redirect tests |
| Stream correctness | **Zero** lost/reordered/duplicated tool/finish events in certified fixture set | Compare canonical event transcripts, include partial failures |
| Admission correctness | **Zero** disallowed route, double-occupancy, unsafe unapproved model substitution or AgentPool ambiguous replay | Property/race/integration tests |

**These targets become the project's benchmark contract for G7 once the owner approves the reference host and workload in G0.** If the target cannot be met on the reference host, submit the measured trade-off and minimal remediation; do **not** falsify numbers, skip a relevant budget mode, or redefine PASS autonomously. A target failure is a performance-gate failure, not a reason to change the product contract. Different hosts may have separate reviewed targets documented as profiles.

## 17.2 Precise measurement rules

- Pin CPU, OS/kernel, Go version, compiler settings, SQLite journal/synchronous mode, disk, power mode, OpenCode build, fixture revision, transport keepalive configuration and resource limits.
- Direct and gateway benchmark the *same* request schema, response stream, upstream fixture service, TLS/profile, concurrency and payload sizes. If direct uses a different model/protocol, numbers are invalid.
- Use monotonic clocks; instrument client send, gateway admission, upstream send, upstream first byte, gateway flush, client first event, stream completion and cancellation completion.
- Record p50/p95/p99 and confidence/spread across at least three independent runs; number of samples and warmup included in report. Track absolute + incremental costs; do not subtract noisy percentiles from unmatched runs and call that a measured per-request delta.
- Profile goroutines, heap allocations, idle sockets, mutex contention and SQLite lock waits. Scale to 1,000 dormant slots and 100 concurrent callers, but avoid a hard product MAX_SLOTS.
- Keep exact stream payload sizes, output lengths, error ratios, TLS and proxies constant across compared cases. Test short/long response classes independently.
- Report cold bootstrap/restart/refresh separately; low hot-path latency does not imply fast reauthorization or unbounded concurrent streams.
- If a connector requires T1/T2, rerun the **same benchmark** with and without that executor and publish the measured overhead. Do not claim an isolated connector has T0 performance.

---

# 18. Release-blocking acceptance suite (builder checklist)

Tests below are **executable specifications**; each must become an automated test or a documented manual/real-upstream conformance protocol with evidence. A project runner SHOULD expose stable test IDs and a machine-readable results matrix; descriptive tests are not considered sufficient without assertions.

## 18.1 EN — Logical placement and enrollment

| Test ID | Scenario and assertion |
|---|---|
| **EN-01** | A/B/C reference matrix + Anthropic-example → committed Slot **A**, no D, no new worker |
| **EN-02** | Same initial matrix + Google-example → committed Slot **B**, no D |
| **EN-03** | Same initial matrix + Codex-example → committed Slot **D**, immutable unique new sequence |
| **EN-04** | Other integration present in slot does not consume position for requested integration |
| **EN-05** | Expired/disabled/cooldown credential keeps occupied position until explicit removal |
| **EN-06** | 100 overlapping same-key commits cannot double-book `(slot,key)` and cannot corrupt slot sequence |
| **EN-07** | Replayed idempotency key returns the exact prior commit result after daemon crash |
| **EN-08** | OAuth cancel/expiry/wrong-state/wrong-issuer/late callback changes no committed neighbor registration |
| **EN-09** | Verified duplicate principal defaults to original-slot reconnect; intentional distinct same-principal API key requires explicit allowed path |
| **EN-10** | `PRINCIPAL_UNKNOWN` and capacity `UNKNOWN` stay explicit, never silently independent |
| **EN-11** | OS/integration exclusive placement rules override first-fit without hidden background process creation |
| **EN-12** | Failed provisional login creates no permanent slot/occupancy and leaves no unmanaged staged secret |
| **EN-13** | Crash at every SQLite/vault/outbox commit cut reconciles idempotently after repeated restart |
| **EN-14** | Deleting registration frees current vacancy without deleting audit/use history or in-flight references |
| **EN-15** | 1/10/100/1,000 dormant slot records cause no proportional process, socket, timer or heartbeat spawning |
| **EN-16** | UI shows committed actual slot and route `VALIDATING` separately; a preview is not falsely labelled final |

## 18.2 RT — Route and protocol semantics

| Test ID | Scenario and assertion |
|---|---|
| **RT-01** | Pinned real OpenCode TUI/headless multi-turn tools run through one AgentPool logical provider using certified exact model |
| **RT-02** | Primary/compaction/title/generate and enabled subagent inference traverse AgentPool or are explicitly denied |
| **RT-03** | Chat vs Responses package/endpoint schema and event stream are never mixed; unsupported feature fails explicitly |
| **RT-04** | Existing stream's route, model, credential generation and billing identity do not change during edit/revoke/429 |
| **RT-05** | Pre-dispatch known-blocked A → later eligible certified B automatically without human or OpenCode restart |
| **RT-06** | Shared capacity-owner observation applies across proven shared routes; unrelated owners unaffected |
| **RT-07** | Provider-state-bound cross-account continuation is pinned or rejected `PROVIDER_STATE_BOUND`, not silently reset |
| **RT-08** | Timeout after send, broken SSE, 429 and cancellation never produce AgentPool-generated ambiguous replay |
| **RT-09** | OpenCode built-in retry/hook behavior pinned and unsafe retry cases vetoed or profile NO-GO |
| **RT-10** | Owner model alias X cannot silently serve unrelated Y or different billing/data class |
| **RT-11** | Response tool events, usage/finish metadata, reasoning signatures and errors match certified semantics |
| **RT-12** | Fake upstream malformed/infinite/slow stream respects bounds, backpressure, timeout and cancellation |
| **RT-13** | HTTP reference profile demonstrably disables implicit WebSocket; experimental WebSocket separate gate |
| **RT-14** | Frozen credential/config epoch compare-and-swap denies revoked route before send while permitting correct active drain behavior |
| **RT-15** | New account enrollment/slot deletion/UI crash has no effect on other already admitted streams |

## 18.3 SEC — Credential and owner boundary

| Test ID | Scenario and assertion |
|---|---|
| **SEC-01** | Revoking/refreshing A does not mutate B or another integration in same logical slot |
| **SEC-02** | Credential never exposed to OpenCode env/config or gateway/client logs or crash diagnostic artifacts |
| **SEC-03** | Concurrent pooled HTTP requests never cross-send bearer headers or TLS identities |
| **SEC-04** | Hostile page cannot add/revoke/disable accounts through CSRF, DNS rebinding or forged Origin/Host |
| **SEC-05** | Workload grant cannot call owner mutation or request arbitrary upstream host/key/model |
| **SEC-06** | Stale enrollment generation/nonce and spoofed callback cannot commit credentials or steal an occupancy |
| **SEC-07** | Secret custody headless reboot, rotation, loss and migration behavior documented and exercised for each shipping OS |
| **SEC-08** | T0 clearly discloses shared-process compromise risk; T2 isolation claim requires OS attack proof |
| **SEC-09** | Provider auth flow uses permitted client identity, terms and endpoint; uncertified/prohibited modes never eligible |
| **SEC-10** | Provider-supplied labels, errors, filenames or OAuth parameters cannot inject commands/HTML, reveal keys or bypass egress |
| **SEC-11** | Immediate policy/security disable blocks future admissions; existing stream policy consistent with documented drain/abort |
| **SEC-12** | No code path inside AgentPool launches OpenCode, executes agent tools, or reads TradeAgent worktree |

## 18.4 PERF/REC — Capacity, latency and recovery

| Test ID | Scenario and assertion |
|---|---|
| **PERF-01** | Warm direct-vs-gateway p50/p95/p99 at 1/20/100 callers on identical H1 response; fast/strict separated |
| **PERF-02** | 1→1,000 dormant logical slots: constant worker process count, bounded metadata memory, no proportional network sessions |
| **PERF-03** | 30-second client stall, long stream, malformed frame, cancel and slow SSE never create unbounded heap/threads |
| **PERF-04** | Strict durable budget cannot overbook under racing dispatch and crash; ambiguous spend not refunded as definite zero |
| **PERF-05** | Per-capacity-owner reservation has no global lock held across provider call/stream |
| **PERF-06** | Refresh storms are jittered/bounded, no bulk credential refresh on each model request |
| **PERF-07** | DB lock/disk full/daemon kill/restored vault yields honest partially unknown outcomes and route readiness |
| **PERF-08** | A route blocked by policy/cert/credential/cooldown never reappears because of stale snapshot |
| **PERF-09** | A standard model call triggers **zero** slot allocation, interactive login, external quota probes, JEV inference or subprocess spawn |
| **PERF-10** | Multi-process optional executor benchmark compares T0 vs T1/T2 and records explicit justification |
| **PERF-11** | Linux/macOS/Windows OS-specific tests only pass for platforms actually executed; no transferred proof |
| **PERF-12** | Restart applies committed outbox and reconciles unsettled durable attempts without automatic redispatch |

## 18.5 Release verdict logic

A test case can be:

- `PASS`: verified assertions with artifact;
- `FAIL`: assertion violated; affected milestone cannot be called complete;
- `NOT_RUN`: no evidence; affected milestone not complete;
- `INCONCLUSIVE`: instrumentation inadequate or observations conflict; repeat with improved instrumentation, do not mark PASS;
- `NOT_APPLICABLE`: only for a genuinely non-targeted OS/protocol/feature, with explicit scope and reviewer approval. Do not use this to skip a test required by an enabled connector.

**Production certification rule:** All mandatory cases applicable to that *exact shipping connector/profile/OS/mode* PASS. Unknown policy/transport/secret safety is blocking for **that connector/profile**; optional WebSocket/T2/JEV tests are not blockers when those features are explicitly disabled. Synthetic-only PASS never certifies a real subscription. Any security or protocol regression after a connector update revokes its certified state until rerun.

---

# 19. Change-control map: v1 → v2 without vision drift

## 19.1 Explicit replacement table

| v1 section / decision | v2 authoritative replacement | The preserved product behavior |
|---|---|---|
| §1 exact A/B/C/N example | v2 §§0.3, 2.2, EN-01..03 | Same visible slots, same first eligible vacancy/overflow |
| §3 object model | v2 §2 (separates transport and capacity from placement) | Registration, principal, capacity owner, route identity remain separate |
| §4 and §4.1 mandatory dedicated activated-slot worker reference | v2 §§1.2, 7.6, 12.3 | **No worker per slot by default**; optional isolation if proven |
| §4.2 shared slot process compromise implications | v2 §§1.2, 5, 12 | Accurate threat boundary; strong isolation never misclaimed |
| §5 placement before browser OAuth | v2 §§3.1–3.5 | Same automatic final placement; provisional login, short atomic late binding |
| §6 auth state lifecycle | v2 §§3, 5, 6 | Official supported auth, independently stored/rotated credentials |
| §7 single central selector | v2 §§7–8 | One logical provider to OpenCode, automatic eligible route selection |
| §7.1 `SlotDispatch` worker IPC as default | v2 T0 in-process certified `Dispatch` contract; T1 IPC conditional | Preselected frozen route remains authority; no second router |
| §8 dashboard | v2 §10 | Same Add Subscription/API + Environments + Accounts/Diagnostics |
| §9 SQLite relations | v2 §9 | Durable slots and occupancy, but no stale history blocking vacancy |
| §10 per-slot supervisor | v2 §§1.2, 11 | Optional bounded infrastructure supervision only if connector requires it |
| §11 failure table | v2 §§5, 8, 11, 16 | Narrow fault blast radius, no hidden retry or false availability |
| §12 first build slot registry/worker then inference | v2 §14 | Prove actual OpenCode inference lane first; UI/slots next |
| §§13–14 acceptance/drift checks | v2 §§15–18, §19.3 | Equivalent user behavior with stronger end-to-end tests |
| §15 non-goals | v2 §13 | TradeAgent/OpenCode/AgentPool responsibility unchanged |
| §16 locked E worker decision | v2 §1.2 | Logical slot guarantee preserved; process count no longer a product requirement |

**This v2 does not authorize changing v5.2's mandatory automatic eligible route selection, no midstream splice, no unauthorized model substitution, upstream policy checks, strict-budget semantics, OpenCode host boundary, or optional off-path JEV roles.** If source documents disagree, raise a specific conflict with citations and a narrow proposed resolution—do not silently alter them.

## 19.2 Audit-finding closure map — all 32 original risks

| Finding | Addressed in v2 |
|---|---|
| F01 slot not security boundary; F02 needless processes; F03 coupled slot failure | §§1.1–1.4, 5, 11–12, U08–U09 |
| F04 early ghost reservation; F05 browser cookies; F06 identity race; F07 unverifiable API identity | §§3.1–3.5, 5, U05, U17 |
| F08 mixed state; F09 worker restarts | §§4–5, 11, EN/REC suites |
| F10 OpenCode retry/other transport; F11 protocol mismatch | §§6, 8, U01–U03, U16, RT suite |
| F12 SQLite hot path; F13 HTTP pool coupling; F14 worker readiness | §§7, 9, 17, U10–U12, U15 |
| F15 quota-owner confusion; F16 ambiguous replay; F17 backpressure | §§7–8, U04, U11, U14 |
| F18 integration-key confusion; F19 unsafe first-fit; F20 occupancy history; F21 slot sequence race | §§2–4, 9, EN suite |
| F22 headless vault; F23 fake process isolation | §§5, 12, U07, U09 |
| F24 stale UI; F25 policy-ineligible subscriptions | §§6, 10, U06, U20 |
| F26 wrong development order; F27 per-slot polling; F28 unbounded runtime growth | §§1, 7, 11, 14, 17, U13 |
| F29 OAuth/rebinding; F30 provider-state switching | §§3, 8, 10, U03, U18 |
| F31 hidden parent conflict; F32 unclean restart duplicate/false completion | §§0, 5, 8–9, 11, 19 |

No F01–F32 finding is silently discarded. Closure in a **document** is a design proposal; each corresponding test still needs actual evidence before implementation closure.

## 19.3 Fifteen-question anti-drift review — answer YES with evidence

1. Does AgentPool still supply inference only, with OpenCode executing all agents and TradeAgent owning work?
2. Does the exact A/B/C/N user example pass under compatible synthetic integration keys?
3. Do logical slots have **no mandatory per-slot runtime**, process, browser profile or quota bucket?
4. Are all credentials/registrations independently protected and tracked, while T0's compromise limitations are honestly disclosed?
5. Are real identity, capacity owner and model route **separate concepts** with `UNKNOWN` where necessary?
6. Are logins provisional and committed by one short generation-fenced transaction, with safe vault reconciliation?
7. Is placement never invoked by an inference request or upstream rate-limit event?
8. Does exactly one central AgentPool selector choose/freeze the complete physical route for a request?
9. Does OpenCode see one logical provider, not the upstream keys or per-slot services?
10. Is A→B a **later eligible request only**, with model/protocol/continuation and budget constraints honored?
11. Does AgentPool refuse to fabricate successful stream data or replay an ambiguous request?
12. Are OpenCode's own retry, compaction, tools, auxiliary calls and WebSocket possibilities tested on the pinned build?
13. Is each real connector mode separately terms-reviewed and stream-certified before the router can choose it?
14. Are the hot path, dormant-slot footprint, security and strict-budget costs measured rather than assumed?
15. Are optional T1/T2/JEV additions independently justified and unable to take over the agent loop or hard eligibility?

**Any NO on an applicable required item is blocking.** The reviewer must reference code path, test ID and artifact, not a reassuring architectural summary.

---

# 20. Source register, research state and builder handoff

## 20.1 Authoritative project context as inspected on 9 October 2026

- [AgentPool v5.2 parent architecture](https://github.com/Nicolas1bhr/agentpool/blob/main/docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md) — source of inference, routing, budget, continuity and JEV invariants. Repo version inspected: blob SHA `43d934ae5da561358aecbe2a3a764b573438e595`.
- [AgentPool provider/connector policy](https://github.com/Nicolas1bhr/agentpool/blob/main/docs/PROVIDER_POLICY.md) — one authorized owner, independently certified permitted flows; inspected blob SHA `017ee80cf627d68373bfcacbae6ddd792e578f58`.
- [AgentPool roadmap](https://github.com/Nicolas1bhr/agentpool/blob/main/ROADMAP.md) — P0–P7, current stage indicated as an inference endpoint skeleton; inspected blob SHA `1fd1ba65861bb21cd5cf971a037a4905353b4723`.
- User-supplied `AgentPool_Environment_Slots_and_Multi_Login_Implementation_Contract_v1.md` — canonical owner enrollment example and anti-drift constraints that v2 preserves or explicitly replaces.
- `AgentPool_Environment_Slots_Multi_Login_v2_Lean_Runtime_Audit_and_Contract.md` — adversarial F01–F32 analysis and v2 design correction used as basis here.

## 20.2 Primary technical references verified 9 October 2026

- [OpenCode V2 — Providers](https://opencode.ai/v2/docs/providers): native `providers/package/models`, package options, timeouts, retries, WebSocket behavior.
- [OpenCode V2 — Build plugin hooks](https://opencode.ai/v2/docs/build/plugins): model-request kinds, HTTP/WebSocket hooks, retry hook and attempt semantics.
- [OpenCode V2 — Config](https://opencode.ai/v2/docs/config): effective configuration and overriding controls.
- [SQLite — Write-Ahead Logging](https://www.sqlite.org/wal.html): concurrency/serialized write behavior and recovery.
- [SQLite — Partial indexes](https://www.sqlite.org/partialindex.html): alternative uniqueness-on-active-state implementation, where appropriate.
- [OAuth 2.0 Security BCP, RFC 9700](https://www.rfc-editor.org/rfc/rfc9700) and [OAuth for Native Apps, RFC 8252](https://www.rfc-editor.org/rfc/rfc8252): PKCE/state/issuer and supported native redirect patterns.
- [Go `net/http` documentation](https://pkg.go.dev/net/http): transport reuse, cancellation, streaming and flush behavior.
- [OpenAI — Sign in with ChatGPT Terms](https://openai.com/policies/sign-in-with-chatgpt-terms/) and [Anthropic — Claude Code legal/compliance](https://code.claude.com/docs/en/legal-and-compliance): example of materially provider-specific subscription/account intermediation constraints; **not** a global ban on the connector category.

**Currency warning:** These links and repository blob SHAs are research evidence as of the above date. At build/release, recheck exact product/runtime version and current provider terms. The plan is deliberately extensible; a changed provider rule changes **eligibility of that connector mode**, not the immutable A/B/C/N vision.

## 20.3 What is still unproven (and where to prove it)

- Actual deployed OpenCode binary and transport profile → **U01, U02, U16**.
- Real A/B same-model safe continuation and quota-owner relation → **U03, U04**.
- Real subscription third-party permission and auth → **U06, U20**.
- Per-OS unattended vault/security boundary → **U07, U09**.
- Crash-safe secret+ledger handling → **U11, U17**.
- Measured latency/memory/performance → **U08, U10, U13–U15, §17**.
- A running production-ready v2 implementation → **G0–G9 and entire §18 applicable suite**.

**Do not call the design implemented, faster, secure against an OS-level attacker, or provider-certified because this document says how to build it.** It is an explicit blueprint with falsifiable gates.

## 20.4 Exact first instruction to the building team

> First inspect the current `agentpool` repository and record its exact commit and existing migrations. Pin the target OpenCode V2 executable/runtime package and create G0's sanitized direct-provider/H1 trace. Then complete G1: one permitted real API connector serving a pinned, genuine streamed multi-turn tool session from real OpenCode through AgentPool's one inference endpoint, with no secondary agent runtime. Do not create per-slot workers or a general subscription-login system before proving that path. Next implement the independent registration/vault model and deterministic late-commit virtual-slot allocator, verifying EN-01/02/03 and concurrency/crash gates. Preserve v5.2 selection and safety semantics throughout. Every unknown must be resolved by its U-test or explicitly remain a disabled connector/profile, never by silently changing the product.

## Final acceptance statement

**The intended finished system is the owner's original idea, made lighter and more dependable:** From one authenticated AgentPool dashboard, add permitted subscription/API connections as many times as legitimate resources and provider rules permit. AgentPool automatically places each into the earliest compatible logical environment with a free position for its integration, or creates the next slot. These slots do not require background processes, agents, containers or expensive idle infrastructure. Independent registrations feed one central, policy-correct, capacity-aware selector. OpenCode sees only AgentPool's truthful logical model profiles and continues performing its agent work as usual. Each inference request uses one frozen, certified physical route; switching applies only to later eligible requests, not to a partially emitted response. Technical uncertainties have written experiments, pass/fail decisions and safe fallback boundaries, not loopholes or vision drift.
