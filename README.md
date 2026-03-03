# supabase_login

CLI tool to manage multiple Supabase accounts. Tokens are stored securely in **macOS Keychain**.

> **macOS only** — This tool uses the macOS Keychain (`security` command) for secure token storage.

## Installation

1. Clone the repository:

```bash
git clone git@github.com:EliaTolin/supabase_login.git
cd supabase_login
```

2. Make the script executable:

```bash
chmod +x supabase_login
```

3. Add an alias to your shell config (`~/.zshrc` or `~/.bashrc`):

```bash
alias supabase_login="/path/to/supabase_login"
```

## Usage

### Add an account

Generate an access token at [supabase.com/dashboard/account/tokens](https://supabase.com/dashboard/account/tokens), then:

```bash
supabase_login add <name> <token>
```

```bash
supabase_login add myproject sbp_xxx...
```

### List saved accounts

```bash
supabase_login list
```

### Login with a specific account

```bash
supabase_login use <name>
```

### Interactive login

Run without arguments to select an account interactively:

```bash
supabase_login
```

```
Select an account:

  1) myproject
  2) another-project

Enter number (1-2):
```

### Remove an account

```bash
supabase_login remove <name>
```

## Requirements

- macOS
- [Supabase CLI](https://supabase.com/docs/guides/cli)
