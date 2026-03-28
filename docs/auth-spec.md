# AIDev Authentication Specification

This document describes how authentication works between the `aidev` CLI/TUI and the Rails backend.

---

## Architecture Overview

```
┌──────────────┐  POST /api/v1/auth/device     ┌──────────────────┐
│              │  ────────────────────────────> │                  │
│   aidev      │  { device_code, user_code,    │   Rails API      │
│   (CLI/TUI)  │    verification_uri }         │   (JWT provider) │
│              │  <──────────────────────────── │                  │
│              │                                │                  │
│              │  ┌─ Opens browser ──┐          │                  │
│              │  │ (auto-opens URL)│          │                  │
│              │  └─────────────────┘          │                  │
│              │                                │                  │
│              │  POST /api/v1/auth/device/token (polling)         │
│              │  ────────────────────────────> │                  │
│              │  { device_code }               │                  │
│              │                                │                  │
│              │  <──────────────────────────── │                  │
│              │  { token, expires_at, user }   │                  │
│              │                                │                  │
│  ┌──────────┐│  GET /instances                │                  │
│  │config.json │  Header: Authorization: Bearer <token>          │
│  └──────────┘│  ───────────────────────────> │                  │
│              │                                │                  │
│              │  <─────────────────────────── │                  │
│              │  [...instances]                │                  │
│              │                                │                  │
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

### 4. Device Flow Login

**User action:** Start TUI or `aidev login` — opens browser for authentication.

**Step 1: Initiate device authorization**

**Request:**
```
POST /api/v1/auth/device
Content-Type: application/json
(empty body)
```

**Response (200):**
```json
{
  "device_code": "dev_abc123xyz456...",
  "user_code": "AIDEV-WXYZ",
  "verification_uri": "https://app.aidev.io/device",
  "expires_in": 300,
  "interval": 5
}
```

**Step 2: Display code and open browser**

The TUI/CLI displays:
```
Code: AIDEV-WXYZ
Visit: https://app.aidev.io/device
```

And automatically opens the verification URL in the default browser.

**Step 3: Poll for completion**

Every `interval` seconds (5 by default), the CLI polls:

**Request:**
```
POST /api/v1/auth/device/token
Content-Type: application/json

{
  "device_code": "dev_abc123xyz456..."
}
```

**Response (200) — Authorized:**
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

**Response (428) — Still waiting:**
```json
{
  "error": "authorization_pending"
}
```

**Response (400) — Code expired or denied:**
```json
{
  "error": "expired_token"
}
```
or
```json
{
  "error": "access_denied"
}
```

**On success (200):**
1. Write token, expires_at, base_url to `config.json` (file mode 0600)
2. Navigate to MainScreen (TUI) or exit with confirmation (CLI)

**On error (400/428/timeout):**
1. Show error message ("Code expired", "Access denied", etc.)
2. TUI: Allow retry with [Enter]
3. CLI: Exit with error, user can retry `aidev login`

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

## Rails Implementation Notes

### Gems Required

```ruby
# Gemfile
gem 'jwt'  # or 'json-jwt'
gem 'securerandom'  # for device code generation
```

### Device Flow Implementation

```ruby
# db/migrate/xxx_create_device_authorizations.rb
create_table :device_authorizations do |t|
  t.string :device_code, null: false, index: { unique: true }
  t.string :user_code, null: false
  t.bigint :user_id, null: true, foreign_key: true
  t.datetime :expires_at, null: false
  t.datetime :approved_at, null: true
  t.string :status, default: 'pending'  # pending, approved, denied, expired
  t.timestamps
end

# app/models/device_authorization.rb
class DeviceAuthorization < ApplicationRecord
  belongs_to :user, optional: true

  DEVICE_CODE_LENGTH = 32
  USER_CODE_LENGTH = 8
  EXPIRES_IN = 300  # 5 minutes
  INTERVAL = 5     # seconds

  def self.create_authorization
    create!(
      device_code: SecureRandom.hex(DEVICE_CODE_LENGTH / 2),
      user_code: SecureRandom.hex(USER_CODE_LENGTH / 2).upcase,
      expires_at: Time.current + EXPIRES_IN.seconds
    )
  end

  def expired?
    expires_at < Time.current
  end

  def approved?
    status == 'approved' && user_id.present?
  end
