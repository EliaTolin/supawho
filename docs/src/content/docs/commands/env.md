---
title: supawho env
description: Print the shell command that sets SUPABASE_ACCESS_TOKEN to a saved account's token, for eval or direnv. Per-shell and per-folder accounts without touching the global login.
lastUpdated: 2026-11-21
---

Print the line that sets `SUPABASE_ACCESS_TOKEN` to a saved account's token, for your shell to evaluate. The Supabase CLI gives that variable precedence over its saved login, so the account applies to **this shell only**, or to **one folder** with direnv, while [`use`](/commands/use/) changes the login everywhere.

## Synopsis

```bash
eval "$(supawho env [name] [--shell sh|bash|zsh|fish|powershell])"
```

- With a `name`, the token of that account is used.
- Without a name, the account named in the nearest `.supawho` file is used, looking in the current folder and then in each parent.
- `--shell` picks the syntax. The default, `sh`, also works for bash and zsh. `pwsh` is accepted for PowerShell.

## Examples

Use an account in the current shell only:

```bash
eval "$(supawho env client-a)"
supabase projects list   # runs as client-a, in this shell only
```

fish and PowerShell:

```bash
supawho env client-a --shell fish | source
```

```powershell
supawho env client-a --shell powershell | Invoke-Expression
```

One account per project folder, with a committed `.supawho` file and [direnv](https://direnv.net/):

```bash
echo client-a > .supawho
echo 'watch_file .supawho
eval "$(supawho env)"' > .envrc
direnv allow
```

The full setup is in [Per-project accounts with direnv](/guides/per-project-accounts/).

## The `.supawho` file

`.supawho` holds only an account **name**, never a token, so it is safe to commit. The first line that is neither blank nor a `#` comment is used:

```text
# Supabase account for this repo
client-a
```

The name refers to an account saved on your machine, so each teammate saves their own token under the same name.

## Behavior

- Output on success is exactly one line, for example `export SUPABASE_ACCESS_TOKEN='sbp_…'`. The token is single-quoted for the chosen shell.
- Every error goes to stderr and nothing is printed on stdout, so a failed `eval` never runs anything. The exit code is `1`.
- When stdout is a terminal, `env` refuses to print the token, so it doesn't end up on screen or in the scrollback, and shows the `eval` command to run instead.
- The token is read from the OS secret store each time. supawho never writes it to a file.
- The global Supabase CLI login is not changed. `unset SUPABASE_ACCESS_TOKEN` (fish: `set -e SUPABASE_ACCESS_TOKEN`) goes back to it.

## Good to know

The variable is inherited by every program started from that shell, such as scripts, `npm run` and editors. That is how the Supabase CLI finds it, and also why `env` is opt-in. See [Security](/guides/security/#per-shell-accounts-with-supawho-env).

## Exit codes

| Situation | Exit code |
|---|---|
| The export line was printed | `0` |
| Unknown account, no `.supawho` found, unknown shell, or stdout is a terminal | `1` |

## See also

- [Per-project accounts with direnv](/guides/per-project-accounts/) · [`use`](/commands/use/) · [`whoami`](/commands/whoami/)
