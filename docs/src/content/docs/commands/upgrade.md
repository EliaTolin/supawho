---
title: supawho upgrade
description: Update supawho to the latest release. Downloads the new binary, verifies its checksum, and swaps it in place — or defers to your package manager.
---

Update supawho to the latest release.

## Synopsis

```bash
supawho upgrade
```

## What it does

1. Checks GitHub for the latest release.
2. If you're already current, it says so and stops.
3. Otherwise it downloads the build for your OS/arch, **verifies its SHA-256 checksum**, and replaces the running binary in place.

```bash
supawho upgrade
```

```
Checking for updates...
Updating 1.5.0 → 1.6.0...
Updated to 1.6.0.
```

## Package-manager installs

If supawho was installed through a package manager, `upgrade` won't overwrite it — it points you to the right command instead:

```
supawho was installed via Homebrew (run: brew upgrade supawho).
```

| Installed via | Update with |
|---|---|
| Homebrew (macOS) | `brew upgrade supawho` |
| Install script / raw binary | `supawho upgrade` |
| `.deb` / `.rpm` / `.apk` | your system package manager |

## Related

Set `SUPAWHO_NO_UPDATE_CHECK=1` to disable the background update notice. See [Updating](/supawho/guides/updating/) for the full picture.
