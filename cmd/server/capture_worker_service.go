package main

import (
	"context"
	"net"
	"net/netip"
	"sync"
	"time"
)

// One explicitly provisioned service snapshot owns its sessions and cancellation.
// It does not create provider/model workers, keys or physical witnesses, or
// install automatic listeners; ListenAndServe binds only the approved endpoint.
// No public API/model input can select policy files or bypass mandatory isolation.
// Close cancels first, then joins serving calls, dial/setup calls and client pumps.
// Stop callbacks must remain nonblocking and must not reenter service Close.
type captureWorkerService struct {
	profiles *captureWorkerProfiles
	lifetime context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	closed   bool
	work     sync.WaitGroup
}

func newCaptureWorkerService(ctx context.Context, path string) (*captureWorkerService, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, errCapturePeer
	}
	profiles, err := loadCaptureWorkerProfiles(path)
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, errCapturePeer
	}
	lifetime, cancel := context.WithCancel(ctx)
	return &captureWorkerService{profiles: profiles, lifetime: lifetime, cancel: cancel}, nil
}
func (s *captureWorkerService) begin(ctx context.Context) (context.Context, func(), error) {
	if s == nil || ctx == nil {
		return nil, nil, errCapturePeer
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.lifetime.Err() != nil || ctx.Err() != nil {
		return nil, nil, errCapturePeer
	}
	s.work.Add(1)
	child, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(s.lifetime, cancel)
	return child, func() { stop(); cancel(); s.work.Done() }, nil
}
func (s *captureWorkerService) Connect(ctx context.Context, worker string, r workerJobRecord, i workerCaptureIntent, stop func(), dialer objectReadControlDialer) (*authenticatedCaptureClient, error) {
	child, finish, err := s.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer finish()
	endpoint, err := s.profiles.endpoint(worker, "worker", r, i)
	if err != nil {
		return nil, err
	}
	// The retained client needs a lifetime beyond this method's setup child.
	live, cancel := context.WithCancel(ctx)
	stopService := context.AfterFunc(s.lifetime, cancel)
	c, err := connectCaptureControl(live, endpoint, r, i, stop, dialer)
	if err != nil {
		stopService()
		cancel()
		return nil, err
	}
	s.mu.Lock()
	if s.closed || s.lifetime.Err() != nil || child.Err() != nil || live.Err() != nil {
		s.mu.Unlock()
		stopService()
		cancel()
		c.StopCaptureAuthority()
		return nil, errCapturePeer
	}
	s.work.Add(1)
	s.mu.Unlock()
	// This callback runs separately from the sole TLS read pump, so it can join it.
	context.AfterFunc(c.CaptureAuthorityLifetime(), func() { defer s.work.Done(); stopService(); cancel(); c.StopCaptureAuthority() })
	return c, nil
}
func (s *captureWorkerService) Serve(ctx context.Context, listener net.Listener, g *controllerCaptureSession) error {
	closeListener := func() {
		if !nilWorkerObservationDependency(listener) {
			_ = listener.Close()
		}
	}
	if g == nil {
		closeListener()
		return errCapturePeer
	}
	child, finish, err := s.begin(ctx)
	if err != nil {
		closeListener()
		g.StopCaptureAuthority()
		return err
	}
	// Join closure and authority revocation before decrementing tracked work,
	// including policy/address refusals before the one-shot server is reached.
	defer func() { defer finish(); closeListener(); g.StopCaptureAuthority() }()
	endpoint, err := s.profiles.endpoint(g.worker, "controller", g.record, g.intent)
	if err != nil || nilWorkerObservationDependency(listener) || listener.Addr() == nil {
		return errCapturePeer
	}
	actual, err := netip.ParseAddrPort(listener.Addr().String())
	expected, _ := netip.ParseAddrPort(endpoint.Address)
	if err != nil || actual != expected {
		return errCapturePeer
	}
	return serveCaptureControlOnce(child, listener, endpoint.Binding, g)
}
func (s *captureWorkerService) Close() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	s.cancel()
	s.work.Wait()
}

// Deployment passes a net.ListenConfig. Test/source injections must honor context
// cancellation and return a listener whose Close unblocks Accept. Only the fixed
// configured TCP address can be bound; neither model metadata nor a caller string
// selects a listener. Any uncertain creation refuses and retains no capture authority.
func (s *captureWorkerService) ListenAndServe(ctx context.Context, g *controllerCaptureSession, factory objectReadControlListenerFactory) error {
	if g == nil {
		return errCapturePeer
	}
	if nilWorkerObservationDependency(factory) || g.lifetime == nil || g.lifetime.Err() != nil {
		g.StopCaptureAuthority()
		return errCapturePeer
	}
	child, finish, err := s.begin(ctx)
	if err != nil {
		g.StopCaptureAuthority()
		return err
	}
	var listener net.Listener
	defer func() {
		defer finish()
		if !nilWorkerObservationDependency(listener) {
			_ = listener.Close()
		}
		g.StopCaptureAuthority()
	}()
	endpoint, err := s.profiles.endpoint(g.worker, "controller", g.record, g.intent)
	if err != nil {
		return err
	}
	listenCtx, cancel := context.WithTimeout(child, 5*time.Second)
	stopSession := context.AfterFunc(g.lifetime, cancel)
	listener, err = factory.Listen(listenCtx, "tcp", endpoint.Address)
	expired := listenCtx.Err() != nil
	stopSession()
	cancel()
	if err != nil || expired || child.Err() != nil || g.lifetime.Err() != nil || nilWorkerObservationDependency(listener) {
		return errCapturePeer
	}
	return s.Serve(child, listener, g)
}
