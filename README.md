# AIDev CLI — Complete Implementation

The terminal user interface (TUI) and command-line tool for managing AI Dev Sandbox instances.

**Status:** ✅ All 6 phases complete — ready for distribution and deployment.

## Overview

The `aidev` CLI provides:
- **TUI**: Interactive terminal interface to manage cloud VM instances
- **Commands**: `login`, `ssh`, `forward`, `instances`, `config`
- **Authentication**: JWT-based auth with config file storage
- **Cross-platform**: macOS, Linux, Windows (single Go binary, ~10 MB)

## Building

### Development Build

```bash
# Build locally
go build -o aidev ./cmd/aidev

# Run the TUI (default)
./aidev
```

### Release Build

```bash
# Requires goreleaser (https://goreleaser.com)
goreleaser release --snapshot --rm-dist

# Full production release (requires GitHub token)
GITHUB_TOKEN=$GH_TOKEN goreleaser release --rm-dist
```

## Commands

```bash
# Launch TUI (default)
aidev

# Login with email/password
aidev login

# Login with API key
aidev login --api-key

# Show configuration
aidev config show

# Manage instances
aidev instances
aidev instances --json  # JSON output for scripting

# SSH directly to instance
aidev ssh my-instance

# Set up port forwarding
aidev forward my-instance 3000 3000

# Check for updates
aidev update

# Show help
aidev --help
aidev --version
```

## Configuration

Config is stored at (XDG-compliant):
- **Linux/macOS**: `~/.config/aidev/config.json`
- **Windows**: `%APPDATA%\aidev\config.json`

File permissions are restricted to `0600` (owner read/write only).

```json
{
  "base_url": "https://api.sandbox.example.com",
  "token": "eyJhbGci...",
  "token_expires_at": "2025-04-27T12:34:56Z",
  "user_email": "alice@example.com"
}
```

## API

The TUI communicates with a Rails 8.1 backend via REST API.

**Spec:** See `docs/rails-api-spec.md`

Key endpoints:
- `POST /api/v1/auth/login` — login with email/password or API key
- `GET /api/v1/instances` — list instances
- `POST /api/v1/instances/:id/start|stop|restart` — control instances
- `GET /api/v1/instances/events` — SSE for real-time updates

## Architecture

```
cmd/aidev/main.go                    # Cobra CLI root + version info
│
├── internal/api/
│   ├── client.go                    # HTTP client with auth + 401 retry
│   └── sse.go                       # SSE client for real-time updates
│
├── internal/auth/
│   └── store.go                     # XDG config file read/write (mode 0600)
│
├── internal/models/
│   ├── auth.go                      # Request/response types
│   ├── instance.go                  # Instance model
│   └── event.go                     # SSE event types
│
├── internal/ssh/
│   ├── ssh.go                       # SSH connection + PTY management
│   └── forward.go                   # SSH port forwarding (background)
│
├── internal/tui/
│   ├── app.go                       # Root Bubble Tea model + state machine
│   └── views/
│       ├── login.go                 # LoginModel (email/password)
│       ├── main.go                  # MainModel (list + detail split)
│       ├── instance_list.go         # InstanceListModel (table with polling)
│       ├── instance_detail.go       # InstanceDetailModel (scrollable viewport)
│       ├── confirm_dialog.go        # Confirmation modal
│       ├── resize_modal.go          # Tier selection modal
│       ├── forward_modal.go         # Port forwarding modal
│       ├── notification.go          # Toast notifications manager
│       └── styles.go                # Color palette & lipgloss styles
│
└── internal/commands/
    ├── main.go                      # Cobra command registration
    ├── tui.go                       # Launch TUI
    ├── login.go                     # Login command
    ├── ssh.go                       # SSH command with instance lookup
    ├── forward.go                   # Port forward command
    ├── instances.go                 # Instance management (list, JSON export)
    ├── config.go                    # Config management (show, set, get, reset)
    └── update.go                    # Self-update command
```

## Tech Stack

- **Language**: Go 1.23+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm architecture)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss) (terminal colors/layout)
- **Components**: [Bubbles](https://github.com/charmbracelet/bubbles) (textinput, table, spinner)
- **CLI**: [Cobra](https://github.com/spf13/cobra) (command structure)
- **Config**: [XDG](https://github.com/adrg/xdg) (portable config paths)

## Development

```bash
# Run tests
make test

# Lint
make lint

# Format code
make fmt

# Watch mode (requires entr)
make dev-watch
```

## Implementation Status

| Phase | Feature | Status |
|-------|---------|--------|
| 1 | Authentication & Login | ✅ Complete |
| 2 | Instance List & Polling | ✅ Complete |
| 3 | Detail Pane & CRUD Operations | ✅ Complete |
| 4 | SSH Connection & Port Forwarding | ✅ Complete |
| 5 | Real-time Updates & Notifications | ✅ Complete |
| 6 | Distribution & Self-Update | ✅ Complete |

### What's Included

- **Phase 1**: Cobra CLI, JWT auth, config storage (XDG-compliant), LoginView
- **Phase 2**: Instance list with polling, table widget, status badges (running/stopped/error)
- **Phase 3**: Scrollable detail pane, confirm dialogs, resize modal, CRUD operations
- **Phase 4**: Certificate-based SSH connection, port forwarding, `aidev ssh <name>` shortcut
- **Phase 5**: SSE real-time updates, toast notifications, port forwarding modal
- **Phase 6**: Goreleaser config, install.sh, self-update command, documentation

## Documentation

### User Guides
- **[Getting Started](docs/GETTING_STARTED.md)** — Installation methods, quick start, command reference
- **[Configuration Guide](docs/CONFIGURATION.md)** — Config file format, environment variables, advanced setup

### Design & Technical Docs
- **[TUI Design](docs/tui-design.md)** — UI/UX spec: screens, navigation, keybindings
- **[API Reference](docs/rails-api-spec.md)** — REST API spec: all endpoints, request/response formats, SSE events
- **[Auth Specification](docs/auth-spec.md)** — Auth spec: JWT lifecycle, config storage, Rails implementation guide

### Installation
- **Binary**: Download from [GitHub Releases](https://github.com/aidev/aidev-cli/releases)
- **Homebrew**: `brew install aidev/tap/aidev` (macOS)
- **Linux**: `curl -sSL https://install.aidev.sh | sh`
- **Windows**: `scoop install aidev`
- **Manual**: Extract binary and add to PATH

## License

TBD
