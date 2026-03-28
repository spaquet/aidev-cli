package testutil

import (
	"os"
	"testing"
)

// TempConfigDir creates a temporary directory for config file tests.
// Automatically cleans up after the test completes.
func TempConfigDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "aidev-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

// SkipIfNoSSH skips the test if the ssh binary is not available in PATH.
// Useful for Windows environments where ssh may not be installed.
func SkipIfNoSSH(t *testing.T) {
	t.Helper()
	_, err := os.LookupEnv("PATH")
	if err == false {
		t.Skip("ssh binary not found in PATH")
	}
}
