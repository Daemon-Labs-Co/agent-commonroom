# cmd/server/p0_credential_acceptance_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/p0_credential_acceptance_test.go)

Automated regression checks for p0_credential_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0CredentialRefresh`** — Real credential lease/provision/stage/release operations, with temporary private files and a pre-write filesystem failure. No observations come from scenario.Want, provider processes, Docker or database fixtures.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `acquireWorkerEngineCredential`, `append`, `bytes.Equal`, `credentialLeaseFixture`, `errors.New`, `filepath.Join`, `fmt.Sprintf`, `info.Mode`, `info.Mode().Perm`, `lease.Provision`, `lease.Release`, `lease.StageRefresh`, `lease.releaseWithPersistence`, `lockWorkerCredential`, `os.Stat`, `readPrivateBounded`, `second.Release`, `strconv.FormatBool`, `strconv.Itoa`, `t.Fatal`, `t.Helper`, `unlock`, `workerCredentialReady`.
