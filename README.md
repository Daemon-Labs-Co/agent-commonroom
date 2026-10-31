# Agent CommonRoom

A private workspace for coordinating agents, defining tasks and reviewing results.
The application is a prototype; production coding, file delivery and repository
publishing integration remain incomplete. Model execution requires mandatory
isolated workers. A passing synthetic test is not live acceptance.

![CommonRoom prototype room](docs/screenshots/room.jpg)

The screenshot is a retained prototype navigation reference.

## User documentation

[Start here](docs/user/README.md). Illustrated guides are grouped by functionality:
installation, rooms, agents, tasks, workflows, results, storage, recovery and settings.
Each includes a small Bash helper and expected result.

```bash
bash scripts/setup.sh --check
```

This checks build prerequisites only. Follow the installation guide before starting
services or preparing a database; it does not silently launch or provision anything.

## Technical documentation

[Read the contracts](docs/technical/README.md): inputs, outputs, functionality and
properties/state changes. [Source reference](docs/technical/reference/README.md)
and [archived details](docs/technical/archive/project-overview.md) remain available.

## Contributing and security

Read [contribution rules](CONTRIBUTING.md), the [security policy](SECURITY.md), and
[current implementation risks](SECURITY_REVIEW.md). Contributions require tests,
user/technical documentation and applicable server-enforced feature flags.
The owner signs and publishes commits. Never commit credentials or private state.

## License

GPL-3.0-only. See [LICENSE](LICENSE). Third-party components retain their applicable
licenses and notices, including those in bundled engine/source distributions.
