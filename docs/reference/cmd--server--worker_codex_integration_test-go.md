# cmd/server/worker_codex_integration_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_codex_integration_test.go)

Automated regression checks for worker_codex_integration; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `codexIntegrationDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { containers map[string]map[string]any volume map[string]any scope toolbroker.Scope engineDone chan error audit *transportTestAuditor creates, starts, helpers, removed int bootstrapped bool mode string }`.
- **Type `codexIntegrationAuditor`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { counter *transportTestAuditor mode string }`.
- **`(codexIntegrationAuditor).RecordIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent toolbroker.Intent) error`.
  Calls: `a.counter.RecordIntent`.
- **`(codexIntegrationAuditor).RecordOutcome`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent toolbroker.Intent, outcome toolbroker.Outcome) error`.
  Calls: `a.counter.RecordOutcome`.
- **`fixtureFlag`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(args []string, key string) string`.
  Calls: `len`.
- **`fixtureLabels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(args []string) map[string]string`.
  Calls: `strings.Cut`.
- **`(*codexIntegrationDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, input []byte, out io.Writer) error`.
  Calls: `append`, `base64.StdEncoding.EncodeToString`, `ctx.Err`, `delete`, `encode`, `errors.New`, `f.audit.calls.Load`, `fixtureFlag`, `fixtureLabels`, `fmt.Errorf`, `hex.EncodeToString`, `int64`, `io.WriteString`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `path.Join`, `sha256.Sum256`, `slices.Contains`, `slices.Index`, `slices.Sort`, `strings.Contains`, `strings.Cut`, `strings.HasPrefix`, `strings.Join`, `strings.Repeat`, `strings.Split`, `strings.TrimPrefix`, `strings.TrimSuffix`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Format`, `write`.
- **`(*codexIntegrationDocker).RunInput`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, input io.Reader, out io.Writer) error`.
  Calls: `ctx.Err`, `errors.New`, `io.ReadAll`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `string`, `strings.Repeat`.
- **`(*codexIntegrationDocker).OpenEngine`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (io.ReadWriteCloser, error)`.
  Calls: `ctx.Err`, `errors.New`, `net.Pipe`, `peer.Close`, `peer.SetDeadline`, `serveSyntheticCodex`, `strings.Repeat`, `time.Now`, `time.Now().Add`.
- **`serveSyntheticCodex`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(stream net.Conn, mode string) error`.
  Calls: `binary.BigEndian.Uint32`, `decoder.Decode`, `encoder.Encode`, `errors.New`, `fmt.Errorf`, `io.ReadFull`, `json.NewDecoder`, `json.NewEncoder`, `len`, `make`, `receive`, `send`, `string`, `strings.Contains`, `uint32`.
- **`TestOwnedCodexSyntheticCallbackAndFencedRetention`** — Regression contract: TestOwnedCodexSyntheticCallbackAndFencedRetention.
  Signature: `func(t *testing.T)`.
  Calls: `auditor.calls.Load`, `awaitAuditResult`, `bundle.closes.Load`, `connectionAuthKey`, `context.WithValue`, `credentialPresent`, `filepath.Join`, `int32`, `len`, `lockWorkerCredential`, `make`, `net.Pipe`, `newWorkerToolSession`, `os.MkdirAll`, `os.ReadDir`, `os.WriteFile`, `peer.Close`, `readPrivateBounded`, `s.RunCodex`, `session.Close`, `shortGatewayRoot`, `string`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `t.TempDir`, `time.After`, `time.Now`, `toolbroker.ServeAudit`, `unlock`, `workerCredentialReady`, `workspacePair`.
