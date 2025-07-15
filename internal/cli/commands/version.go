package commands

import (
	"fmt"

	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand(version, commit, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Show version, build commit, and build date information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(cmd, version, commit, date)
		},
	}

	return cmd
}

func runVersion(cmd *cobra.Command, version, commit, date string) error {
	// Get flags from root command (persistent flags)
	rootCmd := cmd.Root()
	debug, _ := rootCmd.PersistentFlags().GetBool("debug")
	quiet, _ := rootCmd.PersistentFlags().GetBool("quiet")
	jsonOutput, _ := rootCmd.PersistentFlags().GetBool("json")

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