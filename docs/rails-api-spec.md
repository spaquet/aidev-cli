# AIDev Rails API Specification

The Rails 8.1 backend provides a RESTful JSON API for the `aidev` TUI and CLI. All endpoints are under `/api/v1`.

**Base URL:** `https://api.sandbox.example.com` (user-configurable in TUI)

---

## Authentication

All endpoints except `/auth/login` and `/auth/refresh` require:
```
Authorization: Bearer <jwt_token>
```

### POST /api/v1/auth/login

Authenticate with email + password or API key.

**Request:**
```json
{
  "email": "alice@example.com",
  "password": "correct-horse-battery-staple"
}
```

**Alternative (API key):**
```json
{
  "api_key": "aidev_sk_abc123def456"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-04-27T12:34:56Z",
  "user": {
    "id": "user_123",
    "email": "alice@example.com",
    "name": "Alice Chen"
  }
}
```

**Errors:**
- `400 Bad Request` — missing email/password or api_key
- `401 Unauthorized` — invalid credentials
- `429 Too Many Requests` — rate limited after 10 failed attempts

**Notes:**
- Tokens are valid for 24 hours.
- API keys never expire but can be revoked.
- Password hashing: bcrypt with cost ≥ 12.

---

### POST /api/v1/auth/refresh

Refresh an expired token using the same token.

**Request:**
```json
{
  "token": "eyJhbGci..."
}
```

**Response (200):**
```json
{
  "token": "eyJhbGci...",
  "expires_at": "2025-04-28T12:34:56Z"
}
```

**Errors:**
- `401 Unauthorized` — token expired beyond grace period or invalid signature
- `400 Bad Request` — missing token

---

### DELETE /api/v1/auth/logout

Invalidate the current token (optional; TUI just discards token locally).

**Request:**
```
Authorization: Bearer <jwt_token>
```

**Response (204 No Content)**

---

## Instances

### GET /api/v1/instances

List all instances for the authenticated user.

**Query parameters:**
- `region` (optional) — filter by region (e.g., `?region=us-east-1`)

**Response (200):**
```json
{
  "instances": [
    {
      "id": "inst_abc123",
      "name": "my-builder",
      "status": "running",
      "tier": "builder",
      "region": "us-east-1",
      "disk_gb": 80,
      "disk_used_gb": 30,
      "ssh_host": "ssh.us-east-1.sandbox.example.com",
      "ssh_port": 22,
      "ssh_user": "ubuntu",
      "ssh_public_key_fingerprint": "SHA256:abcd1234...",
      "public_urls": [
        "https://8080-my-builder.tunnel.sandbox.example.com"
      ],
      "installed_tools": [
        "claude-code",
        "codex",
        "opencode",
        "gemini-cli",
        "nvim",
        "tmux"
      ],
      "image_version": "2025-03-27",
      "image_update_available": false,
      "created_at": "2025-02-01T00:00:00Z",
      "updated_at": "2025-03-27T12:00:00Z"
    },
    {
      "id": "inst_def456",
      "name": "learning-box",
      "status": "stopped",
      "tier": "starter",
      "region": "us-east-1",
      "disk_gb": 40,
      "disk_used_gb": 5,
      "ssh_host": "ssh.us-east-1.sandbox.example.com",
      "ssh_port": 22,
      "ssh_user": "ubuntu",
      "ssh_public_key_fingerprint": "SHA256:efgh5678...",
      "public_urls": [],
      "installed_tools": [
        "claude-code",
        "nvim"
      ],
      "image_version": "2025-03-24",
      "image_update_available": true,
      "created_at": "2025-03-10T00:00:00Z",
      "updated_at": "2025-03-27T12:00:00Z"
    }
  ]
}
```

**Errors:**
- `401 Unauthorized` — invalid or missing token

---

### POST /api/v1/instances

Create a new instance.

**Request:**
```json
{
  "name": "my-sandbox",
  "tier": "builder",
  "region": "us-east-1"
}
```

**Response (201 Created):**
```json
{
  "id": "inst_xyz789",
  "name": "my-sandbox",
  "status": "provisioning",
  "tier": "builder",
  "region": "us-east-1",
  "disk_gb": 80,
  "disk_used_gb": 0,
  "ssh_host": "ssh.us-east-1.sandbox.example.com",
  "ssh_port": 22,
  "ssh_user": "ubuntu",
  "ssh_public_key_fingerprint": "SHA256:...",
  "public_urls": [],
  "installed_tools": [],
  "image_version": "2025-03-27",
  "image_update_available": false,
  "created_at": "2025-03-27T12:34:56Z",
  "updated_at": "2025-03-27T12:34:56Z"
}
```

