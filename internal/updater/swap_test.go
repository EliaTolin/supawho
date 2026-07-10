package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// makeTarGz builds a .tar.gz containing a single file `binName` with content.
func makeTarGz(t *testing.T, binName string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: binName, Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestUpgradeEndToEnd exercises the real swap path: fetch latest -> download
// archive -> verify checksum -> extract -> replace the target binary on disk.
func TestUpgradeEndToEnd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("archive is .tar.gz; this test targets unix packaging")
	}

	newBinary := []byte("#!/bin/sh\necho supawho v9.9.9\n")
	archive := makeTarGz(t, "supawho", newBinary)

	name := ArchiveName("9.9.9") // e.g. supawho_9.9.9_Darwin_arm64.tar.gz
	sum := sha256.Sum256(archive)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), name)

	mux := http.NewServeMux()
	mux.HandleFunc("/archive", func(w http.ResponseWriter, r *http.Request) { w.Write(archive) })
	mux.HandleFunc("/sums", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(checksums)) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"tag_name": "v9.9.9",
			"html_url": "%s",
			"assets": [
				{"name": %q, "browser_download_url": "%s/archive"},
				{"name": "checksums.txt", "browser_download_url": "%s/sums"}
			]
		}`, srv.URL, name, srv.URL, srv.URL)
	})

	// Point the package at our fake GitHub for the duration of the test.
	orig := latestAPI
	latestAPI = srv.URL + "/latest"
	defer func() { latestAPI = orig }()

	// A throwaway "installed binary" to be replaced.
	target := filepath.Join(t.TempDir(), "supawho")
	if err := os.WriteFile(target, []byte("OLD BINARY v1.0.0"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := upgradeBinary(context.Background(), srv.Client(), "1.0.0", target, &out)
	if err != nil {
		t.Fatalf("upgradeBinary error: %v\noutput:\n%s", err, out.String())
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, newBinary) {
		t.Fatalf("binary was not replaced.\n got: %q\nwant: %q", got, newBinary)
	}
}

// TestUpgradeRejectsTamperedArchive ensures a bad checksum aborts the swap and
// leaves the original binary untouched.
func TestUpgradeRejectsTamperedArchive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("archive is .tar.gz; this test targets unix packaging")
	}

	archive := makeTarGz(t, "supawho", []byte("malicious payload"))
	name := ArchiveName("9.9.9")
	// Deliberately wrong checksum.
	checksums := fmt.Sprintf("%s  %s\n", "0000000000000000000000000000000000000000000000000000000000000000", name)

	mux := http.NewServeMux()
	mux.HandleFunc("/archive", func(w http.ResponseWriter, r *http.Request) { w.Write(archive) })
	mux.HandleFunc("/sums", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(checksums)) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"tag_name":"v9.9.9","assets":[
			{"name":%q,"browser_download_url":"%s/archive"},
			{"name":"checksums.txt","browser_download_url":"%s/sums"}]}`, name, srv.URL, srv.URL)
	})

	orig := latestAPI
	latestAPI = srv.URL + "/latest"
	defer func() { latestAPI = orig }()

	target := filepath.Join(t.TempDir(), "supawho")
	original := []byte("OLD BINARY v1.0.0")
	if err := os.WriteFile(target, original, 0o755); err != nil {
		t.Fatal(err)
	}

	err := upgradeBinary(context.Background(), srv.Client(), "1.0.0", target, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected checksum error, got nil")
	}
	got, _ := os.ReadFile(target)
	if !bytes.Equal(got, original) {
		t.Fatalf("tampered update modified the binary! got %q", got)
	}
}

// TestUpgradeAlreadyCurrent ensures no download happens when up to date.
func TestUpgradeAlreadyCurrent(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"tag_name":"v1.0.0","assets":[]}`)
	})
	orig := latestAPI
	latestAPI = srv.URL + "/latest"
	defer func() { latestAPI = orig }()

	target := filepath.Join(t.TempDir(), "supawho")
	os.WriteFile(target, []byte("x"), 0o755)

	var out bytes.Buffer
	if err := upgradeBinary(context.Background(), srv.Client(), "1.0.0", target, &out); err != nil {
		t.Fatalf("error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Already up to date")) {
		t.Fatalf("output = %q", out.String())
	}
}

// TestMaybeNotifyGuards checks the no-network guard paths.
func TestMaybeNotifyGuards(t *testing.T) {
	// Disabled via env: must not touch the network or print.
	t.Setenv("SUPAWHO_NO_UPDATE_CHECK", "1")
	var out bytes.Buffer
	MaybeNotify("1.0.0", &out)
	if out.Len() != 0 {
		t.Fatalf("expected no output when disabled, got %q", out.String())
	}

	// Dev build: unparsable version is skipped before any network call.
	t.Setenv("SUPAWHO_NO_UPDATE_CHECK", "")
	out.Reset()
	MaybeNotify("dev-abc123", &out)
	if out.Len() != 0 {
		t.Fatalf("expected no output for dev build, got %q", out.String())
	}
}
