# cmd/server/auth_availability_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/auth_availability_test.go)

Automated regression checks for auth_availability; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`loginAvailability`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *ownerAuth)`.
  Calls: `newOwnerAuth`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`badLoginFlood`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *server)`.
  Calls: `httptest.NewRecorder`, `httptest.NewRequest`, `s.login`, `strings.NewReader`.
- **`TestValidBootstrapLoginSurvivesFailedAttemptFlood`** — Regression contract: TestValidBootstrapLoginSurvivesFailedAttemptFlood.
  Signature: `func(t *testing.T)`.
  Calls: `badLoginFlood`, `httptest.NewRecorder`, `httptest.NewRequest`, `loginAvailability`, `s.login`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `w.Body.String`, `w.Header`, `w.Header().Get`.
- **`TestValidSignaturesSurviveFailedAttemptFlood`** — Regression contract: TestValidSignaturesSurviveFailedAttemptFlood.
  Signature: `func(t *testing.T)`.
  Calls: `badLoginFlood`, `base64.StdEncoding.EncodeToString`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Marshal`, `loginAvailability`, `newMockDB`, `pem.EncodeToMemory`, `rsa.GenerateKey`, `rsa.SignPSS`, `s.db.(*mockDB).seed`, `s.login`, `sha256.Sum256`, `string`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `w.Body.String`, `x509.MarshalPKIXPublicKey`.
- **`TestSessionRenewalAtCapacityAndFailureRollback`** — Regression contract: TestSessionRenewalAtCapacityAndFailureRollback.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `fmt.Sprintf`, `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `loginAvailability`, `os.Mkdir`, `r.AddCookie`, `s.login`, `sessionDigest`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `time.Now`, `time.Now().Add`, `w.Result`, `w.Result().Cookies`.
- **`TestUnknownChallengeDoesNotReadMemberDatabase`** — Regression contract: TestUnknownChallengeDoesNotReadMemberDatabase.
  Signature: `func(t *testing.T)`.
  Calls: `httptest.NewRecorder`, `httptest.NewRequest`, `loginAvailability`, `s.login`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
- **`TestMalformedMemberProofConsumesChallengeBeforeDatabaseAccess`** — Regression contract: TestMalformedMemberProofConsumesChallengeBeforeDatabaseAccess.
  Signature: `func(t *testing.T)`.
  Calls: `httptest.NewRecorder`, `httptest.NewRequest`, `loginAvailability`, `s.login`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
