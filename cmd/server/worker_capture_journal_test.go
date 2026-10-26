package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacecapture"
	"agentroom/internal/workspacesnapshot"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type captureJournalGuard struct {
	f             *assembledProbeFixture
	deny, failAck string
	stopped       bool
	seen          []string
}

func (g *captureJournalGuard) AuthorizeCapture(ctx context.Context, r workerJobRecord, i workerCaptureIntent, e workerProbeEvent) error {
	g.f.checkLock()
	if e.Stage == g.deny {
		return errArtifactCustody
	}
	return ctx.Err()
}
func (g *captureJournalGuard) AcknowledgeCapture(ctx context.Context, r workerCaptureRecord, e workerProbeEvent) error {
	g.f.checkLock()
	records, err := readWorkerCaptureRecords(g.f.dir)
	saved, ok := records[captureRecordName(r.Intent)]
	if err != nil || !ok || len(saved.Events) == 0 || saved.Events[len(saved.Events)-1] != e {
		g.f.t.Fatal("ack before local persistence", err)
	}
	g.seen = append(g.seen, e.Stage)
	if e.Stage == g.failAck {
		return errors.New("synthetic controller audit unavailable")
	}
	return ctx.Err()
}
func (g *captureJournalGuard) StopCaptureAuthority() { g.stopped = true }
func captureJournalFixtureScope(t *testing.T) (*workerCaptureJournal, *captureJournalGuard, *workerWorkspaceObservation) {
	t.Helper()
	f := newAssembledProbeFixture(t, "")
	lease, free, err := workerTryLease(f.dir, ".job-"+f.record.JobID+".lease")
	if err != nil || !free {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(func() { cancel(); lease.Close() })
	v := f.record.Pair.Workspace.Identity
	q := workspacecapture.Request{Binding: workspacesnapshot.Binding{Execution: objectcustody.ExecutionScope{CompanyID: "company", RoomID: f.record.RoomID, TaskID: "task", ExecutionID: strings.Repeat("a", 32), JobID: f.record.JobID, RepositoryDigest: f.record.Pair.Repository, DefinitionDigest: strings.Repeat("b", 64)}, BaseCommit: strings.Repeat("c", 40), VolumeName: v.Name, VolumeNonce: v.Nonce, VolumeCreatedAt: v.CreatedAt}, Members: []workspacesnapshot.Member{{OutputName: "result", Name: "result.txt", MaxBytes: 128}}, Challenge: strings.Repeat("e", 32)}
	l := &workerWorkspaceObservation{ownership: lease, active: true, lifetime: ctx, dir: f.dir, dockerRoot: syntheticObservationDockerRoot, record: f.record}
	g := &captureJournalGuard{f: f}
	j, err := newWorkerCaptureJournal(l, q, "sha256:"+strings.Repeat("d", 64), toolbroker.CaptureObjectVolume{Name: "commonroom-custody-" + f.record.JobID, Nonce: strings.Repeat("f", 32), CreatedAt: v.CreatedAt}, g)
	if err != nil {
		t.Fatal(err)
	}
	return j, g, l
}
func captureJournalFixture(t *testing.T) (*workerCaptureJournal, *captureJournalGuard) {
	j, g, _ := captureJournalFixtureScope(t)
	return j, g
}
func beginCaptureJournal(t *testing.T, j *workerCaptureJournal) {
	t.Helper()
	if err := j.BeginCapture(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image); err != nil {
		t.Fatal(err)
	}
}
func TestWorkerCaptureJournalDurableLifecycleAndCapacity(t *testing.T) {
	t.Parallel()
	j, g := captureJournalFixture(t)
	beginCaptureJournal(t, j)
	if n, e := workerCaptureSlots(j.dir); e != nil || n != 1 {
		t.Fatal("missing capture charge", n, e)
	}
	jobs, e := boundedWorkerRecords(j.dir)
	if e != nil {
		t.Fatal(e)
	}
	if workerProbeCapacity(j.dir, jobs, workerContainerLimit) == nil {
		t.Fatal("capture ignored by admission")
	}
	id := strings.Repeat("1", 64)
	if j.BindCapture(t.Context(), j.intent.Request, id) != nil || j.FinishCapture(t.Context(), j.intent.Request, id, true) != nil {
		t.Fatal("normal journal lifecycle")
	}
	records, e := readWorkerCaptureRecords(j.dir)
	if e != nil || len(records) != 1 || !captureRecordSettled(records[captureRecordName(j.intent)]) {
		t.Fatal("missing durable confirmation", e)
	}
	if n, e := workerCaptureSlots(j.dir); e != nil || n != 0 {
		t.Fatal("helper slot not released", n, e)
	}
	if len(g.seen) != 3 || g.stopped {
		t.Fatal("wrong controller ack order", g.seen)
	}
	if j.BeginCapture(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image) == nil {
		t.Fatal("capture replayed")
	}
}
func TestWorkerCaptureJournalKnownRecoveryAndUnknownRetention(t *testing.T) {
	t.Parallel()
	for _, known := range []bool{false, true} {
		t.Run(map[bool]string{true: "known", false: "unknown"}[known], func(t *testing.T) {
			t.Parallel()
			j, _ := captureJournalFixture(t)
			beginCaptureJournal(t, j)
			id := ""
			if known {
				id = strings.Repeat("1", 64)
				if j.BindCapture(t.Context(), j.intent.Request, id) != nil {
					t.Fatal("bind")
				}
			}
			if j.FinishCapture(t.Context(), j.intent.Request, id, false) != nil {
				t.Fatal("uncertain finish")
			}
			if n, e := workerCaptureSlots(j.dir); e != nil || n != 1 {
				t.Fatal("uncertainty released slot", n, e)
			}
			candidate := id
			if candidate == "" {
				candidate = strings.Repeat("9", 64)
			}
			err := j.BeginCaptureRecovery(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image, candidate)
			if !known {
				if err == nil {
					t.Fatal("unknown ID adopted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if j.FinishCaptureRecovery(t.Context(), j.intent.Request, id, false) == nil {
				t.Fatal("unconfirmed recovery released")
			}
			if j.FinishCaptureRecovery(t.Context(), j.intent.Request, id, true) != nil {
				t.Fatal("verified recovery failed")
			}
			if n, e := workerCaptureSlots(j.dir); e != nil || n != 0 {
				t.Fatal("recovered slot charged", n, e)
			}
		})
	}
}
func TestWorkerCaptureJournalAmbiguousPersistenceStopsAuthority(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"local begin", "ack begin", "ack bound", "ack finish", "local confirm"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, g := captureJournalFixture(t)
			real := j.persist
			if mode == "local begin" {
				j.persist = func(string, []byte, func(string) error) error { return errors.New("synthetic sync failed") }
			}
			if mode == "ack begin" {
				g.failAck = "begin"
			}
			err := j.BeginCapture(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image)
			if mode == "ack bound" || mode == "ack finish" || mode == "local confirm" {
				if err != nil {
					t.Fatal(err)
				}
				id := strings.Repeat("1", 64)
				if mode == "ack bound" {
					g.failAck = "bound"
				}
				err = j.BindCapture(t.Context(), j.intent.Request, id)
				if mode != "ack bound" {
					if err != nil {
						t.Fatal(err)
					}
					if mode == "ack finish" {
						g.failAck = "finish"
					} else {
						count := 0
						j.persist = func(p string, b []byte, sync func(string) error) error {
							count++
							if count == 2 {
								return errors.New("synthetic confirmation failed")
							}
							return real(p, b, sync)
						}
					}
					err = j.FinishCapture(t.Context(), j.intent.Request, id, true)
				}
			}
			if err == nil || !g.stopped || !workerProbeStopped(j.dir) {
				t.Fatal("ambiguous state did not stop", err)
			}
			if workerProbeCapacity(j.dir, nil, 0) == nil {
				t.Fatal("integrity stop bypassed")
			}
		})
	}
}
func TestWorkerCaptureJournalRefusesSubstitutionAndUnsafeRecords(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"request", "image", "objects", "bound before begin", "duplicate bound", "malformed", "noncanonical", "symlink", "unsafe mode"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, g := captureJournalFixture(t)
			q := copyCaptureRequest(j.intent.Request)
			switch mode {
			case "request":
				q.Members[0].MaxBytes++
				if j.BeginCapture(t.Context(), q, j.intent.Objects, j.intent.Image) == nil {
					t.Fatal("request substituted")
				}
				return
			case "image":
				if j.BeginCapture(t.Context(), q, j.intent.Objects, "sha256:"+strings.Repeat("9", 64)) == nil {
					t.Fatal("image substituted")
				}
				return
			case "objects":
				o := j.intent.Objects
				o.Nonce = strings.Repeat("9", 32)
				if j.BeginCapture(t.Context(), q, o, j.intent.Image) == nil {
					t.Fatal("volume substituted")
				}
				return
			case "bound before begin":
				if j.BindCapture(t.Context(), q, strings.Repeat("1", 64)) == nil {
					t.Fatal("bound before intent")
				}
				return
			}
			beginCaptureJournal(t, j)
			if mode == "duplicate bound" {
				id := strings.Repeat("1", 64)
				j.BindCapture(t.Context(), q, id)
				if j.BindCapture(t.Context(), q, id) == nil {
					t.Fatal("bound replay")
				}
				return
			}
			path := filepath.Join(j.dir, captureRecordName(j.intent))
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "malformed":
				raw = []byte(`{}`)
			case "noncanonical":
				raw = append(raw, '\n')
			case "symlink":
				target := filepath.Join(t.TempDir(), "synthetic-record")
				os.WriteFile(target, raw, 0600)
				os.Remove(path)
				os.Symlink(target, path)
			case "unsafe mode":
				os.Chmod(path, 0644)
			}
			if mode == "malformed" || mode == "noncanonical" {
				os.WriteFile(path, raw, 0600)
			}
			if _, err = readWorkerCaptureRecords(j.dir); err == nil {
				t.Fatal("unsafe record read")
			}
			if workerProbeCapacity(j.dir, nil, 0) == nil {
				t.Fatal("unsafe records ignored by admission")
			}
			if g.stopped {
				t.Fatal("read-only validation manufactured stop acknowledgment")
			}
		})
	}
}

