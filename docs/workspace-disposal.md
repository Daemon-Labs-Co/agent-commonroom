# Owner-reviewed workspace disposal

The owner can propose permanent deletion of a stopped job's retained workspace,
review its exact volume identity for at least one minute, then explicitly approve
that one deletion. The feature is **off by default**. It never resumes agents or
automatically retries a deletion. Production repository dispatch remains gated.

## Enable or disable

In **Settings → Security → Workspace disposal**, the enabled human owner can
change the `workspaceDisposal` feature flag. It defaults to false and takes effect
immediately without restarting. The server checks it when proposing and claiming
an approval; changing a button or sending a direct API request cannot bypass it.
Paid status and organization rank confer no additional permission.

Disabling prevents new proposals and approvals. It preserves history and pending
proposals, which the owner can still revoke. A deletion already claimed durably
may finish under its bounded cleanup context: switching off cannot undo an admitted
irreversible action. Authentication, exact volume binding, audit persistence,
retention protections and mandatory container isolation have no bypass switch.

The current UI lives in **Worker jobs**, so the optional **Worker job diagnostics**
view must also be enabled to use these controls. Disabling that view does not grant
deletion authority or disable the server's approval checks.

## Review and approve

1. Stop the job using ordinary cancellation/reconciliation. It must have a
   complete version-4 workspace identity, both container roles stopped and no
   supervisor holding its creator lease. Ordinary recovery retains source.
2. In **Worker jobs**, choose **Review workspace disposal** for that job, then
   submit the proposal. This records an intention; it does not delete data.
3. Review the job, volume name, creation nonce and timestamp. The action destroys
   **all contents of that workspace, including unfinished edits**. It does not
   delete room/audit history, provider credentials or separate artifacts.
4. Wait one full minute after the committed intention is observed by the controller. Open Worker jobs and
   refresh status to reveal approval when eligible. Approval is never automatic.
5. Choose **Approve permanent workspace deletion**, type the exact job identity,
   and submit. Alternatively, revoke the pending proposal to retain the workspace.

A proposal expires after one hour. An expired proposal cannot authorize deletion;
create a fresh proposal and review it for another minute if still appropriate.
The current free implementation offers no configurable delay or autonomous agent
approval. Volume identity describes the entire stopped workspace, not a claim
that its contents were exported or attested to a Git revision. Preserve any work
you need before explicitly approving its complete destruction.

A successful response means removal/absence proof and its controller outcome were
recorded. Container-slot release and any exactly bound workspace reservation release
commit in the same transaction. A proposal with `storageReservationId` identifies
the immutable charge being reviewed. The reservation must already bind this job,
company, room, worker, repository and exact volume name, nonce and creation time.
Legacy unbound reservations, bundles and artifacts remain charged. No binding is
guessed from a job ID, reconstructed after removal or accepted from the browser.
Do not infer a storage refund from an empty Docker volume list, edit the ledger
or delete history to regain capacity.

## Uncertainty and recovery

The controller durably claims the proposal once before sending the worker action.
Concurrent clicks or controllers cannot both claim it. Request disconnection or
later feature changes do not reopen the claim. Cleanup has its own actor-bound
40-second context and the worker keeps its existing stricter bounds.

A worker failure, panic, uncertain transport reply or failed outcome audit leaves
a durable claimed/unconfirmed action. The UI says not to repeat it. It cannot be
approved again, and a new proposal for the same unresolved action is refused.
Normal worker diagnostics/reconciliation remain available, but do not erase the
claim or authorize another deletion. A dedicated claim-resolution or retrospective
credit workflow is not implemented yet; administrator investigation may be necessary.
Never fabricate success, reset the journal or resend a destructive request to
resolve uncertainty.

## HTTP and audit contract

`/api/workspace-disposal` requires an authenticated enabled human owner.
POST also requires the current CSRF token. GET returns the current flag, server
time and at most 128 proposal records, including `reviewReadyAt`; unresolved claims take priority. Each controller starts a conservative full-minute wait when it first observes a committed pending proposal. A restart or a newly contacted controller can therefore require another minute. Refresh the review to see its effective ready time. The claim records `reviewObservedAt`, and replay verifies the full delay. Proposal
records are excluded from the general public state projection and generic event
commands cannot manufacture them.

| POST action | Fields | Behavior |
| --- | --- | --- |
| `propose` | `jobId` | Probe approved worker, validate stopped retained identity and publish intention. |
| `approve` | `proposalId`, `confirm` equal to the exact job ID | Recheck current owner, flag, job, volume, worker approval, review delay and expiry; claim once, then execute. |
| `revoke` | `proposalId` | Revoke a pending intention, including while the feature is off; no worker effect. |

Only the exact action fields are accepted. Duplicate keys, aliases, nulls,
unknown fields and oversized payloads are refused. Job/model arguments cannot
select a path, volume, Docker flag or worker. The server obtains the complete
identity from the approved authenticated worker snapshot and binds it through
[the worker disposal contract](workspace-disposal-binding.md).

The controller journals proposal, claim, revocation and verified/unconfirmed
outcome with human attribution. Append-only replay rejects changed scope/identity,
invalid timing, skipped approval and terminal-state rewrites. Expected journal
revisions are checked under the transactional audit-chain lock. No worker action
is sent unless the claim commits. Outcome and container-slot release commit
atomically with `storage.released` when the reviewed reservation is exactly bound.
The release retains the original reservation and references the successful disposal;
replay refuses a duplicate or foreign refund. Raw worker errors, source contents, commands and credentials do not
belong in these records.

The worker transport remains a trusted dependency. Replacing it in source with
an adapter that reports fabricated success invalidates the proof contract. No
model/plugin can inject it. Older workers that cannot understand exact identity
binding must be upgraded; there is no boolean-only compatibility fallback.

## Verification and remaining limits

Automated tests mock every database, SSH and Docker interaction. They exercise
actual HTTP authentication/CSRF/feature checks, full review timing after slow
probes and slow audit publication, expiry/revocation, identity/access changes, restart/replay, concurrent
controllers, atomic outcome failure, uncertainty, panic and disconnected requests.
Synthetic browser tests cover escaping, separate proposal/approval submissions,
server-time gating, default-off switches and absence of retry controls.

Three P0 refusal drivers cover missing approval, unauthorized humans and the
wrong volume identity. The approved-disposal driver additionally exercises the
owner endpoint, worker journal and mocked Docker removal/absence, exact receipt
validation, atomic capacity release and preserved history. Five further catalog
drivers verify stale/revoked approval, failed removal, unavailable absence proof
after removal, cancellation that retains source and refused history deletion.
Uncertain cleanup remains charged and cannot be retried. All nine disposal catalog
scenarios now have hermetic coverage. No live human-approved production disposal,
provider round trip or browser/database deployment was exercised by this change.
The earlier [tool-only live evidence](live-tool-acceptance.md) does not establish
this owner flow's deployment acceptance. Physical quotas, production allocation binding,
artifact delivery and full [release readiness](release-readiness.md) remain open.
