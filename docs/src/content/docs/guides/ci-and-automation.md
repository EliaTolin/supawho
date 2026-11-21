---
title: Using supawho in scripts, and what to use in CI
description: How to call supawho from local scripts, which exit codes to rely on, and why CI pipelines should use SUPABASE_ACCESS_TOKEN instead.
lastUpdated: 2026-11-17
---

supawho is built for the machine you work on. It works well in local scripts; in CI, use the Supabase CLI's own `SUPABASE_ACCESS_TOKEN` variable instead.

## In CI, use `SUPABASE_ACCESS_TOKEN`

CI runners and containers usually have no OS secret vault, so supawho has nowhere to keep tokens, and a pipeline only ever needs one account anyway. Store a personal access token as an encrypted secret and expose it to the Supabase CLI:

```yaml
# GitHub Actions
env:
  SUPABASE_ACCESS_TOKEN: ${{ secrets.SUPABASE_ACCESS_TOKEN }}
  SUPABASE_DB_PASSWORD: ${{ secrets.SUPABASE_DB_PASSWORD }}
  SUPABASE_PROJECT_ID: ${{ secrets.SUPABASE_PROJECT_ID }}
steps:
  - uses: supabase/setup-cli@v1
  - run: supabase link --project-ref "$SUPABASE_PROJECT_ID"
  - run: supabase db push
```

This is the setup described in the Supabase guide on [managing environments](https://supabase.com/docs/guides/cli/managing-environments). Create a dedicated token for CI so you can revoke it without touching the ones on your laptop.

## Local scripts

Every command except the interactive picker runs without prompts when it gets its arguments, so supawho fits in shell scripts, Makefiles and git hooks.

Switch account before a command:

```bash
supawho use client-a && supabase db push
```

Stop if the active project belongs to an account you haven't saved. [`supawho find`](/commands/find/) exits with `1` when no saved account owns the ref:

```bash
ref="abcdefghijklmnopqrst"
if ! supawho find "$ref" > /dev/null; then
  echo "No saved Supabase account can access $ref" >&2
  exit 1
fi
```

### Exit codes

| Result | Exit code |
|---|---|
| Command succeeded | `0` |
| Handled error: account not found, invalid name, project not found, unknown command | `1` |
| Secret vault, Supabase CLI or network failure | `1`, with a `supawho: …` message |

### Update notice

The daily update check runs only when output goes to a terminal, and it prints to standard error, so pipes and redirected output are never affected. To switch it off everywhere:

```bash
export SUPAWHO_NO_UPDATE_CHECK=1
```

## Headless machines and SSH

supawho needs an unlocked secret store:

- **macOS over SSH**: the login keychain is locked in sessions without a graphical login. Unlock it with `security unlock-keychain ~/Library/Keychains/login.keychain-db`.
- **Linux servers, SSH and WSL**: there is usually no Secret Service without a desktop session. Store the tokens in pass instead, with `export SUPAWHO_BACKEND=pass`: see [Headless Linux](/guides/headless-linux/). For unattended jobs, `SUPABASE_ACCESS_TOKEN` is still the simpler choice.

See [Troubleshooting](/guides/troubleshooting/) for the exact error messages.

## Keeping a token out of your history in scripts

`supawho add <name> <token>` is the scriptable form of [`add`](/commands/add/). Read the token from a variable or a secret manager rather than typing it on the command line:

```bash
supawho add client-a "$CLIENT_A_TOKEN"
```
