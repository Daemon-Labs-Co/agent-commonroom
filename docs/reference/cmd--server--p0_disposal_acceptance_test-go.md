# cmd/server/p0_disposal_acceptance_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/p0_disposal_acceptance_test.go)

Automated regression checks for p0_disposal_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0DisposalRefusal`** — Negative cases drive the real owner HTTP boundary and observed journal/ledger state. Positive disposal stays pending until exact storage credit is supported.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `f.calls.Load`, `f.db.seed`, `f.db.snapshot`, `f.now.Store`, `f.post`, `f.propose`, `f.s.load`, `len`, `managedStorageTotal`, `newDisposalFixture`, `p.NotBefore.UnixNano`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `t.Setenv`, `time.Now`, `time.Now().Add`.
