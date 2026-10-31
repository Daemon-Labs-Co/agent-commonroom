package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// This protocol requires an already authenticated, supervisor-owned stream pinned
// to one creator and intent. Never attach it to model pipes, expose its socket to
// containers, or infer authentication from decoded metadata. It transports no
// source bytes and provides no host execution or reconnect fallback.
const captureAuthorityFrameLimit = 128 * 1024

type captureAuthorityRequest struct {
	Version   int                  `json:"version"`
	Sequence  uint64               `json:"sequence"`
	Operation string               `json:"operation"`
	Intent    workerCaptureIntent  `json:"intent"`
	Event     workerProbeEvent     `json:"event"`
	Record    *workerCaptureRecord `json:"record"`
}
type captureAuthorityResponse struct {
	Version  int    `json:"version"`
	Sequence uint64 `json:"sequence"`
	Accepted bool   `json:"accepted"`
}

func writeCaptureAuthorityFrame(w io.Writer, value any) error {
	raw, err := json.Marshal(value)
	if err != nil || len(raw) == 0 || len(raw) > captureAuthorityFrameLimit {
		return errWorkerProbeJournal
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(raw)))
	for _, part := range [][]byte{header[:], raw} {
		for len(part) > 0 {
			n, err := w.Write(part)
			if err != nil || n <= 0 || n > len(part) {
				return errWorkerProbeJournal
			}
			part = part[n:]
		}
	}
	return nil
}
func readCaptureAuthorityFrame(r io.Reader, value any) error {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return errWorkerProbeJournal
	}
	n := binary.BigEndian.Uint32(header[:])
	if n == 0 || n > captureAuthorityFrameLimit {
		return errWorkerProbeJournal
	}
	raw := make([]byte, int(n))
	if _, err := io.ReadFull(r, raw); err != nil || json.Unmarshal(raw, value) != nil {
		return errWorkerProbeJournal
	}
	canonical, err := json.Marshal(value)
	if err != nil || !bytes.Equal(raw, canonical) {
		return errWorkerProbeJournal
	}
	return nil
}
func captureAuthorityBinding(r workerJobRecord, i workerCaptureIntent) bool {
	if !retainedObservationRecord(r) || !validWorkerCaptureRecord(workerCaptureRecord{Version: 1, Intent: i, Events: []workerProbeEvent{{Stage: "begin"}}}) {
		return false
	}
	b := i.Request.Binding
	v := r.Pair.Workspace.Identity
	return b.Execution.JobID == r.JobID && b.Execution.RepositoryDigest == r.Pair.Repository && b.Execution.RoomID == r.RoomID && b.VolumeName == v.Name && b.VolumeNonce == v.Nonce && b.VolumeCreatedAt == v.CreatedAt
}
func captureAuthorityEvent(e workerProbeEvent) bool {
	// Complete history is independently checked by the acknowledgment guard. These
	// are wire-level shapes only; a valid shape grants no transition authority.
	switch e.Stage {
	case "begin":
		return e.ID == "" && !e.Absent
	case "bound", "recovery":
		return probeHex(e.ID, 32) && !e.Absent
	case "finish":
		return (e.ID == "" && !e.Absent) || probeHex(e.ID, 32)
	case "recovery-finish":
		return probeHex(e.ID, 32) && e.Absent
	}
	return false
}
func captureAuthorityRecord(r workerCaptureRecord, i workerCaptureIntent, e workerProbeEvent) bool {
	return validWorkerCaptureRecord(r) && captureIntentEqual(r.Intent, i) && len(r.Events) > 0 && r.Events[len(r.Events)-1] == e && captureAuthorityEvent(e)
}

// Any uncertain exchange poisons this one-shot stream and stops local authority.
// The supplied stop callback must be local and nonblocking, never reenter the link.
type captureAuthorityClient struct {
	mu       sync.Mutex
	stream   io.ReadWriteCloser
	record   workerJobRecord
	intent   workerCaptureIntent
	sequence uint64
	failed   bool
	stop     func()
	once     sync.Once
	notify   sync.Once
	lifetime context.Context
	cancel   context.CancelFunc
	stopped  atomic.Bool
}

func newCaptureAuthorityClient(stream io.ReadWriteCloser, r workerJobRecord, i workerCaptureIntent, stop func()) (*captureAuthorityClient, error) {
	if nilWorkerObservationDependency(stream) || stop == nil || !captureAuthorityBinding(r, i) {
		return nil, errWorkerProbeJournal
	}
	lifetime, cancel := context.WithCancel(context.Background())
	c := &captureAuthorityClient{record: detachedProbeWorkerRecord(r), intent: copyCaptureIntent(i), stop: stop, lifetime: lifetime, cancel: cancel}
	c.stream = newWatchedWorkerAuditLink(lifetime, stream, c.markStopped)
	return c, nil
}

