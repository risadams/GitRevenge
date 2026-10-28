package settings

import (
	"path/filepath"
	"runtime"
	"strings"
)

// samePath reports whether two repository paths refer to the same location. Windows and
// macOS (by default) have case-insensitive file systems.
func samePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
