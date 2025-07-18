package azemailsender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Send sends an email message and returns the response
func (c *Client) Send(message *EmailMessage) (*SendResponse, error) {
	return c.SendWithContext(context.Background(), message)
}

// SendWithContext sends an email message with context support
func (c *Client) SendWithContext(ctx context.Context, message *EmailMessage) (*SendResponse, error) {
	if c.options.Debug {
		c.logger.Printf("[DEBUG] Starting email send process")
		c.logger.Printf("[DEBUG] From: %s", message.SenderAddress)
		c.logger.Printf("[DEBUG] Subject: %s", message.Content.Subject)
	}

	startTime := time.Now()

	// Serialize the message
	body, err := json.Marshal(message)
	if err != nil {
		if c.options.Debug {
			c.logger.Printf("[DEBUG] Failed to marshal message: %v", err)
		}
		return nil, fmt.Errorf("failed to marshal email message: %w", err)
	}

	if c.options.Debug {
		c.logger.Printf("[DEBUG] Message serialized (%d bytes)", len(body))
	}

	// Build the URL
	url := fmt.Sprintf("%s/emails:send?api-version=%s", c.endpoint, c.options.APIVersion)

	if c.options.Debug {
		c.logger.Printf("[DEBUG] API URL: %s", url)
	}

	// Attempt to send with retries
	var lastErr error
	for attempt := 0; attempt <= c.options.MaxRetries; attempt++ {
		if attempt > 0 {
			if c.options.Debug {
				c.logger.Printf("[DEBUG] Retry attempt %d/%d", attempt, c.options.MaxRetries)
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.options.RetryDelay):
				// Continue with retry
			}
		}

		response, err := c.sendSingleAttempt(ctx, url, body)
		if err == nil {
			duration := time.Since(startTime)
			if c.options.Debug {
				c.logger.Printf("[DEBUG] Email sent successfully in %v", duration)
			}

			// Set legacy MessageID for backward compatibility
			response.MessageID = response.ID
			response.Timestamp = time.Now()

			return response, nil
		}

		lastErr = err
		if c.options.Debug {
			c.logger.Printf("[DEBUG] Send attempt %d failed: %v", attempt+1, err)
		}
	}

	return nil, fmt.Errorf("failed to send email after %d attempts: %w", c.options.MaxRetries+1, lastErr)
}

// sendSingleAttempt performs a single send attempt
func (c *Client) sendSingleAttempt(ctx context.Context, url string, body []byte) (*SendResponse, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "azemailsender-go/1.0")

	if c.options.Debug {
		c.logger.Printf("[DEBUG] HTTP Request:")
		c.logger.Printf("[DEBUG]   Method: %s", req.Method)
		c.logger.Printf("[DEBUG]   URL: %s", req.URL.String())
		c.logger.Printf("[DEBUG]   Content-Type: %s", req.Header.Get("Content-Type"))
		c.logger.Printf("[DEBUG]   Body size: %d bytes", len(body))
	}

	// Add authentication
	if err := c.addAuthentication(req, string(body)); err != nil {
		return nil, fmt.Errorf("failed to add authentication: %w", err)
	}

	// Send request
	reqStartTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	requestDuration := time.Since(reqStartTime)

	if c.options.Debug {
		c.logger.Printf("[DEBUG] HTTP Response:")
		c.logger.Printf("[DEBUG]   Status: %s (%d)", resp.Status, resp.StatusCode)
		c.logger.Printf("[DEBUG]   Request duration: %v", requestDuration)
		c.logger.Printf("[DEBUG]   Content-Length: %s", resp.Header.Get("Content-Length"))
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.options.Debug {
		c.logger.Printf("[DEBUG]   Response body: %s", string(respBody))
	}

	// Check for success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError Error
		if err := json.Unmarshal(respBody, &apiError); err != nil {
			// If we can't parse the error, return the raw response
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, apiError.Message)
	}

	// Parse response
	var sendResponse SendResponse
	if err := json.Unmarshal(respBody, &sendResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &sendResponse, nil
}

// GetStatus retrieves the status of a sent email
func (c *Client) GetStatus(messageID string) (*StatusResponse, error) {
	return c.GetStatusWithContext(context.Background(), messageID)
}

