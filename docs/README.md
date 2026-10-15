# Agent CommonRoom documentation

Start here for the current product and technical documentation. Agent CommonRoom is an early prototype; these guides distinguish implemented behavior from intended features.

## Choose a starting point

| You want to… | Read |
| --- | --- |
| Install and try the application | [Main README](../README.md#installation) |
| Use nginx and an existing PostgreSQL server | [nginx and PostgreSQL deployment](deployment/nginx-postgresql.md) |
| Install on a Debian server | [Debian deployment](deployment/debian.md) |
| Connect agents over SSH with explicit configuration | [SSH agent setup](agent-ssh.md) |
| Understand rooms, people, agents, and workflows | [Product concepts and usage](concepts.md) |
| Send tasks to agents and reconcile interrupted dispatch | [Task dispatch](task-dispatch.md) |
| Diagnose database waits that block execution | [Storage deadlines](storage-deadlines.md) |
| Understand sign-in limits and session renewal | [Login availability](login-availability.md) |
| Secure private credentials and worker lease files | [Private file safety](private-file-security.md) |
| Limit daily model launches | [Execution budgets](execution-budgets.md) |
| Execute workflows and recover interrupted work | [Workflow execution](workflow-execution.md) |
| Provision writable development agents | [Development workspaces](development-workspaces.md) |
| Configure mandatory isolated agents | [Container execution](container-isolation.md) |
| Understand the implementation and data flow | [Architecture](architecture.md) |
| Integrate with or debug the HTTP interface | [API overview](api.md) |
| Work on the source and run checks | [Development guide](development.md) |
| Run, back up, and troubleshoot an installation | [Operations guide](operations.md) |
| Report vulnerabilities or understand security limits | [Security policy](../SECURITY.md) and [security review](../SECURITY_REVIEW.md) |
| Understand requirements and planned behavior | [Functional specification](../FUNCTIONAL_SPEC.md) |
| Compare the agreed v0.1 scope with current development | [v0.1 gap assessment — 15 October 2026](v0.1-gap-assessment-2026-10-15.md) |
| Contribute changes | [Contribution guide](../CONTRIBUTING.md) |

## Documentation status

The main README is the installation and capability overview. These guides describe the current source; FUNCTIONAL_SPEC.md also includes planned requirements, and DEVELOPMENT_STATUS.md contains historical notes. A feature described in those planning/history documents is not automatically implemented.

The next documentation passes will add source-level comments for functions, methods, interfaces, and types, then a file-by-file technical reference covering responsibilities, dependencies, callable entry points, variables/properties, and tests. Those passes are not part of this foundation yet.

Update relevant guides whenever behavior changes. Do not include real credentials, live transcripts, or private endpoint details in examples or screenshots.

## Read inside the application

Choose Help in the application header or Help & documentation beside Settings. Contextual Help opens the appropriate guide without discarding the current room or form. The viewer is also available at `/help/` on your installation, including before sign-in. Only the explicit documentation catalog and screenshot assets are served; private state and credentials are excluded.

See [isolated connection sign-in](connection-sign-in.md) for remote Codex/Claude authentication, connection credential storage and the Security feature flag.

[Testing connections](connection-testing.md): per-connection readiness checks and troubleshooting from the Connections screen.

[Help window preference](help-display.md): choose same-window or separate-window documentation per human profile.

- [Security hardening and remaining limits](security-hardening.md)
- [Restricted audit database setup and recovery](database-security.md)

## Free durable job safety

See [worker lifecycle, shutdown verification and diagnostics](worker-jobs.md) for mandatory protocol-v2 execution, durable intent, crash reconciliation, owner diagnostics and the `jobDiagnostics` flag (default on). Shutdown uncertainty never counts as success; no task is replayed automatically.

- [Context disclosure controls](context-disclosure.md)
- [Atomic operations and safe retries](atomic-operations.md)

- [Server structure and responsibilities](server-structure.md)

- [Measured security and refactoring benchmarks](development-benchmarks.md)

Planned security architecture: [Provider/tool credential separation](provider-tool-separation.md) — design and acceptance gates; not yet implemented.

- [Scoped tool audit callbacks](tool-audit-callbacks.md): mandatory metadata protocol, authorization, historical compatibility and integration limits.

- [P0 execution progress — 15 October 2026](p0-execution-progress-2026-10-15.md): verified build/audit changes and remaining runtime gates.
