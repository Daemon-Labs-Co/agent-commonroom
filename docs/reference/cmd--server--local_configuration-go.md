# cmd/server/local_configuration.go

[Module index](cmd.md) · [Open source](../../cmd/server/local_configuration.go)

Go implementation for local configuration; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `localConfiguration`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { DatabaseURL string 'json:"databaseUrl"' RemoteWorkerAlias string 'json:"remoteWorkerAlias"' RemoteWorkerEnabled bool 'json:"remoteWorkerEnabled"' }`.
- **`loadLocalConfiguration`** — loadLocalConfiguration reads a bounded private descriptor, never an unchecked path. Invalid files stop startup; absence preserves environment/default setup. Decode errors are intentionally generic so credential-containing input is not copied into startup logs. Parent directories remain operator-controlled.
  Signature: `func(path string) (localConfiguration, error)`.
  Calls: `bytes.NewReader`, `bytes.TrimSpace`, `d.Decode`, `d.DisallowUnknownFields`, `errors.New`, `json.NewDecoder`, `len`, `new`, `os.IsNotExist`, `readPrivateBounded`.
