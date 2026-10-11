package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"agentroom/internal/room"
	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	db      *pgxpool.Pool
	auth    *ownerAuth
	mu      sync.Mutex
	active  map[string]bool
	cancels map[string]context.CancelFunc
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
	ctx := context.Background()
	var localConfig struct {
		DatabaseURL string `json:"databaseUrl"`
	}
	if raw, e := os.ReadFile(".data/local-config.json"); e == nil {
		if e = json.Unmarshal(raw, &localConfig); e != nil {
			log.Fatal("Invalid local configuration: ", e)
		}
	} else if !os.IsNotExist(e) {
		log.Fatal(e)
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = localConfig.DatabaseURL
	}
	if url == "" {
		url = "postgres://localhost:5432/agentroom?sslmode=disable"
	}
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		log.Fatal("PostgreSQL unavailable: ", err)
	}
	migration, err := os.ReadFile("migrations/001.sql")
	if err != nil {
		log.Fatal(err)
	}
	if _, err = db.Exec(ctx, string(migration)); err != nil {
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
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/integrity", s.integrity)
	mux.HandleFunc("/api/agents", s.agents)
	mux.HandleFunc("/api/provider-auth", s.providerAuth)
	mux.Handle("/", http.FileServer(http.Dir("web")))
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8787"
	}
	keyPath := os.Getenv("OWNER_KEY_FILE")
	if keyPath == "" {
		keyPath = ".data/owner-key"
	}
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
	log.Printf("Owner setup code file: %s (keep private)", keyPath)
	log.Printf("Commonroom: http://%s (local audit protection; external immutable archive not configured)", addr)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; connect-src 'self'; frame-ancestors 'none'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		s.auth.protect(s.actorGuard(mux)).ServeHTTP(w, r)
	})
	log.Fatal((&http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}).ListenAndServe())
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
	var canonical string
	if err = tx.QueryRow(ctx, "SELECT $1::jsonb::text", string(raw)).Scan(&canonical); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO events(id,kind,payload,previous_hash,hash) VALUES($1,$2,$3::jsonb,$4,$5)", id(), kind, canonical, previous, checksum(previous, kind, canonical))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *server) load(ctx context.Context) (room.State, error) {
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
					st.Actors[i] = v
					found = true
				}
			}
			if !found {
				st.Actors = append(st.Actors, v)
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
		case "participant.updated", "participant.session":
			var v room.Participant
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			for i, p := range st.Participants {
				if p.ID == v.ID {
					st.Participants[i] = v
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
		case "incident.opened":
			st.Incident = true
		case "incident.recovered":
			st.Incident = false
		}
	}
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
	reply(w, visibleState(st, requestActor(r), rid))
}
func (s *server) events(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		rows, err := s.db.Query(r.Context(), "SELECT sequence,kind,committed_at,hash FROM events ORDER BY sequence DESC LIMIT 200")
		if err != nil {
			http.Error(w, "Unable to load events", 500)
			return
		}
		defer rows.Close()
		result := []map[string]any{}
		for rows.Next() {
			var seq int64
			var kind, hash string
			var at time.Time
			if err = rows.Scan(&seq, &kind, &at, &hash); err != nil {
				http.Error(w, "Unable to read event", 500)
				return
			}
			result = append(result, map[string]any{"sequence": seq, "kind": kind, "at": at, "hash": hash})
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
		Kind    string          `json:"kind"`
		Payload json.RawMessage `json:"payload"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Invalid JSON", 400)
		return
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
		allowed := req.Kind == "actor.name.updated" || req.Kind == "message.created" || req.Kind == "memory.revised" || req.Kind == "task.revised"
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
		if st.Incident {
			http.Error(w, "Invitations/profile activation blocked during incident", 409)
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
					v.ThreadID = b.ThreadID
					v.Status = b.Status
					found = true
				}
			}
			if !found {
				http.Error(w, "Participant not found", 404)
				return
			}
		}
		p = v
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
	reply(w, map[string]bool{"committed": true})
}
func (s *server) integrity(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	rows, err := s.db.Query(r.Context(), "SELECT kind,payload::text,previous_hash,hash FROM events ORDER BY sequence")
	if err != nil {
		http.Error(w, "Integrity unverified", 503)
		return
	}
	previous := ""
	count := 0
	valid := true
	for rows.Next() {
		var kind, payload, prev, hash string
		if err = rows.Scan(&kind, &payload, &prev, &hash); err != nil {
			rows.Close()
			http.Error(w, "Integrity unverified", 503)
			return
		}
		if prev != previous || hash != checksum(prev, kind, payload) {
			valid = false
		}
		previous = hash
		count++
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		http.Error(w, "Integrity unverified", 503)
		return
	}
	if !valid {
		st, e := s.load(r.Context())
		if e == nil && !st.Incident {
			s.append(r.Context(), "incident.opened", map[string]string{"reason": "checksum or predecessor mismatch"})
		}
	}
	reply(w, map[string]any{"valid": valid, "records": count, "protection": "Local chain only; external immutable archive and authenticated agents not yet implemented"})
}
