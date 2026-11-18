---
title: supawho doctor
description: Health check for supawho — verify that every saved Supabase access token is still valid, and that the Supabase CLI, the active login and supawho itself are in good shape.
---

`doctor` tells you at a glance which saved tokens still work, which ones were revoked, and whether your environment is healthy. Every problem comes with the command that fixes it.

## Synopsis

```bash
supawho doctor [name] [--json]
```

- Without arguments, every saved account is checked, plus the system checks.
- With `name`, only that account is checked, plus the system checks.
- `--json` prints a machine-readable report instead (see below).
- `supawho check` is an alias.

`doctor` only reads. It never changes your accounts, the OS secret store or the Supabase CLI login.

## Example

```bash
supawho doctor
```

On a terminal you get an animated report with a health score. When piped, or with `NO_COLOR` set, the output is plain text:

```
Accounts
PASS  acme        mario@acme.it  142ms
FAIL  old-client  token is invalid or revoked
      -> supawho remove old-client && supawho add old-client <new-token>
System
PASS  Supabase CLI    2.48.3 /opt/homebrew/bin/supabase
PASS  Active account  acme (keychain)
WARN  supawho         1.6.2 -> 1.7.0 available
      -> supawho upgrade

3 passed, 1 warning, 1 failed (score 70%)
```

## Behavior

| Check | PASS | WARN | FAIL |
|---|---|---|---|
| Each account's token (`GET /v1/profile`) | Valid, shows the email and latency | Network error, timeout or 5xx (could not verify), or the OS secret store could not be read | Token rejected (401), or missing from the OS secret store |
| Supabase CLI | Found in `PATH`, shows version and path | — | Not installed |
| Active account | The token the Supabase CLI is using matches a saved account | Not logged in, or logged in with a token that isn't saved in supawho | The active token belongs to an account whose token failed |
| supawho version | Up to date | Update available, check failed, or development build | — |

The active token is read the same way the Supabase CLI does: `SUPABASE_ACCESS_TOKEN`, then the OS keyring (current profile, then the legacy entry), then `~/.supabase/access-token`. The source is shown as `env`, `keychain` or `file`.

The health score counts each warning as half a pass: `round((passed + 0.5 × warnings) / total × 100)`.

Tokens are never printed, not even in the JSON report.

## JSON output

```bash
supawho doctor --json
```

```json
{
  "accounts": [
    { "name": "acme", "status": "pass", "detail": "mario@acme.it", "latency_ms": 142 },
    { "name": "old-client", "status": "fail", "detail": "token is invalid or revoked",
      "hint": "supawho remove old-client && supawho add old-client <new-token>" }
  ],
  "system": [
    { "name": "supabase_cli", "status": "pass", "detail": "2.48.3 /opt/homebrew/bin/supabase" },
    { "name": "active_account", "status": "pass", "detail": "acme (keychain)" },
    { "name": "supawho_version", "status": "warn", "detail": "1.6.2 -> 1.7.0 available",
      "hint": "supawho upgrade" }
  ],
  "summary": { "pass": 3, "warn": 1, "fail": 1, "score": 70 }
}
```

`status` is one of `pass`, `warn` or `fail`. `hint` and `latency_ms` are omitted when empty.

If doctor can't run at all (unknown account, invalid arguments), `--json` still prints JSON, an object with a single `error` field, and exits with `1`:

```json
{"error":"Account 'ghost' not found. Run 'supawho list' to see saved accounts."}
```

## Exit codes

| Situation | Exit code |
|---|---|
| No check failed (warnings allowed) | `0` |
| At least one check failed | `1` |
| Unknown account name or invalid arguments | `1` |

Warnings never change the exit code, so `supawho doctor --json` fits in scripts and CI.

## See also

- [`whoami`](/commands/whoami/) · [`use`](/commands/use/) · [`upgrade`](/commands/upgrade/)
