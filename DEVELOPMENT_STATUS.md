# Development status

First local milestone · 9 October 2026

## Running

Go HTTP server on http://127.0.0.1:8787 with PostgreSQL database `agentroom`. Go and PostgreSQL 17 were installed through Homebrew. PostgreSQL was started directly, without configuring login startup.

## Verified

- Go build and checksum alteration test pass.
- Browser JavaScript syntax check passes.
- HTTP integration: connection registration, room invitation/profile registration, human message commitment, memory revisions, optimistic stale-write rejection, request-token enforcement, and local chain verification pass.
- PostgreSQL rejects UPDATE against events, even when the statement targets no rows.

Development fixture messages and registrations are clearly labelled and retained in the audit history.

## Next milestones

1. Strengthen integration tests and repository structure; room/project CRUD and identities.
2. Complete profile/session capabilities and directory enforcement; implement real CLI and HTTP adapters with bounded concurrency.
3. Task/QA state machine, status reporting, cancellation and watchdog.
4. Memory discussion/consensus workflows and remote authentication/MCP.
5. Immutable archive/checkpoints, checksums covering every record metadata field, protected retention and human purge receipts; incident recovery.
6. TypeScript UI build and complete UI acceptance walkthrough.

The current UI uses dependency-free JavaScript as an initial bridge from the demo. The planned production UI remains TypeScript. No remote agent endpoints are contacted yet. Current checksums cover kind, canonical payload, and predecessor hash; full metadata and independent checkpoint protection remain pending. Local audit triggers do not resist the database owner. Integrity checks are manual at this milestone; production dispatch must verify health automatically.

## First live adapter

ChatGPT-authenticated Codex CLI now supports fresh conversations and subsequent turns resumed by stored thread ID. The room offers New ChatGPT session and per-participant Message controls. A real greeting and follow-up were exercised. Invocations are asynchronous, bounded to three minutes, serialized per participant, read-only, and preceded by a local integrity check. Errors append audit events. Room profile and shared-memory context accompany turns.

This is a Codex conversation authenticated with ChatGPT, not a ChatGPT website conversation. Chat-only instructions are supplied; read-only sandbox does not enforce room-specific read-directory isolation. Do not use this trial for sensitive cross-room isolation. Cancellation, streaming, periodic integrity watchdog, external checkpoint protection, and arbitrary live attachment remain pending.

## Saved conversation support

Existing conversation dialog accepts a Codex/Jerry conversation UUID. It validates the stored thread with app-server thread/read. The default forks saved history into an independent room session; explicit resume continues the original thread. Already-associated original sessions and detected active turns are rejected. Attachment provenance is recorded. This does not attach to a live terminal, import a ChatGPT website transcript, or expose the source transcript as room messages. A history-copy and subsequent recall turn were exercised against the development conversation.

## Optional invitation context and regular CLI invitations

Invitations and profiles support optional notes/transcripts, saved-conversation text preview, source labels, and editable excerpts. Context is persisted in profile revisions and supplied to Codex turns. Standard CLI connections with endpoint `codex` can now be invited and messaged directly. A participant named Jerry was invited through this path and received a harmless context fixture. Execution remains read-only; this is not a live terminal bridge or completed room directory sandbox.

## Requested future work — pluggable memory management

- Add a pluggable memory-management provider and integrate **Fulginn** (name supplied by the owner; exact project/API remains to be confirmed). This is a recorded feature request, not an implemented integration or part of the overnight bug-fixing scope.
- Preserve room/project/application scope, access control, conflict resolution, immutable audit provenance, and retention rules across provider implementations.
- Before implementation, confirm Fulginn’s interface, storage/deployment requirements and migration behavior. Follow the feature-completion rules: meaningful mocked tests, user and technical documentation, and a documented feature flag.

## Worker credential persistence — Codex verified

11 October 2026: the provider launcher fixes the Codex/Claude credential directories and explicitly selects file-based Codex storage. The slim image supplies the native clients' system TLS CA bundle and normalizes non-root launcher permissions. Credential export now uses a separate network-disabled, non-root container with only the login volume mounted read-only, bounded private output and atomic mode-0600 storage. Isolated turns keep stdin open so prompts reach the provider.

`make test`, `make check` and `make secrets` passed. Regression tests cover launcher behavior, actual exporter-script success/failure paths, export isolation arguments, bounded output, prompt forwarding, and preservation of old credentials on invalid exports. Automated tests use synthetic temporary files and mocked interactions, without PostgreSQL, Docker or provider launches.

Live Debian validation: a fresh human-authorized Codex login returned **Connection sign-in saved** for the dedicated `overnight-maintenance` connection. A new SSH process reported `credentialPresent:true`; a fresh isolated container then completed the fixed, tool-disabled connection check with `verified:true`. The worker binary and pinned provider image were updated. This verifies persistence across worker processes and disposable containers, not a host reboot, token expiry/refresh, other connection identities or Claude authentication. Unattended maintenance remains stopped.
