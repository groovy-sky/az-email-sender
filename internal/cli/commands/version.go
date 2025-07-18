package commands

import (
	"fmt"

	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/groovy-sky/azemailsender/internal/simplecli"
)

// NewVersionCommand creates the version command
func NewVersionCommand(version, commit, date string) *simplecli.Command {
	return &simplecli.Command{
		Name:        "version",
		Description: "Show version information",
		Usage:       "version",
		LongDesc:    "Show version, build commit, and build date information",
		Run: func(ctx *simplecli.Context) error {
			return runVersionCommand(ctx, version, commit, date)
		},
	}
}

func runVersionCommand(ctx *simplecli.Context, version, commit, date string) error {
	debug := ctx.GetBool("debug")
	quiet := ctx.GetBool("quiet")
	jsonOutput := ctx.GetBool("json")

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