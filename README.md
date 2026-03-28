# AIDev CLI — Manage AI Dev Sandbox Instances from Your Terminal

A fast, cross-platform terminal user interface and CLI for managing AI Dev Sandbox instances. Written in Go with a rich interactive TUI built on [Bubble Tea](https://github.com/charmbracelet/bubbletea).

**Available for:** macOS, Linux, Windows | **Architectures:** amd64, arm64

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Features](#features)
- [Commands](#commands)
- [Configuration](#configuration)
- [SSH Setup](#ssh-setup)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Documentation](#documentation)

## Installation

### macOS (Homebrew)

```bash
brew install aidev/tap/aidev
```

To upgrade:

```bash
brew upgrade aidev
```

### Linux

```bash
curl -sSL https://install.aidev.sh | sh
```

This script will detect your OS and architecture, download the latest binary, and install it to `~/.local/bin` or `/usr/local/bin`.

### Windows

```bash
scoop install aidev
```

Or download manually from [GitHub Releases](https://github.com/aidev/aidev-cli/releases).

### Manual Installation

1. Download the binary for your platform from [GitHub Releases](https://github.com/aidev/aidev-cli/releases)
2. Extract the archive
3. Move the `aidev` binary to a directory in your `PATH`
4. Make it executable: `chmod +x /usr/local/bin/aidev` (Unix)
5. Verify: `aidev --version`

## Quick Start

### 1. Login

Start the interactive TUI:

```bash
aidev
```

You'll be prompted to log in with:
- **Email + Password**: Your AIDev account credentials
- **API Key**: A long-lived API key from your account settings

Your token is securely stored in `~/.config/aidev/config.json` with restricted permissions (mode 0600).

### 2. View Your Instances

Once logged in, the TUI displays:
- **Left pane (40%)**: List of your instances with status, tier, and CPU/memory usage
- **Right pane (60%)**: Detailed information about the selected instance

Navigate with `[↑↓]` arrow keys, press `[Enter]` to select.

### 3. Connect via SSH

Select an instance and press `[c]` to SSH into it, or use the CLI shortcut:

```bash
aidev ssh my-instance-name
```

The CLI automatically locates your SSH key and connects with certificate authentication.

### 4. Common Operations

While viewing instances in the TUI:

| Key | Action |
|-----|--------|
| `[↑↓]` | Navigate instance list |
| `[c]` | SSH into instance |
| `[s]` | Start instance |
| `[S]` | Stop instance |
| `[r]` | Restart instance |
| `[d]` | Delete instance |
| `[R]` | Resize tier (CPU/memory) |
| `[u]` | Update image |
| `[f]` | Port forwarding |
| `[Ctrl+R]` | Refresh |
| `[?]` | Toggle help |
| `[q]` | Quit |

## Features

| Feature | Status | Notes |
|---------|--------|-------|
| **Authentication** | ✅ Complete | JWT tokens, auto-refresh |
| **Instance Management** | ✅ Complete | List, create, start, stop, restart, delete |
| **Interactive TUI** | ✅ Complete | Real-time polling with responsive UI |
| **SSH Connection** | ✅ Complete | Certificate-based, no passwords |
| **Port Forwarding** | ✅ Complete | Background SSH tunnels |
| **Real-Time Updates** | ✅ Complete | SSE for live instance state |
| **Self-Update** | ✅ Complete | Built-in update checker |
| **Cross-Platform** | ✅ Complete | macOS, Linux, Windows |

## Commands

### Global

```bash
aidev --help              # Show help
aidev --version           # Show version
aidev --api <url>         # Override API base URL
```

### TUI (Interactive)

```bash
aidev              # Launch the interactive TUI (default)
aidev tui          # Explicitly launch the TUI
```

### Authentication

```bash
aidev login                # Authenticate with email/password
aidev login --api-key      # Authenticate with API key
aidev config show          # Show stored configuration
aidev config reset         # Clear configuration and logout
```

### Instance Management

```bash
aidev instances                    # List all instances (JSON)
aidev instances --all              # Include stopped instances
aidev instances --json             # Pretty-print JSON
```

### SSH & Port Forwarding

```bash
aidev ssh my-instance              # SSH directly to instance
aidev forward my-instance 3000 3000 # Forward local:3000 → remote:3000
```

### Updates

```bash
aidev update                       # Check for and install updates
```

## Configuration

### Config File Location

The AIDev CLI stores configuration in an XDG-compliant location:

| OS | Path |
|----|----|
| macOS | `~/.config/aidev/config.json` |
| Linux | `~/.config/aidev/config.json` (or `$XDG_CONFIG_HOME/aidev/config.json`) |
| Windows | `%APPDATA%\aidev\config.json` |

File permissions are automatically set to `0600` (owner read/write only).

### Config File Format

```json
{
  "base_url": "https://api.sandbox.example.com",
  "token": "eyJhbGci...",
  "token_expires_at": "2026-04-27T12:34:56Z",
  "user_email": "alice@example.com"
}
```

### View Configuration

```bash
aidev config show          # Display stored config (tokens masked)
```

### Reset Configuration

```bash
aidev config reset         # Clear all stored data
```

### Environment Variables

Override config values using environment variables:

| Variable | Effect | Example |
|----------|--------|---------|
| `AIDEV_API_URL` | API base URL | `https://api.example.com` |
| `AIDEV_TOKEN` | JWT token (overrides stored) | `eyJhbGc...` |
| `AIDEV_SSH_KEY` | SSH key path | `~/.ssh/id_rsa` |
| `XDG_CONFIG_HOME` | Config directory (Linux) | `~/.config` |

Example:

```bash
AIDEV_API_URL=https://staging-api.example.com aidev tui
```

### Priority Order

Settings are applied in this order (later values override earlier):

1. **Defaults** (built-in)
2. **Config file** (`~/.config/aidev/config.json`)
3. **Environment variables** (`AIDEV_*`)
4. **CLI flags** (`--api`, `--config`, etc.)

## SSH Setup

### Key Detection

The CLI automatically searches for your SSH key in this order:

1. `~/.ssh/id_ed25519` (recommended, Ed25519)
2. `~/.ssh/id_rsa` (RSA)
3. `~/.ssh/id_ecdsa` (ECDSA)
4. `~/.ssh/id_dsa` (DSA, legacy)

### First-Time Setup

Generate an SSH key if you don't have one:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
```

Then add your public key to your AIDev account (sign up or account settings).

### Host Key Verification

On first connection to an instance, you'll see:

```
The authenticity of host 'instance.example.com (1.2.3.4)' can't be established.
ED25519 key fingerprint is SHA256:...
Are you sure you want to continue connecting (yes/no)?
```

Type `yes` to accept. The host key is cached for future connections.

### Custom SSH Config

The CLI respects your `~/.ssh/config` file:

```
Host *.aidev.internal
    User ubuntu
    StrictHostKeyChecking accept-new
    UserKnownHostsFile ~/.ssh/known_hosts.d/aidev
```

## Security

- **Token Storage**: Your API token is stored in `~/.config/aidev/config.json` with `mode 0600` (readable only by you)
- **SSH Keys**: Private SSH keys are never copied or transmitted. Authentication happens locally
- **HTTPS Only**: All API communication uses HTTPS. Never trust HTTP endpoints
- **Known Hosts**: Host keys are cached in `~/.ssh/known_hosts` for verification
- **Auto-Refresh**: Access tokens are automatically refreshed before expiration

## Troubleshooting

### Login Issues

**"Failed to authenticate"**
- Verify your credentials are correct
- Check your internet connection
- Test API connectivity: `curl https://api.sandbox.example.com/health`

**Token expired**
- The CLI automatically refreshes tokens, but you can manually login:
  ```bash
  aidev login
  ```

### SSH Connection Issues

**"Failed to determine SSH key"**
- Ensure you have an SSH key in `~/.ssh/`
- Generate one: `ssh-keygen -t ed25519`
- Check permissions: `chmod 600 ~/.ssh/id_ed25519`

**"Connection refused"**
- Ensure the instance is running (start it with `[s]` in TUI)
- Check security group rules allow SSH (port 22)
- Verify the instance IP is reachable from your network

**"Host key verification failed"**
- Clear your known hosts: `ssh-keygen -R <instance-ip>`
- Try connecting again and accept the new host key

### Instance List Won't Load

**"Failed to fetch instances"**
- Check authentication: `aidev login`
- Verify API connectivity: test with `curl` using your token
- Check internet connection

### Config File Issues

**"Permission denied: ~/.config/aidev/config.json"**
- Check permissions: `ls -la ~/.config/aidev/config.json`
- Should be `-rw-------` (0600)
- Fix if needed: `chmod 600 ~/.config/aidev/config.json`
- Or reset: `aidev config reset`

## Development

### Build Locally

```bash
go build -o aidev ./cmd/aidev
./aidev --version
```

### Run Tests

```bash
make test      # All tests on current platform
```

### Lint

```bash
make lint      # Requires golangci-lint
```

### Cross-Platform Build

```bash
make cross-build   # Builds for Linux, macOS, Windows (amd64, arm64)
```

### Release

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for the full release process.

## Documentation

- **[Architecture](docs/ARCHITECTURE.md)** — System design, components, and implementation decisions
- **[Contributing](docs/CONTRIBUTING.md)** — Development workflow, testing, and release process
- **[TUI Design](docs/tui-design.md)** — UI/UX specification, screens, navigation, keybindings
- **[API Reference](docs/rails-api-spec.md)** — REST API endpoints and request/response formats
- **[Authentication](docs/auth-spec.md)** — JWT lifecycle and token management

## Tech Stack

- **Language**: Go 1.23+
- **TUI**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-inspired architecture)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss) (terminal colors/layout)
- **Components**: [Bubbles](https://github.com/charmbracelet/bubbles) (table, textinput, spinner)
- **CLI**: [Cobra](https://github.com/spf13/cobra) (command framework)
- **Config**: [XDG](https://github.com/adrg/xdg) (portable config paths)

## Getting Help

- **Command Help**: `aidev --help` or `aidev <command> --help`
- **Issues**: Report bugs at https://github.com/aidev/aidev-cli/issues
- **Documentation**: See the [docs/](docs/) folder
- **API Details**: See [docs/rails-api-spec.md](docs/rails-api-spec.md)

## License

TBD
