package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"agentroom/internal/room"
	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	premium           *premiumEntitlements
	workflowLive      map[string]context.CancelFunc
	workflowTurn      func(context.Context, room.Participant, string, string) (string, error)
	workflowReady     func(string) error
	recoveryIdle      func(context.Context, string) (bool, error)
	safetyUIDefault   bool
	stopAll           bool
	integrityFault    bool
	integrityStopPath string
	requestBudgets    map[string]actorRequestBudget
	stopRooms         map[string]bool
	templatesEnabled  bool
	db                eventDatabase
	auth              *ownerAuth
	mu                sync.Mutex
	active            map[string]bool
	cancels           map[string]context.CancelFunc
}

func id() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func checksum(previous, kind, payload string) string {
	h := sha256.Sum256([]byte(previous + "\n" + kind + "\n" + payload))
	return hex.EncodeToString(h[:])
}
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
	developmentEnabled, developmentRegistry, developmentErr := loadControllerDevelopmentConfiguration(".data/development-workspaces.json", os.Getenv("COMMONROOM_DEVELOPMENT_WORKSPACES"), os.Getenv("COMMONROOM_DEVELOPMENT_REPOSITORIES"))
	if developmentErr != nil {
		log.Fatal(developmentErr)
	}
	os.Setenv("COMMONROOM_DEVELOPMENT_WORKSPACES", developmentEnabled)
	os.Setenv("COMMONROOM_DEVELOPMENT_REPOSITORIES", developmentRegistry)
	ctx := context.Background()
	var localConfig struct {
		DatabaseURL         string `json:"databaseUrl"`
		RemoteWorkerAlias   string `json:"remoteWorkerAlias"`
		RemoteWorkerEnabled bool   `json:"remoteWorkerEnabled"`
	}
	if raw, e := os.ReadFile(".data/local-config.json"); e == nil {
		if e = json.Unmarshal(raw, &localConfig); e != nil {
			log.Fatal("Invalid local configuration: ", e)
		}
	} else if !os.IsNotExist(e) {
		log.Fatal(e)
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
		log.Fatal(err)
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		log.Fatal("PostgreSQL unavailable: ", err)
	}
	if err = validateRuntimeDatabase(ctx, db); err != nil {
		log.Fatal(err)
	}
	s := &server{db: db, active: map[string]bool{}, cancels: map[string]context.CancelFunc{}}
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
	mux.HandleFunc("/api/workflow-runs", s.workflowRuns)
	mux.HandleFunc("/api/agent-recovery", s.agentRecovery)
	mux.HandleFunc("/api/premium-features", s.premiumFeatures)
	mux.HandleFunc("/api/protected-audit-archive", s.protectedArchiveStatus)
	mux.HandleFunc("/api/managed-orchestration", s.managedOrchestrationStatus)
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
	s.premium, err = configurePremiumEntitlements(os.Getenv("COMMONROOM_PREMIUM_ISSUER_KEY"), os.Getenv("COMMONROOM_PREMIUM_LICENSE_FILE"), filepath.Dir(keyPath))
	if err != nil {
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
	log.Fatal((&http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 50 * time.Second}).ListenAndServe())
}
func reply(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
func (s *server) session(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	session, ok := s.auth.session(r)
	if !ok {
		http.Error(w, "Sign in required", 401)
		return
	}
	s.auth.mu.Lock()
	pending := s.auth.setupPending
	s.auth.mu.Unlock()
	reply(w, map[string]any{"csrf": session.CSRF, "actorId": requestActor(r), "actorType": "human", "keySetupPending": pending && requestActor(r) == ownerID})
}
func (s *server) append(ctx context.Context, kind string, payload any) error {
	// Background work is a projection of separate telemetry events, not part of
	// a mutable profile/session snapshot. Do not duplicate it into those records.
	if participant, ok := payload.(room.Participant); ok {
		participant.Background = nil
		payload = participant
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(878701)"); err != nil {
		return err
	}
	previous := ""
	err = tx.QueryRow(ctx, "SELECT hash FROM events ORDER BY sequence DESC LIMIT 1").Scan(&previous)
	if err != nil && err.Error() != "no rows in result set" {
		return err
	}
	// PostgreSQL canonical JSON text is used for verification and chain calculation.
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if actor, ok := ctx.Value(actorContextKey{}).(string); ok && actor != "" {
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || fields == nil {
			return errors.New("audit payload must be an object")
		}
		fields["performedBy"], _ = json.Marshal(actor)
		raw, err = json.Marshal(fields)
		if err != nil {
			return err
		}
	}
	var canonical string
	var sequence int64
	if err = tx.QueryRow(ctx, "SELECT nextval('public.events_sequence_seq')").Scan(&sequence); err != nil {
		return err
	}
	meta := auditRecordMetadata{Version: 1, ID: id(), Sequence: sequence, CommittedAt: time.Now().UTC().Truncate(time.Microsecond)}
	raw, err = protectAuditMetadata(raw, meta)
	if err != nil {
		return err
	}
	if err = tx.QueryRow(ctx, "SELECT $1::jsonb::text", string(raw)).Scan(&canonical); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO events(sequence,id,kind,payload,committed_at,previous_hash,hash) VALUES($1,$2,$3,$4::jsonb,$5,$6,$7)", meta.Sequence, meta.ID, kind, canonical, meta.CommittedAt, previous, checksum(previous, kind, canonical))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *server) load(ctx context.Context) (room.State, error) {
	activities := map[string]string{}
	background := map[string][]room.BackgroundActivity{}
	activityStarts := map[string]time.Time{}
	failures := map[string]string{}
	acknowledged := map[string]string{}
	st := room.State{Company: room.Company{ID: "company", Name: "My company"}, HumanCompanyID: "company", Workflows: []room.Workflow{}, Rooms: []room.Room{{ID: "engineering", Name: "Engineering room"}}, Tasks: []room.Task{}, Evidence: []room.Evidence{}, Connections: []room.Connection{}, Participants: []room.Participant{}, Memories: []room.Memory{}, Messages: []room.Message{}}
	rows, err := s.db.Query(ctx, "SELECT kind,payload FROM events ORDER BY sequence")
	if err != nil {
		return st, err
	}
	defer rows.Close()
	for rows.Next() {
		var kind string
		var data []byte
		if err = rows.Scan(&kind, &data); err != nil {
			return st, err
		}
		switch kind {
		case "actor.help.updated":
			var v struct {
				ID   string `json:"id"`
				Mode string `json:"mode"`
			}
			if err := json.Unmarshal(data, &v); err != nil {
				return st, err
			}
			projectActors(&st)
			for i := range st.Actors {
				if st.Actors[i].ID == v.ID {
					st.Actors[i].HelpDisplay = v.Mode
				}
			}
		case "actor.image.updated":
			var v struct {
				ID    string `json:"id"`
				Image string `json:"image"`
			}
			if err := json.Unmarshal(data, &v); err != nil {
				return st, err
			}
			projectActors(&st)
			for i := range st.Actors {
				if st.Actors[i].ID == v.ID {
					st.Actors[i].Image = v.Image
				}
			}
		case "actor.name.updated":
			var v struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if e := json.Unmarshal(data, &v); e != nil {
				return st, e
			}
			for i, a := range st.Actors {
				if a.ID == v.ID {
					st.Actors[i].Name = v.Name
				}
			}
		case "actor.revised":
			var v room.Actor
			if e := json.Unmarshal(data, &v); e != nil {
				return st, e
			}
			found := false
			for i, a := range st.Actors {
				if a.ID == v.ID {
					v.HelpDisplay = a.HelpDisplay
					st.Actors[i] = v
					found = true
				}
			}
			if !found {
				st.Actors = append(st.Actors, v)
			}
		case "company.organization.updated":
			if err := json.Unmarshal(data, &st.Organization); err != nil {
				return st, err
			}
		case "company.updated":
			json.Unmarshal(data, &st.Company)
		case "human.profile.updated":
			var v struct {
				Name  string  `json:"name"`
				Title *string `json:"title"`
			}
			json.Unmarshal(data, &v)
			st.HumanName = v.Name
			if v.Title != nil {
				st.HumanTitle = *v.Title
			}
		case "workflow.run.updated":
			var run room.WorkflowRun
			if json.Unmarshal(data, &run) != nil {
				return st, fmt.Errorf("Invalid workflow run event")
			}
			found := false
			for i, old := range st.WorkflowRuns {
				if old.ID == run.ID {
					st.WorkflowRuns[i] = run
					found = true
					break
				}
			}
			if !found {
				st.WorkflowRuns = append(st.WorkflowRuns, run)
			}
		case "workflow.revised":
			var v room.Workflow
			json.Unmarshal(data, &v)
			found := false
			for i, old := range st.Workflows {
				if old.ID == v.ID {
					st.Workflows[i] = v
					found = true
				}
			}
			if !found {
				st.Workflows = append(st.Workflows, v)
			}

		case "room.created", "room.updated":
			var v room.Room
			json.Unmarshal(data, &v)
			if kind == "room.created" {
				st.Rooms = append(st.Rooms, v)
			} else {
				for i, old := range st.Rooms {
					if old.ID == v.ID {
						st.Rooms[i] = v
					}
				}
			}
		case "task.revised":
			var v room.Task
			json.Unmarshal(data, &v)
			found := false
			for i, t := range st.Tasks {
				if t.ID == v.ID {
					st.Tasks[i] = v
					found = true
				}
			}
			if !found {
				st.Tasks = append(st.Tasks, v)
			}
		case "evidence.created":
			var v room.Evidence
			json.Unmarshal(data, &v)
			st.Evidence = append(st.Evidence, v)

		case "connection.created":
			var v room.Connection
			json.Unmarshal(data, &v)
			st.Connections = append(st.Connections, v)
			st.Actors = append(st.Actors, room.Actor{ID: "bot-" + v.ID, Name: v.Name, Kind: "agent", Role: "Agent", Enabled: true, Revision: 1})

		case "connection.deleted":
			var v struct {
				ID string `json:"id"`
			}
			json.Unmarshal(data, &v)
			connections := st.Connections[:0]
			for _, c := range st.Connections {
				if c.ID != v.ID {
					connections = append(connections, c)
				}
			}
			st.Connections = connections
			participants := st.Participants[:0]
			for _, member := range st.Participants {
				if member.ConnectionID != v.ID {
					participants = append(participants, member)
				}
			}
			st.Participants = participants
		case "connection.updated":
			var v room.Connection
			json.Unmarshal(data, &v)
			for i, old := range st.Connections {
				if old.ID != v.ID {
					continue
				}
				st.Connections[i] = v
				if old.Endpoint != v.Endpoint || old.Protocol != v.Protocol {
					for j, member := range st.Participants {
						if member.ConnectionID == v.ID {
							st.Participants[j].ThreadID = ""
							st.Participants[j].ExecutionMode = ""
							st.Participants[j].Status = ""
							if room.SupportsAgent(v) {
								st.Participants[j].Status = "Idle"
							}
						}
					}
				}
			}
		case "participant.created":
			var v room.Participant
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			st.Participants = append(st.Participants, v)
		case "agent.activity":
			var activity struct {
				ParticipantID string    `json:"participantId"`
				Activity      string    `json:"activity"`
				StartedAt     time.Time `json:"startedAt"`
			}
			json.Unmarshal(data, &activity)
			activities[activity.ParticipantID] = activity.Activity
			activityStarts[activity.ParticipantID] = activity.StartedAt
		case "agent.background":
			var event room.BackgroundActivity
			if json.Unmarshal(data, &event) == nil && backgroundLabels[event.Stage] != "" {
				event.Label = backgroundLabels[event.Stage]
				background[event.ParticipantID] = append(background[event.ParticipantID], event)
				if len(background[event.ParticipantID]) > 64 {
					background[event.ParticipantID] = background[event.ParticipantID][1:]
				}
			}
		case "agent.error":
			var failure struct {
				ParticipantID string `json:"participantId"`
				Error         string `json:"error"`
			}
			json.Unmarshal(data, &failure)
			failures[failure.ParticipantID] = failure.Error
		case "agent.recovered":
			var recovered struct {
				Participant room.Participant `json:"participant"`
			}
			if json.Unmarshal(data, &recovered) == nil {
				for i, p := range st.Participants {
					if p.ID == recovered.Participant.ID && p.RoomID == recovered.Participant.RoomID {
						st.Participants[i] = recovered.Participant
						delete(failures, p.ID)
					}
				}
			}
		case "participant.updated", "participant.session":
			var v room.Participant
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			for i, p := range st.Participants {
				if p.ID == v.ID {
					st.Participants[i] = v
					if kind == "participant.session" && v.Status != "Blocked" {
						delete(failures, v.ID)
					}
					if v.ConnectionID != p.ConnectionID {
						delete(failures, v.ID)
						delete(activities, v.ID)
						delete(activityStarts, v.ID)
					}
				}
			}
		case "memory.revised":
			var v room.Memory
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			found := false
			for i, m := range st.Memories {
				if m.ID == v.ID {
					st.Memories[i] = v
					found = true
				}
			}
			if !found {
				st.Memories = append(st.Memories, v)
			}
		case "participant.removed":
			var v struct {
				ID string `json:"id"`
			}
			json.Unmarshal(data, &v)
			kept := st.Participants[:0]
			for _, p := range st.Participants {
				if p.ID != v.ID {
					kept = append(kept, p)
				}
			}
			st.Participants = kept
		case "message.attention.acknowledged":
			var ack struct {
				MessageID string `json:"messageId"`
				ActorID   string `json:"actorId"`
			}
			json.Unmarshal(data, &ack)
			acknowledged[ack.MessageID] = ack.ActorID
		case "message.created":
			var v room.Message
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			if v.Provider == "" && v.Author != "You" {
				for _, p := range st.Participants {
					if p.Name == v.Author && p.RoomID == v.RoomID {
						for _, c := range st.Connections {
							if c.ID == p.ConnectionID {
								v.Provider = c.Provider
								if v.Provider == "" && room.SupportsAgent(c) {
									v.Provider = providerForEndpoint(c.Endpoint)
								}
							}
						}
					}
				}
			}
			st.Messages = append(st.Messages, v)
		case "feature.flag.updated":
			var v struct {
				Name    string `json:"name"`
				Enabled bool   `json:"enabled"`
			}
			if e := json.Unmarshal(data, &v); e != nil {
				return st, e
			}
			if st.FeatureFlags == nil {
				st.FeatureFlags = map[string]bool{}
			}
			st.FeatureFlags[v.Name] = v.Enabled
		case "agents.stop", "agents.resume":
			var v struct {
				Scope  string `json:"scope"`
				RoomID string `json:"roomId"`
			}
			if e := json.Unmarshal(data, &v); e != nil {
				return st, e
			}
			if v.Scope == "application" {
				st.AgentsStopped = kind == "agents.stop"
			} else if v.Scope == "room" {
				kept := []string{}
				for _, rid := range st.AgentStopRooms {
					if rid != v.RoomID {
						kept = append(kept, rid)
					}
				}
				if kind == "agents.stop" {
					kept = append(kept, v.RoomID)
				}
				st.AgentStopRooms = kept
			}
		case "incident.opened":
			st.Incident = true
		case "incident.recovered":
			st.Incident = false
		}
	}
	projectAgentStatusDetails(&st, activities, failures)
	projectAgentElapsed(&st, activityStarts, time.Now().UTC())
	projectBackground(&st, background)
	for i := range st.Rooms {
		st.Rooms[i].CompanyID = st.Company.ID
	}
	for i := range st.Connections {
		st.Connections[i].CompanyID = st.Company.ID
	}
	for i := range st.Participants {
		st.Participants[i].CompanyID = st.Company.ID
	}
	projectActors(&st)
	projectHumanAttention(&st, acknowledged)
	return st, rows.Err()
}
func (s *server) state(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "Unable to load state", 500)
		return
	}
	rid := requestedRoom(r)
	if requestActor(r) == ownerID {
		for _, room := range st.Rooms {
			if room.ID == rid && room.PersonalActorID != "" && room.PersonalActorID != ownerID {
				if e := s.append(r.Context(), "room.owner_access", map[string]string{"actorId": ownerID, "roomId": rid, "reason": "owner viewed personal room"}); e != nil {
					http.Error(w, "Unable to audit owner access", 503)
					return
				}
			}
		}
	}
	s.mu.Lock()
	if st.FeatureFlags == nil {
		st.FeatureFlags = map[string]bool{}
	}
	st.FeatureFlags["agentSafetyControls"] = s.agentSafetyControlsEnabled(st)
	st.FeatureFlags["humanProfileImages"] = humanProfileImagesEnabled(st)
	for _, feature := range premiumFeatures {
		info := s.premiumFeatureInfo(st, feature, time.Now())
		st.FeatureFlags[feature] = info.Enabled
		if requestActor(r) == ownerID {
			if st.PremiumFeatures == nil {
				st.PremiumFeatures = map[string]room.PremiumFeature{}
			}
			st.PremiumFeatures[feature] = info
		}
	}
	st.AgentRoomStopped = s.agentsBlockedLocked(st, rid)
	st.AgentsStopped = st.AgentsStopped || s.stopAll
	st.Incident = st.Incident || s.integrityFault
	if st.Incident {
		for i := range st.Participants {
			if d := st.Participants[i].StatusDetails; d != nil {
				d.LongRunning = false
				d.Reason = "Audit integrity is failed or unverified."
				d.Next = "The owner must inspect History and explicitly authorize verified recovery."
			}
		}
	}
	if st.AgentRoomStopped && !st.Incident {
		for i := range st.Participants {
			if st.Participants[i].RoomID == rid && st.Participants[i].StatusDetails != nil {
				st.Participants[i].StatusDetails.Reason = "Emergency stop is active."
				st.Participants[i].StatusDetails.Next = "The owner must explicitly resume execution after investigating."
			}
		}
	}
	s.mu.Unlock()
	reply(w, visibleState(st, requestActor(r), rid))
}
func (s *server) events(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		paged := r.URL.Query().Get("page") == "true"
		sql := "SELECT sequence,kind,committed_at,hash,payload FROM events ORDER BY sequence DESC LIMIT 200"
		var args []any
		if paged {
			var err error
			args, err = auditPageBounds(r.URL.Query())
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			sql = auditPageSQL
		}
		refined, flagErr := s.historyPresentationEnabled(r.Context())
		if flagErr != nil {
			http.Error(w, "History settings unavailable", 500)
			return
		}
		rows, err := s.db.Query(r.Context(), sql, args...)
		if err != nil {
			http.Error(w, "Unable to load events", 500)
			return
		}
		defer rows.Close()
		result := []map[string]any{}
		for rows.Next() {
			var seq int64
			var kind, hash string
			var payload []byte
			var at time.Time
			if err = rows.Scan(&seq, &kind, &at, &hash, &payload); err != nil {
				http.Error(w, "Unable to read event", 500)
				return
			}
			event := map[string]any{"sequence": seq, "kind": kind, "at": at, "hash": hash, "summary": auditSummary(kind, payload)}
			if refined {
				event["presentation"] = describeAuditEvent(kind, payload)
			}
			result = append(result, event)
		}
		if rows.Err() != nil {
			http.Error(w, "Unable to read events", 500)
			return
		}
		if paged {
			var next any
			if len(result) > 100 {
				result = result[:100]
				next = result[99]["sequence"]
			}
			reply(w, map[string]any{"events": result, "nextBefore": next})
			return
		}
		reply(w, result)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	var req struct {
		TemplateID string          `json:"templateId"`
		Kind       string          `json:"kind"`
		Payload    json.RawMessage `json:"payload"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	if req.TemplateID != "" {
		if !s.templatesEnabled {
			http.Error(w, "Workflow templates are disabled", 403)
			return
		}
		if req.Kind != "workflow.revised" || !knownWorkflowTemplate(req.TemplateID) {
			http.Error(w, "Unknown workflow template", 400)
			return
		}
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	roomID := requestedRoom(r)
	if !hasRoom(st, roomID) {
		http.Error(w, "Room not found", 404)
		return
	}
	if requestActor(r) != ownerID {
		allowed := req.Kind == "actor.help.updated" || req.Kind == "actor.image.updated" || req.Kind == "actor.name.updated" || req.Kind == "message.created" || req.Kind == "memory.revised" || req.Kind == "task.revised" || req.Kind == "message.attention.acknowledged"
		a, _ := member(st, requestActor(r))
		if req.Kind == "workflow.revised" && a.Role == "Developer" {
			allowed = true
		}
		if !allowed {
			http.Error(w, "Owner permission required", 403)
			return
		}
		if req.Kind == "memory.revised" {
			var mem room.Memory
			json.Unmarshal(req.Payload, &mem)
			oldShared := false
			for _, old := range st.Memories {
				if old.ID == mem.ID && old.Scope != "Room" {
					oldShared = true
				}
			}
			if mem.Scope != "Room" || oldShared {
				http.Error(w, "Shared memory scope changes require owner", 403)
				return
			}
		}
		if req.Kind == "task.revised" {
			var task room.Task
			json.Unmarshal(req.Payload, &task)
			if task.Status == "Accepted" {
				http.Error(w, "Acceptance currently requires owner", 403)
				return
			}
		}
	}
	var p any
	switch req.Kind {
	case "feature.flag.updated":
		var v struct {
			Name    string `json:"name"`
			Enabled *bool  `json:"enabled"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || (!knownPremiumFeature(v.Name) && v.Name != "agentRecovery" && v.Name != "agentBackgroundWork" && v.Name != "agentSafetyControls" && v.Name != "humanProfileImages" && v.Name != "connectionSignIn" && v.Name != "connectionTesting" && v.Name != "helpWindowChoice" && v.Name != "squadProfileActions" && v.Name != "agentInvitationLayout" && v.Name != "standardHelpIcons" && v.Name != "featureFlagTabs" && v.Name != "historyDateGroups" && v.Name != "settingsTabs" && v.Name != "roomSections" && v.Name != "humanProfileDetails" && v.Name != "topAlignedViews" && v.Name != "compactProfileOverlays" && v.Name != "companyOrgChart" && v.Name != "newestMessagesFirst" && v.Name != "roomViewportFit" && v.Name != "featureFlagSwitches" && v.Name != "workflowExecution" && v.Name != "agentStatusDetails" && v.Name != "squadMenuLayout" && v.Name != "humanAttention" && v.Name != "historyRefinement") || v.Enabled == nil {
			http.Error(w, "Choose a supported feature flag and an enabled value", 400)
			return
		}
		if knownPremiumFeature(v.Name) && *v.Enabled && !s.premiumFeatureInfo(st, v.Name, time.Now()).Entitled {
			http.Error(w, "Purchase and a valid signed entitlement are required before enabling this feature", 402)
			return
		}
		p = v
	case "actor.help.updated":
		var v struct {
			ID   string `json:"id"`
			Mode string `json:"mode"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || (v.Mode != "same-window" && v.Mode != "separate-window") {
			http.Error(w, "Choose same-window or separate-window", 400)
			return
		}
		if enabled, ok := st.FeatureFlags["helpWindowChoice"]; ok && !enabled {
			http.Error(w, "Help window choice is disabled", 403)
			return
		}
		v.ID = requestActor(r)
		if a, ok := member(st, v.ID); !ok || a.Kind != "human" {
			http.Error(w, "Human profile required", 403)
			return
		}
		p = v
	case "actor.image.updated":
		var v struct {
			ID    string  `json:"id"`
			Image *string `json:"image"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || v.Image == nil {
			http.Error(w, "Image required (empty removes photo)", 400)
			return
		}
		v.ID = requestActor(r)
		if a, ok := member(st, v.ID); !ok || a.Kind != "human" {
			http.Error(w, "Human profile required", 403)
			return
		}
		if !humanProfileImagesEnabled(st) && *v.Image != "" {
			http.Error(w, "Profile photos are disabled", 403)
			return
		}
		image, err := normalizeProfileImage(*v.Image)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		v.Image = &image
		p = v

	case "actor.name.updated":
		var v struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 {
			http.Error(w, "Name required (maximum 100 characters)", 400)
			return
		}
		v.ID = requestActor(r)
		if v.ID == ownerID {
			http.Error(w, "Use owner profile settings", 400)
			return
		}
		v.Name = strings.TrimSpace(v.Name)
		p = v

	case "company.organization.updated":
		if enabled, exists := st.FeatureFlags["companyOrgChart"]; exists && !enabled {
			http.Error(w, "Company organization chart is disabled", 403)
			return
		}
		var v room.OrganizationChart
		if json.Unmarshal(req.Payload, &v) != nil {
			http.Error(w, "Invalid organization chart", 400)
			return
		}
		if v.Revision != st.Organization.Revision {
			http.Error(w, "Organization changed; reload before saving", 409)
			return
		}
		if err := validateOrganization(v, st.Actors); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		v.Revision++
		p = v
	case "company.updated":
		var v room.Company
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 || len(v.Description) > 2000 {
			http.Error(w, "Company name required (100 characters); description maximum 2000 characters", 400)
			return
		}
		v.ID = st.Company.ID
		v.Name = strings.TrimSpace(v.Name)
		v.Description = strings.TrimSpace(v.Description)
		p = v
	case "human.profile.updated":
		var v struct {
			Name  string  `json:"name"`
			Title *string `json:"title"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 {
			http.Error(w, "Human name required (maximum 100 characters)", 400)
			return
		}
		if v.Title == nil {
			title := st.HumanTitle
			v.Title = &title
		}
		title := strings.TrimSpace(*v.Title)
		if len(title) > 100 {
			http.Error(w, "Title must be at most 100 characters", 400)
			return
		}
		v.Title = &title
		v.Name = strings.TrimSpace(v.Name)
		p = v
	case "workflow.revised":
		var v room.Workflow
		if json.Unmarshal(req.Payload, &v) != nil {
			http.Error(w, "Invalid workflow", 400)
			return
		}
		v.RoomID = roomID
		if e := validateGraph(v); e != nil {
			http.Error(w, e.Error(), 400)
			return
		}
		if v.ID == "" {
			v.ID = id()
			v.Revision = 1
		} else {
			found := false
			for _, old := range st.Workflows {
				if old.ID == v.ID && old.RoomID == roomID {
					found = true
					if old.Revision != v.Revision {
						http.Error(w, "Workflow changed; reload before saving", 409)
						return
					}
					v.Revision++
				}
			}
			if !found {
				http.Error(w, "Workflow not found", 404)
				return
			}
		}
		p = v

	case "room.created", "room.updated":
		var v room.Room
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" {
			http.Error(w, "Room name required", 400)
			return
		}
		if req.Kind == "room.created" {
			v.ID = id()
		} else if !hasRoom(st, v.ID) {
			http.Error(w, "Room not found", 404)
			return
		}
		if v.HumanPresent == nil {
			present := true
			for _, old := range st.Rooms {
				if old.ID == v.ID && old.HumanPresent != nil {
					present = *old.HumanPresent
				}
			}
			v.HumanPresent = &present
		}
		if v.HumanRank == nil {
			rank := 11
			for _, old := range st.Rooms {
				if old.ID == v.ID && old.HumanRank != nil {
					rank = *old.HumanRank
				}
			}
			v.HumanRank = &rank
		}
		if *v.HumanRank < 1 || *v.HumanRank > 11 {
			http.Error(w, "Human seniority must be 1–11", 400)
			return
		}
		v.HumanName = strings.TrimSpace(v.HumanName)
		if v.HumanName == "" {
			for _, old := range st.Rooms {
				if old.ID == v.ID {
					v.HumanName = old.HumanName
				}
			}
		}
		v.Name = strings.TrimSpace(v.Name)
		p = v
	case "task.revised":
		var v room.Task
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Title) == "" {
			http.Error(w, "Task title required", 400)
			return
		}
		v.RoomID = roomID
		switch v.Status {
		case "Assigned", "Working", "Submitted", "Reviewing", "Changes requested", "Accepted":
		default:
			http.Error(w, "Invalid task status", 400)
			return
		}
		for _, pid := range []string{v.Assignee, v.Reviewer} {
			if pid != "" && !hasMember(st, roomID, pid) {
				http.Error(w, "Choose a member of this room", 400)
				return
			}
		}
		if v.ID == "" {
			v.ID = id()
			v.Revision = 1
		} else {
			found := false
			for _, old := range st.Tasks {
				if old.ID == v.ID && old.RoomID == roomID {
					found = true
					if old.Revision != v.Revision {
						http.Error(w, "Task changed; reload", 409)
						return
					}
					v.Revision++
				}
			}
			if !found {
				http.Error(w, "Task not found", 404)
				return
			}
		}
		p = v
	case "evidence.created":
		var v room.Evidence
		if json.Unmarshal(req.Payload, &v) != nil || v.Title == "" || v.Body == "" {
			http.Error(w, "Evidence title and content required", 400)
			return
		}
		found := false
		for _, t := range st.Tasks {
			if t.ID == v.TaskID && t.RoomID == roomID {
				found = true
			}
		}
		if !found {
			http.Error(w, "Task not found", 404)
			return
		}
		switch v.Kind {
		case "Changes", "Checks", "Artifacts", "Review":
		default:
			http.Error(w, "Invalid evidence category", 400)
			return
		}
		v.ID = id()
		v.RoomID = roomID
		v.Source = "Human-provided"
		v.Created = time.Now().UTC()
		p = v

	case "connection.deleted":
		var v struct {
			ID    string `json:"id"`
			Actor string `json:"actor"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || v.ID == "" {
			http.Error(w, "Connection ID required", 400)
			return
		}
		found := false
		for _, c := range st.Connections {
			if c.ID == v.ID {
				found = true
			}
		}
		if !found {
			http.Error(w, "Connection not found", 404)
			return
		}
		for _, member := range st.Participants {
			if member.ConnectionID == v.ID && s.active[member.ID] {
				http.Error(w, "Wait for connected agents to finish before deleting the connection", 409)
				return
			}
		}
		v.Actor = "human"
		p = v
	case "connection.created", "connection.updated":
		var v room.Connection
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" {
			http.Error(w, "Name required", 400)
			return
		}
		if v.Location != "" && v.Location != "Local" && v.Location != "LAN" && v.Location != "Remote/Internet" {
			http.Error(w, "Invalid connection location", 400)
			return
		}
		switch v.Protocol {
		case "CLI", "HTTP", "WebSocket", "MCP":
		default:
			http.Error(w, "Unsupported protocol", 400)
			return
		}

		v.Name = strings.TrimSpace(v.Name)
		v.Endpoint = strings.TrimSpace(v.Endpoint)
		if v.Endpoint == "" {
			http.Error(w, "Endpoint required", 400)
			return
		}
		if req.Kind == "connection.created" {
			v.ID = id()
		} else {
			found := false
			for _, old := range st.Connections {
				if old.ID == v.ID {
					found = true
					if old.Endpoint != v.Endpoint || old.Protocol != v.Protocol {
						for _, member := range st.Participants {
							if member.ConnectionID == v.ID && s.active[member.ID] {
								http.Error(w, "Wait for connected agents to finish before changing the endpoint or protocol", 409)
								return
							}
						}
					}
				}
			}
			if !found {
				http.Error(w, "Connection not found", 404)
				return
			}
		}
		p = v
	case "participant.created", "participant.updated":
		if st.Incident || s.agentsBlockedLocked(st, roomID) {
			http.Error(w, "Invitations/profile activation blocked during an incident or emergency stop", 409)
			return
		}
		var v room.Participant
		if json.Unmarshal(req.Payload, &v) != nil || v.Name == "" {
			http.Error(w, "Name required", 400)
			return
		}
		v.RoomID = roomID
		if v.Rank < 1 || v.Rank > 10 {
			http.Error(w, "Agent seniority must be 1–10", 400)
			return
		}
		if v.Challenge == nil {
			n := 1
			v.Challenge = &n
		}
		if *v.Challenge < 0 || *v.Challenge > 4 {
			http.Error(w, "Challenge must be 0–4", 400)
			return
		}
		if len(v.AdditionalContext) > 40000 {
			http.Error(w, "Additional context exceeds 40,000 bytes; select an excerpt", 400)
			return
		}
		found := false
		for _, c := range st.Connections {
			if c.ID == v.ConnectionID {
				found = true
			}
		}
		if !found {
			http.Error(w, "Register a connection first", 400)
			return
		}
		if req.Kind == "participant.created" {
			v.ThreadID = ""
			v.ExecutionMode = ""
			v.Status = ""
			for _, c := range st.Connections {
				if c.ID == v.ConnectionID && room.SupportsAgent(c) {
					v.Status = "Idle"
				}
			}
			v.ID = id()
		} else {
			found = false
			for _, b := range st.Participants {
				if b.ID == v.ID && b.RoomID == roomID {
					if v.ConnectionID != b.ConnectionID {
						if s.active[b.ID] {
							http.Error(w, "Cannot reassign a working agent", 409)
							return
						}
						v.ThreadID = ""
						v.ExecutionMode = ""
						v.Status = "Idle"
					} else {
						v.ThreadID = b.ThreadID
						v.ExecutionMode = b.ExecutionMode
						v.Status = b.Status
					}
					found = true
				}
			}
			if !found {
				http.Error(w, "Participant not found", 404)
				return
			}
		}
		p = v
	case "message.attention.acknowledged":
		actor, ok := member(st, requestActor(r))
		if !ok || actor.Kind != "human" {
			http.Error(w, "Human account required", 403)
			return
		}
		if !humanAttentionEnabled(st) {
			http.Error(w, "Human attention is disabled", 403)
			return
		}
		var reqAck struct {
			MessageID string `json:"messageId"`
		}
		if json.Unmarshal(req.Payload, &reqAck) != nil {
			http.Error(w, "Invalid acknowledgment", 400)
			return
		}
		found := false
		for _, message := range st.Messages {
			if message.ID == reqAck.MessageID && message.RoomID == roomID && message.ActorType == "agent" && message.HumanRequest != "" {
				found = true
				if message.AttentionAcknowledgedBy != "" {
					reply(w, map[string]bool{"committed": true})
					return
				}
			}
		}
		if !found {
			http.Error(w, "Agent request not found in this room", 404)
			return
		}
		p = map[string]string{"messageId": reqAck.MessageID, "roomId": roomID, "actorId": requestActor(r)}
	case "message.created":
		var v room.Message
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Text) == "" {
			http.Error(w, "Message required", 400)
			return
		}
		v.RoomID = roomID
		v.ID = id()
		v.ActorID = requestActor(r)
		v.ActorType = "human"
		v.HumanRequest = ""
		v.AttentionAcknowledgedBy = ""
		actor, _ := member(st, v.ActorID)
		v.Author = actor.Name
		v.Created = time.Now().UTC()
		p = v
	case "memory.revised":
		var v room.Memory
		if json.Unmarshal(req.Payload, &v) != nil || v.Title == "" {
			http.Error(w, "Title required", 400)
			return
		}
		v.RoomID = roomID
		if v.Scope != "Room" && v.Scope != "Project" && v.Scope != "Application" {
			http.Error(w, "Invalid scope", 400)
			return
		}
		if v.ID == "" {
			v.ID = id()
			v.Revision = 1
		} else {
			found := false
			for _, m := range st.Memories {
				if m.ID == v.ID && (m.RoomID == roomID || m.Scope != "Room") {
					v.RoomID = m.RoomID
					found = true
					if v.Revision != m.Revision {
						http.Error(w, "Memory conflict: reload current revision", 409)
						return
					}
					v.Revision++
				}
			}
			if !found {
				http.Error(w, "Memory not found", 404)
				return
			}
		}
		p = v
	default:
		http.Error(w, "Unsupported operation", 400)
		return
	}
	if err = s.append(r.Context(), req.Kind, p); err != nil {
		log.Print(err)
		http.Error(w, "Commit failed", 500)
		return
	}
	if req.Kind == "feature.flag.updated" {
		var flag struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		}
		json.Unmarshal(req.Payload, &flag)
		if flag.Name == "workflowExecution" && !flag.Enabled {
			for _, cancel := range s.workflowLive {
				cancel()
			}
		}
	}
	result := map[string]any{"committed": true}
	if workflow, ok := p.(room.Workflow); ok {
		result["workflowId"] = workflow.ID
	}
	reply(w, result)
}
func (s *server) integrity(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.Method == "POST" {
		s.recoverIntegrityLocked(w, r)
		return
	}
	s.integrityLocked(w, r)
}

