# Connect and check an agent

[User documentation](README.md) · [Technical reference](../technical/README.md)

![Connect and check an agent: Choose connection → Sign in provider → Check worker → Test reply](../images/user-agents.svg)

Guide illustration; current availability is described below.

## Connect an agent

1. Open Connections and choose the supported provider/approved worker connection.
   New connections start with empty **Provider** and **Connection method** choices.
   Focus or click each field to see its explanation on the right. The setup
   overview appears when the form opens and when all required fields are filled.
   **Connection name** is your label; using a model name does not select that model.
   **CommonRoom server** means the machine hosting the application, not your browser.
   Choose **Another worker machine** for an approved worker on your network or the
   internet. Managed execution still requires Linux with rootless Docker.
2. Use the provider sign-in controls. Keep provider credentials in its official
   authentication tooling, outside task text and model-accessible files.
3. Choose Test connection and inspect each result separately: worker availability,
   sign-in status and an actual provider response are different checks.
4. Invite the configured agent into the intended room.

The managed worker requires supported Linux/rootless Docker isolation. A saved
connection label does not prove the worker can execute development tasks.
See the [execution contract](../technical/execution.md) for supported operations.

## Expected result

A successful test records a real response for that test only. It does not prove
repository editing, artifact persistence, pushing or all release gates.

## Small setup helper

With the controller already running, this read-only helper checks Help is reachable
and prints the page to open. It does not enable features, send tasks or change data.

```bash
bash scripts/user-guide.sh agents
```

The first worker location card shows the hostname of the machine running CommonRoom (for example, `debian`), not the website domain or your browser’s computer. Choose it when the worker belongs on that machine; choose **Another worker machine** otherwise.

### Use the machine hosting CommonRoom

1. Choose the card bearing the server hostname (for example, **debian**).
2. Choose **Containerized CLI** and the provider. The worker address is filled
   automatically; you do not need an SSH alias for the server itself.
3. The deployment operator must configure its isolated worker and enable
   `COMMONROOM_SERVER_WORKER=true` in `runtime.env`, with remote worker mode off.
4. Restart CommonRoom, then test the connection. A missing worker configuration,
   image or Docker prerequisite remains a blocking error. Sign in on the server
   when prompted; your browser machine does not host this worker.

```bash
# On the CommonRoom server, inspect the required private worker configuration.
test -f "$HOME/.config/commonroom-worker/config.json" &&
  stat -c '%a %n' "$HOME/.config/commonroom-worker/config.json"
# After the operator edits runtime.env:
systemctl --user restart commonroom.service
```

Expected result: the test checks the container worker on the named server rather
than asking for another machine. Provider readiness is reported separately.