func TestWorkerCaptureJournalAuthorityDenialRetainsCharges(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"begin", "bound", "finish"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			j, g := captureJournalFixture(t)
			if stage == "begin" {
				g.deny = stage
				if j.BeginCapture(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image) == nil {
					t.Fatal("denied begin granted")
				}
				if n, e := workerCaptureSlots(j.dir); e != nil || n != 0 {
					t.Fatal("denied begin created record", n, e)
				}
				return
			}
			beginCaptureJournal(t, j)
			id := strings.Repeat("1", 64)
			if stage == "bound" {
				g.deny = stage
				if j.BindCapture(t.Context(), j.intent.Request, id) == nil {
					t.Fatal("denied binding granted")
				}
			} else {
				if j.BindCapture(t.Context(), j.intent.Request, id) != nil {
					t.Fatal("bind")
				}
				g.deny = stage
				if j.FinishCapture(t.Context(), j.intent.Request, id, true) == nil {
					t.Fatal("revoked release granted")
				}
			}
			if n, e := workerCaptureSlots(j.dir); e != nil || n != 1 {
				t.Fatal("denial released charge", n, e)
			}
			if g.stopped {
				t.Fatal("ordinary denial manufactured integrity failure")
			}
		})
	}
}
func TestWorkerCaptureJournalBlocksSecondUnsettledCapture(t *testing.T) {
	t.Parallel()
	j, _ := captureJournalFixture(t)
	beginCaptureJournal(t, j)
	other := *j
	other.intent = copyCaptureIntent(j.intent)
	other.intent.Request.Challenge = strings.Repeat("9", 32)
	if other.BeginCapture(t.Context(), other.intent.Request, other.intent.Objects, other.intent.Image) == nil {
		t.Fatal("second capture admitted for unsettled job")
	}
}
