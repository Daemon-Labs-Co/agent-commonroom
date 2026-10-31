# cmd/server/worker_storage_check.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_storage_check.go)

Go implementation for worker storage check; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerStorageCheckPolicy`**: Operator-only, read-only deployment preflight. Independent provisioning supplies expected identity and reservation bounds; observed values never become policy. This command grants no execution/custody/export authority, has no flag bypass, and cannot substitute for fenced live inventory or mandatory container isolation.
  Declaration: `struct { Version int 'json:"version"' Roots []workerStorageCheckRoot 'json:"roots"' }`.
- **Type `workerStorageCheckRoot`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { CompanyID string 'json:"companyId"' Worker string 'json:"worker"' Path string 'json:"path"' Device uint64 'json:"device"' Inode uint64 'json:"inode"' Project uint32 'json:"project"' MaxBytes uint64 'json:"maxBytes"' MaxInodes uint64 'json:"maxInodes"' }`.
- **`(workerStorageCheckRoot).expected`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() projectquota.Expected`.
- **`loadWorkerStorageCheckPolicy`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(path string) (workerStorageCheckPolicy, error)`.
  Calls: `bytes.NewReader`, `dec.Decode`, `dec.DisallowUnknownFields`, `filepath.Clean`, `filepath.IsAbs`, `json.NewDecoder`, `len`, `membershipPartPattern.MatchString`, `readPrivateBounded`, `refuse`, `strictjson.ValidEncoding`, `string`, `uint64`, `uniqueObjectReadProfileJSON`, `workerAliasPattern.MatchString`.
- **`openWorkerStorageCheckRoot`** — Parent directories are operator-controlled; never accept paths from a model or request. The leaf cannot be a symlink/non-directory, and Inspect independently checks the opened descriptor's provisioned device/inode/project. No path retry.
  Signature: `func(path string) (*os.File, error)`.
  Calls: `os.NewFile`, `syscall.Open`, `uintptr`.
- **Type `workerStorageSetInspector`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, []projectquota.Root, uint64) ([]projectquota.Observation, error)`.
- **`checkWorkerStoragePolicy`** — The complete policy validates before any root opens. Keep all caller descriptors pinned until every company group has been inspected; close uncertainty withholds the whole result. Operator labels/budgets still do not prove inventory ownership.
  Signature: `func(ctx context.Context, policy workerStorageCheckPolicy, inspect workerStorageSetInspector) (err error)`.
  Calls: `append`, `ctx.Err`, `files[i].Close`, `inspect`, `len`, `openWorkerStorageCheckRoot`, `root.expected`, `uint64`.
- **`workerStorageCheckMain`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(args []string, stdout, stderr io.Writer) int`.
  Calls: `context.Background`, `runWorkerStorageCheck`.
- **`runWorkerStorageCheck`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, stdout, stderr io.Writer, inspect workerStorageSetInspector) int`.
  Calls: `checkWorkerStoragePolicy`, `ctx.Err`, `fmt.Fprintf`, `fmt.Fprintln`, `len`, `loadWorkerStorageCheckPolicy`, `refuse`.
