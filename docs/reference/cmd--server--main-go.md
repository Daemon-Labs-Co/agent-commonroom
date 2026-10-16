# cmd/server/main.go

[Module index](cmd.md) · [Open source](../../cmd/server/main.go)

Go implementation for main; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`main`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `acquireControllerLease`, `agentSafetyUIEnabled`, `agentWorkerMain`, `cancelStartup`, `connectionEditorEnabled`, `connectionEditorHandler`, `context.Background`, `context.WithTimeout`, `controllerLease.Close`, `db.Close`, `documentationHandler`, `filepath.Dir`, `filepath.Join`, `http.Dir`, `http.FileServer`, `http.HandlerFunc`, `http.NewServeMux`, `len`, `loadControllerDevelopmentConfiguration`, `loadLocalConfiguration`, `loadOwnerKey`, `log.Fatal`, `log.Print`, `log.Printf`, `mux.Handle`, `mux.HandleFunc`, `newHTTPServer`, `newHTTPServer(addr, h).ListenAndServe`, `newOwnerAuth`, `optionalScriptHandler`, `os.Exit`, `os.Getenv`, `os.Setenv`, `parseDailyTurnLimit`, `persistentHumanSessionsEnabled`, `pgxpool.NewWithConfig`, `runDatabaseMigration`, `runtimeDatabaseConfig`, `s.actorGuard`, `s.agentSafetyScript`, `s.auth.configureSessionPersistence`, `s.auth.prepareRSA`, `s.auth.protect`, `s.auth.protect(s.actorGuard(mux)).ServeHTTP`, `s.checkIntegrityNow`, `s.restoreIntegrityStop`, `s.roomPresentationScript`, `s.verifyWorkerAtStartup`, `s.watchIntegrity`, `strings.HasPrefix`, `verifyRuntimeDatabaseStartup`, `w.Header`, `w.Header().Set`, `workflowTemplatesEnabled`.
