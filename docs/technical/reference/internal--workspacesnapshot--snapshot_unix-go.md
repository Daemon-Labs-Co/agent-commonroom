# internal/workspacesnapshot/snapshot_unix.go

[Module index](internal.md) · [Open source](../../../internal/workspacesnapshot/snapshot_unix.go)

Go implementation for snapshot unix; declarations and calls are listed below.

Package: `workspacesnapshot`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Binding`**: Binding is trusted creator/import metadata, not model claims or authority. Execution.RevisionDigest MUST be empty on input: capture computes it from actual bounded source entries. BaseCommit identifies the imported revision; it does not claim that the workspace is clean or that Git HEAD is unchanged.
  Declaration: `struct { Execution objectcustody.ExecutionScope 'json:"execution"' BaseCommit string 'json:"baseCommit"' VolumeName string 'json:"volumeName"' VolumeNonce string 'json:"volumeNonce"' VolumeCreatedAt string 'json:"volumeCreatedAt"' }`.
- **Type `FrozenLease`**: FrozenLease is TRUSTED worker/controller infrastructure, never a model/plugin. VerifyFrozen must authenticate exact job/creator/volume/root identity, establish an actual read-only isolated view, prove all other writers fenced under the creator lease, and check current principal/revision/export policy plus durable authorization for bounded source observation. Existing workspace bytes must already be accounted; capture creates no custody object or on-disk copy. Read-only access in this process alone does not freeze other mounts. Calls must respect context. No production lease implementation is supplied by this package. Tests use explicit synthetic leases; those cannot admit real source work.
  Declaration: `interface { VerifyFrozen(context.Context, *os.Root, Binding) error }`.
- **Type `Member`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { OutputName string 'json:"outputName"' Name string 'json:"name"' MaxBytes int64 'json:"maxBytes"' }`.
- **Type `Entry`**: Entry describes observed source content; it is not an independently observed build/test check. Git's root metadata directory is excluded and never executed.
  Declaration: `struct { Name string 'json:"name"' Kind string 'json:"kind"' Mode uint32 'json:"mode"' Bytes int64 'json:"bytes"' Digest string 'json:"digest,omitempty"' }`.
- **Type `Snapshot`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Binding Binding 'json:"binding"' Scope objectcustody.ExecutionScope 'json:"scope"' Entries []Entry 'json:"entries"' Members []Member 'json:"members"' TotalBytes int64 'json:"totalBytes"' Digest string 'json:"digest"' }`.
- **Type `Source`**: Source borrows a pinned root and live trusted lease; caller keeps both valid until all reads finish. It never owns/deletes source or creates storage files. No constructor accepts a precomputed snapshot or model-provided bytes.
  Declaration: `struct { root *os.Root lease FrozenLease snapshot Snapshot selected map[string][]byte }`.
- **`absent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v any) bool`.
  Calls: `r.IsNil`, `r.Kind`, `reflect.ValueOf`.
- **`active`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) bool`.
  Calls: `ctx.Err`.
- **`validHexDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string, n int) bool`.
  Calls: `hex.DecodeString`, `hex.EncodeToString`, `len`.
- **`validBinding`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(b Binding) bool`.
  Calls: `len`, `objectcustody.ValidExecutionScope`, `strings.Repeat`, `t.IsZero`, `time.Parse`, `validHexDigest`.
- **`safeName`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(name string) bool`.
  Calls: `len`, `path.Clean`, `path.IsAbs`, `strings.ContainsAny`, `strings.EqualFold`, `strings.HasPrefix`, `strings.Split`, `unicode.IsControl`, `utf8.ValidString`.
- **`safeInfo`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(info os.FileInfo, directory bool) bool`.
  Calls: `info.IsDir`, `info.Mode`, `info.Mode().IsRegular`, `info.Mode().Perm`, `info.Sys`, `os.Geteuid`, `uint32`.
- **`validateMembers`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(b Binding, members []Member) bool`.
  Calls: `len`, `objectcustody.ValidIntent`, `safeName`, `strings.Repeat`.
