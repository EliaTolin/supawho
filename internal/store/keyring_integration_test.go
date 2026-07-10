//go:build integration

// Integration tests that hit the real OS secret store.
// Run with: go test -tags integration ./internal/store/
// Skipped by default so CI without a Secret Service can still run `go test ./...`.
package store

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

// A dedicated backend on a throwaway service so we never touch real user data.
func testKeyringStore() *Store {
	return newStore(&keyringBackend{service: "supawho-integration-test"})
}

func cleanup(t *testing.T, names ...string) {
	t.Helper()
	svc := "supawho-integration-test"
	for _, n := range names {
		_ = keyring.Delete(svc, n)
	}
	_ = keyring.Delete(svc, indexKey)
}

func TestKeyringRoundTrip(t *testing.T) {
	s := testKeyringStore()
	defer cleanup(t, "acc1", "acc2", "renamed")

	if err := s.Add("acc1", "tok1"); err != nil {
		t.Fatalf("Add acc1: %v", err)
	}
	if err := s.Add("acc2", "tok2"); err != nil {
		t.Fatalf("Add acc2: %v", err)
	}

	tok, err := s.Get("acc1")
	if err != nil || tok != "tok1" {
		t.Fatalf("Get acc1 = %q, %v", tok, err)
	}

	names, err := s.List()
	if err != nil || len(names) != 2 {
		t.Fatalf("List = %v, %v", names, err)
	}

	if err := s.Rename("acc1", "renamed"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if tok, _ := s.Get("renamed"); tok != "tok1" {
		t.Fatalf("renamed token = %q, want tok1", tok)
	}
	if _, err := s.Get("acc1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("acc1 should be gone, err = %v", err)
	}

	if err := s.Remove("renamed"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if err := s.Remove("acc2"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if names, _ := s.List(); len(names) != 0 {
		t.Fatalf("List after removes = %v, want empty", names)
	}
}
