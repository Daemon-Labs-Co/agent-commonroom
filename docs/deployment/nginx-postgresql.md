# nginx with an existing PostgreSQL installation

[Documentation index](../README.md) · [Debian service setup](debian.md)

Use this guide if nginx and PostgreSQL are already installed on your Debian server. Keep your existing PostgreSQL major version, clusters, authentication rules, and other nginx sites. Do not reinstall PostgreSQL or replace its data directory.

This is a **private reverse proxy over SSH**, not public HTTPS hosting. Agent CommonRoom currently accepts only loopback Host/Origin values on its configured port. Adding a public domain, certificates, or forged proxy headers does not implement a secure public deployment. Public HTTPS requires explicit trusted-origin/proxy support and further security work.

The examples have been checked against application request rules and upstream documentation, but have not been exercised end to end on your server.

## 1. Reuse PostgreSQL safely

Ask the database administrator to identify the intended cluster, port, socket directory, version, and authentication method. Check the server version through that connection rather than relying only on the client version. Local development has used PostgreSQL 17; validate any other existing version with a separate test database before deployment.

Use a dedicated `commonroom` Linux service account, database login, and database. Follow steps 2–4 of the [Debian guide](debian.md) for the account, source/build, and service, adapting its database commands to your existing cluster. Create the login/database only if they do not already exist. Never point CommonRoom at an unrelated application's database: startup applies schema DDL.

For a same-host cluster using local peer authentication and the matching `commonroom` account, the example connection is:

```text
DATABASE_URL="postgres://commonroom@/commonroom?host=/var/run/postgresql&sslmode=disable"
```

Adapt the socket path and port to the existing cluster. This local Unix socket does not need a database password. Preserve peer authentication; do not change it to `trust` to make setup easier.

For a database on another host, use verified TLS and the DBA-provided CA/hostname. An example address without a password is:

```text
DATABASE_URL="postgres://commonroom@db.internal.example:5432/commonroom?sslmode=verify-full&sslrootcert=/etc/agent-commonroom/postgres-ca.pem"
```

The domain is a placeholder. Supply credentials separately through a private service-account PostgreSQL password file or approved credential management. pgx supports PostgreSQL password files; keep the file readable only by the service account. Do not commit actual passwords or put them in connection endpoints. Match the password-file host/port/database/user to the URL, and ensure the runtime account can read the CA file. Coordinate restricted network access and `pg_hba.conf` with the DBA.

The application currently applies migrations as the database owner. It does not need cluster superuser access; separation of migration and runtime roles remains planned.

## 2. pg_vector prerequisite

Assume the DBA has installed **pgvector** for the PostgreSQL server's actual major version. Package/build installation is outside this guide. Once available, a sufficiently privileged administrator can enable it in the dedicated database:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
SELECT extversion FROM pg_extension WHERE extname = 'vector';
```

The extension is named `vector`, even though the project is commonly called pgvector or pg_vector. See [upstream installation documentation](https://github.com/pgvector/pgvector). Agent CommonRoom currently does not store embeddings or query vector indexes; this extension is optional preparation and does not enable semantic memory by itself. Do not grant the application permanent superuser privileges just to install an extension.

## 3. Private nginx configuration

Use two server loopback addresses so nginx and the app can share the **same port** without a bind conflict:

- Agent CommonRoom: `127.0.0.1:8787`.
- nginx private listener: `127.0.0.2:8787`.
- SSH tunnel from your computer: local `127.0.0.1:8787` → server `127.0.0.2:8787`.

On Debian/Linux, the 127/8 range is loopback. Verify these addresses and port availability on your server. Keep existing public nginx sites unchanged; this server block listens only on loopback.

Create `/etc/nginx/sites-available/agent-commonroom` using your normal administrative editor:

```nginx
server {
    listen 127.0.0.2:8787;
    server_name localhost 127.0.0.1;

    client_max_body_size 128k;
    access_log off;

    location / {
        proxy_pass http://127.0.0.1:8787;
        proxy_http_version 1.1;
        proxy_set_header Host $http_host;
        proxy_set_header Connection "";
        proxy_cache off;
        proxy_read_timeout 120s;
    }
}
```

Preserve the original Host **including the port**, and leave Origin unchanged. Do not rewrite Origin, forge trusted-host headers, or strip security headers to bypass application checks. nginx forwards the normal Cookie and Origin headers; there is no public trusted-proxy mode here. Do not configure caching for authenticated pages/API responses.

Enable only the new site using your Debian package layout; first ensure that these paths are not already used by an existing configuration:

```sh
sudo ln -s /etc/nginx/sites-available/agent-commonroom /etc/nginx/sites-enabled/agent-commonroom
sudo nginx -t
sudo systemctl reload nginx
```

If `nginx -t` fails, correct the new configuration before reloading. Installations using `conf.d` instead of `sites-enabled` should use their existing inclusion mechanism rather than duplicating this server block. Do not replace the default site or disable unrelated sites.

## 4. Connect through nginx

Run on your computer:

```sh
ssh -N -o ExitOnForwardFailure=yes -L 127.0.0.1:8787:127.0.0.2:8787 your-server-alias
```

Open `http://127.0.0.1:8787`. Finish owner setup using the [Debian guide](debian.md#6-finish-owner-setup-in-your-local-browser). The original browser Host and Origin now reach CommonRoom through nginx unchanged.

If your computer already uses 8787, choose another port and change **all four** settings together: application LISTEN_ADDR, nginx listener, nginx upstream, and SSH forward/browser port. nginx and the app remain on different loopback IP addresses. A local/server port mismatch fails the current host checks.

Keep this listener private. No public DNS record, internet-facing listen directive, or TLS termination step is required for the SSH-tunnel approach. Public-domain HTTPS hosting is a separate, unimplemented deployment mode.

## Checks and maintenance

- Validate nginx configuration before every reload.
- Confirm the dedicated database connection works under the service account before app startup.
- Preserve `web/`, `docs/`, README/license/security documents, and `migrations/` in the deployment; the in-app Help viewer reads these files.
- Check that login, room loading, `/help/`, and contextual Help work through the tunnel.
- Follow the Debian guide's protected backup/update procedure. An extension install does not change the application's current integrity or isolation guarantees.

## References

- [nginx proxy directives](https://nginx.org/en/docs/http/ngx_http_proxy_module.html)
- [PostgreSQL password files](https://www.postgresql.org/docs/current/libpq-pgpass.html)
- [pgvector](https://github.com/pgvector/pgvector)
