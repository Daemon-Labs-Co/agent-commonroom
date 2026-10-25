//go:build linux || darwin

package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type syntheticArtifactGuard struct {
	source  func(context.Context, room.State, objectcustody.Intent) error
	seal    func(context.Context, room.State, objectcustody.Receipt) error
	quota   func(context.Context, room.StorageReservation) error
	release func(context.Context, room.State, objectcustody.Receipt) error
}

func (g *syntheticArtifactGuard) ValidateSource(c context.Context, s room.State, i objectcustody.Intent) error {
	if g.source != nil {
		return g.source(c, s, i)
	}
	return nil
}
func (g *syntheticArtifactGuard) ValidateSeal(c context.Context, s room.State, r objectcustody.Receipt) error {
	if g.seal != nil {
		return g.seal(c, s, r)
	}
	return nil
}
func (g *syntheticArtifactGuard) EstablishQuota(c context.Context, r room.StorageReservation) error {
	if g.quota != nil {
		return g.quota(c, r)
	}
	return nil
}
func (g *syntheticArtifactGuard) AuthorizeRelease(c context.Context, s room.State, r objectcustody.Receipt) error {
	if g.release != nil {
		return g.release(c, s, r)
	}
	return nil
}

// Real controller admission/replay with only a synthetic task turn and mocked DB.
// This deliberately does not constitute trusted source/quota or live acceptance.
func artifactJournalFixture(t *testing.T, inventory bool) (*server, *mockDB, context.Context, objectcustody.Scope) {
	t.Helper()
	s, db, h, task := structuredTaskFixture(t)
	task.Definition.DesiredOutputs = append(task.Definition.DesiredOutputs, taskcontract.DataContract{Name: "report", Type: "file", Required: false, FileName: "report.txt", MediaType: "text/plain", MaxBytes: 128})
	raw, _ := json.Marshal(task)
	w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	task = artifactState(t, s, t.Context()).Tasks[0]
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	if inventory {
		if err := s.initializeManagedStorage(ctx, "company", storageInventoryFunc(func(context.Context, string) (managedStorageInventory, error) {
			return managedStorageInventory{1, "company", strings.Repeat("d", 64), []room.StorageReservation{}}, nil
		})); err != nil {
			t.Fatal(err)
		}
	}
	ready := make(chan objectcustody.Scope, 1)
	done := make(chan struct{})
	finished := make(chan struct{})
	s.taskTurn = func(turn context.Context, p room.Participant, _, _ string) (string, error) {
		defer close(finished)
		repo := strings.Repeat("b", 64)
		jc := s.managedExecutionJobContext(turn, p, repo, func() (string, error) { return repo, nil })
		reg := jc.Value(managedJobContextKey{}).(managedJobRegistration)
		if err := reg.Register("synthetic-worker", "codex"); err != nil {
			t.Error(err)
			ready <- objectcustody.Scope{}
			return "", err
		}
		if err := reg.Update("submitted"); err != nil {
			t.Error(err)
			ready <- objectcustody.Scope{}
			return "", err
		}
		st, err := s.load(turn)
		if err != nil {
			t.Error(err)
			ready <- objectcustody.Scope{}
			return "", err
		}
		e := st.TaskExecutions[st.Tasks[0].DispatchID]
		ready <- objectcustody.Scope{ExecutionScope: objectcustody.ExecutionScope{CompanyID: "company", RoomID: p.RoomID, TaskID: e.TaskID, ExecutionID: e.ID, JobID: reg.JobID, RepositoryDigest: repo, RevisionDigest: strings.Repeat("c", 64), DefinitionDigest: e.DefinitionDigest}, OutputName: "report", MemberName: "report.txt"}
		<-done
		return `{"result":"synthetic result"}`, nil
	}
	w = dispatchRequest(h, "artifact-journal-task", definedDispatchBody(task))
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	var scope objectcustody.Scope
	select {
	case scope = <-ready:
	case <-time.After(time.Second):
		t.Fatal("mock turn not ready")
	}
	t.Cleanup(func() {
		close(done)
		select {
		case <-finished:
		case <-time.After(time.Second):
			t.Error("mock turn not stopped")
		}
		waitDispatch(t, s)
	})
	if scope.JobID == "" {
		t.Fatal("missing scoped job")
	}
	return s, db, ctx, scope
}
func artifactStore(t *testing.T, j objectcustody.Journal) (*objectcustody.Store, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	store, err := objectcustody.New(root, j)
	if err != nil {
		t.Fatal(err)
	}
	return store, dir
}
func artifactState(t *testing.T, s *server, ctx context.Context) room.State {
	t.Helper()
	st, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return st
}
func artifactTotal(t *testing.T, s *server, ctx context.Context) int64 {
	t.Helper()
	total, err := managedStorageTotal(artifactState(t, s, ctx))
	if err != nil {
		t.Fatal(err)
	}
	return total
}

