---
title: supawho rename
description: Rename a saved Supabase account. The stored token is carried over to the new name.
---

Rename a saved account. The token is carried over to the new name.

## Synopsis

```bash
supawho rename <old> <new>
```

## Example

```bash
supawho rename old-gig acme
```

```
Account 'old-gig' renamed to 'acme'.
```

## Behavior

- Run `rename` with missing arguments to pick the account and enter the new name interactively.
- The new name may not contain commas or whitespace.
- Renaming a name that doesn't exist prints `Account '<old>' not found.` and exits with code `1`.

## See also

- [`add`](/supawho/commands/add/) · [`remove`](/supawho/commands/remove/) · [`list`](/supawho/commands/list/)
