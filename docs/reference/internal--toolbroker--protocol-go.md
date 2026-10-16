# internal/toolbroker/protocol.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/protocol.go)

Go implementation for protocol; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Scope`**: Scope is supplied by trusted supervisor state, never inferred from a model. Version is the bridge protocol version, independent of worker lifecycle versions.
  Declaration: `struct { Version int 'json:"version"' JobID string 'json:"jobId"' Room string 'json:"room"' Participant string 'json:"participant"' Connection string 'json:"connection"' Provider string 'json:"provider"' RepositoryDigest string 'json:"repositoryDigest"' Deadline time.Time 'json:"deadline"' }`.
- **Type `Request`**: Request accepts only typed workspace operations, never container launch options. Path validation here is lexical; execution MUST separately prevent symlink, mount and special-file escapes using trusted workspace descriptors.
  Declaration: `struct { Scope Scope 'json:"scope"' Capability string 'json:"capability"' CallID string 'json:"callId"' Operation string 'json:"operation"' Path string 'json:"path,omitempty"' Content string 'json:"content,omitempty"' Command []string 'json:"command,omitempty"' }`.
- **`identifier`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) bool`.
  Calls: `len`.
- **`digest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string, size int) bool`.
  Calls: `len`.
- **`operation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) bool`.
- **`validRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r Request) bool`.
  Calls: `digest`, `identifier`, `operation`, `validAction`.
- **`validAction`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a Action) bool`.
  Calls: `actionTextBudget`, `len`, `operation`, `path.Clean`, `path.IsAbs`, `strings.ContainsAny`, `strings.ContainsRune`, `strings.HasPrefix`, `utf8.ValidString`.
- **`Decode`** — Decode bounds the whole message and rejects ambiguous, unknown or trailing JSON. It never includes the request or capability in diagnostics.
  Signature: `func(reader io.Reader) (Request, error)`.
  Calls: `bytes.NewReader`, `d.Decode`, `d.DisallowUnknownFields`, `d.Token`, `io.LimitReader`, `io.ReadAll`, `json.NewDecoder`, `len`, `strictjson.ValidEncoding`, `uniqueValue`, `validRequest`.
- **`uniqueValue`** — uniqueValue detects duplicate members at every nesting level before decoding.
  Signature: `func(d *json.Decoder, depth int) error`.
  Calls: `d.More`, `d.Token`, `knownKey`, `uniqueValue`.
- **Type `Authority`**: Authority owns immutable policy and a capability hash. Do not serialize it. The bridge transport must authenticate its expected peer independently.
  Declaration: `struct { mu sync.Mutex scope Scope issued time.Time hash [32]byte allowed map[string]bool calls map[string]bool revoked bool done chan struct{} slot chan struct{} }`.
- **`New`** — New mints authority for at most three minutes using a trusted clock. The raw token is returned once for protected delivery; never log or persist it in chat.
  Signature: `func(scope Scope, allowed []string, now time.Time) (*Authority, string, error)`.
  Calls: `hex.EncodeToString`, `len`, `make`, `operation`, `rand.Read`, `scope.Deadline.UTC`, `scope.Deadline.UTC().Round`, `sha256.Sum256`, `validScope`.
- **`(*Authority).Reserve`** — Reserve atomically consumes a call ID, including failed future executions. Replays are refused, not retried. Future dispatch still requires durable audit intent, cancellation checks and credential-free execution. Caller time is trusted.
  Signature: `func(r Request, now time.Time) error`.
  Calls: `a.mu.Lock`, `a.mu.Unlock`, `json.Marshal`, `len`, `now.Before`, `r.Scope.Deadline.UTC`, `r.Scope.Deadline.UTC().Round`, `sha256.Sum256`, `subtle.ConstantTimeCompare`, `validRequest`.
- **`(*Authority).Revoke`** — Revoke fences all future reservations. It does not stop already running tools; the future supervisor must cancel and verify both container lifecycles.
  Signature: `func()`.
  Calls: `a.mu.Lock`, `a.mu.Unlock`, `close`.
- **`knownKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) bool`.
- **`validScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(scope Scope, now time.Time) bool`.
  Calls: `digest`, `identifier`, `scope.Deadline.After`, `scope.Deadline.Sub`.
- **`actionTextBudget`** — actionTextBudget refuses typed oversized input before hashing/copying/encoding. JSON encoding may expand escaped characters; the serialized cap is still checked afterwards, but its worst-case allocation is now bounded.
  Signature: `func(a Action) bool`.
  Calls: `len`.
- **`(*Authority).permitsAction`** — permitsAction is a defense-in-depth executor gate, not an audit acknowledgement or capability-bearing admission. It never expands or consumes the copied job policy. External requests must still pass the authenticated audited dispatcher.
  Signature: `func(scope Scope, action Action, now time.Time) bool`.
  Calls: `a.mu.Lock`, `a.mu.Unlock`, `now.Before`, `scope.Deadline.UTC`, `scope.Deadline.UTC().Round`, `validAction`.
