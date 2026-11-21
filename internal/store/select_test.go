package store

import (
	"strings"
	"testing"
)

func TestOpenSelectsBackend(t *testing.T) {
	for value, want := range map[string]string{
		"":        BackendKeyring,
		"keyring": BackendKeyring,
		"pass":    BackendPass,
		" PASS ":  BackendPass,
	} {
		s, err := Open(value)
		if err != nil {
			t.Fatalf("Open(%q): %v", value, err)
		}
		if got := s.Backend(); got != want {
			t.Errorf("Open(%q).Backend() = %q, want %q", value, got, want)
		}
	}
}

func TestOpenRejectsUnknownBackend(t *testing.T) {
	_, err := Open("file")
	if err == nil || !strings.Contains(err.Error(), "SUPAWHO_BACKEND") {
		t.Fatalf("Open(file) = %v, want an error naming SUPAWHO_BACKEND", err)
	}
}

func TestMemoryBackendName(t *testing.T) {
	if got := NewMemory().Backend(); got != "memory" {
		t.Errorf("memory Backend() = %q", got)
	}
}
