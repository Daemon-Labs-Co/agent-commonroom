package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/objectread"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type objectReadJournalGuard struct {
	f             *assembledProbeFixture
	deny, failAck string
	stopped       bool
	seen          []string
}

func (g *objectReadJournalGuard) AuthorizeObjectRead(ctx context.Context, r workerJobRecord, i workerObjectReadIntent, e workerProbeEvent) error {
	g.f.checkLock()
	if r.JobID != g.f.record.JobID || i.Request.Receipt.Intent.Scope.JobID != r.JobID || e.Stage == g.deny {
		return errArtifactCustody
	}
	return ctx.Err()
}
func (g *objectReadJournalGuard) AcknowledgeObjectRead(ctx context.Context, r workerObjectReadRecord, e workerProbeEvent) error {
	g.f.checkLock()
	records, err := readWorkerObjectReadRecords(g.f.dir)
	saved, ok := records[objectReadRecordName(r.Intent)]
	if err != nil || !ok || saved.Events[len(saved.Events)-1] != e {
		g.f.t.Fatal("controller ack before exact local persistence", err)
	}
	g.seen = append(g.seen, e.Stage)
	if e.Stage == g.failAck {
		return errors.New("synthetic controller journal failure")
	}
	// A callback must not mutate the actual local record via slice aliases.
	r.Events[0].Stage = "foreign"
	return ctx.Err()
}
func (g *objectReadJournalGuard) StopObjectReadAuthority() { g.stopped = true }
func objectReadJournalFixture(t *testing.T) (*workerObjectReadJournal, *objectReadJournalGuard, *workerWorkspaceObservation, *workerCaptureJournal) {
	t.Helper()
	capture, base, l := captureJournalFixtureScope(t)
	scope := capture.intent.Request.Binding.Execution
	scope.RevisionDigest = strings.Repeat("9", 64)
	q := objectread.Request{Receipt: objectcustody.Receipt{Intent: objectcustody.Intent{ObjectID: strings.Repeat("8", 32), Scope: objectcustody.Scope{ExecutionScope: scope, OutputName: "result", MemberName: "result.txt"}, MaxBytes: 128}, Bytes: 3, Digest: strings.Repeat("7", 64)}, Challenge: capture.intent.Request.Challenge}
	g := &objectReadJournalGuard{f: base.f}
	j, err := newWorkerObjectReadJournal(l, q, capture.intent.Image, capture.intent.Objects, g)
	if err != nil {
		t.Fatal(err)
	}
	return j, g, l, capture
}
func beginObjectReadJournal(t *testing.T, j *workerObjectReadJournal) {
	t.Helper()
	if err := j.BeginObjectRead(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image); err != nil {
		t.Fatal(err)
	}
}
func TestWorkerObjectReadJournalDurableLifecycle(t *testing.T) {
	t.Parallel()
	j, g, _, capture := objectReadJournalFixture(t)
	beginObjectReadJournal(t, j)
	n, err := workerObjectReadSlots(j.dir)
	if err != nil || n != 1 {
		t.Fatal("reader capacity not charged", n, err)
	}
	jobs, err := boundedWorkerRecords(j.dir)
	if err != nil {
		t.Fatal(err)
	}
	if workerProbeCapacity(j.dir, jobs, workerContainerLimit) == nil {
		t.Fatal("reader omitted from shared capacity")
	}
	if capture.BeginCapture(t.Context(), capture.intent.Request, capture.intent.Objects, capture.intent.Image) == nil {
		t.Fatal("capture admitted alongside unresolved reader")
	}
	id := strings.Repeat("1", 64)
	if j.BindObjectRead(t.Context(), j.intent.Request, id) != nil || j.FinishObjectRead(t.Context(), j.intent.Request, id, true) != nil {
		t.Fatal("lifecycle failed")
	}
	records, err := readWorkerObjectReadRecords(j.dir)
	if err != nil || !objectReadRecordSettled(records[objectReadRecordName(j.intent)]) {
		t.Fatal("missing settled durable record", err)
	}
	if n, err = workerObjectReadSlots(j.dir); err != nil || n != 0 || len(g.seen) != 3 || g.stopped {
		t.Fatal("wrong audit/capacity result", n, err, g)
	}
	if j.BeginObjectRead(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image) == nil {
		t.Fatal("reader begin replayed")
	}
}
func TestWorkerObjectReadJournalKnownRecoveryAndUnknownRetention(t *testing.T) {
	t.Parallel()
	for _, known := range []bool{false, true} {
		t.Run(map[bool]string{false: "unknown", true: "known"}[known], func(t *testing.T) {
			t.Parallel()
			j, _, _, _ := objectReadJournalFixture(t)
			beginObjectReadJournal(t, j)
			id := ""
			if known {
				id = strings.Repeat("1", 64)
				if err := j.BindObjectRead(t.Context(), j.intent.Request, id); err != nil {
					t.Fatal(err)
				}
			}
			if err := j.FinishObjectRead(t.Context(), j.intent.Request, id, false); err != nil {
				t.Fatal(err)
			}
			if n, err := workerObjectReadSlots(j.dir); err != nil || n != 1 {
				t.Fatal("uncertainty freed capacity", n, err)
			}
			err := j.BeginObjectReadRecovery(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image, id)
			if !known {
				if err == nil {
					t.Fatal("unknown ID recovered")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if j.FinishObjectReadRecovery(t.Context(), j.intent.Request, id, false) == nil {
				t.Fatal("failed cleanup claimed recovery")
			}
			if n, err := workerObjectReadSlots(j.dir); err != nil || n != 1 {
				t.Fatal("failed recovery refunded capacity", n, err)
			}
			if err := j.FinishObjectReadRecovery(t.Context(), j.intent.Request, id, true); err != nil {
				t.Fatal(err)
			}
			if n, err := workerObjectReadSlots(j.dir); err != nil || n != 0 {
				t.Fatal("durable recovery not settled", n, err)
			}
		})
	}
}
func TestWorkerObjectReadJournalFailuresRetainCapacity(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"local-write", "confirmation-write", "controller-begin", "controller-bound", "controller-finish", "denied-finish", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, g, _, _ := objectReadJournalFixture(t)
			if mode == "controller-begin" {
				g.failAck = "begin"
				if j.BeginObjectRead(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image) == nil {
					t.Fatal("begin ack failure ignored")
				}
			} else {
				beginObjectReadJournal(t, j)
				id := strings.Repeat("1", 64)
				if mode == "controller-bound" {
					g.failAck = "bound"
					if j.BindObjectRead(t.Context(), j.intent.Request, id) == nil {
						t.Fatal("bind ack failure ignored")
					}
				} else {
					if err := j.BindObjectRead(t.Context(), j.intent.Request, id); err != nil {
						t.Fatal(err)
					}
					switch mode {
					case "local-write":
						j.persist = func(string, []byte, func(string) error) error { return errors.New("synthetic write failure") }
					case "confirmation-write":
						persist := j.persist
						writes := 0
						j.persist = func(path string, raw []byte, sync func(string) error) error {
							writes++
							if writes == 2 {
								return errors.New("synthetic confirmation failure")
							}
							return persist(path, raw, sync)
						}
					case "controller-finish":
						g.failAck = "finish"
					case "denied-finish":
						g.deny = "finish"
					case "canceled":
						c, cancel := context.WithCancel(j.lifetime)
						cancel()
						j.lifetime = c
					}
					if j.FinishObjectRead(t.Context(), j.intent.Request, id, true) == nil {
						t.Fatal("finish failure ignored")
					}
				}
			}
			n, err := workerObjectReadSlots(j.dir)
			if err != nil || n != 1 {
				t.Fatal("failed journal freed capacity", n, err)
			}
			if mode == "canceled" {
				records, e := readWorkerObjectReadRecords(j.dir)
				saved := records[objectReadRecordName(j.intent)]
				if e != nil || saved.Events[len(saved.Events)-1].Stage != "finish" || !strings.Contains(strings.Join(g.seen, ","), "finish") {
					t.Fatal("cancellation prevented durable finish uncertainty", e)
				}
			}
			shouldStop := mode != "denied-finish" && mode != "canceled"
			if g.stopped != shouldStop || workerProbeStopped(j.dir) != shouldStop {
				t.Fatal("wrong integrity-stop result", mode, g.stopped)
			}
		})
	}
}
func TestWorkerObjectReadJournalRefusesSubstitutionAndConflicts(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"request", "image", "volume", "wrong-id", "bind-before-begin", "capture-active", "denied-begin", "escaped", "unsafe-record"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, g, l, capture := objectReadJournalFixture(t)
			q, image, objects := j.intent.Request, j.intent.Image, j.intent.Objects
			var err error
			switch mode {
			case "request":
				q.Receipt.Digest = strings.Repeat("2", 64)
				err = j.BeginObjectRead(t.Context(), q, objects, image)
			case "image":
				err = j.BeginObjectRead(t.Context(), q, objects, "sha256:"+strings.Repeat("2", 64))
			case "volume":
				objects.Nonce = strings.Repeat("2", 32)
				err = j.BeginObjectRead(t.Context(), q, objects, image)
			case "bind-before-begin":
				err = j.BindObjectRead(t.Context(), q, strings.Repeat("1", 64))
			case "wrong-id":
				beginObjectReadJournal(t, j)
				id := strings.Repeat("1", 64)
				if e := j.BindObjectRead(t.Context(), q, id); e != nil {
					t.Fatal(e)
				}
				err = j.FinishObjectRead(t.Context(), q, strings.Repeat("2", 64), true)
			case "capture-active":
				beginCaptureJournal(t, capture)
				err = j.BeginObjectRead(t.Context(), q, objects, image)
			case "denied-begin":
				g.deny = "begin"
				err = j.BeginObjectRead(t.Context(), q, objects, image)
			case "escaped":
				ctx, cancel := context.WithCancel(l.lifetime)
				cancel()
				j.lifetime = ctx
				err = j.BeginObjectRead(t.Context(), q, objects, image)
			case "unsafe-record":
				if e := os.WriteFile(filepath.Join(j.dir, ".object-read-foreign.record"), []byte("{}"), 0600); e != nil {
					t.Fatal(e)
				}
				err = j.BeginObjectRead(t.Context(), q, objects, image)
			}
			if err == nil {
				t.Fatal("substitution/conflict accepted", mode)
			}
			if mode != "unsafe-record" && g.stopped {
				t.Fatal("refusal invented persistence fault")
			}
		})
	}
}

