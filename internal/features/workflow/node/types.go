package node

import (
	"context"
	"errors"
)

type Meta struct {
	Type        string      `json:"type"`
	Version     string      `json:"version,omitempty"`
	Platforms   []string    `json:"platforms,omitempty"`
	Label       string      `json:"label,omitempty"`
	Description string      `json:"description,omitempty"`
	Icon        string      `json:"icon,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	DependsOn   []DependsOn `json:"depends_on,omitempty"`
}
type Spec struct {
	Meta
	Actions         []Action    `json:"actions"`
	DefaultAction   string      `json:"default_action,omitempty"`
	GlobalFields    []NodeField `json:"global_fields,omitempty"`
	OutputHandles   []string    `json:"output_handles"`
	HideInputHandle bool        `json:"hide_input_handles"`
}

type NodeField struct {
	Key      string `json:"key"`
	Type     string `json:"type,omitempty"`
	Required bool   `json:"required,omitempty"`
	Default  any    `json:"default,omitempty"`
	Options  any    `json:"options,omitempty"`
	Label    string `json:"label,omitempty"`
	Value    any    `json:"value,omitempty"`
}
type AutoLayout struct {
	Key  string `json:"key"`
	Flex int    `json:"flex"`
	// Supported Component: input, number, number-s1, number-s2, select, textarea, radio, slider, switch, jsonEditor, colorPicker, datePicker, timePicker, text.
	Component string `json:"component,omitempty"`
	// Special Options:
	// - native_file_picker (bool)
	// - native_file_picker_multiple (bool),
	// - native_file_picker_directory (bool),
	// - credential_picker (bool),
	// Regular Options : helper (string), placeholder (string), disabled (bool).
	Options any `json:"options,omitempty"`
}
type Action struct {
	Type         string            `json:"type"`
	Label        string            `json:"label,omitempty"`
	Description  string            `json:"description,omitempty"`
	HasResponse  bool              `json:"has_response,omitempty"`
	ResponseType *NodeResponseType `json:"response_type,omitempty"`
	ResponseVar  string            `json:"response_var,omitempty"`
	Fields       []NodeField       `json:"fields"`
	AutoLayout   [][]AutoLayout    `json:"auto_layout,omitempty"`
}
type DependsOn struct {
	// Node Type
	Node string `json:"node"`
	// Action Type
	Action string `json:"action"`
}
type Node struct {
	Meta
	ID              string `json:"id,omitempty"`
	Action          Action `json:"action"`
	OnError         string `json:"on_error,omitempty"`
	ContinueOnError bool   `json:"continue_on_error,omitempty"`
}
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
type Edge struct {
	Source       string `json:"source"`
	SourceHandle string `json:"sourceHandle"`
	Target       string `json:"target"`
}
type NodeContext struct {
	Stop              <-chan struct{}
	Context           context.Context
	Variables         map[string]any
	InternalVariables map[string]any
	WorkflowID        string
	Iteration         int
	OutputHandle      string
	Metadata          map[string]any
	OnCleanup         []func()
	IsLooping         bool
}
type RunnerContext struct {
	*NodeContext
}

func NewRunnerContext(nc *NodeContext) *RunnerContext {
	return &RunnerContext{NodeContext: nc}
}

type NodeResponseType struct {
	Mime     string `json:"mime,omitempty"`
	Charset  string `json:"charset,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}
type NodeExecutionResult struct {
	Handle   string            `json:"handle"`
	Response any               `json:"response"`
	Type     *NodeResponseType `json:"type,omitempty"`
}

var (
	ErrNodeStopped      = errors.New("Stopped by user")
	ErrWorkflowComplete = errors.New("iteration limit reached")
)

type NodeExecutor interface {
	Execute(ctx *NodeContext) (NodeExecutionResult, error)
}
type NodeExecutorFactory func(action Action) NodeExecutor

type NodeRegistration struct {
	Spec
	ExecutorFactory NodeExecutorFactory
}

type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}
