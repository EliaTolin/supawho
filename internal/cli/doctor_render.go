package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// systemLabels are the human labels for the system checks.
var systemLabels = map[string]string{
	checkStore:   "Secret store",
	checkCLI:     "Supabase CLI",
	checkActive:  "Active account",
	checkVersion: "supawho",
}

// systemLabelWidth is fixed to the widest label so the column never shifts.
const systemLabelWidth = len("Active account")

func checkLabel(c check) string {
	if l, ok := systemLabels[c.Name]; ok {
		return l
	}
	return c.Name
}

func formatLatency(c check) string {
	if c.Latency <= 0 {
		return ""
	}
	return fmt.Sprintf("%dms", c.Latency.Milliseconds())
}

// renderPlain prints the deterministic, escape-free report used for pipes,
// NO_COLOR and tests.
func renderPlain(w io.Writer, r report) {
	fmt.Fprintln(w, "Accounts")
	if len(r.Accounts) == 0 {
		fmt.Fprintln(w, "No accounts saved.")
	}
	width := 0
	for _, c := range r.Accounts {
		width = max(width, utf8.RuneCountInString(c.Name))
	}
	for _, c := range r.Accounts {
		plainRow(w, c, width)
	}
	fmt.Fprintln(w, "System")
	for _, c := range r.System {
		plainRow(w, c, systemLabelWidth)
	}
	fmt.Fprintf(w, "\n%s\n", summaryLine(r))
}

func plainRow(w io.Writer, c check, width int) {
	line := fmt.Sprintf("%s  %-*s  %s", strings.ToUpper(c.Status), width, checkLabel(c), c.Detail)
	if l := formatLatency(c); l != "" {
		line += "  " + l
	}
	fmt.Fprintln(w, line)
	if c.Hint != "" {
		fmt.Fprintf(w, "      -> %s\n", c.Hint)
	}
}

func summaryLine(r report) string {
	warnings := "warnings"
	if r.Warn == 1 {
		warnings = "warning"
	}
	return fmt.Sprintf("%d passed, %d %s, %d failed (score %d%%)", r.Pass, r.Warn, warnings, r.Fail, r.Score)
}

type jsonCheck struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Detail    string `json:"detail"`
	LatencyMS int64  `json:"latency_ms,omitempty"`
	Hint      string `json:"hint,omitempty"`
}

type jsonReport struct {
	Accounts []jsonCheck `json:"accounts"`
	System   []jsonCheck `json:"system"`
	Summary  struct {
		Pass  int `json:"pass"`
		Warn  int `json:"warn"`
		Fail  int `json:"fail"`
		Score int `json:"score"`
	} `json:"summary"`
}

// renderJSON prints the report as indented JSON for scripts and CI.
func renderJSON(w io.Writer, r report) error {
	conv := func(cs []check) []jsonCheck {
		out := make([]jsonCheck, 0, len(cs))
		for _, c := range cs {
			out = append(out, jsonCheck{c.Name, c.Status, c.Detail, c.Latency.Milliseconds(), c.Hint})
		}
		return out
	}
	var jr jsonReport
	jr.Accounts, jr.System = conv(r.Accounts), conv(r.System)
	jr.Summary.Pass, jr.Summary.Warn, jr.Summary.Fail, jr.Summary.Score = r.Pass, r.Warn, r.Fail, r.Score
	b, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	return err
}
