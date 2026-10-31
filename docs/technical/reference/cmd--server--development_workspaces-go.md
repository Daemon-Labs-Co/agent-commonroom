# cmd/server/development_workspaces.go

[Module index](cmd.md) · [Open source](../../../cmd/server/development_workspaces.go)

Go implementation for development workspaces; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `developmentRepository`**: developmentRepository is administrator-owned, never accepted from an event or prompt. Only this bounded immutable bundle is mounted, read-only; an agent gets a separate writable clone. CAN replacing this with a host worktree expose secrets and permit destructive host edits? Yes, BUT it breaks this contract.
  Declaration: `struct { Bundle string 'json:"bundle"' SHA256 string 'json:"sha256"' }`.
- **`controllerDevelopmentRepository`** — controllerDevelopmentRepository is the server-enforced off-by-default feature flag. Enabling it also requires an explicit room-to-snapshot allowlist. Only a deployment administrator can change these environment settings; room members cannot request paths, turn tools on, or grant themselves repository access.
  Signature: `func(roomID, enabled, registry string) (string, error)`.
  Calls: `errors.New`, `json.Unmarshal`, `repositoryDigestPattern.MatchString`, `workerAliasPattern.MatchString`.
- **`providerTurnArgs`** — providerTurnArgs defines three exact policies: tool-free chat, development with native in-container tools (the SEC-04 credential-colocation gap), and development with toolSeparation, where file tools are routed through the separate credential-free tool container via the fixed commonroom-tools MCP bridge instead of running natively here. toolSeparation is ignored unless development is also true: chat already has zero tools, nothing to broker. The development policy uses the mandatory outer Docker sandbox as its execution backend, avoiding nested bwrap user namespaces, and must NEVER be used by a host launcher. Credential-management login has a separate policy. Codex's mcp_servers config path below is a design inference from vendor config syntax, not verified against a live Codex install (not available on this development machine) — see docs/technical/archive/provider-tool-separation.md. Claude's --tools/--mcp-config/--strict-mcp-config combination was checked directly against the installed 'claude --help' (2.1.292): "" disables every built-in tool, and --mcp-config accepts a full stdio server definition that --strict-mcp-config makes exclusive.
  Signature: `func(provider string, development, toolSeparation bool) []string`.
  Calls: `append`.
- **`developmentTurnArgs`** — developmentTurnArgs recognizes either exact development policy prefix (toolSeparation on or off) without needing to know which variant a caller intended — callers that only need "is this a development turn at all" (localIsolatedDevelopmentCommand's own consistency check) can use this alone. A caller that also needs to know which variant, such as remoteAgentCommand building the outbound request, uses toolSeparationTurnArgs alongside it.
  Signature: `func(provider string, args []string) bool`.
  Calls: `len`, `providerTurnArgs`, `slices.Equal`.
- **`toolSeparationTurnArgs`** — toolSeparationTurnArgs reports whether args match the tool-separated development prefix specifically, distinguishing it from the native-tool development variant that developmentTurnArgs alone treats the same way.
  Signature: `func(provider string, args []string) bool`.
  Calls: `len`, `providerTurnArgs`, `slices.Equal`.
- **`validateDevelopmentRegistry`** — validateDevelopmentRegistry bounds the private worker registry and disallows paths whose spelling could change Docker mount-option interpretation.
  Signature: `func(repositories map[string]developmentRepository) error`.
  Calls: `errors.New`, `filepath.IsAbs`, `len`, `repositoryDigestPattern.MatchString`, `strings.ContainsAny`, `workerAliasPattern.MatchString`.
- **`workerDevelopmentBundle`** — workerDevelopmentBundle independently enforces the worker administrator's feature flag, room mapping, snapshot identity and file permissions. A forged controller request cannot choose another room's bundle or an arbitrary mount.
  Signature: `func(key, digest, enabled, registry string) (string, error)`.
  Calls: `errors.New`, `file.Close`, `filepath.EvalSymlinks`, `hash.Sum`, `hex.EncodeToString`, `io.Copy`, `io.LimitReader`, `json.Unmarshal`, `openDevelopmentBundle`, `sha256.New`, `strings.SplitN`, `validateDevelopmentRegistry`.
- **`openDevelopmentBundle`** — openDevelopmentBundle validates the actual opened inode, never a prior stat. Nonblocking/no-follow avoids FIFO hangs and final-component substitution. A bundle may be administrator-owned/readable; unlike credential files it need not have mode 0600. The checksum is still mandatory, including during staging.
  Signature: `func(path string) (*os.File, error)`.
  Calls: `errors.New`, `f.Close`, `f.Stat`, `info.Mode`, `info.Mode().IsRegular`, `info.Mode().Perm`, `info.Size`, `os.OpenFile`.
- **`stageDevelopmentBundle`** — stageDevelopmentBundle copies and hashes the same bounded descriptor into a private per-turn directory. Docker mounts this checked copy, never the source path that an operator could replace after verification. A fork CAN skip this, BUT risks granting a turn an unapproved snapshot. Parent directory is trusted.
  Signature: `func(source, digest, dir string) (string, error)`.
  Calls: `errors.New`, `f.Close`, `filepath.Join`, `hash.Sum`, `hex.EncodeToString`, `io.Copy`, `io.LimitReader`, `io.MultiWriter`, `openDevelopmentBundle`, `os.OpenFile`, `os.Remove`, `out.Chmod`, `out.Close`, `sha256.New`.
- **`developmentRunArgs`** — developmentRunArgs adds only the prevalidated snapshot and fixed bootstrap settings. Network, capabilities, read-only root and all resource limits remain intact. The snapshot is never a writable host mount.
  Signature: `func(args []string, bundle, digest, volume string) []string`.
  Calls: `append`, `len`, `slices.Clone`.
- **`agentExecutionSession`** — agentExecutionSession never resumes a provider session across workspace mode or snapshot changes. Existing room history is supplied separately in prompts. Legacy host sessions remain rejected rather than acquiring container authority.
  Signature: `func(previous, thread, repository string) (string, string, error)`.
  Calls: `errors.New`, `strings.HasPrefix`.
- **`loadControllerDevelopmentConfiguration`** — loadControllerDevelopmentConfiguration preserves operator setup across normal restarts without copying the process environment or credentials. Explicit env values override this dedicated owner-only, non-secret configuration file.
  Signature: `func(path, enabled, registry string) (string, string, error)`.
  Calls: `controllerDevelopmentRepository`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `errors.New`, `json.Marshal`, `json.NewDecoder`, `len`, `new`, `os.IsNotExist`, `readPrivateBounded`, `string`, `strings.NewReader`.
