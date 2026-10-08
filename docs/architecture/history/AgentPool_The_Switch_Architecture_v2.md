# AgentPool — The Switch
## Architecture v2.0 · Team implementation specification

> **A small, local, policy-aware gateway that connects native AI agents to explicitly authorized inference capacity without becoming an agent orchestrator.**
>
> **Keep the agent. Keep its state. Choose an eligible upstream identity only at a proven-safe request boundary. Never invent continuity, billing certainty, or permission.**

| Document control | Value |
|---|---|
| Decision | **Adopt the Switch as the product architecture; gate cross-account automation before release** |
| Status | **Implementation-ready core design; critical provider-dependent capabilities explicitly unproven** |
| Version | 2.0.0 · 2026-10-08 |
| Primary client | Codex CLI / Codex App Server (version-pinned) |
| Secondary client | T3 Code through maintained `codex-acp`, subject to conformance |
| Primary deployment | Single-user local host, macOS first; Linux and Windows via platform adapters |
| Implementation | One Go module, one shipped native executable, embedded SQLite, OS credential store |
| Predecessors | `AgentPool_The_Switch_Architecture(1).md`; `AgentPool_Architecture_Final(2).md` |
| Review input | `AgentPool_Switch_vs_Broker_Critical_Architecture_Audit.md` |
| Research cut | Public documentation/repositories checked 2026-10-08; behavior can change |

**Important meaning of “approved.”** The architecture and work breakdown below are approved *for implementation and verification*. This is **not** a claim that automatic cross-subscription rotation is already technically demonstrated, contractually permitted, or production-ready. Those are independent release gates. No authenticated two-account experiment was conducted while preparing this specification.

---

# 0. Decision in one page

## 0.1 What we are building

AgentPool is a **local, authenticated inference gateway** with a small control plane. It maintains independent user-authorized provider registrations, selects an eligible registration for each *compatible* model request, forwards that request to a provider's **documented** endpoint, streams the response back, and records only the minimal routing/capacity information required to make the next decision.

Its intended invisible user experience is:

```text
Human / TradeAgent / T3 Code
             |
             v
       Native Codex agent
       ├── thread + history
       ├── reasoning + compaction
       ├── tools / MCP / subagents
       ├── permissions + sandbox
       └── workspace + task recovery
             |
             | independently routable Responses request
             v
    +----------------------------+
    |  AgentPool — THE SWITCH    |
    |                            |
    |  caller authentication     |
    |  account eligibility       |
    |  sticky policy routing     |
    |  OAuth renewal             |
    |  exact-stream forwarding   |
    |  narrow failure handling   |
    +----------------------------+
             |
       eligible, authorized
       registration A / B / C
             |
             v
    Official provider endpoint
```

The gateway sees **inference requests**, not “what the agent is trying to accomplish.” It never needs to know that TradeAgent has managers, JEV, worktrees, checkpoints, trading verification, or a hierarchy of coding workers.

## 0.2 The boundary that makes the product small

**AgentPool owns**: OAuth registrations and protected secret references; local caller identity; eligible pools and routing policy; request concurrency; scoped capacity observations; upstream authentication; streaming transport; minimal compatibility records; operational diagnostics.

**Native agent owns**: conversation/thread state; history/context/reasoning; compaction; tools; subagents; permissions; UI; native recovery and native request construction.

**Host (TradeAgent / T3 / script) owns**: long-running job durability; workspace isolation; orchestration; cancellation policy for tasks; semantic retries; reconciliation of external actions; overall service continuity when a native agent crashes.

**Provider owns**: the service's permitted auth/billing semantics, entitlements, model behavior, actual capacity, error meanings, and service policies.

## 0.3 The eight decisions

| ID | Decision | Rationale |
|---|---|---|
| D1 | **Transport-first, Codex-first** | Rotation below the agent is simpler *when tested compatible*. |
| D2 | **One native binary, not a distributed platform** | No message broker, cluster scheduler, transcript DB, job executor, or agent framework. |
| D3 | **Official public Responses route only** | No private ChatGPT backend, browser scraping, cookie harvesting, or undocumented endpoint. |
| D4 | **Account identity and billing are invariants** | Never silently use a different principal, metered API key, workspace, model or credit path. |
| D5 | **No unproven transparent replay** | After downstream commitment, forward failure; do not rerun inference in secret. Even before commitment, retry only a proven rejection. |
| D6 | **Feature admission is proven per exact client/provider combination** | A working custom `base_url` alone does not prove live cross-principal continuity. |
| D7 | **Provider rules are an operational prerequisite** | Multiple registered accounts do not by themselves authorize exhaustion-driven pooling. |
| D8 | **Small fallback, not the old broker** | If the seam fails, retain the account/router core and add only a provider-specific native resume adapter if justified. |

## 0.4 The central reality check

The supported building blocks are real: OpenAI documents independently saved Sign in with ChatGPT (SIWC) registrations, OAuth-backed subscription-plan inference, the public Responses endpoint, and a configurable Codex App Server Responses model provider. **The composition “one live Codex process switches paid registrations every time one is exhausted” is not directly guaranteed by those documents.** The official App Server example actually restarts and resumes after renewal of an environment-supplied credential. AgentPool proposes a separately authenticated **local gateway** to avoid that restart; this must be demonstrated in the actual Codex configuration and with real sessions. [S01–S07]

A second independent blocker exists: OpenAI's EU Terms prohibit circumvention of rate limits or restrictions. SIWC account-switching instructions describe account management and user selection, **not an explicit waiver for automated exhaustion-triggered rotation**. The proposed operating mode requires a documented authorization/policy decision before activation. [S02, S11]

**If either gate fails, do not fake success.** The project still offers secure single-registration gateway operation, observability, standardized launching, and a platform for future expressly supported capacity sources.

---

# 1. Scope and release modes

## 1.1 What v1 does

1. Register an eligible user-authorized SIWC account and verify the exact registration/permission, independent of Codex's ordinary sign-in files.
2. Establish a per-user local gateway and control socket; securely attach a *local* bearer from a supported agent launcher.
3. Forward supported, stateless, streaming Responses calls using the selected account's current OAuth token.
4. Route deterministically among configured eligible registrations **only in operating modes expressly authorized by the provider and owner**, and only after cross-account technical acceptance gates pass.
5. Preserve exact requested model, request content, tool semantics, reasoning artifacts and the native agent process/thread.
6. Distinguish account-specific denial from model-specific denial, app-specific usage cap, provider overload, network failure, unsupported request, unknown condition and auth revocation.
7. Emit actionable diagnostics without logging prompts, raw Responses bodies, access tokens or refresh tokens.
8. Allow Codex CLI and a thin T3/ACP launcher after individually proven integration tests.

## 1.2 Explicitly excluded

- Reimplementing Codex/Claude/Antigravity or inventing an LLM agent runtime.
- Editing/translating/summarizing native agent context, reasoning, encrypted history, tool calls, permissions or workspace state.
- Silent model substitution, prompt simplification, response sanitization, reasoning stripping or use of a different billing provider.
- Automatic API-key/credit overage when a subscription allowance is unavailable.
- Routing consumer Antigravity login through unauthorized third-party software.
- Distributing one person's consumer credentials to other people, accepting public Internet traffic, or selling a shared subscription service.
- Claiming exactly-once provider inference, network uptime, uninterrupted TCP streams, durable AgentPool task execution, or quota prediction where no provider telemetry exists.
- Shipping runtime hooks for every speculative provider now.

## 1.3 Capability levels (never conflate them)

| Level | Name | Can promise | Required proof |
|---|---|---|---|
| L0 | **Local gateway** | Authorized client can reach authenticated gateway + compatible single-provider upstream | Mock protocol/security suite |
| L1 | **Single-account Codex** | One authenticated registration serves a real Codex session | Live SIWC + Codex tests |
| L2 | **Account choice between jobs** | User explicitly selects one approved registration before launch; no live rotation | Identity/billing checks and user authorization |
| L3 | **Live request-boundary switching** | Same running Codex process/thread accepts requests using a newly chosen registration | Two-account session + reasoning/compaction suite |
| L4 | **Automatic capacity-triggered rotation** | Eligible registration chosen transparently on specific quota signals | L3 + provider permission + reliable failure classification |
| L5 | **T3 integration** | T3 + maintained ACP adapter use verified Codex gateway mode | ACP handshake + end-to-end T3 suite |
| L6 | **Other providers** | A specific provider has equivalent verified semantics | Its own policy + adapter certification |

Capabilities are **per exact installation and account policy**, not blanket marketing claims. A v1 release may support L1–L2 without L3–L4; it must disclose that honestly.

## 1.4 Deployment classes

**Personal/local OSS** is the first supported scope: machine-bound operation for the authenticating user, with authorization for each local client and registration.

**Paid SaaS, remotely hosted, multi-user, enterprise/team credential pooling** are *separate* scope and permission decisions. OpenAI's current SIWC OSS overview directs paid or remotely hosted application proponents to an interest process; merely copying a local registration to a server is not blanket permission for a multi-tenant product. [S01, S09]

---

# 2. Architecture: one process, two interfaces

```text
     agentpool run codex             agentpool acp codex
              |                             |
      Codex CLI / App Server      codex-acp -> Codex App Server
              |                             |
              +--------------+--------------+
                             |
              localhost, client-authenticated
                             |
       +---------------------v---------------------+
       |              agentpoold                   |
       |                                           |
       |  DATA PLANE (HTTP / streaming Responses)  |
       |  ingress -> validate -> route -> forward  |
       |  exact event stream + cancellation        |
       |                                           |
       |  CONTROL PLANE (UDS / named pipe)         |
       |  login, logout, accounts, policy, doctor  |
       |                                           |
       |  SMALL CORE                               |
       |  policy / identity / health / selection   |
       |  OAuth / secret store / SQLite            |
       +---------------------+---------------------+
                             |
                    HTTP(S) /v1/responses
                             |
                    Provider's public API
```

One Go executable has command/daemon subcommands and can self-launch a daemon. `agentpool acp codex` is **a launcher**, not a reimplementation of ACP. `agentpool run codex` invokes the real Codex executable; it does not scrape or wrap the terminal UI.

### Data plane

- Bind to a dynamically selected, persistent-for-process loopback TCP port. Default IPv4 `127.0.0.1` only; support `[::1]` only with deliberate dual-stack controls; reject `0.0.0.0` and non-loopback listeners in v1.
- `POST /v1/responses` is the only v1 inference route. Add model list only when a certified client needs it, and do not imply access from an unverified catalog.
- Require a strong **AgentPool-issued local bearer** on every request. The bearer belongs to the local host client, *not* to the provider.
- Preserve request semantics; replace only the upstream authorization and allowed hop-by-hop transport metadata. Restrict the upstream URL to the provider's pinned, official allowlist; never accept arbitrary per-request URL override.
- Use outbound TLS validation and a bounded HTTP client with connection pooling and timeout controls. Do not let caller-supplied `Host`, proxy variables, TLS overrides, or forwarding headers select an untrusted destination.
- Treat the complete request as sensitive. No transparent body logging.

### Control plane

