package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"log"
	"strings"
	"time"
)

type managedJobContextKey struct{}
type managedJobRegistration struct {
	JobID    string
	IssuedAt time.Time
	Deadline time.Time
	Register func(string, string) error
	Update   func(string) error
}

// managedJobContext records intent before SSH dispatch, independent of optional
// diagnostics. Audit failure prevents launch or latches execution off. It never
// stores prompts, provider output, command arguments or credentials.
func (s *server) managedJobContext(ctx context.Context, p room.Participant) context.Context {
	job := room.ManagedJob{Version: 2, JobID: id(), RoomID: p.RoomID, ParticipantID: p.ID, ConnectionID: p.ConnectionID, Status: "allocated", AllocatedAt: time.Now().UTC()}
	job.InitiatorID, _ = ctx.Value(actorContextKey{}).(string)
	job.Deadline, _ = ctx.Deadline()
	if job.Deadline.IsZero() {
		job.Deadline = job.AllocatedAt.Add(3 * time.Minute)
	}
	registered := false
	registration := managedJobRegistration{JobID: job.JobID, IssuedAt: job.AllocatedAt, Deadline: job.Deadline}
	registration.Register = func(alias, provider string) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		if registered || ctx.Err() != nil {
			return errors.New("managed job already allocated or cancelled")
		}
		latest, err := s.load(ctx)
		if err != nil {
			return errors.New("managed job state unavailable")
		}
		if s.agentsBlockedLocked(latest, p.RoomID) {
			return errors.New("managed execution is stopped")
		}
		if job.InitiatorID != "" {
			actor, ok := member(latest, job.InitiatorID)
			if !ok || actor.Kind != "human" || !canAccess(latest, job.InitiatorID, p.RoomID) {
				return errors.New("managed initiator access changed")
			}
		}
		job.CompanyID = latest.Company.ID
		found := false
		for _, member := range latest.Participants {
			if member.ID == p.ID && member.RoomID == p.RoomID && member.ConnectionID == p.ConnectionID {
				found = true
			}
		}
		if binding, ok := ctx.Value(verificationBindingKey{}).(verificationBinding); ok && binding.Participant.ID == p.ID && binding.Participant.RoomID == p.RoomID && binding.Participant.ConnectionID == p.ConnectionID {
			found = false
			for _, connection := range latest.Connections {
				if connection.ID == binding.Connection.ID && connection.Endpoint == binding.Connection.Endpoint && connection.Protocol == "CLI" && connection.Provider == binding.Connection.Provider && connection.Revision == binding.Connection.Revision {
					found = true
				}
			}
			boundAlias, boundProvider, bindingErr := remoteWorkerEndpoint(binding.Connection.Endpoint, alias, "true")
			if bindingErr != nil || boundAlias != alias || boundProvider != provider {
				found = false
			}
			if !connectionTestingEnabled(latest.FeatureFlags) {
				found = false
			}
		}
		if !found {
			return errors.New("managed membership changed before dispatch")
		}
		for _, prior := range latest.ManagedJobs {
			if prior.ParticipantID == p.ID && !prior.ShutdownConfirmed {
				return errors.New("prior worker job has unconfirmed shutdown; inspect job diagnostics")
			}
		}
		if !workerAliasPattern.MatchString(alias) || (provider != "codex" && provider != "claude") {
			return errors.New("invalid worker binding")
		}
		if err = s.admitDailyTurns(latest, time.Now(), 1); err != nil {
			return err
		}
		job.Worker = alias
		job.Provider = provider
		job.UpdatedAt = time.Now().UTC()
		job.DiagnosticCode = "intent_recorded"
		auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if job.InitiatorID != "" {
			auditCtx = context.WithValue(auditCtx, actorContextKey{}, job.InitiatorID)
		}
		if err = s.append(context.WithValue(auditCtx, expectedEventCountKey{}, latest.EventCount), "worker.job.updated", job); err != nil {
			s.tripIntegrityLocked()
			return errors.New("managed job allocation audit failed; launch refused")
		}
		registered = true
		if s.activeJobs == nil {
			s.activeJobs = map[string]string{}
		}
		s.activeJobs[p.ID] = job.JobID
		return nil
	}
	registration.Update = func(status string) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		if !registered {
			return errors.New("managed job is not allocated")
		}
		if status != "submitted" && status != "shutdown_confirmed" && status != "shutdown_unconfirmed" {
			return errors.New("invalid lifecycle transition")
		}
		return s.recordManagedJobLocked(job.JobID, status)
	}
	return context.WithValue(ctx, managedJobContextKey{}, registration)
}

