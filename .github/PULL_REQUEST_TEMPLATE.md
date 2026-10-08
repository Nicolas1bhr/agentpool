## What and why

Spec section(s) this implements or changes:

## Definition of done (spec v5.2 §15.1)

- [ ] Contract stated and tests added for behavior changes
- [ ] At least one deterministic negative/edge-case test per safety rule touched
- [ ] No unreviewed dependency, network target, persisted prompt or authority expansion
- [ ] Nothing marked `PROVEN` without pinned real-client evidence (otherwise `UNVERIFIED`)
- [ ] Capacity changes: pre-dispatch, concurrent reservation, stale observation and zero-hot-path checks
- [ ] Connector changes: certification record per docs/PROVIDER_POLICY.md
- [ ] `make check` passes; CHANGELOG.md updated under `Unreleased`
