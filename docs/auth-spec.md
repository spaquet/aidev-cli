# AIDev Authentication Specification

This document describes how authentication works between the `aidev` CLI/TUI and the Rails backend.

---

## Architecture Overview

```
┌──────────────┐                              ┌──────────────────┐
│              │  POST /auth/login            │                  │
│   aidev      │  (email + password)          │   Rails API      │
│   (CLI/TUI)  │  ───────────────────────────>│   (JWT provider) │
│              │                              │                  │
│              │  <───────────────────────────│                  │
│              │  { token, expires_at, user } │                  │
│              │                              │                  │
│  ┌──────────┐│  GET /instances              │                  │
│  │config.json │  Header: Authorization: Bearer <token>        │
│  └──────────┘│  ───────────────────────────>│                  │
│              │                              │                  │
│              │  <───────────────────────────│                  │
│              │  [...instances]              │                  │
│              │                              │                  │
└──────────────┘                              └──────────────────┘
```

---

## JWT Token Format

The Rails backend issues JWT tokens in the following format:

**Header:**
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Payload:**
```json
{
  "sub": "user_123",
  "email": "alice@example.com",
  "iat": 1711000000,
  "exp": 1711086400,
  "iss": "aidev-api",
  "scope": "api"
}
```

**Fields:**
- `sub` (subject) — unique user ID (user_123)
- `email` — user's email address
- `iat` (issued at) — Unix timestamp when token was created
- `exp` (expiration) — Unix timestamp when token expires (24 hours after `iat`)
- `iss` (issuer) — set to "aidev-api"
- `scope` — "api" (other values: "web", "admin" for future use)

**Secret:** Signed with a strong secret key (≥256 bits) stored in Rails env var `JWT_SECRET_KEY`.

---

## Token Lifecycle in the TUI

### 1. Startup

The TUI reads `$XDG_CONFIG_HOME/aidev/config.json`:

```json
{
  "base_url": "https://api.sandbox.example.com",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_expires_at": "2025-04-27T12:34:56Z",
  "user_email": "alice@example.com"
}
```

**Decision:**
- If file missing → show LoginScreen
- If `token_expires_at` is in the past → refresh token (see step 2)
  - If refresh succeeds → proceed to MainScreen
  - If refresh fails → show LoginScreen
- If token valid → proceed to MainScreen

### 2. Token Refresh

When token is within 1 hour of expiration, the TUI proactively refreshes it.

**Request:**
```
POST /api/v1/auth/refresh
Content-Type: application/json

{ "token": "eyJhbGci..." }
```

**Response:**
```json
{
  "token": "eyJhbGci...",
  "expires_at": "2025-04-28T12:34:56Z"
}
```

**On success:** Update `config.json` with new token and expiration.

**On failure (401):** Delete `config.json`, show LoginScreen.

### 3. API Requests

Every API request includes the token in the Authorization header:

```
GET /api/v1/instances
Authorization: Bearer eyJhbGci...
```

**Response:**
- `200` / `201` / `204` — success, continue
- `401 Unauthorized` — token invalid or expired
  - Attempt refresh (step 2)
  - If refresh succeeds → retry original request with new token
  - If refresh fails → delete config.json, show LoginScreen

### 4. Login Flow

**User action:** Enter email + password (or API key) on LoginScreen.

**Request:**
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "alice@example.com",
  "password": "correct-horse-battery-staple"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGci...",
  "expires_at": "2025-04-27T12:34:56Z",
  "user": {
    "id": "user_123",
    "email": "alice@example.com",
    "name": "Alice Chen"
  }
}
```

**On success:**
1. Write token, expires_at, base_url to `config.json` (file mode 0600)
2. Navigate to MainScreen

**On failure (401):**
1. Show red error banner: "Invalid email or password"
2. Stay on LoginScreen (clear password field)

### 5. Logout

**User action:** Press `q` or select Logout from menu.

**Request:**
```
DELETE /api/v1/auth/logout
Authorization: Bearer eyJhbGci...
```

**Response (204):**
1. Delete `config.json`
2. Navigate to LoginScreen

**Note:** Logout is optional. The TUI can simply discard the token without notifying the server. The token will naturally expire after 24 hours.

---

## Configuration Storage

### File Location

Token and configuration are stored at:

**Linux & macOS:**
```
$XDG_CONFIG_HOME/aidev/config.json
(defaults to ~/.config/aidev/config.json if XDG_CONFIG_HOME not set)
```

**Windows:**
```
%APPDATA%\aidev\config.json
(e.g., C:\Users\alice\AppData\Roaming\aidev\config.json)
```

Use `github.com/adrg/xdg` Go package for portable path handling.

### File Format

```json
{
  "base_url": "https://api.sandbox.example.com",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_expires_at": "2025-04-27T12:34:56Z",
  "user_email": "alice@example.com"
}
```

### File Permissions

- **Creation:** `0600` (read/write for owner only)
- **Read:** Check permissions before reading; warn if world-readable
- **Rotation:** When updating token, overwrite in-place or atomically rename

### Optional: Keyring Integration

On macOS, tokens can be stored in the system Keychain instead of disk:

- Use `github.com/zalando/go-keyring`
- Service: `aidev`
- Account: `{base_url}@{user_email}` (e.g., `https://api.sandbox.example.com@alice@example.com`)
- Store token in Keyring, config.json on disk

