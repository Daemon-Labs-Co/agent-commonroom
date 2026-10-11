# Operations guide

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
