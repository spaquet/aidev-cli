# AIDev CLI — Architecture

This document describes the system architecture, components, and design decisions behind the AIDev CLI.

## Table of Contents

- [System Overview](#system-overview)
- [Component Architecture](#component-architecture)
- [TUI Architecture](#tui-architecture)
- [Authentication Flow](#authentication-flow)
- [SSH & Port Forwarding](#ssh--port-forwarding)
- [Real-Time Updates](#real-time-updates)
- [Distribution & Self-Update](#distribution--self-update)

## System Overview

The AIDev CLI is a cross-platform terminal application that manages cloud VM instances. It combines:

- **Interactive TUI** — Rich terminal interface for instance management (list, detail, CRUD operations)
- **CLI Commands** — Command-line interface for scripting and quick operations
- **API Client** — REST client with JWT authentication and real-time event streaming
- **Configuration** — XDG-compliant config storage with token persistence

```
┌─────────────────────────────────────────────────────────┐
│  User (Terminal)                                        │
├─────────────────────────────────────────────────────────┤
│  AIDev CLI (Go binary, ~10 MB)                          │
│                                                         │
│  ┌──────────────────┐  ┌──────────────────┐             │
│  │  TUI (Bubble Tea)│  │  CLI Commands    │             │
│  │  - Instance List │  │  - login         │             │
│  │  - Details       │  │  - ssh           │             │
│  │  - Modals        │  │  - forward       │             │
│  │  - Notifications │  │  - instances     │             │
│  └──────────────────┘  └──────────────────┘             │
│           ↓                      ↓                       │
│  ┌──────────────────────────────────────┐               │
│  │  API Client + Auth                   │               │
│  │  - HTTP requests + JWT               │               │
│  │  - Token refresh                     │               │
│  │  - SSE for real-time updates         │               │
│  └──────────────────────────────────────┘               │
│           ↓                      ↓                       │
│  ┌──────────────────────────────────────┐               │
│  │  Config Storage (XDG)                │               │
│  │  ~/.config/aidev/config.json (0600)  │               │
│  └──────────────────────────────────────┘               │
│                                                         │
│  ┌──────────────────────────────────────┐               │
│  │  SSH Subprocess                      │               │
│  │  - Certificate auth                  │               │
│  │  - Port forwarding (background)      │               │
│  └──────────────────────────────────────┘               │
└─────────────────────────────────────────────────────────┘
                        ↓
         ┌──────────────────────────┐
         │  AIDev API (Rails 8.1)   │
         │  - /api/v1/auth/*        │
         │  - /api/v1/instances/*   │
         │  - /api/v1/events (SSE)  │
         └──────────────────────────┘
```

## Component Architecture

### Directory Structure

```
cmd/aidev/
└── main.go                    Entry point, Cobra CLI root

internal/
├── api/
│   ├── client.go             HTTP client wrapper, auth, CRUD operations
│   └── sse.go                Server-Sent Events client for real-time updates
│
├── auth/
│   └── store.go              XDG config file read/write, token persistence
│
├── commands/
│   ├── tui.go                Launch TUI command
│   ├── login.go              Login command (email/password or API key)
│   ├── ssh.go                SSH shortcut command
│   ├── forward.go            Port forwarding command
│   ├── instances.go          List instances command
│   ├── config.go             Config management (show, reset)
│   └── update.go             Self-update checker and installer
│
├── models/
│   ├── auth.go               LoginRequest, LoginResponse, User, Config types
│   └── instance.go           Instance, CreateRequest, SSEEvent types
│
├── ssh/
│   └── ssh.go                SSH connection, port forwarding via subprocess
│
└── tui/
    ├── app.go                Root Bubble Tea model, screen state machine
    ├── styles.go             Shared color palette
    │
    └── views/
        ├── login.go          LoginModel — email/password form
        ├── main.go           MainModel — split list + detail layout
        ├── instance_list.go  InstanceListModel — table with polling
        ├── instance_detail.go InstanceDetailModel — scrollable details
        ├── confirm_dialog.go  ConfirmDialogModel — yes/no modal
        ├── resize_modal.go   ResizeModalModel — tier selection
        ├── forward_modal.go  ForwardModalModel — port forwarding input
        ├── notification.go   NotificationManager — toast notifications
        └── styles.go         View-specific lipgloss styles
```

### Component Descriptions

#### `internal/api/`

**HTTP Client** (`client.go`)
- Wraps standard Go http.Client
- Injects Bearer token in Authorization header
- Handles 401 Unauthorized → token refresh → retry
- Methods for all instance operations (list, create, start, stop, etc.)

**SSE Client** (`sse.go`)
- Establishes persistent connection to `/api/v1/instances/events`
- Auto-reconnects on disconnect (5s backoff)
- Parses `event: type\ndata: {json}` format
- Channels events to TUI for real-time updates

#### `internal/auth/`

**Config Store** (`store.go`)
- Reads/writes XDG-compliant config file
- Locations: `~/.config/aidev/config.json` (Unix), `%APPDATA%\aidev\config.json` (Windows)
- File permissions: 0600 (owner read/write only)
- Stores: API URL, JWT token, token expiration, user email
- Detects and refreshes expired tokens

#### `internal/commands/`

Cobra subcommands that implement the CLI interface:
- `tui` — Launch the interactive TUI
- `login` — Authenticate and save token
- `ssh <instance>` — SSH to instance (looks up details from API)
- `forward <instance> <local> <remote>` — Port forward via SSH subprocess
- `instances` — List instances (JSON output)
- `config` — Show/reset configuration
- `update` — Check for and install updates from GitHub

#### `internal/ssh/`

**SSH Handler** (`ssh.go`)
- Spawns `ssh` binary (expects it in PATH)
- Uses first available SSH key: `~/.ssh/id_ed25519`, `id_rsa`, `id_ecdsa`, `id_dsa`
- Certificate-based auth (no password storage)
- Handles PTY allocation for interactive sessions
- Port forwarding via SSH `-L` local forwarding
- Captures exit codes from SSH process

**Why subprocess, not native SSH library?**
- Avoids binary size increase (libssh/crypto libraries)
- Respects user's `~/.ssh/config` and `~/.ssh/known_hosts`
- Easier key rotation without code changes
- Proven standard tool (`openssh`)

#### `internal/models/`

Type definitions for API communication:
- **Auth types**: LoginRequest, LoginResponse, RefreshRequest, RefreshResponse
- **Instance types**: Instance, CreateInstanceRequest, UpdateInstanceRequest
- **SSE types**: SSEEvent (for real-time instance updates)
- **Config type**: Config (stored locally in JSON)

#### `internal/tui/`

**Bubble Tea Framework**
- Elm-inspired architecture: Model → Update → View
- Models are immutable; Updates return new Model + Commands
- No global state (except AppModel contains all screen state)

**Root Model** (`app.go`, AppModel)
- Screen state machine: ScreenLogin ↔ ScreenMain
- Routes messages to active screen
- Handles token refresh on startup

**Main Screen** (`views/main.go`, MainModel)
- Two-pane split layout: 40% list + 60% detail
- Modal overlay system (confirm, resize, forward)
- Polling loop: fetches instances every 5 seconds
- Keyboard handlers for all operations (start, stop, SSH, etc.)
- SSE subscription for real-time updates

**Instance List** (`views/instance_list.go`, InstanceListModel)
- Bubble Tea table component
- Columns: Name, Status (with color badges), Tier, CPU, Memory, Disk
- Selection highlighting
- Sorting by column

**Instance Detail** (`views/instance_detail.go`, InstanceDetailModel)
- Scrollable viewport for large content
- Displays: metadata, SSH command, storage usage, tools, URLs
- Responsive to window size changes

**Modals** (dialog, resize, forward)
- Centered overlays on main view
- Modal state machine (waiting for user input)
- Send response messages back to MainModel

**Notifications** (`views/notification.go`, NotificationManager)
- Multiple notification types: Info, Success, Warning, Error
- Auto-expiring (time-based cleanup)
- Toast-style display (top-right corner)
- Max 5 visible at once

## TUI Architecture

### Screen State Machine

```
┌──────────────┐
│ ScreenLogin  │  LoginModel (email/password form)
│              │  ↓ On successful auth
└──────┬───────┘  Switches to ScreenMain
       │
       └──→ ┌──────────────┐
            │ ScreenMain   │  MainModel (list + detail)
            │              │  ↓ [q] key or logout
            └──────┬───────┘  Switches back to ScreenLogin
                   │
                   └──→ [SSH, forward, resize] operations
                        Exit TUI (return control to subprocess)
```

### Bubble Tea Message Flow

```
User Input  ──→  MainModel.Update()  ──→  Returns (Model, Cmd)
                                               ↓
Cmd         ──→  Execute (fetch API,    ──→  Msg (new data)
                  poll, etc)
                                               ↓
Msg         ──→  MainModel.Update()  ──→  Returns (Model, Cmd)
                                               ↓
Model       ──→  MainModel.View()     ──→  Render to terminal
```

### Example: Fetching Instances

```go
// User presses [Ctrl+R] to refresh
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case key.Runes("r"):  // When Ctrl+R pressed
        return m, fetchInstancesCmd()  // Return a Command
}

// fetchInstancesCmd spawns a goroutine to fetch from API
func fetchInstancesCmd() tea.Cmd {
    return func() tea.Msg {
        instances, err := apiClient.GetInstances()
        return fetchedInstancesMsg{instances, err}  // Send Msg back
    }
}

// Handle the response
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case fetchedInstancesMsg:
        m.instances = msg.instances  // Update model
        // Return view will re-render with new data
        return m, nil
}
```

## Authentication Flow

### Login Process

```
User starts: aidev
    ↓
AppModel checks stored token
    ├─ Token expired → prompt login
    └─ Token valid → show instances
    ↓
User enters email/password (or API key)
    ↓
API request: POST /api/v1/auth/login
    ↓
Backend returns: {token, expires_at, user}
    ↓
Token stored: ~/.config/aidev/config.json (0600)
    ↓
AppModel switches: ScreenLogin → ScreenMain
```

### Token Refresh

Automatic refresh happens when:
1. Token expiration detected (checked before API calls)
2. 401 Unauthorized response from API
3. Manual login command: `aidev login`

```go
// In API client
func (c *Client) callAPI(method, path string) {
    if c.isTokenExpired() {
        c.refreshToken()  // Auto-refresh before request
    }
    resp := http.Request(method, path)

    if resp.StatusCode == 401 {
        c.refreshToken()  // Retry on 401
        resp = http.Request(method, path)
    }
}
```

## SSH & Port Forwarding

### SSH Connection Strategy

Why the TUI quits before SSH connects:

1. **Terminal Sharing**: A TUI can't coexist with an interactive SSH session (both need terminal control)
2. **Clean Design**: User gets shell prompt immediately after TUI quits
3. **SSH Config Respect**: Uses system OpenSSH (respects ~/.ssh/config, known_hosts)

**Flow:**
```
User presses [c] in TUI
    ↓
TUI quits (saves state)
    ↓
CLI executes: ssh ubuntu@instance.ip -i ~/.ssh/id_ed25519
    ↓
User gets interactive shell
    ↓
User exits SSH session
    ↓
Control returns to CLI (optional: return to TUI)
```

### Port Forwarding

Runs SSH in background with local port forwarding:

```bash
ssh -L 3000:localhost:3000 ubuntu@instance.ip -i ~/.ssh/id_ed25519 -N &
```

**Key points:**
- `-L local_port:localhost:remote_port` — Local forwarding
- `-N` — Don't execute command (just forward)
- `&` — Background process
- User presses Ctrl+C to stop forwarding

The `forward_modal` captures user input (local/remote ports) and passes to `ssh.go` which spawns the subprocess.

## Real-Time Updates

### SSE Stream

The API provides `/api/v1/instances/events` (Server-Sent Events stream):

**Connection:**
```go
conn, _ := http.Get("/api/v1/instances/events", bearerToken)
defer conn.Body.Close()

scanner := bufio.NewScanner(conn.Body)
for scanner.Scan() {
    line := scanner.Text()
    // Parse: "event: instance.updated\ndata: {json}"
}
```

**Event Format:**
```
event: instance.started
data: {"id": "inst_123", "status": "running"}

event: instance.stopped
data: {"id": "inst_456", "status": "stopped"}
```

**TUI Integration:**
- SSE client runs in background (separate goroutine)
- Events are sent to TUI via channel
- MainModel updates instance state in real-time
- Notifications display state changes to user

## Distribution & Self-Update

### GoReleaser Configuration

The project uses GoReleaser for automated multi-platform releases. Configuration: `.goreleaser.yml`

**Build Matrix:**
- Platforms: Linux, macOS, Windows
- Architectures: amd64, arm64 (Windows arm64 excluded)
- Output: tar.gz for Unix, zip for Windows
- Binary size: ~10 MB (with strip flags)

**Release Artifacts:**
```
aidev_v0.2.0_linux_amd64.tar.gz
aidev_v0.2.0_linux_arm64.tar.gz
aidev_v0.2.0_darwin_amd64.tar.gz
aidev_v0.2.0_darwin_arm64.tar.gz
aidev_v0.2.0_windows_amd64.zip
checksums.txt (SHA256)
```

**Release Metadata:**
- Automatic changelog generation (feat, fix commits)
- GitHub Release with download links
- Homebrew tap formula auto-generation
- Docker image builds (optional)
- S3 backup upload (optional)

### Install Script

`install.sh` — Bash installer for Unix-like systems

**Features:**
- Auto-detects OS (macOS, Linux) and arch (amd64, arm64)
- Fetches latest version from GitHub API
- Downloads appropriate binary
- Installs to `~/.local/bin` or `/usr/local/bin`
- Verifies installation

**Usage:**
```bash
curl -sSL https://install.aidev.sh | sh
```

### Self-Update Command

`aidev update` — Built-in update mechanism

**Process:**
1. Fetch latest release from GitHub API
2. Compare version (semver comparison)
3. If newer available:
   - Download binary for current platform
   - Extract from archive
   - Atomic replace (backup old binary)
   - Verify installation
   - Report success/failure

**Atomic Replacement:**
```go
// 1. Download to temp file
tmpPath := downloadLatest()

// 2. Backup current binary
cp(currentBinary, currentBinary + ".bak")

// 3. Replace atomically
mv(tmpPath, currentBinary)

// 4. If verification fails, rollback
// cp(currentBinary + ".bak", currentBinary)
```

## Design Decisions

### Why Go?

- Single static binary (~10 MB, no runtime deps)
- Cross-platform with same source code
- Fast compilation and execution
- Strong standard library (no external dependencies for core features)

### Why Bubble Tea + Lipgloss?

- Elm-inspired architecture is proven and maintainable
- Rich terminal UI without external system dependencies
- Clear separation: Model (state) → Update (logic) → View (rendering)
- Growing ecosystem of reusable components (Bubbles)

### Why XDG Config Storage?

- Portable across macOS, Linux, Windows
- Respects user's config directory preferences
- File permissions (0600) for security
- One place to look for config (not scattered across OS-specific paths)

### Why OAuth JWT Tokens?

- Stateless auth (no server sessions needed)
- Automatic expiration
- Can be refreshed without password
- Works with certificate-based SSH

### Why SSH Subprocess, Not Native Library?

- Smaller binary size
- Respects user's SSH config and known_hosts
- Proven standard (OpenSSH)
- No key material in memory (subprocess handles it)

## Testing

### Test Structure

**Unit Tests:** All packages have basic unit tests
- `internal/models/*_test.go` — Model struct tests
- `internal/auth/*_test.go` — Config storage
- `internal/commands/*_test.go` — Platform-specific logic

**Integration Tests:** Minimal (TUI is hard to test)
- API client tested with mock HTTP server
- SSH tests skipped on platforms without ssh binary
- Config storage tested with temp directories

**CI/CD:**
- Lint gate (hard) via golangci-lint
- Tests run on 3-platform matrix (Ubuntu, macOS, Windows)
- Platform-specific tests for OS-dependent code
- All tests must pass before release

## Performance Considerations

- **Instance Polling**: 5-second interval (balance freshness vs API load)
- **SSE Updates**: Real-time state changes (overrides polling for those instances)
- **Terminal Rendering**: Bubble Tea handles double-buffering (minimal flicker)
- **Memory**: ~20-50 MB typical (depends on instance count)
- **Binary Size**: ~10 MB (acceptable for single-use tool)

## Security Considerations

- **Token Storage**: Restricted file permissions (0600)
- **No Password Storage**: Only JWT tokens stored
- **SSH Keys**: Never copied; subprocess handles auth
- **HTTPS Only**: All API communication encrypted
- **Token Expiration**: Auto-refresh before expiry
- **401 Handling**: Refresh token on unauthorized response

## Future Improvements

Potential enhancements (not currently implemented):
- Customizable color themes
- Vim keybindings option
- Instance filtering and search
- SSH key management from TUI
- Secrets management integration
- Multi-account support (separate config files)
- Webhook notifications
- Instance metrics dashboard
