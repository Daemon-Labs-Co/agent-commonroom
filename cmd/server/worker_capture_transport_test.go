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

func TestCaptureAuthorityTransportJournalLifecycle(t *testing.T) {
	t.Parallel()
	j, g := captureJournalFixture(t)
	left, right := net.Pipe()
	var stopped atomic.Int32
	c, err := newCaptureAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- serveCaptureAuthority(t.Context(), right, j.record, j.intent, g) }()
	j.guard = c
	beginCaptureJournal(t, j)
	id := strings.Repeat("1", 64)
	if err := j.BindCapture(t.Context(), j.intent.Request, id); err != nil {
		t.Fatal(err)
	}
	if err := j.FinishCapture(t.Context(), j.intent.Request, id, true); err != nil {
		t.Fatal(err)
	}
	if n, err := workerCaptureSlots(j.dir); err != nil || n != 0 {
		t.Fatal(n, err)
	}
	c.StopCaptureAuthority()
	<-done
	if stopped.Load() != 1 || !g.stopped || strings.Join(g.seen, ",") != "begin,bound,finish" {
		t.Fatal(stopped.Load(), g.stopped, g.seen)
	}
}
func TestCaptureAuthorityTransportDenialRetainsCharge(t *testing.T) {
	t.Parallel()
	j, g := captureJournalFixture(t)
	g.failAck = "begin"
	left, right := net.Pipe()
	var stopped atomic.Int32
	c, err := newCaptureAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- serveCaptureAuthority(t.Context(), right, j.record, j.intent, g) }()
	j.guard = c
	if j.BeginCapture(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image) == nil {
		t.Fatal("accepted denied acknowledgment")
	}
	<-done
	if n, err := workerCaptureSlots(j.dir); err != nil || n != 1 {
		t.Fatal("lost uncertain charge", n, err)
	}
	if !workerProbeStopped(j.dir) || stopped.Load() != 1 || !g.stopped {
		t.Fatal("authority not stopped")
	}
	if c.AuthorizeCapture(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("poisoned link reused")
	}
}
func TestCaptureAuthorityTransportRejectsFramesBeforeGuard(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"sequence", "intent", "operation", "event", "missing-record", "mismatched-record", "authorize-record", "version"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			j, g := captureJournalFixture(t)
			q := captureAuthorityRequest{Version: 1, Sequence: 1, Operation: "authorize", Intent: copyCaptureIntent(j.intent), Event: workerProbeEvent{Stage: "begin"}}
			r := workerCaptureRecord{Version: 1, Intent: copyCaptureIntent(j.intent), Events: []workerProbeEvent{{Stage: "begin"}}}
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
			}
			left, right := net.Pipe()
			done := make(chan error, 1)
			go func() { done <- serveCaptureAuthority(t.Context(), right, j.record, j.intent, g) }()
			_ = writeCaptureAuthorityFrame(left, q)
			var response captureAuthorityResponse
			if readCaptureAuthorityFrame(left, &response) == nil {
				t.Fatal("malformed request accepted")
			}
			left.Close()
			if <-done == nil || !g.stopped || len(g.seen) != 0 {
				t.Fatal("invalid request reached guard")
			}
		})
	}
}
func TestCaptureAuthorityClientCancellationAndBadResponses(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"cancel", "sequence", "version", "denied", "noncanonical"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			j, _ := captureJournalFixture(t)
			left, right := net.Pipe()
			var stopped atomic.Int32
			c, err := newCaptureAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer right.Close()
				var q captureAuthorityRequest
				if readCaptureAuthorityFrame(right, &q) != nil {
					return
				}
				if kind == "cancel" {
					cancel()
					var b [1]byte
					_, _ = right.Read(b[:])
					return
				}
				response := captureAuthorityResponse{Version: 1, Sequence: q.Sequence, Accepted: true}
				switch kind {
				case "sequence":
					response.Sequence++
				case "version":
					response.Version = 2
				case "denied":
					response.Accepted = false
				}
				if kind == "noncanonical" {
					raw := []byte(`{"version":1,"sequence":1,"accepted":true,"extra":1}`)
					var header [4]byte
					binary.BigEndian.PutUint32(header[:], uint32(len(raw)))
					_, _ = right.Write(header[:])
					_, _ = right.Write(raw)
				} else {
					_ = writeCaptureAuthorityFrame(right, response)
				}
			}()
			if c.AuthorizeCapture(ctx, j.record, j.intent, workerProbeEvent{Stage: "begin"}) == nil {
				t.Fatal("bad response accepted")
			}
			<-done
			c.StopCaptureAuthority()
			if stopped.Load() != 1 {
				t.Fatal(stopped.Load())
			}
			if c.AuthorizeCapture(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}) == nil {
				t.Fatal("link revived")
			}
		})
	}
}
func TestCaptureAuthorityFrameBoundsAndCanonicalEncoding(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{}`, `{"version":1,"version":1,"sequence":1,"accepted":true}`, `{"version":1, "sequence":1,"accepted":true}`} {
		var b bytes.Buffer
		var header [4]byte
		binary.BigEndian.PutUint32(header[:], uint32(len(raw)))
		b.Write(header[:])
		b.WriteString(raw)
		var response captureAuthorityResponse
		if readCaptureAuthorityFrame(&b, &response) == nil {
			t.Fatal("noncanonical frame accepted")
		}
	}
	for _, n := range []uint32{0, captureAuthorityFrameLimit + 1} {
		var header [4]byte
		binary.BigEndian.PutUint32(header[:], n)
		var response captureAuthorityResponse
		if readCaptureAuthorityFrame(bytes.NewReader(header[:]), &response) == nil {
			t.Fatal("bad size accepted")
		}
	}
	if writeCaptureAuthorityFrame(io.Discard, strings.Repeat("x", captureAuthorityFrameLimit)) == nil {
		t.Fatal("oversized write accepted")
	}
}
func TestCaptureAuthorityServerCancellationClosesOwnedStream(t *testing.T) {
	t.Parallel()
	j, g := captureJournalFixture(t)
	left, right := net.Pipe()
	defer left.Close()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- serveCaptureAuthority(ctx, right, j.record, j.intent, g) }()
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

type mutatingCaptureTransportGuard struct {
	calls   int
	stopped bool
}

func (g *mutatingCaptureTransportGuard) AuthorizeCapture(ctx context.Context, r workerJobRecord, i workerCaptureIntent, e workerProbeEvent) error {
	g.calls++
	if r.Pair.Workspace.Removed || i.Request.Members[0].Name != "result.txt" {
		return errWorkerProbeJournal
	}
	r.Pair.Workspace.Removed = true
	i.Request.Members[0].Name = "changed"
	return ctx.Err()
}
func (g *mutatingCaptureTransportGuard) AcknowledgeCapture(context.Context, workerCaptureRecord, workerProbeEvent) error {
	return errWorkerProbeJournal
}
func (g *mutatingCaptureTransportGuard) StopCaptureAuthority() { g.stopped = true }
func TestCaptureAuthorityTransportDetachesCallbackMetadata(t *testing.T) {
	t.Parallel()
	j, _ := captureJournalFixture(t)
	left, right := net.Pipe()
	c, err := newCaptureAuthorityClient(left, j.record, j.intent, func() {})
	if err != nil {
		t.Fatal(err)
	}
	g := &mutatingCaptureTransportGuard{}
	done := make(chan error, 1)
	go func() { done <- serveCaptureAuthority(t.Context(), right, j.record, j.intent, g) }()
	for range 2 {
		if c.AuthorizeCapture(t.Context(), j.record, j.intent, workerProbeEvent{Stage: "begin"}) != nil {
			t.Fatal("callback changed binding")
		}
	}
	c.StopCaptureAuthority()
	<-done
	if g.calls != 2 || !g.stopped || j.record.Pair.Workspace.Removed || j.intent.Request.Members[0].Name != "result.txt" {
		t.Fatal("metadata mutation escaped")
	}
}
func TestCaptureAuthorityClientRejectsSubstitutedBindings(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"record", "intent", "event", "history"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			j, _ := captureJournalFixture(t)
			left, right := net.Pipe()
			defer right.Close()
			var stopped atomic.Int32
			c, err := newCaptureAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			r := detachedProbeWorkerRecord(j.record)
			i := copyCaptureIntent(j.intent)
			e := workerProbeEvent{Stage: "begin"}
			switch kind {
			case "record":
				r.RoomID = "other"
			case "intent":
				i.Request.Members[0].MaxBytes++
			case "event":
				e.Stage = "confirmed"
			}
			if kind == "history" {
				bad := workerCaptureRecord{Version: 1, Intent: i, Events: []workerProbeEvent{e}}
				err = c.AcknowledgeCapture(t.Context(), bad, workerProbeEvent{Stage: "bound", ID: strings.Repeat("1", 64)})
			} else {
				err = c.AuthorizeCapture(t.Context(), r, i, e)
			}
			if err == nil || stopped.Load() != 1 {
				t.Fatal("substitution accepted", err, stopped.Load())
			}
		})
	}
}
