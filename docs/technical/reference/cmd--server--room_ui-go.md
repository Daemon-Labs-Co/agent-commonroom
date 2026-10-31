# cmd/server/room_ui.go

[Module index](cmd.md) · [Open source](../../../cmd/server/room_ui.go)

Go implementation for room ui; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).roomPresentationScript`** — roomPresentationScript enforces optional cosmetic modules at delivery time. This cannot change profile authorization, room access or immutable history.
  Signature: `func(root, file, flag string) http.Handler`.
  Calls: `cancel`, `context.WithTimeout`, `http.Error`, `http.HandlerFunc`, `optionalScriptHandler`, `optionalScriptHandler((!exists && defaultEnabled) || (exists && enabled), root, file).ServeHTTP`, `r.Context`, `s.load`, `w.Header`, `w.Header().Set`.
