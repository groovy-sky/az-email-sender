package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/groovy-sky/azemailsender/internal/cli/output"
)

// runConfig executes the config command
func (a *App) runConfig(args []string, globalFlags *GlobalFlags) int {
	if len(args) < 2 {
		a.printConfigUsage()
		return 1
	}

	subcommand := args[1]

	// Handle help
	if subcommand == "--help" || subcommand == "-h" || subcommand == "help" {
		a.printConfigUsage()
		return 0
	}

	switch subcommand {
	case "init":
		return a.runConfigInit(args[1:], globalFlags)
	case "show":
		return a.runConfigShow(args[1:], globalFlags)
	case "env":
		return a.runConfigEnv(args[1:], globalFlags)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown config subcommand '%s'\n", subcommand)
		a.printConfigUsage()
		return 1
	}
}

// runConfigInit creates a default configuration file
func (a *App) runConfigInit(args []string, globalFlags *GlobalFlags) int {
	var path string

	// Create flag set for config init command
	fs := flag.NewFlagSet("config init", flag.ContinueOnError)
	fs.Usage = func() {
		a.printConfigInitUsage()
	}

	fs.StringVar(&path, "path", "./azemailsender.json", "Path for the configuration file")
	fs.StringVar(&path, "p", "./azemailsender.json", "Path for the configuration file")

	// Parse flags
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	formatter := output.NewFormatter(globalFlags.JSON, globalFlags.Quiet, globalFlags.Debug)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		formatter.PrintError(fmt.Errorf("failed to create directory %s: %w", dir, err))
		return 1
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		formatter.PrintError(fmt.Errorf("configuration file already exists at %s", path))
		return 1
	}

	// Create default configuration file
	if err := SaveDefaultConfig(path); err != nil {
		formatter.PrintError(fmt.Errorf("failed to create configuration file: %w", err))
		return 1
	}

	formatter.PrintSuccess("Configuration file created at %s", path)
	return 0
}

// runConfigShow shows current configuration
func (a *App) runConfigShow(args []string, globalFlags *GlobalFlags) int {
	// Create flag set for config show command
	fs := flag.NewFlagSet("config show", flag.ContinueOnError)
	fs.Usage = func() {
		a.printConfigShowUsage()
	}

	// Parse flags
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	formatter := output.NewFormatter(globalFlags.JSON, globalFlags.Quiet, globalFlags.Debug)

	// Load configuration
	cfg, err := LoadConfig(globalFlags.Config)
	if err != nil {
		formatter.PrintError(fmt.Errorf("failed to load configuration: %w", err))
		return 1
	}

	// Hide sensitive data for display
	displayConfig := *cfg
	if displayConfig.AccessKey != "" {
		displayConfig.AccessKey = "***HIDDEN***"
	}
	if displayConfig.ConnectionString != "" {
		displayConfig.ConnectionString = "***HIDDEN***"
	}

	if err := formatter.PrintConfig(displayConfig); err != nil {
		return 1
	}

	return 0
}

// runConfigEnv shows environment variable examples
func (a *App) runConfigEnv(args []string, globalFlags *GlobalFlags) int {
	// Create flag set for config env command
	fs := flag.NewFlagSet("config env", flag.ContinueOnError)
	fs.Usage = func() {
		a.printConfigEnvUsage()
	}

	// Parse flags
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	formatter := output.NewFormatter(globalFlags.JSON, globalFlags.Quiet, globalFlags.Debug)

	if globalFlags.JSON {
		envConfig := map[string]string{
			"AZURE_EMAIL_ENDPOINT":          "https://your-resource.communication.azure.com",
			"AZURE_EMAIL_ACCESS_KEY":        "your-access-key",
			"AZURE_EMAIL_CONNECTION_STRING": "endpoint=https://your-resource.communication.azure.com;accesskey=your-access-key",
			"AZURE_EMAIL_FROM":              "sender@yourdomain.com",
			"AZURE_EMAIL_REPLY_TO":          "reply@yourdomain.com",
			"AZURE_EMAIL_DEBUG":             "false",
			"AZURE_EMAIL_QUIET":             "false",
			"AZURE_EMAIL_JSON":              "false",
		}
		if err := formatter.PrintConfig(envConfig); err != nil {
			return 1
		}
	} else {
		fmt.Print(GetEnvConfigExample())
	}

	return 0
}

// printConfigUsage prints usage information for the config command
func (a *App) printConfigUsage() {
	fmt.Printf(`Manage configuration files and environment variables for %s

Usage:
  %s config <subcommand> [flags]

Available Subcommands:
  init        Create a default configuration file
  show        Show current configuration
  env         Show environment variable examples

Use "%s config <subcommand> --help" for more information about a subcommand.
`, a.name, a.name, a.name)
}

// printConfigInitUsage prints usage information for the config init command
func (a *App) printConfigInitUsage() {
	fmt.Printf(`Create a default configuration file.

Examples:
  # Create config in current directory
  %s config init

  # Create config in specific location
  %s config init --path ~/.config/azemailsender/config.json

Usage:
  %s config init [flags]

Flags:
  -h, --help         help for init
  -p, --path string  Path for the configuration file (default "./azemailsender.json")

Global Flags:
  -c, --config string   Configuration file path
  -d, --debug           Enable debug logging
  -j, --json            Output in JSON format
  -q, --quiet           Suppress output except errors
`, a.name, a.name, a.name)
}

// printConfigShowUsage prints usage information for the config show command
func (a *App) printConfigShowUsage() {
	fmt.Printf(`Show the current configuration loaded from files and environment variables.

Examples:
  # Show current configuration
  %s config show

  # Show configuration from specific file
  %s config show --config ~/.config/azemailsender/config.json

Usage:
  %s config show [flags]

Flags:
  -h, --help   help for show

Global Flags:
  -c, --config string   Configuration file path
  -d, --debug           Enable debug logging
  -j, --json            Output in JSON format
  -q, --quiet           Suppress output except errors
`, a.name, a.name, a.name)
}

// printConfigEnvUsage prints usage information for the config env command
func (a *App) printConfigEnvUsage() {
	fmt.Printf(`Show examples of environment variables that can be used for configuration.

Examples:
  # Show environment variable examples
  %s config env

  # Save environment variables to file
  %s config env > .env

Usage:
  %s config env [flags]

Flags:
  -h, --help   help for env

Global Flags:
  -c, --config string   Configuration file path
  -d, --debug           Enable debug logging
  -j, --json            Output in JSON format
  -q, --quiet           Suppress output except errors
`, a.name, a.name, a.name)
}