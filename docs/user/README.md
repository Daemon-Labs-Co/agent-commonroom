# User documentation

![Choose a guide by functionality](../images/user-getting-started.svg)

Start with the job you want to do. Every guide includes an illustration, ordered
steps, an expected result and a small Bash helper. Images describe the current
prototype or are explicitly labeled guide illustrations; they do not prove live
readiness. Run helper commands from the repository root.

| Functionality | Guide |
| --- | --- |
| Install and open CommonRoom | [Install and open CommonRoom](getting-started.md) |
| Rooms and defined outcomes | [Rooms and defined outcomes](rooms.md) |
| Connect and check an agent | [Connect and check an agent](agents.md) |
| Create and dispatch a task | [Create and dispatch a task](tasks.md) |
| Run a workflow and human gates | [Run a workflow and human gates](workflows.md) |
| Inspect outputs and accept work | [Inspect outputs and accept work](results.md) |
| Storage limits and retained work | [Storage limits and retained work](storage.md) |
| Stop, diagnose and recover work | [Stop, diagnose and recover work](recovery.md) |
| Settings, history and sharing | [Settings, history and sharing](settings.md) |

```bash
bash scripts/user-guide.sh getting-started
```

Use the [technical reference](../technical/README.md) for input/output contracts,
properties and state changes. Legacy details are retained there as archival reference.
