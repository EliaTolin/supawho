package cli

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/EliaTolin/supawho/internal/store"
)

func projectsHarness() *harness {
	h := newHarness("")
	_ = h.st.Add("work", "tok_work")
	_ = h.st.Add("personal", "tok_personal")
	h.app.Lookup = func(token string) ([]Project, error) {
		switch token {
		case "tok_work":
			return []Project{
				{Ref: "ref_zeta", Name: "zeta", Org: "Acme", Region: "eu-west-1", Status: "ACTIVE_HEALTHY"},
				{Ref: "ref_alpha", Name: "alpha", Org: "Acme", Region: "eu-west-1", Status: "INACTIVE"},
			}, nil
		case "tok_personal":
			return []Project{{Ref: "ref_blog", Name: "blog", Org: "Me", Region: "us-east-1", Status: "ACTIVE_HEALTHY"}}, nil
		}
		return nil, errors.New("token is invalid or revoked")
	}
	return h
}

func TestProjectsAll(t *testing.T) {
	h := projectsHarness()
	if code := h.app.Run([]string{"projects"}); code != 0 {
		t.Fatalf("exit = %d\n%s", code, h.out.String())
	}
	out := h.out.String()
	for _, want := range []string{"work (2)", "personal (1)", "ref_alpha", "ref_zeta", "ref_blog", "Acme", "us-east-1", "INACTIVE"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "alpha") > strings.Index(out, "zeta") {
		t.Fatalf("projects are not sorted by name:\n%s", out)
	}
	if strings.Index(out, "work (2)") > strings.Index(out, "personal (1)") {
		t.Fatalf("accounts are not in saved order:\n%s", out)
	}
}

func TestProjectsSingle(t *testing.T) {
	h := projectsHarness()
	if code := h.app.Run([]string{"projects", "personal"}); code != 0 {
		t.Fatalf("exit = %d\n%s", code, h.out.String())
	}
	out := h.out.String()
	if !strings.Contains(out, "ref_blog") || strings.Contains(out, "ref_alpha") {
		t.Fatalf("output = %q", out)
	}
}

func TestProjectsMissingAccount(t *testing.T) {
	h := projectsHarness()
	if code := h.app.Run([]string{"projects", "ghost"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(h.out.String(), "Account 'ghost' not found") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestProjectsNoAccounts(t *testing.T) {
	h := newHarness("")
	h.app.Lookup = func(string) ([]Project, error) { return nil, nil }
	if code := h.app.Run([]string{"projects"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(h.out.String(), "No accounts saved") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestProjectsEmptyAccount(t *testing.T) {
	h := newHarness("")
	_ = h.st.Add("fresh", "tok")
	h.app.Lookup = func(string) ([]Project, error) { return nil, nil }
	if code := h.app.Run([]string{"projects"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(h.out.String(), "No projects") {
		t.Fatalf("output = %q", h.out.String())
	}
}

// A failing account is reported, the others are still listed, and the
// command fails so scripts notice the listing is incomplete.
func TestProjectsFailingAccount(t *testing.T) {
	h := projectsHarness()
	_ = h.st.Add("revoked", "tok_bad")
	if code := h.app.Run([]string{"projects"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	out := h.out.String()
	if !strings.Contains(out, "ref_blog") || !strings.Contains(out, "token is invalid or revoked") {
		t.Fatalf("output = %q", out)
	}
}

func TestProjectsKeychainLocked(t *testing.T) {
	h := newHarness("")
	h.app.Store = store.NewFailing(store.ErrLocked)
	h.app.Lookup = func(string) ([]Project, error) { return nil, nil }
	if code := h.app.Run([]string{"projects"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(h.out.String(), "Keychain is locked") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestProjectsUsage(t *testing.T) {
	for _, args := range [][]string{{"projects", "a", "b"}, {"projects", "--nope"}} {
		h := projectsHarness()
		if code := h.app.Run(args); code != 1 {
			t.Fatalf("%v: exit = %d, want 1", args, code)
		}
		if !strings.Contains(h.out.String(), "Usage: supawho projects [name] [--json]") {
			t.Fatalf("%v: output = %q", args, h.out.String())
		}
	}
}

type projectsDoc struct {
	Accounts []struct {
		Name     string `json:"name"`
		Error    string `json:"error"`
		Projects []struct {
			Ref          string `json:"ref"`
			Name         string `json:"name"`
			Organization string `json:"organization"`
			Region       string `json:"region"`
			Status       string `json:"status"`
		} `json:"projects"`
	} `json:"accounts"`
	Error string `json:"error"`
}

func TestProjectsJSON(t *testing.T) {
	h := projectsHarness()
	_ = h.st.Add("revoked", "tok_bad")
	if code := h.app.Run([]string{"projects", "--json"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	var doc projectsDoc
	if err := json.Unmarshal(h.out.Bytes(), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, h.out.String())
	}
	if len(doc.Accounts) != 3 {
		t.Fatalf("accounts = %d, want 3", len(doc.Accounts))
	}
	w := doc.Accounts[0]
	if w.Name != "work" || len(w.Projects) != 2 || w.Projects[0].Ref != "ref_alpha" || w.Projects[0].Organization != "Acme" {
		t.Fatalf("work = %+v", w)
	}
	bad := doc.Accounts[2]
	if bad.Name != "revoked" || bad.Error == "" || bad.Projects == nil || len(bad.Projects) != 0 {
		t.Fatalf("revoked = %+v", bad)
	}
}

func TestProjectsJSONErrors(t *testing.T) {
	h := projectsHarness()
	if code := h.app.Run([]string{"projects", "ghost", "--json"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	var doc projectsDoc
	if err := json.Unmarshal(h.out.Bytes(), &doc); err != nil || doc.Error == "" {
		t.Fatalf("want a JSON error object, got %q (%v)", h.out.String(), err)
	}

	h = newHarness("")
	h.app.Lookup = func(string) ([]Project, error) { return nil, nil }
	if code := h.app.Run([]string{"projects", "--json"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	doc = projectsDoc{}
	if err := json.Unmarshal(h.out.Bytes(), &doc); err != nil || doc.Accounts == nil || len(doc.Accounts) != 0 {
		t.Fatalf("want an empty accounts array, got %q (%v)", h.out.String(), err)
	}
}
