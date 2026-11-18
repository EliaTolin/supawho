package ui

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// syncBuffer is a bytes.Buffer safe for the LiveList redraw goroutine.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestLiveListPlainOrdered(t *testing.T) {
	var buf bytes.Buffer
	l := NewWithLevel(&buf, LevelNone).NewLiveList()
	a, b := l.Add("a"), l.Add("b")
	b.Done("B")
	a.Done("A")
	if buf.Len() != 0 {
		t.Fatalf("plain LiveList wrote before Stop: %q", buf.String())
	}
	l.Stop()
	if got := buf.String(); got != "A\nB\n" {
		t.Fatalf("output = %q, want %q", got, "A\nB\n")
	}
}

func TestLiveListPlainUnfinishedRow(t *testing.T) {
	var buf bytes.Buffer
	l := NewWithLevel(&buf, LevelNone).NewLiveList()
	l.Add("a")
	l.Stop()
	if got := buf.String(); got != "a\n" {
		t.Fatalf("output = %q, want %q", got, "a\n")
	}
}

func TestLiveListStopRestoresCursor(t *testing.T) {
	buf := &syncBuffer{}
	l := NewWithLevel(buf, LevelTrue).NewLiveList()
	l.Add("a").Done("A")
	l.Stop()
	l.Stop()
	out := buf.String()
	if !strings.Contains(out, "\x1b[?25l") || !strings.HasSuffix(out, "\x1b[?25h") || !strings.Contains(out, "A") {
		t.Fatalf("output = %q", out)
	}
	if n := strings.Count(out, "\x1b[?25h"); n != 1 {
		t.Fatalf("cursor restored %d times, want 1", n)
	}
}

func TestLiveListNoTerminalQueries(t *testing.T) {
	buf := &syncBuffer{}
	l := NewWithLevel(buf, LevelTrue).NewLiveList()
	r := l.Add("a")
	l.Add("b").Done("B")
	r.Done("A")
	l.Stop()
	out := buf.String()
	for _, q := range []string{"\x1b[6n", "\x1b]", "\x1b[c", "\x1b[>c"} {
		if strings.Contains(out, q) {
			t.Fatalf("output contains terminal query %q: %q", q, out)
		}
	}
}

func TestLiveListMultilineRow(t *testing.T) {
	buf := &syncBuffer{}
	l := NewWithLevel(buf, LevelTrue).NewLiveList()
	a := l.Add("a")
	l.Add("b")
	a.Done("A\n  hint")
	l.Stop()
	ups := regexp.MustCompile(`\x1b\[(\d+)A`).FindAllStringSubmatch(buf.String(), -1)
	if len(ups) == 0 || ups[len(ups)-1][1] != "3" {
		t.Fatalf("final redraw should move up 3 lines; moves = %v", ups)
	}
}

func TestBannerGradientPlain(t *testing.T) {
	var a, b bytes.Buffer
	NewWithLevel(&a, LevelNone).BannerGradient()
	NewWithLevel(&b, LevelNone).Banner()
	if a.String() != b.String() {
		t.Fatalf("plain BannerGradient = %q, want %q", a.String(), b.String())
	}
}

func TestLiveListTruncatesWhileAnimating(t *testing.T) {
	buf := &syncBuffer{}
	u := NewWithLevel(buf, LevelTrue)
	u.cols, u.rows = 10, 40
	l := u.NewLiveList()
	l.Add("a").Done(strings.Repeat("x", 30))
	l.Add("b") // still spinning: redraws must not wrap
	animated := buf.String()
	if strings.Contains(animated, strings.Repeat("x", 10)) {
		t.Fatalf("animated redraw wrote a line wider than the terminal: %q", animated)
	}
	l.Stop()
	if !strings.Contains(buf.String(), strings.Repeat("x", 30)) {
		t.Fatalf("final draw must keep the full line: %q", buf.String())
	}
}

func TestStaticListPrintsOnStop(t *testing.T) {
	buf := &syncBuffer{}
	u := NewWithLevel(buf, LevelTrue)
	l := u.NewStaticList()
	l.Add("a").Done("A")
	if buf.String() != "" {
		t.Fatalf("static list wrote before Stop: %q", buf.String())
	}
	l.Stop()
	if got := buf.String(); got != "A\n" {
		t.Fatalf("static output = %q, want %q", got, "A\n")
	}
}

func TestFitsLive(t *testing.T) {
	u := NewWithLevel(&bytes.Buffer{}, LevelTrue)
	if !u.FitsLive(500) {
		t.Fatal("unknown terminal height must allow animation")
	}
	u.rows = 20
	if !u.FitsLive(18) || u.FitsLive(19) {
		t.Fatal("FitsLive must require lines < rows-1")
	}
}

func TestLiveListDoneAfterStopIsSilent(t *testing.T) {
	buf := &syncBuffer{}
	l := NewWithLevel(buf, LevelTrue).NewLiveList()
	r := l.Add("a")
	l.Stop()
	before := buf.String()
	r.Done("late")
	if buf.String() != before {
		t.Fatalf("Done after Stop wrote output: %q", strings.TrimPrefix(buf.String(), before))
	}
}
