# cmd/server/connection_verification_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/connection_verification_test.go)

Automated regression checks for connection_verification; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`verificationFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, room.Connection)`.
  Calls: `db.seed`, `mockServer`, `t.Helper`.
- **`journalVerification`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, status string) error`.
  Calls: `ctx.Value`, `r.Register`, `r.Update`.
- **`TestManagedVerificationRequiresIntentAndShutdownProof`** — Regression contract: TestManagedVerificationRequiresIntentAndShutdownProof.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `db.seed`, `errors.New`, `journalVerification`, `json.Unmarshal`, `len`, `rune`, `s.managedVerification`, `string`, `t.Fatal`, `t.Run`, `t.Setenv`, `verificationFixture`, `verificationParticipant`.
- **`TestVerificationCanBeCancelledAndBlocksConnectionChanges`** — Regression contract: TestVerificationCanBeCancelledAndBlocksConnectionChanges.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `close`, `connectionHasActiveAgent`, `context.Background`, `context.WithTimeout`, `ctx.Done`, `ctx.Err`, `ctx.Value`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Marshal`, `make`, `r.Register`, `r.Update`, `s.load`, `s.managedVerification`, `s.mu.Lock`, `s.mu.Unlock`, `s.prepareAgentsEvent`, `stop`, `t.Fatal`, `t.Setenv`, `time.After`, `verificationFixture`.
- **`TestLegacyVerificationLauncherCannotRun`** — Regression contract: TestLegacyVerificationLauncherCannotRun.
  Signature: `func(t *testing.T)`.
  Calls: `out.Len`, `t.Fatal`, `t.Parallel`, `workerVerifyConnection`.
- **`TestConnectionTestHTTPUsesJournalledVerification`** — Regression contract: TestConnectionTestHTTPUsesJournalledVerification.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `http.NewServeMux`, `journalVerification`, `json.Unmarshal`, `len`, `mockRequest`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `s.load`, `t.Fatal`, `t.Setenv`, `verificationFixture`, `w.Body.Bytes`, `w.Body.String`.
- **`TestVerificationReservationsConsumeRoomCapacity`** — Regression contract: TestVerificationReservationsConsumeRoomCapacity.
  Signature: `func(t *testing.T)`.
  Calls: `rune`, `s.agentAdmission`, `string`, `t.Fatal`, `t.Parallel`, `verificationFixture`.
