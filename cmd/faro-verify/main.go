// Command faro-verify is the public audit CLI.
//
// This is the tool that makes FARO's claim testable by someone who does not
// trust us: it fetches the log over plain HTTP and checks it locally. Nothing it
// prints depends on the log telling the truth about itself.
//
// It follows ÁBACO's shape: stdlib flag only, one file per subcommand, manual
// dispatch. A member of the public should be able to read this program in an
// afternoon and satisfy themselves that it verifies what it says it verifies,
// which argues against a CLI framework and against clever abstractions.
//
// Today it carries `version` and `keygen`. The verification subcommands
// (`checkpoint`, `inclusion`, `consistency`, `monitor`) arrive with the log
// integration; the note key handling they need is already here.
package main

import (
	"flag"
	"fmt"
	"os"
)

var commands = map[string]struct {
	summary string
	run     func(args []string) error
}{
	"version": {"print build information", runVersion},
	"keygen":  {"generate a checkpoint signing key pair", runKeygen},
}

// order fixes the help listing, since map iteration is randomised.
var order = []string{"version", "keygen"}

func main() {
	flag.Usage = usage
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	name := os.Args[1]
	if name == "-h" || name == "--help" || name == "help" {
		usage()
		return
	}
	cmd, ok := commands[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "faro-verify: unknown command %q\n\n", name)
		usage()
		os.Exit(2)
	}
	if err := cmd.run(os.Args[2:]); err != nil {
		fatalf("%v", err)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, "faro-verify — FARO public audit CLI\n\nusage: faro-verify <command> [flags]\n\ncommands:\n")
	for _, name := range order {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", name, commands[name].summary)
	}
	fmt.Fprint(os.Stderr, "\nRun 'faro-verify <command> -h' for the flags of a command.\n")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "faro-verify: "+format+"\n", args...)
	os.Exit(1)
}
