# cmd/tool-helper/main_test.go

[Module index](cmd.md) · [Open source](../../cmd/tool-helper/main_test.go)

Automated regression checks for main; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`testRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*os.Root, string)`.
  Calls: `os.OpenRoot`, `root.Close`, `t.Cleanup`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`wire`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, action toolbroker.Action) *bytes.Reader`.
  Calls: `bytes.NewReader`, `json.Marshal`, `t.Fatal`, `t.Helper`.
- **`TestFileRoundTripMatchesBrokerWire`** — Regression contract: TestFileRoundTripMatchesBrokerWire.
  Signature: `func(t *testing.T)`.
  Calls: `base64.StdEncoding.DecodeString`, `context.Background`, `filepath.Join`, `hex.EncodeToString`, `json.Unmarshal`, `len`, `os.ReadFile`, `output.Bytes`, `output.Reset`, `run`, `sha256.Sum256`, `string`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `testRoot`, `wire`.
- **`TestProtocolRefusesMalformedOrUndeclaredActions`** — Regression contract: TestProtocolRefusesMalformedOrUndeclaredActions.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `decodeAction`, `errors.Is`, `len`, `os.ReadDir`, `output.Len`, `run`, `strings.NewReader`, `strings.Repeat`, `strings.Replace`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `testRoot`.
- **`TestReadBinaryBudgetAndOutsideLink`** — Regression contract: TestReadBinaryBudgetAndOutsideLink.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Repeat`, `context.Background`, `errors.Is`, `filepath.Join`, `os.ReadFile`, `os.Symlink`, `os.WriteFile`, `output.Len`, `output.Reset`, `run`, `string`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `testRoot`, `wire`.
- **Type `countingReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ count int }`.
- **`(*countingReader).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `len`.
- **`TestInputReadStopsAtBudget`** — Regression contract: TestInputReadStopsAtBudget.
  Signature: `func(t *testing.T)`.
  Calls: `decodeAction`, `errors.Is`, `t.Fatal`, `t.Parallel`.
- **`TestOversizeFileRefusedWithoutOutput`** — Regression contract: TestOversizeFileRefusedWithoutOutput.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Repeat`, `context.Background`, `errors.Is`, `filepath.Join`, `os.WriteFile`, `output.Len`, `run`, `t.Fatal`, `t.Parallel`, `testRoot`, `wire`.
- **Type `failedWriter`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(failedWriter).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func([]byte) (int, error)`.
  Calls: `errors.New`.
- **`TestWriteOutputFailureIsUnconfirmed`** — Regression contract: TestWriteOutputFailureIsUnconfirmed.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `errors.Is`, `filepath.Join`, `os.ReadFile`, `run`, `string`, `t.Fatal`, `t.Parallel`, `testRoot`, `wire`.
- **`TestCancelledOrMissingDependenciesRefused`** — Regression contract: TestCancelledOrMissingDependenciesRefused.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `errors.Is`, `run`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `testRoot`, `wire`.
- **`TestRepositoryRootRefusesSymlinkOrMissingDirectory`** — Regression contract: TestRepositoryRootRefusesSymlinkOrMissingDirectory.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `os.Mkdir`, `os.Remove`, `os.Symlink`, `repositoryRoot`, `root.Close`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `testRoot`.
