# Azure Communication Services Email Go Library

A comprehensive Go library for sending emails using Azure Communication Services Email API with extensive debug support and HMAC-SHA256 authentication.

## CLI Tool

This repository also includes a simple command-line interface (CLI) tool that wraps the library for easy email sending from the command line or scripts.

## Features

- **HTTP-based email sending** using Azure Communication Services REST API
- **HMAC-SHA256 authentication** for Azure API with automatic signature generation
- **Multiple authentication methods**: endpoint/access key, connection string, and legacy api-key
- **Fluent message builder interface** for easy email construction
- **Automatic status polling** with customizable intervals and callbacks
- **Support for HTML and plain text content**
- **CC and BCC recipients support**
- **Comprehensive error handling** with wrapped errors
- **Extensive debug logging** throughout the entire process
- **Configurable debug logging** (enable/disable at runtime)
- **Custom logger support** for integration with existing logging systems
- **Thread-safe client implementation**
- **Configurable HTTP timeouts and retry logic**

## Installation

```bash
go get github.com/groovy-sky/azemailsender
```

## Requirements

- Go 1.21 or later
- Azure Communication Services resource with Email enabled
- No external dependencies beyond Go standard library

## CLI Installation and Usage

### Installation

Download the pre-built binary for your platform from the releases page, or build from source:

```bash
# Clone the repository
git clone https://github.com/groovy-sky/azemailsender
cd azemailsender

# Build for your platform
make dev

# Or build for all platforms
make build

# Install locally
make install
```

### CLI Configuration

The CLI tool requires three environment variables:

```bash
export AZURE_EMAIL_ENDPOINT="https://your-resource.communication.azure.com"
export AZURE_EMAIL_ACCESS_KEY="your-access-key"
export AZURE_EMAIL_FROM="sender@yourdomain.com"
```

### CLI Usage

```bash
# Basic usage
azemailsender --to "user@example.com" --subject "Hello" --body "This is a test message"

# Pipe body content from command output
echo "Hello World" | azemailsender --to "user@example.com" --subject "Greeting"

# Read body from file
cat message.txt | azemailsender --to "user@example.com" --subject "File Content"

# Subject extracted from body (first line starting with "Subject: ")
azemailsender --to "user@example.com" --body "Subject: Important Message
This is the email content that will be sent."

# PowerShell usage
echo "Hello from PowerShell" | azemailsender --to "user@example.com"

# Using in scripts
if azemailsender --to "admin@example.com" --subject "Alert" --body "System issue detected"; then
    echo "Alert sent successfully"
else
    echo "Failed to send alert"
fi
```

### CLI Options

- `--to` (required) - Email recipient address
- `--body` (optional) - Email body content. If not provided, reads from stdin
- `--subject` (optional) - Email subject line. If not provided and body starts with "Subject: ", extracts from first line

### Exit Codes

- `0` - Success
- `1` - Error (missing arguments, authentication failure, send failure)

## Go Library Usage

## Go Library Usage

### Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/groovy-sky/azemailsender"
)

func main() {
    // Create client
    client := azemailsender.NewClient(
        "https://your-resource.communication.azure.com",
        "your-access-key",
        nil, // Use default options
    )

    // Build and send email
    message, err := client.NewMessage().
        From("sender@yourdomain.com").
        To("recipient@example.com").
        Subject("Hello from Go!").
        PlainText("This is a test email.").
        Build()
    
    if err != nil {
        log.Fatalf("Failed to build message: %v", err)
    }

    response, err := client.Send(message)
    if err != nil {
        log.Fatalf("Failed to send email: %v", err)
    }
    
    fmt.Printf("Email sent! ID: %s\n", response.ID)
}
```

### With Debug Logging

```go
client := azemailsender.NewClient(
    "https://your-resource.communication.azure.com",
    "your-access-key",
    &azemailsender.ClientOptions{
        Debug: true, // Enable comprehensive debug logging
    },
)
```

### Using Connection String

```go
connectionString := "endpoint=https://your-resource.communication.azure.com;accesskey=your-access-key"
client, err := azemailsender.NewClientFromConnectionString(
    connectionString,
    &azemailsender.ClientOptions{Debug: true},
)
```

## Advanced Usage

### Complex Email with Multiple Recipients

```go
message, err := client.NewMessage().
    From("sender@yourdomain.com").
    To("recipient1@example.com", "John Doe").
    To("recipient2@example.com").
    Cc("manager@example.com", "Manager").
    Bcc("archive@example.com").
    ReplyTo("noreply@yourdomain.com").
    Subject("Complex Email").
    PlainText("Plain text version").
    HTML(`
        <html>
            <body>
                <h1>HTML Email</h1>
                <p>This is an <strong>HTML email</strong>.</p>
            </body>
        </html>
    `).
    Build()