Benefits:
- Token never written to disk
- Integrated with OS credential management
- Survives `~/.config` backup/migration

Implementation:
```go
import "github.com/zalando/go-keyring"

// Store
keyring.Set("aidev", email, token)

// Retrieve
token, err := keyring.Get("aidev", email)
```

---

## API Key Authentication

Users can create static API keys for non-interactive use (scripts, CI/CD).

### Key Format

API keys are prefixed with `aidev_sk_` followed by a 32-character secret:

```
aidev_sk_abc123def456ghi789jkl012mnop
```

### Creation (via web dashboard)

1. User logs in to dashboard
2. Navigate to Settings → API Keys
3. Click "Create new key"
4. Provide optional label and expiration (default: never)
5. Show key once, then hide

Key is stored hashed (bcrypt) in the `api_keys` table:
```sql
CREATE TABLE api_keys (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL FOREIGN KEY,
  key_hash VARCHAR NOT NULL UNIQUE,
  label VARCHAR,
  last_used_at TIMESTAMP,
  expires_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

### Usage in TUI

User can log in with an API key instead of password:

**LoginScreen option:**
```
Email: [alice@example.com__________]
Or paste an API key: [aidev_sk_abc123__________]
```

**Request:**
```
POST /api/v1/auth/login
Content-Type: application/json

{ "api_key": "aidev_sk_abc123..." }
```

**Backend behavior:**
1. Extract prefix: should start with `aidev_sk_`
2. Lookup key hash in `api_keys` table (compare hashed value)
3. If found, issue JWT for the associated user_id
4. Update `last_used_at` timestamp
5. Return token + user info (same as password auth)

### CLI Usage (non-TUI)

Users can also set environment variable:
```bash
export AIDEV_TOKEN=aidev_sk_abc123...
aidev instances list
```

The API client uses env var if present, else reads from config.json.

---

## Rails Implementation Notes

### Gems Required

```ruby
# Gemfile
gem 'jwt'  # or 'json-jwt'
gem 'bcrypt'  # password hashing
gem 'doorkeeper'  # optional: OAuth2 support for future
```

### Minimal JWT Implementation

```ruby
# app/services/auth_service.rb
class AuthService
  SECRET_KEY = ENV.fetch('JWT_SECRET_KEY')

  def self.issue_token(user)
    payload = {
      sub: user.id,
      email: user.email,
      iat: Time.current.to_i,
      exp: 24.hours.from_now.to_i,
      iss: 'aidev-api',
      scope: 'api'
    }
    JWT.encode(payload, SECRET_KEY, 'HS256')
  end

  def self.verify_token(token)
    JWT.decode(token, SECRET_KEY, true, algorithm: 'HS256')
  rescue JWT::DecodeError, JWT::ExpiredSignature
    nil
  end
end

# app/controllers/api/v1/auth_controller.rb
class Api::V1::AuthController < Api::V1::BaseController
  skip_before_action :authorize_user, only: [:login, :refresh]

  def login
    user = User.find_by(email: params[:email])

    if user&.authenticate(params[:password])
      token = AuthService.issue_token(user)
      render json: {
        token: token,
        expires_at: 24.hours.from_now.iso8601,
        user: {
          id: user.id,
          email: user.email,
          name: user.name
        }
      }
    else
      render json: { error: 'Invalid credentials' }, status: :unauthorized
    end
  end

  def refresh
    decoded = AuthService.verify_token(params[:token])
    if decoded
      user = User.find(decoded['sub'])
      new_token = AuthService.issue_token(user)
      render json: {
        token: new_token,
        expires_at: 24.hours.from_now.iso8601
      }
    else
      render json: { error: 'Invalid token' }, status: :unauthorized
    end
  end

  def logout
    # Mark token as revoked (optional, could use Redis blacklist)
    # or just return 204 and let client discard token
    render json: {}, status: :no_content
  end
end

# app/controllers/api/v1/base_controller.rb
class Api::V1::BaseController < ApplicationController
  before_action :authorize_user

  private

  def authorize_user
    token = request.headers['Authorization']&.split(' ')&.last
    decoded = AuthService.verify_token(token) if token
    @current_user = User.find(decoded['sub']) if decoded

    render json: { error: 'Unauthorized' }, status: :unauthorized unless @current_user
  end
