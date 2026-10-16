# Private file and lease safety

**Tier: Free mandatory security hardening.** Security-critical file readers validate
the opened file descriptor instead of inspecting a pathname and reopening it.
This reduces substitution windows in worker records, authentication credentials,
configuration, bootstrap keys and persisted human sessions.

## Enforced rules

Supported Linux/macOS hosts open these files with no-follow and nonblocking flags.
The descriptor must refer to a regular file owned by the current effective user,
with no group/other permissions, exactly one hard link, and a bounded size. Readers
also bound the actual bytes read, so growth after the descriptor check cannot cause
an unbounded read. FIFO/device/directory inputs, symlinks and hard-linked aliases
are rejected. Lease and credential lock files use the same rules before flock.
Open policies reject destructive flags such as truncation before touching a file.

Bounds: worker records/configuration/locks 8 KiB; provider credential JSON,
owner RSA material and session stores 64 KiB; bootstrap setup code 128 bytes.
Caller-specific JSON/key/job-binding checks still apply after reading. This helper
is not a JSON validator, credential classifier or authorization substitute.

## Operational requirements

Keep private files owned by the account running their controller or worker, with
mode 0600 (0400 is suitable for a read-only file). Maintain trusted, owner-controlled
parent directories; use copies rather than hard links for backups. Never repair
permissions by making credentials world-readable. Reinstall worker binaries to
apply worker-side checks; rebuild/restart the controller for session/key checks.
These source changes do not imply the running services were upgraded.

An unsafe credential is reported unavailable; unsafe configuration or session
persistence prevents startup. Unsafe worker state/locks refuse admission or
reconciliation instead of guessing that shutdown occurred. Inspect the affected
file metadata, ownership, link count and parent directory permissions using the
trusted deployment account. Restore a correct private file from a trusted backup;
do not erase or rewrite audit records to resolve a file error. Raw credentials
are not diagnostic output.

## Scope and limitations

`openPrivateRegular` pins descriptor validation to the opened inode;
`readPrivateBounded` bounds actual reads. They are used by worker record/config
readers, lease creation, credential locks/status/export/copy, owner bootstrap/RSA
reads and persisted-session restore. Atomic writers retain their existing
exclusive-create or temporary-file/rename behavior.

Parent path traversal and host administrators remain trusted. A privileged actor
or process with full access to the same deployment account can modify directories,
replace lock inodes, change files, run other programs or bypass these binaries.
No-follow on the leaf does not prevent parent-directory substitution, prove file
content authenticity, provide independent audit rollback detection or guarantee
immutability against the host owner. Independent archives and stronger host
administration boundaries remain separate work.

This hardening has no off switch. Product flags cannot weaken authentication,
audit, isolation or private-file enforcement. Existing session-persistence and
connection/testing/diagnostic preferences retain their normal meanings; enabling
a capability always applies these checks. Source forks **CAN** weaken validation,
**BUT** expose credentials and risk locks being acquired on the wrong inode.

## Tests

Fast parallel-safe tests use only synthetic files in temporary directories. They
cover symlinks, hard links, unsafe permissions, directories/FIFOs, oversize files,
exact read bounds, missing-file bootstrap behavior, descriptor stability after
pathname replacement, aliased leases, credential exports and worker records.
Destructive flags are rejected without truncation. Existing RSA/session, worker
and credential tests run against the same helper. Tests never read live provider
credentials, connect to a database or launch real provider agents. Live installation
upgrade/ownership compatibility remains an operator acceptance check.

## Controller configuration

`.data/local-config.json` (which can contain a database credential) and
`.data/development-workspaces.json` now use the same validated descriptor with an
8 KiB read limit. Both must be owned by the controller account, mode 0600 or
stricter, single-link regular files. Symlinks, hardlinks, shared permissions, FIFOs
and oversized files stop startup. Missing files preserve environment/default
configuration. Development environment overrides retain their documented
precedence; supplying both values avoids reading the development file.

Local configuration accepts one JSON object with only `databaseUrl`,
`remoteWorkerAlias` and `remoteWorkerEnabled`; unknown fields, trailing objects
and invalid types fail with a generic error that excludes input values. Store
only these supported properties. Repair an old shared file by restricting its
permissions and ownership; do not substitute a symlink. Parent directories must
still be trusted. These startup security checks have no optional feature flag.
Synthetic temporary-file tests cover both loaders without real credentials.

## Crash durability of sessions and emergency stops

Saved session replacement syncs the completed temporary file, renames it, then
syncs the parent directory before acknowledging persistence. Disabling session
persistence also syncs removal. Emergency-stop marker creation syncs both file
and directory; repeated stop attempts retry directory synchronization. Owner
recovery syncs marker removal; if it fails, execution remains latched and marker
creation is retried. Failed durability operations are reported rather than
acknowledged as successful. These safeguards are mandatory and have no flag.

A directory-sync error after rename is ambiguous: new state may already be
visible. Do not interpret an error as proof that the old file remains on disk.
Repair storage before restart/retry; the existing runtime latch remains in force.
Filesystem/hardware support, protected parent paths and writable marker storage
remain requirements. Synthetic tests inject directory-sync failure and verify
ordering/cleanup; physical power-loss durability and remote filesystems have not
been verified. These files are operational state, not independent audit archives.

## Exclusive per-turn credential staging

This section describes the existing chat/login credential staging path. The
internal separated Codex engine instead provisions credentials through a bounded
private stdin frame; see [engine bootstrap](codex-engine-bootstrap.md). Do not
apply the host-bind mechanism to its tool container.


The worker copies a bounded valid-JSON credential into a newly created per-turn
file using exclusive creation. Existing files, symlinks and hardlinks are refused
without overwriting/removing them. The staging parent must be a real directory
owned by the worker account with owner-only permissions; a symlinked/public parent
is refused. Host ancestor paths and the worker account remain trusted.

After writing the file, permissions are explicitly set to 0444 for the rootless
container's mapped UID rather than depending on the host umask. The 0700 parent
prevents other ordinary host accounts from traversing to the file; the container
receives a read-only bind. Failed newly-created writes remove their partial file.
Existing per-turn cleanup removes the private directory. No credential contents
are logged in errors. These mandatory controls have no product flag.

This is secure staging, **not separation from tools inside the container**. Tools
sharing the provider UID can still read the mounted/copied credential. Public development admission therefore remains blocked while the replacement
separated engine/tool pipeline completes production acceptance. Synthetic fixtures verify destination
non-overwrite, input bounds, parent privacy and resulting mode; live mapped-UID
readability and crash-temp cleanup have not been verified.

### Provider credential revocation durability

Credential publication and local forgetting now sync the private store's directory
before success is acknowledged. Missing-file sign-out retries also repeat the
sync. See [sign-in durability and limitations](connection-sign-in.md#durable-local-credential-publication-and-sign-out).
No claim of provider-wide revocation or physical erasure is made.
