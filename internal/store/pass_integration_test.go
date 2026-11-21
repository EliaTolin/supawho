//go:build pass_integration

// Integration test against a real pass + gpg, in a throwaway GNUPGHOME and
// PASSWORD_STORE_DIR so it never touches the user's keys or store.
// Run with: go test -tags pass_integration -run TestRealPass ./internal/store/
package store

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRealPassRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("pass"); err != nil {
		t.Skip("pass not installed")
	}
	// gpg-agent's socket path must stay short, so avoid the long macOS TMPDIR.
	gnupg, err := os.MkdirTemp("/tmp", "supawho-gnupg-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", gnupg, "--kill", "gpg-agent").Run()
		_ = os.RemoveAll(gnupg)
	})
	t.Setenv("GNUPGHOME", gnupg)
	t.Setenv("PASSWORD_STORE_DIR", t.TempDir())

	run := func(name string, args ...string) string {
		t.Helper()
		out, err := exec.Command(name, args...).CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
		return string(out)
	}
	run("gpg", "--batch", "--passphrase", "", "--quick-gen-key", "supawho-test@example.invalid", "default", "default", "never")
	run("pass", "init", "supawho-test@example.invalid")

	s := NewPass()
	if err := s.Add("client-a", "sbp_integration_a"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got, err := s.Get("client-a"); err != nil || got != "sbp_integration_a" {
		t.Fatalf("Get = %q, %v", got, err)
	}
	if err := s.Rename("client-a", "client-b"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if names, err := s.List(); err != nil || strings.Join(names, ",") != "client-b" {
		t.Fatalf("List = %v, %v", names, err)
	}
	if _, err := s.Get("client-a"); err != ErrNotFound {
		t.Fatalf("Get renamed-away = %v, want ErrNotFound", err)
	}
	if err := s.Remove("client-b"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if names, _ := s.List(); names != nil {
		t.Fatalf("List after remove = %v", names)
	}
}
