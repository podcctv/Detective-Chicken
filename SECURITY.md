# Security Policy

Please report vulnerabilities privately through GitHub Security Advisories. Do not include real VPS IP addresses, enrollment tokens or agent private keys in public issues.

## Deployment notes

- Terminate TLS before exposing the API to the internet.
- Replace the demo tenant and add user authentication before a public deployment.
- Keep `/etc/detective-chicken/agent.json` readable only by root.
- Use a distributed nonce store when running more than one API replica.
- Pin and verify the Agent release checksum in the installer.
- Review IP.Check.Place's AGPL-3.0 license and every third-party data source's terms before commercial use.
