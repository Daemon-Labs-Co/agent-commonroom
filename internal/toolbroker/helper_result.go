package toolbroker

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"

	"agentroom/internal/strictjson"
	"agentroom/internal/toolworkspace"
)

type helperResult struct {
	Operation string `json:"operation"`
	Bytes     int    `json:"bytes"`
	SHA256    string `json:"sha256"`
	Encoding  string `json:"encoding,omitempty"`
	Data      string `json:"data,omitempty"`
}

// validHelperResult validates bounded evidence before an audited success, not
// filesystem attestation. A compromised image could forge matching metadata;
// image provenance, verified mounts and the trusted helper are still required.
func validHelperResult(raw []byte, action Action) bool {
	if len(raw) > MaxOutputBytes || !strictjson.ValidEncoding(raw) {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return false
	}
	seen := make(map[string]bool, 5)
	var result helperResult
	for decoder.More() {
		key, err := decoder.Token()
		name, ok := key.(string)
		if err != nil || !ok || seen[name] {
			return false
		}
		seen[name] = true
		var value json.RawMessage
		if decoder.Decode(&value) != nil || bytes.Equal(value, []byte("null")) {
			return false
		}
		switch name {
		case "operation":
			if json.Unmarshal(value, &result.Operation) != nil {
				return false
			}
		case "bytes":
			if json.Unmarshal(value, &result.Bytes) != nil {
				return false
			}
		case "sha256":
			if json.Unmarshal(value, &result.SHA256) != nil {
				return false
			}
		case "encoding":
			if json.Unmarshal(value, &result.Encoding) != nil {
				return false
			}
		case "data":
			if json.Unmarshal(value, &result.Data) != nil {
				return false
			}
		default:
			return false
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return false
	}
	if _, err = decoder.Token(); err != io.EOF {
		return false
	}
	if !seen["operation"] || !seen["bytes"] || !seen["sha256"] || result.Operation != action.Operation || result.Bytes < 0 || !digest(result.SHA256, 32) {
		return false
	}
	var data []byte
	switch action.Operation {
	case "read":
		if !seen["encoding"] || result.Encoding != "base64" || result.Bytes > toolworkspace.MaxReadBytes || len(result.Data) > base64.StdEncoding.EncodedLen(toolworkspace.MaxReadBytes) {
			return false
		}
		data, err = base64.StdEncoding.Strict().DecodeString(result.Data)
		if err != nil || len(data) != result.Bytes || base64.StdEncoding.EncodeToString(data) != result.Data {
			return false
		}
	case "write":
		if seen["encoding"] || seen["data"] || result.Bytes != len(action.Content) {
			return false
		}
		data = []byte(action.Content)
	default:
		return false
	}
	digest := sha256.Sum256(data)
	return result.SHA256 == hex.EncodeToString(digest[:])
}
