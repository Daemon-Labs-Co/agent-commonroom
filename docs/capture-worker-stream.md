# Capture worker private stream

The local Linux rootless Docker worker implements `toolbroker.CaptureStarter` with
`dockerCLIRunner.StartCapture`. This is mandatory isolation infrastructure, without
a product toggle or host/SSH fallback. It is not a model-facing launch endpoint and
has not been deployed or live-accepted.

## Fixed invocation and ownership

Only trusted `CaptureRuntime` composition may invoke the adapter, after durable
creation/binding and independent admission of the exact immutable container ID.
The worker validates ID shape and cancellation before invoking anything, then uses
the existing local rootless-daemon validation. Its only command is:

```text
docker start --attach --interactive <exact-container-id>
```

No shell, arbitrary executable, provider process, replacement, name discovery or
reconnect is accepted. The helper must be the reviewed non-TTY capture image with
the fixed source/object mounts validated by the plan. The CLI handles its attach
stream framing; the application sees only helper stdout, never stderr. The process
environment contains only PATH, HOME and XDG_RUNTIME_DIR, as with the existing
approved Docker worker boundary; provider credentials and Docker context overrides
are not inherited. Rootless daemon validation still trusts the configured local
Docker installation and its administrator; it is not proof of image provenance.

The supervisor exclusively owns the returned stdin/stdout. These are never model
conversation or provider credential pipes. Existing private process plumbing owns
stdout separately from `Cmd.Wait`, preserving final buffered protocol bytes even
if the CLI exits before the protocol reader consumes them.

`Wait(ctx)` observes the reaped CLI's sanitized success/failure without closing
stdout. A canceled wait cannot certify success. `Close` cancels the process, closes
both pipe ends and joins its waiter; repeated close is safe. The runtime still
independently inspects the actual helper's exit status, removes its exact ID and
confirms absence. Closing the CLI alone does not prove container shutdown.

## Configuration, permissions and remaining gaps

No user configuration or bypass is added. Trusted assembly must use the same
approved runner/starter, retain creator ownership, authenticate current human/task
policy, enforce real physical source/object quotas and supply durable lifecycle and
custody journals. `StartCapture` itself is neither an authorization nor a quota API.
An external caller or source fork that supplies dishonest dependencies invalidates
the isolation guarantee. Do not call it directly to avoid `CaptureRuntime` checks.

The adapter is implemented but task-dispatch composition remains disabled/incomplete.
Still required are reviewed image/live acceptance, private volume provisioning and
quota enforcement, durable deployed authority/creator integration, recovery,
content validation and authenticated file delivery/review. No complete file-task
workflow or P0 release gate is certified by these tests.

Tests inject a private command factory that launches only the synthetic Go test
helper, never Docker, SSH, a provider or a database. They cover duplex bytes, final
buffered acknowledgment after exit, sanitized failed exit, cancellation, joined
close, fixed arguments/environment and refusal before invocation. Existing engine
stream tests continue to verify compatibility with the shared pipe ownership.
