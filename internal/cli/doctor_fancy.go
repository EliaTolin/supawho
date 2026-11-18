package cli

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/EliaTolin/supawho/internal/cli/ui"
)

// renderFancy runs the checks while animating them: gradient banner, one
// spinner row per check rewritten in place with its badge, and a health box.
func (a *App) renderFancy(names []string) report {
	u := a.UI
	u.BannerGradient()
	noun := "accounts"
	if len(names) == 1 {
		noun = "account"
	}
	fmt.Fprintf(a.Out, "  🩺 %s\n\n", u.Dim(fmt.Sprintf("doctor · checking %d %s", len(names), noun)))

	fmt.Fprintln(a.Out, u.Section("🔑", "Accounts"))
	if len(names) == 0 {
		fmt.Fprintf(a.Out, "  %s\n", u.Dim("No accounts saved."))
	}
	width := 0
	for _, n := range names {
		width = max(width, utf8.RuneCountInString(n))
	}
	// Each account takes up to two lines (result + hint). A block taller than
	// the screen can't be redrawn in place, so it is printed once at the end.
	accounts := u.NewLiveList()
	if !u.FitsLive(2 * len(names)) {
		accounts = u.NewStaticList()
	}
	defer accounts.Stop()
	rows := make([]*ui.Row, len(names))
	for i, n := range names {
		rows[i] = accounts.Add("  " + u.Bold(n))
	}

	labels := []string{checkCLI, checkActive, checkVersion}
	var sysRows []*ui.Row
	system := u.NewLiveList()
	defer system.Stop()

	r := a.diagnose(names, progress{
		account: func(i int, c check) { rows[i].Done(fancyRow(u, c, width)) },
		beforeSystem: func() {
			accounts.Stop()
			fmt.Fprintln(a.Out)
			fmt.Fprintln(a.Out, u.Section("🧰", "System"))
			for _, name := range labels {
				sysRows = append(sysRows, system.Add("  "+u.Bold(systemLabels[name])))
			}
		},
		system: func(i int, c check) { sysRows[i].Done(fancyRow(u, c, systemLabelWidth)) },
	})
	system.Stop()

	fmt.Fprintln(a.Out)
	lines := []string{
		fmt.Sprintf("Health  %s  %s", u.ProgressBar(r.Score, 24), u.Bold(fmt.Sprintf("%d%%", r.Score))),
		u.Dim(fmt.Sprintf("%d passed · %d warnings · %d failed", r.Pass, r.Warn, r.Fail)),
	}
	if r.Score == 100 {
		lines = append(lines, "✨ All systems healthy — you're good to go!")
	}
	for _, line := range strings.Split(strings.TrimSuffix(u.Box(lines), "\n"), "\n") {
		fmt.Fprintln(a.Out, "  "+line)
	}
	return r
}

// fancyRow renders one finished check, plus its hint on a second line.
func fancyRow(u *ui.UI, c check, width int) string {
	detail := c.Detail
	switch c.Status {
	case statusPass:
		detail = u.Color(ui.Cyan, detail)
	case statusWarn:
		detail = u.Color(ui.Yellow, detail)
	case statusFail:
		detail = u.Color(ui.Red, detail)
	}
	line := fmt.Sprintf("  %s %s  %s", u.Badge(c.Status), u.Bold(fmt.Sprintf("%-*s", width, checkLabel(c))), detail)
	if l := formatLatency(c); l != "" {
		line += "  " + u.Dim(l)
	}
	if c.Hint != "" {
		line += "\n         " + u.HintCmd("↳", c.Hint)
	}
	return line
}
