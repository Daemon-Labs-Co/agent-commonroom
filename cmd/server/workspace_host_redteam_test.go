package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func redteamHost(build func(id string) (*workspaceRuntime, error)) (*workspaceHost, *time.Time, func()) {
	tenancy := tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}
	now := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	store := &memoryStore{records: map[string]workspaceRecord{}, clock: func() time.Time { return now }}
	store.records["acme"] = workspaceRecord{ID: "acme", Name: "Acme", Status: workspaceActive, Limits: defaultWorkspaceLimits}
	h := newWorkspaceHost(tenancy, store, build)
	h.now = func() time.Time { return now }
	return h, &now, func() {}
}

func stubRuntime(id string) *workspaceRuntime {
	return &workspaceRuntime{id: id, s: &server{}, stop: func() {}, handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })}
}

// RT-3: a suspend (evict) that lands while a workspace is being built makes the request
// that triggered the build dereference a nil map entry.
func TestRedteamEvictDuringFirstUseDoesNotPanicRouting(t *testing.T) {
	t.Parallel()
	var h *workspaceHost
	h, _, _ = redteamHost(func(id string) (*workspaceRuntime, error) {
		h.evict(id) // platform suspend arriving while the server is still starting
		return stubRuntime(id), nil
	})
	r := httptest.NewRequest("GET", "https://acme.rooms.example.com/api/auth/methods", nil)
	r.Host = "acme.rooms.example.com"
	w := httptest.NewRecorder()
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("host router panicked: %v", p)
		}
	}()
	h.ServeHTTP(w, r)
}

// RT-4: entry.runtime is written inside sync.Once without h.mu while another request
// reads it under h.mu. Detected only with -race (which `make check` runs).
func TestRedteamResolveRuntimeFieldRace(t *testing.T) {
	t.Parallel()
	h, now, _ := redteamHost(func(id string) (*workspaceRuntime, error) {
		time.Sleep(100 * time.Millisecond) // a slow start; no happens-before edge to other requests
		return stubRuntime(id), nil
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.runtimeFor(t.Context(), "acme")
	}()
	time.Sleep(20 * time.Millisecond)
	h.mu.Lock()
	*now = now.Add(10 * time.Second) // registry answer is stale, so the second request re-checks
	h.mu.Unlock()
	go h.runtimeFor(t.Context(), "acme")
	<-done
}
