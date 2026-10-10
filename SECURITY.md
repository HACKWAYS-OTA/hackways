# Security Policy

## Supported Versions

Only the latest release on the `main` branch receives security updates.

| Version | Supported          |
| ------- | ------------------ |
| 1.x / `main` | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

The Hackways team takes the security of our platform, user sessions, attendee registrations, and organizer credentials seriously.

If you discover a security vulnerability within Hackways:

1. **Do NOT open a public GitHub issue.**
2. Send an email to the security team at **security@hackways.local** or utilize GitHub's private vulnerability reporting feature on this repository.
3. Include the following details in your report:
   - Type of vulnerability (e.g. IDOR, SSRF, SQL Injection, Auth bypass, Race condition overbooking)
   - Step-by-step reproduction instructions or proof-of-concept
   - Potential impact of the issue
   - Any suggested mitigations

## Response Timeline

- **Initial Response:** Within 48 hours of report submission
- **Status Update:** Within 5 business days
- **Fix & Public Disclosure:** Coordinated disclosure upon validation and rollout of patch to `main`

## Security Principles in Hackways

- **Zero Secrets in Code:** Credentials and session secrets are managed strictly through environment variables.
- **Server-Side Authorization:** All protected endpoints enforce session verification on the server.
- **SQL Injection Prevention:** All SQL queries are generated via `sqlc` with parameterized bindings; raw string interpolation is prohibited.
- **Password Security:** Hashes stored using `bcrypt` with cost factor >= 12.
- **Session Security:** `HttpOnly`, `SameSite=Lax`, and `Secure` (production) cookies with token hashes in PostgreSQL.
