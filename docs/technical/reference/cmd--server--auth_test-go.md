# cmd/server/auth_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/auth_test.go)

Automated regression checks for auth; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestOwnerBoundary`** — Regression contract: TestOwnerBoundary.
  Signature: `func(t *testing.T)`.
  Calls: `a.protect`, `a.protect(mux).ServeHTTP`, `a.session`, `http.NewServeMux`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Unmarshal`, `mux.HandleFunc`, `newOwnerAuth`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.Bytes`, `w.Body.String`, `w.Result`, `w.Result().Cookies`.
- **`TestOwnerLoginRateLimit`** — Regression contract: TestOwnerLoginRateLimit.
  Signature: `func(t *testing.T)`.
  Calls: `httptest.NewRecorder`, `httptest.NewRequest`, `newOwnerAuth`, `s.login`, `strings.NewReader`, `strings.Repeat`, `t.Fatalf`, `t.Parallel`.
- **`TestOwnerKeyFile`** — Regression contract: TestOwnerKeyFile.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `loadOwnerKey`, `os.Chmod`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
