# Contributing to AgentPool

The [v5.2 specification](docs/architecture/AgentPool_v5.2_Proactive_Routing_JEV_Implementation_Spec.md)
is normative. Read §0.0 before writing code: AgentPool is **only a model provider**. Proposals that
make it launch agents, run tools, hold sessions or select accounts inside OpenCode are out of scope
by design (§0.0.2 anti-patterns).

## Getting started

```bash
make check   # gofmt check, go vet, race-enabled tests
make build   # ./bin/agentpool
```

Go 1.26+. Prefer the standard library; a new dependency needs a demonstrated engineering advantage
and a security/version review (§13.2).

## Definition of done

From spec §15.1. Every PR:

- states its contract and ships tests with any behavior change;
- adds at least one deterministic negative/edge-case test for every safety rule it touches;
- keeps model, protocol and credential authorization traceable before and after;
- tests error behavior and financial uncertainty instead of assuming them;
- for capacity code: includes pre-dispatch proof, shared-owner concurrent reservation, stale
  observation and zero-external-hot-path checks;
- for advisor code: preserves deterministic allow/deny and carries a shadow/evaluation plan and a
  redacted-output privacy proof;
- adds no unreviewed dependency, network target, persisted user prompt or authority expansion;
- updates the compatibility manifest, operator docs and [CHANGELOG.md](CHANGELOG.md) in the same PR;
- **never claims `PROVEN` without pinned real-client evidence.** Fake-server unit tests passing
  is not proof that a real OpenCode build works; mark such paths `UNVERIFIED`.

## Connectors

New provider connectors must follow [docs/PROVIDER_POLICY.md](docs/PROVIDER_POLICY.md):
documented flows only, permitted by the provider's current terms, for one owner's own accounts,
with a certification record.

## Spec changes

The spec is versioned in its header. Propose changes in an issue first; an accepted change lands as
a new spec revision, and the previous one moves to `docs/architecture/history/`.

## Licensing of contributions

By contributing you agree that your contribution is licensed under the
[GNU AGPL v3.0](LICENSE), the license of this repository.

## Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
