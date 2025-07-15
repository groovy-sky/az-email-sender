package commands

import (
	"fmt"

	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	// Get global flags
	debug := viper.GetBool("debug")
	quiet := viper.GetBool("quiet")
	jsonOutput := viper.GetBool("json")

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