# AIDev CLI — Getting Started

Welcome to the AIDev CLI! This guide will help you install and start using the AIDev CLI to manage your AI Dev Sandbox instances.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Command Reference](#command-reference)
- [SSH Configuration](#ssh-configuration)
- [Troubleshooting](#troubleshooting)

## Installation

### macOS (Homebrew)

The easiest way to install on macOS is using Homebrew:

```bash
brew install aidev/tap/aidev
```

To upgrade to the latest version:

```bash
brew upgrade aidev
```

### Linux

Use the convenient install script:

```bash
curl -sSL https://install.aidev.sh | sh
```

This script will:
- Detect your OS and architecture
- Download the latest binary
- Install to `$HOME/.local/bin` or `/usr/local/bin` (whichever is writable)
- Verify the installation

### Windows

Install using Scoop:

```bash
scoop install aidev
```

Or download the latest binary manually from [GitHub Releases](https://github.com/aidev/aidev-cli/releases).

### Manual Installation

1. Download the latest binary for your platform from [GitHub Releases](https://github.com/aidev/aidev-cli/releases)
2. Extract the archive
3. Move the `aidev` binary to a directory in your `PATH` (e.g., `/usr/local/bin` on Unix-like systems)
4. Make it executable: `chmod +x /usr/local/bin/aidev`
5. Verify installation: `aidev --version`

## Quick Start

### 1. Login

Start the TUI and authenticate:

```bash
aidev
```

The interactive terminal UI will prompt you to log in with:
- **Email + Password**: Your AIDev account credentials
- **API Key**: A long-lived API key from your account settings

Your authentication token is securely stored in `~/.config/aidev/config.json` with restricted file permissions (mode 0600).

### 2. View Your Instances

Once logged in, the TUI displays:
- **Left pane**: List of your instances with status, tier, and CPU/memory usage
- **Right pane**: Detailed information about the selected instance

Navigate with `[↑↓]` arrow keys.

### 3. Connect via SSH

Select an instance and press `[c]` to SSH into it:

```
AIDev will automatically:
- Locate your SSH key (~/.ssh/id_ed25519 or similar)
- Connect with certificate authentication
- Present you with a remote shell
- Exit the TUI temporarily during the session
```

Or use the shortcut:

```bash
aidev ssh my-instance-name
```

### 4. Common Operations

While viewing instances in the TUI:

| Key | Action |
|-----|--------|
| `[↑↓]` | Navigate instance list |
| `[Enter]` | View instance details |
| `[c]` | Connect via SSH |
| `[s]` | Start instance |
| `[S]` | Stop instance |
| `[r]` | Restart instance |
| `[d]` | Delete instance |
| `[R]` | Resize tier (change CPU/memory) |
| `[u]` | Update image (if available) |
| `[f]` | Set up port forwarding |
| `[Ctrl+R]` | Refresh instance list |
| `[?]` | Toggle help |
| `[q]` | Quit |

## Command Reference

### Global Commands

```bash
aidev --help              # Show help
aidev --version           # Show version and check for updates
```

### Subcommands

#### `aidev` (default)

Launch the interactive TUI for instance management:

```bash
aidev
```

#### `aidev tui`

Explicitly launch the TUI:

```bash
aidev tui
```

#### `aidev login`

Authenticate and save your credentials:

```bash
aidev login [--api-key]
```

Options:
- `--api-key`: Use API key instead of email/password

Your token is automatically refreshed as needed.

#### `aidev ssh <instance-name>`

Connect directly to an instance via SSH without opening the TUI:

```bash
aidev ssh my-builder
```

#### `aidev instances`

List all instances in JSON format:

```bash
aidev instances
```

Options:
- `-a, --all`: Include stopped instances
- `-j, --json`: Pretty-print JSON

#### `aidev forward <instance-name> <local-port> <remote-port>`

Set up SSH port forwarding in the background:

```bash
aidev forward my-instance 3000 3000
```

This runs the forward as a background process. Press `[Ctrl+C]` to stop.

#### `aidev config`

Manage configuration:

```bash
aidev config get <key>           # Get config value
aidev config set <key> <value>   # Set config value
aidev config show                # Show full config
aidev config reset               # Reset to defaults
```

#### `aidev update`

Check for and install updates:

```bash
aidev update
```

This will:
- Check GitHub for the latest release
- Compare versions
- Download and install if an update is available
- Keep a backup of your previous binary (automatic rollback if needed)

## SSH Configuration

The AIDev CLI uses certificate-based SSH authentication. No passwords are stored or transmitted.

### Key Detection

The CLI automatically searches for your SSH key in this order:

1. `~/.ssh/id_ed25519` (recommended, Ed25519 key)
2. `~/.ssh/id_rsa` (RSA key)
3. `~/.ssh/id_ecdsa` (ECDSA key)
4. `~/.ssh/id_dsa` (DSA key, legacy)

### First-Time Setup

1. Generate an SSH key (if you don't have one):

   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
   ```

2. Add your public key to your AIDev account during signup or in account settings

3. The CLI will use the corresponding private key for authentication

### Host Key Verification

On first connection to an instance, you'll see:

```
The authenticity of host 'instance.example.com (1.2.3.4)' can't be established.
ED25519 key fingerprint is SHA256:...
Are you sure you want to continue connecting (yes/no)?
```

Type `yes` to accept. The host key is cached locally for future connections.

### Advanced: Custom SSH Config

The CLI respects your `~/.ssh/config` file. You can add custom settings:

```
Host *.aidev.internal
    User ubuntu
    StrictHostKeyChecking accept-new
    UserKnownHostsFile ~/.ssh/known_hosts.d/aidev
```

## Troubleshooting

### Login Issues

**"Failed to authenticate"**
- Verify your email/password or API key is correct
- Check your internet connection
- Ensure the API server is reachable: `curl https://api.sandbox.example.com/health`

**Token expired**
- The CLI automatically refreshes tokens, but manual login can help:
  ```bash
  aidev login
  ```

### SSH Connection Issues

**"Failed to determine SSH key"**
- Ensure you have an SSH key in `~/.ssh/`
- Generate one if needed: `ssh-keygen -t ed25519`
- Verify permissions: `chmod 600 ~/.ssh/id_ed25519`

**"Connection refused"**
- Ensure the instance is running: `[s]` to start it in the TUI
- Check your security group rules allow SSH (port 22)
- Verify the instance IP is reachable from your network

**"Host key verification failed"**
- Clear your known hosts: `ssh-keygen -R <instance-ip>`
- Try connecting again and accept the host key

### Instance List Won't Load

**"Failed to fetch instances"**
- Check your authentication: `aidev login`
- Verify API connectivity: `curl -H "Authorization: Bearer <token>" https://api.sandbox.example.com/api/v1/instances`
- Check internet connection

### Config File Issues

**"Permission denied: ~/.config/aidev/config.json"**
- The config file is created with restricted permissions (0600)
- Check permissions: `ls -la ~/.config/aidev/config.json`
- Reset config: `aidev config reset`

## Security Notes

- **Token Storage**: Your API token is stored in `~/.config/aidev/config.json` with `mode 0600` (readable only by you)
- **SSH Keys**: Private SSH keys are never copied or transmitted. Authentication happens locally on your machine
- **HTTPS Only**: All API communication uses HTTPS. Never trust HTTP endpoints
- **Known Hosts**: Host keys are cached in `~/.ssh/known_hosts` for verification

## Getting Help

- **Command Help**: `aidev --help` or `aidev <command> --help`
- **Issues**: Report bugs at https://github.com/aidev/aidev-cli/issues
- **Documentation**: https://github.com/aidev/aidev-cli/docs
- **API Reference**: See `docs/rails-api-spec.md` for backend API details

## Next Steps

1. **Create Your First Instance**: Use the TUI to provision a new instance
2. **Connect via SSH**: Practice connecting to test the setup
3. **Set Up Port Forwarding**: Forward local ports to run development servers
4. **Enable Auto-Update**: Check `[Ctrl+U]` in the TUI to automatically update to new releases

## API Keys

To generate an API key for programmatic access:

1. Log into your AIDev account at https://aidev.example.com
2. Go to Settings → API Keys
3. Create a new key
4. Use with: `aidev login --api-key`

API keys are long-lived and suitable for CI/CD pipelines or scripts.

---

**Version**: Check your installed version with `aidev --version`

**Last Updated**: 2026-03-27
