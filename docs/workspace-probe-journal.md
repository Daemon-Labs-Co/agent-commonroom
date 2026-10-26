# Durable workspace probe lifecycle

The internal worker journal retains read-only mount probe intents and cleanup
outcomes across restarts. It supplies the local persistence portion of
`toolbroker.WorkspaceProbeJournal`; it does not yet assemble a production exporter
or controller transport. Managed file/bundle dispatch remains blocked.

## Usage and authority

Trusted worker assembly constructs `newWorkerProbeJournal` with the absolute,
canonical, owner-private worker job directory and a mandatory `workerProbeGuard`.
It passes that journal to `toolbroker.NewWorkspaceMountProbe`. The caller must
hold the actual creator job lock throughout probe creation, execution and cleanup,
as described in [retained workspace observation](retained-workspace-observation.md).
The journal does not acquire that lock on behalf of callers. There is no HTTP,
model or plugin entry point and no permissive guard.

The guard must bind the exact intent to current human/task/export authority,
creator ownership, approved reviewed image/rootless daemon and full container
reference inventory. `Acknowledge` must durably audit each stage with its exact
intent and deduplicate by job/challenge/stage; begin must also enforce account-wide
capacity across workers. Guard callbacks run under the admission lock and must
not re-enter journal/admission operations. `StopAuthority` must revoke live controller/worker
execution when persistence becomes uncertain. These are trusted requirements,
not proof supplied by an LLM. Production guard/transport assembly remains missing.
The guard must retain its original authorization binding across cleanup's fresh
bounded context; a missing request-context actor cannot become implicit approval.

## Durable transitions

Each unique job/challenge owns one private bounded `.probe-<job>-<challenge>.record` file. Its versioned history contains:

1. `begin`: reserve a local slot and persist exact job/repository/image/volume
   name, nonce and creation time before any Docker operation.
2. `bound`: persist the exact created container ID before helper execution.
3. `finish`: retain actual absence or uncertainty. A known ID may be recorded
   even when binding failed; an empty ID requires `absent=false`.
4. `confirmed`: only after an absent finish, durable controller acknowledgment
   and current authority. This releases the local probe slot, never workspace
   data or storage charges.

A failed/unknown create keeps its begin and explicit uncertain finish. A lost
supervisor retains its last known state. Neither restart nor time alone frees
capacity. Intent IDs cannot be reused, even after confirmation; another challenge
for the same creator is refused while an earlier attempt remains unresolved.
Illegal, duplicate, out-of-order and mismatched transitions refuse. The
[recorded-ID recovery operation](workspace-probe-recovery.md) can now explicitly
reconcile known helper IDs with fresh authority. There is no automatic retry,
unknown-ID adoption, record deletion, refund or volume disposal.

The shared `.admission.lease` serializes journal transitions and normal worker
admission across processes. Unresolved probes, including pending recovery, count against the existing worker
container ceiling alongside unresolved creator roles. Retained creator role slots
remain conservatively charged, even after their individual containers disappear.
Ordinary dispatch also rejects unreconciled live managed container inventory;
this journal does not weaken that check. Local counts cannot replace actual
rootless daemon inventory or account-wide reservations supplied by the guard.

## Persistence and failure behavior

Updates preserve prior lifecycle entries, publish a complete owner-only temporary
file, sync it, rename it and sync the parent directory. Historical entries are
not edited by supported transitions. This is private local operational evidence,
not a tamper-proof substitute for append-only controller audit. Deployment must
trust the worker account, parent directories, filesystem and hardware guarantees.
Source forks can replace the guard or persistence code and invalidate security.

Replay checks regular single-link owner-only files, canonical generated JSON,
version, exact filename/intent and complete transition order. Symlinks, hardlinks,
public files, truncated records, duplicate/unknown JSON fields and malformed
histories refuse. Directory scans stop at 16,384 entries, at most 1,024 probe
records, and 8 KiB per record. Limits include confirmed records; reaching them
requires retention/recovery design, never silent history removal.

Write/sync or controller-audit failures attempt to persist
`.workspace-probe-integrity-stop` and independently call `StopAuthority`. New probes
and normal managed admission refuse while that marker exists, including after a
restart. Marker publication failure is itself an integrity incident: live authority
must remain revoked and operators must preserve the directory for investigation.
There is no automatic clearing endpoint or safe-to-ignore marker exception.
Existing live jobs require the guard's revocation integration; checking admission
alone cannot cancel them.

Known cleanup evidence is retained even after authority revocation or a prior
integrity stop, but no report release is acknowledged. Attempts revoked before
confirmation remain charged. A final authority check also runs after confirmation
is published: revocation during that publication withholds the report, but does
not undo already acknowledged absence or re-charge the removed container slot. A failure after rename can leave visible state without a durable
acknowledgment; the operation still fails and the stop requirement remains.

## Configuration, tests and remaining work

No new product feature or UI is exposed. This mandatory internal isolation/audit
infrastructure has no on/off flag or host fallback. Existing default-off task
feature flags do not authorize bypassing it. Operators cannot enable file export
by creating a journal or changing a feature flag.

Tests use real private temporary files and kernel locks with synthetic guards,
mocked lifecycle Docker fixtures and no database/provider/SSH/credentials. They
cover restart replay, concurrent instances, exact binding, retained capacity,
corrupt/private-file failures, revocation and ambiguous publication/audit failures.
`make test` and `make check` include these tests.

Production controller guard/transport, account-wide and live daemon admission,
production invocation of the guarded probe operation, unknown-ID/incident reconciliation, retention/compaction,
reviewed live Linux image verification, actual frozen source-root authority,
physical storage quotas and authenticated delivery remain outstanding. Local
journal tests certify neither physical quota nor any live P0 release gate.

The [guarded probe operation](guarded-workspace-probe.md) now assembles this journal with actual creator ownership and the fixed helper; controller transport remains required.
