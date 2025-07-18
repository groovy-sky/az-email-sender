package app

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/groovy-sky/azemailsender"
	"github.com/groovy-sky/azemailsender/internal/cli/output"
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

	// Output (from global flags)
	Debug bool
	Quiet bool
	JSON  bool
}

// runSend executes the send command
func (a *App) runSend(args []string, globalFlags *GlobalFlags) int {
	opts := &SendOptions{}
	
	// Create flag set for send command
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	fs.Usage = func() {
		a.printSendUsage()
	}

	// Authentication flags
	fs.StringVar(&opts.Endpoint, "endpoint", "", "Azure Communication Services endpoint")
	fs.StringVar(&opts.Endpoint, "e", "", "Azure Communication Services endpoint")
	fs.StringVar(&opts.AccessKey, "access-key", "", "Access key for authentication")
	fs.StringVar(&opts.AccessKey, "k", "", "Access key for authentication")
	fs.StringVar(&opts.ConnectionString, "connection-string", "", "Connection string for authentication")

	// Email content flags
	fs.StringVar(&opts.From, "from", "", "Sender email address")
	fs.StringVar(&opts.From, "f", "", "Sender email address")
	toFlags := newStringSliceFlag()
	fs.Var(toFlags, "to", "To recipients (can be repeated)")
	fs.Var(toFlags, "t", "To recipients (can be repeated)")
	ccFlags := newStringSliceFlag()
	fs.Var(ccFlags, "cc", "CC recipients (can be repeated)")
	bccFlags := newStringSliceFlag()
	fs.Var(bccFlags, "bcc", "BCC recipients (can be repeated)")
	fs.StringVar(&opts.ReplyTo, "reply-to", "", "Reply-to email address")
	fs.StringVar(&opts.Subject, "subject", "", "Email subject")
	fs.StringVar(&opts.Subject, "s", "", "Email subject")
	fs.StringVar(&opts.Text, "text", "", "Plain text email content")
	fs.StringVar(&opts.HTML, "html", "", "HTML email content")
	fs.StringVar(&opts.TextFile, "text-file", "", "Read plain text content from file")
	fs.StringVar(&opts.HTMLFile, "html-file", "", "Read HTML content from file")

	// Behavior flags
	fs.BoolVar(&opts.Wait, "wait", false, "Wait for email completion")
	fs.BoolVar(&opts.Wait, "w", false, "Wait for email completion")
	fs.DurationVar(&opts.PollInterval, "poll-interval", 5*time.Second, "Status polling interval (when --wait is used)")
	fs.DurationVar(&opts.MaxWaitTime, "max-wait-time", 5*time.Minute, "Maximum wait time (when --wait is used)")

	// Parse flags
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	// Convert string slices
	opts.To = []string(*toFlags)
	opts.Cc = []string(*ccFlags)
	opts.Bcc = []string(*bccFlags)

	// Load configuration
	cfg, err := LoadConfig(globalFlags.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		return 1
	}

	// Merge options with configuration and global flags
	mergeConfigWithSendOptions(opts, cfg, globalFlags)

	// Create output formatter
	formatter := output.NewFormatter(opts.JSON, opts.Quiet, opts.Debug)

	// Validate options
	if err := validateSendOptions(opts); err != nil {
		formatter.PrintError(err)
		return 1
	}

	// Handle content from stdin or files
	if err := loadSendContent(opts); err != nil {
		formatter.PrintError(err)
		return 1
	}

	// Create email client
	client, err := createSendEmailClient(opts)
	if err != nil {
		formatter.PrintError(err)
		return 1
	}

	// Build email message
	message, err := buildSendEmailMessage(client, opts)
	if err != nil {
		formatter.PrintError(err)
		return 1
	}

	formatter.PrintDebug("Sending email to %s", output.FormatRecipients(opts.To))

	// Send email
	response, err := client.Send(message)
	if err != nil {
		formatter.PrintError(err)
		return 1
	}

	// Print send response
	if err := formatter.PrintSendResponse(response); err != nil {
		return 1
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
			return 1
		}

		if err := formatter.PrintStatusResponse(finalStatus); err != nil {
			return 1
		}
	}

	return 0
}

// mergeConfigWithSendOptions merges configuration with command options
func mergeConfigWithSendOptions(opts *SendOptions, cfg *Config, globalFlags *GlobalFlags) {
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

	// Merge email settings
	if opts.From == "" {
		opts.From = cfg.From
	}
	if opts.ReplyTo == "" {
		opts.ReplyTo = cfg.ReplyTo
	}
}

// validateSendOptions validates send command options
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

// loadSendContent loads email content from files or stdin
func loadSendContent(opts *SendOptions) error {
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

// createSendEmailClient creates an email client for sending
func createSendEmailClient(opts *SendOptions) (*azemailsender.Client, error) {
	clientOptions := &azemailsender.ClientOptions{
		Debug: opts.Debug,
	}

	if opts.ConnectionString != "" {
		return azemailsender.NewClientFromConnectionString(opts.ConnectionString, clientOptions)
	}

	return azemailsender.NewClient(opts.Endpoint, opts.AccessKey, clientOptions), nil
}

// buildSendEmailMessage builds an email message
func buildSendEmailMessage(client *azemailsender.Client, opts *SendOptions) (*azemailsender.EmailMessage, error) {
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

// printSendUsage prints usage information for the send command
func (a *App) printSendUsage() {
	fmt.Printf(`Send an email using Azure Communication Services.

Examples:
  # Send a simple email
  %s send --from sender@example.com --to recipient@example.com --subject "Hello" --text "Hello World"

  # Send HTML email with multiple recipients
  %s send --from sender@example.com --to user1@example.com --to user2@example.com --cc manager@example.com --subject "Report" --html "<h1>Monthly Report</h1>"

  # Send email and wait for completion
  %s send --from sender@example.com --to recipient@example.com --subject "Hello" --text "Hello World" --wait

  # Read content from stdin
  echo "Hello from stdin" | %s send --from sender@example.com --to recipient@example.com --subject "Stdin Test"

  # Read content from file
  %s send --from sender@example.com --to recipient@example.com --subject "File Test" --text-file message.txt

Usage:
  %s send [flags]

Flags:
  -k, --access-key string          Access key for authentication
      --bcc strings                BCC recipients (can be repeated)
      --cc strings                 CC recipients (can be repeated)
      --connection-string string   Connection string for authentication
  -e, --endpoint string            Azure Communication Services endpoint
  -f, --from string                Sender email address
  -h, --help                       help for send
      --html string                HTML email content
      --html-file string           Read HTML content from file
      --max-wait-time duration     Maximum wait time (when --wait is used) (default 5m0s)
      --poll-interval duration     Status polling interval (when --wait is used) (default 5s)
      --reply-to string            Reply-to email address
  -s, --subject string             Email subject
      --text string                Plain text email content
      --text-file string           Read plain text content from file
  -t, --to strings                 To recipients (can be repeated)
  -w, --wait                       Wait for email completion

Global Flags:
  -c, --config string   Configuration file path
  -d, --debug           Enable debug logging
  -j, --json            Output in JSON format
  -q, --quiet           Suppress output except errors
`, a.name, a.name, a.name, a.name, a.name, a.name)
}