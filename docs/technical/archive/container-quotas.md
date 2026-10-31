# Managed container quotas

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Free accounts may reserve at most four managed containers. An isolated development
job reserves two slots (engine and tool); a legacy single-container job reserves
one. Different rooms, participants, connections and approved workers share this
ceiling through the account's controller event journal. The local worker also
refuses more than four slots as a second admission boundary.

## Use and behavior

There is no configuration switch, paid bypass or role override. Paid limits are
not yet established, so the current guard applies to all managed registrations.
This mandatory resource safeguard does not need an optional feature flag.

For example, two unresolved development pairs fill the account. A third pair or
another single-container turn is refused before its allocation is committed or
any provider process is launched. Complete the existing work and prove shutdown,
or investigate unresolved jobs using the owner recovery workflow. Do not delete
journals or infer cleanup from a stopped UI, elapsed deadline or missing process.
Existing work above the limit is not automatically killed or erased; new
allocation remains refused until verified recovery restores capacity.

Disabling optional background-work or job-diagnostics features does not disable
quota enforcement, release reservations or delete history. Authentication,
current human access, company binding, repository permission, audit integrity,
daily-turn budgets and worker isolation remain independently required.

## Durable admission contract

`admitManagedContainerQuota` runs during mandatory managed-job registration after
identity validation and before the allocation event append. It counts the full
physical size of every unresolved job in the company, including allocation
intent, submitted work, expired jobs and unconfirmed shutdowns. Legacy jobs with
missing company attribution conservatively consume capacity. Malformed evidence,
a mismatched map identity, duplicate job or missing account binding refuses
admission rather than treating unknown state as free capacity.

Only a valid durable `shutdown_confirmed` record releases the controller's
reservation. Worker cleanup proof and approval/retention safeguards are not
replaced by this accounting function. Partial cleanup remains unresolved. The
controller never releases capacity by merely polling inventory or reaching a
provider deadline.

The allocation append uses the loaded state's event count as an expected
revision. The existing transactional audit-chain advisory lock checks that count
before inserting: two controllers sharing an account journal cannot both commit
the last slot from stale state. A conflict refuses launch; it is not silently
retried. The process-local lock serializes registrations in one controller but is
not the cross-controller correctness boundary.

## Scope and verification limits

Workers must use the same authoritative controller account journal; unrelated
self-hosted installations with separate databases cannot share a quota merely
because they use similar company names. This is container reservation accounting,
not a complete hosted entitlement or tenant-isolation service. Unknown remote
containers must still block worker admission and be reconciled.

The account's hard aggregate 5 GB storage ceiling is still pending. Internal
[storage accounting](storage-accounting.md) now journals bounded capacity
reservations independently of container slots. Complete physical inventory,
worker-enforced quotas and allocation/disposal integration remain required; a
four-container guard does not establish disk quota enforcement. Production
managed development remains gated pending the rest of P0.

Automated tests mock all persistence and exercise mixed paired/single jobs,
different workers, expired reservations, proof-based release, restart, disabled
optional flags, concurrent registrations and stale-controller transaction
conflicts. No database, real Docker workload or provider is used. Live
multi-controller/worker acceptance remains unverified.

The internal [account probe ledger](account-probe-capacity.md) now durably counts unresolved workspace helpers alongside managed jobs across workers; live controller transport and physical evidence remain required.
