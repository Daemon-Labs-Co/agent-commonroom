# internal/toolbroker/workspace_bootstrap.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/workspace_bootstrap.go)

Go implementation for workspace bootstrap; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `StreamingDockerRunner`**: StreamingDockerRunner is used only for approved bounded snapshot streams, never model commands. The production adapter retains all DockerRunner restrictions.
  Declaration: `interface { RunInput(context.Context, []string, io.Reader, io.Writer) error }`.
- **`(*ToolContainerRuntime).Bootstrap`** — Bootstrap initializes the started idle tool container before engine startup. The worker supplies a staged approved bundle, never an arbitrary model file. It consumes/closes input, including on cancellation, and permits one attempt. The supervisor must retain the lease until this call returns and fence both containers on any uncertainty. A receipt is not independent image attestation.
  Signature: `func(ctx context.Context, input io.ReadCloser) (workspacebootstrap.Receipt, error)`.
  Calls: `cancel`, `child.Err`, `closeInput`, `closeOnce.Do`, `context.AfterFunc`, `input.Close`, `io.LimitReader`, `len`, `output.snapshot`, `p.workspace.Store`, `r.context`, `r.lifetime.Err`, `r.mu.Lock`, `r.mu.Unlock`, `r.original.Err`, `r.verify`, `runner.RunInput`, `stop`, `validReceiptJSON`, `workspacebootstrap.ValidReceipt`.
- **`validReceiptJSON`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, receipt *workspacebootstrap.Receipt) bool`.
  Calls: `bytes.NewReader`, `decodeInspection`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `json.NewDecoder`, `len`.
