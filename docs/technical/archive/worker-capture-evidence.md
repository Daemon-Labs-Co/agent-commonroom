# Worker-local physical capture evidence

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


`workspaceCapturePhysicalEvidence` implements the controller's independent capture
observation interface using the actual worker creator-ownership scope. It performs
metadata inspection only: no container launch/removal, source/object reads, quota
refund or storage deletion is available through it.

## Trusted construction

Inside the live `workerWorkspaceObservation` creator callback, trusted setup supplies
the authenticated worker identity, reviewed capture image, exact finite request,
retained private custody volume and mandatory `capturePhysicalBindingGuard`.
Construction requires active ownership/lifetime and a real held creator lease, a
valid reconciled engine/tool record and an exact source volume/repository/job match.
It detaches request members and retained record pointers from caller mutation.

The binding guard remains an explicit, non-optional integration requirement. Each
`BoundCapture` call must independently verify durable local/controller capture binding,
the reviewed image, private object volume, account charges and physically enforced
source/object quotas. It must also bind the supplied mount proof to the exact helper's
exclusively owned authenticated private channel. It returns only the already bound
immutable ID. Empty or discovered-by-name IDs are never admitted. No production
binding/quota guard or permissive default is supplied by this module.

## Per-observation checks

`ValidateCapture` rejects expired/canceled ownership, a substituted record/request/
proof or a different worker/job/execution/definition/company/room binding. It then:

1. Rereads the actual retained workspace record and cancellation fence.
2. Requires the current independent binding/quota/private-proof witness.
3. Runs the capture plan's two independent running-container, both-volume and full
   reference/catalog admission passes against the exact bound ID.
4. Rereads the durable retained record/fences after physical inspection.
5. Requires the binding/quota witness again and refuses a changed ID or revoked
   authority. Callback lifetime/context must still be live.

The controller independently handles current human/task/flag policy and durable audit
acknowledgment before source access. This observer does not replace those permissions.
Evidence callbacks must honor context and avoid reentering controller locks or the
same capture stream. No optional flag can bypass this mandatory isolation boundary;
no host/SSH fallback exists. A source fork with dishonest dependencies defeats these
claims.

## Verification and gaps

Tests use actual synthetic creator lock files and retained records, with mocked
Docker, durable binding, quota and private-channel proof witnesses. They cover healthy
two-pass checks, quota denial before inspection, post-inspection revocation, missing/
foreign/changed IDs, writable source, extra references, replaced volumes, changed
retained records, forged mount proof, missing ownership and authority expiry. The
observer attempts no mutating Docker operation. No provider/database/container or
private fixture is used.

This is not a live rootless/image acceptance result. The mandatory deployed binding/
quota/private-proof witness, durable lifecycle/account/creator composition, physical
quota provisioning, reviewed image and real latency acceptance, unknown-ID incident
handling and authenticated delivery/human review remain incomplete. Required file-task
dispatch stays blocked. Metadata labels/counters are never proof of a disk quota.
