package plugins

import (
	"ekken/internal/features/plugins/kind"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// Mock handler for testing
type mockHandler struct{}

func (h *mockHandler) Inspect(plugin kind.Plugin) (kind.Descriptor, error) {
	var spec struct {
		Node struct {
			Type  string `json:"type"`
			Label string `json:"label"`
		} `json:"node"`
	}
	if err := json.Unmarshal(plugin.Spec, &spec); err != nil {
		return kind.Descriptor{}, err
	}
	return kind.Descriptor{ID: spec.Node.Type, Label: spec.Node.Label}, nil
}
func (h *mockHandler) Validate(plugin kind.Plugin) error { return nil }
func (h *mockHandler) Enable(plugin kind.Plugin) error   { return nil }
func (h *mockHandler) Disable(id string) error           { return nil }

type MockFS struct {
	ReadFileFunc func(name string) ([]byte, error)
	GlobFunc     func(pattern string) ([]string, error)
	MkdirAllFunc func(path string, perm os.FileMode) error
	StatFunc     func(name string) (os.FileInfo, error)
}

func (m *MockFS) ReadFile(name string) ([]byte, error)         { return m.ReadFileFunc(name) }
func (m *MockFS) Glob(pattern string) ([]string, error)        { return m.GlobFunc(pattern) }
func (m *MockFS) MkdirAll(path string, perm os.FileMode) error { return m.MkdirAllFunc(path, perm) }
func (m *MockFS) Stat(name string) (os.FileInfo, error)        { return m.StatFunc(name) }

func TestManager_Load(t *testing.T) {
	// Register mock handler for test
	kind.Register("node", func(config kind.Config) kind.Kind {
		return &mockHandler{}
	})

	fs := &MockFS{
		MkdirAllFunc: func(path string, perm os.FileMode) error { return nil },
		GlobFunc: func(pattern string) ([]string, error) {
			return []string{"plugins/p1/plugin.json"}, nil
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			return []byte(`{
				"kind": "node",
				"spec": {
					"runner": {"command": "node"},
					"node": {
						"type": "n1",
						"label": "N1",
						"tags": ["test"],
						"actions": [{
							"key": "run",
							"label": "Run",
							"fields": [{"key": "f1", "type": "string", "label": "Field 1"}]
						}],
						"outputs": [{"key": "out", "label": "Out", "tone": "success"}]
					}
				}
			}`), nil
		},
	}

	m := NewManager("0.1.0", "/tmp/plugins", 10*time.Second)
	m.SetFileSystem(fs)

	err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	list := m.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(list))
	}

	if list[0].ID != "n1" {
		t.Errorf("Expected plugin ID n1, got %s", list[0].ID)
	}
}
