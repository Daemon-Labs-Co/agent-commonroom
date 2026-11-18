package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func idleFixture(t *testing.T) *platformFixture {
	t.Helper()
	f := newPlatformFixture(t)
	f.ps.persistentSessions = true
	if _, err := f.store.Create(context.Background(), "acme", "Acme", defaultWorkspaceLimits); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *platformFixture) built(id string) bool {
	f.host.mu.Lock()
	defer f.host.mu.Unlock()
	e := f.host.entries[id]
	return e != nil && e.runtime != nil
}

func TestIdleWorkspaceIsReleasedAndRebuiltOnTheNextRequest(t *testing.T) {
	t.Parallel()
	f := idleFixture(t)
	first, err := f.host.runtimeFor(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if got := f.host.evictIdle(f.now.Add(29*time.Minute), 30*time.Minute); len(got) != 0 || !f.built("acme") {
		t.Fatalf("released too early: %v", got)
	}
	if got := f.host.evictIdle(f.now.Add(30*time.Minute), 30*time.Minute); len(got) != 1 || got[0] != "acme" || f.built("acme") {
		t.Fatalf("not released after the idle time: %v", got)
	}
	f.advance(31 * time.Minute)
	second, err := f.host.runtimeFor(context.Background(), "acme")
	if err != nil || second == first {
		t.Fatalf("not rebuilt: %v", err)
	}
	if !f.built("acme") {
		t.Fatal("rebuilt workspace is not registered")
	}
}

func TestUseKeepsAWorkspaceFromBeingReleased(t *testing.T) {
	t.Parallel()
	f := idleFixture(t)
	if _, err := f.host.runtimeFor(context.Background(), "acme"); err != nil {
		t.Fatal(err)
	}
	f.advance(20 * time.Minute)
	if _, err := f.host.runtimeFor(context.Background(), "acme"); err != nil { // used again
		t.Fatal(err)
	}
	f.advance(20 * time.Minute) // 40 minutes since the start but 20 since the last use
	if got := f.host.evictIdle(f.now, 30*time.Minute); len(got) != 0 {
		t.Fatalf("released a workspace used 20 minutes ago: %v", got)
	}
}

func TestWorkspaceIsNeverReleasedWhileBusy(t *testing.T) {
	t.Parallel()
	f := idleFixture(t)
	rt, err := f.host.runtimeFor(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	later := f.now.Add(2 * time.Hour)
	f.host.mu.Lock()
	f.host.entries["acme"].inflight = 1
	f.host.mu.Unlock()
	if got := f.host.evictIdle(later, 30*time.Minute); len(got) != 0 {
		t.Fatal("released with a request in flight")
	}
	f.host.mu.Lock()
	f.host.entries["acme"].inflight = 0
	f.host.mu.Unlock()
	rt.s.mu.Lock()
	rt.s.active["turn"] = true
	rt.s.mu.Unlock()
	if got := f.host.evictIdle(later, 30*time.Minute); len(got) != 0 {
		t.Fatal("released with an agent turn running")
	}
	rt.s.mu.Lock()
	rt.s.active["turn"] = false
	_, cancel := context.WithCancel(context.Background())
	rt.s.cancels["job"] = cancel
	rt.s.mu.Unlock()
	if got := f.host.evictIdle(later, 30*time.Minute); len(got) != 0 {
		t.Fatal("released with cancellable work registered")
	}
	rt.s.mu.Lock()
	delete(rt.s.cancels, "job")
	rt.s.mu.Unlock()
	if got := f.host.evictIdle(later, 30*time.Minute); len(got) != 1 {
		t.Fatal("a quiet idle workspace was not released")
	}
}

func TestWorkspaceWithoutSavedSessionsIsNotReleased(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t) // sessions are not persisted here
	if _, err := f.store.Create(context.Background(), "acme", "Acme", defaultWorkspaceLimits); err != nil {
		t.Fatal(err)
	}
	if _, err := f.host.runtimeFor(context.Background(), "acme"); err != nil {
		t.Fatal(err)
	}
	if got := f.host.evictIdle(f.now.Add(24*time.Hour), time.Minute); len(got) != 0 {
		t.Fatal("released a workspace whose sessions would be lost")
	}
}

func TestSingleWorkspaceModeNeverReleasesAnything(t *testing.T) {
	t.Parallel()
	f := idleFixture(t)
	if _, err := f.host.runtimeFor(context.Background(), "acme"); err != nil {
		t.Fatal(err)
	}
	f.host.tenancy.Mode = tenancySingle
	if got := f.host.evictIdle(f.now.Add(24*time.Hour), time.Minute); len(got) != 0 {
		t.Fatal("single mode released a workspace")
	}
}

func TestARequestInFlightIsCountedWhileItRuns(t *testing.T) {
	t.Parallel()
	f := idleFixture(t)
	f.host.tenancy.Mode = tenancyMulti
	rt, err := f.host.runtimeFor(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	rt.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
	})
	done := make(chan struct{})
	go func() {
		r := httptest.NewRequest("GET", "https://acme.rooms.example.com/slow", nil)
		r.Host = "acme.rooms.example.com"
		f.host.ServeHTTP(httptest.NewRecorder(), r)
		close(done)
	}()
	<-entered
	f.host.mu.Lock()
	busy := f.host.entries["acme"].inflight
	f.host.mu.Unlock()
	if got := f.host.evictIdle(f.now.Add(24*time.Hour), time.Minute); busy != 1 || len(got) != 0 {
		t.Fatalf("in flight %d, released %v", busy, got)
	}
	close(release)
	<-done
	f.host.mu.Lock()
	after := f.host.entries["acme"].inflight
	f.host.mu.Unlock()
	if after != 0 {
		t.Fatalf("in flight %d after the request ended", after)
	}
}

func TestParseWorkspaceIdle(t *testing.T) {
	t.Parallel()
	for raw, want := range map[string]time.Duration{"": 30 * time.Minute, "1": time.Minute, "90": 90 * time.Minute, "10080": 7 * 24 * time.Hour, "0": 0} {
		if got, err := parseWorkspaceIdle(raw); err != nil || got != want {
			t.Errorf("%q: %v %v", raw, got, err)
		}
	}
	for _, raw := range []string{"-1", "10081", "abc", " 5", "1.5", "99999999999999999999"} {
		if _, err := parseWorkspaceIdle(raw); err == nil {
			t.Errorf("%q accepted", raw)
		}
	}
}
