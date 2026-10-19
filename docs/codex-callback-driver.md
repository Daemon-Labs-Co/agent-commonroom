# Reviewed Codex callback driver

`internal/codexbridge.Run` implements a bounded client for the reviewed Codex
app-server protocol pinned to `rust-v0.162.1`, commit
`092d3acd6bec3e3a14bdc7e7a2810ab628ab759d`. This uses the local generated schemas
from that reviewed source, not an assumption about the newest installed CLI.
The driver launches no executable and cannot authenticate a provider or container.
It must run over the reviewed isolated engine stream with a fresh controlled home.

## Turn contract

One invocation initializes the app-server client, starts one ephemeral thread,
then starts one model turn. It provides exactly one nondeferred dynamic function,
`commonroom_tool`, with read/write workspace arguments by default. An explicitly
admitted [isolated command session](isolated-commands.md) also advertises its
fixed terminal Go profiles. The thread requests
read-only sandbox policy, no native approval grants, fixed engine cwd and no
provider/model fallback. The returned thread policy is checked before the turn.
No resume, fork, configuration mutation, external MCP, native shell/patch or
separately billed API fallback is provided.

Callbacks must name exactly that function, have no namespace, and match the
actual returned thread and turn IDs. Both RPC request IDs and provider call IDs
are single-use. Callbacks arriving before the start response are queued within a
small bound and are not dispatched until response identities are known. Unknown
requests, native approval requests, malformed/duplicate JSON, identity mismatch,
failed audit handling, interrupted turns and uncertain EOF abort without replay.

The supervisor's `CodexTurn` accepts only its originally assembled ToolSession
after pair startup, for the Codex provider, and permits one invocation. It binds
the model's operation/path/content to the trusted scope, call ID and privately
held capability before calling the session Dispatcher. Model arguments containing
capability, scope, Docker options, arbitrary commands, unexpected fields, aliases
or duplicate fields are refused. The capability never appears in the model's
schema, app-server request or thread input.

Only final agent text is returned. Reasoning notifications/content are not
exported or stored, and model descriptions of checks are not promoted to trusted
execution evidence. Native execution item types are refused. Actual observed
build/test artifacts remain a separate incomplete P0 requirement.

## Bounds and failure handling

Prompt input is at most 64 KiB, each incoming frame at most 256 KiB, the stream at
most 8 MiB/2,048 frames, prebinding queue at most 16 messages, tool callbacks at
most 64, and final/tool result text at most 16 KiB. JSON nesting is bounded and
duplicate keys are refused before typed decoding. Broker permissions, action
budgets and path containment independently apply and may be stricter. Current
file helper implements read/write only. Opt-in terminal commands use a separate
trusted worker capture path; patch and arbitrary commands are still refused.

The owned engine stream closes on success, protocol refusal or cancellation.
Its Close must interrupt reads/writes. A protocol failure cancels the original
supervisor lifetime. The outer Run/Shutdown lifecycle must still fence both role
containers; stream closure is not container shutdown proof. Source changes could
replace the stream or handler with a host implementation, violating mandatory
isolation. There is no bypass flag or automatic uncontained fallback.

## Verification and remaining work

Synthetic local-stream tests cover initialization, early notification ordering,
thread/turn correlation, callbacks, duplicate call/request IDs, native requests,
completion-policy failures, output bounds, malformed frames and cancellation.
Broker-binding tests cover private authority injection and reject model-supplied
scope/capability. Database/Docker/provider operations are not used by these tests.

This driver has not yet executed against the actual reviewed Linux binary or a
live provider account. The upstream binary's real notification ordering, supported
configuration/customization suppression and the full callback loop still need
acceptance using dedicated synthetic fixtures/accounts. Unexpected notifications
currently fail closed; compatibility is not claimed from schema inspection alone.

The internal engine launcher, process attachment, controlled home, private gateway
and stdin credential provisioning are implemented. Credentials are not host-mounted.
The driver now supports a trusted post-completion finalizer before attachment
closure. Worker custody uses it for private final-credential export; failure or
cancellation withholds the model result. It runs only after valid completion and
has no model-selected callback. See [credential custody](engine-credential-custody.md).
Reviewed release images, live refresh acceptance and the public production worker
entry point remain required.
Claude, Gemini, GitHub Copilot and llama.cpp adapters remain separate P0 work.
Production development admission stays blocked; installing a newer stock CLI or
substituting `codex exec` does not satisfy this reviewed engine contract.

[Codex engine bootstrap](codex-engine-bootstrap.md) describes the fixed
subscription-only process, pinned recipe and outstanding live integration.
