package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

func remoteJobsSnapshot(ctx context.Context, alias string) (workerJobsSnapshot, error) {
	var result workerJobsSnapshot
	var output limitedAgentOutput
	cmd := workerSSHCommand(ctx, alias, "jobs-v2", "")
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if cmd.Run() != nil {
		return result, errors.New("worker_snapshot_unavailable")
	}
	if len(output.Bytes()) > 256*1024 || json.Unmarshal(output.Bytes(), &result) != nil || result.Version != 2 || len(result.Jobs) > 128 || result.Unresolved < 0 || result.Unresolved > 4096 || result.ManagedContainers < 0 || result.ManagedContainers > 4096 {
		return workerJobsSnapshot{}, errors.New("worker_snapshot_invalid")
	}
	seen := map[string]bool{}
	for _, view := range result.Jobs {
		if !validWorkerJobID(view.JobID) || seen[view.JobID] || (view.Provider != "codex" && view.Provider != "claude") || (view.LeaseState != "held" && view.LeaseState != "free" && view.LeaseState != "legacy") {
			return workerJobsSnapshot{}, errors.New("worker_snapshot_invalid")
		}
		seen[view.JobID] = true
		if view.Version != 0 && view.Version != 2 && !pairedLifecycleVersion(view.Version) {
			return workerJobsSnapshot{}, errors.New("worker_snapshot_invalid")
		}
		if view.Version == 2 || pairedLifecycleVersion(view.Version) {
			if !validWorkerRecord(view.workerJobRecord) {
				return workerJobsSnapshot{}, errors.New("worker_snapshot_invalid")
			}
			_, err := newWorkerJobRecord(workerRunRequest{JobID: view.JobID, Provider: view.Provider, Key: view.RoomID + "/" + view.ParticipantID + "/" + view.ConnectionID + "/worker"})
			if err != nil {
				return workerJobsSnapshot{}, errors.New("worker_snapshot_invalid")
			}
		}
	}
	return result, nil
}
func remoteFenceJob(ctx context.Context, alias string, request workerFenceRequest) error {
	expected, err := request.record()
	if err != nil {
		return err
	}
	raw, _ := json.Marshal(request)
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	var output limitedAgentOutput
	action := "cancel-v2"
	if request.Version == 3 {
		action = "cancel-pair-v1"
	}
	if request.Version == workerWorkspaceLifecycleVersion {
		action = "cancel-pair-workspace-v1"
	}
	cmd := workerSSHCommand(ctx, alias, action, encoded)
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if cmd.Run() != nil {
		return errors.New("worker_shutdown_unverified")
	}
	var ack workerFenceAck
	if len(output.Bytes()) > 4096 || json.Unmarshal(output.Bytes(), &ack) != nil || !validWorkerFenceAck(request, expected, ack) {
		return errors.New("worker_shutdown_unverified")
	}
	return nil
}
func (s *server) approvedSnapshot(ctx context.Context) (workerJobsSnapshot, error) {
	alias, _, err := remoteWorkerEndpoint("ssh://"+os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"), os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"), os.Getenv("COMMONROOM_REMOTE_WORKER"))
	if err != nil {
		return workerJobsSnapshot{}, errors.New("worker_not_configured")
	}
	if s.jobSnapshot != nil {
		return s.jobSnapshot(ctx, alias)
	}
	return remoteJobsSnapshot(ctx, alias)
}

type jobDiagnostic struct {
	room.ManagedJob
	WorkerStatus    string     `json:"workerStatus"`
	LeaseState      string     `json:"leaseState"`
	WorkerUpdatedAt *time.Time `json:"workerUpdatedAt,omitempty"`
	CanReconcile    bool       `json:"canReconcile"`
}

