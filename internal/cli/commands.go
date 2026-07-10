package cli

import (
	"errors"
	"fmt"

	"github.com/EliaTolin/supawho/internal/store"
)

// Add saves an account. With no name/token it falls back to the guided flow.
func (a *App) Add(name, token string) error {
	if name == "" || token == "" {
		banner(a.Out)
		return a.guidedAdd()
	}
	if err := validateName(name); err != nil {
		fmt.Fprintln(a.Out, capitalize(err.Error())+".")
		return errHandled
	}
	if err := a.Store.Add(name, token); err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "Account '%s' saved.\n", name)
	return nil
}

func (a *App) guidedAdd() error {
	fmt.Fprintln(a.Out, "  No accounts found. Let's add your first one!")
	fmt.Fprintln(a.Out, "")
	fmt.Fprintln(a.Out, "  Get your token here:")
	fmt.Fprintln(a.Out, "  → https://supabase.com/dashboard/account/tokens")
	fmt.Fprintln(a.Out, "")

	name := a.prompt("  Account name (e.g. myproject): ")
	if name == "" {
		fmt.Fprintln(a.Out, "  Aborted.")
		return errHandled
	}

	token, err := a.readSecret("  Access token (sbp_...): ")
	if err != nil {
		return err
	}
	if token == "" {
		fmt.Fprintln(a.Out, "  Aborted.")
		return errHandled
	}

	if err := a.Add(name, token); err != nil {
		return err
	}

	confirm := a.prompt("\n  Login now as '" + name + "'? [Y/n] ")
	if confirm == "" || isYes(confirm) {
		return a.Use(name)
	}
	return nil
}

// Rename renames a saved account. Missing args trigger interactive prompts.
func (a *App) Rename(oldName, newName string) error {
	if oldName == "" {
		names, err := a.Store.List()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Fprintln(a.Out, "No accounts saved.")
			return errHandled
		}
		banner(a.Out)
		fmt.Fprintln(a.Out, "Select account to rename:")
		fmt.Fprintln(a.Out, "")
		printMenu(a.Out, names)
		choice := a.prompt(fmt.Sprintf("\nEnter number (1-%d): ", len(names)))
		idx, ok := parseSelection(choice, len(names))
		if !ok {
			fmt.Fprintln(a.Out, "Invalid selection.")
			return errHandled
		}
		oldName = names[idx]
	}

	if newName == "" {
		newName = a.prompt("  New name for '" + oldName + "': ")
		if newName == "" {
			fmt.Fprintln(a.Out, "  Aborted.")
			return errHandled
		}
		confirm := a.prompt("  Rename '" + oldName + "' → '" + newName + "'? [Y/n] ")
		if confirm != "" && !isYes(confirm) {
			fmt.Fprintln(a.Out, "  Aborted.")
			return nil
		}
	}

	if err := validateName(newName); err != nil {
		fmt.Fprintln(a.Out, capitalize(err.Error())+".")
		return errHandled
	}

	err := a.Store.Rename(oldName, newName)
	if errors.Is(err, store.ErrNotFound) {
		fmt.Fprintf(a.Out, "Account '%s' not found.\n", oldName)
		return errHandled
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "Account '%s' renamed to '%s'.\n", oldName, newName)
	return nil
}

// Remove deletes a saved account.
func (a *App) Remove(name string) error {
	if name == "" {
		fmt.Fprintln(a.Out, "Usage: supawho remove <name>")
		return errHandled
	}
	if err := a.Store.Remove(name); err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "Account '%s' removed.\n", name)
	return nil
}

// List prints the saved account names.
func (a *App) List() error {
	names, err := a.Store.List()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Fprintln(a.Out, "No accounts saved. Add one with:")
		fmt.Fprintln(a.Out, "  supawho add <name> <token>")
		return nil
	}
	fmt.Fprintln(a.Out, "Saved accounts:")
	for _, n := range names {
		fmt.Fprintf(a.Out, "  - %s\n", n)
	}
	return nil
}

// Use logs in with the token of the named account.
func (a *App) Use(name string) error {
	if name == "" {
		fmt.Fprintln(a.Out, "Usage: supawho use <name>")
		return errHandled
	}
	token, err := a.Store.Get(name)
	if errors.Is(err, store.ErrNotFound) {
		fmt.Fprintf(a.Out, "Account '%s' not found. Run 'supawho list' to see saved accounts.\n", name)
		return errHandled
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "Logging in as '%s'...\n", name)
	if err := a.Login(token); err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "Logged in as '%s'.\n", name)
	return nil
}
