package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacecapture"
	"agentroom/internal/workspacemount"
	"agentroom/internal/workspacesnapshot"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// This fixture drives the real plan/runtime/protocol/local lifecycle journal.
// All daemon, frozen-source, custody and controller evidence remains synthetic.
type ownedCaptureFixture struct {
	base                    *physicalCaptureFixture
	guard                   *captureJournalGuard
	phase                   string
	exists                  bool
	mode                    string
	starts, reserves, seals int
	onStart                 func(context.Context)
	onInspect               func(context.Context)
}

func (f *ownedCaptureFixture) Run(ctx context.Context, args []string, in []byte, out io.Writer) error {
	f.base.probe.checkLock()
	switch args[0] {
	case "create":
		plan, err := toolbroker.NewCaptureContainerPlan(f.base.request, f.base.probe.image, syntheticObservationDockerRoot, f.base.objects)
		if err != nil || strings.Join(args, "\x00") != strings.Join(plan.CreateArgs(), "\x00") {
			f.base.probe.t.Fatal("capture launch policy changed")
		}
		f.exists = true
		f.phase = "created"
		if f.mode == "create" {
			return errors.New("synthetic lost creation reply")
		}
		_, err = io.WriteString(out, f.base.probe.id)
		return err
	case "rm":
		if len(args) != 3 || args[1] != "--force" || args[2] != f.base.probe.id {
			f.base.probe.t.Fatal("unsafe removal", args)
		}
		if f.mode == "cleanup" {
			return errors.New("synthetic cleanup failure")
		}
		f.exists = false
		return nil
	case "ps":
		if !f.exists {
			return nil
		}
	case "inspect":
		if f.onInspect != nil {
			f.onInspect(ctx)
		}
		var b bytes.Buffer
		if err := f.base.Run(ctx, args, in, &b); err != nil {
			return err
		}
		var c map[string]any
		if err := json.Unmarshal(b.Bytes(), &c); err != nil {
			return err
		}
		c["Name"] = "/commonroom-capture-" + f.base.request.Binding.Execution.JobID + "-" + f.base.request.Challenge
		state := c["State"].(map[string]any)
		state["Status"] = f.phase
		state["Running"] = f.phase == "running"
		return json.NewEncoder(out).Encode(c)
	}
	return f.base.Run(ctx, args, in, out)
}
func (f *ownedCaptureFixture) StartCapture(ctx context.Context, id string) (toolbroker.CaptureProcess, error) {
	records, err := readWorkerCaptureRecords(f.base.probe.dir)
	local, exists := records[captureRecordName(workerCaptureIntent{Request: f.base.request, Image: f.base.probe.image, Objects: f.base.objects})]
	if err != nil || !exists || id != f.base.probe.id || len(local.Events) != 2 || local.Events[1] != (workerProbeEvent{Stage: "bound", ID: id}) {
		f.base.probe.t.Fatal("start before durable binding")
	}
	f.starts++
	f.phase = "running"
	parent, helper := net.Pipe()
	p := &ownedCaptureProcess{Conn: parent, f: f, done: make(chan error, 1), joined: make(chan struct{})}
	go func() {
		defer close(p.joined)
		defer helper.Close()
		client, err := workspacecapture.NewCaptureChannel(ctx, helper, f.base.request, func() {})
		if err == nil {
			err = client.AuthorizeFrozenSource(ctx, f.base.request.Binding, workspacemount.Proof{Version: 1, Challenge: f.base.request.Challenge, MountID: 2, ReadOnly: true, RootReadOnly: true})
		}
		scope := f.base.request.Binding.Execution
		scope.RevisionDigest = strings.Repeat("7", 64)
		m := f.base.request.Members[0]
		intent := objectcustody.Intent{ObjectID: strings.Repeat("8", 32), Scope: objectcustody.Scope{ExecutionScope: scope, OutputName: m.OutputName, MemberName: m.Name}, MaxBytes: m.MaxBytes}
		receipt := objectcustody.Receipt{Intent: intent, Bytes: 3, Digest: strings.Repeat("9", 64)}
		if err == nil {
			err = client.Reserve(ctx, intent)
		}
		if err == nil {
			err = client.Seal(ctx, receipt)
		}
		if err == nil {
			err = client.Finish(ctx, workspacecapture.CapturedObjects{Snapshot: workspacesnapshot.Snapshot{Scope: scope, Digest: scope.RevisionDigest}, Receipts: []objectcustody.Receipt{receipt}})
		}
		p.done <- err
	}()
	if f.onStart != nil {
		f.onStart(ctx)
	}
	return p, nil
}

type ownedCaptureProcess struct {
	net.Conn
	f      *ownedCaptureFixture
	done   chan error
	joined chan struct{}
}