func TestArtifactJournalActualBytesDurableChargeSealAndRead(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	guard := &syntheticArtifactGuard{}
	guard.quota = func(_ context.Context, r room.StorageReservation) error {
		st := artifactState(t, &server{db: db}, ctx)
		if st.ArtifactCustody[r.ID].Intent.ObjectID != r.ID || st.ManagedStorage.Reservations[r.ID].Bytes != 128 {
			t.Error("quota called before durable charge/intent")
		}
		return nil
	}
	j, err := newControllerArtifactJournal(s, guard)
	if err != nil {
		t.Fatal(err)
	}
	store, dir := artifactStore(t, j)
	receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic artifact bytes"))
	if err != nil {
		t.Fatal(err)
	}
	restarted := &server{db: db}
	fresh, _ := newControllerArtifactJournal(restarted, guard)
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	readStore, err := objectcustody.New(root, fresh)
	if err != nil {
		t.Fatal(err)
	}
	st := artifactState(t, restarted, ctx)
	record := st.ArtifactCustody[receipt.Intent.ObjectID]
	if record.Receipt == nil || *record.Receipt != receipt || artifactTotal(t, restarted, ctx) != 128 {
		t.Fatal("restart lost exact charge/receipt")
	}
	bytes, err := readStore.Get(ctx, scope, receipt)
	if err != nil || string(bytes) != "synthetic artifact bytes" {
		t.Fatal("actual bytes refused", err)
	}
	st = artifactState(t, restarted, ctx)
	if st.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments != 1 {
		t.Fatal("read acknowledgment missing")
	}
	raw, _ := json.Marshal(st)
	if strings.Contains(string(raw), receipt.Intent.ObjectID) {
		t.Fatal("global state exposed private custody")
	}
	if err := fresh.AuthorizeRead(ctx, scope, receipt); err != nil {
		t.Fatal("restart authorization failed", err)
	}
	// Stopping containers retains both the receipt and full capacity charge.
	job := st.ManagedJobs[scope.JobID]
	job.Status = "shutdown_confirmed"
	job.DiagnosticCode = "shutdown_confirmed"
	job.ShutdownConfirmed = true
	db.seed("worker.job.updated", job)
	if artifactTotal(t, restarted, ctx) != 128 {
		t.Fatal("shutdown refunded bytes")
	}
	if err := fresh.AuthorizeRead(ctx, scope, receipt); err != nil {
		t.Fatal("completed retention unreadable with current guard", err)
	}
}

type artifactReadProbe struct {
	reads  int
	onRead func()
	reader io.Reader
}

