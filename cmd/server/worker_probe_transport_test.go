package main

import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacemount"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type transportProbeGuard struct {
	calls []string
	deny  string
}

func (g *transportProbeGuard) Authorize(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error {
	g.calls = append(g.calls, "authorize")
	if g.deny == "authorize" {
		return errors.New("denied")
	}
	return ctx.Err()
}
func (g *transportProbeGuard) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	g.calls = append(g.calls, e.Stage)
	if g.deny == "acknowledge" {
		return errors.New("denied")
	}
	return ctx.Err()
}
func (g *transportProbeGuard) AuthorizeWorkspaceObservation(ctx context.Context, r workerJobRecord) error {
	g.calls = append(g.calls, "observe")
	r.Pair.Workspace.Removed = true
	if g.deny == "observe" {
		return errors.New("denied")
	}
	return ctx.Err()
}
func (g *transportProbeGuard) StopAuthority() {}
func transportProbeBinding(t *testing.T) (workerJobRecord, toolbroker.WorkspaceProbeIntent) {
	_, r, _ := retainedObservationFixture(t)
	return r, toolbroker.WorkspaceProbeIntent{JobID: r.JobID, Repository: r.Pair.Repository, Volume: r.Pair.Workspace.Identity, Image: "sha256:" + strings.Repeat("d", 64), Challenge: strings.Repeat("e", 32)}
}
func TestProbeAuthorityTransportRoundTrip(t *testing.T) {
	t.Parallel()
	r, i := transportProbeBinding(t)
	left, right := net.Pipe()
	var stopped atomic.Int32
	client, err := newProbeAuthorityClient(left, r, i, func() { stopped.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	guard := &transportProbeGuard{}
	done := make(chan error, 1)
	go func() { done <- serveProbeAuthority(ctx, right, r, i, guard) }()
	for _, call := range []func() error{func() error { return client.Authorize(ctx, i) }, func() error { return client.AuthorizeWorkspaceObservation(ctx, r) }, func() error { return client.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) }, func() error { return client.AuthorizeWorkspaceObservation(ctx, r) }} {
		if err := call(); err != nil {
			t.Fatal(err)
		}
	}
	// Callback changes must never mutate the caller's or the next request's record.
	if r.Pair.Workspace.Removed {
		t.Fatal("callback mutated retained record")
	}
	left.Close()
	<-done
	if strings.Join(guard.calls, ",") != "authorize,observe,begin,observe" || stopped.Load() != 0 {
		t.Fatal(guard.calls, stopped.Load())
	}
}
func TestProbeAuthorityDenialPoisonsLink(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"authorize", "observe", "acknowledge"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			r, i := transportProbeBinding(t)
			left, right := net.Pipe()
			var stopped atomic.Int32
			client, err := newProbeAuthorityClient(left, r, i, func() { stopped.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			guard := &transportProbeGuard{deny: operation}
			done := make(chan error, 1)
			go func() { done <- serveProbeAuthority(t.Context(), right, r, i, guard) }()
			switch operation {
			case "authorize":
				err = client.Authorize(t.Context(), i)
			case "observe":
				err = client.AuthorizeWorkspaceObservation(t.Context(), r)
			case "acknowledge":
				err = client.Acknowledge(t.Context(), i, workerProbeEvent{Stage: "begin"})
			}
			if err == nil {
				t.Fatal("denial acknowledged")
			}
			if client.Authorize(t.Context(), i) == nil {
				t.Fatal("poisoned stream reused")
			}
			<-done
			client.StopAuthority()
			if stopped.Load() != 1 {
				t.Fatal("stop not exactly once", stopped.Load())
			}
		})
	}
}
func TestProbeAuthorityRejectsInvalidFramesBeforeGuard(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"foreign intent", "sequence", "version", "operation", "unexpected event", "unknown field", "duplicate field", "oversized", "truncated"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			r, i := transportProbeBinding(t)
			left, right := net.Pipe()
			defer left.Close()
			guard := &transportProbeGuard{}
			done := make(chan error, 1)
			go func() { done <- serveProbeAuthority(t.Context(), right, r, i, guard) }()
			req := probeAuthorityRequest{Version: 1, Sequence: 1, Operation: "authorize", Intent: i}
			switch mode {
			case "foreign intent":
				req.Intent.Challenge = strings.Repeat("f", 32)
			case "sequence":
				req.Sequence = 2
			case "version":
				req.Version = 2
			case "operation":
				req.Operation = "execute"
			case "unexpected event":
				req.Event.Stage = "begin"
			}
			var wire bytes.Buffer
			if mode == "oversized" {
				var header [4]byte
				binary.BigEndian.PutUint32(header[:], probeAuthorityFrameLimit+1)
				wire.Write(header[:])
			} else if mode == "truncated" {
				wire.Write([]byte{0, 0})
			} else {
				if err := writeProbeAuthorityFrame(&wire, req); err != nil {
					t.Fatal(err)
				}
				if mode == "unknown field" || mode == "duplicate field" {
					raw := wire.Bytes()[4:]
					extra := `,"extra":true}`
					if mode == "duplicate field" {
						extra = `,"version":1}`
					}
					raw = append(append([]byte{}, raw[:len(raw)-1]...), []byte(extra)...)
					wire.Reset()
					var header [4]byte
					binary.BigEndian.PutUint32(header[:], uint32(len(raw)))
					wire.Write(header[:])
					wire.Write(raw)
				}
			}
			left.Write(wire.Bytes())
			if mode == "truncated" {
				left.Close()
			}
			if err := <-done; err == nil {
				t.Fatal("invalid frame accepted")
			}
			if len(guard.calls) != 0 {
				t.Fatal("guard called for invalid frame", guard.calls)
			}
		})
	}
}
func TestProbeAuthorityCancellationAndWrongResponse(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"cancel", "wrong sequence", "wrong version"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			r, i := transportProbeBinding(t)
			left, right := net.Pipe()
			defer right.Close()
			var stopped atomic.Int32
			client, err := newProbeAuthorityClient(left, r, i, func() { stopped.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			read := make(chan struct{})
			go func() {
				var request probeAuthorityRequest
				readProbeAuthorityFrame(right, &request)
				close(read)
				if mode != "cancel" {
					v, s := 1, uint64(1)
					if mode == "wrong sequence" {
						s = 2
					} else {
						v = 2
					}
					writeProbeAuthorityFrame(right, probeAuthorityResponse{v, s, true})
				}
			}()
			done := make(chan error, 1)
			go func() { done <- client.Authorize(ctx, i) }()
			<-read
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("uncertain response accepted")
				}
			case <-time.After(time.Second):
				t.Fatal("cancellation did not unblock")
			}
			if stopped.Load() != 1 {
				t.Fatal("authority not stopped")
			}
		})
	}
}