- Unix domain socket (`0600` parent-owned directory and socket) on macOS/Linux; per-user Windows named pipe with explicit ACL.
- Versioned JSON messages with bounded sizes, strict access control, schema validation and no remote exposure. Small JSON-RPC 2.0 is acceptable; **choose one encoding** before implementation and document errors.
- Commands include `status`, `login`, `logout`, `accounts`, `pools`, `pool`, `policy`, `doctor`, `routes`, `shutdown` and client-lease minting for trusted launchers.
- One local daemon elected via an atomic host lock. Avoid two writers/token refreshers. A lock is a concurrency primitive, not a proof of a live healthy process; stale-lock detection must be safe.

### Resource diagram

```text
1 daemon
  + N account registrations (data, no fixed cap)
  + M local agent clients (bounded active admission)
  + K active upstream requests (bounded)
  + 1 SQLite file (no raw tokens)
  + OS secret store (per-registration credentials)
  + optional short-lived request spools (0600, encrypted disk where available)
```

Capacity and fairness must apply to **requests**, not to imagined durable AgentPool jobs.

---

# 3. The actual Codex/SIWC transport contract

## 3.1 Supported SIWC facts as of this version

Official SIWC documentation describes:

- Each registration has its own issued `client_id`, validated user identity, and credentials; one email may appear in multiple registrations. An `ext_agent_host_id` identifies the running host and must remain stable for that host. [S01–S03]
- The intended scope for ChatGPT-plan inference is `chatgpt.tokens.use.direct`; its presence must be verified **after** OAuth exchange. Authentication alone does not imply plan permission. [S03, S07]
- Authorized access tokens can call `POST https://api.openai.com/v1/responses` with `Authorization: Bearer ...`. **Do not** use private ChatGPT `backend-api` routes. [S04]
- The current preview requires **`store: false` and `stream: true`**. For HTTP, each request supplies its needed history in `input`; it cannot rely on `previous_response_id`. Unsupported parameters and some hosted capabilities must be rejected/avoided. WebSocket continuation is a different connection-bound model. [S05]
- Codex App Server supports a custom Responses provider with `base_url`, `env_key`, `wire_api="responses"`, `requires_openai_auth=false` and `supports_websockets=false`. The official example injects an access token directly and restarts/resumes App Server when renewing. [S06]
- SIWC quota errors may be **app-specific** and do not prove global plan exhaustion. [S07]

## 3.2 Requested integration: two distinct authentication domains

**Native Codex** authenticates only to *AgentPool* using a short-lived local bearer. **AgentPool** authenticates to the provider using the selected registration's SIWC access token.

Conceptual Codex configuration, to be generated and tested for each admitted Codex version (not blindly pasted into arbitrary user dotfiles):

```toml
model_provider = "agentpool"

[model_providers.agentpool]
name = "AgentPool local Responses gateway"
base_url = "http://127.0.0.1:<SELECTED_LOCAL_PORT>/v1"
wire_api = "responses"
env_key = "AGENTPOOL_LOCAL_BEARER"
requires_openai_auth = false
supports_websockets = false
```

The launcher supplies `AGENTPOOL_LOCAL_BEARER` **through the child environment only**. No bearer in `config.toml`, CLI args, launch history, URLs, shell trace, or logs. The exact config/flag precedence must pass an automated test. If the version ignores the requested custom provider because a ChatGPT login takes precedence, **mark that runtime incompatible** rather than guessing the route (Codex has had public reports of such behavior [S15]).

Request path:

```text
Codex Authorization: Bearer <local opaque gateway bearer>
      |
      v
AgentPool validates caller scope + pool + bound client
      |
      | discard local Authorization
      | select approved, eligible registration
      | construct new upstream Authorization only
      v
OpenAI Authorization: Bearer <selected SIWC access token>
```

No token mixing: `issued_client_id(A)` must never be stored alongside `refresh_token(B)`. No ambient `OPENAI_API_KEY`, `CODEX_API_KEY`, custom base URL, inherited `HTTP_PROXY` or mismatched `CODEX_HOME` may silently change the intended path. Note: a local bearer in an environment variable is still a secret; reduce its lifetime/scope, sanitize child environment, and document same-user-process limitations.

## 3.3 Validate the *actual* request, not merely the provider configuration

A Codex version is admitted only if the observed requests through this route are compatible with current SIWC HTTP requirements. At minimum:

| Constraint | Gateway requirement |
|---|---|
| Endpoint and method | Exactly allowed Responses HTTP method/path |
| Streaming | `stream: true`; valid upstream SSE framing and terminal events |
| Storage | `store: false` |
| Stateless history | Request contains required `input` context; no `previous_response_id` |
| No unsupported fields | Detect and reject invalid shape **without silently dropping semantic content** |
| Model | Requested model slug used unchanged; account/model access must be observed, not inferred |
| Tools | Provider-native/local tool invocation passes integration test; unsupported hosted tools fail explicitly |
| Reasoning | Encrypted reasoning and compaction items carried byte-for-byte through gateway |
| WebSocket | Explicitly disabled for this HTTP-only v1 path |
| Unknown extension | Pass through only if understood to be safe/allowed by endpoint; otherwise mark compatibility unknown |

No “helpful” gateway normalization of `input`, `reasoning`, `tools`, `instructions`, content arrays, or SSE event payloads. Reject an incompatible client *before issuing an upstream request* when possible. When a new field's support is uncertain, a version admission check must decide; avoid scanning/re-writing arbitrarily nested payloads just to force a result.

## 3.4 Four different kinds of continuity

1. **Codex local state:** Same process/thread/files/tools remain alive. Controlled primarily by Codex and host.
2. **Server-side Responses state:** SIWC HTTP `store:false` requires resending history; do not assume remote state follows credentials. [S05]
3. **Opaque reasoning/compaction portability:** Possibly model/provider/account-dependent; no generic guarantee between identities. A reported model/deployment switch triggered `invalid_encrypted_content`; that report is not proof of an account-switch failure, but demands cross-account testing. [S14]
4. **Provider caching:** Prompt-cache locality or billing/cost behavior may change when auth identity changes. Never promise unchanged cache hits, token use or latency without measurement.

**L3 cannot ship** until a *single native process and one thread* complete A→B→A tests including long context, tool use, subagents, compaction, encrypted reasoning and cancellation with validated account attribution and no semantic modification. A short chat with two successful HTTP calls is not sufficient.

---
# 4. Permission, identity, billing and data-authority gates

**Permission is a runtime condition, not a paragraph in the README.** A user consented to an OAuth client; that does not necessarily authorize every pooling behavior the developer imagines.

## 4.1 Provider permission gate — P0

The provider's current terms, product rules and SIWC scope must permit the **specific proposed deployment and routing pattern**. Current OpenAI EU Terms prohibit circumvention of service rate limits/restrictions; the technical ability to maintain multiple account registrations does not settle whether exhaustion-triggered rotation is permitted. [S11]

The team must document, with accountable legal/product review and ideally explicit provider clarification:

- Is the project a qualifying **locally hosted open-source client** under the current SIWC flow? If not, obtain the appropriate alternative agreement/approval.
- Are independently authorized registrations allowed to be selected **per request** in a running client?
- Is **automatic selection of another paid account when an applicable rate window is exhausted** permitted under the applicable plan, user, workspace, and product policies?
- Are account/project/workspace controls respected? Are third-party credentials and shared logins forbidden for this use case?
- Are there app-specific limits and refresh/usage restrictions that cannot be interpreted as transferable capacity?
- What conditions would revoke permission; how do we disable the feature promptly?

Until the answer is affirmative for the actual operating context:

```yaml
policy:
  automatic_cross_principal_rotation: disabled
  reason: pending_provider_permission
```

This flag is **not overrideable by a casual `--force` or a config edit** in a released build. Development can use fake providers and controlled approved test accounts, but cannot represent an unresolved compliance risk as a completed feature.

## 4.2 User authority and tenant/data boundary

All accounts in a pool must be independently authorized by their owners, usable for the contemplated content, and belong to one explicitly defined **data-authority group**. Never infer that two accounts with the same email or on the same machine have permission to receive identical code, trade information, documents or workspace prompts.

Policy attaches to each registration:

```yaml
account_authority:
  expected_subject: "verified OIDC sub"
  issued_client_id: "oaiapp_..."
  workspace_binding: "registration-bound workspace or none"
  owner_approval_id: "local consent record"
  authorized_data_domains: ["personal-dev"]
  permitted_clients: ["local-codex", "tradeagent-dev"]
  billing_mode: "subscription-included-only"
```

A pool is legal for a client only if **all** selected accounts meet the same data-authority/permissions requirements. Do not use a registration owned by another person, company or workspace merely because it is logged in locally.

## 4.3 Billing policy — never implicit fallback

V1 default:

```text
subscription-included-only
```

Forbid without a separately specified, auditable feature:

```text
ChatGPT-plan OAuth -> API-key metering
included usage    -> purchased credits/overage
account A          -> ambient account B
personal user      -> unrelated enterprise workspace
model X            -> model Y
provider X         -> provider Y
```

The gateway must separately verify **identity**, **granted inference scope** and **supported billing intent**. Where the provider does not expose authoritative per-request billing confirmation, report **billing evidence unavailable** rather than “verified no charges.” Users must still be able to inspect provider-side usage/limits. Do not make an absolute zero-charge promise that the API does not let you verify.

## 4.4 Disabled providers

As currently documented, Google's Antigravity consumer terms specifically restrict third-party access to Antigravity. Consequently, **consumer Antigravity account rotation is disabled**, regardless of possible ACP technical integration. Supported API/enterprise Gemini offerings may have a separate provider adapter only after explicit permission and billing rules are defined. [S12]

Claude Code subscriptions do not automatically gain a Codex-compatible gateway seam. `CLAUDE_CONFIG_DIR`, headless mode, account isolation and subscription billing need separate current-version verification. Claude is a **future provider investigation**, not a v1 switch promise.

---

# 5. Account registration and credential lifecycle

## 5.1 Data model

A registration is not just an email or a key file:

```text
Registration {
  id                       stable local UUID/label
  provider                 e.g. openai_siwc
  issued_client_id         per registration, never dynamic_agent_client after issuance
  validated_subject        OIDC sub, never raw email as key
  validated_workspace      registration-bound identity if applicable
  host_id                  stable ext_agent_host_id for current host
  authorized_scope_set     includes chatgpt.tokens.use.direct for plan inference
  auth_secret_ref          OS secret-store reference; not token contents
  last_verified_at         timestamp
  token_expiry_at          timestamp
  credential_generation    monotonic
  enabled                  user preference
  policy_state             approved / hold / revoked / unsupported
  data_authority_group     explicit authorization mapping
}
```

Store user-readable email/name as **optional display hints only**. A registration's issued `client_id` and validated identity remain linked even if two registrations show the same email. [S02]

## 5.2 Sign-in state machine

```text
UNREGISTERED
    |
    | durable host identifier + state/nonce/PKCE
    v
AUTHORIZING
    | callback + exact state + issued registration client_id
    v
EXCHANGING
    | code exchanged using issued client_id + code verifier
    v
VALIDATING
    | verify JWKS signature, issuer, audience, expiration, nonce,
    | validated subject, workspace constraints and granted scopes
    +---- missing plan permission --> SIGNED_IN_PLAN_DISABLED
    +---- mismatch / invalid ------> REJECTED (do not replace saved creds)
    v
ACTIVE
    | serialized refresh --> REFRESHING --> ACTIVE
    | revocation/expiry --> REAUTH_REQUIRED
    | user logout       --> REVOKING --> SIGNED_OUT
```

