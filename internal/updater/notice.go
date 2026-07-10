package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const checkInterval = 24 * time.Hour

func cacheFile() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "supawho", "last-update-check"), nil
}

// MaybeNotify prints a one-line "update available" notice to w when a newer
// release exists, at most once per 24h. It is best-effort and silent: any error
// (offline, rate-limited, dev build) is ignored. Disable via SUPAWHO_NO_UPDATE_CHECK.
func MaybeNotify(current string, w io.Writer) {
	if os.Getenv("SUPAWHO_NO_UPDATE_CHECK") != "" {
		return
	}
	// Skip dev builds — their version can't be compared to a release tag.
	if _, ok := parseSemver(current); !ok {
		return
	}

	path, err := cacheFile()
	if err != nil {
		return
	}
	if info, statErr := os.Stat(path); statErr == nil && time.Since(info.ModTime()) < checkInterval {
		return
	}
	// Record the attempt up front so a failing check doesn't retry every run.
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0o644)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rel, err := FetchLatest(ctx, &http.Client{Timeout: 2 * time.Second})
	if err != nil {
		return
	}
	if newer, ok := IsNewer(current, rel.TagName); ok && newer {
		fmt.Fprintf(w, "\nA new version of supawho is available: %s → %s\nRun 'supawho upgrade' to update.\n",
			current, normalize(rel.TagName))
	}
}
