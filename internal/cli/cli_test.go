package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/EliaTolin/supawho/internal/store"
)

type harness struct {
	app    *App
	st     *store.Store
	out    *bytes.Buffer
	logins []string
}

func newHarness(input string) *harness {
	h := &harness{
		st:  store.NewMemory(),
		out: &bytes.Buffer{},
	}
	h.app = &App{
		Store:   h.st,
		Login:   func(tok string) error { h.logins = append(h.logins, tok); return nil },
		In:      strings.NewReader(input),
		Out:     h.out,
		Version: "9.9.9",
	}
	return h
}

func TestUseValid(t *testing.T) {
	h := newHarness("")
	_ = h.st.Add("proj", "sbp_secret")
	if code := h.app.Run([]string{"use", "proj"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if len(h.logins) != 1 || h.logins[0] != "sbp_secret" {
		t.Fatalf("logins = %v, want [sbp_secret]", h.logins)
	}
	if !strings.Contains(h.out.String(), "Logged in as 'proj'.") {
		t.Fatalf("output missing success line:\n%s", h.out.String())
	}
}

func TestUseMissing(t *testing.T) {
	h := newHarness("")
	if code := h.app.Run([]string{"use", "ghost"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if len(h.logins) != 0 {
		t.Fatalf("login should not run for missing account, got %v", h.logins)
	}
	if !strings.Contains(h.out.String(), "not found") {
		t.Fatalf("output missing 'not found':\n%s", h.out.String())
	}
}

func TestUseNoArg(t *testing.T) {
	h := newHarness("")
	if code := h.app.Run([]string{"use"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(h.out.String(), "Usage: supawho use <name>") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestAddValid(t *testing.T) {
	h := newHarness("")
	if code := h.app.Run([]string{"add", "proj", "sbp_tok"}); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	tok, err := h.st.Get("proj")
	if err != nil || tok != "sbp_tok" {
		t.Fatalf("Get = %q, %v", tok, err)
	}
	if !strings.Contains(h.out.String(), "Account 'proj' saved.") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestAddInvalidName(t *testing.T) {
	for _, bad := range []string{"a,b", "a b", "a\tb"} {
		h := newHarness("")
		if code := h.app.Run([]string{"add", bad, "tok"}); code != 1 {
			t.Fatalf("name %q: exit = %d, want 1", bad, code)
		}
		if _, err := h.st.Get(bad); err == nil {
			t.Fatalf("name %q should not have been saved", bad)
		}
		if !strings.Contains(h.out.String(), "may not contain commas or whitespace") {
			t.Fatalf("name %q: output = %q", bad, h.out.String())
		}
	}
}

func TestListEmpty(t *testing.T) {
	h := newHarness("")
	h.app.Run([]string{"list"})
	if !strings.Contains(h.out.String(), "No accounts saved") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestListPopulated(t *testing.T) {
	h := newHarness("")
	_ = h.st.Add("a", "1")
	_ = h.st.Add("b", "2")
	h.app.Run([]string{"list"})
	got := h.out.String()
	if !strings.Contains(got, "  - a") || !strings.Contains(got, "  - b") {
		t.Fatalf("output = %q", got)
	}
}

func TestRemove(t *testing.T) {
	h := newHarness("")
	_ = h.st.Add("a", "1")
	if code := h.app.Run([]string{"remove", "a"}); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if names, _ := h.st.List(); len(names) != 0 {
		t.Fatalf("list = %v, want empty", names)
	}
}

func TestRenameArgs(t *testing.T) {
	h := newHarness("")
	_ = h.st.Add("old", "secret")
	if code := h.app.Run([]string{"rename", "old", "new"}); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	tok, err := h.st.Get("new")
	if err != nil || tok != "secret" {
		t.Fatalf("Get(new) = %q, %v", tok, err)
	}
}

func TestRenameMissing(t *testing.T) {
	h := newHarness("")
	if code := h.app.Run([]string{"rename", "ghost", "new"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(h.out.String(), "not found") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestInteractiveSelect(t *testing.T) {
	h := newHarness("2\n")
	_ = h.st.Add("a", "tok_a")
	_ = h.st.Add("b", "tok_b")
	if code := h.app.Run(nil); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if len(h.logins) != 1 || h.logins[0] != "tok_b" {
		t.Fatalf("logins = %v, want [tok_b]", h.logins)
	}
}

func TestInteractiveInvalidSelection(t *testing.T) {
	for _, in := range []string{"0\n", "5\n", "x\n", "\n"} {
		h := newHarness(in)
		_ = h.st.Add("a", "1")
		_ = h.st.Add("b", "2")
		if code := h.app.Run(nil); code != 1 {
			t.Fatalf("input %q: exit = %d, want 1", in, code)
		}
		if len(h.logins) != 0 {
			t.Fatalf("input %q: no login expected, got %v", in, h.logins)
		}
		if !strings.Contains(h.out.String(), "Invalid selection.") {
			t.Fatalf("input %q: output = %q", in, h.out.String())
		}
	}
}

func TestGuidedAddFlow(t *testing.T) {
	// no accounts + no args → guided add: name, token, then "login now?" (default yes)
	h := newHarness("myproj\nsbp_guided\n\n")
	if code := h.app.Run([]string{"add"}); code != 0 {
		t.Fatalf("exit = %d\noutput:\n%s", code, h.out.String())
	}
	tok, err := h.st.Get("myproj")
	if err != nil || tok != "sbp_guided" {
		t.Fatalf("Get = %q, %v", tok, err)
	}
	if len(h.logins) != 1 || h.logins[0] != "sbp_guided" {
		t.Fatalf("logins = %v, want [sbp_guided]", h.logins)
	}
}

func TestGuidedAddDeclineLogin(t *testing.T) {
	h := newHarness("myproj\nsbp_guided\nn\n")
	if code := h.app.Run([]string{"add"}); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if len(h.logins) != 0 {
		t.Fatalf("declined login should not run, got %v", h.logins)
	}
}

func TestGuidedAddAbortEmptyName(t *testing.T) {
	h := newHarness("\n")
	if code := h.app.Run([]string{"add"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(h.out.String(), "Aborted.") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	h := newHarness("")
	if code := h.app.Run([]string{"frobnicate"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(h.out.String(), "Unknown command: frobnicate") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestWhoamiAll(t *testing.T) {
	h := newHarness("")
	_ = h.st.Add("proj-a", "tok_a")
	_ = h.st.Add("proj-b", "tok_b")
	h.app.Profile = func(token string) (string, []string, error) {
		switch token {
		case "tok_a":
			return "a@example.com", []string{"Org A"}, nil
		case "tok_b":
			return "b@example.com", nil, nil
		}
		return "", nil, errors.New("unknown")
	}
	if code := h.app.Run([]string{"whoami"}); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	out := h.out.String()
	if !strings.Contains(out, "proj-a") || !strings.Contains(out, "a@example.com") || !strings.Contains(out, "Org A") {
		t.Fatalf("output missing proj-a row:\n%s", out)
	}
	if !strings.Contains(out, "proj-b") || !strings.Contains(out, "b@example.com") {
		t.Fatalf("output missing proj-b row:\n%s", out)
	}
}

func TestWhoamiSingle(t *testing.T) {
	h := newHarness("")
	_ = h.st.Add("proj-a", "tok_a")
	_ = h.st.Add("proj-b", "tok_b")
	var asked []string
	h.app.Profile = func(token string) (string, []string, error) {
		asked = append(asked, token)
		return "a@example.com", nil, nil
	}
	if code := h.app.Run([]string{"whoami", "proj-a"}); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if len(asked) != 1 || asked[0] != "tok_a" {
		t.Fatalf("expected only proj-a queried, got %v", asked)
	}
}

func TestWhoamiMissing(t *testing.T) {
	h := newHarness("")
	h.app.Profile = func(string) (string, []string, error) { return "", nil, nil }
	if code := h.app.Run([]string{"whoami", "ghost"}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(h.out.String(), "not found") {
		t.Fatalf("output = %q", h.out.String())
	}
}

func TestWhoamiHandlesPerAccountError(t *testing.T) {
	h := newHarness("")
	_ = h.st.Add("good", "ok")
	_ = h.st.Add("revoked", "bad")
	h.app.Profile = func(token string) (string, []string, error) {
		if token == "bad" {
			return "", nil, errors.New("token is invalid or revoked")
		}
		return "good@example.com", nil, nil
	}
	if code := h.app.Run([]string{"whoami"}); code != 0 {
		t.Fatalf("exit = %d (one bad token must not fail the whole command)", code)
	}
	out := h.out.String()
	if !strings.Contains(out, "good@example.com") || !strings.Contains(out, "invalid or revoked") {
		t.Fatalf("output = %q", out)
	}
}

func TestVersionGolden(t *testing.T) {
	h := newHarness("")
	h.app.Run([]string{"version"})
	if got := h.out.String(); got != "supawho 9.9.9\n" {
		t.Fatalf("version output = %q", got)
	}
}

const wantHelp = "\n" +
	"\x1b[1;32m   ___  _   _ ___  ___  _    _ _  _  ___\x1b[0m\n" +
	"\x1b[1;32m  / __|| | | | _ \\/ _ \\| |  | | || |/ _ \\\x1b[0m\n" +
	"\x1b[1;32m  \\__ \\| |_| |  _/ (_) | |/\\| | __ | (_)\x1b[0m\n" +
	"\x1b[1;32m  |___/ \\___/|_|  \\__\\_\\__/\\__|_||_|\\___/\x1b[0m\n" +
	"\n     🔍 Who are you today?\n\n" +
	"Usage:\n" +
	"  supawho                        Interactive account selection\n" +
	"  supawho add <name> <token>     Save account securely\n" +
	"  supawho rename <old> <new>     Rename a saved account\n" +
	"  supawho remove <name>          Remove account\n" +
	"  supawho list                   List saved accounts\n" +
	"  supawho use <name>             Login with specific account\n" +
	"  supawho whoami [name]          Show the email behind each account\n" +
	"  supawho upgrade                Update supawho to the latest version\n" +
	"  supawho version                Show version\n" +
	"\n" +
	"Tokens are stored securely in your OS secret store.\n"

func TestHelpGolden(t *testing.T) {
	h := newHarness("")
	h.app.Run([]string{"help"})
	if got := h.out.String(); got != wantHelp {
		t.Fatalf("help output mismatch.\n got: %q\nwant: %q", got, wantHelp)
	}
}
