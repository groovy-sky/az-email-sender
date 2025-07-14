package azemailsender

import (
	"fmt"
	"strings"
)

// MessageBuilder provides a fluent interface for building email messages
type MessageBuilder struct {
	client  *Client
	message *EmailMessage
}

// NewMessage creates a new message builder
func (c *Client) NewMessage() *MessageBuilder {
	if c.options.Debug {
		c.logger.Printf("[DEBUG] Creating new message builder")
	}
	
	return &MessageBuilder{
		client: c,
		message: &EmailMessage{
			Recipients: EmailRecipients{
				To:  make([]EmailAddress, 0),
				Cc:  make([]EmailAddress, 0),
				Bcc: make([]EmailAddress, 0),
			},
			ReplyTo: make([]EmailAddress, 0),
		},
	}
}

// From sets the sender address for the email
func (b *MessageBuilder) From(address string) *MessageBuilder {
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Setting sender address: %s", address)
	}
	
	b.message.SenderAddress = address
	return b
}

// To adds a recipient to the "To" field
func (b *MessageBuilder) To(address string, displayName ...string) *MessageBuilder {
	emailAddr := EmailAddress{Address: address}
	if len(displayName) > 0 && displayName[0] != "" {
		emailAddr.DisplayName = displayName[0]
	}
	
	if b.client.options.Debug {
		if emailAddr.DisplayName != "" {
			b.client.logger.Printf("[DEBUG] Adding TO recipient: %s <%s>", emailAddr.DisplayName, emailAddr.Address)
		} else {
			b.client.logger.Printf("[DEBUG] Adding TO recipient: %s", emailAddr.Address)
		}
	}
	
	b.message.Recipients.To = append(b.message.Recipients.To, emailAddr)
	return b
}

// Cc adds a recipient to the "Cc" field
func (b *MessageBuilder) Cc(address string, displayName ...string) *MessageBuilder {
	emailAddr := EmailAddress{Address: address}
	if len(displayName) > 0 && displayName[0] != "" {
		emailAddr.DisplayName = displayName[0]
	}
	
	if b.client.options.Debug {
		if emailAddr.DisplayName != "" {
			b.client.logger.Printf("[DEBUG] Adding CC recipient: %s <%s>", emailAddr.DisplayName, emailAddr.Address)
		} else {
			b.client.logger.Printf("[DEBUG] Adding CC recipient: %s", emailAddr.Address)
		}
	}
	
	b.message.Recipients.Cc = append(b.message.Recipients.Cc, emailAddr)
	return b
}

// Bcc adds a recipient to the "Bcc" field
func (b *MessageBuilder) Bcc(address string, displayName ...string) *MessageBuilder {
	emailAddr := EmailAddress{Address: address}
	if len(displayName) > 0 && displayName[0] != "" {
		emailAddr.DisplayName = displayName[0]
	}
	
	if b.client.options.Debug {
		if emailAddr.DisplayName != "" {
			b.client.logger.Printf("[DEBUG] Adding BCC recipient: %s <%s>", emailAddr.DisplayName, emailAddr.Address)
		} else {
			b.client.logger.Printf("[DEBUG] Adding BCC recipient: %s", emailAddr.Address)
		}
	}
	
	b.message.Recipients.Bcc = append(b.message.Recipients.Bcc, emailAddr)
	return b
}

// ReplyTo adds a reply-to address
func (b *MessageBuilder) ReplyTo(address string, displayName ...string) *MessageBuilder {
	emailAddr := EmailAddress{Address: address}
	if len(displayName) > 0 && displayName[0] != "" {
		emailAddr.DisplayName = displayName[0]
	}
	
	if b.client.options.Debug {
		if emailAddr.DisplayName != "" {
			b.client.logger.Printf("[DEBUG] Adding ReplyTo address: %s <%s>", emailAddr.DisplayName, emailAddr.Address)
		} else {
			b.client.logger.Printf("[DEBUG] Adding ReplyTo address: %s", emailAddr.Address)
		}
	}
	
	b.message.ReplyTo = append(b.message.ReplyTo, emailAddr)
	return b
}

// Subject sets the email subject
func (b *MessageBuilder) Subject(subject string) *MessageBuilder {
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Setting email subject: %s", subject)
	}
	
	b.message.Content.Subject = subject
	return b
}

