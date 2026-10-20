package ui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	hideCursor = "\x1b[?25l"
	showCursor = "\x1b[?25h"
	clearLine  = "\r\x1b[2K"
)

var spinnerFrames = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

// LiveList is a block of rows that each show a spinner until their result
// arrives, then are rewritten in place. Rows keep a fixed position regardless
// of completion order. It only writes escape sequences (cursor up, clear
// line, hide/show cursor) and never queries the terminal.
//
// Without color, or for a static list, nothing is written until Stop, which
// prints every row's final line in order, so piped output stays deterministic.
type LiveList struct {
	u      *UI
	static bool

	mu      sync.Mutex
	stopped bool
	rows    []*Row
	drawn   int // lines currently on screen
	frame   int

	started  bool
	done     chan struct{}
	wg       sync.WaitGroup
	stopOnce sync.Once
}

// Row is one entry of a LiveList.
type Row struct {
	l     *LiveList
	label string
	line  string
	done  bool
}

// NewLiveList starts an empty list that renders through this UI.
func (u *UI) NewLiveList() *LiveList {
	return &LiveList{u: u, done: make(chan struct{})}
}

// NewStaticList is a LiveList that never animates, for blocks too tall to be
// redrawn in place (see FitsLive). Rows keep their colors.
func (u *UI) NewStaticList() *LiveList {
	l := u.NewLiveList()
	l.static = true
	return l
}

func (l *LiveList) animated() bool { return l.u.level != LevelNone && !l.static }

// Add appends a row showing label next to a spinner.
func (l *LiveList) Add(label string) *Row {
	l.mu.Lock()
	defer l.mu.Unlock()
	r := &Row{l: l, label: label}
	l.rows = append(l.rows, r)
	if l.animated() && !l.stopped {
		if !l.started {
			l.start()
		}
		l.redraw(true)
	}
	return r
}

// Done replaces the row's spinner with its final line, which may span several
// lines (e.g. a result plus a hint).
func (r *Row) Done(line string) {
	l := r.l
	l.mu.Lock()
	defer l.mu.Unlock()
	r.line, r.done = line, true
	if l.animated() && l.started && !l.stopped {
		l.redraw(true)
	}
}

// Stop ends the animation, draws the final state and restores the cursor. It
// is safe to call more than once.
func (l *LiveList) Stop() {
	l.stopOnce.Do(func() {
		if !l.animated() {
			l.mu.Lock()
			defer l.mu.Unlock()
			l.stopped = true
			for _, r := range l.rows {
				fmt.Fprintln(l.u.w, r.text())
			}
			return
		}
		if !l.started {
			return
		}
		close(l.done)
		l.wg.Wait()
		l.mu.Lock()
		defer l.mu.Unlock()
		l.stopped = true
		l.redraw(false)
		fmt.Fprint(l.u.w, showCursor)
	})
}

// start hides the cursor and launches the 80ms redraw loop. Ctrl-C, SIGTERM
// and SIGHUP restore the cursor before exiting, so the terminal is never left
// without one. Caller holds l.mu.
func (l *LiveList) start() {
	l.started = true
	fmt.Fprint(l.u.w, hideCursor)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	ticker := time.NewTicker(80 * time.Millisecond)

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		defer ticker.Stop()
		defer signal.Stop(sig)
		for {
			select {
			case <-l.done:
				return
			case s := <-sig:
				l.mu.Lock()
				fmt.Fprint(l.u.w, showCursor+"\n")
				l.mu.Unlock()
				code := 130
				if n, ok := s.(syscall.Signal); ok {
					code = 128 + int(n)
				}
				os.Exit(code)
			case <-ticker.C:
				l.mu.Lock()
				l.frame++
				l.redraw(true)
				l.mu.Unlock()
			}
		}
	}()
}

// redraw rewrites the whole block from its first line. While animating
// (fit=true) every line is cut to the terminal width: a wrapped line would
// take two rows and break the cursor-up arithmetic. The final draw keeps
// lines whole, since nothing is redrawn after it. Caller holds l.mu.
func (l *LiveList) redraw(fit bool) {
	var sb strings.Builder
	if l.drawn > 0 {
		fmt.Fprintf(&sb, "\x1b[%dA", l.drawn)
	}
	lines := 0
	for _, r := range l.rows {
		text := r.text()
		if !r.done {
			frame := string(spinnerFrames[l.frame%len(spinnerFrames)])
			text = l.u.Color(Cyan, frame) + " " + r.label
		}
		for _, line := range strings.Split(text, "\n") {
			if fit && l.u.cols > 0 {
				line = truncate(line, l.u.cols-1)
			}
			sb.WriteString(clearLine + line + "\n")
			lines++
		}
	}
	l.drawn = lines
	fmt.Fprint(l.u.w, sb.String())
}

func (r *Row) text() string {
	if r.done {
		return r.line
	}
	return r.label
}
