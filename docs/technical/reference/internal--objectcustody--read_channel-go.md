# internal/objectcustody/read_channel.go

[Module index](internal.md) · [Open source](../../../internal/objectcustody/read_channel.go)

Go implementation for read channel; declarations and calls are listed below.

Package: `objectcustody`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ReadReleaseGuard`**: ReadReleaseGuard is trusted receiving-side authority, not a model callback. It must pin the authenticated principal/session and recheck current room, receipt/revision/export policy. AcknowledgeRead must durably authorize release before bytes leave ReadChannel. StopReadAuthority must be local/nonblocking. A missing/no-op guard cannot admit production delivery. No isolation bypass exists.
  Declaration: `interface { AuthorizeRead(context.Context, Scope, Receipt) error AcknowledgeRead(context.Context, Receipt) error StopReadAuthority() }`.
- **Type `SealedReader`**: SealedReader must be a trusted isolated custody reader: Get verifies actual bytes and current authority, durably acknowledges release and returns no bytes on failure. Production setup must bind it to the actual quota-backed private object root; never supply a model path, host fallback or fake reader as evidence of stored bytes.
  Declaration: `interface { Get(context.Context, Scope, Receipt) ([]byte, error) }`.
- **Type `readChannelHeader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Operation string 'json:"operation"' Receipt Receipt 'json:"receipt"' }`.
- **`readChannelAbsent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value any) bool`.
  Calls: `reflect.ValueOf`, `v.IsNil`, `v.Kind`.
- **`writeReadChannelBytes`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w io.Writer, raw []byte) error`.
  Calls: `len`, `w.Write`.
- **`writeReadChannelHeader`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w io.Writer, h readChannelHeader) error`.
  Calls: `binary.BigEndian.PutUint32`, `json.Marshal`, `len`, `uint32`, `writeReadChannelBytes`.
- **`readReadChannelHeader`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r io.Reader) (readChannelHeader, error)`.
  Calls: `binary.BigEndian.Uint32`, `bytes.Equal`, `int`, `io.ReadFull`, `json.Marshal`, `json.Unmarshal`, `make`.
- **`readChannelContent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, r Receipt) bool`.
  Calls: `hex.EncodeToString`, `int64`, `len`, `sha256.Sum256`.
- **`ReadChannel`** — ReadChannel owns a ONE-SHOT already authenticated supervisor stream pinned to an exact sealed receipt. It is never model stdio or a public download handler. Both peers must be bound by trusted private setup; JSON fields are not authentication. The raw body is finite, rehashed, and withheld until current durable release acknowledgment. Any uncertainty returns nil and stops receiving-side authority. There is no reconnect, retry, path/URI selection or host execution fallback.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, scope Scope, receipt Receipt, guard ReadReleaseGuard) ([]byte, error)`.
  Calls: `readChannel`.
- **`ReadSupervisedChannel`** — ReadSupervisedChannel borrows the private attached helper stream. On success it leaves the stream open so its owner can Wait for actual process exit and inspect exited policy before Close joins the pumps. Error/cancellation still closes it. Returned bytes remain private: the owner MUST confirm cleanup and current durable release authority before downstream delivery. Mandatory isolation has no fallback.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, scope Scope, receipt Receipt, guard ReadReleaseGuard) ([]byte, error)`.
  Calls: `readChannel`.
- **`readChannel`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, scope Scope, receipt Receipt, guard ReadReleaseGuard, closeOnSuccess bool) (raw []byte, err error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithTimeout`, `ctx.Err`, `guard.AcknowledgeRead`, `guard.AuthorizeRead`, `guard.StopReadAuthority`, `int`, `io.ReadFull`, `make`, `readChannelAbsent`, `readChannelContent`, `readReadChannelHeader`, `stopIO`, `stream.Close`, `stream.Read`, `validReceipt`, `writeReadChannelHeader`.
- **`ServeReadChannel`** — ServeReadChannel releases only the receipt pinned by trusted session setup, using the sealed store's mandatory read/audit guards. It serves one object then closes. A successful write means bytes were authorized onto this stream, never that a user received or accepted them. Failed transfers do not refund/delete stored data.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, scope Scope, receipt Receipt, reader SealedReader) error`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithTimeout`, `ctx.Err`, `readChannelAbsent`, `readChannelContent`, `readReadChannelHeader`, `reader.Get`, `stopIO`, `stream.Close`, `validReceipt`, `writeReadChannelBytes`, `writeReadChannelHeader`.
