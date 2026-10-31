# Check an installation

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Quick start](quick-start.md) · [Operations](operations.md)

Use this after starting the controller to find the stage that needs fixing:

```sh
python3 scripts/check-installation.py
```

Python 3.8+ is required for this optional diagnostic. On Windows run inside WSL2.
Use the **controller's service account** and the same runtime environment, including
`DATABASE_URL` and any custom `OWNER_KEY_FILE`. The script locates its checkout
independently of your current directory. It does not source `.env`, read private
configuration contents, print keys/passwords or ask for a login. A service's
environment is not automatically available to a second terminal; arrange access
through your normal private configuration mechanism. Do not copy credentials into
command arguments or support reports.

## Choose your checks

```sh
# Controller, local files and read-only database catalog checks:
python3 scripts/check-installation.py

# Different configured port or an already-open private SSH tunnel:
python3 scripts/check-installation.py --url http://127.0.0.1:8888

# On the hybrid Linux host: also inspect nginx syntax and Docker identity:
python3 scripts/check-installation.py --hybrid

# No HTTP, database or hybrid-service commands; result remains incomplete:
python3 scripts/check-installation.py --offline
```

URL targets must be numeric loopback HTTP addresses with an explicit port. Ambient
HTTP proxies and redirects are disabled. Start the documented SSH tunnel yourself
when checking from your laptop. Running on the laptop verifies that route; running
on the Linux host checks the local route. A check of one route does not certify
the other. Browser Host/Origin and server ports must match the deployment recipe.

## What the output means

| Check | What it tests | Typical fix |
| --- | --- | --- |
| Application files | Required UI, Help and migration assets exist. | Restore the matching full checkout beside the executable. |
| Controller binary | `bin/commonroom` exists and is executable. | Run `bash scripts/setup.sh --build`. A different service binary location must be verified separately. |
| Go toolchain | Installed Go major/minor meets `go.mod`; automatic toolchain downloads disabled. | Install/activate the required Go version. A prebuilt installation can run without Go, but this remains a build warning. |
| Owner setup | Initial setup code or persistent public key exists. | Start the controller or correct the service account/key path; never generate replacements to silence the check. |
| Private permissions | Checked owner/key/session paths are owned by this account and inaccessible to group/others; direct symlinks rejected. | Restore documented ownership and 0700 directory / 0600 file permissions. |
| Integrity stop | An existing stop marker is flagged. | Investigate and use **Review integrity incident**; do not remove the marker manually. |
| Database | Read-only connection and the controller's catalog privilege contract. | Fix URL, authentication, verified TLS, networking, explicit migration or restricted grants. |
| Browser / script / Help | Expected unauthenticated responses from the running service. | Check startup logs, deployed assets, nginx forwarding and SSH tunnel. |
| Authentication boundary | Unauthenticated `/api/state` returns the expected sign-in rejection. | Investigate unexpected responses or bypasses; do not disable authentication. |
| Hybrid nginx | `nginx -t` under the current account. | Inspect syntax/access errors using an authorized account; no reload is performed. |
| Hybrid Docker | Current context responds and reports rootless mode. | Use the approved worker identity/context and enforceable limits. |

Each output line starts with **PASS**, **FAIL** or **WARN** and identifies the stage.
Failures include the next action. Raw subprocess and network errors are suppressed
because they can include credentials or private backend details; investigate those
through the service's protected logs.

## Database safety and configuration

The script requires `psql` and an explicitly exported runtime `DATABASE_URL`. It
passes the URL through the child environment, not process arguments; privileged
local administrators can still inspect environments. Credentials stay in the
normal private passfile/approved mechanism. Migration credentials are removed
from the database-check child environment.

The SQL is read from `runtimePrivilegeSQL` in the matching
`cmd/server/database_boundary.go`, preventing a separate copied privilege policy.
The client disables startup files and password prompts, sets connection/statement
timeouts, and forces read-only transactions with the documented search path. It
queries only catalog metadata, not room events. It creates no schemas, inserts no
data and does not migrate or repair anything. A connection or missing-schema
failure has intentionally generic guidance; check protected PostgreSQL logs to
identify the precise authentication/network/schema cause.

No environment URL means **incomplete**, even if the running service uses a private
local-config fallback. Never pass migration credentials to get this test green.
The check establishes connectivity with the supplied configuration; the operator
must still use verified TLS for a remote database and confirm it matches the
serving process's configuration.

## Exit codes and limits

- **0:** all selected automated checks passed. Manual acceptance still applies.
- **1:** one or more failures; fix the named stages and rerun.
- **2:** checks are incomplete, or command-line arguments are invalid. No complete
  installation claim is made. Failures take precedence over incomplete checks.

No product feature flag is needed: this is operator verification, not a new
application capability. It changes no application configuration, keys, records,
services or container state. It does not build, start the controller, authenticate,
reload nginx, run migrations, launch providers or automatically repair failures.
Standard diagnostic commands may produce ordinary service access/log entries.

A green result does **not** certify binary freshness, the whole filesystem,
cryptographic key validity, audit-chain integrity, authenticated operations,
nginx's active loaded configuration, remote worker transport, enforced container
limits, provider readiness or unfinished P0 development execution. Docker info
checks only the current context, not a separately configured worker. In the
browser, complete owner sign-in, load a room, open Help and use **Connections →
Test connection** for each approved provider. Keep remaining acceptance gates in
[release readiness](release-readiness.md).

Automated tests use temporary files and mocked commands/HTTP responses. They do
not contact a database, HTTP service, Docker daemon or provider. Live diagnostic
execution on each supported deployment still requires operator acceptance.
