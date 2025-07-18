package app

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/groovy-sky/azemailsender"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
)

// StatusOptions holds options for the status command
type StatusOptions struct {
	// Authentication
	Endpoint         string
	AccessKey        string
	ConnectionString string

	// Behavior
	Wait         bool
	PollInterval time.Duration
	MaxWaitTime  time.Duration

	// Output (from global flags)
	Debug bool
	Quiet bool
	JSON  bool
}

// runStatus executes the status command
func (a *App) runStatus(args []string, globalFlags *GlobalFlags) int {
	opts := &StatusOptions{}

	// Create flag set for status command
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.Usage = func() {
		a.printStatusUsage()
	}

	// Authentication flags
	fs.StringVar(&opts.Endpoint, "endpoint", "", "Azure Communication Services endpoint")
	fs.StringVar(&opts.Endpoint, "e", "", "Azure Communication Services endpoint")
	fs.StringVar(&opts.AccessKey, "access-key", "", "Access key for authentication")
	fs.StringVar(&opts.AccessKey, "k", "", "Access key for authentication")
	fs.StringVar(&opts.ConnectionString, "connection-string", "", "Connection string for authentication")

	// Behavior flags
	fs.BoolVar(&opts.Wait, "wait", false, "Wait for email completion")
	fs.BoolVar(&opts.Wait, "w", false, "Wait for email completion")
	fs.DurationVar(&opts.PollInterval, "poll-interval", 5*time.Second, "Status polling interval (when --wait is used)")
	fs.DurationVar(&opts.MaxWaitTime, "max-wait-time", 5*time.Minute, "Maximum wait time (when --wait is used)")

	// Parse flags
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	// Check for message ID argument
	if fs.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: exactly one message ID required\n")
		a.printStatusUsage()
		return 1
	}

	messageID := fs.Arg(0)

	// Load configuration
	cfg, err := LoadConfig(globalFlags.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		return 1
	}

	// Merge options with configuration and global flags
	mergeConfigWithStatusOptions(opts, cfg, globalFlags)

	// Create output formatter
	formatter := output.NewFormatter(opts.JSON, opts.Quiet, opts.Debug)

	// Validate options
	if err := validateStatusOptions(opts); err != nil {
		formatter.PrintError(err)
		return 1
	}

	// Create email client
	client, err := createStatusEmailClient(opts)
	if err != nil {
		formatter.PrintError(err)
		return 1
	}

	formatter.PrintDebug("Checking status for message ID: %s", messageID)

	if opts.Wait {
		// Wait for completion
		formatter.PrintInfo("Waiting for email completion...")
		
		waitOptions := &azemailsender.WaitOptions{
			PollInterval: opts.PollInterval,
			MaxWaitTime:  opts.MaxWaitTime,
			OnStatusUpdate: func(status *azemailsender.StatusResponse) {
				if !opts.Quiet && !opts.JSON {
					fmt.Printf("Status: %s\n", status.Status)
				}
			},
		}

		finalStatus, err := client.WaitForCompletion(messageID, waitOptions)
		if err != nil {
			formatter.PrintError(fmt.Errorf("waiting for completion failed: %w", err))
			return 1
		}

		if err := formatter.PrintStatusResponse(finalStatus); err != nil {
			return 1
		}
	} else {
		// Check status once
		status, err := client.GetStatus(messageID)
		if err != nil {
			formatter.PrintError(err)
			return 1
		}

		if err := formatter.PrintStatusResponse(status); err != nil {
			return 1
		}
	}

	return 0
}

// mergeConfigWithStatusOptions merges configuration with status command options
func mergeConfigWithStatusOptions(opts *StatusOptions, cfg *Config, globalFlags *GlobalFlags) {
	// Global flags take precedence
	opts.Debug = globalFlags.Debug || cfg.Debug
	opts.Quiet = globalFlags.Quiet || cfg.Quiet
	opts.JSON = globalFlags.JSON || cfg.JSON

	// Merge authentication (CLI flags take precedence)
	if opts.Endpoint == "" {
		opts.Endpoint = cfg.Endpoint
	}
	if opts.AccessKey == "" {
		opts.AccessKey = cfg.AccessKey
	}
	if opts.ConnectionString == "" {
		opts.ConnectionString = cfg.ConnectionString
	}
}

// validateStatusOptions validates status command options
func validateStatusOptions(opts *StatusOptions) error {
	// Check authentication
	hasAuth := false
	if opts.ConnectionString != "" {
		hasAuth = true
	} else if opts.Endpoint != "" && opts.AccessKey != "" {
		hasAuth = true
	}

	if !hasAuth {
		return fmt.Errorf("authentication required: provide either --connection-string or both --endpoint and --access-key")
	}

	return nil
}

// createStatusEmailClient creates an email client for status checking
func createStatusEmailClient(opts *StatusOptions) (*azemailsender.Client, error) {
	clientOptions := &azemailsender.ClientOptions{
		Debug: opts.Debug,
	}

	if opts.ConnectionString != "" {
		return azemailsender.NewClientFromConnectionString(opts.ConnectionString, clientOptions)
	}

	return azemailsender.NewClient(opts.Endpoint, opts.AccessKey, clientOptions), nil
}

// printStatusUsage prints usage information for the status command
func (a *App) printStatusUsage() {
	fmt.Printf(`Check the status of a previously sent email.

Examples:
  # Check status once
  %s status abc123def456

  # Check status and wait for completion
  %s status abc123def456 --wait

  # Check status with custom polling interval
  %s status abc123def456 --wait --poll-interval 10s --max-wait-time 2m

Usage:
  %s status <message-id> [flags]

Flags:
  -k, --access-key string          Access key for authentication
      --connection-string string   Connection string for authentication
  -e, --endpoint string            Azure Communication Services endpoint
  -h, --help                       help for status
      --max-wait-time duration     Maximum wait time (when --wait is used) (default 5m0s)
      --poll-interval duration     Status polling interval (when --wait is used) (default 5s)
  -w, --wait                       Wait for email completion

Global Flags:
  -c, --config string   Configuration file path
  -d, --debug           Enable debug logging
  -j, --json            Output in JSON format
  -q, --quiet           Suppress output except errors
`, a.name, a.name, a.name, a.name)
}