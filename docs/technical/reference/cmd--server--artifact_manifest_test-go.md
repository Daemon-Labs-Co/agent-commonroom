# cmd/server/artifact_manifest_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_manifest_test.go)

Automated regression checks for artifact_manifest; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestStoredArtifactManifestUsesControllerReceiptsAndActualBytes`** — Regression contract: TestStoredArtifactManifestUsesControllerReceiptsAndActualBytes.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `db.seed`, `len`, `newControllerArtifactJournal`, `s.resolveStoredArtifactManifest`, `store.Put`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStoredArtifactManifestRefusesUntrustedOrIncompleteState`** — Regression contract: TestStoredArtifactManifestRefusesUntrustedOrIncompleteState.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `cancel`, `context.WithCancel`, `context.WithValue`, `db.seed`, `errors.New`, `filepath.Join`, `journal.Reserve`, `newControllerArtifactJournal`, `os.WriteFile`, `s.resolveStoredArtifactManifest`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **Type `manifestAckJournal`**: Hook runs after the real journal has released its server lock. This models a concurrent policy change after byte authorization, never reentrant callbacks.
  Declaration: `struct { objectcustody.Journal after func() }`.
- **`(*manifestAckJournal).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `j.Journal.AcknowledgeRead`, `j.after`.
- **`TestStoredArtifactManifestRechecksAfterAllReads`** — Regression contract: TestStoredArtifactManifestRechecksAfterAllReads.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `context.WithValue`, `db.seed`, `newControllerArtifactJournal`, `s.mu.Lock`, `s.mu.Unlock`, `s.resolveStoredArtifactManifest`, `store.Put`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStoredArtifactBundleRequiresCompleteSealedMembers`** — Regression contract: TestStoredArtifactBundleRequiresCompleteSealedMembers.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalOutputFixture`, `artifactState`, `artifactStore`, `len`, `newControllerArtifactJournal`, `s.resolveStoredArtifactManifest`, `store.Put`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`.
