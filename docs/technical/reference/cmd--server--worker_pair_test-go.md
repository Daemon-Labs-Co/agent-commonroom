# cmd/server/worker_pair_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_pair_test.go)

Automated regression checks for worker_pair; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `pairFakeContainer`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ name, execution, repository, job string }`.
- **Type `pairFakeDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { containers map[string]pairFakeContainer removed []string failRemove string failInventory bool replaceName bool }`.
- **`(*pairFakeDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, _ []byte, out io.Writer) error`.
  Calls: `append`, `ctx.Err`, `delete`, `errors.New`, `io.WriteString`, `len`, `sort.Strings`, `strings.Join`, `strings.Repeat`, `validPairContainerID`.
- **`pairRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() workerRunRequest`.
  Calls: `sampleWorkerRequest`, `strings.Repeat`.
- **`pairFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (string, *workerPairSupervisor, *pairFakeDocker)`.
  Calls: `admitWorkerPair`, `pairContainerName`, `pairRequest`, `s.BeginCreate`, `s.BindCreated`, `strings.Repeat`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`pairFence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *workerPairSupervisor) workerFenceRequest`.
  Calls: `pairRequest`.
- **`fencePair`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, s *workerPairSupervisor, runtime *pairFakeDocker) (workerJobRecord, error)`.
  Calls: `fenceWorkerJobContainers`, `pairFence`, `stopWorkerPair`.
- **`TestWorkerPairCreationJournalBindsImmutableRoles`** — Regression contract: TestWorkerPairCreationJournalBindsImmutableRoles.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `filepath.Join`, `os.WriteFile`, `pairRequest`, `readWorkerRecord`, `s.BeginCreate`, `s.BindCreated`, `s.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestWorkerPairShutdownRequiresBothIDsAndReleasedLease`** — Regression contract: TestWorkerPairShutdownRequiresBothIDsAndReleasedLease.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `admitted.Close`, `cancel`, `context.WithTimeout`, `fencePair`, `len`, `pairFixture`, `pairRequest`, `readWorkerRecord`, `s.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerPairPartialShutdownRemainsUnresolved`** — Regression contract: TestWorkerPairPartialShutdownRemainsUnresolved.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `admitted.Close`, `fencePair`, `pairFixture`, `pairRequest`, `readWorkerRecord`, `s.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerPairCreationCrashRecoversOnlyVerifiedIdentity`** — Regression contract: TestWorkerPairCreationCrashRecoversOnlyVerifiedIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `fencePair`, `len`, `pairContainerName`, `pairRequest`, `s.BeginCreate`, `s.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestWorkerPairRefusesDaemonFailureAndNameReplacement`** — Regression contract: TestWorkerPairRefusesDaemonFailureAndNameReplacement.
  Signature: `func(t *testing.T)`.
  Calls: `fencePair`, `len`, `pairFixture`, `s.Close`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkerPairAdmissionReservesContainerSlotsAndRejectsOrphans`** — Regression contract: TestWorkerPairAdmissionReservesContainerSlotsAndRejectsOrphans.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `fakeWorkerRecord`, `pairContainerName`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `workerAdmission`.
- **`TestWorkerPairCannotForgeTerminalOrCrossRoleEvidence`** — Regression contract: TestWorkerPairCannotForgeTerminalOrCrossRoleEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `pairFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validWorkerRecord`, `workerJobSettled`.
- **`TestWorkerPairRuntimeEntryPointRemainsUnavailable`** — Regression contract: TestWorkerPairRuntimeEntryPointRemainsUnavailable.
  Signature: `func(t *testing.T)`.
  Calls: `errors.Is`, `pairRequest`, `t.Fatal`, `t.Parallel`, `validateWorkerRun`.
- **`TestWorkerPairUnknownJobContainerPreventsSettlement`** — Regression contract: TestWorkerPairUnknownJobContainerPreventsSettlement.
  Signature: `func(t *testing.T)`.
  Calls: `fencePair`, `len`, `pairFixture`, `s.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerPairAdmissionRefusesExpiredOrDuplicateJobs`** — Regression contract: TestWorkerPairAdmissionRefusesExpiredOrDuplicateJobs.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `other.Close`, `pairRequest`, `s.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `time.Now`, `time.Now().Add`.
- **`TestWorkerPairRecordCannotDowngradeToLegacyShutdown`** — Regression contract: TestWorkerPairRecordCannotDowngradeToLegacyShutdown.
  Signature: `func(t *testing.T)`.
  Calls: `pairFixture`, `t.Fatal`, `t.Parallel`, `validWorkerRecord`, `workerJobSettled`, `writeWorkerRecord`.
- **`TestWorkerPairRequiresExplicitCancellationProtocol`** — Regression contract: TestWorkerPairRequiresExplicitCancellationProtocol.
  Signature: `func(t *testing.T)`.
  Calls: `fencePair`, `fenceWorkerJobContainers`, `filepath.Join`, `os.IsNotExist`, `os.Stat`, `pairFence`, `pairFixture`, `request.record`, `s.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `validWorkerFenceAck`, `workerFenceAcknowledgement`.
- **`TestWorkerPairIncompleteCleanupCannotClaimSuccess`** — Regression contract: TestWorkerPairIncompleteCleanupCannotClaimSuccess.
  Signature: `func(t *testing.T)`.
  Calls: `fenceWorkerJobContainers`, `pairFence`, `pairFixture`, `readWorkerRecord`, `s.Close`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerPairJournalUsesNonDowngradableLifecycleVersion`** — Regression contract: TestWorkerPairJournalUsesNonDowngradableLifecycleVersion.
  Signature: `func(t *testing.T)`.
  Calls: `pairFixture`, `readWorkerRecord`, `t.Fatal`, `t.Parallel`, `validV2WorkerRecord`, `validWorkerRecord`.
- **`TestWorkerPairCancellationBeforeAdmissionTombstonesPairVersion`** — Regression contract: TestWorkerPairCancellationBeforeAdmissionTombstonesPairVersion.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `admitted.Close`, `fenceWorkerJobContainers`, `pairRequest`, `stopWorkerPair`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestWorkerPairCreationCannotOutliveAdmissionContext`** — Regression contract: TestWorkerPairCreationCannotOutliveAdmissionContext.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `cancel`, `context.Background`, `context.WithCancel`, `pairRequest`, `s.BeginCreate`, `s.Close`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
