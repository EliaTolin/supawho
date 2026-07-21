package supabase

import (
	"net/http"
	"time"
)

// Project is a Supabase project reachable by an access token.
type Project struct {
	Ref    string
	Name   string
	Org    string
	Region string
	Status string
}

// Projects lists the projects an access token can access, via the Management
// API (GET /v1/projects). Used by the reverse lookup: given a project ref,
// find which saved account owns it.
func Projects(token string) ([]Project, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	var raw []struct {
		ID      string `json:"id"`
		Ref     string `json:"ref"`
		Name    string `json:"name"`
		OrgSlug string `json:"organization_slug"`
		Region  string `json:"region"`
		Status  string `json:"status"`
	}
	if err := getJSON(client, token, "/v1/projects", &raw); err != nil {
		return nil, err
	}

	// Resolve organization slugs to human names (best-effort).
	orgName := map[string]string{}
	var orgs []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := getJSON(client, token, "/v1/organizations", &orgs); err == nil {
		for _, o := range orgs {
			orgName[o.Slug] = o.Name
		}
	}

	out := make([]Project, 0, len(raw))
	for _, p := range raw {
		ref := p.Ref
		if ref == "" {
			ref = p.ID
		}
		org := p.OrgSlug
		if name := orgName[p.OrgSlug]; name != "" {
			org = name
		}
		out = append(out, Project{Ref: ref, Name: p.Name, Org: org, Region: p.Region, Status: p.Status})
	}
	return out, nil
}
