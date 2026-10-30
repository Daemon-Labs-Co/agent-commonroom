package main

import (
	"agentroom/internal/strictjson"
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
)

// Private deployment policy, never persisted in a room/task or exposed to models.
// Strict decoding rejects unknown keys and duplicate keys, including case
// variants, while allowing ordinary whitespace and property ordering. No profile
// is usable until the ENTIRE document and all of its
// identities validate. Missing configuration means refusal, not weaker execution.
type objectReadWorkerProfileFile struct {
	Version  int                           `json:"version"`
	Profiles []objectReadWorkerProfileSpec `json:"profiles"`
}
type objectReadWorkerProfileSpec struct {
	CompanyID string                  `json:"companyId"`
	Worker    string                  `json:"worker"`
	Role      string                  `json:"role"`
	Address   string                  `json:"address"`
	Identity  objectReadIdentityFiles `json:"identity"`
}
type objectReadWorkerProfile struct {
	address  string
	identity objectReadPeerIdentity
}
type objectReadWorkerProfiles struct {
	entries map[string]objectReadWorkerProfile
}

func objectReadProfileKey(company, worker, role string) string {
	return company + "/" + worker + "/" + role
}

// Parent directories must remain operator-controlled; the private-file inode
// safety contract applies to policy as well as keys. No discovery, writes, model
// selected address, host/SSH fallback or optional authentication flag exists.
func loadObjectReadWorkerProfiles(path string) (*objectReadWorkerProfiles, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errObjectReadPeer
	}
	raw, err := readPrivateBounded(path, 64*1024)
	if err != nil || !strictjson.ValidEncoding(raw) {
		return nil, errObjectReadPeer
	}
	var file objectReadWorkerProfileFile
	if json.Unmarshal(raw, &file) != nil || file.Version != 1 || len(file.Profiles) == 0 || len(file.Profiles) > 32 {
		return nil, errObjectReadPeer
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&file) != nil || !uniqueObjectReadProfileJSON(raw) {
		return nil, errObjectReadPeer
	}
	result := &objectReadWorkerProfiles{entries: make(map[string]objectReadWorkerProfile)}
	for _, p := range file.Profiles {
		if !membershipPartPattern.MatchString(p.CompanyID) || !workerAliasPattern.MatchString(p.Worker) || (p.Role != "worker" && p.Role != "controller") || !validObjectReadControlAddress(p.Address) {
			return nil, errObjectReadPeer
		}
		key := objectReadProfileKey(p.CompanyID, p.Worker, p.Role)
		if _, exists := result.entries[key]; exists {
			return nil, errObjectReadPeer
		}
		identity, err := loadObjectReadPeerIdentity(p.Identity, p.Role == "controller")
		if err != nil {
			return nil, err
		}
		result.entries[key] = objectReadWorkerProfile{p.Address, identity}
	}
	return result, nil
}

// Resolve only a locally approved company/worker/role. The exact retained record
// and intent are pinned fresh for this operation; decoded peer frames cannot add
// or update a profile. Controller account/evidence checks remain authoritative.
func (p *objectReadWorkerProfiles) endpoint(worker, role string, r workerJobRecord, i workerObjectReadIntent) (objectReadControlEndpoint, error) {
	if p == nil || !workerAliasPattern.MatchString(worker) || (role != "worker" && role != "controller") || !objectReadAuthorityBinding(r, i) {
		return objectReadControlEndpoint{}, errObjectReadPeer
	}
	company := i.Request.Receipt.Intent.Scope.CompanyID
	profile, ok := p.entries[objectReadProfileKey(company, worker, role)]
	if !ok {
		return objectReadControlEndpoint{}, errObjectReadPeer
	}
	binding := objectReadPeerBinding{Identity: profile.identity, Worker: worker, RecordDigest: probeWorkerRecordDigest(r), Intent: i}
	endpoint := objectReadControlEndpoint{Address: profile.address, Binding: binding}
	if !endpoint.valid(r, i) {
		return objectReadControlEndpoint{}, errObjectReadPeer
	}
	return endpoint, nil
}

// Limit nesting and reject duplicate property names before a policy is usable.
// Fold key case because encoding/json matches struct fields case-insensitively.
func uniqueObjectReadProfileJSON(raw []byte) bool {
	d := json.NewDecoder(bytes.NewReader(raw))
	var value func(int) bool
	value = func(depth int) bool {
		if depth > 8 {
			return false
		}
		token, err := d.Token()
		if err != nil {
			return false
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return true
		}
		switch delimiter {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				token, err := d.Token()
				key, ok := token.(string)
				if err != nil || !ok {
					return false
				}
				key = strings.ToLower(key)
				if seen[key] {
					return false
				}
				seen[key] = true
				if !value(depth + 1) {
					return false
				}
			}
			token, err = d.Token()
			return err == nil && token == json.Delim('}')
		case '[':
			for d.More() {
				if !value(depth + 1) {
					return false
				}
			}
			token, err = d.Token()
			return err == nil && token == json.Delim(']')
		}
		return false
	}
	if !value(0) {
		return false
	}
	_, err := d.Token()
	return err == io.EOF
}
