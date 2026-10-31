# Controller setup and preflights

## Input

Explicit startup command, private configuration, runtime DATABASE_URL and separate migration identity. Controller paths resolve from the repository root.

## Output

Started loopback controller, offline validation result, or nonzero refusal. Offline profile validation grants no execution authority.

## Functionality

--migrate performs explicit schema migration. --check-reader-profiles and --check-capture-profiles read private policies/identities before controller or DB startup. --check-worker-storage reads pre-provisioned quota observations. Invalid command intent refuses rather than starting the controller.

## Properties and state changes

Normal startup acquires controller ownership and runtime connections. Profile preflights change no DB, daemon or account state. Migrations change schema only when explicitly selected.

Source: [main.go](../../cmd/server/main.go) · [startup_command.go](../../cmd/server/startup_command.go) · [capture_profiles_check.go](../../cmd/server/capture_profiles_check.go).
