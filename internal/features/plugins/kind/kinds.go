package kind

import (
	"encoding/json"
	"sync"
	"time"
)

// Plugin is the kind-facing plugin envelope. Spec is opaque to the plugins
// manager and is decoded only by the selected kind handler.
type Plugin struct {
	ID           string
	Kind         string
	Spec         json.RawMessage
	SourcePath   string
	ManifestPath string
}

// Descriptor is the stable identity returned by a kind after inspecting spec.
type Descriptor struct {
	ID    string
	Label string
}

// Kind defines the interface for plugin kind-specific lifecycle operations.
type Kind interface {
	// Inspect returns the stable plugin identity for this kind.
	Inspect(plugin Plugin) (Descriptor, error)

	// Validate validates the plugin for this kind.
	Validate(plugin Plugin) error

	// Enable activates the plugin in the kind-owned runtime or registry.
	Enable(plugin Plugin) error

	// Disable deactivates the plugin from the kind-owned runtime or registry.
	Disable(id string) error
}

// PluginExecutor defines the interface for executing plugin operations.
type Executor interface {
	// manifest and context are kind-specific execution inputs.
	Execute(manifest any, config map[string]any, context map[string]any) (ExecuteResult, error)
}

// ExecuteResult represents the result of plugin execution.
type ExecuteResult struct {
	Handle   string         `json:"handle"`
	Response any            `json:"response"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Factory creates a Kind with given config.
type Factory func(config Config) Kind

// Config provides configuration for kind handler creation.
type Config struct {
	ExecTimeout time.Duration
}

// registry manages registered kind handlers.
type registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// globalRegistry is the global kind registry.
var globalRegistry = &registry{
	factories: make(map[string]Factory),
}

// Register registers a kind handler factory.
// Called by kind packages via init() for self-registration.
func Register(name string, factory Factory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[name] = factory
}

// GetHandler returns a kind handler for the given kind name.
// Returns nil if kind is not registered.
func GetHandler(name string, config Config) Kind {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	if factory, ok := globalRegistry.factories[name]; ok {
		return factory(config)
	}
	return nil
}
