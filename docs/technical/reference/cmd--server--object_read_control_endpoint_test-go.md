# cmd/server/object_read_control_endpoint_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_control_endpoint_test.go)

Automated regression checks for object_read_control_endpoint; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticReadDialer`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, string, string) (net.Conn, error)`.
- **`(syntheticReadDialer).DialContext`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, n, a string) (net.Conn, error)`.
  Calls: `f`.
- **Type `syntheticReadListener`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { incoming chan net.Conn closed, started chan struct{} closeOnce, startOnce sync.Once accepts atomic.Int32 failClose bool }`.
- **`newSyntheticReadListener`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() *syntheticReadListener`.
  Calls: `make`.
- **`(*syntheticReadListener).Accept`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (net.Conn, error)`.
  Calls: `close`, `l.accepts.Add`, `l.startOnce.Do`.
- **`(*syntheticReadListener).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `close`, `errors.New`, `l.closeOnce.Do`.
- **`(*syntheticReadListener).Addr`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() net.Addr`.
  Calls: `net.IPv4`.
- **`TestObjectReadControlEndpointLifecycle`** — Regression contract: TestObjectReadControlEndpointLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `append`, `b.Close`, `boundObjectReadFixture`, `c.AcknowledgeObjectRead`, `c.AuthorizeObjectRead`, `c.Deadline`, `c.StopObjectReadAuthority`, `calls.Add`, `calls.Load`, `connectObjectReadControl`, `db.snapshot`, `g.lifetime.Err`, `len`, `listener.accepts.Load`, `make`, `net.Pipe`, `newSyntheticReadListener`, `objectReadSessionFixture`, `serveObjectReadControlOnce`, `strings.Repeat`, `syntheticReadDialer`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.Until`.
- **`TestObjectReadControlEndpointRejectsBeforeDial`** — Regression contract: TestObjectReadControlEndpointRejectsBeforeDial.
  Signature: `func(t *testing.T)`.
  Calls: `boundObjectReadFixture`, `connectObjectReadControl`, `db.snapshot`, `errors.New`, `len`, `objectReadSessionFixture`, `syntheticReadDialer`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObjectReadControlEndpointInvalidIdentityAndDialFailure`** — Regression contract: TestObjectReadControlEndpointInvalidIdentityAndDialFailure.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `a.Write`, `b.Close`, `boundObjectReadFixture`, `cancel`, `connectObjectReadControl`, `context.WithCancel`, `errors.New`, `net.Pipe`, `objectReadSessionFixture`, `strings.Repeat`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObjectReadControlListenerCancellation`** — Regression contract: TestObjectReadControlListenerCancellation.
  Signature: `func(t *testing.T)`.
  Calls: `boundObjectReadFixture`, `cancel`, `context.WithCancel`, `db.snapshot`, `g.StopObjectReadAuthority`, `g.lifetime.Err`, `int32`, `len`, `listener.accepts.Load`, `make`, `newSyntheticReadListener`, `objectReadSessionFixture`, `serveObjectReadControlOnce`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
- **`TestObjectReadControlListenerCloseFailureRefusesSession`** — Regression contract: TestObjectReadControlListenerCloseFailureRefusesSession.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `b.Close`, `boundObjectReadFixture`, `c.StopObjectReadAuthority`, `connectObjectReadControl`, `db.snapshot`, `g.lifetime.Err`, `len`, `make`, `net.Pipe`, `newSyntheticReadListener`, `objectReadSessionFixture`, `serveObjectReadControlOnce`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`.
