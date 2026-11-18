package main

// Per-workspace server construction and host routing. A workspace is a complete,
// independent server value with its own lock, run state, stop and incident flags,
// owner authentication, sessions and data folder; only the database pool and the
// process settings are shared (ADR 0004).

import (
	"agentroom/internal/oidc"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const hostedExecutionMessage = "Agent execution is not available for hosted workspaces on this installation"

// platformSettings are process-wide inputs read once at startup.
type platformSettings struct {
	dailyTurnLimit                            int
	editorEnabled, templatesEnabled, safetyUI bool
	persistentSessions, verifyWorker          bool
	listenAddr, dataRoot, defaultKeyPath      string
	configuredOrigin, extensionRegistry       string
	oidc                                      *oidc.Provider
	// Hosted execution is off unless the operator opts in (see validateHostedExecution).
	hostedExecution bool
	pool            *executionPool
	oidcName        string
}

func loadPlatformSettings(getenv func(string) string, dailyTurnLimit int) (platformSettings, error) {
	ps := platformSettings{dailyTurnLimit: dailyTurnLimit, dataRoot: ".data"}
	var err error
	if ps.editorEnabled, err = connectionEditorEnabled(); err != nil {
		return ps, errors.New("COMMONROOM_CONNECTION_EDITOR must be true or false")
	}
	if ps.templatesEnabled, err = workflowTemplatesEnabled(); err != nil {
		return ps, errors.New("COMMONROOM_WORKFLOW_TEMPLATES must be true or false")
	}
	if ps.safetyUI, err = agentSafetyUIEnabled(); err != nil {
		return ps, errors.New("COMMONROOM_AGENT_SAFETY_UI must be true or false")
	}
	if ps.persistentSessions, err = persistentHumanSessionsEnabled(); err != nil {
		return ps, errors.New("COMMONROOM_PERSISTENT_SESSIONS must be true or false")
	}
	ps.verifyWorker = getenv("COMMONROOM_REMOTE_WORKER") == "true" || getenv("COMMONROOM_SERVER_WORKER") == "true"
	ps.listenAddr = getenv("LISTEN_ADDR")
	if ps.listenAddr == "" {
		ps.listenAddr = "127.0.0.1:8787"
	}
	ps.defaultKeyPath = getenv("OWNER_KEY_FILE")
	if ps.defaultKeyPath == "" {
		ps.defaultKeyPath = filepath.Join(ps.dataRoot, "owner-key")
	}
	ps.configuredOrigin = getenv("COMMONROOM_PUBLIC_ORIGIN")
	ps.extensionRegistry = getenv("COMMONROOM_EXTENSIONS_FILE")
	switch getenv("COMMONROOM_HOSTED_EXECUTION") {
	case "", "false":
	case "true":
		ps.hostedExecution = true
	default:
		return ps, errors.New("COMMONROOM_HOSTED_EXECUTION must be true or false")
	}
	maxTurns := 8
	if raw := strings.TrimSpace(getenv("COMMONROOM_HOSTED_MAX_TURNS")); raw != "" {
		n, convErr := strconv.Atoi(raw)
		if convErr != nil || n < 1 || n > 256 {
			return ps, errors.New("COMMONROOM_HOSTED_MAX_TURNS must be a whole number from 1 to 256")
		}
		maxTurns = n
	}
	ps.pool = newExecutionPool(maxTurns)
	if ps.oidc, ps.oidcName, err = loadOIDCSettings(getenv); err != nil {
		return ps, err
	}
	return ps, nil
}

// loadOIDCSettings reads the installation-wide identity provider, if one is configured.
// The client secret comes from the environment or, preferably, a private file.
func loadOIDCSettings(getenv func(string) string) (*oidc.Provider, string, error) {
	issuer, clientID := strings.TrimSpace(getenv("COMMONROOM_OIDC_ISSUER")), strings.TrimSpace(getenv("COMMONROOM_OIDC_CLIENT_ID"))
	secret, secretFile := getenv("COMMONROOM_OIDC_CLIENT_SECRET"), getenv("COMMONROOM_OIDC_CLIENT_SECRET_FILE")
	if issuer == "" && clientID == "" && secret == "" && secretFile == "" {
		return nil, "", nil
	}
	if secret != "" && secretFile != "" {
		return nil, "", errors.New("set only one of COMMONROOM_OIDC_CLIENT_SECRET and COMMONROOM_OIDC_CLIENT_SECRET_FILE")
	}
	if secretFile != "" {
		raw, err := readPrivateBounded(secretFile, 1024)
		if err != nil {
			return nil, "", errors.New("COMMONROOM_OIDC_CLIENT_SECRET_FILE must be a private file owned by this account")
		}
		secret = strings.TrimSpace(string(raw))
	}
	provider, err := oidc.New(oidc.Config{Issuer: issuer, ClientID: clientID, ClientSecret: secret})
	if err != nil {
		return nil, "", err
	}
	name := strings.TrimSpace(getenv("COMMONROOM_OIDC_NAME"))
	if name == "" {
		name = "Single sign-on"
	}
	if !validWorkspaceName(name) || len(name) > 40 {
		return nil, "", errors.New("COMMONROOM_OIDC_NAME must be 1-40 visible characters")
	}
	return provider, name, nil
}

// keyPath is where a workspace keeps its owner key, sessions and stop marker. The
// default workspace keeps its configured location so existing installs are unchanged.
func (ps platformSettings) keyPath(id string, t tenancySettings) string {
	if t.Mode == tenancySingle {
		return ps.defaultKeyPath
	}
	return filepath.Join(ps.dataRoot, "workspaces", id, "owner-key")
}

// setupCodeEnv exposes a supplied setup code only to the default workspace; a
// hosted workspace always gets its own generated code.
func (ps platformSettings) setupCodeEnv(id string) func(string) string {
	if id == defaultWorkspaceID {
		return os.Getenv
	}
	return func(string) string { return "" }
}

func (ps platformSettings) publicOrigin(id string, t tenancySettings) string {
	if t.Mode == tenancyMulti {
		return t.origin(id)
	}
	return ps.configuredOrigin
}

type workspaceRuntime struct {
	id      string
	s       *server
	handler http.Handler
	stop    context.CancelFunc
	routes  []string // every pattern the workspace serves; used by the isolation tests
}

// routeMux is a ServeMux that remembers what was registered, so tests can attack every route.
type routeMux struct {
	*http.ServeMux
	patterns []string
}

func newRouteMux() *routeMux { return &routeMux{ServeMux: http.NewServeMux()} }

func (m *routeMux) Handle(pattern string, h http.Handler) {
	m.patterns = append(m.patterns, pattern)
	m.ServeMux.Handle(pattern, h)
}

func (m *routeMux) HandleFunc(pattern string, f func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(f))
}

