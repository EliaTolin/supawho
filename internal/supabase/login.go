// Package supabase wraps the external supabase CLI.
package supabase

import (
	"os"
	"os/exec"
	"strings"
)

// accessTokenEnv is the variable the Supabase CLI reads its token from.
const accessTokenEnv = "SUPABASE_ACCESS_TOKEN"

// Login runs `supabase login`, streaming its output to the current
// process's stdio.
func Login(token string) error {
	cmd := loginCommand(token, os.Environ())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// loginCommand builds `supabase login` with the token in the child's
// SUPABASE_ACCESS_TOKEN, which `supabase login` reads when --token is not
// given. Unlike an argument, the environment of a process is not readable by
// other users (ps, /proc/<pid>/cmdline). Any SUPABASE_ACCESS_TOKEN already
// in environ is replaced.
func loginCommand(token string, environ []string) *exec.Cmd {
	env := make([]string, 0, len(environ)+1)
	for _, kv := range environ {
		if !strings.HasPrefix(kv, accessTokenEnv+"=") {
			env = append(env, kv)
		}
	}
	cmd := exec.Command("supabase", "login")
	cmd.Env = append(env, accessTokenEnv+"="+token)
	return cmd
}
