package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/workspacecapture"
	"agentroom/internal/workspacemount"
	"agentroom/internal/workspacesnapshot"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func helperRequest() workspacecapture.Request {
	return workspacecapture.Request{Binding: workspacesnapshot.Binding{Execution: objectcustody.ExecutionScope{CompanyID: "company", RoomID: "engineering", TaskID: "task", ExecutionID: strings.Repeat("a", 32), JobID: strings.Repeat("b", 32), RepositoryDigest: strings.Repeat("c", 64), DefinitionDigest: strings.Repeat("d", 64)}, BaseCommit: strings.Repeat("e", 40), VolumeName: "commonroom-tool-work-" + strings.Repeat("b", 32), VolumeNonce: strings.Repeat("f", 32), VolumeCreatedAt: "2026-10-26T00:00:00Z"}, Members: []workspacesnapshot.Member{{OutputName: "report", Name: "report.txt", MaxBytes: 128}}, Challenge: strings.Repeat("1", 32)}
}
func encodedRequest(r workspacecapture.Request) string {
	raw, _ := json.Marshal(r)
	return base64.RawURLEncoding.EncodeToString(raw)
}

type helperStream struct {
	bytes.Buffer
	closed bool
}

func (s *helperStream) Close() error { s.closed = true; return nil }

type helperAuthority struct{}

func (helperAuthority) AuthorizeFrozenSource(ctx context.Context, _ workspacesnapshot.Binding, _ workspacemount.Proof) error {
	return ctx.Err()
}

type helperJournal struct {
	reserved []objectcustody.Intent
	sealed   []objectcustody.Receipt
}

func (j *helperJournal) Reserve(ctx context.Context, i objectcustody.Intent) error {
	j.reserved = append(j.reserved, i)
	return ctx.Err()
}
func (j *helperJournal) Seal(ctx context.Context, r objectcustody.Receipt) error {
	j.sealed = append(j.sealed, r)
	return ctx.Err()
}
func (*helperJournal) AuthorizeRead(context.Context, objectcustody.Scope, objectcustody.Receipt) error {
	return errors.New("no release")
}
func (*helperJournal) AcknowledgeRead(context.Context, objectcustody.Receipt) error {
	return errors.New("no release")
}
func TestHelperRejectsArgumentsBeforeOpeningRoots(t *testing.T) {
	t.Parallel()
	valid := encodedRequest(helperRequest())
	raw, _ := json.Marshal(helperRequest())
	unknown := base64.RawURLEncoding.EncodeToString(append(append([]byte{}, raw[:len(raw)-1]...), []byte(`,"source":"/host"}`)...))
	duplicate := base64.RawURLEncoding.EncodeToString(append(append([]byte{}, raw[:len(raw)-1]...), []byte(`,"challenge":"`+strings.Repeat("1", 32)+`"}`)...))
	for _, args := range [][]string{nil, {valid, "/host"}, {strings.Repeat("a", maxEncodedRequest+1)}, {"invalid!"}, {valid + "="}, {unknown}, {duplicate}} {
		stream := &helperStream{}
		opened, captured := false, false
		err := run(t.Context(), args, stream, func() (*os.Root, error) { opened = true; return nil, nil }, func(context.Context, workspacecapture.Request, workspacecapture.Authority, *os.Root, objectcustody.Journal) (workspacecapture.CapturedObjects, error) {
			captured = true
			return workspacecapture.CapturedObjects{}, nil
		})
		if err == nil || opened || captured || !stream.closed {
			t.Fatal("invalid helper request used roots/capture", err, opened, captured)
		}
	}
}
func TestHelperSupervisedCustodyAndFinish(t *testing.T) {
	t.Parallel()
	request := helperRequest()
	left, right := net.Pipe()
	dir := t.TempDir()
	if os.Chmod(dir, 0700) != nil {
		t.Fatal("mode")
	}
	journal := &helperJournal{}
	type result struct {
		report workspacecapture.Report
		err    error
	}
	done := make(chan result, 1)
	go func() {
		report, err := workspacecapture.ServeCaptureChannel(t.Context(), right, request, helperAuthority{}, journal)
		done <- result{report, err}
	}()
	capture := func(ctx context.Context, r workspacecapture.Request, authority workspacecapture.Authority, root *os.Root, j objectcustody.Journal) (workspacecapture.CapturedObjects, error) {
		// This CLI test mocks the kernel/capture boundary explicitly. Actual fixed-root
		// scanner plus custody/channel integration is tested in workspacecapture.
		proof := workspacemount.Proof{Version: 1, Challenge: r.Challenge, MountID: 2, ReadOnly: true, RootReadOnly: true}
		if err := authority.AuthorizeFrozenSource(ctx, r.Binding, proof); err != nil {
			return workspacecapture.CapturedObjects{}, err
		}
		scope := r.Binding.Execution
		scope.RevisionDigest = strings.Repeat("8", 64)
		store, err := objectcustody.New(root, j)
		if err != nil {
			return workspacecapture.CapturedObjects{}, err
		}
		receipt, err := store.Put(ctx, objectcustody.Scope{ExecutionScope: scope, OutputName: "report", MemberName: "report.txt"}, 128, strings.NewReader("synthetic helper bytes"))
		if err != nil {
			return workspacecapture.CapturedObjects{}, err
		}
		return workspacecapture.CapturedObjects{Snapshot: workspacesnapshot.Snapshot{Scope: scope, Digest: strings.Repeat("9", 64)}, Receipts: []objectcustody.Receipt{receipt}}, nil
	}
	err := run(t.Context(), []string{encodedRequest(request)}, left, func() (*os.Root, error) { return os.OpenRoot(dir) }, capture)
	served := <-done
	if err != nil || served.err != nil || len(served.report.Receipts) != 1 || len(journal.reserved) != 1 || len(journal.sealed) != 1 {
		t.Fatal("helper lifecycle failed", err, served.err)
	}
	receipt := served.report.Receipts[0]
	data, err := os.ReadFile(filepath.Join(dir, "object-"+receipt.Intent.ObjectID+".blob"))
	if err != nil || string(data) != "synthetic helper bytes" {
		t.Fatal("wrong persisted bytes", err)
	}
}
func TestHelperRootFailureNeverCaptures(t *testing.T) {
	t.Parallel()
	stream := &helperStream{}
	called := false
	err := run(t.Context(), []string{encodedRequest(helperRequest())}, stream, func() (*os.Root, error) { return nil, io.ErrUnexpectedEOF }, func(context.Context, workspacecapture.Request, workspacecapture.Authority, *os.Root, objectcustody.Journal) (workspacecapture.CapturedObjects, error) {
		called = true
		return workspacecapture.CapturedObjects{}, nil
	})
	if err == nil || called || !stream.closed {
		t.Fatal("root failure admitted capture")
	}
}