end

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
  skip_before_action :authorize_user, only: [:device, :device_token, :refresh]

  # Initiate device flow
  def device
    auth = DeviceAuthorization.create_authorization

    render json: {
      device_code: auth.device_code,
      user_code: auth.user_code,
      verification_uri: "#{ENV['DASHBOARD_URL']}/device",
      expires_in: DeviceAuthorization::EXPIRES_IN,
      interval: DeviceAuthorization::INTERVAL
    }
  end

  # Poll for token
  def device_token
    device_code = params[:device_code]
    auth = DeviceAuthorization.find_by(device_code: device_code)

    # Check if code exists
    render json: { error: 'invalid_device_code' }, status: :bad_request and return unless auth

    # Check if expired
    if auth.expired?
      render json: { error: 'expired_token' }, status: :bad_request
      return
    end

    # Check if denied
    if auth.status == 'denied'
      render json: { error: 'access_denied' }, status: :bad_request
      return
    end

    # Check if approved
    unless auth.approved?
      render json: { error: 'authorization_pending' }, status: 428
      return
    end

    # Success: issue token
    user = auth.user
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
  end

  # Refresh token
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

  # Logout
  def logout
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

### Dashboard Implementation

The web dashboard provides the approval UI at `/device`:

1. User visits `https://app.aidev.io/device`
2. Enters the user code (AIDEV-WXYZ) or scans QR code
3. Dashboard looks up `DeviceAuthorization` by `user_code`
4. If found and not expired, displays approval prompt
5. On approve, updates `DeviceAuthorization` with `status: 'approved'` and `user_id: current_user.id`
6. CLI polls `/device/token` endpoint and receives token on next poll

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
describe 'POST /api/v1/auth/device' do
  it 'returns device code and user code' do
    post '/api/v1/auth/device'

    expect(response).to have_http_status(:ok)
    expect(json['device_code']).to be_present
    expect(json['user_code']).to be_present
    expect(json['verification_uri']).to be_present
    expect(json['expires_in']).to eq(300)
    expect(json['interval']).to eq(5)
  end
end

describe 'POST /api/v1/auth/device/token' do
  it 'returns 428 while authorization is pending' do
    auth = create(:device_authorization)
    post '/api/v1/auth/device/token', params: { device_code: auth.device_code }

    expect(response).to have_http_status(428)
    expect(json['error']).to eq('authorization_pending')
  end

  it 'returns token after user approves' do
    user = create(:user)
    auth = create(:device_authorization, user: user, status: 'approved')
    post '/api/v1/auth/device/token', params: { device_code: auth.device_code }

    expect(response).to have_http_status(:ok)
    expect(json['token']).to be_present
    expect(json['user']['email']).to eq(user.email)
  end

  it 'returns 400 for denied authorization' do
    auth = create(:device_authorization, status: 'denied')
    post '/api/v1/auth/device/token', params: { device_code: auth.device_code }

    expect(response).to have_http_status(:bad_request)
    expect(json['error']).to eq('access_denied')
  end

  it 'returns 400 for expired code' do
    auth = create(:device_authorization, expires_at: 10.minutes.ago)
    post '/api/v1/auth/device/token', params: { device_code: auth.device_code }

    expect(response).to have_http_status(:bad_request)
    expect(json['error']).to eq('expired_token')
  end
end
```

### Integration Tests (Go TUI)

```go
func TestDeviceFlow(t *testing.T) {
  // Start test server
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/api/v1/auth/device" && r.Method == "POST" {
      w.Header().Set("Content-Type", "application/json")
      json.NewEncoder(w).Encode(map[string]interface{}{
        "device_code":       "dev_abc123",
        "user_code":         "AIDEV-WXYZ",
        "verification_uri":  "https://example.com/device",
        "expires_in":        300,
        "interval":          5,
      })
    }
    if r.URL.Path == "/api/v1/auth/device/token" && r.Method == "POST" {
      w.Header().Set("Content-Type", "application/json")
      json.NewEncoder(w).Encode(map[string]interface{}{
        "token":      "fake-jwt-token",
        "expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
        "user": map[string]interface{}{
          "id":    "user_123",
          "email": "test@example.com",
          "name":  "Test User",
        },
      })
    }
  }))
  defer server.Close()

  // Test device flow
  client := api.NewClient(server.URL)

  // Step 1: Get device code
  deviceResp, err := client.DeviceAuthorize()
  assert.NoError(t, err)
  assert.Equal(t, "AIDEV-WXYZ", deviceResp.UserCode)

  // Step 2: Poll for token (after user approves in dashboard)
  loginResp, err := client.DevicePoll(deviceResp.DeviceCode)
  assert.NoError(t, err)
  assert.Equal(t, "fake-jwt-token", loginResp.Token)
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