- **`ValidSelection`** — ValidSelection validates finite metadata only, without opening a root or granting source authority. Capture still requires a live trusted frozen lease.
  Signature: `func(b Binding, members []Member) bool`.
  Calls: `validBinding`, `validateMembers`.
- **`verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, root *os.Root, lease FrozenLease, b Binding) error`.
  Calls: `absent`, `active`, `lease.VerifyFrozen`, `root.Lstat`, `safeInfo`, `validBinding`.
- **`Capture`** — Capture enumerates/hashes the actual finite tree twice under the frozen lease, buffers only approved members, and computes the revision identity itself. It never shells out, follows Git configuration, reads .git or fetches URLs. Double observation detects mutations but is NOT a replacement for real fencing.
  Signature: `func(ctx context.Context, root *os.Root, b Binding, members []Member, lease FrozenLease) (*Source, error)`.
  Calls: `append`, `hex.EncodeToString`, `json.Marshal`, `sameEntries`, `scan`, `sha256.Sum256`, `sort.Slice`, `validateMembers`, `verify`.
- **`sameEntries`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a, b []Entry) bool`.
  Calls: `len`.
- **`(*Source).Snapshot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() Snapshot`.
  Calls: `append`.
- **`(*Source).Current`** — Current re-observes the complete tree, not only the selected artifact. It refuses a changed unrelated source file and never edits/adopts the snapshot.
  Signature: `func(ctx context.Context) error`.
  Calls: `sameEntries`, `scan`, `verify`.
- **`(*Source).member`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(scope objectcustody.Scope, max int64) ([]byte, bool)`.
- **`(*Source).Open`** — Open prepares a lazy bounded reader; its first Read revalidates current source authority and the complete revision before returning an independent byte copy. Intended caller is the trusted custody store, not an HTTP/model endpoint. Custody must still reserve/account and establish its physical quota BEFORE reading this stream. Capture itself is a separately authorized bounded source-observation operation under FrozenLease.
  Signature: `func(ctx context.Context, scope objectcustody.Scope, max int64) (io.ReadCloser, error)`.
  Calls: `active`, `s.member`.
- **`(*Source).VerifyReceipt`** — VerifyReceipt ties stored bytes to the observed source member, not a supplied digest string alone. Physical custody quota and root proof remain separate.
  Signature: `func(ctx context.Context, receipt objectcustody.Receipt) error`.
  Calls: `hex.EncodeToString`, `int64`, `len`, `objectcustody.ValidReceipt`, `s.Current`, `s.member`, `sha256.Sum256`.
- **Type `sourceReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { mu sync.Mutex ctx context.Context data, raw []byte offset int closed, started bool source *Source }`.
- **`(*sourceReader).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `active`, `append`, `copy`, `len`, `r.mu.Lock`, `r.mu.Unlock`, `r.source.Current`, `verify`.
- **`(*sourceReader).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `clear`, `r.mu.Lock`, `r.mu.Unlock`.
- **`scan`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, root *os.Root, selected map[string]Member) ([]Entry, map[string][]byte, int64, error)`.
  Calls: `active`, `after.ModTime`, `after.ModTime().Equal`, `after.Mode`, `after.Size`, `append`, `child.Name`, `children[i].Name`, `children[j].Name`, `current.ModTime`, `current.ModTime().Equal`, `current.Mode`, `dir.Lstat`, `dir.Open`, `dir.OpenFile`, `dir.OpenRoot`, `f.Close`, `f.Read`, `f.Stat`, `h.Sum`, `h.Write`, `handle.Close`, `handle.ReadDir`, `hex.EncodeToString`, `info.IsDir`, `info.ModTime`, `info.Mode`, `info.Mode().Perm`, `info.Size`, `int64`, `len`, `make`, `nested.Close`, `nested.Lstat`, `os.SameFile`, `root.Lstat`, `rootInfo.Mode`, `rootInfo.Mode().Perm`, `safeInfo`, `safeName`, `sha256.New`, `sort.Slice`, `uint32`, `walk`.
