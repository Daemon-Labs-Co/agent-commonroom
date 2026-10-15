package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type managedJobContextKey struct{}
type managedJobRegistration struct {
	Version    int
	Repository string
	JobID      string
	IssuedAt   time.Time
	Deadline   time.Time
	Register   func(string, string) error
	Update     func(string) error
}

// managedJobContext records intent before SSH dispatch, independent of optional
// diagnostics. Audit failure prevents launch or latches execution off. It never
// stores prompts, provider output, command arguments or credentials.
func (s *server) managedJobContext(ctx context.Context, p room.Participant) context.Context {
	return s.managedExecutionJobContext(ctx, p, "", nil)
}

// managedPairJobContext journals identity only. It does not enable launch or
// bypass the mandatory development barrier. Repository policy is rechecked at
// registration; absent human provenance cannot become owner authority.
func (s *server) managedPairJobContext(ctx context.Context, p room.Participant, repository string) context.Context {
	return s.managedExecutionJobContext(ctx, p, repository, func() (string, error) {
		return controllerDevelopmentRepository(p.RoomID, os.Getenv("COMMONROOM_DEVELOPMENT_WORKSPACES"), os.Getenv("COMMONROOM_DEVELOPMENT_REPOSITORIES"))
	})
}

func (s *server) managedExecutionJobContext(ctx context.Context, p room.Participant, repository string, policy func() (string, error)) context.Context {
	job := room.ManagedJob{Version: 2, JobID: id(), RoomID: p.RoomID, ParticipantID: p.ID, ConnectionID: p.ConnectionID, Status: "allocated", AllocatedAt: time.Now().UTC()}
	job.InitiatorID, _ = ctx.Value(actorContextKey{}).(string)
	paired := policy != nil || repository != ""
	if paired {
		job.Version = workerPairLifecycleVersion
		job.Execution = workerPairExecution
		job.RepositoryDigest = repository
	}
	job.Deadline, _ = ctx.Deadline()
	if job.Deadline.IsZero() {
		job.Deadline = job.AllocatedAt.Add(3 * time.Minute)
	}
	registered := false
	registration := managedJobRegistration{Version: job.Version, Repository: repository, JobID: job.JobID, IssuedAt: job.AllocatedAt, Deadline: job.Deadline}
	registration.Register = func(alias, provider string) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		if registered || ctx.Err() != nil {
			return errors.New("managed job already allocated or cancelled")
		}
		if paired {
			if job.InitiatorID == "" || !repositoryDigestPattern.MatchString(repository) || policy == nil {
				return errors.New("paired job requires approved repository and human provenance")
			}
			approved, err := policy()
			if err != nil || approved != repository {
				return errors.New("paired repository policy changed")
			}
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
		if !validManagedJob(job) {
			return errors.New("invalid managed job registration")
		}
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
	if job.InitiatorID != "" {
		auditCtx = withInitiatingActor(auditCtx, job.InitiatorID)
	}
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
	switch v.Version {
	case workerLifecycleVersion:
		if v.Execution != "" || v.RepositoryDigest != "" {
			return false
		}
	case workerPairLifecycleVersion:
		if v.Execution != workerPairExecution || !repositoryDigestPattern.MatchString(v.RepositoryDigest) || v.InitiatorID == "" || v.CompanyID == "" || !v.Deadline.After(v.AllocatedAt) || v.Deadline.Sub(v.AllocatedAt) > 3*time.Minute {
			return false
		}
	default:
		return false
	}
	if v.InitiatorID != "" && !membershipPartPattern.MatchString(v.InitiatorID) || v.CompanyID != "" && !membershipPartPattern.MatchString(v.CompanyID) {
		return false
	}
	if !validWorkerJobID(v.JobID) || !membershipPartPattern.MatchString(v.RoomID) || !membershipPartPattern.MatchString(v.ParticipantID) || !membershipPartPattern.MatchString(v.ConnectionID) || !workerAliasPattern.MatchString(v.Worker) || (v.Provider != "codex" && v.Provider != "claude") || v.AllocatedAt.IsZero() || v.UpdatedAt.IsZero() || v.Deadline.IsZero() {
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
	return workerFenceRequest{Version: job.Version, Repository: job.RepositoryDigest, JobID: job.JobID, Provider: job.Provider, Key: strings.Join([]string{job.RoomID, job.ParticipantID, job.ConnectionID, "worker"}, "/")}
}

func hasUnconfirmedManagedJobs(st room.State) bool {
	for _, job := range st.ManagedJobs {
		if !job.ShutdownConfirmed {
			return true
		}
	}
	return false
}

// managedWorkerTransport is the internal transport seam. Production SSH and
// synthetic tests vary here; durable lifecycle ordering stays in this module.
// No adapter may bypass isolated-worker admission or enable paired startup.
type managedWorkerTransport struct {
	command func(context.Context, string, string, string) *exec.Cmd
	fence   func(context.Context, string, workerFenceRequest) error
	monitor func(context.Context, <-chan struct{}, string, string, string)
}

// dispatchManagedWorker owns intent, submission and terminal shutdown evidence.
// Callers supply a validated worker binding and invoke the returned cleanup
// after execution, including command-start failure. Cleanup outlives turn
// cancellation but has its own bounded lifetime and is safe to call repeatedly.
func dispatchManagedWorker(ctx context.Context, alias string, request workerRunRequest, transport managedWorkerTransport) (*exec.Cmd, func(), error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, nil, errors.New("Managed worker context is unavailable or cancelled")
	}
	registration, ok := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	if !ok || registration.Version != workerLifecycleVersion || request.Version != workerLifecycleVersion || !validWorkerJobID(registration.JobID) || registration.Register == nil || registration.Update == nil {
		return nil, nil, errors.New("Durable managed job registration is required")
	}
	if transport.command == nil || transport.fence == nil || transport.monitor == nil {
		return nil, nil, errors.New("Managed worker transport is incomplete")
	}
	request.JobID, request.IssuedAt, request.Deadline = registration.JobID, registration.IssuedAt, registration.Deadline
	encoded, err := encodeWorkerRequest(request)
	if err != nil {
		return nil, nil, err
	}
	if err = registration.Register(alias, request.Provider); err != nil {
		return nil, nil, err
	}
	var once sync.Once
	monitorDone := make(chan struct{})
	cleanup := func() {
		once.Do(func() {
			close(monitorDone)
			cancelCtx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			fence := workerFenceRequest{Version: request.Version, JobID: request.JobID, Provider: request.Provider, Key: request.Key}
			if transport.fence(cancelCtx, alias, fence) != nil {
				_ = registration.Update("shutdown_unconfirmed")
				emitBackground(ctx, "shutdown_unconfirmed", request.JobID)
				fmt.Fprintln(os.Stderr, "Remote worker cancellation unconfirmed:", request.JobID)
				return
			}
			if registration.Update("shutdown_confirmed") != nil {
				emitBackground(ctx, "shutdown_unconfirmed", request.JobID)
				return
			}
			emitBackground(ctx, "shutdown_confirmed", request.JobID)
		})
	}
	if err = registration.Update("submitted"); err != nil {
		cleanup()
		return nil, nil, err
	}
	if _, ok := ctx.Value(backgroundCallbackKey{}).(backgroundCallback); ok {
		go transport.monitor(ctx, monitorDone, alias, request.JobID, request.Provider)
	}
	cmd := transport.command(ctx, alias, "run-v2", encoded)
	if cmd == nil {
		cleanup()
		return nil, nil, errors.New("Managed worker command is unavailable")
	}
	cmd.WaitDelay = 2 * time.Second
	return cmd, cleanup, nil
}

// reconcileManagedJob keeps transport proof and its durable lifecycle outcome
// together. Authorization and approved-worker checks precede this seam; worker
// evidence remains independently verified by the transport adapter.
func (s *server) reconcileManagedJob(ctx context.Context, job room.ManagedJob) (shutdownErr, auditErr error) {
	fence := s.jobFence
	if fence == nil {
		fence = remoteFenceJob
	}
	shutdownErr = fence(ctx, job.Worker, managedJobFence(job))
	s.mu.Lock()
	defer s.mu.Unlock()
	status := "shutdown_confirmed"
	if shutdownErr != nil {
		status = "shutdown_unconfirmed"
	}
	auditErr = s.recordManagedJobLocked(job.JobID, status)
	return
}
