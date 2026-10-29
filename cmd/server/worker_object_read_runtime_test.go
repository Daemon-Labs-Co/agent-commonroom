package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
)

type assembledObjectReadFixture struct {
	*physicalObjectReadFixture
	t        *testing.T
	planArgs []string
	mode     string
	process  *assembledObjectReadProcess
	release  *controllerArtifactReadSession
	onStart  func()
}

func (f *assembledObjectReadFixture) Run(ctx context.Context, args []string, in []byte, out io.Writer) error {
	switch args[0] {
	case "create":
		records, err := readWorkerObjectReadRecords(f.probe.dir)
		r, ok := records[objectReadRecordName(f.intent)]
		if err != nil || !ok || r.Events[len(r.Events)-1].Stage != "begin" || !slices.Equal(args, f.planArgs) {
			f.t.Fatal("create before durable bound policy")
		}
		f.phase = "created"
		_, err = io.WriteString(out, f.probe.id)
		return err
	case "rm":
		if !slices.Equal(args, []string{"rm", "--force", f.probe.id}) {
			f.t.Fatal("foreign cleanup")
		}
		if f.process != nil && !f.process.closed.Load() {
			f.t.Fatal("cleanup before process joined")
		}
		if f.mode == "cleanup" {
			return errors.New("synthetic cleanup failure")
		}
		f.phase = "absent"
		return nil
	}
	return f.physicalObjectReadFixture.Run(ctx, args, in, out)
}
func (f *assembledObjectReadFixture) StartObjectRead(ctx context.Context, id string) (toolbroker.ObjectReadProcess, error) {
	if id != f.probe.id || f.phase != "created" {
		f.t.Fatal("unbound launch")
	}
	records, err := readWorkerObjectReadRecords(f.probe.dir)
	if err != nil || records[objectReadRecordName(f.intent)].Events[1] != (workerProbeEvent{Stage: "bound", ID: id}) {
		f.t.Fatal("start before durable ID")
	}
	left, right := net.Pipe()
	p := &assembledObjectReadProcess{Conn: left, f: f, done: make(chan error, 1)}
	f.process = p
	f.phase = "running"
	go func() {
		p.done <- objectcustody.ServeReadChannel(ctx, right, f.intent.Request.Receipt.Intent.Scope, f.intent.Request.Receipt, syntheticAssembledObjectSource{})
	}()
	if f.onStart != nil {
		f.onStart()
	}
	return p, nil
}

type syntheticAssembledObjectSource struct{}

func (syntheticAssembledObjectSource) Get(ctx context.Context, _ objectcustody.Scope, _ objectcustody.Receipt) ([]byte, error) {
	return []byte("synthetic reader bytes"), ctx.Err()
}

type assembledObjectReadProcess struct {
	net.Conn
	f              *assembledObjectReadFixture
	done           chan error
	waited, closed atomic.Bool
}

