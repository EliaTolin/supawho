# Design: comando `supawho doctor`

Data: 2026-10-20 · Stato: in revisione

## Obiettivo

supawho salva più token Supabase nel keyring di sistema, ma oggi non c'è un modo unico per sapere:

- quali token sono ancora validi e quali sono scaduti o revocati (`whoami` mostra l'errore ma esce sempre con 0);
- se l'ambiente è integro: Supabase CLI installata, token in uso dalla CLI corrispondente a un account salvato, supawho aggiornato.

`doctor` risponde a entrambe le domande con un solo comando, sia per l'uso interattivo sia per script/CI (`--json` + exit code).

**Criterio di successo:** con un comando si vede quali account vanno ri-aggiunti o rimossi e cosa sistemare nell'ambiente, ciascun problema accompagnato dal comando da lanciare.

**Fuori ambito:** correzioni automatiche (`--fix`), modifiche al menu interattivo.

## 1. Comportamento

**Sintassi:** `supawho doctor [nome] [--json]` (alias `check`).

- Senza argomenti: tutti gli account salvati + controlli di sistema.
- Con `nome`: solo quell'account + controlli di sistema. Nome inesistente → messaggio d'errore, exit 1.
- Argomenti sconosciuti (es. `--foo`) → usage, exit 1.
- Solo diagnosi: nessuna scrittura su keyring, file o CLI.
- `doctor` è escluso da `updater.MaybeNotify` in `main.go` (fa già il proprio controllo versione).

### Controlli e classificazione

| Controllo | PASS | WARN | FAIL |
|---|---|---|---|
| Token di ogni account (`/v1/profile`, timeout 10s) | HTTP 200; detail = email, con latenza | errore di rete, 5xx, timeout → "could not verify" | 401 (`ErrUnauthorized`) → hint `supawho remove <n>` poi `supawho add <n>`; nome nell'indice ma token assente nel keyring (`store.ErrNotFound`) → stesso hint |
| Supabase CLI | nel PATH; detail = versione + percorso | — | non trovata → hint con URL di installazione |
| Account attivo | token attivo == token di un account salvato; detail = nome + sorgente (`env`/`keychain`/`file`) | token attivo non salvato in supawho; nessun login (`ErrNoActiveToken`); sorgente illeggibile | token attivo coincide con un account risultato FAIL → hint `supawho use <altro>` |
| Versione supawho | aggiornata | release più recente → hint `supawho upgrade`; check fallito; versione non semver (`dev`) → "skipped" | — |

Con `doctor <nome>` il controllo "Account attivo" confronta comunque con tutti i token salvati (lettura dal keyring, senza chiamate HTTP), ma può dare FAIL solo se l'account attivo è quello controllato e risulta invalido.

**Exit code:** 1 se esiste almeno un FAIL, altrimenti 0. I WARN non influenzano l'exit code.

**Sicurezza:** nessun token viene stampato, loggato o incluso nel JSON. Il token attivo è confrontato solo in memoria.

## 2. Output

Un unico modello dati (`report`) alimenta tre renderer, scelti automaticamente.

### 2.1 Fancy (TTY con colore)

Vincolo: solo sequenze ANSI in scrittura, mai query al terminale (evita byte spuri sul prompt).

- Header: logo `supawho` con gradiente verde Supabase → ciano, sottotitolo `🩺 doctor · checking N accounts`.
- Sezioni `🔑 Accounts` e `🧰 System` con separatore sfumato.
- Righe account: spinner braille (80ms) poi riscrittura sul posto con il risultato. Check in parallelo (max 4), ma **ogni riga occupa una posizione fissa nell'ordine di `Store.List`**.
- Badge a sfondo pieno: ` PASS ` bianco su verde, ` WARN ` nero su giallo, ` FAIL ` bianco su rosso. Nome in bold, email in ciano, latenza in dim.
- Hint su riga sottostante: `↳` magenta + comando evidenziato.
- Box finale arrotondato con health score e barra: rossa < 50%, gialla 50–85%, verde > 85%. A 100%: `✨ All systems healthy — you're good to go!`.
- Cursore nascosto durante lo spinner (`\033[?25l`), sempre ripristinato: `defer` + handler SIGINT.
- Profondità colore: truecolor se `COLORTERM` ∈ {`truecolor`, `24bit`}, altrimenti 256 colori.

**Score:** `round((pass + 0.5·warn) / totale × 100)`. I check di sistema sono sempre 3, quindi il totale non è mai 0; per robustezza la funzione restituisce comunque 100 con totale 0.

### 2.2 Plain (no TTY, `NO_COLOR`, pipe, test)

Niente banner, spinner, box né ANSI. Una riga per check, in ordine fisso:

```
Accounts
PASS  acme        mario@acme.it  142ms
FAIL  old-client  token is invalid or revoked
      -> supawho remove old-client
System
PASS  Supabase CLI    2.48.3 /opt/homebrew/bin/supabase
WARN  supawho         1.6.2 -> 1.7.0 available
      -> supawho upgrade

6 passed, 1 warning, 1 failed (score 81%)
```

Output deterministico, usato dai golden test.

### 2.3 JSON (`--json`)

Solo JSON indentato su stdout, nessun colore/banner. Exit code identico agli altri renderer.

```json
{
  "accounts": [
    {"name": "acme", "status": "pass", "detail": "mario@acme.it", "latency_ms": 142},
    {"name": "old-client", "status": "fail", "detail": "token is invalid or revoked",
     "hint": "supawho remove old-client"}
  ],
  "system": [
    {"name": "supabase_cli", "status": "pass", "detail": "2.48.3 /opt/homebrew/bin/supabase"},
    {"name": "active_account", "status": "pass", "detail": "acme (keychain)"},
    {"name": "supawho_version", "status": "warn", "detail": "1.6.2 -> 1.7.0 available",
     "hint": "supawho upgrade"}
  ],
  "summary": {"pass": 6, "warn": 1, "fail": 1, "score": 81}
}
```

`status` ∈ {`pass`, `warn`, `fail`}. `hint` e `latency_ms` sono omessi quando vuoti/assenti.

## 3. Architettura

Le dipendenze esterne sono funzioni iniettabili su `App` (stesso pattern di `Profile` e `Lookup`), così i test usano fake senza HTTP né keyring.

### `internal/supabase`

- **`profile.go`**: `var ErrUnauthorized = errors.New("token is invalid or revoked")`, restituito da `getJSON` sul 401. Messaggio invariato → `whoami`/`find` non cambiano.
- **`active.go`** (nuovo): `ActiveToken() (token, source string, err error)`. Ordine:
  1. env `SUPABASE_ACCESS_TOKEN` → source `env`;
  2. keyring, service `Supabase CLI`, chiavi `supabase` poi `access-token` → source `keychain`;
  3. file `~/.supabase/access-token` → source `file`.

  Nessuna sorgente → `ErrNoActiveToken`. Le sorgenti sono variabili di package sostituibili nei test. **Da verificare in implementazione** sul sorgente della Supabase CLI: nome service/chiave del keyring (inclusa la variante per profilo) e prefisso eventuale (es. `go-keyring-base64:`).
- **`cli.go`** (nuovo): `CLIInfo() (path, version string, err error)` → `exec.LookPath("supabase")` + `supabase --version` con timeout 5s.

### `internal/cli`

- **`app.go`**: nuovi campi `ActiveToken func() (string, string, error)`, `CLIInfo func() (string, string, error)`, `Latest func(ctx context.Context) (string, error)`; `case "doctor", "check":` → `a.Doctor(args[1:])`.
- **`doctor.go`** (nuovo): parsing argomenti (scansione semplice, niente `flag.FlagSet`), esecuzione check, classificazione errori (`errors.Is` su `supabase.ErrUnauthorized`, `store.ErrNotFound`, `supabase.ErrNoActiveToken`), modello dati:
  ```go
  type check struct {
      Name, Status, Detail, Hint string
      Latency time.Duration
  }
  type report struct {
      Accounts, System []check
      Pass, Warn, Fail, Score int
  }
  ```
  Nessuna logica di rendering. Per l'output fancy espone i risultati tramite callback per-riga, così il renderer può aggiornare lo spinner all'arrivo di ciascun check.
- **`doctor_render.go`** (nuovo): `renderFancy`, `renderPlain`, `renderJSON`. La scelta: `--json` → JSON; altrimenti fancy se `a.UI` ha TTY e colore, plain in tutti gli altri casi.
- **`banner.go`**: riga help `doctor [name] [--json]   Check tokens and environment health`.

### `internal/cli/ui` (solo ANSI, nessuna nuova dipendenza)

- **`ui.go`**: colore da `bool` a livello `none|ansi256|truecolor`; `paint` resta compatibile con i chiamanti esistenti.
- **`style.go`** (nuovo): `Gradient(s, from, to RGB)`, `Badge(status)`, `Section(icon, title)`, `HintCmd(text, cmd)`, `Box(lines)`, `ProgressBar(pct, width)`, `visibleWidth(s)` (ignora ANSI, conta emoji/braille).
- **`live.go`** (nuovo): `LiveList` con `Add(label) *Row`, `Row.Done(line)`, `Stop()`; righe a posizione fissa, ridisegno ogni 80ms con `\r`, `\033[nA`, `\033[2K`; mutex; cursore ripristinato su `Stop` e SIGINT. Senza TTY/colore: `Done` non stampa nulla subito, `Stop` stampa le righe finali in ordine.
- **`banner.go`**: `BannerGradient()` sul logo esistente; `Banner()` invariato.

### `main.go`

Wiring di `supabase.ActiveToken`, `supabase.CLIInfo`, `Latest` (su `updater.FetchLatest` con client 5s); `doctor`/`check` esclusi da `MaybeNotify`.

### Documentazione

- `docs/src/content/docs/commands/doctor.md` (modello: `commands/find.md`): Synopsis, Example, Behavior (tabella check), JSON output, Exit codes, See also.
- Riga in `docs/src/content/docs/commands.md`, voce sidebar in `docs/astro.config.mjs`, tabella usage e feature nel `README.md`.

### Prerequisiti git

1. Commit separato su `main` del refactor UI in sospeso (`internal/cli/ui/` e file collegati): `refactor(ui): …`.
2. Eliminare il vecchio `feat/doctor` (nessun commit proprio, indietro rispetto a `main`) e ricrearlo da `main` aggiornato.

## 4. Test e verifica

### Test automatici

**`internal/cli/cli_test.go`** (harness `newHarness`, fake via closure, output plain):

- tutti i check PASS → exit 0, golden dell'output plain;
- un token `ErrUnauthorized` → FAIL con hint, exit 1;
- errore di rete → WARN, exit 0;
- token mancante nel keyring per un nome in indice → FAIL;
- `doctor <nome>` controlla un solo account; nome inesistente → exit 1;
- argomento sconosciuto → exit 1;
- CLI assente → FAIL, exit 1;
- token attivo non salvato → WARN; `ErrNoActiveToken` → WARN;
- token attivo uguale a un token invalido → FAIL;
- update disponibile → WARN con hint; versione `dev` → WARN "skipped";
- `--json`: unmarshal dell'output, verifica status, summary, assenza di token nel testo;
- help aggiornato (`wantHelp`).

**`internal/supabase`**:

- `profile_test.go`: `errors.Is(err, ErrUnauthorized)` sul 401;
- `active_test.go`: precedenza env > keychain (`supabase` > `access-token`) > file, con `keyring.MockInit()` e HOME temporanea; nessuna sorgente → `ErrNoActiveToken`.

**`internal/cli/ui/ui_test.go`**: `visibleWidth`, interpolazione `Gradient` (inclusa la fallback 256), `Badge`/`Box` plain, `LiveList` non-TTY deterministica, calcolo bucket colore di `ProgressBar`.

**Score**: test di unità sul calcolo (inclusi arrotondamento e totale 0).

### Verifica manuale

1. `go test ./...` e `go vet ./...`.
2. `go build -o /tmp/supawho . && /tmp/supawho doctor` sugli account reali, in iTerm e Terminal: gradiente, spinner, box; nessun byte spurio sul prompt; cursore visibile anche dopo Ctrl-C a metà; riga account attivo corretta dopo `supawho use <x>`.
3. Fallback: `COLORTERM= …` (256 colori), `NO_COLOR=1 …` e `… | cat` (plain).
4. `/tmp/supawho doctor --json | jq .` e `echo $?`.
5. `SUPABASE_ACCESS_TOKEN=bogus /tmp/supawho doctor` → account attivo WARN "not saved".
6. `PATH=/usr/bin:/bin /tmp/supawho doctor` → CLI non trovata, exit 1.
7. `cd docs && npm run build`.
