// Package cli implements the supawho command handlers on top of a store.Store.
// All I/O and the Supabase login call are injectable so the logic is testable
// without touching the real OS secret store or the supabase CLI.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/EliaTolin/supawho/internal/store"
	"github.com/EliaTolin/supawho/internal/updater"
	"golang.org/x/term"
)

// LoginFunc performs the actual `supabase login --token` call.
type LoginFunc func(token string) error

// UpgradeFunc self-updates the running binary to the latest release.
type UpgradeFunc func(ctx context.Context, current string, out io.Writer) error

// ProfileFunc resolves the email and organizations behind an access token.
type ProfileFunc func(token string) (email string, orgs []string, err error)

// App holds the injectable dependencies for the command handlers.
type App struct {
	Store   *store.Store
	Login   LoginFunc
	Upgrade UpgradeFunc
	Profile ProfileFunc
	In      io.Reader
	Out     io.Writer
	Version string

	// readSecret reads a secret without echoing; overridable in tests.
	readSecret func(prompt string) (string, error)
	reader     *bufio.Reader
}

// errHandled signals a user-facing failure whose message was already printed;
// Run maps it to exit code 1 without printing anything further.
var errHandled = errors.New("handled")

// Run dispatches args and returns the process exit code.
func (a *App) Run(args []string) int {
	if a.reader == nil {
		a.reader = bufio.NewReader(a.In)
	}
	if a.readSecret == nil {
		a.readSecret = a.defaultReadSecret
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
	case "upgrade", "update":
		err = a.runUpgrade()
	case "version", "--version", "-v":
		fmt.Fprintf(a.Out, "supawho %s\n", a.Version)
	case "help", "--help", "-h":
		a.help()
	case "":
		err = a.Interactive()
	default:
		fmt.Fprintf(a.Out, "Unknown command: %s\n", cmd)
		fmt.Fprintln(a.Out, "Run 'supawho help' for usage.")
		return 1
	}

	if err != nil {
		if !errors.Is(err, errHandled) {
			fmt.Fprintf(a.Out, "supawho: %v\n", err)
		}
		return 1
	}
	return 0
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
		fmt.Fprintf(a.Out, "supawho was installed via %s.\n", managed.Manager)
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

// --- input helpers ---

func (a *App) prompt(label string) string {
	fmt.Fprint(a.Out, label)
	line, _ := a.reader.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func (a *App) defaultReadSecret(label string) (string, error) {
	fmt.Fprint(a.Out, label)
	if f, ok := a.In.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(a.Out)
		return string(b), err
	}
	// Non-terminal (tests, pipes): read a plain line.
	line, _ := a.reader.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), nil
}

// validateName mirrors the original: names may not contain commas or whitespace,
// since the index is a comma-separated list.
func validateName(name string) error {
	if strings.ContainsAny(name, ", \t\n\r\v\f") {
		return errors.New("account name may not contain commas or whitespace")
	}
	return nil
}

// parseSelection parses a 1-based menu choice against n items.
func parseSelection(s string, n int) (int, bool) {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || i < 1 || i > n {
		return 0, false
	}
	return i - 1, true
}