// PlainText sets the plain text content of the email
func (b *MessageBuilder) PlainText(content string) *MessageBuilder {
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Setting plain text content (%d characters)", len(content))
	}
	
	b.message.Content.PlainText = content
	return b
}

// HTML sets the HTML content of the email
func (b *MessageBuilder) HTML(content string) *MessageBuilder {
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Setting HTML content (%d characters)", len(content))
	}
	
	b.message.Content.Html = content
	return b
}

// AddMultipleRecipients adds multiple recipients to the specified field
func (b *MessageBuilder) AddMultipleRecipients(recipientType string, addresses []string) *MessageBuilder {
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Adding %d recipients to %s field", len(addresses), recipientType)
	}
	
	for _, addr := range addresses {
		switch strings.ToLower(recipientType) {
		case "to":
			b.To(addr)
		case "cc":
			b.Cc(addr)
		case "bcc":
			b.Bcc(addr)
		default:
			if b.client.options.Debug {
				b.client.logger.Printf("[DEBUG] Unknown recipient type: %s", recipientType)
			}
		}
	}
	return b
}

// Validate validates the email message before building
func (b *MessageBuilder) Validate() error {
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Validating email message")
	}
	
	var errors []string
	
	// Check sender address
	if b.message.SenderAddress == "" {
		errors = append(errors, "sender address is required")
	}
	
	// Check subject
	if b.message.Content.Subject == "" {
		errors = append(errors, "subject is required")
	}
	
	// Check content
	if b.message.Content.PlainText == "" && b.message.Content.Html == "" {
		errors = append(errors, "either plain text or HTML content is required")
	}
	
	// Check recipients
	if len(b.message.Recipients.To) == 0 && len(b.message.Recipients.Cc) == 0 && len(b.message.Recipients.Bcc) == 0 {
		errors = append(errors, "at least one recipient is required")
	}
	
	// Validate email addresses
	allRecipients := append(append(b.message.Recipients.To, b.message.Recipients.Cc...), b.message.Recipients.Bcc...)
	for _, recipient := range allRecipients {
		if !isValidEmail(recipient.Address) {
			errors = append(errors, fmt.Sprintf("invalid email address: %s", recipient.Address))
		}
	}
	
	// Validate sender email
	if !isValidEmail(b.message.SenderAddress) {
		errors = append(errors, fmt.Sprintf("invalid sender email address: %s", b.message.SenderAddress))
	}
	
	if len(errors) > 0 {
		if b.client.options.Debug {
			b.client.logger.Printf("[DEBUG] Validation failed with %d errors:", len(errors))
			for _, err := range errors {
				b.client.logger.Printf("[DEBUG]   - %s", err)
			}
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}
	
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Message validation successful")
	}
	
	return nil
}

// Build finalizes and returns the email message
func (b *MessageBuilder) Build() (*EmailMessage, error) {
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Building email message")
	}
	
	if err := b.Validate(); err != nil {
		return nil, err
	}
	
	if b.client.options.Debug {
		b.client.logger.Printf("[DEBUG] Message built successfully:")
		b.client.logger.Printf("[DEBUG]   From: %s", b.message.SenderAddress)
		b.client.logger.Printf("[DEBUG]   Subject: %s", b.message.Content.Subject)
		b.client.logger.Printf("[DEBUG]   To recipients: %d", len(b.message.Recipients.To))
		b.client.logger.Printf("[DEBUG]   CC recipients: %d", len(b.message.Recipients.Cc))
		b.client.logger.Printf("[DEBUG]   BCC recipients: %d", len(b.message.Recipients.Bcc))
		b.client.logger.Printf("[DEBUG]   Has plain text: %t", b.message.Content.PlainText != "")
		b.client.logger.Printf("[DEBUG]   Has HTML: %t", b.message.Content.Html != "")
	}
	
	return b.message, nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Basic validation - contains @ and has parts before and after
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	
	localPart := parts[0]
	domainPart := parts[1]
	
	if len(localPart) == 0 || len(domainPart) == 0 {
		return false
	}
	
	// Domain must contain a dot
	if !strings.Contains(domainPart, ".") {
		return false
	}
	
	return true
}