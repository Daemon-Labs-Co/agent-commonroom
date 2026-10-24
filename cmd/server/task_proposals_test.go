package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func proposalFixture(t *testing.T, enabled bool) (*server, *mockDB, http.Handler) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("feature.flag.updated", map[string]any{"name": "taskProposals", "enabled": enabled})
	db.seed("message.created", room.Message{ID: "source", RoomID: "engineering", ActorID: "synthetic-agent", ActorType: "agent", Author: "Synthetic author", Text: "A suggestion, not approval"})
	db.seed("room.created", room.Room{ID: "foreign", Name: "Foreign"})
	db.seed("message.created", room.Message{ID: "secret", RoomID: "foreign", ActorID: "other", ActorType: "human", Author: "Other", Text: "Synthetic foreign content"})
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-proposals", s.taskProposals)
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/task-dispatch", s.taskDispatch)
	mux.HandleFunc("/api/events", s.events)
	return s, db, s.auth.protect(s.actorGuard(mux))
}
func proposalRequest(h http.Handler, actor, key, body string) *httptest.ResponseRecorder {
	return proposalRoomRequest(h, actor, "engineering", key, body)
}
func proposalRoomRequest(h http.Handler, actor, rid, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/task-proposals?room="+rid, strings.NewReader(body))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: actor})
	r.Header.Set("X-CSRF-Token", "csrf")
	r.Header.Set("Idempotency-Key", key)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

const captureProposal = `{"action":"capture","sourceMessageId":"source","title":"Synthetic proposal","criteria":""}`

func capturedProposal(t *testing.T, h http.Handler) room.TaskProposal {
	t.Helper()
	w := proposalRequest(h, "owner", "capture-proposal", captureProposal)
	var p room.TaskProposal
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	return p
}
func approveProposalBody(p room.TaskProposal) string {
	return fmt.Sprintf(`{"action":"approve","proposalId":%q,"revision":%d,"title":"Reviewed task","criteria":"Explicit synthetic acceptance"}`, p.ID, p.Revision)
}
func TestTaskProposalSeparatesCaptureApprovalAndExecution(t *testing.T) {
	t.Parallel()
	s, db, h := proposalFixture(t, true)
	p := capturedProposal(t, h)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Tasks) != 0 || len(st.ManagedJobs) != 0 || len(s.active) != 0 || p.SourceActorID != "synthetic-agent" || p.SourceAuthor != "Synthetic author" || p.CapturedBy != ownerID || p.Status != "Proposed" {
		t.Fatal("capture became approval or lost authorship", p)
	}
	original := db.snapshot()
	w := proposalRequest(h, "owner", "capture-proposal", captureProposal)
	if w.Code != 200 || w.Header().Get("Idempotent-Replay") != "true" || !reflect.DeepEqual(original, db.snapshot()) {
		t.Fatal("capture retry duplicated proposal")
	}
	body := approveProposalBody(p)
	w = proposalRequest(h, "owner", "approve-proposal", body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err = (&server{db: db}).load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Tasks) != 1 || st.Tasks[0].ProposalID != p.ID || st.Tasks[0].Status != "Draft" || st.Tasks[0].Assignee != "" || len(st.ManagedJobs) != 0 || len(s.active) != 0 {
		t.Fatal("approval dispatched work or lost draft")
	}
	task := st.Tasks[0]
	if st.TaskProposals[p.ID].ReviewedBy != ownerID || st.TaskProposals[p.ID].TaskID != task.ID {
		t.Fatal("approval link missing")
	}
	before := db.snapshot()
	w = proposalRequest(h, "owner", "approve-proposal", body)
	if w.Code != 200 || w.Header().Get("Idempotent-Replay") != "true" || !reflect.DeepEqual(before, db.snapshot()) {
		t.Fatal("approval retry created another task")
	}
	w = proposalRequest(h, "owner", "changed-approval", body)
	if w.Code != 409 {
		t.Fatal("second decision accepted", w.Code)
	}
	w = dispatchRequest(h, "dispatch-draft", fmt.Sprintf(`{"taskId":%q,"revision":1,"mode":"work"}`, task.ID))
	if w.Code != 409 || len(s.active) != 0 {
		t.Fatal("draft dispatched", w.Code)
	}
	// Generic task edits preserve controller-owned source linkage even if omitted.
	w = mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":{"id":%q,"revision":1,"title":"Human edit","criteria":"Reviewed acceptance","status":"Assigned"}}`, task.ID))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err = s.load(t.Context())
	if err != nil || st.Tasks[0].ProposalID != p.ID {
		t.Fatal("task edit erased provenance", err)
	}
}
func TestTaskProposalRefusalsAndScopedRead(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"off", "agent", "human", "foreign-source", "missing-source", "forged-author", "duplicate-field", "stop", "no-key"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, h := proposalFixture(t, mode != "off")
			body := captureProposal
			actor := "owner"
			key := "capture-refusal"
			want := 400
			switch mode {
			case "off":
				want = 403
			case "agent", "human":
				actor = mode
				db.seed("actor.created", room.Actor{ID: mode, Kind: mode, Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
				s.auth.sessions[mode] = ownerSession{ActorID: mode, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
				want = 403
			case "foreign-source":
				body = strings.Replace(body, `"source"`, `"secret"`, 1)
				want = 404
			case "missing-source":
				body = strings.Replace(body, `"source"`, `"absent"`, 1)
				want = 404
			case "forged-author":
				body = strings.TrimSuffix(body, "}") + `,"sourceActorId":"owner"}`
			case "duplicate-field":
				body = strings.TrimSuffix(body, "}") + `,"title":"duplicate"}`
			case "stop":
				s.stopAll = true
				want = 409
			case "no-key":
				key = ""
			}
			before := db.snapshot()
			w := proposalRequest(h, actor, key, body)
			if w.Code != want || !reflect.DeepEqual(before, db.snapshot()) {
				t.Fatal("invalid proposal admitted", w.Code, w.Body.String())
			}
		})
	}
	_, _, h := proposalFixture(t, true)
	capturedProposal(t, h)
	w := mockRequest(h, "GET", "/api/task-proposals", "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "Synthetic foreign content") || !strings.Contains(w.Body.String(), "A suggestion, not approval") {
		t.Fatal("source read leaked or missing", w.Code, w.Body.String())
	}
}
func TestTaskProposalDecisionAtomicityRejectionAndFeatureOff(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"begin", "insert", "commit"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()
			s, db, h := proposalFixture(t, true)
			p := capturedProposal(t, h)
			before := db.snapshot()
			db.failAt = failure
			w := proposalRequest(h, "owner", "failed-approval", approveProposalBody(p))
			if w.Code != 500 || !s.integrityFault || !reflect.DeepEqual(before, db.snapshot()) {
				t.Fatal("approval failure partially committed", w.Code)
			}
			db.failAt = ""
			st, err := s.load(t.Context())
			if err != nil || len(st.Tasks) != 0 || st.TaskProposals[p.ID].Status != "Proposed" {
				t.Fatal("failed decision changed state", err)
			}
		})
	}
	s, db, h := proposalFixture(t, true)
	p := capturedProposal(t, h)
	body := fmt.Sprintf(`{"action":"reject","proposalId":%q,"revision":1}`, p.ID)
	w := proposalRequest(h, "owner", "reject-proposal", body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	db.seed("feature.flag.updated", map[string]any{"name": "taskProposals", "enabled": false})
	st, err := s.load(t.Context())
	if err != nil || st.TaskProposals[p.ID].Status != "Rejected" || len(st.Tasks) != 0 {
		t.Fatal("rejection/off erased history", err)
	}
	if w = proposalRequest(h, "owner", "approve-rejected", approveProposalBody(p)); w.Code != 403 {
		t.Fatal("off permitted decision", w.Code)
	}
	if w = mockRequest(h, "GET", "/api/task-proposals", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Rejected") {
		t.Fatal("off hid retained history")
	}
}
func TestTaskProposalConcurrentApprovalsCreateOneDraft(t *testing.T) {
	t.Parallel()
	s, db, h := proposalFixture(t, true)
	p := capturedProposal(t, h)
	body := approveProposalBody(p)
	other := &server{db: db, auth: s.auth}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-proposals", other.taskProposals)
	h2 := other.auth.protect(other.actorGuard(mux))
	var wg sync.WaitGroup
	codes := make(chan int, 2)
	for i, handler := range []http.Handler{h, h2} {
		wg.Add(1)
		go func(i int, h http.Handler) {
			defer wg.Done()
			codes <- proposalRequest(h, "owner", fmt.Sprintf("approval-race-%d", i), body).Code
		}(i, handler)
	}
	wg.Wait()
	close(codes)
	ok, conflict := 0, 0
	for code := range codes {
		if code == 200 {
			ok++
		} else if code == 409 {
			conflict++
		} else {
			t.Fatal(code)
		}
	}
	st, err := s.load(context.Background())
	if err != nil || ok != 1 || conflict != 1 || len(st.Tasks) != 1 || s.integrityFault || other.integrityFault {
		t.Fatal("approval race duplicated or stopped", err, ok, conflict)
	}
}

