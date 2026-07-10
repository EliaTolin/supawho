// Package updater implements `supawho upgrade` and the passive
// "new version available" check, using the GitHub Releases of this repo.
package updater

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/minio/selfupdate"
)

const (
	owner = "EliaTolin"
	repo  = "supawho"
)

// latestAPI is a var (not a const) so tests can point it at a local server.
var latestAPI = "https://api.github.com/repos/" + owner + "/" + repo + "/releases/latest"

// ErrManaged is returned when the binary was installed by a package manager and
// should be upgraded through it rather than by overwriting itself.
type ErrManaged struct{ Manager string }

func (e *ErrManaged) Error() string {
	return "installed via " + e.Manager
}

// asset is one downloadable file attached to a release.
type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Release is the subset of the GitHub release payload we use.
type Release struct {
	TagName string  `json:"tag_name"`
	HTMLURL string  `json:"html_url"`
	Assets  []asset `json:"assets"`
}

// FetchLatest returns the most recent published release.
func FetchLatest(ctx context.Context, client *http.Client) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned status %d", resp.StatusCode)
	}
	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// normalize strips a leading "v" and surrounding space from a version string.
func normalize(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// parseSemver parses "1.4.0" into its numeric parts. ok is false if the string
// is not a clean MAJOR.MINOR.PATCH (e.g. a "dev-..." build).
func parseSemver(v string) (parts [3]int, ok bool) {
	fields := strings.SplitN(normalize(v), ".", 3)
	if len(fields) != 3 {
		return parts, false
	}
	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil {
			return parts, false
		}
		parts[i] = n
	}
	return parts, true
}

// IsNewer reports whether latest is a strictly higher semver than current.
// If current is not a clean semver (dev build), it is treated as older so an
// explicit upgrade still proceeds; ok is false so the passive check can skip.
func IsNewer(current, latest string) (newer, ok bool) {
	lp, lok := parseSemver(latest)
	if !lok {
		return false, false
	}
	cp, cok := parseSemver(current)
	if !cok {
		return true, false
	}
	for i := 0; i < 3; i++ {
		if lp[i] != cp[i] {
			return lp[i] > cp[i], true
		}
	}
	return false, true
}

// ArchiveName returns the release asset file name for the current OS/arch,
// matching the goreleaser name_template (uname-style: Darwin/Linux/Windows,
// x86_64/arm64).
func ArchiveName(version string) string {
	osName := map[string]string{
		"darwin":  "Darwin",
		"linux":   "Linux",
		"windows": "Windows",
	}[runtime.GOOS]
	archName := runtime.GOARCH
	if archName == "amd64" {
		archName = "x86_64"
	}
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("supawho_%s_%s_%s.%s", normalize(version), osName, archName, ext)
}

// findAsset returns the asset whose name matches exactly.
func findAsset(rel *Release, name string) (asset, bool) {
	for _, a := range rel.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return asset{}, false
}

// ManagedBy returns the package manager that owns execPath, or "" if the binary
// looks self-managed (installed via script or a plain binary copy).
func ManagedBy(execPath string) string {
	p := filepath.ToSlash(execPath)
	switch {
	case strings.Contains(p, "/Cellar/"),
		strings.Contains(p, "/opt/homebrew/"),
		strings.Contains(p, "/linuxbrew/"),
		strings.Contains(p, "/.linuxbrew/"):
		return "Homebrew (run: brew upgrade supawho)"
	case strings.Contains(strings.ToLower(p), "/scoop/"):
		return "Scoop (run: scoop update supawho)"
	case p == "/usr/bin/supawho":
		return "your system package manager (apt/dnf/apk)"
	}
	return ""
}

// download fetches a URL fully into memory.
func download(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// verifyChecksum confirms sha256(data) matches the entry for fileName inside a
// goreleaser-style checksums.txt ("<hex>  <name>" per line).
func verifyChecksum(checksums []byte, fileName string, data []byte) error {
	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	scanner := bufio.NewScanner(bytes.NewReader(checksums))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[1] == fileName {
			if !strings.EqualFold(fields[0], actual) {
				return fmt.Errorf("checksum mismatch for %s", fileName)
			}
			return nil
		}
	}
	return fmt.Errorf("no checksum found for %s", fileName)
}

// extractBinary pulls the supawho executable out of a .tar.gz or .zip archive.
func extractBinary(archive []byte, isZip bool) ([]byte, error) {
	binName := "supawho"
	if runtime.GOOS == "windows" {
		binName = "supawho.exe"
	}
	if isZip {
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == binName {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
		return nil, errors.New("binary not found in archive")
	}

	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if filepath.Base(hdr.Name) == binName {
			return io.ReadAll(tr)
		}
	}
	return nil, errors.New("binary not found in archive")
}

// Run is the default entry point wired into the CLI: it upgrades using a
// standard HTTP client with a sensible timeout.
func Run(ctx context.Context, current string, out io.Writer) error {
	return Upgrade(ctx, &http.Client{Timeout: 60 * time.Second}, current, out)
}

// Upgrade checks for a newer release and, if found, replaces the running binary.
// Progress is written to out. It refuses to run for package-manager installs.
func Upgrade(ctx context.Context, client *http.Client, current string, out io.Writer) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = resolved
	}
	return upgradeBinary(ctx, client, current, execPath, out)
}

// upgradeBinary is the testable core: it targets an explicit path instead of
// os.Executable(), so tests can point it at a throwaway file.
func upgradeBinary(ctx context.Context, client *http.Client, current, execPath string, out io.Writer) error {
	if mgr := ManagedBy(execPath); mgr != "" {
		return &ErrManaged{Manager: mgr}
	}

	fmt.Fprintln(out, "Checking for updates...")
	rel, err := FetchLatest(ctx, client)
	if err != nil {
		return err
	}

	newer, _ := IsNewer(current, rel.TagName)
	if !newer {
		fmt.Fprintf(out, "Already up to date (%s).\n", current)
		return nil
	}
	fmt.Fprintf(out, "Updating %s → %s...\n", current, normalize(rel.TagName))

	name := ArchiveName(rel.TagName)
	archiveAsset, ok := findAsset(rel, name)
	if !ok {
		return fmt.Errorf("no build found for your platform (%s)", name)
	}
	sumsAsset, ok := findAsset(rel, "checksums.txt")
	if !ok {
		return errors.New("checksums.txt missing from release")
	}

	archiveData, err := download(ctx, client, archiveAsset.URL)
	if err != nil {
		return err
	}
	sums, err := download(ctx, client, sumsAsset.URL)
	if err != nil {
		return err
	}
	if err := verifyChecksum(sums, name, archiveData); err != nil {
		return err
	}

	binary, err := extractBinary(archiveData, runtime.GOOS == "windows")
	if err != nil {
		return err
	}

	if err := selfupdate.Apply(bytes.NewReader(binary), selfupdate.Options{TargetPath: execPath}); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("update failed and rollback failed: %w", rerr)
		}
		return err
	}

	fmt.Fprintf(out, "Updated to %s.\n", normalize(rel.TagName))
	return nil
}
