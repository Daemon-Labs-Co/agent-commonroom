# Capture container admission

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


The internal `toolbroker.CaptureContainerPlan` describes and validates the isolated
helper that transfers selected task outputs into private custody. This is a
mandatory security boundary, with no off switch or host/SSH fallback. It is not a
user launch API and does not enable required file-task dispatch.

## Trusted configuration and use

A supervisor supplies a finite `workspacecapture.Request`, reviewed `sha256:` image
ID, approved rootless Docker data root and retained object-volume creation record.
For example, an execution's existing `commonroom-tool-work-<job>` source is mounted
read-only at `/workspace`; its separate `commonroom-custody-<job>` volume is writable
only at `/objects`. Both use recorded nonces and creation timestamps. No arbitrary
volume names, paths, commands or mount options are accepted.

`NewCaptureContainerPlan` validates those inputs. `CreateArgs` returns fixed Docker
creation arguments and a bounded canonical encoded request. It does not execute
anything. Before starting, during capture and after exit, a trusted runtime must
fetch inspection metadata directly and call `VerifyInspection` with the durably
recorded exact container ID and expected `created`, `running` or `exited` phase.
Unknown phases, missing safety metadata or successful-state mismatches fail closed.
`VerifyAdmission` performs two bounded independent inspection/inventory passes.
It requires that both volume references, the job and the exact helper name resolve
only to the recorded ID, and that the source/custody volume catalogs have exactly
their recorded members. Any returning writer, extra volume or changed second-pass
policy refuses admission. The caller must hold creator ownership across both passes.

The policy fixes UID/GID1000, read-only root, no network, no capabilities,
no-new-privileges, private IPC/cgroup namespaces, resource limits, no restart,
no healthcheck and no logs. Only the reviewed helper executable may run. Its
non-TTY interactive stream belongs to the authenticated supervisor, never a model.
Inspection requires exactly the two named local volumes, their canonical approved
mountpoints and ownership labels, no local-driver options, and no other mounts,
host binds, devices, sockets or inherited environment. Docker inspection object
size/depth/duplicate-field restrictions also apply.

## Authorization, limitations and remaining assembly

Names, labels and timestamps are creation evidence, not capabilities. The caller
must still authenticate the human and current task/export policy, own the creator
lease, prove all writers fenced, reserve account capacity, durably journal creation
and bind the exact ID before execution. It must independently enforce physical
source/object quotas, inventory both volumes' references, authenticate the private
capture stream and durably acknowledge custody operations. A reported writable
volume or configured memory limit does not prove a disk quota.

After capture the runtime must observe successful helper exit, remove its exact ID,
confirm absence and current authority, and retain volumes, charges and uncertain
objects on failure. Inspection alone cannot prove image provenance, atomicity,
shutdown, durable custody, exclusion of other containers or a trustworthy daemon.
A source fork that supplies dishonest inspection invalidates isolation guarantees.
No runtime wiring, volume provisioning, physical quotas or release authorization is
added here. Captured reports remain private metadata; public file delivery, format
validation and human acceptance are separate requirements.

Automated tests use synthetic inspection metadata, with no Docker/database/provider
calls. They cover accepted phases and finite launch input, plus writer, volume,
identity, command, privilege, resource and malformed-metadata substitutions. They
do not certify a real image or Linux deployment.

[Capture supervision](capture-supervision.md) now supplies the one-shot lifecycle
around this plan, using mandatory injected daemon/stream/journal dependencies.
Production deployment and physical quota/authority assembly are still required.
