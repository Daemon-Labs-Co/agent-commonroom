# cmd/server/local_configuration_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/local_configuration_test.go)

Automated regression checks for local_configuration; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestLocalConfigurationStrictDecoding`** — Regression contract: TestLocalConfigurationStrictDecoding.
  Signature: `func(t *testing.T)`.
  Calls: `err.Error`, `filepath.Join`, `loadLocalConfiguration`, `os.WriteFile`, `strings.Contains`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.TempDir`.
- **`TestControllerConfigurationFileBoundaries`** — Regression contract: TestControllerConfigurationFileBoundaries.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `loadControllerDevelopmentConfiguration`, `loadLocalConfiguration`, `os.Chmod`, `os.Link`, `os.Symlink`, `os.WriteFile`, `strings.Repeat`, `syscall.Mkfifo`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
