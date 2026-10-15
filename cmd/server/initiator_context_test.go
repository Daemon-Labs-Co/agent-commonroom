package main

import (
	"agentroom/internal/room"
	"context"
	"testing"
	"time"
)

func TestDetachedInitiatorCopiesOnlyAuthenticatedIdentity(t *testing.T) {
	t.Parallel()
	type requestSecretKey struct{}
	request, cancel := context.WithCancel(context.WithValue(context.Background(), actorContextKey{}, "human"))
	request = context.WithValue(request, requestSecretKey{}, "do-not-copy")
	detached := detachedInitiatorContext(request)
	cancel()
	if detached.Err() != nil || detached.Value(actorContextKey{}) != "human" || detached.Value(requestSecretKey{}) != nil {
		t.Fatal("incorrect detached authority/lifetime")
	}
	inherited := context.WithValue(context.Background(), actorContextKey{}, ownerID)
	if withInitiatingActor(inherited, "").Value(actorContextKey{}) != "" {
		t.Fatal("missing recorded initiator inherited owner")
	}
}

func TestTaskDispatchRetainsInitiatorInRegisteredJob(t *testing.T) {
	t.Parallel()
	s, _, h, _ := dispatchFixture(t)
	jobs := make(chan room.ManagedJob, 1)
	s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
		bound := s.managedJobContext(ctx, p)
		registration := bound.Value(managedJobContextKey{}).(managedJobRegistration)
		if err := registration.Register("worker", "codex"); err != nil {
			return "", err
		}
		s.mu.Lock()
		st, err := s.load(context.Background())
		s.mu.Unlock()
		if err != nil {
			return "", err
		}
		jobs <- st.ManagedJobs[registration.JobID]
		return "Synthetic completed result", nil
	}
	w := dispatchRequest(h, "initiator-check", dispatchBody)
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case job := <-jobs:
		if job.InitiatorID != ownerID || job.CompanyID == "" {
			t.Fatal("lost request principal", job)
		}
	case <-time.After(time.Second):
		t.Fatal("task did not register job")
	}
	waitDispatch(t, s)
}

func TestWorkflowRetainsRecordedInitiatorInRegisteredJobs(t *testing.T) {
	t.Parallel()
	s, _, h := executionServer(t)
	jobs := make(chan room.ManagedJob, 2)
	s.workflowTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
		bound := s.managedJobContext(ctx, p)
		registration := bound.Value(managedJobContextKey{}).(managedJobRegistration)
		if err := registration.Register("worker", "codex"); err != nil {
			return "", err
		}
		s.mu.Lock()
		st, err := s.load(context.Background())
		s.mu.Unlock()
		if err != nil {
			return "", err
		}
		jobs <- st.ManagedJobs[registration.JobID]
		if p.ID == "qa" {
			return `{"outcome":"passed","summary":"Synthetic review evidence"}`, nil
		}
		return "Synthetic implementation", nil
	}
	run := startExecution(t, h)
	if run.ActorID != ownerID {
		t.Fatal("workflow lost authenticated initiator")
	}
	awaitExecution(t, s, "awaiting_human")
	for i := 0; i < 2; i++ {
		select {
		case job := <-jobs:
			if job.InitiatorID != run.ActorID || job.CompanyID == "" {
				t.Fatal("workflow job lost principal", job)
			}
		case <-time.After(time.Second):
			t.Fatal("workflow did not register job")
		}
	}
}

func TestDetachedChatContextRegistersInitiatorWithoutRequestLifetime(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	p := room.Participant{ID: "chat-agent", RoomID: "engineering", ConnectionID: "chat-connection"}
	db.seed("participant.created", p)
	request, cancel := context.WithCancel(context.WithValue(context.Background(), actorContextKey{}, ownerID))
	ctx := detachedInitiatorContext(request)
	cancel()
	bound := s.managedJobContext(ctx, p)
	registration := bound.Value(managedJobContextKey{}).(managedJobRegistration)
	if err := registration.Register("worker", "codex"); err != nil {
		t.Fatal(err)
	}
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.ManagedJobs[registration.JobID].InitiatorID != ownerID {
		t.Fatal("chat context lost actor")
	}
}
