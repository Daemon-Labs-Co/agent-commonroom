# cmd/server/worker_storage_check_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_storage_check_test.go)

Automated regression checks for worker_storage_check; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`storageCheckFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (string, workerStorageCheckPolicy)`.
  Calls: `filepath.Join`, `os.Mkdir`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`writeStorageCheckPolicy`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, path string, p workerStorageCheckPolicy)`.
  Calls: `json.Marshal`, `os.WriteFile`, `t.Fatal`, `t.Helper`.
- **`storageCheckObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(e projectquota.Expected) projectquota.Observation`.
- **`TestWorkerStorageCheckSuccessAndDescriptorOwnership`** — Regression contract: TestWorkerStorageCheckSuccessAndDescriptorOwnership.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `errors.Is`, `errout.Len`, `errout.String`, `f.Stat`, `info.IsDir`, `opened.Stat`, `out.String`, `policy.Roots[0].expected`, `runSingleWorkerStorageCheckForTest`, `storageCheckFixture`, `storageCheckObservation`, `strings.Contains`, `t.Fatal`, `t.Fatalf`, `writeStorageCheckPolicy`.
- **`TestWorkerStorageCheckPolicyRefusalBeforeInspection`** — Regression contract: TestWorkerStorageCheckPolicyRefusalBeforeInspection.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `change`, `context.Background`, `out.Len`, `runSingleWorkerStorageCheckForTest`, `storageCheckFixture`, `t.Fatal`, `t.Run`, `writeStorageCheckPolicy`.
- **`TestWorkerStorageCheckStrictPrivateJSON`** — Regression contract: TestWorkerStorageCheckStrictPrivateJSON.
  Signature: `func(t *testing.T)`.
  Calls: `loadWorkerStorageCheckPolicy`, `os.Chmod`, `os.WriteFile`, `storageCheckFixture`, `t.Fatal`, `writeStorageCheckPolicy`.
- **`TestWorkerStorageCheckRefusesWrongRootsAndUncertainEvidence`** — Regression contract: TestWorkerStorageCheckRefusesWrongRootsAndUncertainEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `errors.Is`, `errors.New`, `errout.String`, `f.Close`, `opened.Stat`, `os.Symlink`, `os.WriteFile`, `out.Len`, `runSingleWorkerStorageCheckForTest`, `storageCheckFixture`, `storageCheckObservation`, `strings.Contains`, `t.Fatal`, `t.Run`, `writeStorageCheckPolicy`.
- **`TestWorkerStorageCheckAllOrNothingAndUsage`** — Regression contract: TestWorkerStorageCheckAllOrNothingAndUsage.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `context.Background`, `errout.Reset`, `os.Mkdir`, `out.Len`, `out.Reset`, `runSingleWorkerStorageCheckForTest`, `storageCheckFixture`, `storageCheckObservation`, `t.Fatal`, `writeStorageCheckPolicy`.
- **`TestWorkerStorageCheckPreCanceledAndInvalidWholePolicy`** — Regression contract: TestWorkerStorageCheckPreCanceledAndInvalidWholePolicy.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `cancel`, `context.Background`, `context.WithCancel`, `len`, `runSingleWorkerStorageCheckForTest`, `storageCheckFixture`, `strings.Repeat`, `t.Fatal`, `uint32`, `uint64`, `writeStorageCheckPolicy`.
- **`runSingleWorkerStorageCheckForTest`** — Mock the complete native set query; all kernel interactions remain synthetic.
  Signature: `func(ctx context.Context, args []string, out, errout io.Writer, inspect func(context.Context, *os.File, projectquota.Expected) (projectquota.Observation, error)) int`.
  Calls: `append`, `inspect`, `len`, `make`, `runWorkerStorageCheck`.
- **`TestWorkerStorageCheckAggregateBudgetsAndIndependentCompanies`** — Regression contract: TestWorkerStorageCheckAggregateBudgetsAndIndependentCompanies.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `errors.Is`, `errout.String`, `f.Stat`, `len`, `os.Mkdir`, `out.Len`, `runWorkerStorageCheck`, `storageCheckFixture`, `storageCheckObservation`, `t.Context`, `t.Fatal`, `t.Run`, `uint64`, `writeStorageCheckPolicy`.
