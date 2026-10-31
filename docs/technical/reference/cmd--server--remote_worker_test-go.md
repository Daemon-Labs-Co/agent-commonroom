# cmd/server/remote_worker_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/remote_worker_test.go)

Automated regression checks for remote_worker; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`sampleWorkerRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() workerRunRequest`.
  Calls: `append`, `providerTurnArgs`, `strings.Repeat`, `time.Now`, `time.Now().Add`, `time.Now().Add(179 * time.Second).UTC`, `time.Now().UTC`.
- **`TestRemoteWorkerRegistrationAndFlag`** — Regression contract: TestRemoteWorkerRegistrationAndFlag.
  Signature: `func(t *testing.T)`.
  Calls: `remoteWorkerEndpoint`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerProtocolRejectsArbitraryExecution`** — Regression contract: TestWorkerProtocolRejectsArbitraryExecution.
  Signature: `func(t *testing.T)`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `context.Background`, `decodeWorkerRequest`, `encodeWorkerRequest`, `mutation`, `sampleWorkerRequest`, `strings.Contains`, `strings.Join`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `workerSSHCommand`.
- **`TestWorkerRecordsAndDuplicateSubmission`** — Regression contract: TestWorkerRecordsAndDuplicateSubmission.
  Signature: `func(t *testing.T)`.
  Calls: `fi.Mode`, `fi.Mode().Perm`, `filepath.Join`, `os.Stat`, `readWorkerRecord`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `workerTerminal`, `writeWorkerRecord`.
- **`TestWorkerOutputFailureCancelsExecution`** — Regression contract: TestWorkerOutputFailureCancelsExecution.
  Signature: `func(t *testing.T)`.
  Calls: `b.Len`, `cancel`, `context.Background`, `context.WithCancel`, `ctx.Err`, `t.Fatal`, `t.Parallel`, `w.Write`.
