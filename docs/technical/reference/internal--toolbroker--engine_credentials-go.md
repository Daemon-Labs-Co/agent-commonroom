# internal/toolbroker/engine_credentials.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/engine_credentials.go)

Go implementation for engine credentials; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*EngineContainerRuntime).ExportCredential`** — ExportCredential consumes one attempt after Attach and a running-policy check. It stops the provider child via the private bootstrap and returns sensitive bounded bytes only to trusted worker custody. No retry/adoption/host fallback.
  Signature: `func(ctx context.Context, id string) ([]byte, error)`.
  Calls: `cancel`, `child.Err`, `clear`, `context.WithTimeout`, `done`, `len`, `output.buffer.Bytes`, `output.buffer.Reset`, `output.mu.Lock`, `output.mu.Unlock`, `output.snapshot`, `r.context`, `r.inspect`, `r.mu.Lock`, `r.mu.Unlock`, `r.runner.Run`, `r.verify`.
