package main

// Executable entry point: configuration, dependency wiring and HTTP startup.

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
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
	cancelStartup()
	s := &server{dailyTurnLimit: dailyTurnLimit, db: db, active: map[string]bool{}, cancels: map[string]context.CancelFunc{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/members", s.members)
	mux.HandleFunc("/api/login", s.login)
	mux.HandleFunc("/api/challenge", s.challenge)
	mux.HandleFunc("/api/owner-key/download", s.downloadOwnerKey)
	mux.HandleFunc("/api/logout", s.logout)
	mux.HandleFunc("/api/session", s.session)
	mux.HandleFunc("/api/company/organization", s.organization)
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/integrity", s.integrity)
	mux.HandleFunc("/api/agent-safety", s.agentSafety)
	mux.HandleFunc("/api/agents", s.agents)
	mux.HandleFunc("/api/task-dispatch", s.taskDispatch)
	mux.HandleFunc("/api/governance", s.governance)
	mux.HandleFunc("/api/task-proposals", s.taskProposals)
	mux.HandleFunc("/api/workflow-runs", s.workflowRuns)
	mux.HandleFunc("/api/agent-recovery", s.agentRecovery)
	mux.HandleFunc("/api/job-diagnostics", s.jobDiagnostics)
	mux.HandleFunc("/api/job-recovery", s.jobRecovery)
	mux.HandleFunc("/api/workspace-disposal", s.workspaceDisposal)
	mux.HandleFunc("/api/provider-auth", s.providerAuth)
	mux.HandleFunc("/api/connections/test", s.connectionTest)
	editorEnabled, err := connectionEditorEnabled()
	if err != nil {
		log.Fatal("COMMONROOM_CONNECTION_EDITOR must be true or false")
	}
	s.templatesEnabled, err = workflowTemplatesEnabled()
	if err != nil {
		log.Fatal("COMMONROOM_WORKFLOW_TEMPLATES must be true or false")
	}
	safetyUI, err := agentSafetyUIEnabled()
	if err != nil {
		log.Fatal("COMMONROOM_AGENT_SAFETY_UI must be true or false")
	}
	s.safetyUIDefault = safetyUI
	mux.Handle("/organization-chart.js", s.roomPresentationScript("web", "organization-chart.js", "companyOrgChart"))
	mux.Handle("/viewport-alignment.js", s.roomPresentationScript("web", "viewport-alignment.js", "topAlignedViews"))
	mux.Handle("/overlay-layout.js", s.roomPresentationScript("web", "overlay-layout.js", "compactProfileOverlays"))
	mux.Handle("/human-profile-details.js", s.roomPresentationScript("web", "human-profile-details.js", "humanProfileDetails"))
	mux.Handle("/room-sections.js", s.roomPresentationScript("web", "room-sections.js", "roomSections"))
	mux.Handle("/settings-tabs.js", s.roomPresentationScript("web", "settings-tabs.js", "settingsTabs"))
	mux.Handle("/history-presentation.js", s.roomPresentationScript("web", "history-presentation.js", "historyRefinement"))
	mux.Handle("/human-attention.js", s.roomPresentationScript("web", "human-attention.js", "humanAttention"))
	mux.Handle("/squad-menus.js", s.roomPresentationScript("web", "squad-menus.js", "squadMenuLayout"))
	mux.Handle("/agent-status.js", s.roomPresentationScript("web", "agent-status.js", "agentStatusDetails"))
	mux.Handle("/agent-recovery.js", s.roomPresentationScript("web", "agent-recovery.js", "agentRecovery"))
	mux.Handle("/job-diagnostics.js", s.roomPresentationScript("web", "job-diagnostics.js", "jobDiagnostics"))
	mux.Handle("/background-work.js", s.roomPresentationScript("web", "background-work.js", "agentBackgroundWork"))
	mux.Handle("/room-viewport.js", s.roomPresentationScript("web", "room-viewport.js", "roomViewportFit"))
	mux.Handle("/feature-switches.js", s.roomPresentationScript("web", "feature-switches.js", "featureFlagSwitches"))
	mux.Handle("/room-message-order.js", s.roomPresentationScript("web", "room-message-order.js", "newestMessagesFirst"))
	mux.Handle("/history-date-groups.js", s.roomPresentationScript("web", "history-date-groups.js", "historyDateGroups"))
	mux.Handle("/feature-flag-tabs.js", s.roomPresentationScript("web", "feature-flag-tabs.js", "featureFlagTabs"))
	mux.Handle("/help-icons.js", s.roomPresentationScript("web", "help-icons.js", "standardHelpIcons"))
	mux.Handle("/squad-profiles.js", s.roomPresentationScript("web", "squad-profiles.js", "squadProfileActions"))
	mux.Handle("/agent-invitation.js", s.roomPresentationScript("web", "agent-invitation.js", "agentInvitationLayout"))
	mux.Handle("/agent-safety.js", s.agentSafetyScript("web"))
	mux.Handle("/workflow-templates.js", optionalScriptHandler(s.templatesEnabled, "web", "workflow-templates.js"))
	mux.Handle("/connection-editor.js", connectionEditorHandler(editorEnabled, "web"))
	mux.Handle("/help/", documentationHandler("."))
	mux.Handle("/", http.FileServer(http.Dir("web")))
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8787"
	}
	keyPath := os.Getenv("OWNER_KEY_FILE")
	if keyPath == "" {
		keyPath = ".data/owner-key"
	}
	s.integrityStopPath = filepath.Join(filepath.Dir(keyPath), "integrity-stop")
	s.restoreIntegrityStop()
	key, err := loadOwnerKey(keyPath)
	if err != nil {
		log.Fatal(err)
	}
	s.auth, err = newOwnerAuth(key, addr)
	if err != nil {
		log.Fatal(err)
	}
	if err = s.auth.prepareRSA(keyPath); err != nil {
		log.Fatal(err)
	}
	persistentSessions, err := persistentHumanSessionsEnabled()
	if err != nil {
		log.Fatal("COMMONROOM_PERSISTENT_SESSIONS must be true or false")
	}
	if err = s.auth.configureSessionPersistence(filepath.Join(filepath.Dir(keyPath), "human-sessions.json"), persistentSessions); err != nil {
		log.Fatal("Private session storage unavailable: ", err)
	}
	log.Printf("Owner setup code file: %s (keep private)", keyPath)
	log.Printf("Commonroom: http://%s (local audit protection; external immutable archive not configured)", addr)
	s.checkIntegrityNow()
	if os.Getenv("COMMONROOM_REMOTE_WORKER") == "true" {
		s.verifyWorkerAtStartup(context.Background(), "ssh://"+os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"))
	}
	go s.watchIntegrity(context.Background(), 30*time.Second)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'")
		if strings.HasPrefix(r.URL.Path, "/help/") {
			w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'self'")
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		s.auth.protect(s.actorGuard(mux)).ServeHTTP(w, r)
	})
	log.Fatal(newHTTPServer(addr, h).ListenAndServe())
}
