package supabase

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

// activeEnv isolates ActiveToken from the real machine: mock keyring, empty
// env vars and a temporary home. It returns the temp ~/.supabase directory.
func activeEnv(t *testing.T) string {
	t.Helper()
	keyring.MockInit()
	t.Setenv("SUPABASE_ACCESS_TOKEN", "")
	t.Setenv("SUPABASE_PROFILE", "")
	t.Setenv("SUPABASE_HOME", "")
	home := t.TempDir()
	orig := userHomeDir
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = orig })
	dir := filepath.Join(home, ".supabase")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertActive(t *testing.T, wantTok, wantSrc string) {
	t.Helper()
	tok, src, err := ActiveToken()
	if err != nil || tok != wantTok || src != wantSrc {
		t.Fatalf("ActiveToken() = %q, %q, %v; want %q, %q, nil", tok, src, err, wantTok, wantSrc)
	}
}

func TestActiveTokenEnvWins(t *testing.T) {
	dir := activeEnv(t)
	t.Setenv("SUPABASE_ACCESS_TOKEN", "sbp_env")
	_ = keyring.Set(cliKeyringService, "supabase", "sbp_a")
	writeFile(t, filepath.Join(dir, "access-token"), "sbp_f")
	assertActive(t, "sbp_env", "env")
}

func TestActiveTokenKeychainProfileKey(t *testing.T) {
	activeEnv(t)
	_ = keyring.Set(cliKeyringService, "supabase", "sbp_a")
	_ = keyring.Set(cliKeyringService, "access-token", "sbp_b")
	assertActive(t, "sbp_a", "keychain")
}

func TestActiveTokenKeychainLegacyKey(t *testing.T) {
	activeEnv(t)
	_ = keyring.Set(cliKeyringService, "access-token", "sbp_b")
	assertActive(t, "sbp_b", "keychain")
}

func TestActiveTokenProfileFromEnv(t *testing.T) {
	activeEnv(t)
	t.Setenv("SUPABASE_PROFILE", "work")
	_ = keyring.Set(cliKeyringService, "supabase", "sbp_a")
	_ = keyring.Set(cliKeyringService, "work", "sbp_w")
	assertActive(t, "sbp_w", "keychain")
}

func TestActiveTokenProfileFromFile(t *testing.T) {
	dir := activeEnv(t)
	writeFile(t, filepath.Join(dir, "profile"), "work\n")
	_ = keyring.Set(cliKeyringService, "supabase", "sbp_a")
	_ = keyring.Set(cliKeyringService, "work", "sbp_w")
	assertActive(t, "sbp_w", "keychain")
}

func TestActiveTokenFile(t *testing.T) {
	dir := activeEnv(t)
	writeFile(t, filepath.Join(dir, "access-token"), "sbp_f\n")
	assertActive(t, "sbp_f", "file")
}

func TestActiveTokenSupabaseHome(t *testing.T) {
	activeEnv(t)
	custom := t.TempDir()
	t.Setenv("SUPABASE_HOME", custom)
	writeFile(t, filepath.Join(custom, "access-token"), "sbp_h")
	assertActive(t, "sbp_h", "file")
}

func TestActiveTokenKeyringErrorFallsThrough(t *testing.T) {
	dir := activeEnv(t)
	keyring.MockInitWithError(errors.New("locked"))
	writeFile(t, filepath.Join(dir, "access-token"), "sbp_f")
	assertActive(t, "sbp_f", "file")
}

func TestActiveTokenNone(t *testing.T) {
	activeEnv(t)
	if _, _, err := ActiveToken(); !errors.Is(err, ErrNoActiveToken) {
		t.Fatalf("err = %v, want ErrNoActiveToken", err)
	}
}
