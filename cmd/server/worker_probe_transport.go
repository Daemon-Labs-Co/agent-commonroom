package main

import (
	"agentroom/internal/toolbroker"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"sync"
	"time"
)

// probeAuthorityChannel carries only control messages over an already authenticated,
// supervisor-owned stream. Never connect it to model stdin/stdout or mount its
// socket into a container. The caller binds both peers to one exact reviewed intent
// and retained creator record before serving. Authentication and independent physical
// evidence remain mandatory; decoding a frame grants neither. There is no reconnect
// or automatic replay after uncertain acknowledgment, nor any host execution path.
const probeAuthorityFrameLimit = 4096

type probeAuthorityRequest struct {
	Version   int                             `json:"version"`
	Sequence  uint64                          `json:"sequence"`
	Operation string                          `json:"operation"`
	Intent    toolbroker.WorkspaceProbeIntent `json:"intent"`
	Event     workerProbeEvent                `json:"event"`
}
type probeAuthorityResponse struct {
	Version  int    `json:"version"`
	Sequence uint64 `json:"sequence"`
	Accepted bool   `json:"accepted"`
}

func writeProbeAuthorityFrame(w io.Writer, value any) error {
	raw, err := json.Marshal(value)
	if err != nil || len(raw) > probeAuthorityFrameLimit {
		return errWorkerProbeJournal
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(raw)))
	for _, b := range [][]byte{header[:], raw} {
		for len(b) > 0 {
			n, e := w.Write(b)
			if e != nil || n <= 0 || n > len(b) {
				return errWorkerProbeJournal
			}
			b = b[n:]
		}
	}
	return nil
}
func readProbeAuthorityFrame(r io.Reader, value any) error {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return errWorkerProbeJournal
	}
	n := binary.BigEndian.Uint32(header[:])
	if n == 0 || n > probeAuthorityFrameLimit {
		return errWorkerProbeJournal
	}
	raw := make([]byte, int(n))
	if _, err := io.ReadFull(r, raw); err != nil {
		return errWorkerProbeJournal
	}
	if json.Unmarshal(raw, value) != nil {
		return errWorkerProbeJournal
	}
	canonical, err := json.Marshal(value)
	// Canonical encoding rejects duplicate/unknown fields and alternate spellings.
	if err != nil || !bytes.Equal(raw, canonical) {
		return errWorkerProbeJournal
	}
	return nil
}

// probeAuthorityClient serializes a one-shot link, so responses cannot acknowledge
// another request. Any denial, malformed frame or uncertain I/O poisons the stream
// and invokes the supplied local authority stop exactly once. Cancellation closes
// the owned stream, unblocking both peers; callers must not share it with other work.
type probeAuthorityClient struct {
	mu       sync.Mutex
	stream   io.ReadWriteCloser
	intent   toolbroker.WorkspaceProbeIntent
	record   workerJobRecord
	sequence uint64
	failed   bool
	stop     func()
	once     sync.Once
}

