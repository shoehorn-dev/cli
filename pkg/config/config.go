package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	Version        string              `yaml:"version"`
	CurrentProfile string              `yaml:"current_profile"`
	Profiles       map[string]*Profile `yaml:"profiles"`
}

// Profile represents an authentication profile
type Profile struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
	Auth   *Auth  `yaml:"auth,omitempty"`
}

// Auth contains authentication credentials
type Auth struct {
	ProviderType string    `yaml:"provider_type"`
	Issuer       string    `yaml:"issuer"`
	ClientID     string    `yaml:"client_id"`
	AccessToken  string    `yaml:"access_token,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
	TokenType    string    `yaml:"token_type,omitempty"`
	ExpiresAt    time.Time `yaml:"expires_at,omitempty"`
	User         *User     `yaml:"user,omitempty"`
}

// User contains user information from token
type User struct {
	Email    string `yaml:"email"`
	Name     string `yaml:"name,omitempty"`
	TenantID string `yaml:"tenant_id,omitempty"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".shoehorn", "config.yaml"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".shoehorn")
	return os.MkdirAll(configDir, 0700)
}

// Load reads the config file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Version:        "1.0",
			CurrentProfile: "default",
			Profiles: map[string]*Profile{
				"default": {
					Name:   "Default",
					Server: "http://localhost:8080",
				},
			},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the config file
func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// GetCurrentProfile returns the current active profile
func (c *Config) GetCurrentProfile() (*Profile, error) {
	profile, ok := c.Profiles[c.CurrentProfile]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found", c.CurrentProfile)
	}
	return profile, nil
}

// SetProfile updates or creates a profile
func (c *Config) SetProfile(name string, profile *Profile) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]*Profile)
	}
	c.Profiles[name] = profile
}

// IsAuthenticated checks if the current profile has valid auth
func (c *Config) IsAuthenticated() bool {
	profile, err := c.GetCurrentProfile()
	if err != nil {
		return false
	}
	return profile.Auth != nil && profile.Auth.AccessToken != ""
}

// IsTokenExpired checks if the access token is expired
func (c *Config) IsTokenExpired() bool {
	profile, err := c.GetCurrentProfile()
	if err != nil {
		return true
	}
	if profile.Auth == nil {
		return true
	}
	return time.Now().After(profile.Auth.ExpiresAt)
}
