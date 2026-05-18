package plugins

import (
	"ekken/internal/features/plugins/kind"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestManager_Manage_EnableDisable(t *testing.T) {
	// Register mock handler
	kind.Register("node", func(config kind.Config) kind.Kind {
		return &mockHandler{}
	})

	fs := &MockFS{
		MkdirAllFunc: func(path string, perm os.FileMode) error { return nil },
		GlobFunc: func(pattern string) ([]string, error) {
			return []string{"plugins/test/plugin.json"}, nil
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			if name == "plugins/test/plugin.json" {
				return []byte(`{
					"kind": "node",
					"spec": {
						"runner": {"command": "node"},
						"node": {
							"type": "test_node",
							"label": "Test",
							"tags": ["test"],
							"actions": [{"key": "run", "label": "Run", "fields": []}],
							"outputs": []
						}
					}
				}`), nil
			}
			// plugin_state.json
			return []byte(`{"plugins":{}}`), nil
		},
	}

	m := NewManager("0.1.0", "/tmp/plugins", 10*time.Second)
	m.SetFileSystem(fs)

	// Load plugins
	if err := m.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Test 1: Disable plugin
	if err := m.Manage("test_node", "disable"); err != nil {
		t.Errorf("Disable failed: %v", err)
	}

	list := m.List()
	if len(list) != 1 {
		t.Fatalf("Expected 1 plugin, got %d", len(list))
	}
	if list[0].Status != "disabled" {
		t.Errorf("Expected status disabled, got %s", list[0].Status)
	}

	// Test 2: Enable plugin
	if err := m.Manage("test_node", "enable"); err != nil {
		t.Errorf("Enable failed: %v", err)
	}

	list = m.List()
	if list[0].Status != "enabled" {
		t.Errorf("Expected status enabled, got %s", list[0].Status)
	}

	// Test 3: Invalid action
	if err := m.Manage("test_node", "invalid"); err == nil {
		t.Error("Expected error for invalid action")
	}

	// Test 4: Non-existent plugin
	if err := m.Manage("nonexistent", "enable"); err == nil {
		t.Error("Expected error for non-existent plugin")
	}
}

func TestManager_Manage_EnableInvalidPlugin(t *testing.T) {
	// Register mock handler that fails validation
	kind.Register("bad", func(config kind.Config) kind.Kind {
		return &failingMockHandler{}
	})

	fs := &MockFS{
		MkdirAllFunc: func(path string, perm os.FileMode) error { return nil },
		GlobFunc: func(pattern string) ([]string, error) {
			return []string{"plugins/bad/plugin.json"}, nil
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			if name == "plugins/bad/plugin.json" {
				return []byte(`{"kind": "bad", "spec": {}}`), nil
			}
			return []byte(`{"plugins":{"bad_plugin":{"enabled":false}}}`), nil
		},
	}

	m := NewManager("0.1.0", "/tmp/plugins", 10*time.Second)
	m.SetFileSystem(fs)
	m.Load()

	// Plugin should be in map but disabled
	list := m.List()
	if len(list) != 1 {
		t.Fatalf("Expected 1 plugin in list, got %d", len(list))
	}

	pluginID := list[0].ID
	if list[0].Status == "enabled" {
		t.Error("Invalid plugin should not be enabled")
	}

	// Try to enable invalid plugin
	err := m.Manage(pluginID, "enable")
	if err == nil {
		t.Error("Expected error when enabling invalid plugin")
	}
}

// failingMockHandler always fails validation
type failingMockHandler struct{}

func (h *failingMockHandler) Inspect(plugin kind.Plugin) (kind.Descriptor, error) {
	return kind.Descriptor{ID: "bad_plugin"}, nil
}
func (h *failingMockHandler) Validate(plugin kind.Plugin) error {
	return fmt.Errorf("invalid spec")
}
func (h *failingMockHandler) Enable(plugin kind.Plugin) error { return nil }
func (h *failingMockHandler) Disable(id string) error         { return nil }
