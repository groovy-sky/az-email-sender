package main

import (
    "fmt"
    "log"

    "myproject/azemailsender" // or "github.com/yourname/az-email-sender"
)

func main() {
    sender := azemailsender.New(
        "https://<RESOURCE-NAME>.communication.azure.com",
        "<YOUR_ACCESS_KEY>",
    )

    req := azemailsender.EmailRequest{
        SenderAddress: "sender@yourdomain.com",
        Content: azemailsender.EmailContent{
            Subject:   "Test Email from Go",
            PlainText: "This is a test email sent via Azure Communication Services Email REST API.",
        },
        Recipients: azemailsender.EmailRecipients{
            To: []azemailsender.EmailAddress{
                {Address: "recipient@example.com"},
            },
        },
    }

    resp, err := sender.SendEmail(req)
    if err != nil {
        log.Fatalf("Email failed: %v", err)
    }
    fmt.Printf("Email sent! Message ID: %s\n", resp.MessageId)
}