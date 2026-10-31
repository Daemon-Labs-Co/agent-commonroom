# cmd/server/capture_profiles_check_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/capture_profiles_check_test.go)

Automated regression checks for capture_profiles_check; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `failedCaptureCheckWriter`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(failedCaptureCheckWriter).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func([]byte) (int, error)`.
  Calls: `errors.New`.
- **`TestCaptureProfileOfflineCheck`** — Regression contract: TestCaptureProfileOfflineCheck.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `captureProfileCheckMain`, `captureProfileFixture`, `captureSessionFixture`, `diagnostic.Len`, `diagnostic.String`, `os.WriteFile`, `out.String`, `strings.Contains`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `writeCaptureProfiles`.
