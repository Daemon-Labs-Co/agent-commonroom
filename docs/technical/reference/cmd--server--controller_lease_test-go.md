# cmd/server/controller_lease_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_lease_test.go)

Automated regression checks for controller_lease; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestControllerLeaseExcludesConcurrentStartup`** — Regression contract: TestControllerLeaseExcludesConcurrentStartup.
  Signature: `func(t *testing.T)`.
  Calls: `acquireControllerLease`, `before.Mode`, `before.Mode().Perm`, `filepath.Join`, `first.Close`, `os.SameFile`, `os.Stat`, `second.Close`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `third.Close`.
- **`TestControllerLeaseRejectsLinkedLock`** — Regression contract: TestControllerLeaseRejectsLinkedLock.
  Signature: `func(t *testing.T)`.
  Calls: `acquireControllerLease`, `filepath.Join`, `lease.Close`, `os.Link`, `os.ReadFile`, `os.Symlink`, `os.WriteFile`, `string`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
