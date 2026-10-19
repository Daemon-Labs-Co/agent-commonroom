package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const disposalReviewDelay = time.Minute
const disposalLifetime = time.Hour

var errDisposal = errors.New("workspace disposal binding or transition invalid")

func disposalIdentity(p room.WorkspaceDisposal) toolbroker.WorkspaceVolumeIdentity {
	return toolbroker.WorkspaceVolumeIdentity{Version: 1, Name: p.WorkspaceName, Nonce: p.WorkspaceNonce, CreatedAt: p.WorkspaceCreatedAt}
}
func (s *server) disposalTime() time.Time {
	if s.disposalNow != nil {
		return s.disposalNow().UTC().Round(0)
	}
	return time.Now().UTC().Round(0)
}
func sameDisposalJob(a, b room.ManagedJob) bool {
	a.Status, a.DiagnosticCode, a.UpdatedAt, a.ShutdownConfirmed = b.Status, b.DiagnosticCode, b.UpdatedAt, b.ShutdownConfirmed
	return a == b
}
func validDisposal(p room.WorkspaceDisposal) bool {
	if p.Version != 1 || !validWorkerJobID(p.ID) || p.ActorID != ownerID || !validManagedJob(p.Job) || p.Job.Version != workerWorkspaceLifecycleVersion || p.Job.ShutdownConfirmed || !toolbroker.ValidWorkspaceVolume(p.Job.JobID, disposalIdentity(p), true) || p.CreatedAt.IsZero() || !p.NotBefore.Equal(p.CreatedAt.Add(disposalReviewDelay)) || !p.ExpiresAt.Equal(p.CreatedAt.Add(disposalLifetime)) {
		return false
	}
	switch p.Status {
	case "pending", "revoked":
		return p.ReviewObservedAt.IsZero() && p.ClaimedAt.IsZero() && p.FinishedAt.IsZero()
	case "claimed", "succeeded", "unconfirmed":
		if p.ReviewObservedAt.Before(p.CreatedAt) || p.ClaimedAt.Before(p.ReviewObservedAt.Add(disposalReviewDelay)) || p.ClaimedAt.Before(p.NotBefore) || !p.ClaimedAt.Before(p.ExpiresAt) {
			return false
		}
		if p.Status == "claimed" {
			return p.FinishedAt.IsZero()
		}
		return !p.FinishedAt.IsZero() && !p.FinishedAt.Before(p.ClaimedAt)
	}
	return false
}
func projectWorkspaceDisposal(st *room.State, kind string, raw []byte) error {
	var p room.WorkspaceDisposal
	if json.Unmarshal(raw, &p) != nil || !validDisposal(p) || p.Job.CompanyID != st.Company.ID || kind != "workspace.disposal."+map[string]string{"pending": "proposed", "claimed": "claimed", "revoked": "revoked", "succeeded": "succeeded", "unconfirmed": "unconfirmed"}[p.Status] {
		return errDisposal
	}
	if p.StorageReservationID != "" {
		if !validWorkerJobID(p.StorageReservationID) || st.ManagedStorage == nil {
			return errDisposal
		}
		r, ok := st.ManagedStorage.Reservations[p.StorageReservationID]
		if !ok || !reservationMatchesDisposal(r, p) {
			return errDisposal
		}
		if _, released := st.ManagedStorage.Releases[r.ID]; released {
			return errDisposal
		}
	}
	prior, exists := st.WorkspaceDisposals[p.ID]
	if !exists {
		job, ok := st.ManagedJobs[p.Job.JobID]
		if p.Status != "pending" || !st.FeatureFlags["workspaceDisposal"] || !ok || job != p.Job {
			return errDisposal
		}
		for _, other := range st.WorkspaceDisposals {
			if other.Job.JobID == p.Job.JobID && (other.Status == "claimed" || other.Status == "unconfirmed" || other.Status == "pending" && p.CreatedAt.Before(other.ExpiresAt)) {
				return errDisposal
			}
		}
	} else {
		immutable := p
		immutable.Status, immutable.ReviewObservedAt, immutable.ClaimedAt, immutable.FinishedAt = prior.Status, prior.ReviewObservedAt, prior.ClaimedAt, prior.FinishedAt
		if immutable != prior || prior.Status == "claimed" && p.ReviewObservedAt != prior.ReviewObservedAt {
			return errDisposal
		}
		if !((prior.Status == "pending" && (p.Status == "claimed" && st.FeatureFlags["workspaceDisposal"] || p.Status == "revoked")) || (prior.Status == "claimed" && (p.Status == "succeeded" || p.Status == "unconfirmed"))) {
			return errDisposal
		}
	}
	if st.WorkspaceDisposals == nil {
		st.WorkspaceDisposals = map[string]room.WorkspaceDisposal{}
	}
	st.WorkspaceDisposals[p.ID] = p
	return nil
}
func disposalFence(p room.WorkspaceDisposal) workerFenceRequest {
	identity := disposalIdentity(p)
	request := managedJobFence(p.Job)
	request.DiscardWorkspace = true
	request.ExpectedWorkspace = &identity
	return request
}
func disposalEvent(p room.WorkspaceDisposal) string {
	if p.Status == "pending" {
		return "workspace.disposal.proposed"
	}
	return "workspace.disposal." + strings.TrimSpace(p.Status)
}

// These views are assembled under s.mu. A controller waits a minute from its own
// first observation of COMMITTED state, not an in-transaction timestamp. Restart
// or another controller conservatively restarts that local review barrier.
type disposalReviewView struct {
	room.WorkspaceDisposal
	ReviewReadyAt time.Time `json:"reviewReadyAt"`
}

func (s *server) disposalReview(p room.WorkspaceDisposal) disposalReviewView {
	ready := p.NotBefore
	if p.Status == "pending" {
		if s.disposalSeen == nil {
			s.disposalSeen = map[string]time.Time{}
		}
		seen, ok := s.disposalSeen[p.ID]
		if !ok {
			seen = s.disposalTime()
			if seen.Before(p.CreatedAt) {
				seen = p.CreatedAt
			}
			s.disposalSeen[p.ID] = seen
		}
		if observed := seen.Add(disposalReviewDelay); observed.After(ready) {
			ready = observed
		}
	} else {
		delete(s.disposalSeen, p.ID)
	}
	return disposalReviewView{p, ready}
}
func (s *server) disposalReviews(st room.State) []disposalReviewView {
	views := []disposalReviewView{}
	for _, p := range disposalList(st) {
		views = append(views, s.disposalReview(p))
	}
	return views
}
