# azemailsender-cli

A command-line interface for sending emails using Azure Communication Services. This CLI tool wraps the `azemailsender` Go library to enable sending emails directly from Bash/PowerShell terminals.

## Features

- **Multiple authentication methods** - Support for endpoint/access key, connection string
- **Flexible recipient management** - Support for To, CC, BCC recipients
- **Content options** - Support for both plain text and HTML email content
- **Configuration file support** - Store common settings in JSON config files
- **Environment variable support** - Read sensitive data from environment variables
- **Cross-platform compatibility** - Works on Windows, Linux, and macOS
- **Standard Unix patterns** - Support for stdin piping, meaningful exit codes
- **JSON output** - Machine-readable output for scripting integration
- **Status monitoring** - Check email delivery status and wait for completion

## Installation

### Download Pre-built Binaries

Download the latest release for your platform:
- Linux (amd64): `azemailsender-cli-linux-amd64`
- Linux (arm64): `azemailsender-cli-linux-arm64`
- macOS (amd64): `azemailsender-cli-darwin-amd64`
- macOS (arm64): `azemailsender-cli-darwin-arm64`
- Windows (amd64): `azemailsender-cli-windows-amd64.exe`

### Install Script

```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/groovy-sky/azemailsender/main/scripts/install.sh | bash

# Install to custom directory
INSTALL_DIR="$HOME/bin" curl -sSL https://raw.githubusercontent.com/groovy-sky/azemailsender/main/scripts/install.sh | bash
```

### Build from Source

```bash
git clone https://github.com/groovy-sky/azemailsender.git
cd azemailsender
make build
# Binary will be in dist/azemailsender-cli
```

## Quick Start

### Basic Usage

```bash
# Send a simple email (sender address from AZURE_EMAIL_FROM environment variable)
azemailsender-cli send \
  --endpoint "https://your-resource.communication.azure.com" \
  --access-key "your-access-key" \
  --to "recipient@example.com" \
  --subject "Hello World" \
  --text "This is a test email"
```

### Using Connection String

```bash
azemailsender-cli send \
  --connection-string "endpoint=https://your-resource.communication.azure.com;accesskey=your-access-key" \
  --to "recipient@example.com" \
  --subject "Hello World" \
  --text "This is a test email"
```

### Configuration File

Create a configuration file to avoid repeating common settings:

```bash
# Create default config
azemailsender-cli config init --path ~/.config/azemailsender/config.json

# Edit the config file
{
  "endpoint": "https://your-resource.communication.azure.com",
  "access_key": "your-access-key",
  "from": "sender@yourdomain.com",
  "debug": false,
  "quiet": false,
  "json": false
}

# Send email using config
azemailsender-cli send \
  --config ~/.config/azemailsender/config.json \
  --to "recipient@example.com" \
  --subject "Hello World" \
  --text "This is a test email"
```

### Environment Variables

```bash
# Set environment variables
export AZURE_EMAIL_ENDPOINT="https://your-resource.communication.azure.com"
export AZURE_EMAIL_ACCESS_KEY="your-access-key"
export AZURE_EMAIL_FROM="sender@yourdomain.com"

# Send email (credentials read from environment)
azemailsender-cli send \
  --to "recipient@example.com" \
  --subject "Hello World" \
  --text "This is a test email"
```

## Commands

### send

Send an email message.

```bash
azemailsender-cli send [flags]
```

**Required:**
- Email subject (`--subject` or `AZURE_EMAIL_SUBJECT`)
- At least one recipient (`--to`, `--cc`, or `--bcc`)
- Authentication (`--connection-string` OR `--endpoint` + `--access-key`)
- Sender address (`AZURE_EMAIL_FROM` environment variable)

**Content flags:**
- `--text` - Plain text email content
- `--html` - HTML email content
- `--text-file` - Read plain text content from file
- `--html-file` - Read HTML content from file

