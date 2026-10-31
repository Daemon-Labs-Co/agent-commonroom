package main

import (
	"agentroom/internal/room"
	"agentroom/internal/workspacecapture"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func boundCapturePeers(t *testing.T, g *controllerCaptureSession) (capturePeerBinding, capturePeerBinding) {
	t.Helper()
	client, server := objectReadTestPeers(t)
	b := capturePeerBinding{Identity: client, Worker: g.worker, RecordDigest: probeWorkerRecordDigest(g.record), Intent: copyCaptureIntent(g.intent)}
	s := b
	s.Identity = server
	s.Intent = copyCaptureIntent(g.intent)
	return b, s
}
func TestAuthenticatedCaptureSessionLifecycleAndPermissions(t *testing.T) {
	for _, denied := range []bool{false, true} {
		t.Run(fmt.Sprint("denied=", denied), func(t *testing.T) {
			t.Parallel()
			g, db, ctx, e := captureSessionFixture(t)
			if denied {
				e.operation = func(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error {
					return errControllerCapture
				}
			}
			before := len(db.snapshot())
			client, server := boundCapturePeers(t, g)
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			done := make(chan error, 1)
			go func() { done <- serveAuthenticatedCaptureSession(ctx, right, server, g) }()
			c, err := newAuthenticatedCaptureClient(ctx, left, client, g.record, g.intent, func() {})
			if err != nil {
				t.Fatal(err)
			}
			defer c.StopCaptureAuthority()
			local := workerCaptureRecord{Version: 1, Intent: copyCaptureIntent(g.intent)}
			for _, event := range []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: strings.Repeat("1", 64)}, {Stage: "finish", ID: strings.Repeat("1", 64), Absent: true}} {
				err = c.AuthorizeCapture(ctx, g.record, g.intent, event)
				if denied {
					if err == nil {
						t.Fatal("certificate identity bypassed current permission")
					}
					break
				}
				if err != nil {
					t.Fatal("current authorization", err)
				}
				local.Events = append(local.Events, event)
				if err = c.AcknowledgeCapture(ctx, local, event); err != nil {
					t.Fatal("durable acknowledgment", err)
				}
			}
			c.StopCaptureAuthority()
			<-done
			if g.CaptureAuthorityLifetime().Err() == nil || c.CaptureAuthorityLifetime().Err() == nil {
				t.Fatal("closed session kept authority")
			}
			want := before + 3
			if denied {
				want = before
			}
			if len(db.snapshot()) != want {
				t.Fatal("wrong audit count", len(db.snapshot()), want)
			}
			if !denied {
				saved := artifactState(t, g.server, ctx).WorkspaceCaptures[controllerCaptureKey(g.intent)]
				if saved.ActorID != g.actor || !saved.Released {
					t.Fatal("lost original principal or lifecycle settlement")
				}
			}
		})
	}
}
func TestAuthenticatedCaptureRefusesForeignBindingBeforeAuthority(t *testing.T) {
	for _, mode := range []string{"worker", "record", "intent", "certificate", "worker certificate", "trust", "controller worker", "controller record"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, e := captureSessionFixture(t)
			before := len(db.snapshot())
			client, server := boundCapturePeers(t, g)
			switch mode {
			case "worker":
				client.Worker = "foreign"
			case "record":
				client.RecordDigest = strings.Repeat("0", 64)
			case "intent":
				client.Intent.Request.Members[0].Name = "foreign.txt"
			case "certificate":
				client.Identity.PeerSHA256[0] ^= 1
			case "worker certificate":
				server.Identity.PeerSHA256[0] ^= 1
			case "trust":
				client.Identity.Roots = x509.NewCertPool()
			case "controller worker":
				server.Worker = "foreign"
			case "controller record":
				server.RecordDigest = strings.Repeat("0", 64)
			}
			child, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			done := make(chan error, 1)
			go func() { done <- serveAuthenticatedCaptureSession(child, right, server, g) }()
			c, err := newAuthenticatedCaptureClient(child, left, client, g.record, g.intent, func() {})
			if c != nil {
				c.StopCaptureAuthority()
			}
			other := <-done
			if err == nil || c != nil || other == nil || len(db.snapshot()) != before || e.calls != 0 || e.acks != 0 {
				t.Fatal("foreign peer reached capture authority", mode, err, other)
			}
		})
	}
}
func TestAuthenticatedCaptureRefusesForeignHelloAndReaderProtocol(t *testing.T) {
	for _, mode := range []string{"hello protocol", "hello worker", "hello record", "hello intent", "reader ALPN"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, e := captureSessionFixture(t)
			before := len(db.snapshot())
			client, server := boundCapturePeers(t, g)
			child, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			done := make(chan error, 1)
			go func() { done <- serveAuthenticatedCaptureSession(child, right, server, g) }()
			protocol := captureControlALPN
			if mode == "reader ALPN" {
				protocol = objectReadControlALPN
			}
			conn, err := authenticatePinnedWorkerStream(child, left, client.Identity, false, protocol)
			if mode == "reader ALPN" {
				if err == nil || conn != nil {
					t.Fatal("reader ALPN admitted by capture service")
				}
			} else {
				if err != nil {
					t.Fatal("TLS setup", err)
				}
				hello := client.hello()
				switch mode {
				case "hello protocol":
					hello.Protocol = "object-read-session-v1"
				case "hello worker":
					hello.Worker = "foreign"
				case "hello record":
					hello.RecordDigest = strings.Repeat("0", 64)
				case "hello intent":
					hello.Intent.Request.Challenge = strings.Repeat("9", 32)
				}
				_ = writeCaptureAuthorityFrame(conn, hello)
				var response captureSessionHello
				if readCaptureAuthorityFrame(conn, &response) == nil {
					t.Fatal("foreign hello acknowledged")
				}
			}
			if other := <-done; other == nil || len(db.snapshot()) != before || e.calls != 0 || e.acks != 0 {
				t.Fatal("foreign protocol/binding reached authority", mode, other)
			}
		})
	}
}
func TestAuthenticatedCaptureOriginalContextRevokesIdleSession(t *testing.T) {
	t.Parallel()
	g, _, ctx, _ := captureSessionFixture(t)
	client, server := boundCapturePeers(t, g)
	worker, cancel := context.WithCancel(ctx)
	defer cancel()
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()
	done := make(chan error, 1)
	go func() { done <- serveAuthenticatedCaptureSession(ctx, right, server, g) }()
	c, err := newAuthenticatedCaptureClient(worker, left, client, g.record, g.intent, func() {})
	if err != nil {
		t.Fatal(err)
	}
	defer c.StopCaptureAuthority()
	cancel()
	select {
	case <-c.CaptureAuthorityLifetime().Done():
	case <-time.After(2 * time.Second):
		t.Fatal("idle client outlived original context")
	}
	c.StopCaptureAuthority()
	<-done
	if g.CaptureAuthorityLifetime().Err() == nil {
		t.Fatal("server outlived closed client")
	}
	if c.AuthorizeCapture(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("new context revived original authority")
	}
}
func TestCaptureMutualTLSSupportedVersions(t *testing.T) {
	client, server := objectReadTestPeers(t)
	for _, version := range []uint16{tls.VersionTLS12, tls.VersionTLS13} {
		for _, adapterServer := range []bool{false, true} {
			t.Run(fmt.Sprintf("%x/server=%t", version, adapterServer), func(t *testing.T) {
				t.Parallel()
				local, peer := client, server
				if adapterServer {
					local, peer = server, client
				}
				left, right := net.Pipe()
				defer left.Close()
				defer right.Close()
				ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
				defer cancel()
				config := &tls.Config{MinVersion: version, MaxVersion: version, Certificates: []tls.Certificate{peer.Certificate}, RootCAs: peer.Roots, ClientCAs: peer.Roots, ServerName: peer.PeerDNSName, NextProtos: []string{captureControlALPN}, ClientAuth: tls.RequireAndVerifyClientCert}
				done := make(chan error, 1)
				go func() {
					var conn *tls.Conn
					if adapterServer {
						conn = tls.Client(right, config)
					} else {
						conn = tls.Server(right, config)
					}
					done <- conn.HandshakeContext(ctx)
				}()
				conn, err := authenticatePinnedWorkerStream(ctx, left, local, adapterServer, captureControlALPN)
				other := <-done
				if err != nil || other != nil || conn == nil || conn.ConnectionState().Version != version {
					t.Fatal("supported capture mutual TLS refused", version, err, other)
				}
			})
		}
	}
}

