# cmd/server/artifact_manifest_registry.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_manifest_registry.go)

Go implementation for artifact manifest registry; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `artifactManifestEvent`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' ActorID string 'json:"actorId"' Manifest artifactmanifest.Manifest 'json:"manifest"' PerformedBy string 'json:"performedBy"' Audit *auditRecordMetadata 'json:"_auditRecord,omitempty"' }`.
- **`artifactManifestKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(m artifactmanifest.Manifest) string`.
- **`validateArtifactManifestRecord`** — Replay validates immutable linkage/shape only, without filesystem/network I/O. Only trusted actual-byte assembly may append this event; public client events cannot create it. A recorded manifest is not fresh export/acceptance authority.
  Signature: `func(st room.State, v artifactManifestEvent) error`.
  Calls: `artifactManifestKey`, `validateArtifactManifestLinkage`.
- **`validateArtifactManifestLinkage`** — Linkage is historical metadata validation, never current byte/export proof.
  Signature: `func(st room.State, v artifactManifestEvent) error`.
  Calls: `artifactRecordedActor`, `artifactmanifest.ValidateMetadata`, `len`, `validateArtifactTransition`.
- **`projectArtifactManifest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `append`, `artifactManifestKey`, `bytes.NewReader`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `json.NewDecoder`, `taskcontract.ValidJSON`, `validateArtifactManifestRecord`.
