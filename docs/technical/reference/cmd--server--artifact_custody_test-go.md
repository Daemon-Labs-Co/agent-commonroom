# cmd/server/artifact_custody_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_custody_test.go)

Automated regression checks for artifact_custody; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticArtifactGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { source func(context.Context, room.State, objectcustody.Intent) error seal func(context.Context, room.State, objectcustody.Receipt) error quota func(context.Context, room.StorageReservation) error release func(context.Context, room.State, objectcustody.Receipt) error }`.
- **`(*syntheticArtifactGuard).ValidateSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, s room.State, i objectcustody.Intent) error`.
  Calls: `g.source`.
- **`(*syntheticArtifactGuard).ValidateSeal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, s room.State, r objectcustody.Receipt) error`.
  Calls: `g.seal`.
- **`(*syntheticArtifactGuard).EstablishQuota`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, r room.StorageReservation) error`.
  Calls: `g.quota`.
- **`(*syntheticArtifactGuard).AuthorizeRelease`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, s room.State, r objectcustody.Receipt) error`.
  Calls: `g.release`.
- **`artifactJournalFixture`** — Real controller admission/replay with only a synthetic task turn and mocked DB. This deliberately does not constitute trusted source/quota or live acceptance.
  Signature: `func(t *testing.T, inventory bool) (*server, *mockDB, context.Context, objectcustody.Scope)`.
  Calls: `artifactJournalOutputFixture`, `t.Helper`.
- **`artifactJournalOutputFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, inventory bool, output taskcontract.DataContract) (*server, *mockDB, context.Context, objectcustody.Scope)`.
  Calls: `append`, `artifactState`, `close`, `context.WithValue`, `definedDispatchBody`, `dispatchRequest`, `fmt.Sprintf`, `jc.Value`, `json.Marshal`, `make`, `mockRequest`, `reg.Register`, `reg.Update`, `s.initializeManagedStorage`, `s.load`, `s.managedExecutionJobContext`, `storageInventoryFunc`, `strings.Repeat`, `structuredTaskFixture`, `t.Cleanup`, `t.Context`, `t.Error`, `t.Fatal`, `t.Helper`, `time.After`, `w.Body.String`, `waitDispatch`.
- **`artifactStore`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, j objectcustody.Journal) (*objectcustody.Store, string)`.
  Calls: `objectcustody.New`, `os.Chmod`, `os.OpenRoot`, `root.Close`, `t.Cleanup`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`artifactState`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, s *server, ctx context.Context) room.State`.
  Calls: `s.load`, `t.Fatal`, `t.Helper`.
- **`artifactTotal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, s *server, ctx context.Context) int64`.
  Calls: `artifactState`, `managedStorageTotal`, `t.Fatal`, `t.Helper`.
- **`TestArtifactJournalActualBytesDurableChargeSealAndRead`** — Regression contract: TestArtifactJournalActualBytesDurableChargeSealAndRead.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `db.seed`, `fresh.AuthorizeRead`, `json.Marshal`, `newControllerArtifactJournal`, `objectcustody.New`, `os.OpenRoot`, `readStore.Get`, `root.Close`, `store.Put`, `string`, `strings.Contains`, `strings.NewReader`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **Type `artifactReadProbe`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { reads int onRead func() reader io.Reader }`.
- **`(*artifactReadProbe).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `r.onRead`, `r.reader.Read`.
- **`TestArtifactJournalRefusesBeforeSourceRead`** — Regression contract: TestArtifactJournalRefusesBeforeSourceRead.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `cancel`, `context.WithCancel`, `context.WithValue`, `db.seed`, `errors.New`, `len`, `newControllerArtifactJournal`, `os.ReadDir`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestArtifactJournalSealAndReadFailuresWithholdSuccessAndRetainBytes`** — Regression contract: TestArtifactJournalSealAndReadFailuresWithholdSuccessAndRetainBytes.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `db.seed`, `errors.New`, `j.Reserve`, `j.Seal`, `len`, `newControllerArtifactJournal`, `os.ReadDir`, `store.Get`, `store.Put`, `strings.HasPrefix`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestArtifactJournalStaleControllerConflictIsNotIntegrityIncident`** — Regression contract: TestArtifactJournalStaleControllerConflictIsNotIntegrityIncident.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactTotal`, `cancel`, `context.WithCancel`, `db.seed`, `errors.Is`, `filepath.Join`, `j.Reserve`, `len`, `live.Err`, `newControllerArtifactJournal`, `os.IsNotExist`, `os.Stat`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestArtifactCustodyInternalOnlyAndNilGuard`** — Regression contract: TestArtifactCustodyInternalOnlyAndNilGuard.
  Signature: `func(t *testing.T)`.
  Calls: `fmt.Sprintf`, `mockRequest`, `mockServer`, `newControllerArtifactJournal`, `t.Fatal`, `t.Parallel`.
- **`TestArtifactJournalQuotaAcknowledgmentFailureCannotBeAdopted`** — Regression contract: TestArtifactJournalQuotaAcknowledgmentFailureCannotBeAdopted.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `fresh.AuthorizeRead`, `fresh.Reserve`, `fresh.Seal`, `len`, `newControllerArtifactJournal`, `os.ReadDir`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestArtifactJournalCurrentHumanReadPermissionRecheckedBeforeRelease`** — Regression contract: TestArtifactJournalCurrentHumanReadPermissionRecheckedBeforeRelease.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `context.WithValue`, `db.seed`, `db.snapshot`, `json.Unmarshal`, `len`, `newControllerArtifactJournal`, `store.Get`, `store.Put`, `string`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestArtifactJournalConcurrentControllersCannotReserveSameOutputSlot`** — Regression contract: TestArtifactJournalConcurrentControllersCannotReserveSameOutputSlot.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactTotal`, `close`, `errors.Is`, `fmt.Sprintf`, `j.Reserve`, `len`, `make`, `t.Fatal`, `t.Parallel`.
- **`TestArtifactJournalProjectionRejectsForgedTransitionsAndKeepsPrivateMetadata`** — Regression contract: TestArtifactJournalProjectionRejectsForgedTransitionsAndKeepsPrivateMetadata.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `artifactJournalFixture`, `artifactState`, `delete`, `j.Reserve`, `json.Marshal`, `newControllerArtifactJournal`, `projectArtifactCustody`, `storageRecord`, `string`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestArtifactHistoricalSourceBindingsDoNotReevaluateDeadline`** — Regression contract: TestArtifactHistoricalSourceBindingsDoNotReevaluateDeadline.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactSourceJob`, `artifactState`, `j.Reserve`, `job.AllocatedAt.Add`, `newControllerArtifactJournal`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `validateArtifactTransition`.
- **`TestArtifactJournalAuditFailureCancelsWorkAndSurvivesRestart`** — Regression contract: TestArtifactJournalAuditFailureCancelsWorkAndSurvivesRestart.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactStore`, `cancelModel`, `cancelWorkflow`, `checksum`, `context.WithCancel`, `db.snapshot`, `filepath.Join`, `fresh.agentsBlockedLocked`, `fresh.mu.Lock`, `fresh.mu.Unlock`, `fresh.restoreIntegrityStop`, `freshJournal.current`, `info.Mode`, `info.Mode().Perm`, `model.Err`, `newControllerArtifactJournal`, `os.Stat`, `s.checkIntegrityNow`, `store.Get`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workflow.Err`.
- **`TestArtifactCustodyFencedRetainedSourcePhase`** — Regression contract: TestArtifactCustodyFencedRetainedSourcePhase.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `db.seed`, `int64`, `len`, `newControllerArtifactJournal`, `store.Put`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestArtifactCustodyRetainedPhaseStillRequiresCurrentAuthority`** — Regression contract: TestArtifactCustodyRetainedPhaseStillRequiresCurrentAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactSourceJob`, `artifactState`, `artifactStore`, `db.seed`, `errors.New`, `newControllerArtifactJournal`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`.
