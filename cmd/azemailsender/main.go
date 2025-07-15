package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/groovy-sky/azemailsender"
)

const (
	exitSuccess = 0
	exitError   = 1
)

func main() {
	// Define command line flags
	var (
		to      = flag.String("to", "", "Email recipient address (required)")
		body    = flag.String("body", "", "Email body content (can be piped from stdin)")
		subject = flag.String("subject", "", "Email subject line (if not provided, extract from first line of body)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Simple CLI wrapper for Azure Email Sender\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  AZURE_EMAIL_ENDPOINT    Azure Communication Services endpoint URL\n")
		fmt.Fprintf(os.Stderr, "  AZURE_EMAIL_ACCESS_KEY  Azure access key for authentication\n")
		fmt.Fprintf(os.Stderr, "  AZURE_EMAIL_FROM        Sender email address\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --to \"user@example.com\" --subject \"Hello\" --body \"This is a test message\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  echo \"Hello World\" | %s --to \"user@example.com\" --subject \"Greeting\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --to \"user@example.com\" --body \"Subject: Important Message\\nThis is the email content\"\n", os.Args[0])
	}

	flag.Parse()

	// Validate required flags
	if *to == "" {
		fmt.Fprintf(os.Stderr, "Error: --to flag is required\n")
		flag.Usage()
		os.Exit(exitError)
	}

	// Get environment variables
	endpoint := os.Getenv("AZURE_EMAIL_ENDPOINT")
	accessKey := os.Getenv("AZURE_EMAIL_ACCESS_KEY")
	fromEmail := os.Getenv("AZURE_EMAIL_FROM")

	// Validate environment variables
	if endpoint == "" {
		fmt.Fprintf(os.Stderr, "Error: AZURE_EMAIL_ENDPOINT environment variable is required\n")
		os.Exit(exitError)
	}
	if accessKey == "" {
		fmt.Fprintf(os.Stderr, "Error: AZURE_EMAIL_ACCESS_KEY environment variable is required\n")
		os.Exit(exitError)
	}
	if fromEmail == "" {
		fmt.Fprintf(os.Stderr, "Error: AZURE_EMAIL_FROM environment variable is required\n")
		os.Exit(exitError)
	}

	// Get email body content
	emailBody := *body
	if emailBody == "" {
		// Read from stdin if body not provided
		content, err := readFromStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(exitError)
		}
		emailBody = content
	}

	if emailBody == "" {
		fmt.Fprintf(os.Stderr, "Error: email body is required (provide via --body or stdin)\n")
		os.Exit(exitError)
	}

	// Extract subject from body if not provided and body starts with "Subject: "
	emailSubject := *subject
	if emailSubject == "" {
		extractedSubject, remainingBody := extractSubjectFromBody(emailBody)
		if extractedSubject != "" {
			emailSubject = extractedSubject
			emailBody = remainingBody
		}
	}

	// Create Azure email client
	client := azemailsender.NewClient(endpoint, accessKey, nil)

	// Build the email message
	messageBuilder := client.NewMessage().
		From(fromEmail).
		To(*to).
		PlainText(emailBody)

	if emailSubject != "" {
		messageBuilder = messageBuilder.Subject(emailSubject)
	}

	message, err := messageBuilder.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building email message: %v\n", err)
		os.Exit(exitError)
	}

	// Send the email
	response, err := client.Send(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending email: %v\n", err)
		os.Exit(exitError)
	}

	fmt.Printf("Email sent successfully! Message ID: %s\n", response.ID)
	os.Exit(exitSuccess)
}

// readFromStdin reads content from standard input
func readFromStdin() (string, error) {
	// Check if stdin has data available
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// If no pipe/redirect, return empty
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil
	}

	reader := bufio.NewReader(os.Stdin)
	var content strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				content.WriteString(line)
				break
			}
			return "", err
		}
		content.WriteString(line)
	}

	return strings.TrimSpace(content.String()), nil
}

// extractSubjectFromBody extracts subject from the first line if it starts with "Subject: "
func extractSubjectFromBody(body string) (subject string, remainingBody string) {
	lines := strings.Split(body, "\n")
	if len(lines) == 0 {
		return "", body
	}

	firstLine := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLine, "Subject: ") {
		subject = strings.TrimSpace(strings.TrimPrefix(firstLine, "Subject: "))
		if len(lines) > 1 {
			remainingBody = strings.Join(lines[1:], "\n")
		} else {
			remainingBody = ""
		}
		remainingBody = strings.TrimSpace(remainingBody)
		return subject, remainingBody
	}

	return "", body
}