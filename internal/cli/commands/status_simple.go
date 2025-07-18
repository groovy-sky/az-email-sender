package commands

import (
	"fmt"
	"time"

	"github.com/groovy-sky/azemailsender"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/groovy-sky/azemailsender/internal/simpleconfig"
	"github.com/groovy-sky/azemailsender/internal/simplecli"
)

// NewSimpleStatusCommand creates the status command for simple CLI
func NewSimpleStatusCommand() *simplecli.Command {
	return &simplecli.Command{
		Name:        "status",
		Description: "Check email status",
		Usage:       "status <message-id> [flags]",
		LongDesc: `Check the status of a previously sent email.

Examples:
  # Check status once
  azemailsender-cli status abc123def456

  # Check status and wait for completion
  azemailsender-cli status abc123def456 --wait

  # Check status with custom polling interval
  azemailsender-cli status abc123def456 --wait --poll-interval 10s --max-wait-time 2m`,
		Run: runSimpleStatus,
		Flags: []*simplecli.Flag{
			// Authentication flags
			{
				Name:        "endpoint",
				Short:       "e",
				Description: "Azure Communication Services endpoint",
				Value:       "",
				EnvVar:      "AZURE_EMAIL_ENDPOINT",
			},
			{
				Name:        "access-key",
				Short:       "k",
				Description: "Access key for authentication",
				Value:       "",
				EnvVar:      "AZURE_EMAIL_ACCESS_KEY",
			},
			{
				Name:        "connection-string",
				Description: "Connection string for authentication",
				Value:       "",
				EnvVar:      "AZURE_EMAIL_CONNECTION_STRING",
			},
			// Behavior flags
			{
				Name:        "wait",
				Short:       "w",
				Description: "Wait for email completion",
				Value:       false,
				EnvVar:      "AZURE_EMAIL_WAIT",
			},
			{
				Name:        "poll-interval",
				Description: "Status polling interval (when --wait is used)",
				Value:       "5s",
				EnvVar:      "AZURE_EMAIL_POLL_INTERVAL",
			},
			{
				Name:        "max-wait-time",
				Description: "Maximum wait time (when --wait is used)",
				Value:       "5m",
				EnvVar:      "AZURE_EMAIL_MAX_WAIT_TIME",
			},
		},
	}
}

func runSimpleStatus(ctx *simplecli.Context) error {
	// Check if message ID is provided
	if len(ctx.Args) == 0 {
		return fmt.Errorf("message ID required")
	}
	messageID := ctx.Args[0]

	// Load configuration
	configFile := ctx.GetString("config")
	config, err := simpleconfig.LoadConfig(configFile, ctx.Flags)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create output formatter
	debug := ctx.GetBool("debug")
	quiet := ctx.GetBool("quiet")
	jsonOutput := ctx.GetBool("json")
	formatter := output.NewFormatter(jsonOutput, quiet, debug)

	// Validate authentication
	endpoint := ctx.GetString("endpoint")
	accessKey := ctx.GetString("access-key")
	connectionString := ctx.GetString("connection-string")

	// Use config values if not provided via flags
	if endpoint == "" {
		endpoint = config.Endpoint
	}
	if accessKey == "" {
		accessKey = config.AccessKey
	}
	if connectionString == "" {
		connectionString = config.ConnectionString
	}

	hasAuth := false
	if connectionString != "" {
		hasAuth = true
	} else if endpoint != "" && accessKey != "" {
		hasAuth = true
	}

	if !hasAuth {
		return fmt.Errorf("authentication required: provide either --connection-string or both --endpoint and --access-key")
	}

	// Create email client
	clientOptions := &azemailsender.ClientOptions{
		Debug: debug,
	}

	var client *azemailsender.Client
	if connectionString != "" {
		client, err = azemailsender.NewClientFromConnectionString(connectionString, clientOptions)
	} else {
		client = azemailsender.NewClient(endpoint, accessKey, clientOptions)
	}
	if err != nil {
		formatter.PrintError(err)
		return err
	}

	formatter.PrintDebug("Checking status for message ID: %s", messageID)

	wait := ctx.GetBool("wait")
	if wait {
		// Parse duration strings
		pollIntervalStr := ctx.GetString("poll-interval")
		maxWaitTimeStr := ctx.GetString("max-wait-time")

		// Use config values if not provided via flags
		if pollIntervalStr == "5s" { // default value
			pollIntervalStr = config.PollInterval
		}
		if maxWaitTimeStr == "5m" { // default value
			maxWaitTimeStr = config.MaxWaitTime
		}

		pollInterval, err := time.ParseDuration(pollIntervalStr)
		if err != nil {
			return fmt.Errorf("invalid poll-interval: %w", err)
		}

		maxWaitTime, err := time.ParseDuration(maxWaitTimeStr)
		if err != nil {
			return fmt.Errorf("invalid max-wait-time: %w", err)
		}

		// Wait for completion
		formatter.PrintInfo("Waiting for email completion...")

		waitOptions := &azemailsender.WaitOptions{
			PollInterval: pollInterval,
			MaxWaitTime:  maxWaitTime,
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