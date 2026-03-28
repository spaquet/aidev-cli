# AIDev CLI — Phase 1 Skeleton

The terminal user interface (TUI) and command-line tool for managing AI Dev Sandbox instances.

**Status:** Phase 1 complete — auth and login infrastructure working.

## Overview

The `aidev` CLI provides:
- **TUI**: Interactive terminal interface to manage cloud VM instances
- **Commands**: `login`, `ssh`, `forward`, `instances`, `config`
- **Authentication**: JWT-based auth with config file storage
- **Cross-platform**: macOS, Linux, Windows (single Go binary, ~10 MB)

## Building

```bash
# Build
make build

# Run the TUI (default)
./bin/aidev

# Or use make run
make run

# Cross-compile for all platforms
make cross-build
```

## Commands

```bash
# Launch TUI (default)
aidev

# Login with email/password
aidev login --email alice@example.com --password correcthorsebatterystaple

# Login with API key
aidev login --api-key aidev_sk_abc123...

# Show config
aidev config show

# Logout
aidev config logout

# (Phase 2+)
aidev instances list
aidev ssh my-instance
aidev forward my-instance 3000
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
cmd/aidev/main.go              # Cobra CLI root
│
├── internal/api/
│   └── client.go              # HTTP client with auth + 401 retry
│
├── internal/auth/
│   └── store.go               # XDG config file read/write
│
├── internal/models/
│   ├── auth.go                # Request/response types
│   └── instance.go            # Instance model
│
├── internal/tui/
│   ├── app.go                 # Root Bubble Tea model + screen state
│   ├── styles.go              # Lipgloss styles
│   └── views/
│       ├── login.go           # LoginModel (email/password form)
│       └── styles.go          # View-specific colors
│
└── internal/commands/
    ├── tui.go                 # Launch TUI
    ├── login.go               # Login command
    ├── ssh.go                 # SSH command (Phase 4)
    ├── forward.go             # Port forward command (Phase 4)
    ├── instances.go           # Instance management (Phase 2)
    └── config.go              # Config show/logout
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

## Phase 2–6 Roadmap

- **Phase 2**: Instance list (GET /instances, polling, table with status badges)
- **Phase 3**: Instance detail + CRUD (modals for delete/resize/update)
- **Phase 4**: SSH + port forwarding (`ssh` subprocess, `aidev ssh <name>` shortcut)
- **Phase 5**: SSE real-time updates, image updates, port exposure
- **Phase 6**: Cross-platform distribution (goreleaser, Homebrew tap, curl installer)

## Design Docs

- **`docs/tui-design.md`** — UI/UX spec: screens, navigation, keybindings
- **`docs/rails-api-spec.md`** — REST API spec: all endpoints, request/response formats, SSE events
- **`docs/auth-spec.md`** — Auth spec: JWT lifecycle, config storage, Rails implementation guide

## License

TBD
