# cmd/server/remote_worker.go

[Module index](cmd.md) · [Open source](../../cmd/server/remote_worker.go)

Go implementation for remote worker; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`validWorkerJobID`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v string) bool`.
  Calls: `workerJobPattern.MatchString`.
- **`remoteWorkerEndpoint`** — remoteWorkerEndpoint admits only the explicitly configured worker alias. SSH transports a constrained worker protocol, never arbitrary model commands. Disabling remote transport never restores direct host execution.
  Signature: `func(endpoint, configured, enabled string) (string, string, error)`.
  Calls: `errors.New`, `strings.HasPrefix`, `strings.HasSuffix`, `strings.TrimPrefix`, `strings.TrimSuffix`, `workerAliasPattern.MatchString`.
- **Type `workerRunRequest`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' IssuedAt time.Time 'json:"issuedAt"' Deadline time.Time 'json:"deadline"' JobID string 'json:"jobId"' Provider string 'json:"provider"' Key string 'json:"key"' Args []string 'json:"args"' Repository string 'json:"repository,omitempty"' ToolSeparation bool 'json:"toolSeparation,omitempty"' }`.
- **`encodeWorkerRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v workerRunRequest) (string, error)`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `json.Marshal`, `validateWorkerRun`.
- **`decodeWorkerRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(encoded string) (workerRunRequest, error)`.
  Calls: `base64.RawURLEncoding.DecodeString`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `errors.New`, `json.NewDecoder`, `len`, `new`, `string`, `strings.NewReader`, `validateWorkerRun`.
- **`validateWorkerRun`** — validateWorkerRun forbids arbitrary CLI options, mounts, environment and shell commands. Provider arguments must match the exact chat policy or the explicit development policy bound to an approved snapshot. The worker independently verifies its administrator-owned room registry before any launch.
  Signature: `func(v workerRunRequest) error`.
  Calls: `errors.New`, `len`, `now.Add`, `providerTurnArgs`, `regexp.MustCompile`, `regexp.MustCompile('^[A-Za-z0-9][A-Za-z0-9_-]*$').MatchString`, `repositoryDigestPattern.MatchString`, `requireCredentialSeparatedDevelopment`, `strings.TrimSpace`, `time.Now`, `time.Now().UTC`, `v.Deadline.After`, `v.Deadline.Sub`, `v.IssuedAt.After`, `v.IssuedAt.IsZero`, `validWorkerJobID`, `workerMembership`.
- **`workerSSHCommand`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, alias, action, value string) *exec.Cmd`.
  Calls: `exec.CommandContext`, `shellArgument`.
- **`remoteAgentCommand`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, endpoint, key string, args []string) (*exec.Cmd, func(), error)`.
  Calls: `controllerDevelopmentRepository`, `developmentTurnArgs`, `dispatchManagedWorker`, `errors.New`, `os.Getenv`, `remoteWorkerEndpoint`, `strings.SplitN`, `toolSeparationTurnArgs`.
- **Type `workerConfiguration`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Image string 'json:"image"' AuthDir string 'json:"authDir"' DevelopmentEnabled bool 'json:"developmentEnabled,omitempty"' DevelopmentRepositories map[string]developmentRepository 'json:"developmentRepositories,omitempty"' }`.
- **Type `workerJobRecord`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Pair *workerPairRecord 'json:"pair,omitempty"' Execution string 'json:"execution,omitempty"' Version int 'json:"version,omitempty"' RoomID string 'json:"roomId,omitempty"' ParticipantID string 'json:"participantId,omitempty"' ConnectionID string 'json:"connectionId,omitempty"' Scope string 'json:"scope,omitempty"' StartedAt time.Time 'json:"startedAt,omitempty"' UpdatedAt time.Time 'json:"updatedAt,omitempty"' Deadline time.Time 'json:"deadline,omitempty"' ShutdownConfirmed bool 'json:"shutdownConfirmed"' DiagnosticCode string 'json:"diagnosticCode,omitempty"' JobID string 'json:"jobId"' Status string 'json:"status"' Provider string 'json:"provider"' }`.
- **`workerStateDirectory`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (string, error)`.
  Calls: `filepath.Join`, `os.MkdirAll`, `os.UserHomeDir`.
- **`writeWorkerRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string, v workerJobRecord, exclusive bool) error`.
  Calls: `errors.New`, `f.Close`, `f.Name`, `f.Sync`, `f.Write`, `filepath.Join`, `json.Marshal`, `os.CreateTemp`, `os.OpenFile`, `os.Remove`, `os.Rename`, `syncWorkerDirectory`, `validWorkerRecord`.
- **`readWorkerRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir, job string) (workerJobRecord, error)`.
  Calls: `errors.New`, `filepath.Join`, `json.Unmarshal`, `pairedLifecycleVersion`, `readPrivateBounded`, `validWorkerRecord`.
- **`workerTerminal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(status string) bool`.
- **Type `workerStreamWriter`**: workerStreamWriter bounds remote output independently of the controller.
  Declaration: `struct { writer io.Writer remaining int cancel context.CancelFunc }`.
- **`(*workerStreamWriter).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `errors.New`, `len`, `w.cancel`, `w.writer.Write`.
- **`configureWorker`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `decoder.Decode`, `decoder.DisallowUnknownFields`, `errors.New`, `filepath.IsAbs`, `filepath.Join`, `fmt.Sprint`, `json.Marshal`, `json.NewDecoder`, `new`, `os.Setenv`, `os.UserHomeDir`, `readPrivateBounded`, `string`, `strings.NewReader`, `validAgentImage`, `validateDevelopmentRegistry`.
- **`agentWorkerMain`** — agentWorkerMain is a database-free entry point, invoked only by the installed worker wrapper over authenticated SSH. Provider secrets remain on this host. A fork CAN bypass isolation, BUT removes the documented containment boundary.
  Signature: `func(args []string, in io.Reader, out, diagnostics io.Writer) int`.
  Calls: `admissionCancel`, `admitWorkerJob`, `cancel`, `cleanup`, `cmd.Run`, `configureWorker`, `context.Background`, `context.WithDeadline`, `context.WithTimeout`, `ctx.Done`, `ctx.Err`, `deadline.After`, `decodeFenceRequest`, `decodeWorkerAuditScope`, `decodeWorkerRequest`, `dockerSecurityReady`, `errors.New`, `exec.CommandContext`, `exec.CommandContext(ctx, "docker", "image", "inspect", os.Getenv("COMMONROOM_AGENT_IMAGE")).Run`, `exec.CommandContext(ctx, "docker", "info", "--format", "{{json .}}").Output`, `fail`, `fenceWorkerJob`, `fi.Mode`, `fi.Mode().IsRegular`, `fi.Mode().Perm`, `fi.Size`, `filepath.Join`, `fmt.Fprintln`, `io.LimitReader`, `io.ReadAll`, `isolationAvailable`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `leases.Close`, `len`, `localIsolatedDevelopmentCommand`, `make`, `managedWorkerContainers`, `os.Getenv`, `os.Lstat`, `os.Stat`, `proofCancel`, `readWorkerRecord`, `signal.NotifyContext`, `stopSignals`, `stopWorkerContainer`, `string`, `strings.NewReader`, `ticker.Stop`, `time.NewTicker`, `time.Now`, `time.Now().Add`, `time.Now().UTC`, `validWorkerJobID`, `workerAttachAudit`, `workerFenceAcknowledgement`, `workerIdleProbe`, `workerProviderAuth`, `workerRuntimeReady`, `workerSnapshot`, `workerStateDirectory`, `writeWorkerRecord`.