func TestAuthenticatedCaptureOwnedRuntimeControllerBridge(t *testing.T) {
	for _, off := range []bool{false, true} {
		t.Run(fmt.Sprint("editingDisabled=", off), func(t *testing.T) {
			t.Parallel()
			g, db, ctx, _ := captureSessionFixture(t)
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
			probe := &assembledProbeFixture{t: t, dir: dir, record: detachedProbeWorkerRecord(g.record), id: strings.Repeat("1", 64), image: g.intent.Image}
			f := &ownedCaptureFixture{base: &physicalCaptureFixture{probe: probe, request: copyCaptureRequest(g.intent.Request), objects: g.intent.Objects}}
			client, server := boundCapturePeers(t, g)
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			done := make(chan error, 1)
			go func() { defer close(done); done <- serveAuthenticatedCaptureSession(ctx, right, server, g) }()
			c, err := newAuthenticatedCaptureClient(ctx, left, client, g.record, g.intent, func() {})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { c.StopCaptureAuthority(); <-done }()
			if off {
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
			}
			var report workspacecapture.Report
			err = withWorkerWorkspaceOwnership(ctx, dir, g.record, syntheticObservationDockerRoot, f, probe, func(owner *workerWorkspaceObservation) error {
				runtime, e := newOwnedWorkerCapture(owner, g.intent.Request, g.intent.Image, g.intent.Objects, c, f, f)
				if e != nil {
					return e
				}
				report, e = runtime.Run(ctx)
				return e
			})
			if off {
				if err == nil || len(report.Receipts) != 0 || f.starts != 0 || f.exists {
					t.Fatal("disabled task editing admitted a capture", err)
				}
			} else {
				if err != nil || len(report.Receipts) != 1 || f.starts != 1 || f.exists {
					t.Fatal("authenticated owned capture failed", err)
				}
				st := artifactState(t, g.server, ctx)
				if saved := st.WorkspaceCaptures[controllerCaptureKey(g.intent)]; !saved.Released || saved.ActorID != g.actor {
					t.Fatal("controller capture settlement lost")
				}
				if slots, e := workerCaptureSlots(dir); e != nil || slots != 0 {
					t.Fatal("worker capture settlement lost", slots, e)
				}
			}
		})
	}
}
