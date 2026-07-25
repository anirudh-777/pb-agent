# Threat model

## Assets

- PocketBase superuser and impersonation tokens
- Application records, files, schema, logs, and backups
- Human approval intent
- Plan integrity and audit evidence

## Trust model

The local user and installed `pb-agent` binary are trusted. PocketBase record
content, filenames, API error details, agent-generated arguments, and remote
network responses are untrusted.

## Primary threats and controls

| Threat | Control |
| --- | --- |
| Prompt injection in records | Data never enters policy evaluation; skills label it untrusted |
| Credential disclosure | Keyring/env input, no token flags, recursive output redaction |
| Agent changes arguments after approval | Encrypted payload, signed metadata, and request hash |
| Concurrent modification | Precondition hash fetched again immediately before apply |
| Plan replay, concurrent apply, or retargeting | Apply lock, one-use marker, expiry, environment and connection binding |
| Accidental production write | Read-only default and interactive scoped grant |
| Excessive context or data extraction | Page-size bounds and no unlimited reads |
| Audit leaks record values | Metadata-only JSONL with field names and hashes |
| Malicious TLS endpoint | TLS verification by default; insecure TLS limited to declared dev/test |

## Explicitly out of scope for v0.1

- A compromised local operating system or `pb-agent` binary
- Malicious custom PocketBase server extensions
- Remote MCP transports
- Raw SQL and arbitrary HTTP
- Automated production writes in CI
