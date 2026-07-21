---
title: Getting started
description: Install supawho on macOS, Linux or Windows, then save and switch your first Supabase account.
---

supawho is a single, dependency-free binary. Pick your platform — every option installs the same tool.

## Install

### macOS

```bash
brew install EliaTolin/tap/supawho
```

Or with the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/EliaTolin/supawho/main/install.sh | bash
```

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/EliaTolin/supawho/main/install.sh | bash
```

Or grab a native package (`.deb` / `.rpm` / `.apk`) from the [releases page](https://github.com/EliaTolin/supawho/releases) and install it with your package manager.

:::note
On Linux a Secret Service provider (GNOME Keyring, KWallet, …) must be running to store tokens.
:::

### Windows

```powershell
irm https://raw.githubusercontent.com/EliaTolin/supawho/main/install.ps1 | iex
```

Or download `supawho_*_Windows_x86_64.zip` from the [releases page](https://github.com/EliaTolin/supawho/releases), extract `supawho.exe`, and add it to your `PATH`.

### From source

Requires [Go](https://go.dev/dl/) 1.25+:

```bash
git clone https://github.com/EliaTolin/supawho.git
cd supawho
go install .
```

## Get your Supabase access token

1. Open [supabase.com/dashboard](https://supabase.com/dashboard) → **profile icon** (top right)
2. **Account preferences** → **Access Tokens**
3. **Generate new token**, name it, and copy it

The token starts with `sbp_` and is shown only once — copy it right away.

## Save and switch

```bash
supawho add work sbp_your_token   # save an account (once)
supawho use work                  # switch to it in a second
```

Run `supawho` with no arguments for an interactive picker, or head to the [Commands](/supawho/commands/) page for the full reference.
