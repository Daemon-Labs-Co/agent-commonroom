# cmd/server/worker_lifecycle_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_lifecycle_test.go)

Automated regression checks for worker_lifecycle; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`fakeWorkerRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, job, key string) workerJobRecord`.
  Calls: `newWorkerJobRecord`, `t.Fatal`, `t.Helper`.
- **`TestWorkerLeaseChild`** — Regression contract: TestWorkerLeaseChild.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerJob`, `context.Background`, `fmt.Println`, `leases.Close`, `os.Getenv`, `sampleWorkerRequest`, `t.Fatal`, `time.After`.
- **`TestSupervisorCrashLeavesDurableJobAndBlocksReplacementUntilReconciled`** — Regression contract: TestSupervisorCrashLeavesDurableJobAndBlocksReplacementUntilReconciled.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerJob`, `append`, `bufio.NewScanner`, `cmd.Process.Kill`, `cmd.Start`, `cmd.StdoutPipe`, `cmd.Wait`, `context.Background`, `exec.Command`, `fenceWorkerJob`, `lease.Close`, `make`, `next.Close`, `os.Environ`, `sampleWorkerRequest`, `scanner.Scan`, `scanner.Text`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `time.After`, `workerTryLease`.
- **`TestWorkerCancellationBeforeLaunchTombstonesIdentity`** — Regression contract: TestWorkerCancellationBeforeLaunchTombstonesIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerJob`, `context.Background`, `fenceWorkerJob`, `leases.Close`, `sampleWorkerRequest`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestContainerAbsenceDoesNotProveSupervisorStopped`** — Regression contract: TestContainerAbsenceDoesNotProveSupervisorStopped.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerJob`, `cancel`, `context.Background`, `context.WithTimeout`, `errors.New`, `fenceWorkerJob`, `leases.Close`, `readWorkerRecord`, `sampleWorkerRequest`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestWorkerScopeLeaseBindsMembershipAcrossConnections`** — Regression contract: TestWorkerScopeLeaseBindsMembershipAcrossConnections.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerJob`, `context.Background`, `duplicate.Close`, `leases.Close`, `sampleWorkerRequest`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `workerSnapshot`.
- **`TestWorkerAdmissionGlobalRoomAndUnknownStateLimits`** — Regression contract: TestWorkerAdmissionGlobalRoomAndUnknownStateLimits.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `fakeWorkerRecord`, `fmt.Sprintf`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `workerAdmission`.
- **`TestWorkerFencingRejectsWrongBindingAndUnsafeFiles`** — Regression contract: TestWorkerFencingRejectsWrongBindingAndUnsafeFiles.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `fakeWorkerRecord`, `fenceWorkerJob`, `filepath.Join`, `lease.Close`, `os.Chmod`, `os.IsNotExist`, `os.Stat`, `os.Symlink`, `readWorkerRecord`, `sampleWorkerRequest`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `workerTryLease`, `writeWorkerRecord`.
- **`TestWorkerProtocolRejectsExpiredReplayAndExtendedDeadline`** — Regression contract: TestWorkerProtocolRejectsExpiredReplayAndExtendedDeadline.
  Signature: `func(t *testing.T)`.
  Calls: `encodeWorkerRequest`, `request.IssuedAt.Add`, `sampleWorkerRequest`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestWorkerContainerReservationsIncludeUncreatedAndUnresolvedPairs`** — Regression contract: TestWorkerContainerReservationsIncludeUncreatedAndUnresolvedPairs.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `fakeWorkerRecord`, `fmt.Sprintf`, `strings.Repeat`, `t.Fatalf`, `t.Parallel`, `t.Run`, `workerAdmission`.
