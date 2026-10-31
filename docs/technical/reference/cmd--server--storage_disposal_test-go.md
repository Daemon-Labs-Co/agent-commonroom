# cmd/server/storage_disposal_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/storage_disposal_test.go)

Automated regression checks for storage_disposal; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestBoundWorkspaceStorageDisposal`** — Regression contract: TestBoundWorkspaceStorageDisposal.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: f.db}).load`, `errors.New`, `f.approve`, `f.calls.Load`, `f.now.Store`, `f.propose`, `json.Marshal`, `len`, `managedStorageTotal`, `newDisposalFixture`, `original`, `p.NotBefore.UnixNano`, `projectManagedStorage`, `t.Context`, `t.Fatal`, `t.Run`, `t.Setenv`.
- **`TestStorageReleaseRejectsForgedEvidence`** — Regression contract: TestStorageReleaseRejectsForgedEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `json.Marshal`, `projectManagedStorage`, `s.load`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`.
- **`TestStorageWorkspaceIdentityValidation`** — Regression contract: TestStorageWorkspaceIdentityValidation.
  Signature: `func(t *testing.T)`.
  Calls: `json.Marshal`, `projectManagedStorage`, `storageRecord`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Format`, `validStorageReservation`.
- **`TestStorageReleaseScopeCannotChange`** — Regression contract: TestStorageReleaseScopeCannotChange.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedStorage`, `f.approve`, `f.now.Store`, `f.propose`, `f.s.load`, `managedStorageTotal`, `maps.Clone`, `newDisposalFixture`, `p.NotBefore.UnixNano`, `result.Body.String`, `storageRecord`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Run`, `t.Setenv`, `validStorageRelease`.
