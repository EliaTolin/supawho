---
title: How to use multiple Supabase accounts with the CLI
description: The Supabase CLI holds one login at a time. Three ways to work with multiple Supabase accounts, with steps for each and a table to pick one.
lastUpdated: 2026-11-17
---

The Supabase CLI holds **one login at a time**. To work with multiple Supabase accounts you either log in again on every switch, pass a different access token per command with `SUPABASE_ACCESS_TOKEN`, or save the tokens once and switch by name with supawho.

This guide is about the accounts you use to sign in to the Supabase dashboard and CLI (personal, work, clients). It is not about the end users of your app, which are managed by Supabase Auth.

## Why it comes up

A Supabase account can belong to several organizations, and when one account is a member of every organization you work with, you don't need to switch at all. You need more than one account when clients own their organizations and give you a separate login, when work and personal accounts must stay apart, or when a project sits under an email address you don't normally use.

Every method below relies on a **personal access token**: one per account, created in the dashboard under **Account preferences → Access Tokens**. It starts with `sbp_` and is shown only once.

## Method 1: log in again each time

```bash
supabase logout
supabase login --token sbp_your_token
```

`supabase login` stores the token and every later command uses it. Nothing to install, but you need the right token at hand on every switch, and typing it inline leaves it in your shell history. `supabase login` without `--token` opens the browser instead, where you have to be signed in to the right account first.

**Fits:** two accounts and rare switches.

## Method 2: `SUPABASE_ACCESS_TOKEN` per command or per shell

The CLI reads the token from the `SUPABASE_ACCESS_TOKEN` environment variable, so you can skip login altogether:

```bash
SUPABASE_ACCESS_TOKEN=sbp_your_token supabase projects list
```

Or set it for a whole terminal session:

```bash
export SUPABASE_ACCESS_TOKEN=sbp_your_token
```

This is the method [Supabase documents for CI](https://supabase.com/docs/guides/cli/managing-environments), where the token is an encrypted secret. On a laptop the catch is where the token lives between sessions: in a `.env` or `.envrc` file it is plain text, and one `git add .` away from a commit.

**Fits:** CI pipelines, and one-off commands as another account without changing your login.

## Method 3: save accounts once, switch by name with supawho

[supawho](/guides/introduction/) keeps one token per account in your OS secret vault and runs `supabase login --token` for you:

```bash
supawho add              # once per account: name + token (hidden input)
supawho use client-a     # switch
supawho whoami           # which saved account is which email?
```

After `supawho use`, the Supabase CLI is logged in as that account until you switch again. [Install supawho](/guides/getting-started/) to try it.

**Fits:** three or more accounts, or switching several times a day.

## Which one should you use?

| | Log in again | `SUPABASE_ACCESS_TOKEN` | supawho |
|---|---|---|---|
| Extra install | None | None | One binary |
| Switch | Log out, log in, paste token | Set or change a variable | `supawho use <name>` |
| Where tokens wait between switches | Wherever you keep them | Your shell, an env file, or CI secrets | OS secret vault |
| Scope of a switch | Every terminal | One command or one shell | Every terminal |
| Works in CI | Possible | Yes, the documented way | No, use the variable |
| Tells you which account a token belongs to | No | No | `supawho whoami` |

A longer comparison, including when supawho is the wrong tool, is in [supawho vs. alternatives](/guides/alternatives/).

## Common questions

### Can I be logged in to two Supabase accounts at the same time?

Not through `supabase login`: the CLI stores one token. To act as two accounts side by side, set `SUPABASE_ACCESS_TOKEN` to a different token in each terminal.

### Does switching account change the linked project?

No. `supabase link` stores the project ref inside your project folder, and it stays as it is. After switching, the CLI talks to that project with the new account, so commands fail if that account has no access to it. [`supawho find <project-ref>`](/commands/find/) tells you which saved account owns a project.

### Which account is the CLI using right now?

`supabase projects list` shows the projects the active login can see. To map your saved tokens to emails and organizations, run [`supawho whoami`](/commands/whoami/).
