# Fresh workspace volumes and explicit cleanup

## Purpose and availability

The internal paired worker now allocates fresh tool volumes and can prove their
removal during explicit cancellation/reconciliation. This is mandatory free/core
safety infrastructure, without an off switch, permission exemption or paid bypass.
No production development-launch endpoint is enabled. Normal-completion artifact
retention and release wiring remain unfinished.

New paired starts require lifecycle **version 4**, execution
`separated-pair-workspace-v1`. Version-3 pairs remain readable and recoverable,
but cannot start new work. An older worker rejects version 4 instead of ignoring
workspace metadata and falsely declaring container-only shutdown.

## Allocation and durable identity

The supervisor holds the original membership/job leases. Before Docker volume
creation, it durably records a workspace creation intent containing the fixed
job-specific volume name and a fresh random nonce. Those are non-secret identity
facts, not capabilities, paths or provider credentials. Docker creation uses only
the local driver, mandatory job/repository labels and the nonce label; no driver
options or host bindings are accepted.

`WorkspaceIntent` belongs to the same trusted tool plan used for creation,
bootstrap and actions. `PrepareVolume` refuses successful inventory showing an
existing name, and refuses unavailable/ambiguous inventory. It creates once,
checks the returned name and verifies full local-volume policy. The daemon's
creation time is bound to the nonce and durably recorded before containers can
be created. Container inspection repeats this exact identity check. A new volume
with the same name but another timestamp/nonce cannot silently replace it.

Failed/uncertain creation is not retried or adopted. A crash between creation and
identity publication leaves the original nonce in the journal, allowing cleanup
to discover and verify that exact pending volume. Bootstrap additionally requires
its contents to be empty before writing source. Matching labels alone never make
an old workspace ready.

## Cancellation and reconciliation

Version-4 recovery uses `cancel-pair-workspace-v1`, not either older cancellation
command. The existing owner+CSRF recovery route preserves version, membership,
approved repository and original initiating human; the diagnostics flag cannot
weaken recovery checks. Version-4 acknowledgements require workspace removal
proof as well as both-container shutdown. Legacy requests/acknowledgements cannot
settle a workspace-bearing job. No work is restarted.

The worker tombstones cancellation first. While the creator lease remains held,
it may stop engine/tool containers but cannot remove the workspace or confirm
cleanup. Only after acquiring creator ownership does it inspect the approved
local daemon's data root and reconcile the volume.

Removal requires verified name, nonce, creation time, labels, local driver,
canonical mountpoint and absent driver options. A successful all-container query
must show no references to the volume. Identity is checked again immediately
before removal. With explicit trusted `discardWorkspace:true` authorization, the worker removes
only the intended ephemeral volume, then
requires successful name and job-workspace inventory showing absence. A daemon
error, unknown extra workspace, in-use volume, replacement or partial removal
stays unresolved and continues blocking that membership.

Ordinary cancellation does **not** authorize erasing unfinished source. The
worker protocol's `discardWorkspace` bit defaults false and is refused by older
protocol versions. A present workspace stays retained/unresolved without that
explicit authority; an already absent workspace may still be reconciled safely.
The current HTTP recovery route never grants this bit, so its existing stop UI
cannot silently erase workspace data. The future action-specific human approval
or artifact-retention workflow must authorize discard before public release.

Explicit trusted discard affects only that job's ephemeral tool workspace,
including unfinished edits. It never removes the approved host snapshot,
room/audit history, provider credentials or exported artifacts. The future
normal-success path must finish artifact delivery/retention before teardown.
No existing live workspaces are purged by merely upgrading the code.

## Named-volume limitation and trusted operations

Docker volumes do not expose an immutable removal ID. Cleanup therefore combines
nonce/time verification with exclusive creator ownership; all managed creation and
removal of that name must obey the same lease. An administrator with direct daemon
access can still race or forge a replacement. That administrator, the worker
account and daemon remain trusted. Neither model nor tool container receives
Docker access. Do not describe this as protection against a hostile daemon owner.

Do not delete journals, reuse job IDs, manually relabel old volumes or downgrade
worker binaries to force settlement. Preserve uncertain state for owner review.
There is no automatic destructive fallback after failed verification.

## Verification and remaining work

Tests use temporary journals/leases and mocked Docker/database interactions. They
cover intent-before-create, refusal of existing names, immutable bindings,
creation-ack loss, in-use/replacement/extra volumes, held ownership, partial cleanup,
version downgrade, explicit owner recovery and disabled diagnostics. No live Docker
volume, provider or database is used. Actual rootless creation/removal and daemon
metadata compatibility remain unverified.

Internal Codex assembly and private audit transport are implemented. Public
dispatch, reviewed release images, commands/artifacts, storage quotas, approved
disposal and live acceptance remain P0 work.
The credential-separation admission block remains enforced. See
[paired lifecycle](worker-pairs.md) and [bootstrap](workspace-bootstrap.md).

A source fork **CAN** skip journalling or substitute an unsafe Docker adapter.
**BUT** it then loses fresh-workspace and cleanup guarantees. Native macOS/Windows
workers remain unsupported; the controller may run on macOS with a Linux/rootless
worker.
