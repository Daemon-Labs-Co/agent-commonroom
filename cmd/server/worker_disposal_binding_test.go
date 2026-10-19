package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkspaceDiscardRequiresCompleteExactIdentityBeforeEffects(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"missing", "wrong-name", "wrong-nonce", "wrong-time", "pending-time", "foreign-version", "retain-with-binding", "missing-record"} {
		t.Run(mode, func(t *testing.T) {
			dir, s, adapter := workspacePair(t)
			if err := s.Start(t.Context(), adapter); err != nil {
				t.Fatal(err)
			}
			s.Close()
			request := workspaceFence(s)
			identity := *request.ExpectedWorkspace
			request.ExpectedWorkspace = &identity
			switch mode {
			case "missing":
				request.ExpectedWorkspace = nil
			case "wrong-name":
				identity.Name = "commonroom-tool-work-" + strings.Repeat("e", 32)
			case "wrong-nonce":
				identity.Nonce = strings.Repeat("e", 32)
			case "wrong-time":
				identity.CreatedAt = time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)
			case "pending-time":
				identity.CreatedAt = ""
			case "foreign-version":
				request.Version = 3
			case "retain-with-binding":
				request.DiscardWorkspace = false
			case "missing-record":
				if err := os.Remove(filepath.Join(dir, s.record.JobID+".json")); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := fenceWorkerJobContainers(t.Context(), dir, request, func(context.Context, workerJobRecord) (workerJobRecord, error) {
				t.Fatal("invalid disposal reached cleanup")
				return workerJobRecord{}, nil
			}); err == nil {
				t.Fatal("invalid disposal accepted")
			}
			if _, err := os.Stat(filepath.Join(dir, s.record.JobID+".cancel")); !os.IsNotExist(err) {
				t.Fatal("invalid disposal created cancellation effects")
			}
		})
	}
}
func TestWorkspaceDiscardRechecksIdentityAfterCreatorLeaseWait(t *testing.T) {
	t.Parallel()
	dir, s, adapter := workspacePair(t)
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	request := workspaceFence(s)
	calls := 0
	_, err := fenceWorkerJobContainers(t.Context(), dir, request, func(ctx context.Context, record workerJobRecord) (workerJobRecord, error) {
		calls++
		if owned, _ := ctx.Value(workerCleanupOwnershipKey{}).(bool); owned {
			t.Fatal("stale disposal reached owned cleanup")
		}
		if expected, _ := ctx.Value(workerWorkspaceDiscardKey{}).(*toolbroker.WorkspaceVolumeIdentity); expected != nil {
			t.Fatal("held creator granted discard")
		}
		s.Close()
		copy := *record.Pair
		workspace := *copy.Workspace
		workspace.Identity.Nonce = strings.Repeat("e", 32)
		copy.Workspace = &workspace
		record.Pair = &copy
		if err := writeWorkerRecord(dir, record, false); err != nil {
			t.Fatal(err)
		}
		return record, nil
	})
	if err == nil || calls != 1 {
		t.Fatal("identity changed during wait was accepted", calls, err)
	}
}
func TestWorkspaceDiscardAcknowledgementBindsRemovedIdentity(t *testing.T) {
	t.Parallel()
	dir, s, adapter := workspacePair(t)
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	request := workspaceFence(s)
	expected, err := request.record()
	if err != nil {
		t.Fatal(err)
	}
	f := newWorkspaceCleanup(t, s)
	s.Close()
	record, err := fenceWorkspace(t.Context(), dir, s, f)
	if err != nil {
		t.Fatal(err)
	}
	ack := workerFenceAcknowledgement(request, record)
	if !validWorkerFenceAck(request, expected, ack) || ack.DiscardedWorkspace == nil {
		t.Fatal("exact disposal receipt rejected")
	}
	for _, mode := range []string{"missing", "wrong-nonce", "wrong-time", "not-removed"} {
		changed := ack
		copy := *ack.DiscardedWorkspace
		changed.DiscardedWorkspace = &copy
		switch mode {
		case "missing":
			changed.DiscardedWorkspace = nil
		case "wrong-nonce":
			copy.Nonce = strings.Repeat("e", 32)
		case "wrong-time":
			copy.CreatedAt = time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)
		case "not-removed":
			changed.WorkspaceRemoved = false
		}
		if validWorkerFenceAck(request, expected, changed) {
			t.Fatal("wrong disposal receipt accepted", mode)
		}
	}
	retain := request
	retain.DiscardWorkspace = false
	retain.ExpectedWorkspace = nil
	if validWorkerFenceAck(retain, expected, ack) {
		t.Fatal("unsolicited disposal receipt accepted")
	}
}
func TestFenceWireRefusesAmbiguousDisposalJSON(t *testing.T) {
	t.Parallel()
	_, s, adapter := workspacePair(t)
	defer s.Close()
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	request := workspaceFence(s)
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = decodeFenceRequest(base64.RawURLEncoding.EncodeToString(raw)); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		strings.Replace(string(raw), `"discardWorkspace":true`, `"discardWorkspace":false,"discardWorkspace":true`, 1),
		strings.Replace(string(raw), `"expectedWorkspace":`, `"ExpectedWorkspace":`, 1),
		strings.Replace(string(raw), `"nonce":`, `"Nonce":`, 1),
		strings.Replace(string(raw), `"nonce":`, `"nonce":"ignored","nonce":`, 1),
		strings.Replace(string(raw), `"createdAt":`, `"extra":0,"createdAt":`, 1),
		strings.Replace(string(raw), `"discardWorkspace":true`, `"discardWorkspace":null`, 1),
		string(raw) + `{}`,
	} {
		if _, err := decodeFenceRequest(base64.RawURLEncoding.EncodeToString([]byte(bad))); err == nil {
			t.Fatal("ambiguous destructive request accepted", bad)
		}
	}
	ackRaw, _ := json.Marshal(workerFenceAcknowledgement(request, s.record))
	if !validFenceWire(ackRaw, true) {
		t.Fatal("valid receipt wire rejected")
	}
	if validFenceWire([]byte(strings.Replace(string(ackRaw), `"stopped":false`, `"stopped":true,"stopped":false`, 1)), true) {
		t.Fatal("duplicate receipt accepted")
	}
}
