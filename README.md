# AgentPool

**One model provider in front of every AI account you own — routing is automatic, safety is first.**

AgentPool is a small, local-first inference service. Your agent runtime (OpenCode first) sees a
single model provider called `agentpool`. Behind it sits a pool of independently authenticated
routes — API keys, subscription-backed accounts, enterprise endpoints and local models — and for
every new inference request AgentPool automatically picks an eligible, certified route based on the
requested model alias, capabilities, permissions, session continuity and remaining capacity.

When one route runs out of capacity, cools down or becomes ineligible, the *next* request moves to
another without anyone pressing a button. In-flight streams stay pinned to the route they started
on. Nothing is silently swapped for a different model.

```text
Host (e.g. TradeAgent) ── controls agents ──▶ OpenCode ── model requests ──▶ AgentPool ──▶ route A │ B │ C …
                                                                           (one provider)   (your accounts)
```

> **TradeAgent owns the work. OpenCode executes the agents. AgentPool supplies the inference.**
> AgentPool never owns an agent, its task, its session, its tools, its files or its memory.

## Critical first

The design ([spec v5.2](docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md))
puts correctness and safety ahead of features:

- **Provider only.** AgentPool is a model endpoint, never an agent runner, task queue or tool executor.
- **Fail closed.** Unknown capabilities, unverified routes and ambiguous retries are refused, not guessed.
- **No silent substitution.** A request for model X is never quietly answered by unrelated model Y.
- **Honest status.** `UNKNOWN` is never treated as `PROVEN`; nothing is called working until a
  pinned, real-client conformance test passes.
- **Credentials stay put.** Upstream keys live only in the AgentPool process and the OS vault —
  never in OpenCode's config, prompts or a worker's environment.
- **Loopback only.** No LAN/WAN listener in v1; browser-origin requests are rejected.
- **AI-assisted, never AI-governed.** An optional route-risk advisor (JEV) may *suggest* soft
  ranking penalties off the hot path; deterministic code alone decides capacity, policy, billing,
  credentials and retries.

## Status

**v0.1.0 — skeleton.** The binary builds, serves `/healthz` on loopback, rejects browser origins,
and answers every inference path with an explicit `501 not implemented`. No routing, credentials or
upstream calls exist yet. See [ROADMAP.md](ROADMAP.md) for milestones P0–P7 and their gates.

```bash
make test          # vet + race-enabled tests
make build         # ./bin/agentpool
./bin/agentpool serve
curl http://127.0.0.1:43819/healthz
```

Requires Go 1.26 or newer.

## Whose accounts? Yours.

AgentPool pools accounts that **one owner** holds, for that owner's own use. It is not a tool for
sharing or reselling accounts between people. A connector ships only if it uses a flow its provider
documents and whose terms permit that use — see [docs/PROVIDER_POLICY.md](docs/PROVIDER_POLICY.md).

## Documentation

| Document | Purpose |
|---|---|
| [Spec v5.2 (current)](docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md) | Normative implementation specification |
| [docs/architecture/](docs/architecture/) | Spec index, version history, module map |
| [docs/PROVIDER_POLICY.md](docs/PROVIDER_POLICY.md) | Which connectors may exist and how they are certified |
| [examples/opencode.jsonc](examples/opencode.jsonc) | Illustrative OpenCode provider config (unverified) |
| [ROADMAP.md](ROADMAP.md) | Milestones and release gates |

## Versioning

Software releases follow [Semantic Versioning](https://semver.org) via git tags (`v0.1.0`), with
changes recorded in [CHANGELOG.md](CHANGELOG.md). The **architecture** version (currently 5.2.0) is
separate and lives in the spec's header. Before 1.0.0, minor releases may change interfaces.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security issues: [SECURITY.md](SECURITY.md).

## License

[GNU Affero General Public License v3.0](LICENSE). If you run a modified AgentPool as a network
service, you must offer its source to that service's users.

## Credits

Built by [Nicolas Beeckman](https://github.com/Nicolas1bhr).

The Go skeleton, tests, CI and repository documentation were written with Claude Opus 5.5
(Anthropic), via Claude Code, working from the v5.2 specification.
