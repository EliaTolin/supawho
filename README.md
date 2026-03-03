# supawho

CLI tool to manage multiple Supabase accounts. Tokens are stored securely in **macOS Keychain**.

> **macOS only** — This tool uses the macOS Keychain (`security` command) for secure token storage.

## Installation

1. Clone the repository:

```bash
git clone git@github.com:EliaTolin/supawho.git
```

2. Make the script executable:

```bash
chmod +x supawho/supawho
```

3. Add an alias to your shell config (`~/.zshrc` or `~/.bashrc`):

```bash
echo 'alias supawho="/path/to/supawho/supawho"' >> ~/.zshrc
```

4. Reload the shell:

```bash
source ~/.zshrc
```

## Usage

### Add an account

Generate an access token at [supabase.com/dashboard/account/tokens](https://supabase.com/dashboard/account/tokens), then:

```bash
supawho add <name> <token>
```

```bash
supawho add myproject sbp_xxx...
```

### List saved accounts

```bash
supawho list
```

### Login with a specific account

```bash
supawho use <name>
```

### Interactive login

Run without arguments to select an account interactively:

```bash
supawho
```

```
Select an account:

  1) myproject
  2) another-project

Enter number (1-2):
```

### Remove an account

```bash
supawho remove <name>
```

## Requirements

- macOS
- [Supabase CLI](https://supabase.com/docs/guides/cli)
