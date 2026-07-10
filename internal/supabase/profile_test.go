package supabase

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhoami(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sbp_test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte(`{"primary_email":"me@example.com","username":"me@example.com"}`))
	})
	mux.HandleFunc("/v1/organizations", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"Acme"},{"name":"Beta"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	orig := apiBase
	apiBase = srv.URL
	defer func() { apiBase = orig }()

	id, err := Whoami("sbp_test")
	if err != nil {
		t.Fatalf("Whoami error: %v", err)
	}
	if id.Email != "me@example.com" {
		t.Fatalf("email = %q, want me@example.com", id.Email)
	}
	if len(id.Orgs) != 2 || id.Orgs[0] != "Acme" || id.Orgs[1] != "Beta" {
		t.Fatalf("orgs = %v, want [Acme Beta]", id.Orgs)
	}
}

func TestWhoamiUnauthorized(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/profile", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	orig := apiBase
	apiBase = srv.URL
	defer func() { apiBase = orig }()

	if _, err := Whoami("revoked"); err == nil {
		t.Fatal("expected error for revoked token")
	}
}

// Orgs are best-effort: a failing organizations endpoint must not fail Whoami.
func TestWhoamiOrgsBestEffort(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/profile", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"primary_email":"me@example.com"}`))
	})
	mux.HandleFunc("/v1/organizations", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	orig := apiBase
	apiBase = srv.URL
	defer func() { apiBase = orig }()

	id, err := Whoami("tok")
	if err != nil {
		t.Fatalf("orgs failure should not fail Whoami: %v", err)
	}
	if id.Email != "me@example.com" || len(id.Orgs) != 0 {
		t.Fatalf("got %+v", id)
	}
}
