package pnode

import (
	"ekken/internal/features/workflow/node"
	"encoding/json"
	"fmt"
	"time"
)

type RunnerSpec struct {
	Type    string `json:"type,omitempty"`
	Command string `json:"command"`
}

func (r *RunnerSpec) UnmarshalJSON(data []byte) error {
	var command string
	if err := json.Unmarshal(data, &command); err == nil {
		r.Command = command
		return nil
	}

	type runnerSpec RunnerSpec
	var spec runnerSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("invalid runner spec: %w", err)
	}
	*r = RunnerSpec(spec)
	return nil
}

type PluginSpec struct {
	Runner RunnerSpec `json:"runner"`
	Node   node.Spec  `json:"node"`
}

// ExecuteRequest is the request sent to node plugin binaries
type ExecuteRequest struct {
	Kind    string                 `json:"kind"`
	TypeID  string                 `json:"type_id"`
	Config  map[string]interface{} `json:"config"`
	Context ExecuteContext         `json:"context"`
}

// ExecuteContext contains workflow execution context for nodes
type ExecuteContext struct {
	WorkflowID string                 `json:"workflow_id"`
	Iteration  int                    `json:"iteration"`
	Variables  map[string]interface{} `json:"variables"`
}

// ExecuteResponse is the response from node plugin binaries
type ExecuteResponse struct {
	Handle   string                 `json:"handle"`
	Response interface{}            `json:"response"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

// Executor handles execution of node plugins via external process.
type Executor struct {
	sourcePath  string
	command     string
	execTimeout time.Duration
}

// Handler implements kind.Kind for node plugins.
type Handler struct {
	execTimeout time.Duration
	executors   map[string]*Executor // pluginID -> executor
}
