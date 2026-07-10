package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, latest   string
		wantNewer, wantOK bool
	}{
		{"1.4.0", "v1.4.1", true, true},
		{"1.4.0", "v1.5.0", true, true},
		{"1.4.0", "v2.0.0", true, true},
		{"1.4.0", "v1.4.0", false, true},
		{"1.4.1", "v1.4.0", false, true},
		{"2.0.0", "v1.9.9", false, true},
		{"v1.4.0", "1.4.1", true, true},       // both with/without v
		{"dev-abc123", "v1.4.0", true, false}, // dev build: proceed but ok=false
		{"1.4.0", "nonsense", false, false},
	}
	for _, c := range cases {
		newer, ok := IsNewer(c.current, c.latest)
		if newer != c.wantNewer || ok != c.wantOK {
			t.Errorf("IsNewer(%q,%q) = (%v,%v), want (%v,%v)",
				c.current, c.latest, newer, ok, c.wantNewer, c.wantOK)
		}
	}
}

func TestArchiveNameShape(t *testing.T) {
	// Version's leading v is stripped; must contain the project + version.
	name := ArchiveName("v1.4.0")
	if got := name[:len("supawho_1.4.0_")]; got != "supawho_1.4.0_" {
		t.Fatalf("archive name = %q, unexpected prefix", name)
	}
}

func TestManagedBy(t *testing.T) {
	cases := map[string]bool{
		"/opt/homebrew/Cellar/supawho/1.4.0/bin/supawho":   true,
		"/home/linuxbrew/.linuxbrew/bin/supawho":           true,
		"/Users/me/scoop/apps/supawho/current/supawho.exe": true,
		"/usr/bin/supawho":             true,
		"/usr/local/bin/supawho":       false, // manual install
		"/Users/me/go/bin/supawho":     false,
		"/Users/me/.local/bin/supawho": false,
	}
	for path, wantManaged := range cases {
		got := ManagedBy(path) != ""
		if got != wantManaged {
			t.Errorf("ManagedBy(%q) managed=%v, want %v (msg=%q)", path, got, wantManaged, ManagedBy(path))
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello supawho")
	sum := sha256.Sum256(data)
	line := hex.EncodeToString(sum[:]) + "  supawho_1.4.0_Darwin_arm64.tar.gz\n" +
		"deadbeef  other_file.zip\n"

	if err := verifyChecksum([]byte(line), "supawho_1.4.0_Darwin_arm64.tar.gz", data); err != nil {
		t.Fatalf("valid checksum rejected: %v", err)
	}
	if err := verifyChecksum([]byte(line), "supawho_1.4.0_Darwin_arm64.tar.gz", []byte("tampered")); err == nil {
		t.Fatal("tampered data accepted")
	}
	if err := verifyChecksum([]byte(line), "missing.tar.gz", data); err == nil {
		t.Fatal("missing entry accepted")
	}
}
