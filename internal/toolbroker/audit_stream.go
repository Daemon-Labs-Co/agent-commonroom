package toolbroker

import (
	"bufio"
	"context"
	"io"
	"sync"
	"time"
)

// AuditClient implements Auditor across a private, authenticated controller link.
// The link MUST be separate from model/stdout traffic. Close must interrupt I/O;
// the pair supervisor must stop both containers after any channel failure. A
// source fork CAN feed self-acknowledgements, BUT invalidates durable admission.
type AuditClient struct {
	channel  io.ReadWriteCloser
	reader   *bufio.Reader
	scope    Scope
	mu       sync.Mutex
	sequence int
	pending  *Intent
	failed   bool
}

func NewAuditClient(channel io.ReadWriteCloser, scope Scope) (*AuditClient, error) {
	if channel == nil || !validScope(scope, time.Now()) {
		return nil, ErrAudit
	}
	return &AuditClient{channel: channel, reader: bufio.NewReaderSize(channel, MaxAuditFrameBytes), scope: scope}, nil
}

func (client *AuditClient) RecordIntent(ctx context.Context, intent Intent) error {
	if !client.mu.TryLock() {
		return ErrBusy
	}
	defer client.mu.Unlock()
	if client.pending != nil || intent.Scope != client.scope || !ValidAuditIntent(intent, time.Now()) {
		return client.fail()
	}
	if err := client.exchange(ctx, AuditFrame{Version: 1, Stage: "intent", Intent: intent}); err != nil {
		return err
	}
	client.pending = &intent
	return nil
}

func (client *AuditClient) RecordOutcome(ctx context.Context, intent Intent, outcome Outcome) error {
	if !client.mu.TryLock() {
		return ErrBusy
	}
	defer client.mu.Unlock()
	if client.pending == nil || *client.pending != intent || !ValidAuditOutcome(outcome) {
		return client.fail()
	}
	if err := client.exchange(ctx, AuditFrame{Version: 1, Stage: "outcome", Intent: intent, Outcome: &outcome}); err != nil {
		return err
	}
	client.pending = nil
	return nil
}

func (client *AuditClient) fail() error {
	client.failed = true
	_ = client.channel.Close()
	return ErrAudit
}

func (client *AuditClient) exchange(parent context.Context, frame AuditFrame) error {
	if client.failed || parent == nil || parent.Err() != nil || client.sequence >= 2*MaxCalls {
		return client.fail()
	}
	ctx, cancel := context.WithDeadline(parent, client.scope.Deadline)
	defer cancel()
	if ctx.Err() != nil {
		return client.fail()
	}
	stop := context.AfterFunc(ctx, func() { _ = client.channel.Close() })
	defer stop()
	client.sequence++
	frame.Sequence = client.sequence
	expected, err := auditFrameDigest(frame)
	if err != nil || auditWrite(client.channel, frame) != nil {
		return client.fail()
	}
	var ack auditAck
	if auditRead(client.reader, &ack) != nil || ack.Version != 1 || ack.Digest != expected || !ack.Accepted || ctx.Err() != nil {
		return client.fail()
	}
	if !stop() {
		return client.fail()
	}
	return nil
}

// ServeAudit acknowledges only after the supplied controller auditor returns
// durable success. It rejects reordered/repeated calls and never invokes an
// executor. Scope and peer authentication come from trusted job registration,
// not a frame. EOF/ambiguity is failure; pending intents are not auto-replayed.
func ServeAudit(parent context.Context, channel io.ReadWriteCloser, scope Scope, auditor Auditor) error {
	if parent == nil || channel == nil || auditor == nil || !validScope(scope, time.Now()) {
		return ErrAudit
	}
	ctx, cancel := context.WithDeadline(parent, scope.Deadline)
	defer cancel()
	defer channel.Close()
	stop := context.AfterFunc(ctx, func() { _ = channel.Close() })
	defer stop()
	reader := bufio.NewReaderSize(channel, MaxAuditFrameBytes)
	seen := map[string]bool{}
	var pending *Intent
	for sequence := 1; sequence <= 2*MaxCalls; sequence++ {
		var frame AuditFrame
		if auditRead(reader, &frame) != nil || ctx.Err() != nil || frame.Version != 1 || frame.Sequence != sequence || frame.Intent.Scope != scope {
			return ErrAudit
		}
		var err error
		switch frame.Stage {
		case "intent":
			if pending != nil || frame.Outcome != nil || seen[frame.Intent.CallID] || !ValidAuditIntent(frame.Intent, time.Now()) {
				return ErrAudit
			}
			seen[frame.Intent.CallID] = true
			pending = &frame.Intent
			err = auditor.RecordIntent(ctx, frame.Intent)
		case "outcome":
			if pending == nil || *pending != frame.Intent || frame.Outcome == nil || !ValidAuditOutcome(*frame.Outcome) {
				return ErrAudit
			}
			err = auditor.RecordOutcome(ctx, frame.Intent, *frame.Outcome)
			pending = nil
		default:
			return ErrAudit
		}
		digest, digestErr := auditFrameDigest(frame)
		if digestErr != nil || ctx.Err() != nil {
			return ErrAudit
		}
		if auditWrite(channel, auditAck{Version: 1, Digest: digest, Accepted: err == nil}) != nil || err != nil {
			return ErrAudit
		}
	}
	return ErrAudit // consumed budget closes the link; no fresh authority implied
}
