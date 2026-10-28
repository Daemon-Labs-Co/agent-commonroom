package main

import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacecapture"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// workerCaptureIntent contains finite immutable metadata, never execution rights.
type workerCaptureIntent = toolbroker.CaptureIntent

type workerCaptureRecord struct {
	Version int                 `json:"version"`
	Intent  workerCaptureIntent `json:"intent"`
	Events  []workerProbeEvent  `json:"events"`
}

// workerCaptureJournalGuard MUST authenticate current operation-specific source or
// cleanup authority, actual creator/quota/image/volume evidence, and account-wide
// capacity. AcknowledgeCapture must durably mirror the exact local transition
// before returning; failed persistence must stop controller/worker authority.
// It must not reenter admission locks. No permissive/default implementation exists.
type workerCaptureJournalGuard interface {
	AuthorizeCapture(context.Context, workerJobRecord, workerCaptureIntent, workerProbeEvent) error
	AcknowledgeCapture(context.Context, workerCaptureRecord, workerProbeEvent) error
	StopCaptureAuthority()
}
type workerCaptureJournal struct {
	dir      string
	lifetime context.Context
	record   workerJobRecord
	intent   workerCaptureIntent
	guard    workerCaptureJournalGuard
	persist  func(string, []byte, func(string) error) error
}

