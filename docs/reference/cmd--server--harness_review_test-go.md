# cmd/server/harness_review_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/harness_review_test.go)

Automated regression checks for harness_review; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`harnessFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (harnessReview, ed25519.PublicKey, ed25519.PrivateKey, time.Time)`.
  Calls: `ed25519.GenerateKey`, `now.Add`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `time.Now`, `time.Now().UTC`.
- **`signHarness`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, review harnessReview, key ed25519.PrivateKey) []byte`.
  Calls: `base64.StdEncoding.EncodeToString`, `ed25519.Sign`, `json.Marshal`, `t.Fatal`, `t.Helper`.
- **`TestHarnessReviewSelectionAndStartup`** — Regression contract: TestHarnessReviewSelectionAndStartup.
  Signature: `func(t *testing.T)`.
  Calls: `harnessFixture`, `now.Add`, `quotaJob`, `selectHarnessReview`, `selectedHarnessImage`, `signHarness`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validManagedJob`.
- **`TestHarnessReviewRejectsUnsignedIncompatibleAndAmbiguousInput`** — Regression contract: TestHarnessReviewRejectsUnsignedIncompatibleAndAmbiguousInput.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `harnessFixture`, `make`, `now.Add`, `signHarness`, `string`, `strings.Replace`, `t.Fatal`, `t.Parallel`, `t.Run`, `verifyHarnessReview`.
- **`TestHarnessReviewFlagAvailabilityAndRollback`** — Regression contract: TestHarnessReviewFlagAvailabilityAndRollback.
  Signature: `func(t *testing.T)`.
  Calls: `errors.New`, `harnessFixture`, `quotaJob`, `selectHarnessReview`, `selectedHarnessImage`, `signHarness`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestHarnessReviewBindingCannotChangeOnReplay`** — Regression contract: TestHarnessReviewBindingCannotChangeOnReplay.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `harnessFixture`, `job.UpdatedAt.Add`, `managedPairFixtureVersion`, `s.load`, `signHarness`, `t.Context`, `t.Fatal`, `t.Parallel`, `verifyHarnessReview`.
- **`TestHarnessReviewRegistrationUsesDeploymentPolicy`** — Environment configuration is serialized by testing.Setenv; all other review tests use explicit inputs and can run in parallel without deployment mutation.
  Signature: `func(t *testing.T)`.
  Calls: `base64.StdEncoding.EncodeToString`, `context.WithValue`, `ctx.Value`, `db.seed`, `filepath.Join`, `harnessFixture`, `len`, `managedPairFixtureVersion`, `mockServer`, `os.WriteFile`, `registration.Register`, `s.load`, `s.managedExecutionJobContextVersion`, `signHarness`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Run`, `t.Setenv`, `t.TempDir`.
