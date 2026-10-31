# cmd/server/connection_verification.go

[Module index](cmd.md) · [Open source](../../../cmd/server/connection_verification.go)

Go implementation for connection verification; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `verificationBindingKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **Type `verificationBinding`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Connection room.Connection Participant room.Participant }`.
- **`verificationParticipant`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(rid, cid string) room.Participant`.
  Calls: `hex.EncodeToString`, `sha256.Sum256`.
- **`(*server).managedVerification`** — managedVerification verifies a fixed reply through the journalled protocol-v2 worker. It cannot export room context, resume a session or mount a repository. Source forks CAN bypass registration, BUT lose stop/restart ownership guarantees.
  Signature: `func(ctx context.Context, c room.Connection, rid string) ([]byte, error)`.
  Calls: `adapter.Run`, `bound.Value`, `cancel`, `connectionHasActiveAgent`, `connectionTestingEnabled`, `context.Background`, `context.WithCancel`, `context.WithTimeout`, `context.WithValue`, `deadlineCancel`, `delete`, `errors.New`, `hasRoom`, `json.Marshal`, `os.Getenv`, `remoteWorkerEndpoint`, `run`, `s.agentAdmission`, `s.load`, `s.managedJobContext`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.workflowSafety`, `strings.TrimSpace`, `turnCtx.Err`, `verificationParticipant`.
