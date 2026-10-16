# cmd/server/http_server_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/http_server_test.go)

Automated regression checks for http_server; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestHTTPHeaderBoundary`** — Regression contract: TestHTTPHeaderBoundary.
  Signature: `func(t *testing.T)`.
  Calls: `bufio.NewReader`, `calls.Add`, `calls.Load`, `conn.Close`, `conn.SetDeadline`, `fmt.Fprintf`, `http.HandlerFunc`, `http.ReadResponse`, `listener.Addr`, `listener.Addr().String`, `make`, `net.DialTimeout`, `net.Listen`, `newHTTPServer`, `request`, `response.Body.Close`, `s.Close`, `s.Serve`, `strings.Repeat`, `t.Cleanup`, `t.Fatal`, `t.Fatalf`, `t.Helper`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.WriteHeader`.
- **`TestHTTPTransportLimits`** — Regression contract: TestHTTPTransportLimits.
  Signature: `func(t *testing.T)`.
  Calls: `http.NotFoundHandler`, `newHTTPServer`, `t.Fatalf`, `t.Parallel`.
