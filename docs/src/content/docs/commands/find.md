---
title: supawho find
description: Reverse lookup for Supabase projects — paste a project ref and supawho tells you which saved account owns it, with organization, region, status and email.
---

The reverse of [`whoami`](/supawho/commands/whoami/): paste a Supabase **project ref** and `find` tells you which saved account owns it.

## Synopsis

```bash
supawho find <project-ref>
```

`<project-ref>` is the 20-character project reference (the subdomain of `<ref>.supabase.co`, also shown in your project's URL and settings).

## Example — found

```bash
supawho find sjslmvggunljemhwbkgm
```

```
  ✓ Found in account 'aurora'

  Project       almanaccolo
  Reference     sjslmvggunljemhwbkgm
  Organization  AuroraDigital
  Region        eu-central-1
  Status        ACTIVE_HEALTHY
  Email         mail@auroradigital.it
  Account       aurora

  Switch to it:  supawho use aurora
```

## Example — not found

```bash
supawho find abcdefghijklmnopqrst
```

```
No project matching "abcdefghijklmnopqrst" in any saved account.
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

## See also

- [`whoami`](/supawho/commands/whoami/) · [`use`](/supawho/commands/use/)
