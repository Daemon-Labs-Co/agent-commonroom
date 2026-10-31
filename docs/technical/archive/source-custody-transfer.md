# Captured source transfer into object custody

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


`workspacecapture.CaptureToCustody` now connects the fixed isolated source lease
to the private immutable object store. It returns the actual captured snapshot and
one sealed receipt per selected declared member only after the complete transfer
and final source checks succeed. This is internal infrastructure, not an enabled
file task, public download, format validator or human acceptance.

## Trusted configuration and usage

The reviewed capture helper supplies a supervisor-owned finite `Request` containing
execution/import/volume binding, declared members and fresh challenge. It also supplies
mandatory live capture authority, a separate private quota-backed object root and
the actual durable custody journal. The source is always the fixed `/workspace`
view; there is no source path argument or uncontained fallback.

For example, a declared bounded `report.txt` member can be transferred after the
supervisor has authenticated its task and isolated source and configured physical
storage enforcement. The source scanner computes its revision. The object store
generates a fresh random object identity; the output label never selects a storage
path. Callers own the object-root descriptor until transfer completes.

The object journal must be the real controller/worker admission path. Its `Reserve`
must durably charge maximum capacity and establish the exact physical quota on the
actual private object volume before writes. Directory mode and byte counters do not
supply physical quota. Its source/seal guards must independently validate policy,
content/format requirements and actual source/root identity. No default, no-op or
synthetic journal can admit production work.

## Ordering and failure behavior

1. Capture the finite actual source under the live frozen lease. This separately
   authorized bounded observation buffers selected bytes; it does not create a
   custody object or replace its later quota reservation.
2. Build a private write-only custody adapter bound to that exact computed source
   revision, output/member names and maximum sizes. Revalidate the complete source.
3. For each selected member, prepare a lazy reader. The object store durably reserves
   and establishes physical quota before creating an object or consuming that reader.
   The adapter rechecks source authority after quota setup, refusing changes before
   object creation/transfer while retaining any durable reservation.
4. Stream bounded actual bytes to the immutable store. Before the external journal
   can seal, verify the actual receipt against the captured source and current
   revision. The real journal still enforces source, quota/root, policy and audit.
5. Revalidate each returned receipt and the complete source before returning the
   full result. The outer capture lease also requires a final current-source check.

Transfer authority and journal contexts have the same thirty-second ceiling,
shortened by caller deadlines/cancellation. Callbacks must cooperate; supervised
runtime process deadlines remain required for blocked code.

A failure returns no partial result or receipt prefix. It does **not** roll back
storage: reserved capacity, unsealed files and individually sealed objects remain
recorded. Once a store attempt begins, failure is reported as unconfirmed so an
operator does not mistake it for proof that nothing persisted. A later member's
failure never refunds an earlier member or its own uncertain reservation.
There is no automatic retry, adoption, replacement, deletion or refund API.
Recovery must inspect the durable controller journal rather than rerun blindly.

The adapter refuses both release methods. It never calls public/read authorization
or returns object bytes to a browser/model. Downloads need separate current
human/room/export policy and durable release acknowledgment through the normal
custody read path. A receipt proves stored byte identity, not task acceptance or
an independently observed successful build/test.

## Flags, permissions and verification

This mandatory internal adapter adds no UI, HTTP route or optional security flag.
Current task feature flags and human permissions remain the supervisor/journal's
responsibility and cannot bypass isolation, quota, auditing or retention. Required
file/bundle task dispatch remains blocked. Disabling a product capability does not
delete transferred data or refund retained charges.

Tests use synthetic temporary source/object roots, explicit mocked kernel and
supervisor evidence, and mocked quota/journal methods. They exercise actual persisted
bytes/digests/scopes, quota-before-create ordering, missing/typed-nil journal refusal,
revocation/mutation during quota setup, seal failures, revocation after seal, later
member failure, retained charges/objects, fabricated receipts and denied release.
These tests neither enforce real kernel quotas nor certify live runtime acceptance.

Remaining integration includes the reviewed capture executable/image and supervised
stream, running-view physical validation, deployed authenticated authority/evidence,
real physical quotas, controller custody journal composition, output format validation,
recovery and authenticated file delivery/review. Source forks substituting dishonest
journals, roots or authority invalidate the supported boundary. No live P0 gate is
certified by this adapter.

The [capture helper/private channel](capture-helper-channel.md) now supplies a
fixed-root executable and bounded supervisor RPC for observation/reserve/seal/finish.
It carries metadata only; parent runtime admission, real quotas and public release
remain mandatory independent integrations.
