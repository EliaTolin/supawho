---
title: supawho find
lastUpdated: 2026-11-17
description: Reverse lookup for Supabase projects — paste a project ref and supawho tells you which saved account owns it, with organization, region, status and email.
---

The reverse of [`whoami`](/commands/whoami/): paste a Supabase **project ref** and `find` tells you which saved account owns it.

## Synopsis

```bash
supawho find <project-ref>
```

`<project-ref>` is the 20-character project reference (the subdomain of `<ref>.supabase.co`, also shown in your project's URL and settings).

## Example — found

```bash
supawho find abcdefghijklmnopqrst
```

```
Searching 3 account(s) for "abcdefghijklmnopqrst"...

  ✓ Found in account 'client-a'

  Project       storefront
  Reference     abcdefghijklmnopqrst
  Organization  Client A
  Region        eu-central-1
  Status        ACTIVE_HEALTHY
  Email         you@example.com
  Account       client-a

  Switch to it:  supawho use client-a
```

## Example — not found

```bash
supawho find zyxwvutsrqponmlkjihg
```

```
Searching 3 account(s) for "zyxwvutsrqponmlkjihg"...
No project matching "zyxwvutsrqponmlkjihg" in any saved account.
```

## Exit codes

| Situation | Exit code |
|---|---|
| Project found | `0` |
| Not found in any account | `1` |
| No argument given | `1` |

The non-zero exit on "not found" makes `find` usable in scripts.

## Behavior

- Searches each account's projects via the Management API (`GET /v1/projects`) and resolves the organization name.
- A revoked token skips that account and the search continues.
- `supawho locate` is an alias of `find`.

## See also

- [See the email and organizations behind each account with `supawho whoami`](/commands/whoami/)
- [Switch to the account that owns the project](/commands/use/)
- [Using `find` in scripts](/guides/ci-and-automation/#local-scripts)
