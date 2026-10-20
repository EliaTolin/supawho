# `supawho doctor` — design

Date: 2026-10-20 · Branch: `feat/doctor` (from `main`)

## Goal

A read-only health check that tells the user, in one command, whether their
saved access tokens still work and whether the supawho setup around them is
sound. Used by hand (readable report) and by scripts/CI (`--json` + exit code).

Non-goals: automatic repair, colored/styled output (the pending UI refactor in
`stash@{0}` / `internal/cli/ui/` stays parked), offline mode.

## Command

```
supawho doctor [account] [--json]
```

- `account` (optional) restricts the token checks to that account. The system
  checks always run. An unknown account prints
  `Account '<name>' not found. Run 'supawho list' to see saved accounts.` and
  exits 1 (same wording as `whoami`).
- `--json` may appear before or after `account`.
- Exit code: `1` if at least one check has status `fail`, otherwise `0`.
  `warn` never fails the command.

## Checks

Checks run in this order and are reported in this order.

| Group | Name | ok | warn | fail |
|---|---|---|---|---|
| `keyring` | `keyring` | index read succeeds | — | backend error reading the index |
| `index` | `index` | every name has a token; no duplicates; all names valid | index empty ("no accounts saved") | name without token, duplicate names, or names that fail `validateName` (detail lists them) |
| `network` | `api.supabase.com` | host answers any HTTP status | — | request error / timeout |
| `token` | `<account>` (one per account) | `GET /v1/organizations` → 200 | 403 ("token has limited scope"), or network error/timeout ("could not verify") | 401 ("invalid or revoked") |
| `tool` | `supabase cli` | installed and latest | newer release available, or latest version unknown | binary not in `PATH` |
| `tool` | `supawho` | current is latest | newer release available, or latest version unknown, or running a dev build | — |

Rules:

- If `keyring` fails, `index` and all `token` checks are skipped and reported
  as `warn` with detail `skipped: keyring unavailable`.
- If `network` fails, `token` checks and the "latest version" part of the
  `tool` checks are reported as `warn` with detail `skipped: no network`. The
  `supabase cli` presence check still runs (it is local).
- Token checks for a name already reported as "missing token" by `index` are
  reported as `fail` with detail `no token stored`, and no request is made.
- Token checks run concurrently, each with a 10s timeout; results are emitted
  in index order.
- Every `fail`/`warn` detail ends with a remediation hint where one exists:
  - invalid token → `run 'supawho add <name> <new-token>'`
  - missing token → `run 'supawho remove <name>' or re-add it`
  - supabase CLI missing → `install it: https://supabase.com/docs/guides/cli`
  - supabase CLI outdated → `update it: https://supabase.com/docs/guides/cli/getting-started#updating-the-supabase-cli`
  - supawho outdated → `run 'supawho upgrade'`
- Known limitation: tokens stored in the keyring but missing from the index
  ("orphans") cannot be detected, because the keyring API cannot enumerate
  keys. Documented, not checked.

## Architecture (approach A)

Doctor lives in the `cli` package, alongside `whoami` and `find`, and gets its
outside world through injectable functions on `App`, like those commands do.

### `internal/supabase`

- `profile.go`: add sentinels `ErrUnauthorized` (401) and `ErrForbidden` (403),
  returned by `getJSON`. `ErrUnauthorized`'s message stays
  `token is invalid or revoked`, so existing `whoami` output is unchanged.
- `profile.go`: `func Validate(token string) error` — single
  `GET /v1/organizations`, 10s client timeout, returns `nil`,
  `ErrUnauthorized`, `ErrForbidden`, or a transport/status error.
- `profile.go`: `func Ping() error` — unauthenticated `GET` to `apiBase`, 5s
  timeout; any HTTP response is success.
- `version.go` (new): `func CLIVersion() (string, error)` — runs
  `supabase --version`, reads **stdout only** (the CLI prints its own update
  notice on stderr), returns the trimmed first line. Missing binary returns an
  error wrapping `exec.ErrNotFound`.

### `internal/updater`

- Add `FetchLatestFor(ctx, client, repo string) (*Release, error)` where
  `repo` is `"owner/name"`. `FetchLatest` becomes a thin wrapper over it with
  supawho's repo, so `upgrade` and the update notice are unchanged.
- Reuse `IsNewer` for both version comparisons.

### `internal/cli`

`App` gains:

```go
Validate    func(token string) error             // token check
Ping        func() error                          // network check
ToolVersion func() (string, error)                // installed supabase CLI version
Latest      func(repo string) (string, error)     // latest release tag on GitHub
```

Any nil function makes its check report `warn` with `not available in this
build`, so tests and partial builds never panic.

`doctor.go` (new):

```go
type Check struct {
    Group  string `json:"group"`
    Name   string `json:"name"`
    Status string `json:"status"` // "ok" | "warn" | "fail"
    Detail string `json:"detail"`
}

func (a *App) runChecks(account string) []Check
func (a *App) Doctor(args []string) error
```

`Doctor` parses `[account] [--json]`, calls `runChecks`, renders, and returns
`errHandled` when any check failed.

`app.go`: dispatch `case "doctor":` → `a.Doctor(args[1:])`.
`banner.go`: help line
`supawho doctor [name] [--json]   Check tokens and setup health`.

### `main.go`

Wire `supabase.Validate`, `supabase.Ping`, `supabase.CLIVersion`, and a
`Latest` closure over `updater.FetchLatestFor` (`supabase/cli` and supawho's
repo) with a 10s timeout.

## Output

Text (default) — plain `tabwriter`, no colors, matching `whoami`:

```
Running supawho doctor...
CHECK             STATUS  DETAIL
keyring           ok      OS secret store reachable
index             ok      3 accounts, consistent
api.supabase.com  ok      reachable
token: work       ok      valid
token: client-a   FAIL    invalid or revoked — run 'supawho add client-a <new-token>'
token: client-b   warn    could not verify (timeout)
supabase cli      warn    2.98.2 installed, 2.121.0 available — update it: https://...
supawho           ok      1.6.2 (latest)

1 problem, 2 warnings.
```

The summary line is `All checks passed.` when there is nothing to report.
`FAIL` is uppercase so it stands out without color.

JSON (`--json`) — no progress line, no summary, just one object on stdout:

```json
{
  "ok": false,
  "checks": [
    {"group": "token", "name": "client-a", "status": "fail", "detail": "invalid or revoked — run 'supawho add client-a <new-token>'"}
  ]
}
```

`ok` is `true` iff no check has status `fail`. Exit code matches.

## Testing (TDD)

`internal/cli` (fakes on `App`, memory store):

- one test per severity cell of the Checks table;
- inconsistent index (name in index, key missing in memory backend);
- duplicate and invalid names in the index;
- keyring failure skips index/token checks;
- network failure skips token and latest-version checks;
- single-account filter, and unknown account;
- `--json` output parses and `ok` matches; `--json` before/after account;
- exit code 0 with warnings only, 1 with any fail;
- nil dependency → `warn`, no panic.

`internal/supabase` (`httptest`): `Validate` on 200/401/403/500; `Ping` on any
status and on a closed server; `getJSON` returns the sentinels.

`internal/updater`: `FetchLatestFor` hits the right URL; `FetchLatest`
unchanged.

The memory backend may need a test hook to inject a read error for the
keyring-failure case.

## Docs

- `README.md`: add `doctor` to the commands section.
- `docs/src/content/docs/commands/doctor.md`: new page in the style of
  `whoami.md`; link from `commands.md`.
