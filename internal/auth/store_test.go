package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aidev/cli/internal/models"
	"github.com/aidev/cli/internal/testutil"
)

func TestStore_SaveAndLoad(t *testing.T) {
	tmpDir := testutil.TempConfigDir(t)
	storePath := filepath.Join(tmpDir, "config.json")

	// Use reflection to set private field
	store := &Store{configPath: storePath}

	// Create config data
	config := &models.Config{
		BaseURL:        "https://api.example.com",
		Token:          "sk_test_123456",
		TokenExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		UserEmail:      "test@example.com",
	}

	// Save
	err := store.Save(config)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(storePath); err != nil {
		t.Fatalf("store file not created: %v", err)
	}

	// Load
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify loaded data
	if loaded.UserEmail != config.UserEmail {
		t.Errorf("expected UserEmail=%s, got %s", config.UserEmail, loaded.UserEmail)
	}
	if loaded.Token != config.Token {
		t.Errorf("expected Token=%s, got %s", config.Token, loaded.Token)
	}
	if loaded.BaseURL != config.BaseURL {
		t.Errorf("expected BaseURL=%s, got %s", config.BaseURL, loaded.BaseURL)
	}
}

func TestStore_LoadNonExistent(t *testing.T) {
	tmpDir := testutil.TempConfigDir(t)
	storePath := filepath.Join(tmpDir, "nonexistent.json")

	store := &Store{configPath: storePath}
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error loading nonexistent file, got nil")
	}
	if !IsNoConfigError(err) {
		t.Errorf("expected IsNoConfigError, got %v", err)
	}
}

func TestStore_IsTokenExpired(t *testing.T) {
	tmpDir := testutil.TempConfigDir(t)
	storePath := filepath.Join(tmpDir, "config.json")
	store := &Store{configPath: storePath}

	// Test expired token
	expiredConfig := &models.Config{
		Token:          "sk_expired",
		TokenExpiresAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
	}
	if !store.IsTokenExpired(expiredConfig) {
		t.Error("expected token to be expired")
	}

	// Test valid token
	validConfig := &models.Config{
		Token:          "sk_valid",
		TokenExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}
	if store.IsTokenExpired(validConfig) {
		t.Error("expected token to be valid")
	}

	// Test nil config
	if !store.IsTokenExpired(nil) {
		t.Error("expected nil config to be expired")
	}
}
