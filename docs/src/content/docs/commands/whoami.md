---
title: supawho whoami
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
```

Look up a single account:

```bash
supawho whoami client-a
```

## Behavior

- Queries the Supabase Management API (`/v1/profile` and `/v1/organizations`) with each account's token.
- A revoked or invalid token is reported on its own row without failing the rest.

## See also

- [`find`](/supawho/commands/find/) — the reverse lookup: which account owns a project?
- [How it works](/supawho/guides/how-it-works/)
