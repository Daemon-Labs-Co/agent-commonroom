//go:build linux || darwin

package main

import (
	"agentroom/internal/artifactmanifest"
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"agentroom/internal/workspacesnapshot"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This lease is explicitly synthetic. Actual worker/volume fencing and read-only
// mount attestation remain required before production assembly can use Capture.
type syntheticFrozenArtifactLease struct{ deny bool }

func (l *syntheticFrozenArtifactLease) VerifyFrozen(context.Context, *os.Root, workspacesnapshot.Binding) error {
	if l.deny {
		return errors.New("synthetic source lease revoked")
	}
	return nil
}

type artifactStreamProbe struct {
	io.ReadCloser
	read func()
}

func (p artifactStreamProbe) Read(b []byte) (int, error) { p.read(); return p.ReadCloser.Read(b) }

func TestActualSourceSnapshotFeedsDurableCustodyAndRefusesChangedSource(t *testing.T) {
	for _, mode := range []string{"healthy", "changed source", "revoked source", "forged digest"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			dir := t.TempDir()
			if os.Chmod(dir, 0700) != nil {
				t.Fatal("root mode")
			}
			if err := os.WriteFile(filepath.Join(dir, "report.txt"), []byte("actual synthetic report\n"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "source.go"), []byte("package synthetic\n"), 0644); err != nil {
				t.Fatal(err)
			}
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			execution := scope.ExecutionScope
			execution.RevisionDigest = ""
			binding := workspacesnapshot.Binding{Execution: execution, BaseCommit: strings.Repeat("a", 40), VolumeName: "commonroom-tool-work-" + scope.JobID, VolumeNonce: strings.Repeat("b", 32), VolumeCreatedAt: time.Unix(123, 0).UTC().Format(time.RFC3339Nano)}
			lease := &syntheticFrozenArtifactLease{}
			source, err := workspacesnapshot.Capture(ctx, root, binding, []workspacesnapshot.Member{{OutputName: "report", Name: "report.txt", MaxBytes: 128}}, lease)
			if err != nil {
				t.Fatal(err)
			}
			scope.ExecutionScope = source.Snapshot().Scope
			quotaReady := false
			guard := &syntheticArtifactGuard{}
			guard.source = func(c context.Context, _ room.State, intent objectcustody.Intent) error {
				if intent.Scope != scope || intent.MaxBytes != 128 {
					return errors.New("wrong source contract")
				}
				return source.Current(c)
			}
			guard.quota = func(context.Context, room.StorageReservation) error { quotaReady = true; return nil }
			guard.seal = func(c context.Context, _ room.State, r objectcustody.Receipt) error {
				return source.VerifyReceipt(c, r)
			}
			guard.release = func(c context.Context, _ room.State, r objectcustody.Receipt) error {
				return source.VerifyReceipt(c, r)
			}
			journal, _ := newControllerArtifactJournal(s, guard)
			store, objectDir := artifactStore(t, journal)
			stream, err := source.Open(ctx, scope, 128)
			if err != nil {
				t.Fatal(err)
			}
			defer stream.Close()
			reads := 0
			wrapped := artifactStreamProbe{ReadCloser: stream, read: func() {
				reads++
				if !quotaReady {
					t.Error("stream consumed before quota readiness")
				}
			}}
			// A source change after opening the lazy reader must not expose stale bytes.
			if mode == "changed source" {
				if err := os.WriteFile(filepath.Join(dir, "source.go"), []byte("changed source\n"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "revoked source" {
				lease.deny = true
			}
			stored, err := store.Put(ctx, scope, 128, wrapped)
			if mode == "changed source" || mode == "revoked source" {
				if err == nil || reads != 0 {
					t.Fatal("changed/revoked source admitted", err, reads)
				}
				if len(artifactState(t, s, ctx).ArtifactCustody) != 0 {
					t.Fatal("invalid source charged")
				}
				return
			}
			if err != nil || reads == 0 {
				t.Fatal("actual captured source not stored", err)
			}
			if stored.Intent.Scope.RevisionDigest != source.Snapshot().Scope.RevisionDigest || artifactTotal(t, s, ctx) != 128 {
				t.Fatal("lost observed revision/charge")
			}
			if mode == "forged digest" {
				stored.Digest = strings.Repeat("f", 64)
			}
			raw, err := store.Get(ctx, scope, stored)
			if mode == "forged digest" {
				if err == nil || raw != nil {
					t.Fatal("forged source receipt released bytes")
				}
				return
			}
			if err != nil || string(raw) != "actual synthetic report\n" {
				t.Fatal("actual source bytes not released", err)
			}
			if artifactState(t, &server{db: db}, ctx).ArtifactCustody[stored.Intent.ObjectID].ReadAcknowledgments != 1 {
				t.Fatal("durable read missing")
			}
			state := artifactState(t, s, ctx)
			definition := state.TaskExecutions[scope.ExecutionID].Definition
			manifest, err := artifactmanifest.Resolve(ctx, definition, scope.ExecutionScope, "report", []artifactmanifest.Claim{{Name: "report.txt", Receipt: stored}}, store)
			if err != nil || manifest.TotalBytes != int64(len("actual synthetic report\n")) || manifest.Scope.RevisionDigest != source.Snapshot().Scope.RevisionDigest {
				t.Fatal("actual source did not form finite deliverable", err)
			}
			if err := artifactmanifest.Revalidate(ctx, definition, scope.ExecutionScope, "report", manifest, store); err != nil {
				t.Fatal("stored manifest not revalidated", err)
			}
			entries, err := os.ReadDir(objectDir)
			if err != nil || len(entries) != 1 {
				t.Fatal("object inode missing", err)
			}
		})
	}
}

func TestSourceChangeDuringCopyCannotSealReceipt(t *testing.T) {
	t.Parallel()
	s, _, ctx, scope := artifactJournalFixture(t, true)
	dir := t.TempDir()
	if os.Chmod(dir, 0700) != nil {
		t.Fatal("root mode")
	}
	if os.WriteFile(filepath.Join(dir, "report.txt"), []byte("synthetic report\n"), 0644) != nil {
		t.Fatal("source fixture")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	e := scope.ExecutionScope
	e.RevisionDigest = ""
	source, err := workspacesnapshot.Capture(ctx, root, workspacesnapshot.Binding{Execution: e, BaseCommit: strings.Repeat("a", 40), VolumeName: "commonroom-tool-work-" + e.JobID, VolumeNonce: strings.Repeat("b", 32), VolumeCreatedAt: time.Unix(123, 0).UTC().Format(time.RFC3339Nano)}, []workspacesnapshot.Member{{OutputName: "report", Name: "report.txt", MaxBytes: 128}}, &syntheticFrozenArtifactLease{})
	if err != nil {
		t.Fatal(err)
	}
	scope.ExecutionScope = source.Snapshot().Scope
	guard := &syntheticArtifactGuard{source: func(c context.Context, _ room.State, _ objectcustody.Intent) error { return source.Current(c) }, seal: func(c context.Context, _ room.State, r objectcustody.Receipt) error {
		return source.VerifyReceipt(c, r)
	}}
	journal, _ := newControllerArtifactJournal(s, guard)
	store, objectDir := artifactStore(t, journal)
	stream, err := source.Open(ctx, scope, 128)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	wrapped := &mutatingArtifactSource{ReadCloser: stream, mutate: func() {
		if err := os.WriteFile(filepath.Join(dir, "report.txt"), []byte("changed after source read\n"), 0644); err != nil {
			t.Error(err)
		}
	}}
	if _, err := store.Put(ctx, scope, 128, wrapped); err == nil {
		t.Fatal("changed source sealed")
	}
	st := artifactState(t, s, ctx)
	if artifactTotal(t, s, ctx) != 128 || len(st.ArtifactCustody) != 1 {
		t.Fatal("uncertain copy lost charge")
	}
	for _, r := range st.ArtifactCustody {
		if r.Receipt != nil {
			t.Fatal("changed source receipt persisted")
		}
	}
	entries, _ := os.ReadDir(objectDir)
	if len(entries) != 1 {
		t.Fatal("uncertain bytes deleted")
	}
}

type mutatingArtifactSource struct {
	io.ReadCloser
	mutate func()
}

func (r *mutatingArtifactSource) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if n > 0 && r.mutate != nil {
		r.mutate()
		r.mutate = nil
	}
	return n, err
}
