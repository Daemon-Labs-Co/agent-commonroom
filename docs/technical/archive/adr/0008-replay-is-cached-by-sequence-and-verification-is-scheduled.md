# Cache the replayed state by sequence and schedule chain verification

Status: accepted

Every request replayed the whole event history, and every built workspace replayed it again
every 30 seconds to verify the hash chain, so cost grew with requests times history and with the
number of active workspaces. With many workspaces in one process this is the first scaling limit.

Decision. Behind `COMMONROOM_PROJECTION_CACHE` (default off) the server keeps the replayed
projection and its position `(sequence, hash)`, reads only newer events, and returns a deep copy
of the state with derived views computed on the copy. The recorded history remains the only source
of truth: whenever the cache cannot prove its position is still a prefix of the history, it
rebuilds from the full history. Reducers are one function per event kind, so the replay rules are
no longer one 700-line function.

Chain verification changes with the cache on: incremental every hour (records after the last
verified checkpoint), full at startup and every 24 hours, both configurable. A cache position
mismatch triggers a full verification at once. The owner's integrity view and recovery are always
full. Idle workspaces are released in multi mode so memory follows use, not the number of workspaces.

## Consequences

- Tamper detection is slower for old records: an in-place edit of an old record is found by the next
  full verification (up to 24 hours), not within 30 seconds. New and latest records are caught by the
  next incremental check, the next write or the next cache refresh. A party that can rewrite the
  database as an administrator is the threat this window concerns; the application's own role cannot.
- With the flag off nothing changes: full replay per request and a full check every 30 seconds.
- The original replay is kept in the tests as a frozen reference, and every test that records events
  checks the reducers, the snapshots and the incremental cache against it.
- Cost per request is bounded by the size of the replayed state rather than the length of the history.
  Where state is itself proportional to history (messages), the saving is the avoided database read and
  decoding. A benchmark against an in-memory event source (no database round trip, so the real
  saving is larger) measures about 18 times less time per request at 50,000 events; see
  `projection_bench_test.go`.
- Replay rules still live in `cmd/server`; moving them next to the model in `internal/room` is part of
  splitting that package.