**Fresh authorization requires** OAuth `state`, OIDC `nonce`, PKCE verifier/challenge, exact redirect URI and CSRF-safe callback handling. First-time registration uses `client_id=dynamic_agent_client` only as the entry point and stores the **issued** registration ID for subsequent exchange and sign-in. Host ID is stable across daemon restarts. Consent requests scope correctly and do not infer permission from successful login. [S01, S03]

For returning sign-in, verify newly issued `sub` is the original registration identity **before** replacing the stored tokens; preserve its `client_id`, registration workspace and owner authorization. Sign-out should attempt documented remote revocation and then clear local credentials while reporting whether remote revocation was confirmed. [S02]

## 5.3 Refresh rules

OpenAI's current token reference describes one-hour access tokens, 30-day renewable refresh tokens and **replacement refresh tokens** after successful refresh. Every account has an independent `singleflight` refresh lock. [S08]

1. Refresh early with bounded clock-skew allowance, but only when needed; do not assume all accounts refresh on the same schedule.
2. Request token refresh using the registration's **issued** client ID and current refresh token, with the documented resource parameter and no accidental extra scopes. [S02]
3. Validate refresh result; replace the **entire access/refresh token pair as one secret-store record**, with a monotonic credential generation and durable storage sync. The SQLite generation and OS secret-store update **cannot be assumed to form one cross-store atomic transaction**: use an explicit write/reconcile protocol, with the secret record's generation and a versioned metadata commit; on mismatch quarantine the registration until state is safely reconciled or reauthorization is obtained.
4. If another goroutine is already refreshing generation *g*, wait for its result rather than trying the consumed rotating token again.
5. If refresh receives `invalid_grant` or credentials are revoked, mark registration `REAUTH_REQUIRED`. Do not guess that another token file or account can substitute.
6. On crash between provider refresh success and durable storage, the returned new refresh token may be lost; **detect the uncertain state, stop using the stale token, request reauthorization**. Atomic local writes limit damage but cannot make a remote token rotation transaction exactly once.
7. Never emit token contents to diagnostics or trace events, including failed OAuth redirects.

## 5.4 Secret store

Use an internal `SecretStore` interface implemented by macOS Keychain, Windows supported per-user protection, Linux Secret Service if available, and a *documented explicit* owner-only fallback for headless machines. If an encrypted fallback depends on a local key, specify where that key is protected; file mode `0600` alone is not encryption. No automatic plaintext downgrade.

Persist public metadata and secret **references** in SQLite. Protect database path, temporary spool directory and backups from other users. Support secure logout even if daemon crashes mid-request; revoke future admission immediately, then cancel or drain active requests according to the defined policy.

## 5.5 Principal verification is not a single boolean

Track separate evidence:

```text
identity:       verified OIDC sub + issued client_id + workspace mapping
scopes:         includes direct ChatGPT-plan permission
billing_intent: subscription-only (configured)
billing_proof:  provider confirmation if any, otherwise unavailable
health:         refreshable / reauth-required / blocked
model_access:   observed successful request for exact model/profile
```

Never label `login successful` as `included compute available`, and never label `model/list` as proof of entitlement.

---

# 6. The small persistent model

Only five primary functional objects:

| Object | Purpose | Stored? |
|---|---|---|
| `Registration` | Identity, secret reference, owner and billing permission | Yes; secrets external |
| `Pool` | Named set of authorized registrations + policy | Yes |
| `Binding` | Optional caller/session stickiness, generation and TTL | Yes or ephemeral; not correctness-critical |
| `CapacityObservation` | Scoped provider evidence, confidence, expiry/reset | Yes, bounded, invalidated on restart when stale |
| `RequestAttempt` | In-flight transport lifecycle and selected account | In memory only |

Supporting *small* records: `CapabilityAdmission`, `PolicyDecision`, `ClientLease`, `SchemaMigration`. **No durable AgentPool `Job`, `Turn`, `NativeSession`, `WorkspaceLease`, agent transcript or semantic `Checkpoint` table.**

## 6.1 Pool contract

```yaml
pool:
  id: codex-personal
  provider_route: openai-siwc-responses
  authority_group: personal-dev
  registrations: [acct-a, acct-b]
  requested_billing: subscription-included-only
  selection: sticky-weighted-round-robin
  cross_principal_switch: disabled   # on until legal + conformance gates pass
  default_client_permissions: [infer]
```

Every pool version is validated on atomic reload. Reject references to deleted or unauthorized registrations; missing/unknown policies are errors, not permissive defaults.

## 6.2 Binding contract

```text
binding_key = HMAC(local_secret, client_lease_id || caller_session_hint)
binding = { pool_id, account_id, generation, expires_at }
```

Hash to avoid storing revealing session IDs. If no stable opaque session hint is available through a supported Codex header, **L1 single-account requests may route independently, but L3/L4 live-session switching must not be certified** until session attribution (for example, a trusted launcher-scoped identity) has passed conformance. Do not scrape unstable internal Codex metadata. Session hint must be minted or authorized by AgentPool, not trusted as an arbitrary attacker-selected identity. A binding is an **affinity hint**, not a promise of remote cache continuity; a separately defined short switch barrier protects concurrency.

Update binding using compare-and-swap (CAS) on `generation`; an older routing goroutine cannot overwrite a newer switch. Explicit disable/logout invalidates bindings referencing that registration.

## 6.3 Capacity observations are scoped facts

```yaml
observation:
  registration_id: acct-a
  provider: openai_siwc
  scope:
    kind: app_registration | principal | model | limit_bucket | provider | unknown
    key: "opaque-scoped-identifier-or-none"
  event: available | limited | unavailable | unknown
  error_code: subscription_sharing_usage_limit_exceeded
  reset_at_utc: null
  observed_at_utc: "..."
  valid_until_utc: "..."
  evidence: structured_response
  confidence: authoritative_for_this_request
```

**No generic `account.exhausted = true` derived from one HTTP 429**. The explicit SIWC `subscription_sharing_usage_limit_exceeded` signal says a relevant use path hit a limit; its documentation warns an *app-specific* limit may be responsible. It neither proves a global plan limit nor gives a reset instant. [S07]

Use `unknown` when there is no trustworthy view. Retain explicit observation source, scope, timestamp and freshness; recompute eligibility for the next request rather than persisting a simplistic cooldown flag indefinitely. Provider-specific scopes must not be projected onto a broader scope. Native Codex `account/rateLimits/read` is **not** assumed to expose per-registration SIWC capacity behind this gateway.

## 6.4 State is orthogonal, not one giant enum

Each registration maintains independent dimensions:

```text
auth         = VALID | REFRESHING | REAUTH_REQUIRED | REVOKED | UNKNOWN
policy       = APPROVED | HOLD | DENIED
runtime      = VERIFIED | UNKNOWN | INCOMPATIBLE
capacity     = AVAILABLE | UNKNOWN | LIMITED(scope, evidence, until?)
allocation   = IDLE | ACTIVE(n) | DRAINING
admin        = ENABLED | DISABLED
```

Route eligibility is a pure predicate over these inputs. `READY`, `DRAINING`, `COOLDOWN`, `DISABLED` may be shown as **derived UI labels**, but not used as the underlying source of truth; this avoids conflating a revoked token with a depleted model quota.

## 6.5 Minimal SQLite schema (logical DDL)

The implementation may split tables/indexes further, but must retain constraints. Illustrative, not a ready-to-run complete migration:

```sql
CREATE TABLE registrations (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  client_id TEXT NOT NULL,
  subject TEXT NOT NULL,
  workspace_id TEXT,
  host_id TEXT NOT NULL,
  secret_ref TEXT NOT NULL,
  granted_scopes_json TEXT NOT NULL,
  authority_group TEXT NOT NULL,
  policy_state TEXT NOT NULL,
  billing_mode TEXT NOT NULL,
  credential_generation INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  last_verified_at TEXT,
  UNIQUE(provider, client_id, host_id)
);
CREATE TABLE pools (
  id TEXT PRIMARY KEY,
  route TEXT NOT NULL,
  authority_group TEXT NOT NULL,
  policy_json TEXT NOT NULL,
  version INTEGER NOT NULL
);
CREATE TABLE pool_members (
  pool_id TEXT NOT NULL REFERENCES pools(id),
  registration_id TEXT NOT NULL REFERENCES registrations(id),
  weight INTEGER NOT NULL CHECK(weight > 0),
  PRIMARY KEY (pool_id, registration_id)
);
CREATE TABLE bindings (
  pool_id TEXT NOT NULL,
  opaque_key TEXT NOT NULL,
  registration_id TEXT NOT NULL,
  generation INTEGER NOT NULL,
  expires_at_utc TEXT NOT NULL,
  PRIMARY KEY (pool_id, opaque_key)
);
CREATE TABLE observations (
  id TEXT PRIMARY KEY,
  registration_id TEXT,
  scope_json TEXT NOT NULL,
  code TEXT,
  observed_at_utc TEXT NOT NULL,
  valid_until_utc TEXT,
  reset_at_utc TEXT,
  evidence TEXT NOT NULL
);
CREATE TABLE capability_admissions (
  client_fingerprint TEXT NOT NULL,
  provider_route TEXT NOT NULL,
  feature TEXT NOT NULL,
  status TEXT NOT NULL,
  evidence_ref TEXT,
  checked_at_utc TEXT NOT NULL,
  PRIMARY KEY (client_fingerprint, provider_route, feature)
);
```

Enable foreign keys, restrict paths, set `busy_timeout`, use WAL where supported and appropriate, transactionally migrate, include schema checks and recovery/backup procedure. Retain only bounded historical observations; do not journal prompt bodies. Backups of SQLite **do not** contain raw secrets; backups of secret store are out of scope and must be handled deliberately.

---

# 7. Routing, fairness and scheduling

## 7.1 Eligibility function

For a request `r`, client `c`, registration `a`, policy version `p`:

```text
eligible(a, r, c, p) =
    c has valid local gateway lease permitting requested pool/model
AND a is in pool and enabled
AND provider permission is approved for selected operating mode
AND data-authority group matches both request and account
AND a's OIDC principal/workspace/scopes are valid
AND included-only billing path is supported (no fallback)
AND client's Codex transport/request feature profile is admitted
AND exact requested model is permitted or not known to be disallowed
AND a is not authoritatively limited for the request's specific scope
AND a has concurrency headroom and the provider route is healthy enough
```

`UNKNOWN` quota is normally **eligible for a bounded attempt**, not presumed full capacity. Authentication `UNKNOWN`, policy `HOLD` or client compatibility `UNKNOWN` **are not** eligible: these represent trust/safety gaps, not capacity ambiguity.

## 7.2 Deterministic selection

1. Select an existing sticky bound registration if it remains eligible and not draining for this request.
2. Otherwise filter all policy-approved members by scope and concurrency.
3. Prefer READY members; fallback to those with valid auth and unknown capacity if policy permits an attempted request.
4. Among equivalent eligible candidates, use **stable weighted round-robin**. Do not use random races, ML prediction or complex “best account” scoring.
5. Reserve an in-flight concurrency token, bind with CAS generation and record a request ID. If reservation fails due to a concurrent change, re-evaluate; never oversubscribe blindly.
6. Release the token exactly once on all exit paths; requests cancelled after routing cannot leak capacity tokens.