// hostedGate refuses worker-touching routes in a hosted installation until
// execution is namespaced per workspace.
func hostedGate(hosted bool, next http.HandlerFunc) http.HandlerFunc {
	return hostedGateWith(hosted, hostedExecutionMessage, next)
}

func hostedGateWith(hosted bool, message string, next http.HandlerFunc) http.HandlerFunc {
	if !hosted {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, message, 503)
	}
}

const hostedExtensionMessage = "Extensions are not yet available for hosted workspaces"

const hostedAlwaysMessage = "This is not available for hosted workspaces"

// newWorkspaceRuntime builds one workspace's complete server and HTTP handler.
func newWorkspaceRuntime(ps platformSettings, tenancy tenancySettings, db eventDatabase, id string) (*workspaceRuntime, error) {
	if !validWorkspaceID(id) {
		return nil, errors.New("invalid workspace id")
	}
	hosted := tenancy.Mode == tenancyMulti
	// Worker-touching routes stay refused in a hosted installation until the operator opts
	// in to hosted execution. Workspace disposal, extensions and conversation import stay
	// refused regardless: they reach the host's disk or a registry that is not per workspace.
	executionBlocked := hosted && !ps.hostedExecution
	s := &server{hosted: hosted, hostedExecutionBlocked: hosted && !ps.hostedExecution, dailyTurnLimit: ps.dailyTurnLimit, db: db, active: map[string]bool{}, cancels: map[string]context.CancelFunc{}, workspaceID: id}
	s.extensionRegistry = ps.extensionRegistry
	if hosted {
		s.pool = ps.pool
	}
	s.identity = identitySettings{provider: ps.oidc, name: ps.oidcName, hostedDefault: hosted}
	mux := newRouteMux()
	mux.HandleFunc("/instance-label.js", instanceLabel)
	mux.HandleFunc("/api/members", s.members)
	mux.HandleFunc("/api/auth/methods", s.authMethods)
	mux.HandleFunc("/api/webauthn/login/options", s.passkeyLoginOptions)
	mux.HandleFunc("/api/webauthn/login/verify", s.passkeyLoginVerify)
	mux.HandleFunc("/api/webauthn/register/options", s.passkeyRegisterOptions)
	mux.HandleFunc("/api/webauthn/register/verify", s.passkeyRegisterVerify)
	mux.HandleFunc("/api/webauthn/remove", s.passkeyRemove)
	mux.HandleFunc("/api/invitations", s.createInvitation)
	mux.HandleFunc("/api/invitations/revoke", s.revokeInvitation)
	mux.HandleFunc("/api/invitations/lookup", s.lookupInvitation)
	mux.HandleFunc("/api/invitations/passkey/options", s.invitationPasskeyOptions)
	mux.HandleFunc("/api/invitations/passkey/complete", s.invitationPasskeyComplete)
	mux.HandleFunc("/api/invitations/oidc/start", s.oidcInviteStart)
	mux.HandleFunc("/api/oidc/login", s.oidcLogin)
	mux.HandleFunc("/api/oidc/link", s.oidcLink)
	mux.HandleFunc("/api/oidc/callback", s.oidcCallback)
	mux.HandleFunc("/api/login", s.login)
	mux.HandleFunc("/api/setup-status", s.setupStatus)
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
	mux.HandleFunc("/ext/v1/", hostedGateWith(hosted, hostedExtensionMessage, s.extensionAPI))
	mux.HandleFunc("/api/task-dispatch", s.taskDispatch)
	mux.HandleFunc("/api/task-executions", s.taskExecutions)
	mux.HandleFunc("/api/task-definition-history", s.taskDefinitionHistory)
	mux.HandleFunc("/api/task-input-sources", s.taskInputSources)
	mux.HandleFunc("/api/task-acceptance", s.taskAcceptance)
	mux.HandleFunc("/api/task-progress", s.taskProgress)
	mux.HandleFunc("/api/task-repository-links", s.taskRepositoryLinks)
	mux.HandleFunc("/api/governance", s.governance)
	mux.HandleFunc("/api/task-proposals", s.taskProposals)
	mux.HandleFunc("/api/task-delegations", s.taskDelegations)
	mux.HandleFunc("/api/room-outcome-review", s.roomOutcomeReview)
	mux.HandleFunc("/api/workflow-runs", s.workflowRuns)
	mux.HandleFunc("/api/agent-recovery", hostedGate(executionBlocked, s.agentRecovery))
	mux.HandleFunc("/api/job-diagnostics", hostedGate(executionBlocked, s.jobDiagnostics))
	mux.HandleFunc("/api/job-recovery", hostedGate(executionBlocked, s.jobRecovery))
	mux.HandleFunc("/api/workspace-disposal", hostedGateWith(hosted, hostedAlwaysMessage, s.workspaceDisposal))
	mux.HandleFunc("/api/provider-auth", hostedGate(executionBlocked, s.providerAuth))
	mux.HandleFunc("/api/connections/test", hostedGate(executionBlocked, s.connectionTest))
	editorEnabled := ps.editorEnabled
	s.templatesEnabled = ps.templatesEnabled
	s.safetyUIDefault = ps.safetyUI
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
	mux.Handle("/", staticHandler("web"))
	keyPath := ps.keyPath(id, tenancy)
	s.integrityStopPath = filepath.Join(filepath.Dir(keyPath), "integrity-stop")
	s.restoreIntegrityStop()
	key, keyExpires, err := loadSetupCode(keyPath, ps.setupCodeEnv(id), time.Now())
	if err != nil {
		return nil, err
	}
	s.auth, err = newOwnerAuth(key, ps.listenAddr)
	if err != nil {
		return nil, err
	}
	s.auth.keyExpires = keyExpires
	if hosted {
		s.auth.useHostCookies()
	}
	if hosted && keyExpires.IsZero() {
		// A hosted setup link that is never claimed must not stay valid forever; the
		// platform operator reissues it, and a restart renews the window.
		s.auth.keyExpires = time.Now().Add(setupLinkTTL)
	}
	if err := s.auth.configurePublicOrigin(ps.publicOrigin(id, tenancy)); err != nil {
		return nil, err
	}
	if err = s.prepareOwnerIdentity(context.Background(), keyPath); err != nil {
		return nil, err
	}
	if err = s.auth.configureSessionPersistence(filepath.Join(filepath.Dir(keyPath), "human-sessions.json"), ps.persistentSessions); err != nil {
		return nil, fmt.Errorf("Private session storage unavailable: %w", err)
	}
	log.Printf("Workspace %s: owner setup code file: %s (keep private)", id, keyPath)
	s.checkIntegrityNow()
	if ps.verifyWorker && !hosted {
		s.verifyWorkerAtStartup(context.Background(), approvedWorkerEndpoint())
	}
	watchCtx, stopWatching := context.WithCancel(context.Background())
	go s.watchIntegrity(watchCtx, 30*time.Second)
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
	return &workspaceRuntime{id: id, s: s, handler: h, stop: stopWatching, routes: mux.patterns}, nil
}

