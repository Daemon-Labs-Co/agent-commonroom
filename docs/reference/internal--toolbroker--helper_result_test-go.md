# internal/toolbroker/helper_result_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/helper_result_test.go)

Automated regression checks for helper_result; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`syntheticReadResult`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, content string) []byte`.
  Calls: `base64.StdEncoding.EncodeToString`, `hex.EncodeToString`, `json.Marshal`, `len`, `sha256.Sum256`, `t.Fatal`, `t.Helper`.
- **`TestHelperResultRefusesMalformedAndMismatchedEvidence`** — Regression contract: TestHelperResultRefusesMalformedAndMismatchedEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `json.Marshal`, `json.Unmarshal`, `string`, `strings.Repeat`, `strings.Replace`, `syntheticReadResult`, `t.Fatal`, `t.Parallel`, `validHelperResult`.
- **`TestUnsupportedDeclaredToolRefusedBeforeDocker`** — Regression contract: TestUnsupportedDeclaredToolRefusedBeforeDocker.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `NewToolContainerPlan`, `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `liveFixture`, `mockDocker`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.Now`.
- **`TestWriteResultBindsOriginalContent`** — Regression contract: TestWriteResultBindsOriginalContent.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Replace`, `hex.EncodeToString`, `json.Marshal`, `len`, `sha256.Sum256`, `t.Fatal`, `t.Parallel`, `validHelperResult`.
- **`TestInvalidHelperResultFencesWithoutPublishing`** — Regression contract: TestInvalidHelperResultFencesWithoutPublishing.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `executorFixture`, `io.WriteString`, `out.Len`, `p.RunArgs`, `runtimeMock`, `t.Fatal`, `t.Parallel`, `time.Now`.
