package main

import (
	"net/http"
	"os"
	"strconv"
)

// connectionEditorHandler gates the optional presentation layer at the server.
// Disabling it leaves the existing editor and connection API available.
func connectionEditorHandler(enabled bool, root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if !enabled {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.ServeFile(w, r, root+"/connection-editor.js")
	})
}
func connectionEditorEnabled() (bool, error) {
	value := os.Getenv("COMMONROOM_CONNECTION_EDITOR")
	if value == "" {
		return true, nil
	}
	return strconv.ParseBool(value)
}
