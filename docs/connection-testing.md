# Testing connections

[Documentation home](README.md)

## Use the test action

Open **Connections**, then select **Test connection** beside any registered connection. The same action is available in the footer of **New connection** and **Edit connection**, with results shown in the setup sidebar. The dialog shows a dated checklist and a **Test again** button. Close it with the X in the top-right corner. Only the owner can run checks.

For an approved remote CLI connection, the checklist verifies:

1. The endpoint targets the configured isolated worker and remote transport is enabled.
2. SSH with verified host keys reaches the installed worker.
3. The worker's private configuration, rootless Docker, required CPU/memory/process limits and pinned image are available.
4. A private, valid JSON credential file exists for this connection and provider.
5. The isolated provider agent returns `COMMONROOM_CONNECTION_OK` to a fixed prompt, with tools disabled and no room context. The small test turn may consume provider quota.

**Connection working** requires an actual expected provider response. A healthy worker and a credential file alone never produce success. Provider authentication, availability and the isolated CLI path must work at the time of the check. This does not guarantee future availability or permission to execute arbitrary development tasks.

**Connection test failed** means the provider turn failed, timed out, returned an unexpected reply, or supplied an invalid verification result. Recheck sign-in and provider account access. Raw provider errors are not returned because they may contain sensitive data.

**Sign-in needed** means the runtime is ready but the connection has no supported private credential file. Use its **Sign in** button; see [isolated connection sign-in](connection-sign-in.md).

**Connection blocked** includes disabled remote transport, an unapproved SSH alias, direct host execution, worker configuration problems or an unavailable worker. Check the stage-specific instructions. No fallback to unrestricted execution is allowed.

**Adapter unavailable** applies to registered HTTP, WebSocket and MCP connections. Those execution adapters are not implemented. The test makes no network request to their URLs and does not misrepresent reachability as agent support. Local CLI entries currently require conversion to an approved Linux/rootless-Docker worker endpoint; this diagnostic does not launch a local CLI.

## Feature flag

**Settings → Connections → Connection testing** enables or disables this optional capability. The persisted flag is `connectionTesting`, enabled by default. When disabled, the button is hidden and the API rejects requests. It does not change authentication, sign-in, execution isolation or saved connections.

## API and implementation

`POST /api/connections/test` accepts `{"connectionId":"REGISTERED_ID"}` for a saved connection, or an optional `draft` object with `name`, `protocol`, `endpoint`, `provider`, and `location` from the editor with an authenticated owner session and CSRF token. Draft endpoints are tested only through the same approved-worker allowlist. New drafts and changed execution configurations receive setup-only checks; they cannot access credentials or send provider prompts. A registered connection ID is always resolved against stored state before allowing full verification. Unknown fields and trailing JSON are rejected; the body is limited to 4 KB. Missing connections return 404; disabled capability or insufficient permission returns 403; an integrity incident or agent stop returns 409.

Results contain `status`, `checks` (name, status, detail), `checkedAt`, and `providerResponseTested` (true only when a valid correlated verification outcome was received). A diagnostic failure is a successful diagnostic response with blocked/unsupported status, not a transport exception. Worker operations use `ready-v1` followed by connection-scoped `auth-status-v1` and `verify-v1`, sharing a forty-second controller deadline. The worker test turn has a twenty-second execution deadline, plus bounded container cleanup. SSH startup and output are bounded. Only the fixed diagnostic prompt is sent; no room history, arbitrary model command, credential contents or remote stderr are returned to the client. Failure messages are mapped to safe instructions.

`connection_diagnostics.go` owns the handler, flag policy, staged diagnostic model and bounded worker probe. `web/connection-testing.js` owns the per-row button, dialog, escaped checklist and Settings toggle. Results appear in the current dialog rather than a persisted status badge: they are a snapshot, not ongoing monitoring. Audit event `connection.tested` records actor ID, connection ID and overall status, without credential contents or worker output.

## Verification

Parallel Go tests inject fake worker responses and use mocked database events. They cover ready/signed-out states, bad runtime/auth responses, unavailable workers, endpoint and transport denial, unsupported adapters, owner enforcement and enabled/disabled flags. UI tests verify readiness labels and escaping. No tests connect to PostgreSQL, insert live test data, sign in to real accounts or launch provider agents. Provider response success, failure, invalid result and transport failure are covered with injected fake worker responses.

The creation form uses **Containerized CLI** and **Container worker address**. New connections default to Internal network. Setup guidance explicitly marks a connection as unverified until tested; registering or editing never establishes working status. Other provider icons do not imply executable adapters.

## Connection editor layout

The creation/edit overlay uses a wide two-column desktop layout: connection fields on the left and contextual setup/sign-in/test guidance on the right. The body scrolls independently while the close X and action footer remain available. Below 780px the columns stack; below 520px provider/method and location choices also stack. Address drafts, hidden local provider fields, validation and saved connection edits retain their existing behavior. This layout is part of the optional `COMMONROOM_CONNECTION_EDITOR` presentation feature (enabled by default); disabling it restores the basic editor without changing connection APIs or isolation enforcement.

## Testing before saving

In Create, Test connection uses the fields currently entered without registering a connection. If its worker setup passes, the result says **Save and sign in to finish testing**. Register it, sign in, then test a provider reply.

In Edit, an unchanged endpoint and connection method can receive a full provider check using that saved connection's credentials. If either is changed, the test checks the new worker setup only; save the changes and authenticate before full verification. Unsaved names and provider icons do not change credential authority.

Testing never saves form changes. Results are invalidated when fields or the example address change, including edits made while a request is running. Inline buttons share the owner-only, server-enforced `connectionTesting` feature flag. Draft tests may append audit status metadata but never create or update connection records. Mocked tests verify those boundaries and ensure draft checks never read credentials or launch model turns.

The **Setup guide** button sits at the top of the setup sidebar for every connection method. It opens detailed instructions in a separate help overlay with its own close X. Closing help returns to the connection form with entered values intact. Instructions do not expand the editor body or push its controls down.
