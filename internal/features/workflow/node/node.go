package node

import (
	"fmt"
	"sort"
	"sync"
)

// NodeProvider defines the interface required by the engine to resolve nodes.
type NodeProvider interface {
	GetExecutor(nodeType string, action Action) NodeExecutor
	GetSpec(nodeType string) (Spec, bool)
}

type Registry struct {
	specMu           sync.RWMutex
	specRegistry     map[string]Spec
	executorMu       sync.RWMutex
	executorRegistry map[string]NodeExecutorFactory
}

func NewRegistry() *Registry {
	return &Registry{
		specRegistry:     make(map[string]Spec),
		executorRegistry: make(map[string]NodeExecutorFactory),
	}
}

// GlobalRegistry is the default registry used by built-in nodes.
var GlobalRegistry = NewRegistry()

func (r *Registry) Register(reg NodeRegistration) {
	spec := reg.Spec

	// Auto-fill ResponseVar for actions that have HasResponse
	for i, action := range spec.Actions {
		if action.HasResponse && action.ResponseVar == "" {
			spec.Actions[i].ResponseVar = fmt.Sprintf("%s.%s_", spec.Type, action.Type)
		}
	}

	r.specMu.Lock()
	r.specRegistry[spec.Type] = spec
	r.specMu.Unlock()

	if reg.ExecutorFactory != nil {
		r.executorMu.Lock()
		r.executorRegistry[spec.Type] = reg.ExecutorFactory
		r.executorMu.Unlock()
	}
}

func (r *Registry) Unregister(nodeType string) {
	r.specMu.Lock()
	delete(r.specRegistry, nodeType)
	r.specMu.Unlock()

	r.executorMu.Lock()
	delete(r.executorRegistry, nodeType)
	r.executorMu.Unlock()
}

func (r *Registry) GetSpec(nodeType string) (Spec, bool) {
	r.specMu.RLock()
	defer r.specMu.RUnlock()
	spec, ok := r.specRegistry[nodeType]
	return spec, ok
}

func (r *Registry) AllSpecs() []Spec {
	r.specMu.RLock()
	defer r.specMu.RUnlock()

	result := make([]Spec, 0, len(r.specRegistry))
	for _, spec := range r.specRegistry {
		result = append(result, spec)
	}

	sort.Slice(result, func(i, j int) bool {
		tagI := ""
		if len(result[i].Tags) > 0 {
			tagI = result[i].Tags[0]
		}
		tagJ := ""
		if len(result[j].Tags) > 0 {
			tagJ = result[j].Tags[0]
		}

		if tagI != tagJ {
			return tagI < tagJ
		}
		if result[i].Label != result[j].Label {
			return result[i].Label < result[j].Label
		}
		return result[i].Type < result[j].Type
	})

	return result
}

func (r *Registry) GetExecutor(nodeType string, action Action) NodeExecutor {
	r.executorMu.RLock()
	factory, ok := r.executorRegistry[nodeType]
	r.executorMu.RUnlock()

	if !ok {
		return nil
	}
	return factory(action)
}
