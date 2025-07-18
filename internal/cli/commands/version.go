package commands

import (
	"fmt"

	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/urfave/cli/v2"
)

// NewVersionCommand creates the version command
func NewVersionCommand(version, commit, date string) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version information",
		Description: "Show version, build commit, and build date information",
		Action: func(c *cli.Context) error {
			return runVersion(c, version, commit, date)
		},
	}
}

func runVersion(c *cli.Context, version, commit, date string) error {
	// Get flags from global context
	debug := c.Bool("debug")
	quiet := c.Bool("quiet")
	jsonOutput := c.Bool("json")

	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	versionInfo := map[string]string{
		"version": version,
		"commit":  commit,
		"date":    date,
	}

	if jsonOutput {
		return formatter.PrintConfig(versionInfo)
	}

	fmt.Printf("azemailsender-cli version %s\n", version)
	fmt.Printf("Build commit: %s\n", commit)
	fmt.Printf("Build date: %s\n", date)

	return nil
}