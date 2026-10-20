package main

import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/toolcommand"
	"context"
	"errors"
	"io"
	"net"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

// Substitute only Docker operations. Real session/bootstrap/command capture,
// exact-container cleanup, audit framing and controller persistence run normally.
type commandOutcomeDocker struct {
	*codexIntegrationDocker
	t            *testing.T
	server       *server
	db           *mockDB
	effects      []string
	commandCalls int
	fenced       bool
	failOutcome  bool
	outputBytes  int
	timeout      bool
}

func (r *commandOutcomeDocker) Run(ctx context.Context, args []string, input []byte, out io.Writer) error {
	if r.timeout && args[0] == "rm" {
		deadline, bounded := ctx.Deadline()
		if !slices.Equal(args, []string{"rm", "--force", strings.Repeat("d", 64)}) || ctx.Err() != nil || !bounded || time.Until(deadline) > 10*time.Second {
			r.t.Fatal("timeout cleanup lacked a fresh bounded context or exact container identity")
		}
	}
	if len(args) > 6 && args[0] == "exec" && args[6] == "/bin/mkdir" {
		want := []string{"exec", "--user", "1000:1000", "--workdir", "/workspace", strings.Repeat("d", 64), "/bin/mkdir", "-p", "/workspace/.commonroom-command/cache", "/workspace/.commonroom-command/tmp", "/workspace/.commonroom-command/modules"}
		if !slices.Equal(args, want) || input != nil {
			r.t.Fatal("command setup escaped fixed container policy")
		}
		return nil
	}
	err := r.codexIntegrationDocker.Run(ctx, args, input, out)
	if err == nil && args[0] == "ps" && r.commandCalls == 1 && r.removed == 1 && len(r.containers) == 0 && !r.fenced {
		r.fenced = true
		r.effects = append(r.effects, "tool_container_absent")
	}
	return err
}

func (r *commandOutcomeDocker) RunCommand(ctx context.Context, args []string, out io.Writer) (int, error) {
	if len(args) < 4 || !slices.Equal(args[len(args)-4:], []string{strings.Repeat("d", 64), "/usr/local/go/bin/go", "test", "./..."}) || !slices.Contains(args, "GOPROXY=off") {
		r.t.Fatal("command escaped fixed isolated profile")
	}
	st, err := r.server.load(ctx)
	if err != nil {
		return -1, err
	}
	call := st.ToolAudits[toolAuditKey(r.scope.JobID, "command-outcome")]
	if call.Status != "pending" || call.ActorID != "human" || call.Operation != "command" {
		r.t.Fatal("command ran without committed human-bound intent")
	}
	r.effects = append(r.effects, "intent_durable", "isolated_command")
	r.commandCalls++
	if r.timeout {
		started := time.Now()
		_, _ = io.WriteString(out, "synthetic command output")
		<-ctx.Done()
		if ctx.Err() != context.DeadlineExceeded || time.Since(started) != toolcommand.Timeout {
			r.t.Fatal("command did not reach its fixed timeout", ctx.Err(), time.Since(started))
		}
		r.effects = append(r.effects, "command_cancelled")
		return -1, ctx.Err()
	}
	if r.failOutcome {
		r.db.failAt = "insert"
	}
	if r.outputBytes > 0 {
		_, err = io.WriteString(out, strings.Repeat("x", r.outputBytes))
		if errors.Is(err, toolcommand.ErrOutputLimit) && ctx.Err() != nil {
			r.effects = append(r.effects, "output_limit_reported")
		}
		return 0, err
	}
	_, err = io.WriteString(out, "synthetic command output")
	return 0, err
}

func driveP0CommandOutcomeFailure(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	if scenario.Given["audit_outcome"] != "fails" {
		t.Fatal("unknown command outcome fixture")
	}
	return observeCommandOutcome(t, true)
}

func observeCommandOutcome(t *testing.T, fail bool) pendingP0Observation {
	t.Helper()
	return observeCommandCapture(t, fail, 0)
}

func observeCommandCapture(t *testing.T, fail bool, outputBytes int) pendingP0Observation {
	t.Helper()
	return observeCommandExecution(t, fail, outputBytes, false)
}

