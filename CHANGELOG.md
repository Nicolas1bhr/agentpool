# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html). The architecture version (currently
5.2.0) is tracked separately in the spec header.

## [Unreleased]

### Added
- Orchestration standard for delegated and multi-agent work (`docs/ORCHESTRATION-STANDARD.md`) and a `CLAUDE.md`
  naming this project's gate, protected surfaces and record.
- Independent research paper (2026-10-08) assessing feasibility, continuity limits and provider
  terms for inference capacity pooling (`docs/architecture/`).

## [0.1.0] - 2026-10-08

### Added
- Architecture spec v5.2 (current) and its eight predecessors under `docs/architecture/`.
- `agentpool` binary skeleton: `serve`, `version`, `help`.
- Data-plane skeleton: refuses any non-loopback bind, rejects browser-origin requests, serves
  `/healthz`, and answers every `/v1/*` inference path with an explicit `501 not implemented`.
- Provider and connector policy (`docs/PROVIDER_POLICY.md`).
- Illustrative, unverified OpenCode V2 provider config (`examples/opencode.jsonc`).
- CI (gofmt, vet, race-enabled tests on Linux/macOS, tests on Windows) and a tag-driven release workflow producing
  binaries for linux/darwin/windows × amd64/arm64 with SHA-256 checksums.

[Unreleased]: https://github.com/Nicolas1bhr/agentpool/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Nicolas1bhr/agentpool/releases/tag/v0.1.0
