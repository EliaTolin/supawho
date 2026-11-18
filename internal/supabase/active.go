package supabase

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

// ErrNoActiveToken means the Supabase CLI has no access token in any source.
var ErrNoActiveToken = errors.New("no active Supabase CLI login")

// cliKeyringService is the keyring namespace the Supabase CLI writes to. The
// account is the profile name, with "access-token" as the legacy fallback.
const (
	cliKeyringService = "Supabase CLI"
	legacyKeyringKey  = "access-token"
	defaultProfile    = "supabase"
)

// Sources are vars so tests can swap them.
var userHomeDir = os.UserHomeDir

// ActiveToken returns the token the Supabase CLI would use right now, in the
// CLI's own order: SUPABASE_ACCESS_TOKEN, then the OS keyring (profile key,
// then the legacy key), then <SUPABASE_HOME or ~/.supabase>/access-token.
// source is "env", "keychain" or "file". Unreadable sources are skipped.
func ActiveToken() (token, source string, err error) {
	if tok := strings.TrimSpace(os.Getenv("SUPABASE_ACCESS_TOKEN")); tok != "" {
		return tok, "env", nil
	}

	home := supabaseHome()
	for _, key := range []string{profileName(home), legacyKeyringKey} {
		// keyring.Get decodes the "go-keyring-base64:" prefix itself.
		if tok, err := keyring.Get(cliKeyringService, key); err == nil {
			if tok = strings.TrimSpace(tok); tok != "" {
				return tok, "keychain", nil
			}
		}
	}

	if home != "" {
		if tok := readTrimmed(filepath.Join(home, "access-token")); tok != "" {
			return tok, "file", nil
		}
	}
	return "", "", ErrNoActiveToken
}

// supabaseHome is SUPABASE_HOME or ~/.supabase ("" if neither resolves).
func supabaseHome() string {
	if h := os.Getenv("SUPABASE_HOME"); h != "" {
		return h
	}
	home, err := userHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".supabase")
}

// profileName mirrors the CLI: SUPABASE_PROFILE, then the persisted
// <home>/profile file, then "supabase".
func profileName(home string) string {
	if p := strings.TrimSpace(os.Getenv("SUPABASE_PROFILE")); p != "" {
		return p
	}
	if home != "" {
		if p := readTrimmed(filepath.Join(home, "profile")); p != "" {
			return p
		}
	}
	return defaultProfile
}

func readTrimmed(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
