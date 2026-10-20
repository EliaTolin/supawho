package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EliaTolin/supawho/internal/cli/ui"
	"github.com/EliaTolin/supawho/internal/store"
	"github.com/EliaTolin/supawho/internal/supabase"
)

// newDoctorHarness saves pairs of (name, token) in order and wires fakes that
// pass every check. Token conventions for the fake Profile:
//
//	sbp_bad → ErrUnauthorized, sbp_net → network error echoing the token,
//	anything else → email "<name>@x.io" where name is the account owning it.
func newDoctorHarness(pairs ...string) *harness {
	h := newHarness("")
	owner := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		_ = h.st.Add(pairs[i], pairs[i+1])
		owner[pairs[i+1]] = pairs[i]
	}
	h.app.Profile = func(token string) (string, []string, error) {
		switch token {
		case "sbp_bad":
			return "", nil, supabase.ErrUnauthorized
		case "sbp_net":
			return "", nil, errors.New("dial tcp: timeout for sbp_net")
		}
		return owner[token] + "@x.io", nil, nil
	}
	h.app.CLIInfo = func() (string, string, error) { return "/bin/supabase", "2.48.3", nil }
	h.app.ActiveToken = func() (string, string, error) {
		if len(pairs) >= 2 {
			return pairs[1], "keychain", nil
		}
		return "", "", supabase.ErrNoActiveToken
	}
	h.app.Latest = func(context.Context) (string, error) { return "9.9.9", nil }
	fixed := time.Unix(0, 0)
	h.app.now = func() time.Time { return fixed }
	return h
}

func (h *harness) mustContain(t *testing.T, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(h.out.String(), s) {
			t.Fatalf("output missing %q:\n%s", s, h.out.String())
		}
	}
}

func (h *harness) run(t *testing.T, wantCode int, args ...string) {
	t.Helper()
	if code := h.app.Run(args); code != wantCode {
		t.Fatalf("exit = %d, want %d\noutput:\n%s", code, wantCode, h.out.String())
	}
}

func TestScore(t *testing.T) {
	cases := []struct{ pass, warn, total, want int }{
		{6, 1, 8, 81}, {3, 0, 3, 100}, {0, 0, 0, 100}, {0, 1, 2, 25}, {1, 1, 4, 38},
	}
	for _, c := range cases {
		if got := score(c.pass, c.warn, c.total); got != c.want {
			t.Errorf("score(%d, %d, %d) = %d, want %d", c.pass, c.warn, c.total, got, c.want)
		}
	}
}

// specReport is the example from the design spec §2.2/§2.3.
func specReport() report {
	return report{
		Accounts: []check{
			{Name: "acme", Status: statusPass, Detail: "mario@acme.it", Latency: 142 * time.Millisecond},
			{Name: "old-client", Status: statusFail, Detail: "token is invalid or revoked", Hint: "supawho remove old-client"},
		},
		System: []check{
			{Name: "supabase_cli", Status: statusPass, Detail: "2.48.3 /opt/homebrew/bin/supabase"},
			{Name: "supawho_version", Status: statusWarn, Detail: "1.6.2 -> 1.7.0 available", Hint: "supawho upgrade"},
		},
		Pass: 6, Warn: 1, Fail: 1, Score: 81,
	}
}

func TestRenderPlainGolden(t *testing.T) {
	var buf bytes.Buffer
	renderPlain(&buf, specReport())
	want := "Accounts\n" +
		"PASS  acme        mario@acme.it  142ms\n" +
		"FAIL  old-client  token is invalid or revoked\n" +
		"      -> supawho remove old-client\n" +
		"System\n" +
		"PASS  Supabase CLI    2.48.3 /opt/homebrew/bin/supabase\n" +
		"WARN  supawho         1.6.2 -> 1.7.0 available\n" +
		"      -> supawho upgrade\n" +
		"\n" +
		"6 passed, 1 warning, 1 failed (score 81%)\n"
	if got := buf.String(); got != want {
		t.Fatalf("plain output =\n%s\nwant\n%s", got, want)
	}
}

