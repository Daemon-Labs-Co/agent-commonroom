# cmd/server/worker_auth.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_auth.go)

Go implementation for worker auth; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerAuthRequest`**: Connection credentials are isolated from other connections. Agent execution remains disposable and separately scoped; no host CLI is used for sign-in.
  Declaration: `struct { Version int 'json:"version"' Connection string 'json:"connection"' Provider string 'json:"provider"' }`.
- **`connectionSignInEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(flags map[string]bool) bool`.
- **`connectionAuthKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(connection string) string`.
  Calls: `hex.EncodeToString`, `sha256.Sum256`.
- **`executionConnectionID`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(key string) (string, error)`.
  Calls: `errors.New`, `len`, `strings.SplitN`, `workerAliasPattern.MatchString`.
- **`encodeWorkerAuth`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v workerAuthRequest) (string, error)`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `errors.New`, `json.Marshal`, `workerAliasPattern.MatchString`.
- **`decodeWorkerAuth`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(encoded string) (workerAuthRequest, error)`.
  Calls: `base64.RawURLEncoding.DecodeString`, `d.Decode`, `d.DisallowUnknownFields`, `encodeWorkerAuth`, `errors.New`, `json.NewDecoder`, `len`, `new`, `string`, `strings.NewReader`.
- **`interactiveWorkerLogin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(alias, encoded string) string`.
  Calls: `shellArgument`.
- **`providerCredentialName`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(provider string) string`.
- **`clearWorkerCredential`** — Forget only this connection/provider's stored credential; provider-wide revocation is separate.
  Signature: `func(root, provider string) error`.
  Calls: `clearWorkerCredentialDurably`.
- **`clearWorkerCredentialDurably`** — clearWorkerCredentialDurably acknowledges local revocation only after the directory entry is synced. Missing-file retries must sync too: a previous unlink may have succeeded while its durability acknowledgement failed.
  Signature: `func(root, provider string, syncDirectory func(string) error) error`.
  Calls: `errors.New`, `filepath.Join`, `os.IsNotExist`, `os.Remove`, `providerCredentialName`, `syncDirectory`.
- **`lockWorkerCredential`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(root string) (func(), error)`.
  Calls: `errors.New`, `f.Close`, `f.Fd`, `filepath.Join`, `int`, `openPrivateRegular`, `os.MkdirAll`, `syscall.Flock`.
- **`credentialPresent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(path string) bool`.
  Calls: `filepath.Dir`, `json.Valid`, `len`, `readPrivateBounded`, `workerCredentialReady`.
- **`workerProviderAuth`** — workerProviderAuth permits only provider login/status in a constrained image. Login is a human administrative operation, bounded to fifteen minutes, with the same network gateway and resource limits as agent turns. There is no credential upload endpoint or arbitrary command/mount option.
  Signature: `func(args []string, in io.Reader, out, diagnostics io.Writer) int`.
  Calls: `append`, `cancel`, `clearWorkerCredential`, `clearWorkerCredentialRecovery`, `configureWorker`, `connectionAuthKey`, `context.Background`, `context.WithTimeout`, `create.Run`, `credentialPresent`, `decodeWorkerAuth`, `dockerSecurityReady`, `done`, `errors.New`, `exec.CommandContext`, `exec.CommandContext(cleanup, "docker", "rm", "--force", name).Run`, `exec.CommandContext(cleanup, "docker", "volume", "rm", volume, volume+"-work").Run`, `exec.CommandContext(ctx, "docker", "context", "inspect", "--format", "{{.Endpoints.docker.Host}}").Output`, `exec.CommandContext(ctx, "docker", "info", "--format", "{{json .}}").Output`, `export.Run`, `exported.Bytes`, `fail`, `filepath.Join`, `fmt.Fprintln`, `gateway.Close`, `id`, `isolatedRunArgs`, `isolationAvailable`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `lockWorkerCredential`, `markWorkerCredentialUncertain`, `newProviderGateway`, `os.Getenv`, `os.MkdirTemp`, `os.RemoveAll`, `os.WriteFile`, `persistWorkerCredential`, `providerCredentialName`, `run.Run`, `signal.NotifyContext`, `stop`, `string`, `strings.Contains`, `strings.HasPrefix`, `strings.TrimSpace`, `unlock`, `workerCredentialConnectionQuiescent`, `workerCredentialExportArgs`, `workerCredentialReady`, `workerStateDirectory`, `workerVerifyConnection`.
- **Type `credentialExportOutput`**: credentialExportOutput bounds sensitive output in memory; callers must never forward it to diagnostics. Overflow fails the export instead of truncating it.
  Declaration: `struct{ bytes.Buffer }`.
- **`(*credentialExportOutput).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `b.Buffer.Write`, `b.Len`, `len`.
- **`workerCredentialExportArgs`** — workerCredentialExportArgs reads one known credential from the connection's login volume in a non-root, network-disabled container. The volume is read-only; there are no host paths, provider invocation, gateway, or Docker socket mounts. Source forks CAN weaken these limits, BUT risk exposing every host credential.
  Signature: `func(image, volume, provider string) ([]string, error)`.
  Calls: `errors.New`.
- **`persistWorkerCredential`** — persistWorkerCredential validates only the exported provider file and replaces the connection's private credential atomically and syncs its parent. Invalid exports leave the prior login untouched; post-rename sync failure is ambiguous and must not be acknowledged as saved. No credential bytes enter diagnostics.
  Signature: `func(root, provider, source string) error`.
  Calls: `persistWorkerCredentialDurably`.
- **`persistWorkerCredentialDurably`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(root, provider, source string, syncDirectory func(string) error) error`.
  Calls: `clearWorkerCredentialRecovery`, `errors.New`, `filepath.Join`, `json.Valid`, `len`, `os.MkdirAll`, `providerCredentialName`, `readPrivateBounded`, `replacePrivateDurably`.
- **`workerVerifyConnection`** — workerVerifyConnection rejects the legacy unjournalled verification launcher. Controllers must use registered run-v2 jobs with verified cancellation instead.
  Signature: `func(v workerAuthRequest, out, diagnostics io.Writer) int`.
  Calls: `fmt.Fprintln`.