// integrityLocked requires s.mu; every verification failure stops execution,
// including failures that prevent the incident from being committed.
func (s *server) integrityLocked(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	rows, err := s.db.Query(r.Context(), "SELECT id,sequence,committed_at,kind,payload::text,previous_hash,hash FROM events ORDER BY sequence")
	if err != nil {
		s.tripIntegrityLocked()
		http.Error(w, "Integrity unverified", 503)
		return
	}
	previous := ""
	count := 0
	valid := true
	protectedCount := 0
	for rows.Next() {
		var kind, payload, prev, hash string
		var eventID string
		var sequence int64
		var committedAt time.Time
		if err = rows.Scan(&eventID, &sequence, &committedAt, &kind, &payload, &prev, &hash); err != nil {
			rows.Close()
			s.tripIntegrityLocked()
			http.Error(w, "Integrity unverified", 503)
			return
		}
		if prev != previous || hash != checksum(prev, kind, payload) {
			valid = false
		}
		protected, metadataValid := verifyAuditMetadata(payload, eventID, sequence, committedAt)
		if protected {
			protectedCount++
		}
		if !metadataValid {
			valid = false
		}
		previous = hash
		count++
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		s.tripIntegrityLocked()
		http.Error(w, "Integrity unverified", 503)
		return
	}
	if !valid {
		s.tripIntegrityLocked()
	}
	st, e := s.load(r.Context())
	if e != nil {
		s.tripIntegrityLocked()
		http.Error(w, "Integrity state unverified", 503)
		return
	}
	if st.Incident {
		s.tripIntegrityLocked()
	}
	if s.integrityFault && !st.Incident {
		if s.append(r.Context(), "incident.opened", map[string]string{"reason": "audit verification failed or was unavailable; owner review required"}) != nil {
			http.Error(w, "Incident audit unavailable; execution remains stopped", 503)
			return
		}
	}
	reply(w, map[string]any{"valid": valid, "records": count, "metadataProtectedRecords": protectedCount, "legacyMetadataRecords": count - protectedCount, "executionBlocked": s.integrityFault, "protection": "Local chain only; independent immutable archive and rollback anchor still required"})
}
