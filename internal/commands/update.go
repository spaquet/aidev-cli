package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// NewUpdateCmd creates the update subcommand
func NewUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update aidev to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleUpdate()
		},
	}
}

func handleUpdate() error {
	currentVersion := "0.1.0" // From main.version at build time

	// Fetch latest version from GitHub
	fmt.Println("Checking for updates...")

	latestVersion, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !isNewVersionAvailable(currentVersion, latestVersion) {
		fmt.Printf("You are already running the latest version (%s)\n", currentVersion)
		return nil
	}

	fmt.Printf("Update available: %s → %s\n", currentVersion, latestVersion)

	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Download new version
	fmt.Println("Downloading update...")
	newBinary, err := downloadLatest(latestVersion)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(newBinary)

	// Make executable
	if err := os.Chmod(newBinary, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Backup current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Replace binary
	if err := os.Rename(newBinary, execPath); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Clean up backup
	os.Remove(backupPath)

	fmt.Printf("✓ Successfully updated to version %s\n", latestVersion)
	return nil
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getLatestVersion() (string, error) {
	url := "https://api.github.com/repos/aidev/aidev-cli/releases/latest"

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

func downloadLatest(version string) (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Convert to goreleaser naming
	if goarch == "amd64" {
		goarch = "amd64"
	} else if goarch == "arm64" {
		goarch = "arm64"
	}

	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}

	filename := fmt.Sprintf("aidev_%s_%s_%s%s", version, goos, goarch, ext)
	downloadURL := fmt.Sprintf(
		"https://github.com/aidev/aidev-cli/releases/download/v%s/%s",
		version, filename,
	)

	// Create temp file
	tmpFile, err := os.CreateTemp("", "aidev-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	// Download
	resp, err := http.Get(downloadURL)
	if err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpPath)
		return "", fmt.Errorf("download failed: %d", resp.StatusCode)
	}

	// Write to temp file
	out, err := os.Create(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()

	if err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	// Extract
	extractDir, err := os.MkdirTemp("", "aidev-extract-*")
	if err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	if goos == "windows" {
		// ZIP extraction (would need third-party lib or system unzip)
		// For now, use system unzip
		cmd := exec.Command("unzip", "-q", tmpPath, "-d", extractDir)
		if err := cmd.Run(); err != nil {
			os.RemoveAll(extractDir)
			os.Remove(tmpPath)
			return "", fmt.Errorf("failed to extract: %w", err)
		}
	} else {
		// TAR.GZ extraction
		cmd := exec.Command("tar", "-xzf", tmpPath, "-C", extractDir)
		if err := cmd.Run(); err != nil {
			os.RemoveAll(extractDir)
			os.Remove(tmpPath)
			return "", fmt.Errorf("failed to extract: %w", err)
		}
	}

	os.Remove(tmpPath)

	// Find binary
	var binaryPath string
	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "aidev" || info.Name() == "aidev.exe" {
			binaryPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil || binaryPath == "" {
		os.RemoveAll(extractDir)
		return "", fmt.Errorf("binary not found in archive")
	}

	// Copy to temp location (for atomic replacement)
	finalTmpPath, err := os.CreateTemp("", "aidev-binary-*")
	if err != nil {
		os.RemoveAll(extractDir)
		return "", err
	}
	finalTmpPath.Close()

	src, err := os.Open(binaryPath)
	if err != nil {
		os.RemoveAll(extractDir)
		os.Remove(finalTmpPath.Name())
		return "", err
	}

	dst, err := os.Create(finalTmpPath.Name())
	if err != nil {
		src.Close()
		os.RemoveAll(extractDir)
		return "", err
	}

	_, err = io.Copy(dst, src)
	src.Close()
	dst.Close()

	os.RemoveAll(extractDir)

	if err != nil {
		os.Remove(finalTmpPath.Name())
		return "", err
	}

	return finalTmpPath.Name(), nil
}

func isNewVersionAvailable(current, latest string) bool {
	// Simple version comparison (assumes semver)
	// For production, use a proper semver library
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	for i := 0; i < len(currentParts) && i < len(latestParts); i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}

	return len(latestParts) > len(currentParts)
}
