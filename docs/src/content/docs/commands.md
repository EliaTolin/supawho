---
title: Commands
description: The full supawho command reference — add, use, list, whoami, find, rename, remove and upgrade.
---

| Command | Description |
|---|---|
| `supawho` | Interactive account picker |
| `supawho add <name> <token>` | Save a new account |
| `supawho use <name>` | Switch to an account |
| `supawho list` | Show all saved accounts |
| `supawho whoami [name]` | Reveal the email + organizations behind each account |
| `supawho find <project-ref>` | Find which account owns a project |
| `supawho rename <old> <new>` | Rename a saved account |
| `supawho remove <name>` | Delete an account |
| `supawho upgrade` | Update to the latest version |
| `supawho version` | Print the installed version |

## whoami — who is who

With many accounts it's hard to remember which name maps to which identity. `supawho whoami` asks Supabase and maps each account to its email and organizations:

```
$ supawho whoami
ACCOUNT          EMAIL                     ORGANIZATION
client-a         you@example.com           Client A
side-project     you+dev@example.com       Personal, Labs
old-gig          (token is invalid or revoked)
```

Add a name to look up a single account: `supawho whoami client-a`.

## find — which account owns this project?

The reverse lookup: paste a Supabase project ref and `supawho find` tells you which saved account it lives in, with the project's organization, region, status and the account's email.

```
$ supawho find sjslmvggunljemhwbkgm

  ✓ Found in account 'aurora'

  Project       almanaccolo
  Reference     sjslmvggunljemhwbkgm
  Organization  AuroraDigital
  Region        eu-central-1
  Status        ACTIVE_HEALTHY
  Email         mail@auroradigital.it
  Account       aurora

  Switch to it:  supawho use aurora
```

If no account owns the ref, it exits with code `1` — handy in scripts.

## upgrade — keep it current

```bash
supawho upgrade
```

Downloads the latest release, verifies its checksum, and swaps the binary in place. When installed through a package manager it won't fight it — it points you to the right command instead:

| Installed via | How it updates |
|---|---|
| Homebrew | `brew upgrade supawho` |
| Install script / raw binary | `supawho upgrade` |
| `.deb` / `.rpm` / `.apk` | your system package manager |

Set `SUPAWHO_NO_UPDATE_CHECK=1` to disable the background version check.
