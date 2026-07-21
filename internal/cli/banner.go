package cli

import (
	"fmt"
	"io"
)

const (
	green = "\033[1;32m"
	reset = "\033[0m"
)

func banner(w io.Writer) {
	fmt.Fprintf(w, "\n%s   ___  _   _ ___  ___  _    _ _  _  ___%s\n", green, reset)
	fmt.Fprintf(w, "%s  / __|| | | | _ \\/ _ \\| |  | | || |/ _ \\%s\n", green, reset)
	fmt.Fprintf(w, "%s  \\__ \\| |_| |  _/ (_) | |/\\| | __ | (_)%s\n", green, reset)
	fmt.Fprintf(w, "%s  |___/ \\___/|_|  \\__\\_\\__/\\__|_||_|\\___/%s\n", green, reset)
	fmt.Fprintf(w, "\n     🔍 Who are you today?\n\n")
}

func (a *App) help() {
	banner(a.Out)
	fmt.Fprintln(a.Out, "Usage:")
	fmt.Fprintln(a.Out, "  supawho                        Interactive account selection")
	fmt.Fprintln(a.Out, "  supawho add <name> <token>     Save account securely")
	fmt.Fprintln(a.Out, "  supawho rename <old> <new>     Rename a saved account")
	fmt.Fprintln(a.Out, "  supawho remove <name>          Remove account")
	fmt.Fprintln(a.Out, "  supawho list                   List saved accounts")
	fmt.Fprintln(a.Out, "  supawho use <name>             Login with specific account")
	fmt.Fprintln(a.Out, "  supawho whoami [name]          Show the email behind each account")
	fmt.Fprintln(a.Out, "  supawho find <project-ref>     Find which account owns a project")
	fmt.Fprintln(a.Out, "  supawho upgrade                Update supawho to the latest version")
	fmt.Fprintln(a.Out, "  supawho version                Show version")
	fmt.Fprintln(a.Out, "")
	fmt.Fprintln(a.Out, "Tokens are stored securely in your OS secret store.")
}
