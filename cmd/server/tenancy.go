package main

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Tenancy decides which workspace a request belongs to. In single mode (the
// default) every request belongs to the one workspace "default" and nothing about
// addressing changes. In multi mode the host name decides: <id>.<domain>. A
// session cookie, CSRF token or setup link is bound to its host, so none of them
// means anything on another workspace's host (ADR 0005).
const (
	tenancySingle = "single"
	tenancyMulti  = "multi"
)

type tenancySettings struct {
	Mode   string
	Domain string // multi only: lowercase base domain without a port
}

// Names that must never become workspaces: they are, or could be mistaken for,
// platform or infrastructure addresses.
var reservedWorkspaceIDs = map[string]bool{
	"www": true, "api": true, "admin": true, "platform": true, "app": true, "mail": true,
	"static": true, "assets": true, "ext": true, "help": true, "status": true, "root": true,
	"default": true, "localhost": true, "login": true, "signup": true, "auth": true,
	"cdn": true, "ns1": true, "ns2": true, "smtp": true, "ftp": true, "internal": true,
}

var domainPattern = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

func loadTenancySettings(getenv func(string) string) (tenancySettings, error) {
	mode := strings.TrimSpace(getenv("COMMONROOM_TENANCY"))
	switch mode {
	case "", tenancySingle:
		if getenv("COMMONROOM_WORKSPACE_DOMAIN") != "" {
			return tenancySettings{}, errors.New("COMMONROOM_WORKSPACE_DOMAIN requires COMMONROOM_TENANCY=multi")
		}
		return tenancySettings{Mode: tenancySingle}, nil
	case tenancyMulti:
		domain := strings.TrimSpace(getenv("COMMONROOM_WORKSPACE_DOMAIN"))
		if !domainPattern.MatchString(domain) {
			return tenancySettings{}, errors.New("COMMONROOM_WORKSPACE_DOMAIN must be a lowercase DNS name such as rooms.example.com, without a port or scheme")
		}
		if getenv("COMMONROOM_PUBLIC_ORIGIN") != "" {
			return tenancySettings{}, errors.New("COMMONROOM_PUBLIC_ORIGIN is derived per workspace in multi mode; unset it")
		}
		return tenancySettings{Mode: tenancyMulti, Domain: domain}, nil
	}
	return tenancySettings{}, errors.New("COMMONROOM_TENANCY must be single or multi")
}

// validHostedWorkspaceID is a workspace id that may be served in multi mode.
func validHostedWorkspaceID(id string) bool {
	return validWorkspaceID(id) && id != defaultWorkspaceID && !reservedWorkspaceIDs[id]
}

// workspaceFromHost maps a request Host header to a workspace id. It refuses
// anything that is not exactly one label under the configured domain.
func (t tenancySettings) workspaceFromHost(host string) (string, bool) {
	if t.Mode != tenancyMulti {
		return defaultWorkspaceID, true
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	suffix := "." + t.Domain
	if !strings.HasSuffix(host, suffix) {
		return "", false
	}
	id := strings.TrimSuffix(host, suffix)
	if strings.Contains(id, ".") || !validHostedWorkspaceID(id) {
		return "", false
	}
	return id, true
}

// origin is the exact HTTPS origin of a workspace.
func (t tenancySettings) origin(id string) string {
	return fmt.Sprintf("https://%s.%s", id, t.Domain)
}
