// Package fsutil holds small file system helpers.
package fsutil

import (
	"path/filepath"
	"runtime"
	"strings"
)

// SamePath reports whether two repository paths refer to the same location. Windows and
// macOS (by default) have case-insensitive file systems.
func SamePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
