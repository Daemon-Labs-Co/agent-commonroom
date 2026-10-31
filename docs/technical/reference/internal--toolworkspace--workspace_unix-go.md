# internal/toolworkspace/workspace_unix.go

[Module index](internal.md) · [Open source](../../../internal/toolworkspace/workspace_unix.go)

Go implementation for workspace unix; declarations and calls are listed below.

Package: `toolworkspace`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Workspace`**: Workspace borrows a descriptor-rooted private workspace from trusted setup. The owner must keep root open until all operations finish. os.Root prevents path/symlink traversal outside that root, not bind mounts/devices/kernel escape. Separate verified volumes, dropped privileges and trusted root creation remain mandatory. This object never chooses roots from model input or opens a host root.
  Declaration: `struct{ root *os.Root }`.
- **`New`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(root *os.Root) (*Workspace, error)`.
- **`validPath`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(name string) bool`.
  Calls: `len`, `path.Clean`, `path.IsAbs`, `strings.ContainsAny`, `strings.HasPrefix`.
- **`safeFile`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(info os.FileInfo) bool`.
  Calls: `info.Mode`, `info.Mode().IsRegular`, `info.Mode().Perm`, `info.Sys`, `os.Geteuid`, `uint32`.
- **`active`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) bool`.
  Calls: `ctx.Err`.
- **`(*Workspace).Read`** — Read opens without following the final link or blocking on a FIFO, validates the actual inode and bounds actual bytes as well as stat size. A concurrent regular-file writer may change the contents; this is not a snapshot/CAS API.
  Signature: `func(ctx context.Context, name string) ([]byte, error)`.
  Calls: `active`, `file.Close`, `file.Stat`, `info.Size`, `io.LimitReader`, `io.ReadAll`, `len`, `safeFile`, `validPath`, `w.root.OpenFile`.
- **`(*Workspace).Write`** — Write atomically replaces a regular entry using a private new inode. It never truncates an existing inode, so a late hard-link alias cannot redirect the write. Only 0777 permission bits are preserved; ownership/ACL/xattrs are not migrated. Missing parents and unsafe existing destinations are refused. Rename/sync or post-rename identity uncertainty must never trigger an automatic retry.
  Signature: `func(ctx context.Context, name string, data []byte) error`.
  Calls: `dir.Sync`, `w.write`.
- **`(*Workspace).write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, name string, data []byte, syncDirectory func(*os.File) error) error`.
  Calls: `active`, `dir.Close`, `dir.Stat`, `errors.Is`, `file.Chmod`, `file.Close`, `file.Stat`, `file.Sync`, `file.Write`, `hex.EncodeToString`, `len`, `os.FileMode`, `os.SameFile`, `parentInfo.IsDir`, `parentRoot.Close`, `parentRoot.Lstat`, `parentRoot.OpenFile`, `parentRoot.Remove`, `parentRoot.Rename`, `path.Base`, `path.Dir`, `previous.Mode`, `previous.Mode().Perm`, `rand.Read`, `safeFile`, `syncDirectory`, `validPath`, `w.root.Lstat`, `w.root.OpenRoot`, `w.root.Stat`.
