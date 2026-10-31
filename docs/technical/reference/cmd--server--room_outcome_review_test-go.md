# cmd/server/room_outcome_review_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/room_outcome_review_test.go)

Automated regression checks for room_outcome_review; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`reviewFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, skipTasks ...bool) (*server, *mockDB, http.Handler)`.
  Calls: `db.seed`, `http.NewServeMux`, `len`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `t.Helper`.
- **`readRoomReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler) map[string]json.RawMessage`.
  Calls: `json.Unmarshal`, `mockRequest`, `t.Fatal`, `t.Helper`, `w.Body.Bytes`, `w.Body.String`.
- **`roomReviewBody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler, action string) string`.
  Calls: `json.Marshal`, `json.Unmarshal`, `readRoomReview`, `string`, `t.Helper`.
- **`submitRoomReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, key, body string) *httptest.ResponseRecorder`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`.
- **`TestRoomOutcomeAcceptanceReplayReopenAndRestart`** — Regression contract: TestRoomOutcomeAcceptanceReplayReopenAndRestart.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `context.Background`, `db.seed`, `db.snapshot`, `len`, `readRoomReview`, `reviewFixture`, `roomReviewBody`, `string`, `submitRoomReview`, `t.Fatal`, `t.Parallel`, `w.Body.String`, `w.Header`, `w.Header().Get`.
- **`TestRoomReviewRefusalsAndSnapshotChanges`** — Regression contract: TestRoomReviewRefusalsAndSnapshotChanges.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `len`, `reviewFixture`, `roomReviewBody`, `submitRoomReview`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestRoomReviewHistoryOffAndEventForgery`** — Regression contract: TestRoomReviewHistoryOffAndEventForgery.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `db.seed`, `mockRequest`, `readRoomReview`, `reviewFixture`, `roomReviewBody`, `string`, `submitRoomReview`, `t.Context`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestRoomReviewConfigurationAndMalformedRequests`** — Regression contract: TestRoomReviewConfigurationAndMalformedRequests.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `delete`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Marshal`, `json.Unmarshal`, `len`, `mockRequest`, `r.AddCookie`, `response.Body.String`, `reviewFixture`, `roomReviewBody`, `string`, `strings.NewReader`, `strings.Repeat`, `submitRoomReview`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `w.Body.String`.