func TestRenderJSONShape(t *testing.T) {
	var buf bytes.Buffer
	if err := renderJSON(&buf, specReport()); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	accounts := got["accounts"].([]any)
	first, second := accounts[0].(map[string]any), accounts[1].(map[string]any)
	if first["latency_ms"] != float64(142) || first["status"] != "pass" {
		t.Fatalf("accounts[0] = %v", first)
	}
	if _, ok := second["latency_ms"]; ok || second["hint"] != "supawho remove old-client" {
		t.Fatalf("accounts[1] = %v", second)
	}
	if _, ok := first["hint"]; ok {
		t.Fatalf("empty hint must be omitted: %v", first)
	}
	summary := got["summary"].(map[string]any)
	if summary["score"] != float64(81) || summary["fail"] != float64(1) {
		t.Fatalf("summary = %v", summary)
	}
}

func TestDoctorAllPass(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	h.run(t, 0, "doctor")
	h.mustContain(t, "PASS  a  a@x.io", "PASS  Active account  a (keychain)", "PASS  Supabase CLI", "9.9.9 (latest)", "(score 100%)")
}

func TestDoctorUnauthorized(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a", "b", "sbp_bad")
	h.run(t, 1, "doctor")
	h.mustContain(t, "FAIL  b  token is invalid or revoked", "-> supawho remove b && supawho add b <new-token>")
}

func TestDoctorNetworkWarn(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a", "c", "sbp_net")
	h.run(t, 0, "doctor")
	h.mustContain(t, "WARN  c  could not verify")
}

func TestDoctorMissingToken(t *testing.T) {
	h := newDoctorHarness()
	h.app.Store = store.NewMemoryWithOrphans("ghost")
	h.run(t, 1, "doctor")
	h.mustContain(t, "FAIL  ghost  token missing from the OS secret store")
}

func TestDoctorSingleAccount(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a", "b", "sbp_bad")
	h.run(t, 0, "doctor", "a")
	h.mustContain(t, "PASS  a")
	if strings.Contains(h.out.String(), "  b  ") {
		t.Fatalf("doctor a must not check b:\n%s", h.out.String())
	}
}

func TestDoctorArgs(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"doctor", "ghost"}, "Account 'ghost' not found"},
		{[]string{"doctor", "--foo"}, "Usage: supawho doctor [name] [--json]"},
		{[]string{"doctor", "a", "b"}, "Usage: supawho doctor [name] [--json]"},
	}
	for _, c := range cases {
		h := newDoctorHarness("a", "sbp_ok_a")
		h.run(t, 1, c.args...)
		h.mustContain(t, c.want)
	}
}

func TestDoctorCLIMissing(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	h.app.CLIInfo = func() (string, string, error) { return "", "", errors.New("not found") }
	h.run(t, 1, "doctor")
	h.mustContain(t, "FAIL  Supabase CLI    supabase CLI not found in PATH", "-> install: https://supabase.com/docs/guides/cli")
}

func TestDoctorActiveStates(t *testing.T) {
	cases := []struct {
		name   string
		active func() (string, string, error)
		code   int
		want   []string
	}{
		{"not logged in", func() (string, string, error) { return "", "", supabase.ErrNoActiveToken }, 1,
			[]string{"WARN  Active account  not logged in", "-> supawho use <name>"}},
		{"unreadable", func() (string, string, error) { return "", "", errors.New("boom") }, 1,
			[]string{"WARN  Active account  could not read active token"}},
		{"not saved", func() (string, string, error) { return "sbp_other", "env", nil }, 1,
			[]string{"WARN  Active account  active token is not saved in supawho (env)", "-> supawho add <name> <token>"}},
		{"invalid", func() (string, string, error) { return "sbp_bad", "keychain", nil }, 1,
			[]string{"FAIL  Active account  b (keychain) — token is invalid", "-> supawho use a"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// b is always FAIL, so every case exits 1; the active row is what's asserted.
			h := newDoctorHarness("a", "sbp_ok_a", "b", "sbp_bad")
			h.app.ActiveToken = c.active
			h.run(t, c.code, "doctor")
			h.mustContain(t, c.want...)
		})
	}
}

func TestDoctorActiveInvalidNoAlternative(t *testing.T) {
	h := newDoctorHarness("b", "sbp_bad")
	h.run(t, 1, "doctor")
	h.mustContain(t, "FAIL  Active account  b (keychain)", "-> supawho remove b && supawho add b <new-token>")
}

