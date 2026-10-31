# cmd/server/capture_worker_profiles.go

[Module index](cmd.md) · [Open source](../../../cmd/server/capture_worker_profiles.go)

Go implementation for capture worker profiles; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureWorkerProfileFile`**: Capture-only private deployment policy with an explicit protocol discriminator. Reader policies cannot silently provision capture identities. Shared strict JSON and PEM helpers validate encoding/material only, never operation grants. Private deployment policy, never persisted in a room/task or exposed to models. Strict decoding rejects unknown keys and duplicate keys, including case variants, while allowing ordinary whitespace and property ordering. No profile is usable until the ENTIRE document and all of its identities validate. Missing configuration means refusal, not weaker execution.
  Declaration: `struct { Protocol string 'json:"protocol"' Version int 'json:"version"' Profiles []captureWorkerProfileSpec 'json:"profiles"' }`.
- **Type `captureWorkerProfileSpec`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { CompanyID string 'json:"companyId"' Worker string 'json:"worker"' Role string 'json:"role"' Address string 'json:"address"' Identity objectReadIdentityFiles 'json:"identity"' }`.
- **Type `captureWorkerProfile`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { address string identity objectReadPeerIdentity }`.
- **Type `captureWorkerProfiles`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { entries map[string]captureWorkerProfile }`.
- **`loadCaptureWorkerProfiles`** — Parent directories must remain operator-controlled; the private-file inode safety contract applies to policy as well as keys. No discovery, writes, model selected address, host/SSH fallback or optional authentication flag exists.
  Signature: `func(path string) (*captureWorkerProfiles, error)`.
  Calls: `bytes.NewReader`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `filepath.Clean`, `filepath.IsAbs`, `json.NewDecoder`, `json.Unmarshal`, `len`, `loadObjectReadPeerIdentity`, `make`, `membershipPartPattern.MatchString`, `objectReadProfileKey`, `readPrivateBounded`, `strictjson.ValidEncoding`, `uniqueObjectReadProfileJSON`, `validObjectReadControlAddress`, `workerAliasPattern.MatchString`.
- **`(*captureWorkerProfiles).endpoint`** — Resolve only a locally approved company/worker/role. The exact retained record and intent are pinned fresh for this operation; decoded peer frames cannot add or update a profile. Controller account/evidence checks remain authoritative.
  Signature: `func(worker, role string, r workerJobRecord, i workerCaptureIntent) (captureControlEndpoint, error)`.
  Calls: `captureAuthorityBinding`, `copyCaptureIntent`, `endpoint.valid`, `objectReadProfileKey`, `probeWorkerRecordDigest`, `workerAliasPattern.MatchString`.
