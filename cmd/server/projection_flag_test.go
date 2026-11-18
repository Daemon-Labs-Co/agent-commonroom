package main

import (
	"context"
	"testing"
	"time"
)

func settingsEnv(pairs ...string) func(string) string {
	m := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return func(k string) string { return m[k] }
}

func TestProjectionCacheFlagDefaultsOffAndRejectsOtherValues(t *testing.T) {
	t.Parallel()
	ps, err := loadPlatformSettings(settingsEnv(), 100)
	if err != nil || ps.projectionCache {
		t.Fatalf("default must be off: %v %v", ps.projectionCache, err)
	}
	if ps, err = loadPlatformSettings(settingsEnv("COMMONROOM_PROJECTION_CACHE", "off"), 100); err != nil || ps.projectionCache {
		t.Fatal(ps.projectionCache, err)
	}
	if ps, err = loadPlatformSettings(settingsEnv("COMMONROOM_PROJECTION_CACHE", "on"), 100); err != nil || !ps.projectionCache {
		t.Fatal(ps.projectionCache, err)
	}
	for _, bad := range []string{"true", "1", "ON", "yes", " on"} {
		if _, err = loadPlatformSettings(settingsEnv("COMMONROOM_PROJECTION_CACHE", bad), 100); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	for _, bad := range []map[string]string{{"COMMONROOM_INTEGRITY_FULL_EVERY": "10s"}, {"COMMONROOM_WORKSPACE_IDLE_MINUTES": "-3"}} {
		pairs := []string{}
		for k, v := range bad {
			pairs = append(pairs, k, v)
		}
		if _, err = loadPlatformSettings(settingsEnv(pairs...), 100); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
	ps, err = loadPlatformSettings(settingsEnv(), 100)
	if err != nil || ps.integrity.incrementalEvery != time.Hour || ps.integrity.fullEvery != 24*time.Hour || ps.workspaceIdle != 30*time.Minute {
		t.Fatalf("defaults: %+v %v", ps, err)
	}
}

// A workspace built with the flag off replays in full and verifies every 30 seconds as before;
// with the flag on it uses the cache. Either way it answers the same.
func TestRuntimeFollowsTheProjectionCacheFlag(t *testing.T) {
	for _, on := range []bool{false, true} {
		f := newPlatformFixture(t)
		f.ps.projectionCache = on
		rt, err := newWorkspaceRuntime(f.ps, f.host.tenancy, newMockDB(t), "acme")
		if err != nil {
			t.Fatal(err)
		}
		if got := rt.s.projection.enabled; got != on {
			t.Fatalf("flag %v: cache enabled %v", on, got)
		}
		if rt.s.projection.active() != (on || projectionCacheForTests) {
			t.Fatal("active() does not follow the flag")
		}
		if _, err = rt.s.load(context.Background()); err != nil {
			t.Fatal(err)
		}
		rt.stop()
	}
}
