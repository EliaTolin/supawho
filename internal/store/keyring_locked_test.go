package store

import (
	"errors"
	"os/exec"
	"testing"
)

// exitErr returns a real *exec.ExitError carrying the given exit code.
func exitErr(t *testing.T, code string) error {
	t.Helper()
	err := exec.Command("sh", "-c", "exit "+code).Run()
	if err == nil {
		t.Fatalf("expected exit error for code %s", code)
	}
	return err
}

func TestTranslateKeychainLocked(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	locked := exitErr(t, "36")
	other := exitErr(t, "1")
	plain := errors.New("boom")

	if got := translate("darwin", locked); !errors.Is(got, ErrLocked) {
		t.Errorf("darwin exit 36: got %v, want ErrLocked", got)
	}
	if got := translate("darwin", other); errors.Is(got, ErrLocked) || got != other {
		t.Errorf("darwin exit 1: got %v, want it unchanged", got)
	}
	if got := translate("linux", locked); got != locked {
		t.Errorf("linux exit 36: got %v, want it unchanged", got)
	}
	if got := translate("darwin", plain); got != plain {
		t.Errorf("plain error: got %v, want it unchanged", got)
	}
	if got := translate("darwin", nil); got != nil {
		t.Errorf("nil: got %v", got)
	}
}
