package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/groovy-sky/azemailsender/internal/simpleconfig"
	"github.com/groovy-sky/azemailsender/internal/simplecli"
)

// NewConfigCommand creates the config command
func NewConfigCommand() *simplecli.Command {
	return &simplecli.Command{
		Name:        "config",
		Description: "Manage configuration",
		Usage:       "config [subcommand]",
		LongDesc:    "Manage configuration files and environment variables for azemailsender-cli",
		Run: func(ctx *simplecli.Context) error {
			return fmt.Errorf("subcommand required. Use --help to see available subcommands")
		},
		Subcommands: []*simplecli.Command{
			{
				Name:        "init",
				Description: "Create a default configuration file",
				Usage:       "config init [--path <path>]",
				LongDesc: `Create a default configuration file.

Examples:
  # Create config in current directory
  azemailsender-cli config init

  # Create config in specific location
  azemailsender-cli config init --path ~/.config/azemailsender/config.json`,
				Run: runConfigInit,
				Flags: []*simplecli.Flag{
					{
						Name:        "path",
						Short:       "p",
						Description: "Path for the configuration file",
						Value:       "./azemailsender.json",
					},
				},
			},
			{
				Name:        "show",
				Description: "Show current configuration",
				Usage:       "config show",
				LongDesc: `Show the current configuration loaded from files and environment variables.

Examples:
  # Show current configuration
  azemailsender-cli config show

  # Show configuration from specific file
  azemailsender-cli config show --config ~/.config/azemailsender/config.json`,
				Run: runConfigShow,
			},
			{
				Name:        "env",
				Description: "Show environment variable examples",
				Usage:       "config env",
				LongDesc: `Show examples of environment variables that can be used for configuration.

Examples:
  # Show environment variable examples
  azemailsender-cli config env

  # Save environment variables to file
  azemailsender-cli config env > .env`,
				Run: runConfigEnv,
			},
		},
	}
}

func runConfigInit(ctx *simplecli.Context) error {
	path := ctx.GetString("path")
	debug := ctx.GetBool("debug")
	quiet := ctx.GetBool("quiet")
	jsonOutput := ctx.GetBool("json")

	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("configuration file already exists at %s", path)
	}

	// Create default configuration file
	if err := simpleconfig.SaveDefaultConfig(path); err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}

	return formatter.PrintSuccess("Configuration file created at %s", path)
}

func runConfigShow(ctx *simplecli.Context) error {
	debug := ctx.GetBool("debug")
	quiet := ctx.GetBool("quiet")
	jsonOutput := ctx.GetBool("json")

	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	// Load configuration
	configFile := ctx.GetString("config")
	cfg, err := simpleconfig.LoadConfig(configFile, ctx.Flags)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Hide sensitive data for display
	displayConfig := *cfg
	if displayConfig.AccessKey != "" {
		displayConfig.AccessKey = "***HIDDEN***"
	}
	if displayConfig.ConnectionString != "" {
		displayConfig.ConnectionString = "***HIDDEN***"
	}

	return formatter.PrintConfig(displayConfig)
}

func runConfigEnv(ctx *simplecli.Context) error {
	debug := ctx.GetBool("debug")
	quiet := ctx.GetBool("quiet")
	jsonOutput := ctx.GetBool("json")

	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	if jsonOutput {
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
		return formatter.PrintConfig(envConfig)
	}

	fmt.Print(simpleconfig.GetEnvConfigExample())
	return nil
}