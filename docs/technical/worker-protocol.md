# Worker protocol and transport

Describes the protocol the controller uses to reach an isolated worker as
implemented today. The container runtime extraction in
[runtime-separation](runtime-separation.md) must keep these actions compatible.
Per-action request and response types not listed here are still to be inventoried.

## Input

One worker invocation is one short-lived process: an action name and at most one
value operand.

| Transport | Command | Availability |
| --- | --- | --- |
| Same machine | The application executable run as `--agent-worker ACTION [VALUE]`. | Only when `COMMONROOM_SERVER_WORKER=true`; otherwise the command is empty and fails. |
| Remote | `ssh -T` to an approved alias, running `exec ~/.local/bin/commonroom-worker ACTION 'VALUE'`. | Requires a trusted host key. Options: `BatchMode=yes`, `StrictHostKeyChecking=yes`, `ConnectTimeout=10`, `ServerAliveInterval=10`, `ServerAliveCountMax=2`. |

When the worker runs under a separate account, the controller's SSH key is
authorized only for a forced command, `commonroom-ssh-gate` in the runtime project.
The gate accepts exactly `exec ~/.local/bin/commonroom-worker ACTION ['VALUE']`
with a known action and a base64url/hex value, runs the installed worker directly
with a fixed minimal environment, and refuses chained commands, plain shells and
interactive sessions. It was verified against a real sshd with a throwaway key.

The value is base64 (raw URL alphabet) JSON, hex, or a validated job identifier,
quoted for the remote shell. Actions handled by the worker entry point in source:
`ready-v1`, `verify-v1`, `status-v1`, `jobs-v2`, `idle-v1`, `run-v2`, `cancel-v2`,
`cancel-pair-v1`, `cancel-pair-workspace-v1`, `audit-link-v1`, `auth-login-v1`,
`auth-status-v1` and `auth-logout-v1`. Workspace capture and object reads use
separate helper executables with their own protocols, not covered by this page.

A run request (`run-v2`) is a JSON object with a lifecycle version, job ID, provider,
membership key, provider arguments, optional repository digest, tool-separation
flag, issue time and deadline. A worker holds private configuration in
`~/.config/commonroom-worker/config.json`: an absolute credential directory and a
digest-pinned installed image.

## Output

Success writes JSON to standard output and exits zero. Refusal or failure exits
nonzero with a bounded error. Raw container output is not returned.

| Action | Value | Result |
| --- | --- | --- |
| `ready-v1` | none | `{version: 1, lifecycleVersion: 2, configured, runtimeReady, credentials: {codex, claude}}` after checks, or failure. |
| `status-v1` | job ID | The worker's job record, or failure. |
| `cancel-v2`, `cancel-pair-v1`, `cancel-pair-workspace-v1` | fence request | A fence acknowledgement. |
| `idle-v1` | none | `{version: 1, lifecycleVersion: 2, idle}` within 8 seconds, or failure. |
| `run-v2` | encoded run request | The provider's standard output, streamed; the exit code reports success or failure. |
| `audit-link-v1` | audit scope | A bounded stream over standard input and output until the scope deadline, or a generic refusal. |

The remaining actions follow the same exit-code convention; their payload types are
defined in [remote_worker.go](../../cmd/server/remote_worker.go).

## Functionality

Validation is strict. A request longer than 16,384 characters is refused. Decoding
disallows unknown fields and trailing data. A run request must carry the current
lifecycle version, a valid job ID and a membership key of at most 1,024
characters. Its deadline must be after its issue time by at most three minutes,
not already expired, and not issued more than five seconds in the future. The
provider must be `codex` or `claude`. Arguments must equal the fixed policy for that
provider, with an optional resume identifier of at most 128 permitted characters. No
arbitrary option, mount, environment, shell command or URL is accepted.

`ready-v1` checks, within a 10-second limit: worker configuration, isolation
availability, Docker security and resource-limit support, and that the configured
agent image exists. It reports per provider whether a credential file exists as a
regular file, not group- or world-accessible, of at most 65,536 bytes. Cancellation
has an 18-second limit and requires the request version to match the action.

`run-v2` admits the job within 5 seconds, reads the prompt from standard input
(at most 1 MiB), and bounds standard output and diagnostics to 4 MiB each. Its
deadline never exceeds three minutes. A file named `JOBID.cancel` in the job state
directory, polled every 100 ms, cancels the job. The job record moves through
`starting`, `running`, then `finished`, `failed`, `cancelled` or
`shutdown_unconfirmed`. After the turn the worker requires proof, within 5 seconds,
that the container is gone; without it the record stays `shutdown_unconfirmed`
and the call fails.

Actions that use the container engine (`run-v2`, `jobs-v2`, the cancel actions)
require a local rootless daemon reached through the default Docker context on a
Unix socket. They refuse when `DOCKER_HOST` or `DOCKER_CONTEXT` is set in the
worker's environment, so the worker account must have its rootless context
configured outside the process environment. `ready-v1` does not check these
variables.

A repository-bound development request is refused until a verified
credential-separated provider bridge is configured
([credential_boundary.go](../../cmd/server/credential_boundary.go)).

## Properties and state changes

The worker process environment drops `DATABASE_URL`, `MIGRATION_DATABASE_URL`,
`PG*` and `POSTGRES_*`; the worker has no database role. Job records live in
`~/.local/state/commonroom-worker/jobs` (mode 0700) under the worker account,
and cancellation and status read them. `ready-v1` and `status-v1` change nothing.

The same-machine flag is a transport choice and cannot disable containment,
authorization, job fencing or audit checks. Remote transport alone never
authorizes host execution; there is no direct-host or uncontained fallback.
Extracting the worker into its own project must keep action names, versions and
framing unchanged until a documented version bump.

The separate runtime project serves every action listed above using the worker
code extracted unchanged from this repository, with the existing tests that
exercise only that code. This repository still contains the original; until the
controller stops using it, both must change together under a protocol version
bump. The runtime has not been accepted against live providers.

Source: [remote_worker.go](../../cmd/server/remote_worker.go) · [server_worker.go](../../cmd/server/server_worker.go).
