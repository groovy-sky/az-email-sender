package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/groovy-sky/azemailsender/internal/cli/commands"
	"github.com/groovy-sky/azemailsender/internal/simplecli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// isValidationError determines if an error is related to invalid arguments or usage
func isValidationError(err error) bool {
	errMsg := err.Error()
	validationIndicators := []string{
		"required flag",
		"unknown flag",
		"unknown command",
		"subcommand required",
		"validation failed",
		"at least one recipient required",
		"sender address required",
		"subject required",
		"authentication required",
		"message ID required",
		"invalid",
		"flag requires a value",
	}
	
	for _, indicator := range validationIndicators {
		if strings.Contains(errMsg, indicator) {
			return true
		}
	}
	return false
}

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

	// Add all commands
	app.AddCommand(commands.NewVersionCommand(version, commit, date))
	app.AddCommand(commands.NewConfigCommand())
	app.AddCommand(commands.NewStatusCommand())
	app.AddCommand(commands.NewSendCommand())



	if err := app.Run(); err != nil {
		// Check if we should output JSON and have access to global flags
		args := os.Args[1:]
		jsonOutput := false
		
		// Parse args to find --json flag (simple check)
		for _, arg := range args {
			if arg == "--json" || arg == "-j" {
				jsonOutput = true
				break
			}
			if strings.HasPrefix(arg, "--json=") {
				jsonOutput = true
				break
			}
			// Handle -j combined with other flags like -jq
			if strings.HasPrefix(arg, "-") && strings.Contains(arg, "j") && !strings.HasPrefix(arg, "--") {
				jsonOutput = true
				break
			}
		}
		
		// Determine exit code based on error type
		exitCode := 1
		if isValidationError(err) {
			exitCode = 2
		}
		
		// Output error in appropriate format
		if jsonOutput {
			errorOutput := map[string]interface{}{
				"error":     err.Error(),
				"success":   false,
				"exit_code": exitCode,
			}
			if jsonBytes, jsonErr := json.MarshalIndent(errorOutput, "", "  "); jsonErr == nil {
				fmt.Println(string(jsonBytes))
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(exitCode)
	}
}