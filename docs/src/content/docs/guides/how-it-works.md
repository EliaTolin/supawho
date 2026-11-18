---
title: How supawho works with the Supabase CLI
description: The model behind supawho. It stores Supabase access tokens in the OS secret vault and switches accounts by running supabase login for you.
lastUpdated: 2026-11-17
---

supawho is a thin layer on top of the official Supabase CLI. Here's the whole model.

## Accounts are name → token

Each account you save is a **name** you choose (`work`, `client-a`) mapped to a Supabase **access token** (`sbp_…`). The token is stored in your OS secret vault; the list of names is kept alongside it.

```
work          → sbp_••••••••  (in the OS secret vault)
client-a      → sbp_••••••••
side-project  → sbp_••••••••
```

## Switching = driving the Supabase CLI

When you run [`supawho use work`](/commands/use/), supawho reads that account's token from the vault and runs:

```bash
supabase login --token <token>
```

supawho doesn't reimplement authentication: it drives the Supabase CLI you already have, which is why the [Supabase CLI](https://supabase.com/docs/guides/cli) must be installed and on your `PATH`. After the switch, every `supabase` command runs as that account until you switch again.

The Supabase CLI keeps its own copy of the active token. See [Security](/guides/security/#what-happens-when-you-switch-accounts) for where it goes.

## Identity lookups use the Management API

[`whoami`](/commands/whoami/) and [`find`](/commands/find/) call the **Supabase Management API** with each account's token:

- `GET /v1/profile`: the account's email
- `GET /v1/organizations`: organization names
- `GET /v1/projects`: the projects a token can access

These are read-only calls, made only when you run those commands.

## Where your data lives

| Data | Where |
|---|---|
| Access tokens | OS secret vault (encrypted), service `supabase-cli-accounts` |
| Account names | OS secret vault, same service |
| Config files | None |
| Other files | A timestamp of the last update check in your user cache directory. No tokens. |

The full picture, including what the Supabase CLI stores, is in [Security](/guides/security/).
