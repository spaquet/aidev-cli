# AIDev TUI — Design Specification

## Overview

The AIDev CLI TUI is a terminal user interface for managing AI Dev Sandbox instances. It is built with Go + Bubble Tea, targeting macOS, Linux, and Windows.

**Key principles:**
- Zero-friction: one keystroke to SSH into a VM
- Real-time: instance status updates via SSE
- Opinionated defaults: no configuration needed
- Discoverable: keybindings visible at all times

---

## Screen Architecture

### 1. Login Screen

**Purpose:** Authenticate and store a session token.

**Layout:**
```
┌────────────────────────────────────────┐
│                                        │
│     🔐 AI Dev Sandbox Login            │
│                                        │
│     ────────────────────────────────  │
│                                        │
│     Email:                             │
│     [alice@example.com__________]      │
│                                        │
│     Password:                          │
│     [••••••••••••••__________]          │
│                                        │
│     ────────────────────────────────  │
│                                        │
│     Or paste an API key:               │
│     [aidev_sk_abc123__________]        │
│                                        │
│     ────────────────────────────────  │
│                                        │
│     [Enter] Sign in   [Tab] Next field │
│     [Ctrl+C] Exit                      │
│                                        │
└────────────────────────────────────────┘
```

**Fields:**
- Email or API key (mutually exclusive)
- Password (hidden, only if email chosen)

**Interactions:**
- `Tab` — cycle through fields
- `Enter` — submit login
- `Ctrl+C` — quit

**API call:**
```
POST /api/v1/auth/login
Body: { email, password }  OR  { api_key }
```

**Success:** Store token in config.json, transition to MainScreen.
**Error:** Show red error banner, stay on LoginScreen.

---

### 2. Main Screen (Instance List + Detail)

**Purpose:** Browse and manage instances.

**Layout:**
```
┌─────────────────────────────────────────────────────────────────────────┐
│ AIDev  •  3 instances                                         ⟳ Refresh  │
├─────────────────────────────┬─────────────────────────────────────────┤
│ NAME                        │ SSH CONNECT                             │
│ STATUS     TIER  REGION DISK│                                         │
├─────────────────────────────┼─────────────────────────────────────────┤
│ my-builder      [running]   │ ssh -i ~/.ssh/id_ed25519                │
│ builder  us-e-1  60/80 GB   │ ubuntu@my-builder.tunnel.aidev.sandbox  │
│                             │                                         │
│ learning-box    [running]   │ Region: us-east-1                       │
│ starter  us-e-1  30/40 GB   │ Tier: starter ($9/mo)                   │
│                             │ Disk: 30 GB / 40 GB                     │
│ archive         [stopped]   │ Image: 2025-03-27                       │
│ builder  eu-w-1   5/80 GB   │ Update available: YES                   │
│                             │                                         │
│                             │ Tools installed:                        │
│                             │  • Claude Code                          │
│                             │  • Codex (OpenAI)                       │
│                             │  • nvim                                 │
│                             │                                         │
│                             │ Public URLs (Builder tier):             │
│                             │  https://8080-my-builder.tunnel.a...    │
│                             │                                         │
│                             │ Actions:                                │
│ [↑↓] Navigate  [s]tart [S]top│ [c]onnect [f]orward [u]pdate [d]elete  │
│ [r]estart [R]esize [e]xpose │ [?] Help  [q]uit                       │
└─────────────────────────────┴─────────────────────────────────────────┘
```

**Left pane: Instance List (40% width)**
- Table with columns: NAME, STATUS, TIER, REGION, DISK
- Status badges: `[running]` green, `[stopped]` gray, `[starting]` yellow, `[stopping]` yellow
- Tier badges: `starter` blue, `builder` purple, `pro` gold
- Disk bar: visual progress bar (e.g., `30/80 GB` with a colored bar)
- Selection highlight moves with arrow keys

**Right pane: Instance Detail (60% width)**
- SSH connection command (copy-pasteable)
- Instance metadata (region, tier, disk, image version)
- Installed tools list
- Public URLs (if Builder/Pro)
- Action hints based on current state

**Status bar (bottom)**
```
[↑↓] Navigate  [s]tart [S]top  [r]estart [R]esize [e]xpose
[c]onnect [f]orward [u]pdate [d]elete  [?] Help  [q]uit
```