func TestTaskProposalCannotForgeOrStripProvenance(t *testing.T) {
	t.Parallel()
	s, db, h := proposalFixture(t, true)
	p := capturedProposal(t, h)
	w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":{"title":"Forged task","status":"Assigned","proposalId":%q}}`, p.ID))
	if w.Code != 400 {
		t.Fatal("generic task claimed proposal approval", w.Code, w.Body.String())
	}
	w = mockRequest(h, "POST", "/api/events", `{"kind":"task.proposal.approved","payload":{}}`)
	if w.Code != 400 {
		t.Fatal("generic event injected approval", w.Code)
	}
	w = proposalRequest(h, "owner", "real-approval", approveProposalBody(p))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task := st.Tasks[0]
	task.ProposalID = ""
	db.seed("task.revised", task)
	if _, err = s.load(t.Context()); err == nil {
		t.Fatal("replay accepted stripped provenance")
	}
}
func TestTaskProposalCaptureFailureAndApprovalDisableRace(t *testing.T) {
	t.Parallel()
	t.Run("capture-failure", func(t *testing.T) {
		t.Parallel()
		s, db, h := proposalFixture(t, true)
		before := db.snapshot()
		db.failAt = "commit"
		w := proposalRequest(h, "owner", "capture-failure", captureProposal)
		if w.Code != 500 || !s.integrityFault || !reflect.DeepEqual(before, db.snapshot()) {
			t.Fatal("capture failure was not atomic", w.Code)
		}
	})
	t.Run("disable-race", func(t *testing.T) {
		t.Parallel()
		s, db, h := proposalFixture(t, true)
		p := capturedProposal(t, h)
		other := &server{db: db, auth: s.auth}
		mux := http.NewServeMux()
		mux.HandleFunc("/api/events", other.events)
		otherHandler := other.auth.protect(other.actorGuard(mux))
		once := false
		s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(context.Context) {
			if !once {
				once = true
				w := mockRequest(otherHandler, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskProposals","enabled":false}}`)
				if w.Code != 200 {
					t.Fatal(w.Code)
				}
			}
		}}
		w := proposalRequest(h, "owner", "approval-disable", approveProposalBody(p))
		if w.Code != 409 || s.integrityFault {
			t.Fatal("stale approval crossed disable", w.Code)
		}
		st, err := s.load(t.Context())
		if err != nil || len(st.Tasks) != 0 || st.TaskProposals[p.ID].Status != "Proposed" {
			t.Fatal("disable race created task", err)
		}
	})
}
func TestTaskProposalFlagDefaultsOffAndOwnerCanEnable(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/task-proposals", s.taskProposals)
	h := s.auth.protect(s.actorGuard(mux))
	w := mockRequest(h, "GET", "/api/task-proposals", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"enabled":false`) {
		t.Fatal("default not off", w.Code)
	}
	w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskProposals","enabled":true}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil || !st.FeatureFlags["taskProposals"] {
		t.Fatal("owner enable failed", err)
	}
	db.seed("actor.created", room.Actor{ID: "human", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["human"] = ownerSession{ActorID: "human", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/events", strings.NewReader(`{"kind":"feature.flag.updated","payload":{"name":"taskProposals","enabled":false}}`))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "human"})
	r.Header.Set("X-CSRF-Token", "csrf")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("non-owner changed flag", w.Code)
	}
}

func TestProposalDraftCannotDispatchEvenAfterAssigneeEdit(t *testing.T) {
	t.Parallel()
	s, db, h := proposalFixture(t, true)
	p := capturedProposal(t, h)
	db.seed("connection.created", room.Connection{ID: "connection", Protocol: "CLI", Endpoint: "ssh://synthetic-worker"})
	db.seed("participant.created", room.Participant{ID: "dev", RoomID: "engineering", ConnectionID: "connection", Status: "Idle"})
	w := proposalRequest(h, "owner", "approve-draft", approveProposalBody(p))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task := st.Tasks[0]
	w = mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":{"id":%q,"revision":1,"title":"Assigned draft","criteria":"Acceptance","assignee":"dev","status":"Draft"}}`, task.ID))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	s.taskReady = func(string) error { t.Error("draft reached runtime admission"); return fmt.Errorf("synthetic refusal") }
	w = dispatchRequest(h, "dispatch-assigned-draft", fmt.Sprintf(`{"taskId":%q,"revision":2,"mode":"work"}`, task.ID))
	if w.Code != 409 || len(s.active) != 0 {
		t.Fatal("assigned draft reached dispatch", w.Code)
	}
}

