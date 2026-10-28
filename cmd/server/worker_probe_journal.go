package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacemount"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

var errWorkerProbeJournal = errors.New("workspace probe journal unavailable")

const workerProbeStop = ".workspace-probe-integrity-stop"
const workerProbeRecordLimit = 1024

// workerProbeGuard MUST enforce current human/task/export authority, actual
// creator ownership, approved daemon/image and full reference inventory.
// Acknowledge must durably audit each transition and reserve account-wide probe
// capacity on begin before returning. StopAuthority must revoke controller/worker
// execution on ambiguous persistence. No production transport/default exists.
// Source forks CAN substitute dishonest guards, BUT invalidate these guarantees.
type workerProbeGuard interface {
	Authorize(context.Context, toolbroker.WorkspaceProbeIntent) error
	Acknowledge(context.Context, toolbroker.WorkspaceProbeIntent, workerProbeEvent) error
	StopAuthority()
}
type workerProbeEvent = room.WorkspaceProbeEvent

type workerProbeRecord struct {
	Version int                             `json:"version"`
	Intent  toolbroker.WorkspaceProbeIntent `json:"intent"`
	Events  []workerProbeEvent              `json:"events"`
}

// workerProbeJournal persists bounded lifecycle history under the shared worker
// admission lock. The caller MUST retain creator ownership through probe cleanup.
// This is internal infrastructure, not an exporter, public endpoint or FrozenLease.
type workerProbeJournal struct {
	dir     string
	guard   workerProbeGuard
	persist func(string, []byte, func(string) error) error
}

