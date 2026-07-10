package store

import (
	"errors"
	"reflect"
	"testing"
)

func mustList(t *testing.T, s *Store) []string {
	t.Helper()
	names, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	return names
}

func TestEmptyList(t *testing.T) {
	s := NewMemory()
	if names := mustList(t, s); names != nil {
		t.Fatalf("expected nil list, got %v", names)
	}
}

func TestAddThenList(t *testing.T) {
	s := NewMemory()
	if err := s.Add("proj", "sbp_token"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got := mustList(t, s); !reflect.DeepEqual(got, []string{"proj"}) {
		t.Fatalf("list = %v, want [proj]", got)
	}
	tok, err := s.Get("proj")
	if err != nil || tok != "sbp_token" {
		t.Fatalf("Get = %q, %v; want sbp_token", tok, err)
	}
}

func TestAddDuplicateNoDoubleIndex(t *testing.T) {
	s := NewMemory()
	_ = s.Add("proj", "tok1")
	_ = s.Add("proj", "tok2")
	if got := mustList(t, s); !reflect.DeepEqual(got, []string{"proj"}) {
		t.Fatalf("list = %v, want single [proj]", got)
	}
	tok, _ := s.Get("proj")
	if tok != "tok2" {
		t.Fatalf("token = %q, want tok2 (overwritten)", tok)
	}
}

func TestAddPreservesOrder(t *testing.T) {
	s := NewMemory()
	_ = s.Add("a", "1")
	_ = s.Add("b", "2")
	_ = s.Add("c", "3")
	if got := mustList(t, s); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("list = %v, want [a b c]", got)
	}
}

func TestRename(t *testing.T) {
	s := NewMemory()
	_ = s.Add("a", "1")
	_ = s.Add("old", "secret")
	_ = s.Add("c", "3")

	if err := s.Rename("old", "new"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	// position preserved
	if got := mustList(t, s); !reflect.DeepEqual(got, []string{"a", "new", "c"}) {
		t.Fatalf("list = %v, want [a new c]", got)
	}
	// token carried over
	tok, err := s.Get("new")
	if err != nil || tok != "secret" {
		t.Fatalf("Get(new) = %q, %v; want secret", tok, err)
	}
	// old gone
	if _, err := s.Get("old"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(old) err = %v, want ErrNotFound", err)
	}
}

func TestRenameMissing(t *testing.T) {
	s := NewMemory()
	if err := s.Rename("ghost", "new"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Rename missing err = %v, want ErrNotFound", err)
	}
}

func TestRemove(t *testing.T) {
	s := NewMemory()
	_ = s.Add("a", "1")
	_ = s.Add("b", "2")

	if err := s.Remove("a"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if got := mustList(t, s); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("list = %v, want [b]", got)
	}
	if _, err := s.Get("a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(a) err = %v, want ErrNotFound", err)
	}
}

func TestRemoveNonexistentNoError(t *testing.T) {
	s := NewMemory()
	_ = s.Add("a", "1")
	if err := s.Remove("ghost"); err != nil {
		t.Fatalf("Remove ghost err = %v, want nil", err)
	}
	if got := mustList(t, s); !reflect.DeepEqual(got, []string{"a"}) {
		t.Fatalf("list = %v, want [a]", got)
	}
}

func TestRemoveLastEmptiesList(t *testing.T) {
	s := NewMemory()
	_ = s.Add("only", "1")
	if err := s.Remove("only"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if got := mustList(t, s); got != nil {
		t.Fatalf("list = %v, want nil", got)
	}
}
