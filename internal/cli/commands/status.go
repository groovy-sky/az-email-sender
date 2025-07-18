package commands

import (
	"fmt"
	"time"

	"github.com/groovy-sky/azemailsender"
	"github.com/groovy-sky/azemailsender/internal/cli/config"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/urfave/cli/v2"
	"github.com/spf13/viper"
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

	// Output
	Debug bool
	Quiet bool
	JSON  bool
}

// NewStatusCommand creates the status command
func NewStatusCommand() *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "Check email status",
		ArgsUsage: "<message-id>",
		Description: `Check the status of a previously sent email.

Examples:
  # Check status once
  azemailsender-cli status abc123def456

  # Check status and wait for completion
  azemailsender-cli status abc123def456 --wait

  # Check status with custom polling interval
  azemailsender-cli status abc123def456 --wait --poll-interval 10s --max-wait-time 2m`,
		Flags: []cli.Flag{
			// Authentication flags
			&cli.StringFlag{
				Name:    "endpoint",
				Aliases: []string{"e"},
				Usage:   "Azure Communication Services endpoint",
				EnvVars: []string{"AZURE_EMAIL_ENDPOINT"},
			},
			&cli.StringFlag{
				Name:    "access-key",
				Aliases: []string{"k"},
				Usage:   "Access key for authentication",
				EnvVars: []string{"AZURE_EMAIL_ACCESS_KEY"},
			},
			&cli.StringFlag{
				Name:    "connection-string",
				Usage:   "Connection string for authentication",
				EnvVars: []string{"AZURE_EMAIL_CONNECTION_STRING"},
			},
			// Behavior flags
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for email completion",
			},
			&cli.DurationFlag{
				Name:  "poll-interval",
				Usage: "Status polling interval (when --wait is used)",
				Value: 5 * time.Second,
			},
			&cli.DurationFlag{
				Name:  "max-wait-time",
				Usage: "Maximum wait time (when --wait is used)",
				Value: 5 * time.Minute,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return fmt.Errorf("exactly one message ID required")
			}
			messageID := c.Args().First()
			return runStatus(c, messageID)
		},
	}
}

func runStatus(c *cli.Context, messageID string) error {
	// Create StatusOptions from context
	opts := &StatusOptions{
		Endpoint:         c.String("endpoint"),
		AccessKey:        c.String("access-key"),
		ConnectionString: c.String("connection-string"),
		Wait:             c.Bool("wait"),
		PollInterval:     c.Duration("poll-interval"),
		MaxWaitTime:      c.Duration("max-wait-time"),
		Debug:            c.Bool("debug"),
		Quiet:            c.Bool("quiet"),
		JSON:             c.Bool("json"),
	}

	// Load configuration
	configFile := c.String("config")
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with command-line flags
	if err := mergeStatusOptions(opts, cfg); err != nil {
		return err
	}

	// Create output formatter
	formatter := output.NewFormatter(opts.JSON, opts.Quiet, opts.Debug)

	// Validate options
	if err := validateStatusOptions(opts); err != nil {
		formatter.PrintError(err)
		return err
	}

	// Create email client
	client, err := createStatusEmailClient(opts)
	if err != nil {
		formatter.PrintError(err)
		return err
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
			return err
		}

		return formatter.PrintStatusResponse(finalStatus)
	} else {
		// Check status once
		status, err := client.GetStatus(messageID)
		if err != nil {
			formatter.PrintError(err)
			return err
		}

		return formatter.PrintStatusResponse(status)
	}
}

func mergeStatusOptions(opts *StatusOptions, cfg *config.Config) error {
	// Get global flags from viper
	opts.Debug = viper.GetBool("debug") || opts.Debug
	opts.Quiet = viper.GetBool("quiet") || opts.Quiet
	opts.JSON = viper.GetBool("json") || opts.JSON

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

	return nil
}

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

func createStatusEmailClient(opts *StatusOptions) (*azemailsender.Client, error) {
	clientOptions := &azemailsender.ClientOptions{
		Debug: opts.Debug,
	}

	if opts.ConnectionString != "" {
		return azemailsender.NewClientFromConnectionString(opts.ConnectionString, clientOptions)
	}

	return azemailsender.NewClient(opts.Endpoint, opts.AccessKey, clientOptions), nil
}