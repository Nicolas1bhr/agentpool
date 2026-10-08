# Security policy

AgentPool will hold credentials for AI accounts and sit on the path of every model request, so
security reports are taken seriously even at the skeleton stage.

## Reporting

Use GitHub's private reporting: **Security → Report a vulnerability** on this repository. Please do
not open a public issue for anything exploitable. You should get an acknowledgement within 7 days.

## Supported versions

Pre-1.0: only the latest release receives fixes.

## Scope

- Anything that lets a non-owner reach the control plane, read credentials, or select routes.
- Data-plane exposure beyond loopback, or acceptance of browser-origin requests.
- Credential leakage into logs, OpenCode config, prompts or worker environments.
- Request smuggling, redirect following, SSRF or destination override toward upstreams.
- Silent model substitution, or replay of a request whose outcome is unknown.

The threat model is spec v5.2 §9.
