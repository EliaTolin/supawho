---
title: What is supawho?
description: supawho is a free, open-source CLI that saves multiple Supabase accounts and switches the Supabase CLI between them with one command.
lastUpdated: 2026-11-17
---

**supawho** is a small, open-source command-line tool that saves multiple **Supabase** accounts and switches the Supabase CLI between them with one command. It runs as a single, dependency-free binary on **macOS, Linux and Windows**.

"Account" here means the Supabase account you sign in to the dashboard and CLI with, not the end users of your app.

## The problem

The Supabase CLI holds one login at a time. If you work with more than one account (a personal org, several clients, a work account) you know the routine:

```bash
supabase logout
supabase login      # paste token again
```

Every switch means finding the right access token and pasting it. Tokens end up in shell history, sticky notes, or plain-text files.

## What supawho does

You save each account **once**, then switch with a single command:

```bash
supawho add                    # asks for a name and the token (hidden input)
supawho use client-a           # switch in a second
```

supawho stores each token in your operating system's **native secret vault** (macOS Keychain, Linux Secret Service, or Windows Credential Manager) and never writes it to a file. [Security](/guides/security/) explains exactly what is stored where.

## Beyond switching

- [`supawho whoami`](/commands/whoami/): map every saved account to its real email and organizations
- [`supawho find <ref>`](/commands/find/): the reverse lookup. Which account owns a given project?
- [`supawho upgrade`](/commands/upgrade/): self-update, checksum-verified, on any OS

## Next steps

- [Install supawho and save your first account](/guides/getting-started/)
- [How to use multiple Supabase accounts with the CLI](/guides/multiple-supabase-accounts/): the manual ways and where supawho fits
- [Command reference](/commands/): one page per command
- [How supawho works](/guides/how-it-works/)
