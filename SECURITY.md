# Security policy

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability. Use GitHub private
vulnerability reporting for `anirudh-777/pb-agent`. Include the affected
version, reproduction steps, impact, and any suggested mitigation.

You should receive an acknowledgement within seven days. A coordinated
disclosure date will be agreed after validation.

## Supported versions

Until v1.0, only the latest released `pb-agent` version receives security fixes.

## Security promises

- PocketBase credentials are never accepted as command-line values.
- Mutation plans are encrypted at rest.
- Output and audit paths redact known secret fields.
- Production mutations fail closed without an unexpired scoped grant.

These promises are release-blocking and covered by automated tests.