func (r *artifactReadProbe) Read(p []byte) (int, error) {
	r.reads++
	if r.onRead != nil {
		r.onRead()
		r.onRead = nil
	}
	return r.reader.Read(p)
}
func TestArtifactJournalRefusesBeforeSourceRead(t *testing.T) {
	for _, mode := range []string{"unknown inventory", "capacity", "source", "quota", "commit", "stale definition", "incident", "undeclared output", "wrong member", "wrong repository", "wrong execution", "wrong revision guard", "nonhuman", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, mode != "unknown inventory")
			guard := &syntheticArtifactGuard{}
			quotaCalls := 0
			guard.quota = func(context.Context, room.StorageReservation) error {
				quotaCalls++
				if mode == "quota" {
					return errors.New("synthetic quota unavailable")
				}
				return nil
			}
			if mode == "source" || mode == "wrong revision guard" {
				guard.source = func(context.Context, room.State, objectcustody.Intent) error {
					return errors.New("synthetic source refusal")
				}
			}
			switch mode {
			case "capacity":
				st := artifactState(t, s, ctx)
				j := st.ManagedJobs[scope.JobID]
				db.seed("storage.reserved", room.StorageReservation{Version: 1, ID: strings.Repeat("e", 32), CompanyID: j.CompanyID, JobID: j.JobID, RoomID: j.RoomID, Worker: j.Worker, RepositoryDigest: j.RepositoryDigest, Category: "bundle", Bytes: managedStorageLimit})
			case "commit":
				db.failAt = "commit"
			case "incident":
				db.seed("incident.opened", map[string]any{"reason": "synthetic incident"})
			case "stale definition":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
			case "undeclared output":
				scope.OutputName = "result"
			case "wrong member":
				scope.MemberName = "other.txt"
			case "wrong repository":
				scope.RepositoryDigest = strings.Repeat("a", 64)
			case "wrong execution":
				scope.ExecutionID = strings.Repeat("a", 32)
			case "nonhuman":
				ctx = context.WithValue(ctx, actorContextKey{}, "bot-openai")
			case "cancelled":
				c, cancel := context.WithCancel(ctx)
				cancel()
				ctx = c
			}
			j, _ := newControllerArtifactJournal(s, guard)
			store, dir := artifactStore(t, j)
			probe := &artifactReadProbe{reader: strings.NewReader("synthetic")}
			if _, err := store.Put(ctx, scope, 128, probe); err == nil || probe.reads != 0 {
				t.Fatal("read before admission", err, probe.reads)
			}
			files, _ := os.ReadDir(dir)
			if len(files) != 0 {
				t.Fatal("created file before admission")
			}
			db.failAt = ""
			checkCtx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
			st := artifactState(t, s, checkCtx)
			if mode == "quota" {
				if len(st.ArtifactCustody) != 1 || artifactTotal(t, s, checkCtx) != 128 || quotaCalls != 1 {
					t.Fatal("uncertain quota erased charge")
				}
			} else if len(st.ArtifactCustody) != 0 || quotaCalls != 0 {
				t.Fatal("unexpected admission/physical setup", quotaCalls)
			}
		})
	}
}

