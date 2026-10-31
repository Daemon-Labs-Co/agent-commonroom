# cmd/server/room_outcome_review.go

[Module index](cmd.md) · [Open source](../../../cmd/server/room_outcome_review.go)

Go implementation for room outcome review; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`roomReviewSnapshot`** — Human acceptance is a review record, never a grant to dispatch, dispose of data or skip worker cleanup. Its snapshot becomes stale when room work changes.
  Signature: `func(st room.State, rid string) (string, bool)`.
  Calls: `append`, `enc.Encode`, `h.Sum`, `hex.EncodeToString`, `json.NewEncoder`, `len`, `sha256.New`, `sort.Slice`, `strings.TrimSpace`, `taskHasCurrentAcceptance`, `taskHasValidCurrentOutput`.
- **`roomReviewStatus`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(current room.RoomOutcomeReview, snapshot string) string`.
- **`validRoomReviewTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v room.RoomOutcomeReview) bool`.
  Calls: `roomReviewSnapshot`, `roomReviewStatus`, `strings.TrimSpace`, `utf8.RuneCountInString`, `v.ReviewedAt.IsZero`.
- **`projectRoomOutcomeReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, data []byte) error`.
  Calls: `errors.New`, `json.Unmarshal`, `validRoomReviewTransition`.
- **`(*server).roomOutcomeReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `canAccess`, `errors.Is`, `fenceObjectFields`, `hasRoom`, `http.Error`, `http.MaxBytesReader`, `int`, `io.ReadAll`, `json.Unmarshal`, `len`, `operationError`, `prepareOperation`, `r.Context`, `r.Header.Get`, `reply`, `requestActor`, `requestedRoom`, `roomReviewSnapshot`, `roomReviewStatus`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.validCSRF`, `s.workflowSafety`, `strings.TrimSpace`, `time.Now`, `time.Now().UTC`, `uint`, `utf8.RuneCountInString`, `validRoomReviewTransition`, `withOperationResponse`.
