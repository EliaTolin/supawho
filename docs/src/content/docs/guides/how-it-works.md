---
title: How it works
description: Understand supawho's model — how it stores Supabase access tokens in the OS secret vault and switches accounts by driving the Supabase CLI.
---

supawho is a thin, secure layer on top of the official Supabase CLI. Here's the whole model.

## Accounts are name → token

Each account you save is just a **name** you choose (`work`, `client-a`) mapped to a Supabase **access token** (`sbp_…`). The token is stored in your OS secret vault; the list of names is kept alongside it.

```
work          → sbp_••••••••  (in the OS secret vault)
client-a      → sbp_••••••••
side-project  → sbp_••••••••
```

## Switching = driving the Supabase CLI

When you run [`supawho use work`](/supawho/commands/use/), supawho reads that account's token from the vault and runs:

```bash
supabase login --token <token>
```

So supawho never reimplements authentication — it drives the Supabase CLI you already have. That's why the [Supabase CLI](https://supabase.com/docs/guides/cli) must be installed and on your `PATH`.

## Identity lookups use the Management API

[`whoami`](/supawho/commands/whoami/) and [`find`](/supawho/commands/find/) call the **Supabase Management API** with each account's token:

- `GET /v1/profile` → the account's email
- `GET /v1/organizations` → organization names
- `GET /v1/projects` → the projects a token can access

These are read-only calls, used only when you run those commands.

## Where your data lives

| Data | Where |
|---|---|
| Access tokens | OS secret vault (encrypted) |
| Account names | OS secret vault |
| Anything on disk | Nothing — no config files, no plain-text tokens |

More on the security model in [Security](/supawho/guides/security/).
