package main

import (
	"fmt"
	"log"
	"os"
	"github.com/groovy-sky/azemailsender"
)

// CustomLogger implements the Logger interface
type CustomLogger struct {
	prefix string
}

func (l *CustomLogger) Printf(format string, v ...interface{}) {
	// Add custom prefix and write to stderr for debug messages
	fmt.Fprintf(os.Stderr, l.prefix+" "+format+"\n", v...)
}

func main() {
	// Create custom logger
	customLogger := &CustomLogger{prefix: "[AZURE-EMAIL-DEBUG]"}
	
	// Create client with debug logging and custom logger
	client := azemailsender.NewClient(
		"https://<RESOURCE-NAME>.communication.azure.com",
		"<YOUR_ACCESS_KEY>",
		&azemailsender.ClientOptions{
			Debug:  true,
			Logger: customLogger,
		},
	)

	// Build a more complex email with multiple recipients and HTML content
	message, err := client.NewMessage().
		From("sender@yourdomain.com").
		To("recipient1@example.com", "John Doe").
		To("recipient2@example.com").
		Cc("manager@example.com", "Manager").
		Bcc("archive@example.com").
		ReplyTo("noreply@yourdomain.com").
		Subject("Debug Test Email with Custom Logger").
		PlainText("This is the plain text version of the email.").
		HTML(`
			<html>
				<body>
					<h1>Debug Test Email</h1>
					<p>This email demonstrates <strong>debug logging</strong> with a custom logger.</p>
					<ul>
						<li>Multiple recipients</li>
						<li>HTML content</li>
						<li>Custom debug output</li>
					</ul>
				</body>
			</html>
		`).
		Build()
	
	if err != nil {
		log.Fatalf("Failed to build message: %v", err)
	}

	// Send email with debug output
	fmt.Println("Sending email with debug logging...")
	resp, err := client.Send(message)
	if err != nil {
		log.Fatalf("Email failed: %v", err)
	}
	
	fmt.Printf("Email sent successfully! ID: %s\n", resp.ID)
	
	// Demonstrate status polling with custom callbacks
	fmt.Println("Monitoring email status...")
	waitOptions := &azemailsender.WaitOptions{
		OnStatusUpdate: func(status *azemailsender.StatusResponse) {
			fmt.Printf("Status update: %s at %v\n", status.Status, status.Timestamp)
		},
		OnError: func(err error) {
			fmt.Printf("Status check error: %v\n", err)
		},
	}
	
	finalStatus, err := client.WaitForCompletion(resp.ID, waitOptions)
	if err != nil {
		fmt.Printf("Status monitoring failed: %v\n", err)
	} else {
		fmt.Printf("Final status: %s\n", finalStatus.Status)
	}
}