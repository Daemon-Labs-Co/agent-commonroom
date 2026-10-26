# Retained workspace observation admission

This internal worker boundary admits observation of an exact retained workspace
while holding its real creator job lock. It is one prerequisite for isolated
artifact export, not a working exporter or a source mount attestation. Users
cannot yet download managed task files through it.

## Admission and usage

Trusted worker assembly calls `withWorkerWorkspaceObservation` with the exact
reconciled worker record, approved rootless Docker runner/data root, a current
controller observation guard and an internal callback. There is no HTTP/model
entry point or caller-selected source path. Busy creator locks refuse immediately;
this path never cancels a running creator or automatically retries admission.

Only a version-4 workspace pair in `shutdown_unconfirmed` state is eligible: both
recorded role IDs must have creation and confirmed absence evidence, while the
exact created workspace remains retained. A private durable `cancel` tombstone
must exist. Shutdown remains unconfirmed because retained data is not disposed.
The creator lock is held through initial checks, callback use and final checks.
Concurrent managed cleanup/disposal cannot acquire it. The callback must finish
its work before returning; an escaped observation expires on return or panic.

`Check` re-reads the exact immutable record, verifies the cancellation tombstone,
requires current controller authority and a durable bounded-observation
acknowledgment, and checks actual container and volume inventory. No default guard
exists; nil and typed-nil dependencies refuse. Changing context cannot revive a
cancelled parent lifetime. Controller transport and durable authorization guard
assembly are still missing; tests use explicitly synthetic acknowledgments.

## Actual daemon evidence

The observation checks both fixed role names and the job-label inventory for any
remaining containers. `toolbroker.VerifyRetainedWorkspace` then observes the exact
single job workspace, full labels, creation nonce/time, local driver/scope, no
volume options and canonical daemon mountpoint. It requires no container references,
including stopped containers, and repeats the volume identity/reference pass.
Missing volumes, extra volumes, malformed or uncertain daemon replies and changes
refuse admission. No observation mounts, executes, adopts, removes or refunds data.

Every Docker call must use the same approved rootless daemon. The managed creator
lock excludes cooperating CommonRoom launch/disposal paths; daemon administrators
remain trusted and can race volume names. Docker has no immutable volume ID.
Two observations alone do not prove a permanently frozen filesystem.

## Relationship to source capture

This boundary deliberately does not implement `workspacesnapshot.FrozenLease`.
Before source capture, a production exporter must additionally establish an
attested isolated read-only mount/view and bind the actual source root to this
exact lease, while preserving writer exclusion and enforcing current policy.
No host-root shortcut or optional isolation switch is permitted. A source fork
that skips these contracts invalidates isolation and source claims.

Failures and cancellation retain workspace history, objects and charges. A
successful observation does not assert disposal, quota readiness, an artifact
receipt, useful output or byte-release authority. Existing custody journal and
source-capture requirements remain independently mandatory.

## Verification and availability

Tests seed both real role IDs in a synthetic Docker inventory, reconcile their
removal and retain the volume, then exercise actual private file locks. They cover
competing ownership, cancellation tombstones, lifecycle/record changes, names and
job-label collisions, attached/replaced/extra/missing volumes, revoked authority,
callback failures/panics and escaped observations. Docker, controller authority and
all database/provider interactions are synthetic or mocked; no live worker is used.

This mandatory internal security infrastructure adds no public capability or
product flag. It cannot enable managed file/bundle dispatch. Read-only isolated
transport/root attestation, production controller authorization, physical quota,
complete inventory/accounting integration, recovery/retention and authenticated
public delivery remain outstanding. No live P0 release gate is certified.

See [source capture](workspace-source-capture.md), [custody journal](artifact-custody-journal.md)
and [mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).

The internal [read-only workspace mount probe](readonly-workspace-mount-probe.md) verifies an isolated Linux mount and exact helper cleanup; production export integration remains outstanding.

The internal [guarded workspace probe](guarded-workspace-probe.md) now connects creator ownership, exact helper admission and durable cleanup into one operation; production controller transport and file export remain outstanding.
