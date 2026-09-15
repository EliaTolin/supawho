---
title: supawho remove
description: Delete a saved Supabase account from supawho, removing its token from the OS secret vault.
---

Delete a saved account, removing its token from the OS secret vault.

## Synopsis

```bash
supawho remove <name>
```

## Example

```bash
supawho remove acme
```

```
Account 'acme' removed.
```

## Behavior

- Removing a name that isn't saved is a no-op — it won't error.
- This only removes the account from supawho; it does not revoke the token in Supabase. To revoke a token, use the [Supabase dashboard](https://supabase.com/dashboard/account/tokens).

## See also

- [`list`](/supawho/commands/list/) · [`rename`](/supawho/commands/rename/)
