package cli

// Interactive shows the account menu and logs in with the chosen account.
// With no accounts saved it falls back to the guided add flow.
func (a *App) Interactive() error {
	names, err := a.Store.List()
	if err != nil {
		return err
	}

	a.UI.Banner()
	if len(names) == 0 {
		return a.guidedAdd()
	}

	chosen, err := a.Prompt.Select("Select an account", names)
	if err != nil {
		a.UI.Err("Invalid selection.")
		return errHandled
	}
	return a.Use(chosen)
}

// confirm asks a yes/no question defaulting to yes, swallowing prompt errors.
func (a *App) confirm(title string) bool {
	ok, err := a.Prompt.Confirm(title, true)
	return err == nil && ok
}
