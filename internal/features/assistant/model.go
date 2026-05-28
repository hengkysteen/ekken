package assistant

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	globalModelManager   *ModelManager
	globalModelManagerMu sync.RWMutex
)

func GetGlobalModelManager() *ModelManager {
	globalModelManagerMu.RLock()
	defer globalModelManagerMu.RUnlock()
	return globalModelManager
}

func setGlobalModelManager(m *ModelManager) {
	globalModelManagerMu.Lock()
	defer globalModelManagerMu.Unlock()
	globalModelManager = m
}

func NewModelManager(dataDir string) (*ModelManager, error) {
	path := filepath.Join(dataDir, "models.json")
	m := &ModelManager{
		filePath: path,
	}

	m.setDefaults()

	if err := m.Load(); err != nil {
		slog.Error("Failed to load models.json, using internal defaults", "error", err, "path", path)
	}

	setGlobalModelManager(m)

	return m, nil
}

func (m *ModelManager) setDefaults() {
	m.config.Date = "2026-04-19"
	m.config.System = []ProviderModels{}

	// Loop through all registered provider types and pull their models
	for _, typeName := range RegisteredTypes() {
		models := GetDefaultModels(typeName)
		if len(models) > 0 {
			pModels := ProviderModels{
				Provider: typeName,
				Models:   make([]ModelEntry, len(models)),
			}
			for i, mod := range models {
				pModels.Models[i] = ModelEntry(mod)
			}
			m.config.System = append(m.config.System, pModels)
		}
	}

	m.config.User = make([]ProviderModels, 0)
}

func (m *ModelManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := os.Stat(m.filePath); err == nil {
		bytes, err := os.ReadFile(m.filePath)
		if err != nil {
			return err
		}
		var loadedConfig ModelConfig
		if err := json.Unmarshal(bytes, &loadedConfig); err != nil {
			return fmt.Errorf("json parse failed: %w", err)
		}
		m.config = loadedConfig
		return nil
	}

	return m.saveLocked()
}

func (m *ModelManager) saveLocked() error {
	bytes, err := json.MarshalIndent(m.config, "", "    ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(m.filePath, bytes, 0644)
}

func (m *ModelManager) GetModels(providerID string) []ModelInfo {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []ModelInfo
	seen := make(map[string]bool)

	// 1. User models first
	for _, p := range m.config.User {
		if p.Provider == providerID {
			for _, mod := range p.Models {
				if !seen[mod.Origin] {
					result = append(result, ModelInfo{
						Provider:      providerID,
						Model:         mod.Origin,
						Name:          mod.Name,
						ContextWindow: mod.ContextWindow,
						Type:          "llm",
					})
					seen[mod.Origin] = true
				}
			}
		}
	}

	// 2. System models
	for _, p := range m.config.System {
		if p.Provider == providerID {
			for _, mod := range p.Models {
				if !seen[mod.Origin] {
					result = append(result, ModelInfo{
						Provider:      providerID,
						Model:         mod.Origin,
						Name:          mod.Name,
						ContextWindow: mod.ContextWindow,
						Type:          "llm",
					})
					seen[mod.Origin] = true
				}
			}
		}
	}

	return result
}

func (m *ModelManager) RegisterPluginModels(providerID string, models []ModelEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	foundIndex := -1
	for i, p := range m.config.System {
		if p.Provider == providerID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		m.config.System = append(m.config.System, ProviderModels{
			Provider: providerID,
			Models:   models,
		})
		return m.saveLocked()
	}

	existingModels := m.config.System[foundIndex].Models
	existingMap := make(map[string]bool)
	for _, model := range existingModels {
		existingMap[model.Origin] = true
	}

	appended := false
	for _, newModel := range models {
		if !existingMap[newModel.Origin] {
			m.config.System[foundIndex].Models = append(m.config.System[foundIndex].Models, newModel)
			appended = true
		}
	}

	if appended {
		return m.saveLocked()
	}

	return nil
}

func (m *ModelManager) SyncWithEmbeddedDefaults(force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var defaults []ProviderModels
	for _, typeName := range RegisteredTypes() {
		models := GetDefaultModels(typeName)
		if len(models) > 0 {
			pModels := ProviderModels{
				Provider: typeName,
				Models:   make([]ModelEntry, len(models)),
			}
			for i, mod := range models {
				pModels.Models[i] = ModelEntry(mod)
			}
			defaults = append(defaults, pModels)
		}
	}

	if force {
		m.config.System = defaults
	} else {
		for _, defProv := range defaults {
			foundIndex := -1
			for i, sysProv := range m.config.System {
				if sysProv.Provider == defProv.Provider {
					foundIndex = i
					break
				}
			}

			if foundIndex == -1 {
				m.config.System = append(m.config.System, defProv)
			} else {
				existingModels := m.config.System[foundIndex].Models
				existingMap := make(map[string]bool)
				for _, model := range existingModels {
					existingMap[model.Origin] = true
				}

				for _, defModel := range defProv.Models {
					if !existingMap[defModel.Origin] {
						m.config.System[foundIndex].Models = append(m.config.System[foundIndex].Models, defModel)
					}
				}
			}
		}
	}

	m.config.Date = time.Now().Format("2006-01-02")
	return m.saveLocked()
}
