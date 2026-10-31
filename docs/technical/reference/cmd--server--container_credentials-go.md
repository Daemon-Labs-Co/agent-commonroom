# cmd/server/container_credentials.go

[Module index](cmd.md) · [Open source](../../../cmd/server/container_credentials.go)

Go implementation for container credentials; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`stageContainerCredential`** — stageContainerCredential creates only a new per-turn file in a trusted private directory. Explicit chmod permits the mapped container UID to read it despite the host umask; the private parent and read-only bind remain required. This is NOT credential isolation from model tools sharing the same container/UID.
  Signature: `func(path string, raw []byte) error`.
  Calls: `errors.New`, `f.Chmod`, `f.Close`, `f.Write`, `filepath.Dir`, `info.IsDir`, `info.Mode`, `info.Mode().Perm`, `info.Sys`, `json.Valid`, `len`, `os.Geteuid`, `os.Lstat`, `os.OpenFile`, `os.Remove`, `uint32`.
