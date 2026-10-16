# cmd/server/documentation.go

[Module index](cmd.md) · [Open source](../../cmd/server/documentation.go)

Go implementation for documentation; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `documentationEntry`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID string 'json:"id"' Title string 'json:"title"' Path string 'json:"path"' }`.
- **`documentationHandler`** — documentationHandler serves only the explicit public documentation catalog. Private configuration, arbitrary paths, and symlinked files are never served.
  Signature: `func(root string) http.Handler`.
  Calls: `f.Close`, `f.Stat`, `filepath.Join`, `http.Error`, `http.HandlerFunc`, `http.NotFound`, `http.ServeContent`, `info.ModTime`, `info.Mode`, `info.Mode().IsRegular`, `info.Name`, `json.NewEncoder`, `json.NewEncoder(w).Encode`, `os.Lstat`, `os.Open`, `strings.Split`, `w.Header`, `w.Header().Set`.
