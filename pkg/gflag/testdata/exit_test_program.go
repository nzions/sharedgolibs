//go:build ignore
// +build ignore

package main

import (
	"os"

	"github.com/nzions/sharedgolibs/pkg/gflag"
)

func main() {
	// Create a flagset with ExitOnError (default behavior)
	fs := gflag.NewFlagSet("testprog", gflag.ExitOnError)
	fs.Bool("valid", "v", false, "a valid flag")

	// Parse command line arguments
	fs.Parse(os.Args[1:])

	// If we reach here, parsing was successful
	os.Exit(0)
}