func (p *ownedCaptureProcess) Close() error {
	err := p.Conn.Close()
	<-p.joined
	return err
}
func (p *ownedCaptureProcess) Wait(ctx context.Context) error {
	select {
	case err := <-p.done:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	p.f.phase = "exited"
	return nil
}
func (f *ownedCaptureFixture) AuthorizeFrozenSource(ctx context.Context, _ workspacesnapshot.Binding, _ workspacemount.Proof) error {
	if f.mode == "authority" {
		return errArtifactCustody
	}
	return ctx.Err()
}
func (f *ownedCaptureFixture) Reserve(ctx context.Context, _ objectcustody.Intent) error {
	f.reserves++
	return ctx.Err()
}
func (f *ownedCaptureFixture) Seal(ctx context.Context, _ objectcustody.Receipt) error {
	f.seals++
	return ctx.Err()
}
func (f *ownedCaptureFixture) AuthorizeRead(context.Context, objectcustody.Scope, objectcustody.Receipt) error {
	return errArtifactCustody
}
func (f *ownedCaptureFixture) AcknowledgeRead(context.Context, objectcustody.Receipt) error {
	return errArtifactCustody
}
func ownedCaptureSetup(t *testing.T) (*ownedCaptureFixture, *workerWorkspaceObservation, *workerCaptureJournal) {
	t.Helper()
	j, g, l := captureJournalFixtureScope(t)
	f := &ownedCaptureFixture{base: &physicalCaptureFixture{probe: g.f, request: j.intent.Request, objects: j.intent.Objects}, guard: g}
	// The fixture's reviewed image must match the exact journal/plan image.
	f.base.probe.image = j.intent.Image
	l.runner = f
	return f, l, j
}
func TestOwnedCaptureAssemblesLifecycleAndRetainsUncertainty(t *testing.T) {
	for _, mode := range []string{"healthy", "caller mutation", "authority", "create", "cleanup", "ack", "begin"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f, l, j := ownedCaptureSetup(t)
			f.mode = mode
			if mode == "ack" {
				f.guard.failAck = "bound"
			}
			if mode == "begin" {
				f.guard.deny = "begin"
			}
			q := copyCaptureRequest(j.intent.Request)
			capture, err := newOwnedWorkerCapture(l, q, j.intent.Image, j.intent.Objects, f.guard, f, f)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "caller mutation" {
				q.Members[0].Name = "untrusted-replacement.txt"
			}
			report, err := capture.Run(t.Context())
			records, e := readWorkerCaptureRecords(j.dir)
			if e != nil {
				t.Fatal(e)
			}
			if mode == "healthy" || mode == "caller mutation" {
				if err != nil || len(report.Receipts) != 1 || f.starts != 1 || f.reserves != 1 || f.seals != 1 || f.exists || !captureRecordSettled(records[captureRecordName(j.intent)]) {
					t.Fatal("assembled capture failed", err, report, f.guard.seen)
				}
			} else {
				if err == nil || len(report.Receipts) != 0 {
					t.Fatal("uncertain capture reported success", mode)
				}
				if mode == "begin" {
					if len(records) != 0 || f.exists {
						t.Fatal("denied begin created state")
					}
				} else if len(records) != 1 || (captureRecordSettled(records[captureRecordName(j.intent)]) != (mode == "authority")) {
					t.Fatal("uncertainty lost durable reservation", mode, records)
				}
			}
			if second, e := capture.Run(t.Context()); e == nil || len(second.Receipts) != 0 {
				t.Fatal("capture replayed")
			}
		})
	}
}
func TestOwnedCaptureRefusesMissingOrExpiredSetup(t *testing.T) {
	for _, mode := range []string{"nil owner", "inactive", "no starter", "nil guard", "nil authority", "nil custody", "typed nil guard", "typed nil authority", "typed nil custody", "foreign binding", "expired run", "canceled run"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f, l, j := ownedCaptureSetup(t)
			var guard workerCaptureRuntimeGuard = f.guard
			var authority workspacecapture.Authority = f
			var custody objectcustody.Journal = f
			q := copyCaptureRequest(j.intent.Request)
			switch mode {
			case "nil owner":
				l = nil
			case "inactive":
				l.active = false
			case "no starter":
				l.runner = f.base
			case "nil guard":
				guard = nil
			case "nil authority":
				authority = nil
			case "nil custody":
				custody = nil
			case "typed nil guard":
				var absent *captureJournalGuard
				guard = absent
			case "typed nil authority":
				var absent *ownedCaptureFixture
				authority = absent
			case "typed nil custody":
				var absent *ownedCaptureFixture
				custody = absent
			case "foreign binding":
				q.Binding.Execution.JobID = strings.Repeat("0", 32)
			}
			c, e := newOwnedWorkerCapture(l, q, j.intent.Image, j.intent.Objects, guard, authority, custody)
			if mode == "expired run" || mode == "canceled run" {
				if e != nil {
					t.Fatal(e)
				}
				readCtx, cancel := context.WithCancel(t.Context())
				defer cancel()
				if mode == "expired run" {
					l.mu.Lock()
					l.active = false
					l.mu.Unlock()
				} else {
					cancel()
				}
				if r, e := c.Run(readCtx); e == nil || len(r.Receipts) != 0 {
					t.Fatal("expired ownership revived")
				}
			} else if e == nil || c != nil {
				t.Fatal("invalid setup admitted", mode)
			}
			if f.starts != 0 || len(f.guard.seen) != 0 {
				t.Fatal("invalid setup touched lifecycle")
			}
		})
	}
}