// GetStatusWithContext retrieves the status of a sent email with context support
func (c *Client) GetStatusWithContext(ctx context.Context, messageID string) (*StatusResponse, error) {
	if c.options.Debug {
		c.logger.Printf("[DEBUG] Checking status for message ID: %s", messageID)
	}

	url := fmt.Sprintf("%s/emails/operations/%s?api-version=%s", c.endpoint, messageID, c.options.APIVersion)

	if c.options.Debug {
		c.logger.Printf("[DEBUG] Status check URL: %s", url)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	req.Header.Set("User-Agent", "azemailsender-go/1.0")

	// Add authentication
	if err := c.addAuthentication(req, ""); err != nil {
		return nil, fmt.Errorf("failed to add authentication: %w", err)
	}

	reqStartTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("status request failed: %w", err)
	}
	defer resp.Body.Close()

	requestDuration := time.Since(reqStartTime)

	if c.options.Debug {
		c.logger.Printf("[DEBUG] Status check response: %s (duration: %v)", resp.Status, requestDuration)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read status response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if c.options.Debug {
			c.logger.Printf("[DEBUG] Status check failed: %s", string(respBody))
		}
		return nil, fmt.Errorf("status check failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var statusResponse StatusResponse
	if err := json.Unmarshal(respBody, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	statusResponse.Timestamp = time.Now()

	if c.options.Debug {
		c.logger.Printf("[DEBUG] Current status: %s", statusResponse.Status)
	}

	return &statusResponse, nil
}

// WaitForCompletion waits for an email to reach a final status
func (c *Client) WaitForCompletion(messageID string, options *WaitOptions) (*StatusResponse, error) {
	return c.WaitForCompletionWithContext(context.Background(), messageID, options)
}

// WaitForCompletionWithContext waits for an email to reach a final status with context support
func (c *Client) WaitForCompletionWithContext(ctx context.Context, messageID string, options *WaitOptions) (*StatusResponse, error) {
	if options == nil {
		options = DefaultWaitOptions()
	}

	if c.options.Debug {
		c.logger.Printf("[DEBUG] Starting status polling for message ID: %s", messageID)
		c.logger.Printf("[DEBUG] Poll interval: %v", options.PollInterval)
		c.logger.Printf("[DEBUG] Max wait time: %v", options.MaxWaitTime)
	}

	ctx, cancel := context.WithTimeout(ctx, options.MaxWaitTime)
	defer cancel()

	ticker := time.NewTicker(options.PollInterval)
	defer ticker.Stop()

	attempt := 0

	for {
		attempt++
		if c.options.Debug {
			c.logger.Printf("[DEBUG] Status polling attempt %d", attempt)
		}

		status, err := c.GetStatusWithContext(ctx, messageID)
		if err != nil {
			if c.options.Debug {
				c.logger.Printf("[DEBUG] Status check failed: %v", err)
			}
			if options.OnError != nil {
				options.OnError(err)
			}

			// Don't fail immediately on status check errors, continue polling
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ticker.C:
				continue
			}
		}

		if options.OnStatusUpdate != nil {
			options.OnStatusUpdate(status)
		}

		// Check if we've reached a final status
		if isFinalStatus(status.Status) {
			if c.options.Debug {
				c.logger.Printf("[DEBUG] Final status reached: %s (after %d attempts)", status.Status, attempt)
			}
			return status, nil
		}

		if c.options.Debug {
			c.logger.Printf("[DEBUG] Status still pending: %s", status.Status)
		}

		select {
		case <-ctx.Done():
			if c.options.Debug {
				c.logger.Printf("[DEBUG] Polling timed out after %d attempts", attempt)
			}
			return status, ctx.Err()
		case <-ticker.C:
			// Continue polling
		}
	}
}

// isFinalStatus checks if the given status is a final status
func isFinalStatus(status string) bool {
	finalStatuses := []EmailStatus{
		StatusDelivered,
		StatusFailed,
		StatusCanceled,
	}

	for _, finalStatus := range finalStatuses {
		if status == string(finalStatus) {
			return true
		}
	}

	return false
}
