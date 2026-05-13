// Command samuel is the entry point for the Samuel v2 CLI.
//
// All command wiring, error rendering, and exit-code mapping live in
// internal/commands. main.go stays tiny on purpose: this is the only
// file that imports os.Exit and the only place a non-zero exit code
// originates. Keeping it ~18 lines makes integration tests and the
// release smoke-test stable across versions.
package main

import (
	"fmt"
	"os"

	"github.com/ar4mirez/samuel/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err.Error())
		os.Exit(1)
	}
}
