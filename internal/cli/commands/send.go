package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/groovy-sky/azemailsender"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/groovy-sky/azemailsender/internal/simpleconfig"
	"github.com/groovy-sky/azemailsender/internal/simplecli"
)

// NewSendCommand creates the send command
func NewSendCommand() *simplecli.Command {
	return &simplecli.Command{
		Name:        "send",
		Description: "Send an email",
		Usage:       "send [flags]",
		LongDesc: `Send an email using Azure Communication Services.

Examples:
  # Send a simple email
  azemailsender-cli send --from sender@example.com --to recipient@example.com --subject "Hello" --text "Hello World"

  # Send HTML email with multiple recipients
  azemailsender-cli send --from sender@example.com --to user1@example.com --to user2@example.com --cc manager@example.com --subject "Report" --html "<h1>Monthly Report</h1>"

  # Send email and wait for completion
  azemailsender-cli send --from sender@example.com --to recipient@example.com --subject "Hello" --text "Hello World" --wait

  # Read content from stdin
  echo "Hello from stdin" | azemailsender-cli send --from sender@example.com --to recipient@example.com --subject "Stdin Test"

  # Read content from file
  azemailsender-cli send --from sender@example.com --to recipient@example.com --subject "File Test" --text-file message.txt`,
		Run: runSend,
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
			// Email content flags
			{
				Name:        "from",
				Short:       "f",
				Description: "Sender email address",
				Value:       "",
				Required:    true,
				EnvVar:      "AZURE_EMAIL_FROM",
			},
			{
				Name:        "to",
				Short:       "t",
				Description: "To recipients (can be repeated)",
				Value:       []string{},
			},
			{
				Name:        "cc",
				Description: "CC recipients (can be repeated)",
				Value:       []string{},
			},
			{
				Name:        "bcc",
				Description: "BCC recipients (can be repeated)",
				Value:       []string{},
			},
			{
				Name:        "reply-to",
				Description: "Reply-to email address",
				Value:       "",
				EnvVar:      "AZURE_EMAIL_REPLY_TO",
			},
			{
				Name:        "subject",
				Short:       "s",
				Description: "Email subject",
				Value:       "",
				Required:    true,
			},
			{
				Name:        "text",
				Description: "Plain text email content",
				Value:       "",
			},
			{
				Name:        "html",
				Description: "HTML email content",
				Value:       "",
			},
			{
				Name:        "text-file",
				Description: "Read plain text content from file",
				Value:       "",
			},
			{
				Name:        "html-file",
				Description: "Read HTML content from file",
				Value:       "",
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

func runSend(ctx *simplecli.Context) error {
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

	// Get values from flags and config
	endpoint := ctx.GetString("endpoint")
	accessKey := ctx.GetString("access-key")
	connectionString := ctx.GetString("connection-string")
	from := ctx.GetString("from")
	to := ctx.GetStringSlice("to")
	cc := ctx.GetStringSlice("cc")
	bcc := ctx.GetStringSlice("bcc")
	replyTo := ctx.GetString("reply-to")
	subject := ctx.GetString("subject")
	text := ctx.GetString("text")
	html := ctx.GetString("html")
	textFile := ctx.GetString("text-file")
	htmlFile := ctx.GetString("html-file")
	wait := ctx.GetBool("wait")

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
	if from == "" {
		from = config.From
	}
	if replyTo == "" {
		replyTo = config.ReplyTo
	}

	// Validate authentication
	hasAuth := false
	if connectionString != "" {
		hasAuth = true
	} else if endpoint != "" && accessKey != "" {
		hasAuth = true
	}

	if !hasAuth {
		return fmt.Errorf("authentication required: provide either --connection-string or both --endpoint and --access-key")
	}

	// Check recipients
	if len(to) == 0 && len(cc) == 0 && len(bcc) == 0 {
		return fmt.Errorf("at least one recipient required (--to, --cc, or --bcc)")
	}

	// Check sender
	if from == "" {
		return fmt.Errorf("sender address required (--from)")
	}

	// Check subject
	if subject == "" {
		return fmt.Errorf("subject required (--subject)")
	}

	// Handle content from files
	if textFile != "" {
		content, err := os.ReadFile(textFile)
		if err != nil {
			return fmt.Errorf("failed to read text file %s: %w", textFile, err)
		}
		text = string(content)
	}

	if htmlFile != "" {
		content, err := os.ReadFile(htmlFile)
		if err != nil {
			return fmt.Errorf("failed to read HTML file %s: %w", htmlFile, err)
		}
		html = string(content)
	}

	// Read from stdin if no content provided
	if text == "" && html == "" {
		stat, err := os.Stdin.Stat()
		if err != nil {
			return fmt.Errorf("failed to check stdin: %w", err)
		}

		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped to stdin
			reader := bufio.NewReader(os.Stdin)
			var content strings.Builder

			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						if line != "" {
							content.WriteString(line)
						}
						break
					}
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				content.WriteString(line)
			}

			text = content.String()
		}
	}

	// Validate content
	if text == "" && html == "" {
		return fmt.Errorf("email content required: provide --text, --html, --text-file, --html-file, or pipe content to stdin")
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
		return err
	}

	// Build email message
	builder := client.NewMessage().
		From(from).
		Subject(subject)

	// Add recipients
	for _, recipient := range to {
		builder = builder.To(recipient)
	}
	for _, recipient := range cc {
		builder = builder.Cc(recipient)
	}
	for _, recipient := range bcc {
		builder = builder.Bcc(recipient)
	}

	// Add reply-to if specified
	if replyTo != "" {
		builder = builder.ReplyTo(replyTo)
	}

	// Add content
	if text != "" {
		builder = builder.PlainText(text)
	}
	if html != "" {
		builder = builder.HTML(html)
	}

	message, err := builder.Build()
	if err != nil {
		return err
	}

	formatter.PrintDebug("Sending email to %s", output.FormatRecipients(to))

	// Send email
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	// Print send response
	if err := formatter.PrintSendResponse(response); err != nil {
		return err
	}

	// Wait for completion if requested
	if wait {
		formatter.PrintInfo("Waiting for email completion...")

		// Parse duration strings
		pollIntervalStr := ctx.GetString("poll-interval")
		maxWaitTimeStr := ctx.GetString("max-wait-time")

		// Use config values if defaults
		if pollIntervalStr == "5s" {
			pollIntervalStr = config.PollInterval
		}
		if maxWaitTimeStr == "5m" {
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

		waitOptions := &azemailsender.WaitOptions{
			PollInterval: pollInterval,
			MaxWaitTime:  maxWaitTime,
			OnStatusUpdate: func(status *azemailsender.StatusResponse) {
				if !quiet && !jsonOutput {
					fmt.Printf("Status: %s\n", status.Status)
				}
			},
		}

		finalStatus, err := client.WaitForCompletion(response.ID, waitOptions)
		if err != nil {
			return fmt.Errorf("waiting for completion failed: %w", err)
		}

		return formatter.PrintStatusResponse(finalStatus)
	}

	return nil
}