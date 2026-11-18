---
title: supawho add
description: Save a Supabase account under a name so you can switch to it later. The access token is stored in your OS secret vault, with hidden input by default.
lastUpdated: 2026-11-17
---

Save a Supabase account under a name of your choice. The access token is stored in your OS secret vault; supawho never writes it to a file.

## Synopsis

```bash
supawho add
supawho add <name> <token>
```

- With no arguments, supawho asks for the name and the token. The token input is hidden.
- `<name>`: a label you choose (e.g. `work`, `client-a`). May not contain commas or whitespace.
- `<token>`: a Supabase access token, starting with `sbp_`. See [how to create an access token](/guides/getting-started/#get-your-supabase-access-token).

## Examples

The guided flow keeps the token off your screen and out of your shell history, then offers to log in right away:

```bash
supawho add
```

```
  Account name (e.g. myproject): work
  Access token (sbp_...):
Account 'work' saved.

  Login now as 'work'? [Y/n]
```

The one-line form is meant for scripts:

```bash
supawho add work "$SUPABASE_TOKEN"
```

```
Account 'work' saved.
```

:::caution
Typing the token directly on the command line (`supawho add work sbp_…`) saves it in your shell history. Use the guided flow, or pass it from a variable.
:::

## Behavior

- Adding a name that already exists **overwrites** its token.
- Invalid names (containing a comma or whitespace) are rejected with exit code `1`.

## See also

- [Switch to the account you just saved with `supawho use`](/commands/use/)
- [Rename an account](/commands/rename/) or [remove one](/commands/remove/)
- [Where supawho stores tokens](/guides/security/)
