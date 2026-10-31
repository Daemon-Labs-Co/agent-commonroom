package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func objectReadProfileFixture(t *testing.T, g *controllerObjectReadSession) (string, objectReadWorkerProfileFile) {
	t.Helper()
	client, server := provisionedObjectReadTestFiles(t)
	file := objectReadWorkerProfileFile{Version: 1, Profiles: []objectReadWorkerProfileSpec{
		{g.intent.Request.Receipt.Intent.Scope.CompanyID, g.worker, "worker", "192.0.2.10:443", client},
		{g.intent.Request.Receipt.Intent.Scope.CompanyID, g.worker, "controller", "192.0.2.10:443", server},
	}}
	path := filepath.Join(t.TempDir(), "profiles.json")
	writeObjectReadProfiles(t, path, file)
	return path, file
}
func writeObjectReadProfiles(t *testing.T, path string, file objectReadWorkerProfileFile) {
	t.Helper()
	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

type addressedReadListener struct {
	*syntheticReadListener
	address string
}

func (l *addressedReadListener) Addr() net.Addr {
	address, _ := net.ResolveTCPAddr("tcp", l.address)
	return address
}
func TestObjectReadWorkerServiceLifecycle(t *testing.T) {
	t.Parallel()
	g, db, ctx, _ := objectReadSessionFixture(t)
	before := len(db.snapshot())
	path, _ := objectReadProfileFixture(t, g)
	service, err := newObjectReadWorkerService(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	listener := &addressedReadListener{newSyntheticReadListener(), "192.0.2.10:443"}
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	done := make(chan error, 1)
	go func() { done <- service.Serve(ctx, listener, g) }()
	dial := syntheticReadDialer(func(context.Context, string, string) (net.Conn, error) { listener.incoming <- b; return a, nil })
	c, err := service.Connect(ctx, g.worker, g.record, g.intent, func() {}, dial)
	if err != nil {
		t.Fatal(err)
	}
	defer c.StopObjectReadAuthority()
	event := workerProbeEvent{Stage: "begin"}
	if err = c.AuthorizeObjectRead(ctx, g.record, g.intent, event); err != nil {
		t.Fatal(err)
	}
	if err = c.AcknowledgeObjectRead(ctx, workerObjectReadRecord{Version: 1, Intent: g.intent, Events: []workerProbeEvent{event}}, event); err != nil {
		t.Fatal(err)
	}
	// Closing service must cancel and join the idle client AND controller.
	closed := make(chan struct{})
	go func() { service.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("service shutdown did not join")
	}
	<-done
	if c.ObjectReadAuthorityLifetime().Err() == nil || g.lifetime.Err() == nil || len(db.snapshot()) != before+1 {
		t.Fatal("shutdown lost audit or retained authority")
	}
	if _, err = service.Connect(ctx, g.worker, g.record, g.intent, func() {}, dial); err == nil {
		t.Fatal("closed service revived")
	}
	st := artifactState(t, g.server, ctx)
	if st.ObjectReads[controllerObjectReadKey(g.intent)].Released {
		t.Fatal("shutdown cleared uncertain capacity")
	}
}
func TestObjectReadWorkerProfilesRefuseInvalidPolicy(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"version", "empty", "duplicate profile", "bad role", "bad endpoint", "unknown key", "duplicate key", "case duplicate", "bad identity", "public policy", "foreign company", "unknown worker"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, _, ctx, _ := objectReadSessionFixture(t)
			path, file := objectReadProfileFixture(t, g)
			switch kind {
			case "version":
				file.Version = 2
			case "empty":
				file.Profiles = nil
			case "duplicate profile":
				file.Profiles = append(file.Profiles, file.Profiles[0])
			case "bad role":
				file.Profiles[0].Role = "model"
			case "bad endpoint":
				file.Profiles[0].Address = "example.test:443"
			case "bad identity":
				file.Profiles[1].Identity.PrivateKey = filepath.Join(t.TempDir(), "absent")
			case "foreign company":
				for i := range file.Profiles {
					file.Profiles[i].CompanyID = "foreign"
				}
			case "unknown worker":
				for i := range file.Profiles {
					file.Profiles[i].Worker = "other-worker"
				}
			}
			writeObjectReadProfiles(t, path, file)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "unknown key":
				raw = []byte(strings.Replace(string(raw), `"version": 1`, `"version": 1, "unknown": true`, 1))
			case "duplicate key":
				raw = []byte(strings.Replace(string(raw), `"version": 1`, `"version": 1, "version": 1`, 1))
			case "case duplicate":
				raw = []byte(strings.Replace(string(raw), `"version": 1`, `"version": 1, "VERSION": 1`, 1))
			}
			if err = os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			if kind == "public policy" {
				if err = os.Chmod(path, 0644); err != nil {
					t.Fatal(err)
				}
			}
			service, err := newObjectReadWorkerService(ctx, path)
			if kind == "foreign company" || kind == "unknown worker" {
				if err != nil {
					t.Fatal(err)
				}
				defer service.Close()
				dial := syntheticReadDialer(func(context.Context, string, string) (net.Conn, error) {
					t.Error("unapproved worker dialed")
					return nil, net.ErrClosed
				})
				if c, err := service.Connect(ctx, g.worker, g.record, g.intent, func() {}, dial); err == nil || c != nil {
					t.Fatal("foreign profile selected")
				}
				return
			}
			if err == nil || service != nil {
				if service != nil {
					service.Close()
				}
				t.Fatal("invalid/partial policy accepted")
			}
		})
	}
}
func TestObjectReadWorkerServiceRejectsWrongListenerAndJoinsAccept(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"wrong address", "shutdown waiting"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, _, ctx, _ := objectReadSessionFixture(t)
			path, _ := objectReadProfileFixture(t, g)
			service, err := newObjectReadWorkerService(ctx, path)
			if err != nil {
				t.Fatal(err)
			}
			defer service.Close()
			listener := &addressedReadListener{newSyntheticReadListener(), "192.0.2.10:443"}
			if kind == "wrong address" {
				listener.address = "192.0.2.11:443"
			}
			done := make(chan error, 1)
			go func() { done <- service.Serve(ctx, listener, g) }()
			if kind == "shutdown waiting" {
				<-listener.started
				service.Close()
			}
			select {
			case err := <-done:
				if err != errObjectReadPeer {
					t.Fatal("unsanitized refusal")
				}
			case <-time.After(time.Second):
				t.Fatal("service did not join serving call")
			}
			if kind == "wrong address" && listener.accepts.Load() != 0 {
				t.Fatal("wrong listener consumed peer")
			}
			select {
			case <-listener.closed:
			default:
				t.Fatal("listener leaked")
			}
		})
	}
}

