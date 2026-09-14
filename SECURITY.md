# Security Policy

## Reporting a Vulnerability

Please **do not open a public issue** for security vulnerabilities.

Use GitHub's private vulnerability reporting instead:

1. Go to the repository's **Security** tab
2. Click **Report a vulnerability**
3. Describe the issue, affected versions, and reproduction steps

You can expect an initial response within **7 days**. Once a fix is released,
the report will be disclosed via a GitHub Security Advisory with credit to the
reporter (unless you prefer to remain anonymous).

Direct link: https://github.com/svrforum/NginxProxyGuard/security/advisories/new

## Supported Versions

Only the **latest release** receives security fixes. NginxProxyGuard follows a
rolling release model — please upgrade to the most recent version before
reporting issues against older releases.

| Version | Supported |
|---------|-----------|
| Latest release | ✅ |
| Older releases | ❌ (upgrade first) |

## Scope

In scope:

- The API server (`api/`), admin UI (`ui/`), and nginx/ModSecurity layer (`nginx/`)
- The official Docker images (`svrforum/nginxproxyguard-{api,ui,nginx}`)
- Default `docker-compose.yml` deployment configuration

Out of scope:

- Vulnerabilities in upstream dependencies that are already public (report
  upstream; we track them via Dependabot)
- Issues requiring a fully compromised host or Docker daemon
- Self-modified forks or non-default deployment configurations

## Hardening Guidance for Operators

- **Do not expose the admin UI (port 81) to the internet.** Keep it on your
  LAN/VPN, or front it with its own proxy host protected by access lists and
  2FA.
- Complete the initial setup immediately — since v2.24.6 the server blocks all
  protected APIs until the default credentials are changed.
- Give each operator their own account with the least-privilege built-in role (Viewer or Operator, v2.34.0+) instead of sharing the admin login, and enable 2FA (TOTP) on every local account.
- Use API tokens with the minimum permission scopes needed.
- **If NPG sits behind Cloudflare, another CDN, or an upstream reverse proxy, set Settings → Trusted Proxies to exactly that proxy's address ranges (v2.51.0+).** Without it every request appears to come from the proxy, so IP bans, access lists, geo-blocking and fail2ban act on the proxy's address (one ban can block every visitor) and the global fail2ban jail stays inert. Trusting ranges wider than the proxy lets any visitor spoof their address via `X-Forwarded-For` and bypass those controls.
