# Internal finite artifact manifests

`internal/artifactmanifest` resolves a file or bundle output contract against
actual sealed bytes in `internal/objectcustody`. It has no product API, UI,
managed dispatch hook or production adapter. File/bundle task dispatch still
refuses. No feature flag can bypass custody, audit, permissions or quota.
This internal foundation does not establish a production release gate.

## Finite contract

`Resolve` requires a valid complete task definition, its exact digest and the
company, room, task, execution, job, repository and revision identity. Each claim
must carry a sealed custody receipt for that exact execution, output name and
logical member name. A logical filename is metadata, never a host path. Custody
uses its independently generated object ID under its pinned private store.

A file has exactly its declared filename. A bundle has every declared member
exactly once; there are no optional or undeclared members, reused objects or
implicit directory/archive expansion. Preflight checks the entire claim set,
registered formats, exact reservation caps, individual and aggregate byte limits
before attempting any object read. Bundles support 1–32 members; task contracts
and custody impose a maximum of 16 MiB. The reservation is for the declared
maximum, not merely the actual smaller content.

Every read passes through the real custody store's sealed receipt, current
journal authorization, inode, hash, size and durable release acknowledgment
checks. A failed member returns no partial manifest and never deletes bytes or
refunds capacity. Earlier read acknowledgments record authorization to release
bytes within this process; they do not imply a complete bundle or remote receipt.

The result contains sorted members, actual sealed sizes and hashes, canonical
media types, structural-check versions, definition/execution identities and a
contract digest. Its own SHA-256 digest covers the complete manifest. `Revalidate`
requires exact canonical metadata and repeats current-authority and actual-byte
verification; recomputing a digest cannot authorize forged or stale metadata.
A trusted production journal must recheck actual source revision and authority;
matching caller-supplied strings alone is never proof of origin.

## Registered structural formats

| Media type | Structural check |
| --- | --- |
| `application/octet-stream` | Bounded opaque bytes, including empty content |
| `text/plain`, `text/markdown` | Valid UTF-8; control characters refused except tab, CR and LF |
| `application/json` | Strict bounded JSON, including duplicate-key, Unicode and nesting validation |

Text and JSON optionally accept only `charset=utf-8`, case-normalized. Opaque
bytes accept no parameters. Unsupported formats or parameters refuse before
reads. Empty text/opaque data is valid; empty JSON is invalid. There are no
arbitrary plugins, shell commands, archive extraction or image/document decoders.
Opaque acceptance is not a secret scan, malware scan or validation of semantics.
JSON structure is not proof that a human criterion or independent check passed.

## Required integration and verification

Production still requires a trusted isolated worker export transport, real
source/volume/revision identity, production accounting/custody guard assembly,
complete retained-data inventory, enforced physical quota, retention/disposal workflow,
and authenticated audited public delivery. The [controller journal](artifact-custody-journal.md)
now records durable custody metadata but does not supply those production guards.
No synthetic adapter may activate
managed output delivery. Export cancellation must be transport-aware. Independent
observed checks and owner acceptance remain separate from format validation.

Tests use private temporary files and a mocked in-memory journal, with the real
custody store. They cover canonical ordering, exact members and scopes, full-set
preflight refusal, strict actual content, tampering, revoked authority, missing
seals and release-audit failure. They do not exercise live workers, provider
accounts, a database, kernel quota, crash recovery or public download behavior.
See [custody](object-custody.md), [task definitions](task-definitions.md),
[architecture](architecture.md) and [release readiness](release-readiness.md).

The internal [workspace capture foundation](workspace-source-capture.md) now computes actual bounded tree identity and verifies selected bytes against receipts. Its trusted production frozen-volume lease and transport are still missing; it does not enable public or managed file delivery.

## Controller-selected sealed manifest assembly

The internal server method `resolveStoredArtifactManifest(ctx, scope, outputName,
store)` now joins durable controller custody with the actual-byte resolver. Trusted
assembly supplies the already authenticated human context, exact historical
execution scope, declared output name and approved isolated custody store. No
HTTP route, model tool, host-path selector or permissive store/journal is installed.

