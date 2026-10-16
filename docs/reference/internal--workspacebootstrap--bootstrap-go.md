# internal/workspacebootstrap/bootstrap.go

[Module index](internal.md) · [Open source](../../internal/workspacebootstrap/bootstrap.go)

Go implementation for bootstrap; declarations and calls are listed below.

Package: `workspacebootstrap`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Receipt`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Job string 'json:"job"' Repository string 'json:"repository"' Commit string 'json:"commit"' Branch string 'json:"branch"' }`.
- **`hexDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string, n int) bool`.
  Calls: `len`.
- **`ValidReceipt`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r Receipt, job, digest string) bool`.
  Calls: `hexDigest`.
- **Type `Git`**: Git is a trusted adapter: production invokes only fixed /usr/bin/git with a fresh HOME, no inherited configuration, hooks or network. Tests mock it. Output must be bounded to MaxTreeBytes; error text must not include source. A source fork CAN supply a host runner, BUT invalidates isolation guarantees.
  Declaration: `interface { Run(context.Context, string, ...string) ([]byte, error) }`.
- **`Run`** — Run owns a fresh, empty tool volume. The original bounded bundle is streamed and hashed before Git is invoked. Any failure leaves a nonempty unresolved workspace; it never removes or resets prior content or retries uncertain work. Container/journal lifetime and shutdown remain the worker's responsibility.
  Signature: `func(ctx context.Context, in io.Reader, out io.Writer, workspace, job, digest string, git Git) error`.
  Calls: `ValidReceipt`, `append`, `ctx.Err`, `f.Close`, `f.Sync`, `file.Close`, `file.Name`, `file.Sync`, `file.WriteString`, `filepath.Join`, `git.Run`, `hash.Sum`, `hex.EncodeToString`, `hexDigest`, `info.IsDir`, `info.Mode`, `io.Copy`, `io.LimitReader`, `io.MultiWriter`, `json.Marshal`, `len`, `os.CreateTemp`, `os.Lstat`, `os.MkdirAll`, `os.OpenFile`, `os.ReadDir`, `os.Remove`, `os.WriteFile`, `out.Write`, `run`, `sha256.New`, `string`, `strings.TrimSpace`, `validTree`.
- **Type `contextReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ctx context.Context in io.Reader }`.
- **`(*contextReader).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `r.ctx.Err`, `r.in.Read`.
- **`validTree`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) bool`.
  Calls: `hexDigest`, `len`, `path.Clean`, `path.IsAbs`, `strconv.ParseInt`, `string`, `strings.ContainsAny`, `strings.Cut`, `strings.EqualFold`, `strings.Fields`, `strings.HasPrefix`, `strings.Split`.
