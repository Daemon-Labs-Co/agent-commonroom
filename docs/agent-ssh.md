# Connect an isolated Linux worker over SSH

[Documentation index](README.md) · [Connection concepts](concepts.md#connections-versus-participants)

SSH connects CommonRoom to its constrained Linux worker. It never authorizes running
Codex or Claude directly on the remote host. The controller can run on your Mac;
Docker and provider execution stay on the Linux worker.

## 1. Choose the controller and worker accounts

The account running CommonRoom owns the SSH configuration and private key. For a
Debian controller service this is normally `commonroom`, with configuration under
`/var/lib/commonroom/.ssh/`. For a laptop controller it is your laptop account.

Use a dedicated Linux worker account with a working rootless Docker daemon. See
[container requirements](container-isolation.md) before registering a connection.
Provider CLIs come from the reviewed agent image; installing a host CLI is not a
worker-readiness check.

## 2. Configure one approved alias

Add an entry to the controller account's `~/.ssh/config`, preserving existing entries:

```sshconfig
Host commonroom-worker
    HostName worker.example.com
    User commonroom-worker
    Port 22
    IdentityFile ~/.ssh/id_ed25519_commonroom
    IdentitiesOnly yes
    BatchMode yes
    StrictHostKeyChecking yes
```

Replace the hostname, remote account, port and key path. Protect the `.ssh` directory
with mode 0700 and private keys/configuration with mode 0600. Install only the
public key in the worker account's `authorized_keys`. Keep private keys outside Git.
The CommonRoom owner RSA key is a separate credential and must never be used for SSH.

Unattended SSH cannot answer password or passphrase prompts. A protected key needs
an explicitly managed SSH agent available to the controller service. The Debian
service example does not configure an SSH agent.

## 3. Verify host identity and noninteractive access

Obtain the server fingerprint through a trusted administrator or console before
accepting its key. Establish `known_hosts` interactively under the controller account:

```sh
ssh -o BatchMode=no -o StrictHostKeyChecking=ask commonroom-worker
```

Accept only the verified fingerprint, then leave the shell. Do not disable host-key
checking; `ssh-keyscan` alone does not authenticate a server.

After installing the worker wrapper and private configuration through the
[sign-in setup guide](connection-sign-in.md#worker-configuration), check it:

```sh
ssh -T -o BatchMode=yes commonroom-worker '~/.local/bin/commonroom-worker ready-v1'
```

For a Debian controller, run this as its service account, for example with
`sudo -H -u commonroom`. A successful SSH shell alone does not prove rootless Docker,
image policy or provider access.

## 4. Approve and register the worker

Configure the controller and restart:

```sh
export COMMONROOM_REMOTE_WORKER=true
export COMMONROOM_REMOTE_WORKER_ALIAS=commonroom-worker
```

In **Connections → New connection**, choose **Containerized CLI** and enter:

| Provider | Container worker address |
| --- | --- |
| Codex | `ssh://commonroom-worker` |
| Claude | `ssh://commonroom-worker/claude` |

The alias must match the approved controller configuration. Usernames, ports,
passwords, executable paths and query parameters do not belong in the endpoint.
Use SSH configuration for transport details. Distinct connection records can use
the same worker while retaining separate saved provider credentials.

## 5. Sign in, test, then invite

1. Save the connection.
2. Complete [isolated provider sign-in](connection-sign-in.md).
3. Run [Test connection](connection-testing.md) to verify a fixed provider reply.
4. Invite a fresh room participant using the saved connection.

A saved connection, credential-presence check or successful chat probe does not
prove development readiness. Public development dispatch remains blocked pending
the separated engine/tool integration and acceptance work described in
[development workspaces](development-workspaces.md).

## Troubleshooting

| Symptom | Check |
| --- | --- |
| Alias does not resolve | The controller account's SSH config and `HostName`. |
| Permission denied | Public-key installation, file ownership and unattended key access. |
| Host key changed | Confirm the change with the administrator before updating `known_hosts`. |
| Worker unavailable | Wrapper location, Linux/rootless Docker and reviewed image configuration. |
| Signed in but provider test fails | Provider account access, credential expiry and gateway policy. |
| Works in a terminal only | Repeat as the service account with BatchMode enabled. |

Cancellation is not proof of termination. Use [durable job diagnostics and
reconciliation](worker-jobs.md) for uncertain jobs; never invoke a host CLI as a fallback.
