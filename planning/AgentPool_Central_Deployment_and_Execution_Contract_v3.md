# AgentPool — Central Deployment & Execution Contract
## v3.0 · Planning foundation · One central inference provider · Engineering simplicity without weakened guarantees

> **Document status:** **PROPOSED NORMATIVE DESIGN — READY FOR OWNER REVIEW; NOT YET APPROVED, IMPLEMENTED, DEPLOYED, OR PROVEN.**  
> **Research / repository baseline:** 9 October 2026, Europe/Brussels.  
> **Project:** [`Nicolas1bhr/agentpool`](https://github.com/Nicolas1bhr/agentpool).  
> **Replaces:** `AgentPool_Central_Deployment_and_Execution_Contract_v2.md` **for the central-deployment planning track upon explicit owner approval**. It does **not** replace AgentPool v5.2, Environment Slots & Multi-Login v2, provider policy, or the project roadmap.  
> **Intended reference deployment:** the owner's existing Hostinger KVM 4 VPS, **subject to a real, read-only inventory and measured fit**.  
> **Mutation authorization:** **NONE**. This file authorizes no repository changes, VPS access or modification, DNS/ingress changes, credential enrollment, billing, installation, or production release.  
> **Purpose:** A coherent, durable foundation upon which the remaining AgentPool plans can be completed and an autonomous engineering team can subsequently build the product without reinventing its role or resolving architecture contradictions by guesswork.

---

# Executive contract — read this before anything else

**TradeAgent owns the work. OpenCode executes the agents. AgentPool supplies inference.**

An owner configures OpenCode on a MacBook, desktop, or other permitted machine to use **one model provider named `agentpool`**. OpenCode sends ordinary model inference requests, using a scoped AgentPool credential, to **one central AgentPool** over authenticated public HTTPS. AgentPool selects an **eligible, actually authorized, compatible upstream connection** from a pool of separately registered API keys, legitimately supported account/subscription connections, enterprise connections, and explicitly reachable local-model endpoints. It forwards the request and its stream faithfully. OpenCode continues to own its own session and tools. AgentPool never launches an agent or takes ownership of a task.

The owner manages that pool through a separately authenticated remote dashboard. Registrations are organized into automatically filled **logical environments A/B/C/…/N**. These are grouping/occupancy metadata: **not workers, browser profiles, containers, virtual machines, independent quotas, or active sessions**. A single shared router chooses an eligible route per new inference request. It switches automatically to another compatible registration on subsequent requests when the previous route is unavailable, budget-constrained, or capacity-constrained; it never transfers a live stream or silently changes a model contract.

The reference installation is intentionally modest: **one Go service, one transactional SQLite state database, one owner-controlled credential-root lifecycle, and reuse of the existing safe HTTPS ingress**. The dashboard, admission system, routing logic and credential management are different responsibilities **inside the same authority**, not separate miniature platforms. Auxiliary components are not introduced simply because a feature could be implemented with them.

**Engineering philosophy: complexity is allowed where the real problem is complex; moving parts are not.** Reliability comes from narrow responsibilities, explicit invariants, atomic operations, bounded concurrency, honest error semantics, deterministic recovery and meaningful physical tests. Like a durable engine, it must be understandable, maintainable and repairable without the original builder.

```text
       MacBook                     Desktop                Other permitted hosts
       OpenCode                    OpenCode                     OpenCode
          |                           |                            |
          +---------------------------+----------------------------+
                                      |
                         Authenticated inference HTTPS
                         AgentPool consumer credential
                                      |
                         api.<owner-approved-domain>
                                      |
                       Existing tested TLS ingress
                                      |
       +------------------------------v------------------------------+
       |                    CENTRAL AGENTPOOL                         |
       |                     Hostinger KVM 4                          |
       |                                                               |
       |   One Go service / one routing and authorization authority     |
       |                                                               |
       |   Inference protocol lanes  -> central admission + selector    |
       |          |                        |                             |
       |   frozen route + stream       grants, capacity, budgets        |
       |          |                        |                             |
       |   certified connectors <---- indexed registrations            |
       |                                      |                          |
       |                     SQLite transactional state                |
       |                + encrypted dynamic credential records*         |
       |                         |                                      |
       |                   protected root key                          |
       |                                                               |
       |      Logical slots A/B/C/.../N: metadata only                 |
       +------------------------------+--------------------------------+
                                      |
                           Permitted inference HTTPS
                       (or certified private local route)
                                      |
           Cloud API | approved subscription/account | enterprise | local

       Owner browser --> admin.<owner-approved-domain>
                          separate owner session and permissions
                          SAME central business authority

       Trusted TradeAgent host --> optional scoped grant-issuance API
       TradeAgent --> OpenCode worker control; never AgentPool job control

       * Storage approach is a preferred DESIGN CANDIDATE pending custody
         and crash/recovery proofs; see §8 and amendment C-09.
```

**Definition of success:** OpenCode still behaves as OpenCode; AgentPool still behaves as a model provider. Multiple compatible registrations increase the number of lawful usable inference routes without multiplying the number of idle processes or requiring manual account switching. Everything else is in service of that experience.

---

# 0. Authority, evidence, and how to use this document

## 0.1 Source-of-truth hierarchy

| Priority | Authority | Owns |
|---|---|---|
| 1 | Applicable law, **actual** upstream terms/authorization, proven protocols, explicit owner security/billing limits | What a connector is allowed to do and what is physically possible |
| 2 | [AgentPool v5.2 — Proactive Routing & JEV](https://github.com/Nicolas1bhr/agentpool/blob/main/docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md) | Provider-only boundary; aliases; central routing; continuation; streams; retries; budgets; capacity; optional JEV |
| 3 | [Environment Slots & Multi-Login v2](https://github.com/Nicolas1bhr/agentpool/blob/main/planning/AgentPool_Environment_Slots_and_Multi_Login_Implementation_Contract_v2.md) | A/B/C/N identity, provisional enrollment, atomic late placement, registration/credential generations, related tests |
| 4 | [Provider policy](https://github.com/Nicolas1bhr/agentpool/blob/main/docs/PROVIDER_POLICY.md) | One-owner scope; permitted credentials and auth/inference methods; connector certification |
| 5 | **Approved** amendments in §17 of this document, reconciled into affected parents | Only the precisely named centralization/custody/security deltas |
| 6 | This contract, for central-hosting and remote administration topics not owned by parents | Host topology, deployment, ingress, owner access, recovery, operations, release proof |
| 7 | [ROADMAP](https://github.com/Nicolas1bhr/agentpool/blob/main/ROADMAP.md) and [planning README](https://github.com/Nicolas1bhr/agentpool/blob/main/planning/README.md) | Existing P0–P7 / G0–G9 build gates, planning governance; do not invent a competing program |
| 8 | Later implementation choices | Code organization, dependency versions, release commands, visual design, tuning within the agreed contract |

**Authority rule:** A newer Markdown file does not automatically overrule a specific parent's owned requirements. A section here marked **PROPOSED AMENDMENT** is not yet an authorized parent change. The engineering team must record a conflict, approve the minimum delta, and update/cross-reference the source-of-truth documents before relying on it.

The above parent files were checked against these GitHub blob identities on the baseline date: v5.2 `43d934ae5da561358aecbe2a3a764b573438e595`; slots v2 `936d270ee1ef6d53f66493fbf13bdc40878f953f`; provider policy `017ee80cf627d68373bfcacbae6ddd792e578f58`; roadmap `1fd1ba65861bb21cd5cf971a037a4905353b4723`; planning README `2910c872d7f085662816a0b6f503b47efe1ea83d`. These are snapshots, not permanent version pins. Re-check HEAD, supersession, and relevant vendor docs before construction.

## 0.2 Normative language and proof grades

- **PRODUCT LOCK:** A user-established requirement. Builders cannot replace it with an easier product.
- **INHERITED LOCK:** A requirement already owned by an authoritative parent.
- **DESIGN CANDIDATE:** A concrete proposed engineering decision, ready for review; not yet owner-approved.
- **CONDITIONAL DEFAULT:** A preferred decision with an exact, observable branch condition.
- **FACT UNKNOWN:** Requires host/client/provider inspection, not imagination.
- **EXPERIMENT REQUIRED:** A predeclared falsifiable test is required.
- **BLOCKED:** A named release profile/connector is ineligible until its condition passes. Does not automatically block independent profiles.
- **PROVEN:** Passed an actual dated, version-pinned test with retained safe evidence. Documentation, prototypes, mock-only tests, compile success, and a plausible design do not become production proof.
- **MUST / MUST NOT:** Blocking for the applicable approved scope. **SHOULD:** default with evidenced, owner-reviewed exception. **MAY:** optional, not necessary for release unless specifically selected.

The status of this **document** is distinct from the status of the **running product**. We can approve an architecture without claiming the VPS, connector, authentication, backup, or model semantics have passed a real test.

## 0.3 The reliability law: reduce states, not safeguards

For every additional database, service, process, network hop, queue, worker, global lock or credential class, ask:

1. Which evidenced requirement cannot be satisfied without it?
2. Can an existing transactional boundary or existing service accomplish it instead?
3. What new failure modes, rollback procedures, security boundaries and idle resources does it introduce?
4. How do we **prove** that its benefit exceeds those new costs?
5. Can a replacement engineer diagnose and restore it without secret unwritten knowledge?

Do not delete correctness protections to obtain a tiny benchmark win. Equally, do not implement elaborate recovery for failure modes that a simpler storage or authority model can avoid creating.

---

# 1. Unbreakable product invariants

| ID | Contract | Why it matters |
|---|---|---|
| I-01 | **Inference only:** AgentPool does not run workers, tools, prompts as jobs, task schedulers, memory/handoffs or trading actions. | Protects the product boundary. |
| I-02 | **One OpenCode provider identity:** `agentpool` is the configured model-inference provider. | Clients do not manage upstream registrations. |
| I-03 | **One central selector:** There is no competing dashboard, connector, proxy or slot-level routing authority. | Avoids split-brain routing decisions. |
| I-04 | **Automatic eligible route choice:** Every new request selects under model, permission, continuity, capacity and financial constraints. | Makes AgentPool a real pool, not a manual switch. |
| I-05 | **Model/protocol fidelity:** Never silently substitute model family, revision/variant, capabilities, tool semantics, provider state or billing/data policy. | Preserves caller intent. |
| I-06 | **Immutable physical attempt:** Never switch an in-flight stream or reconstruct ambiguous output as if complete. | Preserves truthful inference and accounting. |
| I-07 | **Logical environments only:** A/B/C/…/N are placement metadata, not allocated execution resources. | Avoids per-slot cost. |
| I-08 | **Independent registrations, evidenced capacity scope:** Multiple accounts and keys can coexist but do not automatically mean independent quotas. | Prevents false headroom. |
| I-09 | **Upstream secrets remain central:** OpenCode holds only AgentPool inference credentials. | Keeps the secret pool out of agents. |
| I-10 | **Separate authorities:** Consumer inference credentials cannot administer, enroll, delegate, reveal secrets or change financial policy. | Prevents privilege confusion. |
| I-11 | **Certified permitted connectors:** Account/subscription modes are a first-class design category; each concrete mode must be independently lawful, supported and proven. | Keeps the intended vision without fabricating rights. |
| I-12 | **Ordinary remote connectivity:** A permitted OpenCode client can use authenticated public HTTPS without a mandatory VPN, sidecar, device approval or local daemon. | Works from Mac, desktop and other machines. |
| I-13 | **Useful remote dashboard:** Owner can do routine pool administration over separately authenticated HTTPS, not SSH. | Makes centralization usable. |
| I-14 | **Honest costs, failures and recoverability:** Unknown quota, provider billing, stream outcome or backup authority must not be presented as certain. | Avoids false confidence. |
| I-15 | **Preserve existing VPS workloads:** Coexistence and bounded resource consumption are release conditions. | AgentPool is not a dedicated appliance. |
| I-16 | **One recoverable authority:** Default topology is one small service and one durable state system; no gratuitous distributed infrastructure. | Promotes comprehensibility. |
| I-17 | **No hidden bypass claim:** AgentPool governs requests addressed to it; TradeAgent governs managed OpenCode egress. | Separates actual responsibilities. |
| I-18 | **No stale-authority resurrection:** A disaster restore is not merely a successful SQLite restore. | Protects revocations and budgets. |
| I-19 | **Critical-path independence:** Dashboard, provider login, expensive probes, optional JEV and unrelated connectors never become synchronous dependencies of healthy inference. | Protects latency and isolation. |
| I-20 | **No false single-host certainty:** Recovery must distinguish invalidating this instance's old grants from proving the old host cannot operate. | Prevents hidden split brain. |

**Reject any feature that makes AgentPool depend on what task the agent is doing or how TradeAgent schedules it.** Understanding model protocol, allowed capabilities, inferred quota evidence, and minimal route-affinity identifiers is not the same as owning agent work.

---

# 2. User journeys: the product as experienced

## UJ-01 — Connect an independent OpenCode installation

1. Owner signs in to the remote AgentPool dashboard and issues a named, scoped inference key.
2. Owner configures the actual pinned OpenCode build with a single `agentpool` provider and advertised, certified model aliases.
3. OpenCode on a MacBook sends ordinary authenticated HTTPS inference and receives normal streamed model responses.
4. Owner repeats on a desktop or other independent machine using a separately revocable inference key if desired.
5. No upstream account credential is placed in OpenCode. No AgentPool client daemon, VPN or device-management program is required.

**Accept when:** both independent clients use the same central AgentPool; streams, tool invocations, multi-turn behavior and failure handling match their certified direct-provider profiles; each key is scoped and independently revocable.

## UJ-02 — Add an API or supported subscription/account

1. Owner opens **Add API** or **Add Subscription**.
2. UI shows actual implemented, permitted connector/auth-mode choices, not a universal subscription promise.
3. Owner performs documented permitted authentication/credential submission; an attempt remains provisional until verified and staged safely.
4. At commit, Environment Slots v2 atomically chooses the **earliest compatible existing slot with a vacant position for that integration**, or creates the next slot.
5. Dashboard distinguishes `REGISTERED` from model-route `ELIGIBLE`. Existing OpenCode configurations remain unchanged merely because another matching credential was added.

**Accept when:** no empty slot causes a process to start; cancellation/replay/identity collision creates no phantom occupancy; route eligibility is independently proven.

## UJ-03 — Normal automatic switch

1. OpenCode requests a certified alias; central selector admits connection A.
2. A later becomes budget-ineligible, authentically capacity-blocked, unavailable or otherwise ineligible.
3. On a **later** compatible OpenCode inference request, selector chooses eligible connection B, without operator intervention.
4. Existing completed output and the OpenCode session remain truthful; no hidden model or upstream-state substitution is introduced.

**Accept when:** different registration is used, same verified effective alias and authorized billing/data scope are preserved, no stream splice or ambiguous replay occurs.

## UJ-04 — TradeAgent unattended managed worker

1. TradeAgent starts and controls OpenCode workers through OpenCode's own host interface.
2. A separately trusted TradeAgent host may request short-lived, **non-escalating** workload inference grants under an owner-issued restricted delegation credential, if that G4 profile has been selected for release.
3. Each worker sees only its own restricted inference grant; the host's delegation authority and upstream credentials remain inaccessible to it.
4. Its normal inference, enabled auxiliary models, compaction, and supported subagent model calls use AgentPool or are explicitly disabled under the managed host's own no-bypass policy.
5. AgentPool never starts, schedules, supervises, recovers, or kills a TradeAgent worker.

**Accept when:** unattended grant renewal/rotation works with real pinned OpenCode; minting more worker tokens does not multiply the principal's budget; compromised worker authorization cannot become host or owner authorization.

## UJ-05 — Failure stays honest and contained

A failed provider registration, network path, vault key, grant or stream results in a typed, limited-scope error. Healthy independent routes continue when their actual dependencies remain healthy. Post-send ambiguity is never hidden behind fabricated success. OpenCode or TradeAgent owns task-level retry, checkpoint and recovery.

## UJ-06 — Restore after host loss

Backups plus separately kept recovery material allow rebuilding or explicitly reauthorizing the pool. The replacement starts in **`RECOVERY_REVIEW`**: old AgentPool-issued grants/sessions cannot be trusted, upstream eligibility and spending are reconciled, and a verified operational cutover addresses any surviving old host. Only then does the owner authorize billable service.

**Accept when:** a backup taken before a revocation or budget spend never silently restores those permissions or allowances; uncertainty is exposed; the team never claims that a recovered DB alone guarantees exclusive active hosting.

---

# 3. Topology and physical deployment

## 3.1 One host, one ordinary public inference provider

The selected reference host is the owner's **existing Hostinger KVM 4**. Nominal plan resources cited by Hostinger are **4 vCPU, 16 GB RAM, 200 GB storage**. These are advertised parameters, not observed free capacity or sustained throughput on the live co-hosted VPS. CPU-credit behavior is plan- and account-dependent; published Hostinger help pages may diverge. Capture real resource and throttling telemetry. Do not assume AgentPool owns unused CPU or disk.

Public endpoints, each requiring independent authorization:

- `https://api.<owner-approved-domain>`: only selected HTTP inference paths and narrowly scoped authorized model metadata.
- `https://admin.<owner-approved-domain>`: owner dashboard, selected approved enrollment callbacks, and narrowly scoped delegated-grant operations if shipped.

The actual domain names, DNS records, certificates, ingress file/container configuration and port assignments are **UNKNOWN UNTIL HOST INVENTORY AND OWNER SELECTION**. Examples are not deployable values. Public TLS is mandatory; directly publicly exposed AgentPool backend sockets are forbidden. IPv6 must be audited explicitly; add AAAA records only for routes with proven equivalent protection.

The daemon/container's listener remains reachable only by the trusted ingress and necessary local privileged control paths. Separate domain names reduce accidental overlap but do **not** create an authorization boundary by themselves.

## 3.2 Runtime shape

**Proposed reference:** one Go daemon/application. Internal responsibilities:

1. HTTP inference transport and certified protocol lanes.
2. One grant/policy/auth core and one route-admission authority.
3. Immutable indexed route catalog with dynamic capacity and spending reservations.
4. One bounded stream-forwarding implementation per certified protocol.
5. One protected registration/credential/refresh lifecycle.
6. One owner-admin adapter and remote dashboard assets.
7. One SQLite transactional state system and protected encryption-key provisioning.
8. One bounded background scheduler for expiring credentials, health and cleanup.

These are modules and task queues **inside one service**, not eight microservices. No slot implies a goroutine, open network connection, process, container, browser or polling timer. No per-client AgentPool process is required.

A crash of this one daemon can terminate current streams and affect dashboard operation. That is an intentional single-instance fault domain; **we do not claim HA**.

## 3.3 Native systemd versus one container — corrected selection rule

**Do not prefer native solely because a loopback socket is reachable.** Once the real host inventory is recorded, score **one** supported deployment profile against mandatory requirements:

1. No regression to unrelated VPS applications, ports, proxy routes, volumes, backups and certificates.
2. Privilege separation: dedicated non-root service identity, read/write path confinement, no access to unrelated credentials, no broad host network or Docker daemon authority.
3. Backend is not publicly reachable and is denied to unrelated service containers/identities as far as the chosen OS threat model can prove.
4. Reliable streaming, restart, update, rollback, bounded drain, restore, metrics and release inspection.
5. Lowest incremental integration complexity with the **actually installed** reverse proxy, service manager and backup convention.
6. Measured resource overhead as a tie-breaker; no speculative preference for a packaging brand.

**Conditional decision:** If a single restricted container integrated into the current host deployment is safer and operationally simpler, select it. If one native non-root systemd service with a private socket provides demonstrably simpler confinement/operations, select native. If neither can meet isolation and coexistence criteria, block deployment and identify the exact failure. **Do not build and maintain both production profiles.**

Container caveat: containers on the same Docker bridge network can communicate; `expose` is not authorization. `127.0.0.1` inside a container is not host loopback. Prefer a dedicated narrowly connected network or proven host-local transport, never publish backend ports for convenience. No Docker socket mounting or privileged containers. Native caveat: filesystem modes and loopback binding alone may not isolate a compromised unrelated same-host process; prove the OS/network boundary actually relied upon.

## 3.4 One established ingress, not a new proxy hierarchy

Reuse the installed TLS ingress **if its version and configuration pass the streaming and isolation tests**. Additional Envoy/Nginx/Caddy/Traefik instances are not justified by default. Require correct host/SNI validation, trusted forwarded headers, safe header handling, no generic inference retries, no SSE buffering, and cancellation propagation.

Document separately: TLS handshake/connection bounds, upstream response-header wait, active-stream idle gap, whole-request timeout, downstream slow-reader allowance and permitted body/frame sizes. A proxy's recognition of `text/event-stream` must be verified with actual ingress middleware and its version, not inferred from a vendor page.

## 3.5 Ordinary upstream HTTPS versus special private routes

Outbound provider requests are ordinary network HTTPS using documented permitted credentials. There is no concept of obtaining **special provider-network approval** to send HTTPS to a commercial API. Keep separate: basic reachability, technical connector compatibility, and actual account/provider authorization.

Prevent SSRF and open-proxy behavior: allowed destinations and paths are certified connector-owned, DNS/re-resolution and redirects are controlled, TLS is verified, client-provided base URLs/headers cannot select an arbitrary origin. Private local-model endpoints reachable on another machine require an explicitly secured reachability profile; they do **not** justify forcing a VPN onto all OpenCode clients.

## 3.6 Host interference budget

Inventory the actual co-hosts first (possibly Traefik, n8n, Redis, PostgreSQL, MinIO and other services; **do not assume they are all still installed**). Establish a protected memory/CPU/file-descriptor/disk and IO envelope for AgentPool. Bound active streams, active connector calls, pending admission and admin tasks. On overload, reject new admission clearly rather than exhausting memory or displacing unrelated services. Reserve disk for DB/WAL/checkpoints and log rotation; a filesystem-full event must not acknowledge an unsafe budget reservation.

**Hosting proof first, but no production changes from this document.**

---

# 4. Compatibility and inference protocol: one provider does not mean one wire language

## 4.1 The exact boundary

```text
TradeAgent  -- normal worker/session/tool lifecycle --> OpenCode
OpenCode   -- selected model provider request -------> AgentPool
AgentPool  -- one permitted physical inference ------> upstream
```

The OpenCode-to-AgentPool contract is model inference, not OpenCode's agent-host API. Only the **model provider** changes. OpenCode still performs agent loops, tool execution, compaction decisions and work-state management. The provider name does not imply that all models have identical tool, reasoning, streaming or context semantics.

## 4.2 Multi-protocol lanes under a single provider identity

A single `agentpool` entry may advertise several truthful model aliases; where the pinned OpenCode version supports it, individual aliases may bind to different supported runtime packages/protocol lanes. Candidate lanes include Chat Completions and Responses; native provider-specific protocols require separate actual client-and-connector proof. A single owner-facing provider identity does **not** authorize funneling every upstream through an impoverished common JSON format.

The documented OpenCode V2 `providers` / `package` / `settings.baseURL` shape is different from general/V1 `provider` / `npm` / `options.baseURL`. Pin exact installed binary and its effective merged configuration before shipping. As of the research baseline, the V2 docs expose a model-level `package` override and configurable headers/chunk timeouts; documentation is **not** a proof that one real pinned client supports every proposed mixed-protocol combination.

**First lane:** one verified HTTP protocol and one permitted real provider route, chosen by the actual OpenCode fixture. Additional lanes are independent certified product profiles, not a new gateway or selector. Stateful WebSocket paths remain disabled until separate state/affinity/cancellation/refresh proof.

## 4.3 Native-first protocol discipline

Use the existing v5.2 compatibility hierarchy:

- **L0:** same effective protocol; credential/model-ID remap and bounded forwarding with minimal event-aware handling.
- **L1:** narrow explicitly proven event/error/usage normalization.
- **L2:** full protocol translation only when capability-by-capability equivalence, continuation, tool semantics and error behavior are proven.
- **L3:** unsupported, reject rather than silently omit features.

Prefer L0 where real protocols agree. Do not declare L0 merely because a text-only stream worked. The gateway must understand enough framing/termination to distinguish completed versus truncated responses; byte forwarding is not an excuse for failing to validate critical stream semantics.

## 4.4 Truthful alias contracts

Each alias binds at least: exact model family/revision/variant, downstream protocol version, input/output/tool/reasoning/vision/structured-output capabilities, context/output limits, billing/data policy, continuation class, approved physical routes, and release/certification evidence. Multiple registrations may serve one alias **only** after their effective contract is proven equivalent. Provider-name similarity is not proof. No implicit fallback between Claude, GPT, Gemini or arbitrary local models.

Some upstream response IDs, reasoning state, cache prefixes or remote sessions are credential/route-bound. If continuity is not provably portable, use required route affinity or fail with a host-understandable model-switch requirement. AgentPool does not invent or migrate agent memory.

## 4.5 Complete call coverage

For managed OpenCode workloads, the effective configuration and network trace must account for primary models, title/small/auxiliary models, compaction, plugin-originated model calls and supported subagent calls. Either each model-generation path uses AgentPool or the managed host explicitly disables it. An unrelated independent machine is not remotely sandboxed merely because it knows AgentPool's URL.

## 4.6 Client-generated retries must be tested before unattended billing

OpenCode V2 documentation reports header/chunk timeouts and retries after timeout; actual retry count and controls must be observed on the pinned build. Distinguish an AgentPool retry from an OpenCode-created **second HTTP request**. An upstream model may already have processed a request when the client decides it timed out.

For the selected client/profile, test: before-send refusal, upstream body fully received then silence, delayed first headers, sparse reasoning chunks, stream truncated after tool data, timeout/EOF, and proxy interruption. Capture whether the host retries, what error details it receives, and whether its configured retry policy/plugin can prevent ambiguous duplicate spending. If no adequate control exists, **do not certify the affected profile for unattended nonduplicate semantics**. Never compensate by placing agent supervision inside AgentPool.

---

# 5. Pool, logical environments and capacity ownership

## 5.1 Distinct identities

| Object | Meaning | Forbidden assumption |
|---|---|---|
| `EnvironmentSlot` | Logical grouping with one position per integration key where permitted | Runtime, separate allowance or credential provider |
| `CredentialRegistration` | One enrolled credential/principal connection and generation | Automatically eligible model route |
| `PhysicalRoute` | Actual provider/auth/model/endpoint/billing/continuation tuple | Generic interchangeable provider account |
| `CapacityOwner` | Actual quota/rate/billing scope backed by evidence | Equal to slot number, login name or key count |
| `LogicalModelAlias` | OpenCode-facing certified model contract | Arbitrary fallback class |
| `CallerPrincipal` | Authorized AgentPool consumer or host/owner authority | Equivalent to an upstream registered principal |

These are separately indexed objects under one business core. No per-slot quota or process. No hard-coded small `MAX_SLOTS` or `MAX_REGISTRATIONS`; practical memory, rate, database and upstream limits remain real.

## 5.2 Provisional enrollment and late placement

Environment Slots v2 owns the exact allocator, collision, identity and race handling. Authenticate and stage a new registration provisionally; evaluate verified/unknown identity and duplicate rules; only at commit find the earliest compatible slot in stable creation order with a vacant exact integration position, or create the next logical slot. Commit database occupancy and registration transactionally. Abort, callback replay, duplicate identity and interrupted OAuth must not create permanent phantom slots.

**Important:** enrolment-time slot placement never decides inference routing. Inference-time routing never creates a slot, logs into a provider, discovers credentials through an interactive flow, or edits occupancy.

## 5.3 Central selector, indexed and bounded

The selector reads an immutable candidate set keyed by logical alias and downstream protocol, with dynamically verified restrictions layered on: grant/audience, account status, actual capability, model continuity, budget, capacity owner, concurrency, credential/config generation and permitted manual pin. Use a deterministic, cheap known-good/affinity-aware strategy with bounded fairness; do not start with a heavy optimizer or make remote quota calls on the hot path.

The final dispatch fence and a concurrent revoke/disable operation must have clear linearization semantics inherited from Slots v2 §7.1. No global lock survives upstream network I/O. Reservation must be scoped to the relevant financial and capacity owner so unrelated routes do not serialize.

Observations of provider capacity have explicit `reported` / `observed` / `estimated` / `UNKNOWN` evidence with timestamps and expiry. Unknown is neither zero nor infinity. A specific 429 should affect the demonstrated scope only; shared quotas should not be multiplied by multiple credential rows. Avoid herd behavior when several accounts share one actual limit.

## 5.4 Genuine proactive routing

Proactive means avoiding **known** insufficient routes before dispatch through cheap local evidence. Provider-side limits that are unreported or probabilistic cannot be guaranteed away. When capacity is unknown, apply bounded concurrency and conservative policy instead of inventing headroom. Optional JEV advice may later softly adjust already-eligible ranking; never let it override hard constraints or block inference.


---

# 6. Request, dispatch, streams and retry safety

## 6.1 One bounded, observable inference pipeline

For each new downstream model request:

```text
RECEIVED
   | reject unsupported path/method, bound headers/body, authenticate bearer
   v
AUTHORIZED
   | evaluate alias, protocol, required capabilities, grant and policies
   v
CANDIDATES_READ
   | indexed immutable candidate snapshot; dynamic eligibility evaluation
   v
ROUTE_SELECTED
   | reserve scoped concurrency/capacity and required financial exposure
   v
ADMITTED_AND_FENCED
   | recheck relevant credential/policy/route epochs; freeze physical route
   v
PREPARING
   | construct certified upstream request with protected credential
   v
MAY_HAVE_SENT  <------ once transmission may have begun; outcome not assumed
   |
   +--> UPSTREAM_REJECTED (typed, evidence-based semantics)
   |
   +--> RESPONSE_READY --> DOWNSTREAM_COMMITTED --> STREAMING
   |                                             |         |
   |                                             |         +--> COMPLETE_VALID
   |                                             |         +--> PARTIAL_FAILURE
   |                                             |         +--> CLIENT_CANCELLED
   |                                             |
   +--> UNKNOWN_OUTCOME (possibly executed/billed)

   Any safe pre-send failure --> NOT_SENT_FAILED (only when proven)
   Every terminal state --> ledger/health/outcome reconciliation as justified
```

**Two commitment points must not be confused:** Upstream work may have been accepted or billed before the first byte of output reaches OpenCode; and output observed by OpenCode cannot be retracted. A network error after possible send is **not** evidence of free, unsuccessful inference.

All interdependent changes relevant to dispatch must be fenced by a bounded atomic/cas mechanism. No SQL write transaction, global mutex, OS process call, OAuth handshake or browser request remains open through upstream inference. The selector is fast; the stream can be long.

## 6.2 Physical attempt immutability

One physical dispatch attempt freezes the selected upstream route, model identity, credential generation, endpoint/TLS profile, billing scope and continuity class. Its stream cannot be migrated to a second account. Credential rotation does not silently modify an in-flight request. In-flight behavior on account revocation is explicit (`DRAIN` future-only versus `ABORT` current requests) and must pass concurrent race tests.

## 6.3 Four different failure classes

| Class | Evidence threshold | Normal permitted behavior |
|---|---|---|
| **LOCAL_REJECT** | Request fails permission/model/budget checks before upstream send | Reject immediately; no upstream work |
| **PROVEN_NOT_SENT** | Transport and connector can establish that no request was transmitted/accepted | Release/reconcile reservation as authorized; potential new route only in a separately certified optimization |
| **DEFINITE_PROVIDER_REFUSAL** | Exact connector semantics prove explicit refusal and its billing/side-effect implications | Mark actual affected scope, return honest error; bounded same-request redispatch only if separately certified/approved |
| **MAY_HAVE_SENT / PARTIAL** | Request may have been accepted, or any model output was delivered | **Never automatically replay**; report uncertainty, preserve financial hold/charge uncertainty, let OpenCode/TradeAgent own task recovery |

A generic 429, 401, 5xx, EOF or timeout is **not automatically** proof of non-execution. Classify based on the exact provider protocol and transmission evidence, never on convenient assumptions. No HTTP proxy middleware is allowed to transparently retry billable inference requests.

## 6.4 Safe redispatch — explicitly **not a baseline behavior**

The earlier adversarial audit identified an opportunity: after a **provably unsent** or provider-certified non-generating refusal, a later physical attempt on equivalent B may avoid an unnecessary OpenCode failure. That is a useful **proposed optional amendment (C-13)** to the parent's current initial no-retry behavior; it is **not activated by this deployment contract**.

If the owner later authorizes it, all must hold: bounded attempt count, clear non-execution proof for preceding attempt, no exposed downstream bytes, same alias/capabilities/billing policy, independent reservations correctly reconciled, no hidden provider-state continuation, protocol-level negative tests and exact OpenCode retry tracing. A new physical attempt must still have one frozen route. A failed or inconclusive proof leaves the existing parent rule in force.

**Do not confuse safe pre-dispatch reselection with retry.** The selector is already permitted to choose B when A's eligibility fails **before** the upstream dispatch boundary.

## 6.5 Streaming transport

The gateway must preserve exact protocol event order, tool-call IDs and argument assembly, valid UTF-8/SSE framing, permitted reasoning contents or signatures, usage/caching metadata, finish conditions, cancellations and well-defined errors. Parse enough event framing to detect a missing terminal marker or malformed event, without buffering whole generations or repackaging an otherwise compatible stream unnecessarily.

- Use a small bounded per-stream buffer; flush events promptly and support arbitrary TCP chunk boundaries.
- Propagate client cancellation to provider transport when supported; do not claim provider billing ceased.
- Bound header size, body size, decoded payload complexity, SSE frame/event size, streams, pending admissions and slow readers.
- Distinguish first-header timeout, idle gap between active chunks, total request lifetime, downstream write timeout and drain limit.
- After downstream headers/output are committed, never pretend a new HTTP success or status can replace partial failure; fail in the protocol's certified way or terminate observably.
- Prevent diagnostic traces/logs from storing prompt bodies, returned reasoning, bearer tokens, provider secrets or raw tool arguments by default.

## 6.6 Actual OpenCode retry interaction

An application-level timeout may result in a **new incoming AgentPool request** even when AgentPool itself never retries. The connected client may drop its original request identity or use a new one. Therefore:

1. Pin actual OpenCode V2 executable, package, transport and effective merged configuration.
2. Record its native timeout and retry behavior under controlled fixture failures.
3. If the host supports a proven narrow retry veto/plugin/policy, apply it in **OpenCode's host configuration**, not in AgentPool's agent control.
4. If the host cannot be constrained for an unattended profile where duplicate billable inference is unacceptable, mark that profile blocked rather than inventing end-to-end exactly-once guarantees.
5. Clearly separate **deduplicating local admission/reservation records** from **guaranteeing that a remote upstream inference happened at most once**. Only an independently certified upstream idempotency facility can help with the latter, and only to its documented extent.

---

# 7. Authorization: one core, distinct non-escalating authorities

## 7.1 Remove parallel authorization engines, not permission boundaries

All authority checks share **one canonical capability evaluation system**. Credentials have different audiences and acquisition mechanisms, but never separate interpretations of model access, budgets, route policy or revocation. The owner dashboard is a protected adapter; it does not have a private alternate router or budget calculator.

**Illustrative common model; schema names are not mandatory:**

```text
Principal
  principal_id
  kind: owner_session | consumer | delegated_host | workload | local_operator
  credential_verifier_or_session_ref
  audience
  generation / epoch
  expiry / revoked_at

EffectiveScope
  allowed_operations
  allowed_aliases / capabilities
  permitted_data_and_billing_scopes
  maximum_aggregate_exposure
  maximum_concurrency
  not_after
  inherited_delegation_ceiling (if any)
```

Evaluate a permission through the **intersection** of the authenticated principal, owner policy, delegation parent (if any), model/physical-route certification, credential epoch and actual dynamic budget/capacity. No action may broaden this intersection because of metadata supplied by the caller, model output, `Host`, forwarded headers, or a dashboard toggle.

## 7.2 Exactly four operating authority classes

| Class | Authentication | Allowed | Forbidden |
|---|---|---|---|
| **Owner** | Strong remote login with tested MFA/session recovery, or verified restricted local bootstrap | Enroll/disable/rotate, inspect, issue/revoke grants, configure permitted policy and budgets | Export raw secrets by default, override actual provider permission |
| **Consumer inference** | High-entropy AgentPool bearer, scoped to inference | Call certified models and minimal permitted catalog | Enroll, see upstream keys, mint grants, change budgets, invoke admin |
| **Delegated trusted host** | Independently revocable restricted host credential | Issue/renew **bounded child inference grants** when G4 is enabled | Increase own ceiling, perform owner admin, run jobs, see provider secrets |
| **Privileged local operator** | Proven OS principal / protected local IPC, limited recovery commands | Secure bootstrap, maintenance and disaster review | Treat arbitrary local app process as owner, accept prompt-supplied control |

An owner **session** is not an inference bearer. A delegated host credential is not an upstream inference key. Workload inference credentials cannot mint their own children. Independent OpenCode clients require no enrollment into a fleet-management system beyond issuance of a permitted inference credential.

## 7.3 Interactive keys

One remote owner can create independently revocable keys for multiple devices. Keys MUST be securely random, unguessable, audience-bound and checked through an efficient indexed verifier. For high-entropy bearer material, a keyed one-way verifier with an installation-protected pepper is a reasonable candidate; do not force an expensive human-password KDF into the streaming admission path by analogy. Use constant-time comparisons where applicable. Store display name, scope, expiry, generation, last-used information and redacted audit metadata; show plaintext once at issue time.

A generic OpenCode config is not assumed to refresh a rotating key transparently. The exact interactive expiry/rotation UX is a **release configuration decision**; test it before promising unattended renewal. Issued keys must have independently enforceable revocation and financial ceilings. Do not embed upstream API keys in config files or generated client instructions.

## 7.4 Remote unattended TradeAgent grants

This feature belongs to the optional/selected G4 managed-host profile, **not** every interactive client's baseline.

1. Owner issues a separate host-delegation credential with non-increasing alias, billing/data policy, aggregate budget, concurrency and expiry ceilings.
2. Trusted host holds delegation material outside the filesystem/environment visible to OpenCode workers.
3. A narrow HTTPS endpoint issues short-lived inference-only grants with scope bounded by the delegated ceiling and current owner policy.
4. **All children charge against a common parent budget and concurrency envelope.** Issuing 100 grants does not grant 100 independent budget allowances.
5. Inference grants cannot delegate, enroll, refresh upstream credentials, change owner policy or access user/session data.
6. Revocation, rotation and crash recovery are generation-fenced. Grant issuance is rate-limited and authenticated independently of inference.
7. The interface contains no tasks, worker start/stop calls, agent IDs as authority, queue, scheduler or fleet-install requirement.

Where delegated work requires a stricter per-worker budget, charge against **both** worker and parent financial limits. A child cannot outlive its valid parent authorization. Propagation and in-flight abort semantics must be explicit, and the centralized verifier must not keep an unsafe long-lived stale allow-cache.

## 7.5 Owner identity and browser security

Freeze exactly **one maintained owner authentication approach** after inspecting which identity system already exists and whether it is appropriate: verified owner OIDC/SSO with strong MFA, or a tested maintained local password/passkey/MFA implementation if no suitable provider exists. The criterion is security, recoverable login and simplest operation, not ideology. Do **not** write a novel identity platform.

Minimum: phishing-resistant MFA where feasible, safe account recovery, secure HttpOnly cookies, SameSite and CSRF protection for state-changing requests, strict callback/Host/Origin verification, session fixation defense, session expiry/revocation, authentication rate limits, recent authentication for high-risk actions, secure audit, and no upstream or recovery secrets in HTML hydration/localStorage/URLs/analytics/logs. Do not allow always-on public first-user signup. An inference bearer must never authenticate a browser owner session.

A selected OAuth connector follows its **provider-documented** flow: exact redirect validation, one-time state binding, PKCE where supported/required, issuer/mix-up defenses, owner-session association, replay protection and credential staging. RFC 9700 is security guidance, not proof of permission to use consumer subscription inference. Some modes require provider-approved device/native flows instead; do not invent a generic website OAuth login for all.

## 7.6 Trust-boundary limits

The owner authority can select budgets and allowable routes **within actual upstream authorization**; it cannot make an unauthorized subscription connector legal by toggling enabled. One co-hosted daemon has access to several credentials; a full daemon/root compromise is a real residual risk. An isolated connector executor or hardened credential compartment is justified only by a demonstrated connector or threat-model requirement, not per logical slot.

---

# 8. Durable state and credential custody — simplify the transaction boundary

## 8.1 One primary transactional database

**Inherited baseline:** SQLite on a real local persistent filesystem, with WAL, schema migrations, durable accounting and exact v2 slot/registration semantics. No default PostgreSQL dependency simply because the VPS may already run Postgres. Do not place the active SQLite database on arbitrary network storage. Long upstream calls, OAuth redirects and streams must never be wrapped by open SQLite write transactions.

Use short transactions for admission, policy/revocation epochs, enrollment placement, audit pointers and required financial holds. Test actual `busy_timeout`, WAL checkpointing, durability/synchronization, concurrent readers, disk-full behavior and crash/restart semantics. In strict durable accounting, preserve the required pre-dispatch fsync/durability fence rather than benchmark it away. SQLite's WAL and online-backup APIs are relevant standards but **host-specific behavior still requires proof**.

## 8.2 Preferred design candidate: encrypted credential blobs in SQLite

**Proposed amendment C-09, not yet an authorized parent replacement.** In v2, SQLite references a separately managed dynamic vault. That is secure only if vault custody is correctly implemented, but it introduces a second mutable consistency boundary for registration and credential generations.

**Prefer a single transactional encrypted-record design if it passes the target host proof:**

```text
SQLite, one atomic transaction:
  Registration { id, principal, integration, status, generation, slot_id }
  CredentialSecret { registration_id, generation, cipher_version,
                     root_key_id, nonce, ciphertext, authentication_tag }
  CredentialStatus { lifecycle, pending/revoked/unknown }
  Audit/Outbox { redacted transition record }

Protected OUTSIDE database:
  Encryption-root key material and disaster-recovery custody
```

Secrets are individually authenticated-encrypted with a maintained authenticated-encryption construction, unique nonces under the actual key rules, authenticated binding to registration/generation, and secure key-version metadata. A database copy **alone must not decrypt credentials**. The daemon should only decrypt the frozen certified registration it is allowed to dispatch; no plaintext disk fallback. Root keys are provisioned through a proven protected service/host credential mechanism; the owner holds separate protected recovery material. Key rotation must be planned and tested without invalidating live legitimate credentials.

**Why this is simpler:** metadata, encrypted credential generation, status and audit transaction references can change in **one ACID database transaction**, removing many avoidable SQLite↔vault orphan/mismatch cases. An online DB backup obtains an internally consistent encrypted-data snapshot. The same transaction can atomically commit a rotated credential payload and generation.

**What remains hard and must not be hand-waved:** a provider can issue/rotate a refresh token **before** AgentPool durably commits it; the provider is an external transactional party. SQLite atomicity cannot roll back a provider-issued invalidation. Design idempotent, provider-specific refresh state and bounded recovery/reauthorization for this irreducible case. Root/daemon compromise still exposes live decrypted secrets; encryption at rest does not create per-credential process isolation. Database snapshots can restore obsolete upstream credentials and obsolete authority; §9 still applies. Stronger T2 separation remains a future proven-need option.

## 8.3 Deterministic custody selection — no endless design fork

At the **custody proof gate**, compare this single-DB encrypted-record design with the smallest secure **OS-native dynamic secret facility** actually available. Select encrypted SQLite unless the alternative demonstrably satisfies the full dynamic lifecycle and disaster-restoration requirements **with fewer or equivalently bounded failure states**. A systemd boot-time encrypted credential facility alone is not a complete dynamic secret store.

**Release gate CD3-CUST01:** If encrypted SQLite satisfies unattended restart, confidential at-rest storage, atomic credential generation, rotation, recovery key separation, attack-boundary controls, crash cuts and off-host restore, it becomes the sole supported default. Otherwise use a proved simpler compliant local custody implementation. If neither passes, block production. Do not maintain two general-purpose secret stores merely for aesthetic flexibility.

## 8.4 Credential refresh and revocation

- Single-flight refresh **per registration**, never provider-wide or slot-wide.
- One monotonic generation per credential; late callback or refresh cannot overwrite a newer credential.
- No interactive login/refresh in a normal inference hot path; bounded background refresh with jitter, expiration-aware scheduling and provider-specific minimum necessary validation.
- A refresh result and new encrypted secret generation commit atomically **locally**; external provider-side issuance uncertainty remains explicit.
- An unsuccessful refresh affects dependent routes only, with `REAUTH_REQUIRED`/`CUSTODY_UNAVAILABLE` where accurate.
- On revoke: block subsequent admission according to fenced generation and explicit `DRAIN`/`ABORT` policy. The owner sees what happens to admitted streams.
- Never confuse removing a local registration with invalidating a provider credential at its source. Report what AgentPool did and what provider revocation remains to be done.

## 8.5 The financial ledger is not optional infrastructure theater

The existing Slots v2 hard-budget modes remain normative:

| Mode | Contract | Trade-off |
|---|---|---|
| `STRICT_DURABLE` | Durably reserve a conservative authorized maximum exposure before upstream send; reconcile actual usage; keep uncertain holds across crash until proven or owner-resolved | Stronger crash-safe local spend authorization; bounded extra synchronous storage cost |
| `FAST_OBSERVATIONAL` | Best-effort/async observation with request and concurrency limits, without pretending to enforce a crash-proof monetary maximum | Lower durable-write overhead; cannot be sold as a hard spend ceiling |

Strict exposure must account for verified model pricing and enforced output/request limits. If cost/rate, cache/reasoning charges, side charges or billing authority make a conservative maximum impossible, reject a strict-mode dispatch or use an explicitly owner-approved conservative bound. A provider's billed total can still differ from local observation; **this is an authorization bound under stated assumptions, not absolute control of external invoices**.

**Shared financial ownership:** aggregate reservations for one physical billing/capacity scope and delegated-grant parent are authoritative. Do not allow new grants, API keys or logical slots to multiply a shared budget. Reservation validity and final route epoch must be ordered with owner revocation. No false refund after uncertain send.

## 8.6 Snapshot, key and restore subjects are distinct

For a valid off-host recovery plan distinguish:

1. A consistent DB snapshot and associated transaction/outbox/ledger/generation state.
2. Encrypted credential records (in the same DB if C-09 is approved), cryptographic versions and state manifests.
3. Separately protected root-key/recovery material; never place unprotected decryption root in the same archive.
4. Current owner authorization, revocation and financial information **after** the backup timestamp, which may not exist in the archive at all.

A successful SQL integrity check proves only the first kind of internal consistency, not the fourth category's present truth.

---

# 9. Recovery — distinguish ordinary restart, old backup and surviving old host

## 9.1 Three recovery classes, three different protocols

| Event | Appropriate behavior |
|---|---|
| **Ordinary crash/restart on intact host/state** | Reopen and integrity-check current SQLite, recover WAL/outbox, reconcile incomplete attempts without replay, preserve valid current authority subject to true expiry/revocation |
| **Local database rollback or older-backup restoration** | Explicit restricted recovery mode; old authority and budgets must not regain validity; financial and credential review |
| **Whole-host loss/replacement** | Additionally verify safe operational cutover from the former host, reinstall identity, root-key custody, DNS/ingress and external provider credential revalidation |

**Do not label any service startup as a disaster restore merely because the daemon was restarted.** Conversely, do not assume a restored database can conclusively detect its own age: a host and database rolled back together may carry no reliable internal evidence that a newer generation ever existed.

## 9.2 Restore mode and fresh authority epoch

When running the explicitly authorized restore workflow:

1. Build a new installation from a verified release and isolated restore artifacts. Do not expose public billable inference or delegated grant issuance.
2. Establish a fresh installation/recovery identity and epoch via independently protected recovery/bootstrap material. Tokens, host-delegation credentials and owner sessions associated with prior epochs are **not accepted by the new instance**.
3. Require an out-of-band owner recovery/bootstrap process and fresh authenticated administration. A password/MFA verifier restored from old backup is not presumed to represent latest owner intent.
4. Start in `RECOVERY_REVIEW`. Mark restored registrations ineligible until validity, permission and actual credential custody are reviewed. Revoked or unknown credentials cannot silently revive.
5. Restore cryptographic access only using separately held recovery material. If the selected custody mode is intentionally non-portable, require provider reauthorization instead of pretending credentials were restored.
6. Treat spending, pending holds, capacity and revocations after the backup as **potentially absent**. Reconcile with trustworthy provider records wherever possible; otherwise require explicit owner-approved conservative reopening/limits.
7. Check the **single-active-host cutover** criteria in §9.3 before new billable admission.
8. Only then record explicit owner release from `RECOVERY_REVIEW`, mint new bounded AgentPool credentials and allow independently eligible routes. Keep auditable, secret-safe evidence.

This avoids restoring old AgentPool-issued authority as though no intervening actions occurred **on the replacement instance**.

## 9.3 The old-host fencing problem — no fake guarantee

A fresh generation in a **new** database does **not** by itself disable an old VPS that is still running with its own old database and keys. If the original machine was merely unreachable from the operator and later returns, two systems could spend from the same upstream registrations.

**Default simple single-host recovery contract:** before enabling the replacement, establish reliable evidence that the original no longer serves inference and cannot regain the production ingress or upstream spending authority without deliberate operator action. Acceptable mechanisms are determined by real provider/host control: confirmed shutdown/reimage, network and ingress isolation, instance retirement, provider-side key revocation/rotation where possible, or other specifically tested operational fencing. DNS change alone is not an absolute fence: old connections, direct reachability or private paths may still exist.

If the former host cannot be controlled or reliably isolated, **exclusive authority cannot be proved with only two independent local SQLite databases**. Either block reopening shared billable upstream credentials pending isolation/rotation, or approve a genuinely failure-independent fencing system based on an evidenced requirement. Do not introduce consensus infrastructure secretly. The residual condition must be visible to the owner.

## 9.4 Backups and RPO/RTO

Encrypted, consistent **off-host** backups are mandatory, with one release manifest, backup hashes and an independent root-key recovery procedure. A raw copy of a live WAL-mode `.db` file is not accepted as a consistent backup method; use the SQLite online backup mechanism or proven quiesced snapshot and test it. Restore rehearsal must include active writes, interrupted refresh, pre-backup revocation, later spending, an old host still reachable, and a missing decryption root.

The previous proposal of **at least daily encrypted backups, 14 daily + 4 weekly retention, 24-hour RPO and prepared-host 4-hour RTO** is **NOT approved or demonstrated**. These may be used as candidate planning targets only after actual billable exposure, off-host restore, physical host access and operational realities have been evaluated. If 24 hours of missing budget activity is unacceptable, increasing backup frequency alone does not solve post-backup authority truth; use more conservative expenditure and a safe recovery review policy.

## 9.5 Upgrade/rollback is not disaster restoration

A known-good binary does not necessarily read a newer schema. Require pre-migration backup, reversible or expand/contract migrations when possible, exact binary+database version compatibility, no accidental old-grant acceptance after rollback, and tested failed-migration recovery. Do not claim zero downtime for a single-daemon stream service. Drain new admissions, allow bounded current streams to finish, cancel/record uncertain outstanding work, then restart. A destructive rollback must not restore previously revoked authority or artificially free already-spent budget.

---

# 10. Dashboard — control the pool without becoming another backend

## 10.1 Required owner-facing surface

| Screen | Owner can do | Must never imply |
|---|---|---|
| **Overview** | Observe slots, registered accounts, actually eligible routes, active streams, blocked reasons, global service/recovery state | Registered = usable; a slot = runtime/quota |
| **Add API / Add Subscription** | Choose actually certified connector/auth mode, authenticate, inspect status, cancel or resolve identity collisions | Every app subscription permits third-party model inference |
| **Environments** | See A/B/C/N integration occupancy and actual committed placement | Slot preview is a guaranteed reservation before commit |
| **Accounts & Routes** | Inspect redacted principal, permitted models, shared capacity evidence, credential status, route-specific eligibility | Two keys imply independent budgets |
| **Grants & Budgets** | Issue, name, narrow, expire, revoke consumer/host grants and manage actual financial scope | Creating a new token resets spending |
| **Diagnostics** | View typed errors, evidence freshness, stream/route health, last safe activity and uncertainty | HTTP listener alive = successful inference |
| **Maintenance & Recovery** | Drain service, inspect backup/restore state, review uncertain credentials/budgets, rotate allowed secrets | Admin can bypass provider authorization or restore stale rights silently |

Use derived truth from the single authoritative state model. Do **not** create a second UI-side registration machine, slot allocator, capacity ledger or routing engine. Enrollment callback handlers call the same transaction and authorization methods used by other owner controls.

## 10.2 Useful and honest status language

`REGISTERING` (provisional), `REGISTERED` (stored, not necessarily usable), `VALIDATING`, `ELIGIBLE_FOR(profile/request)` (derived), `BLOCKED(reason)`, `REAUTH_REQUIRED`, `CUSTODY_UNAVAILABLE`, `COOLDOWN`, and `RECOVERY_REVIEW` are UI concepts. Underlying canonical statuses remain those of Environment Slots v2; do not redefine enum names in a competing source file simply to fit the dashboard.

A route may be valid for alias X but ineligible for alias Y, or eligible in general but blocked for one consumer's spend scope. Show both the stable health/registration state and request-relative eligibility, not one misleading global green tick.

## 10.3 Admin is not in the streaming critical path

Closing the browser, failing an asset request, timing out an OAuth callback, running diagnostics or creating a slot must not terminate already admitted healthy streams. Bound owner login attempts, probes, callback concurrency, bulk listing, diagnostic aggregation and DB usage independently from inference. The shared Go daemon can still crash as one process; do not claim physical process fault isolation that does not exist.

---

# 11. Operational reliability and resource engineering

## 11.1 State of service versus state of models

Separate at least:

- `LIVE`: process responds to a minimal liveness check.
- `READY_CORE`: necessary storage, root-key custody and admission core are operable.
- `SERVING`: public inference endpoint can evaluate authorized requests.
- `ROUTE_AVAILABLE(alias, principal/context)`: at least one route passes **request-specific** eligibility.
- `MAINTENANCE/DRAINING`: no new unsafe admission, current streams handled under a bounded shutdown contract.
- `RECOVERY_REVIEW`: public billable inference and remote delegation fenced.

The process may be `LIVE` with zero model routes available. Do not call that a successful end-to-end inference health check. An external independent watchdog must be able to alert when AgentPool itself is unreachable; otherwise monitoring disappears together with the service.

## 11.2 Bounded global work, independently constrained classes

Implement measurable caps for:

- Active streamed inference and total upstream connections.
- Pending body parsing/admission and pre-dispatch reservations.
- Per-principal and per-real-capacity-owner concurrent calls.
- Admin logins, OAuth exchanges, background refresh and explicit route probes.
- Total memory per stream; total process memory; goroutine/task count; file descriptors.
- Disk/WAL/checkpoint size and safe log rotation.

Protect inference from expensive administrative work with bounded worker pools and cheap prioritization; do not introduce a broker or another daemon merely to label tasks by importance. A full upstream stall must not exhaust the entire Go scheduler or SQLite writer. One failed connector must not mark unrelated registrations offline unless they share a proven dependency.

## 11.3 Failure ownership

| Failure | Required result |
|---|---|
| Provider 429 with known capacity scope | Cool down that capacity scope; route later eligible requests elsewhere; avoid fabricated quota |
| Auth invalid/refresh failure | Invalidate dependent routes and demand permitted reauth; others survive |
| Client cancellation | Cancel upstream where possible; account for possibly still-billed work |
| Broken/ambiguous stream | Observable partial/unknown outcome; never invent terminal success |
| SQLite durable-hold failure | Fail closed for affected strict-mode admission; do not dispatch unreserved spending |
| Root key unavailable | No affected credential decryption or plaintext fallback |
| Public proxy failure | Honest interruption; streams may fail; no transparent external retry |
| Admin outage/slow callback | Inferencing unaffected when core stays healthy |
| Service restart | In-flight streams can terminate; restart state/ledger reconciled, no replay |
| Full host lost | Recovery review, old-host cutover and independent backup/root-key procedure |
| Provider policy revoked | Affected connector/route disabled for later dispatch; legality cannot be bypassed by changing slots |

## 11.4 Lifecycle

```text
SERVING
   -> QUIESCING    stop new unsafe inference/control admission
   -> DRAINING     complete streams within selected finite bound
   -> STOPPING     cancel remainder, mark partial/uncertain attempts honestly
   -> STOPPED
   -> STARTING     validate build/schema, root custody, host identity, ledger
   -> READY       only then reopen permitted services and eligible routes

A restore from older state additionally requires RECOVERY_REVIEW, not READY.
```

## 11.5 Maintenance and repairability

Deliver one install manifest, one configured source of deployment truth, one service supervisor, one managed backup plan, one upgrade/rollback procedure, one operator runbook and one evidence ledger. Every change must be reproducible and inspectable. Prefer standard OS and application mechanisms over AgentPool-specific orchestration frameworks. Maintain dependency pins and patch security-sensitive Go, SQLite library, proxy and authentication components with tested rollouts.

Optional T1 connector isolation or T2 hardened custody is justified only by a reproduced incompatible runtime, unavoidable crash/process defect, or clearly stated secret-compartment threat model with measurable net benefit. It is **never keyed to the logical slot label**.

---

# 12. Performance: measure what OpenCode actually experiences

## 12.1 Inherited engineering baselines

The parent Slots v2 §17 prospective local objectives remain valid **targets pending real host/profile proof**:

| Metric | Prospective goal | Measurement caveat |
|---|---|---|
| Warm local admission / fast observational accounting | p95 ≤ 5 ms at 20 concurrent clients | Gateway work, matched fixture; not cloud model latency |
| Incremental first-event forwarding, fast mode | p95 ≤ 15 ms at 20 concurrent | Direct fixture vs gateway with matching transport/reuse |
| Incremental first-event forwarding, strict durable mode | p95 ≤ 30 ms at 20 concurrent | Durability fence cannot be removed merely to pass |
| 1/20/100 concurrent fixture streams | Bounded correctness and memory | Does **not** claim a provider permits 100 calls |
| 1 → 1,000 logical dormant slots | No per-slot workers/sockets and near-flat per-slot runtime overhead | Also test real registration/route counts |
| Security and stream correctness | No known acceptance violations | Hard gate, not statistical percentile |

Percentile subtraction on unrelated samples is not gateway overhead. Benchmark paired identical request sequences with fixed test fixtures, protocol, connection reuse and warm/cold scenarios. Report local private, proxied public HTTPS and actual remote user experience separately.

## 12.2 Outcome metrics, not just microbenchmarks

Mandatory reporting:

1. Selector/admission latency p50/p95/p99 and contention of hard-budget durable commits.
2. Added proxy/AgentPool time to **first valid model event** against a matched direct baseline.
3. End-to-end time-to-first-output and total completion on real supported client/route profiles, with variability acknowledged.
4. Successful completed request proportion under matching workload conditions.
5. Invalid/failed/wasted upstream dispatches (including 429, auth errors and ambiguous sends) per admitted request.
6. CPU time, RSS/heap, allocations, goroutines, file descriptors, network connections and buffered bytes per active stream; idle baseline with many slots/registrations.
7. VPS CPU-credit depletion behavior, steady-state resource use and the impact on co-hosted unrelated services.
8. Operational restore and credential-rotation costs, not just steady-state request speed.

Selection is successful only if it improves **effective availability and/or capacity use** without materially degrading output correctness, billable safety or perceived latency. AgentPool cannot make a remote model's own token generation inherently faster through the act of proxying it.

## 12.3 A simple, stable selector is the fast selector

Do not start with distributed ranking, online model scoring, request-path provider probing or heavy scheduling. Precompute alias candidate sets; keep observations local with bounded expiry; prefer healthy sticky/fair routes; protect actual shared capacity owners; avoid synchronized reroute oscillation; use deterministic ties for reproducibility. Measure any proposed sophistication against this baseline before accepting its idle work and failure cases.

Optional JEV is outside the request path: advisory inputs are redacted health/capacity summaries; outputs are soft expiring penalties **only for otherwise eligible candidates**; no permission, financial, credential or deterministic math may depend on JEV. With JEV unavailable, behavior is identical to the tested deterministic baseline.

## 12.4 Numeric release thresholds must be set before measurement

The exact acceptable production public proxy overhead, co-host CPU/RAM/IO interference, max concurrent streams, log/disk reserve, backlog rejection threshold and restore objectives depend on the actual KVM 4 workload. They are **unknown now**. The owner and engineering team must predeclare numeric acceptance thresholds after read-only CD3-H01/CD3-H02 inventory and **before** performance trials. A failing result triggers repair or a transparent reduced supported concurrency profile, never silent correctness weakening.

---

# 13. Build, installation, releases and change authority

## 13.1 One supported release artifact

The production profile must be reconstructible from:

- Exact repo commit, signed/checksummed binary or image, dependency/toolchain versions, SBOM/supply-chain evidence appropriate to scope.
- Target OS/kernel/architecture, runtime and ingress config revision; preflight and verified private listener.
- Dedicated service identity, persistent paths, permissions, root-key delivery, restart policy and basic health.
- Actual OpenCode provider configuration template built from a **certified** catalog, not invented model names.
- Owner bootstrap/recovery method, inference key management, and optionally delegated G4 lifecycle.
- Bounded metrics/logging without prompt/token leakage, external outage observation and off-host backups.
- Tested installation, upgrade, drain, rollback, uninstall/revoke and full-host recovery runbooks.

Do not build two permanent install systems to support one approved VPS. Avoid unpinned `latest` runtime images and surprise breaking OpenCode migration during production change windows.

## 13.2 Preflight before touching the live host

First perform **read-only** inventory of services, ingress routes, DNS/TLS, network namespaces, open listeners, installed backup services, OS privilege, storage mount and filesystem, CPU credits, memory, disk and current load. Record a sanitized baseline and practical revert plan. Production changes require separately explicit authorization for DNS, firewall, ingress, secrets, provider billing and service mutations.

A successful architecture design does **not** authorize SSH installation or real billable tests. Staging/private fixture work can precede complete G7 hardening; public internet exposure must not.

## 13.3 Safe upgrade and rollback

1. Validate release checksum/signature, current state and schema compatibility.
2. Ensure verified consistent backup and distinct recovery material are available.
3. Quiesce and drain bounded streams, recording uncertain outcomes honestly.
4. Apply tested idempotent migrations; prevent rollback into an incompatible schema or restored old authority.
5. Start one instance, verify OS confinement, readiness, owner/admin boundary, real OpenCode smoke test and route legitimacy.
6. Roll back with matching compatible state if needed; invoke recovery-review rules when state is actually restored from an old backup.
7. Observe unrelated VPS services and return to normal only when acceptance conditions pass.

No dual active AgentPool instances to 'achieve rolling zero downtime' on the same database without a separate specifically justified concurrency and fencing design. One daemon means maintenance may interrupt streams.

---

# 14. Falsifiable proof catalogue — one ledger, no ceremonial checklists

These proofs **attach to** existing P0–P7 / G0–G9 and v2 U01–U20 / EN / CAP / RT suites. They are not an independently competing milestone program. Tests that require a real upstream connector must separately record live provider evidence; fake upstream tests cannot certify authorization or a vendor's actual semantics. Tests involving money, production keys, DNS or host mutation require separate approval.

**Namespace rule:** This contract uses `CD3-...` test IDs, distinct from inherited Slots v2 `EN-xx`, `RT-xx`, `SEC-xx` and `PERF-xx` IDs. A `CD3-...` proof **links to** existing parent evidence where applicable and must not be implemented as a duplicate test merely because its label differs.

**Pass rule:** Every experiment has a reproducible procedure, a predeclared expected outcome (including numeric thresholds where needed), exact pinned input/build/host, measured observations, sanitized artifact hashes, a reviewer and a PASS/FAIL/INCONCLUSIVE/NOT_RUN/NOT_APPLICABLE verdict. `NOT_APPLICABLE` requires evidence that the capability is excluded from the selected release, not merely inconvenient to test.

## 14.1 Host, network and deployment tests

| ID | Proof | Failure case that MUST be exercised | Gate |
|---|---|---|---|
| CD3-H01 | Actual KVM 4 OS, ingress, containers/services, disk, topology and privilege inventory | Missing/unknown system surface explicitly reported; no inferred host facts | G0 |
| CD3-H02 | Co-host CPU/RAM/IO/credit baseline and admissible headroom | Sustained test under depleted/limited CPU credits; existing service latency measured | G0/G7 |
| CD3-H03 | **One** chosen native/container profile reaches only private ingress | Direct IPv4/IPv6 external backend access and cross-container/user access rejected | G2/G7 |
| CD3-H04 | Dedicated process/service OS confinement | Non-owner cannot read DB/key, privileged socket, secrets or unrelated service mounts; root residual documented | G7 |
| CD3-N01 | Independent MacBook/desktop-style clients over normal HTTPS | Separate networks, TLS/SNI failure, DNS change; no VPN/sidecar | G4/G7 |
| CD3-N02 | Owner/admin/inference separation at ingress and application | Host-header spoofing, X-Forwarded spoofing, cross-origin browser, inference-token owner escalation | G4/G7 |
| CD3-N03 | Streaming through actual installed proxy | Chunk splitting, immediate flush, large events, slow downstream, long silent reasoning intervals, disconnect and cancellation | G4/G7 |
| CD3-N04 | TLS/proxy/runtime failure behavior | Certificate renewal/replacement, proxy restart, IPv4/IPv6 mismatch and external outage with no false replay | G7 |
| CD3-OPS01 | Reproducible install from approved artifacts | Fresh isolated VM/staging host and service restart with correct private paths | G7 |
| CD3-OPS02 | Maintenance and schema-aware rollback | Bad binary, failed migration, interrupted drain, restored older DB, no obsolete authority | G7 |
| CD3-OPS03 | Existing VPS applications stay intact | Diff host services/routes/volumes/ports/resources before and after targeted staging/deployment | G7 |

## 14.2 Real OpenCode, protocol and routing tests

| ID | Proof | Failure case that MUST be exercised | Gate |
|---|---|---|---|
| CD3-OC01 | Version-pinned actual OpenCode TUI/headless config | V1/V2 config/package confusion, merged config bypass, wrong token audience | G0/G4 |
| CD3-OC02 | Same-session direct-provider versus AgentPool fidelity | Tools, parallel tools, roles, reasoning, usage, multimodal, context and finish semantics across certified features | G1/G4 |
| CD3-OC03 | Stream boundaries and truthful endings | SSE across arbitrary TCP chunks, malformed frames, missing terminal marker, early disconnect, partial tool args | G1/G4 |
| CD3-OC04 | OpenCode native timeout/retry profile | Provider received body then timed out; OpenCode retries; prevent unsupported unattended duplicate guarantee | G0/G4 |
| CD3-OC05 | Optional model-level different protocol lane | Chat/Responses mixed model configs, auxiliary request routing, unsupported `package` override; reject unproven combination | G4/G6 |
| CD3-RT01 | One automatic route selector | Registration changes, manual pin/unpin, epoch races and real concurrent requests; no second decision source | G2/G5 |
| CD3-RT02 | A→B after A becomes ineligible for **later** request | Same actual model/variant/stream contract and session; no owner action; no stream splice | G5 |
| CD3-RT03 | Capacity ownership not multiplied by credentials | Two accounts sharing quota, 429 scope and headroom observed; unrelated owner remains healthy | G5 |
| CD3-RT04 | Correct policy and model fidelity | Wrong model family, restricted billing/data, mismatched variant, invalid continuation and expired grant never dispatch | G2/G5 |
| CD3-RT05 | Route generation and revoke linearization | Revoke during reservation, before socket send, during active stream; defined DRAIN/ABORT outcome | G2/G5 |
| CD3-RT06 | Optional C-13 redispatch (only if approved) | Prove unsent/refused status; reject ambiguity or partial output; bounded accounting for each physical attempt | G5/G6 |
| CD3-RT07 | Large logical environment and route catalog | 1,000 dormant slots plus realistic registration/candidate stress; no per-slot process/worker/sockets or unbounded scans | G7 |

## 14.3 Security, financial, credential and recovery tests

| ID | Proof | Failure case that MUST be exercised | Gate |
|---|---|---|---|
| CD3-SEC01 | One common capability evaluator | Wrong audience, stale grant, revoked key, higher claimed role, wrong alias/data scope and unauthorized operation | G2/G7 |
| CD3-SEC02 | Remote owner browser defense | MFA bypass, CSRF, session fixation, Host/Origin/callback misuse, risky owner-action freshness | G3/G7 |
| CD3-SEC03 | One delegation ceiling | 100 child grants cannot multiply parent budget or concurrency; child cannot mint new grant or administer | G4/G7 |
| CD3-SEC04 | Low-cost inference key verification and rotation | Exhaustive rate/size abuse, timing-sensitive behavior review, revoked bearer blocked promptly | G4/G7 |
| CD3-SEC05 | Outbound endpoint control and no secret leakage | SSRF, poisoned redirects/DNS, request header smuggling, non-approved hosts, logs/error bodies/token exposure | G2/G7 |
| CD3-EN01 | Atomic earliest-compatible slot placement | Concurrent same-integration adds, duplicate identity, OAuth replay, cancellations, staged failure | G3 |
| CD3-EN02 | Registration vs route truth | Saved but invalid credential, expired consent, shared quota, missing connector policy, blocked model | G3/G7 |
| CD3-CUST01 | Choice of unified encrypted SQLite custody | Crash at every local write boundary, reboot with key unavailable, key rotation, off-host decryption, compromised DB without key | G2/G7 |
| CD3-CUST02 | External refresh-token boundary | Provider issues new token then daemon crashes before commit; no fabricated safe old-token reuse | G2/G7 |
| CD3-FIN01 | Strict durable exposure and admission | 100 contenders, shared billing owner, disk full, WAL stall, abrupt crash and post-send uncertainty | G2/G5/G7 |
| CD3-FIN02 | Fast observational mode truth | Crash can lose observations; UI/grant never promises a hard cap | G2/G7 |
| CD3-REC01 | Consistent encrypted off-host snapshot | Live writes/refresh/ledger; recover correct DB+secret generations with separate root material | G7 |
| CD3-REC02 | Stale-authority restoration | Backup before revoke/spend and restore after; old consumer/delegation/owner sessions unusable on new instance | G7 |
| CD3-REC03 | Old-host split-brain cutover | Original still reachable after replacement; block shared billable inference until operationally fenced | G7 |
| CD3-REC04 | Failed restore and RTO rehearsal | Missing root, missing backups, stale external entitlement, nonportable token; truthful reauth/blocking | G7 |
| CD3-REL01 | Long streams and shutdown | 1/20/100 synthetic streams; graceful and forced crash; bounded buffers, no unsafe replay | G7 |
| CD3-REL02 | Separate inference/admin work budgets | Many callbacks/logins/probes concurrent with streaming; bounded interference | G3/G7 |
| CD3-PERF01 | Local fast/strict parent metrics | Predeclared percentiles and correct durable holds, no cherry-picked samples | G7 |
| CD3-PERF02 | Public paired benchmark and outcomes | Direct/private/proxied/remote conditions, p95/p99, success and rejected/wasted attempts | G7 |
| CD3-PERF03 | VPS sustained coexistence | CPU-credit exhaustion, IO contention, disk low, slow readers; measured unaffected-service envelope | G7 |
| CD3-OPS04 | Replacement engineer can restore | Follow written runbook with provided approved credentials and no original-author knowledge | G7 |
| CD3-POL01 | Every shipping connector mode is lawful and technically compatible | Terms revision/unsupported subscription class causes ineligibility, not workaround | G6/G7 |

**Test reduction principle:** If an inherited test already proves the exact thing, reuse its experiment ID and evidence instead of implementing duplicate test infrastructure. A test can appear under several requirements without being rerun redundantly, provided the build/host/connector evidence matches.

## 14.4 Canonical experiment record

```yaml
experiment_id: CD3-OC04
status: NOT_RUN                    # PASS | FAIL | INCONCLUSIVE | NOT_RUN | NOT_APPLICABLE
requirement_refs: ["central-v3 §4.6", "parent Slots-v2 U02"]
gate_refs: ["G0", "G4"]
release_profile_id: "<declared exact profile>"
repo_commit: "<full commit>"
parent_blob_shas: ["<full blob identities>"]
client: "<actual OpenCode executable/version/package/config hash>"
host: "<OS/kernel/ingress/filesystem/network/load>"
connector: "<fixture or actual provider/auth-mode/model/policy revision>"
predeclared_method: "<reproducible steps incl. failure injections>"
expected_pass: "<objective outcomes/numeric bounds set BEFORE execution>"
observed: "<results, distributions, error cases, uncertainty>"
evidence: ["<sanitized artifact identifier + checksum>"]
known_limitations: ["<why PASS is profile-specific>"]
reviewer: "<independent authorized reviewer>"
checked_at_utc: "<actual timestamp>"
```

No raw secrets, unredacted prompt bodies, auth headers, customer/private data or recovery keys in proof records. Redact strongly enough to prevent token reconstruction; retain hashes and controlled raw evidence only in separately authorized secret-handling paths when indispensable.

---

# 15. Release profiles and truthful delivery claims

## 15.1 One row for each promised exact profile

| Dimension | Release profile MUST specify |
|---|---|
| OpenCode | Exact V2 executable, environment/runtime package, TUI/headless mode, merged config and version |
| Public model interface | Inference origin, HTTP path/format, stream/error/cancel/continuation support, advertised model alias |
| Caller | Interactive consumer grant or managed delegated-workload grant; expiry/audience/aliases/data/billing/concurrency ceilings |
| Connector | Actual provider, permitted API/OAuth/subscription/enterprise/local mode, documented allowed auth and inference endpoints |
| Route identity | Real provider/model/revision/variant, auth principal, billing owner, capacity scope, protocol/transport |
| Host | Actual VPS OS/ingress packaging, service identity, credential custody, measured maximum concurrency and co-host envelope |
| Assurance | Per-feature pass evidence, terms-check timestamp, recovery class, backup method, audit/build IDs |
| Status | `REQUIRED`, `OPTIONAL_DISABLED`, `DEFERRED` or `BLOCKED` with explicit owner acceptance of exclusions |

No `some Claude/Gemini/GPT subscription probably works` entries. A model name appearing in an example is **not** an entitlement. Actual subscription modes must be certified individually with current explicit provider authorization.

## 15.2 Three precise delivery states

- **DESIGN_FROZEN:** This contract and its parent amendments have been owner-approved, actual unresolved physical facts have exact proof owners/gates, and the chosen supported release profile can be built without a hidden product decision. This **does not** mean it runs.
- **CORE_GATEWAY_COMPLETE:** One central deployed service with full required remote dashboard, protected transactional state/credentials, A/B/C/N enrollment, correct budgets/grants, genuine OpenCode clients over HTTPS, safe automatic equivalent-route next-request selection, proven recovery/operation and all release-blocking security/protocol tests for the named initial connectors.
- **FULL_INTENDED_TARGET_COMPLETE:** Core complete **plus all separately owner-selected must-have provider/auth-mode connector targets**, including actual approved subscription connections where the owner requires them, certified and functioning. A disabled `Add Subscription` placeholder **does not satisfy** a promised working subscription target.

A blocked required connector stays `BLOCKED`; the engineering team must request an explicit owner decision to defer or change scope. It may not mark the project complete by silently downgrading the promise.

---

# 16. Planning board — finish decisions before autonomous construction

This is the **only central-deployment planning queue**. It attaches to P0–P7 / G0–G9; it is not a replacement roadmap or an authorization to build immediately. Each track should produce a short, approved, linked artifact instead of another competing 100-page master plan.

| Order | Planning track | Concrete approved deliverable | Closure proof |
|---|---|---|---|
| **01** | Vision and parent-conflict reconciliation | Approved this v3 and §17 amendment map; parent cross-references | No implicit local-first collision, provider-role drift or hidden new router |
| **02** | Read-only reference VPS inventory | Sanitized actual host/network/ingress/runtime/OS/storage/credit/co-host manifest | CD3-H01/CD3-H02; unknown host facts stated explicitly |
| **03** | Single deployment/ingress profile | ONE native or ONE container selection, explicit private reachability, identity, mounts, listeners, TLS domains | CD3-H03/CD3-H04/CD3-N02 selection; no backend public exposure |
| **04** | OpenCode compatibility baseline | Pinned V2 TUI/headless, package/protocol, all enabled model calls, timeout/retry traces and config | CD3-OC01–CD3-OC04; no invented model config or hidden bypass |
| **05** | Owner/admin + consumer/host authorization | One selected MFA/login approach, common capabilities, interactive-key lifecycle, whether managed G4 is required and its ceiling | CD3-SEC01–CD3-SEC04; no privilege/financial amplification |
| **06** | Transactional state and credential custody | Chosen CD3-CUST01 storage/key design, strict/fast budget semantics, refresh/crash recovery | CD3-CUST01/02, CD3-FIN01/02; one primary durable authority |
| **07** | Actual connector target matrix | Initial real API connector, two equivalent A→B routes, required subscription targets, current terms evidence | CD3-POL01/CD3-OC02/CD3-RT02; no universal connector claims |
| **08** | Remote dashboard and A/B/C/N owner journeys | Screens/actions, authorizations, callback/placement/race contracts using Slots v2 | CD3-EN01/02; no duplicate business engine |
| **09** | Recovery, host cutover and financial reopening | Backup/custody/release manifest, recovery epoch, old-host fencing runbook, measured RPO/RTO targets | CD3-REC01–CD3-REC04; revocations and budgets cannot silently revive |
| **10** | Resource/reliability acceptance | Numeric approved caps for concurrency, latency, idle consumption, per-stream memory, co-host interference, outage/drain | PERF/REL/CD3-H02 goals frozen before runs |
| **11** | Autonomous builder packets and reviewer gates | One P/G-mapped backlog, acceptance fixtures, immutable requirement IDs, test/evidence ledger, change-control authority | Builder can act within contract without deciding a different product |

**Parallel investigation is allowed.** A fake-upstream streaming proof does not have to wait for a finished dashboard; the complete public product still requires all applicable security and recovery gates. Do not publish an intentionally incomplete local G1 prototype on a production public origin.

## 16.1 Builder autonomy versus owner decisions

Builders **may** choose implementation packages, internal Go interfaces, SQL indexes, test fixtures, component styling, reasonably maintained dependencies, benchmark tooling, narrow refactors and exact code layout where acceptance and parent obligations remain unchanged.

Builders **must request owner approval** before altering product authority, changing hosting/provider terms, introducing an extra runtime or database, making a mandatory client sidecar/VPN, changing model substitution semantics, weakening failure/ledger/secret rules, adding new public endpoints, making billable tests, relaxing acceptance, removing a selected subscription goal, rewriting TradeAgent/OpenCode control boundaries or modifying production infrastructure.

A missing unknown is not an engineering taste decision; apply its deterministic proof and fallback. A blocked condition remains blocked rather than guessed away.

## 16.2 The construction sequence must remain dependency-correct

Align rather than duplicate the inherited roadmap:

| Existing milestone/gate | Minimum central deployment emphasis |
|---|---|
| **P0 / G0** | Actual OpenCode and host evidence; first permitted API route; synthetic failure fixtures; pin baseline |
| **P1 / G1** | Thin private authenticated inference path, real TUI/headless stream/tool/failure fidelity against fixture and one permitted provider where authorized |
| **P2 / G2** | One shared permission core, SQLite/credential/accounting, frozen route and safe concurrent admission |
| **Slots G3 / P2 extension** | Full remote owner admin, provisional/late slot placement, actual connector enrollment and negative tests |
| **P3 / G4** | Real independent remote HTTPS clients and selected TradeAgent delegated unattended grant path; no-bypass managed-host evidence |
| **P4 / G5** | Two independently registered equivalent eligible routes, proactive capacity and automatic A→B on later requests, ledger and concurrency correctness |
| **P5 / G6** | Additional provider/protocol/approved subscription modes named by required release matrix |
| **P6 / G7** | Whole KVM 4 security/isolation/backup/restore/co-host/soak/upgrade/rollback acceptance |
| **Conditional G8** | Only genuinely needed T1/T2 connector or custody isolation after measured proof |
| **Optional P7 / G9** | JEV optional off-path calibration, never prerequisite for non-JEV product |

Do not change the project's P numbers merely because central deployment rearranges the implementation dependency order. Record attached gate aliases if the canonical roadmap needs an explicitly approved revision.

---

# 17. Exact proposed amendments to the parent contracts

**Nothing in this section has authority to silently supersede the parents.** On owner approval, reconcile each approved clause in the owning document or append an explicit acceptance-linked parent note. Preserve original semantic constraints unless the exact row below authorizes a change.

## 17.1 Centralization amendments retained and sharpened from v2

| ID | Narrow parent delta | Protected behavior |
|---|---|---|
| **C-01** | Replace local-only inference origin with permitted authenticated **public HTTPS through trusted ingress**, private backend | No client-side AgentPool daemon/mandatory VPN; audience grants required |
| **C-02** | Replace local-only owner dashboard assumption with separately authenticated **remote owner UI** | One central registration and slot authority, no second backend/router |
| **C-03** | Keep protected local operator IPC and add narrow session-authorized remote owner adapter | Inference bearer never receives privileged methods |
| **C-04** | Permit named remote interactive inference credentials under explicit expiry/revocation/rotation policy | Short-lived workload grants remain for selected managed TradeAgent profile |
| **C-05** | For selected G4 unattended host, permit restricted HTTPS child-grant issuance from a non-escalating delegated principal | Not a worker/job/fleet API; parent budgets aggregate |
| **C-06** | Admin browser origin and approved connector callbacks are allowed, while generic browser-origin inference stays rejected unless separately supported | No wildcard CORS or callback owner bypass |
| **C-07** | Interpret one authorized **owner** as permitted to operate multiple independent OpenCode devices | No multi-person account brokerage or subscription-right circumvention |
| **C-08** | Add explicit older-backup `RECOVERY_REVIEW`, stale bearer/session invalidation and financial revalidation | Ordinary restarts differ from restore; no phantom spend reset |

## 17.2 v3 audit-driven scoped amendments

| ID | Narrow proposal | Parent collision / proof needed | Activation |
|---|---|---|---|
| **C-09** | Prefer transactional SQLite encrypted secret records with separately protected root key, instead of a separately mutable dynamic vault | Slots v2 §§5, 9 + v5.2 credential custody; must pass CD3-CUST01/CD3-CUST02 and key recovery | **DESIGN CANDIDATE, NOT PARENT-APPROVED** |
| **C-10** | Define explicit surviving-old-host cutover and acknowledge that a restored DB alone cannot fence the old daemon | Central v2 A-08 recovery guarantee was too strong; CD3-REC02/CD3-REC03 | **REQUIRED CLARIFICATION FOR CENTRAL RECOVERY** |
| **C-11** | Unify capability evaluation across independent owner, consumer, host delegation and local operator credentials | Preserves distinct audiences; CD3-SEC01/CD3-SEC03; no parent business rule change | **DESIGN CANDIDATE** |
| **C-12** | Support certified per-model protocol lanes under one `agentpool` provider identity where pinned OpenCode permits | Parent supports distinct protocol profiles; CD3-OC05 and feature parity prove specific configurations | **DESIGN CANDIDATE; ADDITIONAL LANES RELEASE-SCOPED** |
| **C-13** | Optional bounded same-downstream-request **safe redispatch** after proven nonsend or provider-certified refusal | **Explicit change** to parent initial conservative no-retry behavior; CD3-RT06/CD3-OC04 required | **OFF BY DEFAULT, SEPARATE OWNER DECISION** |
| **C-14** | Replace native-first packaging heuristic with evidence-based **single easiest safe** host profile | No data-plane semantic change; CD3-H01–CD3-H04 | **DESIGN CANDIDATE** |
| **C-15** | Apply aggregate parent-scoped budgets and concurrency to every delegated child grant | Reinforces existing budget/grant invariants; CD3-SEC03/CD3-FIN01 | **REQUIRED ENFORCEMENT CLARIFICATION** |
| **C-16** | Add direct-vs-gateway end-to-end completion and efficiency metrics to inherited latency test targets | No relaxation of parent §17; CD3-PERF01–03 | **PROPOSED RELEASE CRITERIA** |
| **C-17** | Pinned OpenCode retry/failure trace becomes an early unattended proof | Already required by parent U02 and Slots v2 §8.4; CD3-OC04 | **INHERITED REQUIREMENT, PROMOTED IN BUILD ORDER** |
| **C-18** | Separate bounded inference, admin, OAuth/refresh and diagnostic work within the one process | Reinforces no critical-path blocking; CD3-REL02 | **DESIGN CANDIDATE** |
| **C-19** | Distinguish DB recoverability, credential portability and external revocation/billing truth | Clarifies parent vault/recovery; CD3-REC01–04 | **REQUIRED RECOVERY CLARIFICATION** |
| **C-20** | One P/G-mapped decision register and evidence ledger; do not duplicate entire parent roadmaps | Planning governance only, no router change | **DESIGN CANDIDATE** |

**No amendment allows:** arbitrary model substitutions, unauthorized account modes, cross-owner pooling, universal subscription compatibility claims, hidden stream stitching, unsafe ambiguous replays, quota multiplication, per-slot workers, automatic task recovery, credential leakage or an additional selector.

---

# 18. Decision register — what is fixed and what still requires observation

## 18.1 Stable choices for owner consideration

| ID | Proposed decision | Status | Non-negotiable qualifier |
|---|---|---|---|
| D-01 | One central AgentPool on existing KVM 4 | **DESIGN CANDIDATE** | Measure real host feasibility; no forced migration of other apps |
| D-02 | Public ordinary HTTPS inference and separate remote admin origin | **PRODUCT LOCK / CENTRAL AMENDMENT** | Real authorization, not only DNS separation |
| D-03 | One Go business/selection authority, no worker or agent execution | **INHERITED LOCK** | Small, bounded runtime |
| D-04 | Logical A/B/C/N, atomic first-compatible enrollment | **INHERITED LOCK** | Metadata only |
| D-05 | Indexed deterministic selector with strict capacity/budget and model fidelity | **INHERITED LOCK** | JEV optional, no synchronous model scoring |
| D-06 | One primary SQLite transaction boundary | **INHERITED LOCK** | Custody format C-09 subject to proof |
| D-07 | One selected native or container packaging profile | **CONDITIONAL DEFAULT** | CD3-H01–CD3-H04 decide safely |
| D-08 | Shared canonical capability evaluator across separate principal types | **DESIGN CANDIDATE** | No inference-to-owner promotion |
| D-09 | No generic ambiguous automatic replay | **INHERITED LOCK** | C-13 optional and off until explicitly approved |
| D-10 | One-owner account scope across multiple machines | **PRODUCT LOCK / CENTRAL AMENDMENT** | Provider authorization independently applies |
| D-11 | Disaster review plus explicit exclusive-host cutover | **DESIGN CANDIDATE** | Old-host ownership cannot be inferred from DB revision |
| D-12 | Single instance, no default HA/queue/Redis/service mesh/new proxy | **SIMPLICITY LOCK** | Future evidence-based additions require focused amendment |
| D-13 | Real connector certification required even for owner-added subscriptions | **INHERITED LOCK** | No subscription skeleton counted as working target |

## 18.2 Unresolved **facts**, not permission to redesign

| ID | Unknown | Required fact/proof | Release/decision consequence |
|---|---|---|---|
| X-01 | Actual installed VPS OS, proxy, container/network and existing loads | CD3-H01/CD3-H02 read-only manifest | Cannot freeze listener/runtime/limits without it |
| X-02 | Chosen packaging and service identity | CD3-H03/CD3-H04 comparison | One approved profile only |
| X-03 | Real domain names, DNS, IPv6 and certificate route | Owner selection + CD3-N01/CD3-N02 | Public configuration blocked pending choice |
| X-04 | Actual installed OpenCode V2, package, protocol and retry behavior | CD3-OC01–CD3-OC05; pinned trace | Cannot promise arbitrary model/transport behavior |
| X-05 | Selected first certified API and equivalent second physical route | CD3-POL01, CD3-RT02/CD3-RT03 | Automatic pooling proof requires actual two routes |
| X-06 | Required subscription/account connector and authorization mode(s) | Per-mode owner release matrix + current provider documentation and test | Full target blocked for any required unproven mode |
| X-07 | Actual maintained remote owner auth system | CD3-SEC02 and host inventory | Owner UI auth approach frozen before construction |
| X-08 | Interactive key expiry/rotation and G4 delegation scope | Owner limit selection + CD3-SEC03/04 | Do not promise transparent generic key refresh |
| X-09 | Whether C-09 encrypted SQLite or OS-native custody is the smallest safe design | CD3-CUST01/CD3-CUST02 and restore rehearsal | Freeze one; fail closed if neither works |
| X-10 | Measurable budget envelopes, pricing and shared quota evidence | CD3-FIN01, CD3-RT03 and connector pricing docs | No strict unknown-cost admission or invented quota |
| X-11 | Approved numerical concurrency, overhead, RPO/RTO and VPS interference limits | CD3-H02/PERF/REC trials after predeclared bounds | No universal throughput/SLA guarantee |
| X-12 | Whether a lost original host can be isolated during disaster cutover | CD3-REC03 + host-owner authority | Reopening shared credentials blocked until fence |
| X-13 | Whether optional C-13 safe redispatch should ever ship | Proof CD3-RT06 plus explicit owner amendment | Default remains OFF |

**Critical distinction:** `X-## UNKNOWN` can be a well-specified and entirely trustworthy planning outcome. Claiming the unknown fact is known would be untrustworthy.

## 18.3 Rejected starting points

Do not start with: one gateway per OpenCode machine; per-slot CLI processes/browser profiles; a job queue; an OpenCode bridge controlling agents; automatic arbitrary GPT↔Claude↔Gemini swaps; automatic replay of ambiguous inference; a credential-reselling pool; guessed consumer subscription endpoints; second route selector; generic arbitrary URL proxy; Redis/Kubernetes/service mesh or cross-host leader election; separate dynamic vault **and** credential database without proven need; always-on JEV; guaranteed zero-downtime updates; unrestricted wildcard browser inference; production DNS or secret changes before approval.

A future improvement must state a reproduced limitation, explain rejected simpler remedies, show measured benefit and new failure states, and submit one narrow owner-approved amendment. Simplicity is an engineering discipline, not a prohibition on warranted complexity.

---

# 19. Primary external evidence and its limits

These are **official sources consulted for the 9 October 2026 baseline**. Vendor documentation is evidence of a published feature/standard, **not** proof of this owner's installed version, account entitlements, provider terms compliance or production performance. Links must be rechecked at implementation/release; versions may have changed.

| Source | Supports | Does not prove |
|---|---|---|
| [OpenCode V2 — Providers](https://opencode.ai/v2/docs/providers) | `providers`, `package`, `settings.baseURL`, per-model runtime override, documented timeouts and retry behavior | Selected installed V2 binary semantics, mixed-protocol parity or safe headless retries |
| [OpenCode general/V1 provider docs](https://opencode.ai/docs/providers) | Different `provider` / `npm` / `options.baseURL` configuration | Interchangeability with native V2 config |
| [Go official release history](https://go.dev/doc/devel/release) | Go release/patched version history; as of 9 Oct 2026 latest listed major is Go 1.27 with 1.27.2 on 8 Oct | Choice of compiled Go toolchain on the actual host without testing/security review |
| [SQLite WAL](https://www.sqlite.org/wal.html) | Transaction/log behavior, local filesystem and concurrency constraints | Host filesystem correctness with selected `synchronous` and crash cuts |
| [SQLite Online Backup API](https://www.sqlite.org/backup.html) | Consistent online DB backup primitive | Credential-root recovery or post-backup authority correctness |
| [Hostinger VPS limits](https://www.hostinger.com/support/6976044-parameters-and-limits-of-hosting-plans-in-hostinger/) | Nominal KVM 4 resources and public plan-limit descriptions | Present free resources and CPU credits on this VPS |
| [Hostinger CPU credits](https://www.hostinger.com/support/how-cpu-credits-work-on-hostinger-vps/) | CPU-credit accounting and possible throttling | Exact sustained guaranteed share or live host observations; published articles may differ |
| [Traefik HTTP services (v3.5)](https://doc.traefik.io/traefik/v3.5/reference/routing-configuration/http/load-balancing/service/) | Recognized streaming response flush behavior in documented Traefik revision | Actual installed Traefik revision, middlewares, cancellation behavior and AgentPool event parity |
| [Docker bridge networking](https://docs.docker.com/engine/network/drivers/bridge/) | Same-bridge container communication and user-defined network isolation boundaries | Complete OS compromise isolation for the particular host |
| [systemd service credentials](https://github.com/systemd/systemd/blob/main/docs/CREDENTIALS.md) | Protected credential-delivery mechanisms, depending on OS/version | A ready-made transactional dynamic vault, TPM or portable recovery |
| [RFC 9700 — OAuth 2.0 Security Best Current Practice](https://www.rfc-editor.org/rfc/rfc9700) | Redirect/PKCE/state/issuer protections | A provider's authorization to repurpose a consumer subscription for AgentPool inference |
| [AgentPool provider policy](https://github.com/Nicolas1bhr/agentpool/blob/main/docs/PROVIDER_POLICY.md) | Project prohibition on impersonation, scraping private secrets, prohibited account sharing and bypassing plan limits | Provider-specific current account entitlement without a per-connector review |

When documentation conflicts with runtime observations, capture both. When protocol functionality exists but provider policy does not permit that mode, **the connector is ineligible**. That is not a reason to abandon the broader permitted subscription-connector category or to implement covert access workarounds.

---

# 20. Approval, planning stop condition and final builder instruction

## 20.1 What the owner is being asked to approve

Approval of this proposed v3 means agreeing to the product/infrastructure direction and initiating precise parent amendment reconciliation, **not** signing off real host security or authorizing building/deployment. In particular:

1. AgentPool remains a central inference-only provider for ordinary independent OpenCode clients and selected managed TradeAgent workers.
2. The existing KVM 4 remains the reference host unless CD3-H01/CD3-H02 disproves feasibility.
3. The owner dashboard is remote, inference/admin permissions are separate, no client fleet/overlay dependency is introduced.
4. Logical environments, independent registrations, certified connectors and automatic same-effective-model route selection are preserved exactly.
5. The default runtime remains one small Go service and one primary SQLite transaction boundary, with C-09 custody consolidation pursued as a **proof-gated preferred change**.
6. Ambiguous replay stays forbidden; optional C-13 safe redispatch is **not approved by accepting the rest of this document**.
7. Whole-host recovery includes actual operational old-host cutover; restored SQLite alone is not exclusive ownership proof.
8. Deployment packaging, owner identity mechanism, resource caps, connector release targets and actual host details are determined by explicit named proofs, not guessed by builders.
9. The v5.2 / Slots-v2 / policy / P/G source-of-truth hierarchy remains intact.

## 20.2 When this central-deployment planning track is actually complete

**The track is `DESIGN_FROZEN` only when all apply:**

- Owner approves this architecture and accepted amendments have been reconciled with parent contracts.
- Real host inventory and chosen **one** packaging/ingress layout are recorded, including service permissions and exact public/private routes.
- Pinned OpenCode V2 version, selected inference protocol(s), native client retries and no-bypass claims are tested or mapped to blocked release profiles.
- Owner authentication, inference keys and required-versus-deferred TradeAgent delegated grants are unambiguous.
- Credential custody transaction model, root-key provisioning and disaster key recovery are selected with a falsifiable passing path.
- The first actual permitted API route, second equivalent route and required subscription connector targets are declared with truthful current authorization evidence or explicit blockers.
- Dashboard owner journeys map to the existing Slots v2 enrollment state machine and one shared permission core.
- Strict/fast accounting, shared-capacity identity, old-host cutover, budget restore and recovery process are specified for the selected host.
- Numerical latency, concurrency, co-host impact and recovery objectives are frozen **before** the corresponding measurements.
- Every remaining unknown has an assigned proof/gate, a PASS implementation branch, a FAIL/blocking response and owner sign-off mechanism; no builder must invent core product goals.

Until then, this document is a **reliable planning foundation**, not a claim that all engineering uncertainties have disappeared.

## 20.3 What success means in production

A normal allowed request from OpenCode is served by one automatic, certified upstream model route, with faithful streaming, independent protected credentials, enforced caller and billing permissions and correct capacity-aware selection. Adding accounts does not add per-slot processes; opening another OpenCode installation does not create another AgentPool. When a route is known unavailable, future eligible requests can automatically use another equivalent route. When completion is uncertain, the gateway admits uncertainty instead of repeating work or inventing output. When the VPS is lost, security and money are reconciled before resuming.

TradeAgent's existing responsibility for jobs, tools, sessions, memory, checkpoints and trading remains unchanged.

> **Final engineering mandate:** Build a small, sturdy inference engine. A part is added only when it removes a more serious real failure than the ones it introduces. Every permission is explicit, every state transition understandable, every expensive activity kept away from the stream, every backup honest about what it cannot restore, and every promised feature supported by a test. **TradeAgent owns the work. OpenCode executes the agents. AgentPool supplies the inference.**

---

# Appendix A — Concrete change map: central v2 → v3

| v2 issue | v3 correction | Core vision affected? |
|---|---|---|
| Separate vault/database reconciliation as assumed default | C-09 single ACID encrypted-secret design evaluated against actual OS vault; external token refresh uncertainty still modeled | No; simpler custody candidate |
| Restore generation presented as sufficient to prevent old authority | C-10 adds surviving-host operational fencing and limits of purely local restore detection | No; truthful recovery |
| Several credential classes risk several permission engines | C-11 one capability evaluator; independent audiences and non-escalating grant inheritance | No; same required boundaries |
| One initial protocol could constrain full intended model support | C-12 protocol lanes under one `agentpool` identity, model-level certified runtime/profile | No; preserves model fidelity |
| Known pre-send failures and ambiguous sends conflated in optimization debate | C-13 explicitly optional certified safe redispatch, OFF by default; parent ambiguous no-replay retained | No; future narrow optimization only |
| Native systemd selected primarily from private-loopback convenience | C-14 choose the single easiest safe deployment on the real host | No; operational simplification |
| Parent delegated-budget ceiling risk under mint amplification | C-15 parent principal budget and concurrency aggregations required | No; stronger enforcement |
| Proxy overhead microbenchmarks overemphasized | C-16 paired outcome/latency and co-host metrics, no relaxation of parent targets | No; better measurement |
| OpenCode automatic retries might create duplicate billable attempts | C-17 make pinned client retry gate early, with host-side control only if verified | No; closes system-level risk |
| Heavy owner/admin tasks share same daemon with inference | C-18 bounded classes and scheduling rather than extra processes | No; reduces tail-risk |
| Clean database backup conflated with current auth/billing | C-19 explicit restore classes and independent recovery subjects | No; honest guarantees |
| Long checklists risk competing project governance | C-20 one decision and evidence ledger, inherited P/G roadmap | No; simpler handoff |

**No removal of v2's obligations** to a remote dashboard, logical placements, one central selector, public ordinary HTTPS, certified subscription category, independent multi-login, bounded budgets, protected credentials, correct streams, provider policy, recovery or co-host safety is intended or authorized by this v3.

# Appendix B — Change request and experiment governance template

```yaml
change_id: "CDR-000"
status: PROPOSED
initiator: "owner | implementer | reviewer"
problem: "<one reproduced defect or new owner requirement>"
product_invariants: ["I-xx"]
owned_spec_sections: ["v5.2 §...", "Slots v2 §...", "central v3 §..."]
minimal_proposed_delta: "<one precise difference>"
simpler_options_tested: ["<actual evidence of insufficiency>"]
new_failure_and_resource_costs: "<what complexity is introduced>"
security_billing_availability_effects: "<narrowly scoped>"
predeclared_pass_and_negative_tests: ["<exact gate IDs and criteria>"]
owner_approval: PENDING
parent_reconciliation_complete: false
construction_authorized: false
production_mutation_authorized: false
```

# Appendix C — Compact mechanical review for every implementation PR

A reviewer must be able to answer **YES** to each applicable question, with evidence:

1. Does it change only inference plumbing or owner administration, never OpenCode/TradeAgent worker control?
2. Does the new request still have one authenticated, authorized, model-compatible physical route?
3. Are slots and dormant registrations still just metadata, with no automatic idle processes?
4. Is model/protocol/continuation behavior explicit and certified rather than guessed?
5. Are all upstream secret accesses central, scoped and absent from logs/client configuration?
6. Does one canonical permission evaluator enforce inference/owner/delegation boundaries and aggregate budgets?
7. Can a failed provider or admin operation leave unrelated healthy inference working where actual dependencies permit?
8. Are unknown send/usage outcomes still marked uncertain, with no automatic unsafe replay or phantom refund?
9. Are the SQLite/custody mutations as atomic as their local scope allows, with explicit external-provider uncertainty?
10. Can a crash, revoke, refresh race, disk-full event or old backup never silently enlarge authority?
11. Does the selected host packaging avoid new public backend sockets and unrelated app access?
12. Does streaming work under cancellation, long gaps, slow readers and malformed events without unbounded memory?
13. Is any added service/process/database/queue justified by a reproduced need and simpler alternatives evaluated first?
14. Are output metrics measured against matched direct-provider cases, not invented speedups?
15. Is the promised feature live-proven for the exact connector/client/host profile and permitted by current provider policy?
16. Can a different engineer reproduce installation, rollback, diagnosis and recovery from approved artifacts?
17. Are affected inherited tests and the **single** experiment ledger updated without moving a failure threshold after testing?
18. Were every owner-required scope change and any production/billable mutation approved explicitly?

**If any applicable answer is not provably YES, the PR cannot silently claim the corresponding release gate.**
