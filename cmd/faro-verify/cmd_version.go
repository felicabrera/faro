package main

import (
	"flag"
	"fmt"

	"github.com/felicabrera/faro/internal/version"
)

func runVersion(args []string) error {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	fmt.Printf("faro-verify %s\n", version.Read())
	return nil
}
