package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"agentroom/internal/room"
)

const protectedAuditArchive = "protectedAuditArchive"
const managedOrchestration = "managedOrchestration"
const archiveRecordBlock int64 = 100000

var premiumFeatures = []string{protectedAuditArchive, managedOrchestration}

func knownPremiumFeature(name string) bool {
	for _, feature := range premiumFeatures {
		if feature == name {
			return true
		}
	}
	return false
}

const premiumLicenseLimit = 8192

var premiumIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

type premiumEntitlements struct {
	publicKey      ed25519.PublicKey
	installationID string
	licensePath    string
}

// premiumLicenseClaims are issuer-attested commercial grants, not owner editable
// flags or payment receipts. The private signing key never belongs in this app.
type premiumLicenseClaims struct {
	Version                 int       `json:"version"`
	Product                 string    `json:"product"`
	InstallationID          string    `json:"installationId"`
	Features                []string  `json:"features"`
	LicenseID               string    `json:"licenseId"`
	NotBefore               time.Time `json:"notBefore"`
	ExpiresAt               time.Time `json:"expiresAt"`
	ArchiveAdditionalBlocks int       `json:"archiveAdditionalBlocks,omitempty"`
}

type premiumFeatureInfo = room.PremiumFeature

func strictPremiumJSON(raw []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return errors.New("unexpected license data")
	}
	return nil
}

// verifyPremiumLicense checks the exact signed bytes, product, installation,
// feature and validity window on every access. No browser-provided paid bit counts.
func verifyPremiumLicense(raw []byte, publicKey ed25519.PublicKey, installation string, now time.Time) (premiumLicenseClaims, string) {
	var claims premiumLicenseClaims
	var envelope struct {
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	}
	if len(raw) > premiumLicenseLimit || len(publicKey) != ed25519.PublicKeySize || strictPremiumJSON(raw, &envelope) != nil {
		return claims, "invalid"
	}
	payload, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		return claims, "invalid"
	}
	signature, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil || !ed25519.Verify(publicKey, payload, signature) || strictPremiumJSON(payload, &claims) != nil {
		return premiumLicenseClaims{}, "invalid"
	}
	if claims.Version != 1 || claims.Product != "agent-commonroom" || len(claims.Features) == 0 || len(claims.Features) > 16 || claims.InstallationID != installation || !premiumIDPattern.MatchString(claims.LicenseID) || claims.NotBefore.IsZero() || !claims.ExpiresAt.After(claims.NotBefore) {
		return premiumLicenseClaims{}, "invalid"
	}
	seen := map[string]bool{}
	for _, feature := range claims.Features {
		if !knownPremiumFeature(feature) || seen[feature] {
			return premiumLicenseClaims{}, "invalid"
		}
		seen[feature] = true
	}
	if claims.ArchiveAdditionalBlocks < 0 || claims.ArchiveAdditionalBlocks > 1000 || claims.ArchiveAdditionalBlocks > 0 && !seen[protectedAuditArchive] {
		return premiumLicenseClaims{}, "invalid"
	}
	if now.Before(claims.NotBefore) {
		return claims, "not_yet_valid"
	}
	if !now.Before(claims.ExpiresAt) {
		return claims, "expired"
	}
	return claims, "active"
}

func readPremiumFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > premiumLicenseLimit {
		return nil, errors.New("unsafe premium file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, premiumLicenseLimit+1))
	if len(raw) > premiumLicenseLimit {
		return nil, errors.New("oversized premium file")
	}
	return raw, err
}

// configurePremiumEntitlements trusts only deployment-controlled issuer material.
// Source/self-host administrators CAN replace it, BUT it is not a secure DRM
// boundary against operators who control this binary, filesystem and environment.
func configurePremiumEntitlements(publicKey, licensePath, privateDirectory string) (*premiumEntitlements, error) {
	if publicKey == "" {
		return nil, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, errors.New("premium issuer key must be a base64 Ed25519 public key")
	}
	identityPath := filepath.Join(privateDirectory, "premium-installation-id")
	raw, err := readPremiumFile(identityPath)
	if os.IsNotExist(err) {
		raw = []byte(id())
		if err = writePrivateFile(identityPath, raw); err != nil {
			return nil, errors.New("private premium installation identity unavailable")
		}
	} else if err != nil {
		return nil, errors.New("private premium installation identity unavailable")
	}
	if len(raw) != 32 || !workerJobPattern.MatchString(string(raw)) {
		return nil, errors.New("invalid premium installation identity")
	}
	if licensePath == "" {
		licensePath = filepath.Join(privateDirectory, "premium-license.json")
	}
	return &premiumEntitlements{publicKey: decoded, installationID: string(raw), licensePath: licensePath}, nil
}

func (s *server) premiumFeatureInfo(st room.State, feature string, now time.Time) premiumFeatureInfo {
	info := premiumFeatureInfo{Requested: st.FeatureFlags[feature], Status: "unconfigured"}
	if feature == protectedAuditArchive {
		info.IncludedMonthlyRecords = archiveRecordBlock
		info.AdditionalBlockRecords = archiveRecordBlock
	}
	if s.premium == nil {
		return info
	}
	raw, err := readPremiumFile(s.premium.licensePath)
	if os.IsNotExist(err) {
		info.Status = "purchase_required"
		return info
	}
	if err != nil {
		info.Status = "invalid"
		return info
	}
	claims, status := verifyPremiumLicense(raw, s.premium.publicKey, s.premium.installationID, now)
	info.Status = status
	info.Entitled = false
	if status == "active" {
		for _, granted := range claims.Features {
			if granted == feature {
				info.Entitled = true
			}
		}
		if !info.Entitled {
			info.Status = "purchase_required"
		}
	}
	info.Enabled = info.Requested && info.Entitled
	if feature == protectedAuditArchive && info.Entitled {
		info.PurchasedAdditionalBlocks = claims.ArchiveAdditionalBlocks
		info.MonthlyRecordAllowance = archiveRecordBlock * int64(1+claims.ArchiveAdditionalBlocks)
	}
	if status == "active" || status == "expired" || status == "not_yet_valid" {
		info.ExpiresAt = &claims.ExpiresAt
	}
	// Available remains false until a verified archive adapter is implemented.
	return info
}

func (s *server) premiumFeatures(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	result := map[string]premiumFeatureInfo{}
	for _, feature := range premiumFeatures {
		result[feature] = s.premiumFeatureInfo(st, feature, time.Now())
	}
	reply(w, result)
}

// requireProtectedAuditArchive is the mandatory gate for future archive actions:
// owner + current entitlement + saved flag. No caller may rely on UI visibility.
func (s *server) requirePremiumFeature(w http.ResponseWriter, r *http.Request, st room.State, feature string) bool {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return false
	}
	info := s.premiumFeatureInfo(st, feature, time.Now())
	if !info.Entitled {
		http.Error(w, "A valid paid feature entitlement is required", 402)
		return false
	}
	if !info.Enabled {
		http.Error(w, "Premium feature is disabled", 403)
		return false
	}
	return true
}

// protectedArchiveStatus reserves the gated integration boundary; it never
// exports data or claims archival protection until the adapter exists.
func (s *server) protectedArchiveStatus(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	if !s.requirePremiumFeature(w, r, st, protectedAuditArchive) {
		return
	}
	http.Error(w, "Archive adapter is not implemented; no external audit protection is active", 501)
}

func (s *server) managedOrchestrationStatus(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	if !s.requirePremiumFeature(w, r, st, managedOrchestration) {
		return
	}
	http.Error(w, "Managed orchestration is not implemented; no internal planning model is active", 501)
}
