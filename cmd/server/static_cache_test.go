package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStaticFilesRevalidateAndAPIIsUntouched(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>ui</html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "features.js"), []byte("// script"), 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.Handle("/", staticHandler(dir))
	for _, path := range []string{"/", "/features.js?v=x"} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-cache" || w.Header().Get("Last-Modified") == "" {
			t.Fatal(path, w.Code, w.Header())
		}
	}
	// Go redirects /index.html to /; the redirect is not cached either.
	redirect := httptest.NewRecorder()
	mux.ServeHTTP(redirect, httptest.NewRequest("GET", "/index.html", nil))
	if redirect.Code != http.StatusMovedPermanently || redirect.Header().Get("Cache-Control") != "no-cache" {
		t.Fatal("redirect", redirect.Code, redirect.Header())
	}
	// A conditional request still gets a cheap 304 instead of the full file.
	first := httptest.NewRecorder()
	mux.ServeHTTP(first, httptest.NewRequest("GET", "/features.js", nil))
	again := httptest.NewRequest("GET", "/features.js", nil)
	again.Header.Set("If-Modified-Since", first.Header().Get("Last-Modified"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, again)
	if w.Code != http.StatusNotModified || w.Header().Get("Cache-Control") != "no-cache" {
		t.Fatal("conditional request not revalidated", w.Code)
	}
	api := httptest.NewRecorder()
	mux.ServeHTTP(api, httptest.NewRequest("GET", "/api/ping", nil))
	if api.Header().Get("Cache-Control") != "" {
		t.Fatal("API responses must not be changed by the static wrapper")
	}
}