func TestArtifactJournalSealAndReadFailuresWithholdSuccessAndRetainBytes(t *testing.T) {
	for _, mode := range []string{"seal revoked", "seal commit", "seal source mismatch", "read revoked", "read commit", "forged receipt", "foreign scope", "unsealed", "duplicate seal", "duplicate member", "changed revision duplicate"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			guard := &syntheticArtifactGuard{}
			j, _ := newControllerArtifactJournal(s, guard)
			store, dir := artifactStore(t, j)
			probe := &artifactReadProbe{reader: strings.NewReader("synthetic bytes")}
			if mode == "seal revoked" {
				probe.onRead = func() { db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false}) }
			}
			if mode == "seal source mismatch" {
				guard.seal = func(context.Context, room.State, objectcustody.Receipt) error {
					return errors.New("synthetic source hash mismatch")
				}
			}
			if mode == "seal commit" {
				probe.onRead = func() { db.failAt = "commit" }
			}
			receipt, err := store.Put(ctx, scope, 128, probe)
			db.failAt = ""
			if strings.HasPrefix(mode, "seal ") {
				if err == nil {
					t.Fatal("seal failure returned receipt")
				}
				st := artifactState(t, s, ctx)
				if len(st.ArtifactCustody) != 1 || artifactTotal(t, s, ctx) != 128 {
					t.Fatal("seal failure lost charge")
				}
				for _, r := range st.ArtifactCustody {
					if r.Receipt != nil {
						t.Fatal("failed seal projected")
					}
				}
				files, _ := os.ReadDir(dir)
				if len(files) != 1 {
					t.Fatal("uncertain bytes deleted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "read revoked":
				calls := 0
				guard.release = func(context.Context, room.State, objectcustody.Receipt) error {
					calls++
					if calls == 2 {
						return errors.New("synthetic policy revoked")
					}
					return nil
				}
			case "read commit":
				calls := 0
				guard.release = func(context.Context, room.State, objectcustody.Receipt) error {
					calls++
					if calls == 2 {
						db.failAt = "commit"
					}
					return nil
				}
			case "forged receipt":
				receipt.Digest = strings.Repeat("e", 64)
			case "foreign scope":
				scope.RoomID = "elsewhere"
			case "unsealed":
				receipt.Intent.ObjectID = strings.Repeat("e", 32)
			case "duplicate seal":
				if j.Seal(ctx, receipt) == nil {
					t.Fatal("duplicate seal accepted")
				}
				return
			case "duplicate member", "changed revision duplicate":
				next := receipt.Intent
				next.ObjectID = strings.Repeat("e", 32)
				if mode == "changed revision duplicate" {
					next.Scope.RevisionDigest = strings.Repeat("d", 64)
				}
				if j.Reserve(ctx, next) == nil {
					t.Fatal("same output member re-reserved")
				}
				return
			}
			bytes, err := store.Get(ctx, scope, receipt)
			db.failAt = ""
			if err == nil || bytes != nil {
				t.Fatal("failed read released raw bytes")
			}
			st := artifactState(t, s, ctx)
			for _, r := range st.ArtifactCustody {
				if r.ReadAcknowledgments != 0 {
					t.Fatal("failed release acknowledged")
				}
			}
			if artifactTotal(t, s, ctx) != 128 {
				t.Fatal("failed release refunded capacity")
			}
			if mode == "read commit" && !s.integrityFault {
				t.Fatal("audit failure did not latch stop")
			}
		})
	}
}

func TestArtifactJournalStaleControllerConflictIsNotIntegrityIncident(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	s.integrityStopPath = filepath.Join(t.TempDir(), "integrity-stop")
	live, cancel := context.WithCancel(t.Context())
	defer cancel()
	s.cancels["conflict-probe"] = cancel
	calls := 0
	guard := &syntheticArtifactGuard{quota: func(context.Context, room.StorageReservation) error { calls++; return nil }}
	s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(context.Context) {
		db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
	}}
	j, _ := newControllerArtifactJournal(s, guard)
	err := j.Reserve(ctx, objectcustody.Intent{ObjectID: strings.Repeat("a", 32), Scope: scope, MaxBytes: 128})
	if !errors.Is(err, errAuditConflict) || s.integrityFault || calls != 0 {
		t.Fatal("stale reservation admitted or false incident", err, calls)
	}
	if len(artifactState(t, s, ctx).ArtifactCustody) != 0 || artifactTotal(t, s, ctx) != 0 {
		t.Fatal("conflict partially charged")
	}
	if live.Err() != nil {
		t.Fatal("ordinary conflict cancelled work")
	}
	if _, err := os.Stat(s.integrityStopPath); !os.IsNotExist(err) {
		t.Fatal("ordinary conflict persisted incident", err)
	}
	s.db = db
}

func TestArtifactCustodyInternalOnlyAndNilGuard(t *testing.T) {
	t.Parallel()
	s, _, h := mockServer(t)
	var absent *syntheticArtifactGuard
	for _, g := range []artifactCustodyGuard{nil, absent} {
		if _, err := newControllerArtifactJournal(s, g); err == nil {
			t.Fatal("missing trusted guard allowed")
		}
	}
	for _, kind := range []string{"artifact.custody.reserved", "artifact.custody.ready", "artifact.custody.sealed", "artifact.custody.read"} {
		w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":%q,"payload":{}}`, kind))
		if w.Code < 400 {
			t.Fatal("public event injection", kind, w.Code)
		}
	}
}

