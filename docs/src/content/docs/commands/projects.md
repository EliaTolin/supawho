---
title: supawho projects
description: List the Supabase projects of every saved account, or of one account: name, reference, organization, region and status. JSON output for scripts.
lastUpdated: 2026-11-18
---

List the Supabase **projects** each saved account can reach. Where [`find`](/commands/find/) answers "which account owns this project?", `projects` answers "what is in this account?".

## Synopsis

```bash
supawho projects [name] [--json]
```

- With no argument, every saved account is listed.
- With a `name`, only that account is listed.
- `--json` prints machine-readable output.

## Example

```bash
supawho projects
```

```
  client-a (2)
  PROJECT    REFERENCE             ORGANIZATION  REGION        STATUS
  api        abcdefghijklmnopqrst  Client A      eu-central-1  ACTIVE_HEALTHY
  api-dev    tsrqponmlkjihgfedcba  Client A      eu-central-1  INACTIVE

  side-project (1)
  PROJECT    REFERENCE             ORGANIZATION  REGION        STATUS
  blog       qwertyuiopasdfghjklz  Personal      us-east-1     ACTIVE_HEALTHY

  old-gig
  ! skipped (token is invalid or revoked)
```

List a single account:

```bash
supawho projects client-a
```

## JSON output

```bash
supawho projects client-a --json
```

```json
{
  "accounts": [
    {
      "name": "client-a",
      "projects": [
        {
          "ref": "abcdefghijklmnopqrst",
          "name": "api",
          "organization": "Client A",
          "region": "eu-central-1",
          "status": "ACTIVE_HEALTHY"
        }
      ]
    }
  ]
}
```

An account that could not be listed has an empty `projects` array and an `error` field. A usage or setup error prints `{"error": "..."}` instead.

## Exit codes

| Situation | Exit code |
|---|---|
| Every account listed (even with no projects) | `0` |
| No accounts saved | `0` |
| At least one account could not be listed | `1` |
| Account `name` not found, or bad arguments | `1` |

The non-zero exit on a failed account means a script never mistakes an incomplete listing for a complete one.

## Behavior

- Reads each account's projects via the Management API (`GET /v1/projects`) and resolves the organization name.
- Accounts are queried in parallel and shown in saved order; projects are sorted by organization, then name.
- A revoked token is reported for that account and the others are still listed.

## See also

- [`find`](/commands/find/) · [`whoami`](/commands/whoami/) · [`doctor`](/commands/doctor/)
