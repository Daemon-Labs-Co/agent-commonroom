package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"os"
	"strconv"
	"time"
)

// Worker-owner deployment policy only. Never read these values from model tools,
// prompts, room messages or untrusted launch arguments. Disabling this optional
// guard cannot disable hard deadlines, isolation, fencing or source retention.
type workerIdlePolicy struct {
	Enabled bool
	Minutes int
}

func readWorkerIdlePolicy(getenv func(string) string) (workerIdlePolicy, error) {
	p := workerIdlePolicy{true, 5}
	switch getenv("COMMONROOM_AGENT_IDLE_SHUTDOWN") {
	case "", "true":
	case "false":
		p.Enabled = false
	default:
		return p, errors.New("invalid idle shutdown flag")
	}
	if text := getenv("COMMONROOM_AGENT_IDLE_MINUTES"); text != "" {
		n, err := strconv.Atoi(text)
		if err != nil || n < 1 || n > 60 {
			return p, errors.New("idle minutes must be between 1 and 60")
		}
		p.Minutes = n
	}
	return p, nil
}

type workerIdleWatch struct {
	policy workerIdlePolicy
	runner toolbroker.DockerRunner
	since  time.Time
	armed  bool
}

// configureIdle snapshots trusted deployment policy before startup. Explicit
// injection is for reviewed worker assembly/tests, never model/plugin adapters.
// A source fork can omit this guard, but cannot claim automatic idle shutdown.
func (s *workerPairSupervisor) configureIdle(p workerIdlePolicy, runner toolbroker.DockerRunner) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if runner == nil || p.Minutes < 1 || p.Minutes > 60 || s.idle != nil || s.closed || s.record.Status != "starting" {
		return errors.New("idle policy binding refused")
	}
	s.idle = &workerIdleWatch{policy: p, runner: runner}
	return nil
}
func (s *workerPairSupervisor) ensureIdle(runner toolbroker.DockerRunner) error {
	s.mu.Lock()
	configured := s.idle != nil
	s.mu.Unlock()
	if configured {
		return nil
	}
	policy, err := readWorkerIdlePolicy(os.Getenv)
	if err != nil {
		return err
	}
	return s.configureIdle(policy, runner)
}

// Called under mu after successful startup. No idle timer covers preparation.
func (s *workerPairSupervisor) armIdleLocked() {
	if s.idle == nil || s.idle.armed || !s.idle.policy.Enabled {
		return
	}
	s.idle.armed = true
	s.idle.since = time.Now()
	go s.watchIdle()
}
func (s *workerPairSupervisor) watchIdle() {
	for {
		s.mu.Lock()
		delay := time.Until(s.idle.since.Add(time.Duration(s.idle.policy.Minutes) * time.Minute))
		s.mu.Unlock()
		if delay < time.Millisecond {
			delay = time.Millisecond
		}
		timer := time.NewTimer(delay)
		select {
		case <-s.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		expired, _ := s.expireIdle(time.Now())
		if expired {
			return
		}
	}
}

// expireIdle atomically wins against startup, active turns and queued turns.
// It journals intention before cancellation and uses only the existing bounded
// non-destructive Shutdown. A retained workspace keeps full job proof unresolved.
func (s *workerPairSupervisor) expireIdle(now time.Time) (bool, error) {
	if !s.turn.TryLock() {
		s.observeIdleBusy(now)
		return false, nil
	}
	if !s.startup.TryLock() {
		s.turn.Unlock()
		s.observeIdleBusy(now)
		return false, nil
	}
	s.mu.Lock()
	if s.idle == nil || !s.idle.policy.Enabled || s.closed || s.ctx.Err() != nil || s.record.Status != "running" || s.idle.since.IsZero() {
		s.mu.Unlock()
		s.startup.Unlock()
		s.turn.Unlock()
		return false, nil
	}
	if s.pendingTurns > 0 {
		s.idle.since = now
		s.mu.Unlock()
		s.startup.Unlock()
		s.turn.Unlock()
		return false, nil
	}
	if now.Before(s.idle.since.Add(time.Duration(s.idle.policy.Minutes) * time.Minute)) {
		s.mu.Unlock()
		s.startup.Unlock()
		s.turn.Unlock()
		return false, nil
	}
	record := s.record
	record.IdleShutdownRequestedAt = now.UTC().Round(0)
	record.IdleTimeoutMinutes = s.idle.policy.Minutes
	record.DiagnosticCode = "idle_shutdown_requested"
	record.UpdatedAt = time.Now().UTC()
	err := writeWorkerRecord(s.dir, record, false)
	if err == nil {
		s.record = record
	}
	s.closed = true
	s.cancel()
	runner := s.idle.runner
	s.mu.Unlock()
	s.startup.Unlock()
	s.turn.Unlock()
	if err != nil {
		s.Close()
		return true, errors.New("idle intention persistence unavailable")
	}
	_, err = s.Shutdown(context.Background(), runner)
	return true, err
}
func (s *workerPairSupervisor) observeIdleBusy(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.idle != nil && now.After(s.idle.since) {
		s.idle.since = now
	}
}
func (s *workerPairSupervisor) queueTurn() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.ctx.Err() != nil {
		return errors.New("pair supervisor unavailable")
	}
	s.pendingTurns++
	return nil
}
func (s *workerPairSupervisor) finishQueuedTurn() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingTurns--
	if s.idle != nil {
		s.idle.since = time.Now()
	}
}
