# cmd/server/http_server.go

[Module index](cmd.md) · [Open source](../../../cmd/server/http_server.go)

Go implementation for http server; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`newHTTPServer`** — newHTTPServer centralizes mandatory transport limits. These limits supplement authentication and body limits; they are not a substitute for either. Keeping headers small limits unauthenticated parsing work before the handler runs.
  Signature: `func(addr string, handler http.Handler) *http.Server`.