func sameWorkerBinding(job room.ManagedJob, view workerJobView) bool {
	if pairedLifecycleVersion(job.Version) {
		if view.Pair == nil || view.Execution != job.Execution || view.Pair.Repository != job.RepositoryDigest || !view.Deadline.Equal(job.Deadline) {
			return false
		}
	} else if view.Pair != nil || view.Execution != "" {
		return false
	}
	return view.Version == job.Version && (view.Version == 2 || pairedLifecycleVersion(view.Version)) && view.JobID == job.JobID && view.Provider == job.Provider && view.RoomID == job.RoomID && view.ParticipantID == job.ParticipantID && view.ConnectionID == job.ConnectionID
}
func boundedJobDiagnostics(st room.State, snapshot workerJobsSnapshot, snapshotErr error, rid, participant string) []jobDiagnostic {
	jobs := []jobDiagnostic{}
	for _, job := range st.ManagedJobs {
		if (rid != "" && job.RoomID != rid) || (participant != "" && job.ParticipantID != participant) {
			continue
		}
		item := jobDiagnostic{ManagedJob: job, WorkerStatus: "unverified", LeaseState: "unknown"}
		if snapshotErr == nil {
			item.WorkerStatus = "not_observed"
			for _, view := range snapshot.Jobs {
				if view.JobID == job.JobID {
					item.WorkerStatus = "unverified"
					if sameWorkerBinding(job, view) && validWorkerRecord(view.workerJobRecord) {
						switch view.Status {
						case "starting", "running", "finished", "failed", "cancelled", "shutdown_unconfirmed":
							item.WorkerStatus = view.Status
						}
						item.LeaseState = view.LeaseState
						at := view.UpdatedAt
						item.WorkerUpdatedAt = &at
						item.CanReconcile = view.LeaseState == "free" && !job.ShutdownConfirmed
					}
					break
				}
			}
			if item.WorkerStatus == "not_observed" && !snapshot.Truncated && !job.ShutdownConfirmed {
				item.CanReconcile = true
			}
		}
		jobs = append(jobs, item)
	}
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].ShutdownConfirmed != jobs[j].ShutdownConfirmed {
			return !jobs[i].ShutdownConfirmed
		}
		return jobs[i].UpdatedAt.After(jobs[j].UpdatedAt)
	})
	if len(jobs) > 128 {
		jobs = jobs[:128]
	}
	return jobs
}

