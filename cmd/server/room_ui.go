package main

import (
	"context"
	"net/http"
	"time"
)

// roomPresentationScript enforces optional cosmetic modules at delivery time.
// This cannot change profile authorization, room access or immutable history.
func (s *server) roomPresentationScript(root, file, flag string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method not allowed", 405)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		st, err := s.load(ctx)
		if err != nil {
			http.Error(w, "Feature settings unavailable", 503)
			return
		}
		enabled, exists := st.FeatureFlags[flag]
		defaultEnabled := flag != "newestMessagesFirst"
		optionalScriptHandler((!exists && defaultEnabled) || (exists && enabled), root, file).ServeHTTP(w, r)
	})
}
