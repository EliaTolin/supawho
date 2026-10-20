package ui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
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
// Without color the list is silent until Stop, which prints every row's final
// line in order, so piped output stays deterministic.
type LiveList struct {
	u *UI

	mu    sync.Mutex
	rows  []*Row
	drawn int // lines currently on screen
	frame int

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

// Add appends a row showing label next to a spinner.
func (l *LiveList) Add(label string) *Row {
	l.mu.Lock()
	defer l.mu.Unlock()
	r := &Row{l: l, label: label}
	l.rows = append(l.rows, r)
	if l.u.level != LevelNone {
		if !l.started {
			l.start()
		}
		l.redraw()
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
	if l.u.level != LevelNone && l.started {
		l.redraw()
	}
}

// Stop ends the animation, draws the final state and restores the cursor. It
// is safe to call more than once.
func (l *LiveList) Stop() {
	l.stopOnce.Do(func() {
		if l.u.level == LevelNone {
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
		l.redraw()
		fmt.Fprint(l.u.w, showCursor)
	})
}

// start hides the cursor and launches the 80ms redraw loop. Ctrl-C restores
// the cursor before exiting, so the terminal is never left without one.
// Caller holds l.mu.
func (l *LiveList) start() {
	l.started = true
	fmt.Fprint(l.u.w, hideCursor)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
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
			case <-sig:
				l.mu.Lock()
				fmt.Fprint(l.u.w, showCursor+"\n")
				l.mu.Unlock()
				os.Exit(130)
			case <-ticker.C:
				l.mu.Lock()
				l.frame++
				l.redraw()
				l.mu.Unlock()
			}
		}
	}()
}

// redraw rewrites the whole block from its first line. Caller holds l.mu.
func (l *LiveList) redraw() {
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
