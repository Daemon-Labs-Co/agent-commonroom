# cmd/server/artifact_manifest_record.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_manifest_record.go)

Go implementation for artifact manifest record; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).recordStoredArtifactManifest`** — recordStoredArtifactManifest privately assembles actual bytes then durably records the immutable manifest once for this execution/output. No client/model receipt payload, task completion, acceptance or byte download is introduced. Ambiguous append outcomes never return success or automatically retry/adopt.
  Signature: `func(ctx context.Context, scope objectcustody.ExecutionScope, output string, store *objectcustody.Store) (artifactmanifest.Manifest, error)`.
  Calls: `artifactActor`, `cancel`, `context.WithValue`, `ctx.Err`, `ctx.Value`, `errors.Is`, `refuse`, `s.agentsBlockedLocked`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.resolveStoredArtifactManifest`, `s.storageContext`, `s.tripIntegrityLocked`, `validateArtifactManifestRecord`.