func observeCommandExecution(t *testing.T, fail bool, outputBytes int, timeout bool) pendingP0Observation {
	t.Helper()
	overflow := outputBytes > toolcommand.MaxOutputBytes
	s, db, auditor, intent := scopedAuditFixtureWithLifetime(t, 3*time.Minute)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	done := make(chan error, 1)
	go func() { done <- toolbroker.ServeAudit(ctx, peer, intent.Scope, auditor) }()
	runner := &commandOutcomeDocker{codexIntegrationDocker: &codexIntegrationDocker{containers: map[string]map[string]any{}, scope: intent.Scope}, t: t, server: s, db: db, failOutcome: fail, outputBytes: outputBytes, timeout: timeout}
	workerCtx := context.WithValue(ctx, workerLocalJobContextKey{}, intent.Scope.JobID)
	session, token, err := newWorkerToolSessionWithCommands(workerCtx, intent.Scope, []string{"command"},
		"ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), syntheticDockerRoot, time.Now(), client, runner, true)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	volumeIntent, err := session.Runtime.WorkspaceIntent(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.Runtime.PrepareVolume(ctx, volumeIntent); err != nil {
		t.Fatal(err)
	}
	id, err := session.Runtime.Create(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Runtime.Start(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Runtime.Bootstrap(ctx, io.NopCloser(strings.NewReader("synthetic approved bundle"))); err != nil {
		t.Fatal(err)
	}
	request, err := workerCodexToolRequest(intent.Scope, token, "command-outcome", []byte(`{"operation":"command","command":["go","test","./..."]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, dispatchErr := session.Dispatcher.Dispatch(ctx, request)
	client.Close()
	awaitAuditResult(t, done)
	db.failAt = ""
	st, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	call := st.ToolAudits[toolAuditKey(intent.Scope.JobID, request.CallID)]
	if fail && (call.Status != "pending" || !s.integrityFault) {
		t.Fatal("outcome failure lost unresolved intent or integrity stop")
	}
	if overflow {
		if call.Status != "pending" && call.Status != "output_limit" {
			t.Fatal("overflow fabricated success evidence", call.Status)
		}
		if !errors.Is(dispatchErr, toolbroker.ErrAudit) && !errors.Is(dispatchErr, toolbroker.ErrOutputLimit) {
			t.Fatal("overflow failure cause lost", dispatchErr)
		}
		if call.Status == "pending" && !errors.Is(dispatchErr, toolbroker.ErrAudit) {
			t.Fatal("uncommitted outcome was acknowledged")
		}
	}
	if timeout {
		if call.Status != "pending" && call.Status != "failed" {
			t.Fatal("timeout fabricated successful audit evidence", call.Status)
		}
		if !errors.Is(dispatchErr, toolbroker.ErrAudit) && !errors.Is(dispatchErr, toolbroker.ErrExecution) {
			t.Fatal("timeout was not refused", dispatchErr)
		}
		if call.Status == "pending" && !errors.Is(dispatchErr, toolbroker.ErrAudit) {
			t.Fatal("pending timeout outcome was acknowledged")
		}
	}
	if !fail && !overflow && !timeout && (call.Status != "succeeded" || s.integrityFault) {
		t.Fatal("healthy outcome was not committed")
	}
	if runner.commandCalls != 1 || !runner.fenced || runner.volume == nil {
		t.Fatal("command/absence proof missing or source discarded")
	}
	out := pendingP0Observation{State: "allow", Effects: runner.effects, Values: map[string]string{}}
	if errors.Is(dispatchErr, toolbroker.ErrAudit) || errors.Is(dispatchErr, toolbroker.ErrOutputLimit) || errors.Is(dispatchErr, toolbroker.ErrExecution) {
		out.State = "block"
	}
	if len(result) > 0 {
		out.Effects = append(out.Effects, "result_released")
	}
	if fail || overflow || timeout {
		request.CallID = "another-command"
		if more, err := session.Dispatcher.Dispatch(ctx, request); !errors.Is(err, toolbroker.ErrRejected) || len(more) != 0 || runner.commandCalls != 1 {
			t.Fatal("failed command authority was not revoked")
		}
	}
	for _, event := range db.snapshot() {
		if strings.Contains(event.payload, token) || strings.Contains(event.payload, "synthetic command output") {
			t.Fatal("raw output or capability entered audit")
		}
	}
	return out
}

func driveP0CommandTimeout(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	if scenario.Given["command"] != "never-exits" || scenario.Given["timeout"] != "exceeded" {
		t.Fatal("unknown command timeout fixture")
	}
	var out pendingP0Observation
	synctest.Test(t, func(t *testing.T) {
		out = observeCommandExecution(t, false, 0, true)
	})
	return out
}

func TestCommandOutcomeAcceptanceHealthyControl(t *testing.T) {
	t.Parallel()
	out := observeCommandOutcome(t, false)
	want := []string{"intent_durable", "isolated_command", "tool_container_absent", "result_released"}
	if out.State != "allow" || !slices.Equal(out.Effects, want) {
		t.Fatal("healthy control did not release its audited result", out)
	}
}

func driveP0CommandOutputLimit(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	if scenario.Given["output"] != "over-bound" {
		t.Fatal("unknown output fixture")
	}
	return observeCommandCapture(t, false, toolcommand.MaxOutputBytes+1)
}

func TestCommandCaptureExactOutputLimitControl(t *testing.T) {
	t.Parallel()
	out := observeCommandCapture(t, false, toolcommand.MaxOutputBytes)
	want := []string{"intent_durable", "isolated_command", "tool_container_absent", "result_released"}
	if out.State != "allow" || !slices.Equal(out.Effects, want) {
		t.Fatal("exact limit refused", out)
	}
}
