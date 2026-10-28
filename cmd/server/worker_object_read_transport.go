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
const objectReadAuthorityFrameLimit = 32 * 1024

type objectReadAuthorityRequest struct {
	Protocol  string                  `json:"protocol"`
	Version   int                     `json:"version"`
	Sequence  uint64                  `json:"sequence"`
	Operation string                  `json:"operation"`
	Intent    workerObjectReadIntent  `json:"intent"`
	Event     workerProbeEvent        `json:"event"`
	Record    *workerObjectReadRecord `json:"record"`
}
type objectReadAuthorityResponse struct {
	Protocol string `json:"protocol"`
	Version  int    `json:"version"`
	Sequence uint64 `json:"sequence"`
	Accepted bool   `json:"accepted"`
}

func writeObjectReadAuthorityFrame(w io.Writer, value any) error {
	raw, err := json.Marshal(value)
	if err != nil || len(raw) == 0 || len(raw) > objectReadAuthorityFrameLimit {
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
func readObjectReadAuthorityFrame(r io.Reader, value any) error {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return errWorkerProbeJournal
	}
	n := binary.BigEndian.Uint32(header[:])
	if n == 0 || n > objectReadAuthorityFrameLimit {
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
func objectReadAuthorityBinding(r workerJobRecord, i workerObjectReadIntent) bool {
	if !retainedObservationRecord(r) || !validWorkerObjectReadRecord(workerObjectReadRecord{Version: 1, Intent: i, Events: []workerProbeEvent{{Stage: "begin"}}}) {
		return false
	}
	scope := i.Request.Receipt.Intent.Scope
	return scope.JobID == r.JobID && scope.RepositoryDigest == r.Pair.Repository && scope.RoomID == r.RoomID
}
func objectReadAuthorityEvent(e workerProbeEvent) bool { return captureAuthorityEvent(e) }

// The object-read-v1 discriminator prevents cross-protocol frame substitution.
// Any uncertain exchange poisons this one-shot stream and stops local authority.
// The supplied stop callback must be local and nonblocking, never reenter the link.
type objectReadAuthorityClient struct {
	mu       sync.Mutex
	stream   io.ReadWriteCloser
	record   workerJobRecord
	intent   workerObjectReadIntent
	sequence uint64
	failed   bool
	stop     func()
	once     sync.Once
	notify   sync.Once
	lifetime context.Context
	cancel   context.CancelFunc
	stopped  atomic.Bool
}

func newObjectReadAuthorityClient(stream io.ReadWriteCloser, r workerJobRecord, i workerObjectReadIntent, stop func()) (*objectReadAuthorityClient, error) {
	if nilWorkerObservationDependency(stream) || stop == nil || !objectReadAuthorityBinding(r, i) {
		return nil, errWorkerProbeJournal
	}
	lifetime, cancel := context.WithCancel(context.Background())
	c := &objectReadAuthorityClient{record: detachedProbeWorkerRecord(r), intent: i, stop: stop, lifetime: lifetime, cancel: cancel}
	c.stream = newWatchedWorkerAuditLink(lifetime, stream, c.markStopped)
	return c, nil
}

// EOF must revoke authority even between exchanges. The existing bounded private
// read pump is the sole raw-stream reader; it cannot race frame consumers. Its
// callback never joins its own pump or reenters the client.
func (c *objectReadAuthorityClient) markStopped() {
	c.stopped.Store(true)
	c.cancel()
	c.notify.Do(c.stop)
}
func (c *objectReadAuthorityClient) ObjectReadAuthorityLifetime() context.Context {
	if c == nil {
		return nil
	}
	return c.lifetime
}
func (c *objectReadAuthorityClient) StopObjectReadAuthority() {
	if c != nil {
		c.once.Do(func() { _ = c.stream.Close(); c.markStopped() })
	}
}
func (c *objectReadAuthorityClient) exchange(ctx context.Context, operation string, e workerProbeEvent, r *workerObjectReadRecord) error {
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
	fail := func() error { c.failed = true; c.StopObjectReadAuthority(); return errWorkerProbeJournal }
	if ctx.Err() != nil {
		return fail()
	}
	c.sequence++
	stopIO := context.AfterFunc(ctx, c.StopObjectReadAuthority)
	defer stopIO()
	q := objectReadAuthorityRequest{Protocol: "object-read-v1", Version: 1, Sequence: c.sequence, Operation: operation, Intent: c.intent, Event: e, Record: r}
	var response objectReadAuthorityResponse
	if writeObjectReadAuthorityFrame(c.stream, q) != nil || readObjectReadAuthorityFrame(c.stream, &response) != nil || response.Protocol != "object-read-v1" || response.Version != 1 || response.Sequence != c.sequence || !response.Accepted || ctx.Err() != nil {
		return fail()
	}
	return nil
}
func (c *objectReadAuthorityClient) AuthorizeObjectRead(ctx context.Context, r workerJobRecord, i workerObjectReadIntent, e workerProbeEvent) error {
	if c == nil {
		return errWorkerProbeJournal
	}
	a, _ := json.Marshal(r)
	b, _ := json.Marshal(c.record)
	if !bytes.Equal(a, b) || i != c.intent || !objectReadAuthorityEvent(e) {
		c.StopObjectReadAuthority()
		return errWorkerProbeJournal
	}
	return c.exchange(ctx, "authorize", e, nil)
}
func (c *objectReadAuthorityClient) AcknowledgeObjectRead(ctx context.Context, r workerObjectReadRecord, e workerProbeEvent) error {
	if c == nil {
		return errWorkerProbeJournal
	}
	if !objectReadAuthorityRecord(r, c.intent, e) {
		c.StopObjectReadAuthority()
		return errWorkerProbeJournal
	}
	copy := copyObjectReadRecord(r)
	return c.exchange(ctx, "acknowledge", e, &copy)
}

// The guard supplies authenticated current permission, physical/quota evidence and
// durable controller acknowledgment. It must never reenter the worker admission
// lock or call this stream while the worker is waiting for its response. No guard
// is synthesized here. Session failure revokes controller authority exactly once.
func serveObjectReadAuthority(ctx context.Context, stream io.ReadWriteCloser, record workerJobRecord, intent workerObjectReadIntent, guard workerObjectReadJournalGuard) error {
	if ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(stream) || nilWorkerObservationDependency(guard) || !objectReadAuthorityBinding(record, intent) {
		return errWorkerProbeJournal
	}
	defer stream.Close()
	var once sync.Once
	stop := func() { once.Do(guard.StopObjectReadAuthority) }
	defer stop()
	record = detachedProbeWorkerRecord(record)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	stopIO := context.AfterFunc(ctx, func() { _ = stream.Close() })
	defer stopIO()
	for sequence := uint64(1); sequence <= 64; sequence++ {
		var q objectReadAuthorityRequest
		if readObjectReadAuthorityFrame(stream, &q) != nil || q.Protocol != "object-read-v1" || q.Version != 1 || q.Sequence != sequence || q.Intent != intent || !objectReadAuthorityEvent(q.Event) || ctx.Err() != nil {
			return errWorkerProbeJournal
		}
		var err error
		switch q.Operation {
		case "authorize":
			if q.Record != nil {
				return errWorkerProbeJournal
			}
			err = guard.AuthorizeObjectRead(ctx, detachedProbeWorkerRecord(record), intent, q.Event)
		case "acknowledge":
			if q.Record == nil || !objectReadAuthorityRecord(*q.Record, intent, q.Event) {
				return errWorkerProbeJournal
			}
			err = guard.AcknowledgeObjectRead(ctx, copyObjectReadRecord(*q.Record), q.Event)
		default:
			return errWorkerProbeJournal
		}
		if err != nil || ctx.Err() != nil || writeObjectReadAuthorityFrame(stream, objectReadAuthorityResponse{Protocol: "object-read-v1", Version: 1, Sequence: sequence, Accepted: true}) != nil {
			return errWorkerProbeJournal
		}
	}
	return errWorkerProbeJournal
}
