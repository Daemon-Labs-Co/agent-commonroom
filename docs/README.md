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
| Understand the implementation and data flow | [Architecture](architecture.md) |
| Integrate with or debug the HTTP interface | [API overview](api.md) |
| Work on the source and run checks | [Development guide](development.md) |
| Run, back up, and troubleshoot an installation | [Operations guide](operations.md) |
| Report vulnerabilities or understand security limits | [Security policy](../SECURITY.md) and [security review](../SECURITY_REVIEW.md) |
| Understand requirements and planned behavior | [Functional specification](../FUNCTIONAL_SPEC.md) |
| Contribute changes | [Contribution guide](../CONTRIBUTING.md) |

## Documentation status

The main README is the installation and capability overview. These guides describe the current source; FUNCTIONAL_SPEC.md also includes planned requirements, and DEVELOPMENT_STATUS.md contains historical notes. A feature described in those planning/history documents is not automatically implemented.

The next documentation passes will add source-level comments for functions, methods, interfaces, and types, then a file-by-file technical reference covering responsibilities, dependencies, callable entry points, variables/properties, and tests. Those passes are not part of this foundation yet.

Update relevant guides whenever behavior changes. Do not include real credentials, live transcripts, or private endpoint details in examples or screenshots.

## Read inside the application

Choose Help in the application header or Help & documentation beside Settings. Contextual Help opens the appropriate guide without discarding the current room or form. The viewer is also available at `/help/` on your installation, including before sign-in. Only the explicit documentation catalog and screenshot assets are served; private state and credentials are excluded.
