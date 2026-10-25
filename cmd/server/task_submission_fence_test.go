package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestManagedTaskSubmissionRechecksAuthorityAfterAllocation(t *testing.T) {
	for _, scenario := range []string{"healthy", "definitions off", "definition cycle", "reconciled", "role changed", "cancelled"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			s, db, h, task := structuredTaskFixture(t)
			result := make(chan error, 1)
			s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
				workCtx, cancel := context.WithCancel(ctx)
				defer cancel()
				repo := strings.Repeat("b", 64)
				jobCtx := s.managedExecutionJobContext(workCtx, p, repo, func() (string, error) { return repo, nil })
				reg := jobCtx.Value(managedJobContextKey{}).(managedJobRegistration)
				if err := reg.Register("synthetic-worker", "codex"); err != nil {
					result <- err
					return "", err
				}
				switch scenario {
				case "definitions off":
					db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
				case "definition cycle":
					for _, enabled := range []bool{false, true} {
						db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": enabled})
					}
				case "reconciled", "role changed":
					st, err := s.load(t.Context())
					if err != nil {
						result <- err
						return "", err
					}
					changed := st.Tasks[0]
					if scenario == "reconciled" {
						changed.DispatchStatus = "failed"
					} else {
						changed.Assignee = "different"
					}
					db.seed("task.revised", changed)
				case "cancelled":
					cancel()
				}
				err := reg.Update("submitted")
				if scenario == "healthy" {
					if err != nil {
						result <- err
						return "", err
					}
				} else if err == nil {
					result <- errors.New("revoked allocation submitted")
					return "", errors.New("unexpected submission")
				}
				// Cleanup remains permitted even when the original execution context or task
				// authority was revoked. Refusing launch must never hide retained workers.
				if cleanupErr := reg.Update("shutdown_confirmed"); cleanupErr != nil {
					result <- cleanupErr
					return "", cleanupErr
				}
				result <- nil
				if err != nil {
					return "", err
				}
				return `{"result":"synthetic healthy output"}`, nil
			}
			if w := dispatchRequest(h, "submission-fence-"+strings.ReplaceAll(scenario, " ", "-"), definedDispatchBody(task)); w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			select {
			case err := <-result:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("mock turn did not finish")
			}
			st := waitDispatch(t, s)
			if len(st.ManagedJobs) != 1 {
				t.Fatal("missing allocation")
			}
			for _, job := range st.ManagedJobs {
				if !job.ShutdownConfirmed {
					t.Fatal("cleanup refused", job)
				}
			}
			if scenario != "healthy" {
				for _, execution := range st.TaskExecutions {
					if execution.Status == "validated" || len(execution.Outputs) > 0 {
						t.Fatal("revoked turn delivered outputs", execution)
					}
				}
			}
		})
	}
}

func TestManagedDependencySubmissionRechecksSourceAfterAllocation(t *testing.T) {
	for _, scenario := range []string{"dependency off", "dependency cycle", "source revised"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			s, db, h, initial, target, binding := valueDependencyFixture(t)
			result := make(chan error, 1)
			s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
				repo := strings.Repeat("b", 64)
				jobCtx := s.managedExecutionJobContext(ctx, p, repo, func() (string, error) { return repo, nil })
				reg := jobCtx.Value(managedJobContextKey{}).(managedJobRegistration)
				if err := reg.Register("synthetic-worker", "codex"); err != nil {
					result <- err
					return "", err
				}
				if scenario == "source revised" {
					source := initial.Tasks[0]
					source.Status = "Changes requested"
					source.Revision++
					db.seed("task.revised", source)
				} else {
					db.seed("feature.flag.updated", map[string]any{"name": "taskValueDependencies", "enabled": false})
					if scenario == "dependency cycle" {
						db.seed("feature.flag.updated", map[string]any{"name": "taskValueDependencies", "enabled": true})
					}
				}
				if err := reg.Update("submitted"); err == nil {
					result <- errors.New("stale source allocation submitted")
					return "", errors.New("unexpected submission")
				}
				if err := reg.Update("shutdown_confirmed"); err != nil {
					result <- err
					return "", err
				}
				result <- nil
				return "", errors.New("stale input refused")
			}
			if w := dispatchRequest(h, "dependency-submission-"+strings.ReplaceAll(scenario, " ", "-"), valueDispatchBody(t, target, binding)); w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			select {
			case err := <-result:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("mock dependency turn did not finish")
			}
			st := waitDispatch(t, s)
			for _, e := range st.TaskExecutions {
				if e.TaskID == target.ID && (e.Status == "validated" || len(e.Outputs) > 0) {
					t.Fatal("stale input output released", e)
				}
			}
			for _, job := range st.ManagedJobs {
				if !job.ShutdownConfirmed {
					t.Fatal("cleanup refused")
				}
			}
		})
	}
}

func TestManagedSubmissionCannotReuseConsumedOrTerminalAllocation(t *testing.T) {
	for _, terminal := range []string{"submitted", "shutdown_confirmed", "shutdown_unconfirmed"} {
		t.Run(terminal, func(t *testing.T) {
			t.Parallel()
			s, db, _, reg := syntheticManagedJob(t)
			if err := reg.Update("submitted"); err != nil {
				t.Fatal(err)
			}
			if terminal != "submitted" {
				if err := reg.Update(terminal); err != nil {
					t.Fatal(err)
				}
			}
			before := len(db.snapshot())
			if reg.Update("submitted") == nil {
				t.Fatal("consumed allocation reused")
			}
			if len(db.snapshot()) != before {
				t.Fatal("refusal appended submission")
			}
			if err := reg.Update("shutdown_confirmed"); err != nil {
				t.Fatal("cleanup refused", err)
			}
			st, err := s.load(t.Context())
			if err != nil || !st.ManagedJobs[reg.JobID].ShutdownConfirmed {
				t.Fatal("shutdown evidence lost", err)
			}
		})
	}
}
