package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestDetectLevel(t *testing.T) {
	if got := detectLevel(&bytes.Buffer{}); got != LevelNone {
		t.Fatalf("detectLevel(buffer) = %v, want LevelNone", got)
	}
	cases := []struct {
		tty       bool
		noColor   string
		colorterm string
		term      string
		want      Level
	}{
		{true, "", "truecolor", "xterm-256color", LevelTrue},
		{true, "", "24bit", "xterm-256color", LevelTrue},
		{true, "", "", "xterm-256color", Level256},
		{true, "1", "truecolor", "xterm-256color", LevelNone},
		{false, "", "truecolor", "xterm-256color", LevelNone},
		{true, "", "truecolor", "dumb", LevelNone},
	}
	for _, c := range cases {
		if got := detectLevelFor(c.tty, c.noColor, c.colorterm, c.term); got != c.want {
			t.Errorf("detectLevelFor(%v, %q, %q, %q) = %v, want %v", c.tty, c.noColor, c.colorterm, c.term, got, c.want)
		}
	}
}

func TestVisibleWidth(t *testing.T) {
	cases := map[string]int{
		"abc":                        3,
		"\x1b[1;32mabc\x1b[0m":       3,
		"\x1b[38;2;1;2;3mab\x1b[0m":  2,
		"⠋ x":                        3,
		"🩺 doctor":                   9,
		"╭─╮":                        3,
		"✨ ok":                       5,
		"\x1b[48;5;46m PASS \x1b[0m": 6,
	}
	for s, want := range cases {
		if got := VisibleWidth(s); got != want {
			t.Errorf("VisibleWidth(%q) = %d, want %d", s, got, want)
		}
	}
}

func TestGradientEndpoints(t *testing.T) {
	u := NewWithLevel(&bytes.Buffer{}, LevelTrue)
	got := u.Gradient("ab", RGB{0, 0, 0}, RGB{255, 255, 255})
	if !strings.HasPrefix(got, "\x1b[38;2;0;0;0ma") ||
		!strings.Contains(got, "\x1b[38;2;255;255;255mb") ||
		!strings.HasSuffix(got, reset) {
		t.Fatalf("Gradient = %q", got)
	}
	if got := NewWithLevel(&bytes.Buffer{}, LevelNone).Gradient("ab", RGB{}, RGB{}); got != "ab" {
		t.Fatalf("plain Gradient = %q, want ab", got)
	}
}

func TestGradient256(t *testing.T) {
	u := NewWithLevel(&bytes.Buffer{}, Level256)
	if got := u.Gradient("a", RGB{0, 255, 0}, RGB{0, 255, 0}); got != "\x1b[38;5;46ma\x1b[0m" {
		t.Fatalf("Gradient 256 = %q", got)
	}
}

func TestBadge(t *testing.T) {
	plain := NewWithLevel(&bytes.Buffer{}, LevelNone)
	for status, want := range map[string]string{"pass": "PASS", "warn": "WARN", "fail": "FAIL"} {
		if got := plain.Badge(status); got != want {
			t.Errorf("plain Badge(%q) = %q, want %q", status, got, want)
		}
	}
	color := NewWithLevel(&bytes.Buffer{}, LevelTrue)
	if got := VisibleWidth(color.Badge("pass")); got != 6 {
		t.Fatalf("colored badge width = %d, want 6", got)
	}
}

func TestBoxPlain(t *testing.T) {
	u := NewWithLevel(&bytes.Buffer{}, LevelNone)
	want := "╭────╮\n│ ab │\n│ c  │\n╰────╯\n"
	if got := u.Box([]string{"ab", "c"}); got != want {
		t.Fatalf("Box =\n%s\nwant\n%s", got, want)
	}
}

func TestProgressBarPlain(t *testing.T) {
	u := NewWithLevel(&bytes.Buffer{}, LevelNone)
	if got := u.ProgressBar(50, 4); got != "██░░" {
		t.Fatalf("ProgressBar(50, 4) = %q", got)
	}
}

func TestProgressColor(t *testing.T) {
	cases := map[int]string{0: "red", 49: "red", 50: "yellow", 85: "yellow", 86: "green", 100: "green"}
	for pct, want := range cases {
		if got := progressColor(pct); got != want {
			t.Errorf("progressColor(%d) = %q, want %q", pct, got, want)
		}
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"abc", 5, "abc"},
		{"abcdef", 4, "abc…"},
		{"\x1b[1mabcdef\x1b[0m", 4, "\x1b[1mabc…\x1b[0m"},
		{"🩺abc", 3, "🩺…"},
	}
	for _, c := range cases {
		if got := truncate(c.in, c.max); got != c.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
		if VisibleWidth(truncate(c.in, c.max)) > c.max {
			t.Errorf("truncate(%q, %d) is wider than %d", c.in, c.max, c.max)
		}
	}
}
