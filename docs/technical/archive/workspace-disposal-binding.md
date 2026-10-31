# Exact workspace disposal binding

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


A workspace disposal request must identify the exact volume that was reviewed,
including its creation nonce and timestamp. A yes/no discard instruction or a job
name alone is insufficient. This is mandatory retention-safety infrastructure;
there is no flag, paid override or owner-role shortcut around identity checks.

The [owner review workflow](workspace-disposal.md) now exposes this contract through
an off-by-default, human-only approval endpoint. Normal completion, cancellation
and ordinary recovery keep their non-destructive defaults. Exact bound-reservation release is integrated; production allocation and live
disposal acceptance remain P0 work.

## User and operator behavior

Keep source retained until an authorized human has reviewed an exact disposal
proposal. An existing job-recovery button stops/reconciles work; it does not grant
workspace deletion authority. Changing diagnostics flags does not alter this rule.
Do not manufacture worker cancellation payloads as a substitute for approval.

If a crash leaves a creation intent without a known creation timestamp, recovery
must first discover and journal the original volume while retaining it. Discovery
cannot authorize deletion. The resulting complete identity can then be reviewed
for a separate, exact disposal instruction. Unknown or replaced volumes remain
unresolved and are never adopted as disposable data.

Upgrading code does not purge existing workspaces. Old boolean-only destructive
requests are refused. Ordinary requests without disposal fields remain valid.
The optional disposal controls enforce their own authenticated owner permission
and `workspaceDisposal` flag; they never turn off these mandatory checks.

## Internal request contract

A trusted caller must obtain current action-specific human approval **before**
sending `discardWorkspace: true`. The controller review workflow now supplies
that authorization through its durable, one-shot claim. The worker does not authenticate a human from
this boolean or from volume metadata. Its trusted controller/SSH transport and
creator ownership remain required independently.

A destructive version-4 fence must carry `expectedWorkspace`, whose fields are:

| Field | Required binding |
| --- | --- |
| `version` | Workspace identity version 1 |
| `name` | Exact `commonroom-tool-work-<job ID>` volume name |
| `nonce` | Original 32-character lowercase hexadecimal creation nonce |
| `createdAt` | Complete original nonzero RFC3339Nano creation timestamp |

The object must match the worker journal exactly. It identifies the volume, not
the current source contents; source-revision approval and artifact-retention
checks remain separate from this whole-workspace disposal instruction. Job/provider/membership and
repository bindings still apply. The worker rejects missing/partial identities,
foreign names, wrong nonces, changed timestamps, legacy destructive versions and
identity objects supplied with `discardWorkspace: false`.

Requests and acknowledgements accept only exact, unique JSON keys. Duplicate
fields, case aliases, unknown fields, null values, malformed encoding and trailing
objects are rejected, including within the nested identity. Existing transport
size bounds remain; the binding does not authorize arbitrary paths, mounts,
Docker options or artifact deletion.

## Execution ordering

1. Snapshot and validate the request, journal and complete disposal identity
   before creating the cancellation tombstone or reaching cleanup.
2. Cancel the bound job and wait for creator ownership. While its lease is held,
   container interruption can proceed, but workspace deletion remains forbidden.
3. Reload the journal under the creator lease and compare the exact identity
   again. An identity change during the wait refuses destructive cleanup.
4. Carry the exact identity into workspace cleanup, independently check it there,
   and preserve the existing Docker volume nonce/time/labels/mountpoint checks.
5. Require no container references, remove only the matched volume, and verify
   exact-name and job-workspace absence before reporting full shutdown.

A stale binding detected initially has no cancellation/cleanup effects. If an
identity changes after a valid cancellation has begun, cancellation can already
have occurred; this does not permit removal of the changed volume or success.
A missing job journal cannot become permission to erase unknown data.

Docker still removes volumes by name. An administrator with daemon access can
race those names outside managed leases. The host, daemon and trusted transport
remain part of the security boundary. Source changes can substitute cleanup or
skip binding checks, invalidating the retention/isolation contract.

## Receipt contract and compatibility

A successful destructive acknowledgement includes `discardedWorkspace` with the
same complete identity as the admitted `expectedWorkspace`. Container-only proof,
a missing identity, a changed nonce/time or an unconfirmed removal cannot satisfy
that request. A non-destructive request rejects an unsolicited disposal receipt.
Normal non-destructive acknowledgement shapes otherwise remain unchanged.

Both peers must understand the new binding before destructive use. Older workers
reject the unknown request field, and current workers reject old boolean-only
destructive requests. Do not downgrade to an unbound request to regain
compatibility. Verify the updated worker binary as part of deployment acceptance.

A correctly bound receipt is still not permission to release storage capacity.
The owner endpoint correlates authenticated approval, controller audit and the
exact prebound storage reservation captured in the proposal. Only verified
success can release that reservation, atomically with the outcome and container
slots. Unbound charges remain. Expiry, revocation and human permissions are
enforced by the owner workflow. Artifact delivery, production allocation binding
and live verification remain unfinished.

## Verification

Fast tests use temporary worker journals and mocked Docker/transport boundaries.
They cover valid removal, absent/foreign/stale/incomplete bindings, missing journal,
identity replacement while waiting for the creator lease, retained crash discovery,
ordinary recovery, receipt mismatch, ambiguous JSON and container-only proof.
Automated tests launch no Docker workload, provider or database.

The owner workflow has three P0 refusal drivers and one combined approved-disposal
driver with mocked worker removal, receipt and capacity release. Five additional
drivers cover stale approval, remove/absence uncertainty, retained cancellation
and immutable history. Live disposal
acceptance remains pending. The existing tool-only live cleanup used disposable diagnostic data;
it does not establish a human-approved production disposal workflow. See
[workspace volumes](workspace-volumes.md), [storage accounting](storage-accounting.md)
and [release readiness](release-readiness.md).
