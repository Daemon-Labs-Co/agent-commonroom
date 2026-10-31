# cmd/server/workspace_disposal_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_disposal_test.go)

Automated regression checks for workspace_disposal; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `disposalFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { s *server db *mockDB h http.Handler job room.ManagedJob view workerJobView now atomic.Int64 calls atomic.Int32 probes atomic.Int32 }`.
- **`newDisposalFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, enabled bool, bound ...bool) *disposalFixture`.
  Calls: `db.seed`, `f.calls.Add`, `f.now.Load`, `f.now.Store`, `f.probes.Add`, `http.NewServeMux`, `len`, `managedJobFence`, `managedJobFence(f.job).record`, `mockServer`, `mux.HandleFunc`, `quotaJob`, `s.actorGuard`, `s.auth.protect`, `s.load`, `storageRecord`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Helper`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Format`, `time.Now().UTC().UnixNano`, `time.Unix`, `time.Unix(0, f.now.Load()).UTC`, `validWorkerRecord`.
- **`(*disposalFixture).post`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, payload any) *httptest.ResponseRecorder`.
  Calls: `json.Marshal`, `mockRequest`, `string`, `t.Fatal`, `t.Helper`.
- **`(*disposalFixture).propose`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) room.WorkspaceDisposal`.
  Calls: `f.post`, `json.Unmarshal`, `res.Body.Bytes`, `res.Body.String`, `t.Fatal`, `t.Fatalf`, `t.Helper`.
- **`(*disposalFixture).approve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, p room.WorkspaceDisposal) *httptest.ResponseRecorder`.
  Calls: `f.post`.
- **`TestWorkspaceDisposalHTTP`** — Regression contract: TestWorkspaceDisposalHTTP.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `close`, `errors.New`, `f.approve`, `f.calls.Add`, `f.calls.Load`, `f.db.seed`, `f.db.snapshot`, `f.h.ServeHTTP`, `f.now.Add`, `f.now.Load`, `f.now.Store`, `f.post`, `f.probes.Load`, `f.propose`, `f.s.load`, `http.NewServeMux`, `httptest.NewRecorder`, `httptest.NewRequest`, `int64`, `len`, `make`, `managedStorageTotal`, `mockRequest`, `mux.HandleFunc`, `newDisposalFixture`, `other.actorGuard`, `other.auth.protect`, `p.CreatedAt.UnixNano`, `p.ExpiresAt.UnixNano`, `p.NotBefore.Add`, `p.NotBefore.Add(-time.Nanosecond).UnixNano`, `p.NotBefore.Sub`, `p.NotBefore.UnixNano`, `r.Body.String`, `req.AddCookie`, `strings.Contains`, `strings.HasPrefix`, `strings.Join`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Run`, `t.Setenv`, `time.Now`, `time.Now().Add`, `time.Unix`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestDisposalProjectionRefusesForgedTransitions`** — Regression contract: TestDisposalProjectionRefusesForgedTransitions.
  Signature: `func(t *testing.T)`.
  Calls: `disposalEvent`, `f.propose`, `f.s.load`, `json.Marshal`, `newDisposalFixture`, `p.NotBefore.Add`, `projectWorkspaceDisposal`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Run`, `t.Setenv`.
- **`TestPublicEventsCannotForgeDisposalApproval`** — Regression contract: TestPublicEventsCannotForgeDisposalApproval.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `len`, `mockRequest`, `mockServer`, `t.Fatal`.
- **`TestClaimedDisposalSurvivesDisconnectAndFlagDisable`** — Regression contract: TestClaimedDisposalSurvivesDisconnectAndFlagDisable.
  Signature: `func(t *testing.T)`.
  Calls: `action.Err`, `cancel`, `context.WithCancel`, `f.calls.Add`, `f.calls.Load`, `f.db.seed`, `f.h.ServeHTTP`, `f.now.Store`, `f.propose`, `f.s.load`, `httptest.NewRecorder`, `httptest.NewRequest`, `httptest.NewRequest("POST", "http://127.0.0.1:8787/api/workspace-disposal", strings.NewReader(body)).WithContext`, `managedStorageTotal`, `newDisposalFixture`, `p.NotBefore.UnixNano`, `req.AddCookie`, `req.Header.Set`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Setenv`.
- **`TestDisposalAdapterPanicBecomesDurableUncertainty`** — Regression contract: TestDisposalAdapterPanicBecomesDurableUncertainty.
  Signature: `func(t *testing.T)`.
  Calls: `f.approve`, `f.calls.Add`, `f.calls.Load`, `f.now.Store`, `f.propose`, `f.s.load`, `newDisposalFixture`, `p.NotBefore.UnixNano`, `panic`, `response.Body.String`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Setenv`.
- **`TestDisposalOutcomeAndSlotReleaseAreAtomic`** — Regression contract: TestDisposalOutcomeAndSlotReleaseAreAtomic.
  Signature: `func(t *testing.T)`.
  Calls: `f.approve`, `f.calls.Load`, `f.now.Store`, `f.propose`, `f.s.load`, `newDisposalFixture`, `p.NotBefore.UnixNano`, `t.Context`, `t.Fatal`, `t.Setenv`.
- **`TestDisposalFeatureSettingUsesOwnerEventPermissions`** — Regression contract: TestDisposalFeatureSettingUsesOwnerEventPermissions.
  Signature: `func(t *testing.T)`.
  Calls: `f.approve`, `f.calls.Load`, `f.db.seed`, `f.now.Store`, `f.propose`, `mockRequest`, `newDisposalFixture`, `p.NotBefore.UnixNano`, `r.Body.String`, `t.Fatal`, `t.Setenv`, `time.Now`, `time.Now().Add`.
- **`TestDisposalReviewRestartsConservativelyAfterControllerRestart`** — Regression contract: TestDisposalReviewRestartsConservativelyAfterControllerRestart.
  Signature: `func(t *testing.T)`.
  Calls: `f.calls.Load`, `f.now.Store`, `f.propose`, `http.NewServeMux`, `json.Unmarshal`, `len`, `mockRequest`, `mux.HandleFunc`, `newDisposalFixture`, `p.NotBefore.Add`, `p.NotBefore.UnixNano`, `ready.Equal`, `ready.UnixNano`, `response.Body.Bytes`, `response.Body.String`, `restarted.actorGuard`, `restarted.auth.protect`, `t.Fatal`, `t.Setenv`.
- **`TestDisposalReviewStartsAfterAuditPublication`** — Regression contract: TestDisposalReviewStartsAfterAuditPublication.
  Signature: `func(t *testing.T)`.
  Calls: `calls.Add`, `f.approve`, `f.calls.Load`, `f.now.Load`, `f.now.Store`, `f.post`, `initial.Add`, `initial.Add(60 * time.Second).UnixNano`, `json.Unmarshal`, `newDisposalFixture`, `response.Body.Bytes`, `response.Body.String`, `review.ReviewReadyAt.Equal`, `review.ReviewReadyAt.UnixNano`, `t.Fatal`, `t.Setenv`, `time.Unix`, `time.Unix(0, f.now.Load()).UTC`.