**Errors:**
- `400 Bad Request` — invalid tier or region, name already exists, or user quota exceeded
- `401 Unauthorized` — invalid token
- `402 Payment Required` — user's billing is inactive or card declined

**Status progression:**
- `provisioning` → `running` (takes 2–5 minutes)

---

### GET /api/v1/instances/:id

Get details of a specific instance.

**Response (200):**
```json
{
  "id": "inst_abc123",
  "name": "my-builder",
  "status": "running",
  "tier": "builder",
  "region": "us-east-1",
  "disk_gb": 80,
  "disk_used_gb": 30,
  "ssh_host": "ssh.us-east-1.sandbox.example.com",
  "ssh_port": 22,
  "ssh_user": "ubuntu",
  "ssh_public_key_fingerprint": "SHA256:abcd1234...",
  "public_urls": [
    "https://8080-my-builder.tunnel.sandbox.example.com"
  ],
  "installed_tools": [
    "claude-code",
    "codex",
    "nvim"
  ],
  "image_version": "2025-03-27",
  "image_update_available": false,
  "created_at": "2025-02-01T00:00:00Z",
  "updated_at": "2025-03-27T12:00:00Z"
}
```

**Errors:**
- `404 Not Found` — instance does not exist or does not belong to user
- `401 Unauthorized` — invalid token

---

### PATCH /api/v1/instances/:id

Update instance properties (tier, disk size).

**Request:**
```json
{
  "tier": "pro"
}
```

**Response (200):**
```json
{
  "id": "inst_abc123",
  "name": "my-builder",
  "status": "rebooting",
  "tier": "pro",
  "region": "us-east-1",
  "disk_gb": 160,
  "disk_used_gb": 30,
  ...
}
```

**Errors:**
- `400 Bad Request` — invalid tier, downgrade not allowed, or insufficient quota
- `404 Not Found` — instance not found
- `409 Conflict` — instance is not in a state that allows resize (e.g., updating image)

**Notes:**
- Resize triggers an automatic reboot (~2 minutes).
- Instance status becomes `rebooting` during the resize.
- Users cannot downgrade from `pro` to `builder` if they have 3+ public URLs exposed.

---

### DELETE /api/v1/instances/:id

Permanently delete an instance and all its data.

**Response (204 No Content)**

**Errors:**
- `404 Not Found` — instance not found
- `401 Unauthorized` — invalid token

**Notes:**
- This is irreversible. Data is deleted after 30 days (soft-delete for 30 days, then purged).
- The instance remains queryable for 30 days with `status: "deleted"`.

---

## Instance Operations

### POST /api/v1/instances/:id/start

Start a stopped instance.

**Response (200):**
```json
{
  "id": "inst_abc123",
  "status": "starting",
  ...
}
```

**Errors:**
- `409 Conflict` — instance is already running or in a non-compatible state (e.g., deleted)
- `404 Not Found` — instance not found

**Status progression:**
- `starting` → `running` (takes 30–60 seconds)

---

### POST /api/v1/instances/:id/stop

Stop a running instance.

**Response (200):**
```json
{
  "id": "inst_abc123",
  "status": "stopping",
  ...
}
```

**Errors:**
- `409 Conflict` — instance is already stopped

**Status progression:**
- `stopping` → `stopped` (takes 30–60 seconds)

---

### POST /api/v1/instances/:id/restart

Gracefully restart a running instance.

**Response (200):**
```json
{
  "id": "inst_abc123",
  "status": "restarting",
  ...
}
```

**Errors:**
- `409 Conflict` — instance is not running

**Status progression:**
- `restarting` → `running` (takes 60–120 seconds)

---

### POST /api/v1/instances/:id/image-update

Trigger an immediate image update on the instance.

**Request (optional body):**
```json
{}
```

**Response (200):**
```json
{
  "queued": true,
  "estimated_duration_minutes": 5,
  "estimated_completion_at": "2025-03-27T12:45:00Z"
}
```

**Errors:**
- `409 Conflict` — image update already in progress, or instance is not running
- `404 Not Found` — instance not found

**Behavior:**
- VM becomes unavailable during the update (status: `updating`).
- When complete, status returns to `running`.
- SSE sends `instance.image_update_ready` event.

---

## Port Exposure & Public URLs

### POST /api/v1/instances/:id/ports

