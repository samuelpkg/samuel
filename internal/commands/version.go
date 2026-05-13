package commands

import (
	"github.com/spf13/cobra"

	"github.com/ar4mirez/samuel/internal/ui"
)

// Build-time injected via -ldflags "-X .../commands.Version=..."
// Defaults match what `go run` produces locally so unit tests have
// something stable to assert against.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show CLI version information",
	Long: `Display version, commit SHA, and build date for the Samuel CLI.

The values are stamped into the binary at build time via -ldflags;
'dev' / 'none' / 'unknown' indicate an unstamped build (go run / go test).

Examples:
  samuel version              # human-readable
  samuel version --json       # v4 JSON envelope`,
	RunE: runVersion,
}

func init() { rootCmd.AddCommand(versionCmd) }

func runVersion(cmd *cobra.Command, _ []string) error {
	if JSONMode(cmd) {
		ui.PrintJSON(commandPath(cmd), map[string]string{
			"version":   Version,
			"commit":    Commit,
			"buildDate": BuildDate,
		})
		return nil
	}
	ui.Bold("Samuel CLI")
	ui.TableRow("Version", Version)
	ui.TableRow("Commit", Commit)
	ui.TableRow("Built", BuildDate)
	return nil
}
