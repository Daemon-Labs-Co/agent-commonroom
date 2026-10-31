package main

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

type captureLifetimeGuard struct {
	workerCaptureJournalGuard
	lifetime context.Context
}

func (g *captureLifetimeGuard) CaptureAuthorityLifetime() context.Context { return g.lifetime }

func TestOwnedCaptureRequiresContinuousControlAuthority(t *testing.T) {
	for _, mode := range []string{"idle before run", "idle after start", "nil lifetime", "expired lifetime"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f, l, j := ownedCaptureSetup(t)
			left, right := net.Pipe()
			var stopped atomic.Int32
			client, err := newCaptureAuthorityClient(left, j.record, j.intent, func() { stopped.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() {
				defer close(done)
				done <- serveCaptureAuthority(t.Context(), right, j.record, j.intent, f.guard)
			}()
			t.Cleanup(func() { client.StopCaptureAuthority(); <-done })
			var guard workerCaptureRuntimeGuard = client
			if mode == "nil lifetime" {
				guard = &captureLifetimeGuard{workerCaptureJournalGuard: client}
			}
			if mode == "expired lifetime" {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				guard = &captureLifetimeGuard{client, ctx}
			}
			capture, err := newOwnedWorkerCapture(l, j.intent.Request, j.intent.Image, j.intent.Objects, guard, f, f)
			if mode == "nil lifetime" || mode == "expired lifetime" {
				if err == nil || capture != nil {
					t.Fatal("missing live control authority admitted")
				}
				if f.starts != 0 || f.exists {
					t.Fatal("invalid setup launched helper")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			lose := func() {
				right.Close()
				select {
				case <-client.CaptureAuthorityLifetime().Done():
				case <-time.After(2 * time.Second):
					t.Error("control lifetime survived EOF")
				}
			}
			if mode == "idle before run" {
				lose()
			} else {
				f.onStart = func(ctx context.Context) {
					lose()
					select {
					case <-ctx.Done():
					case <-time.After(2 * time.Second):
						t.Error("capture did not follow control lifetime")
					}
				}
			}
			report, err := capture.Run(t.Context())
			if err == nil || report.Version != 0 || len(report.Receipts) != 0 || report.SnapshotDigest != "" {
				t.Fatal("control loss published capture report", err)
			}
			client.StopCaptureAuthority()
			<-done
			if stopped.Load() != 1 {
				t.Fatal("control authority not stopped exactly once", stopped.Load())
			}
			slots, e := workerCaptureSlots(j.dir)
			if e != nil {
				t.Fatal(e)
			}
			if mode == "idle before run" {
				if f.starts != 0 || f.exists || slots != 0 {
					t.Fatal("lost authority admitted capture")
				}
			} else {
				if f.starts != 1 || f.exists || slots != 1 {
					t.Fatal("lost control did not clean exact helper and retain uncertainty", f.starts, f.exists, slots)
				}
			}
			if f.reserves != 0 || f.seals != 0 {
				t.Fatal("lost control admitted source/custody operations")
			}
		})
	}
}

func TestRetainedCaptureRecoveryLosesIdleControlAuthority(t *testing.T) {
	t.Parallel()
	f, l, j := ownedCaptureSetup(t)
	beginCaptureJournal(t, j)
	id := f.base.probe.id
	if err := j.BindCapture(t.Context(), j.intent.Request, id); err != nil {
		t.Fatal(err)
	}
	if err := j.FinishCapture(t.Context(), j.intent.Request, id, false); err != nil {
		t.Fatal(err)
	}
	f.exists = true
	f.phase = "exited"
	left, right := net.Pipe()
	client, err := newCaptureAuthorityClient(left, j.record, j.intent, func() {})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	fresh := &captureJournalGuard{f: f.base.probe}
	go func() {
		defer close(done)
		done <- serveCaptureAuthority(t.Context(), right, j.record, j.intent, fresh)
	}()
	defer func() { client.StopCaptureAuthority(); <-done }()
	once := false
	f.onInspect = func(ctx context.Context) {
		if once {
			return
		}
		once = true
		right.Close()
		select {
		case <-client.CaptureAuthorityLifetime().Done():
		case <-time.After(2 * time.Second):
			t.Error("recovery control loss did not revoke lifetime")
		}
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
			t.Error("recovery did not follow control lifetime")
		}
	}
	if recoverRetainedWorkerCapture(t.Context(), l, j.intent, id, client) == nil {
		t.Fatal("recovery succeeded after idle control loss")
	}
	if !once || !f.exists || f.starts != 0 || f.reserves != 0 || f.seals != 0 {
		t.Fatal("lost control allowed effects beyond known-ID inspection")
	}
	if slots, e := workerCaptureSlots(j.dir); e != nil || slots != 1 {
		t.Fatal("lost control released retained capture capacity", slots, e)
	}
}