func (s *server) recordManagedJobLocked(jobID, status string) error {
	auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	st, err := s.load(auditCtx)
	if err != nil {
		s.tripIntegrityLocked()
		return errors.New("job audit state unavailable")
	}
	job, ok := st.ManagedJobs[jobID]
	if !ok {
		return errors.New("unknown managed job")
	}
	if job.ShutdownConfirmed {
		return nil
	} // late uncertainty cannot reverse verified proof
	job.Status = status
	job.UpdatedAt = time.Now().UTC()
	job.DiagnosticCode = status
	job.ShutdownConfirmed = status == "shutdown_confirmed"
	if err = s.append(auditCtx, "worker.job.updated", job); err != nil {
		s.tripIntegrityLocked()
		return errors.New("job audit update failed")
	}
	if job.ShutdownConfirmed && s.activeJobs[job.ParticipantID] == jobID {
		delete(s.activeJobs, job.ParticipantID)
	}
	if status == "shutdown_unconfirmed" {
		s.stopAll = true
		for _, stop := range s.cancels {
			stop()
		}
		for _, stop := range s.workflowLive {
			stop()
		}
		if s.append(auditCtx, "agents.stop", map[string]string{"scope": "application", "reason": "Worker job shutdown is unverified; inspect job diagnostics and reconcile before explicit resume", "jobId": jobID}) != nil {
			s.tripIntegrityLocked()
		}
	}
	log.Printf("managed job %s participant %s lifecycle %s", jobID, job.ParticipantID, status)
	return nil
}

func validManagedJob(v room.ManagedJob) bool {
	if v.InitiatorID != "" && !membershipPartPattern.MatchString(v.InitiatorID) || v.CompanyID != "" && !membershipPartPattern.MatchString(v.CompanyID) {
		return false
	}
	if v.Version != 2 || !validWorkerJobID(v.JobID) || !membershipPartPattern.MatchString(v.RoomID) || !membershipPartPattern.MatchString(v.ParticipantID) || !membershipPartPattern.MatchString(v.ConnectionID) || !workerAliasPattern.MatchString(v.Worker) || (v.Provider != "codex" && v.Provider != "claude") || v.AllocatedAt.IsZero() || v.UpdatedAt.IsZero() || v.Deadline.IsZero() {
		return false
	}
	switch v.Status {
	case "allocated", "submitted", "shutdown_unconfirmed":
		return !v.ShutdownConfirmed
	case "shutdown_confirmed":
		return v.ShutdownConfirmed
	}
	return false
}
func jobsDebugEnabled(st room.State) bool {
	enabled, exists := st.FeatureFlags["jobDiagnostics"]
	return !exists || enabled
}

func managedJobFence(job room.ManagedJob) workerFenceRequest {
	return workerFenceRequest{Version: 2, JobID: job.JobID, Provider: job.Provider, Key: strings.Join([]string{job.RoomID, job.ParticipantID, job.ConnectionID, "worker"}, "/")}
}

func hasUnconfirmedManagedJobs(st room.State) bool {
	for _, job := range st.ManagedJobs {
		if !job.ShutdownConfirmed {
			return true
		}
	}
	return false
}
