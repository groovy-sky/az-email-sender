package app

import (
	"flag"
	"fmt"

	"github.com/groovy-sky/azemailsender/internal/cli/output"
)

// runVersion executes the version command
func (a *App) runVersion(args []string, globalFlags *GlobalFlags) int {
	// Create flag set for version command
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.Usage = func() {
		a.printVersionUsage()
	}

	// Parse flags
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	formatter := output.NewFormatter(globalFlags.JSON, globalFlags.Quiet, globalFlags.Debug)

	versionInfo := map[string]string{
		"version": a.version,
		"commit":  a.commit,
		"date":    a.date,
	}

	if globalFlags.JSON {
		if err := formatter.PrintConfig(versionInfo); err != nil {
			return 1
		}
	} else {
		fmt.Printf("%s version %s\n", a.name, a.version)
		fmt.Printf("Build commit: %s\n", a.commit)
		fmt.Printf("Build date: %s\n", a.date)
	}

	return 0
}

// printVersionUsage prints usage information for the version command
func (a *App) printVersionUsage() {
	fmt.Printf(`Show version, build commit, and build date information

Usage:
  %s version [flags]

Flags:
  -h, --help   help for version

Global Flags:
  -c, --config string   Configuration file path
  -d, --debug           Enable debug logging
  -j, --json            Output in JSON format
  -q, --quiet           Suppress output except errors
`, a.name)
}