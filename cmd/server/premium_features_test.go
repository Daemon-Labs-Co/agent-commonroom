package main

import (
	"agentroom/internal/room"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func syntheticPremiumLicense(t *testing.T, features []string) (*premiumEntitlements, premiumLicenseClaims, ed25519.PrivateKey) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p := &premiumEntitlements{publicKey: public, installationID: strings.Repeat("b", 32), licensePath: filepath.Join(t.TempDir(), "license.json")}
	claims := premiumLicenseClaims{Version: 1, Product: "agent-commonroom", InstallationID: p.installationID, Features: features, LicenseID: "synthetic-license", NotBefore: time.Now().Add(-time.Hour), ExpiresAt: time.Now().Add(time.Hour)}
	writeSyntheticPremium(t, p, claims, private)
	return p, claims, private
}
func writeSyntheticPremium(t *testing.T, p *premiumEntitlements, claims premiumLicenseClaims, private ed25519.PrivateKey) []byte {
	t.Helper()
	payload, _ := json.Marshal(claims)
	raw, _ := json.Marshal(map[string]string{"payload": base64.StdEncoding.EncodeToString(payload), "signature": base64.StdEncoding.EncodeToString(ed25519.Sign(private, payload))})
	if err := os.WriteFile(p.licensePath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return raw
}
func TestPremiumLicenseRejectsForgeryScopeAndInvalidValidity(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"active", "expired", "future", "wrong-installation", "wrong-product", "unknown-feature", "duplicate-feature", "invalid-window", "signature", "unsigned", "oversized"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			p, claims, key := syntheticPremiumLicense(t, []string{protectedAuditArchive})
			now := time.Now()
			switch mode {
			case "expired":
				claims.ExpiresAt = now
			case "future":
				claims.NotBefore = now.Add(time.Minute)
			case "wrong-installation":
				claims.InstallationID = "other"
			case "wrong-product":
				claims.Product = "other"
			case "unknown-feature":
				claims.Features = []string{"ownerPermissions"}
			case "duplicate-feature":
				claims.Features = []string{protectedAuditArchive, protectedAuditArchive}
			case "invalid-window":
				claims.ExpiresAt = claims.NotBefore
			}
			raw := writeSyntheticPremium(t, p, claims, key)
			if mode == "signature" {
				_, foreign, _ := ed25519.GenerateKey(rand.Reader)
				raw = writeSyntheticPremium(t, p, claims, foreign)
			}
			if mode == "unsigned" {
				raw = []byte(`{"paid":true}`)
			}
			if mode == "oversized" {
				raw = make([]byte, premiumLicenseLimit+1)
			}
			_, status := verifyPremiumLicense(raw, p.publicKey, p.installationID, now)
			if (status == "active") != (mode == "active") {
				t.Fatal(mode, status)
			}
		})
	}
}
func TestPremiumFeatureFlagRequiresCurrentGrantAndOwner(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	enable := func(feature string, on bool) int {
		raw, _ := json.Marshal(map[string]any{"kind": "feature.flag.updated", "payload": map[string]any{"name": feature, "enabled": on, "paid": true}})
		return mockRequest(h, "POST", "/api/events", string(raw)).Code
	}
	if enable(protectedAuditArchive, true) != 402 || len(db.snapshot()) != 0 {
		t.Fatal("unpaid browser enabled flag")
	}
	p, claims, key := syntheticPremiumLicense(t, []string{protectedAuditArchive})
	s.premium = p
	if enable(managedOrchestration, true) != 402 {
		t.Fatal("unrelated entitlement enabled orchestration")
	}
	if enable(protectedAuditArchive, true) != 200 {
		t.Fatal("licensed owner denied")
	}
	st, _ := s.load(context.Background())
	if !s.premiumFeatureInfo(st, protectedAuditArchive, time.Now()).Enabled {
		t.Fatal("enabled grant missing")
	}
	claims.ExpiresAt = time.Now().Add(-time.Second)
	writeSyntheticPremium(t, p, claims, key)
	if s.premiumFeatureInfo(st, protectedAuditArchive, time.Now()).Enabled || enable(protectedAuditArchive, true) != 402 {
		t.Fatal("expired grant accepted")
	}
	if enable(protectedAuditArchive, false) != 200 {
		t.Fatal("expired owner could not disable saved preference")
	}
	claims.ExpiresAt = time.Now().Add(time.Hour)
	writeSyntheticPremium(t, p, claims, key)
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if enable(protectedAuditArchive, true) != 403 {
		t.Fatal("non-owner enabled paid feature")
	}
}
func TestPremiumBoundaryDeniesUnpaidDisabledAndUnimplemented(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/premium-features", s.premiumFeatures)
	mux.HandleFunc("/api/protected-audit-archive", s.protectedArchiveStatus)
	mux.HandleFunc("/api/managed-orchestration", s.managedOrchestrationStatus)
	h := s.auth.protect(s.actorGuard(mux))
	if mockRequest(h, "GET", "/api/protected-audit-archive", "").Code != 402 {
		t.Fatal("unpaid action accepted")
	}
	p, _, _ := syntheticPremiumLicense(t, []string{protectedAuditArchive, managedOrchestration})
	s.premium = p
	for _, feature := range premiumFeatures {
		route := "/api/protected-audit-archive"
		if feature == managedOrchestration {
			route = "/api/managed-orchestration"
		}
		if mockRequest(h, "GET", route, "").Code != 403 {
			t.Fatal("disabled action accepted")
		}
		db.seed("feature.flag.updated", map[string]any{"name": feature, "enabled": true})
		if mockRequest(h, "GET", route, "").Code != 501 {
			t.Fatal("placeholder claimed implementation")
		}
	}
	w := mockRequest(h, "GET", "/api/premium-features", "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "signature") || strings.Contains(w.Body.String(), p.installationID) {
		t.Fatal("unsafe license projection")
	}
	db.seed("actor.revised", room.Actor{ID: "qa", Kind: "human", Role: "QA", Enabled: true})
	s.auth.sessions["owner"] = ownerSession{ActorID: "qa", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if mockRequest(h, "GET", "/api/premium-features", "").Code != 403 {
		t.Fatal("non-owner read licensing")
	}
}
func TestPremiumConfigurationAndFilesFailClosed(t *testing.T) {
	t.Parallel()
	p, _, _ := syntheticPremiumLicense(t, []string{protectedAuditArchive})
	dir := t.TempDir()
	pub := base64.StdEncoding.EncodeToString(p.publicKey)
	configured, err := configurePremiumEntitlements(pub, p.licensePath, dir)
	if err != nil {
		t.Fatal(err)
	}
	again, err := configurePremiumEntitlements(pub, p.licensePath, dir)
	if err != nil || again.installationID != configured.installationID {
		t.Fatal("installation identity changed")
	}
	if _, err = configurePremiumEntitlements("invalid", p.licensePath, dir); err == nil {
		t.Fatal("bad issuer accepted")
	}
	if none, err := configurePremiumEntitlements("", p.licensePath, dir); err != nil || none != nil {
		t.Fatal("unconfigured not locked")
	}
	if err = os.Chmod(p.licensePath, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = readPremiumFile(p.licensePath); err == nil {
		t.Fatal("public license permissions accepted")
	}
	link := filepath.Join(t.TempDir(), "link")
	if err = os.Symlink(p.licensePath, link); err != nil {
		t.Fatal(err)
	}
	if _, err = readPremiumFile(link); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestPremiumArchiveAllowancesAreSignedAndSeparateFromOrchestration(t *testing.T) {
	t.Parallel()
	s, _, _ := mockServer(t)
	p, claims, key := syntheticPremiumLicense(t, []string{protectedAuditArchive, managedOrchestration})
	s.premium = p
	claims.ArchiveAdditionalBlocks = 2
	writeSyntheticPremium(t, p, claims, key)
	info := s.premiumFeatureInfo(room.State{}, protectedAuditArchive, time.Now())
	if info.MonthlyRecordAllowance != 300000 || info.IncludedMonthlyRecords != 100000 || info.AdditionalBlockRecords != 100000 || info.PurchasedAdditionalBlocks != 2 {
		t.Fatal("archive allowance incorrect", info)
	}
	if s.premiumFeatureInfo(room.State{}, managedOrchestration, time.Now()).MonthlyRecordAllowance != 0 {
		t.Fatal("archive quota became AI allowance")
	}
	for _, blocks := range []int{-1, 1001} {
		claims.ArchiveAdditionalBlocks = blocks
		writeSyntheticPremium(t, p, claims, key)
		if s.premiumFeatureInfo(room.State{}, protectedAuditArchive, time.Now()).Entitled {
			t.Fatal("invalid blocks accepted")
		}
	}
	claims.ArchiveAdditionalBlocks = 1
	claims.Features = []string{managedOrchestration}
	writeSyntheticPremium(t, p, claims, key)
	if s.premiumFeatureInfo(room.State{}, managedOrchestration, time.Now()).Entitled {
		t.Fatal("archive blocks granted without archive feature")
	}
}
