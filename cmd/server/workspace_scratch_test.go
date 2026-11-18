//go:build scratchdb

package main

// Real-PostgreSQL verification of workspace storage. These tests are excluded from every normal
// build and from make test (the project forbids database access in automated tests). Run them with
// scripts/check-workspace-database.sh against a scratch cluster you own.

import (
	"agentroom/internal/room"
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func scratchPool(t *testing.T, phase string) *pgxpool.Pool {
	t.Helper()
	if os.Getenv("SCRATCH_PHASE") != phase {
		t.Skip("not the " + phase + " phase")
	}
	pool, err := pgxpool.New(t.Context(), os.Getenv("SCRATCH_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// Phase "pre": the layout before migration 002, served by an unscoped database as today.
func TestScratchLegacyHistoryBeforeMigration(t *testing.T) {
	pool := scratchPool(t, "pre")
	if st, err := verifyWorkspaceStorage(t.Context(), pool); err != nil || st.Present {
		t.Fatal("expected the previous layout", st, err)
	}
	legacy := &server{db: pool, active: map[string]bool{}}
	for i := 0; i < 3; i++ {
		if err := legacy.append(t.Context(), "message.created", room.Message{RoomID: "engineering", ID: "old" + string(rune('0'+i)), Author: "A", Text: "history", ActorType: "human"}); err != nil {
			t.Fatal(err)
		}
	}
	legacy.checkIntegrityNow()
	if legacy.integrityFault {
		t.Fatal("the legacy chain is invalid")
	}
}

// Phase "post": after migrations 002 and 003.
func TestScratchUpgradedHistoryContinuesAndStaysPrivate(t *testing.T) {
	pool := scratchPool(t, "post")
	ctx := context.Background()
	if st, err := verifyWorkspaceStorage(ctx, pool); err != nil || !st.Protected {
		t.Fatal("storage must be protected after migration", st, err)
	}
	db, _ := newScopedDatabase(pool, defaultWorkspaceID)
	s := &server{db: db, active: map[string]bool{}}
	if loaded, err := s.load(ctx); err != nil || loaded.EventCount != 3 {
		t.Fatal("history written before the migration must be the default workspace's", loaded.EventCount, err)
	}
	s.checkIntegrityNow()
	if s.integrityFault {
		t.Fatal("the existing chain failed verification after migration")
	}
	if err := s.append(ctx, "message.created", room.Message{RoomID: "engineering", ID: "new0", Author: "A", Text: "after", ActorType: "human"}); err != nil {
		t.Fatal(err)
	}
	s.checkIntegrityNow()
	if loaded, _ := s.load(ctx); s.integrityFault || loaded.EventCount != 4 {
		t.Fatal("the chain broke when continued after migration", loaded.EventCount)
	}
	other, _ := newScopedDatabase(pool, "beta")
	if st, _ := (&server{db: other, active: map[string]bool{}}).load(ctx); st.EventCount != 0 {
		t.Fatal("another workspace can see the pre-migration history")
	}
}

func TestScratchTwoWorkspacesAreIsolatedAndChainsVerifyIndependently(t *testing.T) {
	pool := scratchPool(t, "post")
	ctx := context.Background()
	mk := func(ws string) *server {
		db, err := newScopedDatabase(pool, ws)
		if err != nil {
			t.Fatal(err)
		}
		return &server{db: db, active: map[string]bool{}}
	}
	a, b := mk("alpha"), mk("beta")
	for i := 0; i < 3; i++ {
		if err := a.append(ctx, "message.created", room.Message{RoomID: "engineering", ID: "a" + string(rune('0'+i)), Author: "A", Text: "alpha secret", ActorType: "human"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.append(ctx, "message.created", room.Message{RoomID: "engineering", ID: "b0", Author: "B", Text: "beta secret", ActorType: "human"}); err != nil {
		t.Fatal(err)
	}
	sa, _ := a.load(ctx)
	sb, _ := b.load(ctx)
	if sa.EventCount != 3 || sb.EventCount != 1 {
		t.Fatal("event counts leaked across workspaces", sa.EventCount, sb.EventCount)
	}
	for _, m := range sa.Messages {
		if m.Text == "beta secret" {
			t.Fatal("alpha read beta's message")
		}
	}
	for _, m := range sb.Messages {
		if m.Text == "alpha secret" {
			t.Fatal("beta read alpha's message")
		}
	}
	a.checkIntegrityNow()
	b.checkIntegrityNow()
	if a.integrityFault || b.integrityFault {
		t.Fatal("per-workspace hash chains did not verify")
	}
	raw := &server{db: pool, active: map[string]bool{}}
	if st, err := raw.load(ctx); err != nil || st.EventCount != 0 {
		t.Fatal("unscoped access saw history", st.EventCount, err)
	}
	if err := raw.append(ctx, "message.created", room.Message{ID: "x"}); err == nil {
		t.Fatal("an unscoped insert was accepted")
	}
}

func TestScratchWorkspaceRegistry(t *testing.T) {
	pool := scratchPool(t, "post")
	reg := databaseWorkspaceRegistry{db: pool}
	ctx := t.Context()
	limits := defaultWorkspaceLimits
	limits.Rooms = 7
	if ok, err := reg.Create(ctx, "acme", "Acme Ltd", limits); err != nil || !ok {
		t.Fatal("create", ok, err)
	}
	if ok, err := reg.Create(ctx, "acme", "Other", limits); err != nil || ok {
		t.Fatal("a duplicate create must report not created", ok, err)
	}
	rec, found, err := reg.Lookup(ctx, "acme")
	if err != nil || !found || rec.Name != "Acme Ltd" || rec.Limits.Rooms != 7 || rec.Created.IsZero() || rec.Status != workspaceActive {
		t.Fatal("lookup", rec, found, err)
	}
	if ok, err := reg.SetStatus(ctx, "acme", workspaceSuspended); err != nil || !ok {
		t.Fatal("suspend", ok, err)
	}
	for _, id := range []string{"beta", "gamma"} {
		reg.Create(ctx, id, id, defaultWorkspaceLimits)
	}
	if list, err := reg.List(ctx, "", 2); err != nil || len(list) != 2 || list[0].ID != "acme" || list[1].ID != "beta" {
		t.Fatal("list", list, err)
	}
	if ok, err := reg.SetStatus(ctx, "acme", workspaceDeleted); err != nil || !ok {
		t.Fatal("delete", ok, err)
	}
	if ok, _ := reg.SetStatus(ctx, "acme", workspaceActive); ok {
		t.Fatal("a deleted workspace was revived")
	}
	if ok, _ := reg.Create(ctx, "acme", "Acme Ltd", limits); ok {
		t.Fatal("a deleted id was reused")
	}
	if _, err := pool.Exec(ctx, "DELETE FROM workspaces WHERE id='beta'"); err == nil {
		t.Fatal("the runtime role deleted a registry row")
	}
}

// scratchProjectionAgreement proves, on a real server, that the incremental cache and the
// incremental integrity check give the same answers as the full replay and the full check.
func scratchProjectionAgreement(t *testing.T, db eventDatabase, appends int) {
	t.Helper()
	ctx := context.Background()
	full := &server{db: db, active: map[string]bool{}}
	cached := &server{db: db, active: map[string]bool{}}
	cached.projection.enabled = true
	cached.integritySettings = integritySettings{incrementalEvery: time.Hour, fullEvery: 24 * time.Hour}
	same := func(step string) {
		want, wantErr := full.loadAt(ctx, equivalenceNow)
		got, gotErr := cached.loadCached(ctx, equivalenceNow)
		if msg := compareStates(want, wantErr, got, gotErr); msg != "" {
			t.Fatalf("%s: %s", step, msg)
		}
	}
	same("empty or existing history")
	start := time.Now()
	cached.checkIntegrityScheduledAt(start) // full, sets the checkpoint
	for i := 0; i < appends; i++ {
		if err := full.append(ctx, "message.created", room.Message{RoomID: "engineering", ID: "proj" + string(rune('a'+i)), Author: "A", Text: "incremental", ActorType: "human"}); err != nil {
			t.Fatal(err)
		}
		same("after append")
	}
	cached.checkIntegrityScheduledAt(start.Add(time.Hour)) // incremental, from the checkpoint
	if cached.integrityFault || cached.integrityPlan.checkpoint == nil || cached.integrityPlan.checkpoint.records < appends {
		t.Fatalf("incremental verification on a real database: fault=%v checkpoint=%+v", cached.integrityFault, cached.integrityPlan.checkpoint)
	}
	cached.checkIntegrityScheduledAt(start.Add(25 * time.Hour)) // full again
	if cached.integrityFault {
		t.Fatal("the full verification failed after incremental ones")
	}
}

func TestScratchProjectionCacheBeforeMigration(t *testing.T) {
	pool := scratchPool(t, "pre")
	// Read-only here: the phases that follow count the history written before the migration.
	scratchProjectionAgreement(t, pool, 0)
}

func TestScratchProjectionCacheAfterMigration(t *testing.T) {
	pool := scratchPool(t, "post")
	db, err := newScopedDatabase(pool, "projection")
	if err != nil {
		t.Fatal(err)
	}
	scratchProjectionAgreement(t, db, 5)
}
