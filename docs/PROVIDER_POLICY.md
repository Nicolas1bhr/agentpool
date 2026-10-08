# Provider and connector policy

AgentPool's connectors are modular (spec v5.2 §4.2–4.3): API-key, OAuth, subscription-backed,
enterprise and local. **Absence of a connector is a backlog item, not an architectural
prohibition.** But a connector is only merged and certified when it meets every rule below.

## 1. One owner, their own accounts

AgentPool pools credentials that **one owner** legitimately holds, for that owner's own workloads
(spec v5.2 document control: "one authorized local owner"; §13.3 `mode: local-owner`). It does not
support, and connectors must not enable:

- pooling or sharing one person's account among several people;
- reselling or brokering access to accounts;
- presenting someone else's credentials as your own.

## 2. Documented flows the provider permits

A connector must use an authentication and inference flow that its provider **documents** and
whose **current terms allow** that use from a third-party tool. Specifically, a connector must not:

- impersonate another vendor's first-party client (spoofed client IDs, user agents, headers or
  fingerprints);
- extract tokens from another application's private storage;
- circumvent rate limits, usage caps, abuse detection or plan restrictions;
- automate a consumer web UI where the provider does not offer a programmatic flow for it.

Provider terms change often, and subscription plans in particular differ in whether they can be
used outside the provider's own apps. A connector that was permitted can become prohibited; it is
then disabled until the terms allow it again.

## 3. Certification record

Every connector PR includes a certification record (kept with the connector under `compatibility/`)
stating:

| Field | Example |
|---|---|
| Provider and credential mode | `<provider>` · subscription OAuth |
| Terms relied on | URL(s) and section of the provider's current terms / developer docs |
| Terms checked on | `2026-10-08` — re-checked at every release |
| Flow | Documented endpoint(s) and auth flow used |
| Capabilities proven | Stream, tools, reasoning, usage reporting — each `PROVEN` with pinned evidence or `UNVERIFIED` |
| Capacity model | What the provider reports about limits; unknowns stay `UNKNOWN` |

A connector without a current terms check is not certified, and the automatic selector never
chooses an uncertified route (spec §4.5).

## 4. Owner responsibility

Running AgentPool does not change your agreement with any provider. You remain responsible for
using each account within its terms.
