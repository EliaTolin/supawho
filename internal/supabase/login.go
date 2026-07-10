// Package supabase wraps the external supabase CLI.
package supabase

import (
	"os"
	"os/exec"
)

// Login runs `supabase login --token <token>`, streaming its output to the
// current process's stdio, exactly like the original bash script.
func Login(token string) error {
	cmd := exec.Command("supabase", "login", "--token", token)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
