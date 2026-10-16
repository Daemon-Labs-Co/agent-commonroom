# cmd/server/private_files_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/private_files_test.go)

Automated regression checks for private_files; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestPrivateFileRejectsUnsafeTypesAndPermissions`** — Regression contract: TestPrivateFileRejectsUnsafeTypesAndPermissions.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Repeat`, `filepath.Join`, `os.Link`, `os.Mkdir`, `os.Symlink`, `os.WriteFile`, `readPrivateBounded`, `syscall.Mkfifo`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestPrivateDescriptorPinsOpenedInode`** — Regression contract: TestPrivateDescriptorPinsOpenedInode.
  Signature: `func(t *testing.T)`.
  Calls: `f.Close`, `filepath.Join`, `io.ReadAll`, `openPrivateRegular`, `os.Rename`, `os.Symlink`, `os.WriteFile`, `readPrivateBounded`, `string`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestPrivateLocksRejectAliasedInodes`** — Regression contract: TestPrivateLocksRejectAliasedInodes.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `lease.Close`, `lockWorkerCredential`, `os.Link`, `os.Mkdir`, `os.Symlink`, `other.Close`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `unlock`, `workerTryLease`.
- **`TestPrivateFileBoundsAndMissingFileSemantics`** — Regression contract: TestPrivateFileBoundsAndMissingFileSemantics.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Repeat`, `filepath.Join`, `len`, `os.IsNotExist`, `os.WriteFile`, `readPrivateBounded`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestPrivateCredentialAndRecordRejectLinks`** — Regression contract: TestPrivateCredentialAndRecordRejectLinks.
  Signature: `func(t *testing.T)`.
  Calls: `credentialPresent`, `filepath.Join`, `os.Symlink`, `os.WriteFile`, `persistWorkerCredential`, `readWorkerRecord`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestPrivateOpenRejectsDestructiveFlagsBeforeTouchingFile`** — Regression contract: TestPrivateOpenRejectsDestructiveFlagsBeforeTouchingFile.
  Signature: `func(t *testing.T)`.
  Calls: `f.Close`, `filepath.Join`, `openPrivateRegular`, `os.ReadFile`, `os.WriteFile`, `string`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
