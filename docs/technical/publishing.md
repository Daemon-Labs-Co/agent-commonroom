# Repository publishing

## Input

Approved repository/branch, exact accepted revision, human publishing authority and protected repository credentials. This is the intended basic-product input contract.

## Output

Intended output: verified commit/push identity or explicit failure. No completed production publishing implementation exists yet.

## Functionality

Current agent instructions explicitly prohibit commit/push, and reviewed tool command profiles do not expose git publishing. A future publisher must consume approved changes outside model credential access. Repository-link metadata is not proof of a push or acceptance.

## Properties and state changes

No new publishing property is currently persisted by an implemented publisher. Future ambiguous remote results must remain recorded as uncertain rather than automatically retried or claimed successful.

Source: [agents.go](../../cmd/server/agents.go) · [task_repository_links.go](../../cmd/server/task_repository_links.go).
