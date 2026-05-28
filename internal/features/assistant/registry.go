package assistant

import (
	"fmt"
	"sync"
)

var (
	registryMU            sync.RWMutex
	providerTypes         = make(map[string]func() IProvider)
	defaultModelsRegistry = make(map[string][]ModelEntry)
	providers             = make(map[string]IProvider)
	providerMeta          = make(map[string]Provider)
)

// Register adds a provider type (factory) and its default models to the global registry
func Register(typeName string, factory func() IProvider, defaultModels []ModelEntry) {
	registryMU.Lock()
	defer registryMU.Unlock()
	providerTypes[typeName] = factory
	defaultModelsRegistry[typeName] = defaultModels
}

// UnregisterProviderType removes a provider type and its default models from the global registry
func UnregisterProviderType(typeName string) {
	registryMU.Lock()
	defer registryMU.Unlock()
	delete(providerTypes, typeName)
	delete(defaultModelsRegistry, typeName)
}


// RegisteredTypes returns all registered provider type names
func RegisteredTypes() []string {
	registryMU.RLock()
	defer registryMU.RUnlock()
	keys := make([]string, 0, len(providerTypes))
	for k := range providerTypes {
		keys = append(keys, k)
	}
	return keys
}

// GetDefaultModels returns the predefined models for a provider type
func GetDefaultModels(typeName string) []ModelEntry {
	registryMU.RLock()
	defer registryMU.RUnlock()
	return defaultModelsRegistry[typeName]
}

// CreateProvider creates a new configured provider instance.
func CreateProvider(providerID string, storedConfig, runtimeConfig map[string]string) error {
	registryMU.Lock()
	defer registryMU.Unlock()

	factory, exists := providerTypes[providerID]
	if !exists {
		return fmt.Errorf("provider type %s not found in registry", providerID)
	}

	p := factory()
	p.Configure(runtimeConfig)

	providers[providerID] = p

	// Store meta data for UI display
	info := p.Info()
	providerMeta[providerID] = Provider{
		ID:          providerID, // ID used for UI list keys
		ProviderID:  providerID, // Reference to the type (groq, etc)
		Name:        info.Name,
		Logo:        info.Logo,
		BaseURL:     info.BaseURL,
		OfficialURL: info.OfficialURL,
		Config:      storedConfig,
	}
	return nil
}

// GetProvider retrieves a provider instance by its ID
func GetProvider(providerID string) (IProvider, error) {
	registryMU.RLock()
	defer registryMU.RUnlock()

	if p, exists := providers[providerID]; exists {
		return p, nil
	}
	return nil, fmt.Errorf("provider %s not found", providerID)
}

// ListProviderTypes returns a list of all registered provider types (The Catalog)
func ListProviderTypes() []ProviderType {
	registryMU.RLock()
	defer registryMU.RUnlock()

	list := make([]ProviderType, 0, len(providerTypes))
	for _, factory := range providerTypes {
		p := factory()
		list = append(list, p.Info())
	}
	return list
}

// ListProviders returns all configured assistant providers
func ListProviders() []Provider {
	registryMU.RLock()
	defer registryMU.RUnlock()

	list := make([]Provider, 0, len(providerMeta))
	for _, info := range providerMeta {
		list = append(list, info)
	}
	return list
}

// RemoveProvider removes a provider from memory
func RemoveProvider(providerID string) {
	registryMU.Lock()
	defer registryMU.Unlock()
	delete(providers, providerID)
	delete(providerMeta, providerID)
}

type ProviderRecord struct {
	ProviderID string
	Config     map[string]string
}
