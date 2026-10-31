# cmd/server/object_read_worker_service.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_worker_service.go)

Go implementation for object read worker service; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadWorkerService`**: One explicitly provisioned service snapshot owns its sessions and cancellation. It does not create provider/model workers, listeners, keys or physical witnesses. No public API/model input can select policy files or bypass mandatory isolation. Close cancels first, then joins serving calls, dial/setup calls and client pumps. Stop callbacks must remain nonblocking and must not reenter service Close.
  Declaration: `struct { profiles *objectReadWorkerProfiles lifetime context.Context cancel context.CancelFunc mu sync.Mutex closed bool work sync.WaitGroup }`.
- **`newObjectReadWorkerService`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, path string) (*objectReadWorkerService, error)`.
  Calls: `context.WithCancel`, `ctx.Err`, `loadObjectReadWorkerProfiles`.
- **`(*objectReadWorkerService).begin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (context.Context, func(), error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithCancel`, `ctx.Err`, `s.lifetime.Err`, `s.mu.Lock`, `s.mu.Unlock`, `s.work.Add`, `s.work.Done`, `stop`.
- **`(*objectReadWorkerService).Connect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, worker string, r workerJobRecord, i workerObjectReadIntent, stop func(), dialer objectReadControlDialer) (*authenticatedObjectReadClient, error)`.
  Calls: `c.ObjectReadAuthorityLifetime`, `c.StopObjectReadAuthority`, `cancel`, `child.Err`, `connectObjectReadControl`, `context.AfterFunc`, `context.WithCancel`, `finish`, `live.Err`, `s.begin`, `s.lifetime.Err`, `s.mu.Lock`, `s.mu.Unlock`, `s.profiles.endpoint`, `s.work.Add`, `s.work.Done`, `stopService`.
- **`(*objectReadWorkerService).Serve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, listener net.Listener, g *controllerObjectReadSession) error`.
  Calls: `closeListener`, `finish`, `g.StopObjectReadAuthority`, `listener.Addr`, `listener.Addr().String`, `listener.Close`, `netip.ParseAddrPort`, `nilWorkerObservationDependency`, `s.begin`, `s.profiles.endpoint`, `serveObjectReadControlOnce`.
- **`(*objectReadWorkerService).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `s.cancel`, `s.mu.Lock`, `s.mu.Unlock`, `s.work.Wait`.
- **Type `objectReadControlListenerFactory`**: Deployment passes a net.ListenConfig. Test/source injections must honor context cancellation and return a listener whose Close unblocks Accept. Only the fixed configured TCP address can be bound; neither model metadata nor a caller string selects a listener. Any uncertain creation refuses and retains no read authority.
  Declaration: `interface { Listen(context.Context, string, string) (net.Listener, error) }`.
- **`(*objectReadWorkerService).ListenAndServe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, g *controllerObjectReadSession, factory objectReadControlListenerFactory) error`.
  Calls: `cancel`, `child.Err`, `context.AfterFunc`, `context.WithTimeout`, `factory.Listen`, `finish`, `g.StopObjectReadAuthority`, `g.lifetime.Err`, `listenCtx.Err`, `listener.Close`, `nilWorkerObservationDependency`, `s.Serve`, `s.begin`, `s.profiles.endpoint`, `stopSession`.
