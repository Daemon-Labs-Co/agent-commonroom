# cmd/server/workspace_capture_physical_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_capture_physical_test.go)

Automated regression checks for workspace_capture_physical; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticCaptureBinding`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { f *physicalCaptureFixture calls int mode string }`.
- **`(*syntheticCaptureBinding).BoundCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, worker string, r workerJobRecord, q workspacecapture.Request, image string, o toolbroker.CaptureObjectVolume, p workspacemount.Proof) (string, error)`.
  Calls: `ctx.Err`, `f.probe.checkLock`, `probeWorkerRecordDigest`, `strings.Repeat`, `workspacecapture.RequestDigest`.
- **Type `physicalCaptureFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { probe *assembledProbeFixture request workspacecapture.Request objects toolbroker.CaptureObjectVolume mode string inspections int }`.
- **`(*physicalCaptureFixture).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `append`, `base64.RawURLEncoding.EncodeToString`, `ctx.Err`, `f.probe.checkLock`, `io.WriteString`, `json.Marshal`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `p.inspection`, `p.t.Fatal`, `r.UpdatedAt.Add`, `strings.Repeat`, `workspacecapture.RequestDigest`, `writeWorkerRecord`.
- **`TestWorkspaceCapturePhysicalEvidence`** — Regression contract: TestWorkspaceCapturePhysicalEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `e.ValidateCapture`, `lease.Close`, `newAssembledProbeFixture`, `newWorkspaceCapturePhysicalEvidence`, `physicalControllerFixture`, `probeControllerState`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerTryLease`.
- **`TestWorkspaceCapturePhysicalMissingOwnership`** — Regression contract: TestWorkspaceCapturePhysicalMissingOwnership.
  Signature: `func(t *testing.T)`.
  Calls: `errors.Is`, `missing.ValidateCapture`, `newWorkspaceCapturePhysicalEvidence`, `strings.Repeat`, `t.Context`, `t.Fatal`.
