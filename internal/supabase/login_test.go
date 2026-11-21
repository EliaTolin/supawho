package supabase

import (
	"slices"
	"strings"
	"testing"
)

const secret = "test-token-not-a-real-secret"

func TestLoginCommandKeepsTokenOutOfArgs(t *testing.T) {
	cmd := loginCommand(secret, []string{"PATH=/usr/bin"})
	for _, arg := range cmd.Args {
		if strings.Contains(arg, secret) {
			t.Fatalf("token visible in argv: %q", cmd.Args)
		}
	}
	if !slices.Equal(cmd.Args[1:], []string{"login"}) {
		t.Fatalf("args = %q, want [supabase login]", cmd.Args)
	}
}

func TestLoginCommandPassesTokenInEnv(t *testing.T) {
	cmd := loginCommand(secret, []string{"PATH=/usr/bin", "SUPABASE_ACCESS_TOKEN=sbp_old", "HOME=/home/me"})
	var tokens []string
	for _, kv := range cmd.Env {
		if v, ok := strings.CutPrefix(kv, "SUPABASE_ACCESS_TOKEN="); ok {
			tokens = append(tokens, v)
		}
	}
	if !slices.Equal(tokens, []string{secret}) {
		t.Fatalf("SUPABASE_ACCESS_TOKEN values = %q, want only the account's token", tokens)
	}
	for _, want := range []string{"PATH=/usr/bin", "HOME=/home/me"} {
		if !slices.Contains(cmd.Env, want) {
			t.Fatalf("env lost %q: %q", want, cmd.Env)
		}
	}
}
