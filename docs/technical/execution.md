# Isolated execution and workflows

## Input

Operator-controlled `COMMONROOM_SERVER_WORKER=true|false` (default `false`),
reserved connection endpoints `ssh://commonroom-server` (Codex) and
`ssh://commonroom-server/claude`. These references retain the existing connection
contract; they do not initiate SSH. Only deployment operators change this flag;
restart is required. Enabling server and remote worker modes together refuses startup.
Worker configuration remains the private `~/.config/commonroom-worker/config.json`
with an absolute credential directory and digest-pinned installed image.

Approved worker target, repository binding, reviewed images, provider configuration and current task/tool authority. Model input supplies no additional privileges.

## Output

The application executable is invoked with `--agent-worker` and a constrained
protocol action. No arbitrary executable, model CLI, shell or URL is accepted.
Database credential environment variables are removed from the child environment.

Bounded provider result, command receipt, lifecycle evidence or explicit refusal/uncertainty.

## Functionality

Connection checks, provider authentication, model requests, audit links, cancellation,
job inspection and startup/resume idle verification use the existing isolated-worker
protocol on the application machine. Interactive sign-in commands run on that machine.
The worker continues to require Linux/rootless Docker, container limits and pinned
images. Missing prerequisites fail closed; no direct host model execution or SSH
fallback exists. Source forks that bypass this entry point invalidate containment.

Managed turns require the approved isolated worker. The current production development check rejects configured repository execution pending the intended credential-separated bridge. Internal RunCodex/capture/read assemblies do not enable public dispatch by themselves. Reviewed command profiles currently permit go test ./... and go build ./... only. Workflows support bounded sequential execution and explicit human gates.

## Properties and state changes

The server identity is `commonroom-server`. Disabling the flag blocks new worker
operations while retaining historical binding validation and saved data. The flag
is a transport choice and cannot disable mandatory containment, authorization,
job fencing or audit checks. UI registration does not prove runtime or provider
readiness; each check must succeed independently.

Durable intent precedes launch. Active/cancellation state changes during execution. Cleanup uncertainty preserves records and capacity. No host/SSH model-execution fallback or isolation off switch exists.

Source: [agents.go](../../cmd/server/agents.go) · [credential_boundary.go](../../cmd/server/credential_boundary.go) · [worker_codex_run.go](../../cmd/server/worker_codex_run.go) · [workflow_runner.go](../../cmd/server/workflow_runner.go).
