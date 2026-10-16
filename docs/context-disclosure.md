# Context disclosure controls (Free)

Agent prompts are an external disclosure to the provider used by that connection.
Room access alone does not authorize cross-room memory export.

## Owner controls

In a room, select **Context sharing**. Allow OpenAI/Codex, Claude, both, or neither.
Exclude earlier room messages, shared memory, or imported reference context as
needed. Stop affected agents before changing policy. The current request and
workflow input are still sent to an allowed provider. Profile instructions and
organization guidance are also included; do not put secrets in them.

In **Shared memory**, edit a revision and select **Human-only** to exclude it from
all agent prompts. Explicitly approve provider export for Application memory that
must cross rooms. Without approval, Application memory is omitted outside its
source room. Room memory is available within its room unless human-only or its
provider allowlist excludes the destination. Project memory currently stays in
its source room; a separate Project authorization model is not implemented.
Personal-room memory cannot be broadened by ordinary writes or exported across
rooms. Unknown providers or missing rooms/connections fail closed.

Disclosure changes reset native session bindings atomically with the policy or
memory revision; application chat history is preserved. Earlier room messages
may still be included unless excluded. Data already sent to a provider cannot
be recalled, and exported repository files use the separate development-workspace
approval boundary. These controls do not claim to prevent all prompt injection,
automatically detect secrets, or provide a provider-side deletion service.

## Configuration and permissions

Security flag `contextPolicyEditor` defaults **on**. Only the owner changes it.
Off hides the room editor and rejects `room.context_export.updated`; saved export
restrictions remain enforced. No restart is needed for policy changes; reload
after changing the editor flag. Memory export grants remain owner-only. Ordinary
room edits cannot reset disclosure policy. Disabling any UI flag cannot bypass
prompt filtering or mandatory container isolation.

Legacy native session bindings without the required disclosure-policy provenance
start fresh on their next managed turn.
Existing rooms allow supported OpenAI/Claude destinations and room-local context.
Existing cross-room Application memory has no explicit export approval and is
therefore excluded. Review required shared notes before running existing agents.

## Technical contract

`contextDestination` checks the participant's actual connection and room policy.
`memoryExportAllowed` enforces source room, human-only status, scope and explicit
provider grants before concatenation. `chatPrompt` separately filters message
room IDs even when given unscoped state. `workflowPrompt` checks destination
policy and room identity, and filters imported references. Prior workflow results
are pinned run reference material, not independent evidence of successful work.
References serialize as JSON with stable source IDs and an untrusted label.
Source labels cannot authorize tools, memory exports or directory access.

The event `room.context_export.updated` stores `{roomId,policy}` with authenticated
actor attribution. Memory revisions stamp `sourceActorId` server-side and contain
`agentExport` and `exportProviders`. Only supported, unique provider values are
accepted. Restrictions are enforced on the server before provider invocation.

Tests use synthetic references and mocked databases. They cover denied providers,
unknown destinations, foreign rooms, unsupported Project sharing, private origins,
human-only memory, explicit grants, both editor states and ordinary-edit bypasses.
Real provider deletion/retention and prevention of all model inference are not
claimed by these tests.
