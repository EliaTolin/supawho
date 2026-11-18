//go:build !windows

package ui

import "os"

// enableVT is a no-op outside Windows: Unix terminals always speak VT.
func enableVT(*os.File) bool { return true }
