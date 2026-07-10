// Command supawho manages multiple Supabase CLI accounts securely, storing
// each account's access token in the OS secret store (macOS Keychain, Linux
// Secret Service, Windows Credential Manager).
package main

import (
	"os"

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
		In:      os.Stdin,
		Out:     os.Stdout,
		Version: current,
	}

	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	code := app.Run(os.Args[1:])

	// Passive "update available" notice: only for everyday interactive use.
	// Skipped for upgrade/version/help and when output is not a terminal
	// (pipes, scripts), so it never pollutes machine-readable output.
	switch cmd {
	case "upgrade", "update", "version", "--version", "-v", "help", "--help", "-h":
	default:
		if term.IsTerminal(int(os.Stdout.Fd())) {
			updater.MaybeNotify(current, os.Stderr)
		}
	}

	os.Exit(code)
}
