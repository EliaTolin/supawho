---
title: Updating
description: How to keep supawho up to date — the built-in self-update, the automatic update notice, and how it defers to your package manager.
---

supawho tells you when a new version ships and can update itself on any OS.

## The update notice

When you run a normal command, supawho checks GitHub for a newer release (at most once a day) and prints a one-line notice:

```
A new version of supawho is available: 1.5.0 → 1.6.0
Run 'supawho upgrade' to update.
```

The check is silent on failure, only runs in an interactive terminal, and can be disabled:

```bash
export SUPAWHO_NO_UPDATE_CHECK=1
```

## Updating

The right command depends on how you installed supawho:

| Installed via | Update with |
|---|---|
| Homebrew (macOS) | `brew upgrade supawho` |
| Install script / raw binary | `supawho upgrade` |
| `.deb` / `.rpm` / `.apk` (Linux) | your system package manager |

[`supawho upgrade`](/supawho/commands/upgrade/) downloads the latest release, verifies its SHA-256 checksum, and swaps the binary in place. If it detects a package-manager install, it won't fight it — it points you to the right command instead.