func (s *server) jobDiagnostics(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "Job state unavailable", 503)
		return
	}
	if !jobsDebugEnabled(st) {
		http.Error(w, "Job diagnostics are disabled", 403)
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope != "" && scope != "room" && scope != "application" {
		http.Error(w, "Choose room/application diagnostics scope", 400)
		return
	}
	rid := requestedRoom(r)
	if scope == "application" {
		rid = ""
	}
	if rid != "" && !hasRoom(st, rid) {
		http.Error(w, "Room not found", 404)
		return
	}
	participant := r.URL.Query().Get("participantId")
	if participant != "" && !membershipPartPattern.MatchString(participant) {
		http.Error(w, "Invalid participant identity", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	snapshot, probeErr := s.approvedSnapshot(ctx)
	detail := "Worker snapshot verified over the approved SSH connection. A held supervisor lease indicates active ownership, not useful model progress."
	if probeErr != nil {
		detail = "Worker state could not be verified. No shutdown is assumed and no task will be replayed."
	}
	known := 0
	untrackedUnresolved := 0
	for _, view := range snapshot.Jobs {
		if _, ok := st.ManagedJobs[view.JobID]; ok {
			known++
		} else if !workerJobSettled(view.workerJobRecord) || view.LeaseState == "held" {
			untrackedUnresolved++
		}
	}
	views := boundedJobDiagnostics(st, snapshot, probeErr, rid, participant)
	s.mu.Lock()
	for i := range views {
		if s.active[views[i].ParticipantID] && s.activeJobs[views[i].ParticipantID] == views[i].JobID {
			views[i].CanReconcile = false
		}
	}
	s.mu.Unlock()
	reply(w, map[string]any{"turnBudget": s.dailyTurnBudget(st, time.Now()), "jobs": views, "verified": probeErr == nil, "checkedAt": time.Now().UTC(), "detail": detail, "untrackedWorkerJobs": len(snapshot.Jobs) - known, "untrackedUnresolvedJobs": untrackedUnresolved, "unresolvedWorkerJobs": snapshot.Unresolved, "managedContainers": snapshot.ManagedContainers, "workerSnapshotTruncated": snapshot.Truncated})
}

// jobRecovery is a required safety operation, not a premium feature. The UI
// diagnostic flag does not disable server cancellation/reconciliation safeguards.
func (s *server) jobRecovery(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope != "" && scope != "room" && scope != "application" {
		http.Error(w, "Choose room/application recovery scope", 400)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	var req struct {
		Action string `json:"action"`
		JobID  string `json:"jobId"`
		Reason string `json:"reason"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || !validWorkerJobID(req.JobID) || (req.Action != "cancel" && req.Action != "reconcile") || strings.TrimSpace(req.Reason) == "" || len(req.Reason) > 500 {
		http.Error(w, "Choose cancel/reconcile and provide a recovery reason", 400)
		return
	}
	s.mu.Lock()
	st, err := s.load(r.Context())
	job, found := st.ManagedJobs[req.JobID]
	if err != nil {
		s.mu.Unlock()
		http.Error(w, "Job state unavailable", 503)
		return
	}
	if !found || (r.URL.Query().Get("scope") != "application" && job.RoomID != requestedRoom(r)) {
		s.mu.Unlock()
		http.Error(w, "Job not found in this room", 404)
		return
	}
	if job.ShutdownConfirmed {
		s.mu.Unlock()
		reply(w, map[string]any{"stopped": true, "jobsRestarted": false})
		return
	}
	if s.active[job.ParticipantID] && (req.Action == "reconcile" || s.activeJobs[job.ParticipantID] != job.JobID) {
		s.mu.Unlock()
		http.Error(w, "Controller work is active; reconcile cannot replace or interrupt another turn", 409)
		return
	}
	auditCtx, auditCancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = s.append(auditCtx, "worker.job.recovery_requested", map[string]string{"jobId": job.JobID, "roomId": job.RoomID, "actorId": ownerID, "action": req.Action, "reason": req.Reason})
	auditCancel()
	if err != nil {
		s.tripIntegrityLocked()
		s.mu.Unlock()
		http.Error(w, "Recovery audit failed; no worker action was sent", 503)
		return
	}
	if req.Action == "cancel" && s.activeJobs[job.ParticipantID] == job.JobID {
		if stop := s.cancels[job.ParticipantID]; stop != nil {
			stop()
		}
	}
	s.mu.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
	defer cancel()
	if req.Action == "reconcile" {
		snapshot, err := s.approvedSnapshot(ctx)
		if err != nil {
			http.Error(w, "Worker state unverified; reconciliation refused", 503)
			return
		}
		for _, view := range snapshot.Jobs {
			if view.JobID == job.JobID && (!sameWorkerBinding(job, view) || !validWorkerRecord(view.workerJobRecord) || view.LeaseState != "free") {
				http.Error(w, "Worker supervisor remains active or binding is unverified; explicitly cancel if intended", 409)
				return
			}
		}
		if snapshot.Truncated {
			http.Error(w, "Worker snapshot is incomplete; reconciliation refused", 409)
			return
		}
	}
	if job.Worker != os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS") {
		http.Error(w, "Recorded worker is no longer approved", 409)
		return
	}
	shutdownErr, auditErr := s.reconcileManagedJob(ctx, job)
	if auditErr != nil {
		http.Error(w, "Worker result could not be audited; execution remains stopped", 503)
		return
	}
	err = shutdownErr
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		http.Error(w, "Worker shutdown unconfirmed; application stop remains active", 503)
		return
	}
	if !s.active[job.ParticipantID] {
		latest, loadErr := s.load(ctx)
		if loadErr == nil {
			for _, p := range latest.Participants {
				if p.ID == job.ParticipantID && p.RoomID == job.RoomID && p.ConnectionID == job.ConnectionID && p.Status == "Working" {
					p.Status = "Stopped"
					if s.append(ctx, "participant.session", p) != nil {
						s.tripIntegrityLocked()
						http.Error(w, "Session status audit failed; execution remains stopped", 503)
						return
					}
				}
			}
		}
	}
	reply(w, map[string]any{"stopped": true, "jobsRestarted": false, "detail": "Supervisor and container shutdown verified. Review partial work, then explicitly resume application execution if stopped."})
}
