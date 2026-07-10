package supabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// apiBase is a var (not a const) so tests can point it at a local server.
var apiBase = "https://api.supabase.com"

// Identity is who a Supabase access token belongs to.
type Identity struct {
	Email string
	Orgs  []string
}

// Whoami resolves the identity behind a Supabase personal access token by
// calling the Management API. The email comes from /v1/profile; the (optional)
// organization names from /v1/organizations.
func Whoami(token string) (Identity, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	var profile struct {
		PrimaryEmail string `json:"primary_email"`
		Username     string `json:"username"`
	}
	if err := getJSON(client, token, "/v1/profile", &profile); err != nil {
		return Identity{}, err
	}

	id := Identity{Email: profile.PrimaryEmail}
	if id.Email == "" {
		id.Email = profile.Username
	}

	// Organizations are a best-effort enrichment; ignore their errors.
	var orgs []struct {
		Name string `json:"name"`
	}
	if err := getJSON(client, token, "/v1/organizations", &orgs); err == nil {
		for _, o := range orgs {
			id.Orgs = append(id.Orgs, o.Name)
		}
	}
	return id, nil
}

func getJSON(client *http.Client, token, path string, v any) error {
	req, err := http.NewRequest(http.MethodGet, apiBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return json.NewDecoder(resp.Body).Decode(v)
	case http.StatusUnauthorized:
		return fmt.Errorf("token is invalid or revoked")
	default:
		return fmt.Errorf("Supabase API returned status %d", resp.StatusCode)
	}
}
