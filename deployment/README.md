# Deploy CommonRoom as an ordinary user

![Deployment: check tools, prepare database, build controller and open workspace](../docs/images/user-getting-started.svg)

This is a setup illustration, not proof of a live deployment. Templates here are
versioned; passwords, generated SQL containing passwords and site configuration
belong outside Git. No helper installs packages or invokes sudo. Existing database,
DNS and HTTPS administrators configure their services through their normal access.

1. Build with `bash scripts/setup.sh --build`. Run the existing migration with the
   migration-owner login; see [database setup](../docs/technical/archive/database-security.md).
2. If no runtime login exists, ask the database administrator to run the role
   template and set its password with psql's interactive `\password` command.
   Administrators can use their own login; becoming an OS superuser is not required.
3. Apply grants **to the migrated application database**, using its actual port:

   ```bash
   psql -p 5432 -d commonroom -U your_database_administrator -v runtime_role=commonroom_runtime -f deployment/postgres/create-runtime-role.sql
   psql -p 5432 -d commonroom -U your_database_administrator -v database=commonroom -v runtime_role=commonroom_runtime -f deployment/postgres/runtime-grants.sql
   ```

   Existing runtime roles skip creation. `GRANT CONNECT` never switches databases.
   The grant template refuses a database-name mismatch and stops on errors.
4. Prepare private configuration and edit its URL, hostname and service paths:

   ```bash
   install -d -m 700 ~/.config/commonroom ~/.config/systemd/user
   install -m 600 deployment/runtime.env.example ~/.config/commonroom/runtime.env
   cp deployment/commonroom.service ~/.config/systemd/user/commonroom.service
   ```

   Configure exactly `COMMONROOM_PUBLIC_ORIGIN=https://agentcommonroom.home.arpa`
   for this hostname. Leave it unset for local-only browser access. It is read on
   restart and can be changed only by the operator controlling service configuration.
   After the service starts, run `commonroom --setup-link` on the server (no arguments)
   to print the link that claims the installation, and give it to the owner. A host
   provisioning many installations may instead supply `COMMONROOM_SETUP_CODE` (see the
   example file). See [getting started](../docs/user/getting-started.md).
5. Point DNS at the server, configure the existing HTTPS proxy using
   [the nginx template](nginx/commonroom.conf), and install a browser-trusted
   certificate for that name. Keep the application listener on loopback. Forward
   the original Host and Origin; never rewrite an untrusted origin to localhost.
6. Start the user service:

   ```bash
   systemctl --user daemon-reload
   systemctl --user enable --now commonroom.service
   curl --fail https://agentcommonroom.home.arpa/ -o /dev/null
   ```

Expected result: the HTTPS origin serves the login UI; unconfigured hosts and
cross-site requests remain rejected. Public-origin sessions use Secure, HttpOnly,
SameSite=Strict cookies. Removing the origin setting restores local-only admission
on restart without deleting application history. User-service startup at boot
without a login depends on the server's existing user-manager policy. No coding
execution or publishing readiness is implied by a successful HTTP check.

## Upgrading an existing installation to workspace storage

Order matters, because migration 002 makes every insert require a workspace setting that only
this version of the controller supplies, and the controller decides at startup whether to supply it:

1. Deploy this version of the controller. Before the migration it serves the previous layout unchanged.
2. Stop the controller. Run `commonroom --migrate` with the migration owner (or apply
   `migrations/002.sql` and `migrations/003.sql` as the database administrator), then the current
   `postgres/runtime-grants.sql` for the runtime role. Existing history becomes workspace `default`.
3. Start the controller. It detects workspace storage, refuses to start if the runtime role could
   bypass row-level security, and from then on scopes every transaction.

Never apply migration 002 while an older controller is running, and never leave the new controller
running across the migration: either way inserts fail until it restarts. Take a backup first
(`pg_dump -Fc`). Single-workspace installations behave exactly as before afterwards.

## Hosting many workspaces

Set `COMMONROOM_TENANCY=multi` and `COMMONROOM_WORKSPACE_DOMAIN` (see
`runtime.env.example`) only on an installation you run for other people. Requirements:

1. Apply `migrations/002.sql` and `migrations/003.sql` (`commonroom --migrate`) and the
   current `postgres/runtime-grants.sql`. The serving role must not own `events` and
   must not hold SUPERUSER or BYPASSRLS; startup refuses otherwise. Prove the database
   isolation on a scratch database with `scripts/check-workspace-isolation.sql`.
2. Wildcard DNS and a wildcard certificate for the domain, and the proxy in
   [commonroom-workspaces.conf](nginx/commonroom-workspaces.conf). Leave
   `COMMONROOM_PUBLIC_ORIGIN` unset; each workspace's origin is derived.
3. Create the control plane's signing key and register it:

   ```bash
   commonroom --platform-keygen ~/.config/commonroom/platform.key control-plane > ~/.config/commonroom/platform-keys.json
   chmod 600 ~/.config/commonroom/platform-keys.json
   # then set COMMONROOM_PLATFORM_KEYS_FILE to that path and restart
   ```
4. Create a workspace and get its setup link (the response also lists limits and status):

   ```bash
   commonroom --platform-request ~/.config/commonroom/platform.key control-plane POST \
     https://platform.rooms.example.com/platform/v1/workspaces '{"id":"acme","name":"Acme Ltd"}'
   ```

5. Optional: let visitors create their own workspace. Set `COMMONROOM_SIGNUP=open` (and, for
   a private beta, `COMMONROOM_SIGNUP_CODE`) and restart. The bare domain then serves the
   sign-up page, so the proxy and certificate must also cover `rooms.example.com` itself (the
   example configuration does) and must set `X-Real-IP`, which the per-address limits use.
   There is no email check: keep `COMMONROOM_SIGNUP_DAILY_LIMIT` and
   `COMMONROOM_SIGNUP_MAX_WORKSPACES` low until you trust the traffic.

To verify the database side on a scratch PostgreSQL server you own, run
`scripts/check-workspace-database.sh` (it creates and drops its own throwaway database).

Hosted workspaces cannot run agents unless you set `COMMONROOM_HOSTED_EXECUTION=true`
(chat-only agents on your isolated worker; see
[the workspaces design](../docs/technical/multi-tenancy.md) and verify on a staging
installation first). Deleting a workspace stops
serving it at once; `scripts/purge-workspace.sql` removes its history afterwards.

For a worker on the application machine, copy `worker-config.json.example` to
`~/.config/commonroom-worker/config.json` with mode `600`. Replace `image` with
the immutable installed CommonRoom agent image ID and `authDir` with an absolute
private credential directory. Set `COMMONROOM_SERVER_WORKER=true` and
`COMMONROOM_REMOTE_WORKER=false` in the service environment, then restart.
The same-machine worker invokes the constrained `--agent-worker` entry point;
it does not require an SSH alias or permit direct host model execution.
