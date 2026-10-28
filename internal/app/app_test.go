package app

import (
	"testing"

	"github.com/risadams/gitrevenge/internal/version"
)

func TestVersion(t *testing.T) {
	if got := New().Version(); got != version.Version {
		t.Fatalf("Version() = %q, want %q", got, version.Version)
	}
}
