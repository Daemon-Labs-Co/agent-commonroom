# Isolated execution and workflows

## Input

Approved worker target, repository binding, reviewed images, provider configuration and current task/tool authority. Model input supplies no additional privileges.

## Output

Bounded provider result, command receipt, lifecycle evidence or explicit refusal/uncertainty.

## Functionality

Managed turns require the approved isolated worker. The current production development check rejects configured repository execution pending the intended credential-separated bridge. Internal RunCodex/capture/read assemblies do not enable public dispatch by themselves. Reviewed command profiles currently permit go test ./... and go build ./... only. Workflows support bounded sequential execution and explicit human gates.

## Properties and state changes

Durable intent precedes launch. Active/cancellation state changes during execution. Cleanup uncertainty preserves records and capacity. No host/SSH model-execution fallback or isolation off switch exists.

Source: [agents.go](../../cmd/server/agents.go) · [credential_boundary.go](../../cmd/server/credential_boundary.go) · [worker_codex_run.go](../../cmd/server/worker_codex_run.go) · [workflow_runner.go](../../cmd/server/workflow_runner.go).
