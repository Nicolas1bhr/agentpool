# Architecture

## Current specification

**[AgentPool v5.2 — Proactive Routing & JEV Implementation Spec](AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md)**
(architecture version 5.2.0, research baseline 2026-10-08). Normative language: `MUST` blocks,
`SHOULD` is default, `MAY` is optional; `DESIGN` = not implemented, `DOCUMENTED` = vendor-published,
`PROVEN` = passing pinned conformance test, and `UNKNOWN` is never treated as `PROVEN`.

v5.2 supersedes all earlier documents **on the location of the provider boundary and all resulting
interfaces**: AgentPool sits *below* OpenCode as a model provider, never above it as an agent bridge.

## Independent research

**[Feasibility, Continuity, and the Limits of Inference Capacity Pooling](AgentPool_Independent_Research_Paper_2026-10-08.md)**
(2026-10-08) — an evidence-based review of the v5.2 design against vendor documentation, public
source code and field reports. Verdict: engineering **conditional GO**; commercial thesis **not yet
investment-proven**. It is research, not a certification, security audit or legal opinion.

## History

Kept for the reasoning they record (spec §16.1 lists which ideas remain binding). Oldest first:

| Document | Notes |
|---|---|
| [Architecture Paper](history/AgentPool_Architecture_Paper.md) | Original concept |
| [Architecture Final](history/AgentPool_Architecture_Final.md) | Researched design baseline |
| [The Switch](history/AgentPool_The_Switch_Architecture.md) | Recommended architecture / blueprint |
| [The Switch v2](history/AgentPool_The_Switch_Architecture_v2.md) | Core design; provider capabilities unproven |
| [The Switch v3](history/AgentPool_The_Switch_Architecture_v3_Implementation_Spec.md) | Approved engineering baseline |
| [v5 OpenCode Inference Gateway](history/AgentPool_v5_OpenCode_Inference_Gateway_Implementation_Spec.md) | Provider boundary moved below OpenCode |
| [v5 Clarified Provider Contract](history/AgentPool_v5_Clarified_Provider_Contract.md) | Provider-only contract clarified |
| [v5.1 Automatic Inference Pool](history/AgentPool_v5.1_Automatic_Inference_Pool_Implementation_Spec.md) | Automatic pooling made core |

## Module map (spec §13.1 → code)

Only one implementation of request admission, identity, attempt-state transitions and policy
evaluation; adapters never duplicate them. Packages are created when their milestone starts.

| Module | Responsibility | State |
|---|---|---|
| `cmd/agentpool` | One binary: `serve`, `version`; later `select`, `status`, `doctor` | Skeleton |
| `internal/version` | Build identity via ldflags | Done |
| `internal/gateway` | HTTP admission, request isolation, stream relay | Skeleton — loopback-only bind, browser-origin rejection, honest 501s |
| `internal/catalog` | Logical aliases, limits, capabilities | Not started (P2) |
| `internal/routing` | Route eligibility, affinity, modes, epochs | Not started (P2, P4) |
| `internal/capacity` | Scoped limiter observations, reservations, headroom | Not started (P4) |
| `internal/adapters/*` | Certified protocol adapters (Chat Completions, Responses, …) | Not started (P1, P5) |
| `internal/credential` | Vault references, refresh, revocation | Not started (P2) |
| `internal/policy` | Service permissions and authority | Not started (P2) |
| `internal/budget` | Reservations, usage, capacity ownership | Not started (P2) |
| `internal/storage` | SQLite migrations and durable events | Not started (P2) |
| `internal/control` | Owner-only IPC (JSON-RPC over Unix socket / named pipe) | Not started (P2) |
| `internal/telemetry` | Safe structured status and audit metadata | Not started (P2) |
| `internal/advisory` | Optional async JEV route-risk advisor | Not started (P7, separate promotion) |
| `specs/` | JSON Schemas and normative state machines | Empty |
| `fixtures/`, `compatibility/` | Replayable synthetic traces; pinned, hash-identified proofs | Created with P1 |