func TestPersonalRoomTaskProposalsReplayWithoutProfileOrHelpEvents(t *testing.T) {
	t.Parallel()
	for _, identity := range []string{"implicit-owner", "member", "connection-agent"} {
		for _, decision := range []string{"approve", "reject"} {
			t.Run(identity+"/"+decision, func(t *testing.T) {
				t.Parallel()
				_, db, h := proposalFixture(t, true)
				actorID := ownerID
				sourceKind := "human"
				switch identity {
				case "member":
					actorID = "member"
					db.seed("actor.revised", room.Actor{ID: actorID, Kind: "human", Role: "Developer", Name: "Synthetic member", Enabled: true})
				case "connection-agent":
					actorID = "bot-native"
					sourceKind = "agent"
					db.seed("connection.created", room.Connection{ID: "native", Name: "Synthetic agent", Protocol: "CLI", Endpoint: "codex"})
				}
				rid := "personal-" + actorID
				db.seed("message.created", room.Message{ID: "personal-source", RoomID: rid, ActorID: actorID, ActorType: sourceKind, Author: "Personal author", Text: "Synthetic personal suggestion"})
				body := `{"action":"capture","sourceMessageId":"personal-source","title":"Private suggestion","criteria":""}`
				w := proposalRoomRequest(h, "owner", rid, "capture-personal", body)
				var p room.TaskProposal
				if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
					t.Fatal(w.Code, w.Body.String())
				}
				// Fresh controller replay must work immediately after the first capture.
				restarted := &server{db: db}
				st, err := restarted.load(t.Context())
				if err != nil || !hasRoom(st, rid) || st.TaskProposals[p.ID].RoomID != rid {
					t.Fatal("personal capture poisoned replay", err)
				}
				w = mockRequest(h, "GET", "/api/state?room="+rid, "")
				if w.Code != 200 {
					t.Fatal("personal state unavailable after capture", w.Code, w.Body.String())
				}
				body = fmt.Sprintf(`{"action":"reject","proposalId":%q,"revision":1}`, p.ID)
				if decision == "approve" {
					body = approveProposalBody(p)
				}
				w = proposalRoomRequest(h, "owner", rid, "decide-personal", body)
				if w.Code != 200 {
					t.Fatal(w.Code, w.Body.String())
				}
				st, err = restarted.load(t.Context())
				if err != nil {
					t.Fatal("personal decision poisoned replay", err)
				}
				want := "Rejected"
				if decision == "approve" {
					want = "Approved"
					if len(st.Tasks) != 1 || st.Tasks[0].RoomID != rid {
						t.Fatal("draft escaped personal room")
					}
				}
				if st.TaskProposals[p.ID].Status != want {
					t.Fatal("decision lost")
				}
				w = mockRequest(h, "GET", "/api/task-proposals?room="+rid, "")
				if w.Code != 200 || !strings.Contains(w.Body.String(), "Synthetic personal suggestion") || strings.Contains(w.Body.String(), "Synthetic foreign content") {
					t.Fatal("personal read failed or leaked", w.Code)
				}
				// Shared-room lists must never include the private proposal/source.
				w = mockRequest(h, "GET", "/api/task-proposals?room=engineering", "")
				if w.Code != 200 || strings.Contains(w.Body.String(), "Synthetic personal suggestion") {
					t.Fatal("private proposal leaked to shared room")
				}
				// Retained private history remains replayable with the feature disabled.
				db.seed("feature.flag.updated", map[string]any{"name": "taskProposals", "enabled": false})
				if _, err = restarted.load(t.Context()); err != nil {
					t.Fatal("disabled private history failed", err)
				}
			})
		}
	}
}