Weights adjust distribution across accounts, **not** identity/billing guarantees. Separate capacity admission for known provider constraints and local load-shedding protects the service from self-inflicted saturation.

## 7.3 Prewarm and drain are optional optimizations

`PREWARM`: validate selected next registration/refresh if needed, **without generating billable inferences** just to probe. Avoid background refresh storms.

`DRAIN`: cease assigning *new* requests on a narrowly scoped, trustworthy near-limit observation; let existing requests finish. Never terminate an active stream just because a percentage crossed a threshold.

The current SIWC docs do not guarantee a real-time numerical usage snapshot for independent registrations. Accordingly:

```text
percent-based predictive drain = OFF unless per-registration telemetry is proven
```

When reliable telemetry exists, threshold settings must remain configured policy and scope-aware rather than pretending “95% used” universally means five percent of all plan capacity remains.

## 7.4 Per-session switching barrier

When a known session/process is bound to registration A, switching its binding to B **must not happen while a prior A request for that same binding is still active** unless a provider/agent compatibility test expressly proves this safe. The initial design uses a simple **drain barrier**:

```text
binding S generation 12 -> A; outstanding(A,S)=2
mark S SWITCH_PENDING; no new A requests admitted for S
wait bounded time for both A requests to finish/fail
if A request cannot safely finish -> block or fail S (no secret replay)
CAS generation 12 -> 13; bind S -> B
only then admit S's next B request
```

This is *not* workspace management or a durable job lease. It is a short-lived in-memory ordering guard for a logical caller/session. Other independent sessions and accounts continue normally. On daemon crash the barrier and active streams disappear; the host receives transport failure, and the gateway does not infer that old work committed. If the session cannot be distinguished reliably from other clients, **disable L3/L4**, rather than claiming a safe same-thread switch on an unstable hint.

The barrier may reduce throughput in pathological concurrent-agent cases; measure and document it rather than weakening correctness invisibly. Subagents get their own trusted identity or inherit a conservative parent process binding as established by conformance, never an untested assumption.

## 7.5 Concurrency and fairness

- Default conservative per-registration active request cap, configurable only within verified provider limits.
- Global maximum active requests, maximum queued admission, per-client max concurrency, memory budget, body size and total spool disk budget.
- Queue admission locally for short bounded periods only; no hours-long hanging inference requests. Reject overflow promptly with a structured error and suggested host-level retry policy.
- Never use hidden parallel hedging/speculative duplicate inference in v1. Precisely one authoritative upstream attempt at any instant.
- Fairness across clients: small bounded per-client admission quotas and a FIFO or simple weighted-fair queue; do not allow one TradeAgent process to starve all interactive clients.
- Client-side native Codex retries and gateway retries can multiply; identify request lineage where supported and cap the combined effective budget through conservative gateway defaults.

## 7.6 Sleep, shutdown and clocks

Persist reset instants (when actually supplied) as absolute UTC. Evaluate elapsed time on waking or daemon restart. Use monotonic elapsed time for *live deadlines* only, not persisted quota resets. Handle backward system-clock changes by marking timing suspect and seeking fresh observations rather than assuming reset. Apply bounded jitter when many requests become eligible simultaneously.

Graceful shutdown: close new admission; cancel or bounded-drain active streams; close secret handles; flush metadata; exit with correct status. Forced crash: active HTTP streams fail; native agents/hosts own recovery. Persisted bindings never imply that an in-flight request survived a crash.

---
# 8. Streaming and retry safety: the gateway's hardest invariant

> **Do not confuse “no bytes delivered to Codex” with “the provider did no work.”** The former concerns the client's observable state; the latter may be unknowable. Automatic retry is allowed only in an explicitly recognized safe rejection case under the selected policy.

## 8.1 RequestAttempt lifecycle

```text
RECEIVED
   |
   v
VALIDATED  --(unsafe shape/policy/auth)--> REJECTED_LOCAL
   |
   v
ADMITTED   --(no approved capacity)----> NO_CAPACITY_LOCAL
   |
   v
SENT_UPSTREAM
   |
   +-- explicit rejection, no accepted inference or downstream output
   |       | classify scope; policy authorizes switch?
   |       +--> RETRYABLE_REJECTION --> NEW ATTEMPT (bounded)
   |
   +-- unknown upstream outcome / timeout before event
   |       +--> AMBIGUOUS_FAILURE --> SURFACE_ERROR, no blind rotation
   |
   v
BOOTSTRAP_OBSERVED
   | validate framing, status, first event and any structured failure
   v
DOWNSTREAM_COMMITTED
   | exactly one selected upstream attempt is now authoritative
   | stream faithfully, preserve backpressure
   +-- complete ----------------------> COMPLETED
   +-- provider failure -------------> FAILED_AFTER_COMMIT
   +-- client cancel/disconnect ------> CANCELLED
```

`DOWNSTREAM_COMMITTED` means the first downstream observable upstream response information (HTTP response status/headers or response body bytes) has been sent. Treat the entire response, not only the first text token, as observable. Before that boundary, buffer enough to classify and decide, bounded by explicit limits. Once committed, **no transparent attempt B**.

For an early structured provider failure *within* an SSE connection, classification depends on the exact event, whether inference was accepted, and whether any output is already committed. Do not assume a `response.failed` SSE event is an HTTP rejection or safe retry. Some failures can arrive after upstream work begins. [S05, S07]

## 8.2 Safe and unsafe examples

**Explicit rejection before accepted work (eligible for a bounded policy-approved retry):**

```text
Codex -> gateway -> registration A
A replies with recognized account-specific quota rejection,
   no accepted inference, no downstream response committed
Gateway closes A attempt, reserves eligible B,
   sends exact same request body (authorized mode only)
B returns valid stream -> forwarded to Codex
```

**Ambiguous network loss before first event (do not quietly retry):**

```text
Gateway sent request to A
TCP connection breaks before provider response
Did A already accept/execute? UNKNOWN
=> return an explicit ambiguous failure; do not silently duplicate it
```

**Mid-stream failure (never hidden retry):**

```text
Provider A begins an SSE stream
Gateway commits events to Codex
Provider A stream fails or returns terminal response.failed
=> convey failure without manufacturing response.completed
=> host/native agent owns any restart/reconciliation
```

**Concurrency / disconnect:** A retryable bootstrap decision must first cancel/close the old upstream request, wait for its local transport termination (bounded), and only then dispatch B. Do not race two upstream streams for the same incoming request. A client disconnect cancels current upstream context and releases leases, but cannot guarantee provider computation already accepted is undone.

## 8.3 Four explicit classifications

| Class | Evidence | Decision |
|---|---|---|
| `EXPLICIT_NO_ACCEPT_REJECTION` | Recognized structured failure indicating request was rejected before inference acceptance | Eligible for one bounded retry if allowed |
| `AMBIGUOUS_PRECOMMIT_FAILURE` | Timeout, EOF, disconnect or unknown error after request was sent | **Never silently replay**; send actionable error |
| `POSTCOMMIT_FAILURE` | Any failure after forwarding response status/headers/events | **Never silently replay**; preserve failure semantics |
| `LOCAL_NONRETRYABLE` | Invalid request, permissions, client version, billing or policy | Reject locally, no provider attempt |

Strict policy favors a visible error over an uncertain duplicate. Retry configuration must not transform an unrecognized 401/403/429/503 into an assumed exhausted account or transient retry.

## 8.4 Bounded bootstrap algorithm

```go
// Pseudocode only: captures safety contract, not a full HTTP implementation.
func ServeResponses(ctx context.Context, request Incoming) Result {
    validated := ValidateIncoming(request) // permission, shape, body cap
    if !validated.OK { return Reject(validated.Error) }

    snapshot := FreezeRequestBody(request) // memory or secure bounded spool
    defer snapshot.CloseAndDestroy()

    for attemptNo := 0; attemptNo < MaxSafeAttempts; attemptNo++ {
        slot, err := ReserveEligible(request) // policy and concurrency checked
        if err != nil { return NoCapacity(err) }

        handle := BeginReservedAttempt(ctx, slot, snapshot)
        outcome := handle.InspectBootstrap() // handle owns reservation/stream

        switch outcome.Kind {
        case ExplicitRejection:
            ObserveScopedFailure(slot, outcome)
            handle.CloseAndReleaseOnce() // close A before any attempt B
            if AllowedSafeRetry(outcome, request, attemptNo) {
                continue
            }
            return ForwardOrSurface(outcome)
        case ValidStream:
            CommitExactlyOnce(outcome.HeadersAndBootstrap)
            result := handle.StreamRemainingWithBackpressure(ctx) // no retry
            handle.CloseAndReleaseOnce() // after streaming really ends
            return result
        case AmbiguousFailure:
            handle.CloseAndReleaseOnce()
            return SurfaceAmbiguousFailure(outcome) // no retry
        default:
            handle.CloseAndReleaseOnce()
            return SurfaceFailure(outcome)
        }
    }
    return SurfaceBoundedRetryExhausted()
}
```

Implementation requirements:

- `MaxSafeAttempts` is small and includes the first attempt; no infinite cycling.
- Preserve the original request bytes as much as protocol permits. For JSON validation, parse within bounded limits; avoid reserializing/normalizing unknown reasoning/compaction content.
- Retain request body in memory only below threshold; for larger allowed payloads, spool to an owner-only temporary file and delete it on **every** exit path. Cap total disk budget; optionally reject bodies too large for configured safe replay.
- Distinguish *request transport reuse* from *upstream inference duplication*. Local body availability alone does not prove replay safety.
- Bound upstream connect/handshake/first-event/idle/total times separately. Long valid model silence may be normal; do not misclassify it as quota.
- Forward SSE framing, metadata and terminal events according to the exact supported Responses contract; ensure flushing and cancellation semantics match Codex expectations.
- Avoid decompression/recompression transformations unless a certified compatibility implementation requires them; remove invalid hop-by-hop headers and recalculate content length/transfer encoding correctly.
- Do not emit a synthetic successful completion on EOF, abrupt socket close or `response.failed`. Verify that the exact native Codex version interprets post-commit failures as failures.
- Wire downstream cancellation to upstream cancellation promptly; measure cancellation races and acknowledge that provider-side compute may already be running.

## 8.5 Error taxonomy and intended actions

