package main

import "runtime/debug"

// version is stamped at build time by goreleaser via
// -ldflags "-X main.version=<tag>". Left empty for plain `go build`/`go install`,
// in which case resolveVersion() derives it from the Go build metadata.
var version = ""

// resolveVersion returns the best available version string:
//  1. the tag injected at release time (goreleaser), if present;
//  2. otherwise the module version from `go install ...@vX.Y.Z`;
//  3. otherwise a "dev-<commit>" string from the local Git checkout;
//  4. otherwise "dev".
func resolveVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	// Set when installed with `go install github.com/EliaTolin/supawho@vX.Y.Z`.
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}

	// Local build: fall back to the VCS revision embedded by the Go toolchain.
	var revision, suffix string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				suffix = "-dirty"
			}
		}
	}
	if revision != "" {
		if len(revision) > 7 {
			revision = revision[:7]
		}
		return "dev-" + revision + suffix
	}
	return "dev"
}
