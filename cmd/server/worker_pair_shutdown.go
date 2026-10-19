package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"strings"
	"time"
)

// Shutdown is the supervisor's bounded stop-and-reconcile path. It cancels first,
// waits for in-flight startup to unwind before releasing creator ownership, then
// uses the existing durable tombstone/lease/identity fences to remove both roles.
// The caller supplies a fresh cleanup context, not the cancelled model lifetime.
// No default runner, direct host execution, restart or workspace discard authority
// exists here. A source fork CAN skip this method, BUT cannot claim shutdown proof.
// A misbehaving launch adapter that ignores cancellation keeps Close waiting;
// never release its lease early or turn a timeout into absence evidence.
func (s *workerPairSupervisor) Shutdown(ctx context.Context, runner toolbroker.DockerRunner) (workerJobRecord, error) {
	s.reconcile.Lock()
	defer s.reconcile.Unlock()
	s.Close()
	s.mu.Lock()
	original := s.record
	listener := s.auditListener
	dir := s.dir
	s.mu.Unlock()
	if listener != nil {
		listener.Close()
	}
	if ctx == nil || ctx.Err() != nil || runner == nil || !validWorkerRecord(original) || original.Pair == nil {
		return original, errors.New("pair cleanup context or binding unavailable")
	}
	cleanup, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	request := workerFenceRequest{Version: original.Version, JobID: original.JobID, Provider: original.Provider, Repository: original.Pair.Repository, Key: strings.Join([]string{original.RoomID, original.ParticipantID, original.ConnectionID, "worker"}, "/")}
	// Deliberately do not populate DiscardWorkspace: cancellation and normal exit
	// do not authorize destructive workspace deletion. Present workspaces remain
	// unresolved until the owner explicitly approves their exact disposal.
	result, err := fenceWorkerJobContainers(cleanup, dir, request, func(ctx context.Context, record workerJobRecord) (workerJobRecord, error) {
		return stopWorkerPair(ctx, record, runner)
	})
	if !validWorkerRecord(result) || !sameWorkerFenceBinding(result, original) {
		return original, errors.New("pair shutdown evidence unavailable")
	}
	s.mu.Lock()
	s.record = result
	credential := s.engineCredential
	s.mu.Unlock()
	if credential != nil {
		releaseErr := credential.Release(result)
		err = errors.Join(err, releaseErr)
		credential.mu.Lock()
		released := credential.released
		credential.mu.Unlock()
		if released {
			s.mu.Lock()
			if s.engineCredential == credential {
				s.engineCredential = nil
			}
			s.mu.Unlock()
		}
	}
	return result, err
}

// Run owns one admitted isolated model turn and always reconciles after startup
// or turn failure, cancellation, normal return, or panic. turn is trusted worker
// driver code for the reviewed engine, never a model-selected host function. It
// MUST honor its context and return before calling Close/Shutdown; those methods
// wait for the turn to unwind. This method provides no host execution fallback.
// No result is a successful lifecycle result unless fenced cleanup also succeeds.
// Present retained workspaces deliberately keep that proof unresolved.
func (s *workerPairSupervisor) Run(ctx context.Context, adapter workerPairLaunchAdapter, runner toolbroker.DockerRunner, turn func(context.Context) error) (record workerJobRecord, err error) {
	defer func() {
		var shutdownErr error
		// Cleanup must not inherit the cancelled model, request or audit lifetime.
		record, shutdownErr = s.Shutdown(context.Background(), runner)
		err = errors.Join(err, shutdownErr)
	}()
	if ctx == nil || ctx.Err() != nil || adapter == nil || runner == nil || turn == nil {
		return record, errors.New("paired turn dependencies unavailable")
	}
	s.turn.Lock()
	defer s.turn.Unlock()
	s.mu.Lock()
	deadline := s.record.Deadline
	s.mu.Unlock()
	lifetime, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	stop := context.AfterFunc(s.ctx, cancel)
	defer stop()
	if err = s.Start(lifetime, adapter); err != nil {
		return record, err
	}
	err = turn(lifetime)
	if lifetime.Err() != nil {
		err = errors.Join(err, lifetime.Err())
	}
	if s.ctx.Err() != nil {
		err = errors.Join(err, s.ctx.Err())
	}
	return record, err
}