| Event | Classification | Account state / action | Retry? |
|---|---|---|---|
| `subscription_sharing_usage_limit_exceeded` 429 | App-registration usage path limited; scope may be narrower than whole plan | Preserve actual source/scope; no invented reset | Only explicit rejection + approved rotation mode |
| `subscription_sharing_usage_unavailable` 503 | Usage visibility/service unavailable | Provider/transient observation, not exhausted | Bounded later retry, not account-exhaustion cycling |
| `subscription_sharing_user_not_eligible` 403 | Selected user/workspace/policy blocked | Disable affected permission path; explain | No immediate reroute without independent valid authorization |
| `subscription_sharing_unsupported_capability` 400 | Request feature mismatch | Mark client/feature incompatible, inspect `error.param` | Never replay same body unchanged to another account |
| `subscription_sharing_route_not_supported` 403 | Invalid endpoint/routing path | Mark adapter incompatible | No |
| OAuth `invalid_grant` | Session invalid or revoked | `REAUTH_REQUIRED` | After explicit reauthorization |
| Unknown 401/403/429 | Unclassified auth/permission/throttle | Preserve unknown; surface or bounded policy-specific handling | No speculative capacity switch |
| Provider 500/502/503/network outage | Provider/transport health, not proven account depletion | Backoff/breaker by upstream route | Do not spray accounts |
| SSE terminal `response.failed` | Failure may be after provider acceptance | Preserve provider failure and commit state | No after commit |
| Malformed SSE/JSON | Protocol incompatible or upstream defect | Fail request, consider client/provider quarantine | No semantic rewrite |
| Local 401/403 | Invalid or unauthorized local client bearer | Reject before upstream | No |
| Every eligible account limited | Local `NO_CAPACITY` with evidence/freshness | Host waits or surfaces status | Not a long-held HTTP request |

**No fabricated reset clock.** An unavailable reset timestamp is shown as `unknown`; the gateway can perform cautious later admission when evidence expires, subject to a bounded attempt budget. `Retry-After` is forwarded or used only when the provider actually supplied it for the relevant scope.

## 8.6 Request-level error envelope

For a local failure *before any upstream response is committed*, use a stable error JSON envelope and appropriate HTTP status. Example is a **proposed AgentPool extension**, not an OpenAI response schema:

```json
{
  "error": {
    "code": "agentpool_no_capacity",
    "message": "No approved, eligible registration is currently available for this request.",
    "retryable": true,
    "next_known_reset_at": null,
    "request_id": "redacted-opaque-id"
  }
}
```

The Codex adapter must be tested to ensure such local errors are understandable. **After commitment**, AgentPool cannot safely replace the HTTP status; propagate the actual stream failure/closure and rely on tested native-agent handling. An optional status endpoint can expose diagnosis correlated by request ID without changing streaming protocol content.

---

# 9. Security model and threat boundaries

## 9.1 Trusted and untrusted actors

Even on localhost, other processes, browser tabs, compromised packages and adjacent users may attempt to consume the user's registered inference capacity. **Loopback is a network boundary, not an authentication policy.**

| Threat | Control |
|---|---|
| Malicious local web page (CSRF / DNS rebinding) | Require unguessable bearer; verify `Host`/origin context as appropriate; no CORS wildcard; no cookie-based admin; reject browser preflights by default |
| Another local process sends unauthenticated requests | Bearer mandatory; per-client scopes; owner-only IPC for lease issuance |
| Credential exposure in logs/CLI | Structured redaction; never print tokens; no credentials in argv/querystrings |
| Ambient API key redirects billing | Construct sanitized child environment; exact custom provider; no implicit credential fallback |
| Malicious upstream URL or proxy redirection | Fixed TLS origin allowlist, constrained redirects (ideally none), proxy environment disabled unless explicit trusted enterprise configuration |
| Secret-store or metadata file copied | Platform credential store/owner controls; avoid plaintext fallback; threat-model same-user process compromise honestly |
| Stolen local bearer | Short-lived narrow scope, rotation on launcher exit, explicit revocation; local port not Internet exposed |
| Cross-pool data leakage | Pool authority group, client-to-pool ACL, explicit account authorization |
| Host update changes request route | Exact-version admission tests, canary and deny on known-bad build |
| DoS via huge body/slow client/SSE stalls | Body, connection, concurrency, disk, queue and deadline budgets |
| Prompt injection in model output | Native agent owns approval/tool sandbox; AgentPool never elevates tool permissions |

## 9.2 Local bearer lifecycle

A trusted local launcher requests a scoped client lease over owner-authenticated IPC:

```text
lease_id, pool_ids, allowed_modes, expiry, max_concurrency, random_secret
```

Return the bearer only over the protected local IPC response. The launcher sets it in the child's environment; the gateway stores only a secret hash/MAC-equivalent verifier where possible. Bind the local bearer to allowed pools and client identity. Support revocation, time-bounded leases and renewal **without** killing an active native agent when appropriate; if Codex reads `env_key` only once on launch, plan for local bearer long enough for a session or a controlled compatible refresh mechanism. Do not promise rotation of local bearer without testing the actual client.

On Linux and macOS, same-UID malware may still steal process environment or instrument the owner. This is a **single-user trust model**, not a defense against a fully compromised user account. For multi-user machines require distinct operating-system accounts and isolated home directories.

## 9.3 Host and filesystem hardening

- Deterministic canonical paths and owner-only directories, no persisted shell `~` or working-directory-dependent path resolution.
- Create sockets and temp files using secure atomic primitives (`O_EXCL` / equivalent); defend against symlink attacks and unintended world-writable parents.
- OS-aware process signal/cancellation semantics; do not leave orphan child processes with token-bearing environments.
- Set restrictive file permissions before writing secrets. Include backup, upgrade, rollback, corruption and disk-full tests.
- Hash or redact opaque session keys in telemetry. Apply retention limits to source status errors as those may include upstream request metadata.
- Do not dump goroutines, process environment or core dumps containing tokens in ordinary support output; offer sanitized support bundle separately.
- Enforce outbound endpoint pinning at HTTP layer and TLS certificate validation; never disable TLS verification as a workaround.

## 9.4 Privacy and observability

Prompts/code, reasoning payloads, model output, tool parameters, access tokens and ID tokens are **excluded from logs by default**. Diagnostic information may include:

```text
UTC time, local request ID, hashed client ID, pool ID,
registration label (no token), model slug, client/runtime fingerprint,
route decision reason, scoped failure code, status, latency,
bytes in/out, active requests, cancellation result
```

Any trace capture containing private request/response data requires an explicit user opt-in with time-bound retention and an obvious purge operation; it is **never necessary for normal operation**. Deep conformance tests prefer synthetic prompts and sanitized raw event traces.

---

# 10. Provider and runtime compatibility

## 10.1 Design goal

A small gateway can still be upgrade-proof if it refuses to rely on unverified version behavior. A large dynamic plugin system is not necessary. What matters is a **small tested capability registry**.

```yaml
capability_admission:
  client_kind: codex
  executable_sha256: "hash-of-exact-binary"
  client_version: "captured-at-runtime"
  provider_route: openai-siwc-responses-http
  source_protocol: responses-sse
  features:
    custom_base_url: verified
    local_env_bearer: verified
    websocket_disabled: verified
    stateless_history: verified
    encrypted_history_cross_principal: unverified
    compaction_cross_principal: unverified
    t3_acp_custom_gateway: unverified
  verified_with: "test-manifest-id"
  verified_at: "UTC timestamp"
  known_bad: false
```

Every protected capability has evidence strength: `DOCUMENTED`, `SOURCE_OBSERVED`, `MOCK_TESTED`, `LIVE_TESTED`, `POLICY_APPROVED`. Use the strongest evidence required by that capability: a UI display string can use docs; **cross-account live continuation requires live tests**, and **provider permission requires an appropriate policy decision**.

## 10.2 Admission algorithm

```text
1. hash/detect exact Codex executable and adapter build
2. check policy and provider-route compatibility record
3. known-bad fingerprint?                  -> BLOCK
4. changed/unknown executable?             -> non-consuming local conformance
5. incompatible source request shape?      -> BLOCK or degrade feature
6. valid local identity/auth?              -> admit L0/L1 as proven
7. live cross-account certificate present? -> only then consider L3
8. policy clearance + L3 + scoped error?   -> only then enable L4
9. ACP/T3 evidence present?                -> only then admit L5
```

A known-good **version range** is advisory; a particular binary fingerprint and local probe give stronger confidence. Signed compatibility fixture updates can add known-bad manifests, but no remote update may silently enable previously disabled sensitive features.

## 10.3 Upgrade handling

- Pin a known-good Codex/codex-acp version for release images or documented installs; don't follow `latest` silently on critical infrastructure.
- On unknown new binary, preserve configuration and registrations but disable advanced switching until conformance returns success.
- Keep model/provider protocol checks in one `compat` package; avoid scattered `if version` logic across router and OAuth code.
- Maintain a redacted replay fixture corpus for different SSE events, errors, encrypted context placeholders and interrupted streams.
- Before a version upgrade, run canary sessions; rollback is a supported deployment workflow.
- Treat a provider change to scope or Terms as a policy-event admission change, not merely a software bug.

## 10.4 Minimal provider adapter contract

Do not implement a universal agent adapter. The core data-plane contract should express provider-specific *transport* facts only:

```go
// Illustrative stable-domain shape; error and secret types are opaque.
type UpstreamRoute interface {
    Name() string
    ValidateRequestShape(body []byte, profile CapabilityProfile) error
    SelectEndpoint(request RequestMeta) (PinnedEndpoint, error)
    GetAuthorization(ctx context.Context, reg RegistrationID) (SecretHandle, error)
    ClassifyFailure(response UpstreamFailure) FailureClassification
    ReadCapacity(ctx context.Context, reg RegistrationID) ([]CapacityObservation, error)
}
```

The gateway owns HTTP request lifecycle, local authentication, concurrency, body reuse, SSE stream and admission. The route owns exact provider endpoint, legal auth mode, shape restrictions, structured error semantics and supported capacity sources. No methods for agents, conversations, worktrees, manager hierarchy, model prompts, checkpoint/resume or tools.

A future provider requiring **native process/session replacement** gets a separately scoped launcher/adapter mode, not an accreting dozen-method gateway interface.

---

# 11. Native client integrations

## 11.1 Codex CLI (`agentpool run codex`)

Contract:

1. Resolve exact installed Codex executable/fingerprint; run capability admission.
2. Ensure daemon active; mint local bearer and opaque session hint through owner-only IPC.
3. Generate ephemeral, version-tested Codex model-provider overrides in the supported mechanism; **do not overwrite the user's default Codex config**.
4. Construct clean child environment with only documented runtime essentials and the local bearer; remove conflicting upstream keys and unapproved proxy settings.
5. Execute native Codex with unchanged terminal IO and native exit status. AgentPool is not a TUI emulator.
6. On exit, revoke client lease, clear env storage and report errors while not leaking credentials.

Design first for real local `codex` invocation. `codex app-server` is a separate direct integration that may be easier for TradeAgent's programmatic workers. Both must pass independent shape tests.

## 11.2 TradeAgent

Use the gateway as an **inference dependency**, not an orchestration participant:

```text
TradeAgent supervisor / managers / JEV
       |
       | launches native Codex worker (separate isolated worktree)
       v
Codex App Server / CLI
       |
       | authenticated Responses request
       v
AgentPool
       |
       v
Allowed provider account/route
```

TradeAgent retains: task/retry state machine; worker supervision; worktree ownership; progress verification; JEV context lifecycle; checkpoints; tool permission semantics; semantic idempotency; reconciliation of deployments/trades/external effects. **A failed network stream can still terminate or confuse a native coding worker**. TradeAgent must define what to do when Codex reports a nonrecoverable transport failure, and its policy must avoid duplicate real-world side effects. AgentPool will not pretend that its small routing database ensures continued execution after a Codex crash.

Recommended host integration API:

