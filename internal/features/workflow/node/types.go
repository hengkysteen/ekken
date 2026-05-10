package node

import (
	"context"
	"errors"
)

type NodeField struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  any    `json:"default,omitempty"`
	Options  any    `json:"options,omitempty"`
	Label    string `json:"label"`
}

type Form struct {
	Key  string `json:"key"`
	Flex int    `json:"flex"`
	// Supported values: input, number, number-s1, number-s2, select, textarea, radio, slider, switch, jsonEditor, colorPicker, datePicker, timePicker, text.
	Component string `json:"component,omitempty"`
	// Special Options:
	// - native_file_picker (bool)
	// - native_file_picker_multiple (bool),
	// - native_file_picker_directory (bool),
	// - credential_picker (bool),

	// Regular Options : helper (string), placeholder (string), disabled (bool).
	FormOptions any `json:"form_options,omitempty"`
}

type NodeAction struct {
	Key          string            `json:"key"`
	Label        string            `json:"label"`
	Description  string            `json:"description"`
	HasResponse  bool              `json:"has_response"`
	ResponseType *NodeResponseType `json:"response_type,omitempty"`
	Response     string            `json:"response,omitempty"`
	Fields       []NodeField       `json:"fields"`
	Form         [][]Form          `json:"form,omitempty"`
}

type NodeDependency struct {
	Node   string `json:"node"`
	Action string `json:"action"`
}

type NodeMetadata struct {
	Type      string           `json:"type"`
	Label     string           `json:"label,omitempty"`
	Icon      string           `json:"icon,omitempty"`
	Tags      []string         `json:"tags,omitempty"`
	DependsOn []NodeDependency `json:"depends_on,omitempty"`
}

type Node struct {
	NodeMetadata
	ID              string              `json:"id,omitempty"`
	Config          map[string]any      `json:"config"`
	ResponseVar     string              `json:"response_var,omitempty"`
	OnError         string              `json:"on_error,omitempty"`
	ContinueOnError bool                `json:"continue_on_error,omitempty"`
	Nodes           []Node              `json:"nodes,omitempty"`
	Edges           []Edge              `json:"edges,omitempty"`
	Positions       map[string]Position `json:"positions,omitempty"`
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

type NodeOutputDef struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

type NodeSpec struct {
	NodeMetadata
	Kind          string          `json:"kind,omitempty"`
	Parent        string          `json:"parent,omitempty"`
	Description   string          `json:"description"`
	ParentConfig  []NodeField     `json:"parent_config,omitempty"`
	SupportsNodes bool            `json:"supports_nested_nodes"`
	Actions       []NodeAction    `json:"actions"`
	GlobalFields  []NodeField     `json:"global_fields,omitempty"`
	GlobalForm    [][]Form        `json:"global_form,omitempty"`
	DefaultAction string          `json:"default_action,omitempty"`
	Outputs       []NodeOutputDef `json:"outputs,omitempty"`
}

var (
	ErrNodeStopped      = errors.New("Stopped by user")
	ErrWorkflowComplete = errors.New("iteration limit reached")
)

type NodeExecutor interface {
	Execute(ctx *NodeContext) (NodeExecutionResult, error)
}

type NodeExecutorFactory func(config map[string]any, childNodes []Node) NodeExecutor

type NodeRegistration struct {
	NodeSpec
	ExecutorFactory NodeExecutorFactory
}

type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}
