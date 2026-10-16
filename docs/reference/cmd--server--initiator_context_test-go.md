# cmd/server/initiator_context_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/initiator_context_test.go)

Automated regression checks for initiator_context; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestDetachedInitiatorCopiesOnlyAuthenticatedIdentity`** — Regression contract: TestDetachedInitiatorCopiesOnlyAuthenticatedIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `context.WithValue`, `detached.Err`, `detached.Value`, `detachedInitiatorContext`, `t.Fatal`, `t.Parallel`, `withInitiatingActor`, `withInitiatingActor(inherited, "").Value`.
- **`TestTaskDispatchRetainsInitiatorInRegisteredJob`** — Regression contract: TestTaskDispatchRetainsInitiatorInRegisteredJob.
  Signature: `func(t *testing.T)`.
  Calls: `bound.Value`, `context.Background`, `dispatchFixture`, `dispatchRequest`, `make`, `registration.Register`, `s.load`, `s.managedJobContext`, `s.mu.Lock`, `s.mu.Unlock`, `t.Fatal`, `t.Parallel`, `time.After`, `w.Body.String`, `waitDispatch`.
- **`TestWorkflowRetainsRecordedInitiatorInRegisteredJobs`** — Regression contract: TestWorkflowRetainsRecordedInitiatorInRegisteredJobs.
  Signature: `func(t *testing.T)`.
  Calls: `awaitExecution`, `bound.Value`, `context.Background`, `executionServer`, `make`, `registration.Register`, `s.load`, `s.managedJobContext`, `s.mu.Lock`, `s.mu.Unlock`, `startExecution`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestDetachedChatContextRegistersInitiatorWithoutRequestLifetime`** — Regression contract: TestDetachedChatContextRegistersInitiatorWithoutRequestLifetime.
  Signature: `func(t *testing.T)`.
  Calls: `bound.Value`, `cancel`, `context.Background`, `context.WithCancel`, `context.WithValue`, `db.seed`, `detachedInitiatorContext`, `mockServer`, `registration.Register`, `s.load`, `s.managedJobContext`, `t.Fatal`, `t.Parallel`.
