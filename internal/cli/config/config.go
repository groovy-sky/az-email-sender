package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the CLI configuration
type Config struct {
	// Authentication
	Endpoint         string `mapstructure:"endpoint"`
	AccessKey        string `mapstructure:"access-key"`
	ConnectionString string `mapstructure:"connection-string"`

	// Email settings
	From    string `mapstructure:"from"`
	ReplyTo string `mapstructure:"reply-to"`

	// Output settings
	Debug bool `mapstructure:"debug"`
	Quiet bool `mapstructure:"quiet"`
	JSON  bool `mapstructure:"json"`

	// Wait settings
	Wait        bool   `mapstructure:"wait"`
	PollInterval string `mapstructure:"poll-interval"`
	MaxWaitTime  string `mapstructure:"max-wait-time"`
}

// Load loads configuration from file, environment variables, and command line flags
func Load(configFile string) (*Config, error) {
	v := viper.New()
	
	// Set defaults
	v.SetDefault("debug", false)
	v.SetDefault("quiet", false)
	v.SetDefault("json", false)
	v.SetDefault("wait", false)
	v.SetDefault("poll-interval", "5s")
	v.SetDefault("max-wait-time", "5m")

	// Environment variable setup
	v.SetEnvPrefix("AZURE_EMAIL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Bind environment variables explicitly for Unmarshal to work
	v.BindEnv("endpoint", "AZURE_EMAIL_ENDPOINT")
	v.BindEnv("access-key", "AZURE_EMAIL_ACCESS_KEY")
	v.BindEnv("connection-string", "AZURE_EMAIL_CONNECTION_STRING")
	v.BindEnv("from", "AZURE_EMAIL_FROM")
	v.BindEnv("reply-to", "AZURE_EMAIL_REPLY_TO")
	v.BindEnv("debug", "AZURE_EMAIL_DEBUG")
	v.BindEnv("quiet", "AZURE_EMAIL_QUIET")
	v.BindEnv("json", "AZURE_EMAIL_JSON")
	v.BindEnv("wait", "AZURE_EMAIL_WAIT")
	v.BindEnv("poll-interval", "AZURE_EMAIL_POLL_INTERVAL")
	v.BindEnv("max-wait-time", "AZURE_EMAIL_MAX_WAIT_TIME")

	// Load configuration file if specified
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}
	} else {
		// Look for config file in common locations
		v.SetConfigName("azemailsender")
		v.SetConfigType("json")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/azemailsender")
		v.AddConfigPath("/etc/azemailsender")
		
		// Try to read config file (ignore if not found)
		v.ReadInConfig()
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// SaveDefaultConfig creates a default configuration file
func SaveDefaultConfig(path string) error {
	defaultConfig := `{
  "endpoint": "https://your-resource.communication.azure.com",
  "access-key": "your-access-key",
  "from": "sender@yourdomain.com",
  "reply-to": "",
  "debug": false,
  "quiet": false,
  "json": false,
  "wait": false,
  "poll-interval": "5s",
  "max-wait-time": "5m"
}`

	return os.WriteFile(path, []byte(defaultConfig), 0644)
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