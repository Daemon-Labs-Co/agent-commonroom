# tools/harness-review/main.go

[Module index](tools.md) · [Open source](../../tools/harness-review/main.go)

Go implementation for main; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(args []string, out io.Writer, now time.Time) error`.
  Calls: `base64.StdEncoding.EncodeToString`, `base64.StdEncoding.Strict`, `base64.StdEncoding.Strict().DecodeString`, `ed25519.GenerateKey`, `ed25519.PrivateKey`, `flag.NewFlagSet`, `flags.NArg`, `flags.Parse`, `flags.SetOutput`, `flags.String`, `fmt.Fprintln`, `harnessrelease.Publish`, `harnessrelease.ReadPrivate`, `harnessrelease.Sign`, `string`, `strings.TrimSuffix`.
- **`main`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `fmt.Fprintln`, `os.Exit`, `run`, `time.Now`, `time.Now().UTC`.
