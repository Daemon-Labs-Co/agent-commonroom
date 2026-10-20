package main

import (
	"agentroom/internal/harnessrelease"
	"agentroom/internal/room"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"time"
)

var errHarnessReview = harnessrelease.ErrReview

type harnessReview = harnessrelease.Review
type harnessReviewEnvelope = harnessrelease.Envelope

func verifyHarnessReview(raw []byte, key ed25519.PublicKey, now time.Time) (room.HarnessSelection, error) {
	return harnessrelease.Verify(raw, key, now)
}
func validHarnessSelection(v room.HarnessSelection) bool { return harnessrelease.ValidSelection(v) }

func validManagedHarness(job room.ManagedJob) bool {
	return job.Harness == (room.HarnessSelection{}) || (job.Version == workerWorkspaceLifecycleVersion && job.Provider == "codex" && validHarnessSelection(job.Harness) && !job.Harness.CheckedAt.Before(job.AllocatedAt) && !job.Harness.CheckedAt.After(job.UpdatedAt))
}

// Selection runs only on new authorized allocations. Off leaves the existing
// explicit immutable-image contract intact; it never enables public dispatch.
func selectHarnessReview(enabled string, source func() ([]byte, error), key ed25519.PublicKey, now time.Time, jobs map[string]room.ManagedJob) (room.HarnessSelection, error) {
	if enabled == "" || enabled == "false" {
		return room.HarnessSelection{}, nil
	}
	if enabled != "true" || source == nil {
		return room.HarnessSelection{}, errHarnessReview
	}
	raw, err := source()
	if err != nil {
		return room.HarnessSelection{}, errHarnessReview
	}
	selected, err := verifyHarnessReview(raw, key, now)
	if err != nil {
		return room.HarnessSelection{}, err
	}
	for _, prior := range jobs {
		if prior.Harness.Sequence > selected.Sequence || (prior.Harness.Sequence == selected.Sequence && prior.Harness.ReviewDigest != selected.ReviewDigest) {
			return room.HarnessSelection{}, errHarnessReview
		}
	}
	return selected, nil
}
func deploymentHarnessReview(now time.Time, jobs map[string]room.ManagedJob) (room.HarnessSelection, error) {
	enabled := os.Getenv("COMMONROOM_HARNESS_UPGRADE_CHECKS")
	// Disabled mode must not read files or require a review key.
	if enabled == "" || enabled == "false" {
		return room.HarnessSelection{}, nil
	}
	if !filepath.IsAbs(os.Getenv("COMMONROOM_HARNESS_REVIEW_FILE")) {
		return room.HarnessSelection{}, errHarnessReview
	}
	key, err := base64.StdEncoding.Strict().DecodeString(os.Getenv("COMMONROOM_HARNESS_REVIEW_KEY"))
	if err != nil {
		return room.HarnessSelection{}, errHarnessReview
	}
	return selectHarnessReview(enabled, func() ([]byte, error) { return readPrivateBounded(os.Getenv("COMMONROOM_HARNESS_REVIEW_FILE"), 4096) }, ed25519.PublicKey(key), now, jobs)
}

// The allocation is trusted authenticated controller input, never model input.
// Recheck expiry only before startup; a running engine is never upgraded here.
func selectedHarnessImage(job room.ManagedJob, configured string, now time.Time) (string, error) {
	if job.Harness == (room.HarnessSelection{}) {
		return configured, nil
	}
	if !validManagedHarness(job) || !job.Harness.ExpiresAt.After(now) || now.Before(job.Harness.CheckedAt) || (configured != "" && configured != job.Harness.Image) {
		return "", errHarnessReview
	}
	return job.Harness.Image, nil
}
