package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPHeaderBoundary(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	s := newHTTPServer("127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- s.Serve(listener) }()
	t.Cleanup(func() { _ = s.Close(); <-done })
	request := func(size int) int {
		t.Helper()
		conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		if _, err = fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: localhost\r\nX-Test: %s\r\nConnection: close\r\n\r\n", strings.Repeat("a", size)); err != nil {
			t.Fatal(err)
		}
		response, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		return response.StatusCode
	}
	if status := request(1024); status != http.StatusNoContent {
		t.Fatalf("ordinary header: %d", status)
	}
	// net/http permits parser lookahead beyond MaxHeaderBytes. Exceed both the
	// configured budget and lookahead, without depending on an exact byte cutoff.
	if status := request(32 << 10); status != http.StatusRequestHeaderFieldsTooLarge {
		t.Fatalf("oversized header: %d", status)
	}
	if calls.Load() != 1 {
		t.Fatalf("oversized request reached handler; calls=%d", calls.Load())
	}
	if status := request(1024); status != http.StatusNoContent {
		t.Fatalf("server unavailable after rejection: %d", status)
	}
}

func TestHTTPTransportLimits(t *testing.T) {
	t.Parallel()
	s := newHTTPServer("127.0.0.1:8787", http.NotFoundHandler())
	if s.MaxHeaderBytes != 16<<10 || s.ReadHeaderTimeout != 5*time.Second || s.ReadTimeout != 15*time.Second || s.WriteTimeout != 50*time.Second || s.IdleTimeout != 15*time.Second {
		t.Fatalf("mandatory HTTP limits changed: %+v", s)
	}
}
