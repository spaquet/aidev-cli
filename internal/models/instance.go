package models

// Instance represents a cloud VM instance
type Instance struct {
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	Status                  string   `json:"status"`
	Tier                    string   `json:"tier"`
	Region                  string   `json:"region"`
	DiskGB                  int      `json:"disk_gb"`
	DiskUsedGB              int      `json:"disk_used_gb"`
	SSHHost                 string   `json:"ssh_host"`
	SSHPort                 int      `json:"ssh_port"`
	SSHUser                 string   `json:"ssh_user"`
	SSHPublicKeyFingerprint string   `json:"ssh_public_key_fingerprint"`
	PublicURLs              []string `json:"public_urls"`
	InstalledTools          []string `json:"installed_tools"`
	ImageVersion            string   `json:"image_version"`
	ImageUpdateAvailable    bool     `json:"image_update_available"`
	CreatedAt               string   `json:"created_at"`
	UpdatedAt               string   `json:"updated_at"`
}

// InstancesResponse wraps a list of instances
type InstancesResponse struct {
	Instances []Instance `json:"instances"`
}

// CreateInstanceRequest is sent to POST /instances
type CreateInstanceRequest struct {
	Name   string `json:"name"`
	Tier   string `json:"tier"`
	Region string `json:"region"`
}

// UpdateInstanceRequest is sent to PATCH /instances/:id
type UpdateInstanceRequest struct {
	Tier      *string `json:"tier,omitempty"`
	DiskGB    *int    `json:"disk_gb,omitempty"`
}

// ImageUpdateResponse is returned from POST /instances/:id/image-update
type ImageUpdateResponse struct {
	Queued                    bool   `json:"queued"`
	EstimatedDurationMinutes  int    `json:"estimated_duration_minutes"`
	EstimatedCompletionAt     string `json:"estimated_completion_at"`
}

// SSEEvent represents a server-sent event
type SSEEvent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
