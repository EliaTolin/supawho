package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// errInvalidSelection is returned by Select when the choice is not a valid
// option.
var errInvalidSelection = errors.New("invalid selection")

// Prompter collects interactive input using plain, robust terminal I/O: a
// numbered menu and line reads, with password masking on real terminals. It
// deliberately avoids full TUI frameworks, so it never takes over the screen,
// never queries the terminal, and never blocks on startup.
type Prompter interface {
	// Select prints a numbered menu and returns the chosen option.
	Select(title string, options []string) (string, error)
	// Input reads a single line of text.
	Input(title string) (string, error)
	// Password reads a secret, masked when reading from a real terminal.
	Password(title string) (string, error)
	// Confirm asks a yes/no question with the given default.
	Confirm(title string, def bool) (bool, error)
}

// NewPrompter builds a prompter over in/out, coloring menus only on a terminal.
func NewPrompter(in io.Reader, out io.Writer) Prompter {
	return &prompter{in: in, r: bufio.NewReader(in), w: out, color: colorEnabled(out)}
}

type prompter struct {
	in    io.Reader
	r     *bufio.Reader
	w     io.Writer
	color bool
}

func (p *prompter) paint(code, s string) string { return colorize(p.color, code, s) }

func (p *prompter) readLine() string {
	line, _ := p.r.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func (p *prompter) Select(title string, options []string) (string, error) {
	fmt.Fprintf(p.w, "  %s\n\n", p.paint(bold, title))
	for i, o := range options {
		fmt.Fprintf(p.w, "    %s %s\n", p.paint(green, fmt.Sprintf("%d)", i+1)), o)
	}
	fmt.Fprintf(p.w, "\n  %s", p.paint(dim, fmt.Sprintf("Enter number (1-%d): ", len(options))))

	i, err := strconv.Atoi(strings.TrimSpace(p.readLine()))
	if err != nil || i < 1 || i > len(options) {
		return "", errInvalidSelection
	}
	return options[i-1], nil
}

func (p *prompter) Input(title string) (string, error) {
	fmt.Fprintf(p.w, "  %s: ", title)
	return strings.TrimSpace(p.readLine()), nil
}

func (p *prompter) Password(title string) (string, error) {
	fmt.Fprintf(p.w, "  %s: ", title)
	if f, ok := p.in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(p.w)
		return string(b), err
	}
	return p.readLine(), nil
}

func (p *prompter) Confirm(title string, def bool) (bool, error) {
	hint := "[Y/n]"
	if !def {
		hint = "[y/N]"
	}
	fmt.Fprintf(p.w, "  %s %s ", title, p.paint(dim, hint))
	s := strings.TrimSpace(p.readLine())
	if s == "" {
		return def, nil
	}
	return s == "y" || s == "Y", nil
}
