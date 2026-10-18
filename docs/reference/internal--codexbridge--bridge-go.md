# internal/codexbridge/bridge.go

[Module index](internal.md) · [Open source](../../internal/codexbridge/bridge.go)

Go implementation for bridge; declarations and calls are listed below.

Package: `codexbridge`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Result`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ ThreadID, TurnID, Text string }`.
- **Type `Handler`**: Handler receives no capability or launch options from the model. The trusted caller binds scope/capability and dispatches through the original ToolSession. A source fork CAN supply a host handler, BUT violates mandatory isolation.
  Declaration: `func(context.Context, string, json.RawMessage) ([]byte, error)`.
- **Type `message`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID json.RawMessage 'json:"id,omitempty"' Method string 'json:"method,omitempty"' Params json.RawMessage 'json:"params,omitempty"' Result json.RawMessage 'json:"result,omitempty"' Error json.RawMessage 'json:"error,omitempty"' JSONRPC string 'json:"jsonrpc,omitempty"' }`.
- **Type `driver`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ctx context.Context stream io.ReadWriteCloser reader *bufio.Reader bytes, frames int pending []message result Result seen map[string]bool requests map[string]bool handler Handler }`.
- **`Run`** — Run owns one fresh thread and one turn, never resume/fork/config/review methods. Close MUST interrupt both stream directions; container shutdown remains the supervisor's responsibility. Cancellation closes the stream without replay.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, prompt, model string, handler Handler) (Result, error)`.
  Calls: `RunWithFinalizer`.
- **`RunWithFinalizer`** — RunWithFinalizer gives trusted worker custody one bounded post-turn operation before closing the attachment. It runs only after validated completion, never on a provider failure. No model selects this callback. Failure withholds text; cancellation still interrupts both stream directions and supervisor fencing remains mandatory. The finalizer must honor ctx and keep credentials private.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, prompt, model string, handler Handler, finalize func(context.Context) error) (Result, error)`.
  Calls: `bufio.NewReaderSize`, `context.AfterFunc`, `ctx.Err`, `d.handle`, `d.read`, `d.response`, `d.send`, `finalize`, `identifier.MatchString`, `json.Unmarshal`, `len`, `modelName.MatchString`, `readOnlySandbox`, `stop`, `stream.Close`, `utf8.ValidString`.
- **`(*driver).send`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v any) error`.
  Calls: `append`, `d.ctx.Err`, `d.stream.Write`, `json.Marshal`, `len`.
- **`(*driver).read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (message, error)`.
  Calls: `bytes.Equal`, `bytes.NewReader`, `d.reader.ReadSlice`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `exactObject`, `json.NewDecoder`, `len`, `strictjson.ValidEncoding`, `uniqueJSON`.
- **`(*driver).response`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(id string) (json.RawMessage, error)`.
  Calls: `append`, `d.read`, `len`, `string`.
- **`(*driver).handle`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(m message) (bool, error)`.
  Calls: `append`, `bytes.NewReader`, `d.ctx.Err`, `d.handler`, `d.send`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `exactObject`, `identifier.MatchString`, `json.NewDecoder`, `json.RawMessage`, `json.Unmarshal`, `len`, `string`, `utf8.Valid`, `validRPCID`.
- **`validRPCID`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) bool`.
  Calls: `bytes.NewReader`, `decoder.Decode`, `decoder.UseNumber`, `identifier.MatchString`, `json.NewDecoder`, `json.Unmarshal`, `n.Int64`.
- **`uniqueJSON`** — Reject duplicate keys at every depth before decoding typed protocol fields.
  Signature: `func(raw []byte) bool`.
  Calls: `bytes.NewReader`, `decoder.More`, `decoder.Token`, `json.Delim`, `json.NewDecoder`, `strings.ToLower`, `value`.
- **`exactObject`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, allowed, required []string) bool`.
  Calls: `json.Unmarshal`.
- **`readOnlySandbox`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) bool`.
  Calls: `bytes.TrimSpace`, `exactObject`, `json.Unmarshal`, `string`.
