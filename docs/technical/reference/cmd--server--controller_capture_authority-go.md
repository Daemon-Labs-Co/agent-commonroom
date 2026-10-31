# cmd/server/controller_capture_authority.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_capture_authority.go)

Go implementation for controller capture authority; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerCaptureEvidence`**: controllerCaptureEvidence MUST independently authenticate creator ownership, exact reviewed running helper/source/object volumes, actual frozen mount proof, writer exclusion and physically enforced storage accounting. It must not reenter server locks or this capture stream. No permissive/default implementation exists.
  Declaration: `interface { ValidateCapture(context.Context, room.State, workerJobRecord, workspacecapture.Request, workspacemount.Proof) error }`.
- **Type `controllerCaptureAuthorization`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' ActorID string 'json:"actorId"' Worker string 'json:"worker"' RecordDigest string 'json:"recordDigest"' Request workspacecapture.Request 'json:"request"' Proof workspacemount.Proof 'json:"proof"' }`.
- **Type `controllerCaptureAuthority`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server lifetime context.Context actor, worker string record workerJobRecord request workspacecapture.Request evidence controllerCaptureEvidence }`.
- **`copyCaptureRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(q workspacecapture.Request) workspacecapture.Request`.
  Calls: `slices.Clone`.
- **`captureDefinitionBound`** — Historical audit validation intentionally has no current deadline/connection or physical callback: replay must not turn a past observation into a fresh grant.
  Signature: `func(st room.State, q workspacecapture.Request, worker string) bool`.
  Calls: `validManagedJob`, `workspacecapture.ValidateContractRequest`.
- **`validControllerCaptureAuthorization`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v controllerCaptureAuthorization) bool`.
  Calls: `artifactRecordedActor`, `captureDefinitionBound`, `probeHex`, `workspacemount.ValidProof`.
- **`projectControllerCaptureAuthorization`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `json.Unmarshal`, `validControllerCaptureAuthorization`.
- **`newControllerCaptureAuthority`** — Private session context supplies the authenticated human and worker. Peer/model frames may not choose either identity. The record/request and lifetime are pinned; canceled authority cannot be revived with a new operation context. This source observation is not lifecycle admission, custody sealing or public export authority.
  Signature: `func(ctx context.Context, s *server, worker string, r workerJobRecord, q workspacecapture.Request, e controllerCaptureEvidence) (*controllerCaptureAuthority, error)`.
  Calls: `copyCaptureRequest`, `ctx.Err`, `ctx.Value`, `detachedProbeWorkerRecord`, `membershipPartPattern.MatchString`, `nilWorkerObservationDependency`, `retainedObservationRecord`, `workerAliasPattern.MatchString`, `workspacecapture.RequestDigest`.
- **`liveControllerCaptureBinding`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v controllerCaptureAuthorization, r workerJobRecord) bool`.
  Calls: `artifactCaptureJobStatus`, `coordinationJobAllowed`, `delegatedJobAllowed`, `managedTaskCurrent`, `sameWorkerBinding`, `time.Now`, `time.Now().Before`, `validControllerCaptureAuthorization`.
- **`(*controllerCaptureAuthority).current`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, p workspacemount.Proof) (room.State, controllerCaptureAuthorization, error)`.
  Calls: `artifactActor`, `copyCaptureRequest`, `ctx.Err`, `g.lifetime.Err`, `g.server.agentsBlockedLocked`, `g.server.load`, `liveControllerCaptureBinding`, `probeWorkerRecordDigest`.
- **`(*controllerCaptureAuthority).AuthorizeFrozenSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, b workspacesnapshot.Binding, p workspacemount.Proof) error`.
  Calls: `cancel`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `context.WithValue`, `copyCaptureRequest`, `ctx.Err`, `detachedProbeWorkerRecord`, `errors.Is`, `g.current`, `g.evidence.ValidateCapture`, `g.lifetime.Err`, `g.server.storageContext`, `nilWorkerObservationDependency`, `s.append`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `stop`, `storageCancel`, `workspacemount.ValidProof`.
