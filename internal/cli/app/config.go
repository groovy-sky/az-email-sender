package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	Wait         bool   `json:"wait"`
	PollInterval string `json:"poll-interval"`
	MaxWaitTime  string `json:"max-wait-time"`
}

// LoadConfig loads configuration from file, environment variables, and command line flags
func LoadConfig(configFile string) (*Config, error) {
	// Start with defaults
	config := &Config{
		Debug:        false,
		Quiet:        false,
		JSON:         false,
		Wait:         false,
		PollInterval: "5s",
		MaxWaitTime:  "5m",
	}

	// Load from config file if specified or found
	if err := loadConfigFile(config, configFile); err != nil {
		return nil, err
	}

	// Override with environment variables
	loadConfigFromEnv(config)

	return config, nil
}

// loadConfigFile loads configuration from a JSON file
func loadConfigFile(config *Config, configFile string) error {
	var filePath string

	if configFile != "" {
		// Use specified config file
		filePath = configFile
	} else {
		// Look for config file in common locations
		searchPaths := []string{
			"./azemailsender.json",
			filepath.Join(os.Getenv("HOME"), ".config", "azemailsender", "azemailsender.json"),
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
		// No config file found, which is okay
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if configFile != "" {
			// User specified a config file that doesn't exist
			return fmt.Errorf("failed to read config file %s: %w", filePath, err)
		}
		// Auto-discovered file doesn't exist, which is okay
		return nil
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", filePath, err)
	}

	return nil
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv(config *Config) {
	// Authentication
	if val := os.Getenv("AZURE_EMAIL_ENDPOINT"); val != "" {
		config.Endpoint = val
	}
	if val := os.Getenv("AZURE_EMAIL_ACCESS_KEY"); val != "" {
		config.AccessKey = val
	}
	if val := os.Getenv("AZURE_EMAIL_CONNECTION_STRING"); val != "" {
		config.ConnectionString = val
	}

	// Email settings
	if val := os.Getenv("AZURE_EMAIL_FROM"); val != "" {
		config.From = val
	}
	if val := os.Getenv("AZURE_EMAIL_REPLY_TO"); val != "" {
		config.ReplyTo = val
	}

	// Output settings
	if val := os.Getenv("AZURE_EMAIL_DEBUG"); val != "" {
		config.Debug = parseBool(val)
	}
	if val := os.Getenv("AZURE_EMAIL_QUIET"); val != "" {
		config.Quiet = parseBool(val)
	}
	if val := os.Getenv("AZURE_EMAIL_JSON"); val != "" {
		config.JSON = parseBool(val)
	}

	// Wait settings
	if val := os.Getenv("AZURE_EMAIL_WAIT"); val != "" {
		config.Wait = parseBool(val)
	}
	if val := os.Getenv("AZURE_EMAIL_POLL_INTERVAL"); val != "" {
		config.PollInterval = val
	}
	if val := os.Getenv("AZURE_EMAIL_MAX_WAIT_TIME"); val != "" {
		config.MaxWaitTime = val
	}
}

// parseBool parses a string as a boolean value
func parseBool(s string) bool {
	val, err := strconv.ParseBool(s)
	if err != nil {
		// Try to parse common string representations
		switch strings.ToLower(s) {
		case "1", "yes", "on", "enable", "enabled":
			return true
		case "0", "no", "off", "disable", "disabled":
			return false
		}
	}
	return val
}

// SaveDefaultConfig creates a default configuration file
func SaveDefaultConfig(path string) error {
	defaultConfig := Config{
		Endpoint:         "https://your-resource.communication.azure.com",
		AccessKey:        "your-access-key",
		ConnectionString: "",
		From:             "sender@yourdomain.com",
		ReplyTo:          "",
		Debug:            false,
		Quiet:            false,
		JSON:             false,
		Wait:             false,
		PollInterval:     "5s",
		MaxWaitTime:      "5m",
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