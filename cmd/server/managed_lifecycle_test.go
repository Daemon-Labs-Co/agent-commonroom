package main

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

func TestManagedDispatchOwnsLifecycle(t *testing.T) {
	for _, mode := range []string{"success", "registration-failed", "submission-failed", "shutdown-failed", "audit-failed", "command-unavailable"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			var events []string
			now := time.Now().UTC()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			reg := managedJobRegistration{Version: 2, JobID: "0123456789abcdef0123456789abcdef", IssuedAt: now, Deadline: now.Add(time.Minute)}
			reg.Register = func(alias, provider string) error {
				events = append(events, "register")
				if alias != "worker" || provider != "codex" {
					t.Fatal("binding changed")
				}
				if mode == "registration-failed" {
					return errors.New("refused")
				}
				return nil
			}
			reg.Update = func(status string) error {
				events = append(events, status)
				if status == "submitted" && mode == "submission-failed" || status == "shutdown_confirmed" && mode == "audit-failed" {
					return errors.New("audit unavailable")
				}
				return nil
			}
			ctx = context.WithValue(ctx, managedJobContextKey{}, reg)
			transport := managedWorkerTransport{
				command: func(commandCtx context.Context, alias, action, encoded string) *exec.Cmd {
					events = append(events, "command")
					request, err := decodeWorkerRequest(encoded)
					if err != nil || action != "run-v2" || request.JobID != reg.JobID || !request.Deadline.Equal(reg.Deadline) {
						t.Fatal("durable identity lost")
					}
					if mode == "command-unavailable" {
						return nil
					}
					return exec.CommandContext(commandCtx, "synthetic-not-started")
				},
				fence: func(cleanupCtx context.Context, alias string, fence workerFenceRequest) error {
					events = append(events, "fence")
					if cleanupCtx.Err() != nil {
						t.Fatal("turn cancellation prevented cleanup")
					}
					deadline, ok := cleanupCtx.Deadline()
					if !ok || time.Until(deadline) > 35*time.Second {
						t.Fatal("cleanup lifetime unbounded")
					}
					if fence.JobID != reg.JobID || fence.Provider != "codex" || fence.Key != "engineering/agent/connection/worker" || fence.Version != 2 {
						t.Fatal("fence binding changed")
					}
					if mode == "shutdown-failed" {
						return errors.New("unconfirmed")
					}
					return nil
				},
				monitor: func(context.Context, <-chan struct{}, string, string, string) {},
			}
			request := workerRunRequest{Version: 2, Provider: "codex", Key: "engineering/agent/connection/worker", Args: append(providerTurnArgs("codex", false, false), "-")}
			cmd, cleanup, err := dispatchManagedWorker(ctx, "worker", request, transport)
			if mode == "registration-failed" {
				if err == nil || cmd != nil || cleanup != nil {
					t.Fatal("failed registration admitted")
				}
				if !reflect.DeepEqual(events, []string{"register"}) {
					t.Fatal(events)
				}
				return
			}
			if mode == "submission-failed" {
				if err == nil || cmd != nil || cleanup != nil {
					t.Fatal("failed submission admitted")
				}
				if !reflect.DeepEqual(events, []string{"register", "submitted", "fence", "shutdown_confirmed"}) {
					t.Fatal(events)
				}
				return
			}
			if mode == "command-unavailable" {
				if err == nil || cmd != nil || cleanup != nil {
					t.Fatal("missing command admitted")
				}
				if !reflect.DeepEqual(events, []string{"register", "submitted", "command", "fence", "shutdown_confirmed"}) {
					t.Fatal(events)
				}
				return
			}
			if err != nil || cmd == nil || cleanup == nil {
				t.Fatalf("dispatch: %v", err)
			}
			cancel()
			cleanup()
			cleanup()
			status := "shutdown_confirmed"
			if mode == "shutdown-failed" {
				status = "shutdown_unconfirmed"
			}
			if !reflect.DeepEqual(events, []string{"register", "submitted", "command", "fence", status}) {
				t.Fatal(events)
			}
		})
	}
}

func TestManagedDispatchRejectsUnavailableContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, ctx := range []context.Context{nil, ctx} {
		cmd, cleanup, err := dispatchManagedWorker(ctx, "worker", workerRunRequest{}, managedWorkerTransport{})
		if err == nil || cmd != nil || cleanup != nil {
			t.Fatal("unavailable context admitted")
		}
	}
}
