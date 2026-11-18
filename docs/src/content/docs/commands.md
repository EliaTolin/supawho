---
title: supawho commands
description: Every supawho command at a glance, with aliases. add, use, list, whoami, find, rename, remove, upgrade, version and help, each linked to its reference page.
lastUpdated: 2026-11-17
---

supawho has a small, focused command set. Each one has its own reference page with syntax, examples and exit codes.

| Command | Description |
|---|---|
| [`supawho`](/commands/interactive/) | Interactive account picker |
| [`supawho add`](/commands/add/) | Save a new account (asks for name and token) |
| [`supawho add <name> <token>`](/commands/add/) | Save a new account in one line, for scripts |
| [`supawho use <name>`](/commands/use/) | Switch to an account |
| [`supawho list`](/commands/list/) | List saved accounts |
| [`supawho whoami [name]`](/commands/whoami/) | Reveal the email + organizations behind each account |
| [`supawho find <project-ref>`](/commands/find/) | Find which account owns a project |
| [`supawho rename <old> <new>`](/commands/rename/) | Rename a saved account |
| [`supawho remove <name>`](/commands/remove/) | Delete an account |
| [`supawho upgrade`](/commands/upgrade/) | Update to the latest version |
| [`supawho version`](/commands/version/) | Print the installed version |
| `supawho help` | Print the list of commands |

## Aliases

| Command | Also accepted |
|---|---|
| `whoami` | `who` |
| `find` | `locate` |
| `upgrade` | `update` |
| `version` | `--version`, `-v` |
| `help` | `--help`, `-h` |

## Conventions

- `<required>` arguments are shown in angle brackets; `[optional]` in square brackets.
- Commands that succeed exit with code `0`; handled errors (not found, bad input) exit with `1`.
- Account names may not contain commas or whitespace, and are case-sensitive.
- An unknown command prints `Unknown command: <name>` and exits with `1`.