Before any read, the controller checks current human room access, account identity,
incident/stop/integrity state, immutable definition digest and every matching
custody record's exact sealed receipt, quota readiness and unreleased reservation.
Claims come from replayed controller records, never caller/provider receipt lists.
It selects the COMPLETE output member set for the exact execution/revision and
refuses unsealed or malformed matching records. The existing finite resolver then
rejects missing/extra/duplicate members and bad formats/caps before object reads.

The server lock is released before reading: each store read's real journal must
reacquire that lock to authorize/acknowledge bytes and verify its independent
physical release guard. Actual sealed content, size/hash and structural format
are checked. No server lock is held across storage callbacks. After ALL members
resolve, the controller reloads current permissions/integrity and compares the
complete canonical receipt set and definition again. Revocation, incident,
changed custody or uncertain state withholds the entire manifest; no partial
manifest is returned. The operation uses the existing bounded storage context.

Individual durable read acknowledgments remain authorization history if a later
member/format/final-policy check fails. They do not certify a complete bundle,
delivery, acceptance, refund or deletion. A manifest contains metadata only: it
neither accepts/completes a task nor provides a byte-download capability. Turning
`taskDefinitions` editing off preserves authorized historical verification; it
never bypasses the mandatory custody/account/export checks. This internal security
verification has no separate optional bypass flag.

This is controller assembly, not production delivery. The passed store MUST be
assembled with the actual controller journal and independently authenticated,
quota-backed isolated custody; an arbitrary host store or dishonest source fork
voids the guarantee. Current source/revision/retention/export policy and actual
writer fences still belong to the mandatory physical guard. Snapshot comparison
is not proof of atomic filesystem custody. Durable internal registry recording is described below. Attaching a manifest to
task outputs, current-byte revalidation before export, public downloads, observed
execution checks and human acceptance remain separate production integration tasks.

Tests use real controller projection/journal plus private synthetic temporary
custody files and mocked DB/physical guards. They cover actual file and complete
bundle assembly, missing/unsealed members, actual-byte tampering and JSON format
failure, foreign scope, human/agent authorization, editing-flag states, cancellation,
incident/integrity refusal and access revocation after durable read acknowledgment.
No database, provider, Docker, network or native quota service is used. Synthetic
manifest success does not close a live release gate or enable file-task dispatch.


## Immutable controller registry

`recordStoredArtifactManifest` now performs trusted actual-byte assembly and
records one `artifact.manifest.recorded` audit event for each execution/output.
It rechecks current human room access, incident/stop/integrity state, canonical
metadata and the complete exact sealed custody/reservation set immediately before
the append. Every member must already have a durable byte-read acknowledgment.
The append uses the observed event count under the transactional audit-chain lock:
a concurrent revision is a conflict, never permission to write stale metadata.

The event contains version1, actor ID, performedBy and the canonical manifest;
the normal audit layer adds protected record metadata. Projection validates the
actor stamp, canonical definition/output/receipt shape, exact complete membership,
sealed/quota-backed/unreleased custody and prior read acknowledgments without
filesystem I/O. Each execution/output registry key is immutable: duplicate events,
replacement, deletion and automatic adoption are refused. New task executions
receive their own keys. Returned member slices do not mutate replayed history.
Registry data is internal (`json:"-"`); generic state responses do not expose it.
Client `/api/events` requests cannot submit this trusted event, even as owner.

Any append failure returns no success manifest. Uncertain persistence trips the
existing integrity latch, retains sealed objects, full reservations and prior
read acknowledgments, and must be reconciled from the audit. A revision conflict
also returns no success. No automatic retry, object deletion or storage credit is
inferred. Recording does not complete/accept a task, attach a task output, prove
client delivery or authorize a later read. Fresh actual-byte/export checks remain
mandatory. Turning task-definition editing off preserves authorized historical
verification/recording; no flag can weaken the security checks.

