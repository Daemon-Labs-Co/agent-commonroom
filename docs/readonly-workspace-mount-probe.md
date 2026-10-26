# Read-only workspace mount probe

This internal prerequisite checks a retained workspace mount inside an isolated
Linux helper. It does not export files, return a source root, fence writers, or
provide `workspacesnapshot.FrozenLease`. Users still cannot download managed task
files through this path; required file/bundle dispatch remains blocked.

## Trusted worker contract

`toolbroker.NewWorkspaceMountProbe` requires exact job/repository identity,
created workspace name/nonce/time, a fresh 32-character lowercase hexadecimal
challenge, a reviewed local immutable helper image ID (`sha256:` plus 64 lowercase
hexadecimal characters), an approved rootless daemon runner/data root, and a
mandatory `WorkspaceProbeJournal`. Nil and typed-nil dependencies refuse.

The caller must hold the real creator job lock across the entire operation,
including cleanup. Admission must establish reconciled role absence, cancellation,
current human/task/export authority and the exact retained volume. See
[retained workspace observation](retained-workspace-observation.md). Do not invoke
its zero-reference `Check` while the probe is attached: perform admission before
creation and final checks after confirmed removal. Production assembly must also
exclude every other container reference throughout the mount lifetime; this probe
alone does not implement that inventory/fencing policy.

The journal has no permissive default or production implementation:

- `Begin` must authenticate current authority/creator ownership, reserve an actual
  bounded physical container slot and durably record creation intent before any
  Docker operation. It must deduplicate attempts and retain uncertain charges.
- `Bound` must persist the exact created container ID and recheck authority before
  execution. Failure prevents start and triggers exact-ID cleanup.
- `Finish` must durably record confirmed absence or uncertainty, including IDs
  learned before a failed `Bound`. Inventory/create failures still call `Finish`
  with an empty ID and `absent=false`; the pending intent and charge remain
  uncertain for reconciliation. It must recheck current authority before any
  report release. Persistence/integrity failure must stop controller authority;
  an acknowledgment from a test fake is not durable production evidence.

One probe object admits one attempt. Name collisions refuse without adoption.
Unknown create replies retain uncertainty and must be reconciled under trusted
ownership; this code does not guess an ID from a name or automatically retry.

## Isolation and observation

The fixed helper runs as UID/GID 1000 with a read-only root, no network, all
capabilities dropped, no-new-privileges, private IPC/cgroup namespaces, bounded
memory/CPU/process/file descriptors, no logging, no stdin and no restart policy.
Its only environment entry is the fixed `PATH=/`. Its only workspace mount is the
exact local volume, read-only with `volume-nocopy`; no host bind, credentials,
socket, provider, shell or source content is supplied. Docker policy and volume
identity are inspected before start and again after successful exit.

The helper reads only `/workspace` directory metadata, fixed kernel mountinfo
and filesystem flags. It checks ownership/mode, read-only workspace and rootfs
flags, exactly one explicit read-only workspace mount and no workspace submounts.
Writable underlying filesystem superblock flags are permitted for a read-only
bind view; effective per-mount flags and independent statfs checks must be
read-only. Reports bind the challenge and mount ID. Output is bounded; duplicate,
unknown, malformed or mismatched fields refuse.

No report is released until the exact created container ID has been removed,
observed absent and acknowledged by `Finish`. Cleanup gets a separate bounded
context even after caller cancellation. Failures retain uncertainty and never
remove the workspace volume. The report describes a mount that has already
ended; it cannot authorize later source reads or serve as an enduring lease.

## Build and configuration

Build from the repository root with `docker build -f docker/export/Dockerfile
--build-arg GO_IMAGE=<reviewed-Go-image>@sha256:<reviewed-digest> .` after replacing
both placeholders with approved values. The base has no mutable default. The
compile step disables networking/module downloads. The final scratch image
contains only the helper and an empty owned workspace directory. The dedicated
Dockerignore admits only the reviewed helper source/build recipe. Record and
review the resulting local image ID before passing it to the worker.

This is mandatory internal security infrastructure with no public route/UI and
no on/off isolation flag. Existing product flags do not authorize bypasses.
Linux rootless Docker is required; unsupported platforms refuse, with no local,
SSH or uncontained fallback. Source forks can substitute dishonest runners,
journals or helper images, invalidating these managed security guarantees.

## Verification and remaining gaps

Automated tests use synthetic daemon/journal responses and no database, provider,
SSH, Docker daemon, private data or credentials. They exercise authorization and
journal ordering, exact-ID cleanup, cancellation, drift, malformed reports and
uncertain lifecycle outcomes. Parser/CLI tests use synthetic kernel metadata and
injected observations. Image recipe/context checks and an offline Linux cross-build
verify packaging inputs, not a running image.

Production controller transport and invocation, full lifetime reference
inventory, reviewed image build/live Linux mount verification, physical quota,
actual source-root binding/FrozenLease, retention recovery and authenticated file
delivery remain outstanding. No live P0 gate or production readiness is certified.

The internal [durable probe journal](workspace-probe-journal.md) retains lifecycle uncertainty across restarts and counts unresolved probes during worker admission; production controller assembly remains required.

The [guarded probe operation](guarded-workspace-probe.md) now holds creator ownership and connects the local journal with exact helper/reference checks through cleanup.
