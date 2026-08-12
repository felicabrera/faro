package main

import (
	"flag"
	"fmt"

	"golang.org/x/mod/sumdb/note"
)

// runKeygen generates the log's checkpoint signing key pair.
//
// The format is the one from golang.org/x/mod/sumdb/note, the same signed-note
// format used by the Go checksum database and by tlog-tiles logs generally. That
// choice is what lets third-party verifiers, written against the C2SP specs and
// not against our code, check a FARO checkpoint.
//
// The name given here becomes the checkpoint's origin line, the string a
// verifier pins to be sure a proof came from this log and not another. It should
// identify the log unambiguously, which in practice means a hostname the
// operator controls.
func runKeygen(args []string) error {
	fs := flag.NewFlagSet("keygen", flag.ExitOnError)
	name := fs.String("name", "", "log origin, e.g. faro.example.uy/elections (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("keygen: --name is required; it becomes the checkpoint origin that verifiers pin")
	}

	skey, vkey, err := note.GenerateKey(nil, *name)
	if err != nil {
		return fmt.Errorf("keygen: generating the key pair: %w", err)
	}

	fmt.Printf("origin:     %s\n", *name)
	fmt.Printf("public key: %s\n", vkey)
	fmt.Println()
	fmt.Println("Publish the public key and the origin. Anyone verifying this log needs both.")
	fmt.Println()
	fmt.Println("Private key (set as FARO_SIGNING_KEY; whoever holds it can sign a tree head):")
	fmt.Printf("  %s\n", skey)
	return nil
}
