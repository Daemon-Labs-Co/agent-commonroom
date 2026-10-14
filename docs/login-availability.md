# Login availability and session renewal

**Tier: Free authentication hardening.** Failed sign-in requests cannot use the
shared failed-attempt response budget to reject an otherwise valid owner or invited
human proof. Authentication, account enablement, room permissions, origin checks
and signature replay protection still apply.

## Behavior

The first five failed proof attempts in a controller's one-minute window return
401; additional failures return 429 with Retry-After. A valid owner bootstrap key
(during its normal setup lifetime), owner RSA signature or enabled invited-human
signature can still sign in during that window. Valid logins do not clear the
failed-attempt budget for other requests. Proof validity is checked before the
failure-response limit; this is not a pre-verification CPU throttle.

Requests are limited to 4 KiB; unknown fields and trailing JSON are rejected.
Unknown/expired challenges cannot cause a member database lookup. Member challenges
are consumed before lookup; malformed or unbounded signature encodings are rejected
before reading state. A failed attempt cannot keep replaying the same challenge to
force repeated full-state loads. RSA/public-key/account checks are unchanged.

Up to 32 active sessions are supported. A browser with an existing session can
replace it at capacity without opening a 33rd slot. A new browser at capacity must
sign out elsewhere or wait for expiry. A rejected or failed replacement does not
revoke the old browser session. The new session cookie is emitted only after
persistence succeeds; if saving fails, prior in-memory session entries are restored.
This includes both raw and persisted hashed-cookie representations.

## What to do

Use a fresh challenge and your key if an earlier attempt failed; challenges are
single-use. A correct proof can proceed despite previous failed attempts. For
503 persistence errors, keep the old browser session and repair private session
storage; do not broaden permissions or disable authentication. For capacity errors
in a new browser, sign out of an unused session or wait for its eight-hour expiry.

## Flags, operations and limitations

These authentication checks have no off switch. The existing optional
`COMMONROOM_PERSISTENT_SESSIONS` preference controls restart persistence, not proof
validation, failure handling, session capacity or origin checks. Enabled persistence
applies atomic replacement/rollback; disabled persistence still applies capacity
and renewal rules. Private-file safety checks remain mandatory.

This addresses the simple shared failed-attempt lockout and session replacement
failure paths. It does not guarantee availability under volumetric traffic,
challenge-pool exhaustion, valid-sized cryptographic garbage, slow storage or a
compromised host. Signature checking necessarily consumes work before determining
whether proof is valid. The challenge pool remains bounded at 128 and two minutes.
Public/non-loopback hosting remains unsupported; a reverse proxy cannot substitute
for an explicitly supported remote-authentication boundary. Authentication
failures must never log key, signature, challenge or bearer-cookie contents.

## Technical contract and tests

`login` separates proof validity from failed-response counting, validates strict
single-object JSON and computes replacement capacity before changing sessions.
It restores retired entries if `persistSessionsLocked` fails. `verifyMemberLogin`
requires the authentication mutex and consumes/bounds proof before database replay.
No schema migration or provider call is involved.

Parallel-safe tests use generated synthetic RSA material, in-memory DB mocks and
temporary files. Tests cover valid owner/invited proofs during failure floods,
continued rejection and replay denial, exact-capacity replacement, persistence
rollback for raw/hashed sessions, malformed/unknown challenges without DB access
and strict request decoding. Existing session restart and RSA lifecycle tests
remain in the full suite. Live production login/deployment is not claimed.
