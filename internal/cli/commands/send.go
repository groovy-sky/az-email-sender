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

// SendOptions holds options for the send command
type SendOptions struct {
	// Authentication
	Endpoint         string
	AccessKey        string
	ConnectionString string

	// Email content
	From     string
	To       []string
	Cc       []string
	Bcc      []string
	ReplyTo  string
	Subject  string
	Text     string
	HTML     string
	TextFile string
	HTMLFile string

	// Behavior
	Wait         bool
	PollInterval time.Duration
	MaxWaitTime  time.Duration

	// Output
	ConfigFile string
	Debug      bool
	Quiet      bool
	JSON       bool
}

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

func runSend(cmd *cobra.Command, opts *SendOptions) error {
	// Load configuration
	configFile, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with command-line flags and viper settings
	if err := mergeOptions(opts, cfg); err != nil {
		return err
	}

	// Create output formatter
	formatter := output.NewFormatter(opts.JSON, opts.Quiet, opts.Debug)

	// Validate options
	if err := validateSendOptions(opts); err != nil {
		formatter.PrintError(err)
		return err
	}

	// Handle content from stdin or files
	if err := loadContent(opts); err != nil {
		formatter.PrintError(err)
		return err
	}

	// Create email client
	client, err := createEmailClient(opts)
	if err != nil {
		formatter.PrintError(err)
		return err
	}

	// Build email message
	message, err := buildEmailMessage(client, opts)
	if err != nil {
		formatter.PrintError(err)
		return err
	}

	formatter.PrintDebug("Sending email to %s", output.FormatRecipients(opts.To))

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
	if opts.Wait {
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

		finalStatus, err := client.WaitForCompletion(response.ID, waitOptions)
		if err != nil {
			formatter.PrintError(fmt.Errorf("waiting for completion failed: %w", err))
			return err
		}

		return formatter.PrintStatusResponse(finalStatus)
	}

	return nil
}

func mergeOptions(opts *SendOptions, cfg *config.Config) error {
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

	// Merge email settings
	if opts.From == "" {
		opts.From = cfg.From
	}
	if opts.ReplyTo == "" {
		opts.ReplyTo = cfg.ReplyTo
	}

	return nil
}

func validateSendOptions(opts *SendOptions) error {
	var errors []string

	// Check authentication
	hasAuth := false
	if opts.ConnectionString != "" {
		hasAuth = true
	} else if opts.Endpoint != "" && opts.AccessKey != "" {
		hasAuth = true
	}

	if !hasAuth {
		errors = append(errors, "authentication required: provide either --connection-string or both --endpoint and --access-key")
	}

	// Check recipients
	if len(opts.To) == 0 && len(opts.Cc) == 0 && len(opts.Bcc) == 0 {
		errors = append(errors, "at least one recipient required (--to, --cc, or --bcc)")
	}

	// Check sender
	if opts.From == "" {
		errors = append(errors, "sender address required (--from)")
	}

	// Check subject
	if opts.Subject == "" {
		errors = append(errors, "subject required (--subject)")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}

func loadContent(opts *SendOptions) error {
	// Read from files if specified
	if opts.TextFile != "" {
		content, err := os.ReadFile(opts.TextFile)
		if err != nil {
			return fmt.Errorf("failed to read text file %s: %w", opts.TextFile, err)
		}
		opts.Text = string(content)
	}

	if opts.HTMLFile != "" {
		content, err := os.ReadFile(opts.HTMLFile)
		if err != nil {
			return fmt.Errorf("failed to read HTML file %s: %w", opts.HTMLFile, err)
		}
		opts.HTML = string(content)
	}

	// Read from stdin if no content provided
	if opts.Text == "" && opts.HTML == "" {
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

			opts.Text = content.String()
		}
	}

	// Validate content
	if opts.Text == "" && opts.HTML == "" {
		return fmt.Errorf("email content required: provide --text, --html, --text-file, --html-file, or pipe content to stdin")
	}

	return nil
}

func createEmailClient(opts *SendOptions) (*azemailsender.Client, error) {
	clientOptions := &azemailsender.ClientOptions{
		Debug: opts.Debug,
	}

	if opts.ConnectionString != "" {
		return azemailsender.NewClientFromConnectionString(opts.ConnectionString, clientOptions)
	}

	return azemailsender.NewClient(opts.Endpoint, opts.AccessKey, clientOptions), nil
}

func buildEmailMessage(client *azemailsender.Client, opts *SendOptions) (*azemailsender.EmailMessage, error) {
	builder := client.NewMessage().
		From(opts.From).
		Subject(opts.Subject)

	// Add recipients
	for _, to := range opts.To {
		builder = builder.To(to)
	}
	for _, cc := range opts.Cc {
		builder = builder.Cc(cc)
	}
	for _, bcc := range opts.Bcc {
		builder = builder.Bcc(bcc)
	}

	// Add reply-to if specified
	if opts.ReplyTo != "" {
		builder = builder.ReplyTo(opts.ReplyTo)
	}

	// Add content
	if opts.Text != "" {
		builder = builder.PlainText(opts.Text)
	}
	if opts.HTML != "" {
		builder = builder.HTML(opts.HTML)
	}

	return builder.Build()
}