func TestArtifactJournalQuotaAcknowledgmentFailureCannotBeAdopted(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	guard := &syntheticArtifactGuard{quota: func(context.Context, room.StorageReservation) error { db.failAt = "commit"; return nil }}
	j, _ := newControllerArtifactJournal(s, guard)
	store, dir := artifactStore(t, j)
	probe := &artifactReadProbe{reader: strings.NewReader("synthetic")}
	if _, err := store.Put(ctx, scope, 128, probe); err == nil || probe.reads != 0 {
		t.Fatal("unacknowledged quota released write authority", err)
	}
	db.failAt = ""
	st := artifactState(t, s, ctx)
	if !s.integrityFault || len(st.ArtifactCustody) != 1 || artifactTotal(t, s, ctx) != 128 {
		t.Fatal("lost uncertain charge/stop")
	}
	files, _ := os.ReadDir(dir)
	if len(files) != 0 {
		t.Fatal("created file before quota acknowledgment")
	}
	fresh, _ := newControllerArtifactJournal(&server{db: db}, &syntheticArtifactGuard{})
	for _, r := range st.ArtifactCustody {
		if r.QuotaEstablished || r.Receipt != nil {
			t.Fatal("failed ready event projected")
		}
		fake := objectcustody.Receipt{Intent: r.Intent, Bytes: 0, Digest: strings.Repeat("a", 64)}
		if fresh.Seal(ctx, fake) == nil || fresh.AuthorizeRead(ctx, scope, fake) == nil || fresh.Reserve(ctx, r.Intent) == nil {
			t.Fatal("uncertain object adopted/retried")
		}
	}
}

func TestArtifactJournalCurrentHumanReadPermissionRecheckedBeforeRelease(t *testing.T) {
	for _, mode := range []string{"allowed member", "no room", "disabled member", "revoked during read", "anonymous", "agent", "incident", "flags off"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			guard := &syntheticArtifactGuard{}
			j, _ := newControllerArtifactJournal(s, guard)
			store, _ := artifactStore(t, j)
			receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic bytes"))
			if err != nil {
				t.Fatal(err)
			}
			actor := room.Actor{ID: "reader", Kind: "human", Enabled: true, Revision: 1, RoomIDs: []string{scope.RoomID}}
			if mode == "no room" {
				actor.RoomIDs = []string{"elsewhere"}
			}
			if mode == "disabled member" {
				actor.Enabled = false
			}
			db.seed("actor.revised", actor)
			readCtx := context.WithValue(ctx, actorContextKey{}, actor.ID)
			if mode == "incident" {
				db.seed("incident.opened", map[string]any{"reason": "synthetic incident"})
			}
			if mode == "flags off" {
				for _, name := range []string{"taskDefinitions", "jobDiagnostics", "agentBackgroundWork", "contextPolicyEditor"} {
					db.seed("feature.flag.updated", map[string]any{"name": name, "enabled": false})
				}
			}
			if mode == "anonymous" {
				readCtx = context.WithValue(ctx, actorContextKey{}, "")
			}
			if mode == "agent" {
				readCtx = context.WithValue(ctx, actorContextKey{}, "bot-openai")
			}
			if mode == "revoked during read" {
				guard.release = func(context.Context, room.State, objectcustody.Receipt) error {
					actor.Enabled = false
					actor.Revision++
					db.seed("actor.revised", actor)
					return nil
				}
			}
			bytes, err := store.Get(readCtx, scope, receipt)
			st := artifactState(t, s, ctx)
			reads := st.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments
			if mode == "allowed member" || mode == "flags off" {
				if err != nil || string(bytes) != "synthetic bytes" || reads != 1 {
					t.Fatal("authorized read failed", err)
				}
				last := db.snapshot()[len(db.snapshot())-1]
				var event artifactCustodyEvent
				if json.Unmarshal([]byte(last.payload), &event) != nil || event.ActorID != actor.ID {
					t.Fatal("read audit lost human attribution")
				}
			} else if err == nil || bytes != nil || reads != 0 {
				t.Fatal("revoked read released bytes", err, reads)
			}
		})
	}
}

