package store

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakePass puts testdata/fake-pass on PATH as `pass`, backed by a temp dir.
// The store is initialised unless uninit is true.
func fakePass(t *testing.T, uninit bool) (dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake pass is a shell script")
	}
	bin := t.TempDir()
	src, err := filepath.Abs("testdata/fake-pass")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(src, filepath.Join(bin, "pass")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	dir = t.TempDir()
	t.Setenv("FAKE_PASS_DIR", dir)
	t.Setenv("FAKE_PASS_FAIL", "")
	if !uninit {
		if err := os.WriteFile(filepath.Join(dir, ".gpg-id"), []byte("test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestPassRoundTrip(t *testing.T) {
	dir := fakePass(t, false)
	s := NewPass()
	if err := s.Add("client-a", "sbp_secret_a"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := s.Add("client-b", "sbp_secret_b"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	got, err := s.Get("client-a")
	if err != nil || got != "sbp_secret_a" {
		t.Fatalf("Get = %q, %v", got, err)
	}
	names, err := s.List()
	if err != nil || strings.Join(names, ",") != "client-a,client-b" {
		t.Fatalf("List = %v, %v", names, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "supawho", "client-a")); err != nil {
		t.Errorf("entry not stored under the supawho/ prefix: %v", err)
	}
	if err := s.Rename("client-a", "client-c"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if err := s.Remove("client-b"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	names, _ = s.List()
	if strings.Join(names, ",") != "client-c" {
		t.Fatalf("List after rename/remove = %v", names)
	}
}

func TestPassMissingEntryIsNotFound(t *testing.T) {
	fakePass(t, false)
	s := NewPass()
	if _, err := s.Get("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get missing = %v, want ErrNotFound", err)
	}
	if names, err := s.List(); err != nil || names != nil {
		t.Fatalf("List on empty store = %v, %v", names, err)
	}
	if err := s.Remove("nope"); err != nil {
		t.Fatalf("Remove missing = %v, want nil", err)
	}
}

func TestPassNeverPutsSecretsInArgv(t *testing.T) {
	dir := fakePass(t, false)
	s := NewPass()
	if err := s.Add("client-a", "sbp_topsecret"); err != nil {
		t.Fatal(err)
	}
	argv, err := os.ReadFile(filepath.Join(dir, ".argv"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(argv), "sbp_topsecret") {
		t.Fatalf("token passed as an argument:\n%s", argv)
	}
}

func TestPassUninitialisedStoreExplains(t *testing.T) {
	fakePass(t, true)
	err := NewPass().Add("client-a", "sbp_x")
	if err == nil || !strings.Contains(err.Error(), "pass init") {
		t.Fatalf("Add on uninitialised store = %v, want a 'pass init' hint", err)
	}
}

func TestPassErrorsDoNotLeakStdout(t *testing.T) {
	fakePass(t, false)
	s := NewPass()
	if err := s.Add("client-a", "sbp_leak"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FAKE_PASS_FAIL", "gpg: decryption failed: No secret key")
	_, err := s.Get("client-a")
	if err == nil || !strings.Contains(err.Error(), "decryption failed") {
		t.Fatalf("Get = %v, want the gpg error", err)
	}
	if strings.Contains(err.Error(), "sbp_leak") {
		t.Fatalf("error leaks the token: %v", err)
	}
}

func TestPassNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := NewPass().Get("client-a"); !errors.Is(err, ErrPassNotInstalled) {
		t.Fatalf("Get without pass = %v, want ErrPassNotInstalled", err)
	}
}

func TestPassRejectsUnsafeKeys(t *testing.T) {
	dir := fakePass(t, false)
	s := NewPass()
	for _, k := range []string{"", "../escape", "a/b", ".hidden", "-flag"} {
		if err := s.Add(k, "sbp_x"); err == nil {
			t.Errorf("Add(%q) succeeded, want an error", k)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "escape")); err == nil {
		t.Fatal("path traversal wrote outside the supawho/ prefix")
	}
}
