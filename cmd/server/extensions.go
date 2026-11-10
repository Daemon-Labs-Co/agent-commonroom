package main

// Extension layer, first version. An extension is a separately deployed service
// that an administrator lists in a registry file. It proves its identity by
// signing each request with its own key; the server stores only the public key.
// An extension is never a human or the owner: it may read the messages of the
// rooms it is granted and post attributed notices there. A notice may carry
// [PROPOSE TASK] lines, which become ordinary task suggestions for a human to
// capture. It cannot approve, dispatch, accept, change settings or reach any
// other endpoint. Removing the registry entry removes all access.

import (
	"agentroom/internal/room"
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	extensionContractVersion = 1
	extensionClockSkew       = 60 * time.Second
	extensionRequestsPerMin  = 30
	maxExtensionNoticeRunes  = 4000
	maxExtensions            = 16
)

var extensionIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,31}$`)

type extensionEntry struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	PublicKey string   `json:"publicKey"`
	Rooms     []string `json:"rooms"`
	key       ed25519.PublicKey
}

// extensionsEnabled defaults off; while off the registry is ignored and every
// extension request is refused.
func extensionsEnabled(st room.State) bool {
	return st.FeatureFlags["extensions"]
}

// loadExtensionRegistry reads the administrator's registry on every request, so
// removing an entry takes effect immediately. An unreadable or invalid registry
// grants nothing.
func loadExtensionRegistry(path string) (map[string]extensionEntry, error) {
	if path == "" {
		return nil, errors.New("no extension registry is configured")
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 65536 {
		return nil, errors.New("extension registry is missing, not a regular file or too large")
	}
	if info.Mode().Perm()&0o022 != 0 {
		return nil, errors.New("extension registry must not be writable by group or others")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []extensionEntry
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&entries); err != nil || len(entries) > maxExtensions {
		return nil, errors.New("extension registry is invalid")
	}
	out := map[string]extensionEntry{}
	for _, e := range entries {
		key, keyErr := base64.StdEncoding.DecodeString(e.PublicKey)
		name := strings.TrimSpace(e.Name)
		if !extensionIDPattern.MatchString(e.ID) || keyErr != nil || len(key) != ed25519.PublicKeySize || name == "" || utf8.RuneCountInString(name) > 60 || len(e.Rooms) > 64 {
			return nil, errors.New("extension registry has an invalid entry")
		}
		if _, dup := out[e.ID]; dup {
			return nil, errors.New("extension registry has a duplicate id")
		}
		e.Name, e.key = name, ed25519.PublicKey(key)
		out[e.ID] = e
	}
	return out, nil
}

// extensionSigningInput is what an extension signs: the contract version, its id,
// the timestamp, the method, the full request path with query, and the body hash.
func extensionSigningInput(id, timestamp, method, target string, body []byte) []byte {
	sum := sha256.Sum256(body)
	return []byte(strings.Join([]string{"commonroom-extension-v" + strconv.Itoa(extensionContractVersion), id, timestamp, method, target, hex.EncodeToString(sum[:])}, "\n"))
}

type extensionLimiter struct {
	mu     sync.Mutex
	window map[string][]time.Time
	seen   map[string]time.Time
}

// admit enforces the per-extension rate and refuses a replayed signature.
func (l *extensionLimiter) admit(id, signature string, now time.Time) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.window == nil {
		l.window, l.seen = map[string][]time.Time{}, map[string]time.Time{}
	}
	for sig, at := range l.seen {
		if now.Sub(at) > 2*extensionClockSkew {
			delete(l.seen, sig)
		}
	}
	if _, replay := l.seen[signature]; replay {
		return errors.New("Request was already used")
	}
	recent := l.window[id][:0]
	for _, at := range l.window[id] {
		if now.Sub(at) < time.Minute {
			recent = append(recent, at)
		}
	}
	if len(recent) >= extensionRequestsPerMin {
		l.window[id] = recent
		return errors.New("Extension request rate exceeded")
	}
	l.window[id] = append(recent, now)
	l.seen[signature] = now
	return nil
}

// authenticateExtension verifies the flag, registry entry, timestamp, signature,
// replay and rate, and returns the entry and the request body.
func (s *server) authenticateExtension(w http.ResponseWriter, r *http.Request, st room.State) (extensionEntry, []byte, bool) {
	if !extensionsEnabled(st) {
		http.Error(w, "Extensions are disabled", 403)
		return extensionEntry{}, nil, false
	}
	registry, err := loadExtensionRegistry(s.extensionRegistry)
	if err != nil {
		http.Error(w, "Extension registry unavailable", 403)
		return extensionEntry{}, nil, false
	}
	entry, ok := registry[r.Header.Get("X-Extension-Id")]
	if !ok {
		http.Error(w, "Unknown extension", 403)
		return extensionEntry{}, nil, false
	}
	stamp := r.Header.Get("X-Extension-Timestamp")
	seconds, err := strconv.ParseInt(stamp, 10, 64)
	now := time.Now()
	if err != nil || now.Sub(time.Unix(seconds, 0)) > extensionClockSkew || time.Unix(seconds, 0).Sub(now) > extensionClockSkew {
		http.Error(w, "Request timestamp missing or outside the allowed window", 403)
		return extensionEntry{}, nil, false
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 65536))
	if err != nil {
		http.Error(w, "Request body too large", 413)
		return extensionEntry{}, nil, false
	}
	signature, err := base64.StdEncoding.DecodeString(r.Header.Get("X-Extension-Signature"))
	if err != nil || !ed25519.Verify(entry.key, extensionSigningInput(entry.ID, stamp, r.Method, r.URL.RequestURI(), body), signature) {
		http.Error(w, "Invalid extension signature", 403)
		return extensionEntry{}, nil, false
	}
	if err = s.extensionLimits.admit(entry.ID, r.Header.Get("X-Extension-Signature"), now); err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return extensionEntry{}, nil, false
	}
	return entry, body, true
}

func extensionGranted(entry extensionEntry, roomID string) bool {
	for _, id := range entry.Rooms {
		if id == roomID {
			return true
		}
	}
	return false
}

// extensionAPI serves /ext/v1/rooms/{room}/messages (GET) and
// /ext/v1/rooms/{room}/notices (POST). Nothing else exists in this version.
func (s *server) extensionAPI(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/ext/v1/"), "/")
	if len(parts) != 3 || parts[0] != "rooms" || parts[1] == "" {
		http.Error(w, "Not found", 404)
		return
	}
	roomID, resource := parts[1], parts[2]
	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	entry, body, ok := s.authenticateExtension(w, r, all)
	if !ok {
		return
	}
	if !hasRoom(all, roomID) || !extensionGranted(entry, roomID) {
		http.Error(w, "Room not granted to this extension", 403)
		return
	}
	st := scoped(all, roomID)
	switch {
	case resource == "messages" && r.Method == http.MethodGet:
		if policy := roomExportPolicy(all, roomID); policy != nil && policy.ExcludeMessages {
			reply(w, map[string]any{"version": extensionContractVersion, "messages": []any{}, "excluded": true})
			return
		}
		type message struct {
			ID        string    `json:"id"`
			ActorType string    `json:"actorType"`
			Author    string    `json:"author"`
			Text      string    `json:"text"`
			Created   time.Time `json:"created"`
		}
		after, out := r.URL.Query().Get("after"), []message{}
		start := messagesAfterBoundary(st.Messages, roomID, roomContextBoundary(all, roomID))
		found := after == ""
		for _, m := range st.Messages[start:] {
			if m.RoomID != roomID {
				continue
			}
			if found {
				out = append(out, message{ID: m.ID, ActorType: m.ActorType, Author: m.Author, Text: m.Text, Created: m.Created})
			}
			if m.ID == after {
				found = true
			}
		}
		if len(out) > 100 {
			out = out[:100]
		}
		reply(w, map[string]any{"version": extensionContractVersion, "messages": out})
	case resource == "notices" && r.Method == http.MethodPost:
		if all.Incident {
			http.Error(w, "Writes are blocked during an integrity incident", 409)
			return
		}
		var req struct {
			Text string `json:"text"`
		}
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()
		text := ""
		if dec.Decode(&req) == nil {
			text = strings.TrimSpace(req.Text)
		}
		if text == "" || utf8.RuneCountInString(text) > maxExtensionNoticeRunes {
			http.Error(w, "Notice text required (4,000 characters maximum)", 400)
			return
		}
		m := room.Message{ID: id(), RoomID: roomID, ActorID: "extension:" + entry.ID, ActorType: "extension", Author: entry.Name, Text: text, Created: time.Now().UTC()}
		if err = s.appendBatch(r.Context(), []auditChange{{Kind: "message.created", Payload: m}, {Kind: "extension.notice", Payload: map[string]string{"extensionId": entry.ID, "roomId": roomID, "messageId": m.ID}}}); err != nil {
			operationError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		reply(w, map[string]any{"version": extensionContractVersion, "messageId": m.ID})
	default:
		http.Error(w, "Not found", 404)
	}
}

func roomExportPolicy(st room.State, roomID string) *room.ContextExportPolicy {
	for _, r := range st.Rooms {
		if r.ID == roomID {
			return r.ContextExport
		}
	}
	return nil
}
