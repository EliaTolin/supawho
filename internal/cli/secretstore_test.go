package cli

import (
	"strings"
	"testing"

	"github.com/EliaTolin/supawho/internal/store"
)

func runFailing(t *testing.T, goos string, err error) string {
	t.Helper()
	h := newHarness("")
	h.app.Store = store.NewFailing(err)
	h.app.GOOS = goos
	if code := h.app.Run([]string{"list"}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	return h.out.String()
}

func TestLinuxNoSecretStoreExplainsPass(t *testing.T) {
	out := runFailing(t, "linux", store.ErrNoSecretStore)
	for _, want := range []string{
		"no secret store is available",
		"export SUPAWHO_BACKEND=pass",
		"https://supawho.dev/guides/headless-linux/",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestLinuxLockedKeyringIsNotMacOS(t *testing.T) {
	out := runFailing(t, "linux", store.ErrLocked)
	if strings.Contains(out, "macOS") || strings.Contains(out, "security unlock-keychain") {
		t.Errorf("Linux user gets macOS advice:\n%s", out)
	}
	for _, want := range []string{"keyring is locked", "SUPAWHO_BACKEND=pass"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPassNotInstalledExplains(t *testing.T) {
	out := runFailing(t, "linux", store.ErrPassNotInstalled)
	for _, want := range []string{"pass is not installed", "pass init"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}
