# Shared audited tool session

The worker's internal tool session now constructs its runtime and dispatcher from
one authority and one plan. Successful bootstrap opens readiness on that exact
plan, bound to the original job lifetime and immutable container ID. This removes
the previous assembly gap where bootstrap and dispatcher could hold independent
plans and never share readiness.

There is no user setting, product toggle or enabled development endpoint. This
is mandatory free/core infrastructure. Existing worker-job context, original
scope and controller audit checks still apply; this assembly cannot authenticate
a peer or turn missing human provenance into owner permissions.

`newWorkerToolSession` first checks the worker job context. The trusted library
constructor accepts an explicit approved Docker adapter and private authenticated
audit channel, creates the per-job authority/capability, tool plan/runtime,
AuditClient and dispatcher, and binds all of them to the original deadline.
The capability is returned once for protected supervisor-to-engine delivery; it
is not a serialized session field or an audit/history value. The worker must
authenticate and separate the audit channel from model stdout before construction.

The supervisor journals workspace intent, prepares the fresh volume, records its
identity, creates/binds the tool container, verifies aggregate pair resources,
starts the idle tool and bootstraps source through the session runtime. It must
complete those steps and the engine-side requirements before enabling turns.
Dispatcher file actions remain refused until bootstrap succeeds. Every action
still requires durable controller intent and outcome acknowledgements.

Each action inspects the original bootstrap container ID, verifies that returned
metadata reports that same ID, checks full current volume/runtime policy, and
executes the fixed helper only by that immutable ID. A replacement with matching
name, labels and settings cannot inherit readiness. Runtime inspection and output
validation do not establish independent image/daemon attestation.

Close revokes authority, cancels runtime lifetime and closes blocked audit I/O.
It is idempotent and fresh caller contexts cannot revive it. Close is not proof
that processes or volumes are gone: paired fencing and released creator ownership
remain mandatory. Failed construction closes the owned audit link after initial
input validation. There is no no-op audit or host-execution fallback.

Tests use real framed local audit channels and mocked Docker responses across
allocation, creation, bootstrap and dispatch. They verify intent-before-execution,
failed audit/receipt refusal, suppression of unaudited output, cancellation,
capability serialization exclusion and replacement-ID refusal. No live provider,
Docker workload or database is used. The private authenticated SSH audit transport,
reviewed engine driver/images and actual production pair adapter remain P0 work.

See [worker pairs](worker-pairs.md), [workspace volumes](workspace-volumes.md) and
[bootstrap](workspace-bootstrap.md). A source fork can replace trusted assembly,
but then invalidates isolation and durable-audit guarantees.

See the [private worker audit transport](worker-audit-transport.md) for
SSH separation, single-use socket ownership and remaining production integration.

See [worker pair assembly](worker-pair-assembly.md) for exact session/runtime
binding, snapshot ownership, startup ordering and remaining engine integration.

The supervisor-owned `AcceptToolSession` attachment now watches the private link
for idle EOF and cancels the original lifetime on loss. See
[audit session supervision](worker-audit-transport.md#session-attachment-and-disconnect-supervision).
Cancellation still requires explicit fenced container cleanup before shutdown
or quota release can be confirmed.
