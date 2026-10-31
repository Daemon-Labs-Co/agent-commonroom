# internal/workspacecapture/channel.go

[Module index](internal.md) · [Open source](../../../internal/workspacecapture/channel.go)

Go implementation for channel; declarations and calls are listed below.

Package: `workspacecapture`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Report`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Scope objectcustody.ExecutionScope 'json:"scope"' SnapshotDigest string 'json:"snapshotDigest"' Receipts []objectcustody.Receipt 'json:"receipts"' }`.
- **Type `captureFrame`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Sequence uint64 'json:"sequence"' BindingDigest string 'json:"bindingDigest"' Operation string 'json:"operation"' Proof *workspacemount.Proof 'json:"proof,omitempty"' Intent *objectcustody.Intent 'json:"intent,omitempty"' Receipt *objectcustody.Receipt 'json:"receipt,omitempty"' Report *Report 'json:"report,omitempty"' }`.
- **Type `captureAck`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Sequence uint64 'json:"sequence"' Accepted bool 'json:"accepted"' }`.
- **`RequestDigest`** — RequestDigest validates finite metadata only. Authentication and reviewed helper, creator, source and physical quota proof remain mandatory trusted setup duties.
  Signature: `func(request Request) (string, error)`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `len`, `sha256.Sum256`, `workspacemount.ValidChallenge`, `workspacesnapshot.ValidSelection`.
- **`controlWrite`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w io.Writer, value any) error`.
  Calls: `binary.BigEndian.PutUint32`, `json.Marshal`, `len`, `uint32`, `w.Write`.
- **`controlRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r io.Reader, value any) error`.
  Calls: `binary.BigEndian.Uint32`, `bytes.Equal`, `int`, `io.ReadFull`, `json.Marshal`, `json.Unmarshal`, `make`.
- **`memberIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(request Request, i objectcustody.Intent) bool`.
  Calls: `objectcustody.ValidIntent`.
- **`memberKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i objectcustody.Intent) string`.
- **`digest32`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) bool`.
  Calls: `hex.DecodeString`, `hex.EncodeToString`, `len`.
- **`validReport`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(request Request, r Report) bool`.
  Calls: `digest32`, `len`, `memberIntent`, `memberKey`, `objectcustody.ValidExecutionScope`, `objectcustody.ValidReceipt`.
- **Type `CaptureChannel`**: CaptureChannel uses ONLY an authenticated supervisor-owned stream connected to the reviewed helper, never model stdin/stdout. Both peers pin the same Request before I/O. No reconnect, retry, read/release operation or source bytes exist in this protocol. Kernel/worker/volume/quota proof comes from trusted authorities, not successful frame decoding. A fork exposing/substituting this stream loses those guarantees; mandatory managed isolation has no host fallback.
  Declaration: `struct { mu sync.Mutex stream io.ReadWriteCloser lifetime context.Context cancel context.CancelFunc request Request digest string sequence uint64 failed, finished bool stop func() stopOnce sync.Once }`.
- **`NewCaptureChannel`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, request Request, stop func()) (*CaptureChannel, error)`.
  Calls: `RequestDigest`, `append`, `context.WithTimeout`, `ctx.Err`, `isNilDependency`.
- **`(*CaptureChannel).fail`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `c.cancel`, `c.stop`, `c.stopOnce.Do`, `c.stream.Close`.
- **`(*CaptureChannel).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `c.cancel`, `c.stream.Close`.
- **`(*CaptureChannel).exchange`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, frame captureFrame) error`.
  Calls: `c.fail`, `c.lifetime.Deadline`, `c.lifetime.Err`, `c.mu.Lock`, `c.mu.Unlock`, `c.stream.Close`, `call.Err`, `cancel`, `closeIO`, `context.AfterFunc`, `context.WithDeadline`, `context.WithTimeout`, `controlRead`, `controlWrite`, `ctx.Err`, `parentStop`, `shortCancel`.
- **`(*CaptureChannel).AuthorizeFrozenSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, b workspacesnapshot.Binding, p workspacemount.Proof) error`.
  Calls: `c.exchange`, `workspacemount.ValidProof`.
- **`(*CaptureChannel).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i objectcustody.Intent) error`.
  Calls: `c.exchange`, `memberIntent`.
- **`(*CaptureChannel).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `c.exchange`, `memberIntent`, `objectcustody.ValidReceipt`.
- **`(*CaptureChannel).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Scope, objectcustody.Receipt) error`.
- **`(*CaptureChannel).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
- **`(*CaptureChannel).Finish`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, result CapturedObjects) error`.
  Calls: `append`, `c.exchange`, `len`, `validReport`.
- **`ServeCaptureChannel`** — ServeCaptureChannel acknowledges only after current authority and durable journal callbacks succeed. Its result is private metadata, not public delivery: callers must still verify helper exit/cleanup and policy before publishing it. Callbacks must not reenter this same stream. Partial writes/seals survive transport errors.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, request Request, authority Authority, journal objectcustody.Journal) (Report, error)`.
  Calls: `serveCaptureChannel`.
- **`ServeSupervisedCaptureChannel`** — ServeSupervisedCaptureChannel borrows the supervisor's attached process stream. On protocol success it leaves that stream open: the caller MUST wait for actual successful process exit, verify exited-container policy, and finally close/join the process before owned cleanup/report release. Error/cancellation still closes the stream promptly. Authentication, custody and mandatory isolation are unchanged; this does not permit model pipes, a host launch or an unsupervised fallback.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, request Request, authority Authority, journal objectcustody.Journal) (Report, error)`.
  Calls: `serveCaptureChannel`.
- **`serveCaptureChannel`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, request Request, authority Authority, journal objectcustody.Journal, closeOnSuccess bool) (report Report, err error)`.
  Calls: `RequestDigest`, `append`, `authority.AuthorizeFrozenSource`, `cancel`, `closeIO`, `context.AfterFunc`, `context.WithTimeout`, `controlRead`, `controlWrite`, `ctx.Err`, `errors.Join`, `isNilDependency`, `journal.Reserve`, `journal.Seal`, `len`, `memberIntent`, `memberKey`, `missingAuthority`, `missingJournal`, `objectcustody.ValidReceipt`, `stream.Close`, `uint64`, `validReport`, `workspacemount.ValidProof`.
