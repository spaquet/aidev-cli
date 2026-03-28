# AIDev CLI — Configuration Guide

This guide covers configuration options for the AIDev CLI, including the config file format, environment variables, and advanced customization.

## Table of Contents

- [Config File Location](#config-file-location)
- [Config File Format](#config-file-format)
- [Environment Variables](#environment-variables)
- [CLI Flags](#cli-flags)
- [Advanced Configuration](#advanced-configuration)
- [Security Considerations](#security-considerations)

## Config File Location

The AIDev CLI stores configuration in an XDG-compliant location:

| OS | Location |
|----|----------|
| macOS | `~/.config/aidev/config.json` |
| Linux | `~/.config/aidev/config.json` (or `$XDG_CONFIG_HOME/aidev/config.json`) |
| Windows | `%APPDATA%\aidev\config.json` |

File permissions are automatically set to `0600` (readable/writable by owner only) for security.

### View Your Config

```bash
aidev config show
```

This displays your current configuration (with sensitive values masked).

### Reset Config

```bash
aidev config reset
```

This resets the config to defaults and removes stored tokens.

## Config File Format

The config file is stored as JSON. Here's the complete structure:

```json
{
  "api_url": "https://api.sandbox.example.com",
  "user": {
    "email": "user@example.com",
    "id": "usr_123abc"
  },
  "auth": {
    "token": "eyJhbGc...",
    "token_expires_at": "2026-04-27T20:03:00Z",
    "refresh_token": "ref_456def"
  },
  "preferences": {
    "theme": "auto",
    "notifications": true,
    "default_editor": "vim"
  },
  "ssh": {
    "key_path": "~/.ssh/id_ed25519",
    "known_hosts_file": "~/.ssh/known_hosts",
    "strict_host_key_checking": "accept-new"
  }
}
```

### Field Descriptions

#### `api_url` (string)
The base URL for the AIDev API. Default: `https://api.sandbox.example.com`

#### `user` (object)
Current logged-in user information:
- `email` (string): User's email address
- `id` (string): User ID from the API

#### `auth` (object)
Authentication tokens (managed automatically):
- `token` (string): Current JWT access token
- `token_expires_at` (string): ISO 8601 expiration timestamp
- `refresh_token` (string): Token for refreshing the access token

**Note**: Tokens are refreshed automatically. You don't need to manually manage these.

#### `preferences` (object)
User preferences:
- `theme` (string): `"auto"`, `"light"`, or `"dark"`
- `notifications` (boolean): Enable/disable toast notifications
- `default_editor` (string): Default editor for tasks (`vim`, `nano`, `emacs`, etc.)

#### `ssh` (object)
SSH configuration:
- `key_path` (string): Path to your SSH private key (default: auto-detected)
- `known_hosts_file` (string): Path to SSH known_hosts file
- `strict_host_key_checking` (string): `"accept-new"` or `"yes"`

## Environment Variables

You can override config values using environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `AIDEV_API_URL` | API base URL | `https://api.example.com` |
| `AIDEV_TOKEN` | JWT token (overrides stored token) | `eyJhbGc...` |
| `AIDEV_SSH_KEY` | SSH private key path | `~/.ssh/id_rsa` |
| `AIDEV_CONFIG_DIR` | Config directory | `~/.config/aidev` |
| `XDG_CONFIG_HOME` | XDG config home (Linux) | `~/.config` |

### Example: Override API URL

```bash
AIDEV_API_URL=https://staging-api.example.com aidev
```

### Example: Use Custom SSH Key

```bash
AIDEV_SSH_KEY=~/.ssh/special_key aidev ssh my-instance
```

## CLI Flags

Global flags apply to all commands:

```bash
aidev [global-flags] [command] [command-flags]
```

### Global Flags

| Flag | Description |
|------|-------------|
| `--api <url>` | Override API base URL |
| `--config <path>` | Override config file path |
| `--version` | Show version and check for updates |
| `--help` | Show help text |

### Examples

```bash
# Use staging API
aidev --api https://staging-api.example.com tui

# Use custom config location
aidev --config /etc/aidev/config.json instances
```

## Advanced Configuration

### Custom SSH Key

The CLI auto-detects your SSH key from `~/.ssh/id_*`. To use a specific key:

```bash
# Set in config
aidev config set ssh.key_path ~/.ssh/special_key

# Or use environment variable
export AIDEV_SSH_KEY=~/.ssh/special_key
aidev ssh my-instance
```

### Multiple Accounts

To switch between accounts:

1. Log out: `aidev config reset`
2. Log in with new account: `aidev login`

Each account uses the same config file location, so switching will overwrite the previous account's token.

Alternatively, use multiple config files with the `--config` flag:

```bash
AIDEV_CONFIG_DIR=~/.config/aidev-staging aidev login
AIDEV_CONFIG_DIR=~/.config/aidev-staging aidev tui
```

### Theming

Currently, the TUI uses a fixed color scheme optimized for both light and dark terminals. Future versions may support customizable themes:

```bash
aidev config set preferences.theme dark
```

Supported values: `auto`, `light`, `dark`

### Notifications

Disable toast notifications (useful for scripts):

```bash
aidev config set preferences.notifications false
```

### API Server Self-Signed Certificates

If your API server uses self-signed certificates:

```bash
export INSECURE_SKIP_VERIFY=true
aidev login
```

**Warning**: Only use this in development. Self-signed certs are a security risk in production.

### Proxy Configuration

HTTP proxies can be configured using standard environment variables:

```bash
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
aidev tui
```

## Config Priority

Settings are applied in this order (later values override earlier ones):

1. **Defaults** (built-in)
2. **Config file** (`~/.config/aidev/config.json`)
3. **Environment variables** (`AIDEV_*`)
4. **CLI flags** (`--api`, `--config`, etc.)

Example: If you set `api_url` in your config file but also pass `--api https://other.com`, the CLI flag takes precedence.

## Security Considerations

### Token Storage

- **Permissions**: Config file is created with `mode 0600` (readable only by owner)
- **Encryption**: Tokens are stored in plain text. Consider using:
  - File encryption (e.g., `ecryptfs` on Linux, FileVault on macOS)
  - Credential storage system (e.g., `pass`, 1Password, Keychain)
- **Expiration**: Access tokens expire automatically and are refreshed as needed

### SSH Key Security

- **Private Keys**: Never stored in config; always loaded from `~/.ssh/`
- **Permissions**: SSH keys should have `mode 0600`
- **No Passwords**: The CLI never stores or transmits SSH passwords

### API URL Validation

Always use HTTPS for your API URL:

```bash
# ✅ Good
aidev --api https://api.example.com

# ❌ Bad
aidev --api http://api.example.com
```

### Credential Handling

Never commit your config file to version control:

```bash
# .gitignore
~/.config/aidev/config.json
```

### Reset on Shared Systems

If using AIDev CLI on a shared system, reset your config when done:

```bash
aidev config reset
```

This removes your token and auth information.

## Troubleshooting Configuration

### Config File Not Found

```
Error: Failed to load config: no such file or directory
```

Solution: Run `aidev login` to create the config file.

### Permission Denied

```
Error: Permission denied: ~/.config/aidev/config.json
```

Solution: Check file permissions and ownership:

```bash
ls -la ~/.config/aidev/config.json
# Should show: -rw------- (0600)

# If permissions are wrong, reset them:
chmod 600 ~/.config/aidev/config.json
```

### Token Expired

```
Error: Token expired, please login again
```

Solution: The CLI should auto-refresh tokens. If this persists:

```bash
aidev login
```

### Config Directory Missing

```
Error: config directory does not exist
```

Solution: Create the directory:

```bash
mkdir -p ~/.config/aidev
aidev login
```

### Environment Variables Not Applying

Ensure you're exporting variables correctly:

```bash
export AIDEV_API_URL=https://staging.example.com
aidev --version  # Should use staging URL

# Not exported (won't work):
AIDEV_API_URL=https://staging.example.com aidev --version
```

Actually, the above **does** work for a single command. To persist across commands:

```bash
export AIDEV_API_URL=https://staging.example.com
```

## Examples

### Development Setup

Use staging API and custom SSH key:

```bash
aidev config set api_url https://staging-api.example.com
export AIDEV_SSH_KEY=~/.ssh/staging_key
aidev tui
```

### Production Setup

Use production API with notifications:

```bash
aidev config set api_url https://api.example.com
aidev config set preferences.notifications true
```

### CI/CD Integration

Disable notifications and use API key:

```bash
export AIDEV_CONFIG_DIR=/tmp/aidev-ci
export AIDEV_TOKEN=$(cat /secrets/aidev-token)
aidev instances --json | jq '.[]' | head -5
```

### Multiple Environments

Keep separate configs:

```bash
# Staging
aidev --config ~/.config/aidev-staging/config.json tui

# Production
aidev --config ~/.config/aidev-prod/config.json tui
```

## Configuration Defaults

Here are the built-in defaults:

```json
{
  "api_url": "https://api.sandbox.example.com",
  "preferences": {
    "theme": "auto",
    "notifications": true,
    "default_editor": "vim"
  },
  "ssh": {
    "strict_host_key_checking": "accept-new"
  }
}
```

To see your current configuration:

```bash
aidev config show
```

To reset to defaults:

```bash
aidev config reset
```

---

**Version**: Check your installed version with `aidev --version`

**Last Updated**: 2026-03-27

**Related Documentation**:
- [Getting Started](GETTING_STARTED.md) — Installation and quick start
- [API Reference](rails-api-spec.md) — Backend API specification
- [TUI Design](tui-design.md) — User interface documentation
