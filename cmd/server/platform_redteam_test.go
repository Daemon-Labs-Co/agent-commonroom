package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// RT-1: a captured platform request can be replayed because replay memory is keyed on
// the raw Base64 header text, and Go's StdEncoding decoder accepts non-zero padding bits.
func TestRedteamPlatformReplayViaNonCanonicalBase64(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	stamp := strconv.FormatInt(f.now.Unix(), 10)
	list := "/platform/v1/workspaces"
	sig := ed25519.Sign(f.priv, platformSigningInput(f.keyID, stamp, "GET", list, nil))
	canonical := base64.StdEncoding.EncodeToString(sig)
	// 64 bytes encode to 86 chars + "=="; the last char carries 2 data bits and 4 ignored bits.
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	last := canonical[85]
	idx := -1
	for i := range alphabet {
		if alphabet[i] == last {
			idx = i
		}
	}
	variant := canonical[:85] + string(alphabet[idx^1]) + "=="
	if variant == canonical {
		t.Fatal("test bug")
	}
	if raw, err := base64.StdEncoding.DecodeString(variant); err != nil || string(raw) != string(sig) {
		t.Skip("decoder is strict")
	}
	do := func(header string) int {
		r := httptest.NewRequest("GET", "https://platform.rooms.example.com"+list, nil)
		r.Host = "platform.rooms.example.com"
		r.Header.Set("X-Platform-Id", f.keyID)
		r.Header.Set("X-Platform-Timestamp", stamp)
		r.Header.Set("X-Platform-Signature", header)
		w := httptest.NewRecorder()
		f.host.ServeHTTP(w, r)
		return w.Code
	}
	if c := do(canonical); c != http.StatusOK {
		t.Fatal("original request refused", c)
	}
	if c := do(canonical); c != 429 {
		t.Fatal("exact replay not refused", c)
	}
	if c := do(variant); c != 429 {
		t.Fatalf("replay with re-encoded signature accepted: status %d, want 429", c)
	}
}

// RT-5: reissuing a setup link must cut off whoever already signed in with the old code.
func TestRedteamReissuedSetupLinkLeavesOldBootstrapSessionAlive(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	rt := f.runtime("acme")
	oldCode, _, pending := rt.s.auth.setupSnapshot()
	if !pending {
		t.Fatal("setup should be pending")
	}
	// An attacker who obtained the leaked setup link signs in as owner.
	w := f.request("acme", "POST", "/api/login", map[string]string{"key": oldCode}, nil, nil)
	attacker := sessionFrom(t, w)
	if c := f.request("acme", "GET", "/api/session", nil, &attacker, nil).Code; c != 200 {
		t.Fatal("attacker session should work before reissue", c)
	}
	// The operator notices and reissues the link.
	if r := f.call("POST", "/platform/v1/workspaces/acme/setup-link", ""); r.Code != 200 {
		t.Fatal("reissue", r.Code, r.Body.String())
	}
	if c := f.request("acme", "GET", "/api/session", nil, &attacker, nil).Code; c == 200 {
		t.Fatalf("session created with the revoked setup code still works after reissue (status %d)", c)
	}
}

// RT-8: platformRequestsPerMin (120) is declared but never used; the platform API is held
// to the extension limit of 30 a minute, so a control plane provisioning in bulk is refused.
func TestRedteamPlatformRateLimitIsTheDeclaredOne(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	for i := 1; i <= platformRequestsPerMin; i++ {
		// Distinct signatures: vary the query so each request is different and fresh.
		w := f.call("GET", "/platform/v1/workspaces?limit="+strconv.Itoa(1+i%100)+"&after=a"+strconv.Itoa(i), "")
		if w.Code == 429 {
			t.Fatalf("request %d of the declared %d per minute was refused with 429", i, platformRequestsPerMin)
		}
	}
}
