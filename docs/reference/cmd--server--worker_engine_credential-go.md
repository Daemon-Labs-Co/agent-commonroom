# cmd/server/worker_engine_credential.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_engine_credential.go)

Go implementation for worker engine credential; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerEngineCredential`**: workerEngineCredential is worker-only custody, not a model tool or controller payload. It serializes engine use with the existing sign-in/sign-out lock. Cancellation does NOT unlock: a still-running container may retain credentials. The supervisor must fence both roles before releasing. A source fork can leak these bytes or forge evidence; the worker account remains a trusted boundary. No host credential mount is used. Refreshed-token persistence remains separate.
  Declaration: `struct { mu sync.Mutex scope toolbroker.Scope original context.Context raw []byte unlock func() released bool provisioned bool }`.
- **`acquireWorkerEngineCredential`** — acquireWorkerEngineCredential takes a bounded private snapshot while holding the connection lock. authRoot is trusted worker configuration, never a request path. The configured root and its ancestry must be secured by deployment.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, authRoot, journalDir string) (*workerEngineCredential, error)`.
  Calls: `clear`, `connectionAuthKey`, `ctx.Err`, `ctxJob`, `filepath.Clean`, `filepath.IsAbs`, `filepath.Join`, `lockWorkerCredential`, `providerCredentialName`, `readPrivateBounded`, `scope.Deadline.UTC`, `scope.Deadline.UTC().Round`, `string`, `time.Now`, `toolbroker.ValidAuditScope`, `unlock`, `validEngineSubscriptionCredential`, `workerAliasPattern.MatchString`, `workerCredentialJournalReady`.
- **`workerCredentialJournalReady`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string, scope toolbroker.Scope) bool`.
  Calls: `boundedWorkerRecords`, `pairRoleSettled`, `validWorkerAuditScope`, `validWorkerRecord`, `workerJobSettled`.
- **`validEngineSubscriptionCredential`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) bool`.
  Calls: `engineCredentialFields`, `json.Unmarshal`, `len`, `strictjson.ValidEncoding`, `string`.
- **`(*workerEngineCredential).Provision`** — Provision sends one bounded private authentication frame before the callback protocol begins. Only pass the inspected engine attachment, never a model, audit, controller or log writer. Even partial failure consumes this attempt.
  Signature: `func(ctx context.Context, out io.Writer) error`.
  Calls: `binary.BigEndian.PutUint32`, `c.mu.Lock`, `c.mu.Unlock`, `c.original.Err`, `ctxJob`, `len`, `make`, `out.Write`, `time.Now`, `toolbroker.ValidAuditScope`, `uint32`.
- **`(*workerEngineCredential).Release`** — Release requires durable pair-fencing evidence for the original identity. Retained source volumes do not require retaining provider credentials once both execution roles are confirmed absent (or never began creation). No timer, cancelled context or process exit can substitute this evidence.
  Signature: `func(record workerJobRecord) error`.
  Calls: `c.mu.Lock`, `c.mu.Unlock`, `c.unlock`, `clear`, `matchesWorkerCredentialScope`, `pairRoleSettled`, `validWorkerRecord`.
- **`matchesWorkerCredentialScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(record workerJobRecord, scope toolbroker.Scope) bool`.
  Calls: `record.Deadline.Equal`.
- **`engineCredentialFields`** — Exact top-level and token objects reject duplicate and case-aliased fields. Unknown schema fields require review before a newer client can be admitted.
  Signature: `func(raw []byte, allowed []string) (map[string]json.RawMessage, bool)`.
  Calls: `bytes.NewReader`, `d.Decode`, `d.More`, `d.Token`, `json.Delim`, `json.NewDecoder`, `slices.Contains`.