func TestDoctorActiveSingleTarget(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a", "b", "sbp_bad")
	h.app.ActiveToken = func() (string, string, error) { return "sbp_bad", "keychain", nil }
	h.run(t, 0, "doctor", "a")
	h.mustContain(t, "PASS  Active account  b (keychain)")
}

func TestDoctorVersion(t *testing.T) {
	cases := []struct {
		name    string
		version string
		latest  func(context.Context) (string, error)
		want    []string
	}{
		{"update", "9.9.9", func(context.Context) (string, error) { return "10.0.0", nil },
			[]string{"WARN  supawho         9.9.9 -> 10.0.0 available", "-> supawho upgrade"}},
		{"dev", "dev", func(context.Context) (string, error) { return "10.0.0", nil },
			[]string{"WARN  supawho         dev (dev build), update check skipped"}},
		{"offline", "9.9.9", func(context.Context) (string, error) { return "", errors.New("offline") },
			[]string{"WARN  supawho         could not check for updates"}},
		{"dev offline", "dev", func(context.Context) (string, error) { return "", errors.New("offline") },
			[]string{"WARN  supawho         dev (dev build), update check skipped"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newDoctorHarness("a", "sbp_ok_a")
			h.app.Version = c.version
			h.app.Latest = c.latest
			h.run(t, 0, "doctor")
			h.mustContain(t, c.want...)
		})
	}
}

func TestDoctorMissingDeps(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	h.app.CLIInfo, h.app.ActiveToken, h.app.Latest = nil, nil, nil
	h.run(t, 0, "doctor")
	if n := strings.Count(h.out.String(), "not available in this build"); n != 3 {
		t.Fatalf("want 3 'not available' rows, got %d:\n%s", n, h.out.String())
	}
}

func TestDoctorNoProfile(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	h.app.Profile = nil
	h.run(t, 1, "doctor")
	h.mustContain(t, "doctor is not available in this build")
}

func TestDoctorJSON(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a", "b", "sbp_bad")
	h.run(t, 1, "doctor", "--json")
	var got struct {
		Accounts []map[string]any `json:"accounts"`
		System   []struct {
			Name string `json:"name"`
		} `json:"system"`
		Summary struct{ Pass, Warn, Fail, Score int } `json:"summary"`
	}
	if err := json.Unmarshal(h.out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, h.out.String())
	}
	var names []string
	for _, s := range got.System {
		names = append(names, s.Name)
	}
	if fmt.Sprint(names) != "[supabase_cli active_account supawho_version]" {
		t.Fatalf("system names = %v", names)
	}
	if len(got.Accounts) != 2 || got.Accounts[1]["status"] != "fail" || got.Summary.Fail != 1 {
		t.Fatalf("report = %+v", got)
	}
	if strings.Contains(h.out.String(), "sbp_") {
		t.Fatalf("JSON leaks a token:\n%s", h.out.String())
	}
}

func TestDoctorNeverPrintsTokens(t *testing.T) {
	for _, args := range [][]string{{"doctor"}, {"doctor", "--json"}} {
		h := newDoctorHarness("c", "sbp_net")
		h.run(t, 0, args...)
		if strings.Contains(h.out.String(), "sbp_net") {
			t.Fatalf("%v leaks a token:\n%s", args, h.out.String())
		}
	}
}

func TestDoctorNoAccounts(t *testing.T) {
	h := newDoctorHarness()
	h.app.ActiveToken = func() (string, string, error) { return "sbp_x", "file", nil }
	h.run(t, 0, "doctor")
	h.mustContain(t, "Accounts\nNo accounts saved.\nSystem\n", "WARN  Active account  active token is not saved in supawho (file)")
}

func TestDoctorPlainHasNoEscapes(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a", "b", "sbp_bad")
	h.run(t, 1, "doctor")
	if strings.Contains(h.out.String(), "\x1b") {
		t.Fatalf("plain output contains escape codes: %q", h.out.String())
	}
}

func TestDoctorCheckAlias(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	h.run(t, 0, "check")
	h.mustContain(t, "Accounts", "PASS  a")
}