**Recipient flags:**
- `--to, -t` - To recipients (can be repeated)
- `--cc` - CC recipients (can be repeated)
- `--bcc` - BCC recipients (can be repeated)
- `--reply-to` - Reply-to email address

**Authentication flags:**
- `--endpoint, -e` - Azure Communication Services endpoint
- `--access-key, -k` - Access key for authentication
- `--connection-string` - Connection string for authentication

**Behavior flags:**
- `--wait, -w` - Wait for email completion
- `--poll-interval` - Status polling interval (default: 5s)
- `--max-wait-time` - Maximum wait time (default: 5m)

**Examples:**

```bash
# Simple email (sender address from AZURE_EMAIL_FROM environment variable)
azemailsender-cli send --to recipient@example.com --subject "Hello" --text "Hello World"

# HTML email with multiple recipients
azemailsender-cli send --to user1@example.com --to user2@example.com --cc manager@example.com --subject "Report" --html "<h1>Monthly Report</h1>"

# Send email and wait for completion
azemailsender-cli send --to recipient@example.com --subject "Hello" --text "Hello World" --wait

# Read content from stdin
echo "Hello from stdin" | azemailsender-cli send --to recipient@example.com --subject "Stdin Test"

# Read content from file
azemailsender-cli send --to recipient@example.com --subject "File Test" --text-file message.txt
```

### status

Check the status of a previously sent email.

```bash
azemailsender-cli status <message-id> [flags]
```

**Examples:**

```bash
# Check status once
azemailsender-cli status abc123def456

# Check status and wait for completion
azemailsender-cli status abc123def456 --wait

# Check status with custom polling
azemailsender-cli status abc123def456 --wait --poll-interval 10s --max-wait-time 2m
```

### config

Manage configuration files and environment variables.

```bash
azemailsender-cli config [command]
```

**Subcommands:**
- `init` - Create a default configuration file
- `show` - Show current configuration
- `env` - Show environment variable examples

**Examples:**

```bash
# Create default config
azemailsender-cli config init --path ~/.config/azemailsender/config.json

# Show current config
azemailsender-cli config show

# Show environment variable examples
azemailsender-cli config env
```

### version

Show version information.

```bash
azemailsender-cli version
```

## Configuration

Configuration is loaded in the following order (later sources override earlier ones):

1. Default values
2. Configuration file
3. Environment variables
4. Command-line flags

### Configuration File

The configuration file uses JSON format:

```json
{
  "endpoint": "https://your-resource.communication.azure.com",
  "access_key": "your-access-key",
  "connection_string": "endpoint=https://your-resource.communication.azure.com;accesskey=your-access-key",
  "from": "sender@yourdomain.com",
  "reply_to": "noreply@yourdomain.com",
  "debug": false,
  "quiet": false,
  "json": false,
  "wait": false,
  "poll_interval": "5s",
  "max_wait_time": "5m"
}
```

**Configuration file locations (searched in order):**
1. Path specified by `--config` flag
2. `./azemailsender.json` (current directory)
3. `$HOME/.config/azemailsender/azemailsender.json`
4. `/etc/azemailsender/azemailsender.json`

### Environment Variables

All environment variables use the `AZURE_EMAIL_` prefix:

- `AZURE_EMAIL_ENDPOINT` - Azure Communication Services endpoint
- `AZURE_EMAIL_ACCESS_KEY` - Access key for authentication
- `AZURE_EMAIL_CONNECTION_STRING` - Connection string for authentication
- `AZURE_EMAIL_FROM` - Default sender email address
- `AZURE_EMAIL_REPLY_TO` - Default reply-to email address
- `AZURE_EMAIL_DEBUG` - Enable debug logging (true/false)
- `AZURE_EMAIL_QUIET` - Suppress output except errors (true/false)
- `AZURE_EMAIL_JSON` - Output in JSON format (true/false)

## Global Flags

These flags are available for all commands:

