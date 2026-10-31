# cmd/server/members_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/members_test.go)

Automated regression checks for members; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestPersonalRoomsAndVisibility`** — Regression contract: TestPersonalRoomsAndVisibility.
  Signature: `func(t *testing.T)`.
  Calls: `canAccess`, `hasRoom`, `len`, `projectActors`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `visibleState`.
- **`TestInvitationAuthorizationWithMockDB`** — Regression contract: TestInvitationAuthorizationWithMockDB.
  Signature: `func(t *testing.T)`.
  Calls: `a.protect`, `base64.StdEncoding.EncodeToString`, `call`, `context.Background`, `db.snapshot`, `h.ServeHTTP`, `hasRoom`, `http.NewServeMux`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Marshal`, `json.Unmarshal`, `member`, `mux.HandleFunc`, `newMockDB`, `newOwnerAuth`, `pem.EncodeToMemory`, `r.AddCookie`, `r.Header.Set`, `rsa.GenerateKey`, `rsa.SignPSS`, `s.actorGuard`, `s.load`, `sha256.Sum256`, `string`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.Bytes`, `w.Body.String`, `x509.MarshalPKIXPublicKey`.
- **`TestPersonalLabelRefreshPreservesRoomSettingsAndActorAuthority`** — Regression contract: TestPersonalLabelRefreshPreservesRoomSettingsAndActorAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `canAccess`, `member`, `personalRoomNameVersion`, `projectActors`, `reflect.DeepEqual`, `t.Fatal`, `t.Parallel`.
