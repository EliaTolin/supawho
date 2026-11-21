---
title: 'Per-project Supabase accounts with direnv'
description: Give every project folder its own Supabase account. supawho env and direnv set SUPABASE_ACCESS_TOKEN when you cd into the folder, so two terminals can use two accounts at once.
lastUpdated: 2026-11-21
---

[`supawho use`](/commands/use/) changes the Supabase CLI login for the whole machine. With two terminals open on two clients' projects, the second switch silently changes the account under the first one.

The Supabase CLI reads `SUPABASE_ACCESS_TOKEN` before its saved login. [`supawho env`](/commands/env/) prints that variable for a saved account, and [direnv](https://direnv.net/) sets it when you enter a folder and removes it when you leave. Each folder gets its own account, and the token never leaves the OS secret store for a file.

## Set it up

1. **Install direnv** and hook it into your shell, as its [installation guide](https://direnv.net/docs/installation.html) explains (`brew install direnv`, `apt install direnv`, …).

2. **Name the account in the project root.** The file holds only the name, so you can commit it:

   ```bash
   echo client-a > .supawho
   ```

3. **Create an `.envrc`** next to it:

   ```bash
   # .envrc
   watch_file .supawho
   eval "$(supawho env)"
   ```

   `watch_file` makes direnv reload when `.supawho` changes. You can also skip the file and write the name directly: `eval "$(supawho env client-a)"`.

4. **Allow it once:**

   ```bash
   direnv allow
   ```

Now `cd` into the folder and every `supabase` command runs as `client-a`, in that folder and its subfolders only. `supawho whoami` confirms it:

```
Active: client-a (from SUPABASE_ACCESS_TOKEN)
This folder: client-a (/home/you/code/client-a/.supawho)
```

## Without direnv

In any shell, for that shell only:

```bash
eval "$(supawho env client-a)"            # bash, zsh
supawho env client-a --shell fish | source
```

```powershell
supawho env client-a --shell powershell | Invoke-Expression
```

On Windows, put the PowerShell line in a script you run when you open the project.

## For teams

Commit `.supawho` and `.envrc`, not the token. Each teammate saves their own token under the agreed name:

```bash
supawho add client-a
```

If an account is missing, `supawho env` says so on stderr, direnv shows it, and nothing else changes.

## Security trade-offs

- The token stays in the OS secret store and is read each time direnv loads the folder. It is never written to `.envrc`, `.supawho` or any other file.
- While you are in the folder, the token is in your shell's environment, and every program started from it inherits it: scripts, `npm run`, editors. That is the point, since the Supabase CLI is one of them, but it is wider than [`use`](/commands/use/), which hands the token to the Supabase CLI only.
- `.supawho` decides which of **your** accounts a folder uses. After pulling changes from others, check it like any other config. direnv asks you to allow `.envrc` again when it changes, but not `.supawho`.
- `supawho env` refuses to print the token to a terminal, so it doesn't end up on screen or in your scrollback.
