package config

import (
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	CredentialsPath string
	TokenPath       string
	CalendarID      string
	Timezone        string
	DryRun          bool
	Verbose         bool
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		CredentialsPath: "credentials.json",
		TokenPath:       getDefaultTokenPath(),
		CalendarID:      "primary",
		Timezone:        "local",
		DryRun:          false,
		Verbose:         false,
	}
}

// getDefaultTokenPath returns platform-appropriate token storage path
func getDefaultTokenPath() string {
	// Try to use user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fall back to current directory
		return "token.json"
	}

	tokenDir := filepath.Join(configDir, "calendar-event-generator")
	return filepath.Join(tokenDir, "token.json")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check if credentials file exists (only if not dry-run)
	if !c.DryRun {
		if _, err := os.Stat(c.CredentialsPath); os.IsNotExist(err) {
			// Don't error here, let the auth module handle the friendly error message
		}
	}

	return nil
}
