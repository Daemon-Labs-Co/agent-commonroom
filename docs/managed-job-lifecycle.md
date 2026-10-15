# Managed job lifecycle module

Managed job dispatch records intent before preparing isolated worker execution,
records submission before returning a command, and requires verified shutdown
before recording completion. The managed job module owns this ordering in
`cmd/server/managed_jobs.go`. The SSH adapter in `remote_worker.go` retains
endpoint approval and development repository policy checks.

## Use and configuration

Users continue to run managed turns and use owner-authorized job recovery as
before. No additional setup or permission is introduced. Existing remote-worker
configuration, authenticated membership checks, initiating-human attribution,
daily limits, and repository approval remain enforced. Turning optional job
diagnostics off does not disable recovery or lifecycle recording.

This is an internal refactor of mandatory execution safeguards, with no product
on/off flag. It enables no new execution path. Paired startup remains unavailable;
the legacy dispatch interface rejects paired registrations instead of downgrading
them. See [worker pairs](worker-pairs.md) for paired recovery and remaining work.

## Technical contract

`dispatchManagedWorker` is the dispatch seam. The caller supplies the approved
worker and constrained run request; the module supplies the durable job identity,
original issued time and deadline. It validates and encodes the request, records
registration and submission, and returns the transport command plus idempotent
cleanup. The caller must invoke cleanup after command completion **or command
start failure**. This retains the existing command ownership interface without
moving process execution into the lifecycle module.

The internal `managedWorkerTransport` adapter seam varies SSH transport in
production and synthetic transport in tests. It cannot authorize execution:
registration still checks current membership and permissions, and the worker
independently validates admission, request scope and container isolation.
Changing source to bypass these checks would invalidate the security contract;
this module does not make direct host or provider execution safe.

Cleanup cancels background monitoring and requests scoped shutdown with a fresh
35-second deadline even if the original turn is cancelled. Failed shutdown records
`shutdown_unconfirmed`, preserving the application stop. A shutdown audit failure
also reports uncertainty; successful transport alone is never reported as durable
completion. Submission audit failure triggers the same cleanup before dispatch
returns an error. Cleanup never restarts or replays work.

`reconcileManagedJob` keeps explicit recovery's transport evidence and durable
status update together. The recovery handler still authenticates and authorizes
the human, checks current worker approval, and validates reconciliation inventory
before crossing that seam. Recorded immutable job identity determines fencing;
request parameters cannot substitute another repository or participant.

## Verification and limits

Synthetic tests cross the dispatch interface to cover registration rejection,
submission audit failure, successful cancellation after turn cancellation,
shutdown failure, terminal audit failure, bound fencing, original timing and
idempotent cleanup. Existing mocked registration and HTTP recovery tests cover
permission and provenance behavior. No database, provider agent or Docker workload
is launched. Live SSH, provider and container verification remain separate P0 work.

The module increases depth by concentrating lifecycle choreography. This gives
callers leverage and maintainers locality: protocol submission and shutdown status
ordering change in one place. Deleting the module would spread those requirements
back into dispatch and recovery callers.
