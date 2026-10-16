# cmd/server/events_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/events_test.go)

Automated regression checks for events; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`mockServer`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler)`.
  Calls: `auth.protect`, `http.NewServeMux`, `mux.HandleFunc`, `newMockDB`, `newOwnerAuth`, `s.actorGuard`, `strings.Repeat`, `t.Fatal`, `time.Now`, `time.Now().Add`.
- **`mockRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, method, path, body string) *httptest.ResponseRecorder`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`.
- **`TestEventValidationNoWrites`** — Regression contract: TestEventValidationNoWrites.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `len`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestEventProjectionRoundTrip`** — Regression contract: TestEventProjectionRoundTrip.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.snapshot`, `json.Marshal`, `len`, `mockRequest`, `mockServer`, `s.load`, `string`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `w.Body.String`.
- **`TestAuditReadAndIntegrityFailures`** — Regression contract: TestAuditReadAndIntegrityFailures.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `db.snapshot`, `errors.New`, `handler`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Unmarshal`, `len`, `mockServer`, `s.events`, `s.integrity`, `s.load`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.Bytes`, `w.Body.String`.
