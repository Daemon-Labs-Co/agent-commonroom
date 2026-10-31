# internal/objectcustody/store_unix.go

[Module index](internal.md) · [Open source](../../../internal/objectcustody/store_unix.go)

Go implementation for store unix; declarations and calls are listed below.

Package: `objectcustody`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Store`**: Store borrows a descriptor-pinned, private, quota-backed object root. Trusted setup must keep it open and prevent all other writers/mount/device replacement. It must never be the engine credential volume or a model-selected/host worktree. os.Root is path containment, not kernel/container or physical-quota proof.
  Declaration: `struct { root *os.Root journal Journal syncFile func(*os.File) error syncDirectory func(*os.File) error closeFile func(*os.File) error }`.
- **`New`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(root *os.Root, journal Journal) (*Store, error)`.
  Calls: `absent`, `dir.Close`, `f.Close`, `f.Sync`, `store.directory`.
- **`absent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value any) bool`.
  Calls: `reflect.ValueOf`, `v.IsNil`, `v.Kind`.
- **`active`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) bool`.
  Calls: `ctx.Err`.
- **`safeDirectory`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(info os.FileInfo) bool`.
  Calls: `info.IsDir`, `info.Mode`, `info.Mode().Perm`, `info.Sys`, `os.Geteuid`, `uint32`.
- **`safeObject`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(info os.FileInfo) bool`.
  Calls: `info.Mode`, `info.Mode().IsRegular`, `info.Mode().Perm`, `info.Sys`, `os.Geteuid`, `uint32`.
- **`(*Store).directory`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (*os.File, error)`.
  Calls: `dir.Close`, `dir.Stat`, `s.root.OpenFile`, `safeDirectory`.
- **`objectName`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(id string) string`.
- **`(*Store).Put`** — Put reads actual bounded bytes only after accounting/admission acknowledgment. The source must be a context-aware isolated tool export, never engine output, a model path/claim, credentials or an unverified live mutable source. This primitive proves stored byte identity only; source quiescence, content/format policy and revision attestation remain mandatory caller/journal obligations.
  Signature: `func(ctx context.Context, scope Scope, maxBytes int64, source io.Reader) (receipt Receipt, err error)`.
  Calls: `absent`, `active`, `current.Size`, `errors.Is`, `file.Seek`, `file.Stat`, `file.Write`, `hash.Sum`, `hash.Write`, `hex.EncodeToString`, `info.Size`, `int`, `int64`, `io.Copy`, `io.LimitReader`, `len`, `make`, `objectName`, `observed.Sum`, `os.SameFile`, `rand.Read`, `s.closeFile`, `s.directory`, `s.journal.Reserve`, `s.journal.Seal`, `s.root.Lstat`, `s.root.OpenFile`, `s.syncDirectory`, `s.syncFile`, `safeObject`, `sha256.New`, `source.Read`, `validScope`.
- **`(*Store).Get`** — Get verifies the actual private inode, size and content BEFORE acknowledging release. No raw bytes are returned on any authorization, integrity, cancellation, audit or owned-descriptor cleanup failure. It never accepts a model-provided filesystem path or URI. Cleanup rechecks cancellation before returning bytes.
  Signature: `func(ctx context.Context, scope Scope, receipt Receipt) (raw []byte, err error)`.
  Calls: `active`, `after.Size`, `clear`, `current.Size`, `file.Stat`, `hex.EncodeToString`, `info.Size`, `int64`, `io.LimitReader`, `io.ReadAll`, `len`, `objectName`, `os.SameFile`, `s.closeFile`, `s.directory`, `s.journal.AcknowledgeRead`, `s.journal.AuthorizeRead`, `s.root.Lstat`, `s.root.OpenFile`, `safeObject`, `sha256.Sum256`, `validReceipt`.
