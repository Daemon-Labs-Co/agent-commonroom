# Tasks, execution records and acceptance

## Input

Versioned seven-part task definition; assignee/reviewer; bounded typed input bindings; current task revision; human initiator. taskDefinitions defaults off.

## Output

Immutable execution records, validated value outputs, recorded failures and explicit human acceptance decisions.

## Functionality

Definitions specify task, expected inputs, process, desired outputs, criteria, execution limits and failure behavior. At least one output is required. Required file/bundle outputs currently block Executable until verified artifact integration. Structured duration is1–180seconds, response at most16KiB and tool calls1–64. Reviews reuse current work inputs.

## Properties and state changes

Dispatch advances task revision/status and creates pending intent before execution. Definition/input/configuration changes invalidate stale lineage. Acceptance does not follow automatically from an agent response.

Source: [task_dispatch.go](../../cmd/server/task_dispatch.go) · [task_execution.go](../../cmd/server/task_execution.go) · [task_acceptance.go](../../cmd/server/task_acceptance.go).
