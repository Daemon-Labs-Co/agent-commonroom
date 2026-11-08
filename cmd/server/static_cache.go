package main

import "net/http"

// staticHandler serves the browser UI from dir. Responses carry
// "Cache-Control: no-cache", so browsers revalidate (cheap 304 via
// Last-Modified) on every load instead of reusing a heuristically cached page
// or script after a deploy. It applies only to the UI files, never to /api.
func staticHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		files.ServeHTTP(w, r)
	})
}
