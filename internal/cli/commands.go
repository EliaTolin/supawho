package cli

import (
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

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

// Whoami shows the email (and organizations) behind each saved account by
// querying the Supabase Management API. With a name, only that account is shown.
func (a *App) Whoami(name string) error {
	if a.Profile == nil {
		return errors.New("whoami is not available in this build")
	}

	var names []string
	if name != "" {
		if _, err := a.Store.Get(name); errors.Is(err, store.ErrNotFound) {
			fmt.Fprintf(a.Out, "Account '%s' not found. Run 'supawho list' to see saved accounts.\n", name)
			return errHandled
		} else if err != nil {
			return err
		}
		names = []string{name}
	} else {
		all, err := a.Store.List()
		if err != nil {
			return err
		}
		if len(all) == 0 {
			fmt.Fprintln(a.Out, "No accounts saved. Add one with:")
			fmt.Fprintln(a.Out, "  supawho add <name> <token>")
			return nil
		}
		names = all
	}

	fmt.Fprintln(a.Out, "Looking up accounts on Supabase...")
	tw := tabwriter.NewWriter(a.Out, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "ACCOUNT\tEMAIL\tORGANIZATION")
	for _, n := range names {
		token, err := a.Store.Get(n)
		if err != nil {
			fmt.Fprintf(tw, "%s\t(not found)\t\n", n)
			continue
		}
		email, orgs, err := a.Profile(token)
		if err != nil {
			fmt.Fprintf(tw, "%s\t(%v)\t\n", n, err)
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", n, email, strings.Join(orgs, ", "))
	}
	return tw.Flush()
}

// Find is the reverse lookup: given a Supabase project ref (or a name
// substring), it reports which saved account owns a matching project.
func (a *App) Find(query string) error {
	if query == "" {
		fmt.Fprintln(a.Out, "Usage: supawho find <project-ref>")
		return errHandled
	}
	if a.Lookup == nil {
		return errors.New("find is not available in this build")
	}

	names, err := a.Store.List()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Fprintln(a.Out, "No accounts saved.")
		return errHandled
	}

	fmt.Fprintf(a.Out, "Searching %d account(s) for %q...\n", len(names), query)
	found := false
	for _, n := range names {
		token, err := a.Store.Get(n)
		if err != nil {
			continue
		}
		projects, err := a.Lookup(token)
		if err != nil {
			fmt.Fprintf(a.Out, "  %s: skipped (%v)\n", n, err)
			continue
		}
		for _, p := range projects {
			if p.Ref == query {
				found = true
				a.printMatch(n, token, p)
			}
		}
	}

	if !found {
		fmt.Fprintf(a.Out, "No project matching %q in any saved account.\n", query)
		return errHandled
	}
	return nil
}

// printMatch renders a rich detail block for a found project, enriching it with
// the owning account's email (best-effort via the Supabase API).
func (a *App) printMatch(account, token string, p Project) {
	var email string
	if a.Profile != nil {
		if e, _, err := a.Profile(token); err == nil {
			email = e
		}
	}

	field := func(label, value string) {
		if value != "" {
			fmt.Fprintf(a.Out, "  %-14s%s\n", label, value)
		}
	}

	fmt.Fprintf(a.Out, "\n  ✓ Found in account '%s'\n\n", account)
	field("Project", p.Name)
	field("Reference", p.Ref)
	field("Organization", p.Org)
	field("Region", p.Region)
	field("Status", p.Status)
	field("Email", email)
	field("Account", account)
	fmt.Fprintf(a.Out, "\n  Switch to it:  supawho use %s\n\n", account)
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
