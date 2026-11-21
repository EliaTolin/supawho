// Command supawho manages multiple Supabase CLI accounts securely, storing
// each account's access token in the OS secret store (macOS Keychain, Linux
// Secret Service, Windows Credential Manager).
package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/EliaTolin/supawho/internal/cli"
	"github.com/EliaTolin/supawho/internal/store"
	"github.com/EliaTolin/supawho/internal/supabase"
	"github.com/EliaTolin/supawho/internal/updater"
	"golang.org/x/term"
)

func main() {
	current := resolveVersion()
	app := &cli.App{
		Store:   store.NewKeyring(),
		Login:   supabase.Login,
		Upgrade: updater.Run,
		Profile: func(token string) (string, []string, error) {
			id, err := supabase.Whoami(token)
			return id.Email, id.Orgs, err
		},
		Lookup: func(token string) ([]cli.Project, error) {
			projects, err := supabase.Projects(token)
			if err != nil {
				return nil, err
			}
			out := make([]cli.Project, len(projects))
			for i, p := range projects {
				out[i] = cli.Project{Ref: p.Ref, Name: p.Name, Org: p.Org, Region: p.Region, Status: p.Status}
			}
			return out, nil
		},
		ActiveToken: supabase.ActiveToken,
		CLIInfo:     supabase.CLIInfo,
		Latest: func(ctx context.Context) (string, error) {
			rel, err := updater.FetchLatest(ctx, &http.Client{Timeout: 5 * time.Second})
			if err != nil {
				return "", err
			}
			return strings.TrimPrefix(rel.TagName, "v"), nil
		},
		In:      os.Stdin,
		Out:     os.Stdout,
		Err:     os.Stderr,
		Version: current,
	}

	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	code := app.Run(os.Args[1:])

	// Passive "update available" notice: only for everyday interactive use.
	// Skipped for upgrade/version/help, for doctor (which checks the version
	// itself), and when output is not a terminal
	// (pipes, scripts), so it never pollutes machine-readable output.
	switch cmd {
	case "env", "upgrade", "update", "version", "--version", "-v", "help", "--help", "-h", "doctor", "check":
	default:
		if term.IsTerminal(int(os.Stdout.Fd())) {
			updater.MaybeNotify(current, os.Stderr)
		}
	}

	os.Exit(code)
}
