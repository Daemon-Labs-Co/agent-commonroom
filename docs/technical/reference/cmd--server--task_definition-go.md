# cmd/server/task_definition.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_definition.go)

Go implementation for task definition; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`taskDefinitionDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t room.Task) string`.
  Calls: `taskcontract.Digest`.
- **`prepareTaskDefinition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request, st room.State, req eventRequest, v *room.Task) bool`.
  Calls: `bytes.Equal`, `bytes.TrimSpace`, `err.Error`, `http.Error`, `int`, `json.Unmarshal`, `requestActor`, `taskHasValidCurrentOutput`, `taskcontract.CriteriaText`, `taskcontract.Decode`, `taskcontract.Same`, `uint`.
- **`taskDefinitionContextAllowed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, p room.Participant) bool`.
  Calls: `ctx.Value`, `taskDefinitionDigest`, `taskExecutionPendingCurrent`.
- **`validateTaskDefinitionProjection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, t room.Task, raw []byte) error`.
  Calls: `errors.New`, `json.Unmarshal`, `taskHasCurrentAcceptance`, `taskHasValidCurrentOutput`, `taskcontract.CriteriaText`, `taskcontract.Same`, `taskcontract.Validate`.
