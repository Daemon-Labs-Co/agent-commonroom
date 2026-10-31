# cmd/server/artifact_manifest_read_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_manifest_read_test.go)

Automated regression checks for artifact_manifest_read; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRecordedArtifactMemberRequiresFreshBytesAndExactManifest`** — Regression contract: TestRecordedArtifactMemberRequiresFreshBytesAndExactManifest.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `cancel`, `context.WithCancel`, `context.WithValue`, `db.seed`, `errors.New`, `filepath.Join`, `len`, `newControllerArtifactJournal`, `os.WriteFile`, `s.readRecordedArtifactMember`, `s.recordStoredArtifactManifest`, `store.Put`, `string`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestRecordedArtifactBundleMemberRefusesChangedSibling`** — Regression contract: TestRecordedArtifactBundleMemberRefusesChangedSibling.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalOutputFixture`, `artifactStore`, `filepath.Join`, `newControllerArtifactJournal`, `os.WriteFile`, `s.readRecordedArtifactMember`, `s.recordStoredArtifactManifest`, `store.Put`, `string`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
