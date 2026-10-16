# cmd/server/controller_lease.go

[Module index](cmd.md) · [Open source](../../cmd/server/controller_lease.go)

Go implementation for controller lease; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerLease`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ file *os.File }`.
- **`acquireControllerLease`** — acquireControllerLease enforces one controller per trusted local data directory before database/identity/worker startup. Never unlink the lease while a process can hold it: that would create a second lock inode. This is not a distributed database lease; separate directories/hosts must not share the same installation.
  Signature: `func(path string) (*controllerLease, error)`.
  Calls: `errors.New`, `f.Close`, `f.Fd`, `filepath.Dir`, `int`, `openPrivateRegular`, `os.MkdirAll`, `syscall.Flock`.
- **`(*controllerLease).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `int`, `lease.file.Close`, `lease.file.Fd`, `syscall.Flock`.