func TestProposalReplayDoesNotInventUnknownOrFuturePersonalRooms(t *testing.T) {
	t.Parallel()
	for _, future := range []bool{false, true} {
		t.Run(fmt.Sprint(future), func(t *testing.T) {
			t.Parallel()
			s, db, _ := proposalFixture(t, true)
			rid := "personal-absent"
			db.seed("message.created", room.Message{ID: "invalid-source", RoomID: rid, ActorID: "absent", Author: "Unregistered", Text: "Synthetic invalid scope"})
			db.seed("task.proposal.captured", map[string]any{"id": "invalid-proposal", "companyId": "company", "roomId": rid, "sourceMessageId": "invalid-source", "sourceActorId": "absent", "sourceActorType": "", "sourceAuthor": "Unregistered", "capturedBy": ownerID, "title": "Invalid scope", "criteria": "", "status": "Proposed", "revision": 1, "performedBy": ownerID})
			if future {
				db.seed("actor.revised", room.Actor{ID: "absent", Kind: "human", Role: "Developer", Enabled: true})
			}
			if _, err := s.load(t.Context()); err == nil {
				t.Fatal("unknown or future room made historical proposal valid")
			}
		})
	}
}

func TestProposalReplayRefreshesPersonalRoomLabelsAfterLaterRenames(t *testing.T) {
	t.Parallel()
	for _, proposalRoom := range []string{"engineering", "personal-member"} {
		t.Run(proposalRoom, func(t *testing.T) {
			t.Parallel()
			s, db, h := proposalFixture(t, true)
			db.seed("actor.revised", room.Actor{ID: "member", Kind: "human", Name: "Earlier member", Role: "Developer", Enabled: true})
			db.seed("human.profile.updated", map[string]string{"name": "Earlier owner"})
			db.seed("connection.created", room.Connection{ID: "renamed", Name: "Earlier connection", Protocol: "CLI", Endpoint: "codex"})
			db.seed("message.created", room.Message{ID: "rename-source", RoomID: proposalRoom, ActorID: "member", ActorType: "human", Author: "Earlier member", Text: "Synthetic suggestion before rename"})
			w := proposalRoomRequest(h, "owner", proposalRoom, "capture-before-rename", `{"action":"capture","sourceMessageId":"rename-source","title":"Before rename","criteria":""}`)
			var p room.TaskProposal
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
				t.Fatal(w.Code, w.Body.String())
			}
			db.seed("actor.name.updated", map[string]string{"id": "member", "name": "Current member"})
			db.seed("human.profile.updated", map[string]string{"name": "Current owner"})
			db.seed("connection.updated", room.Connection{ID: "renamed", Name: "Current connection", Protocol: "CLI", Endpoint: "codex"})
			for _, controller := range []*server{s, {db: db}} {
				st, err := controller.load(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				for actorID, want := range map[string]string{"member": "Current member", ownerID: "Current owner", "bot-renamed": "Current connection"} {
					actor, ok := member(st, actorID)
					if !ok || actor.Name != want {
						t.Fatal("actor label stale", actorID, actor.Name)
					}
					found := false
					for _, r := range st.Rooms {
						if r.ID == "personal-"+actorID {
							found = true
							if r.Name != want+" · Personal" || r.PersonalActorID != actorID {
								t.Fatal("personal room label or identity stale", r)
							}
						}
					}
					if !found {
						t.Fatal("personal room missing", actorID)
					}
				}
				if st.TaskProposals[p.ID].SourceAuthor != "Earlier member" {
					t.Fatal("rename rewrote historical source authorship")
				}
			}
		})
	}
}

