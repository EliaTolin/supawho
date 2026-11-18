---
title: supawho use
lastUpdated: 2026-11-17
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

- Requires the [Supabase CLI](https://supabase.com/docs/guides/cli) on your `PATH`: supawho drives it, it doesn't reimplement login.
- The switch applies to every terminal, until you switch again.
- If the account doesn't exist, it prints `Account '<name>' not found. Run 'supawho list' to see saved accounts.` and exits with code `1`. Names are case-sensitive.
- The token is passed to `supabase login` as an argument. See [what happens when you switch](/guides/security/#what-happens-when-you-switch-accounts).

## See also

- [Confirm which email an account maps to with `supawho whoami`](/commands/whoami/)
- [List saved accounts](/commands/list/)
- [Troubleshooting a switch that seems to have no effect](/guides/troubleshooting/#the-supabase-cli-still-acts-as-another-account-after-supawho-use)
