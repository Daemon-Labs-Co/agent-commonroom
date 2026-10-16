# cmd/server/agents.go

[Module index](cmd.md) · [Open source](../../cmd/server/agents.go)

Go implementation for agents; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `codexAdapter`**: codexAdapter runs Codex conversations through the current local/SSH launcher. All model turns use the mandatory isolated worker; no direct fallback exists. The required execution contract is mandatory isolation, with no product toggle or fallback to host execution. Chat read-only provider settings are defense in depth, not room filesystem, credential or network boundaries. A source fork CAN bypass that contract, BUT doing so exposes the OS account's readable files, credentials and network access to compromised agent behavior. See docs/architecture.md#mandatory-container-isolation-design-contract.
  Declaration: `struct { endpoint, executionKey string development bool }`.
- **`(codexAdapter).Capabilities`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() room.Capabilities`.
- **`(codexAdapter).Run`** — Run starts or resumes a model turn through the isolated worker; every turn uses it and fails closed if containment cannot be established. Existing external session IDs do not confer permission to run on the host.
  Signature: `func(ctx context.Context, thread, prompt string, onThread func(string) error) (string, error)`.
  Calls: `append`, `bufio.NewScanner`, `cleanup`, `cmd.Process.Kill`, `cmd.Start`, `cmd.StdoutPipe`, `cmd.Wait`, `codexBackgroundStage`, `ctx.Err`, `emitBackground`, `errors.New`, `fmt.Errorf`, `isolatedAgentCommand`, `json.Unmarshal`, `len`, `make`, `onThread`, `providerTurnArgs`, `scanner.Buffer`, `scanner.Bytes`, `scanner.Err`, `scanner.Scan`, `strings.NewReader`, `strings.TrimSpace`.
- **`(*server).agents`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `cancel`, `chatPrompt`, `check.Bytes`, `context.Len`, `context.String`, `context.WithTimeout`, `context.WriteString`, `detachedInitiatorContext`, `e.Error`, `exec.CommandContext`, `exec.CommandContext(ctx, "codex", "login", "status").CombinedOutput`, `exec.LookPath`, `hasRoom`, `historyRPC`, `http.Error`, `http.MaxBytesReader`, `id`, `isolationAvailable`, `json.NewDecoder`, `json.NewDecoder(r.Body).Decode`, `json.Unmarshal`, `len`, `member`, `operationError`, `os.Getwd`, `personalName`, `prepareOperation`, `r.Clone`, `r.Context`, `r.WithContext`, `reply`, `requestActor`, `requestedRoom`, `room.SupportsAgent`, `s.agentAdmission`, `s.agentsBlockedLocked`, `s.append`, `s.appendBatch`, `s.integrityLocked`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.runAgent`, `s.validCSRF`, `scoped`, `string`, `strings.Contains`, `strings.TrimSpace`, `time.Now`, `time.Now().UTC`, `w.WriteHeader`, `withOperationResponse`.
- **Type `bufferResponse`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { header http.Header body *bytes.Buffer status int }`.
- **`(*bufferResponse).Header`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() http.Header`.
- **`(*bufferResponse).WriteHeader`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(n int)`.
- **`(*bufferResponse).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v []byte) (int, error)`.
  Calls: `b.body.Write`.
- **`(*server).runAgent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, cancel context.CancelFunc, p room.Participant, prompt, endpoint string)`.
  Calls: `cancel`, `s.executeAgentTurn`, `s.finishAgentTurn`.
- **`(*server).executeAgentTurn`** — executeAgentTurn is shared by chat and workflows. It always uses the isolated worker. A source-level host fallback CAN bypass containment, BUT would expose host secrets and invalidate the mandatory execution boundary.
  Signature: `func(ctx context.Context, p room.Participant, prompt, endpoint string) (text string, err error)`.
  Calls: `adapter.Run`, `agentExecutionSession`, `controllerDevelopmentRepository`, `ctx.Err`, `errors.New`, `finish`, `nativeContextThread`, `os.Getenv`, `requireCredentialSeparatedDevelopment`, `s.agentsBlockedLocked`, `s.append`, `s.backgroundContext`, `s.load`, `s.managedJobContext`, `s.mu.Lock`, `s.mu.Unlock`, `strings.HasPrefix`, `strings.HasSuffix`.
- **`(*server).finishAgentTurn`** — finishAgentTurn discards cancelled results even if dispatch has been resumed.
  Signature: `func(ctx context.Context, p room.Participant, text, endpoint string, err error)`.
  Calls: `s.finishAgentTurnWithTask`.
- **`(*server).finishAgentTurnWithTask`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, p room.Participant, text, endpoint string, err error, taskID, dispatchID string)`.
  Calls: `append`, `context.Background`, `ctx.Err`, `delete`, `err.Error`, `id`, `providerForEndpoint`, `s.agentsBlockedLocked`, `s.appendBatch`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `taskOutcome`, `time.Now`, `time.Now().UTC`.
- **Type `sessionAdapter`**: sessionAdapter abstracts model turns, not a security boundary. Implementations must use mandatory container execution; exposing an unrestricted Run implementation must not become an alternate dispatch path.
  Declaration: `interface { Run(context.Context, string, string, func(string) error) (string, error) }`.
- **Type `claudeAdapter`**: claudeAdapter invokes Claude through the mandatory isolated worker. Chat disables tools; approved development workspaces enable fixed built-in tools under the same container boundary. MCP remains disabled. A source fork CAN remove isolation, BUT doing so exposes account files/credentials and permits host-side effects if provider safeguards are bypassed. No supported flag may enable that bypass.
  Declaration: `struct { endpoint, executionKey string development bool }`.
- **`(claudeAdapter).Capabilities`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() room.Capabilities`.
- **`(claudeAdapter).Run`** — Run starts a Claude turn with controlled egress, private session storage and resource limits before launch; isolation failure must block execution rather than fall back to this launcher.
  Signature: `func(ctx context.Context, thread, prompt string, onThread func(string) error) (string, error)`.
  Calls: `append`, `cleanup`, `cmd.Run`, `ctx.Err`, `errors.New`, `fmt.Errorf`, `isolatedAgentCommand`, `json.Unmarshal`, `len`, `onThread`, `output.Bytes`, `providerTurnArgs`, `strings.NewReader`, `strings.TrimSpace`.