// Returning from the owner callback cannot release the creator lock while a
// canceled helper operation is still joining its trusted dependencies.
type blockingCaptureGuard struct{ entered, canceled, release chan struct{} }

func (g *blockingCaptureGuard) AuthorizeCapture(ctx context.Context, _ workerJobRecord, _ workerCaptureIntent, _ workerProbeEvent) error {
	close(g.entered)
	<-ctx.Done()
	close(g.canceled)
	<-g.release
	return ctx.Err()
}
func (g *blockingCaptureGuard) AcknowledgeCapture(context.Context, workerCaptureRecord, workerProbeEvent) error {
	return errArtifactCustody
}
func (g *blockingCaptureGuard) StopCaptureAuthority() {}

// This explicitly synthetic guard tests owner cancellation independently of
// transport loss; production must supply an authenticated cancelable lifetime.
func (g *blockingCaptureGuard) CaptureAuthorityLifetime() context.Context {
	return context.Background()
}
func TestOwnedCaptureTeardownJoinsBeforeCreatorUnlock(t *testing.T) {
	for _, recovery := range []bool{false, true} {
		name := "capture"
		if recovery {
			name = "recovery"
		}
		t.Run(name, func(t *testing.T) { t.Parallel(); checkOwnedCaptureJoin(t, recovery) })
	}
}
func checkOwnedCaptureJoin(t *testing.T, recovery bool) {
	t.Helper()
	f, l, j := ownedCaptureSetup(t)
	if recovery {
		beginCaptureJournal(t, j)
		if err := j.BindCapture(t.Context(), j.intent.Request, f.base.probe.id); err != nil {
			t.Fatal(err)
		}
		if err := j.FinishCapture(t.Context(), j.intent.Request, f.base.probe.id, false); err != nil {
			t.Fatal(err)
		}
	}
	// Release fixture ownership so the real ownership callback acquires its lease.
	l.ownership.Close()
	g := &blockingCaptureGuard{make(chan struct{}), make(chan struct{}), make(chan struct{})}
	done := make(chan error, 1)
	runDone := make(chan error, 1)
	go func() {
		done <- withWorkerWorkspaceOwnership(t.Context(), l.dir, l.record, l.dockerRoot, f, f.base.probe, func(owner *workerWorkspaceObservation) error {
			if recovery {
				go func() { runDone <- recoverRetainedWorkerCapture(t.Context(), owner, j.intent, f.base.probe.id, g) }()
			} else {
				c, e := newOwnedWorkerCapture(owner, j.intent.Request, j.intent.Image, j.intent.Objects, g, f, f)
				if e != nil {
					return e
				}
				go func() { _, e := c.Run(t.Context()); runDone <- e }()
			}
			select {
			case <-g.entered:
				return errors.New("synthetic owner exit")
			case <-time.After(2 * time.Second):
				return errors.New("runtime did not begin")
			}
		})
	}()
	select {
	case <-g.canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("owner did not cancel runtime")
	}
	lease, free, e := workerTryLease(l.dir, ".job-"+l.record.JobID+".lease")
	if lease != nil {
		lease.Close()
	}
	if e != nil || free {
		t.Fatal("creator lock released before join", e)
	}
	select {
	case <-done:
		t.Fatal("owner returned before runtime joined")
	default:
	}
	close(g.release)
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("owner failure lost")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("owner failed to join")
	}
	select {
	case e := <-runDone:
		if e == nil {
			t.Fatal("canceled capture succeeded")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("capture not joined")
	}
	lease, free, e = workerTryLease(l.dir, ".job-"+l.record.JobID+".lease")
	if lease != nil {
		lease.Close()
	}
	if e != nil || !free {
		t.Fatal("creator lock not released after join", e)
	}
}
