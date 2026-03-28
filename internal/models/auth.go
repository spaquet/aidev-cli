package models

// LoginRequest represents credentials for logging in
type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

// LoginResponse is returned from the auth/login endpoint
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	User      User   `json:"user"`
}

// User represents authenticated user info
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// RefreshRequest is sent to auth/refresh
type RefreshRequest struct {
	Token string `json:"token"`
}

// RefreshResponse is returned from auth/refresh
type RefreshResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// DeviceAuthResponse is returned from auth/device
type DeviceAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DevicePollRequest is sent to auth/device/token
type DevicePollRequest struct {
	DeviceCode string `json:"device_code"`
}

// Config represents the stored configuration file
type Config struct {
	BaseURL        string `json:"base_url"`
	Token          string `json:"token"`
	TokenExpiresAt string `json:"token_expires_at"`
	UserEmail      string `json:"user_email"`
}
