---
title: 'supawho on headless Linux, SSH and WSL with pass'
description: Use supawho on servers, SSH sessions and WSL, where no Secret Service is running, by storing the tokens in pass (GPG-encrypted) instead.
lastUpdated: 2026-11-21
---

On Linux, supawho normally keeps tokens in the Secret Service (GNOME Keyring or KWallet). Servers, SSH sessions, WSL and containers usually have no Secret Service running, so supawho has nowhere to put the tokens and stops with:

```text
supawho: no secret store is available, so your accounts can't be read or saved.
```

On those machines, store the tokens in [pass](https://www.passwordstore.org/), the standard Unix password manager. Each token becomes a GPG-encrypted entry, and `gpg-agent` remembers your passphrase for a while, so you type it once per session rather than on every command. pass needs no D-Bus and no desktop.

## Set it up

1. **Install pass and a terminal pinentry.**

   ```bash
   # Debian / Ubuntu
   sudo apt install pass pinentry-curses
   # Fedora / RHEL
   sudo dnf install pass pinentry
   # Alpine
   sudo apk add pass pinentry
   ```

2. **Create a GPG key**, if you don't have one on this machine. Choose a strong passphrase: it is what protects the tokens.

   ```bash
   gpg --full-generate-key
   ```

3. **Let GPG ask for the passphrase in your terminal.** Add this to `~/.bashrc` (or `~/.zshrc`):

   ```bash
   export GPG_TTY=$(tty)
   ```

4. **Create the password store**, with the email or key ID of your GPG key:

   ```bash
   pass init you@example.com
   ```

5. **Tell supawho to use pass.** Add this to the same shell profile:

   ```bash
   export SUPAWHO_BACKEND=pass
   ```

Then use supawho as usual:

```bash
supawho add client-a
supawho use client-a
```

`supawho doctor` shows `Secret store  pass` when it is set up correctly.

## How long the passphrase is remembered

`gpg-agent` caches it for 10 minutes after the last use by default. To keep it for an hour, up to 8 hours in total, put this in `~/.gnupg/gpg-agent.conf`:

```text
default-cache-ttl 3600
max-cache-ttl 28800
```

Then reload the agent with `gpgconf --reload gpg-agent`.

## What is stored where

- Each account is an entry under `supawho/` in your password store (`~/.password-store/supawho/<name>.gpg`), plus `supawho/_account_list` with the account names.
- Tokens are handed to pass on standard input and read from its output. They never appear in command-line arguments, in temporary files or in error messages.
- You can inspect the entries with `pass ls supawho`, and back them up or sync them like any pass store, e.g. with `pass git`.

## What it protects against

pass protects the tokens **at rest**: someone who copies the store, a backup or the whole disk still needs your GPG private key and its passphrase. Like an unlocked Secret Service, it does not protect against software already running as your user while `gpg-agent` holds the passphrase. See [Security](/guides/security/).

## Good to know

- supawho never switches backend on its own. If `SUPAWHO_BACKEND=pass` is set and pass is missing, it stops and tells you, instead of silently using another store.
- Accounts saved in the OS keyring are not moved to pass automatically. Add them again with `SUPAWHO_BACKEND=pass` set.
- With pass, account names cannot contain `/` or start with `.` or `-`.
- In CI, use the Supabase CLI's own `SUPABASE_ACCESS_TOKEN` variable instead: see [CI and automation](/guides/ci-and-automation/).