// workspaceHost routes each request to its workspace's handler. In single mode
// that is always the one default workspace. In multi mode the host name picks the
// workspace, the registry must know and allow it, and its server is built on first
// use and kept.
type workspaceHost struct {
	tenancy  tenancySettings
	registry workspaceRegistry
	build    func(id string) (*workspaceRuntime, error)
	now      func() time.Time
	platform http.Handler   // served on platform.<domain>; nil when not configured
	signup   *signupService // served on the bare domain; nil unless COMMONROOM_SIGNUP=open
	meter    requestMeter

	mu      sync.Mutex
	entries map[string]*workspaceEntry
	missing map[string]time.Time // recent unknown hosts, so probing cannot hammer the registry
}

type workspaceEntry struct {
	once    sync.Once
	runtime *workspaceRuntime
	err     error
	checked time.Time
	status  string
	limits  workspaceLimits
}

const (
	registryRecheck  = 5 * time.Second  // how long a registry answer is trusted
	missingRemember  = 30 * time.Second // how long an unknown workspace is remembered
	maxMissingMemory = 4096
)

func newWorkspaceHost(t tenancySettings, registry workspaceRegistry, build func(string) (*workspaceRuntime, error)) *workspaceHost {
	return &workspaceHost{tenancy: t, registry: registry, build: build, now: time.Now, entries: map[string]*workspaceEntry{}, missing: map[string]time.Time{}}
}

