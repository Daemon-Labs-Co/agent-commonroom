package main

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type syntheticReadDialer func(context.Context, string, string) (net.Conn, error)

func (f syntheticReadDialer) DialContext(c context.Context, n, a string) (net.Conn, error) {
	return f(c, n, a)
}

type syntheticReadListener struct {
	incoming             chan net.Conn
	closed, started      chan struct{}
	closeOnce, startOnce sync.Once
	accepts              atomic.Int32
	failClose            bool
}

func newSyntheticReadListener() *syntheticReadListener {
	return &syntheticReadListener{incoming: make(chan net.Conn, 1), closed: make(chan struct{}), started: make(chan struct{})}
}
func (l *syntheticReadListener) Accept() (net.Conn, error) {
	l.accepts.Add(1)
	l.startOnce.Do(func() { close(l.started) })
	select {
	case c := <-l.incoming:
		return c, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}
func (l *syntheticReadListener) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })
	if l.failClose {
		return errors.New("synthetic listener close failure")
	}
	return nil
}
func (l *syntheticReadListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1234}
}

func TestObjectReadControlEndpointLifecycle(t *testing.T) {
	t.Parallel()
	g, db, ctx, _ := objectReadSessionFixture(t)
	before := len(db.snapshot())
	client, server := boundObjectReadFixture(t, g)
	listener := newSyntheticReadListener()
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	var calls atomic.Int32
	endpoint := objectReadControlEndpoint{"192.0.2.10:443", client}
	dial := syntheticReadDialer(func(c context.Context, n, address string) (net.Conn, error) {
		calls.Add(1)
		if n != "tcp" || address != endpoint.Address {
			t.Error("dial escaped provisioned TCP endpoint")
		}
		if deadline, ok := c.Deadline(); !ok || time.Until(deadline) > 5*time.Second {
			t.Error("unbounded dial")
		}
		listener.incoming <- b
		return a, nil
	})
	done := make(chan error, 1)
	go func() { done <- serveObjectReadControlOnce(ctx, listener, server, g) }()
	c, err := connectObjectReadControl(ctx, endpoint, g.record, g.intent, func() {}, dial)
	if err != nil {
		t.Fatal(err)
	}
	defer c.StopObjectReadAuthority()
	select {
	case <-listener.closed:
	default:
		t.Fatal("listener survived accepted connection")
	}
	local := workerObjectReadRecord{Version: 1, Intent: g.intent}
	for _, event := range []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: strings.Repeat("1", 64)}, {Stage: "finish", ID: strings.Repeat("1", 64), Absent: true}} {
		if err = c.AuthorizeObjectRead(ctx, g.record, g.intent, event); err != nil {
			t.Fatal(err)
		}
		local.Events = append(local.Events, event)
		if err = c.AcknowledgeObjectRead(ctx, local, event); err != nil {
			t.Fatal(err)
		}
	}
	c.StopObjectReadAuthority()
	<-done
	if calls.Load() != 1 || listener.accepts.Load() != 1 || len(db.snapshot()) != before+3 || g.lifetime.Err() == nil {
		t.Fatal("endpoint retried, lost audit or retained authority")
	}
}
func TestObjectReadControlEndpointRejectsBeforeDial(t *testing.T) {
	t.Parallel()
	for _, address := range []string{"https://example.test", "example.test:443", "192.0.2.10", "192.0.2.10:0", "0.0.0.0:443", "224.0.0.1:443", "[::]:443", "[ff02::1]:443", "[fe80::1%en0]:443", "[::ffff:192.0.2.10]:443"} {
		t.Run(address, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, _ := objectReadSessionFixture(t)
			client, _ := boundObjectReadFixture(t, g)
			before := len(db.snapshot())
			dial := syntheticReadDialer(func(context.Context, string, string) (net.Conn, error) {
				t.Error("invalid address dialed")
				return nil, errors.New("private error")
			})
			c, err := connectObjectReadControl(ctx, objectReadControlEndpoint{address, client}, g.record, g.intent, func() {}, dial)
			if c != nil || err != errObjectReadPeer || len(db.snapshot()) != before {
				t.Fatal("invalid address accepted or unsanitized")
			}
		})
	}
}
func TestObjectReadControlEndpointInvalidIdentityAndDialFailure(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"identity", "nil key", "record", "canceled", "dial failure", "expired dial"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, _, ctx, _ := objectReadSessionFixture(t)
			client, _ := boundObjectReadFixture(t, g)
			a, b := net.Pipe()
			defer a.Close()
			defer b.Close()
			calls := 0
			child, cancel := context.WithCancel(ctx)
			defer cancel()
			switch kind {
			case "identity":
				client.Identity.Certificate = tls.Certificate{}
			case "nil key":
				var key ed25519.PrivateKey
				client.Identity.Certificate.PrivateKey = key
			case "record":
				client.RecordDigest = strings.Repeat("0", 64)
			case "canceled":
				cancel()
			}
			dial := syntheticReadDialer(func(c context.Context, _, _ string) (net.Conn, error) {
				calls++
				if kind == "expired dial" {
					cancel()
					return a, nil
				}
				return a, errors.New("sensitive fixture error")
			})
			c, err := connectObjectReadControl(child, objectReadControlEndpoint{"192.0.2.10:443", client}, g.record, g.intent, func() {}, dial)
			if c != nil || err != errObjectReadPeer {
				t.Fatal("failed setup returned client or leaked error")
			}
			want := 0
			if kind == "dial failure" || kind == "expired dial" {
				want = 1
				if _, e := a.Write([]byte("x")); e == nil {
					t.Fatal("failed dial leaked connection")
				}
			}
			if calls != want {
				t.Fatal("unexpected dial or retry", calls)
			}
		})
	}
}
func TestObjectReadControlListenerCancellation(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"serving context", "pinned session", "invalid binding"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, _ := objectReadSessionFixture(t)
			before := len(db.snapshot())
			_, server := boundObjectReadFixture(t, g)
			listener := newSyntheticReadListener()
			child, cancel := context.WithCancel(ctx)
			defer cancel()
			if kind == "invalid binding" {
				server.Worker = "other-worker"
			}
			done := make(chan error, 1)
			go func() { done <- serveObjectReadControlOnce(child, listener, server, g) }()
			if kind != "invalid binding" {
				<-listener.started
				if kind == "serving context" {
					cancel()
				} else {
					g.StopObjectReadAuthority()
				}
			}
			select {
			case err := <-done:
				if err != errObjectReadPeer {
					t.Fatal("unsanitized refusal")
				}
			case <-time.After(time.Second):
				t.Fatal("cancellation did not join Accept")
			}
			select {
			case <-listener.closed:
			default:
				t.Fatal("listener leaked")
			}
			want := int32(1)
			if kind == "invalid binding" {
				want = 0
			}
			if listener.accepts.Load() != want || g.lifetime.Err() == nil || len(db.snapshot()) != before {
				t.Fatal("invalid listener effects or authority retained")
			}
		})
	}
}

func TestObjectReadControlListenerCloseFailureRefusesSession(t *testing.T) {
	t.Parallel()
	g, db, ctx, e := objectReadSessionFixture(t)
	before := len(db.snapshot())
	client, server := boundObjectReadFixture(t, g)
	listener := newSyntheticReadListener()
	listener.failClose = true
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	listener.incoming <- b
	done := make(chan error, 1)
	go func() { done <- serveObjectReadControlOnce(ctx, listener, server, g) }()
	dial := syntheticReadDialer(func(context.Context, string, string) (net.Conn, error) { return a, nil })
	c, err := connectObjectReadControl(ctx, objectReadControlEndpoint{"192.0.2.10:443", client}, g.record, g.intent, func() {}, dial)
	if c != nil {
		c.StopObjectReadAuthority()
	}
	if err != errObjectReadPeer || c != nil || <-done != errObjectReadPeer {
		t.Fatal("listener closure failure admitted session")
	}
	if e.calls != 0 || e.acks != 0 || len(db.snapshot()) != before || g.lifetime.Err() == nil {
		t.Fatal("failed listener reached authority or persistence")
	}
}
