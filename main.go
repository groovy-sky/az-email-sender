package azemailsender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type EmailContent struct {
	Subject   string `json:"subject"`
	PlainText string `json:"plainText,omitempty"`
	Html      string `json:"html,omitempty"`
}

type EmailAddress struct {
	Address     string `json:"address"`
	DisplayName string `json:"displayName,omitempty"`
}

type EmailRecipients struct {
	To  []EmailAddress `json:"to"`
	Cc  []EmailAddress `json:"cc,omitempty"`
	Bcc []EmailAddress `json:"bcc,omitempty"`
}

type EmailRequest struct {
	SenderAddress string          `json:"senderAddress"`
	Content       EmailContent    `json:"content"`
	Recipients    EmailRecipients `json:"recipients"`
	ReplyTo       []EmailAddress  `json:"replyTo,omitempty"`
}

type EmailResponse struct {
	MessageId string `json:"messageId"`
}

// EmailSender is the main client for sending emails via Azure Communication Services Email REST API.
type EmailSender struct {
	Endpoint   string
	AccessKey  string
	HttpClient *http.Client
}

// New creates a new EmailSender client.
func New(endpoint, accessKey string) *EmailSender {
	fmt.Printf("[INF] Creating EmailSender with endpoint: %s\n", endpoint)
	return &EmailSender{
		Endpoint:   endpoint,
		AccessKey:  accessKey,
		HttpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendEmail sends an email using Azure Communication Services Email REST API.
func (s *EmailSender) SendEmail(req EmailRequest) (*EmailResponse, error) {
	fmt.Printf("[INF] Preparing to send email from: %s\n", req.SenderAddress)
	fmt.Printf("[INF] Email subject: %s\n", req.Content.Subject)
	fmt.Printf("[INF] Email recipients: %+v\n", req.Recipients)

	// Correct URL format as per official documentation
	url := fmt.Sprintf("%s/emails:send?api-version=2023-03-31", s.Endpoint)
	fmt.Printf("[INF] API URL: %s\n", url)

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("[ERR] Failed to marshal request: %v\n", err)
		return nil, err
	}
	fmt.Printf("[INF] Marshalled request body: %s\n", string(body))

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("[ERR] Failed to create HTTP request: %v\n", err)
		return nil, err
	}

	// Set headers exactly as shown in the official documentation
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", s.AccessKey)
	fmt.Printf("[INF] Set HTTP headers\n")

	resp, err := s.HttpClient.Do(httpReq)
	if err != nil {
		fmt.Printf("[ERR] HTTP request failed: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	fmt.Printf("[INF] Received HTTP response with status: %s\n", resp.Status)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("[ERR] Email send failed. Status: %d, Body: %s\n", resp.StatusCode, string(b))
		return nil, fmt.Errorf("failed to send email: %s", string(b))
	}

	var emailResp EmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		fmt.Printf("[ERR] Failed to decode email response: %v\n", err)
		return nil, err
	}
	fmt.Printf("[INF] Email sent successfully. MessageId: %s\n", emailResp.MessageId)
	return &emailResp, nil
}