// preload builds the single default workspace eagerly so a configuration error stops
// startup instead of failing the first request.
func (h *workspaceHost) preload() error {
	rt, err := h.runtimeFor(context.Background(), defaultWorkspaceID)
	if err != nil {
		return err
	}
	_ = rt
	return nil
}

var errWorkspaceUnknown = errors.New("workspace not found")
var errWorkspaceUnavailable = errors.New("workspace unavailable")

func (h *workspaceHost) runtimeFor(ctx context.Context, id string) (*workspaceRuntime, error) {
	return h.resolve(ctx, id, false)
}

// adminRuntime is runtimeFor for the platform operator: a suspended workspace can still
// be reached so the action can be recorded in its history; it is never served to users.
func (h *workspaceHost) adminRuntime(ctx context.Context, id string) (*workspaceRuntime, error) {
	return h.resolve(ctx, id, true)
}

func (h *workspaceHost) resolve(ctx context.Context, id string, allowInactive bool) (*workspaceRuntime, error) {
	now := h.now()
	h.mu.Lock()
	if until, ok := h.missing[id]; ok {
		if now.Before(until) {
			h.mu.Unlock()
			return nil, errWorkspaceUnknown
		}
		delete(h.missing, id)
	}
	entry := h.entries[id]
	if entry == nil {
		entry = &workspaceEntry{}
		h.entries[id] = entry
	}
	needCheck := h.tenancy.Mode == tenancyMulti && (entry.checked.IsZero() || now.Sub(entry.checked) >= registryRecheck)
	h.mu.Unlock()
	if needCheck {
		rec, found, err := h.registry.Lookup(ctx, id)
		if err != nil {
			return nil, errWorkspaceUnavailable
		}
		h.mu.Lock()
		if !found || rec.Status == workspaceDeleted {
			if len(h.missing) >= maxMissingMemory {
				h.missing = map[string]time.Time{}
			}
			h.missing[id] = now.Add(missingRemember)
			if entry.runtime == nil {
				delete(h.entries, id)
			}
			h.mu.Unlock()
			return nil, errWorkspaceUnknown
		}
		entry.checked, entry.status, entry.limits = now, rec.Status, rec.Limits
		rt := entry.runtime
		h.mu.Unlock()
		if rt != nil {
			limits := rec.Limits
			rt.s.setLimits(&limits)
		}
	}
	h.mu.Lock()
	status := entry.status
	h.mu.Unlock()
	if h.tenancy.Mode == tenancyMulti && status != workspaceActive && !(allowInactive && status == workspaceSuspended) {
		return nil, errWorkspaceUnavailable
	}
	entry.once.Do(func() {
		built, buildErr := h.build(id)
		// The registry recheck reads these fields under h.mu, so they are written under it.
		h.mu.Lock()
		entry.runtime, entry.err = built, buildErr
		limits := entry.limits
		h.mu.Unlock()
		if buildErr == nil && h.tenancy.Mode == tenancyMulti {
			built.s.setLimits(&limits)
		}
	})
	if entry.err != nil {
		h.mu.Lock()
		if h.entries[id] == entry {
			delete(h.entries, id) // a failed build is retried by the next request
		}
		h.mu.Unlock()
		log.Printf("Workspace %s could not start: %v", id, entry.err)
		return nil, errWorkspaceUnavailable
	}
	return entry.runtime, nil
}