```

### Status Monitoring

```go
// Send email
response, err := client.Send(message)
if err != nil {
    log.Fatal(err)
}

// Monitor status with custom callbacks
waitOptions := &azemailsender.WaitOptions{
    PollInterval: 5 * time.Second,
    MaxWaitTime:  2 * time.Minute,
    OnStatusUpdate: func(status *azemailsender.StatusResponse) {
        fmt.Printf("Status: %s at %v\n", status.Status, status.Timestamp)
    },
    OnError: func(err error) {
        fmt.Printf("Error: %v\n", err)
    },
}

finalStatus, err := client.WaitForCompletion(response.ID, waitOptions)
if err != nil {
    log.Printf("Monitoring failed: %v", err)
} else {
    fmt.Printf("Final status: %s\n", finalStatus.Status)
}
```

### Custom Logger

```go
type CustomLogger struct{}

func (l *CustomLogger) Printf(format string, v ...interface{}) {
    // Your custom logging logic
    log.Printf("[CUSTOM] "+format, v...)
}

client := azemailsender.NewClient(
    endpoint, accessKey,
    &azemailsender.ClientOptions{
        Debug:  true,
        Logger: &CustomLogger{},
    },
)
```

## Configuration Options

### ClientOptions

```go
type ClientOptions struct {
    Debug       bool          // Enable debug logging
    Logger      Logger        // Custom logger implementation
    HTTPTimeout time.Duration // HTTP client timeout
    APIVersion  string        // Azure API version
    MaxRetries  int          // Maximum retry attempts
    RetryDelay  time.Duration // Delay between retries
}
```

### WaitOptions

```go
type WaitOptions struct {
    PollInterval   time.Duration                    // How often to check status
    MaxWaitTime    time.Duration                    // Maximum time to wait
    OnStatusUpdate func(*StatusResponse)            // Called on each status check
    OnError        func(error)                      // Called on errors
}
```

## Authentication Methods

### 1. HMAC-SHA256 (Recommended)

```go
client := azemailsender.NewClient(endpoint, accessKey, options)
```

### 2. Connection String

```go
client, err := azemailsender.NewClientFromConnectionString(connectionString, options)
```

### 3. Legacy API Key

```go
client := azemailsender.NewClientWithAccessKey(endpoint, accessKey, options)
```

## Debug Output

When debug logging is enabled, the library provides comprehensive information about:

- Client initialization and configuration
- Message building steps and validation
- HTTP request details (URL, headers, body size)
- Authentication signature generation process
- Response details and timing information
- Status polling attempts and results
- Error details and troubleshooting information

Example debug output:
```
[DEBUG] Client initialized with endpoint: https://your-resource.communication.azure.com
[DEBUG] Authentication method: HMAC-SHA256
[DEBUG] API Version: 2024-07-01-preview
[DEBUG] Creating new message builder
[DEBUG] Setting sender address: sender@example.com
[DEBUG] Adding TO recipient: recipient@example.com
[DEBUG] Setting email subject: Test Email
[DEBUG] Setting plain text content (20 characters)
[DEBUG] Message validation successful
[DEBUG] Starting email send process
[DEBUG] Generating HMAC signature
[DEBUG] HTTP Request: POST https://your-resource.communication.azure.com/emails:send
[DEBUG] Email sent successfully in 1.234s
```

## Error Handling

The library provides detailed error information with context:

```go
response, err := client.Send(message)
if err != nil {
    // Errors are wrapped with context
    fmt.Printf("Send failed: %v\n", err)
    
    // You can unwrap to get the root cause if needed
    if rootErr := errors.Unwrap(err); rootErr != nil {
        fmt.Printf("Root cause: %v\n", rootErr)
    }
}
```

## Examples

The repository includes several example implementations:

- [`example/main.go`](example/main.go) - Comprehensive example with multiple scenarios
- [`example/simple/main.go`](example/simple/main.go) - Simple usage example
- [`example/debug-only/main.go`](example/debug-only/main.go) - Debug-focused example with custom logger

## API Compatibility

This library uses Azure Communication Services Email API version `2024-07-01-preview` by default. You can specify a different version in the client options:

```go
options := &azemailsender.ClientOptions{
    APIVersion: "2023-03-31", // Use older API version
}
```

## Thread Safety

The client is thread-safe and can be used concurrently from multiple goroutines. Each request is independent and doesn't share state.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues related to Azure Communication Services itself, please refer to the [official Azure documentation](https://docs.microsoft.com/en-us/azure/communication-services/).

For library-specific issues, please open an issue in this repository.