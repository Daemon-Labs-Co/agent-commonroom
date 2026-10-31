# Pinned controller artifact read sessions

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


This internal adapter supplies the receiving-side release guard for the
[private sealed-object read channel](sealed-object-read-channel.md). It pins one
human principal, worker, exact sealed receipt and session lifetime to the existing
controller custody journal. It is mandatory authorization infrastructure, not a
public download feature or optional isolation switch.

## Setup and identity

Already authenticated private setup calls `newControllerArtifactReadSession` with
the original human context, approved worker alias, exact receipt and mandatory
`artifactReleaseEvidence`. Message fields and later operation contexts cannot
select another principal, worker, job, output, revision, object, size or digest.
The worker must still match the retained managed job on each check; knowing a
receipt alone grants no access.

The independent release witness must authenticate the actual approved isolated
reader/root, sealed object identity, physical quota/account state and current
export/retention policy. It must cooperate with cancellation and must not reenter
the controller lock or waiting read stream. No deployed/default witness is supplied.
Container labels, directory modes, peer metadata or synthetic fixtures do not prove
physical quota or authenticated custody. Source forks can substitute witnesses or
private-session setup and invalidate these guarantees.

## Read and audit behavior

Pass the adapter as `objectcustody.ReadReleaseGuard`. `AuthorizeRead` requires exact
scope/receipt equality and uses the pinned human context. The controller journal
checks current human room access, incident/emergency/integrity stops, historical
execution/definition/output binding and exact sealed receipt/reservation linkage.
A release witness then checks current isolated-reader/export policy. After that
callback, the adapter reloads current access and linkage and rejects changed audit
revisions. Callback latency cannot silently preserve revoked access.

After the read channel verifies the actual complete bytes, `AcknowledgeRead` repeats
current authorization and appends the release acknowledgment under the journal's
audit revision fence. No bytes leave the read channel on failure. Journal persistence
failures retain storage and trip the existing controller integrity latch. An
acknowledgment means bytes were authorized for downstream release; it is not evidence
of user receipt, task completion or human acceptance.

Existing policy intentionally permits human reads of retained completed objects
when task-definition editing and unrelated optional features are disabled. Turning
those flags off blocks new source work; it does not erase retained data. Current
room/export permission, emergency stops, audit integrity and retention safeguards
remain mandatory for all reads. No new user-visible feature or feature flag is added.

Session cancellation propagates into active calls. `StopReadAuthority` immediately
and nonblockingly cancels this session; a fresh operation context cannot revive it.
Read-transfer uncertainty stops the affected receiving session, without inventing a
storage refund or global persistence failure. Actual journal persistence faults still
trip the global integrity stop. Recovery requires fresh authenticated setup.

The adapter explicitly refuses source validation, quota establishment and sealing:
passing it as a general custody guard cannot turn a read session into a writer.
It grants no file path, host execution, public endpoint or authentication fallback.

## Verification and remaining integration

Tests join actual temporary synthetic sealed custody objects, the mocked controller
database/journal and the private read channel. Source and receiving supervisors each
record their required authorization-to-release acknowledgment. Tests also cover
principal pinning, worker/receipt/scope substitution, missing/agent principal, denied
witness, canceled/permanently stopped sessions, flags-off retained reads, stale audit
snapshots, callback-time revocation, incident/integrity stops and persistence failure.
No real database, Docker daemon, provider, credential or private message is used.

Deployed authenticated transport, isolated reader/helper orchestration, actual
release/quota/account witnesses and physical volume provisioning remain incomplete.
End-to-end task artifact delivery/human review, reviewed rootless image/latency/live
acceptance and unknown-ID incident handling remain required. Required file-task
dispatch stays blocked. Synthetic byte transfer with real controller code is not
production/live P0 acceptance; there is no supported uncontained fallback.

See [controller custody journal](artifact-custody-journal.md) and
[mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).