func TestArtifactJournalConcurrentControllersCannotReserveSameOutputSlot(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	other := &server{db: db}
	guard := &syntheticArtifactGuard{}
	journals := []*controllerArtifactJournal{{server: s, guard: guard}, {server: other, guard: guard}}
	start := make(chan struct{})
	results := make(chan error, 2)
	for i, j := range journals {
		go func(i int, j *controllerArtifactJournal) {
			<-start
			results <- j.Reserve(ctx, objectcustody.Intent{ObjectID: fmt.Sprintf("%032x", 100+i), Scope: scope, MaxBytes: 128})
		}(i, j)
	}
	close(start)
	success := 0
	for range 2 {
		if err := <-results; err == nil {
			success++
		} else if !errors.Is(err, errAuditConflict) && !errors.Is(err, errArtifactCustody) {
			t.Fatal(err)
		}
	}
	st := artifactState(t, s, ctx)
	if success != 1 || len(st.ArtifactCustody) != 1 || artifactTotal(t, s, ctx) != 128 || s.integrityFault || other.integrityFault {
		t.Fatal("output slot overbooked or ordinary conflict latched", success)
	}
	for _, r := range st.ArtifactCustody {
		if !r.QuotaEstablished || r.Receipt != nil {
			t.Fatal("reservation lifecycle corrupt")
		}
	}
}

