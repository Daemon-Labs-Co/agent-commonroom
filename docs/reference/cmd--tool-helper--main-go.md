# cmd/tool-helper/main.go

[Module index](cmd.md) · [Open source](../../cmd/tool-helper/main.go)

Go implementation for main; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `fileAction`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { operation string path string content string }`.
- **`decodeAction`** — decodeAction accepts exactly the broker Action wire fields. Revalidating here does not authenticate a caller: only the broker may authorize/admit operations.
  Signature: `func(in io.Reader) (fileAction, error)`.
  Calls: `bytes.Equal`, `bytes.NewReader`, `decoder.Decode`, `decoder.More`, `decoder.Token`, `io.LimitReader`, `io.ReadAll`, `json.Delim`, `json.NewDecoder`, `json.Unmarshal`, `len`, `make`, `strictjson.ValidEncoding`.
- **Type `fileResult`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Operation string 'json:"operation"' Bytes int 'json:"bytes"' SHA256 string 'json:"sha256"' Encoding string 'json:"encoding,omitempty"' Data string 'json:"data,omitempty"' }`.
- **`run`** — run borrows a trusted root for one operation. There is no command, patch, directory creation or alternate-root fallback. Output failure after a write is ambiguous and must not cause replay. The supervisor owns hard cancellation.
  Signature: `func(ctx context.Context, in io.Reader, out io.Writer, root *os.Root) error`.
  Calls: `append`, `base64.StdEncoding.EncodeToString`, `ctx.Err`, `decodeAction`, `errors.Is`, `hex.EncodeToString`, `json.Marshal`, `len`, `out.Write`, `sha256.Sum256`, `toolworkspace.New`, `workspace.Read`, `workspace.Write`.
- **`openWorkspace`** — openWorkspace uses a fixed container mount; neither argv nor the payload may select roots. os.Root is path containment, not container/mount attestation.
  Signature: `func() (*os.Root, error)`.
  Calls: `os.OpenRoot`, `parent.Close`, `repositoryRoot`.
- **`repositoryRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(parent *os.Root) (*os.Root, error)`.
  Calls: `info.IsDir`, `info.Mode`, `os.SameFile`, `parent.Lstat`, `parent.OpenRoot`, `root.Close`, `root.Stat`.
- **`main`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `errors.Is`, `fmt.Fprintln`, `len`, `openWorkspace`, `os.Exit`, `root.Close`, `run`, `signal.NotifyContext`, `stop`.
