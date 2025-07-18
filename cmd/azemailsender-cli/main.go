package main

import (
	"fmt"
	"os"

	"github.com/groovy-sky/azemailsender/internal/cli/commands"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := &cli.App{
		Name:  "azemailsender-cli",
		Usage: "Azure Communication Services Email CLI",
		Description: `A command-line interface for sending emails using Azure Communication Services.
Supports multiple authentication methods, flexible recipient management,
and both plain text and HTML email content.`,
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Configuration file path",
				EnvVars: []string{"AZEMAILSENDER_CONFIG"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable debug logging",
				EnvVars: []string{"AZEMAILSENDER_DEBUG"},
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress output except errors",
				EnvVars: []string{"AZEMAILSENDER_QUIET"},
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Output in JSON format",
				EnvVars: []string{"AZEMAILSENDER_JSON"},
			},
		},
		Commands: []*cli.Command{
			commands.NewSendCommand(),
			commands.NewStatusCommand(),
			commands.NewConfigCommand(),
			commands.NewVersionCommand(version, commit, date),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}