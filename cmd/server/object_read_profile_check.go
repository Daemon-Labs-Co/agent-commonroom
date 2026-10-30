package main

import (
	"fmt"
	"io"
)

// Offline operator preflight: no controller lease, database, daemon, listener,
// provider, model or network access. Success validates private configuration only;
// it grants no execution/export authority and proves no physical/live acceptance.
func objectReadProfileCheckMain(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "Usage: --check-reader-profiles /absolute/private/profiles.json")
		return 2
	}
	profiles, err := loadObjectReadWorkerProfiles(args[0])
	if err != nil {
		if err == errObjectReadCertificatePEM {
			fmt.Fprintln(stderr, err.Error())
		} else {
			fmt.Fprintln(stderr, "Reader worker profile validation refused; check private file ownership, JSON policy and certificate configuration.")
		}
		return 1
	}
	if _, err = fmt.Fprintf(stdout, "Reader worker profiles validated (%d entries). Network, storage and live acceptance have not been verified.\n", len(profiles.entries)); err != nil {
		return 1
	}
	return 0
}
