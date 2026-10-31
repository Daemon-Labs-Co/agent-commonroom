# cmd/server/connection_ui_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/connection_ui_test.go)

Automated regression checks for connection_ui; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestConnectionEditorFeatureFlag`** — Regression contract: TestConnectionEditorFeatureFlag.
  Signature: `func(t *testing.T)`.
  Calls: `connectionEditorHandler`, `connectionEditorHandler(enabled, root).ServeHTTP`, `filepath.Join`, `httptest.NewRecorder`, `httptest.NewRequest`, `os.WriteFile`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `w.Body.Len`, `w.Body.String`, `w.Header`, `w.Header().Get`.
