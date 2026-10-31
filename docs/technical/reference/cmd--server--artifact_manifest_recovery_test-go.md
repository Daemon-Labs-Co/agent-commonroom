# cmd/server/artifact_manifest_recovery_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_manifest_recovery_test.go)

Automated regression checks for artifact_manifest_recovery; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRecordedArtifactRecoveryRequiresCurrentCustodyAndAuthority`** — Reopening descriptors and replaying immutable metadata must not carry the old controller's release authority across restart. Physical boundaries are mocked; this verifies recovery protocol behavior, never deployed custody readiness.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactManifestKey`, `artifactState`, `artifactStore`, `artifactTotal`, `context.WithValue`, `db.seed`, `errors.New`, `filepath.Join`, `int64`, `newControllerArtifactJournal`, `objectcustody.New`, `os.Chmod`, `os.OpenRoot`, `os.Remove`, `os.WriteFile`, `restarted.readRecordedArtifactMember`, `root.Close`, `s.recordStoredArtifactManifest`, `store.Put`, `string`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
