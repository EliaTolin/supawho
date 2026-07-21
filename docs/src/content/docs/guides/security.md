---
title: Security
description: How supawho stores your Supabase tokens — in the operating system's native secret vault, never on disk.
---

Your tokens **never touch the filesystem**. supawho stores each one in your operating system's native secret store — the same encrypted vault the OS uses for passwords and certificates:

- **macOS** → Keychain
- **Linux** → Secret Service (GNOME Keyring / KWallet, via `libsecret`)
- **Windows** → Credential Manager

| | supawho | Plain-text files |
|---|:---:|:---:|
| Encrypted at rest | ✅ | ❌ |
| Protected by system login | ✅ | ❌ |
| Hidden from other processes | ✅ | ❌ |
| Survives an accidental `git add .` | ✅ | ❌ |

## Updates are verified

`supawho upgrade` downloads the release archive over HTTPS and checks its SHA-256 against the published `checksums.txt` before replacing the binary. A mismatch aborts the update and leaves the current binary untouched.

## Requirements

- macOS, Linux, or Windows — a single static binary, no runtime to install
- On Linux, a Secret Service provider must be running (GNOME Keyring, KWallet, …)
- The [Supabase CLI](https://supabase.com/docs/guides/cli), available in your `PATH`
