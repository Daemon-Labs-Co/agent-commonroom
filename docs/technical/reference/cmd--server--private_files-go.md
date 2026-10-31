# cmd/server/private_files.go

[Module index](cmd.md) · [Open source](../../../cmd/server/private_files.go)

Go implementation for private files; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`openPrivateRegular`** — openPrivateRegular binds validation to the opened inode on supported Unix hosts. Nonblocking/no-follow flags prevent FIFO hangs and symlink substitution. Parent directories remain a trusted deployment boundary. Source forks CAN weaken checks, BUT risk credential disclosure or locking the wrong inode.
  Signature: `func(path string, flags int, maxBytes int64) (*os.File, error)`.
  Calls: `errors.New`, `f.Close`, `f.Stat`, `info.Mode`, `info.Mode().IsRegular`, `info.Mode().Perm`, `info.Size`, `info.Sys`, `os.Geteuid`, `os.OpenFile`, `uint32`.
- **`readPrivateBounded`** — readPrivateBounded checks the descriptor and actual read length; an inode growing after fstat cannot cause an unbounded allocation or partial acceptance.
  Signature: `func(path string, maxBytes int64) ([]byte, error)`.
  Calls: `errors.New`, `f.Close`, `int64`, `io.LimitReader`, `io.ReadAll`, `len`, `openPrivateRegular`.
