# Constrained tool-container launch

## Purpose and current availability

The internal `toolbroker.ToolContainerRuntime` implements the tool half of the
paired worker launcher. It performs fixed Docker creation, full reported-policy
inspection and immutable-ID startup. It does not launch a model, bootstrap a
workspace, authenticate a controller, or enable managed development. No paired
launch endpoint or production pair adapter calls it yet.

Users have no new configuration or button. Existing development settings cannot
bypass the credential-separation admission block. This is mandatory security
infrastructure with no on/off flag, owner exemption or paid bypass.

## Worker contract

Only a trusted isolated worker may construct this module. Before creation, that
worker must authenticate the controller, acquire the paired job/membership
leases, persist creation intent, and bootstrap a fresh job-specific volume from
the approved snapshot without provider credentials. This module verifies that
volume exists and reports the expected local driver, name, labels, canonical
mountpoint and no driver options. It cannot attest the volume's contents or its
history. Reusing a previously credential-bearing volume is forbidden.

`CreateArgs` derives the same fixed tool policy as `RunArgs`, but uses Docker
`create` without detached-run or automatic-removal flags. The container remains
stopped while both pair identities are durably bound. Explicit cleanup preserves
crash evidence instead of allowing automatic deletion to erase it.

`Create` refuses an existing fixed-name container and unavailable/ambiguous
inventory. It checks the prepared volume before creation, preventing Docker's
implicit creation of a missing volume. It accepts only a full immutable ID and
verifies the stopped container, image, labels, environment, mounts, namespace and
resource settings against the trusted plan. Only the `created` state is accepted;
an exited or running container cannot be adopted as a fresh creation.

The paired supervisor must durably bind both role IDs and validate their aggregate
resources before calling `Start`. `Start` rechecks the complete stopped policy,
starts only the recorded immutable ID, requires a matching response and verifies
running policy afterward. A changed name cannot retarget execution. Missing or
changed evidence refuses success. No model supplies Docker flags, host paths,
images, environment, namespace choices or container identities.

The constructor binds the original job lifetime; a fresh call context cannot
revive cancelled or revoked authority. The existing trusted Docker runner must
still verify the approved local rootless daemon on every operation. There is no
direct host execution or remote-context fallback.

## Uncertain outcomes and cleanup

A runtime instance attempts creation once and startup once. Failed inspection,
creation errors, malformed IDs, changed policy, start errors, mismatched replies
or missing running evidence never cause an automatic retry. A failed creation
cannot later acquire startup authority merely because subsequent inspection
looks acceptable.

The supervisor retains the durable creation intent for recovery, including the
case where Docker created a container but its reply was lost. The runtime never
blindly removes a name or declares a whole pair stopped. Existing fixed-name
identity discovery, immutable-ID cleanup, both-role inventory and released
supervisor ownership remain required. See [paired lifecycle](worker-pairs.md).
Volume cleanup and bootstrap lifecycle remain separate unfinished P0 work.

## Tests and verification limits

Fast tests inject Docker responses and synthetic volume/container metadata; no
Docker daemon, database or provider is used. They exercise creation before start,
immutable IDs, original lifetime/revocation, unavailable inventory, name collision,
missing/unsafe volumes, unsafe container policy, exited/running adoption refusal,
changed policy before startup, ambiguous start responses, incomplete startup and
refusal to replay uncertain operations.

The same full inspection implementation checks created and running policy with
an explicit expected state. Existing inspection tests continue covering unwanted
mounts, credentials, environment, driver options, namespaces and resources.

No live rootless runtime or reviewed tool image has been verified for this change.
The reviewed engine driver, fresh bootstrap, authenticated audit transport,
production pair integration and runtime acceptance gates remain open. This module
is an implementation step toward P0 closure, not evidence that managed development
is ready. See [execution progress](p0-execution-progress-2026-10-15.md).

A source fork **CAN** call the runtime outside the paired supervisor, substitute
an unsafe runner or skip inspection. **BUT** that invalidates the admission,
credential-separation and crash-recovery guarantees.
