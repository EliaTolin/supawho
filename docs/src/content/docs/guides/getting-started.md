---
title: Get started with supawho
description: Install supawho on macOS, Linux or Windows, save your first Supabase account and switch between accounts with one command.
lastUpdated: 2026-11-17
---

supawho is a single, dependency-free binary. In a few minutes you'll have it installed, one account saved, and the Supabase CLI logged in as that account.

## Before you start

supawho switches accounts by driving the official Supabase CLI, so you need the [Supabase CLI](https://supabase.com/docs/guides/cli) installed and on your `PATH`. Check with:

```bash
supabase --version
```

## Install

### macOS

```bash wrap
brew install EliaTolin/tap/supawho
```

Or with the install script:

```bash wrap
curl -fsSL https://raw.githubusercontent.com/EliaTolin/supawho/main/install.sh | bash
```

### Linux

```bash wrap
curl -fsSL https://raw.githubusercontent.com/EliaTolin/supawho/main/install.sh | bash
```

Or grab a native package (`.deb` / `.rpm` / `.apk`) from the [releases page](https://github.com/EliaTolin/supawho/releases) and install it with your package manager.

:::note
On Linux a Secret Service provider (GNOME Keyring, KWallet, …) must be running to store tokens.
:::

### Windows

```powershell wrap
irm https://raw.githubusercontent.com/EliaTolin/supawho/main/install.ps1 | iex
```

Or download the Windows `.zip` for your CPU (`x86_64` or `arm64`) from the [releases page](https://github.com/EliaTolin/supawho/releases), extract `supawho.exe`, and add it to your `PATH`.

### From source

Requires [Go](https://go.dev/dl/) 1.25+:

```bash
git clone https://github.com/EliaTolin/supawho.git
cd supawho
go install .
```

Builds are published for Intel/AMD (`x86_64`) and ARM (`arm64`) on all three systems.

### Check the install

```bash
supawho version
```

## Get your Supabase access token

1. Open [supabase.com/dashboard](https://supabase.com/dashboard) → **profile icon** (top right)
2. **Account preferences** → **Access Tokens**
3. **Generate new token**, name it, and copy it

The token starts with `sbp_` and is shown only once, so copy it right away.

## Save your first account

Run [`supawho add`](/commands/add/) with no arguments. It asks for a name, then for the token with hidden input, and offers to log you in right away:

```bash
supawho add
```

The token goes into your OS secret vault. See [Security](/guides/security/) for the details.

:::tip
`supawho add <name> <token>` does the same in one line, but your shell saves the token in its history. Prefer the prompt on your own machine.
:::

## Add a second account and switch

Repeat `supawho add` with the token of another account, then switch by name with [`supawho use`](/commands/use/):

```bash
supawho use work
supawho use client-a
```

```
Logging in as 'client-a'...
Logged in as 'client-a'.
```

## Check which account is which

[`supawho whoami`](/commands/whoami/) asks Supabase for the email and organizations behind every saved account:

```bash
supawho whoami
```

```
ACCOUNT          EMAIL                     ORGANIZATION
work             you@example.com           Acme Inc
client-a         you+client@example.com    Client A
```

To confirm the Supabase CLI itself is on the right account, list the projects it can see:

```bash
supabase projects list
```

## Next steps

- Run [`supawho`](/commands/interactive/) with no arguments for an interactive account picker
- [All supawho commands](/commands/)
- [Find which account owns a project](/commands/find/)
- [Troubleshooting](/guides/troubleshooting/) if something doesn't work
