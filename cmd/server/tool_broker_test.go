package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"agentroom/internal/toolbroker"
)

func TestToolSeparationEnabledDefaultsOff(t *testing.T) {
	t.Parallel()
	st := room.State{}
	if toolSeparationEnabled(st) {
		t.Fatal("expected default off")
	}
	st.FeatureFlags = map[string]bool{"toolSeparation": true}
	if !toolSeparationEnabled(st) {
		t.Fatal("expected enabled once set")
	}
}

func TestNewToolDispatcherAssemblesWithoutTouchingDocker(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := &server{db: db}
	now := time.Now()
	scope := sampleToolScope(t)
	image := "ghcr.io/example/commonroom-tool@sha256:" + strings.Repeat("a", 64)
	// NewContainerExecutor only validates inputs; it does not invoke docker
	// until Dispatch actually calls Execute. Construction must stay real-Docker-free.
	dispatcher, token, err := s.newToolDispatcher(scope, []string{"read"}, image, "/var/lib/docker", now)
	if err != nil {
		t.Fatal(err)
	}
	if dispatcher == nil {
		t.Fatal("expected a non-nil dispatcher")
	}
	if len(token) != 64 {
		t.Fatalf("expected a 64-hex-char capability token, got %q", token)
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

// This proves the real Auditor round-trips intent -> outcome through an
// actual Dispatcher.Dispatch call (not just called directly, as in
// tool_broker_audit_test.go), using toolbroker's own Executor interface with a
// trusted in-test fake standing in for real Docker/container execution —
// exactly as the plan called for, and consistent with how
// internal/toolbroker's own tests stub DockerRunner rather than shelling out.
func TestDispatchRoundTripsThroughRealAuditor(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name           string
		fail           bool
		wantErr        bool
		wantFinalState string
	}{
		{"success", false, false, "succeeded"},
		{"executor failure", true, true, "failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := newMockDB(t)
			s := &server{db: db}
			now := time.Now()
			scope := sampleToolScope(t)
			authority, token, err := toolbroker.New(scope, []string{"read"}, now)
			if err != nil {
				t.Fatal(err)
			}
			dispatcher, err := toolbroker.NewDispatcher(authority, &storedToolAuditor{s: s}, fakeToolExecutor{fail: tc.fail})
			if err != nil {
				t.Fatal(err)
			}
			req := toolbroker.Request{Scope: scope, Capability: token, CallID: "call-dispatch", Operation: "read", Path: "notes.txt"}
			_, dispatchErr := dispatcher.Dispatch(t.Context(), req)
			if tc.wantErr != (dispatchErr != nil) {
				t.Fatalf("Dispatch error = %v, wantErr = %v", dispatchErr, tc.wantErr)
			}
			st, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			got, ok := st.ToolCalls[req.CallID]
			if !ok {
				t.Fatal("expected the call's audit record to be replayed")
			}
			if got.Status != tc.wantFinalState {
				t.Fatalf("Status = %q, want %q", got.Status, tc.wantFinalState)
			}
		})
	}
}
