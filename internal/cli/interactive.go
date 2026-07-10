package cli

import (
	"fmt"
	"io"
	"strings"
)

// Interactive shows the account menu and logs in with the chosen account.
// With no accounts saved it falls back to the guided add flow.
func (a *App) Interactive() error {
	names, err := a.Store.List()
	if err != nil {
		return err
	}

	banner(a.Out)
	if len(names) == 0 {
		return a.guidedAdd()
	}

	fmt.Fprintln(a.Out, "Select an account:")
	fmt.Fprintln(a.Out, "")
	printMenu(a.Out, names)

	choice := a.prompt(fmt.Sprintf("\nEnter number (1-%d): ", len(names)))
	idx, ok := parseSelection(choice, len(names))
	if !ok {
		fmt.Fprintln(a.Out, "Invalid selection.")
		return errHandled
	}
	return a.Use(names[idx])
}

func printMenu(w io.Writer, names []string) {
	for i, n := range names {
		fmt.Fprintf(w, "  %d) %s\n", i+1, n)
	}
}

func isYes(s string) bool {
	s = strings.TrimSpace(s)
	return s == "y" || s == "Y"
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
