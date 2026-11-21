---
title: 'Security: how supawho stores Supabase access tokens'
description: Where supawho keeps your Supabase access tokens, what the OS secret vault protects, what happens when you switch accounts, and what is written to disk.
lastUpdated: 2026-11-21
---

supawho **never writes an access token to a file**. Each token is stored in your operating system's native secret store, the same encrypted vault the OS uses for passwords and certificates:

- **macOS**: Keychain
- **Linux**: Secret Service (GNOME Keyring / KWallet, via `libsecret`)
- **Windows**: Credential Manager

All entries live under the service name `supabase-cli-accounts`, so you can inspect or delete them with your OS tools (Keychain Access, Seahorse, Credential Manager).

## What the vault protects

| | OS secret vault | Plain-text file (`.env`, notes, dotfiles) |
|---|:---:|:---:|
| Encrypted at rest | ✅ | ❌ |
| Unlocked by your system login | ✅ | ❌ |
| Kept out of project folders, so `git add .` can't commit it | ✅ | ❌ |

The vault protects tokens at rest and from other user accounts on the machine. It is not a sandbox: software that already runs as you can ask the vault for the same entries, as with any command-line credential helper.

## What happens when you switch accounts

[`supawho use <name>`](/commands/use/) reads the token from the vault and runs `supabase login --token <token>`. Two consequences are worth knowing:

- **The token is passed as a command-line argument.** For the fraction of a second that `supabase login` runs, the token is visible in the process list to other processes running as your user (and to administrators). On a shared machine, treat that as exposure.
- **The Supabase CLI keeps its own copy of the active token.** According to the [Supabase CLI reference](https://supabase.com/docs/reference/cli/supabase-login), the token is stored in native credentials storage, and "if native credentials storage is unavailable, it will be written to a plain text file at `~/.supabase/access-token`". That file, when it exists, is written by the Supabase CLI, not by supawho.

## Per-shell accounts with `supawho env`

[`supawho env`](/commands/env/) prints `export SUPABASE_ACCESS_TOKEN=…` for your shell to evaluate, usually from direnv. The token is read from the vault each time and never written to a file, but while it is set it is in the shell's environment and inherited by every program started from that shell, not only the Supabase CLI. `env` refuses to print to a terminal, so the token doesn't land in your scrollback, and it prints every error to stderr, so a failed `eval` runs nothing. A `.supawho` file holds only an account name and is safe to commit. See [Per-project accounts](/guides/per-project-accounts/#security-trade-offs).

## Add tokens without leaving them in your shell history

Run [`supawho add`](/commands/add/) with no arguments. It asks for the name and reads the token with hidden input, so the token never appears on screen or in your shell history:

```bash
supawho add
```

The one-line form `supawho add <name> <token>` is handy in scripts, but your shell records the whole command, token included, in its history file. If you used it, remove that line from your history or rotate the token.

## What supawho writes to disk

One file, with no secrets in it: `last-update-check` in your user cache directory (`~/Library/Caches/supawho/` on macOS, `~/.cache/supawho/` on Linux, `%LocalAppData%\supawho\` on Windows). It holds the timestamp of the last update check. supawho has no config file.

## Network requests

- **Supabase Management API** (`api.supabase.com`): only when you run [`whoami`](/commands/whoami/) or [`find`](/commands/find/), using the token of each saved account.
- **GitHub API** (`api.github.com`): the update check, at most once a day and only in an interactive terminal, plus [`supawho upgrade`](/commands/upgrade/). Set `SUPAWHO_NO_UPDATE_CHECK=1` to turn the check off.

supawho sends no telemetry.

## Updates are verified

`supawho upgrade` downloads the release archive over HTTPS and checks its SHA-256 against the published `checksums.txt` before replacing the binary. A mismatch aborts the update and leaves the current binary untouched.

## Removing and revoking tokens

[`supawho remove <name>`](/commands/remove/) deletes the token from the vault. It does not revoke it: to invalidate a token, delete it in the [Supabase dashboard](https://supabase.com/dashboard/account/tokens).

## Requirements

- macOS, Linux, or Windows: a single static binary, no runtime to install
- On Linux, a Secret Service provider must be running (GNOME Keyring, KWallet, …)
- The [Supabase CLI](https://supabase.com/docs/guides/cli), available in your `PATH`
