# Security Policy

Please report vulnerabilities privately through GitHub Security Advisories. Do not include real VPS IP addresses, enrollment tokens or agent private keys in public issues.

## Deployment notes

- Terminate TLS before exposing the API to the internet.
- Keep public registration disabled unless new users are expected; the first registered account is the initial administrator.
- Back up `/data/state.json`, restrict its filesystem permissions, and migrate to PostgreSQL plus a distributed session/nonce store before running multiple API replicas.
- Keep `/etc/detective-chicken/agent.json` readable only by root.
- Use a distributed nonce store when running more than one API replica.
- Publish checksums with Agent releases and pin them in the generated installer before using an external binary mirror.
- Review IP.Check.Place's AGPL-3.0 license and every third-party data source's terms before commercial use.
