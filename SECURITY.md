# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in MeshDNS, please report it privately.

**Do not open a public issue.** Instead, email:

📧 `security@meshdns.dev`

We'll acknowledge your report within **48 hours** and provide a timeline
for resolution within **5 business days**.

## Supported versions

| Version | Supported |
|---------|-----------|
| latest `main` branch | ✅ |
| All older versions | ❌ |

## What to expect

1. You submit a report via email
2. We confirm receipt within 48 hours
3. We investigate and determine impact
4. We develop and test a fix
5. We release the fix and publish an advisory
6. We credit you in the advisory (unless you prefer anonymity)

## Disclosure policy

We follow a **90-day disclosure** window:
- Fix released within 90 days of the initial report
- Advisory published alongside the fix
- Extensions granted by mutual agreement if more time is needed

## Scope

Security issues in scope include:
- Authentication bypass (write_key)
- SQL injection (store layer)
- Server-side request forgery via health probes
- Information disclosure (write_key exposure in logs/events)
- Denial of service (resource exhaustion)

## Out of scope

- Issues requiring physical access to the host
- Theoretical attacks with no practical exploit
- Social engineering
- Issues already publicly disclosed