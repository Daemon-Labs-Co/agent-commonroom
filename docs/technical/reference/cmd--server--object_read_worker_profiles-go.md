# cmd/server/object_read_worker_profiles.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_worker_profiles.go)

Go implementation for object read worker profiles; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadWorkerProfileFile`**: Private deployment policy, never persisted in a room/task or exposed to models. Strict decoding rejects unknown keys and duplicate keys, including case variants, while allowing ordinary whitespace and property ordering. No profile is usable until the ENTIRE document and all of its identities validate. Missing configuration means refusal, not weaker execution.
  Declaration: `struct { Version int 'json:"version"' Profiles []objectReadWorkerProfileSpec 'json:"profiles"' }`.
- **Type `objectReadWorkerProfileSpec`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { CompanyID string 'json:"companyId"' Worker string 'json:"worker"' Role string 'json:"role"' Address string 'json:"address"' Identity objectReadIdentityFiles 'json:"identity"' }`.
- **Type `objectReadWorkerProfile`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { address string identity objectReadPeerIdentity }`.
- **Type `objectReadWorkerProfiles`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { entries map[string]objectReadWorkerProfile }`.
- **`objectReadProfileKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(company, worker, role string) string`.
- **`loadObjectReadWorkerProfiles`** — Parent directories must remain operator-controlled; the private-file inode safety contract applies to policy as well as keys. No discovery, writes, model selected address, host/SSH fallback or optional authentication flag exists.
  Signature: `func(path string) (*objectReadWorkerProfiles, error)`.
  Calls: `bytes.NewReader`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `filepath.Clean`, `filepath.IsAbs`, `json.NewDecoder`, `json.Unmarshal`, `len`, `loadObjectReadPeerIdentity`, `make`, `membershipPartPattern.MatchString`, `objectReadProfileKey`, `readPrivateBounded`, `strictjson.ValidEncoding`, `uniqueObjectReadProfileJSON`, `validObjectReadControlAddress`, `workerAliasPattern.MatchString`.
- **`(*objectReadWorkerProfiles).endpoint`** — Resolve only a locally approved company/worker/role. The exact retained record and intent are pinned fresh for this operation; decoded peer frames cannot add or update a profile. Controller account/evidence checks remain authoritative.
  Signature: `func(worker, role string, r workerJobRecord, i workerObjectReadIntent) (objectReadControlEndpoint, error)`.
  Calls: `endpoint.valid`, `objectReadAuthorityBinding`, `objectReadProfileKey`, `probeWorkerRecordDigest`, `workerAliasPattern.MatchString`.
- **`uniqueObjectReadProfileJSON`** — Limit nesting and reject duplicate property names before a policy is usable. Fold key case because encoding/json matches struct fields case-insensitively.
  Signature: `func(raw []byte) bool`.
  Calls: `bytes.NewReader`, `d.More`, `d.Token`, `json.Delim`, `json.NewDecoder`, `strings.ToLower`, `value`.
