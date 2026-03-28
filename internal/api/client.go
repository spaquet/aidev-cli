package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aidev/cli/internal/models"
)

// Client wraps HTTP operations with auth and error handling
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the JWT token for authenticated requests
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetToken returns the current token
func (c *Client) GetToken() string {
	return c.token
}

// Login authenticates with email/password or API key
func (c *Client) Login(email, password, apiKey string) (*models.LoginResponse, error) {
	req := models.LoginRequest{}
	if apiKey != "" {
		req.APIKey = apiKey
	} else {
		req.Email = email
		req.Password = password
	}

	var resp models.LoginResponse
	if err := c.postJSON("/api/v1/auth/login", req, &resp, false); err != nil {
		return nil, err
	}

	c.SetToken(resp.Token)
	return &resp, nil
}

// Refresh refreshes the current token
func (c *Client) Refresh(token string) (*models.RefreshResponse, error) {
	req := models.RefreshRequest{Token: token}
	var resp models.RefreshResponse

	if err := c.postJSON("/api/v1/auth/refresh", req, &resp, false); err != nil {
		return nil, err
	}

	c.SetToken(resp.Token)
	return &resp, nil
}

// Logout invalidates the current token
func (c *Client) Logout() error {
	return c.delete("/api/v1/auth/logout", true)
}

// DeviceAuthorize initiates a device authorization flow
func (c *Client) DeviceAuthorize() (*models.DeviceAuthResponse, error) {
	var resp models.DeviceAuthResponse
	if err := c.postJSON("/api/v1/auth/device", nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DevicePoll polls for the token result in device authorization flow
func (c *Client) DevicePoll(deviceCode string) (*models.LoginResponse, error) {
	req := models.DevicePollRequest{DeviceCode: deviceCode}
	var resp models.LoginResponse
	if err := c.postJSON("/api/v1/auth/device/token", req, &resp, false); err != nil {
		return nil, err
	}
	c.SetToken(resp.Token)
	return &resp, nil
}

// GetInstances fetches the list of instances
func (c *Client) GetInstances() (*models.InstancesResponse, error) {
	var resp models.InstancesResponse
	if err := c.getJSON("/api/v1/instances", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetInstance fetches a single instance by ID
func (c *Client) GetInstance(id string) (*models.Instance, error) {
	var inst models.Instance
	if err := c.getJSON(fmt.Sprintf("/api/v1/instances/%s", id), &inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

// CreateInstance creates a new instance
func (c *Client) CreateInstance(req *models.CreateInstanceRequest) (*models.Instance, error) {
	var inst models.Instance
	if err := c.postJSON("/api/v1/instances", req, &inst, true); err != nil {
		return nil, err
	}
	return &inst, nil
}

// UpdateInstance updates an instance (tier, disk, etc.)
func (c *Client) UpdateInstance(id string, req *models.UpdateInstanceRequest) (*models.Instance, error) {
	var inst models.Instance
	if err := c.patchJSON(fmt.Sprintf("/api/v1/instances/%s", id), req, &inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

// DeleteInstance deletes an instance
func (c *Client) DeleteInstance(id string) error {
	return c.delete(fmt.Sprintf("/api/v1/instances/%s", id), true)
}

// StartInstance starts a stopped instance
func (c *Client) StartInstance(id string) (*models.Instance, error) {
	var inst models.Instance
	if err := c.postJSON(fmt.Sprintf("/api/v1/instances/%s/start", id), nil, &inst, true); err != nil {
		return nil, err
	}
	return &inst, nil
}

// StopInstance stops a running instance
func (c *Client) StopInstance(id string) (*models.Instance, error) {
	var inst models.Instance
	if err := c.postJSON(fmt.Sprintf("/api/v1/instances/%s/stop", id), nil, &inst, true); err != nil {
		return nil, err
	}
	return &inst, nil
}

// RestartInstance restarts an instance
func (c *Client) RestartInstance(id string) (*models.Instance, error) {
	var inst models.Instance
	if err := c.postJSON(fmt.Sprintf("/api/v1/instances/%s/restart", id), nil, &inst, true); err != nil {
		return nil, err
	}
	return &inst, nil
}

// TriggerImageUpdate triggers an image update on an instance
func (c *Client) TriggerImageUpdate(id string) (*models.ImageUpdateResponse, error) {
	var resp models.ImageUpdateResponse
	if err := c.postJSON(fmt.Sprintf("/api/v1/instances/%s/image-update", id), nil, &resp, true); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Helper methods

func (c *Client) getJSON(path string, dst interface{}) error {
	url := c.baseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	return c.do(req, dst)
}

func (c *Client) postJSON(path string, src interface{}, dst interface{}, requireAuth bool) error {
	var body io.Reader
	if src != nil {
		data, err := json.Marshal(src)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	return c.doWithAuth(req, dst, requireAuth)
}

func (c *Client) patchJSON(path string, src interface{}, dst interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	url := c.baseURL + path
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	return c.doWithAuth(req, dst, true)
}

func (c *Client) delete(path string, requireAuth bool) error {
	url := c.baseURL + path
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	return c.doWithAuth(req, nil, requireAuth)
}

func (c *Client) do(req *http.Request, dst interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	if dst != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) doWithAuth(req *http.Request, dst interface{}, requireAuth bool) error {
	if requireAuth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Handle 401: token expired or invalid
	if resp.StatusCode == http.StatusUnauthorized {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	if resp.StatusCode >= 400 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	if dst != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return err
		}
	}

	return nil
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

// IsUnauthorized checks if an error is a 401 Unauthorized
func IsUnauthorized(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsAuthorizationPending checks if an error is a 428 Precondition Required (authorization pending)
func IsAuthorizationPending(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == 428
	}
	return false
}
