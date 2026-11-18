package main

import (
	"strings"
	"testing"
	"time"
)

func TestTenancySettings(t *testing.T) {
	t.Parallel()
	env := func(kv ...string) func(string) string {
		m := map[string]string{}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return func(k string) string { return m[k] }
	}
	if got, err := loadTenancySettings(env()); err != nil || got.Mode != tenancySingle {
		t.Fatal("single must be the default", got, err)
	}
	if got, err := loadTenancySettings(env("COMMONROOM_TENANCY", "single")); err != nil || got.Mode != tenancySingle {
		t.Fatal(got, err)
	}
	if got, err := loadTenancySettings(env("COMMONROOM_TENANCY", "multi", "COMMONROOM_WORKSPACE_DOMAIN", "rooms.example.com")); err != nil || got.Domain != "rooms.example.com" {
		t.Fatal(got, err)
	}
	for name, e := range map[string]func(string) string{
		"unknown mode":            env("COMMONROOM_TENANCY", "shared"),
		"domain without multi":    env("COMMONROOM_WORKSPACE_DOMAIN", "rooms.example.com"),
		"multi without domain":    env("COMMONROOM_TENANCY", "multi"),
		"domain with scheme":      env("COMMONROOM_TENANCY", "multi", "COMMONROOM_WORKSPACE_DOMAIN", "https://rooms.example.com"),
		"domain with port":        env("COMMONROOM_TENANCY", "multi", "COMMONROOM_WORKSPACE_DOMAIN", "rooms.example.com:443"),
		"uppercase domain":        env("COMMONROOM_TENANCY", "multi", "COMMONROOM_WORKSPACE_DOMAIN", "Rooms.Example.com"),
		"single label domain":     env("COMMONROOM_TENANCY", "multi", "COMMONROOM_WORKSPACE_DOMAIN", "localhost"),
		"wildcard domain":         env("COMMONROOM_TENANCY", "multi", "COMMONROOM_WORKSPACE_DOMAIN", "*.example.com"),
		"public origin conflicts": env("COMMONROOM_TENANCY", "multi", "COMMONROOM_WORKSPACE_DOMAIN", "rooms.example.com", "COMMONROOM_PUBLIC_ORIGIN", "https://x.example.com"),
	} {
		if _, err := loadTenancySettings(e); err == nil {
			t.Fatal("accepted:", name)
		}
	}
}

func TestWorkspaceFromHost(t *testing.T) {
	t.Parallel()
	single := tenancySettings{Mode: tenancySingle}
	for _, host := range []string{"127.0.0.1:8787", "anything.example.com", "", "agentcommonroom.home.arpa"} {
		if id, ok := single.workspaceFromHost(host); !ok || id != defaultWorkspaceID {
			t.Fatal("single mode must always serve the default workspace", host)
		}
	}
	multi := tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}
	for host, want := range map[string]string{
		"acme.rooms.example.com":      "acme",
		"ACME.Rooms.Example.COM":      "acme",
		"acme.rooms.example.com:8443": "acme",
		"acme.rooms.example.com.":     "acme",
		"a1-b2.rooms.example.com":     "a1-b2",
	} {
		if id, ok := multi.workspaceFromHost(host); !ok || id != want {
			t.Fatal(host, id, ok)
		}
	}
	for _, host := range []string{
		"", "rooms.example.com", "example.com", "acme.example.com", "acme.rooms.example.org",
		"a.b.rooms.example.com", ".rooms.example.com", "-a.rooms.example.com", "a_b.rooms.example.com",
		"www.rooms.example.com", "api.rooms.example.com", "admin.rooms.example.com", "platform.rooms.example.com",
		"default.rooms.example.com", "acme.rooms.example.com.evil.test", "xacme.rooms.example.com.attacker.com",
		"acme.rooms.example.com@evil.test", "acme rooms.example.com", strings.Repeat("a", 64) + ".rooms.example.com",
	} {
		if id, ok := multi.workspaceFromHost(host); ok {
			t.Fatal("served a host that is not exactly one workspace label:", host, id)
		}
	}
	if multi.origin("acme") != "https://acme.rooms.example.com" {
		t.Fatal(multi.origin("acme"))
	}
}

func TestDatabaseWorkspaceRegistry(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	created := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	db.workspaces = map[string][]any{"acme": {"acme", "Acme Ltd", "active", `{"rooms":7}`, created}, "old": {"old", "Old", "deleted", `{}`, created}}
	r := databaseWorkspaceRegistry{db: db}
	rec, found, err := r.Lookup(t.Context(), "acme")
	if err != nil || !found || rec.Name != "Acme Ltd" || rec.Status != workspaceActive || rec.Limits.Rooms != 7 || rec.Limits.Members != defaultWorkspaceLimits.Members {
		t.Fatal(rec, found, err)
	}
	if _, found, err := r.Lookup(t.Context(), "nobody"); found || err != nil {
		t.Fatal("unknown workspace found", found, err)
	}
	for _, id := range []string{"www", "default", "A", "a.b", "", "x'; DROP TABLE workspaces;--"} {
		if _, found, err := r.Lookup(t.Context(), id); found || err != nil {
			t.Fatal("invalid or reserved id reached the registry", id)
		}
	}
	db.failAt = "query"
	if _, _, err := r.Lookup(t.Context(), "acme"); err == nil || strings.Contains(err.Error(), "mock") {
		t.Fatal("registry failure must be generic", err)
	}
}
