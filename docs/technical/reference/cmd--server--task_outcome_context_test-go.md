# cmd/server/task_outcome_context_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_outcome_context_test.go)

Automated regression checks for task_outcome_context; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestTaskDeliverableContextDisclosure`** — Regression contract: TestTaskDeliverableContextDisclosure.
  Signature: `func(t *testing.T)`.
  Calls: `chatPrompt`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskDispatchPrompt`.
- **`TestTaskDeliverableDispatchWorkAndReview`** — Regression contract: TestTaskDeliverableDispatchWorkAndReview.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `ctx.Done`, `ctx.Err`, `db.seed`, `dispatchFixture`, `dispatchRequest`, `fmt.Sprintf`, `make`, `replay.Header`, `replay.Header().Get`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `w.Body.String`.
- **`TestTaskDeliverableBoundsAndNoEffectsOnRefusal`** — Regression contract: TestTaskDeliverableBoundsAndNoEffectsOnRefusal.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `dispatchFixture`, `dispatchRequest`, `json.Marshal`, `len`, `mockRequest`, `s.load`, `string`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskDispatchPrompt`, `time.Now`, `time.Now().Add`, `w.Body.String`.