Portable `artifactmanifest.ValidateMetadata` shares the canonical finite contract
preflight with actual-byte `Resolve`. It validates shape/digests only and performs
NO read, authorization or structural content scan. It can accept historical
metadata while access is revoked or the backing store is absent; that is replay,
not permission. `Revalidate` still repeats actual custody reads. Malformed receipt
identities are now rejected for the whole set before any member read.

Tests cover durable recording/restart, immutable keys/caller copies, generic-state
privacy, public-event injection, malformed/forged replay, missing read evidence,
editing-flag behavior and failed commit with retained custody/history. Native
source/quota/export guards, live storage acceptance, public review/download flows
and task-output attachment remain missing. No synthetic registry record closes
those gates or certifies present physical bytes after source disposal.

## Private recorded-member revalidation

`readRecordedArtifactMember(ctx, executionId, outputName, manifestDigest,
memberName, store)` now provides an internal bridge from the immutable registry
to actual verified member bytes. Trusted authenticated setup supplies the approved
isolated store and retains independent physical custody/writer/export fences for
the WHOLE operation. There is no default store, HTTP route, model tool, arbitrary
host path or execution fallback. This is not yet a public download feature.

The controller first loads the exact recorded execution/output and caller-pinned
manifest digest under current enabled human room/account permissions, incident/
stop/integrity checks and canonical complete custody linkage. Unknown member
names, stale digests and foreign executions refuse BEFORE any custody read.
The logical member name must match a declared recorded file exactly; it never
becomes a filesystem path. Object IDs/receipts come only from the trusted record.

The operation revalidates ALL finite members through actual custody bytes and
current journal/physical export checks, then reads the selected member again
through its exact sealed receipt. A valid member cannot be released from a bundle
whose sibling is now corrupt or unauthorized. The selected read checks authority
again rather than reusing its earlier verification. After that read, the controller
reloads current permissions, immutable record and complete custody linkage before
returning bytes. Revocation, cancellation or change withholds/clears the result.
Server locks are never held across store reads; the whole call uses the bounded
storage context. Matching metadata/repeated reads do not replace real writer
exclusion or independently prove current physical volume ownership.

Durable acknowledgments from verification and the selected read remain history
when a later step fails. They count internal authorizations, NOT user downloads
or delivery confirmations. No failed read refunds capacity, deletes data,
replaces the manifest or accepts/completes a task. Turning task-definition editing
off preserves authorized historical reading; no optional flag can bypass current
custody/export checks. The registry remains private in generic state responses.

Tests cover actual file/bundle member reads, stale digest/unknown member/foreign
execution and agent refusal without reads, byte tampering, corrupt sibling,
physical release revocation between verification and selected read, permission
revocation AFTER the selected acknowledgment, cancellation before/during reads,
editing-off history and retained charge/audit evidence. All stores use synthetic
private temporary files with mocked DB and physical guards. Production independent
custody, complete inventory, public transport/export and human review/acceptance
assembly and live provider/runtime acceptance remain required.

## Recovery after controller restart

Restart does not turn a recorded digest into a download capability. Trusted
assembly must reopen the independently provisioned private custody root and
construct a NEW controller journal with current physical/export guards. Never
carry an old controller's authorization decision into the restarted reader.
The existing recorded execution/output/digest selects history; every actual
read still checks current permissions, stops, bytes and durable acknowledgments.

Missing or changed blobs, unsafe root permissions, a denied physical guard,
an incident or revoked room membership withhold the result. Do not recreate
missing bytes from provider claims, replace an immutable manifest, adopt an
unsealed object or refund its reservation to make recovery appear successful.
An uncertain read-acknowledgment write also withholds bytes and latches the
controller's integrity stop. Follow the existing integrity recovery procedure;
restart alone is not permission to clear the stop or retry delivery.

Synthetic restart tests reopen the root with a fresh controller journal/guard
without reusing the original authorization objects, then check healthy and editing-disabled historical
reads plus each refusal above. A subsequent replay must retain the same manifest,
full capacity charge and only acknowledged read history. These tests mock physical
custody and persistence; they do not prove independently retained storage survives
source-container disposal, deploy a recovery service or close a live release gate.
