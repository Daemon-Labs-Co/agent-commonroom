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
	if e := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Public guide"), 0600); e != nil {
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
	if e := os.Remove(filepath.Join(root, "README.md")); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink(filepath.Join(root, ".data", "owner-key"), filepath.Join(root, "README.md")); e != nil {
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
	if e := os.Symlink(filepath.Join(root, ".data"), filepath.Join(root, "docs")); e != nil {
		t.Fatal(e)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/help/document/index", nil))
	if w.Code != 404 {
		t.Fatal("symlinked docs directory served")
	}
}
