# cmd/server/room_ui_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/room_ui_test.go)

Automated regression checks for room_ui; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRoomPresentationFlags`** — Regression contract: TestRoomPresentationFlags.
  Signature: `func(t *testing.T)`.
  Calls: `check`, `db.seed`, `filepath.Join`, `handler.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `mockRequest`, `mockServer`, `os.WriteFile`, `s.roomPresentationScript`, `t.Fatal`, `t.Helper`, `t.Parallel`, `t.Run`, `t.TempDir`, `time.Now`, `time.Now().Add`, `w.Body.String`, `w.Header`, `w.Header().Get`.
