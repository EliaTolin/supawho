---
title: Introduction
description: supawho is a free, open-source CLI to manage multiple Supabase accounts and switch between them in seconds — with tokens stored in your OS secret vault, never on disk.
---

**supawho** is a small, open-source command-line tool that lets you manage multiple **Supabase** accounts and switch between them in seconds. It runs as a single, dependency-free binary on **macOS, Linux and Windows**.

## The problem

If you work with more than one Supabase account — a personal org, several clients, a work account — you know the dance:

```bash
supabase logout
supabase login      # paste token again
```

Every switch means finding the right access token and pasting it. Tokens end up in shell history, sticky notes, or plain-text files.

## What supawho does

You save each account **once**, then switch with a single command:

```bash
supawho add client-a sbp_xxx   # save it once
supawho use client-a           # switch in a second
```

Under the hood, supawho stores each token in your operating system's **native secret vault** — macOS Keychain, Linux Secret Service, or Windows Credential Manager — so tokens are encrypted and **never written to a file**. See [Security](/supawho/guides/security/) for details.

## Beyond switching

- [`supawho whoami`](/supawho/commands/whoami/) — map every saved account to its real email and organizations
- [`supawho find <ref>`](/supawho/commands/find/) — the reverse lookup: which account owns a given project?
- [`supawho upgrade`](/supawho/commands/upgrade/) — self-update, checksum-verified, on any OS

## Next steps

- [Getting started](/supawho/guides/getting-started/) — install and save your first account
- [Commands](/supawho/commands/) — the full reference, one page per command
- [How it works](/supawho/guides/how-it-works/) — the model behind supawho
