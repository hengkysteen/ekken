package passistant

import (
	"ekken/internal/features/plugins/kind"
	"encoding/json"
	"fmt"
	"time"
)

type RegistryPort interface {
	Register(providerID string, runner RunnerSpec, provider ProviderSpec, sourcePath string, execTimeout time.Duration) error
	Unregister(providerID string) error
}

var GlobalRegistry RegistryPort

type Handler struct {
	execTimeout time.Duration
}

func NewKind(execTimeout time.Duration) kind.Kind {
	return &Handler{
		execTimeout: execTimeout,
	}
}

func decodePluginSpec(plugin kind.Plugin) (PluginSpec, error) {
	if len(plugin.Spec) == 0 {
		return PluginSpec{}, fmt.Errorf("plugin.spec is required")
	}
	var spec PluginSpec
	if err := json.Unmarshal(plugin.Spec, &spec); err != nil {
		return PluginSpec{}, fmt.Errorf("failed to unmarshal assistant plugin spec: %w", err)
	}
	return spec, nil
}

func (h *Handler) Inspect(plugin kind.Plugin) (kind.Descriptor, error) {
	spec, err := decodePluginSpec(plugin)
	if err != nil {
		return kind.Descriptor{}, err
	}
	if err := h.Validate(plugin); err != nil {
		return kind.Descriptor{}, err
	}
	return kind.Descriptor{
		ID:    spec.Provider.ID,
		Label: spec.Provider.Name,
	}, nil
}

func (h *Handler) Validate(plugin kind.Plugin) error {
	return ValidateManifest(plugin)
}

func (h *Handler) Enable(plugin kind.Plugin) error {
	if err := h.Validate(plugin); err != nil {
		return err
	}
	spec, err := decodePluginSpec(plugin)
	if err != nil {
		return err
	}

	if GlobalRegistry == nil {
		return fmt.Errorf("assistant GlobalRegistry port is not set")
	}

	return GlobalRegistry.Register(spec.Provider.ID, spec.Runner, spec.Provider, plugin.SourcePath, h.execTimeout)
}

func (h *Handler) Disable(id string) error {
	if GlobalRegistry == nil {
		return fmt.Errorf("assistant GlobalRegistry port is not set")
	}
	return GlobalRegistry.Unregister(id)
}

func ValidateManifest(plugin kind.Plugin) error {
	spec, err := decodePluginSpec(plugin)
	if err != nil {
		return err
	}
	if spec.Runner.Command == "" {
		return fmt.Errorf("plugin.spec.runner.command is required")
	}
	if spec.Provider.ID == "" {
		return fmt.Errorf("plugin.spec.provider.id is required")
	}
	if spec.Provider.Name == "" {
		return fmt.Errorf("plugin.spec.provider.name is required")
	}
	for i, model := range spec.Provider.Models {
		if model.Name == "" {
			return fmt.Errorf("plugin.spec.provider.models[%d].name is required", i)
		}
		if model.Origin == "" {
			return fmt.Errorf("plugin.spec.provider.models[%d].origin is required", i)
		}
		if model.ContextWindow <= 0 {
			return fmt.Errorf("plugin.spec.provider.models[%d].context_window must be greater than 0", i)
		}
	}
	return nil
}
