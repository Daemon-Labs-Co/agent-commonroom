# cmd/server/worker_capture_lifetime_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_lifetime_test.go)

Automated regression checks for worker_capture_lifetime; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureLifetimeGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { workerCaptureJournalGuard lifetime context.Context }`.
- **`(*captureLifetimeGuard).CaptureAuthorityLifetime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() context.Context`.
- **`TestOwnedCaptureRequiresContinuousControlAuthority`** — Regression contract: TestOwnedCaptureRequiresContinuousControlAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `capture.Run`, `client.CaptureAuthorityLifetime`, `client.CaptureAuthorityLifetime().Done`, `client.StopCaptureAuthority`, `close`, `context.WithCancel`, `ctx.Done`, `len`, `lose`, `make`, `net.Pipe`, `newCaptureAuthorityClient`, `newOwnedWorkerCapture`, `ownedCaptureSetup`, `right.Close`, `serveCaptureAuthority`, `stopped.Add`, `stopped.Load`, `t.Cleanup`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `workerCaptureSlots`.
- **`TestRetainedCaptureRecoveryLosesIdleControlAuthority`** — Regression contract: TestRetainedCaptureRecoveryLosesIdleControlAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `client.CaptureAuthorityLifetime`, `client.CaptureAuthorityLifetime().Done`, `client.StopCaptureAuthority`, `close`, `ctx.Done`, `j.BindCapture`, `j.FinishCapture`, `make`, `net.Pipe`, `newCaptureAuthorityClient`, `ownedCaptureSetup`, `recoverRetainedWorkerCapture`, `right.Close`, `serveCaptureAuthority`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.After`, `workerCaptureSlots`.
