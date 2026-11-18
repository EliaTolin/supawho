package ui

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// Level is the color depth the terminal supports.
type Level int

const (
	// LevelNone emits no escape codes at all (pipes, NO_COLOR, tests).
	LevelNone Level = iota
	// Level256 uses the xterm 256-color palette.
	Level256
	// LevelTrue uses 24-bit RGB colors.
	LevelTrue
)

// detectLevel picks the color depth for w: none unless w is a terminal and
// NO_COLOR is unset, truecolor when COLORTERM advertises it, else 256.
// On Windows the console must also accept VT sequences, otherwise output
// falls back to plain text.
func detectLevel(w io.Writer) Level {
	f, ok := w.(*os.File)
	tty := ok && term.IsTerminal(int(f.Fd())) && enableVT(f)
	return detectLevelFor(tty, os.Getenv("NO_COLOR"), os.Getenv("COLORTERM"), os.Getenv("TERM"))
}

func detectLevelFor(tty bool, noColor, colorterm, termEnv string) Level {
	switch {
	case !tty || noColor != "" || termEnv == "dumb":
		return LevelNone
	case colorterm == "truecolor" || colorterm == "24bit":
		return LevelTrue
	default:
		return Level256
	}
}

// terminalSize reads the size of a terminal writer via ioctl (never an escape
// query); 0, 0 when w is not a terminal.
func terminalSize(w io.Writer) (cols, rows int) {
	f, ok := w.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return 0, 0
	}
	cols, rows, err := term.GetSize(int(f.Fd()))
	if err != nil {
		return 0, 0
	}
	return cols, rows
}

// FitsLive reports whether a block of lines can be animated in place: taller
// blocks would scroll, and cursor-up cannot reach above the top of the screen.
func (u *UI) FitsLive(lines int) bool {
	return u.rows == 0 || lines < u.rows-1
}

// RGB is a 24-bit color, downsampled to the 256 palette when needed.
type RGB struct{ R, G, B uint8 }

// Brand colors shared by the fancy renderers.
var (
	SupaGreen = RGB{62, 207, 142}
	Cyan      = RGB{34, 211, 238}
	Magenta   = RGB{217, 70, 239}
	Red       = RGB{239, 68, 68}
	Yellow    = RGB{250, 204, 21}
	White     = RGB{255, 255, 255}
	Black     = RGB{0, 0, 0}
)

// xterm256 maps c onto the 6×6×6 color cube of the xterm palette.
func xterm256(c RGB) int {
	q := func(v uint8) int { return int(math.Round(float64(v) * 5 / 255)) }
	return 16 + 36*q(c.R) + 6*q(c.G) + q(c.B)
}

// fg and bg return the escape sequence for a foreground/background color at
// this UI's level ("" when color is off).
func (u *UI) fg(c RGB) string {
	switch u.level {
	case LevelTrue:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.R, c.G, c.B)
	case Level256:
		return fmt.Sprintf("\x1b[38;5;%dm", xterm256(c))
	}
	return ""
}

func (u *UI) bg(c RGB) string {
	switch u.level {
	case LevelTrue:
		return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c.R, c.G, c.B)
	case Level256:
		return fmt.Sprintf("\x1b[48;5;%dm", xterm256(c))
	}
	return ""
}

// Color paints s in a single foreground color.
func (u *UI) Color(c RGB, s string) string {
	if u.level == LevelNone {
		return s
	}
	return u.fg(c) + s + reset
}

// Bold and Dim style s when color is on.
func (u *UI) Bold(s string) string { return u.paint(bold, s) }
func (u *UI) Dim(s string) string  { return u.paint(dim, s) }

