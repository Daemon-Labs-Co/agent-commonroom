package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Operator tools for the platform API (ADR 0007): create a signing key, and send one
// signed request. The private key never leaves its file and is never printed.

// platformKeygenMain creates KEYFILE (0600, refusing to overwrite) and prints the registry
// entry for COMMONROOM_PLATFORM_KEYS_FILE.
func platformKeygenMain(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || !extensionIDPattern.MatchString(args[1]) {
		fmt.Fprintln(stderr, "Usage: commonroom --platform-keygen KEYFILE ID   (ID: lowercase letters, digits and dashes, starting with a letter)")
		return 2
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintln(stderr, "Key generation failed")
		return 1
	}
	if err := writePrivateFile(args[0], []byte(base64.StdEncoding.EncodeToString(priv)+"\n")); err != nil {
		fmt.Fprintln(stderr, "The key file could not be created (it must not already exist)")
		return 1
	}
	entry, _ := json.MarshalIndent([]map[string]any{{"id": args[1], "name": "Platform control plane", "publicKey": base64.StdEncoding.EncodeToString(pub), "rooms": []string{}}}, "", "  ")
	fmt.Fprintf(stdout, "%s\n", entry)
	fmt.Fprintln(stderr, "Save the JSON above as COMMONROOM_PLATFORM_KEYS_FILE (chmod 600). Keep KEYFILE private: it can create and delete workspaces.")
	return 0
}

// platformRequestMain sends one signed request: KEYFILE ID METHOD URL [JSON-BODY].
func platformRequestMain(args []string, stdout, stderr io.Writer, now func() time.Time) int {
	if len(args) < 4 || len(args) > 5 {
		fmt.Fprintln(stderr, "Usage: commonroom --platform-request KEYFILE ID METHOD URL [JSON-BODY]")
		return 2
	}
	keyFile, id, method, target := args[0], args[1], strings.ToUpper(args[2]), args[3]
	body := ""
	if len(args) == 5 {
		body = args[4]
	}
	u, err := url.Parse(target)
	loopback := err == nil && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost")
	if err != nil || u.Host == "" || !(u.Scheme == "https" || (u.Scheme == "http" && loopback)) || !strings.HasPrefix(u.Path, "/platform/v1/") {
		fmt.Fprintln(stderr, "URL must be https://platform.<your domain>/platform/v1/... (http is allowed only for localhost)")
		return 2
	}
	switch method {
	case "GET", "POST", "PUT":
	default:
		fmt.Fprintln(stderr, "METHOD must be GET, POST or PUT")
		return 2
	}
	raw, err := readPrivateBounded(keyFile, 1024)
	if err != nil {
		fmt.Fprintln(stderr, "The key file could not be read (it must be private to this account)")
		return 1
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil || len(decoded) != ed25519.PrivateKeySize {
		fmt.Fprintln(stderr, "The key file is not a platform key")
		return 1
	}
	stamp := strconv.FormatInt(now().Unix(), 10)
	signature := ed25519.Sign(ed25519.PrivateKey(decoded), platformSigningInput(id, stamp, method, u.RequestURI(), []byte(body)))
	req, err := http.NewRequest(method, target, strings.NewReader(body))
	if err != nil {
		fmt.Fprintln(stderr, "Invalid request")
		return 2
	}
	req.Header.Set("X-Platform-Id", id)
	req.Header.Set("X-Platform-Timestamp", stamp)
	req.Header.Set("X-Platform-Signature", base64.StdEncoding.EncodeToString(signature))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(stderr, "Request failed: the platform host could not be reached")
		return 1
	}
	defer resp.Body.Close()
	reply, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	fmt.Fprintf(stderr, "HTTP %d\n", resp.StatusCode)
	stdout.Write(reply)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 1
	}
	return 0
}
