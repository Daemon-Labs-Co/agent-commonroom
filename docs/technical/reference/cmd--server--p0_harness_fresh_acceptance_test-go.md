# cmd/server/p0_harness_fresh_acceptance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_harness_fresh_acceptance_test.go)

Automated regression checks for p0_harness_fresh_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0HarnessFreshPair`** — Exercise signed controller allocation through the owned Codex assembly path. Docker and the provider protocol peer are synthetic; no real provider runs.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `(&server{db: db}).load`, `admitWorkerPair`, `append`, `awaitAuditResult`, `base64.StdEncoding.EncodeToString`, `bundle.closes.Load`, `connectionAuthKey`, `context.WithValue`, `counter.calls.Load`, `ctx.Value`, `db.seed`, `filepath.Join`, `harnessFixture`, `job.Harness.CheckedAt.Before`, `len`, `make`, `managedJobFence`, `managedPairFixtureVersion`, `net.Pipe`, `newWorkerToolSession`, `os.MkdirAll`, `os.WriteFile`, `pairRequest`, `peer.Close`, `readWorkerRecord`, `reflect.DeepEqual`, `registration.Register`, `s.load`, `s.managedExecutionJobContextVersion`, `session.Close`, `shortGatewayRoot`, `signHarness`, `strconv.FormatBool`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `t.Setenv`, `t.TempDir`, `time.After`, `time.Now`, `toolbroker.ServeAudit`, `validManagedHarness`, `worker.Close`, `worker.RunCodex`.
- **`TestCodexProviderOwnedPairContract`** — Keep the catalog scenario pending until its host-launch observation has an independently measured boundary; this regression proves only the owned pair.
  Signature: `func(t *testing.T)`.
  Calls: `driveP0HarnessFreshPair`, `reflect.DeepEqual`, `t.Fatal`.