func TestArtifactJournalProjectionRejectsForgedTransitionsAndKeepsPrivateMetadata(t *testing.T) {
	t.Parallel()
	s, _, ctx, scope := artifactJournalFixture(t, true)
	guard := &syntheticArtifactGuard{}
	j, _ := newControllerArtifactJournal(s, guard)
	intent := objectcustody.Intent{ObjectID: strings.Repeat("a", 32), Scope: scope, MaxBytes: 128}
	if err := j.Reserve(ctx, intent); err != nil {
		t.Fatal(err)
	}
	receipt := objectcustody.Receipt{Intent: intent, Bytes: 3, Digest: strings.Repeat("d", 64)}
	base := artifactState(t, s, ctx)
	for _, mode := range []string{"unknown kind", "wrong version", "wrong actor", "foreign company", "wrong intent", "oversized receipt", "missing receipt", "extra reservation", "read before seal", "duplicate ready", "duplicate reserve", "missing charge", "released charge", "counter overflow", "missing actor", "agent actor", "disabled actor", "foreign actor room", "missing job", "missing execution", "undeclared member", "foreign worker reservation", "foreign stored worker"} {
		t.Run(mode, func(t *testing.T) {
			event := artifactCustodyEvent{Version: 1, ActorID: ownerID, Intent: intent, Receipt: &receipt}
			kind := "artifact.custody.sealed"
			switch mode {
			case "unknown kind":
				kind = "artifact.custody.adopted"
			case "wrong version":
				event.Version = 2
			case "wrong actor":
				event.ActorID = "../owner"
			case "missing actor":
				event.ActorID = "missing-human"
			case "agent actor":
				for _, actor := range base.Actors {
					if actor.Kind == "agent" {
						event.ActorID = actor.ID
						break
					}
				}
				if event.ActorID == ownerID {
					t.Fatal("synthetic agent missing")
				}
			case "disabled actor", "foreign actor room":
				event.ActorID = "replay-human"
			case "foreign company":
				event.Intent.Scope.CompanyID = "foreign"
			case "wrong intent":
				event.Intent.MaxBytes--
			case "oversized receipt":
				copy := receipt
				copy.Bytes = 129
				event.Receipt = &copy
			case "missing receipt":
				event.Receipt = nil
			case "extra reservation":
				event.Reservation = storageRecord(77, 1, "artifact")
			case "read before seal":
				kind = "artifact.custody.read"
			case "duplicate ready":
				kind = "artifact.custody.ready"
				event.Receipt = nil
			case "duplicate reserve":
				kind = "artifact.custody.reserved"
				event.Receipt = nil
			}
			candidate := base
			if mode == "disabled actor" || mode == "foreign actor room" {
				candidate.Actors = append([]room.Actor(nil), base.Actors...)
				candidate.Actors = append(candidate.Actors, room.Actor{ID: "replay-human", Kind: "human", Enabled: mode != "disabled actor", RoomIDs: []string{"elsewhere"}})
			}
			if mode == "missing job" {
				candidate.ManagedJobs = map[string]room.ManagedJob{}
			}
			if mode == "missing execution" {
				candidate.TaskExecutions = map[string]room.TaskExecution{}
			}
			if mode == "undeclared member" {
				event.Intent.Scope.MemberName = "undeclared.txt"
				r := *event.Receipt
				r.Intent = event.Intent
				event.Receipt = &r
			}
			candidate.ArtifactCustody = map[string]room.ArtifactCustodyObject{intent.ObjectID: base.ArtifactCustody[intent.ObjectID]}
			ledger := *base.ManagedStorage
			ledger.Reservations = map[string]room.StorageReservation{}
			for k, v := range base.ManagedStorage.Reservations {
				ledger.Reservations[k] = v
			}
			ledger.Releases = map[string]room.StorageRelease{}
			candidate.ManagedStorage = &ledger
			if mode == "foreign worker reservation" {
				kind = "artifact.custody.reserved"
				event.Receipt = nil
				event.Reservation = ledger.Reservations[intent.ObjectID]
				event.Reservation.Worker = "other-worker"
				delete(candidate.ArtifactCustody, intent.ObjectID)
				delete(ledger.Reservations, intent.ObjectID)
			}
			if mode == "foreign stored worker" {
				r := ledger.Reservations[intent.ObjectID]
				r.Worker = "other-worker"
				ledger.Reservations[intent.ObjectID] = r
			}
			if mode == "missing charge" {
				delete(ledger.Reservations, intent.ObjectID)
			}
			if mode == "released charge" {
				ledger.Releases[intent.ObjectID] = room.StorageRelease{Version: 1, ReservationID: intent.ObjectID, DisposalID: "synthetic"}
			}
			if mode == "counter overflow" {
				old := candidate.ArtifactCustody[intent.ObjectID]
				old.Receipt = &receipt
				old.ReadAcknowledgments = math.MaxInt64
				candidate.ArtifactCustody[intent.ObjectID] = old
				kind = "artifact.custody.read"
			}
			before, _ := json.Marshal(base.ArtifactCustody)
			raw, _ := json.Marshal(event)
			if projectArtifactCustody(&candidate, kind, raw) == nil {
				t.Fatal("forged transition accepted", mode)
			}
			after, _ := json.Marshal(base.ArtifactCustody)
			if string(before) != string(after) {
				t.Fatal("refusal mutated prior history")
			}
		})
	}
}

func TestArtifactHistoricalSourceBindingsDoNotReevaluateDeadline(t *testing.T) {
	t.Parallel()
	s, _, ctx, scope := artifactJournalFixture(t, true)
	j, _ := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
	intent := objectcustody.Intent{ObjectID: strings.Repeat("a", 32), Scope: scope, MaxBytes: 128}
	if err := j.Reserve(ctx, intent); err != nil {
		t.Fatal(err)
	}
	st := artifactState(t, s, ctx)
	job := st.ManagedJobs[scope.JobID]
	job.AllocatedAt = time.Now().Add(-2 * time.Minute)
	job.UpdatedAt = job.AllocatedAt
	job.Deadline = job.AllocatedAt.Add(time.Minute)
	st.ManagedJobs = map[string]room.ManagedJob{job.JobID: job}
	if _, ok := artifactSourceJob(st, intent); ok {
		t.Fatal("expired job gained current authority")
	}
	receipt := objectcustody.Receipt{Intent: intent, Bytes: 3, Digest: strings.Repeat("d", 64)}
	event := artifactCustodyEvent{Version: 1, ActorID: ownerID, Intent: intent, Receipt: &receipt}
	if err := validateArtifactTransition(st, "artifact.custody.sealed", event); err != nil {
		t.Fatal("historical relationship rechecked wall clock", err)
	}
}

