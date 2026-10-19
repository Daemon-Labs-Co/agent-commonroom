package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

func disposalOwner(st room.State, actor string) bool {
	a, ok := member(st, actor)
	return actor == ownerID && ok && a.Enabled && a.Kind == "human"
}
func disposalList(st room.State) []room.WorkspaceDisposal {
	list := []room.WorkspaceDisposal{}
	for _, p := range st.WorkspaceDisposals {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		unresolved := func(p room.WorkspaceDisposal) bool { return p.Status == "claimed" || p.Status == "unconfirmed" }
		if unresolved(list[i]) != unresolved(list[j]) {
			return unresolved(list[i])
		}
		if list[i].CreatedAt.Equal(list[j].CreatedAt) {
			return list[i].ID < list[j].ID
		}
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	if len(list) > 128 {
		list = list[:128]
	}
	return list
}

// Reviewed workspace disposal is off by default, owner-only and one-shot. Worker
// operations require a committed claim. No automatic retry, host fallback or
// storage refund exists. A crash leaves a durable unresolved claim for review.
func (s *server) workspaceDisposal(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Method == "POST" && !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	var req struct {
		Action     string `json:"action"`
		JobID      string `json:"jobId"`
		ProposalID string `json:"proposalId"`
		Confirm    string `json:"confirm"`
	}
	if r.Method == "POST" {
		raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
		if err != nil {
			http.Error(w, "Invalid disposal request", 400)
			return
		}
		if _, err = fenceObjectFields(raw, "action", "jobId", "proposalId", "confirm"); err != nil || json.Unmarshal(raw, &req) != nil || (req.Action != "propose" && req.Action != "approve" && req.Action != "revoke") {
			http.Error(w, "Choose propose, approve or revoke", 400)
			return
		}
	}
	s.mu.Lock()
	st, err := s.load(r.Context())
	if err != nil || s.integrityFault {
		s.mu.Unlock()
		http.Error(w, "Disposal state unavailable", 503)
		return
	}
	if !disposalOwner(st, requestActor(r)) {
		s.mu.Unlock()
		http.Error(w, "Enabled human owner required", 403)
		return
	}
	if r.Method == "GET" {
		reviews := s.disposalReviews(st)
		now := s.disposalTime()
		s.mu.Unlock()
		reply(w, map[string]any{"enabled": st.FeatureFlags["workspaceDisposal"], "proposals": reviews, "now": now})
		return
	}
	if !st.FeatureFlags["workspaceDisposal"] && req.Action != "revoke" {
		s.mu.Unlock()
		http.Error(w, "Workspace disposal is disabled", 403)
		return
	}
	now := s.disposalTime()
	var p room.WorkspaceDisposal
	if req.Action == "propose" {
		job, ok := st.ManagedJobs[req.JobID]
		if !ok || !validManagedJob(job) || job.Version != workerWorkspaceLifecycleVersion || job.ShutdownConfirmed || s.active[job.ParticipantID] || req.ProposalID != "" || req.Confirm != "" {
			s.mu.Unlock()
			http.Error(w, "Stopped workspace job required", 409)
			return
		}
		p = room.WorkspaceDisposal{Version: 1, ID: id(), ActorID: ownerID, Job: job, CreatedAt: now, NotBefore: now.Add(disposalReviewDelay), ExpiresAt: now.Add(disposalLifetime), Status: "pending"}
	} else {
		var ok bool
		p, ok = st.WorkspaceDisposals[req.ProposalID]
		if !ok || p.Status != "pending" || req.JobID != "" {
			s.mu.Unlock()
			http.Error(w, "Pending proposal required", 409)
			return
		}
		if req.Action == "revoke" {
			if req.Confirm != "" {
				s.mu.Unlock()
				http.Error(w, "Unexpected confirmation", 400)
				return
			}
			p.Status = "revoked"
			err = s.persistDisposal(r.Context(), st, p)
			s.mu.Unlock()
			if err != nil {
				http.Error(w, "Revocation audit failed", 503)
				return
			}
			reply(w, p)
			return
		}
		if req.Confirm != p.Job.JobID || now.Before(s.disposalReview(p).ReviewReadyAt) || !now.Before(p.ExpiresAt) {
			s.mu.Unlock()
			http.Error(w, "Review period, expiry or exact job confirmation failed", 409)
			return
		}
	}
	s.mu.Unlock()
	// Snapshot through the approved authenticated worker transport before claiming.
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	snapshot, err := s.approvedSnapshot(ctx)
	if err != nil || snapshot.Truncated {
		http.Error(w, "Complete worker state required", 503)
		return
	}
	matched := false
	for _, view := range snapshot.Jobs {
		if view.JobID != p.Job.JobID {
			continue
		}
		if !sameWorkerBinding(p.Job, view) || !validWorkerRecord(view.workerJobRecord) || view.LeaseState != "free" || view.Pair == nil || view.Pair.Workspace == nil || view.Pair.Workspace.Removed || !pairRoleSettled(view.Pair.Engine) || !pairRoleSettled(view.Pair.Tool) {
			break
		}
		identity := view.Pair.Workspace.Identity
		if req.Action == "propose" {
			p.WorkspaceName, p.WorkspaceNonce, p.WorkspaceCreatedAt = identity.Name, identity.Nonce, identity.CreatedAt
		} else if disposalIdentity(p) != identity {
			break
		}
		matched = true
		break
	}
	if !matched || !validDisposal(p) {
		http.Error(w, "Retained workspace identity and stopped containers required", 409)
		return
	}
	s.mu.Lock()
	latest, err := s.load(r.Context())
	currentJob, found := latest.ManagedJobs[p.Job.JobID]
	if err != nil || s.integrityFault || !disposalOwner(latest, requestActor(r)) || !latest.FeatureFlags["workspaceDisposal"] || !found || !sameDisposalJob(currentJob, p.Job) || currentJob.ShutdownConfirmed || s.active[currentJob.ParticipantID] || currentJob.Worker != os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS") {
		s.mu.Unlock()
		http.Error(w, "Disposal authorization changed", 409)
		return
	}
	if req.Action == "propose" {
		// The review minute starts when the intention is published, not before a
		// potentially slow worker probe.
		p.StorageReservationID = disposalStorageReservation(latest, p)
		p.CreatedAt = s.disposalTime()
		p.NotBefore = p.CreatedAt.Add(disposalReviewDelay)
		p.ExpiresAt = p.CreatedAt.Add(disposalLifetime)
	}
	if req.Action == "approve" {
		prior, ok := latest.WorkspaceDisposals[p.ID]
		now = s.disposalTime()
		if !ok || prior != p || now.Before(s.disposalReview(p).ReviewReadyAt) || !now.Before(p.ExpiresAt) {
			s.mu.Unlock()
			http.Error(w, "Proposal changed, expired or already used", 409)
			return
		}
		p.Status = "claimed"
		p.ReviewObservedAt = s.disposalSeen[p.ID]
		p.ClaimedAt = now
	}
	err = s.persistDisposal(r.Context(), latest, p)
	var review disposalReviewView
	if err == nil && req.Action == "propose" {
		review = s.disposalReview(p)
	}
	s.mu.Unlock()
	if err != nil {
		http.Error(w, "Disposal audit failed; no worker action sent", 503)
		return
	}
	if req.Action == "propose" {
		reply(w, review)
		return
	}
	// Claim commits approval once. Disconnect/flag changes do not undo the admitted
	// irreversible action. Cleanup/outcome have their own bounded actor-bound context.
	action, cancelAction := context.WithTimeout(context.WithoutCancel(r.Context()), 40*time.Second)
	defer cancelAction()
	fence := s.jobFence
	if fence == nil {
		fence = remoteFenceJob
	}
	workerErr := callDisposalWorker(action, p.Job.Worker, disposalFence(p), fence)
	s.mu.Lock()
	defer s.mu.Unlock()
	latest, loadErr := s.load(action)
	if loadErr != nil {
		http.Error(w, "Disposal outcome unresolved; do not retry", 503)
		return
	}
	prior, ok := latest.WorkspaceDisposals[p.ID]
	if !ok || prior != p {
		http.Error(w, "Disposal claim unresolved; do not retry", 503)
		return
	}
	p.FinishedAt = s.disposalTime()
	p.Status = "succeeded"
	if workerErr != nil {
		p.Status = "unconfirmed"
	}
	if !validDisposal(p) {
		http.Error(w, "Disposal outcome time unresolved; do not retry", 503)
		return
	}
	// Keep physical outcome and controller container-slot release atomic. Storage
	// credit is released only for the immutable exact reservation reviewed at proposal.
	changes := []auditChange{{Kind: disposalEvent(p), Payload: p}}
	if workerErr == nil {
		job := latest.ManagedJobs[p.Job.JobID]
		job.Status = "shutdown_confirmed"
		job.ShutdownConfirmed = true
		job.DiagnosticCode = "workspace_disposal_confirmed"
		job.UpdatedAt = p.FinishedAt
		changes = append(changes, auditChange{Kind: "worker.job.updated", Payload: job})
		if p.StorageReservationID != "" {
			release := room.StorageRelease{Version: 1, ReservationID: p.StorageReservationID, DisposalID: p.ID}
			candidate := latest
			candidate.WorkspaceDisposals = map[string]room.WorkspaceDisposal{p.ID: p}
			candidate.ManagedJobs = map[string]room.ManagedJob{job.JobID: job}
			if !validStorageRelease(candidate, release) {
				http.Error(w, "Disposal credit unresolved; do not retry", 503)
				return
			}
			changes = append(changes, auditChange{Kind: "storage.released", Payload: release})
		}
	}
	if s.appendBatch(context.WithValue(action, expectedEventCountKey{}, latest.EventCount), changes) != nil {
		http.Error(w, "Disposal outcome audit unresolved; do not retry", 503)
		return
	}
	if workerErr != nil {
		http.Error(w, "Disposal unconfirmed; do not retry", 503)
		return
	}
	reply(w, p)
}
func (s *server) persistDisposal(ctx context.Context, st room.State, p room.WorkspaceDisposal) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	candidate := st
	// Projection owns its map; validation must not mutate the loaded snapshot.
	candidate.WorkspaceDisposals = map[string]room.WorkspaceDisposal{}
	for k, v := range st.WorkspaceDisposals {
		candidate.WorkspaceDisposals[k] = v
	}
	if err = projectWorkspaceDisposal(&candidate, disposalEvent(p), raw); err != nil {
		return err
	}
	return s.append(context.WithValue(ctx, expectedEventCountKey{}, st.EventCount), disposalEvent(p), p)
}

// A trusted adapter panic is still an unknown external outcome. Persist generic
// uncertainty rather than log a raw panic or make the claim replayable.
func callDisposalWorker(ctx context.Context, alias string, request workerFenceRequest, fence func(context.Context, string, workerFenceRequest) error) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("workspace disposal outcome unconfirmed")
		}
	}()
	return fence(ctx, alias, request)
}
