// harness-review signs and exclusively publishes an operator-reviewed release.
// Run on the trusted review workstation; never copy signing keys to workers.
package main

import (
	"agentroom/internal/harnessrelease"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func run(args []string, out io.Writer, now time.Time) error {
	flags := flag.NewFlagSet("harness-review", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	generate := flags.String("generate-key", "", "absolute path for a NEW offline review signing key")
	payloadPath := flags.String("review", "", "absolute path to canonical review JSON")
	keyPath := flags.String("signing-key", "", "absolute path to owner-only base64 Ed25519 private key")
	output := flags.String("output", "", "absolute path to a NEW signed envelope file")
	if flags.Parse(args) != nil || flags.NArg() != 0 {
		return harnessrelease.ErrReview
	}
	if *generate != "" {
		if *payloadPath != "" || *keyPath != "" || *output != "" {
			return harnessrelease.ErrReview
		}
		public, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return harnessrelease.ErrReview
		}
		if err := harnessrelease.Publish(*generate, []byte(base64.StdEncoding.EncodeToString(private)+"\n")); err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, "COMMONROOM_HARNESS_REVIEW_KEY="+base64.StdEncoding.EncodeToString(public))
		return err
	}
	if *payloadPath == "" || *keyPath == "" || *output == "" {
		return harnessrelease.ErrReview
	}
	payload, err := harnessrelease.ReadPrivate(*payloadPath, 4096)
	if err != nil {
		return err
	}
	encoded, err := harnessrelease.ReadPrivate(*keyPath, 128)
	if err != nil {
		return err
	}
	// Permit one final line ending in key files, never embedded whitespace.
	key, err := base64.StdEncoding.Strict().DecodeString(strings.TrimSuffix(strings.TrimSuffix(string(encoded), "\n"), "\r"))
	if err != nil {
		return harnessrelease.ErrReview
	}
	envelope, err := harnessrelease.Sign(payload, ed25519.PrivateKey(key), now)
	if err != nil {
		return err
	}
	if err = harnessrelease.Publish(*output, envelope); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "Signed release review published. Keep the signing key offline; stage the exact reviewed image separately.")
	return err
}
func main() {
	if err := run(os.Args[1:], os.Stdout, time.Now().UTC()); err != nil {
		fmt.Fprintln(os.Stderr, "Release review publication refused. Check canonical input, key, validity, permissions and a new output path.")
		os.Exit(1)
	}
}