// The sole bounded raw-stream pump revokes idle authority on EOF/error. This
// callback must never join its own pump or reenter the private control stream.
func (c *captureAuthorityClient) markStopped() {
	c.stopped.Store(true)
	c.cancel()
	c.notify.Do(c.stop)
}
func (c *captureAuthorityClient) CaptureAuthorityLifetime() context.Context {
	if c == nil {
		return nil
	}
	return c.lifetime
}
func (c *captureAuthorityClient) StopCaptureAuthority() {
	if c != nil {
		c.once.Do(func() { _ = c.stream.Close(); c.markStopped() })
	}
}
func (c *captureAuthorityClient) exchange(ctx context.Context, operation string, e workerProbeEvent, r *workerCaptureRecord) error {
	if c == nil || ctx == nil {
		return errWorkerProbeJournal
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failed || c.stopped.Load() {
		return errWorkerProbeJournal
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	stopLifetime := context.AfterFunc(c.lifetime, cancel)
	defer stopLifetime()
	fail := func() error { c.failed = true; c.StopCaptureAuthority(); return errWorkerProbeJournal }
	if ctx.Err() != nil {
		return fail()
	}
	c.sequence++
	stopIO := context.AfterFunc(ctx, c.StopCaptureAuthority)
	defer stopIO()
	q := captureAuthorityRequest{Version: 1, Sequence: c.sequence, Operation: operation, Intent: c.intent, Event: e, Record: r}
	var response captureAuthorityResponse
	if writeCaptureAuthorityFrame(c.stream, q) != nil || readCaptureAuthorityFrame(c.stream, &response) != nil || response.Version != 1 || response.Sequence != c.sequence || !response.Accepted || ctx.Err() != nil || c.stopped.Load() {
		return fail()
	}
	return nil
}
func (c *captureAuthorityClient) AuthorizeCapture(ctx context.Context, r workerJobRecord, i workerCaptureIntent, e workerProbeEvent) error {
	if c == nil {
		return errWorkerProbeJournal
	}
	a, _ := json.Marshal(r)
	b, _ := json.Marshal(c.record)
	if !bytes.Equal(a, b) || !captureIntentEqual(i, c.intent) || !captureAuthorityEvent(e) {
		c.StopCaptureAuthority()
		return errWorkerProbeJournal
	}
	return c.exchange(ctx, "authorize", e, nil)
}
func (c *captureAuthorityClient) AcknowledgeCapture(ctx context.Context, r workerCaptureRecord, e workerProbeEvent) error {
	if c == nil {
		return errWorkerProbeJournal
	}
	if !captureAuthorityRecord(r, c.intent, e) {
		c.StopCaptureAuthority()
		return errWorkerProbeJournal
	}
	copy := copyCaptureRecord(r)
	return c.exchange(ctx, "acknowledge", e, &copy)
}

// The guard supplies authenticated current permission, physical/quota evidence and
// durable controller acknowledgment. It must never reenter the worker admission
// lock or call this stream while the worker is waiting for its response. No guard
// is synthesized here. Session failure revokes controller authority exactly once.
func serveCaptureAuthority(ctx context.Context, stream io.ReadWriteCloser, record workerJobRecord, intent workerCaptureIntent, guard workerCaptureJournalGuard) error {
	if ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(stream) || nilWorkerObservationDependency(guard) || !captureAuthorityBinding(record, intent) {
		return errWorkerProbeJournal
	}
	defer stream.Close()
	var once sync.Once
	stop := func() { once.Do(guard.StopCaptureAuthority) }
	defer stop()
	record = detachedProbeWorkerRecord(record)
	intent = copyCaptureIntent(intent)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	stopIO := context.AfterFunc(ctx, func() { _ = stream.Close() })
	defer stopIO()
	for sequence := uint64(1); sequence <= 64; sequence++ {
		var q captureAuthorityRequest
		if readCaptureAuthorityFrame(stream, &q) != nil || q.Version != 1 || q.Sequence != sequence || !captureIntentEqual(q.Intent, intent) || !captureAuthorityEvent(q.Event) || ctx.Err() != nil {
			return errWorkerProbeJournal
		}
		var err error
		switch q.Operation {
		case "authorize":
			if q.Record != nil {
				return errWorkerProbeJournal
			}
			err = guard.AuthorizeCapture(ctx, detachedProbeWorkerRecord(record), copyCaptureIntent(intent), q.Event)
		case "acknowledge":
			if q.Record == nil || !captureAuthorityRecord(*q.Record, intent, q.Event) {
				return errWorkerProbeJournal
			}
			err = guard.AcknowledgeCapture(ctx, copyCaptureRecord(*q.Record), q.Event)
		default:
			return errWorkerProbeJournal
		}
		if err != nil || ctx.Err() != nil || writeCaptureAuthorityFrame(stream, captureAuthorityResponse{Version: 1, Sequence: sequence, Accepted: true}) != nil {
			return errWorkerProbeJournal
		}
	}
	return errWorkerProbeJournal
}
