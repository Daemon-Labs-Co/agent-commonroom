package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestObjectReadAuthorityTransportJournalLifecycle(t *testing.T) {
	t.Parallel()
	j, g := objectReadTransportFixture(t)
	left, right := net.Pipe()
	var stopped atomic.Int32
	c, err := newObjectReadAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- serveObjectReadAuthority(t.Context(), right, j.record, j.intent, g) }()
	j.guard = c
	beginObjectReadJournal(t, j)
	id := strings.Repeat("1", 64)
	if err := j.BindObjectRead(t.Context(), j.intent.Request, id); err != nil {
		t.Fatal(err)
	}
	if err := j.FinishObjectRead(t.Context(), j.intent.Request, id, true); err != nil {
		t.Fatal(err)
	}
	if n, err := workerObjectReadSlots(j.dir); err != nil || n != 0 {
		t.Fatal(n, err)
	}
	c.StopObjectReadAuthority()
	<-done
	if stopped.Load() != 1 || !g.stopped || strings.Join(g.seen, ",") != "begin,bound,finish" {
		t.Fatal(stopped.Load(), g.stopped, g.seen)
	}
}
func TestObjectReadAuthorityTransportDenialRetainsCharge(t *testing.T) {
	t.Parallel()
	j, g := objectReadTransportFixture(t)
	g.failAck = "begin"
	left, right := net.Pipe()
	var stopped atomic.Int32
	c, err := newObjectReadAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- serveObjectReadAuthority(t.Context(), right, j.record, j.intent, g) }()
	j.guard = c
	if j.BeginObjectRead(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image) == nil {
		t.Fatal("accepted denied acknowledgment")
	}
	<-done
	if n, err := workerObjectReadSlots(j.dir); err != nil || n != 1 {
		t.Fatal("lost uncertain charge", n, err)
	}
	if !workerProbeStopped(j.dir) || stopped.Load() != 1 || !g.stopped {
		t.Fatal("authority not stopped")
	}
	if c.AuthorizeObjectRead(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("poisoned link reused")
	}
}
func TestObjectReadAuthorityTransportRejectsFramesBeforeGuard(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"sequence", "intent", "operation", "event", "missing-record", "mismatched-record", "authorize-record", "version", "protocol"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			j, g := objectReadTransportFixture(t)
			q := objectReadAuthorityRequest{Protocol: "object-read-v1", Version: 1, Sequence: 1, Operation: "authorize", Intent: j.intent, Event: workerProbeEvent{Stage: "begin"}}
			r := workerObjectReadRecord{Version: 1, Intent: j.intent, Events: []workerProbeEvent{{Stage: "begin"}}}
			switch kind {
			case "sequence":
				q.Sequence = 2
			case "intent":
				q.Intent.Request.Challenge = strings.Repeat("7", 32)
			case "operation":
				q.Operation = "execute"
			case "event":
				q.Event.Stage = "confirmed"
			case "missing-record":
				q.Operation = "acknowledge"
			case "mismatched-record":
				q.Operation = "acknowledge"
				r.Intent.Image = "sha256:" + strings.Repeat("8", 64)
				q.Record = &r
			case "authorize-record":
				q.Record = &r
			case "version":
				q.Version = 2
			case "protocol":
				q.Protocol = "capture-v1"
			}
			left, right := net.Pipe()
			done := make(chan error, 1)
			go func() { done <- serveObjectReadAuthority(t.Context(), right, j.record, j.intent, g) }()
			_ = writeObjectReadAuthorityFrame(left, q)
			var response objectReadAuthorityResponse
			if readObjectReadAuthorityFrame(left, &response) == nil {
				t.Fatal("malformed request accepted")
			}
			left.Close()
			if <-done == nil || !g.stopped || len(g.seen) != 0 {
				t.Fatal("invalid request reached guard")
			}
		})
	}
}
func TestObjectReadAuthorityClientCancellationAndBadResponses(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"cancel", "sequence", "version", "denied", "noncanonical", "protocol"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			j, _ := objectReadTransportFixture(t)
			left, right := net.Pipe()
			var stopped atomic.Int32
			c, err := newObjectReadAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer right.Close()
				var q objectReadAuthorityRequest
				if readObjectReadAuthorityFrame(right, &q) != nil {
					return
				}
				if kind == "cancel" {
					cancel()
					var b [1]byte
					_, _ = right.Read(b[:])
					return
				}
				response := objectReadAuthorityResponse{Protocol: "object-read-v1", Version: 1, Sequence: q.Sequence, Accepted: true}
				switch kind {
				case "sequence":
					response.Sequence++
				case "version":
					response.Version = 2
				case "denied":
					response.Accepted = false
				case "protocol":
					response.Protocol = "capture-v1"
				}
				if kind == "noncanonical" {
					raw := []byte(`{"version":1,"sequence":1,"accepted":true,"extra":1}`)
					var header [4]byte
					binary.BigEndian.PutUint32(header[:], uint32(len(raw)))
					_, _ = right.Write(header[:])
					_, _ = right.Write(raw)
				} else {
					_ = writeObjectReadAuthorityFrame(right, response)
				}
			}()
			if c.AuthorizeObjectRead(ctx, j.record, j.intent, workerProbeEvent{Stage: "begin"}) == nil {
				t.Fatal("bad response accepted")
			}
			<-done
			c.StopObjectReadAuthority()
			if stopped.Load() != 1 {
				t.Fatal(stopped.Load())
			}
			if c.AuthorizeObjectRead(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}) == nil {
				t.Fatal("link revived")
			}
		})
	}
}
func TestObjectReadAuthorityFrameBoundsAndCanonicalEncoding(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{}`, `{"version":1,"version":1,"sequence":1,"accepted":true}`, `{"version":1, "sequence":1,"accepted":true}`} {
		var b bytes.Buffer
		var header [4]byte
		binary.BigEndian.PutUint32(header[:], uint32(len(raw)))
		b.Write(header[:])
		b.WriteString(raw)
		var response objectReadAuthorityResponse
		if readObjectReadAuthorityFrame(&b, &response) == nil {
			t.Fatal("noncanonical frame accepted")
		}
	}
	for _, n := range []uint32{0, objectReadAuthorityFrameLimit + 1} {
		var header [4]byte
		binary.BigEndian.PutUint32(header[:], n)
		var response objectReadAuthorityResponse
		if readObjectReadAuthorityFrame(bytes.NewReader(header[:]), &response) == nil {
			t.Fatal("bad size accepted")
		}
	}
	if writeObjectReadAuthorityFrame(io.Discard, strings.Repeat("x", objectReadAuthorityFrameLimit)) == nil {
		t.Fatal("oversized write accepted")
	}
}
func TestObjectReadAuthorityServerCancellationClosesOwnedStream(t *testing.T) {
	t.Parallel()
	j, g := objectReadTransportFixture(t)
	left, right := net.Pipe()
	defer left.Close()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- serveObjectReadAuthority(ctx, right, j.record, j.intent, g) }()
	cancel()
	// Cancellation can occur before server validation, when ownership has not yet
	// transferred. Closing our peer still guarantees no goroutine is stranded.
	left.Close()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled serve succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("server did not terminate")
	}
}

type mutatingObjectReadTransportGuard struct {
	calls   int
	stopped bool
}

func (g *mutatingObjectReadTransportGuard) AuthorizeObjectRead(ctx context.Context, r workerJobRecord, i workerObjectReadIntent, e workerProbeEvent) error {
	g.calls++
	if r.Pair.Workspace.Removed || i.Request.Receipt.Intent.Scope.MemberName != "result.txt" {
		return errWorkerProbeJournal
	}
	r.Pair.Workspace.Removed = true
	i.Request.Receipt.Intent.Scope.MemberName = "changed"
	return ctx.Err()
}
func (g *mutatingObjectReadTransportGuard) AcknowledgeObjectRead(context.Context, workerObjectReadRecord, workerProbeEvent) error {
	return errWorkerProbeJournal
}
func (g *mutatingObjectReadTransportGuard) StopObjectReadAuthority() { g.stopped = true }
func TestObjectReadAuthorityTransportDetachesCallbackMetadata(t *testing.T) {
	t.Parallel()
	j, _ := objectReadTransportFixture(t)
	left, right := net.Pipe()
	c, err := newObjectReadAuthorityClient(left, j.record, j.intent, func() {})
	if err != nil {
		t.Fatal(err)
	}
	g := &mutatingObjectReadTransportGuard{}
	done := make(chan error, 1)
	go func() { done <- serveObjectReadAuthority(t.Context(), right, j.record, j.intent, g) }()
	for range 2 {
		if c.AuthorizeObjectRead(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}) != nil {
			t.Fatal("callback changed binding")
		}
	}
	c.StopObjectReadAuthority()
	<-done
	if g.calls != 2 || !g.stopped || j.record.Pair.Workspace.Removed || j.intent.Request.Receipt.Intent.Scope.MemberName != "result.txt" {
		t.Fatal("metadata mutation escaped")
	}
}
func TestObjectReadAuthorityClientRejectsSubstitutedBindings(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"record", "intent", "event", "history"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			j, _ := objectReadTransportFixture(t)
			left, right := net.Pipe()
			defer right.Close()
			var stopped atomic.Int32
			c, err := newObjectReadAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			r := detachedProbeWorkerRecord(j.record)
			i := j.intent
			e := workerProbeEvent{Stage: "begin"}
			switch kind {
			case "record":
				r.RoomID = "other"
			case "intent":
				i.Request.Receipt.Bytes++
			case "event":
				e.Stage = "confirmed"
			}
			if kind == "history" {
				bad := workerObjectReadRecord{Version: 1, Intent: i, Events: []workerProbeEvent{e}}
				err = c.AcknowledgeObjectRead(t.Context(), bad, workerProbeEvent{Stage: "bound", ID: strings.Repeat("1", 64)})
			} else {
				err = c.AuthorizeObjectRead(t.Context(), r, i, e)
			}
			if err == nil || stopped.Load() != 1 {
				t.Fatal("substitution accepted", err, stopped.Load())
			}
		})
	}
}

func objectReadTransportFixture(t *testing.T) (*workerObjectReadJournal, *objectReadJournalGuard) {
	t.Helper()
	j, g, _, _ := objectReadJournalFixture(t)
	return j, g
}

func TestObjectReadAuthorityTransportFrameBudget(t *testing.T) {
	t.Parallel()
	j, g := objectReadTransportFixture(t)
	left, right := net.Pipe()
	c, err := newObjectReadAuthorityClient(left, j.record, j.intent, func() {})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { defer close(done); done <- serveObjectReadAuthority(t.Context(), right, j.record, j.intent, g) }()
	defer func() { c.StopObjectReadAuthority(); <-done }()
	for range 64 {
		if err = c.AuthorizeObjectRead(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}); err != nil {
			t.Fatal(err)
		}
	}
	if c.AuthorizeObjectRead(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("control frame budget unbounded")
	}
}

func TestObjectReadAuthorityIdleEOFRevokesLifetime(t *testing.T) {
	t.Parallel()
	j, _ := objectReadTransportFixture(t)
	left, right := net.Pipe()
	var stopped atomic.Int32
	c, err := newObjectReadAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	defer c.StopObjectReadAuthority()
	_ = right.Close()
	select {
	case <-c.ObjectReadAuthorityLifetime().Done():
	case <-time.After(time.Second):
		t.Fatal("idle EOF did not revoke authority")
	}
	c.StopObjectReadAuthority()
	if stopped.Load() != 1 {
		t.Fatal("stop callback not once", stopped.Load())
	}
	if c.AuthorizeObjectRead(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("EOF link revived")
	}
}
