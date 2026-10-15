package toolbroker

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuditStreamCommitsBeforeToolAndResult(t *testing.T) {
	t.Parallel()
	authority, request := liveFixture(t)
	left, right := net.Pipe()
	t.Cleanup(func() { left.Close(); right.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stage atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- ServeAudit(ctx, right, authority.scope, mockAudit{
			intent: func(context.Context, Intent) error {
				if !stage.CompareAndSwap(0, 1) {
					return errors.New("bad intent order")
				}
				return nil
			},
			outcome: func(context.Context, Intent, Outcome) error {
				if !stage.CompareAndSwap(2, 3) {
					return errors.New("bad outcome order")
				}
				return nil
			},
		})
	}()
	client, err := NewAuditClient(left, authority.scope)
	if err != nil {
		t.Fatal(err)
	}
	dispatch := dispatcher(t, authority, client, mockExecutor(func(_ context.Context, _ Scope, _ Action, out io.Writer) error {
		if !stage.CompareAndSwap(1, 2) {
			return errors.New("tool before durable intent")
		}
		_, err := io.WriteString(out, "ok")
		return err
	}))
	output, err := dispatch.Dispatch(ctx, request)
	if err != nil || string(output) != "ok" || stage.Load() != 3 {
		t.Fatal("unacknowledged result", err, stage.Load())
	}
	left.Close()
	if <-done != ErrAudit {
		t.Fatal("closed link reported success")
	}
}

func TestAuditStreamPersistenceFailureNeverRunsTool(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"intent", "outcome"} {
		t.Run(failure, func(t *testing.T) {
			authority, request := liveFixture(t)
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			var executed atomic.Bool
			done := make(chan error, 1)
			go func() {
				done <- ServeAudit(context.Background(), right, authority.scope, mockAudit{
					intent: func(context.Context, Intent) error {
						if failure == "intent" {
							return ErrAudit
						}
						return nil
					},
					outcome: func(context.Context, Intent, Outcome) error {
						if failure == "outcome" {
							return ErrAudit
						}
						return nil
					},
				})
			}()
			client, _ := NewAuditClient(left, authority.scope)
			dispatch := dispatcher(t, authority, client, mockExecutor(func(context.Context, Scope, Action, io.Writer) error { executed.Store(true); return nil }))
			output, err := dispatch.Dispatch(context.Background(), request)
			if err != ErrAudit || output != nil || executed.Load() != (failure == "outcome") {
				t.Fatal("audit failure admitted work/result", err)
			}
			if authority.Reserve(Request{}, time.Now()) == nil {
				t.Fatal("authority not fenced")
			}
			if <-done != ErrAudit {
				t.Fatal("persistence failure not returned")
			}
		})
	}
}

func TestAuditClientRejectsStaleMalformedAndPartialAcknowledgements(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"digest", "denied", "duplicate", "case", "null", "missing", "oversize", "partial"} {
		t.Run(mode, func(t *testing.T) {
			authority, _ := liveFixture(t)
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer right.Close()
				var frame AuditFrame
				if auditRead(bufio.NewReaderSize(right, MaxAuditFrameBytes), &frame) != nil {
					return
				}
				digest, _ := auditFrameDigest(frame)
				ack := auditAck{1, digest, true}
				raw, _ := json.Marshal(ack)
				switch mode {
				case "digest":
					raw = []byte(`{"version":1,"digest":"` + strings.Repeat("f", 64) + `","accepted":true}`)
				case "denied":
					raw = []byte(`{"version":1,"digest":"` + digest + `","accepted":false}`)
				case "duplicate":
					raw = []byte(`{"version":1,"version":1,"digest":"` + digest + `","accepted":true}`)
				case "case":
					raw = []byte(`{"Version":1,"digest":"` + digest + `","accepted":true}`)
				case "null":
					raw = []byte(`{"version":1,"digest":"` + digest + `","accepted":null}`)
				case "missing":
					raw = []byte(`{"digest":"` + digest + `","accepted":true}`)
				case "oversize":
					raw = []byte(strings.Repeat("x", MaxAuditFrameBytes+1))
				case "partial":
					_, _ = right.Write(raw[:10])
					return
				}
				_, _ = right.Write(append(raw, '\n'))
			}()
			client, _ := NewAuditClient(left, authority.scope)
			intent := Intent{authority.scope, "call", "read", strings.Repeat("b", 64)}
			if client.RecordIntent(context.Background(), intent) != ErrAudit {
				t.Fatal("bad acknowledgement accepted")
			}
			if client.RecordIntent(context.Background(), intent) != ErrAudit {
				t.Fatal("failed channel reused")
			}
			<-done
		})
	}
}

func TestAuditCancellationInterruptsBlockedReadAndWrite(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"write", "read"} {
		t.Run(mode, func(t *testing.T) {
			authority, _ := liveFixture(t)
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, _ := NewAuditClient(left, authority.scope)
			started := make(chan struct{})
			if mode == "read" {
				go func() {
					var frame AuditFrame
					_ = auditRead(bufio.NewReaderSize(right, MaxAuditFrameBytes), &frame)
					close(started)
				}()
			} else {
				close(started)
			}
			done := make(chan error, 1)
			go func() {
				done <- client.RecordIntent(ctx, Intent{authority.scope, "call", "read", strings.Repeat("b", 64)})
			}()
			<-started
			cancel()
			select {
			case err := <-done:
				if err != ErrAudit {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("cancelled I/O leaked")
			}
		})
	}
}

func TestAuditServerRejectsForgedScopeAndReorderedFramesBeforePersistence(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"scope", "sequence", "outcome-first", "invalid-operation", "duplicate-key", "capability", "null"} {
		t.Run(mode, func(t *testing.T) {
			authority, _ := liveFixture(t)
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			var persisted atomic.Bool
			done := make(chan error, 1)
			go func() {
				done <- ServeAudit(context.Background(), right, authority.scope, mockAudit{intent: func(context.Context, Intent) error { persisted.Store(true); return nil }})
			}()
			frame := AuditFrame{Version: 1, Sequence: 1, Stage: "intent", Intent: Intent{authority.scope, "call", "read", strings.Repeat("b", 64)}}
			switch mode {
			case "scope":
				frame.Intent.Scope.JobID = strings.Repeat("f", 32)
			case "sequence":
				frame.Sequence = 2
			case "outcome-first":
				frame.Stage = "outcome"
			case "invalid-operation":
				frame.Intent.Operation = "shell"
			}
			raw, _ := json.Marshal(frame)
			switch mode {
			case "duplicate-key":
				raw = append([]byte(`{"version":1,`), raw[1:]...)
			case "capability":
				raw = append([]byte(`{"capability":"secret-must-not-enter-audit",`), raw[1:]...)
			case "null":
				raw = []byte(`{"version":1,"sequence":1,"stage":"intent","intent":null}`)
			}
			_, _ = left.Write(append(raw, '\n'))
			if <-done != ErrAudit || persisted.Load() {
				t.Fatal("forged audit frame persisted")
			}
		})
	}
}
