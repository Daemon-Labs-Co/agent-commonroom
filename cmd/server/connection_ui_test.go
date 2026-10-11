package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConnectionEditorFeatureFlag(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "connection-editor.js"), []byte("/* guided editor */"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []bool{true, false} {
		w := httptest.NewRecorder()
		connectionEditorHandler(enabled, root).ServeHTTP(w, httptest.NewRequest("GET", "/connection-editor.js", nil))
		if enabled && (w.Code != 200 || !strings.Contains(w.Body.String(), "guided editor")) {
			t.Fatal("enabled editor not served")
		}
		if !enabled && (w.Code != 204 || w.Body.Len() != 0) {
			t.Fatal("disabled editor was served")
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("flag change could be cached")
		}
		w = httptest.NewRecorder()
		connectionEditorHandler(enabled, root).ServeHTTP(w, httptest.NewRequest("POST", "/connection-editor.js", nil))
		if w.Code != 405 {
			t.Fatal("mutation accepted")
		}
	}
}
