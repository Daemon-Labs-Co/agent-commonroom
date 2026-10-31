# internal/workspacebootstrap/bootstrap_test.go

[Module index](internal.md) · [Open source](../../../internal/workspacebootstrap/bootstrap_test.go)

Automated regression checks for bootstrap; read test names for individual contracts.

Package: `workspacebootstrap`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `fakeGit`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T mode string calls []string }`.
- **`(*fakeGit).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, args ...string) ([]byte, error)`.
  Calls: `append`, `errors.New`, `f.t.Fatal`, `filepath.Join`, `len`, `os.MkdirAll`, `os.ReadFile`, `os.WriteFile`, `string`, `strings.Contains`, `strings.HasPrefix`, `strings.Repeat`.
- **`TestBootstrapHashBeforeGitAndFreshWorkspace`** — Regression contract: TestBootstrapHashBeforeGitAndFreshWorkspace.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `ValidReceipt`, `bytes.NewReader`, `filepath.Join`, `hex.EncodeToString`, `json.Unmarshal`, `len`, `os.IsNotExist`, `os.Stat`, `os.WriteFile`, `output.Bytes`, `output.Len`, `sha256.Sum256`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **Type `shortOutput`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(shortOutput).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
- **`TestBootstrapTreeLimits`** — Regression contract: TestBootstrapTreeLimits.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `bytes.Repeat`, `record`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validTree`.
- **`TestBootstrapCancellationAndSymlink`** — Regression contract: TestBootstrapCancellationAndSymlink.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `cancel`, `context.WithCancel`, `filepath.Join`, `len`, `os.ReadDir`, `os.Symlink`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
