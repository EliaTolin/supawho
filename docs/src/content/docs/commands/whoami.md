---
title: supawho whoami
lastUpdated: 2026-11-21
description: Reveal the real email and organizations behind each saved Supabase account, so you can tell which account is which.
---

Map every saved account to its real **email** and **organizations**. With many accounts, `whoami` answers "which one is this?" at a glance.

## Synopsis

```bash
supawho whoami [name]
```

- With no argument, every saved account is looked up.
- With a `name`, only that account is looked up.

## Example

```bash
supawho whoami
```

```
ACCOUNT          EMAIL                     ORGANIZATION
client-a         you@example.com           Client A
side-project     you+dev@example.com       Personal, Labs
old-gig          (token is invalid or revoked)

Active: client-a (from the Supabase CLI login)
```

Look up a single account:

```bash
supawho whoami client-a
```

## Behavior

- Queries the Supabase Management API (`/v1/profile` and `/v1/organizations`) with each account's token.
- A revoked or invalid token is reported on its own row without failing the rest.
- If a token can't read its profile but can list organizations, the row shows them with `(unavailable)` as the email.
- Below the table, `Active:` names the saved account the Supabase CLI is using and where its token comes from: `SUPABASE_ACCESS_TOKEN` or the Supabase CLI login. A token that isn't saved in supawho is reported as such.
- If a [`.supawho`](/commands/env/#the-supawho-file) file applies to the current folder, `This folder:` shows the account it names and the file's path.
- `supawho who` is an alias of `whoami`.

## See also

- [Find which account owns a project with `supawho find`](/commands/find/)
- [How supawho talks to the Supabase Management API](/guides/how-it-works/#identity-lookups-use-the-management-api)
