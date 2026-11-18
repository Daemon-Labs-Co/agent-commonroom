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
	if tenancy.Mode == tenancyMulti && settings.hostedExecution {
		if err = validateHostedExecution(os.Getenv); err != nil {
			log.Fatal(err)
		}
		log.Print("Hosted agent execution is ENABLED: chat-only agents run on the configured worker for every workspace, within per-workspace and installation limits")
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
	if tenancy.Mode == tenancyMulti {
		keyFile := os.Getenv("COMMONROOM_PLATFORM_KEYS_FILE")
		host.platform = &platformAPI{tenancy: tenancy, store: databaseWorkspaceRegistry{db: db}, host: host, keyFile: keyFile, now: time.Now}
		if keyFile == "" {
			log.Print("COMMONROOM_PLATFORM_KEYS_FILE is not set: the platform API at platform." + tenancy.Domain + " refuses every request")
		}
	}
	signup, err := loadSignupSettings(os.Getenv, tenancy)
	if err != nil {
		log.Fatal(err)
	}
	if signup.enabled {
		host.signup = newSignupService(signup, tenancy, databaseWorkspaceRegistry{db: db}, host)
		log.Printf("Built-in sign-up is ENABLED at https://%s (limits: %d new workspaces a day, %d in total; sign-up code %s)", tenancy.Domain, signup.dailyLimit, signup.maxTotal, map[bool]string{true: "required", false: "not required"}[signup.code != ""])
	}
	if tenancy.Mode == tenancySingle {
		if err = host.preload(); err != nil {
			log.Fatal(err)
		}
		log.Printf("Commonroom: http://%s (local audit protection; external immutable archive not configured)", settings.listenAddr)
	} else {
		log.Printf("Commonroom: hosting workspaces at *.%s behind the HTTPS proxy on %s (hosted agent execution is off unless COMMONROOM_HOSTED_EXECUTION=true)", tenancy.Domain, settings.listenAddr)
	}
	h := host
	log.Fatal(newHTTPServer(settings.listenAddr, h).ListenAndServe())
}
