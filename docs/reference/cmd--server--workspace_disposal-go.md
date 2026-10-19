# cmd/server/workspace_disposal.go

[Module index](cmd.md) · [Open source](../../cmd/server/workspace_disposal.go)

Go implementation for workspace disposal; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`disposalOwner`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, actor string) bool`.
  Calls: `member`.
- **`disposalList`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) []room.WorkspaceDisposal`.
  Calls: `append`, `len`, `list[i].CreatedAt.After`, `list[i].CreatedAt.Equal`, `sort.Slice`, `unresolved`.
- **`(*server).workspaceDisposal`** — Reviewed workspace disposal is off by default, owner-only and one-shot. Worker operations require a committed claim. No automatic retry, host fallback or storage refund exists. A crash leaves a durable unresolved claim for review.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `callDisposalWorker`, `cancel`, `cancelAction`, `context.WithTimeout`, `context.WithValue`, `context.WithoutCancel`, `disposalEvent`, `disposalFence`, `disposalIdentity`, `disposalOwner`, `fenceObjectFields`, `http.Error`, `http.MaxBytesReader`, `id`, `io.ReadAll`, `json.Unmarshal`, `now.Add`, `now.Before`, `os.Getenv`, `p.CreatedAt.Add`, `pairRoleSettled`, `r.Context`, `reply`, `requestActor`, `s.appendBatch`, `s.approvedSnapshot`, `s.disposalReview`, `s.disposalReviews`, `s.disposalTime`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.persistDisposal`, `s.validCSRF`, `sameDisposalJob`, `sameWorkerBinding`, `validDisposal`, `validManagedJob`, `validWorkerRecord`.
- **`(*server).persistDisposal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, p room.WorkspaceDisposal) error`.
  Calls: `context.WithValue`, `disposalEvent`, `json.Marshal`, `projectWorkspaceDisposal`, `s.append`.
- **`callDisposalWorker`** — A trusted adapter panic is still an unknown external outcome. Persist generic uncertainty rather than log a raw panic or make the claim replayable.
  Signature: `func(ctx context.Context, alias string, request workerFenceRequest, fence func(context.Context, string, workerFenceRequest) error) (err error)`.
  Calls: `errors.New`, `fence`, `recover`.
