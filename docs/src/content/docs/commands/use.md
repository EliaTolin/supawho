---
title: supawho use
description: Switch to a saved Supabase account by name. supawho logs you in by driving the Supabase CLI with the stored token.
---

Switch to a saved account. supawho reads the account's token from the OS secret vault and logs you in by running `supabase login --token`.

## Synopsis

```bash
supawho use <name>
```

## Example

```bash
supawho use client-a
```

```
Logging in as 'client-a'...
Logged in as 'client-a'.
```

## Behavior

- Requires the [Supabase CLI](https://supabase.com/docs/guides/cli) on your `PATH` — supawho drives it, it doesn't reimplement login.
- If the account doesn't exist, it prints `Account '<name>' not found.` and exits with code `1`. Run [`list`](/supawho/commands/list/) to see exact names (they're case-sensitive).

## See also

- [`whoami`](/supawho/commands/whoami/) — confirm which email an account maps to
- [`list`](/supawho/commands/list/)
