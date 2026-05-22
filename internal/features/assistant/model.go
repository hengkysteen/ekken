package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func NewModelManager(dataDir string) (*ModelManager, error) {
	path := filepath.Join(dataDir, "models.json")
	m := &ModelManager{
		filePath: path,
	}

	m.setDefaults()

	if err := m.Load(); err != nil {
		slog.Error("Failed to load models.json, using internal defaults", "error", err, "path", path)
	}

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

func (m *ModelManager) SyncFromRegistry(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry error: %d", resp.StatusCode)
	}

	var registryConfig ModelConfig
	if err := json.NewDecoder(resp.Body).Decode(&registryConfig); err != nil {
		return err
	}

	m.mu.Lock()
	m.config.Date = registryConfig.Date
	m.config.System = registryConfig.System
	err = m.saveLocked()
	m.mu.Unlock()

	return err
}