func TestProbeAuthorityChannelRunsRealProbeLifecycle(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"healthy", "audit begin", "audit bound", "audit finish"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f := newAssembledProbeFixture(t, mode)
			var done chan error
			connect := func(ctx context.Context, r workerJobRecord, i toolbroker.WorkspaceProbeIntent) (io.ReadWriteCloser, error) {
				left, right := net.Pipe()
				done = make(chan error, 1)
				go func() { done <- serveProbeAuthority(ctx, right, r, i, f) }()
				return left, nil
			}
			var stopped atomic.Int32
			proof, err := probeRetainedWorkerWorkspaceChannel(t.Context(), f.dir, f.record, syntheticObservationDockerRoot, f, "sha256:"+strings.Repeat("d", 64), connect, func() { stopped.Add(1) })
			if done == nil {
				t.Fatal("channel not connected")
			}
			<-done
			if mode == "healthy" {
				if err != nil || !workspacemount.ValidProof(proof, f.challenge) || !f.removed || stopped.Load() != 0 {
					t.Fatal("healthy lifecycle failed", err, proof)
				}
				if len(f.events) != 3 || f.events[0].Stage != "begin" || f.events[1].Stage != "bound" || f.events[2].Stage != "finish" {
					t.Fatal(f.events)
				}
			} else {
				if err == nil || proof != (workspacemount.Proof{}) || stopped.Load() != 1 || !workerProbeStopped(f.dir) {
					t.Fatal("failed audit released proof or authority", err, proof, stopped.Load())
				}
				if mode == "audit begin" && f.creates != 0 {
					t.Fatal("created before durable acknowledgment")
				}
				if mode == "audit bound" && f.started {
					t.Fatal("started before durable binding")
				}
			}
		})
	}
}

func TestProbeAuthorityReplayDoesNotRepeatGuard(t *testing.T) {
	t.Parallel()
	r, i := transportProbeBinding(t)
	left, right := net.Pipe()
	defer left.Close()
	guard := &transportProbeGuard{}
	done := make(chan error, 1)
	go func() { done <- serveProbeAuthority(t.Context(), right, r, i, guard) }()
	request := probeAuthorityRequest{1, 1, "authorize", i, workerProbeEvent{}}
	if writeProbeAuthorityFrame(left, request) != nil {
		t.Fatal("write")
	}
	var response probeAuthorityResponse
	if readProbeAuthorityFrame(left, &response) != nil || !response.Accepted {
		t.Fatal("first request refused")
	}
	if writeProbeAuthorityFrame(left, request) != nil {
		t.Fatal("write replay")
	}
	if <-done == nil || len(guard.calls) != 1 {
		t.Fatal("replayed authority invoked", guard.calls)
	}
}
func TestProbeAuthorityRejectsUnboundRecord(t *testing.T) {
	t.Parallel()
	r, i := transportProbeBinding(t)
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()
	changed := i
	changed.Repository = strings.Repeat("f", 64)
	if _, err := newProbeAuthorityClient(left, r, changed, func() {}); err == nil {
		t.Fatal("foreign repository accepted")
	}
	if err := serveProbeAuthority(t.Context(), right, r, changed, &transportProbeGuard{}); err == nil {
		t.Fatal("foreign server binding accepted")
	}
	client, err := newProbeAuthorityClient(left, r, i, func() {})
	if err != nil {
		t.Fatal(err)
	}
	if client.Authorize(t.Context(), changed) == nil {
		t.Fatal("client switched intent")
	}
	changedRecord := r
	pair := *r.Pair
	workspace := *pair.Workspace
	workspace.Removed = true
	pair.Workspace = &workspace
	changedRecord.Pair = &pair
	if client.AuthorizeWorkspaceObservation(t.Context(), changedRecord) == nil {
		t.Fatal("client switched retained record")
	}
}