```text
start_native_worker(pool_ref, workspace_ref, model_policy, host_client_id)
observe_native_worker_exit()
query_agentpool_status(pool_ref)
on_typed_no_capacity_or_transport_error -> host backoff / resume / escalate
```

`start_native_worker` belongs in **TradeAgent** (or a tiny launcher invocation), not an AgentPool durable-jobs API. Hosts may choose whether to wait for capacity; the gateway's request should not hang indefinitely.

## 11.3 T3 Code + maintained ACP adapter

Preferred path when supported:

```text
T3 Code
  |
  | negotiated ACP
  v
agentclientprotocol/codex-acp
  |
  v
Codex App Server (custom Responses provider)
  |
  v
AgentPool
```

The maintained `agentclientprotocol/codex-acp` advertises a **custom gateway auth path when the ACP client opts into that capability**. This is not blanket evidence that every T3 version opts in. Its documented runtime options include `CODEX_CONFIG`, `MODEL_PROVIDER`, `CODEX_PATH` and others. [S13]

`agentpool acp codex` performs *only*: create scoped bearer/endpoint; construct a version-tested `CODEX_CONFIG`/`MODEL_PROVIDER` launch profile; scrub conflicting auth; exec the pinned maintained `codex-acp`; leave ACP stdio and protocol events to that upstream adapter. Avoid vendoring ACP behavior without an observed integration limitation.

Required tests with actual T3 build: session create; model selection; streaming; permissions; tool/MCP calls; session resume; shutdown; recovery after disconnect; gateway auth capability negotiation; zero competing native auth handshakes; account attribution. If T3 does not support the advertised custom gateway path in that build, mark T3 as unsupported until there is a tested **standard** configuration path, not a T3 fork by default.

## 11.4 Other clients/providers

- **OpenCode / other custom-provider CLIs:** Candidate future client only if exact wire protocol matches an authorized provider endpoint and session/feature conformance passes; “OpenAI-compatible” is not a guarantee.
- **Claude Code:** Separate account/config roots, subscription availability, headless mode, and any custom endpoint must be researched/verified independently; subscription route is *not* a generic Responses gateway.
- **Google consumer Antigravity:** Disabled because of documented third-party access restriction; use authorized Gemini API/enterprise integrations if a future requirements decision selects them. [S12]
- **Generic ACP agents:** Their ACP surface is an *agent protocol*, not proof of safe bearer swapping below model requests. ACP transport compatibility and subscription entitlement are separate axes.

---
# 12. CLI and control-plane contract

## 12.1 Human-facing commands

```bash
agentpool init                         # secure storage, host identity, local policy
agentpool start                        # start local daemon (or register user service)
agentpool stop                         # graceful shutdown
agentpool status                       # health, pools, currently supported levels

agentpool login openai                 # official SIWC browser/OIDC/PKCE flow
agentpool accounts                     # labels, permission, validity; never tokens
agentpool logout acct-a                # revoke and clear selected registration
agentpool account disable acct-a
agentpool account enable acct-a

agentpool pools                        # display configured groups
agentpool pool show codex-personal
agentpool pool create codex-personal   # guided config, validates authority
agentpool policy status                # show provider-feature clearance

agentpool run codex --pool codex-personal
agentpool acp codex --pool codex-personal
agentpool doctor                       # non-consuming preflight
agentpool doctor --deep                # explicitly consent to real inference/test
agentpool route explain --pool codex-personal --model <slug>
agentpool diagnostics export --redacted
```

Commands are the **intended surface**, not evidence that current Codex/T3 builds already support each one. No configuration command can enable a gated feature merely by forcing a boolean. Account selection for L2 can be explicit (`--account acct-a`), but L3/L4 depend on admission.

## 12.2 Control RPCs (version 1)

```text
v1/status.get
v1/pools.list
v1/pools.get
v1/accounts.list
v1/accounts.beginLogin
v1/accounts.cancelLogin
v1/accounts.logout
v1/accounts.setEnabled
v1/clientLeases.issue
v1/clientLeases.revoke
v1/route.explain
v1/compatibility.get
v1/doctor.run
v1/daemon.shutdown
```

All RPCs must specify protocol version, correlation ID, bounded input, typed status and error. Authenticate the OS IPC caller and separately authorize dangerous account mutation. Control RPC does not accept raw model prompts. If JSON-RPC 2.0 is selected, use its standard envelope and a documented, versioned `error.data` structure; no arbitrary stringified internal Go errors over the socket.

Suggested stable status response concepts:

```json
{
  "api_version": 1,
  "daemon": {"state": "healthy", "pid": 1234},
  "features": {
    "codex_single_account": "admitted",
    "live_cross_account": "blocked_unverified",
    "automatic_rotation": "blocked_policy_review",
    "t3_codex_acp": "pending_conformance"
  },
  "pools": [{"id": "codex-personal", "eligible": 1, "unknown_capacity": 1}],
  "active_requests": 2
}
```

Example is illustrative, **not** a frozen production JSON schema. The team should generate client/server types from a checked-in formal schema and run wire-compatibility tests.

## 12.3 Explainability by design

Every routing admission/rejection must be explainable with a deterministic reason code:

```text
ALLOW_STICKY
ALLOW_WEIGHTED_SELECTION
DENY_LOCAL_AUTH
DENY_CLIENT_CAPABILITY
DENY_DATA_AUTHORITY
DENY_PROVIDER_POLICY
DENY_UNSUPPORTED_BILLING
DENY_REGISTRATION_REVOKED
DENY_QUOTA_SCOPE
DENY_CONCURRENCY
DENY_NO_CAPACITY
```

Status may show registration labels and source of observations, but not sensitive identities or request content. An operator must be able to answer **which registration served this request and why**, within local confidentiality boundaries, without relying on guesswork from model output.

---

# 13. Configuration and installation

## 13.1 Config example

```yaml
schema_version: 2

server:
  bind: "127.0.0.1:0"
  control_transport: "user-local"
  allow_remote: false

security:
  require_local_bearer: true
  credential_store: "platform"
  log_request_bodies: false
  log_response_bodies: false
  upstream_endpoints: ["https://api.openai.com/v1"]

limits:
  active_global: 32
  active_per_local_client: 8
  active_per_registration: 4
  max_body_mib: 16
  max_spool_mib: 256
  bootstrap_inspection_kib: 64
  max_safe_upstream_attempts: 2
  max_local_admission_wait_ms: 2000

pools:
  codex-personal:
    provider_route: "openai-siwc-responses-http"
    authority_group: "personal-dev"
    registrations: ["acct-a", "acct-b"]
    routing: "sticky-weighted-round-robin"
    billing: "subscription-included-only"
    automatic_cross_principal_rotation: false
    predictive_draining: false

integrations:
  codex:
    runtime: "external-pinned"
    require_known_good_fingerprint: true
  t3_acp:
    enabled: false  # enable after certified integration
```

Values are **illustrative starting caps**, not measured capacity/recommended OpenAI limits. The implementation must supply sensible platform and memory-specific defaults after load testing. Configuration is non-secret, owner-only, schema-validated, loaded atomically with a last-known-good rollback; unknown safety-sensitive fields fail validation.

## 13.2 Packaging

- A single signed/reproducibly built Go executable (`agentpool`) containing CLI, daemon, gateway and launchers.
- One Go module, no mandatory container runtime or external database. SQLite via a maintained driver; evaluate CGO versus pure-Go trade-offs on each supported OS, including SQLite build options and signed release reproducibility.
- Platform secret-store and IPC implementations selected behind build tags/interfaces.
- Native user-service integration: macOS LaunchAgent first; systemd user unit later; Windows per-user background service/task later. Avoid running as root/administrator.
- Managed auto-updates **not** a v1 core feature; provide version pin, explicit install and reversible schema migrations. Updates must never reset permissions or silently raise capability level.

## 13.3 Minimal repository

```text
agentpool/
├── cmd/agentpool/main.go
├── internal/
│   ├── control/          # local IPC and typed admin RPC
│   ├── gateway/          # HTTP ingress, SSE forwarding, deadlines, cancellation
│   ├── routing/          # eligibility + sticky selection + CAS generation
│   ├── capacity/         # observations/scopes/freshness
│   ├── auth/siwc/        # OIDC, PKCE, registrations, refresh
│   ├── credentials/      # secret store abstractions + platform backends
│   ├── store/            # SQLite, migrations, atomic config
│   ├── compatibility/    # version admissions and conformance
│   ├── launch/           # native Codex and maintained codex-acp launchers
│   ├── policy/           # provider operating-mode and data authority gates
│   └── diagnostics/      # metrics/status/redaction/support bundles
├── schema/               # control API and config schemas
├── migrations/
├── testdata/             # deterministic redacted protocol fixtures
├── tests/
│   ├── contract/
│   ├── fault/
│   ├── security/
│   ├── live-gated/
│   └── upgrade/
├── compat/               # pinned known-good and known-bad manifests
├── docs/
│   ├── RUNBOOK.md
│   ├── SECURITY.md
│   ├── PROVIDER_POLICY.md
│   └── EVIDENCE_REGISTER.md
├── go.mod
└── README.md
```

Keep packages honest and few. These are source organization boundaries, **not microservices or runtime plugins**. Never create eleven sub-daemons because the directory tree has eleven folders.

---

# 14. Implementation roadmap with hard exit conditions

**Engineering can start immediately with mocks and the core. Automatic exhaustion-triggered multi-registration use remains disabled until its policy and technical gates pass.** Each phase produces independently testable artifacts; no phase is “done” because it compiled.

## Phase 0 — Proof and permission (must happen before claiming L3/L4)

**Deliverables:** Provider-permission record; exact current SIWC/Codex compatibility constraints; pinned Codex runtime candidate; reproducible two-registration experimental plan; data authority policy.

**Exit:** `G0` documented; critical feature decision explicitly `ALLOWED`, `DENIED`, or `UNRESOLVED`. Unresolved does not count as allowed. Obtain explicit technical testing authorization and do not launch exhaustion-triggered rotation without applicable clearance.

## Phase 1 — Pure routing kernel + simulated upstream (L0)

Implement immutable request metadata, policy predicate, scoped observations, deterministic selection, bounded body retention, local bearer validation, SSE forwarding, cancellation, CAS binding generation, atomic in-flight admission, clean shutdown. Create synthetic/mock provider and deterministic fault-injection test harness **before** OAuth.

**Exit:** race tests, property tests, fuzzing and adversarial stream test suite pass; zero transparent post-commit retries; no credential content in logs; predictable resource limits under stress.

## Phase 2 — Official SIWC registration and single-account inference (L1)

Implement official browser authorization, PKCE/nonce/JWKS validation, workspace/scope checks, OS secret store, safe refresh token rotation, sign-out/revocation, direct `POST /v1/responses` and current preview validation. Begin with one explicitly authorized account.

**Exit:** real single-account request passes without ambient API key; negative auth/security tests pass; credentials survive restart where expected; logout blocks future requests; accidental paid fallback not possible in implementation path.

## Phase 3 — Codex gateway integration (L1/L2)

Implement pinned custom-provider launcher, local bearer env wiring, exact HTTP/SSE shape detection and native App Server/CLI tests. Validate correct process exit and that unsupported Codex versions fail closed.

