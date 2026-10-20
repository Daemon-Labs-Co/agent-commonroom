# cmd/server/p0_harness_acceptance_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/p0_harness_acceptance_test.go)

Automated regression checks for p0_harness_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0HarnessReviewRefusal`** — These negative cases exercise the real allocation gate with mocked audit storage. They establish rejection before launch authority exists, not live image compatibility or successful provider execution.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `base64.StdEncoding.EncodeToString`, `context.WithValue`, `ctx.Value`, `db.seed`, `db.snapshot`, `ed25519.GenerateKey`, `errors.Is`, `filepath.Join`, `harnessFixture`, `len`, `mockServer`, `os.WriteFile`, `registration.Register`, `s.managedExecutionJobContextVersion`, `signHarness`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `t.Setenv`, `t.TempDir`.
- **Type `harnessDigestDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls [][]string image string }`.
- **`(*harnessDigestDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, _ []byte, out io.Writer) error`.
  Calls: `append`, `ctx.Err`, `errors.New`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `strings.Repeat`, `strings.Split`.
- **`driveP0HarnessDigestRefusal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `errors.Is`, `harnessFixture`, `now.Add`, `runtime.Close`, `runtime.Create`, `selectHarnessReview`, `signHarness`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `toolbroker.NewEngineContainerRuntime`.
