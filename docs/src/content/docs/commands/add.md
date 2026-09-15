---
title: supawho add
description: Save a Supabase account under a name so you can switch to it later. The access token is stored in your OS secret vault.
---

Save a Supabase account under a name of your choice. The access token is stored in your OS secret vault, never on disk.

## Synopsis

```bash
supawho add <name> <token>
```

- `<name>` — a label you choose (e.g. `work`, `client-a`). May not contain commas or whitespace.
- `<token>` — a Supabase access token, starting with `sbp_`. See [Getting started](/supawho/guides/getting-started/) for how to create one.

## Examples

```bash
supawho add work sbp_1234567890abcdef
```

```
Account 'work' saved.
```

Run `add` with no arguments to use the interactive guided flow, which prompts for the name and token (the token input is hidden) and offers to log in right away:

```bash
supawho add
```

## Behavior

- Adding a name that already exists **overwrites** its token.
- Invalid names (containing a comma or whitespace) are rejected with exit code `1`.

## See also

- [`use`](/supawho/commands/use/) — switch to the account you just saved
- [`rename`](/supawho/commands/rename/) · [`remove`](/supawho/commands/remove/)
