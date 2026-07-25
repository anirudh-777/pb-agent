# pb-agent

`pb-agent` is a safe, agent-first gateway for inspecting, testing, and changing
[PocketBase](https://pocketbase.io/) projects. It exposes a JSON-first CLI,
immutable plan/apply mutations, environment-aware policy, secret redaction, and
local audit evidence.

> [!IMPORTANT]
> The project is pre-release. Its compatibility baseline is PocketBase 0.39.8.

## Install

Install the latest GitHub release on macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/anirudh-777/pb-agent/main/install.sh | sh
```

The installer detects the current platform, verifies the release archive
against its published SHA-256 checksum, and installs to `/usr/local/bin` when
writable or `~/.local/bin` otherwise.

To pin a version or choose the destination:

```sh
PB_AGENT_VERSION=v0.1.0-rc.4 \
PB_AGENT_INSTALL_DIR="$HOME/.local/bin" \
  sh install.sh
```

You can also build from source:

```sh
go install github.com/anirudh-777/pb-agent/cmd/pb-agent@latest
```

Building from source requires Go 1.26.5 or newer.

Planned release channels:

```sh
npm install -g pb-agent
brew install anirudh-777/tap/pb-agent
scoop install pb-agent
```

Install the agent skill separately in supported coding agents:

```sh
npx skills add anirudh-777/pb-agent --skill pb-agent -g -y
```

The skill teaches an agent the safe workflow; the installer above provides the
`pb-agent` executable it invokes.

## Quick start

```sh
pb-agent init --name local --url http://127.0.0.1:8090 --environment development
pb-agent --human connection token-help
printf '%s' "$POCKETBASE_SUPERUSER_TOKEN" | pb-agent connection add \
  --name local \
  --url http://127.0.0.1:8090 \
  --environment development \
  --token-stdin

pb-agent --connection local doctor
pb-agent --connection local records list --collection posts
```

To generate the token, open the PocketBase Dashboard, select
**Collections → `_superusers` → your dedicated superuser → Impersonate**, and
choose the shortest practical duration. See
[PocketBase authentication](docs/AUTHENTICATION.md) for complete generation,
storage, CI, and revocation instructions.

Mutations always use an immutable plan:

```sh
printf '{"title":"Hello"}' > /tmp/post.json
pb-agent --connection local plan record-create \
  --collection posts \
  --data-file /tmp/post.json

pb-agent apply --plan pln_REVIEWED_PLAN_ID
```

Every command emits a versioned JSON envelope. Use `pb-agent capabilities` for
the machine-readable capability catalog.

## Safety model

- Credentials come from the OS keychain or environment, never command arguments.
- Stored mutation payloads are encrypted locally.
- Plans expire after 15 minutes and are bound to one PocketBase instance.
- Updates and deletes abort when the target changed after planning.
- Staging and production mutations need a scoped, short-lived access grant.
- Audits contain metadata and hashes, not record values or secrets.
- Record contents are untrusted data and never affect policy.

See [authentication](docs/AUTHENTICATION.md),
[threat model](docs/THREAT_MODEL.md), [compatibility](docs/COMPATIBILITY.md),
and [security policy](SECURITY.md).

## Current scope

Implemented reads:

- Health and capability probes
- Collections and records
- Guest, configured-user, and supplied-user list-rule testing
- File downloads with refusal to overwrite local files
- Logs and backups

Implemented plan/apply mutations:

- Record create, update, upsert, delete, and batch
- Collection create, update, and delete
- Backup create, restore, and delete

Settings, mail, realtime, raw HTTP, SQL, PocketBase process management, and MCP
are intentionally deferred.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). The project is licensed under
[Apache-2.0](LICENSE).
