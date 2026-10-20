package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

// These drivers cover refusal before execution, not successful command isolation,
// compiler behavior, descendant shutdown or live worker/provider compatibility.
type commandAdmissionProbe struct{ calls int }

func (p *commandAdmissionProbe) Run(context.Context, []string, []byte, io.Writer) error {
	p.calls++
	return errors.New("unexpected mocked Docker operation")
}

type commandAdmissionLink struct{ calls int }

func (p *commandAdmissionLink) Read([]byte) (int, error) {
	p.calls++
	return 0, io.EOF
}
func (p *commandAdmissionLink) Write([]byte) (int, error) {
	p.calls++
	return 0, io.ErrClosedPipe
}
func (*commandAdmissionLink) Close() error { return nil }

func driveP0CommandAdmission(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	_, _, _, intent := scopedAuditFixture(t)
	ctx := context.WithValue(t.Context(), workerLocalJobContextKey{}, intent.Scope.JobID)
	runner, channel := &commandAdmissionProbe{}, &commandAdmissionLink{}
	allowed, enabled := []string{"read"}, true
	if scenario.Given["command_feature"] == "off" {
		allowed, enabled = []string{"command"}, false
	} else if scenario.Given["permission"] != "missing" {
		t.Fatal("unknown command refusal fixture")
	}
	session, token, err := newWorkerToolSessionWithCommands(ctx, intent.Scope, allowed,
		"ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), syntheticDockerRoot,
		time.Now(), channel, runner, enabled)
	if session != nil {
		defer session.Close()
	}
	if enabled && (err != nil || session == nil || session.CommandsEnabled()) {
		t.Fatal("read-only session fixture must construct without command authority", err)
	}
	var result []byte
	if err == nil {
		request, requestErr := workerCodexToolRequest(intent.Scope, token, "command-refusal",
			[]byte(`{"operation":"command","command":["go","test","./..."]}`))
		if requestErr != nil {
			t.Fatal("valid command fixture refused by parser", requestErr)
		}
		result, err = session.Dispatcher.Dispatch(ctx, request)
	}
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if errors.Is(err, toolbroker.ErrRejected) {
		out.State = "deny"
	}
	if runner.calls != 0 {
		out.Effects = append(out.Effects, "tool_executed")
	}
	if channel.calls != 0 {
		out.Effects = append(out.Effects, "audit_transport_used")
	}
	if len(result) != 0 {
		out.Effects = append(out.Effects, "result_released")
	}
	return out
}

func driveP0CommandIntentFailure(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	if scenario.Given["audit_intent"] != "fails" {
		t.Fatal("unknown audit failure fixture")
	}
	s, db, auditor, intent := scopedAuditFixture(t)
	before := db.snapshot()
	db.failAt = "insert"
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	clientLink, controllerLink := net.Pipe()
	defer clientLink.Close()
	defer controllerLink.Close()
	done := make(chan error, 1)
	go func() { done <- toolbroker.ServeAudit(ctx, controllerLink, intent.Scope, auditor) }()
	client, err := toolbroker.NewAuditClient(clientLink, intent.Scope)
	if err != nil {
		t.Fatal(err)
	}
	authority, token, err := toolbroker.New(intent.Scope, []string{"command"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Revoke()
	executed := false
	dispatcher, err := toolbroker.NewDispatcher(authority, client,
		scopedAuditExecutor(func(context.Context, toolbroker.Scope, toolbroker.Action, io.Writer) error {
			executed = true
			return errors.New("unexpected command execution")
		}))
	if err != nil {
		t.Fatal(err)
	}
	request, err := workerCodexToolRequest(intent.Scope, token, "command-intent-failure",
		[]byte(`{"operation":"command","command":["go","test","./..."]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, dispatchErr := dispatcher.Dispatch(ctx, request)
	clientLink.Close()
	awaitAuditResult(t, done)
	db.failAt = ""
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if errors.Is(dispatchErr, toolbroker.ErrAudit) {
		out.State = "block"
	}
	if executed {
		out.Effects = append(out.Effects, "tool_executed")
	}
	if len(result) != 0 {
		out.Effects = append(out.Effects, "result_released")
	}
	if !reflect.DeepEqual(before, db.snapshot()) {
		out.Effects = append(out.Effects, "uncommitted_audit_persisted")
	}
	if !s.integrityFault {
		out.Effects = append(out.Effects, "integrity_stop_missing")
	}
	return out
}
