# Object-reader container admission

The object-reader plan describes and verifies one credential-free isolated reader
for an exact sealed receipt/challenge. It is a security boundary foundation, not
permission to launch or release bytes. It creates no product switch or host fallback.

## Configuration and fixed policy

Construct `toolbroker.NewObjectReadContainerPlan` from the exact finite request,
reviewed lowercase SHA-256 image digest, independently authenticated rootless Docker
data-root observation and retained object-volume creation identity. The request must
fit the helper's bounded argument, and the volume must be the exact job custody
volume with a nonce and creation timestamp. These are metadata bindings, not proof
of actual storage, physical quotas or current human permission.

Create arguments pin the image with no pull, fixed entrypoint and private supervisor
stdin/stdout, UID/GID1000, `/objects` working directory and a single read-only object
volume mount with nocopy. The root is read-only; network, capabilities, privilege,
healthchecks, restart, logs, extra environment, host/device/credential mounts are
excluded. Memory/CPU/process/file limits are fixed. No workspace source or engine
credential volume belongs in this container.

The supervisor MUST first durably reserve account capacity and bind the actual
creator/worker, approved image/volume, finite request and exact created ID. It must
fence writers and independently enforce current release/retention/physical quota
policy. Never run `CreateArgs` directly as a substitute for that lifecycle.

## Actual observation

`VerifyInspection` accepts only a durably known exact ID and created/running/exited
phase. It compares actual bounded container and volume metadata with exact image,
arguments, labels, state, rootless helper restrictions, volume nonce/creation time/
mountpoint and read-only single mount. Bind-backed volume options, writable views,
extra mounts, changed entrypoint/environment or unsafe state fail closed.

`VerifyAdmission` fetches container/volume metadata, all references to the object
volume/job/helper name, and the job's custody-volume catalog twice. Every container
reference must be the exact known helper ID and the catalog must be the exact retained
volume. Extra writers, drift between passes, ambiguous/missing observations or stale
identities refuse. The operation has a five-second deadline and performs only bounded
read-only queries. It never discovers/adopts an ID, creates/starts/removes a container,
releases bytes or refunds storage.

A shared private host-policy validator preserves the same restrictions for capture
helpers; image/arguments/mounts and authority remain independently checked by each
plan. No synthetic observation is inserted into production validation. Daemon
administrators and actual creator/writer exclusion remain trusted external boundaries.

## Verification and remaining work

Tests use synthetic complete inspection metadata and mocked Docker runners. They
cover fixed arguments and all three phases, unsafe image/entrypoint/environment/
privilege/root/network/TTY/restart/exit/mount policies, volume nonce/replacement/options,
extra references/catalog entries, second-pass drift and malformed/unpinned metadata.
Existing capture admission tests remain passing after sharing host restrictions.
No real Docker, database, provider, credentials or private messages are used.

Supervised creation/start/exit/known-ID cleanup/recovery and shared account/local/
controller lifecycle journals still require assembly. Deployed authentication,
actual release/quota/account witnesses, physically enforced storage provisioning,
end-to-end task artifact delivery/human review and reviewed image/latency/live
acceptance remain gaps. Required file-task dispatch stays blocked. This admission
validator is not a live P0 gate pass. Source forks can falsify runners/metadata or
launch on a host and invalidate the mandatory isolation contract.

See [fixed helper](isolated-object-reader.md),
[controller read session](controller-artifact-read-session.md) and
[mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).
