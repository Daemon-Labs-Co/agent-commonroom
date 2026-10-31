# cmd/server/governance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/governance_test.go)

Automated regression checks for governance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`governanceFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, enabled bool) (*server, *mockDB, http.Handler)`.
  Calls: `db.seed`, `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `t.Helper`.
- **`governanceRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, actorCookie, key, body string) *httptest.ResponseRecorder`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`.
- **`TestGovernanceTransferReplayAndRestart`** — Regression contract: TestGovernanceTransferReplayAndRestart.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `db.seed`, `db.snapshot`, `governanceFixture`, `governanceRequest`, `len`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.String`, `w.Header`, `w.Header().Get`.
- **`TestGovernanceRefusalsAndAtomicFailure`** — Regression contract: TestGovernanceRefusalsAndAtomicFailure.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `governanceFixture`, `governanceRequest`, `len`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestGovernanceFirstRegistrationAndNoAutomaticReplacement`** — Regression contract: TestGovernanceFirstRegistrationAndNoAutomaticReplacement.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `governanceFixture`, `json.Unmarshal`, `mockRequest`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `w.Body.Bytes`, `w.Body.String`.
- **`TestGovernanceConcurrentTransfers`** — Regression contract: TestGovernanceConcurrentTransfers.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `close`, `context.Background`, `governanceFixture`, `governanceRequest`, `make`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestGovernanceProjectionRefusesForgedAuthority`** — Regression contract: TestGovernanceProjectionRefusesForgedAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `governanceFixture`, `json.Marshal`, `projectGovernance`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestAutomaticAppointmentFailureLatchesStop`** — Regression contract: TestAutomaticAppointmentFailureLatchesStop.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `governanceFixture`, `governanceRequest`, `len`, `mockRequest`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAppointmentBatchesStopOnFailureButNotConflict`** — Regression contract: TestAppointmentBatchesStopOnFailureButNotConflict.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `context.WithValue`, `db.snapshot`, `governanceFixture`, `initialGovernanceChanges`, `len`, `s.appendAppointmentBatch`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAgentHandlersPersistInitialAppointmentAtomically`** — Regression contract: TestAgentHandlersPersistInitialAppointmentAtomically.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `fmt.Sprintf`, `governanceFixture`, `governanceRequest`, `json.RawMessage`, `json.Unmarshal`, `len`, `mockRequest`, `reflect.DeepEqual`, `s.load`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `w.Body.Bytes`, `w.Body.String`.
