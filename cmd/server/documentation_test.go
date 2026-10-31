package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentationAllowlist(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if e := os.MkdirAll(filepath.Join(root, "docs", "technical", "archive"), 0700); e != nil {
		t.Fatal(e)
	}
	guide := filepath.Join(root, "docs", "technical", "archive", "installation.md")
	if e := os.WriteFile(guide, []byte("# Public guide"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.Mkdir(filepath.Join(root, ".data"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, ".data", "owner-key"), []byte("private fixture"), 0600); e != nil {
		t.Fatal(e)
	}
	h := documentationHandler(root)
	for _, tc := range []struct {
		method, path string
		status       int
	}{{"GET", "/help/document/installation", 200}, {"HEAD", "/help/document/installation", 200}, {"GET", "/help/catalog", 200}, {"POST", "/help/catalog", 405}, {"GET", "/help/document/../.data/owner-key", 404}, {"GET", "/help/document/.data/owner-key", 404}, {"GET", "/help/document/unknown", 404}} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status {
			t.Fatalf("%s %s: %d", tc.method, tc.path, w.Code)
		}
		if strings.Contains(w.Body.String(), "private fixture") {
			t.Fatal("private data exposed")
		}
		if tc.method == "HEAD" && w.Body.Len() != 0 {
			t.Fatal("HEAD returned body")
		}
	}
	if e := os.Remove(guide); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink(filepath.Join(root, ".data", "owner-key"), guide); e != nil {
		t.Fatal(e)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/help/document/installation", nil))
	if w.Code != 404 {
		t.Fatal("symlinked credential served")
	}
	if e := os.WriteFile(filepath.Join(root, ".data", "README.md"), []byte("private fixture"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.Remove(guide); e != nil {
		t.Fatal(e)
	}
	if e := os.RemoveAll(filepath.Join(root, "docs")); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink(filepath.Join(root, ".data"), filepath.Join(root, "docs")); e != nil {
		t.Fatal(e)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/help/document/index", nil))
	if w.Code != 404 {
		t.Fatal("symlinked docs directory served")
	}
}

// Catalog changes must keep contextual Help destinations available and unique.
func TestDocumentationCatalogTargets(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, entry := range documentationCatalog {
		if entry.ID == "" || entry.Title == "" || seen[entry.ID] {
			t.Fatalf("invalid/duplicate Help entry %q", entry.ID)
		}
		seen[entry.ID] = true
		w := httptest.NewRecorder()
		documentationHandler("../..").ServeHTTP(w, httptest.NewRequest("GET", "/help/document/"+entry.ID, nil))
		if w.Code != 200 {
			t.Errorf("Help %s (%s): HTTP %d", entry.ID, entry.Path, w.Code)
		}
	}
}

func TestDocumentationAudienceIllustrations(t *testing.T) {
	t.Parallel()
	h := documentationHandler("../..")
	for _, topic := range []string{"getting-started", "rooms", "agents", "tasks", "workflows", "results", "storage", "recovery", "settings"} {
		for _, method := range []string{"GET", "HEAD"} {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(method, "/help/image/user-"+topic+".svg", nil))
			if w.Code != 200 || w.Header().Get("Content-Type") != "image/svg+xml" {
				t.Fatal(topic, w.Code, w.Header())
			}
			if method == "HEAD" && w.Body.Len() != 0 {
				t.Fatal("HEAD image returned body")
			}
		}
	}
	for _, path := range []string{"/help/image/unlisted.svg", "/help/image/../../.data/private.svg"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatal("unlisted image served", path, w.Code)
		}
	}
}
