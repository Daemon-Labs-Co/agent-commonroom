//go:build !linux || !amd64

package projectquota

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUnsupportedPlatformDoesNotSubstituteHostEvidence(t *testing.T) {
	t.Parallel()
	root, err := os.Open(filepath.Clean(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	e, _ := quotaFixture()
	if obs, err := Inspect(context.Background(), root, e); err != ErrUnavailable || obs != (Observation{}) {
		t.Fatal("unsupported platform claimed evidence", obs, err)
	}
}
