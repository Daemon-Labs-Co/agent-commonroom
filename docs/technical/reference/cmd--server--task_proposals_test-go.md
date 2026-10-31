# cmd/server/task_proposals_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_proposals_test.go)

Automated regression checks for task_proposals; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`proposalFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, enabled bool) (*server, *mockDB, http.Handler)`.
  Calls: `db.seed`, `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `t.Helper`.
- **`proposalRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, actor, key, body string) *httptest.ResponseRecorder`.
  Calls: `proposalRoomRequest`.
- **`proposalRoomRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, actor, rid, key, body string) *httptest.ResponseRecorder`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`.
- **`capturedProposal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler) room.TaskProposal`.
  Calls: `json.Unmarshal`, `proposalRequest`, `t.Fatal`, `t.Helper`, `w.Body.Bytes`, `w.Body.String`.
- **`approveProposalBody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.TaskProposal) string`.
  Calls: `fmt.Sprintf`.
- **`TestTaskProposalSeparatesCaptureApprovalAndExecution`** — Regression contract: TestTaskProposalSeparatesCaptureApprovalAndExecution.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `approveProposalBody`, `capturedProposal`, `db.snapshot`, `dispatchRequest`, `fmt.Sprintf`, `len`, `mockRequest`, `proposalFixture`, `proposalRequest`, `reflect.DeepEqual`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `w.Body.String`, `w.Header`, `w.Header().Get`.
- **`TestTaskProposalRefusalsAndScopedRead`** — Regression contract: TestTaskProposalRefusalsAndScopedRead.
  Signature: `func(t *testing.T)`.
  Calls: `capturedProposal`, `db.seed`, `db.snapshot`, `mockRequest`, `proposalFixture`, `proposalRequest`, `reflect.DeepEqual`, `strings.Contains`, `strings.Replace`, `strings.TrimSuffix`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestTaskProposalDecisionAtomicityRejectionAndFeatureOff`** — Regression contract: TestTaskProposalDecisionAtomicityRejectionAndFeatureOff.
  Signature: `func(t *testing.T)`.
  Calls: `approveProposalBody`, `capturedProposal`, `db.seed`, `db.snapshot`, `fmt.Sprintf`, `len`, `mockRequest`, `proposalFixture`, `proposalRequest`, `reflect.DeepEqual`, `s.load`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestTaskProposalConcurrentApprovalsCreateOneDraft`** — Regression contract: TestTaskProposalConcurrentApprovalsCreateOneDraft.
  Signature: `func(t *testing.T)`.
  Calls: `approveProposalBody`, `capturedProposal`, `close`, `context.Background`, `fmt.Sprintf`, `http.NewServeMux`, `len`, `make`, `mux.HandleFunc`, `other.actorGuard`, `other.auth.protect`, `proposalFixture`, `proposalRequest`, `s.load`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestTaskProposalCannotForgeOrStripProvenance`** — Regression contract: TestTaskProposalCannotForgeOrStripProvenance.
  Signature: `func(t *testing.T)`.
  Calls: `approveProposalBody`, `capturedProposal`, `db.seed`, `fmt.Sprintf`, `mockRequest`, `proposalFixture`, `proposalRequest`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestTaskProposalCaptureFailureAndApprovalDisableRace`** — Regression contract: TestTaskProposalCaptureFailureAndApprovalDisableRace.
  Signature: `func(t *testing.T)`.
  Calls: `approveProposalBody`, `capturedProposal`, `db.snapshot`, `http.NewServeMux`, `len`, `mockRequest`, `mux.HandleFunc`, `other.actorGuard`, `other.auth.protect`, `proposalFixture`, `proposalRequest`, `reflect.DeepEqual`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestTaskProposalFlagDefaultsOffAndOwnerCanEnable`** — Regression contract: TestTaskProposalFlagDefaultsOffAndOwnerCanEnable.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `h.ServeHTTP`, `http.NewServeMux`, `httptest.NewRecorder`, `httptest.NewRequest`, `mockRequest`, `mockServer`, `mux.HandleFunc`, `r.AddCookie`, `r.Header.Set`, `s.actorGuard`, `s.auth.protect`, `s.load`, `strings.Contains`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestProposalDraftCannotDispatchEvenAfterAssigneeEdit`** — Regression contract: TestProposalDraftCannotDispatchEvenAfterAssigneeEdit.
  Signature: `func(t *testing.T)`.
  Calls: `approveProposalBody`, `capturedProposal`, `db.seed`, `dispatchRequest`, `fmt.Errorf`, `fmt.Sprintf`, `len`, `mockRequest`, `proposalFixture`, `proposalRequest`, `s.load`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestPersonalRoomTaskProposalsReplayWithoutProfileOrHelpEvents`** — Regression contract: TestPersonalRoomTaskProposalsReplayWithoutProfileOrHelpEvents.
  Signature: `func(t *testing.T)`.
  Calls: `approveProposalBody`, `db.seed`, `fmt.Sprintf`, `hasRoom`, `json.Unmarshal`, `len`, `mockRequest`, `proposalFixture`, `proposalRoomRequest`, `restarted.load`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.Bytes`, `w.Body.String`.
- **`TestProposalReplayDoesNotInventUnknownOrFuturePersonalRooms`** — Regression contract: TestProposalReplayDoesNotInventUnknownOrFuturePersonalRooms.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `fmt.Sprint`, `proposalFixture`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestProposalReplayRefreshesPersonalRoomLabelsAfterLaterRenames`** — Regression contract: TestProposalReplayRefreshesPersonalRoomLabelsAfterLaterRenames.
  Signature: `func(t *testing.T)`.
  Calls: `controller.load`, `db.seed`, `json.Unmarshal`, `member`, `proposalFixture`, `proposalRoomRequest`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.Bytes`, `w.Body.String`.
- **`TestPersonalRoomCustomNamesSurviveProposalReplayAndRenames`** — Regression contract: TestPersonalRoomCustomNamesSurviveProposalReplayAndRenames.
  Signature: `func(t *testing.T)`.
  Calls: `capturedProposal`, `controller.load`, `db.seed`, `fmt.Sprintf`, `json.Marshal`, `json.Unmarshal`, `mockRequest`, `proposalFixture`, `s.load`, `saveRoom`, `string`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.Bytes`, `w.Body.String`.
- **`TestStalePersonalRoomSettingsCannotFreezeAutomaticNames`** — Regression contract: TestStalePersonalRoomSettingsCannotFreezeAutomaticNames.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: s.db}).load`, `capturedProposal`, `db.seed`, `db.snapshot`, `json.Marshal`, `json.Unmarshal`, `mockRequest`, `proposalFixture`, `readRoom`, `reflect.DeepEqual`, `rename`, `save`, `string`, `t.Context`, `t.Fatal`, `t.Helper`, `t.Parallel`, `t.Run`, `w.Body.Bytes`, `w.Body.String`.
