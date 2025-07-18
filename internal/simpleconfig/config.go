package simpleconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config represents the CLI configuration
type Config struct {
	// Authentication
	Endpoint         string `json:"endpoint"`
	AccessKey        string `json:"access-key"`
	ConnectionString string `json:"connection-string"`

	// Email settings
	From    string `json:"from"`
	ReplyTo string `json:"reply-to"`

	// Output settings
	Debug bool `json:"debug"`
	Quiet bool `json:"quiet"`
	JSON  bool `json:"json"`

	// Wait settings
	Wait         bool          `json:"wait"`
	PollInterval time.Duration `json:"poll-interval"`
	MaxWaitTime  time.Duration `json:"max-wait-time"`
}

// LoadConfig loads configuration with priority: defaults -> config file -> env vars -> CLI flags
func LoadConfig(configFile string, cliFlags map[string]interface{}) (*Config, error) {
	// Start with defaults
	config := &Config{
		Debug:        false,
		Quiet:        false,
		JSON:         false,
		Wait:         false,
		PollInterval: 5 * time.Second,
		MaxWaitTime:  5 * time.Minute,
	}

	// Load from config file (if exists)
	if err := loadFromFile(config, configFile); err != nil {
		return nil, err
	}

	// Override with environment variables
	loadFromEnv(config)

	// Override with CLI flags
	loadFromFlags(config, cliFlags)

	return config, nil
}

// loadFromFile loads configuration from JSON file
func loadFromFile(config *Config, configFile string) error {
	var filePath string

	if configFile != "" {
		filePath = configFile
	} else {
		// Look for config file in common locations
		searchPaths := []string{
			"./azemailsender.json",
			os.ExpandEnv("$HOME/.config/azemailsender/azemailsender.json"),
			"/etc/azemailsender/azemailsender.json",
		}

		for _, path := range searchPaths {
			if _, err := os.Stat(path); err == nil {
				filePath = path
				break
			}
		}
	}

	if filePath == "" {
		return nil // No config file found, that's OK
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if configFile != "" {
			// If explicitly specified, it's an error
			return fmt.Errorf("failed to read config file %s: %w", filePath, err)
		}
		// If auto-discovered, ignore the error
		return nil
	}

	// Parse durations as strings in JSON, then convert
	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", filePath, err)
	}

	// Convert back to JSON to unmarshal into struct (handles most fields)
	jsonData, _ := json.Marshal(rawConfig)
	if err := json.Unmarshal(jsonData, config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Handle duration fields manually
	if pollInterval, ok := rawConfig["poll-interval"].(string); ok {
		if d, err := time.ParseDuration(pollInterval); err == nil {
			config.PollInterval = d
		}
	}
	if maxWaitTime, ok := rawConfig["max-wait-time"].(string); ok {
		if d, err := time.ParseDuration(maxWaitTime); err == nil {
			config.MaxWaitTime = d
		}
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	envMap := map[string]*string{
		"AZURE_EMAIL_ENDPOINT":          &config.Endpoint,
		"AZURE_EMAIL_ACCESS_KEY":        &config.AccessKey,
		"AZURE_EMAIL_CONNECTION_STRING": &config.ConnectionString,
		"AZURE_EMAIL_FROM":              &config.From,
		"AZURE_EMAIL_REPLY_TO":          &config.ReplyTo,
	}

	for envVar, field := range envMap {
		if value := os.Getenv(envVar); value != "" {
			*field = value
		}
	}

	boolEnvMap := map[string]*bool{
		"AZURE_EMAIL_DEBUG": &config.Debug,
		"AZURE_EMAIL_QUIET": &config.Quiet,
		"AZURE_EMAIL_JSON":  &config.JSON,
		"AZURE_EMAIL_WAIT":  &config.Wait,
	}

	for envVar, field := range boolEnvMap {
		if value := os.Getenv(envVar); value != "" {
			*field = parseBool(value)
		}
	}

	// Duration environment variables
	if value := os.Getenv("AZURE_EMAIL_POLL_INTERVAL"); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			config.PollInterval = d
		}
	}
	if value := os.Getenv("AZURE_EMAIL_MAX_WAIT_TIME"); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			config.MaxWaitTime = d
		}
	}
}

// loadFromFlags loads configuration from CLI flags
func loadFromFlags(config *Config, flags map[string]interface{}) {
	if val, ok := flags["endpoint"].(string); ok && val != "" {
		config.Endpoint = val
	}
	if val, ok := flags["access-key"].(string); ok && val != "" {
		config.AccessKey = val
	}
	if val, ok := flags["connection-string"].(string); ok && val != "" {
		config.ConnectionString = val
	}
	if val, ok := flags["from"].(string); ok && val != "" {
		config.From = val
	}
	if val, ok := flags["reply-to"].(string); ok && val != "" {
		config.ReplyTo = val
	}
	if val, ok := flags["debug"].(bool); ok {
		config.Debug = val
	}
	if val, ok := flags["quiet"].(bool); ok {
		config.Quiet = val
	}
	if val, ok := flags["json"].(bool); ok {
		config.JSON = val
	}
	if val, ok := flags["wait"].(bool); ok {
		config.Wait = val
	}
	if val, ok := flags["poll-interval"].(time.Duration); ok && val > 0 {
		config.PollInterval = val
	}
	if val, ok := flags["max-wait-time"].(time.Duration); ok && val > 0 {
		config.MaxWaitTime = val
	}
}

// parseBool parses boolean from string
func parseBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

// SaveDefaultConfig creates a default configuration file
func SaveDefaultConfig(path string) error {
	defaultConfig := map[string]interface{}{
		"endpoint":       "https://your-resource.communication.azure.com",
		"access-key":     "your-access-key",
		"from":           "sender@yourdomain.com",
		"reply-to":       "",
		"debug":          false,
		"quiet":          false,
		"json":           false,
		"wait":           false,
		"poll-interval":  "5s",
		"max-wait-time":  "5m",
	}

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// GetEnvConfigExample returns example environment variable configuration
func GetEnvConfigExample() string {
	return `# Azure Communication Services Email Environment Variables
export AZURE_EMAIL_ENDPOINT="https://your-resource.communication.azure.com"
export AZURE_EMAIL_ACCESS_KEY="your-access-key"
export AZURE_EMAIL_FROM="sender@yourdomain.com"
export AZURE_EMAIL_REPLY_TO="reply@yourdomain.com"
export AZURE_EMAIL_DEBUG="false"
export AZURE_EMAIL_QUIET="false" 
export AZURE_EMAIL_JSON="false"`
}