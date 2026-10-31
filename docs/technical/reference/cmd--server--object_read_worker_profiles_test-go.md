# cmd/server/object_read_worker_profiles_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_worker_profiles_test.go)

Automated regression checks for object_read_worker_profiles; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`objectReadProfileFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, g *controllerObjectReadSession) (string, objectReadWorkerProfileFile)`.
  Calls: `filepath.Join`, `provisionedObjectReadTestFiles`, `t.Helper`, `t.TempDir`, `writeObjectReadProfiles`.
- **`writeObjectReadProfiles`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, path string, file objectReadWorkerProfileFile)`.
  Calls: `json.MarshalIndent`, `os.WriteFile`, `t.Fatal`, `t.Helper`.
- **Type `addressedReadListener`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *syntheticReadListener address string }`.
- **`(*addressedReadListener).Addr`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() net.Addr`.
  Calls: `net.ResolveTCPAddr`.
- **`TestObjectReadWorkerServiceLifecycle`** — Regression contract: TestObjectReadWorkerServiceLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `artifactState`, `b.Close`, `c.AcknowledgeObjectRead`, `c.AuthorizeObjectRead`, `c.ObjectReadAuthorityLifetime`, `c.ObjectReadAuthorityLifetime().Err`, `c.StopObjectReadAuthority`, `close`, `controllerObjectReadKey`, `db.snapshot`, `g.lifetime.Err`, `len`, `make`, `net.Pipe`, `newObjectReadWorkerService`, `newSyntheticReadListener`, `objectReadProfileFixture`, `objectReadSessionFixture`, `service.Close`, `service.Connect`, `service.Serve`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestObjectReadWorkerProfilesRefuseInvalidPolicy`** — Regression contract: TestObjectReadWorkerProfilesRefuseInvalidPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `filepath.Join`, `newObjectReadWorkerService`, `objectReadProfileFixture`, `objectReadSessionFixture`, `os.Chmod`, `os.ReadFile`, `os.WriteFile`, `service.Close`, `service.Connect`, `string`, `strings.Replace`, `syntheticReadDialer`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `writeObjectReadProfiles`.
- **`TestObjectReadWorkerServiceRejectsWrongListenerAndJoinsAccept`** — Regression contract: TestObjectReadWorkerServiceRejectsWrongListenerAndJoinsAccept.
  Signature: `func(t *testing.T)`.
  Calls: `listener.accepts.Load`, `make`, `newObjectReadWorkerService`, `newSyntheticReadListener`, `objectReadProfileFixture`, `objectReadSessionFixture`, `service.Close`, `service.Serve`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
- **`TestObjectReadWorkerServiceShutdownDuringDial`** — Regression contract: TestObjectReadWorkerServiceShutdownDuringDial.
  Signature: `func(t *testing.T)`.
  Calls: `c.Done`, `c.Err`, `close`, `make`, `newObjectReadWorkerService`, `objectReadProfileFixture`, `objectReadSessionFixture`, `service.Close`, `service.Connect`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`, `time.After`.
- **Type `syntheticReadListenerFactory`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, string, string) (net.Listener, error)`.
- **`(syntheticReadListenerFactory).Listen`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, n, a string) (net.Listener, error)`.
  Calls: `f`.
- **`TestObjectReadWorkerServiceListenerFactoryFailureAndCancellation`** — Regression contract: TestObjectReadWorkerServiceListenerFactoryFailureAndCancellation.
  Signature: `func(t *testing.T)`.
  Calls: `c.Deadline`, `c.Done`, `c.Err`, `cancel`, `close`, `context.WithCancel`, `g.lifetime.Err`, `listener.accepts.Load`, `make`, `newObjectReadWorkerService`, `newSyntheticReadListener`, `objectReadProfileFixture`, `objectReadSessionFixture`, `service.Close`, `service.ListenAndServe`, `syntheticReadListenerFactory`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
- **Type `blockedReadServiceClose`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *addressedReadListener entered, release chan struct{} once sync.Once }`.
- **`(*blockedReadServiceClose).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `close`, `l.addressedReadListener.Close`, `l.once.Do`.
- **`TestObjectReadWorkerServiceJoinsRefusedListenerClosure`** — Regression contract: TestObjectReadWorkerServiceJoinsRefusedListenerClosure.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `g.lifetime.Err`, `listener.accepts.Load`, `make`, `newObjectReadWorkerService`, `newSyntheticReadListener`, `objectReadProfileFixture`, `objectReadSessionFixture`, `service.Close`, `service.Serve`, `t.Fatal`, `t.Parallel`, `time.After`.
