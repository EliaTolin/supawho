package cli

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/EliaTolin/supawho/internal/store"
	"github.com/EliaTolin/supawho/internal/supabase"
	"github.com/EliaTolin/supawho/internal/updater"
)

const (
	statusPass = "pass"
	statusWarn = "warn"
	statusFail = "fail"
)

// System check names, as they appear in the JSON output.
const (
	checkCLI     = "supabase_cli"
	checkActive  = "active_account"
	checkVersion = "supawho_version"
)

const (
	doctorUsage    = "Usage: supawho doctor [name] [--json]"
	installCLIHint = "install it: https://supabase.com/docs/guides/local-development/cli/getting-started"
	maxConcurrent  = 4
)

// check is the outcome of a single diagnostic.
type check struct {
	Name, Status, Detail, Hint string
	Latency                    time.Duration
}

// report is everything doctor found; every renderer works from it alone.
type report struct {
	Accounts, System        []check
	Pass, Warn, Fail, Score int
}

// progress lets a renderer react as each check completes. Either field may be
// nil; calls are serialized.
type progress struct {
	account func(i int, c check)
	system  func(i int, c check)
}

// Doctor checks every saved token (or just one) against the Supabase API plus
// the local environment, prints the report and fails when anything is broken.
func (a *App) Doctor(args []string) error {
	target, jsonOut, ok := parseDoctorArgs(args)
	if !ok {
		a.UI.Err(doctorUsage)
		return errHandled
	}
	if a.Profile == nil {
		return errors.New("doctor is not available in this build")
	}

	names, err := a.Store.List()
	if err != nil {
		return err
	}
	if target != "" {
		if !slices.Contains(names, target) {
			a.UI.Err("Account '%s' not found. Run 'supawho list' to see saved accounts.", target)
			return errHandled
		}
		names = []string{target}
	}

	var r report
	switch {
	case jsonOut:
		r = a.diagnose(names, progress{})
		if err := renderJSON(a.Out, r); err != nil {
			return err
		}
	default:
		r = a.diagnose(names, progress{})
		renderPlain(a.Out, r)
	}
	if r.Fail > 0 {
		return errHandled
	}
	return nil
}

// parseDoctorArgs accepts at most one account name and the --json flag.
func parseDoctorArgs(args []string) (target string, jsonOut, ok bool) {
	for _, arg := range args {
		switch {
		case arg == "--json":
			jsonOut = true
		case strings.HasPrefix(arg, "-"), target != "":
			return "", false, false
		default:
			target = arg
		}
	}
	return target, jsonOut, true
}

// savedToken is a saved account with its token, used to recognize the
// active login without any network call.
type savedToken struct{ name, token string }

// diagnose runs the account checks for names (concurrently, results kept in
// order) and then the system checks.
func (a *App) diagnose(names []string, p progress) report {
	now := a.now
	if now == nil {
		now = time.Now
	}

	// Every saved token is read up front: the active-account check compares
	// against all of them, and any error text is scrubbed of all of them.
	all, _ := a.Store.List()
	var saved []savedToken
	secrets := []string{}
	for _, n := range all {
		if tok, err := a.Store.Get(n); err == nil {
			saved = append(saved, savedToken{n, tok})
			secrets = append(secrets, tok)
		}
	}
	active, source, activeErr := "", "", error(nil)
	if a.ActiveToken != nil {
		active, source, activeErr = a.ActiveToken()
		if active != "" {
			secrets = append(secrets, active)
		}
	}

	var r report
	var mu sync.Mutex
	r.Accounts = make([]check, len(names))
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	for i, n := range names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			c := a.checkAccount(n, secrets, now)
			mu.Lock()
			defer mu.Unlock()
			r.Accounts[i] = c
			if p.account != nil {
				p.account(i, c)
			}
		}()
	}
	wg.Wait()

	steps := []func() check{
		a.checkCLI,
		func() check {
			return checkActiveAccount(a.ActiveToken != nil, active, source, activeErr, saved, r.Accounts)
		},
		a.checkVersion,
	}
	for i, step := range steps {
		c := step()
		c.Detail = redact(c.Detail, secrets)
		r.System = append(r.System, c)
		if p.system != nil {
			p.system(i, c)
		}
	}

	for _, c := range slices.Concat(r.Accounts, r.System) {
		switch c.Status {
		case statusPass:
			r.Pass++
		case statusWarn:
			r.Warn++
		case statusFail:
			r.Fail++
		}
	}
	r.Score = score(r.Pass, r.Warn, r.Pass+r.Warn+r.Fail)
	return r
}

