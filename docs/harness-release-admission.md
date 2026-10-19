# Reviewed harness release admission

Agent CommonRoom can now check a signed operator review when registering a new
version-4 Codex development allocation. The selected version and immutable image
are recorded with its allocation audit. The private Codex worker assembly consumes
that selection and refuses a conflicting worker image or expired review before
credential acquisition. Running engines retain their allocation's original image.

This is internal admission groundwork. Public development dispatch remains blocked.
It does not discover upstream releases, build/publish images, stage them on workers,
or establish live compatibility. Other providers and older lifecycle protocols are
not supported by this review profile. The P0 upgrade gate remains open.

## Deployment configuration

Only trusted deployment administrators may configure these process environment
values. Room members, models, plugins and ordinary owner UI settings cannot change
the review key, file or flag through an application API.

| Setting | Default | Meaning |
| --- | --- | --- |
| `COMMONROOM_HARNESS_UPGRADE_CHECKS` | `false` (unset) | Exactly `true` enables review admission for newly registered paired allocations. Exactly `false` disables it; other values refuse registration. |
| `COMMONROOM_HARNESS_REVIEW_FILE` | unset | Absolute path to the signed review envelope. Required when enabled. |
| `COMMONROOM_HARNESS_REVIEW_KEY` | unset | Standard base64 encoding of the pinned 32-byte Ed25519 public review key. Required when enabled. This public key is not a provider credential. |

Restart the controller after changing its deployment environment. The file is read
afresh for each allocation, so an administrator can atomically replace it without
restarting the controller. Parent directories must be trusted. The file must be a
single-link regular file owned by the server account, with owner-only permissions
(e.g. `0600`), no symlink and at most 4096 bytes. Workers must already have the exact
image staged: the engine runtime uses `--pull=never`.

Disabled mode preserves the existing operator-specified immutable-image contract.
It does not permit mutable tags, host execution, native/SSH model launches, or
bypassing authorization, quotas, audit, retention or the public dispatch barrier.
Previously recorded selections remain enforced even after disabling the flag;
history and running jobs are preserved.

## Review format and publication

The deployment's review/signing pipeline must produce `json.Marshal` canonical
bytes in this exact field order for the signed payload:

```json
{"schema":1,"sequence":2,"provider":"codex","version":"synthetic-reviewed-2","image":"registry.example/engine@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","profile":"codex-separated-callback-v1","authProtocol":1,"auditProtocol":2,"workspaceProtocol":4,"issuedAt":"2026-10-19T12:00:00Z","expiresAt":"2026-10-20T12:00:00Z"}
```

This is a synthetic format example, not an approved release. Sign the exact payload
bytes with Ed25519. The envelope contains `review` as that JSON object followed by
`signature` as standard base64. Serialize the envelope with `json.Marshal` too;
whitespace, trailing newlines, unknown/duplicate/case-altered fields and alternate
encodings are rejected. The automated tests demonstrate signing with disposable
in-memory keys. Keep production signing private keys outside this application.

The profile pins the current separated Codex callback contract: credential framing
1, private audit 2 and workspace lifecycle 4. Versions are opaque labels (1–80 ASCII
letters, digits, dots, underscores or hyphens; first character alphanumeric).
Images must be a registry/repository reference with a lowercase SHA-256 digest.
Issuance must not be in the future and expiry must be after the admission time.
Synchronize deployment clocks. Expiry is checked again at worker startup, not used
to replace a running engine.

Use a positive, increasing sequence for each published review. Allocation history
rejects a sequence below any previously admitted review and rejects different
payloads with the same sequence. The same review may admit another fresh job.
The check and allocation audit share the existing serialized, expected-event-count
admission boundary. Failed review or audit means no active allocation. The durable
snapshot includes the sequence, version, image, payload SHA-256, check time and
expiry, never credentials or signing private keys. Lifecycle replay cannot change
this binding. Audit records are the rollback floor; no separate watermark survives
loss or unauthorized replacement of that history. Preserve audit integrity and
coordinate key rotation with the sequence history.

## Security and operational limits

A valid signature establishes the configured review authority. It does not prove
image contents match a safe harness, validate upstream provenance or measure
compatibility. Those are responsibilities of the trusted review pipeline; Docker's
existing immutable-image inspection still applies. An invalid digest format or
worker override is rejected here; actual image/digest verification remains in the
container runtime. No network update or package installation runs in this path.

The allocation passed to private worker assembly is trusted authenticated controller
input. Selection metadata is not a portable signed authorization token. Production
controller-to-worker launch/policy assembly is still outstanding. Source forks can
bypass private helpers or dishonestly supply adapters; such changes invalidate the
security claim. Mandatory container isolation has no off switch.

Tests use disposable signing keys, synthetic manifests, mocked audit storage and
no real database, provider, Docker daemon or image publication. Live review
publication, staged-image admission and production dispatch remain unverified.
