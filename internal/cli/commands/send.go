package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/groovy-sky/azemailsender"
	"github.com/groovy-sky/azemailsender/internal/cli/config"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
	"github.com/urfave/cli/v2"
)

// NewSendCommand creates the send command
func NewSendCommand() *cli.Command {
	return &cli.Command{
		Name:  "send",
		Usage: "Send an email",
		Description: `Send an email using Azure Communication Services.

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
		Flags: []cli.Flag{
			// Authentication flags
			&cli.StringFlag{
				Name:     "endpoint",
				Aliases:  []string{"e"},
				Usage:    "Azure Communication Services endpoint",
				EnvVars:  []string{"AZURE_EMAIL_ENDPOINT"},
			},
			&cli.StringFlag{
				Name:     "access-key",
				Aliases:  []string{"k"},
				Usage:    "Access key for authentication",
				EnvVars:  []string{"AZURE_EMAIL_ACCESS_KEY"},
			},
			&cli.StringFlag{
				Name:     "connection-string",
				Usage:    "Connection string for authentication",
				EnvVars:  []string{"AZURE_EMAIL_CONNECTION_STRING"},
			},
			// Email content flags
			&cli.StringFlag{
				Name:     "from",
				Aliases:  []string{"f"},
				Usage:    "Sender email address",
				EnvVars:  []string{"AZURE_EMAIL_FROM"},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:    "to",
				Aliases: []string{"t"},
				Usage:   "To recipients (can be repeated)",
			},
			&cli.StringSliceFlag{
				Name:  "cc",
				Usage: "CC recipients (can be repeated)",
			},
			&cli.StringSliceFlag{
				Name:  "bcc",
				Usage: "BCC recipients (can be repeated)",
			},
			&cli.StringFlag{
				Name:    "reply-to",
				Usage:   "Reply-to email address",
				EnvVars: []string{"AZURE_EMAIL_REPLY_TO"},
			},
			&cli.StringFlag{
				Name:     "subject",
				Aliases:  []string{"s"},
				Usage:    "Email subject",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "text",
				Usage: "Plain text email content",
			},
			&cli.StringFlag{
				Name:  "html",
				Usage: "HTML email content",
			},
			&cli.StringFlag{
				Name:  "text-file",
				Usage: "Read plain text content from file",
			},
			&cli.StringFlag{
				Name:  "html-file",
				Usage: "Read HTML content from file",
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
			return runSend(c)
		},
	}
}

func runSend(c *cli.Context) error {
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

	// Create output formatter
	formatter := output.NewFormatter(jsonOutput, quiet, debug)

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

	// Get email content flags
	from := c.String("from")
	to := c.StringSlice("to")
	cc := c.StringSlice("cc")
	bcc := c.StringSlice("bcc")
	replyTo := c.String("reply-to")
	subject := c.String("subject")
	text := c.String("text")
	html := c.String("html")
	textFile := c.String("text-file")
	htmlFile := c.String("html-file")

	// Validate recipients
	if len(to) == 0 {
		err := fmt.Errorf("at least one recipient is required (use --to)")
		formatter.PrintError(err)
		return err
	}

	// Handle content from stdin or files
	if err := loadContentForCli(&text, &html, textFile, htmlFile); err != nil {
		formatter.PrintError(err)
		return err
	}

	// Validate content
	if text == "" && html == "" {
		err := fmt.Errorf("email content is required: provide --text, --html, --text-file, --html-file, or pipe content to stdin")
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

	// Build email message
	builder := client.NewMessage().From(from).Subject(subject)

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
		formatter.PrintError(err)
		return err
	}

	formatter.PrintDebug("Sending email to %s", output.FormatRecipients(to))

	// Send email
	response, err := client.Send(message)
	if err != nil {
		formatter.PrintError(err)
		return err
	}

	// Print send response
	if err := formatter.PrintSendResponse(response); err != nil {
		return err
	}

	// Wait for completion if requested
	wait := c.Bool("wait")
	if wait {
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

		finalStatus, err := client.WaitForCompletion(response.ID, waitOptions)
		if err != nil {
			formatter.PrintError(fmt.Errorf("waiting for completion failed: %w", err))
			return err
		}

		return formatter.PrintStatusResponse(finalStatus)
	}

	return nil
}

// loadContentForCli handles loading content from files or stdin for urfave/cli
func loadContentForCli(text, html *string, textFile, htmlFile string) error {
	// Read from files if specified
	if textFile != "" {
		content, err := os.ReadFile(textFile)
		if err != nil {
			return fmt.Errorf("failed to read text file %s: %w", textFile, err)
		}
		*text = string(content)
	}

	if htmlFile != "" {
		content, err := os.ReadFile(htmlFile)
		if err != nil {
			return fmt.Errorf("failed to read HTML file %s: %w", htmlFile, err)
		}
		*html = string(content)
	}

	// Read from stdin if no content provided
	if *text == "" && *html == "" {
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

			*text = content.String()
		}
	}

	return nil
}

