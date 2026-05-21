package pnode

import (
	"ekken/internal/features/plugins/kind"
	"ekken/internal/features/workflow/node"
	"time"
)

// NewKind creates a new node kind handler (for registration).
func NewKind(execTimeout time.Duration) kind.Kind {
	return NewHandler(execTimeout)
}

// NewHandler creates a new node kind handler.
func NewHandler(execTimeout time.Duration) *Handler {
	return &Handler{
		execTimeout: execTimeout,
		executors:   make(map[string]*Executor),
	}
}

// Inspect returns the workflow node type represented by this plugin.
func (h *Handler) Inspect(plugin kind.Plugin) (kind.Descriptor, error) {
	spec, err := decodePluginSpec(plugin)
	if err != nil {
		return kind.Descriptor{}, err
	}
	if err := validateNodeSpec(spec.Node); err != nil {
		return kind.Descriptor{}, err
	}
	return kind.Descriptor{
		ID:    spec.Node.Type,
		Label: spec.Node.Label,
	}, nil
}

// Validate validates a node plugin manifest.
func (h *Handler) Validate(plugin kind.Plugin) error {
	return ValidateManifest(plugin)
}

// Enable registers a node plugin with the global node registry.
func (h *Handler) Enable(plugin kind.Plugin) error {
	if err := h.Validate(plugin); err != nil {
		return err
	}

	spec, err := decodePluginSpec(plugin)
	if err != nil {
		return err
	}

	executor := NewExecutor(spec.Runner.Command, h.execTimeout)
	executor.SetSourcePath(plugin.SourcePath)
	h.executors[spec.Node.Type] = executor

	node.GlobalRegistry.Register(node.NodeRegistration{
		Spec:            spec.Node,
		ExecutorFactory: h.createExecutorFactory(spec.Node.Type, executor),
	})

	return nil
}

// Disable removes a node plugin from the global registry.
func (h *Handler) Disable(id string) error {
	node.GlobalRegistry.Unregister(id)
	delete(h.executors, id)
	return nil
}

// createExecutorFactory creates a factory function for node executors.
func (h *Handler) createExecutorFactory(nodeType string, executor kind.Executor) node.NodeExecutorFactory {
	return func(action node.Action) node.NodeExecutor {
		config := make(map[string]interface{})
		for _, f := range action.Fields {
			val := f.Value
			if val == nil {
				val = f.Default
			}
			config[f.Key] = val
		}
		config["action"] = action.Label

		return &PluginProcessExecutor{
			executor: executor.(*Executor),
			nodeType: nodeType,
			config:   config,
		}
	}
}
