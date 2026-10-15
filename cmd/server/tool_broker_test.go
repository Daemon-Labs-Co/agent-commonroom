package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNewToolDispatcherRequiresWorkerAndAuditChannel(t *testing.T) {
	t.Parallel()
	scope := sampleToolScope(t)
	image := "ghcr.io/example/commonroom-tool@sha256:" + strings.Repeat("a", 64)
	for _, mode := range []string{"missing worker", "wrong job", "missing channel", "cancelled", "valid"} {
		t.Run(mode, func(t *testing.T) {
			client, peer := net.Pipe()
			defer client.Close()
			defer peer.Close()
			var channel io.ReadWriteCloser = client
			ctx := context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID)
			switch mode {
			case "missing worker":
				ctx = t.Context()
			case "wrong job":
				ctx = context.WithValue(ctx, workerLocalJobContextKey{}, strings.Repeat("c", 32))
			case "missing channel":
				channel = nil
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			dispatcher, token, cleanup, err := newToolDispatcher(ctx, scope, []string{"read"}, image, "/var/lib/docker", time.Now(), channel)
			if cleanup != nil {
				defer cleanup()
			}
			if mode == "valid" {
				if err != nil || dispatcher == nil || len(token) != 64 || cleanup == nil {
					t.Fatal("assembly failed", err)
				}
			} else if err == nil || dispatcher != nil || token != "" || cleanup != nil {
				t.Fatal("unbound assembly accepted")
			}
		})
	}
}

type fakeToolExecutor struct{ fail bool }

func (f fakeToolExecutor) Execute(ctx context.Context, scope toolbroker.Scope, action toolbroker.Action, out io.Writer) error {
	if f.fail {
		return errors.New("synthetic executor failure")
	}
	_, err := out.Write([]byte(`{"ok":true}`))
	return err
}

func TestDispatchRoundTripsThroughScopedControllerChannel(t *testing.T) {
	t.Parallel()
	for _, failed := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "executor failure"}[failed], func(t *testing.T) {
			t.Parallel()
			s, _, auditor, intent := scopedAuditFixture(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			client, peer := net.Pipe()
			defer client.Close()
			done := make(chan error, 1)
			go func() { done <- toolbroker.ServeAudit(ctx, peer, intent.Scope, auditor) }()
			audit, err := toolbroker.NewAuditClient(client, intent.Scope)
			if err != nil {
				t.Fatal(err)
			}
			authority, token, err := toolbroker.New(intent.Scope, []string{"read"}, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			dispatcher, err := toolbroker.NewDispatcher(authority, audit, fakeToolExecutor{fail: failed})
			if err != nil {
				t.Fatal(err)
			}
			req := toolbroker.Request{Scope: intent.Scope, Capability: token, CallID: "roundtrip", Operation: "read", Path: "notes.txt"}
			_, dispatchErr := dispatcher.Dispatch(ctx, req)
			if failed != (dispatchErr != nil) {
				t.Fatal("unexpected dispatch result", dispatchErr)
			}
			st, err := s.load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			want := "succeeded"
			if failed {
				want = "failed"
			}
			got := st.ToolAudits[toolAuditKey(intent.Scope.JobID, req.CallID)]
			if got.Status != want || got.ActorID != auditor.actorID {
				t.Fatal("scoped result/attribution lost", got)
			}
			if len(st.ToolCalls) != 0 {
				t.Fatal("new action used legacy unscoped projection")
			}
			cancel()
			client.Close()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("audit stream did not close")
			}
		})
	}
}

// Credential separation and durable audit are mandatory, not rollout toggles.
func TestToolSeparationFlagCannotGrantOrDisableAudit(t *testing.T) {
	t.Parallel()
	for _, enabled := range []string{"true", "false"} {
		t.Run(enabled, func(t *testing.T) {
			_, db, h := mockServer(t)
			before := len(db.snapshot())
			response := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"toolSeparation","enabled":`+enabled+`}}`)
			if response.Code != 400 || len(db.snapshot()) != before {
				t.Fatal("mandatory audit exposed as product toggle", response.Code)
			}
		})
	}
}

func TestWorkerCancellationClosesAuditDespiteFreshDispatchContext(t *testing.T) {
	t.Parallel()
	scope := sampleToolScope(t)
	worker, cancel := context.WithCancel(context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID))
	defer cancel()
	client, peer := net.Pipe()
	defer peer.Close()
	image := "ghcr.io/example/commonroom-tool@sha256:" + strings.Repeat("a", 64)
	dispatcher, token, cleanup, err := newToolDispatcher(worker, scope, []string{"read"}, image, "/var/lib/docker", time.Now(), client)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	entered := make(chan struct{})
	go func() { var first [1]byte; _, _ = peer.Read(first[:]); close(entered) }()
	result := make(chan error, 1)
	go func() {
		_, err := dispatcher.Dispatch(context.Background(), toolbroker.Request{Scope: scope, Capability: token, CallID: "cancelled", Operation: "read", Path: "file"})
		result <- err
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("audit never started")
	}
	cancel()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("cancelled worker admitted action")
		}
	case <-time.After(time.Second):
		t.Fatal("worker cancellation failed to interrupt audit")
	}
	// Replacing the Dispatch context cannot resurrect the revoked authority.
	if _, err := dispatcher.Dispatch(context.Background(), toolbroker.Request{Scope: scope, Capability: token, CallID: "again", Operation: "read", Path: "file"}); err == nil {
		t.Fatal("cancelled worker authority survived")
	}
	cleanup()
	cleanup()
}