**Real-time updates:**
- Instance list polls `/api/v1/instances` every 10 seconds
- SSE stream (`GET /api/v1/instances/events`) pushes status changes, image updates, deletes
- Spinner in top-right corner while syncing

---

### 3. Modal: Confirm Delete

**Trigger:** Press `d` on an instance.

**Layout:**
```
┌──────────────────────────────────────┐
│  ⚠️  Delete instance?                │
│                                      │
│  Are you sure you want to delete     │
│  "my-builder"? This cannot be        │
│  undone. All data will be lost.      │
│                                      │
│                                      │
│  [y] Delete   [n] Cancel             │
└──────────────────────────────────────┘
```

**Interactions:**
- `y` — confirm delete, send DELETE request, wait for 204, remove from list
- `n` or `Esc` — cancel, dismiss modal
- `Ctrl+C` — cancel (like `n`)

---

### 4. Modal: Resize Tier

**Trigger:** Press `R` on an instance.

**Layout:**
```
┌──────────────────────────────────────────────┐
│  📦 Resize Instance                          │
│  Current tier: builder ($25/mo)              │
│                                              │
│  Select new tier:                            │
│    starter  ($9/mo)   2 vCPU, 4 GB RAM       │
│  ▶ builder  ($25/mo)  4 vCPU, 8 GB RAM       │
│    pro      ($59/mo)  8 vCPU, 16 GB RAM      │
│                                              │
│  Note: Resize requires a reboot (~2 min)    │
│                                              │
│  [↑↓] Select   [Enter] Confirm   [Esc] Back │
└──────────────────────────────────────────────┘
```

**Interactions:**
- `↑↓` — navigate tier options
- `Enter` — send PATCH /api/v1/instances/:id with new tier, dismiss modal
- `Esc` — cancel

---

### 5. Modal: Port Forwarding

**Trigger:** Press `f` on an instance.

**Layout:**
```
┌────────────────────────────────────────┐
│  🔗 Local Port Forwarding              │
│                                        │
│  Forward local port to VM:             │
│  Local port:                           │
│  [3000_________]                       │
│                                        │
│  Remote port (on VM):                  │
│  [3000_________]                       │
│                                        │
│  [Enter] Start forwarding  [Esc] Back  │
└────────────────────────────────────────┘
```

**Interactions:**
- `Tab` — cycle between fields
- `Enter` — start port forwarding in background, show "Forwarding 3000→3000" in detail pane
- `Esc` — cancel

**Background process:**
- Exec `ssh -N -L 3000:localhost:3000 ubuntu@host` in subprocess
- Show persistent "Forwarding" status in detail pane
- Press `F` to kill the forwarding subprocess

---

### 6. Modal: Image Update

**Trigger:** Press `u` on an instance with `image_update_available: true`.

**Layout:**
```
┌──────────────────────────────────────────┐
│  📦 Update Image                         │
│                                          │
│  New version: 2025-03-27                │
│  Current:     2025-03-24                │
│                                          │
│  Changes:                                │
│   • Claude Code v0.2.5 → v0.3.0         │
│   • Ubuntu security patches             │
│   • Codex CLI v1.2.1                    │
│                                          │
│  Estimated time: 5 minutes               │
│  VM will be unavailable during update.   │
│                                          │
│  [Enter] Update now   [Esc] Cancel       │
└──────────────────────────────────────────┘
```

**Interactions:**
- `Enter` — POST /api/v1/instances/:id/image-update, dismiss modal
- `Esc` — cancel

**SSE follow-up:**
- When update completes, SSE sends `instance.image_update_ready` event
- TUI shows toast notification: "Image update ready!"

---

### 7. Toast Notifications

**Examples:**
```
✅ Instance started (top-right, 3 seconds)

⚠️  SSH connection lost (red, top-right, 10 seconds)

📬 Image update ready (blue, top-right, 5 seconds)

🔄 Syncing... (spinner, bottom-left)
```

**Positioning:** Top-right corner, non-blocking, auto-dismiss after 3–10 seconds.

---

## Navigation Flow

