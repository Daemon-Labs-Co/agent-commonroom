# Recover uncertain worker credentials

[Connection sign-in](connection-sign-in.md) · [Release readiness](release-readiness.md)

A provider can refresh or rotate subscription tokens during an engine turn. If the
worker cannot establish durable custody of the resulting credential, the previous
saved file may be stale even when it is valid JSON and all containers are gone.
CommonRoom blocks that connection until credentials are explicitly recovered.

## What you see

In **Connections**, the sign-in control explains that credential recovery is
required. The sign-in dialog asks you to finish or reconcile prior work before
signing in again. Status reports `credentialPresent:false` and
`recoveryRequired:true`; it never returns credential contents. Other connections
remain independent, including connections using the same provider account.

## Recover the connection

1. Stop or finish existing work. Reconcile uncertain engine/tool creation until
   both execution roles are confirmed absent or never started. A crashed process,
   cancelled request or released authentication lock is insufficient.
2. Choose **Sign in** for the exact affected connection and complete the isolated
   provider flow. If worker records remain unresolved, authentication changes are
   refused; inspect job diagnostics before retrying.
3. Wait for **Connection sign-in saved**, then choose **Check sign-in** and
   **Test connection**. Provider validity still requires that live connection test.

Source workspaces can remain retained while execution roles are fenced. Recovery
must not discard source or alter chat/audit history. Explicit **Sign out** removes
this connection's stored credential and clears recovery only after durable local
revocation; provider-wide token revocation is a separate operation. Failed login
or export leaves the connection blocked rather than trusting its old file.

Do not delete the recovery marker, copy another connection's login or relax file
permissions to make a status check pass. Complete a new authorized sign-in.

## Worker contract

The marker is `.credential-recovery-required` inside the existing private
`authDir/<connection hash>/<provider>/` directory. It contains only a version and
recovery boolean, never tokens, transcripts or model output. Its existence is the
blocking signal; corrupt files, symlinks, directories and inspection failures are
also treated as unresolved custody.

The separated engine lease acquires the connection lock and checks both the job
journal and recovery state before reading its private credential snapshot. Before
writing any authentication frame bytes, it atomically publishes the marker,
fsyncs its file and parent directory, and consumes the one-shot provisioning
attempt. Persistence failure sends no credential bytes. Partial transport failure
keeps the marker. Releasing the lease after both-role fencing clears memory and the lock. Token
uncertainty clears only if a valid final snapshot is durably saved; absence alone
still does not establish custody.

Status and the existing isolated execution preparation path also refuse a marked
credential. Status is an observation rather than execution authority: later
launches must perform their own checks. Connection/provider paths scope the
marker; no global reset exists.

Sign-in and sign-out acquire the same connection lock and scan bounded private
worker records. Any corrupt/unknown record fails closed. For matching records,
legacy jobs need confirmed shutdown; paired jobs need both roles fenced. Retained
source alone does not prevent reauthentication. A newly created unprovisioned job
cannot obtain the credential lock until authentication finishes.

A login marks uncertainty before starting the provider process, since a failed or
interrupted login may itself rotate tokens. Successful private export is validated,
atomically written and synced before marker removal and parent-directory sync.
Persistence failures are reported without raw backend errors. A failure syncing
the marker's removal is still reported as unconfirmed; the credential replacement
was already durable, so the new credential can be visible while acknowledgement
is uncertain. Retrying sign-in/sign-out must respect the same lock and journal
checks. Never acknowledge a failed persistence operation as success.

## Controls and security limits

This custody gate is mandatory security infrastructure. It has no feature flag,
owner override, paid bypass or host-execution fallback. Existing
`connectionSignIn` controls whether the owner may use the authenticated, authorized
and audited sign-in/status/sign-out API. Disabling that feature does not clear the
marker or permit execution with uncertain credentials. Re-enable it to perform
owner-driven recovery.

The configured private authentication root and its ancestry, worker filesystem,
Docker daemon and service account remain trusted. A host administrator or source
fork can remove the marker, forge journal evidence or bypass checks; doing so
invalidates the custody guarantee. Filesystem durability depends on the operating
system, filesystem and hardware honoring sync operations. This mechanism is not
provider-wide revocation or administrator-resistant evidence storage.

## Verification and remaining P0 work

Tests use synthetic credentials, temporary private files, injected persistence
failures and mocked providers/database interactions. They cover marker-before-byte
ordering, partial/failed provisioning, crash-style reacquisition, uncertain-role
blocking, retained-source recovery, connection isolation, persistence ordering,
status/UI handling and combined callback shutdown paths. No real provider or
Docker workload is launched.

The internal separated-engine turn now captures a final private credential after
its reviewed provider child closes, then persists it only after both containers
are fenced and matching absence evidence is read back from the durable journal.
A successful commit clears recovery without a new sign-in. See the
[private custody protocol](engine-credential-custody.md) for ordering and limits.

Failed/interrupted turns or invalid/unavailable exports retain recovery. An
unpublished absence record keeps the lease and snapshot for later reconciliation.
Persistence failure releases only safely fenced execution custody, preserves the
recovery requirement where the replacement was not durably committed, and never
acknowledges success. The engine image must be rebuilt, reviewed and pinned for
this new export protocol; old images fail closed rather than falling back.

Four credential acceptance drivers now exercise real custody operations with
mocked external boundaries. Live provider refresh, cross-process crash behavior
and deployment acceptance remain outstanding. This does not enable public
repository-development dispatch or complete all P0 gates.
