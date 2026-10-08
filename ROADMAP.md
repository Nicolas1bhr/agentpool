# Roadmap

Milestones come from [spec v5.2 §15](docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md):
**complete a thin slice first.** A milestone is done when its gate passes, not when its code exists.

| Milestone | Scope | Gate | Status |
|---|---|---|---|
| **P0 — Freeze & measure** | Pin the OpenCode V2 release and provider package; enumerate target API-key and account-based connector modes; identify the host integration seam | Real config and captured HTTP trace; written connector capability profile | Not started |
| **P1 — Gateway proof** | One authenticated loopback endpoint, fake upstream, one exact protocol, real OpenCode TUI/headless fixture | Full multi-turn stream with tools; no bypass; no hidden fake success | **Skeleton** (loopback bind, origin rejection, honest 501s) |
| P2 — Minimal core | Go service, multi-registration schema and automatic selector interface, vault, local IPC, one alias, one API key, request/stream state machine, budget grant | Security/stream fault tests pass; no upstream token in OpenCode; single route uses the future multi-route selector | Not started |
| P3 — Host connection | Scoped per-worker gateway tokens, private OpenCode config, metadata correlation, strict egress and tool boundaries (TradeAgent as first host) | Same native OpenCode host/session/tool interfaces with direct-provider and AgentPool fixtures | Not started |
| P4 — Automatic multi-route switching & proactive admission | Indexed route registry; two certified same-model routes; `auto_eligible`; per-owner rate evidence, reservations, headroom; optional pin/override; immutable epochs | **Without operator action**, known-insufficient route A is excluded before dispatch and B selected; unrelated streams continue | Not started |
| P5 — New protocol/provider | One more adapter with full stream/tool/error/usage fixtures | That backend's route matrix earns `PROVEN` | Not started |
| P6 — Production hardening | Cross-platform vault/IPC/sandbox proofs, signed releases, soak, connector auditing, usage reconciliation, hot-path latency | All release-blocking gates pass on target OS | Not started |
| P7 — Optional JEV risk advisor | Async typed risk observations on redacted provider-health summaries; shadow → calibrate → opt-in soft penalty | JEV-01..09 and held-out calibration pass; nothing on the hot path | Not started |

## Definition of done for every PR (spec §15.1)

See [CONTRIBUTING.md](CONTRIBUTING.md#definition-of-done).