func newWorkerProbeJournal(dir string, guard workerProbeGuard) (*workerProbeJournal, error) {
	if nilWorkerObservationDependency(guard) || !filepath.IsAbs(dir) || filepath.Clean(dir) != dir {
		return nil, errWorkerProbeJournal
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return nil, errWorkerProbeJournal
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) {
		return nil, errWorkerProbeJournal
	}
	return &workerProbeJournal{dir: dir, guard: guard, persist: replacePrivateDurably}, nil
}
func probeHex(s string, n int) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == n && hex.EncodeToString(b) == s
}
func validWorkerProbeIntent(i toolbroker.WorkspaceProbeIntent) bool {
	return validWorkerJobID(i.JobID) && probeHex(i.Repository, 32) && strings.HasPrefix(i.Image, "sha256:") && probeHex(strings.TrimPrefix(i.Image, "sha256:"), 32) && workspacemount.ValidChallenge(i.Challenge) && toolbroker.ValidWorkspaceVolume(i.JobID, i.Volume, true)
}
func probeRecordName(i toolbroker.WorkspaceProbeIntent) string {
	return ".probe-" + i.JobID + "-" + i.Challenge + ".record"
}
func validWorkerProbeRecord(r workerProbeRecord) bool {
	if r.Version != 1 || !validWorkerProbeIntent(r.Intent) || len(r.Events) < 1 || len(r.Events) > 6 {
		return false
	}
	return validWorkerHelperEvents(r.Events)
}
func validWorkerHelperEvents(events []workerProbeEvent) bool {
	if len(events) < 1 || len(events) > 6 {
		return false
	}
	stage, id := "", ""
	for _, e := range events {
		switch e.Stage {
		case "begin":
			if stage != "" || e.ID != "" || e.Absent {
				return false
			}
		case "bound":
			if stage != "begin" || !probeHex(e.ID, 32) || e.Absent {
				return false
			}
			id = e.ID
		case "finish":
			if stage != "begin" && stage != "bound" {
				return false
			}
			if e.ID != "" && !probeHex(e.ID, 32) || id != "" && e.ID != id || e.Absent && e.ID == "" {
				return false
			}
			id = e.ID
		case "recovery":
			if (stage != "bound" && stage != "finish") || id == "" || e.ID != id || e.Absent {
				return false
			}
		case "recovery-finish":
			if stage != "recovery" || e.ID != id || !e.Absent {
				return false
			}
		case "recovered":
			if stage != "recovery-finish" || e.ID != id || !e.Absent {
				return false
			}
		case "confirmed":
			if stage != "finish" || !events[len(events)-2].Absent || !e.Absent || e.ID != id {
				return false
			}
		default:
			return false
		}
		stage = e.Stage
	}
	return true
}
func workerProbeStopped(dir string) bool {
	_, err := os.Lstat(filepath.Join(dir, workerProbeStop))
	return !os.IsNotExist(err)
}
func readWorkerProbeRecords(dir string) (map[string]workerProbeRecord, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entries, err := f.ReadDir(16385)
	if err != nil && err != io.EOF || len(entries) > 16384 {
		return nil, errWorkerProbeJournal
	}
	records := map[string]workerProbeRecord{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".probe-") {
			continue
		}
		if len(records) >= workerProbeRecordLimit {
			return nil, errWorkerProbeJournal
		}
		raw, err := readPrivateBounded(filepath.Join(dir, entry.Name()), 8192)
		if err != nil {
			return nil, errWorkerProbeJournal
		}
		var r workerProbeRecord
		if json.Unmarshal(raw, &r) != nil || !validWorkerProbeRecord(r) || probeRecordName(r.Intent) != entry.Name() {
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

// All callers hold .admission.lease. Missing/invalid evidence never frees a slot.
// The per-worker ceiling is conservative: retained creator roles also stay charged.
func workerProbeCapacity(dir string, jobs []workerJobRecord, next int) error {
	if workerProbeStopped(dir) {
		return errWorkerProbeJournal
	}
	records, err := readWorkerProbeRecords(dir)
	if err != nil {
		return err
	}
	captureSlots, err := workerCaptureSlots(dir)
	if err != nil {
		return err
	}
	readerSlots, err := workerObjectReadSlots(dir)
	if err != nil {
		return err
	}
	count := next + captureSlots + readerSlots
	for _, job := range jobs {
		if !workerJobSettled(job) {
			count += workerContainerSlots(job)
		}
	}
	for _, r := range records {
		if !workerProbeSettled(r) {
			count++
		}
	}
	if count > workerContainerLimit {
		return errWorkerProbeJournal
	}
	return nil
}
func (j *workerProbeJournal) fail() error {
	// Marker failures cannot become permission to continue: revoke live authority
	// independently. A deployment must treat missing/failed stop durability as an
	// integrity incident; there is no automatic reset/recovery endpoint.
	_ = replacePrivateDurably(filepath.Join(j.dir, workerProbeStop), []byte("stop"), syncParentDirectory)
	j.guard.StopAuthority()
	return errWorkerProbeJournal
}
func probeRecordMatchesWorker(dir string, i toolbroker.WorkspaceProbeIntent) bool {
	r, e := readWorkerRecord(dir, i.JobID)
	return e == nil && retainedObservationRecord(r) && r.Pair.Repository == i.Repository && r.Pair.Workspace.Identity == i.Volume
}
func (j *workerProbeJournal) transition(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	if j == nil || ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(j.guard) || j.persist == nil || !validWorkerProbeIntent(i) {
		return errWorkerProbeJournal
	}
	lock, err := workerWaitLease(ctx, j.dir, ".admission.lease")
	if err != nil {
		return err
	}
	defer lock.Close()
	records, err := readWorkerProbeRecords(j.dir)
	if err != nil {
		return j.fail()
	}
	name := probeRecordName(i)
	r, exists := records[name]
	if e.Stage == "begin" {
		if exists || len(records) >= workerProbeRecordLimit || workerProbeStopped(j.dir) {
			return errWorkerProbeJournal
		}
		for _, old := range records {
			if old.Intent.JobID == i.JobID && !workerProbeSettled(old) {
				return errWorkerProbeJournal
			}
		}
		jobs, err := boundedWorkerRecords(j.dir)
		if err != nil || workerProbeCapacity(j.dir, jobs, 1) != nil {
			return errWorkerProbeJournal
		}
		r = workerProbeRecord{Version: 1, Intent: i}
	} else if !exists || r.Intent != i {
		return errWorkerProbeJournal
	}
	if e.Stage != "finish" {
		if workerProbeStopped(j.dir) || !probeRecordMatchesWorker(j.dir, i) || j.guard.Authorize(ctx, i) != nil || ctx.Err() != nil {
			return errWorkerProbeJournal
		}
	}
	r.Events = append(r.Events, e)
	if !validWorkerProbeRecord(r) {
		return errWorkerProbeJournal
	}
	raw, _ := json.Marshal(r)
	if j.persist(filepath.Join(j.dir, name), raw, syncParentDirectory) != nil {
		return j.fail()
	}
	if j.guard.Acknowledge(ctx, i, e) != nil || ctx.Err() != nil {
		return j.fail()
	}
	if e.Stage == "finish" {
		// Cleanup evidence is retained even when release authority was revoked. Such
		// attempts stay charged; only a current final acknowledgment confirms release.
		if workerProbeStopped(j.dir) || !probeRecordMatchesWorker(j.dir, i) || j.guard.Authorize(ctx, i) != nil || ctx.Err() != nil {
			return errWorkerProbeJournal
		}
		if e.Absent {
			r.Events = append(r.Events, workerProbeEvent{Stage: "confirmed", ID: e.ID, Absent: true})
			raw, _ = json.Marshal(r)
			if j.persist(filepath.Join(j.dir, name), raw, syncParentDirectory) != nil {
				return j.fail()
			}
		}
	}
	if j.guard.Authorize(ctx, i) != nil || ctx.Err() != nil {
		return errWorkerProbeJournal
	}
	return nil
}
func (j *workerProbeJournal) Begin(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error {
	return j.transition(ctx, i, workerProbeEvent{Stage: "begin"})
}
func (j *workerProbeJournal) Bound(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string) error {
	return j.transition(ctx, i, workerProbeEvent{Stage: "bound", ID: id})
}
func (j *workerProbeJournal) Finish(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string, absent bool) error {
	return j.transition(ctx, i, workerProbeEvent{Stage: "finish", ID: id, Absent: absent})
}

func workerProbeSettled(r workerProbeRecord) bool {
	if len(r.Events) == 0 {
		return false
	}
	stage := r.Events[len(r.Events)-1].Stage
	return stage == "confirmed" || stage == "recovered"
}