// platformHost reports whether the request is for the operator's platform API.
func (h *workspaceHost) platformHost(host string) bool {
	if h.tenancy.Mode != tenancyMulti {
		return false
	}
	if bare, _, err := net.SplitHostPort(host); err == nil {
		host = bare
	}
	return strings.ToLower(strings.TrimSuffix(host, ".")) == "platform."+h.tenancy.Domain
}

func (h *workspaceHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.signupHost(r.Host) {
		h.signup.ServeHTTP(w, r)
		return
	}
	if h.platformHost(r.Host) {
		if h.platform == nil {
			http.Error(w, "Not found", 404)
			return
		}
		h.platform.ServeHTTP(w, r)
		return
	}
	id, ok := h.tenancy.workspaceFromHost(r.Host)
	if !ok {
		http.Error(w, "Not found", 404)
		return
	}
	rt, err := h.runtimeFor(r.Context(), id)
	switch {
	case errors.Is(err, errWorkspaceUnknown):
		http.Error(w, "Not found", 404)
	case err != nil:
		http.Error(w, "This workspace is unavailable", 503)
	default:
		if h.tenancy.Mode == tenancyMulti && strings.HasPrefix(r.URL.Path, "/api/") {
			h.mu.Lock()
			current := h.entries[id]
			limit := 0
			if current != nil {
				limit = current.limits.RequestsPerMinute
			}
			h.mu.Unlock()
			if current == nil {
				// Evicted (suspended or deleted) between resolving and routing: fail closed.
				http.Error(w, "This workspace is unavailable", 503)
				return
			}
			if !h.meter.admit(id, limit, h.now()) {
				tooManyRequests(w)
				return
			}
		}
		rt.handler.ServeHTTP(w, r)
	}
}

// invalidate makes the next request recheck the registry (after a status or limits change).
func (h *workspaceHost) invalidate(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.missing, id)
	if e := h.entries[id]; e != nil {
		e.checked = time.Time{}
	}
}

// evict stops a workspace's server and forgets it; it is rebuilt only if the registry allows.
func (h *workspaceHost) evict(id string) {
	h.mu.Lock()
	e := h.entries[id]
	delete(h.entries, id)
	delete(h.missing, id)
	h.mu.Unlock()
	if e != nil && e.runtime != nil {
		e.runtime.stop()
	}
}

// stopAll ends every built workspace's background work (shutdown and tests).
func (h *workspaceHost) stopAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, e := range h.entries {
		if e.runtime != nil {
			e.runtime.stop()
		}
	}
}

func logPlatformAuditGap(id, action string) {
	log.Printf("Workspace %s: platform action %q changed the registry but could not be recorded in the workspace history", id, action)
}

// validateHostedExecution is checked at startup when hosted execution is switched on. Hosted
// agents are chat-only: repository workspaces and isolated commands read the host's disk
// through process-wide settings, so they must be off, and a worker must be configured.
func validateHostedExecution(getenv func(string) string) error {
	for _, name := range []string{"COMMONROOM_DEVELOPMENT_WORKSPACES", "COMMONROOM_ISOLATED_COMMANDS", "COMMONROOM_WORKER_DEVELOPMENT"} {
		if v := getenv(name); v != "" && v != "false" {
			return errors.New("COMMONROOM_HOSTED_EXECUTION=true runs chat-only agents: unset " + name + " (it gives agents access to the host's repositories or commands)")
		}
	}
	if getenv("COMMONROOM_REMOTE_WORKER") != "true" && getenv("COMMONROOM_SERVER_WORKER") != "true" {
		return errors.New("COMMONROOM_HOSTED_EXECUTION=true needs an isolated worker: set COMMONROOM_REMOTE_WORKER or COMMONROOM_SERVER_WORKER to true")
	}
	return nil
}
