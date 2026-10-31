# cmd/server/gateway_connections_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/gateway_connections_test.go)

Automated regression checks for gateway_connections; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticGatewayListener`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { incoming chan net.Conn done chan struct{} once sync.Once }`.
- **`(*syntheticGatewayListener).Accept`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (net.Conn, error)`.
- **`(*syntheticGatewayListener).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `close`, `l.once.Do`.
- **`(*syntheticGatewayListener).Addr`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() net.Addr`.
- **`TestGatewayListenerBoundsAndReleasesConnections`** — Regression contract: TestGatewayListenerBoundsAndReleasesConnections.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `b.Close`, `base.Close`, `c.Close`, `errors.Is`, `firstPeer.Close`, `l.Accept`, `len`, `make`, `net.Pipe`, `nextPeer.Close`, `peer.Close`, `peer.Read`, `peer.SetReadDeadline`, `secondPeer.Close`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.After`, `time.Now`, `time.Now().Add`.
- **`TestGatewayClientConcurrentClose`** — Regression contract: TestGatewayClientConcurrentClose.
  Signature: `func(t *testing.T)`.
  Calls: `c.Close`, `len`, `make`, `net.Pipe`, `peer.Close`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestGatewayCapsClientsBeforeHeaders`** — Regression contract: TestGatewayCapsClientsBeforeHeaders.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `c.Close`, `clients[0].Close`, `context.Background`, `excess.Close`, `excess.Read`, `excess.SetReadDeadline`, `filepath.Join`, `g.Close`, `net.DialTimeout`, `newProviderGateway`, `os.MkdirTemp`, `os.RemoveAll`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.Now`, `time.Now().Add`, `waitGatewaySlots`.
- **`waitGatewaySlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, l *gatewayListener, want int)`.
  Calls: `len`, `t.Fatalf`, `t.Helper`, `time.Now`, `time.Now().Add`, `time.Now().Before`, `time.Sleep`.
