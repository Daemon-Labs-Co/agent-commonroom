package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// A setup code proves the person claiming a new installation was given access
// by whoever runs it. On a local install the server writes the code to a private
// file. A hosted or containerised deployment has no file the owner can read, so
// its operator may supply the code in the environment and hand the owner a setup
// link instead. Both paths are the same bootstrap secret: it is valid only while
// first-owner setup is pending, is cleared once the owner key is claimed and is
// never served over HTTP or written to the audit log.
const (
	setupCodeEnv      = "COMMONROOM_SETUP_CODE"
	setupCodeTTLEnv   = "COMMONROOM_SETUP_CODE_TTL_HOURS"
	defaultSetupHours = 24
	maximumSetupHours = 168
)

var setupCodePattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// loadSetupCode returns the setup code and, for an operator-supplied code, the
// moment it stops working. Unset configuration keeps the generated private file.
// A supplied code expires so an unclaimed link cannot stay usable indefinitely;
// restarting the service with a fresh code is how an operator reissues one.
func loadSetupCode(path string, getenv func(string) string, now time.Time) (code string, expires time.Time, err error) {
	supplied := strings.TrimSpace(getenv(setupCodeEnv))
	if supplied == "" {
		code, err = loadOwnerKey(path)
		return code, time.Time{}, err
	}
	if !setupCodePattern.MatchString(supplied) {
		return "", time.Time{}, fmt.Errorf("%s must be 64 lowercase hexadecimal characters (for example: openssl rand -hex 32)", setupCodeEnv)
	}
	hours := defaultSetupHours
	if raw := strings.TrimSpace(getenv(setupCodeTTLEnv)); raw != "" {
		hours, err = strconv.Atoi(raw)
		if err != nil || hours < 1 || hours > maximumSetupHours {
			return "", time.Time{}, fmt.Errorf("%s must be a whole number of hours from 1 to %d", setupCodeTTLEnv, maximumSetupHours)
		}
	}
	return supplied, now.Add(time.Duration(hours) * time.Hour), nil
}

// setupLinkMain prints the address an owner opens to claim a pending
// installation. The code travels after "#", which browsers never send to the
// server or to a proxy, so it stays out of access logs.
func setupLinkMain(args []string, getenv func(string) string, home string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "--setup-link takes no arguments")
		return 2
	}
	getenv, err := withServiceSettings(getenv, home)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	code := strings.TrimSpace(getenv(setupCodeEnv))
	path := getenv("OWNER_KEY_FILE")
	if path == "" {
		path = ".data/owner-key"
	}
	if code != "" {
		if !setupCodePattern.MatchString(code) {
			fmt.Fprintf(stderr, "%s is not a valid setup code\n", setupCodeEnv)
			return 1
		}
	} else {
		raw, readErr := readPrivateBounded(path, 128)
		if readErr != nil {
			fmt.Fprintln(stderr, "No setup is pending: no setup code file was found. If the owner key was already claimed, sign in with it.")
			return 1
		}
		code = strings.TrimSpace(string(raw))
	}
	// Once the private key was claimed the public key remains and the code is spent.
	if _, err := os.Stat(path + ".public.pem"); err == nil {
		if _, err := os.Stat(path + ".private.pem"); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(stderr, "No setup is pending: the owner key has already been claimed.")
			return 1
		}
	}
	origin := getenv("COMMONROOM_PUBLIC_ORIGIN")
	if origin == "" {
		addr := getenv("LISTEN_ADDR")
		if addr == "" {
			addr = "127.0.0.1:8787"
		}
		if _, _, err := net.SplitHostPort(addr); err != nil {
			fmt.Fprintln(stderr, "LISTEN_ADDR is not a valid host:port")
			return 1
		}
		origin = "http://" + addr
	}
	fmt.Fprintf(stdout, "%s/#setup=%s\n", strings.TrimRight(origin, "/"), code)
	fmt.Fprintln(stderr, "Open this link in a browser to claim the installation. Treat it like a password.")
	return 0
}

// Settings the setup link needs, read from the service's settings file when the
// command is run from an ordinary shell. Nothing else in that file is ever read
// or returned: it also holds the database login and other secrets.
var serviceSettingNames = map[string]bool{
	"OWNER_KEY_FILE":           true,
	"COMMONROOM_PUBLIC_ORIGIN": true,
	"LISTEN_ADDR":              true,
	setupCodeEnv:               true,
}

// withServiceSettings lets `commonroom --setup-link` work with no arguments and
// no exported variables, the same way on a personal machine and on a host. A
// value in the process environment wins (containers); otherwise the settings
// file the service itself starts from is used: COMMONROOM_ENV_FILE if set, else
// ~/.config/commonroom/runtime.env, the location the deployment guide installs.
// A missing file is fine; an unsafe one is refused like any private file.
func withServiceSettings(getenv func(string) string, home string) (func(string) string, error) {
	path := getenv("COMMONROOM_ENV_FILE")
	explicit := path != ""
	if !explicit {
		if home == "" {
			return getenv, nil
		}
		path = filepath.Join(home, ".config", "commonroom", "runtime.env")
	}
	raw, err := readPrivateBounded(path, 64*1024)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !explicit {
			return getenv, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("The settings file named by COMMONROOM_ENV_FILE was not found")
		}
		return nil, errors.New("The settings file must be owned by you and readable only by you (chmod 600): " + path)
	}
	file := parseServiceSettings(string(raw))
	return func(name string) string {
		if v := getenv(name); v != "" {
			return v
		}
		return file[name]
	}, nil
}

// parseServiceSettings reads KEY=VALUE lines like a systemd EnvironmentFile:
// full-line # and ; comments, optional matching quotes, no expansion.
func parseServiceSettings(text string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "export "))
		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}
		name, value, found := strings.Cut(line, "=")
		name = strings.TrimSpace(name)
		if !found || !serviceSettingNames[name] {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		out[name] = value
	}
	return out
}

// reissueSetupCode replaces a pending workspace's setup code with a fresh one that
// expires after ttl, for a platform operator resending a setup link. The previous code
// stops working immediately. It refuses once the owner key has been claimed.
func (a *ownerAuth) reissueSetupCode(ttl time.Duration, now time.Time) (string, time.Time, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.setupPending || a.setupPath == "" {
		return "", time.Time{}, errors.New("setup is not pending")
	}
	code := id() + id()
	tmp := a.setupPath + ".new"
	if err := os.WriteFile(tmp, []byte(code+"\n"), 0o600); err != nil {
		return "", time.Time{}, errors.New("setup code could not be saved")
	}
	if err := os.Rename(tmp, a.setupPath); err != nil {
		os.Remove(tmp)
		return "", time.Time{}, errors.New("setup code could not be saved")
	}
	a.key, a.keyExpires, a.attempts = code, now.Add(ttl), 0
	// The old code is revoked, so anything it signed in is revoked too. While setup
	// is pending the only sessions that can exist are owner sessions from the code.
	for token, session := range a.sessions {
		if session.ActorID == ownerID {
			delete(a.sessions, token)
		}
	}
	if err := a.persistSessionsLocked(); err != nil {
		return "", time.Time{}, errors.New("old sessions could not be revoked")
	}
	return code, a.keyExpires, nil
}

// setupSnapshot returns the pending setup code and expiry, if setup is pending.
func (a *ownerAuth) setupSnapshot() (code string, expires time.Time, pending bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.setupPending {
		return "", time.Time{}, false
	}
	return a.key, a.keyExpires, true
}
