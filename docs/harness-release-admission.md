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

## Offline review publication tool

Build or run `tools/harness-review` on the trusted review workstation. It uses the
same contract validator as controller admission; it performs no network requests,
provider calls, database access or Docker operations. This is administrator tooling,
not a room/model capability, so it has no separate product toggle. The server still
enforces `COMMONROOM_HARNESS_UPGRADE_CHECKS` as documented above.

First generate a new signing key in a private directory outside the checkout:

```sh
go run ./tools/harness-review --generate-key /absolute/private/review-signing.key
```

The tool prints only `COMMONROOM_HARNESS_REVIEW_KEY=<public key>`. Configure that
public value on the controller. The new private key file is owner-only and cannot
overwrite an existing path. Keep it offline on the review workstation; never copy
it to the controller, workers, source repository or chat. Back it up through your
organization's approved private-key process. Existing keys are also accepted as
standard base64-encoded 64-byte Ed25519 private keys, with an optional final line
ending. Their seed and public components must agree.

After reviewing the harness build and image, author the canonical payload below
as an owner-only file with current issuance/expiry, a new sequence and the exact
reviewed image digest. It must contain no final newline. Publish its signed envelope:

```sh
go run ./tools/harness-review \
  --review /absolute/private/review.json \
  --signing-key /absolute/private/review-signing.key \
  --output /absolute/private/review-0002.signed.json
```

Copy only the signed envelope to the controller's trusted private review directory,
set its owner and permissions, then atomically replace the configured review file.
Stage the exact image on the workers separately before admitting work. A signed
review is public metadata; its owner-only deployment file prevents local substitution.

Publication uses a synced temporary file and an exclusive hard link in the same
directory. Readers cannot see a partial envelope; existing files and symlinks are
never replaced. Exactly one competing publisher can claim a new destination.
Temporary files are removed on ordinary completion and errors. An interrupted
process can leave a hidden `.harness-review-*` file containing an envelope or, during
key generation, a private key: treat it as private and remove it through your normal
secure workstation process. Directory crash durability is filesystem-dependent.
The tool does not automatically rotate keys, replace live manifests, or perform
review/image compatibility analysis. Failures report a generic refusal without
printing private key contents.

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
encodings are rejected. The publication tool performs this serialization and signature step after validating
the canonical payload. Tests use disposable in-memory keys.

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
worker override is rejected here; the container runtime additionally requires the exact reviewed reference in
Docker's `RepoDigests` before creation and binds all subsequent container checks
to the inspected local image ID. A missing or mismatched repository digest blocks
creation even when the local image ID and other policy fields look valid. This
metadata check relies on the trusted daemon; it is not an image-content attestation. No network update or package installation runs in this path.

The allocation passed to private worker assembly is trusted authenticated controller
input. Selection metadata is not a portable signed authorization token. Production
controller-to-worker launch/policy assembly is still outstanding. Source forks can
bypass private helpers or dishonestly supply adapters; such changes invalidate the
security claim. Mandatory container isolation has no off switch.

Tests cover publication/signing round trips, offline key generation, exclusive
concurrent publication, unsafe key files, invalid keys, expiry and output refusal.
Admission tests use disposable signing keys, synthetic manifests, mocked audit storage and
no real database, provider, Docker daemon or image publication. Live review
publication, staged-image admission and production dispatch remain unverified.

Combined synthetic worker tests now consume a signed-review selection through
RunCodex, create the selected image, complete a callback and verify fenced role
cleanup with workspace retention. A separate acceptance driver proves staged
digest mismatch rejection before creation. These checks use fake Docker and a
synthetic engine peer, not a real model or staged registry image.
