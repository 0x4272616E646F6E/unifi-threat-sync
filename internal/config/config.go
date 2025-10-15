package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the entire application configuration
type Config struct {
	UniFi  UniFiConfig  `yaml:"unifi"`
	Sync   SyncConfig   `yaml:"sync"`
	Feeds  FeedsList    `yaml:"feeds"`
	Health HealthConfig `yaml:"health"`
}

// UniFiConfig holds UniFi controller settings
type UniFiConfig struct {
	URL       string `yaml:"url"`
	Site      string `yaml:"site"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	GroupName string `yaml:"groupName"`
	Ruleset   string `yaml:"ruleset"`
	RuleIndex int    `yaml:"ruleIndex"`
}

// SyncConfig holds synchronization settings
type SyncConfig struct {
	Interval time.Duration `yaml:"interval"`
}

// HealthConfig holds health check server settings
type HealthConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// FeedConfig represents a single threat feed configuration
type FeedConfig struct {
	Name    string                 `yaml:"name"`
	URL     string                 `yaml:"url"`
	Parser  string                 `yaml:"parser"`
	Enabled bool                   `yaml:"enabled"`
	Timeout string                 `yaml:"timeout"`
	Auth    map[string]interface{} `yaml:"auth"`
	Params  map[string]interface{} `yaml:"params"`
}

// FeedsList is a slice of FeedConfig with helper methods
type FeedsList []FeedConfig

// EnabledCount returns the number of enabled feeds
func (f FeedsList) EnabledCount() int {
	count := 0
	for _, feed := range f {
		if feed.Enabled {
			count++
		}
	}
	return count
}

// GetEnabled returns only enabled feeds
func (f FeedsList) GetEnabled() []FeedConfig {
	enabled := make([]FeedConfig, 0, len(f))
	for _, feed := range f {
		if feed.Enabled {
			enabled = append(enabled, feed)
		}
	}
	return enabled
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	cfg.setDefaults()

	return &cfg, nil
}

// setDefaults sets default values for optional fields
func (c *Config) setDefaults() {
	// UniFi defaults
	if c.UniFi.Site == "" {
		c.UniFi.Site = "default"
	}
	if c.UniFi.GroupName == "" {
		c.UniFi.GroupName = "uts-block-list"
	}
	if c.UniFi.Ruleset == "" {
		c.UniFi.Ruleset = "WAN_OUT"
	}
	if c.UniFi.RuleIndex == 0 {
		c.UniFi.RuleIndex = 2000
	}

	// Sync defaults
	if c.Sync.Interval == 0 {
		c.Sync.Interval = 60 * time.Minute
	}

	// Health defaults
	if c.Health.Port == 0 {
		c.Health.Port = 8080
	}

	// Feed defaults
	for i := range c.Feeds {
		// Default enabled to true if not specified
		if c.Feeds[i].Timeout == "" {
			c.Feeds[i].Timeout = "30s"
		}
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate UniFi config
	if c.UniFi.URL == "" {
		return fmt.Errorf("unifi.url is required")
	}
	if !strings.HasPrefix(c.UniFi.URL, "http://") && !strings.HasPrefix(c.UniFi.URL, "https://") {
		return fmt.Errorf("unifi.url must start with http:// or https://")
	}
	if c.UniFi.Username == "" {
		return fmt.Errorf("unifi.username is required")
	}
	if c.UniFi.Password == "" {
		return fmt.Errorf("unifi.password is required")
	}

	// Validate sync config
	if c.Sync.Interval < time.Minute {
		return fmt.Errorf("sync.interval must be at least 1 minute")
	}

	// Validate feeds
	if len(c.Feeds) == 0 {
		return fmt.Errorf("at least one feed must be configured")
	}

	enabledCount := 0
	for i, feed := range c.Feeds {
		if !feed.Enabled {
			continue
		}
		enabledCount++

		if feed.Name == "" {
			return fmt.Errorf("feed[%d].name is required", i)
		}
		if feed.URL == "" {
			return fmt.Errorf("feed[%d].url is required", i)
		}
		if feed.Parser == "" {
			return fmt.Errorf("feed[%d].parser is required", i)
		}
		if !strings.HasPrefix(feed.URL, "http://") && !strings.HasPrefix(feed.URL, "https://") {
			return fmt.Errorf("feed[%d].url must start with http:// or https://", i)
		}
	}

	if enabledCount == 0 {
		return fmt.Errorf("at least one feed must be enabled")
	}

	return nil
}