func captureRecordName(i workerCaptureIntent) string {
	return ".capture-" + i.Request.Binding.Execution.JobID + "-" + i.Request.Challenge + ".record"
}
func validWorkerCaptureRecord(r workerCaptureRecord) bool {
	if r.Version != 1 || !validWorkerHelperEvents(r.Events) {
		return false
	}
	return toolbroker.ValidCaptureMetadata(r.Intent.Request, r.Intent.Image, r.Intent.Objects)
}
func readWorkerCaptureRecords(dir string) (map[string]workerCaptureRecord, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entries, err := f.ReadDir(16385)
	if err != nil && err != io.EOF || len(entries) > 16384 {
		return nil, errWorkerProbeJournal
	}
	records := map[string]workerCaptureRecord{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".capture-") {
			continue
		}
		if len(records) >= workerProbeRecordLimit {
			return nil, errWorkerProbeJournal
		}
		raw, err := readPrivateBounded(filepath.Join(dir, entry.Name()), 128*1024)
		var r workerCaptureRecord
		if err != nil || json.Unmarshal(raw, &r) != nil || !validWorkerCaptureRecord(r) || captureRecordName(r.Intent) != entry.Name() {
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
func captureRecordSettled(r workerCaptureRecord) bool {
	if len(r.Events) == 0 {
		return false
	}
	s := r.Events[len(r.Events)-1].Stage
	return s == "confirmed" || s == "recovered"
}
func workerCaptureSlots(dir string) (int, error) {
	records, err := readWorkerCaptureRecords(dir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, r := range records {
		if !captureRecordSettled(r) {
			count++
		}
	}
	return count, nil
}
func newWorkerCaptureJournal(l *workerWorkspaceObservation, q workspacecapture.Request, image string, objects toolbroker.CaptureObjectVolume, g workerCaptureJournalGuard) (*workerCaptureJournal, error) {
	if l == nil || nilWorkerObservationDependency(g) {
		return nil, errWorkerProbeJournal
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || !retainedObservationRecord(l.record) {
		return nil, errWorkerProbeJournal
	}
	if _, err := toolbroker.NewCaptureContainerPlan(q, image, l.dockerRoot, objects); err != nil {
		return nil, err
	}
	v := l.record.Pair.Workspace.Identity
	b := q.Binding
	if b.Execution.JobID != l.record.JobID || b.Execution.RepositoryDigest != l.record.Pair.Repository || b.VolumeName != v.Name || b.VolumeNonce != v.Nonce || b.VolumeCreatedAt != v.CreatedAt {
		return nil, errWorkerProbeJournal
	}
	return &workerCaptureJournal{l.dir, l.lifetime, detachedProbeWorkerRecord(l.record), workerCaptureIntent{Request: copyCaptureRequest(q), Image: image, Objects: objects}, g, replacePrivateDurably}, nil
}
func (j *workerCaptureJournal) fail() error {
	_ = replacePrivateDurably(filepath.Join(j.dir, workerProbeStop), []byte("stop"), syncParentDirectory)
	j.guard.StopCaptureAuthority()
	return errWorkerProbeJournal
}
func copyCaptureIntent(i workerCaptureIntent) workerCaptureIntent {
	i.Request = copyCaptureRequest(i.Request)
	return i
}
func copyCaptureRecord(r workerCaptureRecord) workerCaptureRecord {
	r.Intent = copyCaptureIntent(r.Intent)
	r.Events = slices.Clone(r.Events)
	return r
}
func captureIntentEqual(a, b workerCaptureIntent) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return bytes.Equal(x, y)
}
func (j *workerCaptureJournal) matches(q workspacecapture.Request) bool {
	return j != nil && captureIntentEqual(j.intent, workerCaptureIntent{Request: q, Image: j.intent.Image, Objects: j.intent.Objects})
}
func (j *workerCaptureJournal) transition(ctx context.Context, event workerProbeEvent) error {
	if j == nil || ctx == nil || ctx.Err() != nil || j.lifetime == nil || j.lifetime.Err() != nil || nilWorkerObservationDependency(j.guard) || j.persist == nil {
		return errWorkerProbeJournal
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(j.lifetime, cancel)
	defer stop()
	lock, err := workerWaitLease(child, j.dir, ".admission.lease")
	if err != nil {
		return err
	}
	defer lock.Close()
	records, err := readWorkerCaptureRecords(j.dir)
	if err != nil {
		return j.fail()
	}
	name := captureRecordName(j.intent)
	r, exists := records[name]
	if event.Stage == "begin" {
		if exists || len(records) >= workerProbeRecordLimit || workerProbeStopped(j.dir) {
			return errWorkerProbeJournal
		}
		for _, old := range records {
			if old.Intent.Request.Binding.Execution.JobID == j.record.JobID && !captureRecordSettled(old) {
				return errWorkerProbeJournal
			}
		}
		readers, err := readWorkerObjectReadRecords(j.dir)
		if err != nil {
			return j.fail()
		}
		for _, old := range readers {
			if old.Intent.Request.Receipt.Intent.Scope.JobID == j.record.JobID && !objectReadRecordSettled(old) {
				return errWorkerProbeJournal
			}
		}
		jobs, err := boundedWorkerRecords(j.dir)
		if err != nil || workerProbeCapacity(j.dir, jobs, 1) != nil {
			return errWorkerProbeJournal
		}
		r = workerCaptureRecord{Version: 1, Intent: j.intent}
	} else if !exists || !captureIntentEqual(r.Intent, j.intent) {
		return errWorkerProbeJournal
	}
	if _, err = readRetainedWorkspaceObservation(j.dir, j.record); err != nil {
		return errWorkerProbeJournal
	}
	if event.Stage != "finish" {
		if workerProbeStopped(j.dir) || j.guard.AuthorizeCapture(child, detachedProbeWorkerRecord(j.record), copyCaptureIntent(j.intent), event) != nil || child.Err() != nil {
			return errWorkerProbeJournal
		}
	}
	r.Events = append(r.Events, event)
	if !validWorkerCaptureRecord(r) {
		return errWorkerProbeJournal
	}
	raw, _ := json.Marshal(r)
	if j.persist(filepath.Join(j.dir, name), raw, syncParentDirectory) != nil {
		return j.fail()
	}
	if j.guard.AcknowledgeCapture(child, copyCaptureRecord(r), event) != nil || child.Err() != nil {
		return j.fail()
	}
	if workerProbeStopped(j.dir) || j.guard.AuthorizeCapture(child, detachedProbeWorkerRecord(j.record), copyCaptureIntent(j.intent), event) != nil || child.Err() != nil {
		return errWorkerProbeJournal
	}
	if event.Stage == "finish" && event.Absent || event.Stage == "recovery-finish" {
		stage := "confirmed"
		if event.Stage == "recovery-finish" {
			stage = "recovered"
		}
		r.Events = append(r.Events, workerProbeEvent{Stage: stage, ID: event.ID, Absent: true})
		if !validWorkerCaptureRecord(r) {
			return j.fail()
		}
		raw, _ = json.Marshal(r)
		if j.persist(filepath.Join(j.dir, name), raw, syncParentDirectory) != nil {
			return j.fail()
		}
	}
	return nil
}
func (j *workerCaptureJournal) BeginCapture(ctx context.Context, q workspacecapture.Request, o toolbroker.CaptureObjectVolume, image string) error {
	if !j.matches(q) || o != j.intent.Objects || image != j.intent.Image {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "begin"})
}
func (j *workerCaptureJournal) BindCapture(ctx context.Context, q workspacecapture.Request, id string) error {
	if !j.matches(q) {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "bound", ID: id})
}
func (j *workerCaptureJournal) FinishCapture(ctx context.Context, q workspacecapture.Request, id string, absent bool) error {
	if !j.matches(q) {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "finish", ID: id, Absent: absent})
}
func (j *workerCaptureJournal) BeginCaptureRecovery(ctx context.Context, q workspacecapture.Request, o toolbroker.CaptureObjectVolume, image, id string) error {
	if !j.matches(q) || o != j.intent.Objects || image != j.intent.Image {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "recovery", ID: id})
}
func (j *workerCaptureJournal) FinishCaptureRecovery(ctx context.Context, q workspacecapture.Request, id string, absent bool) error {
	if !j.matches(q) || !absent {
		return errWorkerProbeJournal
	}
	return j.transition(ctx, workerProbeEvent{Stage: "recovery-finish", ID: id, Absent: true})
}

var _ toolbroker.CaptureLifecycleJournal = (*workerCaptureJournal)(nil)
var _ toolbroker.CaptureRecoveryJournal = (*workerCaptureJournal)(nil)
