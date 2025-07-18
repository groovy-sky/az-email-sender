package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/groovy-sky/azemailsender/internal/cli/config"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates the config command
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Manage configuration files and environment variables for azemailsender-cli",
	}

	cmd.AddCommand(newConfigInitCommand())
	cmd.AddCommand(newConfigShowCommand())
	cmd.AddCommand(newConfigEnvCommand())

	return cmd
}

func newConfigInitCommand() *cobra.Command {
	var path string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default configuration file",
		Long: `Create a default configuration file.

Examples:
  # Create config in current directory
  azemailsender-cli config init

  # Create config in specific location
  azemailsender-cli config init --path ~/.config/azemailsender/config.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigInit(cmd, path)
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "./azemailsender.json", "Path for the configuration file")

	return cmd
}

func newConfigShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long: `Show the current configuration loaded from files and environment variables.

Examples:
  # Show current configuration
  azemailsender-cli config show

  # Show configuration from specific file
  azemailsender-cli config show --config ~/.config/azemailsender/config.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(cmd, args)
		},
	}

	return cmd
}

func newConfigEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Show environment variable examples",
		Long: `Show examples of environment variables that can be used for configuration.

Examples:
  # Show environment variable examples
  azemailsender-cli config env

  # Save environment variables to file
  azemailsender-cli config env > .env`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigEnv(cmd, args)
		},
	}

	return cmd
}

func runConfigInit(cmd *cobra.Command, path string) error {
	// Get flags from root command since these are persistent flags
	rootCmd := cmd.Root()
	debug, _ := rootCmd.PersistentFlags().GetBool("debug")
	quiet, _ := rootCmd.PersistentFlags().GetBool("quiet")
	jsonOutput, _ := rootCmd.PersistentFlags().GetBool("json")

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

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Get flags from root command since these are persistent flags
	rootCmd := cmd.Root()
	debug, _ := rootCmd.PersistentFlags().GetBool("debug")
	quiet, _ := rootCmd.PersistentFlags().GetBool("quiet")
	jsonOutput, _ := rootCmd.PersistentFlags().GetBool("json")

	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	// Load configuration
	configFile, _ := cmd.Flags().GetString("config")
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

func runConfigEnv(cmd *cobra.Command, args []string) error {
	// Get flags from root command since these are persistent flags
	rootCmd := cmd.Root()
	debug, _ := rootCmd.PersistentFlags().GetBool("debug")
	quiet, _ := rootCmd.PersistentFlags().GetBool("quiet")
	jsonOutput, _ := rootCmd.PersistentFlags().GetBool("json")

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
