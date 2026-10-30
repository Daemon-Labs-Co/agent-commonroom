package main

import (
	"context"
	"net"
	"net/netip"
	"sync"
	"time"
)

// One explicitly provisioned service snapshot owns its sessions and cancellation.
// It does not create provider/model workers, listeners, keys or physical witnesses.
// No public API/model input can select policy files or bypass mandatory isolation.
// Close cancels first, then joins serving calls, dial/setup calls and client pumps.
// Stop callbacks must remain nonblocking and must not reenter service Close.
type objectReadWorkerService struct {
	profiles *objectReadWorkerProfiles
	lifetime context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	closed   bool
	work     sync.WaitGroup
}

func newObjectReadWorkerService(ctx context.Context, path string) (*objectReadWorkerService, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, errObjectReadPeer
	}
	profiles, err := loadObjectReadWorkerProfiles(path)
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, errObjectReadPeer
	}
	lifetime, cancel := context.WithCancel(ctx)
	return &objectReadWorkerService{profiles: profiles, lifetime: lifetime, cancel: cancel}, nil
}
func (s *objectReadWorkerService) begin(ctx context.Context) (context.Context, func(), error) {
	if s == nil || ctx == nil {
		return nil, nil, errObjectReadPeer
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.lifetime.Err() != nil || ctx.Err() != nil {
		return nil, nil, errObjectReadPeer
	}
	s.work.Add(1)
	child, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(s.lifetime, cancel)
	return child, func() { stop(); cancel(); s.work.Done() }, nil
}
func (s *objectReadWorkerService) Connect(ctx context.Context, worker string, r workerJobRecord, i workerObjectReadIntent, stop func(), dialer objectReadControlDialer) (*authenticatedObjectReadClient, error) {
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
	c, err := connectObjectReadControl(live, endpoint, r, i, stop, dialer)
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
		c.StopObjectReadAuthority()
		return nil, errObjectReadPeer
	}
	s.work.Add(1)
	s.mu.Unlock()
	// This callback runs separately from the sole TLS read pump, so it can join it.
	context.AfterFunc(c.ObjectReadAuthorityLifetime(), func() { defer s.work.Done(); stopService(); cancel(); c.StopObjectReadAuthority() })
	return c, nil
}
func (s *objectReadWorkerService) Serve(ctx context.Context, listener net.Listener, g *controllerObjectReadSession) error {
	if !nilWorkerObservationDependency(listener) {
		defer listener.Close()
	}
	if g == nil {
		return errObjectReadPeer
	}
	defer g.StopObjectReadAuthority()
	child, finish, err := s.begin(ctx)
	if err != nil {
		return err
	}
	defer finish()
	endpoint, err := s.profiles.endpoint(g.worker, "controller", g.record, g.intent)
	if err != nil || nilWorkerObservationDependency(listener) || listener.Addr() == nil {
		return errObjectReadPeer
	}
	actual, err := netip.ParseAddrPort(listener.Addr().String())
	expected, _ := netip.ParseAddrPort(endpoint.Address)
	if err != nil || actual != expected {
		return errObjectReadPeer
	}
	return serveObjectReadControlOnce(child, listener, endpoint.Binding, g)
}
func (s *objectReadWorkerService) Close() {
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
// selects a listener. Any uncertain creation refuses and retains no read authority.
type objectReadControlListenerFactory interface {
	Listen(context.Context, string, string) (net.Listener, error)
}

func (s *objectReadWorkerService) ListenAndServe(ctx context.Context, g *controllerObjectReadSession, factory objectReadControlListenerFactory) error {
	if g == nil {
		return errObjectReadPeer
	}
	defer g.StopObjectReadAuthority()
	if nilWorkerObservationDependency(factory) || g.lifetime == nil || g.lifetime.Err() != nil {
		return errObjectReadPeer
	}
	child, finish, err := s.begin(ctx)
	if err != nil {
		return err
	}
	defer finish()
	endpoint, err := s.profiles.endpoint(g.worker, "controller", g.record, g.intent)
	if err != nil {
		return err
	}
	listenCtx, cancel := context.WithTimeout(child, 5*time.Second)
	stopSession := context.AfterFunc(g.lifetime, cancel)
	listener, err := factory.Listen(listenCtx, "tcp", endpoint.Address)
	expired := listenCtx.Err() != nil
	stopSession()
	cancel()
	if !nilWorkerObservationDependency(listener) {
		defer listener.Close()
	}
	if err != nil || expired || child.Err() != nil || g.lifetime.Err() != nil || nilWorkerObservationDependency(listener) {
		return errObjectReadPeer
	}
	return s.Serve(child, listener, g)
}
