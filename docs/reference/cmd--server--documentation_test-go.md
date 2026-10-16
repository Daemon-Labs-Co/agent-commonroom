# cmd/server/documentation_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/documentation_test.go)

Automated regression checks for documentation; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestDocumentationAllowlist`** — Regression contract: TestDocumentationAllowlist.
  Signature: `func(t *testing.T)`.
  Calls: `documentationHandler`, `filepath.Join`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `os.Mkdir`, `os.Remove`, `os.Symlink`, `os.WriteFile`, `strings.Contains`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.TempDir`, `w.Body.Len`, `w.Body.String`.
- **`TestDocumentationCatalogTargets`** — Catalog changes must keep contextual Help destinations available and unique.
  Signature: `func(t *testing.T)`.
  Calls: `documentationHandler`, `documentationHandler("../..").ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `t.Errorf`, `t.Fatalf`, `t.Parallel`.
