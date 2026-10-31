# cmd/server/artifact_manifest_registry_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_manifest_registry_test.go)

Automated regression checks for artifact_manifest_registry; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestArtifactManifestRecordedOnceAndReplayed`** — Regression contract: TestArtifactManifestRecordedOnceAndReplayed.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactManifestKey`, `artifactState`, `artifactStore`, `db.snapshot`, `json.Marshal`, `newControllerArtifactJournal`, `s.recordStoredArtifactManifest`, `store.Put`, `string`, `strings.Contains`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
- **`TestArtifactManifestCommitFailureRetainsCustodyWithoutSuccess`** — Regression contract: TestArtifactManifestCommitFailureRetainsCustodyWithoutSuccess.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `len`, `newControllerArtifactJournal`, `s.recordStoredArtifactManifest`, `store.Put`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestArtifactManifestProjectionRejectsForgeryAndClientInjection`** — Regression contract: TestArtifactManifestProjectionRejectsForgeryAndClientInjection.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `artifactJournalFixture`, `artifactState`, `artifactStore`, `db.snapshot`, `http.HandlerFunc`, `json.Marshal`, `len`, `mockRequest`, `newControllerArtifactJournal`, `projectArtifactManifest`, `response.Body.String`, `s.actorGuard`, `s.auth.protect`, `s.resolveStoredArtifactManifest`, `store.Put`, `string`, `strings.Contains`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestArtifactManifestRecordingAfterEditingDisabled`** — Regression contract: TestArtifactManifestRecordingAfterEditingDisabled.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactManifestKey`, `artifactState`, `artifactStore`, `db.seed`, `newControllerArtifactJournal`, `s.recordStoredArtifactManifest`, `store.Put`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
- **`TestArtifactManifestConcurrentRecordHasOneWinner`** — Regression contract: TestArtifactManifestConcurrentRecordHasOneWinner.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `close`, `db.snapshot`, `len`, `make`, `newControllerArtifactJournal`, `s.recordStoredArtifactManifest`, `store.Put`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
