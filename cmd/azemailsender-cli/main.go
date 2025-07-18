package main

import (
	"fmt"
	"os"

	"github.com/groovy-sky/azemailsender/internal/cli/commands"
	"github.com/groovy-sky/azemailsender/internal/simplecli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Create global CLI context
	app := simplecli.NewGlobalContext("azemailsender-cli", 
		`A command-line interface for sending emails using Azure Communication Services.
Supports multiple authentication methods, flexible recipient management,
and both plain text and HTML email content.`)

	// Add global flags
	app.AddGlobalFlag(&simplecli.Flag{
		Name:        "config",
		Short:       "c",
		Description: "Configuration file path",
		Value:       "",
	})
	app.AddGlobalFlag(&simplecli.Flag{
		Name:        "debug",
		Short:       "d",
		Description: "Enable debug logging",
		Value:       false,
	})
	app.AddGlobalFlag(&simplecli.Flag{
		Name:        "quiet",
		Short:       "q",
		Description: "Suppress output except errors",
		Value:       false,
	})
	app.AddGlobalFlag(&simplecli.Flag{
		Name:        "json",
		Short:       "j",
		Description: "Output in JSON format",
		Value:       false,
	})

	// Add commands using new framework
	app.AddCommand(commands.NewSimpleVersionCommand(version, commit, date))
	app.AddCommand(commands.NewSimpleConfigCommand())

	// Placeholder commands (to be migrated)
	app.AddCommand(&simplecli.Command{
		Name:        "send",
		Description: "Send an email",
		Usage:       "send [flags]",
		LongDesc:    "Send an email using Azure Communication Services.",
		Run: func(ctx *simplecli.Context) error {
			return fmt.Errorf("send command will be implemented in next step")
		},
	})

	app.AddCommand(&simplecli.Command{
		Name:        "status",
		Description: "Check email status",
		Usage:       "status <message-id>",
		LongDesc:    "Check the status of a previously sent email.",
		Run: func(ctx *simplecli.Context) error {
			return fmt.Errorf("status command will be implemented in next step")
		},
	})

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}