package supabase

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// lookPath is a var so tests can point it at a fake binary.
var lookPath = exec.LookPath

// CLIInfo locates the supabase CLI in PATH and reports its version. A binary
// that is found but fails `--version` yields version "unknown", not an error.
func CLIInfo() (path, version string, err error) {
	path, err = lookPath("supabase")
	if err != nil {
		return "", "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--version")
	// Don't hang if a wrapper script leaves a child holding stdout open.
	cmd.WaitDelay = time.Second
	out, err := cmd.Output()
	version = strings.TrimSpace(string(out))
	if err != nil || version == "" {
		version = "unknown"
	}
	return path, version, nil
}
