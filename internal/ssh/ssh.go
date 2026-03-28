package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// KeyPath returns the path to the user's SSH private key
// Tries common key names in order: id_ed25519, id_rsa, id_ecdsa, id_dsa
func KeyPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")

	// Try common key names in order
	keyNames := []string{
		"id_ed25519",
		"id_rsa",
		"id_ecdsa",
		"id_dsa",
	}

	for _, name := range keyNames {
		keyPath := filepath.Join(sshDir, name)
		if _, err := os.Stat(keyPath); err == nil {
			return keyPath, nil
		}
	}

	return "", fmt.Errorf("no SSH key found in ~/.ssh (tried: %v)", keyNames)
}

// ConnectOptions for SSH connection
type ConnectOptions struct {
	Host       string
	Port       int
	User       string
	KeyPath    string // If empty, uses default
	StrictHost bool   // Enable StrictHostKeyChecking (default: true)
}

// Connect launches an interactive SSH session
// This spawns an SSH subprocess with inherited stdin/stdout/stderr
// The parent process waits for SSH to exit
func Connect(opts ConnectOptions) error {
	// Find key if not specified
	if opts.KeyPath == "" {
		keyPath, err := KeyPath()
		if err != nil {
			return err
		}
		opts.KeyPath = keyPath
	}

	// Verify key exists
	if _, err := os.Stat(opts.KeyPath); err != nil {
		return fmt.Errorf("SSH key not found: %s", opts.KeyPath)
	}

	// Build SSH command
	args := []string{
		"-i", opts.KeyPath,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "UserKnownHostsFile=~/.ssh/known_hosts",
	}

	if opts.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", opts.Port))
	}

	// Host user@host
	hostArg := fmt.Sprintf("%s@%s", opts.User, opts.Host)
	args = append(args, hostArg)

	// Prepare command
	cmd := exec.Command("ssh", args...)

	// Inherit stdio for interactive session
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run SSH command
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// SSH exited with a code, propagate it
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return &ExitError{Code: status.ExitStatus()}
			}
		}
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	return nil
}

// PortForwardOptions for SSH port forwarding
type PortForwardOptions struct {
	Host       string
	Port       int
	User       string
	KeyPath    string
	LocalPort  int // Local port to bind to
	RemotePort int // Remote port on the VM
}

// ForwardPort sets up local port forwarding via SSH
// Returns a function to stop the forwarding
func ForwardPort(opts PortForwardOptions) (func() error, error) {
	// Find key if not specified
	if opts.KeyPath == "" {
		keyPath, err := KeyPath()
		if err != nil {
			return nil, err
		}
		opts.KeyPath = keyPath
	}

	// Verify key exists
	if _, err := os.Stat(opts.KeyPath); err != nil {
		return nil, fmt.Errorf("SSH key not found: %s", opts.KeyPath)
	}

	// Build SSH command for port forwarding
	// ssh -i key -N -L localhost:3000:localhost:3000 user@host
	args := []string{
		"-i", opts.KeyPath,
		"-N", // Don't execute command
		"-o", "StrictHostKeyChecking=accept-new",
		"-L", fmt.Sprintf("localhost:%d:localhost:%d", opts.LocalPort, opts.RemotePort),
	}

	if opts.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", opts.Port))
	}

	hostArg := fmt.Sprintf("%s@%s", opts.User, opts.Host)
	args = append(args, hostArg)

	// Prepare command (run in background)
	cmd := exec.Command("ssh", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Start the forwarding process
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start port forwarding: %w", err)
	}

	// Return a cleanup function
	stopFn := func() error {
		if cmd.Process != nil {
			return cmd.Process.Kill()
		}
		return nil
	}

	return stopFn, nil
}

// GetSSHCommand builds the SSH command string (for display/documentation)
func GetSSHCommand(opts ConnectOptions) string {
	cmd := "ssh"

	if opts.KeyPath != "" {
		cmd += fmt.Sprintf(" -i %s", opts.KeyPath)
	}

	if opts.Port != 22 {
		cmd += fmt.Sprintf(" -p %d", opts.Port)
	}

	cmd += fmt.Sprintf(" %s@%s", opts.User, opts.Host)
	return cmd
}

// ExitError represents an SSH exit code
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("SSH exited with code %d", e.Code)
}
