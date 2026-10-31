# cmd/server/identity_defaults_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/identity_defaults_test.go)

Automated regression checks for identity_defaults; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestMissingRequestActorCannotAcquireOwnerAuthority`** — Regression contract: TestMissingRequestActorCannotAcquireOwnerAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `c.handler`, `context.WithValue`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.Context`, `r.WithContext`, `requestActor`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
- **`TestSessionWithoutActorIsRejectedAndRemoved`** — Regression contract: TestSessionWithoutActorIsRejectedAndRemoved.
  Signature: `func(t *testing.T)`.
  Calls: `a.protect`, `a.protect(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP`, `a.session`, `http.HandlerFunc`, `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `newOwnerAuth`, `r.AddCookie`, `sessionDigest`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
