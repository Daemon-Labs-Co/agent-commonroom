# cmd/server/worker_audit_transport.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_audit_transport.go)

Go implementation for worker audit transport; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerAuditListener`**: The audit socket is supervisor-only: never mount its directory into either container. SSH authenticates the approved worker account; that account and its Docker daemon remain trusted. This is not protection from a hostile host user with the same UID. A source fork CAN expose this socket, BUT loses audit trust. There is exactly one accepted link per supervisor lifetime; reconnect needs a new job and reconciliation, never replay of an uncertain audit outcome.
  Declaration: `struct { listener *net.UnixListener parent context.Context scope toolbroker.Scope stop func() bool mu sync.Mutex accepted bool closed bool channel io.ReadWriteCloser }`.
- **`workerAuditSocket`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir, job string) string`.
  Calls: `filepath.Join`, `hex.EncodeToString`, `sha256.Sum256`.
- **`newWorkerAuditListener`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, scope toolbroker.Scope) (*workerAuditListener, error)`.
  Calls: `context.AfterFunc`, `ctxJob`, `fi.IsDir`, `fi.Mode`, `fi.Mode().Perm`, `filepath.Join`, `listener.Addr`, `listener.Addr().String`, `listener.Close`, `listener.SetDeadline`, `marker.Close`, `marker.Sync`, `net.ListenUnix`, `os.Chmod`, `os.Lstat`, `os.OpenFile`, `syncWorkerDirectory`, `time.Now`, `toolbroker.ValidAuditScope`, `workerAuditSocket`.
- **`(*workerAuditListener).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `channel.Close`, `l.listener.Close`, `l.mu.Lock`, `l.mu.Unlock`, `l.stop`.
- **`auditScopeProof`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(scope toolbroker.Scope) []byte`.
  Calls: `json.Marshal`, `sha256.Sum256`.
- **`(*workerAuditListener).Accept`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (io.ReadWriteCloser, error)`.
  Calls: `auditScopeProof`, `conn.Close`, `conn.SetDeadline`, `conn.Write`, `context.AfterFunc`, `ctx.Err`, `equalAuditProof`, `io.ReadFull`, `l.listener.AcceptUnix`, `l.listener.Close`, `l.mu.Lock`, `l.mu.Unlock`, `l.parent.Err`, `l.scope.Deadline.Before`, `make`, `owned.Close`, `stop`, `stopAccept`, `stopConn`, `stopParent`, `time.Now`, `time.Now().Add`.
- **`equalAuditProof`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a, b []byte) bool`.
  Calls: `len`.
- **Type `workerAuditConn`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn stop func() bool }`.
- **`(*workerAuditConn).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `c.Conn.Close`, `c.stop`.
- **`validWorkerAuditScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(record workerJobRecord, scope toolbroker.Scope) bool`.
  Calls: `record.Deadline.Equal`, `time.Now`, `toolbroker.ValidAuditScope`, `validWorkerRecord`, `workerJobSettled`, `workerTerminal`.
- **`decodeWorkerAuditScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(encoded string) (toolbroker.Scope, error)`.
  Calls: `base64.RawURLEncoding.DecodeString`, `json.Marshal`, `json.Unmarshal`, `len`, `string`, `time.Now`, `toolbroker.ValidAuditScope`.
- **`workerAttachAudit`** — workerAttachAudit forwards only a registered private audit socket, over its own SSH session. It never executes provider commands or consumes model output.
  Signature: `func(ctx context.Context, dir string, scope toolbroker.Scope, in io.Reader, out io.Writer) error`.
  Calls: `(&net.Dialer{}).DialContext`, `auditScopeProof`, `conn.Close`, `conn.SetDeadline`, `conn.Write`, `context.AfterFunc`, `fi.IsDir`, `fi.Mode`, `fi.Mode().Perm`, `io.ReadFull`, `lease.Close`, `os.Lstat`, `readWorkerRecord`, `relayWorkerAudit`, `scope.Deadline.Before`, `stop`, `time.Now`, `time.Now().Add`, `validWorkerAuditScope`, `workerAuditSocket`, `workerTryLease`.
- **`relayWorkerAudit`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, conn net.Conn, in io.Reader, out io.Writer) error`.
  Calls: `conn.Close`, `ctx.Done`, `ctx.Err`, `errors.New`, `input.Close`, `io.Copy`, `make`, `output.Close`.
- **Type `commandAuditLink`**: commandAuditLink owns a dedicated SSH process and both pipes. No stdout or stderr from the model can enter this channel. Close kills and reaps the link; it is not proof that the separate managed containers have stopped.
  Declaration: `struct { input io.WriteCloser output io.ReadCloser command *exec.Cmd cancel context.CancelFunc once sync.Once done chan struct{} }`.
- **`(*commandAuditLink).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `l.output.Read`.
- **`(*commandAuditLink).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `l.input.Write`.
- **`(*commandAuditLink).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `l.cancel`, `l.input.Close`, `l.once.Do`, `l.output.Close`.
- **`startCommandAuditLink`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(command *exec.Cmd, cancel context.CancelFunc) (*commandAuditLink, error)`.
  Calls: `cancel`, `close`, `command.Start`, `command.StdinPipe`, `command.StdoutPipe`, `command.Wait`, `input.Close`, `make`, `output.Close`.
- **`(*server).serveRemoteToolAudit`** — serveRemoteToolAudit admits only the configured host-key-verified SSH worker and the original registered human/job. It never retries a disconnected link. Production pair assembly must run this concurrently with its supervisor and cancel/fence the pair on any transport failure before admitting more tools.
  Signature: `func(ctx context.Context, endpoint string, scope toolbroker.Scope) error`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `context.WithDeadline`, `json.Marshal`, `link.Close`, `os.Getenv`, `remoteWorkerEndpoint`, `s.toolAuditor`, `startCommandAuditLink`, `toolbroker.ServeAudit`, `workerSSHCommand`.
- **`(*workerPairSupervisor).OpenAuditListener`** — OpenAuditListener derives ownership from the admitted supervisor, never from model-supplied context. Closed, abandoned and legacy supervisors cannot listen.
  Signature: `func(scope toolbroker.Scope) (*workerAuditListener, error)`.
  Calls: `context.WithValue`, `newWorkerAuditListener`, `s.available`, `s.mu.Lock`, `s.mu.Unlock`, `validWorkerAuditScope`.
