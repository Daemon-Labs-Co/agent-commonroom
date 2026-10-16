# cmd/server/worker_engine_gateway.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_engine_gateway.go)

Go implementation for worker engine gateway; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerEngineGateway`**: workerEngineGateway owns one fresh per-job socket in a private host directory. Only its socket inode is mounted into the engine, never the parent directory. Socket mode 0666 supports the container's rootless subordinate UID; the private mode-0700 host directory prevents access by other host accounts. This is NOT a credential file or a replacement for provider destination/address restrictions. The worker account/host remain trusted. Source forks can replace this boundary.
  Declaration: `struct { mu sync.Mutex scope toolbroker.Scope original, lifetime context.Context cancel context.CancelFunc gateway *providerGateway dir string directory, socket os.FileInfo closed bool once sync.Once stop func() bool closeError error }`.
- **`privateGatewayDirectory`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(path string) (os.FileInfo, error)`.
  Calls: `fi.IsDir`, `fi.Mode`, `fi.Mode().Perm`, `fi.Sys`, `os.Geteuid`, `os.Lstat`, `uint32`.
- **`newWorkerEngineGateway`** — root is an existing private worker-controlled staging root. Its ancestors and daemon selection are deployment trust. No caller/model may choose root/socket. A long Unix socket path is refused, never moved to a shared-directory fallback.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, root string) (*workerEngineGateway, error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithDeadline`, `ctxJob`, `filepath.Clean`, `filepath.IsAbs`, `filepath.Join`, `g.ready`, `g.shutdown`, `gateway.listener.(*gatewayListener).Listener.(*net.UnixListener).SetUnlinkOnClose`, `newProviderGateway`, `os.Lstat`, `os.MkdirTemp`, `os.Remove`, `privateGatewayDirectory`, `scope.Deadline.UTC`, `scope.Deadline.UTC().Round`, `time.Now`, `toolbroker.ValidAuditScope`.
- **`(*workerEngineGateway).ready`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope toolbroker.Scope) error`.
  Calls: `ctxJob`, `g.gateway.ctx.Err`, `g.lifetime.Err`, `g.mu.Lock`, `g.mu.Unlock`, `g.original.Err`, `os.Geteuid`, `os.Lstat`, `os.SameFile`, `privateGatewayDirectory`, `socket.Mode`, `socket.Mode().Perm`, `socket.Sys`, `time.Now`, `toolbroker.ValidAuditScope`, `uint32`.
- **`(*workerEngineGateway).Close`** — Close revokes network access; it never proves either container has stopped. Only the owned empty directory is removed, never recursively deleted. Changed ancestry/inodes or extra files leave an explicit cleanup error for recovery.
  Signature: `func() error`.
  Calls: `g.shutdown`, `g.stop`.
- **`(*workerEngineGateway).shutdown`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `g.cancel`, `g.gateway.Close`, `g.mu.Lock`, `g.mu.Unlock`, `g.once.Do`, `os.IsNotExist`, `os.Lstat`, `os.Remove`, `os.SameFile`, `privateGatewayDirectory`, `socket.Mode`.
