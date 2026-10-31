# Storage accounting and physical enforcement

## Input

Complete independently fenced account inventory, finite workspace/bundle/artifact reservations, current company identity and independently provisioned root/quota identities.

## Output

Admitted reservation, observed quota evidence, capacity/identity refusal or retained uncertainty.

## Functionality

The free-account accounting ceiling is5,000,000,000 decimal bytes. Unknown inventory blocks growth. Existing over-limit data remains retained. Native projectquota observation supports XFS/Linuxamd64 and does not provision quotas. Inspected Jerry storage is ext4; deployed XFS enforcement and mapping are unverified.

## Properties and state changes

Reservations increase charged capacity before allocation. Uncertain/retained data stays charged. Release requires independent verified disposal; low observed usage or container exit alone does not refund capacity.

Source: [storage_quota.go](../../cmd/server/storage_quota.go) · [worker_storage_check.go](../../cmd/server/worker_storage_check.go) · [storage_disposal.go](../../cmd/server/storage_disposal.go).
