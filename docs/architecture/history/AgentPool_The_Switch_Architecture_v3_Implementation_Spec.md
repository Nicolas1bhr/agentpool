# AgentPool — The Switch
## v3.0 · Implementation-ready architecture and acceptance specification

> **One native agent, one continuous logical work session, many independently authorized inference registrations — selected at safe request boundaries.**
>
> **The Switch does not manufacture extra quota, reset an account's limits, or change a provider's limits. It respects each registration's entitlements, then selects another independently entitled registration when the current one cannot serve a new request.**

| Control | Decision |
|---|---|
| Document status | **Approved engineering baseline; feature release is evidence-gated** |
| Architecture | Local, single-user, authenticated inference gateway; *not* an agent orchestrator |
| Date / version | 2026-10-08 · v3.0 |
| First runtime | Pinned Codex CLI and Codex App Server |
| Second runtime | T3 Code via pinned `codex-acp` **if certified**; evaluate native T3 switching independently |
| Deployment | macOS per-user first; Linux and Windows adapters later |
| Implementation | One Go module, one executable, SQLite metadata, OS-backed secrets, localhost HTTP and owner-only IPC |
| First upstream route | Documented Sign in with ChatGPT (SIWC) public Responses API, subject to the actual grant |
| Non-goals | Agent orchestration; rewriting native context; scraping private routes; hidden billing fallback; bypassing per-account limits |
| Source of truth | This specification, its accompanying normative schemas, test fixtures, and signed release admission manifest |
| Technical evidence | Official documentation and public repositories checked 2026-10-08; live integration **not claimed** |
| Supersedes | `AgentPool_The_Switch_Architecture_v2(1).md` |

