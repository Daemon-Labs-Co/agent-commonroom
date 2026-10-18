package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// Real credential lease/provision/stage/release operations, with temporary
// private files and a pre-write filesystem failure. No observations come from
// scenario.Want, provider processes, Docker or database fixtures.
func driveP0CredentialRefresh(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	ctx, _, scope, root, private, record, journal := credentialLeaseFixture(t)
	lease, err := acquireWorkerEngineCredential(ctx, scope, root, journal)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release(record)
	before, err := readPrivateBounded(filepath.Join(private, "auth.json"), 65536)
	if err != nil {
		t.Fatal(err)
	}
	var frame bytes.Buffer
	if err = lease.Provision(ctx, &frame); err != nil {
		t.Fatal(err)
	}
	out := pendingP0Observation{State: "deny", Values: map[string]string{}}
	refresh := []byte(syntheticRefreshedCredential)
	if scenario.ID == "refresh-foreign-connection" {
		foreign := scope
		foreign.Connection = "foreign-connection"
		if err = lease.StageRefresh(foreign, refresh); err == nil {
			out.State = "allow"
		}
		saved, _ := readPrivateBounded(filepath.Join(private, "auth.json"), 65536)
		if !bytes.Equal(saved, before) {
			out.Effects = append(out.Effects, "credential_replaced")
		}
		return out
	}
	overlaps := 0
	if scenario.ID == "refresh-concurrent" {
		// The second worker contends while the first owns provisioning custody.
		if second, e := acquireWorkerEngineCredential(ctx, scope, root, journal); e == nil {
			overlaps++
			second.Release(record)
		}
	}
	if err = lease.StageRefresh(scope, refresh); err != nil {
		t.Fatal(err)
	}
	if scenario.ID == "refresh-atomic" {
		out.Effects = append(out.Effects, "private_refresh_validated")
	}
	if scenario.ID == "refresh-persist-fails" {
		err = lease.releaseWithPersistence(record, func(path string, raw []byte) error {
			if path != filepath.Join(private, "auth.json") || !bytes.Equal(raw, refresh) {
				t.Fatal("foreign credential persistence")
			}
			return errors.New("synthetic pre-write failure")
		})
	} else {
		err = lease.Release(record)
	}
	saved, readErr := readPrivateBounded(filepath.Join(private, "auth.json"), 65536)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err != nil {
		out.State = "block"
	} else {
		out.State = "allow"
	}
	if scenario.ID == "refresh-persist-fails" {
		if !workerCredentialReady(private) {
			out.Effects = append(out.Effects, "auth_recovery_required")
		}
		if err == nil {
			out.Effects = append(out.Effects, "refresh_success_reported")
		}
		out.Values["prior_credential_preserved"] = strconv.FormatBool(bytes.Equal(saved, before))
		return out
	}
	if scenario.ID == "refresh-atomic" {
		if bytes.Equal(saved, refresh) && workerCredentialReady(private) {
			out.Effects = append(out.Effects, "credential_persisted")
		}
		info, e := os.Stat(filepath.Join(private, "auth.json"))
		if e != nil {
			t.Fatal(e)
		}
		out.Values["mode"] = fmt.Sprintf("%04o", info.Mode().Perm())
		if bytes.Equal(saved, refresh) {
			out.Values["connection"] = "original"
		}
		unlock, e := lockWorkerCredential(private)
		if e == nil && lease.released {
			out.Effects = append(out.Effects, "credential_lease_released")
		}
		if unlock != nil {
			unlock()
		}
	} else if scenario.ID == "refresh-concurrent" {
		second, e := acquireWorkerEngineCredential(ctx, scope, root, journal)
		lost := 0
		if e != nil {
			lost++
		} else {
			if !bytes.Equal(second.raw, refresh) {
				lost++
			}
			second.Release(record)
		}
		out.Values["overlapping_refreshes"] = strconv.Itoa(overlaps)
		out.Values["lost_updates"] = strconv.Itoa(lost)
		if overlaps == 0 && lost == 0 {
			out.Effects = append(out.Effects, "refresh_serialized")
		}
	}
	return out
}
