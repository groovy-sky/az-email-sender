# Azure Communication Services Email Go Library

A comprehensive Go library for sending emails using Azure Communication Services Email API with extensive debug support and HMAC-SHA256 authentication.

## 🚀 Now with CLI Tool!

This library now includes **azemailsender-cli**, a command-line interface that makes it easy to send emails directly from Bash/PowerShell terminals. Perfect for automation, scripting, and integration workflows.

**[📖 CLI Documentation](CLI.md)** | **[🔧 CLI Quick Start](#cli-quick-start)**

## Installation

### Go Library

```bash
go get github.com/groovy-sky/azemailsender
```

### CLI Tool

**Download pre-built binaries:**
- [Latest releases](https://github.com/groovy-sky/azemailsender/releases)

**Install script:**
```bash
curl -sSL https://raw.githubusercontent.com/groovy-sky/azemailsender/master/scripts/install.sh | bash
```

**Build from source:**
```bash
git clone https://github.com/groovy-sky/azemailsender.git
cd azemailsender
make build
# Binary will be in dist/azemailsender-cli
```

## CLI Quick Start

```bash
# Send a simple email
azemailsender-cli send \
  --endpoint "https://your-resource.communication.azure.com" \
  --access-key "your-access-key" \
  --from "sender@yourdomain.com" \
  --to "recipient@example.com" \
  --subject "Hello World" \
  --text "This is a test email"

# Use environment variables
export AZURE_EMAIL_ENDPOINT="https://your-resource.communication.azure.com"
export AZURE_EMAIL_ACCESS_KEY="your-access-key"
export AZURE_EMAIL_FROM="sender@yourdomain.com"

echo "Email content from stdin" | azemailsender-cli send \
  --to "recipient@example.com" \
  --subject "Test Email"

# Send HTML email with multiple recipients
azemailsender-cli send \
  --from "sender@yourdomain.com" \
  --to "user1@example.com" --to "user2@example.com" \
  --cc "manager@example.com" \
  --subject "Team Update" \
  --html "<h1>Important Update</h1><p>Please review the attached information.</p>" \
  --wait
```

## Requirements

- Go 1.21 or later
- Azure Communication Services resource with Email enabled
- No external dependencies beyond Go standard library (for library usage)

## Quick Start

### Basic Usage

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

## CLI Tool

The **azemailsender-cli** tool provides a command-line interface for the library, enabling:


### CLI Examples

```bash
# Basic email sending
azemailsender-cli send --from sender@example.com --to user@example.com --subject "Test" --text "Hello"

# Pipeline integration  
generate-report | azemailsender-cli send --to team@company.com --subject "Daily Report" --html-file report.html

# Automation with config
azemailsender-cli config init
azemailsender-cli send --to alerts@company.com --subject "System Alert" --text "Service is down"

# JSON output for scripting
result=$(azemailsender-cli send --to user@example.com --subject "Test" --text "Hello" --json)
message_id=$(echo "$result" | jq -r '.id')
```

For complete CLI documentation, examples, and usage patterns, see **[CLI.md](CLI.md)**.

## Building

### Library Only
```bash
go build
```

### CLI Tool
```bash
# Build for current platform
make build

# Build for all platforms  
make build-all

# Install locally
make install
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues related to Azure Communication Services itself, please refer to the [official Azure documentation](https://docs.microsoft.com/en-us/azure/communication-services/).

For library-specific issues, please open an issue in this repository.
