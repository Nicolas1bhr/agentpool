# Working on AgentPool

The spec in `docs/architecture/` (v5.2) is normative — read its §0.0 before writing code. Contribution rules: `CONTRIBUTING.md`;
connector rules: `docs/PROVIDER_POLICY.md`.

## Orchestration

Before any delegated or multi-agent work — one agent or a fleet — read
[`docs/ORCHESTRATION-STANDARD.md`](docs/ORCHESTRATION-STANDARD.md), the shared standard (version 2026-10-08,
synced from TradeAgent). Every brief tells its agent to read it and this file first. The project facts it relies on:

- **Gate:** `make check` (gofmt, vet, race-enabled tests).
- **Protected surfaces** (red-first test + one watched mutant): credential custody (upstream keys never leave the process and the OS vault); the loopback-only data plane and browser-origin rejection; model identity (no silent substitution); no replay of a request whose outcome is unknown.
- **Record:** `CHANGELOG.md`; nothing is called `PROVEN` without pinned real-client evidence.
