---
title: supawho vs. supabase login vs. SUPABASE_ACCESS_TOKEN
description: An honest comparison of the ways to switch Supabase accounts in the CLI. What each one costs, where the token ends up, and when supawho is not the right tool.
lastUpdated: 2026-11-21
---

There are three ways to switch Supabase accounts in the CLI: run `supabase login` again, set `SUPABASE_ACCESS_TOKEN`, or use supawho. supawho is the only one that remembers several accounts for you, and it is built on the first: under the hood it runs `supabase login` with the saved token.

For step-by-step instructions on each method, see [How to use multiple Supabase accounts with the CLI](/guides/multiple-supabase-accounts/).

## At a glance

| | `supabase login` | `SUPABASE_ACCESS_TOKEN` | supawho |
|---|---|---|---|
| Made by | Supabase | Supabase | Community, open source (MIT) |
| Accounts remembered | One | None, you supply the token | As many as you save |
| Where your tokens are kept | The active one in the CLI's storage; the others wherever you put them | Shell profile, env file, or CI secret store | OS secret vault (Keychain, Secret Service, Credential Manager) |
| Switching | `supabase logout`, `supabase login`, paste | Change the variable | `supawho use <name>` |
| Affects | Every terminal | One command or one shell | Every terminal |
| CI and containers | Possible | Yes | No |
| Identify an account | No | No | `whoami` shows email and organizations |
| Find the account that owns a project | No | No | `find <project-ref>` |

## `supabase login`

The built-in way. The Supabase CLI stores the token in native credentials storage (the OS keyring) and, per its [reference](https://supabase.com/docs/reference/cli/supabase-login), falls back to a plain-text file at `~/.supabase/access-token` when that isn't available.

It is enough when you have one account, or two that you rarely swap. The cost shows up with more: the CLI remembers only the current token, so the others have to live somewhere, and that somewhere tends to be a notes file.

## `SUPABASE_ACCESS_TOKEN`

Setting the variable makes the CLI skip the stored login for that command or shell. It is the right tool for CI, where the token sits in the provider's encrypted secrets, and for running a single command as another account.

For everyday switching it has two drawbacks. The token has to be stored somewhere to be exported, usually a plain-text file. And a variable left set in one terminal quietly overrides whatever account you think you are logged in as.

## supawho

supawho stores one token per account in the OS secret vault and switches by running `supabase login` with the one you pick. On top of that it adds [`whoami`](/commands/whoami/), which maps saved accounts to emails and organizations, and [`find`](/commands/find/), which tells you which account owns a project ref.

What it doesn't change: the token is still handed to the Supabase CLI, which stores the active one exactly as `supabase login` does. The [Security](/guides/security/) page describes the whole path.

## When supawho is not the right tool

- **CI, containers and servers.** There is usually no OS secret vault to store tokens in. Use `SUPABASE_ACCESS_TOKEN` from your secret manager. See [CI and automation](/guides/ci-and-automation/).
- **One account with access to everything.** If a single Supabase account is a member of all the organizations you work with, there is nothing to switch.
- **Two accounts in parallel.** supawho changes the login for the whole machine. To act as two accounts at the same moment, set `SUPABASE_ACCESS_TOKEN` per terminal.
- **Machines where you can't install software.** `supabase login --token` works with what is already there.
