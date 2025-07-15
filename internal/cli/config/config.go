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
	AccessKey        string `mapstructure:"access_key"`
	ConnectionString string `mapstructure:"connection_string"`

	// Email settings
	From    string `mapstructure:"from"`
	ReplyTo string `mapstructure:"reply_to"`

	// Output settings
	Debug bool `mapstructure:"debug"`
	Quiet bool `mapstructure:"quiet"`
	JSON  bool `mapstructure:"json"`

	// Wait settings
	Wait        bool   `mapstructure:"wait"`
	PollInterval string `mapstructure:"poll_interval"`
	MaxWaitTime  string `mapstructure:"max_wait_time"`
}

// Load loads configuration from file, environment variables, and command line flags
func Load(configFile string) (*Config, error) {
	v := viper.New()
	
	// Set defaults
	v.SetDefault("debug", false)
	v.SetDefault("quiet", false)
	v.SetDefault("json", false)
	v.SetDefault("wait", false)
	v.SetDefault("poll_interval", "5s")
	v.SetDefault("max_wait_time", "5m")

	// Environment variable setup
	v.SetEnvPrefix("AZURE_EMAIL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

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
  "access_key": "your-access-key",
  "from": "sender@yourdomain.com",
  "debug": false,
  "quiet": false,
  "json": false,
  "wait": false,
  "poll_interval": "5s",
  "max_wait_time": "5m"
}`

	return os.WriteFile(path, []byte(defaultConfig), 0644)
}

// GetEnvConfigExample returns example environment variable configuration
func GetEnvConfigExample() string {
	return `# Azure Communication Services Email Environment Variables
export AZURE_EMAIL_ENDPOINT="https://your-resource.communication.azure.com"
export AZURE_EMAIL_ACCESS_KEY="your-access-key"
export AZURE_EMAIL_FROM="sender@yourdomain.com"
export AZURE_EMAIL_DEBUG="false"
export AZURE_EMAIL_QUIET="false" 
export AZURE_EMAIL_JSON="false"`
}