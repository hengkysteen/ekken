package plugins

import (
	"ekken/internal/features/plugins/kind"
	"ekken/internal/logger"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Load loads all plugins from the plugin directory.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Info("Loading plugins", "dir", m.pluginDir)
	if err := m.fs.MkdirAll(m.pluginDir, 0o755); err != nil {
		return err
	}

	// Load state from plugin_state.json
	state := m.loadState()

	// Unregister all existing plugins
	for pluginID, plugin := range m.plugins {
		handler := kind.GetHandler(plugin.manifest.Kind, kind.Config{
			ExecTimeout: m.execTimeout,
		})
		if handler != nil {
			handler.Disable(pluginID)
		}
	}
	m.plugins = make(map[string]*runtimePlugin)

	manifestPaths, err := m.pluginManifestPaths()
	if err != nil {
		return err
	}

	for _, manifestPath := range manifestPaths {
		plugin := m.loadPlugin(manifestPath)

		handler := kind.GetHandler(plugin.manifest.Kind, kind.Config{
			ExecTimeout: m.execTimeout,
		})

		var pluginID string
		if handler != nil && plugin.status == "enabled" {
			kindPlugin := plugin.kindPlugin("")
			if descriptor, err := handler.Inspect(kindPlugin); err == nil {
				pluginID = descriptor.ID
			}
		}

		if pluginID == "" {
			pluginID = filepath.Base(filepath.Dir(manifestPath))
		}

		m.plugins[pluginID] = plugin

		// Hanya registrasi jika statusnya ENABLED di state
		entry, exists := state.Plugins[pluginID]
		isEnabled := !exists || entry.Enabled

		if isEnabled && plugin.status == "enabled" {
			if handler != nil {
				if err := handler.Enable(plugin.kindPlugin(pluginID)); err != nil {
					logger.Error("Failed to enable plugin", "id", pluginID, "error", err)
					plugin.status = "error"
					plugin.reason = fmt.Sprintf("enable failed: %v", err)
					m.plugins[pluginID] = plugin
				} else {
					logger.Info("Plugin registered", "id", pluginID, "kind", plugin.manifest.Kind)
				}
			}
		} else if !isEnabled {
			plugin.status = "disabled"
			m.plugins[pluginID] = plugin
		}

		// Update state jika belum ada
		if !exists {
			if state.Plugins == nil {
				state.Plugins = make(map[string]PluginStateEntry)
			}
			state.Plugins[pluginID] = PluginStateEntry{Enabled: true}
		}
	}

	m.saveState(state)
	return nil
}

func (m *Manager) pluginManifestPaths() ([]string, error) {
	patterns := []string{
		filepath.Join(m.pluginDir, "*", "plugin.json"),
		filepath.Join(m.pluginDir, "*", "*", "plugin.json"),
	}

	seen := make(map[string]bool)
	var manifestPaths []string
	for _, pattern := range patterns {
		matches, err := m.fs.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			if seen[match] {
				continue
			}
			seen[match] = true
			manifestPaths = append(manifestPaths, match)
		}
	}
	sort.Strings(manifestPaths)
	return manifestPaths, nil
}

func (m *Manager) loadState() PluginsState {
	statePath := filepath.Join(m.pluginDir, "plugin_state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return PluginsState{Plugins: make(map[string]PluginStateEntry)}
	}

	var state PluginsState
	if err := json.Unmarshal(data, &state); err != nil {
		return PluginsState{Plugins: make(map[string]PluginStateEntry)}
	}
	if state.Plugins == nil {
		state.Plugins = make(map[string]PluginStateEntry)
	}
	return state
}

func (m *Manager) saveState(state PluginsState) {
	statePath := filepath.Join(m.pluginDir, "plugin_state.json")
	data, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(statePath, data, 0o644)
}

func (m *Manager) loadPlugin(manifestPath string) *runtimePlugin {
	sourcePath := filepath.Dir(manifestPath)
	plugin := &runtimePlugin{
		sourcePath:   sourcePath,
		manifestPath: manifestPath,
		status:       "error",
	}
	raw, err := m.fs.ReadFile(manifestPath)
	if err != nil {
		plugin.manifest = Manifest{}
		plugin.reason = err.Error()
		return plugin
	}

	if err := json.Unmarshal(raw, &plugin.manifest); err != nil {
		plugin.manifest = Manifest{}
		plugin.reason = fmt.Sprintf("invalid manifest JSON: %v", err)
		return plugin
	}

	if plugin.manifest.Source == "" {
		plugin.manifest.Source = SourceLocal
	}

	// Common validation
	if err := validateManifest(plugin.manifest); err != nil {
		plugin.reason = err.Error()
		logger.Error("Plugin validation failed", "path", manifestPath, "error", err)
		return plugin
	}

	// Kind-specific validation
	handler := kind.GetHandler(plugin.manifest.Kind, kind.Config{
		ExecTimeout: m.execTimeout,
	})
	if handler == nil {
		plugin.reason = fmt.Sprintf("unsupported plugin kind: %s", plugin.manifest.Kind)
		return plugin
	}

	if err := handler.Validate(plugin.kindPlugin("")); err != nil {
		plugin.reason = err.Error()
		logger.Error("Plugin validation failed", "path", manifestPath, "error", err)
		return plugin
	}

	plugin.status = "enabled"
	return plugin
}
