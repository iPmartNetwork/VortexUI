# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.0.x   | ✅        |
| < 1.0   | ❌        |

We support the latest minor release. Please upgrade before reporting issues.

## Reporting a Vulnerability

**Please do not open a public issue for security vulnerabilities.**

Instead, report privately through GitHub:

1. Go to the repository's **Security** tab → **Report a vulnerability**
   ([Private vulnerability reporting](https://github.com/iPmartNetwork/VortexUI/security/advisories/new)).
2. Describe the issue, affected version/commit, and reproduction steps.

If private reporting is unavailable, contact the maintainers via the
[iPmartNetwork](https://github.com/iPmartNetwork) organization.

### What to expect

- **Acknowledgement** within 72 hours.
- An assessment and, for confirmed issues, a fix timeline.
- Coordinated disclosure once a patch is available — we'll credit you unless you
  prefer to remain anonymous.

## Scope

Because VortexUI manages proxy infrastructure, we are especially interested in:

- Authentication / authorization bypass (JWT, RBAC, 2FA, API tokens)
- Panel ↔ node mTLS or gRPC weaknesses
- Subscription token leakage or enumeration
- Privilege escalation between admins/resellers
- Injection (SQL, command, template) and SSRF in routing/geo updates

Out of scope: issues requiring a compromised host/root, and findings in
third-party cores (Xray-core, sing-box) — report those upstream.