// Gradient paints each rune of s on a linear ramp from one color to another.
func (u *UI) Gradient(s string, from, to RGB) string {
	if u.level == LevelNone {
		return s
	}
	n := utf8.RuneCountInString(s)
	lerp := func(a, b uint8, t float64) uint8 { return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t)) }
	var sb strings.Builder
	i := 0
	for _, r := range s {
		t := 0.0
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		sb.WriteString(u.fg(RGB{lerp(from.R, to.R, t), lerp(from.G, to.G, t), lerp(from.B, to.B, t)}))
		sb.WriteRune(r)
		i++
	}
	sb.WriteString(reset)
	return sb.String()
}

// Badge renders a status ("pass", "warn", "fail") as a solid label, or as the
// bare upper-case word without color.
func (u *UI) Badge(status string) string {
	label := strings.ToUpper(status)
	if u.level == LevelNone {
		return label
	}
	back, fore := SupaGreen, White
	switch status {
	case "warn":
		back, fore = Yellow, Black
	case "fail":
		back, fore = Red, White
	}
	return u.bg(back) + u.fg(fore) + bold + " " + label + " " + reset
}

// Section renders a section title over a gradient rule.
func (u *UI) Section(icon, title string) string {
	return "  " + icon + " " + u.Bold(title) + "\n  " + u.Gradient(strings.Repeat("─", 44), SupaGreen, Cyan)
}

// HintCmd renders a remediation hint: a magenta marker and the command.
func (u *UI) HintCmd(text, cmd string) string {
	return u.Color(Magenta, text) + " " + u.Bold(u.Color(Cyan, cmd))
}

// Box frames lines in a rounded border, padding every row to the widest one.
func (u *UI) Box(lines []string) string {
	w := 0
	for _, l := range lines {
		w = max(w, VisibleWidth(l))
	}
	edge := strings.Repeat("─", w+2)
	var sb strings.Builder
	sb.WriteString(u.Dim("╭"+edge+"╮") + "\n")
	for _, l := range lines {
		pad := strings.Repeat(" ", w-VisibleWidth(l))
		sb.WriteString(u.Dim("│") + " " + l + pad + " " + u.Dim("│") + "\n")
	}
	sb.WriteString(u.Dim("╰"+edge+"╯") + "\n")
	return sb.String()
}

// ProgressBar renders pct (0–100) as a bar of width cells, colored red below
// 50%, yellow up to 85% and green above.
func (u *UI) ProgressBar(pct, width int) string {
	pct = min(max(pct, 0), 100)
	full := int(math.Round(float64(pct) * float64(width) / 100))
	c := map[string]RGB{"red": Red, "yellow": Yellow, "green": SupaGreen}[progressColor(pct)]
	return u.Color(c, strings.Repeat("█", full)) + u.Dim(strings.Repeat("░", width-full))
}

func progressColor(pct int) string {
	switch {
	case pct < 50:
		return "red"
	case pct <= 85:
		return "yellow"
	default:
		return "green"
	}
}

// VisibleWidth is the number of terminal cells s occupies: ANSI escape
// sequences count zero, emoji count two, everything else one.
func VisibleWidth(s string) int {
	w := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
				j++
			}
			i = j + 1
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		switch {
		case r == 0xFE0F || r == 0x200D:
			// variation selector / zero-width joiner
		case r >= 0x1F300 || r == 0x2728: // emoji, ✨
			w += 2
		default:
			w++
		}
	}
	return w
}

// truncate shortens s to at most max visible cells, ending in "…". Escape
// sequences are kept, and a reset is appended if any color was open.
func truncate(s string, max int) string {
	if VisibleWidth(s) <= max {
		return s
	}
	var sb strings.Builder
	w, colored := 0, false
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
				j++
			}
			if j < len(s) {
				j++
			}
			sb.WriteString(s[i:j])
			colored = true
			i = j
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		rw := VisibleWidth(string(r))
		if w+rw > max-1 {
			break
		}
		sb.WriteRune(r)
		w += rw
		i += size
	}
	sb.WriteString("…")
	if colored {
		sb.WriteString(reset)
	}
	return sb.String()
}
