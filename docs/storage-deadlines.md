# Bounded storage waits

**Tier: Free mandatory execution safeguard.** State replay, event transactions and
integrity verification have a maximum five-second cooperative storage context.
Earlier caller deadlines are preserved. A stalled database operation must not hold
the controller mutex indefinitely and prevent emergency cancellation.

## Behavior and recovery

State reads return an error if querying/streaming/replay exceeds the context.
Commands that cannot load state reject the operation rather than acting on a
partial projection. An append batch bounds begin, advisory-lock wait, queries,
inserts and commit together; failure rolls back without dispatching new work.
Existing caller-specific storage-failure handling remains in force. Integrity
verification timeout latches the execution stop and cancels active contexts.

Application emergency stop cancels contexts before auditing its own stop. It may
still wait for the controller mutex, but cooperative storage work now releases
that mutex on deadline. The regression test stalls a mocked state query while it
owns the mutex, then verifies that the authenticated application stop completes
and cancels the synthetic turn. Worker shutdown still requires separate fencing
proof; cancellation of a context alone does not prove physical termination.

If reads/commits time out, investigate storage availability, lock contention and
history size. Do not disable authentication, audit integrity or container isolation.
Inspect pending intent/outcomes before retrying; failed persistence does not prove
an external action never ran. Follow existing owner incident/worker recovery when
execution is latched off. No actions are automatically replayed after recovery.

## Implementation and limits

`storageContext` supplies the cap. `load`, `appendBatch` and `integrityLocked` use it.
Replay also checks context between records and before/after projection. The private
`storageTimeout` field permits shorter deterministic tests, never a longer or
unlimited production wait. No environment/product off switch is provided.

This is not an absolute real-time guarantee: contexts require cooperative database
and network implementations. Filesystem stalls, arbitrary non-cooperative code,
CPU-heavy projections and other locks are not forcibly interrupted. Not every
read-only administrative SQL query is covered by these three entry points. Large
histories still need bounded/persisted projections and operational monitoring.
Multi-controller behavior remains unsupported. Source forks **CAN** remove the
cap, **BUT** then storage stalls can block cancellation indefinitely.

## Verification

Mocked tests stall query start, row iteration, transaction begin and insert work,
then assert deadline failure and no committed partial data. They preserve shorter
caller deadlines, reject internal unlimited overrides, test emergency-stop mutex
release and confirm integrity timeout cancels jobs. Fixtures and transport are
synthetic; no database connections or live providers are used. Runtime deployment
and actual driver/network outage acceptance are not claimed by unit tests.

## History reads

History endpoints now share the same maximum five-second storage deadline across
latest history-setting lookup and the event query/row scan. The setting helper
also applies the cap when called directly and preserves an earlier caller
deadline. Paged and legacy bounded-history responses fail rather than returning
partial rows when context expires, even if the backend omits a row error.

These mandatory limits have no feature flag: switching enriched history
presentation off does not remove the deadline. If history times out, repair or
optimize storage/indexes and retry; do not interpret an error as empty history or
raise limits to mask a stalled backend. Authentication and existing owner access
remain unchanged. The failure uses the existing history API error response.

Mocked settings/query/row stalls use short deadlines and synthetic history only.
Tests verify no partial page is acknowledged and earlier caller cancellation is
retained. Live storage contention and history index performance are unverified.
Cooperative DB cancellation remains required; this is not a general volumetric
request or total-storage quota.
