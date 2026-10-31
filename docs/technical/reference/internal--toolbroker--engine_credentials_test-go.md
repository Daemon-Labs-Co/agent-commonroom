# internal/toolbroker/engine_credentials_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/engine_credentials_test.go)

Automated regression checks for engine_credentials; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `exportEngineFake`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *engineDockerFake exports int mode string }`.
- **`(*exportEngineFake).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, input []byte, out io.Writer) error`.
  Calls: `errors.New`, `f.engineDockerFake.Run`, `io.WriteString`, `len`, `slices.Equal`, `strings.Repeat`.
- **`TestEnginePrivateCredentialExportOneShotAndPolicyBound`** — Regression contract: TestEnginePrivateCredentialExportOneShotAndPolicyBound.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `len`, `r.Attach`, `r.ExportCredential`, `stream.Close`, `streamFixture`, `string`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
