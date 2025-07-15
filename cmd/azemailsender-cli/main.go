package main

import (
	"fmt"
	"os"

	"github.com/groovy-sky/azemailsender/internal/cli/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "azemailsender-cli",
		Short: "Azure Communication Services Email CLI",
		Long: `A command-line interface for sending emails using Azure Communication Services.
Supports multiple authentication methods, flexible recipient management,
and both plain text and HTML email content.`,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Bind flags to viper
			viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
			viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
			viper.BindPFlag("quiet", cmd.PersistentFlags().Lookup("quiet"))
			viper.BindPFlag("json", cmd.PersistentFlags().Lookup("json"))
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file path")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress output except errors")
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output in JSON format")

	// Add commands
	rootCmd.AddCommand(commands.NewSendCommand())
	rootCmd.AddCommand(commands.NewStatusCommand())
	rootCmd.AddCommand(commands.NewConfigCommand())
	rootCmd.AddCommand(commands.NewVersionCommand(version, commit, date))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}