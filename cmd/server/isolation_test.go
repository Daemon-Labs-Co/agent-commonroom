package main

import (
	"agentroom/internal/room"
	"bytes"
	"context"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProviderAddressDenial(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"127.0.0.1", "10.0.0.1", "192.168.1.1", "172.16.0.1", "169.254.169.254", "::1", "fc00::1", "fe80::1", "0.0.0.0", "100.64.0.1", "198.18.0.1", "0.1.2.3", "240.1.2.3", "192.88.99.1", "64:ff9b::a00:1", "64:ff9b:1::a00:1", "2002:a00:1::", "2001::1", "100::1", "100:0:0:1::1", "3ffe::1", "3fff::1", "5f00::1", "::ffff:127.0.0.1", "::ffff:169.254.169.254", "2001:2::1", "2001:20::1"} {
		if publicProviderIP(net.ParseIP(value)) {
			t.Fatal("unsafe address accepted", value)
		}
	}
	for _, value := range []string{"1.1.1.1", "::ffff:1.1.1.1", "2606:4700:4700::1111", "2001:4860:4860::8888"} {
		if !publicProviderIP(net.ParseIP(value)) {
			t.Fatal("ordinary public address rejected", value)
		}
	}
	for _, ip := range []net.IP{nil, {}, {1, 2, 3}} {
		if publicProviderIP(ip) {
			t.Fatal("invalid address accepted")
		}
	}
}
func TestGatewayRejectsBeforeNetworkAccess(t *testing.T) {
	t.Parallel()
	g := &providerGateway{ctx: context.Background()}
	for _, target := range []string{"attacker.example:443", "api.openai.com:80", "127.0.0.1:443", "api.openai.com"} {
		r := httptest.NewRequest("CONNECT", "http://example", nil)
		r.Host = target
		w := httptest.NewRecorder()
		g.connect(w, r, "codex")
		if w.Code != 403 {
			t.Fatal("unauthorized tunnel accepted", target)
		}
	}
	r := httptest.NewRequest("GET", "http://api.openai.com", nil)
	r.Host = "api.openai.com:443"
	w := httptest.NewRecorder()
	g.connect(w, r, "codex")
	if w.Code != 403 {
		t.Fatal("non-tunnel request accepted")
	}
}
func TestAgentOutputBound(t *testing.T) {
	t.Parallel()
	var b limitedAgentOutput
	if _, err := b.Write(bytes.Repeat([]byte("x"), 4*1024*1024)); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Write([]byte("x")); err == nil || b.Len() != 4*1024*1024 {
		t.Fatal("output limit failed")
	}
	if strings.Contains(b.String(), "secret") {
		t.Fatal("unexpected fixture")
	}
}
func TestLegacySessionCannotLaunch(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	p := structParticipantForIsolation()
	db.seed("participant.created", p)
	ctx, cancel := context.WithCancel(context.Background())
	s.runAgent(ctx, cancel, p, "fixture", "codex")
	st, err := s.load(context.Background())
	if err != nil || st.Participants[0].Status != "Blocked" || len(st.Messages) != 0 {
		t.Fatal("legacy host session was launched")
	}
}

func structParticipantForIsolation() room.Participant {
	return room.Participant{ID: "fixture", RoomID: "engineering", Name: "Fixture", ThreadID: "host-session"}
}

func TestClientCannotForgeContainerSessionProvenance(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "c", Name: "Fixture", Protocol: "CLI", Endpoint: "codex"})
	w := mockRequest(h, "POST", "/api/events", `{"kind":"participant.created","payload":{"name":"Fixture","roomId":"engineering","connectionId":"c","role":"Builder","rank":1,"threadId":"forged","executionMode":"container-v1"}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(context.Background())
	if err != nil || len(st.Participants) != 1 || st.Participants[0].ThreadID != "" || st.Participants[0].ExecutionMode != "" {
		t.Fatal("client forged session provenance")
	}
}
