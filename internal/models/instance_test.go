package models

import (
	"testing"
)

func TestInstance_ZeroValue(t *testing.T) {
	var inst Instance
	if inst.ID != "" {
		t.Errorf("expected zero ID, got %q", inst.ID)
	}
	if inst.Name != "" {
		t.Errorf("expected zero Name, got %q", inst.Name)
	}
	if inst.Status != "" {
		t.Errorf("expected zero Status, got %q", inst.Status)
	}
}

func TestInstance_Fields(t *testing.T) {
	inst := Instance{
		ID:     "test-123",
		Name:   "my-instance",
		Status: "running",
	}

	if inst.ID != "test-123" {
		t.Errorf("expected ID=test-123, got %q", inst.ID)
	}
	if inst.Name != "my-instance" {
		t.Errorf("expected Name=my-instance, got %q", inst.Name)
	}
	if inst.Status != "running" {
		t.Errorf("expected Status=running, got %q", inst.Status)
	}
}
