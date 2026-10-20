package supabase

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// fakeCLI installs an executable "supabase" script and points lookPath at it.
func fakeCLI(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell scripts are not executable on windows")
	}
	path := filepath.Join(t.TempDir(), "supabase")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := lookPath
	lookPath = func(string) (string, error) { return path, nil }
	t.Cleanup(func() { lookPath = orig })
	return path
}

func TestCLIInfoFound(t *testing.T) {
	want := fakeCLI(t, "echo 2.48.3")
	path, version, err := CLIInfo()
	if err != nil || path != want || version != "2.48.3" {
		t.Fatalf("CLIInfo() = %q, %q, %v; want %q, 2.48.3, nil", path, version, err, want)
	}
}

func TestCLIInfoBadVersion(t *testing.T) {
	want := fakeCLI(t, "exit 1")
	path, version, err := CLIInfo()
	if err != nil || path != want || version != "unknown" {
		t.Fatalf("CLIInfo() = %q, %q, %v; want %q, unknown, nil", path, version, err, want)
	}
}

func TestCLIInfoMissing(t *testing.T) {
	orig := lookPath
	lookPath = func(string) (string, error) { return "", exec.ErrNotFound }
	t.Cleanup(func() { lookPath = orig })
	path, _, err := CLIInfo()
	if err == nil || path != "" {
		t.Fatalf("CLIInfo() = %q, _, %v; want \"\", error", path, err)
	}
}