func TestObjectReadWorkerServiceShutdownDuringDial(t *testing.T) {
	t.Parallel()
	g, _, ctx, _ := objectReadSessionFixture(t)
	path, _ := objectReadProfileFixture(t, g)
	service, err := newObjectReadWorkerService(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	started := make(chan struct{})
	dial := syntheticReadDialer(func(c context.Context, _, _ string) (net.Conn, error) {
		close(started)
		<-c.Done()
		return nil, c.Err()
	})
	done := make(chan error, 1)
	go func() { _, err := service.Connect(ctx, g.worker, g.record, g.intent, func() {}, dial); done <- err }()
	<-started
	closed := make(chan struct{})
	go func() { service.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("service did not join canceled dial")
	}
	if err := <-done; err != errObjectReadPeer {
		t.Fatal("canceled dial escaped sanitized refusal")
	}
}

type syntheticReadListenerFactory func(context.Context, string, string) (net.Listener, error)

func (f syntheticReadListenerFactory) Listen(c context.Context, n, a string) (net.Listener, error) {
	return f(c, n, a)
}
func TestObjectReadWorkerServiceListenerFactoryFailureAndCancellation(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"failure", "late result", "shutdown"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, _, ctx, _ := objectReadSessionFixture(t)
			path, _ := objectReadProfileFixture(t, g)
			service, err := newObjectReadWorkerService(ctx, path)
			if err != nil {
				t.Fatal(err)
			}
			defer service.Close()
			listener := &addressedReadListener{newSyntheticReadListener(), "192.0.2.10:443"}
			started := make(chan struct{})
			child, cancel := context.WithCancel(ctx)
			defer cancel()
			factory := syntheticReadListenerFactory(func(c context.Context, n, a string) (net.Listener, error) {
				if n != "tcp" || a != "192.0.2.10:443" {
					t.Error("unconfigured bind")
				}
				if _, ok := c.Deadline(); !ok {
					t.Error("unbounded listen")
				}
				close(started)
				if kind == "late result" {
					cancel()
					return listener, nil
				}
				if kind == "shutdown" {
					<-c.Done()
					return listener, c.Err()
				}
				return listener, net.ErrClosed
			})
			done := make(chan error, 1)
			go func() { done <- service.ListenAndServe(child, g, factory) }()
			<-started
			if kind == "shutdown" {
				service.Close()
			}
			select {
			case err := <-done:
				if err != errObjectReadPeer {
					t.Fatal("unsanitized listener failure")
				}
			case <-time.After(time.Second):
				t.Fatal("listen did not join")
			}
			select {
			case <-listener.closed:
			default:
				t.Fatal("failed listen leaked listener")
			}
			if listener.accepts.Load() != 0 || g.lifetime.Err() == nil {
				t.Fatal("failed listen accepted or retained authority")
			}
		})
	}
}

type blockedReadServiceClose struct {
	*addressedReadListener
	entered, release chan struct{}
	once             sync.Once
}

func (l *blockedReadServiceClose) Close() error {
	l.once.Do(func() { close(l.entered); <-l.release })
	return l.addressedReadListener.Close()
}
func TestObjectReadWorkerServiceJoinsRefusedListenerClosure(t *testing.T) {
	t.Parallel()
	g, _, ctx, _ := objectReadSessionFixture(t)
	path, _ := objectReadProfileFixture(t, g)
	service, err := newObjectReadWorkerService(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	listener := &blockedReadServiceClose{addressedReadListener: &addressedReadListener{newSyntheticReadListener(), "192.0.2.11:443"}, entered: make(chan struct{}), release: make(chan struct{})}
	serving := make(chan error, 1)
	go func() { serving <- service.Serve(ctx, listener, g) }()
	select {
	case <-listener.entered:
	case <-time.After(time.Second):
		close(listener.release)
		t.Fatal("refusal did not close listener")
	}
	closed := make(chan struct{})
	go func() { service.Close(); close(closed) }()
	select {
	case <-closed:
		close(listener.release)
		t.Fatal("service returned before refused-listener closure joined")
	case <-time.After(20 * time.Millisecond):
	}
	close(listener.release)
	if <-serving != errObjectReadPeer {
		t.Fatal("wrong listener admitted")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("service did not join released closure")
	}
	if listener.accepts.Load() != 0 || g.lifetime.Err() == nil {
		t.Fatal("wrong listener effects or retained authority")
	}
}
