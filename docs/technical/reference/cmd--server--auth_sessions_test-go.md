# cmd/server/auth_sessions_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/auth_sessions_test.go)

Automated regression checks for auth_sessions; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestHumanSessionSurvivesReloadAndRestart`** — Regression contract: TestHumanSessionSurvivesReloadAndRestart.
  Signature: `func(t *testing.T)`.
  Calls: `a.configureSessionPersistence`, `a.session`, `filepath.Join`, `fresh.configureSessionPersistence`, `fresh.session`, `httptest.NewRecorder`, `httptest.NewRequest`, `info.Mode`, `info.Mode().Perm`, `newOwnerAuth`, `os.ReadFile`, `os.Stat`, `request.AddCookie`, `request.Header.Set`, `resumed.Expires.Equal`, `s.login`, `s.logout`, `string`, `strings.Contains`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `third.configureSessionPersistence`, `third.session`, `w.Body.String`, `w.Result`, `w.Result().Cookies`.
- **`TestPrivateSessionStoreFailuresAndDisabledState`** — Regression contract: TestPrivateSessionStoreFailuresAndDisabledState.
  Signature: `func(t *testing.T)`.
  Calls: `a.configureSessionPersistence`, `a.persistSessionsLocked`, `filepath.Join`, `fresh.configureSessionPersistence`, `len`, `newOwnerAuth`, `os.IsNotExist`, `os.Stat`, `os.Symlink`, `os.WriteFile`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `time.Now`, `time.Now().Add`.
- **`TestSessionExpiryIdentityAndWriteFailure`** — Regression contract: TestSessionExpiryIdentityAndWriteFailure.
  Signature: `func(t *testing.T)`.
  Calls: `a.configureSessionPersistence`, `a.persistSessionsLocked`, `a.session`, `filepath.Join`, `fresh.configureSessionPersistence`, `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `newOwnerAuth`, `r.AddCookie`, `r.Header.Set`, `rotated.configureSessionPersistence`, `s.logout`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `time.Now`, `time.Now().Add`.
- **`TestPersistentSessionFlagParsing`** — Regression contract: TestPersistentSessionFlagParsing.
  Signature: `func(t *testing.T)`.
  Calls: `persistentHumanSessionsEnabled`, `t.Fatal`, `t.Run`, `t.Setenv`.
