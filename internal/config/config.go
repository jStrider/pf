package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DefaultStore     string                  `yaml:"default_store"`
	Stores          map[string]StoreConfig  `yaml:"stores"`
	AgeKeyPath      string                  `yaml:"age_key_path"`
	ClipboardTimeout time.Duration          `yaml:"clipboard_timeout"`
	AuditLog        bool                    `yaml:"audit_log"`
}

// StoreConfig represents a password store configuration
type StoreConfig struct {
	Path       string   `yaml:"path"`
	Recipients []string `yaml:"recipients"`
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	cfg := &Config{
		ClipboardTimeout: 45 * time.Second,
		Stores:          make(map[string]StoreConfig),
	}

	// Get config path
	configPath := cfg.GetConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Unmarshal YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if cfg.AgeKeyPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		cfg.AgeKeyPath = filepath.Join(home, ".pf", "age-key.txt")
	}

	return cfg, nil
}

// GetConfigPath returns the configuration file path
func (c *Config) GetConfigPath() string {
	// Check viper for custom config path
	if configPath := viper.GetString("config"); configPath != "" {
		return configPath
	}

	// Check environment variable
	if configPath := os.Getenv("PF_CONFIG"); configPath != "" {
		return configPath
	}

	// Default to ~/.pf/config.yaml
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pf", "config.yaml")
}