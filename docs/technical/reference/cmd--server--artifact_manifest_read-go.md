# cmd/server/artifact_manifest_read.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_manifest_read.go)

Go implementation for artifact manifest read; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).recordedManifestSnapshot`** — recordedManifestSnapshot is private authenticated setup, not a public listing. Exact digest pins the version reviewed by the caller. Metadata never proves current bytes; actual release MUST still use the approved isolated store guard.
  Signature: `func(ctx context.Context, execution, output, digest string) (taskcontract.Definition, artifactmanifest.Manifest, error)`.
  Calls: `append`, `artifactActor`, `ctx.Err`, `ctx.Value`, `membershipPartPattern.MatchString`, `refuse`, `repositoryDigestPattern.MatchString`, `s.agentsBlockedLocked`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `validWorkerJobID`, `validateArtifactManifestLinkage`.
- **`(*server).readRecordedArtifactMember`** — readRecordedArtifactMember revalidates ALL finite members before returning one exact declared member's actual sealed bytes. No caller-supplied path/receipt, permissive source, host fallback or public route exists. Trusted assembly must supply actual isolated custody with independent physical/export/writer fences. Never hold server locks across Store.Get; its journal owns current authorization. A manifest digest or prior audit record alone can NEVER authorize delivery.
  Signature: `func(ctx context.Context, execution, output, digest, member string, store *objectcustody.Store) ([]byte, error)`.
  Calls: `artifactmanifest.Revalidate`, `cancel`, `clear`, `ctx.Err`, `reflect.DeepEqual`, `refuse`, `s.recordedManifestSnapshot`, `s.storageContext`, `store.Get`.
