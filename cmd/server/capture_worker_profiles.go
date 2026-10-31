package main

import (
	"agentroom/internal/strictjson"
	"bytes"
	"encoding/json"
	"path/filepath"
)

// Capture-only private deployment policy with an explicit protocol discriminator.
// Reader policies cannot silently provision capture identities. Shared strict JSON
// and PEM helpers validate encoding/material only, never operation grants.
// Private deployment policy, never persisted in a room/task or exposed to models.
// Strict decoding rejects unknown keys and duplicate keys, including case
// variants, while allowing ordinary whitespace and property ordering. No profile
// is usable until the ENTIRE document and all of its
// identities validate. Missing configuration means refusal, not weaker execution.
type captureWorkerProfileFile struct {
	Protocol string                     `json:"protocol"`
	Version  int                        `json:"version"`
	Profiles []captureWorkerProfileSpec `json:"profiles"`
}
type captureWorkerProfileSpec struct {
	CompanyID string                  `json:"companyId"`
	Worker    string                  `json:"worker"`
	Role      string                  `json:"role"`
	Address   string                  `json:"address"`
	Identity  objectReadIdentityFiles `json:"identity"`
}
type captureWorkerProfile struct {
	address  string
	identity objectReadPeerIdentity
}
type captureWorkerProfiles struct {
	entries map[string]captureWorkerProfile
}

// Parent directories must remain operator-controlled; the private-file inode
// safety contract applies to policy as well as keys. No discovery, writes, model
// selected address, host/SSH fallback or optional authentication flag exists.
func loadCaptureWorkerProfiles(path string) (*captureWorkerProfiles, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errCapturePeer
	}
	raw, err := readPrivateBounded(path, 64*1024)
	if err != nil || !strictjson.ValidEncoding(raw) {
		return nil, errCapturePeer
	}
	var file captureWorkerProfileFile
	if json.Unmarshal(raw, &file) != nil || file.Protocol != captureControlALPN || file.Version != 1 || len(file.Profiles) == 0 || len(file.Profiles) > 32 {
		return nil, errCapturePeer
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&file) != nil || !uniqueObjectReadProfileJSON(raw) {
		return nil, errCapturePeer
	}
	result := &captureWorkerProfiles{entries: make(map[string]captureWorkerProfile)}
	for _, p := range file.Profiles {
		if !membershipPartPattern.MatchString(p.CompanyID) || !workerAliasPattern.MatchString(p.Worker) || (p.Role != "worker" && p.Role != "controller") || !validObjectReadControlAddress(p.Address) {
			return nil, errCapturePeer
		}
		key := objectReadProfileKey(p.CompanyID, p.Worker, p.Role)
		if _, exists := result.entries[key]; exists {
			return nil, errCapturePeer
		}
		identity, err := loadObjectReadPeerIdentity(p.Identity, p.Role == "controller")
		if err != nil {
			if err == errObjectReadCertificatePEM {
				return nil, err
			}
			return nil, errCapturePeer
		}
		result.entries[key] = captureWorkerProfile{p.Address, identity}
	}
	return result, nil
}

// Resolve only a locally approved company/worker/role. The exact retained record
// and intent are pinned fresh for this operation; decoded peer frames cannot add
// or update a profile. Controller account/evidence checks remain authoritative.
func (p *captureWorkerProfiles) endpoint(worker, role string, r workerJobRecord, i workerCaptureIntent) (captureControlEndpoint, error) {
	if p == nil || !workerAliasPattern.MatchString(worker) || (role != "worker" && role != "controller") || !captureAuthorityBinding(r, i) {
		return captureControlEndpoint{}, errCapturePeer
	}
	company := i.Request.Binding.Execution.CompanyID
	profile, ok := p.entries[objectReadProfileKey(company, worker, role)]
	if !ok {
		return captureControlEndpoint{}, errCapturePeer
	}
	binding := capturePeerBinding{Identity: profile.identity, Worker: worker, RecordDigest: probeWorkerRecordDigest(r), Intent: copyCaptureIntent(i)}
	endpoint := captureControlEndpoint{Address: profile.address, Binding: binding}
	if !endpoint.valid(r, i) {
		return captureControlEndpoint{}, errCapturePeer
	}
	return endpoint, nil
}
