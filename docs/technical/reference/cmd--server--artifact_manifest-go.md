# cmd/server/artifact_manifest.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_manifest.go)

Go implementation for artifact manifest; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).artifactManifestClaims`** — artifactManifestClaims selects exact sealed receipts from replayed controller custody, never provider/model claims. It grants no byte/source authority; every read still requires the trusted store's current journal and physical guards.
  Signature: `func(ctx context.Context, scope objectcustody.ExecutionScope, output string) (taskcontract.Definition, []artifactmanifest.Claim, error)`.
  Calls: `append`, `artifactActor`, `ctx.Err`, `ctx.Value`, `len`, `objectcustody.ValidExecutionScope`, `refuse`, `s.agentsBlockedLocked`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `sort.Slice`, `taskcontract.Digest`, `validateArtifactTransition`.
- **`(*server).resolveStoredArtifactManifest`** — resolveStoredArtifactManifest joins the durable controller's receipt selection with actual private stored bytes and finite structural validation. Trusted assembly supplies a real isolated store/journal; no default/permissive adapter, path selection, public download, task completion or acceptance is introduced. Never hold server locks across Store.Get: its journal reacquires them. Recheck access, integrity and the COMPLETE receipt set after all member acknowledgments. Matching metadata is not a writer fence or independent source/custody proof.
  Signature: `func(ctx context.Context, scope objectcustody.ExecutionScope, output string, store *objectcustody.Store) (artifactmanifest.Manifest, error)`.
  Calls: `artifactmanifest.Resolve`, `cancel`, `ctx.Err`, `reflect.DeepEqual`, `refuse`, `s.artifactManifestClaims`, `s.storageContext`.
