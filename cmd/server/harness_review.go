package main

import (
	"agentroom/internal/room"
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var errHarnessReview = errors.New("harness release review unavailable, invalid or incompatible")
var harnessVersionPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}$`)
var harnessImagePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9./:_-]{0,254}@sha256:[a-f0-9]{64}$`)

// Only trusted deployment administrators supply reviews and pin signing keys.
// This source-level seam CAN be weakened by forks, BUT then image approval is
// untrusted. A signature proves review authority, not image safety or provenance.
type harnessReview struct {
	Schema            int       `json:"schema"`
	Sequence          int64     `json:"sequence"`
	Provider          string    `json:"provider"`
	Version           string    `json:"version"`
	Image             string    `json:"image"`
	Profile           string    `json:"profile"`
	AuthProtocol      int       `json:"authProtocol"`
	AuditProtocol     int       `json:"auditProtocol"`
	WorkspaceProtocol int       `json:"workspaceProtocol"`
	IssuedAt          time.Time `json:"issuedAt"`
	ExpiresAt         time.Time `json:"expiresAt"`
}
type harnessReviewEnvelope struct {
	Review    json.RawMessage `json:"review"`
	Signature string          `json:"signature"`
}

// Canonical bytes remove ambiguous duplicate, unknown and case-folded JSON
// fields. The signed payload must be exactly json.Marshal(harnessReview).
func canonicalHarnessJSON(raw []byte, value any) bool {
	if len(raw) == 0 || len(raw) > 4096 || json.Unmarshal(raw, value) != nil {
		return false
	}
	canonical, err := json.Marshal(value)
	return err == nil && bytes.Equal(raw, canonical)
}
func verifyHarnessReview(raw []byte, key ed25519.PublicKey, now time.Time) (room.HarnessSelection, error) {
	var envelope harnessReviewEnvelope
	var review harnessReview
	if len(key) != ed25519.PublicKeySize || !canonicalHarnessJSON(raw, &envelope) || !canonicalHarnessJSON(envelope.Review, &review) {
		return room.HarnessSelection{}, errHarnessReview
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(envelope.Signature)
	if err != nil || !ed25519.Verify(key, envelope.Review, signature) || review.Schema != 1 || review.Provider != "codex" || review.Profile != "codex-separated-callback-v1" || review.AuthProtocol != 1 || review.AuditProtocol != 2 || review.WorkspaceProtocol != 4 || review.IssuedAt.IsZero() || review.IssuedAt.After(now) || !review.ExpiresAt.After(now) || !review.ExpiresAt.After(review.IssuedAt) {
		return room.HarnessSelection{}, errHarnessReview
	}
	digest := sha256.Sum256(envelope.Review)
	selection := room.HarnessSelection{Sequence: review.Sequence, Version: review.Version, Image: review.Image, ReviewDigest: hex.EncodeToString(digest[:]), CheckedAt: now.UTC(), ExpiresAt: review.ExpiresAt}
	if !validHarnessSelection(selection) {
		return room.HarnessSelection{}, errHarnessReview
	}
	return selection, nil
}
func validHarnessSelection(v room.HarnessSelection) bool {
	return v.Sequence > 0 && harnessVersionPattern.MatchString(v.Version) && len(v.Image) <= 327 && harnessImagePattern.MatchString(v.Image) && repositoryDigestPattern.MatchString(v.ReviewDigest) && !v.CheckedAt.IsZero() && v.ExpiresAt.After(v.CheckedAt)
}
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
