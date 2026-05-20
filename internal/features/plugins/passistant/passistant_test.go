package passistant

import (
	"ekken/internal/features/plugins/kind"
	"testing"
	"time"
)

type mockRegistry struct {
	registeredID   string
	unregisteredID string
	err            error
}

func (m *mockRegistry) Register(providerID string, runner RunnerSpec, provider ProviderSpec, sourcePath string, execTimeout time.Duration) error {
	m.registeredID = providerID
	return m.err
}

func (m *mockRegistry) Unregister(providerID string) error {
	m.unregisteredID = providerID
	return m.err
}

func TestHandler_Lifecycle(t *testing.T) {
	validSpec := []byte(`{
		"runner": {"command": "echo"},
		"provider": {
			"id": "mock-prov",
			"name": "Mock Provider",
			"models": [
				{"name": "Model A", "origin": "a", "context_window": 1000}
			]
		}
	}`)

	handler := NewKind(1 * time.Second)

	// Test Validate with empty spec
	err := handler.Validate(kind.Plugin{Spec: nil})
	if err == nil {
		t.Error("expected error for empty spec, got nil")
	}

	// Test Validate with invalid JSON
	err = handler.Validate(kind.Plugin{Spec: []byte("{invalid")})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}

	// Test Validate with valid spec
	err = handler.Validate(kind.Plugin{Spec: validSpec})
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Test Inspect
	desc, err := handler.Inspect(kind.Plugin{Spec: validSpec})
	if err != nil {
		t.Fatalf("unexpected Inspect error: %v", err)
	}
	if desc.ID != "mock-prov" || desc.Label != "Mock Provider" {
		t.Errorf("unexpected descriptor: %+v", desc)
	}

	// Test Enable and Disable with mock registry
	reg := &mockRegistry{}
	GlobalRegistry = reg

	err = handler.Enable(kind.Plugin{Spec: validSpec, ID: "plugin-1"})
	if err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}
	if reg.registeredID != "mock-prov" {
		t.Errorf("expected registered ID 'mock-prov', got %q", reg.registeredID)
	}

	err = handler.Disable("plugin-1")
	if err != nil {
		t.Fatalf("Disable returned error: %v", err)
	}
	if reg.unregisteredID != "plugin-1" {
		t.Errorf("expected unregistered ID 'plugin-1', got %q", reg.unregisteredID)
	}
}

func TestValidateManifest_Failures(t *testing.T) {
	cases := []struct {
		name string
		spec string
	}{
		{"missing command", `{"provider": {"id": "p", "name": "n"}}`},
		{"missing id", `{"runner": {"command": "cmd"}, "provider": {"name": "n"}}`},
		{"missing name", `{"runner": {"command": "cmd"}, "provider": {"id": "p"}}`},
		{"invalid model name", `{"runner": {"command": "cmd"}, "provider": {"id": "p", "name": "n", "models": [{"origin": "o", "context_window": 1}]}}`},
		{"invalid model origin", `{"runner": {"command": "cmd"}, "provider": {"id": "p", "name": "n", "models": [{"name": "n", "context_window": 1}]}}`},
		{"invalid context window", `{"runner": {"command": "cmd"}, "provider": {"id": "p", "name": "n", "models": [{"name": "n", "origin": "o", "context_window": 0}]}}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateManifest(kind.Plugin{Spec: []byte(tc.spec)})
			if err == nil {
				t.Error("expected validation failure, got success")
			}
		})
	}
}
