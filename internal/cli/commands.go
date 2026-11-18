package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/EliaTolin/supawho/internal/store"
)

// Add saves an account. With no name/token it falls back to the guided flow.
func (a *App) Add(name, token string) error {
	if name == "" || token == "" {
		a.UI.Banner()
		return a.guidedAdd()
	}
	if err := validateName(name); err != nil {
		a.UI.Err("%s.", err.Error())
		return errHandled
	}
	if err := a.Store.Add(name, token); err != nil {
		return err
	}
	a.UI.OK("Account '%s' saved.", name)
	return nil
}

func (a *App) guidedAdd() error {
	a.UI.Info("No accounts found. Let's add your first one!")
	a.UI.Line()
	a.UI.Hint("Get your token here:")
	a.UI.Hint("→ https://supabase.com/dashboard/account/tokens")
	a.UI.Line()

	name, _ := a.Prompt.Input("Account name (e.g. myproject)")
	if name == "" {
		a.UI.Info("Aborted.")
		return errHandled
	}

	token, err := a.Prompt.Password("Access token (sbp_...)")
	if err != nil {
		return err
	}
	if token == "" {
		a.UI.Info("Aborted.")
		return errHandled
	}

	if err := a.Add(name, token); err != nil {
		return err
	}

	if a.confirm("Login now as '" + name + "'?") {
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
			a.UI.Info("No accounts saved.")
			return errHandled
		}
		a.UI.Banner()
		chosen, err := a.Prompt.Select("Select account to rename", names)
		if err != nil {
			a.UI.Err("Invalid selection.")
			return errHandled
		}
		oldName = chosen
	}

	if newName == "" {
		newName, _ = a.Prompt.Input("New name for '" + oldName + "'")
		if newName == "" {
			a.UI.Info("Aborted.")
			return errHandled
		}
		if !a.confirm("Rename '" + oldName + "' → '" + newName + "'?") {
			a.UI.Info("Aborted.")
			return nil
		}
	}

	if err := validateName(newName); err != nil {
		a.UI.Err("%s.", err.Error())
		return errHandled
	}

	err := a.Store.Rename(oldName, newName)
	if errors.Is(err, store.ErrNotFound) {
		a.UI.Err("Account '%s' not found.", oldName)
		return errHandled
	}
	if errors.Is(err, store.ErrExists) {
		a.UI.Err("Account '%s' already exists.", newName)
		return errHandled
	}
	if err != nil {
		return err
	}
	a.UI.OK("Account '%s' renamed to '%s'.", oldName, newName)
	return nil
}

// Remove deletes a saved account.
func (a *App) Remove(name string) error {
	if name == "" {
		a.UI.Err("Usage: supawho remove <name>")
		return errHandled
	}
	if err := a.Store.Remove(name); err != nil {
		return err
	}
	a.UI.OK("Account '%s' removed.", name)
	return nil
}

// List prints the saved account names.
func (a *App) List() error {
	names, err := a.Store.List()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		a.UI.Info("No accounts saved. Add one with:")
		a.UI.Hint("supawho add <name> <token>")
		return nil
	}
	a.UI.Title("Saved accounts", len(names))
	a.UI.Line()
	for _, n := range names {
		a.UI.Bullet("%s", n)
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
			a.UI.Err("Account '%s' not found. Run 'supawho list' to see saved accounts.", name)
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
			a.UI.Info("No accounts saved. Add one with:")
			a.UI.Hint("supawho add <name> <token>")
			return nil
		}
		names = all
	}

	a.UI.Hint("Looking up accounts on Supabase…")
	a.UI.Line()
	rows := make([][]string, 0, len(names))
	for _, n := range names {
		token, err := a.Store.Get(n)
		if err != nil {
			rows = append(rows, []string{n, "(not found)", ""})
			continue
		}
		email, orgs, err := a.Profile(token)
		if err != nil {
			rows = append(rows, []string{n, fmt.Sprintf("(%v)", err), ""})
			continue
		}
		if email == "" {
			email = "(unavailable)"
		}
		rows = append(rows, []string{n, email, strings.Join(orgs, ", ")})
	}
	a.UI.Table([]string{"ACCOUNT", "EMAIL", "ORGANIZATION"}, rows)
	return nil
}

// Find is the reverse lookup: given a Supabase project ref, it reports which
// saved account owns a matching project.
func (a *App) Find(query string) error {
	if query == "" {
		a.UI.Err("Usage: supawho find <project-ref>")
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
		a.UI.Err("No accounts saved.")
		return errHandled
	}

	a.UI.Info("Searching %d account(s) for %q...", len(names), query)
	found := false
	for _, n := range names {
		token, err := a.Store.Get(n)
		if err != nil {
			continue
		}
		projects, err := a.Lookup(token)
		if err != nil {
			a.UI.Warn("%s: skipped (%v)", n, err)
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
		a.UI.Err("No project matching %q in any saved account.", query)
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

	a.UI.Line()
	a.UI.OK("Found in account '%s'", account)
	a.UI.Line()
	a.UI.Field("Project", p.Name)
	a.UI.Field("Reference", p.Ref)
	a.UI.Field("Organization", p.Org)
	a.UI.Field("Region", p.Region)
	a.UI.Field("Status", p.Status)
	a.UI.Field("Email", email)
	a.UI.Field("Account", account)
	a.UI.Line()
	a.UI.Hint("Switch to it:  supawho use %s", account)
	a.UI.Line()
}

// Use logs in with the token of the named account.
func (a *App) Use(name string) error {
	if name == "" {
		a.UI.Err("Usage: supawho use <name>")
		return errHandled
	}
	token, err := a.Store.Get(name)
	if errors.Is(err, store.ErrNotFound) {
		a.UI.Err("Account '%s' not found. Run 'supawho list' to see saved accounts.", name)
		return errHandled
	}
	if err != nil {
		return err
	}
	a.UI.Info("Logging in as '%s'...", name)
	if err := a.Login(token); err != nil {
		return err
	}
	a.UI.OK("Logged in as '%s'.", name)
	return nil
}
