package main

// Platform API (ADR 0007). The operator's control plane calls it, signed with an
// Ed25519 key listed in COMMONROOM_PLATFORM_KEYS_FILE, to create, inspect, suspend,
// resume and delete workspaces, set their limits and reissue setup links. It is
// served only on platform.<COMMONROOM_WORKSPACE_DOMAIN>. It never reads or writes a
// workspace's rooms, messages, members or settings, and returns no workspace content.

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	platformContractVersion = 1
	platformRequestsPerMin  = 120
	platformClockSkew       = 60 * time.Second
	platformListMax         = 100
	setupLinkTTL            = 24 * time.Hour
)

func platformSigningInput(id, timestamp, method, target string, body []byte) []byte {
	sum := sha256.Sum256(body)
	return []byte(strings.Join([]string{"commonroom-platform-v" + strconv.Itoa(platformContractVersion), id, timestamp, method, target, hex.EncodeToString(sum[:])}, "\n"))
}

type platformAPI struct {
	tenancy tenancySettings
	store   workspaceStore
	host    *workspaceHost
	keyFile string
	limiter extensionLimiter
	now     func() time.Time
}

type platformWorkspaceView struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Status       string          `json:"status"`
	Origin       string          `json:"origin"`
	Limits       workspaceLimits `json:"limits"`
	Created      time.Time       `json:"created"`
	SetupPending *bool           `json:"setupPending,omitempty"`
	SetupLink    string          `json:"setupLink,omitempty"`
	SetupExpires *time.Time      `json:"setupExpires,omitempty"`
}

func (p *platformAPI) view(rec workspaceRecord) platformWorkspaceView {
	return platformWorkspaceView{ID: rec.ID, Name: rec.Name, Status: rec.Status, Origin: p.tenancy.origin(rec.ID), Limits: rec.Limits, Created: rec.Created}
}

// authenticate verifies the signature, freshness, replay and rate. Failures are
// uniform so a caller learns nothing about which key ids exist.
func (p *platformAPI) authenticate(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	deny := func() ([]byte, bool) { http.Error(w, "Platform authentication failed", 403); return nil, false }
	if p.keyFile == "" {
		return deny()
	}
	registry, err := loadExtensionRegistry(p.keyFile)
	if err != nil {
		return deny()
	}
	entry, ok := registry[r.Header.Get("X-Platform-Id")]
	stamp := r.Header.Get("X-Platform-Timestamp")
	seconds, parseErr := strconv.ParseInt(stamp, 10, 64)
	now := p.now()
	if !ok || parseErr != nil || now.Sub(time.Unix(seconds, 0)) > platformClockSkew || time.Unix(seconds, 0).Sub(now) > platformClockSkew {
		return deny()
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 16384))
	if err != nil {
		http.Error(w, "Request body too large", 413)
		return nil, false
	}
	signature, err := base64.StdEncoding.DecodeString(r.Header.Get("X-Platform-Signature"))
	if err != nil || !ed25519.Verify(entry.key, platformSigningInput(entry.ID, stamp, r.Method, r.URL.RequestURI(), body), signature) {
		return deny()
	}
	if err = p.limiter.admitLimit("platform:"+entry.ID, string(signature), now, platformRequestsPerMin); err != nil {
		http.Error(w, "Platform request rate exceeded or request already used", http.StatusTooManyRequests)
		return nil, false
	}
	return body, true
}

