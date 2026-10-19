# Private engine credential custody

[Credential recovery](worker-credential-recovery.md) · [Bootstrap](codex-engine-bootstrap.md)

A completed model turn does not establish custody of tokens refreshed inside its
engine. The internal Codex worker now captures those credentials through a private
channel and commits them only after execution is fenced. This is implemented
internally with hermetic tests; reviewed image and real-provider acceptance remain
required. Public repository-development dispatch is still blocked.

## Normal sequence

1. The connection lease and durable uncertainty marker precede credential
   provisioning. Other turns and sign-in/sign-out cannot take the same lease.
2. The callback driver validates `turn/completed`. A trusted finalizer runs before
   closing the Docker attachment, so stdin EOF cannot prematurely remove the
   temporary engine credential home. Failure withholds the model result.
3. The runtime re-inspects the original running engine and makes one fixed
   `docker exec` invocation as UID 1000. It executes only the bootstrap's private
   export client, not a provider, model tool or host shell. There is no retry.
4. The client sends exactly `COMMONROOM-CREDENTIAL-EXPORT-V1\n` to
   `/tmp/commonroom-engine/credential-export.sock` inside that engine. The private
   0700 parent protects the socket; it is not mounted into tools or the controller.
5. Bootstrap accepts one request, sends SIGTERM to its reviewed provider child,
   waits for `close`, then opens its fixed auth file using no-follow/nonblocking
   descriptor checks. The file must be regular, single-link, UID-1000-owned,
   owner-only and 2–65536 bytes. JSON must be subscription-mode with an access token
   and no API-key fallback. The worker independently enforces the strict schema.
6. Bootstrap replies with the same magic, a four-byte big-endian length and the
   exact JSON bytes. The private client rejects wrong magic, truncation, trailing
   bytes and oversized output. Its deadline is ten seconds; the runtime bounds
   invocation to twelve seconds and the original job lifetime. Bootstrap remains
   alive for export and supervised fencing; child exit is not container absence.
7. The worker re-inspects engine identity/policy and stages one snapshot under the
   existing lease. Connection/job scope and account identity must match. Different
   account IDs, duplicate/unknown fields, API keys and malformed token types refuse.
8. Normal supervisor shutdown removes and reconciles both immutable container
   identities. Before committing, custody reads the private journal and requires
   matching full scope and engine/tool absence evidence. Unpublished evidence
   cannot authorize saving or lease release with a staged snapshot.
9. Custody compares the official credential with its original snapshot, writes the
   final bytes through private atomic replacement, syncs the file/directory, and
   clears the recovery marker. It then clears owned buffers and releases the lock.
   A changed credential is never overwritten. Retained source is independent.

No credential payload, token hash, file path chosen by a model, or export result
is sent to model callbacks, application audit or controller responses. The trusted
worker/Docker account necessarily handles the private bytes. Worker subprocess
stderr is not forwarded into model or controller output.

## Failure and recovery

| Failure | Behavior |
| --- | --- |
| Provider/callback/audit failure | Do not invoke successful-turn finalization; fence execution and retain recovery. |
| Missing old-image export protocol, invalid frame or timeout | No fallback or reconnect; withhold the turn result and keep recovery. |
| Unsafe credential inode or strict-schema/account mismatch | Reject the snapshot and keep the prior official file. |
| Container absence or durable journal proof unavailable | Keep the staged snapshot and lease for reconciliation; a process crash loses the snapshot but leaves recovery required. |
| Existing official credential changed | Do not overwrite it; retain recovery and report custody failure. |
| Atomic replacement fails before commit | Preserve the old credential, retain recovery and report failure. |
| Post-rename directory sync fails | New bytes may be visible but are not acknowledged as durable; keep recovery. |
| Marker removal sync fails | Report failure even though credential replacement was already durable. The marker may be absent; this is uncertain acknowledgement, not unknown token bytes. |

Release errors remain errors on repeated calls rather than turning a failed save
into success. Execution that is actually fenced may release the lock for explicit
human recovery after a save failure; unresolved execution must not do so. Use the
[recovery procedure](worker-credential-recovery.md) instead of manually deleting
state or replaying uncertain jobs.

## Controls, deployment and limits

This is mandatory authentication/isolation infrastructure with no bypass flag,
paid exception or host fallback. Existing execution/sign-in permissions and flags
remain enforced by their server paths. Disabling optional UI does not clear
uncertainty or weaken the engine boundary. Runtime paths, images, helper source
and Docker selection are trusted worker policy; models cannot supply them.

Rebuild and review the engine image, then select its immutable digest together
with matching worker code. Existing pinned images do not gain this protocol when
source files change. Unsupported images fail closed. Verify the reviewed binary
stops its token writers on shutdown; the worker's final container fencing remains
required regardless. The trusted bridge-only binary must not leave independent
credential-writing descendants. A host administrator/source fork can bypass the
protocol or forge metadata; this is not protection against that administrator.

Live acceptance must exercise real subscription refresh, graceful shutdown,
crashes between export/fencing/rename/sync, ambiguous persistence and recovery,
and two worker processes contending for the same connection. Record reviewed
image/binary versions and redacted observations, never credentials. No image
build, provider session or Docker workload was run for the automated tests.

Hermetic tests cover the bootstrap, private framing/client, policy drift, one-shot
runtime export, finalizer-before-close ordering, account/scope rejection,
connection contention, durable-journal checks and combined callback/shutdown flow.
Four prepared P0 credential scenarios now have real custody-backed drivers; the
remaining 55 scenarios and all live acceptance are still outstanding.
