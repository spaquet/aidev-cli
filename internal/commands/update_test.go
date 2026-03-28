package commands

import (
	"runtime"
	"testing"
)

func TestUpdateCommand_ArchiveFormat(t *testing.T) {
	tests := []struct {
		goos     string
		expected string
	}{
		{"windows", ".zip"},
		{"linux", ".tar.gz"},
		{"darwin", ".tar.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			// Determine expected format based on OS
			ext := ".tar.gz"
			if tt.goos == "windows" {
				ext = ".zip"
			}

			if ext != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, ext)
			}
		})
	}
}

func TestUpdateCommand_CurrentOS(t *testing.T) {
	// Verify that the current runtime.GOOS is one of the supported ones
	supportedGOOS := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"windows": true,
	}

	if !supportedGOOS[runtime.GOOS] {
		t.Fatalf("current GOOS %q is not in supported list", runtime.GOOS)
	}
}