func TestPersonalRoomCustomNamesSurviveProposalReplayAndRenames(t *testing.T) {
	t.Parallel()
	for _, timing := range []string{"before-proposal", "after-proposal"} {
		for _, custom := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/custom=%t", timing, custom), func(t *testing.T) {
				t.Parallel()
				s, db, h := proposalFixture(t, true)
				db.seed("actor.revised", room.Actor{ID: "member", Kind: "human", Name: "Earlier member", Role: "Developer", Enabled: true})
				rid := "personal-member"
				saveRoom := func() {
					st, err := s.load(t.Context())
					if err != nil {
						t.Fatal(err)
					}
					var target room.Room
					for _, r := range st.Rooms {
						if r.ID == rid {
							target = r
						}
					}
					if target.ID == "" {
						t.Fatal("personal room missing")
					}
					if custom {
						target.Name = "Research den"
					}
					target.Description = "Saved description"
					body, err := json.Marshal(map[string]any{"kind": "room.updated", "payload": target})
					if err != nil {
						t.Fatal(err)
					}
					w := mockRequest(h, "POST", "/api/events?room="+rid, string(body))
					if w.Code != 200 {
						t.Fatal(w.Code, w.Body.String())
					}
				}
				if timing == "before-proposal" {
					saveRoom()
				}
				capturedProposal(t, h) // Shared-room proposal must not alter personal naming policy.
				if timing == "after-proposal" {
					saveRoom()
				}
				db.seed("actor.name.updated", map[string]string{"id": "member", "name": "Current member"})
				saveRoom() // Later settings saves retain the selected naming policy.
				for _, controller := range []*server{s, {db: db}} {
					st, err := controller.load(t.Context())
					if err != nil {
						t.Fatal(err)
					}
					for _, r := range st.Rooms {
						if r.ID == rid {
							want := "Current member · Personal"
							if custom {
								want = "Research den"
							}
							if r.Name != want || r.Description != "Saved description" || r.PersonalActorID != "member" || r.CustomPersonalName != custom {
								t.Fatal("saved naming policy or settings lost", r)
							}
						}
					}
				}
				// Verify the actual room list response after refresh, not just projection.
				w := mockRequest(h, "GET", "/api/state?room="+rid, "")
				var state room.State
				if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &state) != nil {
					t.Fatal(w.Code, w.Body.String())
				}
				want := "Current member · Personal"
				if custom {
					want = "Research den"
				}
				found := false
				for _, r := range state.Rooms {
					if r.ID == rid {
						found = true
						if r.Name != want {
							t.Fatal("room list label differs", r.Name)
						}
					}
				}
				if !found {
					t.Fatal("room missing in refreshed state")
				}
				if strings.Contains(w.Body.String(), "CustomPersonalName") || strings.Contains(w.Body.String(), "customPersonalName") {
					t.Fatal("internal naming policy exposed")
				}
			})
		}
	}
}

