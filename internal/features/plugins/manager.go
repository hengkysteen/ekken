package plugins

import (
	"ekken/internal/features/plugins/kind"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Manager manages plugin lifecycle and execution.
type Manager struct {
	appVersion  string
	pluginDir   string
	execTimeout time.Duration
	fs          FileSystem
	mu          sync.RWMutex
	plugins     map[string]*runtimePlugin
}

// NewManager creates a new plugin manager.
func NewManager(appVersion, pluginDir string, execTimeout time.Duration) *Manager {
	if execTimeout <= 0 {
		execTimeout = 60 * time.Second
	}

	return &Manager{
		appVersion:  normalizeVersion(appVersion),
		pluginDir:   pluginDir,
		execTimeout: execTimeout,
		fs:          osFS{},
		plugins:     make(map[string]*runtimePlugin),
	}
}

// SetFileSystem sets a custom filesystem implementation (for testing).
func (m *Manager) SetFileSystem(fs FileSystem) {
	m.fs = fs
}

// List returns summaries of all loaded plugins.
func (m *Manager) List() []PluginList {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]PluginList, 0, len(m.plugins))
	for id, plugin := range m.plugins {
		var iconStr string
		if len(plugin.manifest.Spec) > 0 {
			var spec struct {
				Node struct {
					Icon string `json:"icon"`
				} `json:"node"`
				Provider struct {
					Icon string `json:"icon"`
				} `json:"provider"`
			}
			_ = json.Unmarshal(plugin.manifest.Spec, &spec)

			if spec.Node.Icon != "" {
				iconStr = spec.Node.Icon
			} else if spec.Provider.Icon != "" {
				iconStr = spec.Provider.Icon
			}
		}

		summaries = append(summaries, PluginList{
			ID:          id,
			Icon:        iconStr,
			Manifest:    plugin.manifest,
			SourcePath:  plugin.sourcePath,
			Status:      plugin.status,
			Reason:      plugin.reason,
			IsInstalled: true,
			IsEnabled:   plugin.status == "enabled",
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].ID < summaries[j].ID
	})

	return summaries
}

// Manage enables or disables a plugin.
func (m *Manager) Manage(id string, action string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("plugin not found: %s", id)
	}

	state := m.loadState()
	if state.Plugins == nil {
		state.Plugins = make(map[string]PluginStateEntry)
	}

	switch action {
	case "enable":
		handler := kind.GetHandler(plugin.manifest.Kind, kind.Config{
			ExecTimeout: m.execTimeout,
		})
		if handler == nil {
			return fmt.Errorf("unsupported plugin kind: %s", plugin.manifest.Kind)
		}

		kindPlugin := plugin.kindPlugin(id)
		if err := handler.Validate(kindPlugin); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		if err := handler.Enable(kindPlugin); err != nil {
			return fmt.Errorf("failed to enable plugin: %w", err)
		}

		state.Plugins[id] = PluginStateEntry{Enabled: true}
		plugin.status = "enabled"
		plugin.reason = ""
		m.plugins[id] = plugin

	case "disable":
		handler := kind.GetHandler(plugin.manifest.Kind, kind.Config{
			ExecTimeout: m.execTimeout,
		})
		if handler != nil {
			handler.Disable(id)
		}

		state.Plugins[id] = PluginStateEntry{Enabled: false}
		plugin.status = "disabled"
		m.plugins[id] = plugin

	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	m.saveState(state)
	return nil
}

func (m *Manager) Uninstall(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pluginID, plugin, exists := m.findPluginLocked(id)
	if !exists {
		return fmt.Errorf("plugin not found: %s", id)
	}

	handler := kind.GetHandler(plugin.manifest.Kind, kind.Config{
		ExecTimeout: m.execTimeout,
	})
	if handler != nil {
		if err := handler.Disable(pluginID); err != nil {
			return fmt.Errorf("failed to disable plugin: %w", err)
		}
	}

	if err := os.RemoveAll(plugin.sourcePath); err != nil {
		return err
	}

	state := m.loadState()
	delete(state.Plugins, pluginID)
	delete(m.plugins, pluginID)
	m.saveState(state)
	return nil
}

func (m *Manager) findPluginLocked(id string) (string, *runtimePlugin, bool) {
	if plugin, exists := m.plugins[id]; exists {
		return id, plugin, true
	}
	for pluginID, plugin := range m.plugins {
		if filepath.Base(plugin.sourcePath) == id {
			return pluginID, plugin, true
		}
	}
	return "", nil, false
}