**Notation.** **MUST**/**MUST NOT** are testable requirements. **SHOULD** permits a documented, reviewable exception. **MAY** means optional. Code and JSON below are *implementation contracts or illustrative pseudocode as labeled*, not claims that existing executables implement them. A feature marked `UNVERIFIED` is not production-supported.

---

# 0. The product, precisely

## 0.1 The intent — paid capacity selection, not defeating limits

AgentPool serves workloads that have legitimately authenticated and independently usable inference registrations. When registration A reaches a provider-enforced usage limit, the Switch **stops directing applicable new requests to A**; if B is separately authorized, actually has access to the requested model, fits the workload's data authority and may serve that request, B can be selected at the next safe request boundary. A's usage limit is **not** reset, modified, obscured or evaded at the level of A. B's usage is attributable to B. This is the intended meaning of **account rotation** in every AgentPool design document and code review.

Do **not** incorrectly describe this technical mechanism as: forging additional tokens, breaking a limiter, tricking an exhausted account into serving more traffic, credential theft, browser-cookie harvesting, creating synthetic identities, cross-tenant credential sharing, or bypassing a provider's authentication system. None of those are the product design.

Equally, **do not confuse a technical description with a blanket contractual conclusion**. Buying separate subscriptions is not, by itself, evidence that every product permits automated aggregation or rate-window-triggered selection. The provider's applicable product and contractual rules, the user's authority over each identity, and any limits intended to apply across registrations still govern. An explicit service-level restriction is not overridden by this definition. A neutral, reviewable provider-operating-mode decision is a release gate; engineers MUST NOT repeatedly label independent, paid registrations "rate-limit circumvention" merely because rotation is mentioned, nor assert the feature is permitted solely because registrations are paid. Document the actual provision/approval and the actual behavior.

**Account rotation requirements:**

1. Every registration authenticates legitimately and independently through documented methods.
2. The provider enforces that registration's existing limits; AgentPool neither modifies them nor spoofs identity to change them.
3. One inferencing request uses exactly one selected principal and one explicit billing path; no merged credentials or quota counters.
4. All selected principals are authorized for the workload's confidentiality boundary, model and service route.
5. No silent API-key, credit, overage, workspace, model or provider fallback.
6. Selection uses the provider's documented capabilities, with explicit release clearance for the *specific automatic cross-registration mode* where needed.
7. Unknown policy/identity is not presumed allowed; unknown remaining quota may produce one bounded, safe attempt.
8. When continuation is unsupported, the product reports a typed stop rather than disguising an agent/process restart as transparency.

**Important accounting distinction.** Subscription plan usage is not necessarily a fixed inventory of prepaid tokens. Provider usage can be shared across applications, constrained by app-level settings, and governed by different time windows or credits. UI labels MUST say *observed capacity*, *account limit*, or *provider usage*, not invent a literal token wallet. For example, OpenAI documents that certain ChatGPT Plus usage windows are shared across participating apps; new SIWC registrations do not automatically mint independent plan capacity within the *same* ChatGPT account. [R02, R07]

**Independent paid entitlement is not the same as independent OAuth registration.** The operator may add many registrations, but the selector MUST distinguish the registration used to authenticate from the underlying subscriber/workspace allowance that may be shared among registrations. The intended A→B capacity expansion generally requires independently entitled *provider principals*, not two OAuth clients registered against the same subscription. Model an opaque, verified-or-unknown `capacity_owner_key` at the provider account/workspace level. Two registrations with the same validated subject/workspace are conservatively treated as sharing an allowance unless the provider expressly exposes a finer billing/entitlement boundary. App-specific limits remain scoped to their individual registrations. Do not promise independent quota merely because two sign-ins succeeded.

## 0.2 Outside appearance; internal reality

```text
      Human / T3 Code / TradeAgent / local application
                            |
                native Codex / App Server
       (thread, tools, compaction, reasoning, sandbox)
                            |
                  independently routable
                 HTTP Responses request
                            |
               +-------------------------+
               |       AGENTPOOL         |
               |       THE SWITCH        |
               | ----------------------- |
               | local client identity   |
               | workload authority      |
               | approved account pool   |
               | selection and barrier   |
               | OAuth/token refresh     |
               | exact streaming         |
               | scoped usage evidence   |
               +------------+------------+
                            |
             +--------------+---------------+
             |              |               |
       Registration A Registration B Registration C
       own identity   own identity    own identity
       own limits     own limits      own limits
             |              |               |
             +--------------+---------------+
                       official API
```

The Switch manages inference **routing**, not agent intent. TradeAgent's managers/JEV/worker recovery, T3's UI, and Codex's native thread/compaction/tool implementation remain external.

## 0.3 Outcomes we promise and do not promise

| Promise | Status / exact meaning |
|---|---|
| Native client stays the engine | Design invariant |
| Secure local gateway and official upstream auth | Buildable; must pass real L1 proof |
| Same active thread can send later requests via B | **L3: conditional, live-test dependent** |
| Automatically use B when A is specifically limited | **L4: conditional on L3, error evidence and operating-mode clearance** |
| Running SSE response switches A→B without interruption | **NOT PROMISED** |
| A server-generated opaque reasoning/compaction item always works under B | **NOT ASSUMED** |
| Every account has a measurable, predictable reset timer | **NOT ASSUMED** |
| Exact-once external tool actions or live task durability | Host/native agent obligation |
| Guaranteed no extra provider charges with unverified billing evidence | **NOT PROMISED** |
| Multi-user/SaaS/shared consumer credentials | Not in v1 |

## 0.4 Five architecture decisions that are settled

- **Transport first.** Keep Codex intact; route authenticated compatible model calls below it.
- **Go binary, not distributed machinery.** One service with internal control and data planes.
- **Official endpoint and grant.** No ChatGPT private backend, token scraping or undocumented consumer app protocol.
- **Strict correctness before convenience.** Explicitly classify failures; no hidden semantic replay, account identity changes or degraded-model fallbacks.
- **Proof before expansion.** Run a thin vertical-slice conformance harness *before* building an elaborate daemon/config UI.

---

# 1. Responsibility boundaries and supported modes

| Owner | Owns | MUST NOT silently delegate to AgentPool |
|---|---|---|
| Native Codex | Context, encrypted reasoning, compaction, tools, sandbox, thread/session storage, agent-side events | Semantic history reconstruction and tool idempotency |
| TradeAgent/T3/host | Task durability, worktree, real-world side effects, restart/pause/resume, host caller policy | A durable job executor inside gateway |
| AgentPool | Registrations, protected secrets, local leases, data-authority ACL, route selection, scoped capacity, HTTP/SSE forwarding and evidence | Agent intentions, summaries, manual or model-driven handoff |
| Provider | Credential scopes, plan entitlements, quotas, actual billing, errors, model behavior, service rules | Invented quota resets or legal interpretation |

### Capability levels

| Level | Public label | Prerequisite | Scope |
|---|---|---|---|
| L0 | Simulated gateway | Protocol/security mock tests | No real provider |
| L1 | One authorized SIWC account | Live OAuth, request + tool + compaction-route proof | One registration |
| L2 | Manual account at launch | L1 for each account, explicit selection and authorities | Separate jobs or compatible resume |
| L3 | Live request-boundary switch | L2 + A→B→A same process/thread + opaque context tests | Exact fingerprint/profile only |
| L4 | Automatic scoped-capacity rotation | L3 + classified safe rejection + permitted operating mode | Exact certified profile only |
| L5 | T3 through ACP | L1/L3 as needed + actual ACP/T3 test | Exact versions and features |
| L6 | Additional provider | Independent official endpoint, billing/policy + runtime test | Separate capability, never implied |

**Rule:** A successful `base_url` override does not prove L3 or L4. A feature is `PASS` only if the corresponding real acceptance evidence exists and remains current. Release manifests MUST declare actual capabilities, not aspiration.

### Explicitly out of scope

- Independent transcript database, agent scheduling, worktree/branch management, subagent handoffs, MCP execution, and task checkpoints.
- Editing provider prompts, encrypted reasoning, compaction records, function calls, or model outputs to improve compatibility.
- Automatically replacing a requested model/provider/service tier with another.
- Consumer Antigravity integration where third-party access is restricted; API/enterprise options are separate products.
- Remote ingress, public Internet exposure, credential resale, pooled accounts across unrelated people, and commercial multitenant operation.
- Probing quota through fabricated inference without explicit budget/consent.

---

# 2. Proof-first feasibility gate (before substantial implementation)

This phase is **mandatory** because the central product benefit depends on facts that cannot be inferred from documentation alone. The smallest implementation should answer the hardest question first.

## 2.1 Proof harness artifacts

Ship `tests/live-gated/` scripts and a deterministic fake Responses server that can:

1. Run a pinned Codex App Server against a custom `base_url` using **only a local bearer**.
2. Verify actual requests use `POST /v1/responses`, `stream:true`, `store:false`, full necessary `input` history, permitted tool shape, and no WebSocket.
3. Validate turn success by native `turn/completed` status **and** upstream `response.completed`; a native message delta is not proof of success.
4. Force local 429/403/503, early SSE rejection, `response.failed`, partial SSE, unknown EOF, bad JSON, cancellation, and server restarts.
5. Confirm whether Codex creates any extra outbound requests beyond `/v1/responses`, including **automatic and explicit compaction**, model catalog, web search or other features used by the intended workload.
6. Confirm the pinned custom-provider retry settings actually disable native HTTP and SSE retries and that the host does not silently resubmit after a committed failure.
7. Exercise a multi-turn, tool-using session A→B→A using separately owned, authorized test registrations and the same native PID and thread ID (if this test pattern is permitted); verify identity attribution.
8. Stress opaque reasoning and compaction portability after a long session and after threshold-triggered compaction, not merely a two-message chat.
9. Test command-backed local credential refresh; the old `env_key` path is an explicit fallback, not a promise of renewal.
10. Compare native T3 compatible-account switching with ACP→gateway behavior before building an ACP-specific compatibility layer.

If (7) or (8) fails, preserve L1/L2; do not secretly rewrite history, strip opaque content, downgrade the model or rebrand a process restart as transparent switching. Capture exact build fingerprints and sanitized traces.

## 2.2 Compaction transport is a release blocker

The general OpenAI API documents both server-side compaction within `/responses` and a distinct `POST /responses/compact` endpoint. **The SIWC plan-sharing preview does not thereby grant that second endpoint.** [R05, R12] Codex may handle compaction locally or differently by build and configuration. **L1-long-session and L3 MUST test actual Codex traffic and automatic/manual compaction before certifying stability.**

If Codex requires unsupported `/responses/compact`, do not proxy it to an unauthorized route or turn it into invented requests. Either configure a documented compatible native compaction mode *without semantic alteration*, certify a separately supported route after provider grant, or mark the runtime/feature incompatible.

## 2.3 Explicit feature permission record

Use a neutral, evidence-based `OperatingModeApproval` document, not a blanket claim about the intent. Required fields:

```yaml
provider_operating_mode:
  route: openai-siwc-responses-http
  deployment_class: local-single-user
  pattern: independent-paid-registrations-request-boundary-selection
  account_ownership: individually-authorized
  interpretation: respects-each-registration-provider-limits
  approval_status: UNRESOLVED # APPROVED | DENIED | UNRESOLVED
  basis: []                # precise Terms section, provider communication, agreement, counsel note
  reviewed_at_utc: null
  reviewer: null
  recheck_triggers: [terms-change, route-change, deployment-change]
```

A legal/provider-operating-mode finding is distinct from a technical proof. Documentation may demonstrate how to maintain multiple OAuth registrations without deciding whether exhaustion-triggered automation is allowed in a particular deployment. `UNRESOLVED` blocks **automatic mode L4**, not mock development, security work or ordinary approved single-account usage. No user-editable `--force` bypass in release builds.

**Language style rule for maintainers and automated agents:** refer to this project as **authorized independent-account routing**. When identifying an actual policy concern, describe the precise provider provision and disputed behavior; do not categorically mischaracterize the engineering mechanism as an exploit, and do not claim verified permission without evidence.

---

# 3. Process architecture and transports

```text
                   agentpool run codex --pool P
                                  |
                          pinned Codex process
                                  |
                         local Responses HTTP
                                  |
             +--------------------v---------------------+
             |             agentpool daemon             |
             |                                         |
             |  DATA PLANE  127.0.0.1:<ephemeral port>  |
             |  bearer -> parse -> policy -> selector  |
             |  -> upstream HTTPS -> raw SSE -> client |
             |                                         |
             |  CONTROL PLANE  owner-only UDS / pipe   |
             |  login; leases; pool; health; policy    |
             |                                         |
             |  CORE / SQLite / OS secret store        |
             +--------------------+---------------------+
                                  |
                        api.openai.com HTTPS
```

One binary contains CLI, daemon, local credential helper, Codex launcher and optional ACP launcher. The binary does **not** implement ACP or Codex's TUI. No remote listeners in v1.

## 3.1 HTTP data plane

| Property | Required implementation |
|---|---|
| Listen | `127.0.0.1:0`; optional `[::1]` only after dual-stack tests; no wildcard address |
| Route | `POST /v1/responses` only for inference in v1; `GET /healthz` local readiness without secrets; separate control IPC for details |
| Local auth | Strong random opaque bearer, scoped to lease/pool/data authority; constant-time validation of protected token verifier |
| Body | Bounded JSON; reject unsupported SIWC profile before network dispatch; preserve raw bytes on success |
| Outbound | Exactly pinned `https://api.openai.com/v1/responses` with TLS hostname verification; no redirects, arbitrary proxy, request-controlled URL or remote DNS authority selection |
| Header rules | Reconstruct upstream request using an allowlist; never forward local Authorization, Host, cookies, client-set OpenAI organization/project/billing headers or hop-by-hop headers blindly |
| SSE | Byte-preserving upstream event-body forwarding, bounded bootstrap classification, response-flush and backpressure, terminal-state verification |
| Cancellation | Downstream close/context cancellation propagates to upstream; no guarantee already-accepted provider work is cancelled |
| Observability | Metadata only, never request/response content or bearer/token dumps |

A local health endpoint must not become an admin endpoint and must not reveal registration identities. The control plane should provide detailed status through authenticated OS-local RPC.

**HTTP implementation precision.** Configure the Go upstream transport to disable automatic decompression; request identity encoding where supported, and correctly strip hop-by-hop headers, `Connection`-nominated headers, invalid `Content-Length`/`Transfer-Encoding` combinations and sensitive downstream-only headers. A response must never claim the compressed upstream length after transparent decompression. Use strict TLS hostname verification (not fragile static IP or certificate pinning), no redirects, and a bounded connection pool. Avoid automatically propagating client-requested `Host`, `OpenAI-Organization`, `OpenAI-Project`, cookies, proxy forwarding metadata or query parameters. A checked-in request/response header policy must be covered by adversarial fixtures.

## 3.2 Owner-only control IPC

macOS/Linux: private parent directory mode `0700`, UDS owner-only; reject symlinks and unsafe paths, verify peer OS user where supported. Windows: per-user named pipe with verified ACL and user identity. One daemon elected through an atomic exclusive lock with process-identity/health reconciliation. Lock alone is not proof of health; a stale lock is not license to start two refreshers.

Use **JSON-RPC 2.0 over length-delimited messages** with a 1 MiB control-message cap, bounded nesting, a published `schema/control-v1.json`, response correlation, and a stable typed `error.data`. Protocol major version mismatch fails closed. Mutating commands require both owner identity and a configured administrative caller role. A same-UID hostile process remains within the explicitly disclosed trust limit unless additional OS entitlement checks are installed.

## 3.3 Service lifecycle

States: `STARTING`, `READY`, `DRAINING`, `STOPPED`, `DEGRADED`, `FAULTED`. Readiness requires SQLite migrations, secret-store availability *or an explicit no-registration mode*, policy validity, safe binding, and functioning RPC. Health failures are not mistaken for capacity exhaustion.

Shutdown: close admission, optionally drain bounded active streams, checkpoint SQLite, destroy uncommitted secure spools, revoke launcher leases as appropriate, exit. On kill/crash: in-flight HTTP fails; restart *must not* treat prior requests as completed and *must not* automatically resend them.

macOS first deployment uses a user LaunchAgent with opt-in automatic restart and no root privileges. A restart loop must surface `FAULTED`; it cannot silently redirect clients to another Codex account.

---

# 4. Identity, account ownership and data authority

## 4.1 The five identities

Do not conflate:

1. **OS owner** — human who operates the local daemon.
2. **Local client** — a Codex process, TradeAgent worker, ACP bridge, or explicitly authorized launcher.
3. **Workload authority** — confidentiality and intended workspace/domain for the incoming inference request.
4. **Provider principal** — validated OIDC subject and applicable workspace/tenant.
5. **Provider app registration** — issued SIWC `client_id`, OAuth scopes, host ID and tokens.

Two registrations with the same email or subject may be distinct OAuth grants but **not** independent paid allowances. Two different subjects may not have permission to see the same code. Never infer either from an email label.

## 4.2 Registration record

```go
// Conceptual stable domain model. Secrets are referenced, never embedded.
type Registration struct {
    ID                 RegistrationID
    Route              RouteID
    IssuedClientID     string
    OIDCSubject        string
    WorkspaceID        *string
    CapacityOwnerKey   string // privacy-preserving provider principal/workspace entitlement grouping
    HostID             string
    GrantedScopes      []string
    CredentialRef      string
    CredentialVersion  uint64
    OwnerApprovalID    string
    AuthorityGroupID   string
    BillingMode        BillingMode
    Enabled            bool
    PolicyState        PolicyState
    VerifiedAt         time.Time
}
```

Registration identity is the verified tuple `{route, issued_client_id, host_id, subject, workspace}`. Tokens must never be paired with metadata from a different tuple. `email` is a display hint only.

`capacity_owner_key` is derived from validated principal/workspace identity and known entitlement boundaries, using a keyed, opaque local identifier. It is not a guessed per-account remaining-token value and not necessarily proof of separately paid plans. A principal-scoped provider limit applies across matching capacity owners, whereas an app-registration-specific limit affects only matching registrations. The exact scoping rule is provider-adapter data, never a global hardcoded assumption.

## 4.3 Workload lease / authority binding

The gateway cannot safely inspect prompts to learn whether they contain client data. Therefore the trusted **host** must bind the workload's authority to a newly issued local lease *before inference*.

```text
Lease {
  lease_id:              opaque UUID
  client_id:             registered local app/launcher ID
  process_family_id:     trusted launcher process family
  pool_id:               selected pool
  authority_group_id:    immutable classification
  permitted_models:      exact allowlist or declared all-certified
  permitted_mode:        L1 | L2 | L3 | L4
  concurrency_cap:       small positive integer
  issued_at, expires_at
  secret_verifier:       hash/HMAC, no plaintext bearer in SQLite
  revoked_at:           nullable
}
```

**No arbitrary caller-provided `authority_group` header authorizes a change.** The local bearer selects its server-side lease. Reclassifying a running worker requires a new explicitly approved lease and a safe session boundary. Editing pool membership never broadens existing lease authority. Client → pool → registration eligibility uses the intersection of all independent authority constraints.

Default local trust policy only grants credentials to launched, owner-approved applications. Document same-UID compromise explicitly; OS-local bearer does not defend against malware with full user privileges.

## 4.4 Billing and model invariants

- Default `subscription-included-only` and selected model slug remain immutable across a request and its allowed safe alternate attempt.
- No ambient `OPENAI_API_KEY`, `CODEX_API_KEY`, `OPENAI_BASE_URL`, inherited proxy or system Codex login may change the route.
- No automatic charged credits, API metering, trial credits or pay-as-you-go fallback. Such a feature would require a separate explicit product design, opt-in, hard spend caps, test and receipt evidence.
- Account-specific model catalog is an indication of what can be selected; successful inference on that exact model is stronger evidence, but not permanent entitlement.
- If authoritative per-request billing evidence is unavailable, output `billing_evidence: unavailable`, not `zero_charges_verified`.

---

# 5. Official SIWC registration and credential lifecycle

## 5.1 First registration

Use only official documented SIWC endpoints and parameters. The initial browser authorization starts with `client_id=dynamic_agent_client`, a stable per-installation `ext_agent_host_id`, actual stable `agent_name_hint`, `state`, `nonce`, PKCE S256, exact loopback redirect URI, requested scopes and resource. Validate the callback's `state` and issued client ID before exchanging a code. Exchange **with the issued client ID**; `dynamic_agent_client` is not the saved credential identity. [R01, R03]

Validate OIDC signature against authorized issuer JWKS, issuer, audience, expiry/not-before, nonce, subject/workspace and scope; reject invalid login without overwriting a previously valid registration. Required plan scope: `chatgpt.tokens.use.direct`. Optional `offline_access` is needed for refresh. Never treat authentication alone as plan inference entitlement.

Returning authorization must match original subject/workspace/client identity before credential replacement. Consent withdrawal and logout must invalidate local routing immediately; attempt documented remote revocation and truthfully report whether confirmed.

## 5.2 Refresh handling

As currently documented, access tokens live about one hour and each successful refresh produces a **replacement** refresh token with a new 30-day lifespan. Future revisions may change this; use returned expiry and `earliest_refresh_at`. [R04]

State machine:

```text
UNREGISTERED -> AUTHORIZING -> EXCHANGING -> VALIDATING -> ACTIVE
                                      \-> INVALID / PLAN_DISABLED
ACTIVE -> REFRESH_INTENT_DURABLE -> REFRESHING -> COMMIT_SECRET -> ACTIVE
                  |                      |
                  |                      +-> definite invalid_grant -> REAUTH_REQUIRED
                  +-> crash/uncertain exchange -> QUARANTINED_UNCERTAIN
ACTIVE -> REVOKING -> REVOKED/SIGNED_OUT
ACTIVE -> EXPIRED/UNUSABLE -> REAUTH_REQUIRED
```

**Refresh transaction protocol (required):**

1. Per-registration singleflight mutex across all goroutines; daemon election prevents second local writer.
2. Re-read current persisted secret and generation, verify expected subject and issued client ID.
3. Persist `RefreshIntent(registration, old_generation, attempt_id, started_at)` in SQLite **before** sending refresh. This records potential uncertainty, not a duplicate retry authorization.
4. Send official refresh with saved issued `client_id`, saved `refresh_token`, exact resource, no accidental scope additions.
5. Validate response and store the *new complete credential pair* as a versioned, atomic **single secret-store value**. Force durable write where platform supports it.
6. Commit the new credential generation and clear intent in SQLite. On startup reconcile SQLite generation with secret record generation; fail closed on mismatch unless monotonic recovery is demonstrably safe.
7. Crash after provider acceptance but before securely saving replacement is intrinsically uncertain; do **not** repeat a potentially consumed refresh token. Require reauthentication if it cannot be safely reconciled.
8. `invalid_grant` / documented terminal refresh errors → `REAUTH_REQUIRED`; temporary network/5xx → retain credentials and bounded later retry only when refresh outcome is known safe.
9. Every outcome is redacted; token contents, redirect URLs containing hints and credential payloads never enter logging or support artifacts.

This design deliberately admits the remote-token-rotation/local-storage atomicity limit rather than claiming exactly-once refresh across two systems.

## 5.3 OS secret store

Use a `SecretStore` abstraction: macOS Keychain first; Windows supported user-protected credentials; Linux Secret Service when available. A headless encrypted-file fallback is **explicit opt-in** and must specify where its actual encryption key is protected; file permissions are necessary but not encryption. No silent plaintext fallback. SQLite contains only public identifiers, metadata, generation and secret references.

Logout with active requests: immediately revoke **new admission**, invalidate bindings, and cancel active streams for revoked/compromised or data-authority-removed registrations. Ordinary user-requested account disable may optionally bounded-drain current streams if this behavior was displayed and authorized. Either way, the decision is explicit and recorded without prompt content.

## 5.4 Refresh/lease distinction

Provider OAuth tokens are held by AgentPool; Codex sees only the *local gateway bearer*. Refreshing the upstream token does not require restarting Codex **if** its gateway path is already using a valid local credential, because AgentPool constructs each new upstream request authorization. Local bearer refresh is a separate problem and is solved in Section 6.

---

# 6. Codex integration contract — exact configuration

## 6.1 Default: command-backed local bearer

Current Codex docs support a command-backed bearer for custom model providers using `[model_providers.<id>.auth]`, which **must not** be combined with `env_key` or `requires_openai_auth`. The command prints a token to stdout. [R08]

The v3 preferred implementation uses a **local lease helper** that authenticates over private IPC, retrieves/mints the current local bearer for the already authorized process family, prints only the bearer to stdout, and exits. It receives a non-secret opaque launcher context reference in its args. The IPC server must not mint privileged leases based solely on an untrusted, guessable context string.

Conceptual launcher-generated config (not an unconditional copy/paste into every user's Codex installation):

```toml
model_provider = "agentpool"

[model_providers.agentpool]
name = "AgentPool local Responses gateway"
base_url = "http://127.0.0.1:<RUNTIME_PORT>/v1"
wire_api = "responses"
supports_websockets = false
request_max_retries = 0
stream_max_retries = 0

[model_providers.agentpool.auth]
command = "/absolute/path/to/agentpool"
args = ["lease-token", "--context", "<OPAQUE_LAUNCH_CONTEXT>"]
refresh_interval_ms = 60000
timeout_ms = 3000
```

The actual command behavior, refresh cadence and error handling MUST be probed against the pinned Codex build. Do not assume `auth.command` supports the same behavior on all platform builds or through ACP. The helper never prints upstream SIWC OAuth tokens. If a malicious local process can invoke it, authorization must still require an owner-approved process-family lease; same-user OS compromise remains out of scope.

**Peer-binding requirement.** The launcher registers the intended child process with OS process identity and start-time evidence when available; the helper validates the peer process, approved executable/session family and still-active lease through owner-only IPC. If the platform cannot establish the expected peer binding reliably, document the narrower same-UID trust guarantee and limit renewal authority to the original preapproved, short-scope lease. A visible `--context` value is never accepted as independent proof of client identity.

If the configuration cannot be certified, use the documented env-key mode:

```toml
model_provider = "agentpool"
[model_providers.agentpool]
name = "AgentPool local Responses gateway"
base_url = "http://127.0.0.1:<RUNTIME_PORT>/v1"
wire_api = "responses"
env_key = "AGENTPOOL_LOCAL_BEARER"
requires_openai_auth = false
supports_websockets = false
request_max_retries = 0
stream_max_retries = 0
```

**Fallback limitation:** An environment-provided bearer can remain fixed throughout the lifetime of a native process. Renew upstream provider OAuth inside AgentPool, but the local lease for that Codex process must have an explicitly bounded and suitable lifetime; no claim of endless lease refresh until proved. A deliberate native restart-and-resume operation may be required when that local lease expires.

## 6.2 Native retry disabling is a safety invariant

Published Codex custom-provider defaults include HTTP and SSE retry counts (`request_max_retries=4`, `stream_max_retries=5`). **Both MUST be explicitly set to `0` in the certified critical profile, and observed to take effect**. [R08]

A gateway's own "do not replay after commit" rule is not end-to-end if Codex or TradeAgent invisibly resubmits the inference. The complete native → gateway → host pipeline MUST ensure that:

- No automatic native stream reconnection causes a *second* authoritative inference after a committed first stream.
- A native parser failure after `response.completed` does not silently recreate a turn that already led to tool actions.
- A host supervisor cannot blindly restart/replay a turn after ambiguous transport loss; host decides whether tools/external effects require reconciliation.
- A disabled retry setting is verified in a fault harness, not simply present in a config file.
- Native Codex's SSE idle timeout must be explicitly certified against gateway and upstream idle-time budgets; a gateway allowing longer silence does not prevent an earlier native timeout.

If a future Codex version changes retry configuration, mark it incompatible until recertified. If native retry is eventually re-enabled for benign failures, it requires a new end-to-end lineage and idempotency proof.

## 6.3 Launch algorithm

1. Verify executable SHA-256, version, supported configuration precedence, model profile and feature settings.
2. Confirm daemon readiness and effective `OperatingModeApproval` + client capability admission.
3. Authenticate local launcher; issue least-privilege process-family lease bound to pool, authority, model and mode.
4. Generate ephemeral profile outside shared user `config.toml` and outside source repositories; permissions owner-only. Do not overwrite ordinary Codex login or home.
5. Set exact local provider endpoint and credential mechanism; scrub ambient API keys/base URL/proxy overrides and unexpected `CODEX_HOME` only where they would change routing. Preserve necessary OS/terminal/workspace environment variables.
6. Enforce disabled native retries; verify tests cover config layer precedence and actual outgoing wire traffic.
7. Execute real `codex` or `codex app-server`, preserving native IO/protocol and exit status; own child lifecycle and revoke lease when session exits.
8. If config is ignored, native auth takes precedence, the model changes, or requests bypass gateway: **stop and classify incompatible**. Do not quietly switch to a private route or charged API key.

## 6.4 SIWC HTTP shape admission

At minimum v1 permits only supported `POST /v1/responses` HTTP with:

- `store:false`, `stream:true`, required context in `input`, no `previous_response_id`.
- No unsupported fields documented in the current SIWC preview (`background`, `conversation`, `max_output_tokens`, `max_tool_calls`, `metadata`, `moderation`, `multi_agent`, `prompt`, `prompt_cache_retention`, `safety_identifier`, `temperature`, `top_logprobs`, `top_p`, `truncation`, `user`, etc.; refresh list at release). [R05]
- No invalid explicit `role:"system"` message shape where disallowed. Preserve native `instructions` and acceptable developer input, never silently rewrite.
- Requested model unchanged and evidence-backed; local tools and compatible custom/function tools permitted; unsupported hosted tools fail visibly.
- WebSocket disabled and stateless HTTP history present.
- Encrypted reasoning and native compaction items passed **byte-for-byte**; unrecognized feature is rejected as incompatible instead of rewritten.
- Provider-native exceptions/feature sets evaluated against the actual current preview, not inferred from generic OpenAI Responses API documentation.

## 6.5 App Server and originator identity

When initiating Codex App Server directly, set stable `initialize.params.clientInfo.{name,title,version}`. OpenAI explicitly documents that the app name is used for attribution and should agree with `agent_name_hint` used at SIWC registration. [R06] The launcher must certify how the selected Codex/ACP build sets or forwards that originator. Do not spoof T3 as AgentPool or vice versa; preserve the real calling application's identity consistently with the registration's authorized app identity.

## 6.6 T3 and ACP

The pinned maintained `agentclientprotocol/codex-acp` starts native Codex App Server and exposes gateway authentication **only when the ACP client advertises the relevant gateway capability**. [R09]

`agentpool acp codex` may do only: validate binary/config, obtain local lease, provide tested `CODEX_CONFIG`/`MODEL_PROVIDER`, exec maintained `codex-acp`, pass ACP stdio untouched. If the actual T3 build does not negotiate gateway auth or preserve the intended provider, mark that integration unsupported rather than fabricating an ACP layer.

**Alternative supported product path:** T3 Code documents native multiple ChatGPT accounts with compatible thread continuation and shared/shadow `CODEX_HOME` arrangements. [R10] Evaluate that as a client-specific account-routing path; it may be better than the HTTP gateway for T3 even if AgentPool remains the policy/status facade. Native T3 switching is *not* evidence of automatic quota-triggered request-boundary rotation and must not be mislabeled L4.

---

---

# 7. Routing domain and persistence

## 7.1 Minimal persistent objects

| Object | Stored? | Purpose |
|---|---|---|
| `Registration` | SQLite metadata + secret-store reference | One specific issued provider grant |
| `Pool` / `PoolMember` | SQLite | Approved candidates, weights, authority and policy |
| `ClientLease` | SQLite metadata/hashed verifier | Local client privileges and routing/session family |
| `Binding` | SQLite, not in-flight correctness | Affinity for one trusted lease/session group |
| `CapacityObservation` | SQLite bounded | Scope, freshness and evidence of capacity/limits |
| `CapabilityAdmission` | SQLite + signed evidence manifest | Exact-build conformance and feature availability |
| `PolicyClearance` | Signed/local reviewed artifact | Applicable operating mode and approval status |
| `RefreshIntent` | SQLite | Recovery of OAuth token-rotation uncertainty |
| `RequestAttempt` | **memory only** | Actual in-flight stream/transport and reservation |

No `Job`, `Agent`, `Thread`, `ToolCall`, `Checkpoint`, `Workspace`, `TaskScheduler` or `MessageTranscript` persistence in AgentPool. A `Binding` represents routing affinity, **not** native conversation ownership.

## 7.2 Scoped capacity is not a scalar token counter

```go
type CapacityObservation struct {
    RegistrationID RegistrationID
    Route          RouteID
    ScopeKind      ScopeKind // app_registration | principal | model | bucket | provider | unknown
    ScopeKey       string    // opaque provider evidence, if available
    Event          CapacityEvent // available | limited | unavailable | unknown
    ProviderCode   string
    Evidence       EvidenceKind // structured_error | trusted_telemetry | manual | heuristic
    ObservedAt     time.Time
    ExpiresAt      *time.Time
    ResetAt        *time.Time
    Confidence     Confidence // scope-specific, never universal
}
```

An explicit `subscription_sharing_usage_limit_exceeded` event may be app-specific. **Do not infer the entire ChatGPT plan, other app registrations, unrelated accounts, other models or a guessed reset time are exhausted.** The provider documents `subscription_sharing_usage_unavailable` separately from a limit. [R07]

A successful request is evidence that the account served *that exact request at that moment*. It does not prove unlimited remaining capacity, a predictable quota balance, or zero charge. Querying model listings and ordinary Codex account rate-limits does not automatically provide per-SIWC-registration headroom behind the gateway. Do not fabricate pre-exhaustion percentages.

## 7.3 Selection logic

For request `r`, lease `l`, account `a`, immutable policy snapshot `p`:

```text
eligible(a,r,l,p) =
    valid_lease(l) and l.pool == p.pool and l.authority == p.authority
 and a.enabled and a.member_of(p.pool)
 and a.owner_approved and a.authority_allows(l.authority)
 and provider_mode_admitted(p.mode, l.mode)
 and a.oidc_identity_valid and a.scope_allows_request
 and a.billing_mode == subscription-included-only
 and a.provider_route == requested_route
 and exact_model_is_certified_or_safely_probeable(a, r.model)
 and client_request_profile_is_admitted(r,l)
 and not authoritatively_limited(a, matching_request_scope(r))
 and admission_capacity_available(a,l,p)
 and not a.revoked and not a.policy_hold
```

**Unknown distinctions:**

- `unknown_capacity` → optional one bounded attempt if identity, consent, profile and model admission are otherwise valid.
- `unknown_permission`, `unknown_data_authority`, `unknown_billing_path`, `unknown_client_shape` → **deny**.
- `unknown_account_model_entitlement` → either safe probe under explicit policy and same requested model or deny; never silently substitute.
- `unknown_provider_limit_scope` → no broad account cooldown without evidence; use cautious local retry/backoff, not global quota extrapolation.

## 7.4 Deterministic allocator

1. Compute eligible registrations under a versioned immutable policy/lease snapshot.
2. Prefer the current binding if eligible and no switch/drain is pending.
3. Otherwise choose among eligible candidates with deterministic **smooth weighted round-robin** and stable tie-breaking by registration ID. Weights are explicit user preferences, not estimates of plan size.
4. Reserve global, per-client, per-account and per-binding concurrency **atomically under one allocation critical section** before network dispatch. Reservation stores snapshot policy version and binding generation.
5. Recheck revocation epochs and authority before sending upstream; cancellation/disable can revoke an already reserved dispatch.
6. Release tokens once on all exits, including panic/error/cancel, with resource tracking and leak assertions.
7. Under quota evidence, record precise scope, remove only the affected candidates and reevaluate. Do not spin or select a just-rejected account again in one request lineage.
8. A registration that returns a provider error not specifically indicating exhaustion must not be treated as empty fuel.

**Concurrency implementation decision:** a single in-memory allocator lock for the single-daemon v1 is preferable to a clever distributed reservation scheme. Bounded critical-section work only, no network calls while holding the lock. Global fairness is enforced by small FIFO per-client queues with round-robin queue selection. Concurrency caps below are conservative placeholders, **not documented provider limits**.

**Revocation ordering.** `Reserve` issues an epoch-bound dispatch permit. `BeginReservedAttempt` and the revocation control path synchronize on that permit state so no **new application-level dispatch** is begun after revocation has taken effect. Already dispatched bytes may still be in the OS/network/provider pipeline; a later cancellation cannot erase them. The runbook must describe that unavoidable external timing limit instead of promising instantaneous remote revocation of an in-flight request.

## 7.5 Affinity and conservative switch barrier

There is no reliable assumption that a Codex inference HTTP request contains a stable native thread ID. Therefore the v1 safe identity is a trusted **launcher process-family lease**, not an attacker-supplied session header. One lease can conservatively cover multiple native child/subagent threads. This may reduce parallelism but prevents hidden cross-account overlap. An improved actual per-thread ID is allowed only through a documented, conformance-tested client integration.

```text
ACTIVE(S, A, generation=12, outstanding=2)
 -> SWITCH_PENDING(S, A -> B)
 -> STOP_NEW_DISPATCH_FOR_S
 -> WAIT_FOR_ALL_ACCEPTED_A_REQUESTS_TO_END (bounded)
 -> if timeout: FAIL_OR_PAUSE_S (no blind B dispatch)
 -> RECHECK_B_AUTHORITY, MODEL, BILLING, POLICY, CAPACITY
 -> CAS_BINDING(S, generation=12 -> 13, account=B)
 -> UNBLOCK_NEW_REQUESTS_FOR_S
```

A **switch barrier is about concurrent inference lifetimes**, not copying native context. When the previous registration has active streams, new B requests from that lease do not start until the barrier completes. If a process contains multiple independent sessions and this conservative grouping is unacceptable, use separate launch leases/processes or decline L3 for that runtime.

Bindings are keyed by `HMAC(local_binding_key, lease_id || certified_session_hint)` and have TTL. Persisted affinity is an optimization only; no persisted binding implies a stream survived a crash. On restart, set outstanding counts to zero only because all old streams are known to have been severed; native host must reconcile before resuming, and L3 needs compatible re-admission.

CAS and policy epoch prevent a stale goroutine from resetting a newer selection. Account logout, authority revoke and policy change invalidate relevant bindings immediately.

## 7.6 SQLite logical schema (migration 0001)

This DDL defines **minimal constraints and storage boundaries**; repository migrations must include indexes, triggers as appropriate and schema integration tests. Values serialized as JSON must be validated to a versioned schema before commit. SQLite foreign keys MUST be enabled on every connection.

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE registrations (
  id TEXT PRIMARY KEY,
  route_id TEXT NOT NULL,
  issued_client_id TEXT NOT NULL,
  oidc_subject TEXT NOT NULL,
  workspace_id TEXT NOT NULL DEFAULT '',
  capacity_owner_key TEXT NOT NULL,
  host_id TEXT NOT NULL,
  secret_ref TEXT NOT NULL UNIQUE,
  credential_generation INTEGER NOT NULL DEFAULT 0 CHECK (credential_generation >= 0),
  granted_scopes_json TEXT NOT NULL,
  authority_group_id TEXT NOT NULL,
  owner_approval_id TEXT NOT NULL,
  billing_mode TEXT NOT NULL CHECK (billing_mode IN ('subscription-included-only')),
  policy_state TEXT NOT NULL CHECK (policy_state IN ('APPROVED','HOLD','DENIED','REVOKED')),
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
  verified_at_utc TEXT,
  UNIQUE(route_id, issued_client_id, host_id)
);

CREATE TABLE pools (
  id TEXT PRIMARY KEY,
  route_id TEXT NOT NULL,
  authority_group_id TEXT NOT NULL,
  billing_mode TEXT NOT NULL CHECK (billing_mode = 'subscription-included-only'),
  operating_mode TEXT NOT NULL CHECK (operating_mode IN ('L1','L2','L3','L4')),
  policy_json TEXT NOT NULL,
  version INTEGER NOT NULL CHECK (version > 0)
);

CREATE TABLE pool_members (
  pool_id TEXT NOT NULL REFERENCES pools(id) ON DELETE CASCADE,
  registration_id TEXT NOT NULL REFERENCES registrations(id) ON DELETE RESTRICT,
  weight INTEGER NOT NULL CHECK (weight BETWEEN 1 AND 100),
  PRIMARY KEY(pool_id, registration_id)
);

CREATE TABLE client_leases (
  id TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  process_family_id TEXT NOT NULL,
  pool_id TEXT NOT NULL REFERENCES pools(id),
  authority_group_id TEXT NOT NULL,
  permitted_mode TEXT NOT NULL,
  permitted_models_json TEXT NOT NULL,
  secret_verifier BLOB NOT NULL,
  concurrency_cap INTEGER NOT NULL CHECK (concurrency_cap BETWEEN 1 AND 32),
  issued_at_utc TEXT NOT NULL,
  expires_at_utc TEXT NOT NULL,
  revoked_at_utc TEXT
);

CREATE TABLE bindings (
  pool_id TEXT NOT NULL REFERENCES pools(id),
  lease_id TEXT NOT NULL REFERENCES client_leases(id),
  opaque_session_key BLOB NOT NULL,
  registration_id TEXT NOT NULL REFERENCES registrations(id),
  generation INTEGER NOT NULL CHECK (generation >= 0),
  expires_at_utc TEXT NOT NULL,
  PRIMARY KEY(pool_id, lease_id, opaque_session_key)
);

CREATE TABLE capacity_observations (
  id TEXT PRIMARY KEY,
  registration_id TEXT REFERENCES registrations(id) ON DELETE CASCADE,
  route_id TEXT NOT NULL,
  scope_kind TEXT NOT NULL,
  scope_key TEXT NOT NULL DEFAULT '',
  event TEXT NOT NULL,
  provider_code TEXT,
  evidence_kind TEXT NOT NULL,
  confidence TEXT NOT NULL,
  observed_at_utc TEXT NOT NULL,
  expires_at_utc TEXT,
  reset_at_utc TEXT
);

CREATE TABLE refresh_intents (
  registration_id TEXT PRIMARY KEY REFERENCES registrations(id) ON DELETE CASCADE,
  expected_generation INTEGER NOT NULL,
  attempt_id TEXT NOT NULL UNIQUE,
  started_at_utc TEXT NOT NULL,
  status TEXT NOT NULL CHECK(status IN ('PREPARED','SENT','RESULT_UNCERTAIN'))
);

CREATE TABLE capability_admissions (
  profile_digest TEXT NOT NULL,
  route_id TEXT NOT NULL,
  feature TEXT NOT NULL,
  state TEXT NOT NULL CHECK (state IN ('PASS','FAIL','UNVERIFIED','STALE')),
  evidence_ref TEXT NOT NULL,
  verified_at_utc TEXT NOT NULL,
  expires_at_utc TEXT NOT NULL,
  PRIMARY KEY(profile_digest, route_id, feature)
);

CREATE TABLE schema_migrations (
  version INTEGER PRIMARY KEY,
  checksum TEXT NOT NULL,
  applied_at_utc TEXT NOT NULL
);

CREATE INDEX ix_obs_reg_scope_time ON capacity_observations
 (registration_id, scope_kind, scope_key, observed_at_utc);
CREATE INDEX ix_bindings_account ON bindings(registration_id);
CREATE INDEX ix_leases_expiry ON client_leases(expires_at_utc);
```

**DB guardrails:** all writes via parameterized queries; no user-supplied SQL; owner-only file and parent; `busy_timeout`, WAL if platform filesystem supports it, backup API for consistent snapshots, migration checksums, integrity tests, file lock and corruption quarantine. Foreign keys do not by themselves verify that a pool and its members have compatible authority groups; enforce this in **atomic application transaction validation** plus negative tests.

## 7.7 Single-process allocation vs persisted data

In-flight counters, request attempts, stream handles, active switch barriers and cancellation maps are ephemeral and **must never be reconstructed as successful work** from SQLite. Persist only account/lease/policy/affinity/observations needed to make a fresh decision after restart.

Operational logs are not a hidden transcript DB; use bounded rotation and minimal sanitized metadata. For tests, synthetic fixtures can retain content under an isolated test-only storage root.

---

# 8. Streaming and retry safety — normative state machine

## 8.1 One downstream request, bounded eligible attempts

```text
RECEIVED
   | validate bearer, authority, JSON shape and body limits
VALIDATED --------------------------------------> LOCAL_DENIED
   | reserve account under current policy
RESERVED ---------------------------------------> NO_CAPACITY
   | send exact request to official route
SENT_UPSTREAM
   +-- explicit no-accept rejection --------------> CLASSIFY -> RELEASE
   |                                                 | if permitted, safe
   |                                                 v
   |                                         ANOTHER ELIGIBLE ATTEMPT
   +-- no response / unknown accept ----------------> AMBIGUOUS_FAILURE
   +-- valid stream bootstrap ----------------------> COMMITTED
                                                       | raw SSE, no replay
                                                       +-> COMPLETED
                                                       +-> POSTCOMMIT_FAILED
                                                       +-> CANCELLED
```

**Critical distinction:** "no bytes yet delivered to Codex" does not prove the provider did no work. An alternate attempt is allowed only when the provider's **structured response proves the first attempt was rejected before inference acceptance**, the initial upstream connection has been terminated, and the mode authorizes switching. A malformed EOF or timeout is not such proof.

## 8.2 Downstream commitment

Commit starts with the *first downstream-visible upstream status/headers/body information*. Precommit buffering may inspect status and an initial bounded SSE prefix solely to determine whether a recognized early failure is safely replayable. The implementation cannot assume HTTP 200 means success, and cannot treat a `response.failed` event as a no-accept rejection without specific evidence. SIWC documents failure events occurring during an SSE stream. [R07]

Once committed:

- Do not switch to registration B within the current HTTP response.
- Do not issue a second inference on the same incoming request.
- Do not fabricate `response.completed` if SSE ended abruptly or returned `response.failed`/`response.incomplete`.
- Preserve underlying SSE bytes/events, tool-call data, event IDs, ordering and response termination.
- If HTTP status is already sent, cannot replace it with a late local error; close/report stream failure in the tested native client's expected way and expose a correlated sanitized diagnostic over control IPC.

## 8.3 Safe retry classes

| Class | Exact evidence | Gateway behavior |
|---|---|---|
| `SAFE_EXPLICIT_NO_ACCEPT` | Versioned, recognized structured rejection proven to represent no accepted inference | At most one alternate authorized attempt if L4 + pool policy allow; no duplicate upstream in parallel |
| `AMBIGUOUS_PRECOMMIT` | Network reset, timeout, no trustworthy response, unexpected EOF | Return typed ambiguous error, **no automatic replay** |
| `COMMITTED_FAILURE` | Any failure after downstream-visible upstream information | Forward failure/terminate; **no automatic replay** |
| `UNSUPPORTED_REQUEST` | Invalid route, model, parameter, tool shape or build | Local/provider error, no account cycling |
| `AUTH_FAILED` | Invalid local lease, OAuth revocation/terminal failure | Disable affected admission, no silent credential replacement |
| `PROVIDER_UNAVAILABLE` | Provider-wide overload or failure | Scoped breaker/backoff; do not spray registrations |
| `NO_AUTHORIZED_CAPACITY` | No account eligible after scoped filtering | Typed local response, no long blocking socket |

**Provider-specific proof rule:** Not every HTTP 429 guarantees non-acceptance and not every app-specific limit means a whole paid subscription exhausted. The route classifier maps only documented exact error shapes to `SAFE_EXPLICIT_NO_ACCEPT`; unknown response codes fail conservative.

## 8.4 Bootstrap and streaming algorithm

```go
// Pseudocode: reviewable safety lifecycle, not an entire Go handler.
func HandleResponses(ctx context.Context, in Incoming) Result {
    meta, snapshot, err := ValidateAndFreeze(in) // body caps, shape, authority
    if err != nil { return LocalError(err) }
    defer snapshot.DestroySecurely()

    attempted := map[RegistrationID]bool{}
    for n := 0; n < 2; n++ { // Includes the first attempt; configurable only within proven bounds.
        reservation, err := allocator.Reserve(ctx, meta, attempted)
        if err != nil { return TypedNoCapacity(err) }
        handle := upstream.Begin(ctx, reservation, snapshot)
        outcome := handle.ReadBoundedBootstrap()
        switch outcome.Kind {
        case SafeExplicitNoAccept:
            attempted[reservation.Registration] = true
            capacity.ObserveScoped(outcome.Evidence, reservation)
            handle.CloseCancelReleaseOnce()
            if n == 0 && policy.AllowsAlternate(meta, outcome) { continue }
            return ExposeRejection(outcome)
        case ValidStream:
            result := handle.CommitHeadersAndSSEThenStream(ctx, outcome)
            handle.CloseCancelReleaseOnce()
            return result // NEVER hidden-retry here.
        case Ambiguous:
            handle.CloseCancelReleaseOnce()
            return TypedAmbiguousFailure(outcome)
        default:
            handle.CloseCancelReleaseOnce()
            return TypedFailure(outcome)
        }
    }
    return TypedSafeAttemptBudgetExhausted()
}
```

Necessary implementation refinements:

- Route reservation and barrier completion are **explicit prerequisite states** of `Begin`; the pseudocode assumes that contract.
- `CloseCancelReleaseOnce()` must be idempotent even under downstream close, timeout, panic and racing shutdown.
- `snapshot` is immutable and replayable only for recognized safe no-accept failures. A readable copy of a request does not prove inference can be resent safely.
- Bootstrap classifier has both byte and elapsed-time caps; over-cap bootstrap without safe classification must **commit a valid stream or fail without replay**, never allocate unbounded memory.
- Do not buffer the full streaming response; apply bounded buffering and Go HTTP backpressure, periodically flush.
- Verify response terminal event/EOF rather than relying on HTTP status alone; HTTP 200 can carry `response.failed`.
- No proxy-side JSON reserialization after validation; preserve raw request body whenever the route permits.
- Strict total body cap, spool cap and stream timeouts prevent local DoS. Do not impose an arbitrary short total timeout that kills legitimate long reasoning gaps; first-byte, idle and optional total deadlines have separate documented defaults.

## 8.5 Local error envelope

Only before downstream commitment may AgentPool return its own stable JSON error:

```json
{
  "error": {
    "code": "agentpool_no_authorized_capacity",
    "message": "No approved registration can serve this request under the current pool and workload policy.",
    "retryable": true,
    "next_known_reset_at": null,
    "request_id": "opaque-request-id"
  }
}
```

A local error is an AgentPool extension, **not** an OpenAI response object. Codes and mapping are defined in `schema/errors-v1.json`; exact Codex versions must be observed to display or recover from them correctly.

## 8.6 Failure classification table

| Evidence | Meaning | Action |
|---|---|---|
| `subscription_sharing_usage_limit_exceeded` | Quota relevant to app/registration/path, not proof of global plan exhaustion | Mark exact known scope; safe alternate attempt only if provider rejected before inference and L4 allowed |
| `subscription_sharing_usage_unavailable` | Usage service unavailable | Transient backoff, **not** account empty |
| `subscription_sharing_user_not_eligible` | User/plan eligibility issue | Disable affected route; diagnostic, no speculative cycling |
| `subscription_sharing_unsupported_capability` | Request/model/tool shape not supported | Incompatibility, not capacity issue |
| `subscription_sharing_route_not_supported` | Endpoint/protocol unsupported | Incompatibility; do not invent alternate API route |
| `subscription_sharing_invalid_user` | Subscriber context unavailable/invalid | Diagnose/re-auth as appropriate |
| `chatpass_v2_scope_not_authorized` | Grant/scope incorrect | Stop affected registration; require explicit correction |
| OAuth `invalid_grant` | Refresh credential unusable | Reauthenticate correct registration |
| Unknown 401/403/429 | Unknown scope/permission/limit | Surface exact safe status; no inferred depletion |
| 5xx/timeout | Provider/network may be unavailable | Breaker/backoff; no simultaneous account spray |
| SSE `response.failed`/`response.incomplete` | Inference did not complete successfully | Preserve semantics; no hidden continuation |
| Client-side disconnection | Local transport terminated | Cancel upstream; provider-side cancellation may be uncertain |

All provider-specific codes require exact shape tests and revision tracking. The table is derived from official error documentation [R07] and does not authorize adding undocumented retries.

---

# 9. Concurrency, deadlines, fairness and capacity inference

## 9.1 Bounds, defaults and overload response

Initial tunable **engineering candidate values**, to be load-tested and adjusted per machine; these are *not* provider quota values:

```yaml
limits:
  active_requests_global: 24
  active_requests_per_client: 6
  active_requests_per_registration: 3
  active_requests_per_process_family: 6
  pending_admission_global: 64
  pending_admission_per_client: 12
  admission_max_wait_ms: 1500
  request_body_max_mib: 16
  request_body_memory_threshold_mib: 1
  total_private_spool_max_mib: 128
  sse_bootstrap_max_kib: 64
  control_message_max_mib: 1
  max_safe_upstream_attempts: 2
  upstream_connect_timeout_ms: 10000
  upstream_response_header_timeout_ms: 120000
  stream_idle_timeout_ms: 360000
  stream_total_timeout_ms: 0       # no artificial total cutoff; host can cancel
```

The outstanding request cap must include attempts awaiting bootstrap, not just currently emitting deltas. If queues are full, return a typed local overload immediately; do not indefinitely block critical worker sockets. No hidden speculative hedging or parallel duplicate inference.

## 9.2 Queue fairness

Use per-client FIFO queues and deterministic round-robin dequeue; one TradeAgent manager cannot starve a local interactive T3 user. A switch-pending process-family group blocks only itself, not unrelated groups. Separate scopes for global/provider/account concurrency avoid turning global outages into account exhaustion.

## 9.3 Reset and cooling behavior

Only provider-supplied `reset_at` or `Retry-After` can be advertised as known. Otherwise use an internally chosen *observation expiry/reprobe time*, labeled `retry_probe_after`, **never presented as the provider's actual reset**. Track the evidence's scope and age. A recognized exhausted app-registration should not be hammered every incoming request. Reprobe at a conservatively bounded cadence only if the operating mode and quota state allow it.

If accurate numerical per-registration telemetry is unavailable, leave `predictive_drain` off. "Prewarming" can renew a soon-expiring valid OAuth token and check local readiness; it should not burn provider inference requests to estimate capacity without explicit test authorization.

## 9.4 Wall time and sleep/wake

Store UTC provider reset instants and evidence times. Use monotonic duration for live deadlines. Mark observations suspect after significant backward wall-clock changes, sleep/wake, or offset inconsistencies; do not mark unknown accounts newly full because the laptop woke. On startup discard stale ephemeral availability assumptions, reconcile credentials and re-evaluate eligibility.

---

# 10. Security and privacy threat model

## 10.1 Threat boundaries

| Threat | Mandatory control |
|---|---|
| Unauthenticated local process | Strong local bearer; owner-only IPC; per-client authorization; no open management HTTP |
| Malicious web page targeting localhost | Strict Host validation, no CORS wildcard, no cookie auth/preflight acceptance, unpredictable port **plus bearer** |
| Local bearer stolen | Least privilege, refresh/revoke, protected command helper; acknowledge same-UID process compromise limit |
| Wrong provider identity billed | Independent issued client/token pair and workspace verification, pinned route, fail-closed billing mode |
| Cross-client or cross-project prompt exposure | Immutable authority-bound leases; pool-member intersection; negative tests |
| Arbitrary URL/SSRF or network interception | Fixed official origin, TLS checks, no redirects or untrusted proxy env, reject custom per-request endpoint |
| Prompt leakage in trace/spool | No content logging, owner-only ephemeral spool, strict deletion, encrypted storage where available and explicit risk warning |
| Native automatic retry | Force retry zero; prove by synthetic post-commit failure injection |
| Token refresh race/crash | Singleflight, durable intent, versioned secret-store record, reconcile-or-reauth |
| Compromised update / config drift | Signed artifact provenance, checksum manifests, config validation, rollback and quarantine |
| Resource exhaustion | Bounded request/body/SSE/queues/spool/memory; cancel stalled downstream |
| Prompt injection | Native agent owns tools/permissions; gateway never adds or executes tools |

## 10.2 Secret handling

Never log credentials, auth redirect hints, bearer env, prompt bodies, SSE data or opaque reasoning. No secrets in CLI args, URLs, shell history, persistent TOML/YAML, support bundles, crash diagnostics or git trees. Command-backed auth writes bearer **only to the native client's inherited stdout pipe**, never to terminal stderr, and verifies the process-family capability before issuing. The helper must have a noninteractive error path that emits no secret.

OS credential stores protect against disk theft better than raw permissions but cannot save a fully compromised owner session. Secure deletion of SSD/temp snapshots is not guaranteed; minimize spooling and use full-disk encryption where appropriate. For installations requiring stricter confidentiality, disable disk spool and reject oversized safe-replay candidates.

## 10.3 Logging, metrics and privacy

Allowed default event fields:

```text
UTC timestamp, opaque request ID, hashed client/lease label,
pool label, registration alias, provider route, requested model slug,
admission decision code, policy/version fingerprint,
scoped provider error code, HTTP status, first-byte/total latency,
bytes in/out, stream terminal category, cancellations, active slots,
credential generation (number only), stale/fresh evidence status
```

Not allowed: raw URL containing token, email/subject by default, input/output text, headers containing secrets, exact session IDs, tool arguments, encrypted content, raw tokens, full provider error bodies or stack traces exposing request content. Deep debugging requires explicit capture consent, encrypted owner-only fixture storage, retention deadline and purge.

## 10.4 Filesystem and process isolation

Secure canonical app data root; owner-only parent dirs; no symlink traversal, `O_EXCL` for socket/spool temp creation; no symlink target in world-writable temp; path independence from shell current directory. On macOS preserve TCC/sandbox considerations and avoid a privileged LaunchDaemon. `agentpool doctor` checks permissions and reports inconsistencies without exposing protected content.

---

# 11. Runtime compatibility and version admission

## 11.1 Effective runtime profile, not merely a binary hash

A Codex SHA-256 alone does not capture behavioral changes from config, ACP build, feature flags, provider model profile, compaction mode or server requirements. Use:

```yaml
runtime_profile:
  client_kind: codex-app-server
  codex_sha256: "..."
  codex_version: "..."
  acp_sha256: null
  host_kind: tradeagent
  os_arch: darwin-arm64
  launch_config_sha256: "..."   # secret-free effective normalized config
  model_slug: "..."
  route_id: openai-siwc-responses-http
  profile_schema: 1
  features:
    http_stateless_history: PASS
    native_http_retries_disabled: PASS
    native_stream_retries_disabled: PASS
    local_bearer_refresh: UNVERIFIED
    automatic_compaction_route: UNVERIFIED
    manual_compaction_route: UNVERIFIED
    opaque_context_cross_principal: UNVERIFIED
    process_family_barrier: PASS
    t3_acp_gateway_auth: UNVERIFIED
  evidence_manifest_sha256: "..."
  validated_at_utc: "..."
  expires_at_utc: "..."
```

Admission statuses: `DOCUMENTED`, `SOURCE_OBSERVED`, `MOCK_TESTED`, `LIVE_TESTED`, `POLICY_REVIEWED` are **evidence types**, not synonyms for `PASS`. For L3, `LIVE_TESTED` is mandatory. For L4, real technical proof **and** applicable operating-mode clearance are mandatory.

## 11.2 Fail-closed upgrades

- Pin Codex and ACP installation versions for critical deployments; do not follow `latest` or `@preview` implicitly.
- Signed update manifests may mark known-bad builds, but **may not remotely enable** a previously gated advanced feature.
- New binary/config/profile → re-run non-consuming local checks. L3/L4 remain off until necessary conformance passes.
- Maintain a testable compatibility matrix by full profile digest rather than scattering version-string conditionals.
- The relevant documentation revision and provider error-schema mapping must be recorded in evidence.
- Releasing a new binary requires canary, rollback, schema migration verification and a record of whether the old compatibility evidence survives.

## 11.3 Minimal provider route interface

```go
// Provider-specific transport facts, not native agent intelligence.
type UpstreamRoute interface {
    ID() RouteID
    ValidateShape(rawBody []byte, profile EffectiveProfile) (RequestMeta, error)
    Endpoint(meta RequestMeta) PinnedEndpoint
    Credential(ctx context.Context, reg RegistrationID) (SecretHandle, error)
    ClassifyEarlyFailure(status int, prefix []byte) FailureClass
    ObserveProviderSignal(signal ResponseSignal) *CapacityObservation
}
```

There is intentionally **no** generic `SwapConversation`, `SummarizeContext`, `CheckpointJob`, `RunAgent`, `ExecuteTool` or `RecreateThread` method. If another provider requires restarting a native process, use a separately scoped optional **native resume adapter**, outside the core gateway.

---

# 12. Configuration, CLI and control protocol

## 12.1 Versioned config example

The checked-in source config schema is authoritative. YAML below is a human-facing canonical example; secrets are never in config:

```yaml
schema_version: 3
server:
  bind: 127.0.0.1:0
  allow_remote: false
  control_transport: os-owner-local
security:
  local_bearer_required: true
  credential_store: platform
  upstream_host_allowlist: [api.openai.com]
  follow_redirects: false
  trust_proxy_environment: false
  log_request_content: false
limits:
  active_requests_global: 24
  active_requests_per_client: 6
  active_requests_per_registration: 3
  pending_admission_global: 64
  request_body_max_mib: 16
  total_private_spool_max_mib: 128
  max_safe_upstream_attempts: 2
pools:
  personal-codex:
    route_id: openai-siwc-responses-http
    authority_group_id: personal-dev
    billing_mode: subscription-included-only
    registrations: [acct-a, acct-b]
    weights: {acct-a: 1, acct-b: 1}
    routing: sticky-smooth-weighted-round-robin
    operating_mode: L1         # L3 and L4 require admission and clearance
    automatic_rotation: false
    predictive_drain: false
integrations:
  codex:
    pinned_binary: /absolute/path/to/codex
    credential_mechanism: auth-command
    http_retries: 0
    stream_retries: 0
    require_effective_profile_admission: true
  t3_codex_acp:
    enabled: false
```

The schema MUST reject unknown safety-sensitive settings, invalid weights, unauthorized pool members, unsupported route and billing strings, negative caps and conflicting credential methods. Read config atomically, validate entire graph, publish new policy epoch only if valid, retain last-known-good on failure. Changes to authority/policy must invalidate or revoke affected client leases rather than silently upgrade them.

## 12.2 Human-facing CLI

```bash
agentpool init
agentpool start
agentpool stop
agentpool status
agentpool doctor
agentpool doctor --deep                 # opt-in real inference/cost/plan usage

agentpool login openai                 # official SIWC browser flow
agentpool accounts                     # aliases, scoped evidence, not secrets
agentpool account disable acct-a
agentpool account enable acct-a
agentpool logout acct-a
agentpool pools
agentpool pool show personal-codex
agentpool pool create personal-codex
agentpool policy status
agentpool capability status

agentpool run codex --pool personal-codex --authority personal-dev
agentpool run codex --pool personal-codex --account acct-a
agentpool acp codex --pool personal-codex
agentpool route explain --pool personal-codex --model <slug>
agentpool diagnostics export --redacted
agentpool lease-token --context <launcher-context>  # internal helper, no terminal UI
```

`agentpool run codex` does not scrape the TUI. `agentpool acp codex` does not reinterpret ACP. User-facing commands have stable exit codes; failure outputs never print credentials.

## 12.3 JSON-RPC 2.0 methods

```text
v1/status.get
v1/health.get
v1/accounts.list
v1/accounts.beginLogin
v1/accounts.cancelLogin
v1/accounts.logout
v1/accounts.setEnabled
v1/pools.list
v1/pools.get
v1/pools.applyConfig
v1/clientLeases.issue
v1/clientLeases.renew
v1/clientLeases.revoke
v1/routing.explain
v1/capability.get
v1/policy.get
v1/doctor.run
v1/diagnostics.export
v1/daemon.shutdown
```

Example return:

```json
{
  "jsonrpc": "2.0",
  "id": "req-348",
  "result": {
    "schema_version": 1,
    "daemon": {"state": "READY", "version": "3.0.0"},
    "capabilities": {
      "codex_l1": "UNVERIFIED",
      "codex_l3": "UNVERIFIED",
      "automatic_rotation_l4": "UNVERIFIED",
      "t3_l5": "UNVERIFIED"
    },
    "pools": [{"id": "personal-codex", "eligible": 0, "unknown_capacity": 0}],
    "active_requests": 0
  }
}
```

All example statuses are placeholders, not proof. Error envelope for control plane:

```json
{
  "jsonrpc": "2.0",
  "id": "req-348",
  "error": {
    "code": -32010,
    "message": "Capability not admitted",
    "data": {
      "schema_version": 1,
      "reason": "AP_CAPABILITY_UNVERIFIED",
      "request_id": "opaque-request-id",
      "retryable": false
    }
  }
}
```

## 12.4 Typed reason codes

```text
AP_LOCAL_UNAUTHORIZED
AP_LOCAL_LEASE_EXPIRED
AP_LOCAL_LEASE_REVOKED
AP_AUTHORITY_MISMATCH
AP_POOL_POLICY_HOLD
AP_OPERATING_MODE_NOT_APPROVED
AP_CLIENT_INCOMPATIBLE
AP_UNSUPPORTED_REQUEST
AP_UNSUPPORTED_MODEL
AP_UNSUPPORTED_BILLING
AP_REGISTRATION_REAUTH_REQUIRED
AP_REGISTRATION_DISABLED
AP_QUOTA_SCOPE_LIMITED
AP_NO_AUTHORIZED_CAPACITY
AP_LOCAL_OVERLOADED
AP_SAFE_ATTEMPT_BUDGET_EXHAUSTED
AP_AMBIGUOUS_UPSTREAM_OUTCOME
AP_POSTCOMMIT_UPSTREAM_FAILURE
AP_PROVIDER_UNAVAILABLE
AP_COMPATIBILITY_STALE
AP_SECRETS_UNAVAILABLE
AP_METADATA_CORRUPT
AP_SWITCH_BARRIER_TIMED_OUT
AP_CANCELLED
```

Every routing decision must answer: which registration was chosen (alias only), why it was eligible, which specific quota observation was used, and which policy/lease version was applied. An error reason MUST NOT claim an account is globally exhausted based on one app-specific 429.

---

---

# 13. Repository and development conventions

## 13.1 Minimal Go module

```text
agentpool/
├── cmd/agentpool/main.go
├── internal/
│   ├── daemon/              # startup, lock, supervised shutdown
│   ├── control/             # local JSON-RPC 2.0, OS peer auth
│   ├── httpgateway/         # admission, body budget, SSE transport
│   ├── routing/             # pure eligibility, allocator, affinity barrier
│   ├── capacity/            # provider evidence and scope classification
│   ├── auth/siwc/           # OAuth, OIDC, PKCE, refresh state machine
│   ├── secrets/             # OS secret-store adapters
│   ├── policy/              # operating mode, authorities, config epochs
│   ├── persistence/         # SQLite DB and migrations
│   ├── compatibility/       # runtime profile and evidence admission
│   ├── launcher/            # native Codex and maintained ACP launch
│   └── observability/       # redacted metrics/diagnostics
├── schema/
│   ├── config-v3.schema.json
│   ├── control-v1.schema.json
│   ├── errors-v1.schema.json
│   └── admission-v1.schema.json
├── migrations/0001_initial.sql
├── compat/
│   ├── manifests/
│   └── known-bad/
├── testdata/
│   ├── streams/             # synthetic SSE fixtures
│   ├── oidc/                # synthetic security fixtures
│   └── runtime/             # mock client configurations
├── tests/
│   ├── contract/
│   ├── fuzz/
│   ├── fault/
│   ├── race/
│   ├── security/
│   ├── live-gated/
│   └── upgrade/
├── docs/
│   ├── RUNBOOK.md
│   ├── SECURITY.md
│   ├── PROVIDER_OPERATING_MODE.md
│   ├── EVIDENCE_REGISTER.md
│   └── COMPATIBILITY_MATRIX.md
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

This is one Go process with internal package boundaries, **not** a microservice diagram. Prefer maintained standard-library facilities; add dependencies only for SQLite, OS-secure stores, and well-reviewed OAuth/JWT validation when justified. Record each dependency's maintenance, license, security update path and signed/reproducible binary supply-chain implications. Select SQLite driver after comparing CGO, cross-compilation, platform code-signing and embedded SQLite features; do not assert a driver choice without a reproducible cross-platform build.

## 13.2 Core type discipline

- Use typed `RegistrationID`, `PoolID`, `LeaseID`, `RouteID`, `AuthorityGroupID`, `ProfileDigest`, and `RequestID` in Go; avoid positional strings that can swap clients/accounts.
- Use dedicated `SecretHandle` with narrow access; never hand a raw OAuth token to routing or logging code.
- Separate pure `EvaluateEligibility(snapshot)` from mutating `Reserve()` and transport `Dispatch()`.
- Make every resource release/stream state transition idempotent and race-testable.
- Parsing request JSON for admission **cannot** modify the bytes eventually dispatched to the provider.
- Return typed error categories, not string comparisons on arbitrary HTTP bodies, except narrow vendor-adapter parsing of documented error codes.
- One upstream request attempt owns one selected registration, one credential generation snapshot, one cancellation context and one request-lineage reference.
- All state mutations produce a monotonically increasing policy/credential/binding epoch with stale-write rejection.

## 13.3 Required documentation accompanying the code

The `.md` alone is a product/architecture implementation specification, not a compiled executable. Before first merge, create real checked-in JSON schemas, SQL migrations and test fixtures corresponding to the contracts above. They are the executable authority when examples and code diverge; any divergence must generate an architecture review and change to this spec, not silently redefine the behavior.

---

# 14. Hard acceptance gates and test campaign

Every test report records: test ID, spec invariant, immutable runtime profile digest, OS and architecture, provider route, model, sanitized test dataset, registration aliases, expected/observed events, outcome, trace artifact hash, execution timestamp, reviewer, expiration and link to the applicable operating-mode decision. Raw access tokens and real user prompts are forbidden in CI artifacts.

## 14.1 Blocking gates

| ID | Gate | Mandatory pass evidence | Blocks |
|---|---|---|---|
| G00 | Operating-mode review | Specific deployment/pattern considered; documented approved/denied/unresolved decision | L4 |
| G01 | Official SIWC registration | Issued client IDs, correct identity, scopes and workspace checks; two separate authorized registrations for cross-principal testing | L1/L3 |
| G02 | Exact custom-provider routing | Real Codex traffic reaches only approved gateway route; no ambient billed/API key fallback | L1 |
| G03 | Native retry elimination | Wire fixture proves zero second attempts after post-commit failure, parser failure and dropped SSE | L1 critical profile |
| G04 | Native compaction routes | Actual automatic/manual compaction and resumed multi-turn execution use supported route, no semantic stripping | Long-session L1, L3 |
| G05 | Same PID/thread switching | A→B→A with same PID/thread ID, exact model and tool context, no unauthorized overlap | L3 |
| G06 | Opaque reasoning/compaction | Reasoning/compaction encrypted payloads compatible in repeated real cross-registration long sessions | L3 |
| G07 | Trustworthy session grouping | Conservative process-family grouping and drain barrier prevent overlapping principal switch | L3/L4 |
| G08 | Safe rejection and scoped observation | Documented explicit no-accept rejection switches once; app-scope not promoted to global | L4 |
| G09 | Ambiguous/postcommit faults | No automatic replay, no fabricated native success, host sees actionable failure | All critical |
| G10 | OAuth lifecycle | Serialized refresh, rotated-token crash window, logout, revoked grants, wrong principal, expired local bearer | L1+ |
| G11 | Local security | Browser localhost attacks, wrong bearer, unsafe UDS, symlink, cross-project lease, SSRF all denied | All |
| G12 | Resource/fairness | Bounded memory/spool/queue/stream and no starvation/leak under load/race | Operational |
| G13 | T3/ACP | Actual T3 initialization/auth capability, stream, tools, cancel and recovery through pinned ACP | L5 |
| G14 | T3 native path comparison | Document native compatible switching and its constraints; justified adapter choice | T3 design decision |
| G15 | Upgrade/downgrade | Unknown runtime quarantines high-risk mode; rollback restores known-good profile | Operational |
| G16 | Host crash recovery | Real TradeAgent or harness resumes/pause/escalates without duplicate tool side effects | TradeAgent release |
| G17 | Long lifecycle | Simulated 30+ day credential expiry and clock faults, extended sleep/wake, real soak | Operational |
| G18 | Release attestation | Signed manifest, issue-free safety regression, explicit capability label and owner approval | Every tagged release |

**A mock-only G05/G06 test cannot certify a real cross-principal SIWC deployment.** An ordinary provider's general API capabilities do not substitute for its SIWC plan-sharing preview. If G04 fails, L1 can only be labeled a *limited short-session profile*, not durable native Codex support.

## 14.2 Deterministic SSE and transport fixtures

| ID | Injected condition | Required result |
|---|---|---|
| F01 | Canonical short success with `response.completed` | Exactly one success; exact bytes/events |
| F02 | Long SSE, heartbeats and valid tool-call deltas | Backpressure, no premature timeout or rewrite |
| F03 | HTTP 429 documented explicit no-accept rejection | Scoped observation; one alternate at most if enabled |
| F04 | App-specific 429, unknown reset | Never mark all principals globally exhausted |
| F05 | Unknown 429 | Surface, no speculative rotation |
| F06 | 503 usage visibility unavailable | Provider transient, not quota empty |
| F07 | TCP disconnect before headers | Ambiguous; no hidden retry |
| F08 | Headers sent then connection breaks before first event | Postcommit in downstream semantics; native sees failure |
| F09 | One valid SSE event, abrupt EOF | Failure, not completed; no retry |
| F10 | `response.failed` after deltas | Failure, no synthetic completion |
| F11 | `response.incomplete` | Not success |
| F12 | `response.completed` then clean EOF | Success only once |
| F13 | Malformed SSE, giant event, invalid UTF-8/JSON boundaries | Bounded and safe, not a parser panic |
| F14 | Downstream cancel before upstream acceptance | Prompt cancellation and release |
| F15 | Downstream cancel after model output | No alternate; report cancelled |
| F16 | Upstream ignores cancellation or half-closes | Bound resources and accurately report uncertainty |
| F17 | Native auto HTTP retry on gateway 503 | Zero native retries in admitted profile |
| F18 | Native SSE retry after partial `response.completed` parser issue | Zero native retries in admitted profile |
| F19 | Two same-lease concurrent streams then switch | Wait barrier or typed timeout, never account overlap |
| F20 | Switch CAS from stale goroutine | Older generation loses, new principal preserved |
| F21 | Cross-project lease to forbidden registration | Denied before any upstream byte |
| F22 | Pool edit while request is reserved | Recheck/epoch fence and correct deny/cancel behavior |
| F23 | Logout while streaming | Defined drain/cancel semantics and no new admission |
| F24 | OAuth refresh succeeds then process killed before save | Unknown state quarantined; no stale replay |
| F25 | SQLite secret generation mismatch on restart | Quarantine or safe reconcile only |
| F26 | Disk full, locked Keychain, corrupted WAL | No fabricated valid auth state |
| F27 | Fake local web origin, DNS rebinding, CORS/preflight | No unauthorized access |
| F28 | Randomized response chunks split inside SSE boundaries | Exactly correct stream semantics |
| F29 | Invalid request `previous_response_id` | Reject, never strip and forward |
| F30 | Unknown SIWC field / hosted tool | Explicit incompatibility, no model change |
| F31 | Native auto-compaction or `/compact` | Observed endpoint/feature matches certified route |
| F32 | A→B with opaque encrypted compaction item | Live pass or mark capability unsupported |
| F33 | Client bearer renewal during long native process | No native interruption or accurate limited-mode label |
| F34 | Heterogeneous eligible pool (different models/workspaces) | Wrong candidate filtered, never silent downgrade |
| F35 | Multiple independent workers under provider 5xx | Circuit breaker stops retry/account spray |
| F36 | Clock backwards, sleep beyond reset hint | Freshness invalidated; no invented provider reset |
| F37 | Daemon crash during active SSE | Native/host knows failure; nothing marked complete |
| F38 | Upstream headers attempt cookie/billing-origin injection | Header allowlist and no credential leakage |

## 14.3 Property and fuzz suites

**Properties:**

```text
P01: selected registration is in approved pool for current immutable policy snapshot
P02: a request's selected principal never changes after upstream acceptance
P03: no request is forwarded with a model/billing/authority outside lease scope
P04: revoked or disabled accounts never receive new dispatch after revocation fence
P05: one failed app-registration limit never becomes a global limit without evidence
P06: one input request cannot have parallel authoritative provider attempts
P07: any stream with no terminal completed event is not reported successful
P08: stale binding CAS cannot overwrite newer binding generation
P09: every admitted request releases its concurrency reservation exactly once
P10: unsafe/unknown auth, policy and client compatibility always fail closed
P11: a provider's reset estimate is never synthesized from unrelated evidence
P12: successful internal retry requires exact documented no-accept rejection
P13: token rotation does not mix issued_client_id, sub, workspace or generations
P14: process-family switch barrier covers all supported subagent inference requests
P15: native/host retry policy cannot duplicate a committed inference unnoticed
P16: config reload cannot broaden an already minted authority lease
```

Fuzz JSON request admission, SSE/event framing, response error code parser, OAuth callback `state`/PKCE validation, OIDC ID token claims, config schema, JSON-RPC envelopes, UTF-8 boundaries, header rewriting and SQLite migration recoverability. Run `go test -race ./...` for supported race-capable environments plus deterministic simulated scheduling and high-contention stress. A fuzz suite does not replace live/native protocol testing.

## 14.4 Release scorecard

**Targets** (not observed results):

| Measure | Release requirement |
|---|---|
| Wrong provider account, workspace, model or billing path | **Zero** in exhaustive applicable tests; any finding blocks release |
| Hidden inference replay after commitment | **Zero** across gateway, native client and test host |
| Unauthorized operating mode or data crossing | **Zero** |
| Secrets or private prompt/reasoning in default logs | **Zero** by automated scan |
| False native success on incomplete SSE | **Zero** |
| Concurrency cap violations / leaked reservations | **Zero** in fault/race suites |
| Service availability | Publish measured rate/failure classes, do not invent SLA |
| Rotation continuation success | Publish measured per exact admitted client/model and failure class |
| Successful capacity rescue rate | Report numerator/denominator for correctly classified eligible alternate attempts |
| Median/p95/p99 route overhead | Measure against pinned direct single-account baseline |
| Native tool/action duplication | **Zero** in verified host fault harness; real-world exactly-once not promised |
| Provider reset correctness | No fabricated times, unknown stays unknown |
| Token refresh reliability | Crash-window tested; uncertain cases fail closed |
| Resource budgets | No exceeded configured caps under specified stress envelope |

A percentage target cannot excuse a correctness-invariant failure. One-day soak is insufficient for refresh rotations, prolonged unattended workers and new Codex build changes; use accelerated lifecycle tests plus an extended real soak.

---

# 15. Implementation sequence: decisions, owners, outputs

**Do not begin with a giant CLI or a generic provider plugin framework.** Build the critical vertical slice, prove the seam, then harden it into a daemon.

| Phase | Scope / deliverables | Exit condition |
|---|---|---|
| **0 — Product/permission register** | Intended paid-account routing defined as in §0; scope, account authority, exact provider operating-mode record | Approval/denial/unresolved recorded honestly; no automatic mode implied |
| **1 — Thin Codex harness** | Pinned Codex + mock Responses HTTP, retry-zero test, compaction route capture, T3 native/ACP comparison | G02–G04 mock/native parts pass; unsupported endpoint risk understood |
| **2 — Pure kernel** | Typed eligibility, immutable snapshots, allocator, scoped observations, binding barrier, SSE state machine and bounded spool | Race/property/fault suite passes with fake provider |
| **3 — SIWC + secure storage** | Official browser login, identities/grants, Keychain, refresh-intent reconciliation, logout | One real single-account model request; G01/G10/G11 |
| **4 — True L1 Codex** | Tested gateway launcher, full tool turn, automatic/manual compaction and failure reporting | G02–G04, G09 complete; no charged fallback |
| **5 — L2 + multiple registrations** | Independent registrations, explicit manual account choice, authority and model intersection | Two approved accounts used separately with accurate identity and billing-intent evidence |
| **6 — L3 feasibility** | Safe same-process A→B→A experiment with long context, subagents, encrypted compaction, concurrency and valid authority | G05/G06/G07 live pass; otherwise L3 off |
| **7 — L4 automatic routing** | Only after positive operating-mode decision and L3 pass: scoped limit detection, bounded safe alternate selection, no-capacity status | G00/G08/G09 pass in approved profile |
| **8 — T3** | If needed, thin `codex-acp` launcher, actual T3 capability negotiation; compare native account path | G13/G14 plus account attribution |
| **9 — Reliability and release** | Signed binary, rollback, process supervision, fault/soak, docs, status, telemetry, manifest | All release gates pass; publish actual capabilities only |
| **10 — Evidence-driven expansion** | Other Codex clients, independently permitted provider routes, optional native resume integration | New adapter-specific decision record and proof suite |

Parallelize only independent work: kernel mocks, documentation, threat model and policy research can proceed together. OAuth live tests and policy-specific automation are individually gated. New provider research does not become a dependency of the initial Codex gateway.

## 15.1 PR definition of done

A pull request changing protected code must include its new/updated test case, relevant invariant ID, exact negative-path behavior, migration/revert path if needed, resource/backpressure impact if transport-related, permission/authority impact if security-related, and compatibility changes if native-dependent. A reviewer should be able to answer **why a retry is safe** without reading Codex's conversation history.

## 15.2 Named engineering ownership

| Area | Reviewer responsibility |
|---|---|
| Architecture | Preserve simple gateway boundary and ADR/invariant register |
| Security/auth | OAuth, OIDC, Keychain, lease trust, threat model, identity attribution |
| Protocol/transport | Request admission, SSE correctness, hidden replay detection |
| Routing | Scoped capacity, fairness, barriers, atomic reservation |
| Integrations | Pinned Codex/App Server/ACP/T3 conformance and native retry controls |
| Provider/terms | Precise, neutral operating-mode evidence, permitted usage, billing path |
| Quality/release | Fuzz/race/fault, migration, signing, compatibility and runbooks |

Roles may share one engineer but no review responsibility disappears.

---

# 16. Operations: failure runbook and observability

## 16.1 Operator decision table

| Incident | AgentPool | Native agent / host |
|---|---|---|
| Registration reaches an applicable usage window limit | Record provider's exact scoped evidence; cease eligible dispatch while it applies | Keep task/thread; next request may use approved B if L4 certified, else typed wait |
| App-specific usage cap; plan still usable elsewhere | Do not mark the whole plan globally empty | Display actual app-limited state, no invented universal reset |
| Unknown 429 or unavailable usage backend | Treat as unknown/transient, not fuel empty | Backoff/retry only through native/host safe policy |
| All eligible registrations limited/unavailable | Return `AP_NO_AUTHORIZED_CAPACITY` | Host pauses durable task or informs user; do not keep HTTP open for hours |
| Token near expiry | Serialized upstream refresh | Usually no client disruption with valid local bearer |
| Local credential helper cannot renew | Fail local auth, no ambient fallback | Relaunch/resume only via tested path |
| Revoked/invalid grant | Quarantine registration; require correct reauthorization | Host may use other approved registration only when mode/authority allow |
| Wrong work authority/account membership | Deny and record reason | Correct pool/lease explicitly; no forced request reroute |
| Unsupported compaction request | Mark runtime feature incompatible | Use tested native method or host pause; never strip context silently |
| Opaque history rejected after switch | Quarantine L3 for profile | Host-managed native resume only if separately certified |
| Stream fails after commit | No alternate inference | Host reconciles existing tool effects before any replay |
| Provider outage | Route-level circuit breaker/backoff | Pause/wait; do not cycle accounts rapidly |
| Daemon dies | All active streams fail; safe restart and re-admit | Host owns task durability / semantic reconciliation |
| SQLite corrupt | Fail closed, preserve diagnostic snapshot | Restore verified backup; never invent registrations |
| Secret store unavailable | Deny inference for affected registrations | User unlocks store; no plaintext downgrade |
| Clock jump/sleep | Mark capacity timers suspect and reevaluate | Native jobs may retry through host after wake |
| Malicious local request | Reject, redacted security signal | No effect on provider or native worker |
| New unverified native build | L3/L4 downgraded/disabled | Canary and recertify; continue only verified lower capability |
| Terms or provider access requirement changes | Reevaluate precise operating mode; feature-specific gate | Clearly report policy/profile hold |

## 16.2 TradeAgent contract

TradeAgent owns worker launch, worktrees, JEV-driven context hygiene, manager/subagent lifetime, external tool effects, deployments/trades and durable task checkpoints. The gateway may provide **availability and request attribution**, not task progress or worker semantic state.

A recommended integration API **inside TradeAgent**, not AgentPool:

```text
launch_native_codex_worker(pool_ref, authority_group, model_policy, worktree)
observe_worker_events_and_exit()
query_agentpool_status(pool_ref)
on_capacity_unavailable -> PAUSE_JOB/REQUEUE_AT_SAFE_BOUNDARY
on_ambiguous_inference -> RECONCILE_TOOL_EFFECTS_BEFORE_RESUME
on_native_client_incompatible -> REPORT/USE_CERTIFIED_PROFILE
```

For trading-related tooling, reissuing a model request after uncertain native tool side effects must be treated as high risk. A model inference gateway cannot guarantee exactly-once execution of an order, payment, deployment or other irreversible action. No hidden host turn replay.

## 16.3 Status surface

The status command must include: local daemon health, registered aliases and their independent credential health, applied authority group, selected model route, capacity evidence with **scope and age**, exact feature admission states, currently active requests, pending barriers, source of known reset time, and reasons for no-capacity. Do not show "all exhausted" where evidence says only one app's registration hit a limit. Provide the provider's official usage-management location, not manufactured subscription token counts.

## 16.4 Rollback and backup

Before schema migration, snapshot SQLite through its backup API and record schema version/checksum; stop active admission, reconcile rotating token intents, then migrate transactionally. Rollback must not blindly restore **old refresh tokens** or overwritten security grants. DB restoration can safely restore metadata only if secret-generation matching is revalidated. When mismatch cannot be reconciled, quarantine and reauthorize.

No hot binary update during active sensitive sessions without a tested drain plan. A rollback restores pinned native/ACP/runtime compatibility along with daemon code, config and evidence manifests. It must not silently enable an older unreviewed operating mode.

---

# 17. Explicit release invariants

| ID | Normative invariant |
|---|---|
| I01 | Each upstream inference request uses exactly one independently authorized registered principal; provider enforces that principal's own limits. |
| I02 | Rotation does not reset, forge, inflate, transfer or modify per-account entitlement/usage limits. |
| I03 | Technical paid-registration selection is not categorically mislabeled an exploit; provider-specific approval remains independently reviewable. |
| I04 | A mode without verified applicable operating permission or necessary conformance never silently activates. |
| I05 | Requested model, provider route, workspace/authority and billing intent remain unchanged unless explicitly authorized through a separate request. |
| I06 | No ambient API credit/key/overage fallback or private provider endpoint. |
| I07 | One downstream inference can never have parallel authoritative attempts. |
| I08 | Automatic alternate attempts require documented definite no-accept rejection and exact permitted mode. |
| I09 | Neither AgentPool nor admitted native Codex silently retries a committed inference stream. |
| I10 | Unknown outcome is never represented as definite rejection or successful completion. |
| I11 | Only actual `response.completed` terminal evidence permits upstream-completed classification. |
| I12 | Capacity evidence is scoped, sourced, timestamped and TTL-limited; no invented quota/reset. |
| I13 | Unknown capacity is distinct from unknown identity, consent, billing or client compatibility. |
| I14 | Same workload authority must cover every selected registration; leases cannot be broadened by config drift. |
| I15 | Encrypted context, compaction, tools, instructions and outputs are never semantically rewritten by gateway. |
| I16 | Actual native compaction route and effect must be proven for long-running support. |
| I17 | Same-session/account boundary requires the binding/stream drain barrier; no A/B overlap for unproven profile. |
| I18 | Stale binding, credential or policy generation cannot overwrite a newer state. |
| I19 | OAuth rotations are serialized and crash uncertainty is quarantined, not disguised as fresh auth. |
| I20 | Local bearer and upstream credentials are distinct; no credential/PII/prompt body in default telemetry. |
| I21 | Outbound HTTPS is pinned to the documented official route; client cannot redirect egress. |
| I22 | Request, spool, queue, stream and concurrency resource budgets remain bounded in fault tests. |
| I23 | User cancellation releases resources and never starts a replacement inference. |
| I24 | Daemon crash cannot preserve an HTTP stream or claim native work completed. |
| I25 | TradeAgent/T3/native agent retain task, tool, workspace, UI, context and side-effect ownership. |
| I26 | Known-unknown version or profile changes disable unproven high-risk capability. |
| I27 | Adding a legitimately registered account is data/configuration work, not a new code path. |
| I28 | No unsupported consumer Antigravity connection or other undocumented provider protocol is implemented. |
| I29 | A local lease bound to a long-running client remains valid/reissuable only under its exact existing authority and revocation rules. |
| I30 | A release's capability labels correspond to dated, inspectable evidence, never synthetic or wishful results. |

**Every invariant maps to at least one deterministic test or a documented external approval gate.** For I03, use a specification/communications review test that confirms descriptions distinguish *independent paid-account rotation* from actual documented abuse. I04's permission decision cannot be "unit-tested" into approval.

---

# 18. Deferred choices: maintain the small design

| Future need | Preferred first step | Do not prematurely do |
|---|---|---|
| Alternative native coding client | Test official custom Responses transport and versioned auth | Universal CLI emulator |
| Claude subscription account selection | Review sanctioned CLI/auth/config capabilities and native thread semantics | Pretend it is SIWC/OpenAI-compatible |
| Gemini/Google models | Use documented API or authorized enterprise route, independent billing/permissions | Route unauthorized consumer Antigravity session through a proxy |
| T3 account management | Evaluate documented native compatible-account switching | Build a redundant ACP fork without evidence |
| Native process restart required | Thin documented provider-specific resume launcher with host-owned durability | Reintroduce old durable broker/job/transcript database |
| Remote or multi-user hosting | Separate product, threat model, deployment and permission decision | Expose local bearer listener to LAN/public Internet |
| Forecasting quota | Add only authoritative, registration-scoped provider telemetry | Guess remaining plan percentage from 429s |
| Future higher availability | Measure daemon fault/recovery and simplify single-host supervision first | Kubernetes/message bus/distributed locks in v1 |

**Scope discipline rule:** if a feature requires AgentPool to inspect a task transcript, change a native model's reasoning, own tools, or replay an arbitrary coding step, it belongs in the native client/host or a separately approved integration — not inside the Switch's core.

---

# 19. Source register and evidence hierarchy

The research below is source evidence for **documented building blocks as of 2026-10-08**, not proof that an actual two-account AgentPool build was tested. Each source's revision, publication/check date and material constraints must be captured in release evidence. Official documentation is authoritative for supported interfaces; a public repository is version-specific implementation evidence; third-party issues are only diagnostic signals. Policy-specific approval is recorded separately.

| ID | Primary source | Implementer use |
|---|---|---|
| [R01] | [OpenAI SIWC overview](https://developers.openai.com/siwc/token-sharing-open-source) | Local OSS grant context and official path |
| [R02] | [SIWC accounts and sessions](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions) | Multiple saved OAuth registrations, correct refresh, app usage and plan limits |
| [R03] | [SIWC registration/sign-in](https://developers.openai.com/siwc/token-sharing-open-source/sign-in) | Initial dynamic client, issued client ID, host ID, OIDC/PKCE/nonce |
| [R04] | [SIWC token reference](https://developers.openai.com/siwc/token-sharing-open-source/token-reference) | One-hour access tokens, replacement 30-day refresh, claims |
| [R05] | [SIWC preview limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations) | `store:false`/`stream:true`, stateless HTTP, unsupported fields/tools/routes |
| [R06] | [SIWC Codex App Server](https://developers.openai.com/siwc/token-sharing-open-source/codex-app-server) | Custom provider, native `clientInfo`, documented restart/resume on env-token renewal |
| [R07] | [SIWC errors and recovery](https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery) | Scoped quota and usage-unavailable distinctions, revocation/error types |
| [R08] | [Codex configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference) | Custom-provider `auth.command`, retry counts, WebSocket flag, env-key distinction |
| [R09] | [Maintained codex-acp](https://github.com/agentclientprotocol/codex-acp) | Conditional custom gateway auth and native App Server adapter |
| [R10] | [T3 Code Codex account documentation](https://github.com/pingdotgg/t3code/blob/main/docs/user/providers-codex.md) | Compatible account switching, shared/shadow home and alternative integration |
| [R11] | [EU Terms of Use](https://openai.com/en-GB/policies/eu-terms-of-use/) | Applicable service restriction review, **not** an automatic characterization of this architecture |
| [R12] | [OpenAI Responses compaction guide](https://developers.openai.com/api/docs/guides/compaction) | Distinct general-API compaction modes, **not** implied SIWC entitlement |
| [R13] | [SIWC models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference) | Official `POST /v1/responses`, model catalog and terminal streamed events |

Internal architecture lineage: `AgentPool_The_Switch_Architecture_v2(1).md`, its audit findings, earlier Switch and historical broker specification. This v3 document preserves the small Switch boundary and deliberately replaces v2's unresolved native retry and bearer-lifecycle abstractions with explicit implementation and proof contracts.

---

# 20. Final decision — what the team is authorized to build

The implementation team is authorized to build a **small, owner-authenticated, policy-aware local gateway** that uses documented provider inference registrations, independently tracks each account's capacity evidence, and routes compatible native Codex requests at safe inference boundaries. It is a deterministic transport service, not a new agent runtime.

The team MUST preserve the meaning of legitimate paid-registration rotation from §0: **when A is limited, respect A's limit; select authorized B for a new request when permitted and technically compatible; do not tamper with either account's quota, identity, model or billing path.** They also MUST keep provider-specific operating-mode permission distinct from that engineering description and release only capabilities for which proof exists.

**Immediate engineering mandate:** run the exact Codex transport/compaction/retry/credential harness first. Implement the mock routing kernel and secure SIWC single-account path next. Certify live A→B→A before enabling same-thread switching. Enable automatic capacity-triggered selection only after its technical and applicable provider operating-mode gates pass. Compare native T3 account continuity before writing an unnecessary bridge. Keep TradeAgent in full ownership of task durability and side effects.

**Success criterion:** From the native agent's perspective, ordinary completed inference turns continue correctly through the same authenticated gateway and, for certified configurations only, subsequent turns can use another independently eligible account without native history rewriting. If that cannot be demonstrated, the Switch stops safely and explains exactly which constraint prevented it. **Never purchase apparent continuity by sacrificing correctness.**
