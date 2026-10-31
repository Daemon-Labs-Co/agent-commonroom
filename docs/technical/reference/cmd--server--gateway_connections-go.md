# cmd/server/gateway_connections.go

[Module index](cmd.md) · [Open source](../../../cmd/server/gateway_connections.go)

Go implementation for gateway connections; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `gatewayListener`**: gatewayListener caps accepted connections before net/http allocates per-client handler state. Excess clients are closed immediately, without parsing secrets or performing DNS. A source fork CAN remove it, BUT exposes worker resources to untrusted pre-header clients. It does not prevent volumetric socket/CPU attacks.
  Declaration: `struct { net.Listener slots chan struct{} }`.
- **`(*gatewayListener).Accept`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (net.Conn, error)`.
  Calls: `conn.Close`, `l.Listener.Accept`.
- **Type `gatewayClient`**: gatewayClient releases admission once even when HTTP shutdown and tunnel cleanup both close the same connection. Keep the slot for the whole tunnel.
  Declaration: `struct { net.Conn once sync.Once release func() closeErr error }`.
- **`(*gatewayClient).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `c.Conn.Close`, `c.once.Do`, `c.release`.
