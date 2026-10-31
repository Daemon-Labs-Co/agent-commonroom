# cmd/object-read/main.go

[Module index](cmd.md) · [Open source](../../../cmd/object-read/main.go)

Go implementation for main; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`decodeRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(args []string) (objectread.Request, error)`.
  Calls: `base64.RawURLEncoding.DecodeString`, `base64.RawURLEncoding.EncodeToString`, `bytes.Equal`, `json.Marshal`, `json.Unmarshal`, `len`, `objectread.ValidRequest`.
- **`run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, stream io.ReadWriteCloser, serve func(context.Context, objectread.Request, io.ReadWriteCloser) error) error`.
  Calls: `ctx.Err`, `decodeRequest`, `serve`, `stream.Close`.
- **Type `supervisorStdio`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ in, out *os.File }`.
- **`(*supervisorStdio).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(b []byte) (int, error)`.
  Calls: `s.in.Read`.
- **`(*supervisorStdio).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(b []byte) (int, error)`.
  Calls: `s.out.Write`.
- **`(*supervisorStdio).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `errors.Join`, `s.in.Close`, `s.out.Close`.
- **`main`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `fmt.Fprintln`, `os.Exit`, `run`, `signal.NotifyContext`, `stop`.