- `--config, -c` - Configuration file path
- `--debug, -d` - Enable debug logging
- `--quiet, -q` - Suppress output except errors
- `--json, -j` - Output in JSON format

## Output Formats

### Standard Output

```bash
$ azemailsender-cli send --from sender@example.com --to recipient@example.com --subject "Test" --text "Hello"
Email sent successfully!
Message ID: abc123def456
```

### JSON Output

```bash
$ azemailsender-cli send --from sender@example.com --to recipient@example.com --subject "Test" --text "Hello" --json
{
  "id": "abc123def456",
  "status": "Queued",
  "timestamp": "2023-12-07T10:30:00Z"
}
```

### Debug Output

```bash
$ azemailsender-cli send --from sender@example.com --to recipient@example.com --subject "Test" --text "Hello" --debug
[DEBUG] Client initialized with endpoint: https://your-resource.communication.azure.com
[DEBUG] Authentication method: HMAC-SHA256
[DEBUG] Creating new message builder
[DEBUG] Setting sender address: sender@example.com
[DEBUG] Adding TO recipient: recipient@example.com
[DEBUG] Setting email subject: Test
[DEBUG] Setting plain text content (5 characters)
[DEBUG] Message validation successful
[DEBUG] Starting email send process
[DEBUG] Email sent successfully in 1.234s
Email sent successfully!
Message ID: abc123def456
```

## Error Handling

The CLI uses standard Unix exit codes:

- `0` - Success
- `1` - General error
- `2` - Misuse of shell command (invalid arguments)

Error messages are written to stderr:

```bash
$ azemailsender-cli send --from sender@example.com --subject "Test" --text "Hello"
Error: validation failed:
  at least one recipient required (--to, --cc, or --bcc)
```

## Integration Examples

### Shell Script

```bash
#!/bin/bash
# Send daily report

# Set up environment variables
export AZURE_EMAIL_FROM="reports@company.com"

REPORT_FILE="/tmp/daily-report.html"
generate_report > "$REPORT_FILE"

azemailsender-cli send \
  --to "manager@company.com" \
  --cc "team@company.com" \
  --subject "Daily Report - $(date +%Y-%m-%d)" \
  --html-file "$REPORT_FILE" \
  --wait

if [ $? -eq 0 ]; then
  echo "Report sent successfully"
else
  echo "Failed to send report" >&2
  exit 1
fi
```

### PowerShell Script

```powershell
# Send notification email
$env:AZURE_EMAIL_ENDPOINT = "https://your-resource.communication.azure.com"
$env:AZURE_EMAIL_ACCESS_KEY = "your-access-key"
$env:AZURE_EMAIL_FROM = "notifications@company.com"

$content = @"
Hello,

This is an automated notification.

Best regards,
System
"@

$content | azemailsender-cli send `
  --to "admin@company.com" `
  --subject "System Notification" `
  --json

if ($LASTEXITCODE -eq 0) {
  Write-Host "Notification sent successfully"
} else {
  Write-Error "Failed to send notification"
  exit 1
}
```

### Cron Job

```bash
# Add to crontab: 0 9 * * 1 /path/to/weekly-report.sh

#!/bin/bash
# Weekly report cron job

export AZURE_EMAIL_ENDPOINT="https://your-resource.communication.azure.com"
export AZURE_EMAIL_ACCESS_KEY="your-access-key"
export AZURE_EMAIL_FROM="reports@company.com"

# Generate and send weekly report
echo "Weekly summary for week of $(date +%Y-%m-%d)" | \
azemailsender-cli send \
  --to "management@company.com" \
  --subject "Weekly Report - $(date +%Y-%m-%d)" \
  --quiet

# Log result
if [ $? -eq 0 ]; then
  logger "Weekly report sent successfully"
else
  logger "Failed to send weekly report"
fi
```

## Building

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install locally
make install

# Run tests
make test

# Clean build artifacts
make clean
```

### Manual Build

```bash
go build -o azemailsender-cli ./cmd/azemailsender-cli
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.