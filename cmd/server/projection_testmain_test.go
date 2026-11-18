package main

import (
	"os"
	"testing"
)

// COMMONROOM_TEST_PROJECTION_CACHE=1 runs the whole package with the incremental cache on, so
// every existing test also proves the cache gives the same results as the full replay.
func TestMain(m *testing.M) {
	projectionCacheForTests = os.Getenv("COMMONROOM_TEST_PROJECTION_CACHE") == "1"
	os.Exit(m.Run())
}
