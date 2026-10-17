# cmd/server/isolation.go

[Module index](cmd.md) · [Open source](../../cmd/server/isolation.go)

Go implementation for isolation; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `limitedAgentOutput`**: limitedAgentOutput bounds captured stdout/stderr without disclosing secrets.
  Declaration: `struct{ bytes.Buffer }`.
- **`(*limitedAgentOutput).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `b.Buffer.Write`, `b.Len`, `errors.New`, `len`.
- **`isolationAvailable`** — isolationAvailable rejects unsupported hosts/configuration. There is no flag or direct-launch fallback. A fork CAN bypass this, BUT exposes host credentials, files and network access and invalidates the containment contract.
  Signature: `func(endpoint string) error`.
  Calls: `errors.New`, `exec.LookPath`, `os.Getenv`, `remoteWorkerEndpoint`, `strings.HasPrefix`, `validAgentImage`.
- **`verifiedLocalDockerContext`** — verifiedLocalDockerContext rejects a daemon that cannot enforce required limits and any non-local or environment-overridden Docker context. Shared by every caller that is about to shell out to "docker" for isolated execution (the single-container agent path here and the tool-container runner in tool_broker_runner.go) so this check is defined once, not duplicated.
  Signature: `func(ctx context.Context) error`.
  Calls: `cancel`, `context.WithTimeout`, `dockerSecurityReady`, `errors.New`, `exec.CommandContext`, `exec.CommandContext(checkCtx, "docker", "context", "inspect", "--format", "{{.Endpoints.docker.Host}}").Output`, `exec.CommandContext(checkCtx, "docker", "info", "--format", "{{json .}}").Output`, `os.Getenv`, `string`, `strings.HasPrefix`, `strings.TrimSpace`.
- **`validAgentImage`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value string) bool`.
  Calls: `hex.DecodeString`, `len`, `strings.ContainsAny`, `strings.HasPrefix`, `strings.Split`.
- **`dockerSecurityReady`** — dockerSecurityReady fails closed when the daemon cannot enforce required limits.
  Signature: `func(raw []byte) bool`.
  Calls: `json.Unmarshal`, `strings.Contains`, `strings.Join`.
- **`isolatedRunArgs`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(image, name, volume, socket, credential, provider string, args []string) []string`.
  Calls: `append`.
- **`isolatedAgentCommand`** — isolatedAgentCommand prepares a network-less container and a per-turn Unix CONNECT gateway. Only a single copied provider credential is mounted. The private named home volume belongs to this membership/connection, not a host directory. Missing prerequisites fail closed. Trusted Docker images/runtime and the worker host remain part of the security boundary.
  Signature: `func(ctx context.Context, endpoint, key string, args ...string) (*exec.Cmd, func(), error)`.
  Calls: `ctx.Value`, `errors.New`, `localIsolatedAgentCommand`, `remoteAgentCommand`, `strings.HasPrefix`, `validWorkerJobID`.
- **`localIsolatedAgentCommand`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, endpoint, key, jobID string, args ...string) (*exec.Cmd, func(), error)`.
  Calls: `localIsolatedDevelopmentCommand`.
- **`localIsolatedDevelopmentCommand`** — Development changes only the private workspace and provider tool policy; the rootless container boundary remains mandatory. A source bypass CAN expose host authority, BUT invalidates containment and must never be an automatic fallback.
  Signature: `func(ctx context.Context, endpoint, key, jobID, repository string, toolSeparation bool, args ...string) (*exec.Cmd, func(), error)`.
  Calls: `cleanupDir`, `connectionAuthKey`, `context.Background`, `context.WithTimeout`, `developmentRunArgs`, `developmentTurnArgs`, `errors.New`, `exec.CommandContext`, `exec.CommandContext(cleanupCtx, "docker", "rm", "--force", name).Run`, `executionConnectionID`, `filepath.Dir`, `filepath.Join`, `fmt.Fprintln`, `gateway.Close`, `hex.EncodeToString`, `id`, `isolatedRunArgs`, `isolationAvailable`, `json.Valid`, `len`, `newProviderGateway`, `once.Do`, `os.Getenv`, `os.MkdirTemp`, `os.RemoveAll`, `readPrivateBounded`, `requireCredentialSeparatedDevelopment`, `sha256.Sum256`, `stageContainerCredential`, `stageDevelopmentBundle`, `stop`, `strings.TrimSpace`, `toolSeparationTurnArgs`, `validWorkerJobID`, `verifiedLocalDockerContext`, `workerCredentialReady`, `workerDevelopmentBundle`.
- **Type `providerGateway`**: providerGateway permits HTTPS tunnels to exact provider hosts only. DNS results are checked and the selected public IP is dialed directly to resist rebinding. TLS stays end-to-end; an allowed provider can still receive supplied context.
  Declaration: `struct { socket string server *http.Server listener net.Listener mu sync.Mutex connections map[net.Conn]bool closed bool slots chan struct{} ctx context.Context cancel context.CancelFunc lookupIP func(context.Context, string, string) ([]net.IP, error) }`.
- **`providerHostAllowed`** — Claude subscription and Console OAuth both exchange tokens at platform.claude.com. Permit exact authentication hosts only; never broaden this to a wildcard or host fallback.
  Signature: `func(provider, host string) bool`.
- **`publicProviderIP`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ip net.IP) bool`.
  Calls: `ip.IsGlobalUnicast`, `ip.IsLinkLocalUnicast`, `ip.IsLoopback`, `ip.IsPrivate`, `ip.IsUnspecified`, `ip.To4`, `native.Contains`, `net.ParseCIDR`, `reserved.Contains`.
- **`newProviderGateway`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, socket, provider string) (*providerGateway, error)`.
  Calls: `context.WithCancel`, `g.connect`, `g.server.Serve`, `http.HandlerFunc`, `l.Close`, `make`, `net.Listen`, `os.Chmod`.
- **`(*providerGateway).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `c.Close`, `g.cancel`, `g.mu.Lock`, `g.mu.Unlock`, `g.server.Close`.
- **`(*providerGateway).connect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request, provider string)`.
  Calls: `(&net.Dialer{Timeout: 10 * time.Second}).DialContext`, `buffer.Flush`, `buffer.WriteString`, `cancel`, `client.Close`, `client.SetDeadline`, `close`, `context.WithTimeout`, `delete`, `g.ctx.Err`, `g.mu.Lock`, `g.mu.Unlock`, `http.Error`, `io.Copy`, `io.LimitReader`, `ips[0].String`, `len`, `lookup`, `make`, `net.JoinHostPort`, `net.SplitHostPort`, `providerHostAllowed`, `publicProviderIP`, `time.Now`, `time.Now().Add`, `upstream.Close`, `upstream.SetDeadline`, `w.(http.Hijacker).Hijack`.
