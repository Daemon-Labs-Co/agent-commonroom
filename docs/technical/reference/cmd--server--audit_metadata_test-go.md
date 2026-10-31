# cmd/server/audit_metadata_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/audit_metadata_test.go)

Automated regression checks for audit_metadata; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestNewAuditMetadataDetectsIdentityOrderAndTimeTampering`** — Regression contract: TestNewAuditMetadataDetectsIdentityOrderAndTimeTampering.
  Signature: `func(t *testing.T)`.
  Calls: `at.Add`, `protectAuditMetadata`, `string`, `t.Fatal`, `t.Parallel`, `time.Date`, `verifyAuditMetadata`.
- **`TestNewAppendOverwritesClaimedActorAndChecksMetadata`** — Regression contract: TestNewAppendOverwritesClaimedActorAndChecksMetadata.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `context.WithValue`, `db.snapshot`, `entry.at.Add`, `json.Unmarshal`, `mockServer`, `s.append`, `s.checkIntegrityNow`, `string`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `verifyAuditMetadata`.
- **`TestProtectedAuditFormatCannotDowngradeToLegacy`** — Regression contract: TestProtectedAuditFormatCannotDowngradeToLegacy.
  Signature: `func(t *testing.T)`.
  Calls: `active.Err`, `cancel`, `context.Background`, `context.WithCancel`, `db.seedLegacy`, `json.Unmarshal`, `mockRequest`, `mockServer`, `s.append`, `t.Error`, `t.Fatal`, `t.Parallel`, `w.Body.Bytes`, `w.Body.String`.
- **`TestLegacyPrefixTransitionsToProtectedAuditWithoutRewriting`** — Regression contract: TestLegacyPrefixTransitionsToProtectedAuditWithoutRewriting.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `db.seedLegacy`, `db.snapshot`, `json.Unmarshal`, `mockRequest`, `mockServer`, `s.append`, `t.Fatal`, `t.Parallel`, `w.Body.Bytes`, `w.Body.String`.
