# Cancellation, cleanup and recovery

## Input

Exact durable job/helper identity, original retained binding and fresh current cleanup permission. Unknown IDs are not cleanup authority.

## Output

Verified exact-ID absence with durable acknowledgment, or explicit uncertainty preserving history and capacity.

## Functionality

Stop requests cancel contexts; they do not certify remote termination. Known-ID reconciliation does not start a new provider, adopt by name, read/export objects, delete volumes or refund quota. Identical interrupted recovery checkpoints can be explicitly reauthorized where the journal permits. Integrity stops require separate verified recovery.

## Properties and state changes

Active contexts are revoked and tracked work joins before creator ownership ends. Listener closure/authority revocation finish before service shutdown completes. Recovery preserves existing audit/acknowledgment history and does not auto-accept results.

Source: [worker_capture_runtime.go](../../cmd/server/worker_capture_runtime.go) · [worker_object_read_runtime.go](../../cmd/server/worker_object_read_runtime.go).
