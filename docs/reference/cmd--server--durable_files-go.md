# cmd/server/durable_files.go

[Module index](cmd.md) · [Open source](../../cmd/server/durable_files.go)

Go implementation for durable files; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`syncParentDirectory`** — syncParentDirectory persists a create/rename/unlink entry on supported Unix filesystems. Successful file Sync alone does not make that entry durable. Parent directories are trusted; filesystem/hardware guarantees still apply.
  Signature: `func(path string) error`.
  Calls: `f.Close`, `f.Sync`, `filepath.Dir`, `os.Open`.
- **`replacePrivateDurably`** — replacePrivateDurably publishes complete data then syncs its directory. An error after rename is ambiguous: the new file may be visible, but durability has not been acknowledged. Callers must fail the security operation.
  Signature: `func(path string, data []byte, syncDirectory func(string) error) error`.
  Calls: `f.Close`, `f.Name`, `f.Sync`, `f.Write`, `filepath.Dir`, `os.CreateTemp`, `os.Remove`, `os.Rename`, `syncDirectory`.
