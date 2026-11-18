package main

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func envMap(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

var suppliedCode = strings.Repeat("ab", 32)

func TestSuppliedSetupCodeValidationAndExpiry(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 17, 9, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "owner-key")
	code, expires, err := loadSetupCode(path, envMap(map[string]string{setupCodeEnv: " " + suppliedCode + "\n"}), now)
	if err != nil || code != suppliedCode || !expires.Equal(now.Add(24*time.Hour)) {
		t.Fatal(code, expires, err)
	}
	if _, statErr := os.Stat(path); statErr == nil {
		t.Fatal("a supplied code must not write a file")
	}
	if _, expires, _ = loadSetupCode(path, envMap(map[string]string{setupCodeEnv: suppliedCode, setupCodeTTLEnv: "168"}), now); !expires.Equal(now.Add(168 * time.Hour)) {
		t.Fatal("custom lifetime ignored", expires)
	}
	for _, bad := range []string{"short", strings.Repeat("AB", 32), strings.Repeat("g", 64), suppliedCode + "00"} {
		if _, _, err := loadSetupCode(path, envMap(map[string]string{setupCodeEnv: bad}), now); err == nil || strings.Contains(err.Error(), bad) {
			t.Fatal("invalid code accepted or echoed", bad, err)
		}
	}
	for _, hours := range []string{"0", "169", "-1", "x", "1.5"} {
		if _, _, err := loadSetupCode(path, envMap(map[string]string{setupCodeEnv: suppliedCode, setupCodeTTLEnv: hours}), now); err == nil {
			t.Fatal("invalid lifetime accepted", hours)
		}
	}
	// Unset keeps the generated private file, which never expires.
	generated, expires, err := loadSetupCode(path, envMap(nil), now)
	if err != nil || len(generated) != 64 || !expires.IsZero() {
		t.Fatal(generated, expires, err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("generated code was not persisted")
	}
}

func TestSuppliedSetupCodeStopsWorkingWhenExpiredOrClaimed(t *testing.T) {
	t.Parallel()
	login := func(a *ownerAuth, key string) int {
		s := &server{auth: a}
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/login", strings.NewReader(`{"key":"`+key+`"}`))
		r.Host = "127.0.0.1:8787"
		r.Header.Set("Content-Type", "application/json")
		s.login(w, r)
		return w.Code
	}
	newAuth := func(expires time.Time) *ownerAuth {
		a, _ := newOwnerAuth(suppliedCode, "127.0.0.1:8787")
		a.keyExpires = expires
		return a
	}
	if code := login(newAuth(time.Now().Add(time.Hour)), suppliedCode); code != 200 {
		t.Fatal("valid supplied code refused", code)
	}
	if code := login(newAuth(time.Now().Add(-time.Second)), suppliedCode); code != 401 {
		t.Fatal("expired supplied code accepted", code)
	}
	if code := login(newAuth(time.Time{}), suppliedCode); code != 200 {
		t.Fatal("a code with no expiry must keep working", code)
	}
	// Claimed installations clear the code at startup whatever supplied it.
	a := newAuth(time.Now().Add(time.Hour))
	path := filepath.Join(t.TempDir(), "owner-key")
	if err := a.prepareRSA(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path + ".private.pem"); err != nil {
		t.Fatal(err)
	}
	b := newAuth(time.Now().Add(time.Hour))
	if err := b.prepareRSA(path); err != nil || b.key != "" {
		t.Fatal("claimed installation kept a bootstrap code", err)
	}
	if code := login(b, suppliedCode); code != 401 {
		t.Fatal("claimed installation accepted the old setup code", code)
	}
}

func TestSetupStatusReportsExpiredCode(t *testing.T) {
	t.Parallel()
	status := func(expires time.Time, pending bool) string {
		a, _ := newOwnerAuth(suppliedCode, "127.0.0.1:8787")
		a.keyExpires, a.setupPending = expires, pending
		w := httptest.NewRecorder()
		(&server{auth: a}).setupStatus(w, httptest.NewRequest("GET", "/api/setup-status", nil))
		return strings.TrimSpace(w.Body.String())
	}
	if got := status(time.Now().Add(time.Hour), true); got != `{"setupCodeExpired":false,"setupRequired":true}` {
		t.Fatal(got)
	}
	if got := status(time.Now().Add(-time.Hour), true); got != `{"setupCodeExpired":true,"setupRequired":true}` {
		t.Fatal(got)
	}
	if got := status(time.Now().Add(-time.Hour), false); got != `{"setupCodeExpired":false,"setupRequired":false}` {
		t.Fatal(got)
	}
	if strings.Contains(status(time.Now(), true), suppliedCode) {
		t.Fatal("status disclosed the code")
	}
}

func TestSetupLinkCommand(t *testing.T) {
	t.Parallel()
	run := func(values map[string]string, args ...string) (int, string, string) {
		var out, diagnostic bytes.Buffer
		code := setupLinkMain(args, envMap(values), &out, &diagnostic)
		return code, out.String(), diagnostic.String()
	}
	dir := t.TempDir()
	key := filepath.Join(dir, "owner-key")
	generated, err := loadOwnerKey(key)
	if err != nil {
		t.Fatal(err)
	}
	code, out, _ := run(map[string]string{"OWNER_KEY_FILE": key})
	if code != 0 || out != "http://127.0.0.1:8787/#setup="+generated+"\n" {
		t.Fatal(code, out)
	}
	code, out, _ = run(map[string]string{"OWNER_KEY_FILE": key, "COMMONROOM_PUBLIC_ORIGIN": "https://rooms.example.test", "LISTEN_ADDR": "127.0.0.1:9000"})
	if code != 0 || out != "https://rooms.example.test/#setup="+generated+"\n" {
		t.Fatal("public origin ignored", out)
	}
	code, out, _ = run(map[string]string{"OWNER_KEY_FILE": filepath.Join(dir, "absent"), setupCodeEnv: suppliedCode, "LISTEN_ADDR": "127.0.0.1:9000"})
	if code != 0 || out != "http://127.0.0.1:9000/#setup="+suppliedCode+"\n" {
		t.Fatal("supplied code ignored", out)
	}
	// Spent: the owner key was claimed, so no link is offered.
	os.WriteFile(key+".public.pem", []byte("public"), 0o600)
	if code, out, diagnostic := run(map[string]string{"OWNER_KEY_FILE": key}); code != 1 || out != "" || !strings.Contains(diagnostic, "already been claimed") {
		t.Fatal(code, out, diagnostic)
	}
	os.WriteFile(key+".private.pem", []byte("pending"), 0o600)
	if code, _, _ := run(map[string]string{"OWNER_KEY_FILE": key}); code != 0 {
		t.Fatal("pending setup was refused")
	}
	if code, out, _ := run(map[string]string{"OWNER_KEY_FILE": filepath.Join(dir, "none")}); code != 1 || out != "" {
		t.Fatal("missing code produced a link", code, out)
	}
	if code, out, _ := run(map[string]string{setupCodeEnv: "bad"}); code != 1 || out != "" {
		t.Fatal("invalid supplied code produced a link")
	}
	if code, out, _ := run(nil, "extra"); code != 2 || out != "" {
		t.Fatal("arguments accepted")
	}
	var out2, diagnostic bytes.Buffer
	if handled, code := startupCommand([]string{"--setup-link", "x"}, &out2, &diagnostic); !handled || code != 2 || out2.Len() != 0 {
		t.Fatal("startup dispatch accepted arguments")
	}
}
