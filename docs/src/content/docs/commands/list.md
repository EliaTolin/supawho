---
title: supawho list
description: List the names of every Supabase account saved in supawho.
---

List the names of all saved accounts.

## Synopsis

```bash
supawho list
```

## Example

```bash
supawho list
```

```
Saved accounts:
  - work
  - client-a
  - side-project
```

## Behavior

- With no accounts saved, it tells you how to add one.
- `list` shows only names — to see the email or organization behind each, use [`whoami`](/supawho/commands/whoami/).

## See also

- [`whoami`](/supawho/commands/whoami/) — names plus emails and organizations
- [`add`](/supawho/commands/add/)
