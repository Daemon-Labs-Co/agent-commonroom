# cmd/server/job_diagnostics.go

[Module index](cmd.md) · [Open source](../../cmd/server/job_diagnostics.go)

Go implementation for job diagnostics; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`remoteJobsSnapshot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, alias string) (workerJobsSnapshot, error)`.
  Calls: `cmd.Run`, `errors.New`, `json.Unmarshal`, `len`, `newWorkerJobRecord`, `output.Bytes`, `pairedLifecycleVersion`, `validWorkerJobID`, `validWorkerRecord`, `workerSSHCommand`.
- **`remoteFenceJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, alias string, request workerFenceRequest) error`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `cmd.Run`, `errors.New`, `json.Marshal`, `json.Unmarshal`, `len`, `output.Bytes`, `request.record`, `validWorkerFenceAck`, `workerSSHCommand`.
- **`(*server).approvedSnapshot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (workerJobsSnapshot, error)`.
  Calls: `errors.New`, `os.Getenv`, `remoteJobsSnapshot`, `remoteWorkerEndpoint`, `s.jobSnapshot`.
- **Type `jobDiagnostic`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { room.ManagedJob WorkerStatus string 'json:"workerStatus"' LeaseState string 'json:"leaseState"' WorkerUpdatedAt *time.Time 'json:"workerUpdatedAt,omitempty"' CanReconcile bool 'json:"canReconcile"' }`.
- **`sameWorkerBinding`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(job room.ManagedJob, view workerJobView) bool`.
  Calls: `pairedLifecycleVersion`, `view.Deadline.Equal`.
- **`boundedJobDiagnostics`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, snapshot workerJobsSnapshot, snapshotErr error, rid, participant string) []jobDiagnostic`.
  Calls: `append`, `jobs[i].UpdatedAt.After`, `len`, `sameWorkerBinding`, `sort.Slice`, `validWorkerRecord`.
- **`(*server).jobDiagnostics`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `boundedJobDiagnostics`, `cancel`, `context.WithTimeout`, `hasRoom`, `http.Error`, `jobsDebugEnabled`, `len`, `membershipPartPattern.MatchString`, `r.Context`, `r.URL.Query`, `r.URL.Query().Get`, `reply`, `requestActor`, `requestedRoom`, `s.approvedSnapshot`, `s.dailyTurnBudget`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `time.Now`, `time.Now().UTC`, `workerJobSettled`.
- **`(*server).jobRecovery`** — jobRecovery is a required safety operation, not a premium feature. The UI diagnostic flag does not disable server cancellation/reconciliation safeguards.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `auditCancel`, `cancel`, `context.Background`, `context.WithTimeout`, `http.Error`, `http.MaxBytesReader`, `json.NewDecoder`, `json.NewDecoder(r.Body).Decode`, `len`, `os.Getenv`, `r.Context`, `r.URL.Query`, `r.URL.Query().Get`, `reply`, `requestActor`, `requestedRoom`, `s.append`, `s.approvedSnapshot`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.reconcileManagedJob`, `s.tripIntegrityLocked`, `s.validCSRF`, `sameWorkerBinding`, `stop`, `strings.TrimSpace`, `validWorkerJobID`, `validWorkerRecord`.
