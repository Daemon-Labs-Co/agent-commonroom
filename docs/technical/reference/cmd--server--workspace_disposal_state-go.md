# cmd/server/workspace_disposal_state.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_disposal_state.go)

Go implementation for workspace disposal state; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`disposalIdentity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.WorkspaceDisposal) toolbroker.WorkspaceVolumeIdentity`.
- **`(*server).disposalTime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() time.Time`.
  Calls: `s.disposalNow`, `s.disposalNow().UTC`, `s.disposalNow().UTC().Round`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Round`.
- **`sameDisposalJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a, b room.ManagedJob) bool`.
- **`validDisposal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.WorkspaceDisposal) bool`.
  Calls: `disposalIdentity`, `p.ClaimedAt.Before`, `p.ClaimedAt.IsZero`, `p.CreatedAt.Add`, `p.CreatedAt.IsZero`, `p.ExpiresAt.Equal`, `p.FinishedAt.Before`, `p.FinishedAt.IsZero`, `p.NotBefore.Equal`, `p.ReviewObservedAt.Add`, `p.ReviewObservedAt.Before`, `p.ReviewObservedAt.IsZero`, `toolbroker.ValidWorkspaceVolume`, `validManagedJob`, `validWorkerJobID`.
- **`projectWorkspaceDisposal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, kind string, raw []byte) error`.
  Calls: `json.Unmarshal`, `p.CreatedAt.Before`, `reservationMatchesDisposal`, `validDisposal`, `validWorkerJobID`.
- **`disposalFence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.WorkspaceDisposal) workerFenceRequest`.
  Calls: `disposalIdentity`, `managedJobFence`.
- **`disposalEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.WorkspaceDisposal) string`.
  Calls: `strings.TrimSpace`.
- **Type `disposalReviewView`**: These views are assembled under s.mu. A controller waits a minute from its own first observation of COMMITTED state, not an in-transaction timestamp. Restart or another controller conservatively restarts that local review barrier.
  Declaration: `struct { room.WorkspaceDisposal ReviewReadyAt time.Time 'json:"reviewReadyAt"' }`.
- **`(*server).disposalReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.WorkspaceDisposal) disposalReviewView`.
  Calls: `delete`, `observed.After`, `s.disposalTime`, `seen.Add`, `seen.Before`.
- **`(*server).disposalReviews`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) []disposalReviewView`.
  Calls: `append`, `disposalList`, `s.disposalReview`.
