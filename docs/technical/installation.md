# Controller setup and preflights

## Input

Explicit startup command, private configuration, runtime DATABASE_URL and separate migration identity. Controller paths resolve from the repository root.

Optional COMMONROOM_PUBLIC_ORIGIN accepts one exact lowercase HTTPS origin with
no path, credentials, query, fragment or wildcard. It is unset by default and
operator-controlled through process/service configuration, not room events.

## Output

Started loopback controller, offline validation result, or nonzero refusal. Offline profile validation grants no execution authority.

## Functionality

--migrate performs explicit schema migration. --check-reader-profiles and --check-capture-profiles read private policies/identities before controller or DB startup. --check-worker-storage reads pre-provisioned quota observations. Invalid command intent refuses rather than starting the controller.

Database setup uses one dedicated database with separate migration-owner and
restricted runtime logins. Runtime grants apply only after migration, on a
connection explicitly selecting that database and port. GRANT CONNECT grants
access without changing the current database. A missing public.events relation
requires checking the connection and migration result. Provisioning SQL should
stop on its first error; rollback undoes changes made in that transaction.

With a public origin configured, the listener remains numeric loopback-only and
the TLS proxy forwards the original Host and Origin. Only the configured public
host and matching HTTPS Origin are accepted in addition to existing local
origins. Cross-site requests remain refused; forwarded headers do not establish
trust. Existing login, authorization and CSRF checks still apply. Session and
logout cookies gain Secure in addition to HttpOnly and SameSite=Strict. The proxy
must serve a certificate trusted by the browser for the configured name.

Versioned [deployment templates](../../deployment/README.md) provide SQL, private
environment examples, nginx configuration and a user service; they contain no
passwords and invoke no privilege escalation or package installation.

## Properties and state changes

Normal startup acquires controller ownership and runtime connections. Profile preflights change no DB, daemon or account state. Migrations change schema only when explicitly selected.

Public-origin changes require restart. Unsetting the value restores local-only
host admission without deleting history. Invalid values refuse startup. Secure
cookies require HTTPS browser access when public-origin mode is enabled.

Source: [main.go](../../cmd/server/main.go) · [startup_command.go](../../cmd/server/startup_command.go) · [capture_profiles_check.go](../../cmd/server/capture_profiles_check.go).
