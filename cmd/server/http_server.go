package main

import (
	"net/http"
	"time"
)

// newHTTPServer centralizes mandatory transport limits. These limits supplement
// authentication and body limits; they are not a substitute for either. Keeping
// headers small limits unauthenticated parsing work before the handler runs.
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr: addr, Handler: handler,
		MaxHeaderBytes:    16 << 10,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      50 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
}
