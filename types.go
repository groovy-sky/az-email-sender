package azemailsender

import (
	"log"
	"time"
)

// Logger interface for custom logging implementations
type Logger interface {
	Printf(format string, v ...interface{})
}

// ClientOptions provides configuration options for the email client
type ClientOptions struct {
	// Debug enables comprehensive debug logging
	Debug bool
	
	// Logger is a custom logger implementation. If nil, uses standard log package
	Logger Logger
	
	// HTTPTimeout sets the HTTP client timeout
	HTTPTimeout time.Duration
	
	// APIVersion specifies the Azure Communication Services API version
	APIVersion string
	
	// MaxRetries sets the maximum number of retry attempts for failed requests
	MaxRetries int
	
	// RetryDelay sets the delay between retry attempts
	RetryDelay time.Duration
}

// DefaultClientOptions returns default client options
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		Debug:       false,
		Logger:      log.Default(),
		HTTPTimeout: 30 * time.Second,
		APIVersion:  "2024-07-01-preview",
		MaxRetries:  3,
		RetryDelay:  time.Second,
	}
}

// EmailAddress represents an email address with optional display name
type EmailAddress struct {
	Address     string `json:"address"`
	DisplayName string `json:"displayName,omitempty"`
}

// EmailContent represents the content of an email message
type EmailContent struct {
	Subject   string `json:"subject"`
	PlainText string `json:"plainText,omitempty"`
	Html      string `json:"html,omitempty"`
}

// EmailRecipients represents all recipient types for an email
type EmailRecipients struct {
	To  []EmailAddress `json:"to"`
	Cc  []EmailAddress `json:"cc,omitempty"`
	Bcc []EmailAddress `json:"bcc,omitempty"`
}

// EmailMessage represents a complete email message ready to be sent
type EmailMessage struct {
	SenderAddress string          `json:"senderAddress"`
	Content       EmailContent    `json:"content"`
	Recipients    EmailRecipients `json:"recipients"`
	ReplyTo       []EmailAddress  `json:"replyTo,omitempty"`
}

// SendResponse represents the response from sending an email
type SendResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status,omitempty"`
	Error       *Error `json:"error,omitempty"`
	Timestamp   time.Time
	MessageID   string // Legacy field for backward compatibility
}

// Error represents an error response from the Azure API
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target,omitempty"`
	Details []Error `json:"details,omitempty"`
}

// StatusResponse represents the status of a sent email
type StatusResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Error     *Error    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// WaitOptions provides configuration for waiting for email completion
type WaitOptions struct {
	// PollInterval sets how often to check the status
	PollInterval time.Duration
	
	// MaxWaitTime sets the maximum time to wait for completion
	MaxWaitTime time.Duration
	
	// OnStatusUpdate is called each time the status is checked
	OnStatusUpdate func(status *StatusResponse)
	
	// OnError is called when an error occurs during polling
	OnError func(err error)
}

// DefaultWaitOptions returns default wait options
func DefaultWaitOptions() *WaitOptions {
	return &WaitOptions{
		PollInterval: 5 * time.Second,
		MaxWaitTime:  5 * time.Minute,
		OnStatusUpdate: func(status *StatusResponse) {
			// Default: do nothing
		},
		OnError: func(err error) {
			// Default: do nothing
		},
	}
}

// EmailStatus represents the possible statuses of an email
type EmailStatus string

const (
	StatusQueued     EmailStatus = "Queued"
	StatusOutForDelivery EmailStatus = "OutForDelivery"
	StatusDelivered  EmailStatus = "Delivered"
	StatusFailed     EmailStatus = "Failed"
	StatusCanceled   EmailStatus = "Canceled"
)

// AuthMethod represents the authentication method
type AuthMethod int

const (
	AuthMethodAccessKey AuthMethod = iota
	AuthMethodConnectionString
	AuthMethodHMAC
)

// ParsedConnectionString represents a parsed Azure Communication Services connection string
type ParsedConnectionString struct {
	Endpoint  string
	AccessKey string
}