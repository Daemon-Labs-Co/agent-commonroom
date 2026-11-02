package main

import (
	"agentroom/internal/room"
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

type ownerSetupRaceDB struct {
	*mockDB
	injected bool
}

func (d *ownerSetupRaceDB) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	rows, err := d.mockDB.Query(ctx, query, args...)
	if err == nil && !d.injected {
		d.injected = true
		d.seed("actor.revised", room.Actor{ID: "concurrent-human", Kind: "human", Enabled: true})
	}
	return rows, err
}

func TestFirstOwnerRegistrationRejectsConcurrentHistory(t *testing.T) {
	t.Parallel()
	db := &ownerSetupRaceDB{mockDB: newMockDB(t)}
	path := filepath.Join(t.TempDir(), "owner-key")
	a, _ := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
	s := &server{db: db, auth: a}
	if s.prepareOwnerIdentity(t.Context(), path) == nil {
		t.Fatal("owner registration raced another account")
	}
	if len(db.snapshot()) != 1 {
		t.Fatal("failed registration changed existing history")
	}
}

func TestFirstOwnerIdentityPersistsOnlyPublicKey(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	path := filepath.Join(t.TempDir(), "owner-key")
	code, err := loadOwnerKey(path)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := newOwnerAuth(code, "127.0.0.1:8787")
	s := &server{db: db, auth: a}
	if err := s.prepareOwnerIdentity(t.Context(), path); err != nil {
		t.Fatal(err)
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	actor, ok := member(st, ownerID)
	if !ok || !strings.Contains(actor.PublicKey, "BEGIN PUBLIC KEY") || !a.setupPending {
		t.Fatal("missing first-owner registration")
	}
	for _, event := range db.snapshot() {
		if strings.Contains(event.payload, "PRIVATE KEY") || strings.Contains(event.payload, code) {
			t.Fatal("private credential persisted in database")
		}
	}
	before := len(db.snapshot())
	if err := s.prepareOwnerIdentity(t.Context(), path); err != nil {
		t.Fatal(err)
	}
	if len(db.snapshot()) != before {
		t.Fatal("restart registered another owner")
	}
	w := httptest.NewRecorder()
	s.setupStatus(w, httptest.NewRequest("GET", "/api/setup-status", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"setupRequired":true`) || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("setup state missing")
	}
	a.setupPending = false
	w = httptest.NewRecorder()
	s.setupStatus(w, httptest.NewRequest("GET", "/api/setup-status", nil))
	if !strings.Contains(w.Body.String(), `"setupRequired":false`) {
		t.Fatal("returning installation shown as new")
	}
}

func TestExistingHistoryNeverCreatesReplacementOwner(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	db.seed("actor.revised", room.Actor{ID: "existing-user", Kind: "human", Enabled: true})
	path := filepath.Join(t.TempDir(), "owner-key")
	a, _ := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
	s := &server{db: db, auth: a}
	if s.prepareOwnerIdentity(t.Context(), path) == nil {
		t.Fatal("replaced owner on existing database")
	}
	if _, err := os.Stat(path + ".private.pem"); !os.IsNotExist(err) {
		t.Fatal("generated unauthorized private key")
	}
	if err := a.prepareRSA(path); err != nil {
		t.Fatal(err)
	}
	if s.prepareOwnerIdentity(t.Context(), path) == nil {
		t.Fatal("pending setup admitted with another user")
	}
}

func TestRegisteredOwnerPublicKeyMismatchRefused(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	path := filepath.Join(t.TempDir(), "owner-key")
	a, _ := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
	if err := a.prepareRSA(path); err != nil {
		t.Fatal(err)
	}
	db.seed("actor.revised", room.Actor{ID: ownerID, Kind: "human", Enabled: true, PublicKey: "different registered key"})
	s := &server{db: db, auth: a}
	if s.prepareOwnerIdentity(t.Context(), path) == nil {
		t.Fatal("changed registered owner key")
	}
}