**Exit:** one real multi-turn Codex thread, tools, code edits, cancellation and native completion through one SIWC account, with evidence that all inference traffic uses the intended local gateway. No claims of rotation yet.

## Phase 4 — Cross-principal continuity feasibility (L3, gated)

With appropriately authorized A and B, test same-model requests on the same native Codex thread over live A→B→A changes. Include opaque reasoning/compaction, model feature differences, multiple local workers and cache/latency measurement. Record provider request attribution without exposing tokens. Repeat across exact declared versions.

**Exit:** all `G3`–`G5` tests pass with sanitized traces **and** provider permission for that behavior is confirmed. Failure means L3 stays OFF; do not strip context or silently start a new thread.

## Phase 5 — Policy-approved automatic rotation (L4, gated)

Only after Phase 4 and policy approval: implement scoped capacity-triggered switching, safe explicit-rejection bootstrap retry, bounded backoff, prewarm, optional genuine data-driven draining and clean exhausted/no-capacity errors.

**Exit:** all capacity-error/failover tests pass; false-global-exhaustion and hidden-replay counts remain zero; billing/permission monitors show expected behavior; no surprise fallback.

## Phase 6 — T3 integration (L5)

Build thin `codex-acp` launcher using the current maintained adapter and documented config options; validate actual T3 ACP negotiation and session lifecycle.

**Exit:** real T3 can create/continue sessions, stream tools, handle model selection, cancel/restart and relay failure, with clean account provenance, under pinned versions.

## Phase 7 — Operational hardening and release

Run long-duration mixed workload with macOS sleep/wake, daemon crashes, process kills, token renewal, expired sessions, corrupted DB, disk exhaustion, provider 429/503, network flaps, multiple Codex workers and subagents. Add signed binary/update provenance, runbook, diagnostics, compatibility manifests, security review, rollback and support escalation guidance.

**Exit:** full gated acceptance suite and reliability scorecard below. Publish only capability levels actually passed; do not promote L4 based on synthetic tests alone.

## Phase 8 — Expansion by evidence, not ambition

Investigate (in order): additional Codex launch clients, permitted non-OpenAI provider routes, Claude-specific account isolation and runtime switching, authorized enterprise/API Gemini, optional provider-specific native resume fallback. **Each** gets its own policy and protocol compatibility review. No generic “swap any provider underneath any agent” claim.

---

# 15. Acceptance matrix: the proof campaign

Each gate records: purpose; exact runtime/build hashes; OS; registered account identifiers **redacted**; model; synthetic test script; expected state; observed state; network/stream evidence; timestamps; result; operator approval. Security-sensitive evidence must be kept in a restricted artifact store, not in public CI.

| Gate | Test | Pass criterion | Release dependency |
|---|---|---|---|
| **G0** | Provider permission for actual automated subscription rotation | Appropriate documented clearance or explicit feature prohibition | L4 blocking |
| **G1** | Two independent real SIWC registrations A/B | Verified separate client IDs, subjects/workspace, inference scope and deliberate user authority; no ambient API-key usage | L3 blocking |
| **G2** | Codex custom provider through mock Responses gateway | All inference reaches gateway; bearer/path/WebSocket-disabled settings verified | L1 blocking |
| **G3** | Same live Codex process + thread A→B→A | Original PID/thread persists; context, tools, model and telemetry correct | L3 blocking |
| **G4** | Encrypted reasoning and compaction across registration switch | No encrypted-content errors, silent context loss, reasoning transformation or false completion | L3 blocking |
| **G5** | Explicit pre-output rejection and safe alternate attempt | No duplicate downstream stream; old attempt closed; exact original request; right account/scope | L4 blocking |
| **G6** | Ambiguous precommit, postcommit and cancellation failures | No invisible reroute, clear failure, bounded time/memory, correct native interpretation | All |
| **G7** | Concurrent Codex/TradeAgent-style workers + subagents | No principal crossovers, stale binding rollback, quota global misclassification or unbounded queue | L4 blocking |
| **G8** | Actual T3 → ACP → Codex → gateway → SIWC | Full client session/tools/model/permission/cancel path works | L5 blocking |
| **G9** | Upgrade canary, malformed SSE, unknown field, bad-client fixture | Capability downgrade/quarantine rather than hidden incorrect behavior | All |
| **G10** | Full-day soak with token refresh, sleep/wake, all-capacity-loss | All observable failures diagnosed; no safety invariant violations; availability measured honestly | Operational release |
| **G11** | Security abuse suite (localhost abuse, auth replay, leak scan) | Unauthorized access and egress blocked, logs redacted, tokens isolated | All |
| **G12** | Disk-full, SQLite recovery, secret-store failure, clock rollback | No state corruption mistaken for capacity or credentials; clear repair path | Operational release |
| **G13** | Host/native-agent crash and recovery | Gateway reports transport failure, host native recovery works at proven boundary; no false durable-job promise | TradeAgent integration |

### 15.1 Deterministic stream fixtures (must have)

```text
S01 success short SSE
S02 long valid SSE with heartbeats
S03 early explicit 429 by relevant scope
S04 app-specific cap 429 with unknown reset
S05 generic unknown 429
S06 provider-wide 503
S07 no headers, connection reset
S08 headers delivered, no SSE body, disconnect
S09 first SSE semantic event then abrupt EOF
S10 response.failed after partial delta
S11 response.completed plus late disconnect
S12 malformed SSE frame and oversized frame
S13 client cancel before headers / during stream
S14 upstream ignores cancellation
S15 two concurrent requests from same binding
S16 token refresh races and revoked credentials
S17 clock jump/sleep and stale reset observation
S18 local bearer missing/wrong/expired/revoked
S19 foreign workspace/client attempts cross-pool inference
S20 unsupported SIWC field, previous_response_id, WebSocket route
S21 encrypted reasoning and compaction payload passthrough
S22 process crash during secret replacement / disk full
S23 two overlapping same-session requests while a switch is pending
```

### 15.2 Required automated methods

- **Property tests:** no stale binding overwrite; pool membership/authority invariant; no route to disabled account; scoped 429 cannot globally exhaust every model without evidence.
- **Fuzz:** bounded JSON parser, SSE framing, headers/transfer encoding, malformed HTTP, OIDC callback validation, config schema and error envelopes.
- **Race/concurrency:** Go race detector; refresh singleflight; cross-store secret generation reconciliation; cancellation; config reload; request reserve/release; per-session drain barrier; daemon singleton election.
- **Security:** credential redaction snapshots, CSRF and localhost-origin attempts, filesystem symlink attacks, revoked local bearer, outbound DNS/proxy/redirect abuse.
- **Fault injection:** abrupt process termination, secret store locked/unavailable, disk full, SQLite busy/corrupt, OS sleep/wake, network half-close, provider timeout.
- **Live-gated:** two-account OAuth, same-process thread continuity, encrypted/compaction behavior, actual native Codex and T3. Never substitute mocks for these proof obligations.

### 15.3 Quantitative release scorecard

The following are **targets for engineering validation, not measured results**:

| Metric | Target before operational release |
|---|---|
| Incorrect principal/billing path selected | **0** in all certified tests; release-blocking any instance |
| Hidden post-commit transparent retries | **0** by invariant |
| Unauthorized provider-policy mode activated | **0** |
| Secret leaks in ordinary logs/support export | **0** |
| False `response.completed` on interrupted/failed stream | **0** |
| Stale binding generation wins | **0** under deterministic race suite |
| Memory/disk/active request over configured caps | **0** under fault and stress suite |
| Auth refresh races causing accidental reuse of replaced token | **0** in controlled concurrent tests |
| Sustained/soak availability | Report actual rate and failure classes, no invented 99.99% promise |
| Rotation median/p95 impact | **Measure** latency/throughput/cache penalty; set SLO after live baseline |
| Native agent continuation | **Verified** only for exact model/client/session conditions that passed G3/G4 |

No arbitrary numeric pass percentage excuses one safety-invariant violation. A 24-hour successful soak is insufficient if a significant untested upgrade, provider policy change or model difference invalidates the certified seam.

---

# 16. Failure runbook: what operators and hosts will actually see

| Incident | AgentPool action | Native agent/host action |
|---|---|---|
| One registered account's token expires | Serialized refresh; preserve registration identity | Usually none if refresh works before next dispatch |
| Refresh grant revoked | Stop admission; require login | Surface auth issue; use another account only if mode permitted |
| Provider rejects unsupported request field | Reject/mark compatibility mismatch, do not strip context | Upgrade/downgrade native client or disable feature deliberately |
| Applicable account quota rejection | Record exact scope and code; safe retry only after all gates | Continue if valid new attempt succeeds; otherwise wait/fail clearly |
| All approved capacity unavailable | Structured no-capacity with best evidence and optional reset | Back off or pause durable job according to host policy |
| Provider global overload | Provider-level breaker/backoff | Retry after provider recovery, not swap every account |
| Stream fails after response begins | End/fail; no invisible replay | Native Codex and host determine safe resume/reconciliation |
| Gateway daemon crashes | Active connection fails; restart reconstructs routing/auth metadata | Native agent may need transport retry/session recovery; TradeAgent owns job state |
| Compaction/reasoning incompatible after switch | L3 revoked/quarantined for profile | Revert to proven single-principal path or host-managed safe handoff only if explicitly supported |
| Provider updates Terms | Disable affected operating mode immediately | Existing jobs receive clear capacity/policy unavailable result |
| A pool includes an unauthorized workspace | Reject admission; audit config | Owner must reconfigure authorization group |
| Local bearer stolen | Revoke lease, stop admission, rotate affected local credentials | Relaunch authorized client |
| SQLite corrupt or secret store locked | Fail closed, preserve files, provide recovery guidance | Host receives unavailable diagnostic, never reroutes to arbitrary ambient login |
| Machine sleeps beyond known quota reset | Re-evaluate observations on wake, bounded probe if safe | Retry or resume under host policy |

## 16.1 No-capacity policy for TradeAgent

TradeAgent is a durable orchestration system: it may pause a coding worker or requeue verified work without requiring the gateway to keep an HTTP socket open for hours. Define a typed `capacity_unavailable` response and host backoff policy; let JEV and TradeAgent's supervisor keep doing their own jobs. AgentPool must not silently spin with repeated attempts on capacity-impaired registrations.

## 16.2 Crash recovery limitation

A local HTTP gateway cannot preserve a socket that no longer exists, cannot guarantee an in-flight provider request had no effect, and cannot resurrect a dead Codex process. It *can* retain account registrations, route policy and previously observed capacity, then safely accept new requests after restart. That is the correct scoped promise.

---

# 17. Deferred fallback: when the switch seam does not work

Do **not** conceal a failed L3 test with heuristic history rewriting or pretend an alternate agent process is the same session. Instead choose one of the following explicit product outcomes:

| Outcome | Action |
|---|---|
| L3 works for exact client/model/identity profile | Certify that combination; allow L4 only after policy clearance |
| L3 fails for encrypted context, compaction or auth | Disable L3 for that combination; retain L1/L2 single-account gateway |
| Codex supports a verified native resume at a safe turn boundary | Consider a **small Codex-specific resume launcher** outside the gateway; host orchestrator maintains durable jobs |
| Provider supports a permitted official multi-registration API with better semantics | Prefer that documented interface even if it changes transport adapter |
| Terms prohibit automatic exhaustion-triggered rotation | Do not implement/enable it; retain authorized single-account or explicitly permitted billing paths |
| Another provider has different session/auth design | Give it a dedicated adapter and conformance track; do not retrofit process rotation into every transport request |

