# az-email-sender

A lightweight Go library for sending emails via Azure Communication Services Email REST API.

## Features

- Send plain text or HTML emails
- To, CC, BCC support
- Minimal dependencies (just the Go standard library)
- Simple, idiomatic Go API

## Installation

```bash
go get github.com/groovy-sky/az-email-sender/azemailsender
```

## Usage

1. Ensure you have an Azure Communication Services resource with Email enabled.

2. Use the following code:

```go
import "github.com/groovy-sky/az-email-sender/azemailsender"

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
```

## Example

See [`example/main.go`](example/main.go) for a runnable example.

## License

MIT License (see [LICENSE](LICENSE))