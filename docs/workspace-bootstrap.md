# Credential-free workspace bootstrap

## Purpose and availability

The internal tool runtime can initialize a fresh isolated workspace from the
worker-approved Git bundle. It verifies the snapshot, checks checkout limits and
creates a job-specific feature branch before the engine starts. Users have no new
setting or launch path: production pair assembly and live verification remain
unfinished, and development admission is still blocked.

This is mandatory free/core execution infrastructure. There is no off switch,
owner exemption, paid bypass or direct host-execution alternative. Existing
optional development settings cannot bypass these checks.

## Startup sequence

The trusted worker must authenticate the controller, enforce room/snapshot
approval independently, acquire the paired leases, stage the immutable approved
bundle and allocate a fresh empty job-specific local volume. Fresh-volume allocation and cleanup are provided by the version-4 worker
lifecycle, separately from this bootstrap helper. Reusing old provider or
development volumes is forbidden.

Both pair IDs and aggregate resources must be verified before either starts.
The tool starts an idle process in `/workspace`, then its fixed bootstrap helper
receives the staged bundle over stdin. The supervisor checks the returned receipt
against the original job and repository digest before it starts the engine.
The engine receives no source volume, bundle or tool-container credentials.

The tool runtime opens readiness only on the same trusted plan whose bootstrap
completed successfully and passed fresh runtime inspection. File tools reject
unready plans before talking to Docker. A new plan or matching volume labels do
not establish readiness. Readiness retains the original job lifetime; fresh call
contexts cannot revive it after cancellation, and an action ending after that
lifetime cannot return success output. The shared audited-session constructor supplies that one plan to runtime and
dispatcher; full production pair assembly remains blocked.

## Snapshot and Git contract

The helper supports Linux inside the credential-free tool container. It accepts
only the worker-supplied job ID and approved SHA-256 digest; it cannot select host
paths, executables, Docker options or environment. Its workspace is fixed at
`/workspace`. The volume must be empty, and its root cannot be a symlink.

The bundle is streamed to a private exclusive file, bounded at **256 MiB**, hashed
and flushed before Git runs. Empty, oversized, unreadable or mismatched input
fails. The worker closes the stream on return or cancellation, including a blocked
read. The helper never rewrites or deletes a prior workspace to make it fresh.

Git runs from fixed `/usr/bin/git`, with a fresh temporary HOME and an explicit
environment. System/global configuration, inherited Git variables, replacement
objects, prompts, hooks and fsmonitor are disabled. Only local-file transport is
permitted, and the container has no network. Clone uses no checkout, hardlinks or
templates. It removes the local origin and never initializes submodules.

Before checkout, bounded tree metadata admits at most **10,000 regular files**
and **256 MiB of blob content**. Absolute/escaping/ambiguous paths, `.git` path
segments, duplicate paths, symlinks, submodules, malformed object metadata and
negative/oversized sizes fail. Metadata output is limited to **2 MiB**. Highest-
priority Git attributes disable filters, encoding conversion, ident expansion
and newline transforms so source attributes cannot run programs or expand the
admitted checkout unexpectedly.

Checkout creates `feature/commonroom-<job-id>` at the verified commit. The helper
checks the resulting HEAD and verifies writable source and Git metadata using
exclusive temporary probes. Only then does it remove the staged bundle and emit
a receipt with version, job ID, approved digest, commit and branch. No source,
credentials or Git stderr is included. Raw bundle/source data never enters an
audit event. Existing durable job registration already binds the initiating
human and approved snapshot; production receipt persistence remains integration
work.

## Failure and recovery

The runtime permits one bootstrap attempt. Hash, clone, tree, checkout, output or
receipt failure does not enable tools or start the engine. Failed or partial
workspaces remain unresolved; no automatic retry, reset or adoption occurs.
A lost success receipt also requires explicit recovery instead of replay.

The supervisor retains creator ownership through bootstrap. Cancellation and
helper return alone do not prove descendant/container shutdown. Existing
immutable-ID fencing, both-role inventory and released-lease proof are still
required. Leftover volumes require explicit version-4 cleanup and staged source remains
the worker staging owner's responsibility; this helper does not remove them or declare the job settled.

The limits bound accepted input/tree content, not a filesystem quota or complete
protection against malicious Git objects. Docker memory/CPU/process limits remain
mandatory; worker disk quotas and image/Git provenance need separate verification.
An approved snapshot may itself contain secrets. This feature does not claim to
scan source or implement the planned secret-detection feature.

## Image and verification

The tool image recipe builds both Go helpers without network downloads and copies
only explicit allowlisted source/artifacts. The externally pinned reviewed base
must provide `/usr/local/bin/node` and `/usr/bin/git`; the recipe refuses a base
without either executable. It installs no mutable packages. Digest pinning alone
does not prove image contents or Git safety. No reviewed release image was built
or deployed for this change.

Automated tests use temporary directories and mocked Git/Docker/database
interactions. They cover checksum/freshness, partial checkout, unsafe/oversized
trees, controlled command/environment construction, output failures, forged
receipts, blocked-input cancellation, readiness and engine-start ordering.
No real Git clone, provider, Docker workload or database is launched by those
tests. Real Git/bundle compatibility, image contents and rootless startup require
separate live verification before enabling managed development. See
[execution progress](p0-execution-progress-2026-10-15.md).

A source fork **CAN** call the helper on a host, substitute Git/runner code or skip
receipt checks. **BUT** doing so invalidates isolation, snapshot and recovery
claims. Native macOS/Windows managed workers remain unsupported; the supported
worker target is Linux/rootless Docker with the existing Mac controller option.

See [workspace volumes](workspace-volumes.md) for allocation, explicit discard authorization and retention limits.
