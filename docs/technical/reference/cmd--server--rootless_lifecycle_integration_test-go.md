# cmd/server/rootless_lifecycle_integration_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/rootless_lifecycle_integration_test.go)

Automated regression checks for rootless_lifecycle_integration; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRootlessCancellationIntegration`** — Opt-in maintenance smoke test: a credential-free, network-disabled Node timer, never a provider agent. It creates no database data or live job-record fixtures.
  Signature: `func(t *testing.T)`.
  Calls: `admission.Close`, `cancel`, `cmd.Run`, `configureWorker`, `context.Background`, `context.WithTimeout`, `done`, `exec.CommandContext`, `fenceWorkerJob`, `id`, `len`, `managedWorkerContainers`, `os.Getenv`, `request.record`, `stopWorkerContainer`, `t.Error`, `t.Fatal`, `t.Skip`, `t.TempDir`, `workerIdleProbe`, `workerStateDirectory`, `workerWaitLease`, `writeWorkerRecord`.
