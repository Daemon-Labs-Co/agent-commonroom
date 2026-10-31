# cmd/server/history.go

[Module index](cmd.md) · [Open source](../../../cmd/server/history.go)

Go implementation for history; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`historyRPC`** — historyRPC uses the supported app-server protocol without starting model turns. This is a current host-side history utility, not an isolated agent worker. Under the mandatory-isolation design, human-authorized history import must be a separate trusted capability; agents must not inherit access to the user's whole session store. Imported context never authorizes host execution. See docs/technical/archive/architecture.md#mandatory-container-isolation-design-contract.
  Signature: `func(ctx context.Context, method string, params any) (json.RawMessage, error)`.
  Calls: `bufio.NewScanner`, `cancel`, `cmd.Start`, `cmd.StdinPipe`, `cmd.StdoutPipe`, `cmd.Wait`, `context.WithTimeout`, `encoder.Encode`, `errors.New`, `exec.CommandContext`, `in.Close`, `json.NewEncoder`, `json.Unmarshal`, `make`, `request`, `scanner.Buffer`, `scanner.Bytes`, `scanner.Err`, `scanner.Scan`.
