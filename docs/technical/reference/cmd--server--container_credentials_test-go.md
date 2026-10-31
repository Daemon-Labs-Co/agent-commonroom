# cmd/server/container_credentials_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/container_credentials_test.go)

Automated regression checks for container_credentials; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestCredentialStagingRefusesExistingAndAliasedTargets`** — Regression contract: TestCredentialStagingRefusesExistingAndAliasedTargets.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `os.Link`, `os.ReadFile`, `os.Symlink`, `os.WriteFile`, `stageContainerCredential`, `string`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestCredentialStagingPrivacySizeAndPermissions`** — Regression contract: TestCredentialStagingPrivacySizeAndPermissions.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `info.Mode`, `info.Mode().Perm`, `os.Chmod`, `os.IsNotExist`, `os.Lstat`, `os.Stat`, `os.Symlink`, `stageContainerCredential`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
