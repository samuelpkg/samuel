package commands

import (
	"strings"

	"github.com/spf13/cobra"
)

// commandPath returns the invoked command path as it should appear in
// JSON output's `command` field. cobra.Command.CommandPath() includes
// the binary name ("samuel"); we strip that so JSON consumers see
// "version", "init", "run done", etc.
func commandPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	return strings.TrimPrefix(cmd.CommandPath(), "samuel ")
}
