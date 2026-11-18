package ui

import "testing"

func TestColorize(t *testing.T) {
	if got := colorize(false, green, "x"); got != "x" {
		t.Fatalf("color off = %q, want %q", got, "x")
	}
	if got := colorize(true, green, "x"); got != green+"x"+reset {
		t.Fatalf("color on = %q, want %q", got, green+"x"+reset)
	}
}

func TestColorEnabledOnBuffer(t *testing.T) {
	// A non-*os.File writer is never a terminal, so color must stay off —
	// this is what keeps piped/redirected output free of escape codes.
	if colorEnabled(&countWriter{}) {
		t.Fatal("colorEnabled must be false for a non-terminal writer")
	}
}

type countWriter struct{ n int }

func (c *countWriter) Write(p []byte) (int, error) { c.n += len(p); return len(p), nil }
