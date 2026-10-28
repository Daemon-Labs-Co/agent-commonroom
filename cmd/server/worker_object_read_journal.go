package main

import (
	"agentroom/internal/objectread"
	"agentroom/internal/toolbroker"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// workerObjectReadIntent contains finite immutable metadata, never execution rights.
// This local adapter requires a retained-workspace creator lease. It grants no host
// read path, public download or isolation toggle. Source forks can fake guards and
// invalidate the mandatory containment/physical-quota contract. Account-wide audit,
// actual sealed-receipt custody and export authority MUST be independently checked.
type workerObjectReadIntent = toolbroker.ObjectReadIntent

type workerObjectReadRecord struct {
	Version int                    `json:"version"`
	Intent  workerObjectReadIntent `json:"intent"`
	Events  []workerProbeEvent     `json:"events"`
}

// workerObjectReadJournalGuard MUST authenticate current operation-specific reads or
// cleanup authority, actual creator/quota/image/volume evidence, and account-wide
// capacity. AcknowledgeObjectRead must durably mirror the exact local transition
// before returning; failed persistence must stop controller/worker authority.
// An exact already-begun recovery may be acknowledged again after fresh current
// authorization. It resumes the same logical cleanup, never creates a new reader.
// It must not reenter creator/admission locks. Finishing uncertainty must be
// acknowledged even when read authority was canceled; it grants no export rights. No permissive/default implementation exists.
type workerObjectReadJournalGuard interface {
	AuthorizeObjectRead(context.Context, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error
	AcknowledgeObjectRead(context.Context, workerObjectReadRecord, workerProbeEvent) error
	StopObjectReadAuthority()
}
type workerObjectReadJournal struct {
	owner    *workerWorkspaceObservation
	dir      string
	lifetime context.Context
	record   workerJobRecord
	intent   workerObjectReadIntent
	guard    workerObjectReadJournalGuard
	persist  func(string, []byte, func(string) error) error
}

func objectReadRecordName(i workerObjectReadIntent) string {
	return ".object-read-" + i.Request.Receipt.Intent.Scope.JobID + "-" + i.Request.Challenge + ".record"
}
func validWorkerObjectReadRecord(r workerObjectReadRecord) bool {
	if r.Version != 1 || !validWorkerHelperEvents(r.Events) {
		return false
	}
	return toolbroker.ValidObjectReadMetadata(r.Intent.Request, r.Intent.Image, r.Intent.Objects)
}
func readWorkerObjectReadRecords(dir string) (map[string]workerObjectReadRecord, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entries, err := f.ReadDir(16385)
	if err != nil && err != io.EOF || len(entries) > 16384 {
		return nil, errWorkerProbeJournal
	}
	records := map[string]workerObjectReadRecord{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".object-read-") {
			continue
		}
		if len(records) >= workerProbeRecordLimit {
			return nil, errWorkerProbeJournal
		}
		raw, err := readPrivateBounded(filepath.Join(dir, entry.Name()), 16*1024)
		var r workerObjectReadRecord
		if err != nil || json.Unmarshal(raw, &r) != nil || !validWorkerObjectReadRecord(r) || objectReadRecordName(r.Intent) != entry.Name() {
			return nil, errWorkerProbeJournal
		}
		canonical, _ := json.Marshal(r)
		if !bytes.Equal(raw, canonical) {
			return nil, errWorkerProbeJournal
		}
		records[entry.Name()] = r
	}
	return records, nil
}
func objectReadRecordSettled(r workerObjectReadRecord) bool {
	if len(r.Events) == 0 {
		return false
	}
	s := r.Events[len(r.Events)-1].Stage
	return s == "confirmed" || s == "recovered"
}
func workerObjectReadSlots(dir string) (int, error) {
	records, err := readWorkerObjectReadRecords(dir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, r := range records {
		if !objectReadRecordSettled(r) {
			count++
		}
	}
	return count, nil
}
func newWorkerObjectReadJournal(l *workerWorkspaceObservation, q objectread.Request, image string, objects toolbroker.CaptureObjectVolume, g workerObjectReadJournalGuard) (*workerObjectReadJournal, error) {
	if l == nil || nilWorkerObservationDependency(g) {
		return nil, errWorkerProbeJournal
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || !retainedObservationRecord(l.record) {
		return nil, errWorkerProbeJournal
	}
	if _, err := toolbroker.NewObjectReadContainerPlan(q, image, l.dockerRoot, objects); err != nil {
		return nil, err
	}
	b := q.Receipt.Intent.Scope
	if b.JobID != l.record.JobID || b.RepositoryDigest != l.record.Pair.Repository || b.RoomID != l.record.RoomID {
		return nil, errWorkerProbeJournal
	}
	return &workerObjectReadJournal{l, l.dir, l.lifetime, detachedProbeWorkerRecord(l.record), workerObjectReadIntent{Request: q, Image: image, Objects: objects}, g, replacePrivateDurably}, nil
}
func (j *workerObjectReadJournal) fail() error {
	_ = replacePrivateDurably(filepath.Join(j.dir, workerProbeStop), []byte("stop"), syncParentDirectory)
	j.guard.StopObjectReadAuthority()
	return errWorkerProbeJournal
}
func copyObjectReadRecord(r workerObjectReadRecord) workerObjectReadRecord {
	r.Events = slices.Clone(r.Events)
	return r
}
func objectReadIntentEqual(a, b workerObjectReadIntent) bool { return a == b }
func (j *workerObjectReadJournal) matches(q objectread.Request) bool {
	return j != nil && objectReadIntentEqual(j.intent, workerObjectReadIntent{Request: q, Image: j.intent.Image, Objects: j.intent.Objects})
}
func (j *workerObjectReadJournal) transition(ctx context.Context, event workerProbeEvent) error {
	if j == nil || ctx == nil || ctx.Err() != nil || j.owner == nil || j.lifetime == nil || (event.Stage != "finish" && j.lifetime.Err() != nil) || nilWorkerObservationDependency(j.guard) || j.persist == nil {
		return errWorkerProbeJournal
	}
	// Keep the actual creator lease alive through persistence and callbacks.
	// Finishing uncertainty may outlive canceled read authority, never ownership.
	j.owner.mu.Lock()
	defer j.owner.mu.Unlock()
	if !j.owner.active || j.owner.ownership == nil || j.owner.ownership.file == nil {
		return errWorkerProbeJournal
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	if event.Stage != "finish" {
		stop := context.AfterFunc(j.lifetime, cancel)
		defer stop()
	}
	lock, err := workerWaitLease(child, j.dir, ".admission.lease")
	if err != nil {
		return err
	}
	defer lock.Close()
	records, err := readWorkerObjectReadRecords(j.dir)
	if err != nil {
		return j.fail()
	}
	name := objectReadRecordName(j.intent)
	r, exists := records[name]
	if event.Stage == "begin" {
		if exists || len(records) >= workerProbeRecordLimit || workerProbeStopped(j.dir) {
			return errWorkerProbeJournal
		}
		for _, old := range records {
			if old.Intent.Request.Receipt.Intent.Scope.JobID == j.record.JobID && !objectReadRecordSettled(old) {
				return errWorkerProbeJournal
			}
		}
		captures, err := readWorkerCaptureRecords(j.dir)
		if err != nil {
			return j.fail()
		}
		for _, old := range captures {
			if old.Intent.Request.Binding.Execution.JobID == j.record.JobID && !captureRecordSettled(old) {
				return errWorkerProbeJournal
			}
		}
		jobs, err := boundedWorkerRecords(j.dir)
		if err != nil || workerProbeCapacity(j.dir, jobs, 1) != nil {
			return errWorkerProbeJournal
		}
		r = workerObjectReadRecord{Version: 1, Intent: j.intent}
	} else if !exists || !objectReadIntentEqual(r.Intent, j.intent) {
		return errWorkerProbeJournal
	}
	if _, err = readRetainedWorkspaceObservation(j.dir, j.record); err != nil {
		return errWorkerProbeJournal
	}
	if event.Stage != "finish" {
		if workerProbeStopped(j.dir) || j.guard.AuthorizeObjectRead(child, detachedProbeWorkerRecord(j.record), j.intent, event) != nil || child.Err() != nil {
			return errWorkerProbeJournal
		}
	}
	// A failed/interrupted cleanup may leave the logical recovery already begun.
	// Resume only that exact recorded ID; keep finite history unchanged while
	// repeating persistence, current authorization and durable acknowledgment.
	resumingRecovery := event.Stage == "recovery" && len(r.Events) > 0 && r.Events[len(r.Events)-1] == event
	if !resumingRecovery {
		r.Events = append(r.Events, event)
	}
	if !validWorkerObjectReadRecord(r) {
		return errWorkerProbeJournal
	}
	raw, _ := json.Marshal(r)
	if j.persist(filepath.Join(j.dir, name), raw, syncParentDirectory) != nil {
		return j.fail()
	}
	if j.guard.AcknowledgeObjectRead(child, copyObjectReadRecord(r), event) != nil || child.Err() != nil {
		return j.fail()
	}
	if j.lifetime.Err() != nil || workerProbeStopped(j.dir) || j.guard.AuthorizeObjectRead(child, detachedProbeWorkerRecord(j.record), j.intent, event) != nil || child.Err() != nil {
		return errWorkerProbeJournal
	}
	if event.Stage == "finish" && event.Absent || event.Stage == "recovery-finish" {
		stage := "confirmed"
		if event.Stage == "recovery-finish" {
			stage = "recovered"
		}
		r.Events = append(r.Events, workerProbeEvent{Stage: stage, ID: event.ID, Absent: true})
		if !validWorkerObjectReadRecord(r) {
			return j.fail()
		}
		raw, _ = json.Marshal(r)
		if j.persist(filepath.Join(j.dir, name), raw, syncParentDirectory) != nil {
			return j.fail()
		}
	}
	return nil
}
func (j *workerObjectReadJournal) BeginObjectRead(ctx context.Context, q objectread.Request, o toolbroker.CaptureObjectVolume, image string) error {
	if !j.matches(q) || o != j.intent.Objects || image != j.intent.Image {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "begin"})
}
func (j *workerObjectReadJournal) BindObjectRead(ctx context.Context, q objectread.Request, id string) error {
	if !j.matches(q) {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "bound", ID: id})
}
func (j *workerObjectReadJournal) FinishObjectRead(ctx context.Context, q objectread.Request, id string, absent bool) error {
	if !j.matches(q) {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "finish", ID: id, Absent: absent})
}
func (j *workerObjectReadJournal) BeginObjectReadRecovery(ctx context.Context, q objectread.Request, o toolbroker.CaptureObjectVolume, image, id string) error {
	if !j.matches(q) || o != j.intent.Objects || image != j.intent.Image {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "recovery", ID: id})
}
func (j *workerObjectReadJournal) FinishObjectReadRecovery(ctx context.Context, q objectread.Request, id string, absent bool) error {
	if !j.matches(q) || !absent {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "recovery-finish", ID: id, Absent: true})
}

var _ toolbroker.ObjectReadLifecycleJournal = (*workerObjectReadJournal)(nil)
var _ toolbroker.ObjectReadRecoveryJournal = (*workerObjectReadJournal)(nil)
