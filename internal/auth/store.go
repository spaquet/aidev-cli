package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/aidev/cli/internal/models"
)

const configFileName = "config.json"

var errNoConfig = errors.New("config file not found")

// Store handles reading and writing the config file
type Store struct {
	configPath string
}

// NewStore creates a new auth store using XDG config directories
func NewStore() (*Store, error) {
	configDir := filepath.Join(xdg.ConfigHome, "aidev")
	configPath := filepath.Join(configDir, configFileName)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	return &Store{configPath: configPath}, nil
}

// Load reads the config from disk
func (s *Store) Load() (*models.Config, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errNoConfig
		}
		return nil, err
	}

	var config models.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save writes the config to disk with restricted permissions (0600)
func (s *Store) Save(config *models.Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write with 0600 permissions (owner read/write only)
	if err := os.WriteFile(s.configPath, data, 0600); err != nil {
		return err
	}

	return nil
}

// Delete removes the config file
func (s *Store) Delete() error {
	if err := os.Remove(s.configPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsTokenExpired checks if the stored token is past expiration
func (s *Store) IsTokenExpired(config *models.Config) bool {
	if config == nil || config.TokenExpiresAt == "" {
		return true
	}

	expTime, err := time.Parse(time.RFC3339, config.TokenExpiresAt)
	if err != nil {
		return true
	}

	// Expire 1 minute before actual expiration to allow refresh
	return time.Now().After(expTime.Add(-1 * time.Minute))
}

// IsNoConfigError checks if the error is "no config file"
func IsNoConfigError(err error) bool {
	return errors.Is(err, errNoConfig)
}