```
startup
  └─ read config.json
      ├─ [no token or expired]
      │  └─ show LoginScreen
      │      └─ user logs in
      │          └─ write token + expires_at
      │              └─ transition to MainScreen
      │
      └─ [token valid]
         └─ show MainScreen
             ├─ GET /instances (polling every 10s)
             ├─ SSE /instances/events subscribe
             └─ user selects instance
                 └─ show detail pane
                     ├─ [c] keypress → exit TUI, exec `ssh ubuntu@host`
                     ├─ [f] keypress → show forward modal
                     ├─ [u] keypress → show update modal (if available)
                     ├─ [d] keypress → show delete confirm modal
                     ├─ [R] keypress → show resize modal
                     ├─ [s] keypress → POST /instances/:id/start
                     ├─ [S] keypress → POST /instances/:id/stop
                     └─ [r] keypress → POST /instances/:id/restart
```

---

## Keybindings Reference

### Navigation (MainScreen)
| Key | Action |
|---|---|
| `↑` / `↓` | Navigate instance list |
| `←` / `→` | (reserved for pane focus in future) |
| `Enter` / `Space` | (not used; selection automatic) |

### Operations
| Key | Action |
|---|---|
| `c` | Connect via SSH |
| `f` | Forward local port |
| `u` | Update image |
| `d` | Delete instance |
| `s` | Start instance |
| `S` | Stop instance |
| `r` | Restart instance |
| `R` | Resize tier |
| `e` | Expose port (public URL) |

### Global
| Key | Action |
|---|---|
| `Ctrl+C` / `q` | Quit |
| `Ctrl+R` | Force refresh instances |
| `?` | Show help |

### Modal-specific
| Key | Action |
|---|---|
| `Enter` | Confirm action |
| `Esc` | Dismiss modal |
| `y` / `n` | Yes / No (confirm modal only) |
| `Tab` | Next field (form modals) |

---

## Visual Style

**Color scheme (dark terminal):**
- Status badges: ✅ `#90EE90` green, ⏸️ `#A9A9A9` gray, 🟡 `#FFD700` yellow
- Tier badges: `#6495ED` starter blue, `#9370DB` builder purple, `#FFB347` pro gold
- Links/SSH text: `#87CEEB` sky blue (underlined)
- Error text: `#FF6B6B` red
- Success text: `#51CF66` bright green
- Borders: `#404040` dark gray

**Typography:**
- Monospace font (terminal default)
- Field labels in bold
- Hints in dim/gray text
- Status strings in dim text + colored prefix

**Responsive:**
- Below 80 columns: collapse to list-only, show warning "Terminal too narrow"
- Below 24 rows: hide footer hints, show "↑↓" tip

---

## SSH Connect Behavior

When user presses `[c]`:

1. **TUI Teardown:** Exit Bubble Tea alt-screen mode, restore terminal
2. **Spawn SSH:** Exec `ssh -i ~/.ssh/id_ed25519 ubuntu@host` with inherited stdin/stdout/stderr
3. **Block:** The `aidev` process waits for SSH subprocess to exit
4. **SSH Session:** User interacts with remote shell normally
5. **Exit:** SSH subprocess closes
6. **Restart TUI:** (optional) Restart the TUI, or just exit and user runs `aidev` again

This pattern is used by k9s, lazygit, and other mature TUIs.

---

## Error Handling

**API errors:**
- 401 Unauthorized → refresh token, retry. If refresh fails → logout, show LoginScreen
- 404 Not Found → show toast "Instance not found. Refresh?" and remove from list
- 5xx Server Error → show toast "Server error. Retry?" with exponential backoff

**Network errors:**
- Connection timeout → show spinner "Connecting..." in status bar
- SSE disconnect → auto-reconnect with exponential backoff
- No response → show toast "Network timeout" after 30 seconds

**User errors:**
- Invalid email/password → show red banner "Invalid credentials"
- Malformed API key → show red banner "Invalid API key format"

---

## Future Extensions

- **Logs:** Press `l` to stream VM logs in a split pane
- **Metrics:** Press `m` to show CPU/disk/memory graphs
- **Templates:** Create instances from templates
- **Settings:** Press `=` to configure refresh rate, theme, etc.
- **Workspaces:** Manage multiple API endpoints / user accounts