end
```

### Password Security

- Minimum 12 characters
- Require uppercase, lowercase, digit, special char
- Hash with bcrypt cost ≥ 12:
  ```ruby
  BCrypt::Password.create(password, cost: 12)
  ```
- Enforce rate limiting: max 10 failed login attempts per IP per minute
- Never log passwords, only log login attempt (email, IP, timestamp, success/failure)

### Secret Key Generation

```bash
# Generate 256-bit secret for JWT signing
openssl rand -hex 32
# Output: abc123def456...

# Store in .env or Rails credentials
EDITOR=nano rails credentials:edit
```

In production, rotate this secret every 6 months. During rotation:
1. Keep old secret in `OLD_JWT_SECRET_KEY`
2. Try decoding with current key, fall back to old key
3. Re-issue new token with current key
4. Eventually stop accepting old key

### Token Blacklist / Logout (Optional)

For production, implement a token revocation list (Redis-backed):

```ruby
# app/services/token_blacklist.rb
class TokenBlacklist
  def self.revoke(token)
    decoded = AuthService.verify_token(token)
    exp_time = Time.at(decoded['exp'])
    ttl = (exp_time - Time.current).to_i

    Redis.current.setex("revoked_#{token}", ttl, true)
  end

  def self.revoked?(token)
    Redis.current.exists?("revoked_#{token}")
  end
end

# In base_controller
def authorize_user
  # ... verify token ...
  render json: { error: 'Token revoked' }, status: :unauthorized if TokenBlacklist.revoked?(token)
  # ...
end
```

---

## Security Considerations

### HTTPS Only

- All API endpoints must use HTTPS
- TUI should warn if base_url is http:// (require user confirmation)
- Never transmit tokens over HTTP

### Token Storage

- **TUI:** Store in `config.json` with file mode `0600`
- **Web dashboard:** HTTP-only cookie (not localStorage)
- **CI/CD:** Environment variable (injected at runtime, never committed)

### Token Expiration

- **Access token:** 24 hours
- **Refresh token:** (not used in current design; token is long-lived)
- Users must re-login every 24 hours

### Rate Limiting

- Max 10 login attempts per IP per minute
- Max 100 API requests per user per minute
- Return `429 Too Many Requests` with `Retry-After` header

### Audit Logging

Log all auth events:
```
2025-03-27T12:34:56Z [login] alice@example.com success
2025-03-27T12:35:00Z [api_call] user_123 GET /instances 200
2025-03-27T12:36:00Z [login] bob@example.com failure (invalid password)
2025-03-27T12:37:00Z [logout] alice@example.com
```

Never log:
- Passwords
- Tokens (only hash or truncate)
- API keys (only hash or truncate)

---

## Testing

### Unit Tests (Rails)

```ruby
describe 'POST /api/v1/auth/login' do
  it 'returns token for valid credentials' do
    user = create(:user, password: 'secret123')
    post '/api/v1/auth/login', params: { email: user.email, password: 'secret123' }

    expect(response).to have_http_status(:ok)
    expect(json['token']).to be_present
    expect(json['user']['email']).to eq(user.email)
  end

  it 'returns 401 for invalid password' do
    user = create(:user)
    post '/api/v1/auth/login', params: { email: user.email, password: 'wrong' }

    expect(response).to have_http_status(:unauthorized)
  end

  it 'enforces rate limiting' do
    user = create(:user)
    15.times do
      post '/api/v1/auth/login', params: { email: user.email, password: 'wrong' }
    end

    expect(response).to have_http_status(:too_many_requests)
  end
end
```

### Integration Tests (Go TUI)

```go
func TestLoginAndFetch(t *testing.T) {
  // Start test server
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/api/v1/auth/login" {
      w.Header().Set("Content-Type", "application/json")
      json.NewEncoder(w).Encode(map[string]interface{}{
        "token":      "fake-jwt-token",
        "expires_at": time.Now().Add(24 * time.Hour),
        "user": map[string]interface{}{
          "id":    "user_123",
          "email": "test@example.com",
          "name":  "Test User",
        },
      })
    }
  }))
  defer server.Close()

  // Test login
  client := api.NewClient(server.URL)
  resp, err := client.Login("test@example.com", "password")

  assert.NoError(t, err)
  assert.Equal(t, "fake-jwt-token", resp.Token)
}
```

---

## Future: OAuth2 / OIDC

For web dashboard or mobile app, implement OAuth2:

```
POST /oauth/token
  grant_type: "authorization_code"
  code: "..."
  client_id: "..."
  client_secret: "..."
  redirect_uri: "https://myapp.example.com/callback"
```

Use `doorkeeper` gem for Rails. TUI does not need OAuth (direct API key auth is sufficient).
