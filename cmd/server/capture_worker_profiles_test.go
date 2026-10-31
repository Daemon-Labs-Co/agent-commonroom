package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func captureProfileFixture(t *testing.T, g *controllerCaptureSession) (string, captureWorkerProfileFile) {
	t.Helper()
	client, server := provisionedObjectReadTestFiles(t)
	file := captureWorkerProfileFile{Protocol: captureControlALPN, Version: 1, Profiles: []captureWorkerProfileSpec{
		{g.intent.Request.Binding.Execution.CompanyID, g.worker, "worker", "192.0.2.10:443", client},
		{g.intent.Request.Binding.Execution.CompanyID, g.worker, "controller", "192.0.2.10:443", server},
	}}
	path := filepath.Join(t.TempDir(), "profiles.json")
	writeCaptureProfiles(t, path, file)
	return path, file
}
func writeCaptureProfiles(t *testing.T, path string, file captureWorkerProfileFile) {
	t.Helper()
	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCaptureWorkerServiceLifecycle(t *testing.T) {
	t.Parallel()
	g, db, ctx, _ := captureSessionFixture(t)
	before := len(db.snapshot())
	path, _ := captureProfileFixture(t, g)
	service, err := newCaptureWorkerService(ctx, path)
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
	defer c.StopCaptureAuthority()
	event := workerProbeEvent{Stage: "begin"}
	if err = c.AuthorizeCapture(ctx, g.record, g.intent, event); err != nil {
		t.Fatal(err)
	}
	if err = c.AcknowledgeCapture(ctx, workerCaptureRecord{Version: 1, Intent: g.intent, Events: []workerProbeEvent{event}}, event); err != nil {
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
	if c.CaptureAuthorityLifetime().Err() == nil || g.lifetime.Err() == nil || len(db.snapshot()) != before+1 {
		t.Fatal("shutdown lost audit or retained authority")
	}
	if _, err = service.Connect(ctx, g.worker, g.record, g.intent, func() {}, dial); err == nil {
		t.Fatal("closed service revived")
	}
	st := artifactState(t, g.server, ctx)
	if st.WorkspaceCaptures[controllerCaptureKey(g.intent)].Released {
		t.Fatal("shutdown cleared uncertain capacity")
	}
}
func TestCaptureWorkerProfilesRefuseInvalidPolicy(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"protocol", "missing protocol", "version", "empty", "duplicate profile", "bad role", "bad endpoint", "unknown key", "duplicate key", "case duplicate", "bad identity", "public policy", "foreign company", "unknown worker"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, _, ctx, _ := captureSessionFixture(t)
			path, file := captureProfileFixture(t, g)
			switch kind {
			case "protocol":
				file.Protocol = objectReadControlALPN
			case "missing protocol":
				file.Protocol = ""
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
			writeCaptureProfiles(t, path, file)
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
			service, err := newCaptureWorkerService(ctx, path)
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
func TestCaptureWorkerServiceRejectsWrongListenerAndJoinsAccept(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"wrong address", "shutdown waiting"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, _, ctx, _ := captureSessionFixture(t)
			path, _ := captureProfileFixture(t, g)
			service, err := newCaptureWorkerService(ctx, path)
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
				if err != errCapturePeer {
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

func TestCaptureWorkerServiceShutdownDuringDial(t *testing.T) {
	t.Parallel()
	g, _, ctx, _ := captureSessionFixture(t)
	path, _ := captureProfileFixture(t, g)
	service, err := newCaptureWorkerService(ctx, path)
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
	if err := <-done; err != errCapturePeer {
		t.Fatal("canceled dial escaped sanitized refusal")
	}
}

func TestCaptureWorkerServiceListenerFactoryFailureAndCancellation(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"failure", "late result", "shutdown"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, _, ctx, _ := captureSessionFixture(t)
			path, _ := captureProfileFixture(t, g)
			service, err := newCaptureWorkerService(ctx, path)
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
				if err != errCapturePeer {
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

func TestCaptureWorkerServiceJoinsRefusedSessionRevocation(t *testing.T) {
	t.Parallel()
	g, _, ctx, _ := captureSessionFixture(t)
	path, _ := captureProfileFixture(t, g)
	service, err := newCaptureWorkerService(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	listener := &addressedReadListener{newSyntheticReadListener(), "192.0.2.11:443"}
	// Delay the actual integrity stop after cancellation. Close must still join
	// that stop, even when policy rejected the listener before Accept.
	g.server.mu.Lock()
	serving := make(chan error, 1)
	go func() { serving <- service.Serve(ctx, listener, g) }()
	select {
	case <-g.lifetime.Done():
	case <-time.After(time.Second):
		g.server.mu.Unlock()
		t.Fatal("refusal did not cancel authority")
	}
	closed := make(chan struct{})
	go func() { service.Close(); close(closed) }()
	select {
	case <-closed:
		g.server.mu.Unlock()
		t.Fatal("service returned before refused-session revocation joined")
	case <-time.After(20 * time.Millisecond):
	}
	g.server.mu.Unlock()
	if <-serving != errCapturePeer {
		t.Fatal("wrong listener admitted")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("service did not join released revocation")
	}
	if listener.accepts.Load() != 0 {
		t.Fatal("wrong listener accepted a peer")
	}
}
