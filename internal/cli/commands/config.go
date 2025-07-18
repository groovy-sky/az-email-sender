package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/groovy-sky/azemailsender/internal/cli/config"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/urfave/cli/v2"
)

// NewConfigCommand creates the config command
func NewConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage configuration",
		Description: "Manage configuration files and environment variables for azemailsender-cli",
		Subcommands: []*cli.Command{
			newConfigInitCommand(),
			newConfigShowCommand(),
			newConfigEnvCommand(),
		},
	}
}

func newConfigInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Create a default configuration file",
		Description: `Create a default configuration file.

Examples:
  # Create config in current directory
  azemailsender-cli config init

  # Create config in specific location
  azemailsender-cli config init --path ~/.config/azemailsender/config.json`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Path for the configuration file",
				Value:   "./azemailsender.json",
			},
		},
		Action: func(c *cli.Context) error {
			path := c.String("path")
			return runConfigInit(c, path)
		},
	}
}

func newConfigShowCommand() *cli.Command {
	return &cli.Command{
		Name:  "show",
		Usage: "Show current configuration",
		Description: `Show the current configuration loaded from files and environment variables.

Examples:
  # Show current configuration
  azemailsender-cli config show

  # Show configuration from specific file
  azemailsender-cli config show --config ~/.config/azemailsender/config.json`,
		Action: func(c *cli.Context) error {
			return runConfigShow(c)
		},
	}
}

func newConfigEnvCommand() *cli.Command {
	return &cli.Command{
		Name:  "env",
		Usage: "Show environment variable examples",
		Description: `Show examples of environment variables that can be used for configuration.

Examples:
  # Show environment variable examples
  azemailsender-cli config env

  # Save environment variables to file
  azemailsender-cli config env > .env`,
		Action: func(c *cli.Context) error {
			return runConfigEnv(c)
		},
	}
}

func runConfigInit(c *cli.Context, path string) error {
	// Get global flags
	debug := c.Bool("debug")
	quiet := c.Bool("quiet")
	jsonOutput := c.Bool("json")

	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		formatter.PrintError(fmt.Errorf("failed to create directory %s: %w", dir, err))
		return err
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("configuration file already exists at %s", path)
	}

	// Create default configuration file
	if err := config.SaveDefaultConfig(path); err != nil {
		formatter.PrintError(fmt.Errorf("failed to create configuration file: %w", err))
		return err
	}

	return formatter.PrintSuccess("Configuration file created at %s", path)
}

func runConfigShow(c *cli.Context) error {
	// Get global flags from parent context
	debug := c.Bool("debug")
	quiet := c.Bool("quiet")
	jsonOutput := c.Bool("json")

	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	// Load configuration
	configFile := c.String("config")
	cfg, err := config.Load(configFile)
	if err != nil {
		formatter.PrintError(fmt.Errorf("failed to load configuration: %w", err))
		return err
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

func runConfigEnv(c *cli.Context) error {
	// Get global flags
	debug := c.Bool("debug")
	quiet := c.Bool("quiet")
	jsonOutput := c.Bool("json")

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

	fmt.Print(config.GetEnvConfigExample())
	return nil
}