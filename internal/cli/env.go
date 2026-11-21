package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/EliaTolin/supawho/internal/store"
	"golang.org/x/term"
)

// folderFile is the per-project file that names the account to use. It holds
// only an account name, never a token, so it can be committed.
const folderFile = ".supawho"

const envUsage = "Usage: supawho env [name] [--shell sh|bash|zsh|fish|powershell]"

// Env prints the shell command that sets SUPABASE_ACCESS_TOKEN to the token
// of a saved account, for `eval "$(supawho env <name>)"` in a shell or an
// .envrc. Without a name it uses the nearest .supawho file.
//
// Its stdout is evaluated by a shell, so it only ever carries the export
// line: every message goes to Err, and nothing is printed to a terminal,
// where the token would end up in the scrollback.
func (a *App) Env(args []string) error {
	name, shell, ok := parseEnvArgs(args)
	if !ok {
		return a.envFail("%s", envUsage)
	}
	format, ok := exportFormats[shell]
	if !ok {
		return a.envFail("unknown shell %q. %s", shell, envUsage)
	}
	if a.terminal(a.Out) {
		cmd := "supawho env"
		if len(args) > 0 {
			cmd += " " + strings.Join(args, " ")
		}
		return a.envFail("env prints your access token for the shell to evaluate, so it won't print it to a terminal.\nRun it as: eval \"$(%s)\"", cmd)
	}

	from := ""
	if name == "" {
		var err error
		if name, from, err = a.folderAccount(); err != nil {
			return a.envFail("%v", err)
		}
		if name == "" {
			dir, _ := a.workDir()
			return a.envFail("no account name given and no %s file found in %s or its parents.\n%s", folderFile, dir, envUsage)
		}
	}

	// Only saved accounts are used: the name may come from a .supawho file
	// in a folder someone else wrote.
	names, err := a.Store.List()
	if err != nil {
		return a.envFail("%v", err)
	}
	token := ""
	if slices.Contains(names, name) {
		token, err = a.Store.Get(name)
	} else {
		err = store.ErrNotFound
	}
	if errors.Is(err, store.ErrNotFound) {
		if from != "" {
			return a.envFail("Account '%s' (from %s) not found. Run 'supawho list' to see saved accounts.", printable(name), from)
		}
		return a.envFail("Account '%s' not found. Run 'supawho list' to see saved accounts.", printable(name))
	}
	if err != nil {
		return a.envFail("%v", err)
	}
	if strings.ContainsFunc(token, isControl) {
		return a.envFail("the token saved for '%s' contains unexpected characters. Save it again with: supawho add %s <token>", printable(name), printable(name))
	}
	fmt.Fprintf(a.Out, format, quoteFor(shell, token))
	return nil
}

// exportFormats maps a shell to its "set and export a variable" line.
var exportFormats = map[string]string{
	"sh":         "export " + accessTokenVar + "=%s\n",
	"bash":       "export " + accessTokenVar + "=%s\n",
	"zsh":        "export " + accessTokenVar + "=%s\n",
	"fish":       "set -gx " + accessTokenVar + " %s\n",
	"powershell": "$env:" + accessTokenVar + " = %s\n",
	"pwsh":       "$env:" + accessTokenVar + " = %s\n",
}

const accessTokenVar = "SUPABASE_ACCESS_TOKEN"

// quoteFor single-quotes s so the shell takes it literally.
func quoteFor(shell, s string) string {
	switch shell {
	case "fish":
		s = strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s)
	case "powershell", "pwsh":
		s = strings.ReplaceAll(s, `'`, `''`)
	default:
		s = strings.ReplaceAll(s, `'`, `'\''`)
	}
	return "'" + s + "'"
}

func parseEnvArgs(args []string) (name, shell string, ok bool) {
	shell = "sh"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--shell":
			if i+1 >= len(args) {
				return "", "", false
			}
			i++
			shell = args[i]
		case strings.HasPrefix(arg, "--shell="):
			shell = strings.TrimPrefix(arg, "--shell=")
		case strings.HasPrefix(arg, "-"), name != "":
			return "", "", false
		default:
			name = arg
		}
	}
	return name, strings.ToLower(shell), true
}

// folderAccount walks up from the working directory to the nearest .supawho
// file and returns the account name in it (the first line that is neither
// blank nor a # comment) and the file's path. name and path are "" when no
// file is found.
func (a *App) folderAccount() (name, path string, err error) {
	dir, err := a.workDir()
	if err != nil {
		return "", "", err
	}
	for {
		path = filepath.Join(dir, folderFile)
		if info, statErr := os.Stat(path); statErr == nil && info.Mode().IsRegular() {
			b, err := os.ReadFile(path)
			if err != nil {
				return "", "", err
			}
			for _, line := range strings.Split(string(b), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					return line, path, nil
				}
			}
			return "", "", fmt.Errorf("%s has no account name", path)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", nil
		}
		dir = parent
	}
}

func (a *App) workDir() (string, error) {
	if a.getwd != nil {
		return a.getwd()
	}
	return os.Getwd()
}

func (a *App) terminal(w io.Writer) bool {
	if a.isTerminal != nil {
		return a.isTerminal(w)
	}
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// envFail prints "supawho: <msg>" to Err, never to Out.
func (a *App) envFail(format string, args ...any) error {
	w := a.Err
	if w == nil {
		w = io.Discard
	}
	fmt.Fprintf(w, "supawho: "+format+"\n", args...)
	return errHandled
}

func isControl(r rune) bool { return r < 0x20 || r == 0x7f }

// printable replaces control characters, so a name read from a file can't
// send escape sequences to the terminal.
func printable(s string) string {
	return strings.Map(func(r rune) rune {
		if isControl(r) {
			return '?'
		}
		return r
	}, s)
}
