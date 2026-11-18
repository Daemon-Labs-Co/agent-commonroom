package main

// Executable entry point: configuration, dependency wiring and HTTP startup.

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"time"
)

func main() {
	if handled, code := startupCommand(os.Args[1:], os.Stdout, os.Stderr); handled {
		os.Exit(code)
	}
	if len(os.Args) > 1 && os.Args[1] == "--agent-worker" {
		os.Exit(agentWorkerMain(os.Args[2:], os.Stdin, os.Stdout, os.Stderr))
	}
	if len(os.Args) > 1 && os.Args[1] == "--migrate" {
		if err := runDatabaseMigration(context.Background()); err != nil {
			log.Fatal(err)
		}
		log.Print("Audit schema migration completed")
		return
	}
	controllerLease, leaseErr := acquireControllerLease(".data/controller.lock")
	if leaseErr != nil {
		log.Fatal("Controller startup refused: ", leaseErr)
	}
	defer controllerLease.Close()
	developmentEnabled, developmentRegistry, developmentErr := loadControllerDevelopmentConfiguration(".data/development-workspaces.json", os.Getenv("COMMONROOM_DEVELOPMENT_WORKSPACES"), os.Getenv("COMMONROOM_DEVELOPMENT_REPOSITORIES"))
	if developmentErr != nil {
		log.Fatal(developmentErr)
	}
	os.Setenv("COMMONROOM_DEVELOPMENT_WORKSPACES", developmentEnabled)
	os.Setenv("COMMONROOM_DEVELOPMENT_REPOSITORIES", developmentRegistry)
	dailyTurnLimit, budgetErr := parseDailyTurnLimit(os.Getenv("COMMONROOM_DAILY_TURN_LIMIT"))
	if budgetErr != nil {
		log.Fatal(budgetErr)
	}
	ctx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	localConfig, configErr := loadLocalConfiguration(".data/local-config.json")
	if configErr != nil {
		log.Fatal("Private local configuration unavailable: ", configErr)
	}
	if os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS") == "" && localConfig.RemoteWorkerAlias != "" {
		os.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", localConfig.RemoteWorkerAlias)
	}
	if os.Getenv("COMMONROOM_REMOTE_WORKER") == "" && localConfig.RemoteWorkerEnabled {
		os.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	}
	if flag := os.Getenv("COMMONROOM_REMOTE_WORKER"); flag != "" && flag != "true" && flag != "false" {
		log.Fatal("COMMONROOM_REMOTE_WORKER must be true or false")
	}
	if flag := os.Getenv("COMMONROOM_SERVER_WORKER"); flag != "" && flag != "true" && flag != "false" {
		log.Fatal("COMMONROOM_SERVER_WORKER must be true or false")
	}
	if os.Getenv("COMMONROOM_SERVER_WORKER") == "true" && os.Getenv("COMMONROOM_REMOTE_WORKER") == "true" {
		log.Fatal("Configure one worker transport: server or remote")
	}
	if err := serverEndpointAliasConfigError(); err != nil {
		log.Fatal(err)
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = localConfig.DatabaseURL
	}
	if url == "" {
		url = "postgres://localhost:5432/agentroom?sslmode=disable"
	}
	config, err := runtimeDatabaseConfig(url)
	if err != nil {
		log.Fatal(err)
	}
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("Runtime database pool unavailable; inspect database administration logs")
	}
	defer db.Close()
	if err = verifyRuntimeDatabaseStartup(ctx, db.Ping, db); err != nil {
		log.Fatal(err)
	}
	storage, err := verifyWorkspaceStorage(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	cancelStartup()
	tenancy, err := loadTenancySettings(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	if tenancy.Mode == tenancyMulti && !storage.Present {
		log.Fatal("COMMONROOM_TENANCY=multi requires workspace storage: run --migrate (migrations/002.sql and 003.sql) first")
	}
	settings, err := loadPlatformSettings(os.Getenv, dailyTurnLimit)
	if err != nil {
		log.Fatal(err)
	}
	// With workspace storage in place every query runs inside a transaction scoped to
	// one workspace; before the migration the previous layout is served unchanged.
	workspaceDatabase := func(id string) (eventDatabase, error) {
		if !storage.Present {
			return db, nil
		}
		return newScopedDatabase(db, id)
	}
	host := newWorkspaceHost(tenancy, databaseWorkspaceRegistry{db: db}, func(id string) (*workspaceRuntime, error) {
		scoped, scopeErr := workspaceDatabase(id)
		if scopeErr != nil {
			return nil, scopeErr
		}
		return newWorkspaceRuntime(settings, tenancy, scoped, id)
	})
	if tenancy.Mode == tenancySingle {
		if err = host.preload(); err != nil {
			log.Fatal(err)
		}
		log.Printf("Commonroom: http://%s (local audit protection; external immutable archive not configured)", settings.listenAddr)
	} else {
		log.Printf("Commonroom: hosting workspaces at *.%s behind the HTTPS proxy on %s (agent execution disabled for hosted workspaces)", tenancy.Domain, settings.listenAddr)
	}
	h := host
	log.Fatal(newHTTPServer(settings.listenAddr, h).ListenAndServe())
}
