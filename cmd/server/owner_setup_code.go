package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
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
func setupLinkMain(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "--setup-link takes no arguments")
		return 2
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
		raw, err := readPrivateBounded(path, 128)
		if err != nil {
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
