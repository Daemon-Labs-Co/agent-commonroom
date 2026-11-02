package main

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Pending code entry is ephemeral, owner-only and bound to one connection and
// worker. Authorization codes never become audit events or configuration.
type providerLoginSession struct {
	ID, Alias, Provider string
	Context             context.Context
	Cancel              context.CancelFunc
	Input               *os.File
	Ready, Submitted    bool
}

var claudeAuthorizationCode = regexp.MustCompile(`^[A-Za-z0-9._~+/#=\-]+$`)

func (s *server) providerLoginCode(w http.ResponseWriter, r *http.Request, connection, alias, provider, loginID, code string) {
	code = strings.TrimSpace(code)
	if provider != "claude" || len(code) < 8 || len(code) > 2048 || !claudeAuthorizationCode.MatchString(code) {
		http.Error(w, "Enter the authorization code supplied by Claude", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session := s.providerLoginSessions[connection]
	if session == nil || session.ID != loginID || session.Alias != alias || session.Provider != provider || session.Context.Err() != nil || !session.Ready || session.Submitted || session.Input == nil {
		http.Error(w, "This Claude sign-in is expired, already submitted or belongs to another connection. Start sign-in again.", 409)
		return
	}
	session.Submitted = true
	if _, err := session.Input.Write([]byte(code + "\n")); err != nil {
		session.Cancel()
		http.Error(w, "The worker could not receive the authorization code. Start sign-in again.", 503)
		return
	}
	reply(w, map[string]bool{"submitted": true})
}
