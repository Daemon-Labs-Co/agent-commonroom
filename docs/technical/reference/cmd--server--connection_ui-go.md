# cmd/server/connection_ui.go

[Module index](cmd.md) · [Open source](../../../cmd/server/connection_ui.go)

Go implementation for connection ui; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`connectionEditorHandler`** — connectionEditorHandler gates the optional presentation layer at the server. Disabling it leaves the existing editor and connection API available.
  Signature: `func(enabled bool, root string) http.Handler`.
  Calls: `optionalScriptHandler`.
- **`optionalScriptHandler`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(enabled bool, root, filename string) http.Handler`.
  Calls: `http.Error`, `http.HandlerFunc`, `http.ServeFile`, `w.Header`, `w.Header().Set`, `w.WriteHeader`.
- **`connectionEditorEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (bool, error)`.
  Calls: `os.Getenv`, `strconv.ParseBool`.