Expose a port on the instance to a public URL.

**Request:**
```json
{
  "port": 3000,
  "protocol": "https",
  "subdomain": "api"
}
```

**Response (201 Created):**
```json
{
  "port": 3000,
  "protocol": "https",
  "url": "https://api-my-builder.tunnel.sandbox.example.com",
  "ssl_certificate_expires_at": "2026-03-27T00:00:00Z",
  "created_at": "2025-03-27T12:34:56Z"
}
```

**Errors:**
- `400 Bad Request` — invalid port (must be 1024–65535), protocol not `http`/`https`, or subdomain already taken
- `402 Payment Required` — user on starter tier (no public URLs), or quota exceeded (max 3 for builder, unlimited for pro)
- `404 Not Found` — instance not found
- `409 Conflict` — instance not running

**Notes:**
- Starter tier: 0 public URLs
- Builder tier: 1 public URL + auto-SSL via Let's Encrypt
- Pro tier: 3 public URLs + custom domain support + SSL
- Subdomains are auto-generated if not specified: `{port}-{instance-name}`
- URL format: `https://{subdomain}.tunnel.sandbox.example.com` (custom domain: `https://{custom-domain}`)

---

### DELETE /api/v1/instances/:id/ports/:port

Revoke public URL for a port.

**Response (204 No Content)**

**Errors:**
- `404 Not Found` — port exposure not found

---

## Real-time Events (SSE)

### GET /api/v1/instances/events

Subscribe to real-time instance events via Server-Sent Events.

**Headers:**
```
GET /api/v1/instances/events HTTP/1.1
Host: api.sandbox.example.com
Authorization: Bearer <jwt_token>
Accept: text/event-stream
Connection: keep-alive
```

**Response: 200 OK (streaming)**

The server sends `Content-Type: text/event-stream` and continuously sends events.

---

### Event Types

#### instance.status_changed

Fired when instance status changes.

```
event: instance.status_changed
data: {"id":"inst_abc123","status":"running","timestamp":"2025-03-27T12:34:56Z"}
```

#### instance.image_update_ready

Fired when an image update completes.

```
event: instance.image_update_ready
data: {"id":"inst_abc123","image_version":"2025-03-27","timestamp":"2025-03-27T12:34:56Z"}
```

#### instance.deleted

Fired when an instance is deleted.

```
event: instance.deleted
data: {"id":"inst_abc123","timestamp":"2025-03-27T12:34:56Z"}
```

#### instance.disk_warning

Fired when disk usage exceeds 80%.

```
event: instance.disk_warning
data: {"id":"inst_abc123","disk_used_gb":64,"disk_gb":80,"timestamp":"2025-03-27T12:34:56Z"}
```

#### ping (keep-alive)

Sent every 30 seconds to keep the connection alive.

```
: ping
```

---

## Error Response Format

All errors use this JSON format:

```json
{
  "error": {
    "code": "invalid_tier",
    "message": "Tier 'mega' does not exist",
    "details": {
      "valid_tiers": ["starter", "builder", "pro"]
    }
  }
}
```

**Common HTTP status codes:**
- `400 Bad Request` — validation error (missing field, invalid value)
- `401 Unauthorized` — missing or invalid token
- `402 Payment Required` — billing issue (tier exceeded, card declined, quota exceeded)
- `404 Not Found` — resource not found
- `409 Conflict` — resource state conflict (e.g., trying to start an already-running instance)
- `429 Too Many Requests` — rate limited
- `500 Internal Server Error` — server error (retry with exponential backoff)

---

## Rate Limiting

- **General:** 100 requests per minute per user
- **Login:** 10 attempts per minute per IP
- **Create instance:** 5 per minute per user

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1711000000
```

---

## Pagination (Future)

Currently, all list endpoints return all resources. Future versions may add:
```
GET /api/v1/instances?page=1&per_page=20
```

Response:
```json
{
  "instances": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 42,
    "total_pages": 3
  }
}
```

---

## CORS (if used by web UI)

The backend should include CORS headers for the web dashboard:
```
Access-Control-Allow-Origin: https://dashboard.sandbox.example.com
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE
Access-Control-Allow-Headers: Authorization, Content-Type
Access-Control-Allow-Credentials: true
```

For the CLI, CORS is not needed (no browser).

---

## Versioning

All endpoints are under `/api/v1`. When breaking changes occur:
- Create new `/api/v2` endpoints
- Keep `/api/v1` functional for 6+ months
- Document migration path
