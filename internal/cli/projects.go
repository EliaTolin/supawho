package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
)

const projectsUsage = "Usage: supawho projects [name] [--json]"

// accountProjects is one account's listing: its projects, or why they could
// not be listed.
type accountProjects struct {
	name     string
	projects []Project
	err      error
}

// Projects lists the Supabase projects of every saved account (or just one).
// It fails when any account could not be listed, so a script never mistakes
// an incomplete listing for a complete one.
func (a *App) Projects(args []string) error {
	target, jsonOut, ok := parseNameArgs(args)
	if !ok {
		return a.fail(jsonOut, projectsUsage)
	}
	if a.Lookup == nil {
		return a.fail(jsonOut, "projects is not available in this build")
	}

	names, err := a.Store.List()
	if err != nil {
		if jsonOut {
			return a.fail(jsonOut, err.Error())
		}
		return err // Run explains a locked Keychain
	}
	if target != "" {
		if !slices.Contains(names, target) {
			return a.fail(jsonOut, fmt.Sprintf("Account '%s' not found. Run 'supawho list' to see saved accounts.", target))
		}
		names = []string{target}
	}

	if !jsonOut {
		if len(names) == 0 {
			a.UI.Info("No accounts saved. Add one with:")
			a.UI.Hint("supawho add <name> <token>")
			return nil
		}
		a.UI.Hint("Looking up projects on Supabase…")
	}

	results := a.listProjects(names)
	if jsonOut {
		if err := renderProjectsJSON(a, results); err != nil {
			return err
		}
	} else {
		a.renderProjects(results)
	}
	for _, r := range results {
		if r.err != nil {
			return errHandled
		}
	}
	return nil
}

// listProjects fetches each account's projects concurrently, keeping the
// saved order. Tokens are read one at a time, since every read may prompt.
func (a *App) listProjects(names []string) []accountProjects {
	results := make([]accountProjects, len(names))
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	for i, n := range names {
		results[i].name = n
		token, err := a.Store.Get(n)
		if err != nil {
			results[i].err = err
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			projects, err := a.Lookup(token)
			slices.SortFunc(projects, func(x, y Project) int {
				if c := strings.Compare(strings.ToLower(x.Org), strings.ToLower(y.Org)); c != 0 {
					return c
				}
				return strings.Compare(strings.ToLower(x.Name), strings.ToLower(y.Name))
			})
			results[i].projects, results[i].err = projects, err
		}()
	}
	wg.Wait()
	return results
}

func (a *App) renderProjects(results []accountProjects) {
	for _, r := range results {
		a.UI.Line()
		if r.err != nil {
			a.UI.Title(r.name, -1)
			a.UI.Warn("skipped (%v)", r.err)
			continue
		}
		a.UI.Title(r.name, len(r.projects))
		if len(r.projects) == 0 {
			a.UI.Hint("No projects.")
			continue
		}
		rows := make([][]string, 0, len(r.projects))
		for _, p := range r.projects {
			rows = append(rows, []string{p.Name, p.Ref, p.Org, p.Region, p.Status})
		}
		a.UI.Table([]string{"PROJECT", "REFERENCE", "ORGANIZATION", "REGION", "STATUS"}, rows)
	}
}

func renderProjectsJSON(a *App, results []accountProjects) error {
	type project struct {
		Ref          string `json:"ref"`
		Name         string `json:"name"`
		Organization string `json:"organization"`
		Region       string `json:"region"`
		Status       string `json:"status"`
	}
	type account struct {
		Name     string    `json:"name"`
		Projects []project `json:"projects"`
		Error    string    `json:"error,omitempty"`
	}
	doc := struct {
		Accounts []account `json:"accounts"`
	}{Accounts: make([]account, 0, len(results))}
	for _, r := range results {
		acc := account{Name: r.name, Projects: make([]project, 0, len(r.projects))}
		if r.err != nil {
			acc.Error = r.err.Error()
		}
		for _, p := range r.projects {
			acc.Projects = append(acc.Projects, project{p.Ref, p.Name, p.Org, p.Region, p.Status})
		}
		doc.Accounts = append(doc.Accounts, acc)
	}
	enc := json.NewEncoder(a.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}