func TestWorkerObjectReadJournalCannotEscapeCreatorLease(t *testing.T) {
	t.Parallel()
	j, _, l, _ := objectReadJournalFixture(t)
	beginObjectReadJournal(t, j)
	l.mu.Lock()
	l.active = false
	l.mu.Unlock()
	if j.FinishObjectRead(t.Context(), j.intent.Request, "", false) == nil {
		t.Fatal("escaped journal used closed creator lease")
	}
	records, err := readWorkerObjectReadRecords(j.dir)
	if err != nil || len(records[objectReadRecordName(j.intent)].Events) != 1 {
		t.Fatal("escaped journal appended", err)
	}
}

func TestWorkerObjectReadJournalRequiresFreshAttemptAndPinnedLease(t *testing.T) {
	t.Parallel()
	j, g, l, _ := objectReadJournalFixture(t)
	beginObjectReadJournal(t, j)
	next := j.intent.Request
	next.Challenge = strings.Repeat("2", 32)
	another, err := newWorkerObjectReadJournal(l, next, j.intent.Image, j.intent.Objects, g)
	if err != nil {
		t.Fatal(err)
	}
	if another.BeginObjectRead(t.Context(), next, j.intent.Objects, j.intent.Image) == nil {
		t.Fatal("overlapping attempt admitted")
	}
	id := strings.Repeat("1", 64)
	if err = j.BindObjectRead(t.Context(), j.intent.Request, id); err != nil {
		t.Fatal(err)
	}
	if err = j.FinishObjectRead(t.Context(), j.intent.Request, id, true); err != nil {
		t.Fatal(err)
	}
	if err = another.BeginObjectRead(t.Context(), next, j.intent.Objects, j.intent.Image); err != nil {
		t.Fatal("fresh authorized attempt after settlement refused", err)
	}
	var missing *objectReadJournalGuard
	if _, err = newWorkerObjectReadJournal(l, next, j.intent.Image, j.intent.Objects, missing); err == nil {
		t.Fatal("missing guard admitted")
	}
	wrong := next
	wrong.Receipt.Intent.Scope.RoomID = "foreign-room"
	if _, err = newWorkerObjectReadJournal(l, wrong, j.intent.Image, j.intent.Objects, g); err == nil {
		t.Fatal("foreign room journal admitted")
	}
}
