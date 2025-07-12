package main

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
	return &EmailSender{
		Endpoint:   endpoint,
		AccessKey:  accessKey,
		HttpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendEmail sends an email using Azure Communication Services Email REST API.
func (s *EmailSender) SendEmail(req EmailRequest) (*EmailResponse, error) {
	url := fmt.Sprintf("%s/emails:send?api-version=2023-03-31", s.Endpoint)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", s.AccessKey)

	resp, err := s.HttpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to send email: %s", string(b))
	}

	var emailResp EmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		return nil, err
	}
	return &emailResp, nil
}