func TestStalePersonalRoomSettingsCannotFreezeAutomaticNames(t *testing.T) {
	t.Parallel()
	for _, identity := range []string{"member", "owner", "connection"} {
		t.Run(identity, func(t *testing.T) {
			t.Parallel()
			s, db, h := proposalFixture(t, true)
			actorID := "member"
			switch identity {
			case "member":
				db.seed("actor.revised", room.Actor{ID: actorID, Name: "Earlier", Kind: "human", Role: "Developer", Enabled: true})
			case "owner":
				actorID = ownerID
				db.seed("human.profile.updated", map[string]string{"name": "Earlier"})
			case "connection":
				actorID = "bot-dialog"
				db.seed("connection.created", room.Connection{ID: "dialog", Name: "Earlier", Protocol: "CLI", Endpoint: "codex"})
			}
			capturedProposal(t, h)
			rid := "personal-" + actorID
			readRoom := func() room.Room {
				t.Helper()
				w := mockRequest(h, "GET", "/api/state?room="+rid, "")
				var st room.State
				if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &st) != nil {
					t.Fatal("room state unavailable", w.Code, w.Body.String())
				}
				for _, r := range st.Rooms {
					if r.ID == rid {
						return r
					}
				}
				t.Fatal("room missing")
				return room.Room{}
			}
			rename := func(name string) {
				switch identity {
				case "member":
					db.seed("actor.name.updated", map[string]string{"id": actorID, "name": name})
				case "owner":
					db.seed("human.profile.updated", map[string]string{"name": name})
				case "connection":
					db.seed("connection.updated", room.Connection{ID: "dialog", Name: name, Protocol: "CLI", Endpoint: "codex"})
				}
			}
			save := func(r room.Room) *httptest.ResponseRecorder {
				body, err := json.Marshal(map[string]any{"kind": "room.updated", "payload": r})
				if err != nil {
					t.Fatal(err)
				}
				return mockRequest(h, "POST", "/api/events?room="+rid, string(body))
			}
			stale := readRoom()
			rename("Current")
			current := readRoom()
			if stale.Revision != current.Revision || stale.NameVersion == "" || stale.NameVersion == current.NameVersion {
				t.Fatal("derived-name precondition not independent of room revision")
			}
			stale.Description = "Only description edited"
			before := db.snapshot()
			w := save(stale)
			if w.Code != 409 || !reflect.DeepEqual(before, db.snapshot()) {
				t.Fatal("stale dialog froze name", w.Code, w.Body.String())
			}
			current.Description = "Fresh description"
			w = save(current)
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			rename("Latest")
			latest := readRoom()
			if latest.Name != "Latest · Personal" || latest.Description != "Fresh description" {
				t.Fatal("description save disabled automatic labels", latest)
			}
			// A fresh intentional rename is still allowed and remains custom.
			latest.Name = "Explicit custom name"
			w = save(latest)
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			rename("Newest")
			custom := readRoom()
			if custom.Name != "Explicit custom name" {
				t.Fatal("explicit custom name lost", custom)
			}
			// Renaming the actor does not invalidate a dialog for an unchanged custom label.
			rename("Final")
			custom.Description = "Custom description"
			w = save(custom)
			if w.Code != 200 {
				t.Fatal("custom label precondition changed unnecessarily", w.Code)
			}
			missing := readRoom()
			missing.NameVersion = ""
			if w = save(missing); w.Code != 409 {
				t.Fatal("missing personal-name precondition accepted", w.Code)
			}
			if _, err := (&server{db: s.db}).load(t.Context()); err != nil {
				t.Fatal("restart failed", err)
			}
		})
	}
}
