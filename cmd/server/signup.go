package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Built-in sign-up (multi mode only, off by default). A visitor to the bare domain picks a
// name and an address; the server creates a workspace with the default limits and returns
// its first-owner setup link in the same response. There is no email, so the protections are
// a per-address rate limit, a daily and a total cap, and an optional shared sign-up code.
const (
	defaultSignupDailyLimit  = 50
	defaultSignupMaxTotal    = 1000
	signupPerAddressLimit    = 5
	signupAttemptsPerAddress = 20
	signupWindow             = time.Hour
	signupBodyLimit          = 2048
	signupTrackedAddresses   = 10000
)

type signupSettings struct {
	enabled    bool
	code       string
	dailyLimit int
	maxTotal   int
}

// loadSignupSettings reads the operator's sign-up configuration. Sign-up is a decision to
// let strangers create workspaces, so it needs the multi-workspace mode and fails startup
// otherwise instead of silently doing nothing.
func loadSignupSettings(getenv func(string) string, tenancy tenancySettings) (signupSettings, error) {
	s := signupSettings{dailyLimit: defaultSignupDailyLimit, maxTotal: defaultSignupMaxTotal}
	switch getenv("COMMONROOM_SIGNUP") {
	case "", "closed":
		return s, nil
	case "open":
		s.enabled = true
	default:
		return s, errors.New("COMMONROOM_SIGNUP must be closed or open")
	}
	if tenancy.Mode != tenancyMulti {
		return s, errors.New("COMMONROOM_SIGNUP=open requires COMMONROOM_TENANCY=multi")
	}
	s.code = getenv("COMMONROOM_SIGNUP_CODE")
	if s.code != "" && (len(s.code) < 8 || len(s.code) > 128 || s.code != strings.TrimSpace(s.code)) {
		return s, errors.New("COMMONROOM_SIGNUP_CODE must be 8-128 characters without surrounding spaces")
	}
	for _, v := range []struct {
		name string
		dst  *int
		max  int
	}{{"COMMONROOM_SIGNUP_DAILY_LIMIT", &s.dailyLimit, 100000}, {"COMMONROOM_SIGNUP_MAX_WORKSPACES", &s.maxTotal, 1000000}} {
		raw := getenv(v.name)
		if raw == "" {
			continue
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > v.max {
			return s, errors.New(v.name + " must be an integer from 1 to " + strconv.Itoa(v.max))
		}
		*v.dst = n
	}
	return s, nil
}

type signupService struct {
	settings signupSettings
	tenancy  tenancySettings
	store    workspaceStore
	host     *workspaceHost
	web      http.Handler
	now      func() time.Time

	mu       sync.Mutex // serialises check-then-create so the caps cannot be raced
	attempts map[string][]time.Time
	created  map[string][]time.Time
}

func newSignupService(settings signupSettings, tenancy tenancySettings, store workspaceStore, host *workspaceHost) *signupService {
	return &signupService{settings: settings, tenancy: tenancy, store: store, host: host, web: staticHandler("web"), now: time.Now}
}

// signupHost reports whether the request is for the bare domain.
func (h *workspaceHost) signupHost(host string) bool {
	if h.signup == nil || h.tenancy.Mode != tenancyMulti {
		return false
	}
	if bare, _, err := net.SplitHostPort(host); err == nil {
		host = bare
	}
	return strings.ToLower(strings.TrimSuffix(host, ".")) == h.tenancy.Domain
}

// clientAddress is the address used for rate limiting. Behind the proxy every connection
// comes from loopback and the proxy overwrites X-Real-IP, so the header is trusted only
// when the direct peer is loopback.
func clientAddress(r *http.Request) string {
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		peer = r.RemoteAddr
	}
	if ip := net.ParseIP(peer); ip != nil && ip.IsLoopback() {
		if real := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); real != nil {
			return real.String()
		}
	}
	return peer
}

func (s *signupService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'")
	switch {
	case r.URL.Path == "/api/signup/config" && r.Method == http.MethodGet:
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"domain": s.tenancy.Domain, "codeRequired": s.settings.code != ""})
	case r.URL.Path == "/api/signup" && r.Method == http.MethodPost:
		w.Header().Set("Cache-Control", "no-store")
		s.signup(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/"):
		http.Error(w, "Not found", 404)
	case r.Method != http.MethodGet && r.Method != http.MethodHead:
		http.Error(w, "Method not allowed", 405)
	case r.URL.Path == "/" || r.URL.Path == "/index.html":
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/signup.html"
		s.web.ServeHTTP(w, r2)
	case r.URL.Path == "/signup.js" || r.URL.Path == "/signup.css" || r.URL.Path == "/style.css" || r.URL.Path == "/assets/commonroom-mark.png":
		s.web.ServeHTTP(w, r)
	default:
		http.Error(w, "Not found", 404)
	}
}

