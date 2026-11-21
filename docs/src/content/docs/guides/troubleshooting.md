---
title: Troubleshooting supawho
description: Fixes for supawho error messages. Account not found, locked macOS Keychain, locked Linux keyring, Supabase CLI not found, revoked tokens and upgrade messages.
lastUpdated: 2026-11-17
---

Find the message you see below. Each heading is the text supawho or your system prints.

## `Account '<name>' not found. Run 'supawho list' to see saved accounts.`

No account is saved under that name. Names are case-sensitive, so `Work` and `work` are different accounts. Run [`supawho list`](/commands/list/) to see the exact names, or [`supawho rename`](/commands/rename/) to fix one.

## `supawho: the macOS Keychain is locked, so your accounts can't be read or saved.`

The login keychain is locked, which usually happens over SSH or in a session without a graphical login. Unlock it, then run the command again:

```bash
security unlock-keychain ~/Library/Keychains/login.keychain-db
```

## Linux: `your keyring is locked`

Unlock the login keyring (GNOME Keyring or KWallet) and try again.

## Linux: `no secret store is available`

No Secret Service is running in this session, which is normal on servers, over SSH, in WSL and in containers. Store the tokens in pass instead: see [Headless Linux](/guides/headless-linux/).

## `SUPAWHO_BACKEND=pass is set, but pass is not installed`

Install pass, create a store with `pass init <your-gpg-key-id>`, and try again. To go back to the OS keyring, `unset SUPAWHO_BACKEND`.

## `supawho: exec: "supabase": executable file not found in $PATH`

supawho switches accounts by running the Supabase CLI, and it can't find it. [Install the Supabase CLI](https://supabase.com/docs/guides/cli) and check that `supabase --version` works in the same terminal. On Windows the message ends with `%PATH%`.

## `(token is invalid or revoked)` in `supawho whoami`

Supabase rejected that account's token: it was deleted or has expired. Create a new one in [Supabase Dashboard → Access Tokens](https://supabase.com/dashboard/account/tokens), then save it under the same name. Adding an existing name replaces its token:

```bash
supawho add
```

## The Supabase CLI still acts as another account after `supawho use`

Check whether `SUPABASE_ACCESS_TOKEN` is set in that terminal:

```bash
echo "$SUPABASE_ACCESS_TOKEN"
```

When the variable is set, the Supabase CLI uses it instead of the stored login, so switching has no visible effect. Unset it (`unset SUPABASE_ACCESS_TOKEN`) and look for it in your shell profile or a `.env` / `.envrc` file.

## `Account name may not contain commas or whitespace.`

Pick a name without spaces or commas, for example `client-a` instead of `client a`.

## `supawho was installed via Homebrew (run: brew upgrade supawho).`

Not an error. supawho was installed through a package manager, so [`supawho upgrade`](/commands/upgrade/) leaves the update to it. Run the command it suggests. The same applies to `.deb`, `.rpm` and `.apk` installs.

## `Unknown command: <name>`

Run `supawho help` for the list of commands, or see the [command reference](/commands/).

## Still stuck?

[Open an issue on GitHub](https://github.com/EliaTolin/supawho/issues) with the command you ran, the full output, your OS and the output of `supawho version`. Never paste an access token into an issue.