func (p *assembledObjectReadProcess) Wait(ctx context.Context) error {
	if !p.waited.CompareAndSwap(false, true) {
		return errors.New("duplicate wait")
	}
	err := <-p.done
	p.f.phase = "exited"
	if p.f.mode == "exit" {
		return errors.New("synthetic unsuccessful exit")
	}
	return errors.Join(err, ctx.Err())
}
func (p *assembledObjectReadProcess) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil
	}
	_ = p.Conn.Close()
	if !p.waited.Load() {
		<-p.done
	}
	if p.f.mode == "close" {
		return errors.New("synthetic close failure")
	}
	return nil
}
func TestWorkerObjectReadRuntimeControllerBridge(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"healthy", "cleanup", "exit", "close", "release denied", "release revoked", "audit failure", "idle authority loss", "creator loss"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, _ := objectReadSessionFixture(t)
			dir := t.TempDir()
			if err := os.Chmod(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := writeWorkerRecord(dir, g.record, true); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, g.record.JobID+".cancel"), []byte("cancel"), 0600); err != nil {
				t.Fatal(err)
			}
			lease, free, err := workerTryLease(dir, ".job-"+g.record.JobID+".lease")
			if err != nil || !free {
				t.Fatal(err)
			}
			defer lease.Close()
			p := &assembledProbeFixture{t: t, dir: dir, record: detachedProbeWorkerRecord(g.record), id: strings.Repeat("1", 64)}
			f := &assembledObjectReadFixture{physicalObjectReadFixture: &physicalObjectReadFixture{probe: p, intent: g.intent, phase: "vacant"}, t: t, mode: mode}
			creator, stopCreator := context.WithCancel(ctx)
			defer stopCreator()
			l := &workerWorkspaceObservation{ownership: lease, active: true, lifetime: creator, dir: dir, dockerRoot: syntheticObservationDockerRoot, record: detachedProbeWorkerRecord(g.record), runner: f}
			witness := &syntheticObjectReadWitness{}
			evidence, err := newWorkspaceObjectReadEvidence(l, g.worker, g.intent, witness)
			if err != nil {
				t.Fatal(err)
			}
			g.evidence = evidence
			releaseEvidence := &syntheticArtifactGuard{}
			calls := 0
			releaseEvidence.release = func(context.Context, room.State, objectcustody.Receipt) error {
				calls++
				if mode == "release denied" || mode == "release revoked" && calls > 1 {
					return errors.New("synthetic export denial")
				}
				return nil
			}
			release, err := newControllerArtifactReadSession(ctx, g.server, g.worker, g.intent.Request.Receipt, releaseEvidence)
			if err != nil {
				t.Fatal(err)
			}
			f.release = release
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			clientBinding, serverBinding := boundObjectReadFixture(t, g)
			listener := newSyntheticReadListener()
			done := make(chan error, 1)
			go func() { done <- serveObjectReadControlOnce(ctx, listener, serverBinding, g); close(done) }()
			dialer := syntheticReadDialer(func(context.Context, string, string) (net.Conn, error) {
				listener.incoming <- right
				return left, nil
			})
			endpoint := objectReadControlEndpoint{Address: "192.0.2.10:443", Binding: clientBinding}
			c, err := connectObjectReadControl(ctx, endpoint, g.record, g.intent, func() {}, dialer)
			if err != nil {
				_ = left.Close()
				<-done
				t.Fatal(err)
			}
			defer func() { c.StopObjectReadAuthority(); <-done }()
			runtime, err := newWorkerObjectReadRuntime(l, g.intent, f, c, release)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := toolbroker.NewObjectReadContainerPlan(g.intent.Request, g.intent.Image, l.dockerRoot, g.intent.Objects)
			if err != nil {
				t.Fatal(err)
			}
			f.planArgs = plan.CreateArgs()
			if mode == "creator loss" {
				f.onStart = func() { stopCreator(); <-creator.Done() }
			}
			if mode == "idle authority loss" {
				f.onStart = func() { _ = right.Close(); <-c.ObjectReadAuthorityLifetime().Done() }
			}
			if mode == "audit failure" {
				db.failAt = "commit"
			}
			raw, err := runtime.Run(ctx)
			db.failAt = ""
			if mode == "audit failure" {
				if !g.server.integrityFault {
					t.Fatal("persistence failure not latched")
				}
				g.server.integrityFault = false
			}
			if mode == "healthy" {
				if err != nil || string(raw) != "synthetic reader bytes" {
					t.Fatal("healthy isolated bridge", string(raw), err)
				}
			} else if err == nil || raw != nil {
				t.Fatal("unsafe bytes released", mode, err)
			}
			if f.process != nil && !f.process.closed.Load() {
				t.Fatal("process not joined")
			}
			if mode == "healthy" {
				if n, e := workerObjectReadSlots(dir); e != nil || n != 0 {
					t.Fatal(n, e)
				}
				if n, e := controllerObjectReadSlots(artifactState(t, g.server, ctx)); e != nil || n != 0 {
					t.Fatal(n, e)
				}
				if witness.operations == 0 || witness.lifecycles == 0 {
					t.Fatal("physical lifecycle evidence bypassed")
				}
				if _, e := runtime.Run(ctx); e == nil {
					t.Fatal("one-shot reader reused")
				}
			} else if mode == "cleanup" || mode == "idle authority loss" {
				if n, e := workerObjectReadSlots(dir); e != nil || n != 1 {
					t.Fatal("cleanup failure released local capacity", n, e)
				}
				if n, e := controllerObjectReadSlots(artifactState(t, g.server, ctx)); e != nil || n != 1 {
					t.Fatal("cleanup failure released controller capacity", n, e)
				}
			}
			st := artifactState(t, g.server, ctx)
			if _, ok := st.ManagedStorage.Releases[g.intent.Request.Receipt.Intent.ObjectID]; ok {
				t.Fatal("read refunded retained object")
			}
		})
	}
}

