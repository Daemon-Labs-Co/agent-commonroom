# cmd/server/workspace_capture_physical.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_capture_physical.go)

Go implementation for workspace capture physical; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `capturePhysicalBindingGuard`**: capturePhysicalBindingGuard is mandatory independent custody infrastructure. It MUST authenticate exact durable local/controller lifecycle binding, helper image/object volume, account charges and actual physical source/object quotas, plus the proof's origin on this helper's exclusively owned private channel. It returns only the already durably bound immutable ID, never name discovery. Calls must not reenter server locks or the capture stream. No default exists.
  Declaration: `interface { BoundCapture(context.Context, room.State, string, workerJobRecord, workspacecapture.Request, string, toolbroker.CaptureObjectVolume, workspacemount.Proof) (string, error) }`.
- **Type `workspaceCapturePhysicalEvidence`**: workspaceCapturePhysicalEvidence observes worker-local metadata under the actual creator ownership lifetime. It never launches, reads source/object bytes or changes accounting. A dishonest runner/binding guard invalidates isolation.
  Declaration: `struct { lifetime context.Context dir, root, worker, image string record workerJobRecord request workspacecapture.Request objects toolbroker.CaptureObjectVolume runner toolbroker.DockerRunner binding capturePhysicalBindingGuard plan *toolbroker.CaptureContainerPlan }`.
- **`newWorkspaceCapturePhysicalEvidence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, worker, image string, request workspacecapture.Request, objects toolbroker.CaptureObjectVolume, binding capturePhysicalBindingGuard) (*workspaceCapturePhysicalEvidence, error)`.
  Calls: `copyCaptureRequest`, `detachedProbeWorkerRecord`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `nilWorkerObservationDependency`, `probeHex`, `retainedObservationRecord`, `strings.HasPrefix`, `strings.TrimPrefix`, `toolbroker.NewCaptureContainerPlan`, `workerAliasPattern.MatchString`.
- **`(*workspaceCapturePhysicalEvidence).ValidateCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, record workerJobRecord, request workspacecapture.Request, proof workspacemount.Proof) error`.
  Calls: `cancel`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `copyCaptureRequest`, `ctx.Err`, `detachedProbeWorkerRecord`, `e.binding.BoundCapture`, `e.lifetime.Err`, `e.plan.VerifyAdmission`, `nilWorkerObservationDependency`, `probeHex`, `probeWorkerRecordDigest`, `readRetainedWorkspaceObservation`, `sameWorkerBinding`, `stop`, `validManagedJob`, `workspacecapture.RequestDigest`, `workspacemount.ValidProof`.
