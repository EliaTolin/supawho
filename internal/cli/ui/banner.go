package ui

import "fmt"

var logo = []string{
	`   ___  _   _ ___  ___  _    _ _  _  ___`,
	`  / __|| | | | _ \/ _ \| |  | | || |/ _ \`,
	`  \__ \| |_| |  _/ (_) | |/\| | __ | (_)`,
	`  |___/ \___/|_|  \__\_\__/\__|_||_|\___/`,
}

// Banner prints the supawho logo and tagline. It is shown only on interactive
// surfaces (the account menu, guided add, help), never in scriptable command
// output.
func (u *UI) Banner() {
	fmt.Fprintln(u.w)
	for _, line := range logo {
		fmt.Fprintln(u.w, u.paint(green, line))
	}
	fmt.Fprintf(u.w, "\n     🔍 %s\n\n", u.paint(dim, "Who are you today?"))
}
