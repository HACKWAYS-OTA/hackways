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
   - Type of vulnerability (e.g., IDOR, SSRF, SQL Injection, Auth bypass, Race condition overbooking)
   - Step-by-step reproduction instructions or proof-of-concept
   - Potential impact of the issue
   - Any suggested mitigations

## Security Architecture Reference

For detailed threat models, row-level concurrency lock verification, cookie scoping, and secrets management, consult our comprehensive documentation in:
- **[docs/SECURITY.md](../docs/SECURITY.md)**
- **[docs/ARCHITECTURE_PLAN.md](../docs/ARCHITECTURE_PLAN.md)**
