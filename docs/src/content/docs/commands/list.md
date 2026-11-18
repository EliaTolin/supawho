---
title: supawho list
description: List the names of every Supabase account saved in supawho. Names only; use whoami to see the email and organizations behind each.
lastUpdated: 2026-11-17
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
- `list` shows only names — to see the email or organization behind each, use [`whoami`](/commands/whoami/).

## See also

- [Show emails and organizations with `supawho whoami`](/commands/whoami/)
- [Save another account](/commands/add/)
