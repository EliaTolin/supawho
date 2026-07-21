---
title: supawho (interactive)
description: Run supawho with no arguments to pick a Supabase account from an interactive menu and log in.
---

Running `supawho` with no arguments opens an interactive picker: choose an account by number and supawho logs you in.

## Synopsis

```bash
supawho
```

## Example

```
   ___  _   _ ___  ___  _    _ _  _  ___
  / __|| | | | _ \/ _ \| |  | | || |/ _ \
  \__ \| |_| |  _/ (_) | |/\| | __ | (_)
  |___/ \___/|_|  \__\_\__/\__|_||_|\___/

     🔍 Who are you today?

Select an account:

  1) myproject
  2) another-project

Enter number (1-2): 2
Logging in as 'another-project'...
Logged in as 'another-project'.
```

## Behavior

- If **no accounts** are saved yet, supawho starts a short guided flow to add your first one (see [`add`](/supawho/commands/add/)).
- An invalid selection prints `Invalid selection.` and exits with code `1`.

## See also

- [`use`](/supawho/commands/use/) — switch to a specific account by name
- [`list`](/supawho/commands/list/) — see saved accounts