func reAddHint(name string) string {
	return fmt.Sprintf("supawho remove %s && supawho add %s <new-token>", name, name)
}

func (a *App) checkAccount(name string, secrets []string, now func() time.Time) check {
	c := check{Name: name}
	token, err := a.Store.Get(name)
	if err != nil {
		c.Status, c.Hint = statusFail, reAddHint(name)
		c.Detail = "token missing from the OS secret store"
		if !errors.Is(err, store.ErrNotFound) {
			c.Detail = redact("could not read token: "+err.Error(), secrets)
		}
		return c
	}

	start := now()
	email, orgs, err := a.Profile(token)
	c.Latency = now().Sub(start)
	switch {
	case errors.Is(err, supabase.ErrUnauthorized):
		c.Status, c.Detail, c.Hint = statusFail, supabase.ErrUnauthorized.Error(), reAddHint(name)
		c.Latency = 0
	case err != nil:
		c.Status, c.Detail = statusWarn, redact("could not verify: "+err.Error(), secrets)
		c.Latency = 0
	default:
		c.Status, c.Detail = statusPass, email
		if email == "" {
			c.Detail = strings.Join(orgs, ", ")
		}
	}
	return c
}

func (a *App) checkCLI() check {
	c := check{Name: checkCLI}
	if a.CLIInfo == nil {
		c.Status, c.Detail = statusWarn, "not available in this build"
		return c
	}
	path, version, err := a.CLIInfo()
	if err != nil {
		c.Status, c.Detail, c.Hint = statusFail, "supabase CLI not found in PATH", installCLIHint
		return c
	}
	c.Status, c.Detail = statusPass, version+" "+path
	return c
}

// checkActiveAccount matches the CLI's active token against the saved ones.
// It fails only when the match is an account that failed in this run.
func checkActiveAccount(available bool, active, source string, activeErr error, saved []savedToken, accounts []check) check {
	c := check{Name: checkActive}
	switch {
	case !available:
		c.Status, c.Detail = statusWarn, "not available in this build"
		return c
	case errors.Is(activeErr, supabase.ErrNoActiveToken):
		c.Status, c.Detail, c.Hint = statusWarn, "not logged in", "supawho use <name>"
		return c
	case activeErr != nil:
		c.Status, c.Detail = statusWarn, "could not read active token"
		return c
	}

	i := slices.IndexFunc(saved, func(s savedToken) bool { return s.token == active })
	if i < 0 {
		c.Status, c.Detail, c.Hint = statusWarn, fmt.Sprintf("active token is not saved in supawho (%s)", source), "supawho add <name> <token>"
		return c
	}
	name := saved[i].name
	c.Status, c.Detail = statusPass, fmt.Sprintf("%s (%s)", name, source)

	if j := slices.IndexFunc(accounts, func(a check) bool { return a.Name == name }); j >= 0 && accounts[j].Status == statusFail {
		c.Status, c.Detail = statusFail, fmt.Sprintf("%s (%s) — token is invalid", name, source)
		c.Hint = reAddHint(name)
		if k := slices.IndexFunc(accounts, func(a check) bool { return a.Status == statusPass }); k >= 0 {
			c.Hint = "supawho use " + accounts[k].Name
		}
	}
	return c
}

func (a *App) checkVersion() check {
	c := check{Name: checkVersion}
	if a.Latest == nil {
		c.Status, c.Detail = statusWarn, "not available in this build"
		return c
	}
	latest, err := a.Latest(context.Background())
	if err != nil {
		c.Status, c.Detail = statusWarn, "could not check for updates"
		return c
	}
	newer, ok := updater.IsNewer(a.Version, latest)
	switch {
	case !ok:
		c.Status, c.Detail = statusWarn, a.Version+" (dev build), update check skipped"
	case newer:
		c.Status, c.Detail, c.Hint = statusWarn, fmt.Sprintf("%s -> %s available", a.Version, latest), "supawho upgrade"
	default:
		c.Status, c.Detail = statusPass, a.Version+" (latest)"
	}
	return c
}

// score weighs a warning as half a pass; with nothing checked it is 100.
func score(pass, warn, total int) int {
	if total == 0 {
		return 100
	}
	return int(math.Round((float64(pass) + 0.5*float64(warn)) / float64(total) * 100))
}

// redact masks every known token in s, so error text can never leak one.
func redact(s string, secrets []string) string {
	for _, tok := range secrets {
		if tok != "" {
			s = strings.ReplaceAll(s, tok, "***")
		}
	}
	return s
}
