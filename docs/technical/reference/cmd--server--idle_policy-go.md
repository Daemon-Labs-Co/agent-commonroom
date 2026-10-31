# cmd/server/idle_policy.go

[Module index](cmd.md) · [Open source](../../../cmd/server/idle_policy.go)

Go implementation for idle policy; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`idlePolicyForState`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) (workerIdlePolicy, error)`.
  Calls: `errors.New`.
- **`prepareIdleSettings`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, req eventRequest, st room.State) (any, bool)`.
  Calls: `fenceObjectFields`, `http.Error`, `json.Unmarshal`, `len`.
- **`projectIdleSettings`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `errors.New`, `json.Unmarshal`.
- **`validManagedIdlePolicy`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(job room.ManagedJob) bool`.
  Calls: `pairedLifecycleVersion`.
- **`(*workerPairSupervisor).configureJobIdle`** — configureJobIdle consumes an immutable, durably admitted controller snapshot. The reviewed authenticated assembly must pass its exact registered job, never room/model-selected data. This method grants no execution/transport authority. Omitting it in a source fork loses controller-policy enforcement; legacy local assembly retains the worker deployment default until production dispatch lands.
  Signature: `func(job room.ManagedJob, runner toolbroker.DockerRunner, announce ...workerIdleAnnounce) error`.
  Calls: `errors.New`, `len`, `s.mu.Lock`, `s.mu.Unlock`, `sameWorkerBinding`, `validManagedJob`.