// sameOrigin requires the browser's Origin to be exactly the sign-up origin, so another
// site cannot make a visitor's browser create workspaces.
func (s *signupService) sameOrigin(r *http.Request) bool {
	u, err := url.Parse(r.Header.Get("Origin"))
	return err == nil && u.Scheme == "https" && strings.EqualFold(u.Host, s.tenancy.Domain) && u.Path == ""
}

func (s *signupService) refuse(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}

func within(list []time.Time, now time.Time, window time.Duration) []time.Time {
	kept := list[:0]
	for _, at := range list {
		if now.Sub(at) < window {
			kept = append(kept, at)
		}
	}
	return kept
}

func (s *signupService) signup(w http.ResponseWriter, r *http.Request) {
	if !s.sameOrigin(r) || !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		s.refuse(w, 403, "Sign-up must be sent from this page")
		return
	}
	var req struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, signupBodyLimit))
	dec.DisallowUnknownFields()
	if dec.Decode(&req) != nil {
		s.refuse(w, 400, "Provide a workspace name, an address and, if asked, the sign-up code")
		return
	}
	now := s.now()
	address := clientAddress(r)

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.attempts == nil {
		s.attempts, s.created = map[string][]time.Time{}, map[string][]time.Time{}
	}
	if len(s.attempts) >= signupTrackedAddresses {
		for key, list := range s.attempts {
			if len(within(list, now, signupWindow)) == 0 {
				delete(s.attempts, key)
			}
		}
		if len(s.attempts) >= signupTrackedAddresses {
			s.refuse(w, 429, "Sign-up is busy; try again later")
			return
		}
	}
	// Every attempt counts, including wrong codes and taken names, so guessing is bounded.
	recent := within(s.attempts[address], now, signupWindow)
	if len(recent) >= signupAttemptsPerAddress {
		s.attempts[address] = recent
		w.Header().Set("Retry-After", "3600")
		s.refuse(w, 429, "Too many attempts from this address; try again later")
		return
	}
	s.attempts[address] = append(recent, now)

	if s.settings.code != "" && subtle.ConstantTimeCompare([]byte(req.Code), []byte(s.settings.code)) != 1 {
		s.refuse(w, 403, "The sign-up code is not correct")
		return
	}
	if !validHostedWorkspaceID(req.ID) || !validWorkspaceName(req.Name) {
		s.refuse(w, 400, "Choose an address of 2-63 lowercase letters, digits or hyphens that is not reserved, and a name of 1-100 characters")
		return
	}
	made := within(s.created[address], now, signupWindow)
	s.created[address] = made
	if len(made) >= signupPerAddressLimit {
		s.refuse(w, 429, "This address has created enough workspaces for now")
		return
	}
	total, today, err := s.counts(r.Context(), now)
	if err != nil {
		s.refuse(w, 503, "Sign-up is unavailable right now")
		return
	}
	if total >= s.settings.maxTotal || today >= s.settings.dailyLimit {
		s.refuse(w, 503, "Sign-up is full for now; try again later")
		return
	}
	created, err := s.store.Create(r.Context(), req.ID, req.Name, defaultWorkspaceLimits)
	if err != nil {
		s.refuse(w, 503, "Sign-up is unavailable right now")
		return
	}
	if !created {
		// Never return an existing workspace or its setup link: that would let a visitor
		// take over someone else's workspace before they claim it.
		s.refuse(w, 409, "That address is taken; choose another")
		return
	}
	s.created[address] = append(made, now)
	s.host.invalidate(req.ID)
	rt, err := s.host.adminRuntime(r.Context(), req.ID)
	if err != nil {
		s.refuse(w, 503, "The workspace was created but could not be started; contact the operator with the address "+req.ID)
		return
	}
	if err = rt.s.append(r.Context(), "platform.workspace.action", map[string]any{"action": "created", "by": "signup", "name": req.Name}); err != nil {
		logPlatformAuditGap(rt.id, "created")
	}
	code, expires, pending := rt.s.auth.setupSnapshot()
	if !pending {
		s.refuse(w, 503, "The workspace was created but is not ready; contact the operator with the address "+req.ID)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(map[string]any{"id": req.ID, "name": req.Name, "origin": s.tenancy.origin(req.ID), "setupLink": s.tenancy.origin(req.ID) + "/#setup=" + code, "setupExpires": expires})
}

// counts returns the live workspaces and those created in the last day. Deleted workspaces
// do not count towards the total but do count towards the day, so create-and-delete cycling
// cannot get around the daily cap. The scan stops as soon as the total cap is reached.
func (s *signupService) counts(ctx context.Context, now time.Time) (total, today int, err error) {
	after := ""
	for {
		page, err := s.store.List(ctx, after, 100)
		if err != nil {
			return 0, 0, err
		}
		if len(page) == 0 {
			return total, today, nil
		}
		for _, rec := range page {
			if rec.Status != workspaceDeleted {
				total++
			}
			if now.Sub(rec.Created) < 24*time.Hour {
				today++
			}
			after = rec.ID
		}
		if total >= s.settings.maxTotal {
			return total, today, nil
		}
	}
}
