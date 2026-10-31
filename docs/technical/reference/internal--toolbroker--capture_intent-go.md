# internal/toolbroker/capture_intent.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/capture_intent.go)

Go implementation for capture intent; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `CaptureIntent`**: CaptureIntent is immutable finite lifecycle metadata, not source/quota/launch authority. Each use still requires current authenticated supervisor admission.
  Declaration: `struct { Request workspacecapture.Request 'json:"request"' Image string 'json:"image"' Objects CaptureObjectVolume 'json:"objects"' }`.
