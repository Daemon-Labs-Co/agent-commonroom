# Quick start by environment

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[README](../../../README.md) · [Detailed installation](installation.md)

Get the human workspace running first; provider sign-in and isolated agent workers
are a separate stage. These scripts build and optionally run the controller. They
do not install system packages, create database roles, change nginx, overwrite
credentials or enable unfinished repository development. Run from a trusted source
checkout. No script requires administrator privileges. No product flag is needed
for these operator build helpers.

## 1. Choose your platform

Install Git, Go 1.26+ and a PostgreSQL client/server or an approved connection to an
existing server. Use the [installation choices](installation-options.md) for package,
mise, Docker and nginx instructions. Clone the repository, then change into it.
All scripts resolve their own repository root, including when invoked elsewhere.

| Environment | Build command | Notes |
| --- | --- | --- |
| macOS | `bash scripts/setup-macos.sh` | Homebrew or official Go installer; Linux worker required for agent turns. |
| Linux | `bash scripts/setup-linux.sh` | Existing or separate PostgreSQL; optional nginx. |
| Windows | `powershell -File scripts/setup-windows.ps1 -LinuxRepoPath /home/YOU/agent-commonroom` | Install WSL2 first; keep the checkout and Go inside Linux. Replace the example path. |
| Hybrid Linux host | `bash scripts/setup-hybrid.sh` | Checks nginx, Docker, psql and SSH tools before building; does not reconfigure them. |

Windows native controller/worker execution is not a supported setup here. Inside
WSL2 you can also run `bash scripts/setup-linux.sh` directly. The PowerShell wrapper
uses the default WSL distribution. Run database/start commands in its Linux shell
so credentials do not have to cross the Windows/WSL environment boundary. WSL
worker resource-limit enforcement still needs separate validation; a Debian
worker is the documented agent deployment path.

Use `--check` for a non-mutating build-prerequisite check, `--build` (the default)
to compile, `--migrate` to apply the schema explicitly, and `--start` to rebuild
and start in the foreground. `--check` finds Go; the build enforces the version in
`go.mod`. It does not test PostgreSQL, proxy routing or worker admission.

## 2. Prepare the database once

Follow [restricted database setup](database-security.md#administrator-setup).
Use a **dedicated database**, separate migration owner and restricted runtime
login. Set passwords interactively with `psql`'s `\password`; use a private passfile
or approved peer authentication. Do not include passwords in URLs or command
arguments. The scripts never create or modify these logins for you.

After the administrator creates roles and configures authentication:

```sh
export MIGRATION_DATABASE_URL='postgres://commonroom_migrator@localhost:5432/commonroom?sslmode=disable'
bash scripts/setup.sh --migrate
unset MIGRATION_DATABASE_URL
```

The example disables TLS only for a same-machine database. For a separate database
server use `sslmode=verify-full`, its real hostname and the DBA-provided CA.
**Apply the restricted runtime grants after migration** using the SQL in the
database guide. Never start the web application as the migration owner.

## 3. Open your workspace

```sh
export DATABASE_URL='postgres://commonroom_runtime@localhost:5432/commonroom?sslmode=disable'
bash scripts/setup.sh --start
```

Open **http://127.0.0.1:8787/**. Find your private setup code in `.data/owner-key`,
create the owner account and keep a secure backup of the downloaded PEM sign-in
key. Create a room and send a message. Provider credentials are not required for
human collaboration. Ctrl+C stops the foreground controller; stop active work
before restarting an established installation.

The scripts inherit your environment; they do **not** source `.env` files. Avoid
printing environment variables or copying private state into support requests.
For a persistent Linux service follow [the Debian service guide](deployment/debian.md).
Keep the source assets beside the binary; this is not a standalone static executable.

## 4. Hybrid: Linux controller, nginx, Docker and separate PostgreSQL

Use this sequence for a Mac/Windows browser talking to your Linux host:

1. On Linux, install Go and the PostgreSQL client, retaining your existing nginx
   and Docker installation. Run `bash scripts/setup-hybrid.sh --check`, then build.
2. On the separate PostgreSQL server, ask the DBA to create the dedicated database
   and two roles. Permit only the Linux service host and approved administration
   paths; configure verified TLS and private password-file authentication.
3. On Linux, run `--migrate` with the migration URL. Have the DBA apply runtime
   grants, then remove migration credentials from the serving account.
4. Set `DATABASE_URL` for the runtime role with `sslmode=verify-full` and the CA
   path. Start the service bound to `127.0.0.1:8787`.
5. Add the private nginx listener at `127.0.0.2:8787` using the exact
   [nginx configuration](deployment/nginx-postgresql.md#3-private-nginx-configuration).
   Validate before reloading; preserve other sites.
6. From your laptop, open the tunnel below and use `http://127.0.0.1:8787/`.
7. Configure a dedicated **rootless** Docker worker identity on Linux, the approved
   images and host-key-verified SSH transport using [worker setup](connection-sign-in.md).
   Even on the same host, agent turns use that verified worker transport.

```sh
ssh -N -o ExitOnForwardFailure=yes \
  -L 127.0.0.1:8787:127.0.0.2:8787 your-linux-server
```

A running Docker daemon is not proof of rootless enforcement or an approved worker.
The application currently validates loopback Host/Origin and matching ports;
public domains and internet-facing nginx hosting are unsupported. The tunnel
keeps your browser's expected origin intact. See the [deployment diagram](installation-options.md#hybrid-deployment).

## Verify the running installation

In a second terminal, use the same controller service account, checkout and runtime
environment, then run `python3 scripts/check-installation.py`. On the hybrid Linux
host add `--hybrid`. This diagnoses each stage and gives repair instructions.
See [installation diagnostics](installation-check.md) for checks, flags, exit codes
and limits. Windows users run it inside WSL, not native PowerShell.

## Verification and limits

The setup scripts have automated command-stub tests without database access or
provider launches. macOS/Linux live installation, PowerShell/WSL execution, package
installation, nginx, remote PostgreSQL TLS and live worker admission still require
operator acceptance. The scripts do not claim those checks have passed.

Troubleshoot by stage: missing Go → prerequisite setup; migration failure → owner
connection/authentication; startup privilege refusal → runtime grants; browser
origin failure → exact loopback addresses/ports; agent readiness failure → worker
admission and provider sign-in. See [operations](operations.md) for recovery.
