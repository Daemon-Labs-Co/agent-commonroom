package main

import (
	"fmt"
	"io"
)

// Validate command intent before controller leases, configuration or DB startup.
// A typo in an operator preflight must never silently start the application or
// migrate a database. Diagnostics never echo arbitrary arguments or credentials.
// This startup safety boundary has no optional feature-flag bypass.
func startupCommand(args []string, stdout, stderr io.Writer) (handled bool, code int) {
	if len(args) == 0 {
		return false, 0
	}
	usage := func() (bool, int) {
		fmt.Fprintln(stderr, "Invalid command or arguments. Use --help for supported commands.")
		return true, 2
	}
	switch args[0] {
	case "--check-reader-profiles":
		return true, objectReadProfileCheckMain(args[1:], stdout, stderr)
	case "--check-worker-storage":
		return true, workerStorageCheckMain(args[1:], stdout, stderr)
	case "--help", "-h":
		if len(args) != 1 {
			return usage()
		}
		_, err := fmt.Fprintln(stdout, "Usage: commonroom [command]\nNo command: start the controller.\n--check-reader-profiles ABSOLUTE_PATH: offline private profile validation.\n--check-worker-storage ABSOLUTE_PATH: read-only local storage observations.\n--migrate: explicitly run database migration.\n--agent-worker: internal isolated worker protocol.\n--help: show this help without starting services.")
		if err != nil {
			return true, 1
		}
		return true, 0
	case "--migrate":
		if len(args) != 1 {
			return usage()
		}
		return false, 0
	case "--agent-worker":
		return false, 0 // dedicated bounded protocol validates its own operands
	default:
		return usage()
	}
}
