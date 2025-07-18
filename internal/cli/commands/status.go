package commands

import (
	"fmt"
	"time"

	"github.com/groovy-sky/azemailsender"
	"github.com/groovy-sky/azemailsender/internal/cli/config"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/urfave/cli/v2"
)


// NewStatusCommand creates the status command
func NewStatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Check email status",
		Description: `Check the status of a previously sent email.

Examples:
  # Check status once
  azemailsender-cli status abc123def456

  # Check status and wait for completion
  azemailsender-cli status abc123def456 --wait

  # Check status with custom polling interval
  azemailsender-cli status abc123def456 --wait --poll-interval 10s --max-wait-time 2m`,
		ArgsUsage: "<message-id>",
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
				return fmt.Errorf("expected exactly one argument: message-id")
			}
			return runStatus(c, c.Args().Get(0))
		},
	}
}

func runStatus(c *cli.Context, messageID string) error {
	// Load configuration
	configFile := c.String("config")
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get flags and merge with configuration
	debug := c.Bool("debug")
	quiet := c.Bool("quiet")
	jsonOutput := c.Bool("json")

	// Merge authentication (CLI flags take precedence)
	endpoint := c.String("endpoint")
	if endpoint == "" {
		endpoint = cfg.Endpoint
	}
	accessKey := c.String("access-key")
	if accessKey == "" {
		accessKey = cfg.AccessKey
	}
	connectionString := c.String("connection-string")
	if connectionString == "" {
		connectionString = cfg.ConnectionString
	}

	// Create output formatter
	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	// Validate authentication
	hasAuth := false
	if connectionString != "" {
		hasAuth = true
	} else if endpoint != "" && accessKey != "" {
		hasAuth = true
	}

	if !hasAuth {
		err := fmt.Errorf("authentication required: provide either --connection-string or both --endpoint and --access-key")
		formatter.PrintError(err)
		return err
	}

	// Create email client
	clientOptions := &azemailsender.ClientOptions{
		Debug: debug,
	}

	var client *azemailsender.Client
	if connectionString != "" {
		client, err = azemailsender.NewClientFromConnectionString(connectionString, clientOptions)
		if err != nil {
			formatter.PrintError(err)
			return err
		}
	} else {
		client = azemailsender.NewClient(endpoint, accessKey, clientOptions)
	}

	formatter.PrintDebug("Checking status for message ID: %s", messageID)

	wait := c.Bool("wait")
	if wait {
		// Wait for completion
		formatter.PrintInfo("Waiting for email completion...")
		
		waitOptions := &azemailsender.WaitOptions{
			PollInterval: c.Duration("poll-interval"),
			MaxWaitTime:  c.Duration("max-wait-time"),
			OnStatusUpdate: func(status *azemailsender.StatusResponse) {
				if !quiet && !jsonOutput {
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

