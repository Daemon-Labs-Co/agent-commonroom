package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRetainedCaptureRecoveryUsesDurableKnownIDAndNoStarter(t *testing.T) {
	for _, mode := range []string{"present", "already absent", "cleanup retry", "denied", "wrong ID", "unknown ID", "ack failure", "expired owner", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f, l, j := ownedCaptureSetup(t)
			beginCaptureJournal(t, j)
			id := f.base.probe.id
			if mode != "unknown ID" {
				if err := j.BindCapture(t.Context(), j.intent.Request, id); err != nil {
					t.Fatal(err)
				}
			} else {
				id = ""
			}
			if err := j.FinishCapture(t.Context(), j.intent.Request, id, false); err != nil {
				t.Fatal(err)
			}
			// A fresh cleanup guard gets no source/export permission from the old one.
			guard := &captureJournalGuard{f: f.base.probe}
			f.exists = mode != "already absent"
			f.phase = "exited"
			if mode == "denied" {
				guard.deny = "recovery"
			}
			if mode == "wrong ID" || mode == "unknown ID" {
				id = strings.Repeat("9", 64)
			}
			if mode == "ack failure" {
				guard.failAck = "recovery"
			}
			if mode == "cleanup retry" {
				f.mode = "cleanup"
			}
			if mode == "expired owner" {
				l.active = false
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if mode == "canceled" {
				cancel()
			}
			err := recoverRetainedWorkerCapture(ctx, l, j.intent, id, guard)
			if mode == "cleanup retry" {
				if err == nil {
					t.Fatal("failed removal reported success")
				}
				if slots, e := workerCaptureSlots(j.dir); e != nil || slots != 1 {
					t.Fatal("failed cleanup lost capacity", slots, e)
				}
				before := f.base.inspections
				denied := &captureJournalGuard{f: f.base.probe, deny: "recovery"}
				if recoverRetainedWorkerCapture(t.Context(), l, j.intent, id, denied) == nil || f.base.inspections != before {
					t.Fatal("retained recovery bypassed fresh cleanup authority")
				}
				f.mode = ""
				guard = &captureJournalGuard{f: f.base.probe}
				err = recoverRetainedWorkerCapture(t.Context(), l, j.intent, id, guard)
			}
			healthy := mode == "present" || mode == "already absent" || mode == "cleanup retry"
			if (err == nil) != healthy {
				t.Fatal(mode, err)
			}
			want := 1
			if healthy {
				want = 0
			}
			if slots, e := workerCaptureSlots(j.dir); e != nil || slots != want {
				t.Fatal("wrong retained helper charge", slots, e)
			}
			if f.starts != 0 || f.reserves != 0 || f.seals != 0 {
				t.Fatal("cleanup created capture/source/custody authority")
			}
			if !healthy && mode != "canceled" && !guard.stopped {
				t.Fatal("failed cleanup retained operation authority")
			}
			if mode == "denied" || mode == "wrong ID" || mode == "unknown ID" || mode == "expired owner" || mode == "ack failure" || mode == "canceled" {
				if f.base.inspections != 0 {
					t.Fatal("daemon inspected before durable cleanup admission")
				}
			}
			if healthy {
				inspections := f.base.inspections
				again := &captureJournalGuard{f: f.base.probe}
				if recoverRetainedWorkerCapture(t.Context(), l, j.intent, id, again) == nil || f.base.inspections != inspections {
					t.Fatal("settled recovery replayed")
				}
			}
			if healthy && f.exists {
				t.Fatal("success while exact helper still present")
			}
			records, e := readWorkerCaptureRecords(j.dir)
			if e != nil {
				t.Fatal(e)
			}
			record := records[captureRecordName(j.intent)]
			if healthy && !captureRecordSettled(record) {
				t.Fatal("known cleanup not durably settled")
			}
			if mode == "ack failure" && !errors.Is(err, errWorkerProbeJournal) {
				t.Fatal("audit failure not retained", err)
			}
		})
	}
}
