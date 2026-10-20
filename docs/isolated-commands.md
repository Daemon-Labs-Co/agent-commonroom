# Bounded isolated Go commands

This internal worker capability lets an explicitly authorized Codex callback run
one fixed build or test command against its initialized repository. It is off by
default. Public production development dispatch remains blocked; this capability
alone does not make repository development release-ready.

## Enable or disable

The installation owner/operator can set `COMMONROOM_ISOLATED_COMMANDS=true` in
the **worker process environment**, then restart workers. Empty or `false` disables
it; other values reject session construction. This is trusted deployment policy,
not a model argument, room preference or public configuration endpoint.

A trusted admitted session must separately include the `command` operation grant.
The flag does not confer permission. A disabled worker rejects sessions containing
that grant. Enabled sessions without it retain file-only access. The model schema
advertises commands only when both checks pass; broker authorization independently
enforces the grant before audit intent or Docker execution. The authenticated
controller audit channel and immutable job/room/source scope remain mandatory.

Disable the setting and restart workers to prevent new command sessions. Drain or
cancel existing work through the normal supervisor first: already constructed
sessions retain their admitted setting. Disabling does not erase source, history,
receipts or retained volumes, and never disables container isolation.

## Requests and results

The callback accepts exactly these two argument profiles:

```json
{"operation":"command","command":["go","test","./..."]}
{"operation":"command","command":["go","build","./..."]}
```

No executable paths, extra arguments, shell strings, working directories,
environments, image choices, timeout changes or simultaneous file content/path
are accepted. Native provider shell, approval and patch tools remain refused.
This implements Go checks only; it is not a general command runner.

The worker runs `/usr/local/go/bin/go` in `/workspace/repository`, as UID/GID
1000, in the inspected credential-free tool container. The reviewed tool image
now includes the Go SDK from its digest-pinned build stage, including its license
notices. Review and rebuild the image before enabling; an old helper-only image
cannot run these commands. Dependencies must already be available in the approved
snapshot or vendored source: module and checksum downloads, automatic toolchain
installation, Go environment files, workspace discovery and CGO are disabled.
The existing network-none, read-only root, memory/CPU/PID and mount restrictions
apply. Go cache/module/temp directories are fixed under
`/workspace/.commonroom-command/`; they remain with the retained workspace volume.

Execution is bounded to 120 seconds or the remaining admitted scope lifetime,
whichever is shorter. Combined stdout/stderr is limited to 8 KiB. Overflow,
cancellation or execution uncertainty withholds the receipt and fences authority.
The trusted worker captures output; untrusted repository processes cannot author
the broker's receipt. Output is still untrusted text and can contain misleading
claims. Do not treat a model's description as observed execution evidence.

A receipt contains `operation`, `bytes`, output `sha256`, `encoding` (`base64`),
`data`, `clientExitCode`, `completion`, and the fixed argv `commandDigest`.
Zero Docker CLI status produces `completion: "succeeded"`. A nonzero CLI status
produces `"unconfirmed"`: it can indicate a compiler/test failure **or** a Docker
client/daemon failure. It does not certify the test process's exit status.
Broker audit success means the bounded receipt was delivered, not that tests
passed. Existing audit intent/outcome rules still gate execution and output;
command output itself is not copied into decision metadata.

## Terminal container lifecycle

A command is terminal for its tool container. Before releasing any receipt, the
worker removes the exact inspected immutable container ID and verifies its
absence. This kills background descendants, including detached processes; merely
waiting for the top-level command is insufficient. Failed or ambiguous cleanup
withholds output and requires reconciliation. Later file/command calls in that
session are refused. The paired supervisor must still stop and reconcile both
roles and preserve the source volume; this does not prove pair disposal or grant
permission to delete source. Start a separately admitted fresh pair for more work.

## Verification and remaining P0 work

Fast tests use synthetic runners, private local streams and mocked Docker/audit
interactions. They cover both flag states, missing grants, fixed profiles, output
bounds, nonzero status, malformed receipts, failed cleanup and terminal reuse.
Process-adapter tests use harmless local test subprocesses; no provider, database,
Docker daemon or repository compiler is launched by the automated suite.

Initial [tool-only live checks](live-tool-acceptance.md) now verify an image build,
a cold synthetic Go test, kernel limit values and disposable-container cleanup.
Full Linux rootless-Docker/paired-worker/provider verification remains outstanding.
The command digest and initial source-scope digest do **not** attest the current
working-tree revision after edits. Revision-bound artifact collection, broader
language/provider support, hard aggregate 5 GB disk enforcement and production
admission remain P0 gaps. Do not promote these internal receipts to a completed
release acceptance gate. Source changes can replace the runner or bypass the
broker; doing so violates the mandatory isolation/audit contract. No host fallback
or isolation off switch exists.

See [callback binding](codex-callback-driver.md), [audited sessions](audited-tool-session.md),
[tool runtime](tool-container-runtime.md), and [release readiness](release-readiness.md).

The P0 catalog now binds three command admission refusals to actual worker-session,
broker and controller audit code: feature off, missing command grant and failed
intent persistence. Denied policy reaches neither Docker nor audit I/O; failed
persistence reaches neither executor nor result release and latches the integrity
stop. These are synthetic refusal checks, not evidence of successful command
execution, revision binding, timeout cleanup or live-provider compatibility.
They introduce no product flag: the existing default-off command setting and
mandatory audit/authorization protections remain unchanged.

Outcome-persistence acceptance additionally runs real session bootstrap, terminal
command execution logic and exact-ID tool-container absence checks with mocked
Docker. An injected controller outcome-insert failure withholds the captured
result, retains the pending intent and source, latches integrity shutdown and
revokes further authority. A healthy control commits the outcome before releasing
its result. Neither path runs a real compiler or establishes live container/SSH
compatibility. Audit remains mandatory regardless of command feature settings.