func TestArtifactJournalAuditFailureCancelsWorkAndSurvivesRestart(t *testing.T) {
	for _, stage := range []string{"reserve", "ready", "seal", "read"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			s.integrityStopPath = filepath.Join(t.TempDir(), "integrity-stop")
			model, cancelModel := context.WithCancel(t.Context())
			defer cancelModel()
			workflow, cancelWorkflow := context.WithCancel(t.Context())
			defer cancelWorkflow()
			s.cancels["synthetic-probe"] = cancelModel
			s.workflowLive = map[string]context.CancelFunc{"synthetic-workflow": cancelWorkflow}
			guard := &syntheticArtifactGuard{}
			if stage == "ready" {
				guard.quota = func(context.Context, room.StorageReservation) error { db.failAt = "commit"; return nil }
			}
			j, _ := newControllerArtifactJournal(s, guard)
			store, _ := artifactStore(t, j)
			source := &artifactReadProbe{reader: strings.NewReader("synthetic artifact")}
			if stage == "reserve" {
				db.failAt = "commit"
			}
			if stage == "seal" {
				source.onRead = func() { db.failAt = "commit" }
			}
			receipt, err := store.Put(ctx, scope, 128, source)
			if stage == "read" {
				if err != nil {
					t.Fatal(err)
				}
				calls := 0
				guard.release = func(context.Context, room.State, objectcustody.Receipt) error {
					calls++
					if calls == 2 {
						db.failAt = "commit"
					}
					return nil
				}
				bytes, readErr := store.Get(ctx, scope, receipt)
				err = readErr
				if bytes != nil {
					t.Fatal("failed audit released bytes")
				}
			}
			if err == nil || !s.integrityFault || model.Err() != context.Canceled || workflow.Err() != context.Canceled {
				t.Fatal("audit failure did not stop/cancel", err)
			}
			info, statErr := os.Stat(s.integrityStopPath)
			if statErr != nil || info.Mode().Perm() != 0600 {
				t.Fatal("restart marker missing/private mode wrong", statErr)
			}
			db.failAt = ""
			// Restart before any periodic probe can persist an incident. The
			// rolled-back audit chain is healthy, so only the marker preserves
			// the failed-persistence stop in this exact regression.
			previous := ""
			for _, event := range db.snapshot() {
				if event.kind == "incident.opened" {
					t.Fatal("restart fixture already has durable incident")
				}
				if event.previous != previous || event.hash != checksum(previous, event.kind, event.payload) {
					t.Fatal("rollback damaged audit chain")
				}
				previous = event.hash
			}
			fresh := &server{db: db, integrityStopPath: s.integrityStopPath}
			fresh.restoreIntegrityStop()
			if !fresh.agentsBlockedLocked(room.State{}, scope.RoomID) {
				t.Fatal("restart forgot failed required persistence")
			}
			freshJournal, _ := newControllerArtifactJournal(fresh, &syntheticArtifactGuard{})
			fresh.mu.Lock()
			_, currentErr := freshJournal.current(ctx, objectcustody.Intent{ObjectID: strings.Repeat("f", 32), Scope: scope, MaxBytes: 128}, false)
			fresh.mu.Unlock()
			if currentErr == nil {
				t.Fatal("restarted controller granted custody authority")
			}

			s.checkIntegrityNow()
			if !s.integrityFault {
				t.Fatal("healthy audit probe auto-resumed")
			}
		})
	}
}
