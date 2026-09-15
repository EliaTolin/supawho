---
title: Commands overview
description: Every supawho command at a glance — add, use, list, whoami, find, rename, remove, upgrade and version — with links to each command's reference page.
---

supawho has a small, focused command set. Each one has its own reference page with syntax, examples and exit codes.

| Command | Description |
|---|---|
| [`supawho`](/commands/interactive/) | Interactive account picker |
| [`supawho add <name> <token>`](/commands/add/) | Save a new account |
| [`supawho use <name>`](/commands/use/) | Switch to an account |
| [`supawho list`](/commands/list/) | List saved accounts |
| [`supawho whoami [name]`](/commands/whoami/) | Reveal the email + organizations behind each account |
| [`supawho find <project-ref>`](/commands/find/) | Find which account owns a project |
| [`supawho rename <old> <new>`](/commands/rename/) | Rename a saved account |
| [`supawho remove <name>`](/commands/remove/) | Delete an account |
| [`supawho upgrade`](/commands/upgrade/) | Update to the latest version |
| [`supawho version`](/commands/version/) | Print the installed version |

## Conventions

- `<required>` arguments are shown in angle brackets; `[optional]` in square brackets.
- Commands that succeed exit with code `0`; handled errors (not found, bad input) exit with `1`.
- Account names may not contain commas or whitespace.
