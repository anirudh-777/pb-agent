# PocketBase authentication

`pb-agent` uses a PocketBase **superuser impersonation token**. It does not ask
for or store a superuser password.

## Generate a token

1. Open the PocketBase Dashboard for the target instance.
2. Open **Collections** and select the `_superusers` system collection.
3. Select the dedicated superuser record that `pb-agent` should use.
4. Open **Impersonate**, choose the shortest practical duration, and generate
   the token.
5. Copy the generated token.

PocketBase describes these nonrenewable superuser impersonation tokens in its
[API keys documentation](https://pocketbase.io/docs/authentication/#api-keys).

You can print the same instructions from the CLI:

```sh
pb-agent --human connection token-help
```

## Store the token

Place the token in a temporary environment variable, then pipe it through
standard input:

```sh
printf '%s' "$POCKETBASE_SUPERUSER_TOKEN" | pb-agent connection add \
  --name local \
  --url http://127.0.0.1:8090 \
  --environment development \
  --token-stdin
```

The token is stored in the operating system credential manager. It is not
written to `pb-agent.yaml`, command history, output, plans, or audit records.

For CI, provide the token through the secret environment variable
`PB_AGENT_TOKEN`. Do not run `connection add` in CI.

## Security and revocation

- Create a dedicated PocketBase superuser for agent access.
- Use the shortest token duration that fits the workflow.
- Never pass a token as a command argument.
- Never commit a token or place it in `pb-agent.yaml`.
- Changing the dedicated superuser password invalidates tokens issued for that
  superuser.
- Changing the shared `_superusers` token secret invalidates issued tokens for
  every superuser.
