// Package harnessrelease verifies the bounded, operator-signed release contract.
// It grants no launch authority and does not prove image contents or compatibility.
package harnessrelease

import (
	"agentroom/internal/room"
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

var digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

var ErrReview = errors.New("harness release review unavailable, invalid or incompatible")
var versionPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}$`)
var imagePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9./:_-]{0,254}@sha256:[a-f0-9]{64}$`)

// Only trusted deployment administrators supply reviews and pin signing keys.
// This source-level seam CAN be weakened by forks, BUT then image approval is
// untrusted. A signature proves review authority, not image safety or provenance.
type Review struct {
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
type Envelope struct {
	Review    json.RawMessage `json:"review"`
	Signature string          `json:"signature"`
}

// Canonical bytes remove ambiguous duplicate, unknown and case-folded JSON
// fields. The signed payload must be exactly json.Marshal(Review).
func canonicalJSON(raw []byte, value any) bool {
	if len(raw) == 0 || len(raw) > 4096 || json.Unmarshal(raw, value) != nil {
		return false
	}
	canonical, err := json.Marshal(value)
	return err == nil && bytes.Equal(raw, canonical)
}
func Verify(raw []byte, key ed25519.PublicKey, now time.Time) (room.HarnessSelection, error) {
	var envelope Envelope
	var review Review
	if len(key) != ed25519.PublicKeySize || !canonicalJSON(raw, &envelope) || !canonicalJSON(envelope.Review, &review) {
		return room.HarnessSelection{}, ErrReview
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(envelope.Signature)
	if err != nil || !ed25519.Verify(key, envelope.Review, signature) || review.Schema != 1 || review.Provider != "codex" || review.Profile != "codex-separated-callback-v1" || review.AuthProtocol != 1 || review.AuditProtocol != 2 || review.WorkspaceProtocol != 4 || review.IssuedAt.IsZero() || review.IssuedAt.After(now) || !review.ExpiresAt.After(now) || !review.ExpiresAt.After(review.IssuedAt) {
		return room.HarnessSelection{}, ErrReview
	}
	digest := sha256.Sum256(envelope.Review)
	selection := room.HarnessSelection{Sequence: review.Sequence, Version: review.Version, Image: review.Image, ReviewDigest: hex.EncodeToString(digest[:]), CheckedAt: now.UTC(), ExpiresAt: review.ExpiresAt}
	if !ValidSelection(selection) {
		return room.HarnessSelection{}, ErrReview
	}
	return selection, nil
}
func ValidSelection(v room.HarnessSelection) bool {
	return v.Sequence > 0 && versionPattern.MatchString(v.Version) && len(v.Image) <= 327 && imagePattern.MatchString(v.Image) && digestPattern.MatchString(v.ReviewDigest) && !v.CheckedAt.IsZero() && v.ExpiresAt.After(v.CheckedAt)
}
