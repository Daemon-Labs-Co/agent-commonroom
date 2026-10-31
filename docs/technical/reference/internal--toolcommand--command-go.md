# internal/toolcommand/command.go

[Module index](internal.md) · [Open source](../../../internal/toolcommand/command.go)

Go implementation for command; declarations and calls are listed below.

Package: `toolcommand`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`Valid`** — Valid permits only reviewed profiles; argv never selects a shell, executable path, environment, working directory, timeout or image.
  Signature: `func(argv []string) bool`.
  Calls: `slices.Equal`.
- **Type `Spec`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Program string Args, Env []string Dir string }`.
- **Type `Runner`**: Behavior contract; required method signatures are shown in the declaration.
  Declaration: `interface { Run(context.Context, Spec, io.Writer) (int, error) }`.
- **Type `Receipt`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Operation string 'json:"operation"' Bytes int 'json:"bytes"' SHA256 string 'json:"sha256"' Encoding string 'json:"encoding"' Data string 'json:"data"' ExitCode int 'json:"clientExitCode"' Completion string 'json:"completion"' CommandDigest string 'json:"commandDigest"' }`.
- **`Digest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(argv []string) string`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `sha256.Sum256`.
- **`Run`** — Run captures a combined bounded stdout/stderr receipt. ExitCode is the observed Docker CLI status. Nonzero remains unconfirmed: a daemon failure can look like a compiler failure, so it must not be certified as a completed test result. Success here is NOT descendant absence; the outer worker MUST remove/verify the tool container before releasing the receipt or executing more model work.
  Signature: `func(ctx context.Context, argv []string, runner Runner) (Receipt, error)`.
  Calls: `Digest`, `Valid`, `append`, `base64.StdEncoding.EncodeToString`, `cancel`, `context.WithTimeout`, `ctx.Err`, `hex.EncodeToString`, `len`, `out.mu.Lock`, `out.mu.Unlock`, `runner.Run`, `sha256.Sum256`.
- **Type `boundedWriter`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { mu sync.Mutex data []byte overflow, closed bool cancel context.CancelFunc }`.
- **`(*boundedWriter).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `append`, `len`, `w.cancel`, `w.mu.Lock`, `w.mu.Unlock`.
