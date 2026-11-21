package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// envHarness runs `supawho env` with stdout and stderr captured separately,
// in dir as the working directory.
func envHarness(t *testing.T, dir string) (*harness, *bytes.Buffer) {
	t.Helper()
	h := newHarness("")
	errOut := &bytes.Buffer{}
	h.app.Err = errOut
	h.app.getwd = func() (string, error) { return dir, nil }
	return h, errOut
}

func TestEnvPrintsPosixExport(t *testing.T) {
	h, errOut := envHarness(t, t.TempDir())
	_ = h.st.Add("client-a", "sbp_aaa")
	if code := h.app.Run([]string{"env", "client-a"}); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, errOut)
	}
	if got, want := h.out.String(), "export SUPABASE_ACCESS_TOKEN='sbp_aaa'\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut)
	}
}

func TestEnvShells(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"--shell", "zsh"}, "export SUPABASE_ACCESS_TOKEN='a'\\''b'\n"},
		{[]string{"--shell=bash"}, "export SUPABASE_ACCESS_TOKEN='a'\\''b'\n"},
		{[]string{"--shell", "fish"}, "set -gx SUPABASE_ACCESS_TOKEN 'a\\'b'\n"},
		{[]string{"--shell", "powershell"}, "$env:SUPABASE_ACCESS_TOKEN = 'a''b'\n"},
		{[]string{"--shell", "pwsh"}, "$env:SUPABASE_ACCESS_TOKEN = 'a''b'\n"},
	}
	for _, c := range cases {
		h, errOut := envHarness(t, t.TempDir())
		_ = h.st.Add("x", "a'b")
		if code := h.app.Run(append([]string{"env", "x"}, c.args...)); code != 0 {
			t.Fatalf("%v: exit = %d, stderr: %s", c.args, code, errOut)
		}
		if got := h.out.String(); got != c.want {
			t.Fatalf("%v: stdout = %q, want %q", c.args, got, c.want)
		}
	}
}

func TestEnvFishEscapesBackslash(t *testing.T) {
	h, _ := envHarness(t, t.TempDir())
	_ = h.st.Add("x", `a\b`)
	h.app.Run([]string{"env", "x", "--shell", "fish"})
	if got, want := h.out.String(), `set -gx SUPABASE_ACCESS_TOKEN 'a\\b'`+"\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

// failsQuietly asserts the exit code is 1, nothing reached stdout (it is
// meant for eval) and stderr explains the problem.
func failsQuietly(t *testing.T, h *harness, errOut *bytes.Buffer, args []string, wantErr string) {
	t.Helper()
	if code := h.app.Run(args); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if h.out.Len() != 0 {
		t.Fatalf("stdout = %q, want empty: it is evaluated by the shell", h.out)
	}
	if !strings.Contains(errOut.String(), wantErr) {
		t.Fatalf("stderr = %q, want it to mention %q", errOut, wantErr)
	}
}

func TestEnvUnknownAccount(t *testing.T) {
	h, errOut := envHarness(t, t.TempDir())
	failsQuietly(t, h, errOut, []string{"env", "ghost"}, "Account 'ghost' not found")
}

func TestEnvUnknownShell(t *testing.T) {
	h, errOut := envHarness(t, t.TempDir())
	_ = h.st.Add("x", "sbp_x")
	failsQuietly(t, h, errOut, []string{"env", "x", "--shell", "tcsh"}, `unknown shell "tcsh"`)
}

func TestEnvRejectsTokenWithNewline(t *testing.T) {
	h, errOut := envHarness(t, t.TempDir())
	_ = h.st.Add("x", "sbp_x\nrm -rf ~")
	failsQuietly(t, h, errOut, []string{"env", "x"}, "unexpected characters")
}

func TestEnvRefusesTerminal(t *testing.T) {
	h, errOut := envHarness(t, t.TempDir())
	h.app.isTerminal = func(io.Writer) bool { return true }
	_ = h.st.Add("x", "sbp_x")
	failsQuietly(t, h, errOut, []string{"env", "x"}, `eval "$(supawho env x)"`)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnvReadsNearestSupawhoFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".supawho"), "# account for this repo\n\n  client-b  \n")
	sub := filepath.Join(root, "apps", "web")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	h, errOut := envHarness(t, sub)
	_ = h.st.Add("client-b", "sbp_bbb")
	if code := h.app.Run([]string{"env"}); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, errOut)
	}
	if got, want := h.out.String(), "export SUPABASE_ACCESS_TOKEN='sbp_bbb'\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestEnvSupawhoFileUnknownAccountNamesTheFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".supawho"), "$(touch pwned)\n")
	h, errOut := envHarness(t, root)
	failsQuietly(t, h, errOut, []string{"env"}, filepath.Join(root, ".supawho"))
}

func TestEnvWithoutNameOrFile(t *testing.T) {
	h, errOut := envHarness(t, t.TempDir())
	failsQuietly(t, h, errOut, []string{"env"}, "no .supawho file")
}

func TestEnvEmptySupawhoFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".supawho"), "# nothing\n")
	h, errOut := envHarness(t, root)
	failsQuietly(t, h, errOut, []string{"env"}, "has no account name")
}

func TestWhoamiShowsActiveSourceAndFolderAccount(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".supawho"), "client-a\n")
	h, _ := envHarness(t, root)
	h.app.Profile = func(string) (string, []string, error) { return "me@example.com", nil, nil }
	h.app.ActiveToken = func() (string, string, error) { return "sbp_bbb", "env", nil }
	_ = h.st.Add("client-a", "sbp_aaa")
	_ = h.st.Add("client-b", "sbp_bbb")
	if code := h.app.Run([]string{"whoami"}); code != 0 {
		t.Fatalf("exit = %d\n%s", code, h.out)
	}
	out := h.out.String()
	for _, want := range []string{
		"Active: client-b (from SUPABASE_ACCESS_TOKEN)",
		"This folder: client-a (" + filepath.Join(root, ".supawho") + ")",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestWhoamiActiveTokenNotSaved(t *testing.T) {
	h, _ := envHarness(t, t.TempDir())
	h.app.Profile = func(string) (string, []string, error) { return "me@example.com", nil, nil }
	h.app.ActiveToken = func() (string, string, error) { return "sbp_other", "keychain", nil }
	_ = h.st.Add("client-a", "sbp_aaa")
	h.app.Run([]string{"whoami"})
	if out := h.out.String(); !strings.Contains(out, "Active: a token not saved in supawho (from the Supabase CLI login)") {
		t.Fatalf("output:\n%s", out)
	}
	if strings.Contains(h.out.String(), "This folder") {
		t.Fatalf("no .supawho here, but output mentions it:\n%s", h.out)
	}
}
