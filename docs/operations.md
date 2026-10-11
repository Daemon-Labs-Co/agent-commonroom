# Operations guide

> Managed agent execution now requires the [Linux/rootless-Docker worker](container-isolation.md). Direct SSH model dispatch is blocked until a remote isolated worker is implemented. Earlier SSH endpoint setup in this guide does not bypass that restriction.


[Documentation index](README.md)

## Supported installations

Run privately on loopback on your computer, or on a Debian server accessed through SSH forwarding. See the main README and deployment/debian.md for installation. Public reverse-proxy/internet hosting is not currently supported. The same browser and server port is required by host/origin checks.

## Configuration and assets

| Item | Behavior |
| --- | --- |
| DATABASE_URL | Overrides local config; otherwise falls back to .data/local-config.json and the development default. |
| LISTEN_ADDR | Defaults to 127.0.0.1:8787; numeric loopback only. |
| OWNER_KEY_FILE | Defaults to .data/owner-key; companion public/private files use suffixes. |
| web/ and migrations/ | Required alongside the executable; run from repository root. |
| .env.example | Documents environment variables; the server does not load .env automatically. |

The server applies migration DDL on startup. Review schema changes before upgrading; a versioned migration runner and separate runtime database role are not implemented.

## Sign-in and keys

First installation creates a setup code and an RSA key pair. The owner downloads the private key and proves possession through browser signature login; successful key login removes pending private material and the setup code from the server. The retained public key is needed for future sign-in.

Sessions expire after eight hours and are lost on restart. Member access and enabled status are checked on requests. Only the owner invites/disables humans or changes their roles and room access. There is no self-registration, general key recovery, rotation, or encrypted-PEM import.

Keep downloaded owner/member credentials in protected storage with backups. Do not include them in bug reports. Deleting a file is not a guarantee of physical erasure from storage or backups.

## Agent health

A participant's application status is normally Idle, Working, or Blocked. Working indicates an in-memory turn; Blocked can indicate provider/process errors. Native provider login and executable availability are separate from a saved connection's presence.

Investigate connection protocol/endpoint, service PATH and HOME, CLI login, SSH host verification/authentication, and agent.error events. A successful SSH login by your own account does not prove the application service account can connect.

Turns have a three-minute context deadline; native session IDs persist, but the work queue and cancellation handles do not. Before maintenance, finish or cancel active work and inspect remote jobs. Restarting the server or removing a participant is not a guaranteed remote-process cutoff.

## Integrity and history

The owner’s **History & retention** screen shows plain-language descriptions and timestamps for the latest 200 audit records. Search filters these recent descriptions and event types. Expand **Verification details** for the sequence number, original event type, and checksum. Descriptions are derived from the original event payload without rewriting it; message text, agent instructions, endpoints, and raw provider errors are not included. Older events that lack names use a generic description. This view is not a replacement for provider diagnostics or a complete searchable archive.

The history screen inspects recent event metadata and can request chain verification. Verification failure records an incident and blocks some agent operations. Startup/periodic detection and complete active-process shutdown are incomplete. Do not treat an incident flag as proof every agent has disconnected.

The checksum chain and database trigger provide local protection, not administrator-resistant immutability. No independent rollback anchor, external WORM archive, or complete metadata protection exists. Preserve evidence and backups; do not edit/delete/re-hash history to make a failed check pass. There is no complete recovery runbook implemented yet.

Retention, trash, and human-authorized purge are product requirements, not implemented operational features. Do not schedule database deletion on the assumption that the application's retention UI enforces it.

## Backup and update procedure

1. Choose a maintenance window and finish/cancel active work; inspect remote processes separately.
2. Stop Agent CommonRoom.
3. Back up PostgreSQL, local configuration/public-key/state files, and the deployed source version. Protect provider credentials separately and maintain the human-held private key backup.
4. Keep private backups outside the repository, with encrypted off-machine copies.
5. Deploy a reviewed commit/release and build with its required Go version.
6. Review startup migrations, start the service, sign in, and check rooms/history/provider status.

The Debian guide provides concrete commands. Test restore in an isolated installation. Do not overwrite a newer live event log with an older backup as a routine rollback; that loses authoritative history and the current chain may not detect it. Backups do not substitute for the planned independent immutable archive.

## Troubleshooting and escalation

Start with installation troubleshooting in the main README, then Debian service/tunnel checks if applicable. Share error summaries, application version, and redacted diagnostics; avoid raw private transcripts and credentials. Follow SECURITY.md for vulnerabilities and SECURITY_REVIEW.md for known limitations. No production service-level guarantees or automated recovery tools exist yet.

## Guided connection editor flag

`COMMONROOM_CONNECTION_EDITOR=true` (default) serves the compact editor. An operator can set it to `false` and restart to restore the previous form. This is a presentation flag: connection data, running agents, the API, and security checks remain unchanged. See [usage and technical behavior](concepts.md#guided-connection-editor). Invalid boolean values prevent startup.

## Workflow template catalog flag

`COMMONROOM_WORKFLOW_TEMPLATES=true` enables the main-menu catalog by default. Set `false` and restart to disable catalog serving and explicit template application. Existing workflows and ordinary editing remain available; no records are removed and no agent work is stopped. Only the server operator can change this environment setting. See [workflow templates](concepts.md#workflow-templates).

## Emergency stop procedure

1. As owner, select **Stop all agents**, or **Stop room agents** for a scoped stop. Confirm the action and optional reason.
2. Verify the emergency banner and readable stop event. Cancellation is requested; do not infer remote termination from the banner.
3. Inspect affected hosts, processes and provider/tool access. Use host/network containment where the application alone cannot stop ongoing effects.
4. If audit recording fails, the current process remains stopped locally. Do not restart until the failure is investigated and the stop is recorded successfully; an unrecorded latch is not durable.
5. Resume explicitly after investigation. Application resume retains separate room stops; cancelled turns are not restarted. Integrity incidents block resume.

The server must be reachable; serialized operations can delay the stop. Use host-level controls if it is unavailable. The owner can hide controls in **Settings → Feature flags → Security**. `COMMONROOM_AGENT_SAFETY_UI` supplies the initial default; a saved setting overrides it. Neither disables backend enforcement. With controls hidden, use an existing owner session’s protected [agent-safety API](api.md#emergency-agent-safety); do not expose credentials in URLs or shared scripts. These controls do not provide automatic rogue-behavior detection or guaranteed remote/child-process termination.

Profile photos use bounded immutable event storage. Revisions increase audit storage and remain until eligible human-authorized purge. Disable new uploads and display through Settings → Feature flags → Cosmetic; this is a presentation option, not historical erasure. No filesystem upload directory or additional image service is required.

### Private human session store

Unexpired human sessions are stored in `human-sessions.json` beside `OWNER_KEY_FILE` (normally `.data/`). Keep it owner-only; never copy it into the web directory, commit it, or include it in troubleshooting output. Corrupt/unsafe stores block startup. To force fresh login, stop the application, remove this private session file, then restart; this does not alter immutable history. Restart persistence is enabled by default via `COMMONROOM_PERSISTENT_SESSIONS`; set false and restart to clear the saved store and use memory-only sessions. Only one application process may use a store. See [the session contract](concepts.md#human-login-persistence) for expiry, revocation and verification limits.