func TestWorkerObjectReadRuntimeRequiresBoundary(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"owner", "runner", "starter", "guard", "release", "closed"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g, _, ctx, _ := objectReadSessionFixture(t)
			dir := t.TempDir()
			if err := os.Chmod(dir, 0700); err != nil {
				t.Fatal(err)
			}
			lease, free, err := workerTryLease(dir, ".job-"+g.record.JobID+".lease")
			if err != nil || !free {
				t.Fatal(err)
			}
			defer lease.Close()
			fake := &assembledObjectReadFixture{}
			var runner toolbroker.DockerRunner = fake
			var starter toolbroker.ObjectReadStarter = fake
			var guard workerObjectReadRuntimeGuard = g
			release, e := newControllerArtifactReadSession(ctx, g.server, g.worker, g.intent.Request.Receipt, &syntheticArtifactGuard{})
			if e != nil {
				t.Fatal(e)
			}
			var releaseGuard objectcustody.ReadReleaseGuard = release
			l := &workerWorkspaceObservation{ownership: lease, active: true, lifetime: ctx, dir: dir, dockerRoot: syntheticObservationDockerRoot, record: detachedProbeWorkerRecord(g.record), runner: runner}
			switch mode {
			case "owner":
				l = nil
			case "runner":
				l.runner = (*assembledObjectReadFixture)(nil)
			case "starter":
				starter = (*assembledObjectReadFixture)(nil)
			case "guard":
				guard = (*controllerObjectReadSession)(nil)
			case "release":
				releaseGuard = (*controllerArtifactReadSession)(nil)
			case "closed":
				l.active = false
			}
			if _, err := newWorkerObjectReadRuntime(l, g.intent, starter, guard, releaseGuard); err == nil {
				t.Fatal("missing isolation dependency accepted")
			}
		})
	}
}

