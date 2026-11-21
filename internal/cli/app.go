// Package cli implements the supawho command handlers on top of a store.Store.
// All I/O and the Supabase login call are injectable so the logic is testable
// without touching the real OS secret store or the supabase CLI.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/EliaTolin/supawho/internal/cli/ui"
	"github.com/EliaTolin/supawho/internal/store"
	"github.com/EliaTolin/supawho/internal/updater"
)

// LoginFunc performs the actual `supabase login --token` call.
type LoginFunc func(token string) error

// UpgradeFunc self-updates the running binary to the latest release.
type UpgradeFunc func(ctx context.Context, current string, out io.Writer) error

// ProfileFunc resolves the email and organizations behind an access token.
type ProfileFunc func(token string) (email string, orgs []string, err error)

// Project is a Supabase project owned by an account, used by the reverse lookup.
type Project struct {
	Ref    string
	Name   string
	Org    string
	Region string
	Status string
}

// LookupFunc lists the projects an access token can access.
type LookupFunc func(token string) ([]Project, error)

// App holds the injectable dependencies for the command handlers.
type App struct {
	Store   *store.Store
	Login   LoginFunc
	Upgrade UpgradeFunc
	Profile ProfileFunc
	Lookup  LookupFunc
	In      io.Reader
	Out     io.Writer
	// Err receives messages that must stay off Out, such as the errors of
	// `env`, whose stdout is evaluated by a shell.
	Err     io.Writer
	Version string

	// Doctor probes. ActiveToken reports the token the supabase CLI uses and
	// where it came from; CLIInfo locates the CLI; Latest returns the newest
	// supawho release without the leading "v".
	ActiveToken func() (token, source string, err error)
	CLIInfo     func() (path, version string, err error)
	Latest      func(ctx context.Context) (string, error)

	// UI and Prompt handle presentation and interactive input. Both default
	// from In/Out on first Run, so tests only need to set In/Out.
	UI     *ui.UI
	Prompt ui.Prompter

	// now times the token checks; defaults to time.Now.
	now func() time.Time
	// getwd and isTerminal default to os.Getwd and a real terminal check.
	getwd      func() (string, error)
	isTerminal func(io.Writer) bool
}

// errHandled signals a user-facing failure whose message was already printed;
// Run maps it to exit code 1 without printing anything further.
var errHandled = errors.New("handled")

// Run dispatches args and returns the process exit code.
func (a *App) Run(args []string) int {
	if a.UI == nil {
		a.UI = ui.New(a.Out)
	}
	if a.Prompt == nil {
		a.Prompt = ui.NewPrompter(a.In, a.Out)
	}

	var err error
	switch cmd := arg(args, 0); cmd {
	case "add":
		err = a.Add(arg(args, 1), arg(args, 2))
	case "rename":
		err = a.Rename(arg(args, 1), arg(args, 2))
	case "remove":
		err = a.Remove(arg(args, 1))
	case "list":
		err = a.List()
	case "use":
		err = a.Use(arg(args, 1))
	case "whoami", "who":
		err = a.Whoami(arg(args, 1))
	case "find", "locate":
		err = a.Find(arg(args, 1))
	case "env":
		err = a.Env(args[1:])
	case "projects":
		err = a.Projects(args[1:])
	case "doctor", "check":
		err = a.Doctor(args[1:])
	case "upgrade", "update":
		err = a.runUpgrade()
	case "version", "--version", "-v":
		fmt.Fprintf(a.Out, "supawho %s\n", a.Version)
	case "help", "--help", "-h":
		a.help()
	case "":
		err = a.Interactive()
	default:
		a.UI.Err("Unknown command: %s", cmd)
		a.UI.Hint("Run 'supawho help' for usage.")
		return 1
	}

	if err != nil {
		switch {
		case errors.Is(err, errHandled):
		case errors.Is(err, store.ErrLocked):
			a.keychainLocked()
		default:
			a.UI.Err("%v", err)
		}
		return 1
	}
	return 0
}

// keychainLocked explains a locked macOS Keychain, which otherwise surfaces
// as a bare "exit status 36" from /usr/bin/security.
func (a *App) keychainLocked() {
	fmt.Fprintln(a.Out, "supawho: the macOS Keychain is locked, so your accounts can't be read or saved.")
	fmt.Fprintln(a.Out)
	fmt.Fprintln(a.Out, "Unlock it, then run the command again:")
	fmt.Fprintln(a.Out, "  security unlock-keychain ~/Library/Keychains/login.keychain-db")
	fmt.Fprintln(a.Out)
	fmt.Fprintln(a.Out, "This usually happens over SSH or in a session without a graphical login.")
}

// runUpgrade self-updates the binary, or explains how to upgrade when the
// binary is managed by a package manager (Homebrew, Scoop, apt, ...).
func (a *App) runUpgrade() error {
	if a.Upgrade == nil {
		return errors.New("upgrade is not available in this build")
	}
	err := a.Upgrade(context.Background(), a.Version, a.Out)
	var managed *updater.ErrManaged
	if errors.As(err, &managed) {
		a.UI.Info("supawho was installed via %s.", managed.Manager)
		return nil
	}
	return err
}

func arg(args []string, i int) string {
	if i < len(args) {
		return args[i]
	}
	return ""
}

// validateName mirrors the original: names may not contain commas or whitespace,
// since the index is a comma-separated list.
func validateName(name string) error {
	if strings.ContainsAny(name, ", \t\n\r\v\f") {
		return errors.New("account name may not contain commas or whitespace")
	}
	return nil
}
