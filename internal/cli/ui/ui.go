// Package ui centralizes supawho's terminal presentation: a single styling
// vocabulary and a Prompter abstraction for interactive input.
//
// Styling uses plain ANSI escape codes, emitted only when the output writer is
// a real terminal and NO_COLOR is unset. Crucially it never *queries* the
// terminal (unlike full TUI renderers), so one-shot commands can't leave stray
// response bytes on the shell prompt. Pipes and tests get plain text.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"golang.org/x/term"
)

const (
	green  = "\033[1;32m"
	red    = "\033[1;31m"
	yellow = "\033[0;33m"
	dim    = "\033[2m"
	bold   = "\033[1m"
	reset  = "\033[0m"
)

// UI renders styled status lines to a fixed writer.
type UI struct {
	w     io.Writer
	color bool
}

// New builds a UI, enabling color only for real terminals without NO_COLOR.
func New(w io.Writer) *UI {
	return &UI{w: w, color: colorEnabled(w)}
}

func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// colorize wraps s in an ANSI code when on, otherwise returns it unchanged.
func colorize(on bool, code, s string) string {
	if !on {
		return s
	}
	return code + s + reset
}

// paint colors s when this UI has color enabled.
func (u *UI) paint(code, s string) string { return colorize(u.color, code, s) }

// OK prints a success line: a green check plus the message.
func (u *UI) OK(format string, a ...any) {
	fmt.Fprintf(u.w, "  %s %s\n", u.paint(green, "✓"), fmt.Sprintf(format, a...))
}

// Err prints an error line: a red cross plus the message.
func (u *UI) Err(format string, a ...any) {
	fmt.Fprintf(u.w, "  %s %s\n", u.paint(red, "✗"), fmt.Sprintf(format, a...))
}

// Warn prints a warning line: a yellow bang plus the message.
func (u *UI) Warn(format string, a ...any) {
	fmt.Fprintf(u.w, "  %s %s\n", u.paint(yellow, "!"), fmt.Sprintf(format, a...))
}

// Info prints a plain, indented line (no symbol).
func (u *UI) Info(format string, a ...any) {
	fmt.Fprintf(u.w, "  %s\n", fmt.Sprintf(format, a...))
}

// Hint prints a dimmed, indented line for secondary guidance.
func (u *UI) Hint(format string, a ...any) {
	fmt.Fprintf(u.w, "  %s\n", u.paint(dim, fmt.Sprintf(format, a...)))
}

// Title prints a bold section header, with an optional dimmed count suffix.
func (u *UI) Title(s string, count int) {
	if count >= 0 {
		fmt.Fprintf(u.w, "  %s %s\n", u.paint(bold, s), u.paint(dim, fmt.Sprintf("(%d)", count)))
		return
	}
	fmt.Fprintf(u.w, "  %s\n", u.paint(bold, s))
}

// Bullet prints a list item marked with a green bullet.
func (u *UI) Bullet(format string, a ...any) {
	fmt.Fprintf(u.w, "    %s %s\n", u.paint(green, "●"), fmt.Sprintf(format, a...))
}

// Field prints a "label   value" row with the label dimmed and aligned.
func (u *UI) Field(label, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(u.w, "  %s  %s\n", u.paint(dim, fmt.Sprintf("%-13s", label)), value)
}

// Line prints a blank line.
func (u *UI) Line() { fmt.Fprintln(u.w) }

// Table prints columns aligned with tabwriter, indented to the two-space
// gutter. No box-drawing borders, so it never breaks on narrow terminals.
func (u *UI) Table(headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(u.w, 0, 2, 2, ' ', 0)
	fmt.Fprintf(tw, "  %s\n", strings.Join(headers, "\t"))
	for _, r := range rows {
		fmt.Fprintf(tw, "  %s\n", strings.Join(r, "\t"))
	}
	tw.Flush()
}
