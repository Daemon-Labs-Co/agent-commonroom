# cmd/workspace-capture/main.go

[Module index](cmd.md) · [Open source](../../../cmd/workspace-capture/main.go)

Go implementation for main; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureFunc`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, workspacecapture.Request, workspacecapture.Authority, *os.Root, objectcustody.Journal) (workspacecapture.CapturedObjects, error)`.
- **`decodeRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(args []string) (workspacecapture.Request, error)`.
  Calls: `base64.RawURLEncoding.DecodeString`, `base64.RawURLEncoding.EncodeToString`, `bytes.Equal`, `json.Marshal`, `json.Unmarshal`, `len`, `workspacecapture.RequestDigest`.
- **`run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, stream io.ReadWriteCloser, openObjects func() (*os.Root, error), capture captureFunc) error`.
  Calls: `cancel`, `capture`, `channel.Close`, `channel.Finish`, `context.WithCancel`, `ctx.Err`, `decodeRequest`, `openObjects`, `root.Close`, `stream.Close`, `workspacecapture.NewCaptureChannel`.
- **Type `supervisorStdio`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ in, out *os.File }`.
- **`(*supervisorStdio).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `s.in.Read`.
- **`(*supervisorStdio).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `s.out.Write`.
- **`(*supervisorStdio).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `errors.Join`, `s.in.Close`, `s.out.Close`.
- **`main`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `fmt.Fprintln`, `os.Exit`, `os.OpenRoot`, `run`, `signal.NotifyContext`, `stop`.
