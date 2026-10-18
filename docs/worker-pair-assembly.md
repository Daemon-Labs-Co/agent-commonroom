# Worker pair assembly

The admitted workspace-aware supervisor can now assemble its exact audited tool
session with a separate reviewed engine launcher and approved snapshot stream.
This connects the existing tool runtime, durable pair ordering and bootstrap
readiness. `RunCodex` uses this assembly for its owned internal lifecycle. Public
development launch remains gated; synthetic assembly tests do not establish live
image or provider compatibility.

## Ownership and startup

Trusted worker code calls `supervisor.AssemblePair(session, engine, bundle)` once,
then passes the returned adapter to `supervisor.Start`. The session must match
the supervisor's complete registered scope: job, participant, connection,
provider, room, repository and deadline. Legacy, cancelled and already assembled
supervisors are refused. A refused assembly leaves bundle ownership with its
caller; an accepted assembly owns the bundle until bootstrap or cancellation.
Never supply a model-selected host path, image or executable.

The adapter routes the tool role through the session's original
`ToolContainerRuntime`. The dispatcher therefore uses the same authority, plan,
workspace readiness, container identity and lifetime that bootstrap verified.
`ToolSession.Scope()` returns only a copy of immutable nonsecret registration
metadata; the raw capability remains privately delivered once.

The supervisor still owns the sequence:

1. Journal and prepare a fresh workspace volume.
2. Journal creation intent, create and verify each stopped role container.
3. Bind both immutable IDs durably and verify aggregate resource limits.
4. Start the verified idle tool container.
5. Consume the approved bounded bundle and validate its bootstrap receipt.
6. Start the separately verified engine by its bound immutable ID.

Adapter role/job/repository checks reject mismatched routing. Unknown IDs cannot
be inspected, started or bootstrapped through it. Creation, start and bootstrap
attempts are single-use, including uncertain failures. A failed or mismatched
bootstrap never satisfies engine readiness. The tool must have successfully
started before bootstrap, and engine/tool verification must precede tool start.
All calls remain bound to the original supervisor lifetime and deadline even if
the caller supplies a fresh context.

On supervisor cancellation an unconsumed bundle closes exactly once and the
shared session is revoked. `supervisor.Close` retains its existing lease fencing
through in-flight startup. Closing source/audit handles is not proof that either
container or workspace has been removed: the existing explicit stop/reconcile
and retention/discard requirements remain mandatory. Do not release quota from
an adapter error alone.

## Engine trust boundary and status

`workerEngineLaunch` is an internal trusted seam. The Codex implementation creates
fresh stopped containers, verifies the immutable image and mount/network policy,
starts only by immutable ID and honors cancellation. Credentials arrive through
private stdin rather than a host-file mount. Source code could replace it with host execution, but
that would violate the mandatory isolation boundary. There is no automatic host
or direct-SSH provider fallback.

Automated tests use injected synthetic role runtimes and temporary journals to
exercise startup routing, ordering, one-shot failures and cancellation. A real
`ToolSession` constructor test confirms exact runtime sharing and ownership
cleanup. Existing toolbroker tests separately cover full tool inspection,
bootstrap and audited dispatch. Combined `RunCodex` tests exercise the complete
internal path with mocked Docker and synthetic protocol peers; no live provider
run has been verified.

Reviewed release images, public transport/startup coordination, live refreshed-token
persistence, bounded commands/artifacts, retention/discard workflow, hard disk
quotas, provider coverage/upgrades/idle policy and live image/SSH/rootless/provider
acceptance remain P0 work. Current turns end after work; no idle session pool has
yet been introduced. The owner-configurable idle interval (default five minutes) must
be wired at the production session boundary without treating active work as idle.

This assembly is mandatory internal isolation infrastructure, with no product
flag or permission bypass. Existing authenticated human/delegation, account quota,
repository and audit checks are still required before production registration.

The supervisor-owned `AcceptToolSession` attachment now watches the private link
for idle EOF and cancels the original lifetime on loss. See
[audit session supervision](worker-audit-transport.md#session-attachment-and-disconnect-supervision).
Cancellation still requires explicit fenced container cleanup before shutdown
or quota release can be confirmed.

See [paired turn shutdown](worker-pair-shutdown.md) for unconditional
reconciliation, active-turn lease fencing and workspace retention.

[Codex callback driver](codex-callback-driver.md) documents the bounded
app-server protocol, private broker binding and remaining engine acceptance gates.

See [reviewed engine-container runtime](engine-container-runtime.md) for
fresh creation, reported policy verification and remaining production integration.
