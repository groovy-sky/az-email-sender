package main

import (
	"fmt"
	"log"
	"time"
	"github.com/groovy-sky/azemailsender"
)

func main() {
	// Create client with debug logging enabled
	client := azemailsender.NewClient(
		"https://<RESOURCE-NAME>.communication.azure.com",
		"<YOUR_ACCESS_KEY>",
		&azemailsender.ClientOptions{Debug: true},
	)

	// Example 1: Basic email with fluent interface
	fmt.Println("=== Example 1: Basic Email ===")
	message, err := client.NewMessage().
		From("sender@yourdomain.com").
		To("recipient@example.com").
		Subject("Test Email from Go").
		PlainText("This is a test email sent via Azure Communication Services Email REST API.").
		Build()
	
	if err != nil {
		log.Fatalf("Failed to build message: %v", err)
	}

	resp, err := client.Send(message)
	if err != nil {
		log.Fatalf("Email failed: %v", err)
	}
	fmt.Printf("Email sent! Message ID: %s\n", resp.ID)

	// Example 2: Complex email with multiple recipients and HTML
	fmt.Println("\n=== Example 2: Complex Email ===")
	complexMessage, err := client.NewMessage().
		From("sender@yourdomain.com").
		To("recipient1@example.com", "John Doe").
		To("recipient2@example.com").
		Cc("manager@example.com", "Manager").
		Bcc("archive@example.com").
		ReplyTo("noreply@yourdomain.com").
		Subject("Complex Test Email").
		PlainText("This is the plain text version.").
		HTML(`
			<html>
				<body>
					<h1>Welcome!</h1>
					<p>This is an <strong>HTML email</strong> with multiple recipients.</p>
				</body>
			</html>
		`).
		Build()
	
	if err != nil {
		log.Fatalf("Failed to build complex message: %v", err)
	}

	complexResp, err := client.Send(complexMessage)
	if err != nil {
		log.Fatalf("Complex email failed: %v", err)
	}
	fmt.Printf("Complex email sent! Message ID: %s\n", complexResp.ID)

	// Example 3: Status monitoring
	fmt.Println("\n=== Example 3: Status Monitoring ===")
	waitOptions := &azemailsender.WaitOptions{
		PollInterval: 2 * time.Second,
		MaxWaitTime:  30 * time.Second,
		OnStatusUpdate: func(status *azemailsender.StatusResponse) {
			fmt.Printf("Status update: %s\n", status.Status)
		},
	}
	
	finalStatus, err := client.WaitForCompletion(resp.ID, waitOptions)
	if err != nil {
		fmt.Printf("Status monitoring failed: %v\n", err)
	} else {
		fmt.Printf("Final status: %s\n", finalStatus.Status)
	}

	// Example 4: Connection string authentication
	fmt.Println("\n=== Example 4: Connection String Auth ===")
	connectionString := "endpoint=https://<RESOURCE-NAME>.communication.azure.com;accesskey=<YOUR_ACCESS_KEY>"
	connClient, err := azemailsender.NewClientFromConnectionString(connectionString, &azemailsender.ClientOptions{Debug: true})
	if err != nil {
		log.Fatalf("Failed to create client from connection string: %v", err)
	}
	
	connMessage, err := connClient.NewMessage().
		From("sender@yourdomain.com").
		To("recipient@example.com").
		Subject("Connection String Test").
		PlainText("This email was sent using connection string authentication.").
		Build()
	
	if err != nil {
		log.Fatalf("Failed to build connection string message: %v", err)
	}

	connResp, err := connClient.Send(connMessage)
	if err != nil {
		log.Fatalf("Connection string email failed: %v", err)
	}
	fmt.Printf("Connection string email sent! Message ID: %s\n", connResp.ID)
}
