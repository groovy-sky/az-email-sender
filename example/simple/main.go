package main

import (
	"fmt"
	"github.com/groovy-sky/azemailsender"
	"log"
)

func main() {
	// Simple usage example without debug logging
	client := azemailsender.NewClient(
		"https://<RESOURCE-NAME>.communication.azure.com",
		"<YOUR_ACCESS_KEY>",
		nil, // Use default options
	)

	message, err := client.NewMessage().
		From("sender@yourdomain.com").
		To("recipient@example.com").
		Subject("Simple Test Email").
		PlainText("Hello, this is a simple test email!").
		Build()

	if err != nil {
		log.Fatalf("Failed to build message: %v", err)
	}

	resp, err := client.Send(message)
	if err != nil {
		log.Fatalf("Email failed: %v", err)
	}

	fmt.Printf("Email sent successfully! ID: %s\n", resp.ID)
}