If a native resume fallback is justified, its safety constraints come from the older broker's verified principles: explicit safe turn boundary, authenticated session domain, process version pinning, no duplicate active writer, and host-owned reconciliation. **Do not transplant** the entire old broker's jobs, leases, handoff agents, cross-provider checkpoints and runtime manager into the v1 Switch.

This is how we preserve simplicity while refusing to bury an important failure mode.

---

# 18. Architecture invariants — tests should map to these IDs

| ID | Invariant |
|---|---|
| **I01** | Every upstream request is attributable to one explicit eligible registration, not a mixed client ID/token pair. |
| **I02** | No request silently changes its model, provider route, billing path or data-authority group. |
| **I03** | Automatic exhaustion-triggered rotation is OFF until policy and technical gates have passed. |
| **I04** | At most one upstream attempt can be authoritative at any time for a downstream request. |
| **I05** | Zero hidden retries after downstream response commitment. |
| **I06** | Unknown upstream acceptance state is not called an explicit rejection. |
| **I07** | An account-scoped/app-scoped/model-scoped limit does not automatically become global plan exhaustion. |
| **I08** | Unknown *capacity* can be attempted under bounded policy; unknown *identity/permission/compatibility* blocks admission. |
| **I09** | A model catalog is not an entitlement guarantee. |
| **I10** | OAuth renewals are serialized by registration and stale rotated refresh tokens are never intentionally reused. |
| **I11** | Stale binding generations cannot overwrite newer account selection. |
| **I12** | Local and upstream credentials are separated, scoped and redacted. |
| **I13** | No normal request/response/reasoning content is persistently logged. |
| **I14** | User/client cancellation releases local resources and does not create an invisible replacement request. |
| **I15** | Request/body/stream/queue concurrency and time budgets remain bounded. |
| **I16** | A crash does not magically preserve HTTP streams; no successful native-agent completion is fabricated. |
| **I17** | Runtime/protocol change cannot silently enable unverified high-risk capability. |
| **I18** | TradeAgent owns orchestration, checkpoints, tool side effects and durable job recovery. |
| **I19** | Adding an authorized registration is data/configuration, not a code change. |
| **I20** | No use of provider-prohibited private routes, unauthorized consumer Antigravity, credential scraping or rate-limit evasions. |
| **I21** | L3 keeps Codex's native process and thread **only when the exact certified test conditions hold**. |
| **I22** | No silent semantic mutation of opaque reasoning/compaction/tool content to force an upstream success. |
| **I23** | An exposed reset is treated as a timestamped provider observation, not an invented timer. |
| **I24** | An unverified billing outcome is reported as unverified, never as guaranteed free or guaranteed within allowance. |
| **I25** | A known session's new account binding cannot outrun its outstanding prior-account requests; absent reliable session identity, L3/L4 stay disabled. |

Every invariant must have at least one automated test or a named human/provider approval record. Unknown is not pass.

---

# 19. Team implementation conventions and governance

## 19.1 Small code; deep evidence

Prefer a few boring modules with explicit boundaries over a complicated universal framework. Review any new persistent object, worker background loop, transport protocol, daemon, external dependency or feature flag against: **“Can this behavior be implemented by the native agent or the provider?”** If yes, do not duplicate it in AgentPool.

## 19.2 PR definition of done

Every PR includes a failure-mode test; relevant policy/compat changes; release gate impact; benchmark or resource impact if streaming changes; a schema migration/rollback check if persistence changes; and documented negative-path behavior. AgentPool engineers should be able to review **why a retry is safe** without reading Codex's internal conversation state.

## 19.3 Required ownership

| Area | Accountable team responsibility |
|---|---|
| Core architecture | Maintain product boundary, invariant registry and ADR changes |
| Auth/security | OAuth correctness, Keychain/secret store, local bearer, threat model |
| Transport | Responses shape, SSE, cancellation, explicit rejection vs ambiguous state |
| Provider/policy | SIWC docs and Terms watch, account/workspace data authority, billing restrictions |
| Integrations | Exact-version Codex CLI/App Server and maintained ACP/T3 conformance |
| Quality/reliability | Fault injection, soak, release metrics, migration/rollback |

One engineer can own multiple roles in a small team, but none of these responsibilities should silently disappear.

## 19.4 Release labels

Use explicit labels rather than a single optimistic “AgentPool works” badge:

```text
CODEX_L1_SINGLE_ACCOUNT=PASSED
CODEX_L2_MANUAL_ACCOUNT_SELECTION=PASSED
CODEX_L3_LIVE_CROSS_ACCOUNT=UNVERIFIED
CODEX_L4_AUTOMATIC_EXHAUSTION_ROTATION=POLICY_HOLD
T3_L5=UNVERIFIED
GOOGLE_CONSUMER_ANTIGRAVITY=UNSUPPORTED
```

These are **illustrative initial statuses**, not claims of tests passed today. On project inception set even L1/L2 to `UNVERIFIED` until the actual tests run.

---

# 20. Source and evidence register

**Interpretation rule:** Primary vendor documentation describes supported pieces and obligations. Open-source source code/repository READMEs show an implementation snapshot. GitHub issue reports are troubleshooting signals **not normative product guarantees**. Product-specific permission requires a separate policy decision. No source listed here proves an actual two-account AgentPool integration has already passed.

## Vendor documentation (primary)

- **[S01] OpenAI — SIWC open-source plan usage overview:** https://developers.openai.com/siwc/token-sharing-open-source  
  Basis: local OSS scope, agent host identity, public Responses route, provider registration context.
- **[S02] OpenAI — Accounts and sessions:** https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions  
  Basis: multiple saved account registrations, separate client/token pairs, serialized refresh, sign-out and usage UI.
- **[S03] OpenAI — Registration and sign-in:** https://developers.openai.com/siwc/token-sharing-open-source/sign-in  
  Basis: `dynamic_agent_client` initial registration, issued `client_id`, PKCE, `state`, OIDC nonce/JWKS/identity/scope validation.
- **[S04] OpenAI — Models and inference:** https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference  
  Basis: current model list and public `POST /v1/responses` with OAuth bearer; no ChatGPT private backend.
- **[S05] OpenAI — Preview limitations:** https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations  
  Basis: `store:false`, `stream:true`, stateless HTTP history, unsupported request keys and WebSocket boundaries.
- **[S06] OpenAI — Codex App Server with SIWC:** https://developers.openai.com/siwc/token-sharing-open-source/codex-app-server  
  Basis: custom model provider settings and app-server token-injection behavior; documented restart/resume on token renewal.
- **[S07] OpenAI — Errors and recovery:** https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery  
  Basis: exact subscription-sharing errors; app-specific limit ambiguity; 503 usage unavailable and route/capability failures.
- **[S08] OpenAI — Token reference:** https://developers.openai.com/siwc/token-sharing-open-source/token-reference  
  Basis: OAuth token lifetimes and replacement refresh token semantics as currently documented.
- **[S09] OpenAI — Self-hosted VM guidance:** https://developers.openai.com/siwc/token-sharing-open-source/self-hosted-vms  
  Basis: distinct host identity, protected credential transfer and current limitations, not automatic team-server permission.
- **[S10] Codex configuration reference:** https://learn.chatgpt.com/docs/config-file/config-reference  
  Basis: configurable model providers; installation/version-specific settings must be rechecked.
- **[S11] OpenAI — Europe Terms of Use (updated 2026-01-16):** https://openai.com/en-GB/policies/eu-terms-of-use/  
  Basis: restriction on circumventing rate limits/restrictions; does not itself provide approval for the proposed pool rotation.
- **[S12] Google Antigravity FAQ and Terms:** https://www.antigravity.google/docs/faq/ and https://www.antigravity.google/terms  
  Basis: explicit third-party access restriction affecting consumer Antigravity.

## Open-source implementation and issue evidence

- **[S13] Maintained ACP Codex adapter:** https://github.com/agentclientprotocol/codex-acp  
  Basis: `codex-acp` is stdio ACP server using Codex App Server; documented custom gateway auth is conditional on client opt-in; environment configuration options.
- **[S14] Codex issue #17541 — `invalid_encrypted_content` on model/deployment switch:** https://github.com/openai/codex/issues/17541  
  Basis: a *reported* encrypted-history portability failure in a different scenario; motivates account-switch compaction/reasoning tests, not evidence that changing accounts necessarily fails.
- **[S15] Codex issue #37245 — reported ignored custom provider under some ChatGPT login conditions:** https://github.com/openai/codex/issues/37245  
  Basis: negative-version conformance case; do not assert universal applicability or a permanent upstream bug.
- **[S16] T3 Code repository:** https://github.com/pingdotgg/t3code/blob/main/packages/contracts/src/settings.ts and https://github.com/pingdotgg/t3code/blob/main/docs/user/providers-codex.md  
  Basis: client-specific provider settings and Codex integration; real ACP gateway negotiation remains to be demonstrated in target release.

## Internal provenance

- **[I-A]** `AgentPool_Architecture_Final(2).md` — earlier durable execution broker reference for fallback-only safety patterns.
- **[I-B]** `AgentPool_The_Switch_Architecture(1).md` — immediate predecessor and core simplification.
- **[I-C]** `AgentPool_Switch_vs_Broker_Critical_Architecture_Audit.md` — independent assessment of missing proofs and P0/P1 blockers.

**Evidence freshness:** The provider docs and repositories can change after 2026-10-08. At each release gate and before production rollout, verify current sources and record their revision/date. Do not describe preview endpoints as evergreen stable interfaces.

---

# 21. Final architectural contract

The product should *feel* like a fuel selector, but behave like audited infrastructure.

```text
                 THE SWITCH, IN 10 RULES

  1. The native agent is the engine; never rebuild it unnecessarily.
  2. AgentPool controls inference authorization, not agent cognition.
  3. One logical request has one authoritative upstream attempt.
  4. Transparent retries require an explicit proven-safe rejection.
  5. Opaque native context always remains opaque and unmodified.
  6. Every account, billing path and data authority is explicit.
  7. Quota is scoped evidence; unknown is not “exhausted.”
  8. Versions and protocol features are certified, not assumed.
  9. Provider permission is a hard gate, not an afterthought.
 10. When proof runs out, stop safely and explain why.
```

## Engineering authorization

**Build the small local gateway, official SIWC single-registration path, secure routing core, native Codex launcher, deterministic fault harness and version compatibility system.** Those are the right foundations irrespective of whether live cross-account rotation passes.

**Do not enable or promise automatic exhaustion-triggered account switching until both (a) the applicable provider permission decision and (b) a real, repeatable same-process, same-thread, encrypted-history/compaction continuity test have passed.** Do not silently replace that proof with a fallback broker or a different billing path.

If the seam passes, AgentPool becomes the elegantly small infrastructure originally envisioned. If it does not, we keep a valuable, secure component and adapt only the missing part. **This is the implementation architecture I would put in front of a team: small in responsibilities, precise in invariants, and explicit about every claim the test suite has not yet earned.**
