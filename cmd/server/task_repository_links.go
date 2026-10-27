package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var errTaskRepositoryLinks = errors.New("invalid task repository links")
var repositorySlug = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,99}$`)
var repositoryNumber = regexp.MustCompile(`^[1-9][0-9]{0,9}$`)
var repositoryCommit = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)
var repositoryBranch = regexp.MustCompile(`^[A-Za-z0-9._/-]{1,200}$`)

// No network access or repository authority is obtained from a saved reference.
// This accepts a bounded safe URL subset, not proof an external object exists.
func validTaskRepositoryLink(v room.TaskRepositoryLink) bool {
	s := v.URL
	if len(s) > 1024 || !strings.HasPrefix(s, "https://github.com/") || strings.ContainsAny(s, "%\\?#\r\n\t ") {
		return false
	}
	u, e := url.Parse(s)
	if e != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) < 4 || !repositorySlug.MatchString(parts[0]) || !repositorySlug.MatchString(parts[1]) {
		return false
	}
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			return false
		}
	}
	switch v.Kind {
	case "issue":
		return len(parts) == 4 && parts[2] == "issues" && repositoryNumber.MatchString(parts[3])
	case "pull-request":
		return len(parts) == 4 && parts[2] == "pull" && repositoryNumber.MatchString(parts[3])
	case "commit":
		return len(parts) == 4 && parts[2] == "commit" && repositoryCommit.MatchString(parts[3])
	case "branch":
		return parts[2] == "tree" && repositoryBranch.MatchString(strings.Join(parts[3:], "/"))
	}
	return false
}
func taskRepositoryLinksKey(company, rid, tid string) string { return company + "/" + rid + "/" + tid }
func taskForRepositoryLinks(st room.State, rid, tid string) (room.Task, bool) {
	for _, t := range st.Tasks {
		if t.ID == tid && t.RoomID == rid {
			return t, true
		}
	}
	return room.Task{}, false
}
func validateTaskRepositoryLinks(st room.State, v room.TaskRepositoryLinks) error {
	t, ok := taskForRepositoryLinks(st, v.RoomID, v.TaskID)
	if !ok || !sharedTaskValueRoom(st, v.RoomID) || !st.FeatureFlags["taskRepositoryLinks"] || v.Version != 1 || v.CompanyID != st.Company.ID || v.ChangedBy != ownerID || !artifactRecordedActor(st, v.ChangedBy, v.RoomID) || v.ChangedAt.IsZero() || v.TaskRevision != t.Revision || v.Links == nil || len(v.Links) > 16 {
		return errTaskRepositoryLinks
	}
	old := st.TaskRepositoryLinks[taskRepositoryLinksKey(v.CompanyID, v.RoomID, v.TaskID)]
	if old.Revision == int(^uint(0)>>1) || v.Revision != old.Revision+1 {
		return errTaskRepositoryLinks
	}
	seen := map[string]bool{}
	for _, link := range v.Links {
		key := strings.ToLower(link.URL)
		if !validTaskRepositoryLink(link) || seen[key] {
			return errTaskRepositoryLinks
		}
		seen[key] = true
	}
	return nil
}
func projectTaskRepositoryLinks(st *room.State, raw []byte) error {
	var v room.TaskRepositoryLinks
	if json.Unmarshal(raw, &v) != nil {
		return errTaskRepositoryLinks
	}
	if err := validateTaskRepositoryLinks(*st, v); err != nil {
		return err
	}
	if st.TaskRepositoryLinks == nil {
		st.TaskRepositoryLinks = map[string]room.TaskRepositoryLinks{}
	}
	st.TaskRepositoryLinks[taskRepositoryLinksKey(v.CompanyID, v.RoomID, v.TaskID)] = v
	return nil
}

type taskRepositoryLinksRequest struct {
	TaskID       string                    `json:"taskId"`
	TaskRevision int                       `json:"taskRevision"`
	Revision     int                       `json:"revision"`
	Links        []room.TaskRepositoryLink `json:"links"`
}

func (s *server) taskRepositoryLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Method == "POST" && !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	tid := r.URL.Query().Get("taskId")
	var req taskRepositoryLinksRequest
	if r.Method == "POST" {
		raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 24576))
		if err != nil || !taskcontract.ValidJSON(raw, 24576) {
			http.Error(w, "Invalid bounded link JSON", 400)
			return
		}
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if dec.Decode(&req) != nil || dec.Decode(new(any)) != io.EOF || req.TaskID == "" || req.TaskRevision < 1 || req.Revision < 0 || req.Links == nil || len(req.Links) > 16 || tid != "" && tid != req.TaskID || r.Header.Get("Idempotency-Key") == "" {
			http.Error(w, "Task, revisions, links and operation key required", 400)
			return
		}
		tid = req.TaskID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	rid := requestedRoom(r)
	if !sharedTaskValueRoom(st, rid) || !artifactActor(st, r.Context(), rid) {
		http.Error(w, "Shared room unavailable", 404)
		return
	}
	if !st.FeatureFlags["taskRepositoryLinks"] {
		http.Error(w, "Task repository links are disabled", 403)
		return
	}
	task, ok := taskForRepositoryLinks(st, rid, tid)
	if !ok {
		http.Error(w, "Task unavailable", 404)
		return
	}
	key := taskRepositoryLinksKey(st.Company.ID, rid, tid)
	current, exists := st.TaskRepositoryLinks[key]
	if !exists {
		current = room.TaskRepositoryLinks{Version: 1, CompanyID: st.Company.ID, RoomID: rid, TaskID: tid, Links: []room.TaskRepositoryLink{}}
	}
	owner := requestActor(r) == ownerID
	if r.Method == "GET" {
		reply(w, map[string]any{"record": current, "taskRevision": task.Revision, "title": task.Title, "canEdit": owner})
		return
	}
	if !owner {
		http.Error(w, "Owner permission required", 403)
		return
	}
	ctx, replay := prepareOperation(w, r, st, req)
	if replay {
		return
	}
	if st.Incident || s.integrityFault {
		http.Error(w, "Audit integrity must be restored before editing links", 409)
		return
	}
	if current.Revision != req.Revision || task.Revision != req.TaskRevision || current.Revision == int(^uint(0)>>1) {
		http.Error(w, "Task or links changed; reopen repository links", 409)
		return
	}
	next := room.TaskRepositoryLinks{Version: 1, CompanyID: st.Company.ID, RoomID: rid, TaskID: tid, TaskRevision: task.Revision, Revision: current.Revision + 1, Links: req.Links, ChangedBy: ownerID, ChangedAt: time.Now().UTC()}
	if validateTaskRepositoryLinks(st, next) != nil {
		http.Error(w, "Use at most 16 distinct GitHub issue, branch, commit or pull request links", 400)
		return
	}
	response := map[string]any{"committed": true, "record": next}
	if err = s.appendBatch(withOperationResponse(ctx, 200, response), []auditChange{{Kind: "task.repository_links.updated", Payload: next}}); err != nil {
		if !errors.Is(err, errAuditConflict) {
			s.tripIntegrityLocked()
		}
		operationError(w, err)
		return
	}
	reply(w, response)
}
