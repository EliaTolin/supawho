package supabase

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjects(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sbp_test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte(`[
			{"id":"aaa1112223334445556","ref":"aaa1112223334445556","name":"Alpha","organization_slug":"acme"},
			{"id":"bbb1112223334445556","ref":"bbb1112223334445556","name":"Beta","organization_slug":"acme"}
		]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	orig := apiBase
	apiBase = srv.URL
	defer func() { apiBase = orig }()

	got, err := Projects("sbp_test")
	if err != nil {
		t.Fatalf("Projects error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Ref != "aaa1112223334445556" || got[0].Name != "Alpha" || got[0].Org != "acme" {
		t.Fatalf("got[0] = %+v", got[0])
	}
}

// When ref is absent, id is used as the fallback.
func TestProjectsRefFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":"onlyid00000000000000","name":"Gamma"}]`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	orig := apiBase
	apiBase = srv.URL
	defer func() { apiBase = orig }()

	got, err := Projects("t")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(got) != 1 || got[0].Ref != "onlyid00000000000000" {
		t.Fatalf("got = %+v", got)
	}
}