func newProbeAuthorityClient(stream io.ReadWriteCloser, record workerJobRecord, intent toolbroker.WorkspaceProbeIntent, stop func()) (*probeAuthorityClient, error) {
	if nilWorkerObservationDependency(stream) || stop == nil || !retainedObservationRecord(record) || !validWorkerProbeIntent(intent) || record.JobID != intent.JobID || record.Pair.Repository != intent.Repository || record.Pair.Workspace.Identity != intent.Volume {
		return nil, errWorkerProbeJournal
	}
	pair := *record.Pair
	workspace := *pair.Workspace
	pair.Workspace = &workspace
	record.Pair = &pair
	return &probeAuthorityClient{stream: stream, intent: intent, record: record, stop: stop}, nil
}
func (c *probeAuthorityClient) StopAuthority() { c.once.Do(func() { c.stream.Close(); c.stop() }) }
func (c *probeAuthorityClient) exchange(ctx context.Context, operation string, event workerProbeEvent) error {
	if c == nil || ctx == nil || ctx.Err() != nil {
		return errWorkerProbeJournal
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if c.failed {
		return errWorkerProbeJournal
	}
	c.sequence++
	cancelIO := context.AfterFunc(ctx, func() { c.StopAuthority() })
	defer cancelIO()
	request := probeAuthorityRequest{1, c.sequence, operation, c.intent, event}
	var response probeAuthorityResponse
	if writeProbeAuthorityFrame(c.stream, request) != nil || readProbeAuthorityFrame(c.stream, &response) != nil || response.Version != 1 || response.Sequence != c.sequence || !response.Accepted || ctx.Err() != nil {
		c.failed = true
		c.StopAuthority()
		return errWorkerProbeJournal
	}
	return nil
}
func (c *probeAuthorityClient) Authorize(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error {
	if c == nil || i != c.intent {
		return errWorkerProbeJournal
	}
	return c.exchange(ctx, "authorize", workerProbeEvent{})
}
func (c *probeAuthorityClient) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	if c == nil || i != c.intent {
		return errWorkerProbeJournal
	}
	return c.exchange(ctx, "acknowledge", e)
}
func (c *probeAuthorityClient) AuthorizeWorkspaceObservation(ctx context.Context, r workerJobRecord) error {
	if c == nil {
		return errWorkerProbeJournal
	}
	a, _ := json.Marshal(r)
	b, _ := json.Marshal(c.record)
	if !bytes.Equal(a, b) {
		return errWorkerProbeJournal
	}
	return c.exchange(ctx, "observe", workerProbeEvent{})
}

// serveProbeAuthority invokes an independently assembled trusted controller guard.
// Its authority comes from the authenticated session, never request-supplied actor,
// worker or room IDs. The fixed record/intent cannot be selected by the peer.
// Guard errors close this link without disclosing diagnostics. This transport does
// not implement physical evidence, create a FrozenLease or release source bytes.
func serveProbeAuthority(ctx context.Context, stream io.ReadWriteCloser, record workerJobRecord, intent toolbroker.WorkspaceProbeIntent, guard workerWorkspaceProbeGuard) error {
	if ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(stream) || nilWorkerObservationDependency(guard) || !retainedObservationRecord(record) || !validWorkerProbeIntent(intent) || record.JobID != intent.JobID || record.Pair.Repository != intent.Repository || record.Pair.Workspace.Identity != intent.Volume {
		return errWorkerProbeJournal
	}
	defer stream.Close()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cancelIO := context.AfterFunc(ctx, func() { stream.Close() })
	defer cancelIO()
	for sequence := uint64(1); ; sequence++ {
		var request probeAuthorityRequest
		if readProbeAuthorityFrame(stream, &request) != nil || request.Version != 1 || request.Sequence != sequence || request.Intent != intent || ctx.Err() != nil {
			return errWorkerProbeJournal
		}
		var err error
		switch request.Operation {
		case "authorize":
			if request.Event != (workerProbeEvent{}) {
				return errWorkerProbeJournal
			}
			err = guard.Authorize(ctx, intent)
		case "observe":
			if request.Event != (workerProbeEvent{}) {
				return errWorkerProbeJournal
			}
			// Detach pointers on each callback; a callback cannot rewrite later bindings.
			copy := record
			pair := *record.Pair
			workspace := *pair.Workspace
			pair.Workspace = &workspace
			copy.Pair = &pair
			err = guard.AuthorizeWorkspaceObservation(ctx, copy)
		case "acknowledge":
			err = guard.Acknowledge(ctx, intent, request.Event)
		default:
			return errWorkerProbeJournal
		}
		if err != nil || ctx.Err() != nil {
			return errWorkerProbeJournal
		}
		if writeProbeAuthorityFrame(stream, probeAuthorityResponse{1, sequence, true}) != nil {
			return errWorkerProbeJournal
		}
	}
}