func TestDoctorPreservesOrder(t *testing.T) {
	pairs := []string{}
	for i := 0; i < 6; i++ {
		pairs = append(pairs, fmt.Sprintf("n%d", i), fmt.Sprintf("sbp_ok_%d", i))
	}
	h := newDoctorHarness(pairs...)
	base := h.app.Profile
	h.app.Profile = func(token string) (string, []string, error) {
		// earlier accounts finish last
		i := int(token[len(token)-1] - '0')
		time.Sleep(time.Duration(6-i) * 5 * time.Millisecond)
		return base(token)
	}
	h.run(t, 0, "doctor")
	out := h.out.String()
	last := -1
	for i := 0; i < 6; i++ {
		at := strings.Index(out, fmt.Sprintf("PASS  n%d", i))
		if at < last {
			t.Fatalf("rows out of order:\n%s", out)
		}
		last = at
	}
}

// lockedBuffer is safe for the spinner goroutine of the fancy renderer.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func fancyRun(t *testing.T, h *harness, wantCode int) string {
	t.Helper()
	buf := &lockedBuffer{}
	h.app.Out = buf
	h.app.UI = ui.NewWithLevel(buf, ui.LevelTrue)
	if code := h.app.Run([]string{"doctor"}); code != wantCode {
		t.Fatalf("exit = %d, want %d\n%s", code, wantCode, buf.String())
	}
	return buf.String()
}

func TestDoctorFancySmoke(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a", "b", "sbp_bad")
	out := fancyRun(t, h, 1)
	for _, want := range []string{"Accounts", "System", "FAIL", "supawho remove b", "Health", "\x1b[?25l", "\x1b[?25h"} {
		if !strings.Contains(out, want) {
			t.Fatalf("fancy output missing %q:\n%q", want, out)
		}
	}
	if strings.Contains(out, "sbp_") {
		t.Fatalf("fancy output leaks a token: %q", out)
	}
	if strings.Count(out, "\x1b[?25l") != strings.Count(out, "\x1b[?25h") {
		t.Fatalf("cursor hidden/shown counts differ: %q", out)
	}
}

func TestDoctorFancyHealthy(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	out := fancyRun(t, h, 0)
	if !strings.Contains(out, "All systems healthy") {
		t.Fatalf("fancy output missing healthy message: %q", out)
	}
}

func TestDoctorFancyJSONStaysPlain(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	buf := &lockedBuffer{}
	h.app.Out = buf
	h.app.UI = ui.NewWithLevel(buf, ui.LevelTrue)
	h.run(t, 0, "doctor", "--json")
	if strings.Contains(buf.String(), "\x1b") || !json.Valid([]byte(buf.String())) {
		t.Fatalf("--json on a color terminal must be pure JSON: %q", buf.String())
	}
}

func TestDoctorJSONErrors(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"doctor", "ghost", "--json"}, "Account 'ghost' not found. Run 'supawho list' to see saved accounts."},
		{[]string{"doctor", "--json", "--foo"}, "Usage: supawho doctor [name] [--json]"},
	}
	for _, c := range cases {
		h := newDoctorHarness("a", "sbp_ok_a")
		h.run(t, 1, c.args...)
		var got struct{ Error string }
		if err := json.Unmarshal(h.out.Bytes(), &got); err != nil || got.Error != c.want {
			t.Fatalf("%v: want JSON error %q, got %q (%v)", c.args, c.want, h.out.String(), err)
		}
	}
}

func TestDoctorKeyringReadErrorWarns(t *testing.T) {
	h := newDoctorHarness()
	h.app.Store = store.NewMemoryWithBrokenSecrets(errors.New("keychain locked"), "a")
	h.run(t, 0, "doctor")
	h.mustContain(t, "WARN  a  could not read token: keychain locked")
	if strings.Contains(h.out.String(), "supawho remove a") {
		t.Fatalf("a read error must not suggest deleting the account:\n%s", h.out.String())
	}
}

func TestDoctorDevBuildSkipsNetwork(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	h.app.Version = "dev"
	called := false
	h.app.Latest = func(context.Context) (string, error) { called = true; return "10.0.0", nil }
	h.run(t, 0, "doctor")
	if called {
		t.Fatal("a dev build must not query the latest release")
	}
}

func TestRenderPlainUnicodeAlign(t *testing.T) {
	var buf bytes.Buffer
	renderPlain(&buf, report{Accounts: []check{
		{Name: "café", Status: statusPass, Detail: "x"},
		{Name: "abcd", Status: statusPass, Detail: "y"},
	}})
	if !strings.Contains(buf.String(), "PASS  café  x\n") {
		t.Fatalf("non-ASCII names misaligned:\n%s", buf.String())
	}
}