func (p *platformAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	body, ok := p.authenticate(w, r)
	if !ok {
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/platform/v1/"), "/"), "/")
	if !strings.HasPrefix(r.URL.Path, "/platform/v1/") || parts[0] != "workspaces" || len(parts) > 3 {
		http.Error(w, "Not found", 404)
		return
	}
	switch {
	case len(parts) == 1 && r.Method == "POST":
		p.create(w, r, body)
	case len(parts) == 1 && r.Method == "GET":
		p.list(w, r)
	case len(parts) == 2 && r.Method == "GET":
		p.get(w, r, parts[1])
	case len(parts) == 3 && r.Method == "POST" && (parts[2] == "suspend" || parts[2] == "resume" || parts[2] == "delete"):
		p.transition(w, r, parts[1], parts[2])
	case len(parts) == 3 && r.Method == "PUT" && parts[2] == "limits":
		p.setLimits(w, r, parts[1], body)
	case len(parts) == 3 && r.Method == "POST" && parts[2] == "setup-link":
		p.reissue(w, r, parts[1])
	default:
		http.Error(w, "Not found", 404)
	}
}

func (p *platformAPI) fail(w http.ResponseWriter, err error) {
	if errors.Is(err, errRegistryUnavailable) {
		http.Error(w, "Workspace registry unavailable", 503)
		return
	}
	http.Error(w, "Request refused", 400)
}

// record appends a platform action to the workspace's own history, so its owner can
// see what the operator did and the audit chain stays complete.
func (p *platformAPI) record(r *http.Request, rt *workspaceRuntime, action string, detail map[string]any) {
	detail["action"] = action
	detail["by"] = "platform"
	if err := rt.s.append(r.Context(), "platform.workspace.action", detail); err != nil {
		// The registry change already happened; surface the gap in the operator log.
		logPlatformAuditGap(rt.id, action)
	}
}

func (p *platformAPI) create(w http.ResponseWriter, r *http.Request, body []byte) {
	var req struct {
		ID     string          `json:"id"`
		Name   string          `json:"name"`
		Limits json.RawMessage `json:"limits"`
	}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if dec.Decode(&req) != nil || !validHostedWorkspaceID(req.ID) || !validWorkspaceName(req.Name) {
		http.Error(w, "Provide id (a lowercase DNS label that is not reserved) and name (1-100 characters)", 400)
		return
	}
	limits, err := parseWorkspaceLimits(req.Limits)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	created, err := p.store.Create(r.Context(), req.ID, req.Name, limits)
	if err != nil {
		p.fail(w, err)
		return
	}
	rec, found, err := p.store.Lookup(r.Context(), req.ID)
	if err != nil || !found {
		p.fail(w, errRegistryUnavailable)
		return
	}
	if rec.Status == workspaceDeleted {
		http.Error(w, "That workspace id was used by a deleted workspace and cannot be reused", 409)
		return
	}
	if !created && rec.Name != req.Name {
		http.Error(w, "A workspace with that id already exists", 409)
		return
	}
	p.host.invalidate(req.ID)
	rt, err := p.host.adminRuntime(r.Context(), req.ID)
	if err != nil {
		http.Error(w, "Workspace could not be started", 503)
		return
	}
	if created {
		p.record(r, rt, "created", map[string]any{"name": rec.Name})
	}
	view := p.view(rec)
	p.attachSetup(&view, rt)
	status := 200
	if created {
		status = 201
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(view)
}

// attachSetup adds the current setup link when the owner has not yet claimed the workspace.
func (p *platformAPI) attachSetup(view *platformWorkspaceView, rt *workspaceRuntime) {
	code, expires, pending := rt.s.auth.setupSnapshot()
	view.SetupPending = &pending
	if pending {
		view.SetupLink = p.tenancy.origin(rt.id) + "/#setup=" + code
		if !expires.IsZero() {
			view.SetupExpires = &expires
		}
	}
}

func (p *platformAPI) get(w http.ResponseWriter, r *http.Request, id string) {
	rec, found, err := p.store.Lookup(r.Context(), id)
	if err != nil {
		p.fail(w, err)
		return
	}
	if !found {
		http.Error(w, "Workspace not found", 404)
		return
	}
	view := p.view(rec)
	if rec.Status == workspaceActive {
		if rt, err := p.host.adminRuntime(r.Context(), id); err == nil {
			_, _, pending := rt.s.auth.setupSnapshot()
			view.SetupPending = &pending // the link itself is only issued by create and setup-link
		}
	}
	reply(w, view)
}

func (p *platformAPI) list(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > platformListMax {
			http.Error(w, "limit must be 1-100", 400)
			return
		}
		limit = n
	}
	after := r.URL.Query().Get("after")
	if len(after) > 63 {
		http.Error(w, "invalid cursor", 400)
		return
	}
	recs, err := p.store.List(r.Context(), after, limit)
	if err != nil {
		p.fail(w, err)
		return
	}
	views := make([]platformWorkspaceView, 0, len(recs))
	next := ""
	for _, rec := range recs {
		views = append(views, p.view(rec))
		next = rec.ID
	}
	if len(recs) < limit {
		next = ""
	}
	reply(w, map[string]any{"workspaces": views, "next": next})
}

func (p *platformAPI) transition(w http.ResponseWriter, r *http.Request, id, action string) {
	rec, found, err := p.store.Lookup(r.Context(), id)
	if err != nil {
		p.fail(w, err)
		return
	}
	if !found || rec.Status == workspaceDeleted {
		http.Error(w, "Workspace not found", 404)
		return
	}
	target := map[string]string{"suspend": workspaceSuspended, "resume": workspaceActive, "delete": workspaceDeleted}[action]
	if rec.Status != target {
		// Record first so the history shows the action even for a workspace about to
		// stop serving; the platform reaches suspended workspaces too.
		if rt, err := p.host.adminRuntime(r.Context(), id); err == nil {
			p.record(r, rt, action, map[string]any{})
		}
		if _, err := p.store.SetStatus(r.Context(), id, target); err != nil {
			p.fail(w, err)
			return
		}
	}
	if target == workspaceActive {
		p.host.invalidate(id)
	} else {
		p.host.evict(id)
	}
	rec.Status = target
	reply(w, p.view(rec))
}

func (p *platformAPI) setLimits(w http.ResponseWriter, r *http.Request, id string, body []byte) {
	limits, err := parseWorkspaceLimits(body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	rec, found, err := p.store.Lookup(r.Context(), id)
	if err != nil {
		p.fail(w, err)
		return
	}
	if !found || rec.Status == workspaceDeleted {
		http.Error(w, "Workspace not found", 404)
		return
	}
	if _, err = p.store.SetLimits(r.Context(), id, limits); err != nil {
		p.fail(w, err)
		return
	}
	p.host.invalidate(id)
	if rt, err := p.host.adminRuntime(r.Context(), id); err == nil {
		p.record(r, rt, "limits", map[string]any{"limits": limits})
	}
	rec.Limits = limits
	reply(w, p.view(rec))
}

func (p *platformAPI) reissue(w http.ResponseWriter, r *http.Request, id string) {
	rec, found, err := p.store.Lookup(r.Context(), id)
	if err != nil {
		p.fail(w, err)
		return
	}
	if !found || rec.Status != workspaceActive {
		http.Error(w, "Workspace not found", 404)
		return
	}
	rt, err := p.host.runtimeFor(r.Context(), id)
	if err != nil {
		http.Error(w, "Workspace could not be started", 503)
		return
	}
	code, expires, err := rt.s.auth.reissueSetupCode(setupLinkTTL, p.now())
	if err != nil {
		http.Error(w, "The owner has already claimed this workspace", 409)
		return
	}
	p.record(r, rt, "setup-link-reissued", map[string]any{})
	view := p.view(rec)
	pending := true
	view.SetupPending, view.SetupLink, view.SetupExpires = &pending, p.tenancy.origin(id)+"/#setup="+code, &expires
	reply(w, view)
}