func TestWorkerObjectReadRecoveryControllerBridge(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"present", "already absent", "cleanup retry", "denied", "wrong ID", "unknown ID", "foreign ID", "editing off"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			initial, db, ctx, _ := objectReadSessionFixture(t)
			dir := t.TempDir()
			if err := os.Chmod(dir, 0700); err != nil {
				t.Fatal(err)
			}
			record, intent := detachedProbeWorkerRecord(initial.record), initial.intent
			if err := writeWorkerRecord(dir, record, true); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, record.JobID+".cancel"), []byte("cancel"), 0600); err != nil {
				t.Fatal(err)
			}
			lease, free, err := workerTryLease(dir, ".job-"+record.JobID+".lease")
			if err != nil || !free {
				t.Fatal(err)
			}
			defer lease.Close()
			probe := &assembledProbeFixture{t: t, dir: dir, record: record, id: strings.Repeat("1", 64)}
			f := &assembledObjectReadFixture{physicalObjectReadFixture: &physicalObjectReadFixture{probe: probe, intent: intent, phase: "vacant"}, t: t}
			l := &workerWorkspaceObservation{ownership: lease, active: true, lifetime: ctx, dir: dir, dockerRoot: syntheticObservationDockerRoot, record: record, runner: f}
			witness := &syntheticObjectReadWitness{}
			evidence, err := newWorkspaceObjectReadEvidence(l, initial.worker, intent, witness)
			if err != nil {
				t.Fatal(err)
			}
			open := func() (*controllerObjectReadSession, *objectReadAuthorityClient) {
				session, err := newControllerObjectReadSession(ctx, initial.server, initial.worker, record, intent, evidence)
				if err != nil {
					t.Fatal(err)
				}
				left, right := net.Pipe()
				c, err := newObjectReadAuthorityClient(left, record, intent, func() {})
				if err != nil {
					t.Fatal(err)
				}
				done := make(chan error, 1)
				go func() { defer close(done); done <- serveObjectReadAuthority(ctx, right, record, intent, session) }()
				t.Cleanup(func() { c.StopObjectReadAuthority(); <-done })
				return session, c
			}
			_, c := open()
			journal, err := newWorkerObjectReadJournal(l, intent.Request, intent.Image, intent.Objects, c)
			if err != nil {
				t.Fatal(err)
			}
			if journal.BeginObjectRead(ctx, intent.Request, intent.Objects, intent.Image) != nil {
				t.Fatal("begin")
			}
			id := probe.id
			if mode != "unknown ID" {
				f.phase = "created"
				if journal.BindObjectRead(ctx, intent.Request, id) != nil {
					t.Fatal("bind")
				}
			} else {
				id = ""
			}
			if journal.FinishObjectRead(ctx, intent.Request, id, false) != nil {
				t.Fatal("uncertain finish")
			}
			c.StopObjectReadAuthority()
			// A new authenticated cleanup session cannot reuse the canceled read session.
			// No automatic stream reconnection or unknown-ID adoption occurs.
			if mode == "already absent" {
				f.phase = "absent"
			}
			if mode == "cleanup retry" {
				f.mode = "cleanup"
			}
			if mode == "denied" {
				witness.deny = true
			}
			if mode == "foreign ID" {
				f.physicalObjectReadFixture.mode = "foreign ID"
			}
			if mode == "editing off" {
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
				defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
			}
			id = probe.id
			if mode == "wrong ID" {
				id = strings.Repeat("9", 64)
			}
			_, fresh := open()
			before := f.calls
			err = recoverRetainedWorkerObjectRead(ctx, l, intent, id, fresh)
			if mode == "cleanup retry" {
				if err == nil {
					t.Fatal("failed cleanup settled")
				}
				if n, e := workerObjectReadSlots(dir); e != nil || n != 1 {
					t.Fatal("failed cleanup lost charge", n, e)
				}
				f.mode = ""
				_, retry := open()
				err = recoverRetainedWorkerObjectRead(ctx, l, intent, id, retry)
			}
			healthy := mode == "present" || mode == "already absent" || mode == "cleanup retry" || mode == "editing off"
			if (err == nil) != healthy {
				t.Fatal(mode, err)
			}
			want := 1
			if healthy {
				want = 0
			}
			if n, e := workerObjectReadSlots(dir); e != nil || n != want {
				t.Fatal("worker capacity", n, e)
			}
			if n, e := controllerObjectReadSlots(artifactState(t, initial.server, ctx)); e != nil || n != want {
				t.Fatal("controller capacity", n, e)
			}
			if (mode == "denied" || mode == "wrong ID" || mode == "unknown ID") && f.calls != before {
				t.Fatal("daemon before fresh known-ID authority")
			}
			if f.process != nil {
				t.Fatal("recovery started a reader")
			}
			st := artifactState(t, initial.server, ctx)
			if st.ArtifactCustody[intent.Request.Receipt.Intent.ObjectID].ReadAcknowledgments != 0 {
				t.Fatal("cleanup exported bytes")
			}
			if _, ok := st.ManagedStorage.Releases[intent.Request.Receipt.Intent.ObjectID]; ok {
				t.Fatal("cleanup refunded retained storage")
			}
		})
	}
}
