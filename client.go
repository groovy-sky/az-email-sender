package azemailsender

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents the Azure Communication Services Email client
type Client struct {
	endpoint   string
	accessKey  string
	authMethod AuthMethod
	options    *ClientOptions
	httpClient *http.Client
	logger     Logger
}

// NewClient creates a new email client with endpoint and access key
func NewClient(endpoint, accessKey string, options *ClientOptions) *Client {
	if options == nil {
		options = DefaultClientOptions()
	}

	// Ensure API version is set
	if options.APIVersion == "" {
		options.APIVersion = DefaultAPIVersion
	}

	// Ensure logger is set
	if options.Logger == nil {
		options.Logger = &noOpLogger{}
	}

	client := &Client{
		endpoint:   strings.TrimSuffix(endpoint, "/"),
		accessKey:  accessKey,
		authMethod: AuthMethodHMAC,
		options:    options,
		logger:     options.Logger,
		httpClient: &http.Client{
			Timeout: options.HTTPTimeout,
		},
	}

	if client.options.Debug {
		client.logger.Printf("[DEBUG] Client initialized with endpoint: %s", client.endpoint)
		client.logger.Printf("[DEBUG] Authentication method: HMAC-SHA256")
		client.logger.Printf("[DEBUG] API Version: %s", client.options.APIVersion)
		client.logger.Printf("[DEBUG] HTTP Timeout: %v", client.options.HTTPTimeout)
		client.logger.Printf("[DEBUG] Max Retries: %d", client.options.MaxRetries)
	}

	return client
}

// NewClientFromConnectionString creates a new email client from a connection string
func NewClientFromConnectionString(connectionString string, options *ClientOptions) (*Client, error) {
	if options == nil {
		options = DefaultClientOptions()
	}

	// Ensure API version is set
	if options.APIVersion == "" {
		options.APIVersion = DefaultAPIVersion
	}

	parsed, err := parseConnectionString(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	client := NewClient(parsed.Endpoint, parsed.AccessKey, options)
	client.authMethod = AuthMethodConnectionString

	if client.options.Debug {
		client.logger.Printf("[DEBUG] Client created from connection string")
		client.logger.Printf("[DEBUG] Parsed endpoint: %s", parsed.Endpoint)
	}

	return client, nil
}

// NewClientWithAccessKey creates a new email client using access key authentication (legacy)
func NewClientWithAccessKey(endpoint, accessKey string, options *ClientOptions) *Client {
	if options == nil {
		options = DefaultClientOptions()
	}

	client := NewClient(endpoint, accessKey, options)
	client.authMethod = AuthMethodAccessKey

	if client.options.Debug {
		client.logger.Printf("[DEBUG] Client created with access key authentication (legacy)")
	}

	return client
}

// parseConnectionString parses an Azure Communication Services connection string
func parseConnectionString(connectionString string) (*ParsedConnectionString, error) {
	parts := strings.Split(connectionString, ";")
	parsed := &ParsedConnectionString{}

	for _, part := range parts {
		if strings.HasPrefix(part, "endpoint=") {
			parsed.Endpoint = strings.TrimPrefix(part, "endpoint=")
		} else if strings.HasPrefix(part, "accesskey=") {
			parsed.AccessKey = strings.TrimPrefix(part, "accesskey=")
		}
	}

	if parsed.Endpoint == "" {
		return nil, fmt.Errorf("endpoint not found in connection string")
	}

	if parsed.AccessKey == "" {
		return nil, fmt.Errorf("access key not found in connection string")
	}

	return parsed, nil
}

// generateHMACSignature generates HMAC-SHA256 signature for Azure API authentication
func (c *Client) generateHMACSignature(method, uri, host, dateHeader, body string) string {
	if c.options.Debug {
		c.logger.Printf("[DEBUG] Generating HMAC signature")
		c.logger.Printf("[DEBUG] Method: %s", method)
		c.logger.Printf("[DEBUG] URI: %s", uri)
		c.logger.Printf("[DEBUG] Host: %s", host)
		c.logger.Printf("[DEBUG] Date: %s", dateHeader)
		c.logger.Printf("[DEBUG] Body length: %d bytes", len(body))
	}

	// Create string to sign
	stringToSign := fmt.Sprintf("%s\n%s\n%s;%s;%s", method, uri, host, dateHeader, body)

	if c.options.Debug {
		c.logger.Printf("[DEBUG] String to sign: %s", stringToSign)
	}

	// Decode the access key
	decodedKey, err := base64.StdEncoding.DecodeString(c.accessKey)
	if err != nil {
		if c.options.Debug {
			c.logger.Printf("[DEBUG] Failed to decode access key: %v", err)
		}
		return ""
	}

	// Create HMAC
	h := hmac.New(sha256.New, decodedKey)
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if c.options.Debug {
		c.logger.Printf("[DEBUG] Generated signature: %s", signature)
	}

	return signature
}

// addAuthentication adds authentication headers to the HTTP request
func (c *Client) addAuthentication(req *http.Request, body string) error {
	if c.options.Debug {
		c.logger.Printf("[DEBUG] Adding authentication headers (method: %v)", c.authMethod)
	}

	switch c.authMethod {
	case AuthMethodAccessKey:
		// Legacy API key authentication
		req.Header.Set("api-key", c.accessKey)
		if c.options.Debug {
			c.logger.Printf("[DEBUG] Added api-key header")
		}
	case AuthMethodHMAC, AuthMethodConnectionString:
		// HMAC-SHA256 authentication
		dateHeader := time.Now().UTC().Format(time.RFC1123)
		req.Header.Set("Date", dateHeader)

		parsedURL, err := url.Parse(req.URL.String())
		if err != nil {
			return fmt.Errorf("failed to parse URL: %w", err)
		}

		signature := c.generateHMACSignature(req.Method, parsedURL.Path+"?"+parsedURL.RawQuery, parsedURL.Host, dateHeader, body)

		authHeader := fmt.Sprintf("HMAC-SHA256 SignedHeaders=date;host;x-ms-content-sha256&Signature=%s", signature)
		req.Header.Set("Authorization", authHeader)

		// Add content hash
		h := sha256.New()
		h.Write([]byte(body))
		contentHash := base64.StdEncoding.EncodeToString(h.Sum(nil))
		req.Header.Set("x-ms-content-sha256", contentHash)

		if c.options.Debug {
			c.logger.Printf("[DEBUG] Added HMAC-SHA256 authentication headers")
			c.logger.Printf("[DEBUG] Authorization: %s", authHeader)
			c.logger.Printf("[DEBUG] Content hash: %s", contentHash)
		}
	default:
		return fmt.Errorf("unsupported authentication method: %v", c.authMethod)
	}

	return nil
}

// SetDebug enables or disables debug logging at runtime
func (c *Client) SetDebug(enabled bool) {
	c.options.Debug = enabled
	if enabled {
		c.logger.Printf("[DEBUG] Debug logging enabled")
	}
}

// SetLogger sets a custom logger for the client
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
	c.options.Logger = logger
	if c.options.Debug {
		c.logger.Printf("[DEBUG] Custom logger set")
	}
}

// noOpLogger is a logger that does nothing
type noOpLogger struct{}

func (l *noOpLogger) Printf(format string, v ...interface{}) {
	// Do nothing
}
