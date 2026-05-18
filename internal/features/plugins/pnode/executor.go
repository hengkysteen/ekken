package pnode

import (
	"bytes"
	"context"
	"ekken/internal/features/plugins/kind"
	"ekken/internal/features/workflow/node"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NewExecutor creates a new node plugin executor.
func NewExecutor(command string, execTimeout time.Duration) *Executor {
	return &Executor{
		command:     command,
		execTimeout: execTimeout,
	}
}

// SetSourcePath sets the plugin source directory.
func (e *Executor) SetSourcePath(path string) {
	e.sourcePath = path
}

// Execute runs the plugin with the given configuration.
func (e *Executor) Execute(manifest interface{}, config map[string]interface{}, data map[string]interface{}) (kind.ExecuteResult, error) {
	return e.ExecuteWithContext(context.Background(), manifest, config, data)
}

// ExecuteWithContext runs the plugin and lets the workflow cancellation stop the process.
func (e *Executor) ExecuteWithContext(ctx context.Context, manifest interface{}, config map[string]interface{}, data map[string]interface{}) (kind.ExecuteResult, error) {
	var nodeType string
	switch m := manifest.(type) {
	case string:
		nodeType = m
	case interface{ GetNodeType() string }:
		nodeType = m.GetNodeType()
	default:
		nodeType = "unknown"
	}

	command := e.resolveCommand()
	if command == "" {
		return kind.ExecuteResult{}, fmt.Errorf("plugin has empty backend command")
	}

	if ctx == nil {
		ctx = context.Background()
	}
	callCtx, cancel := context.WithTimeout(ctx, e.execTimeout)
	defer cancel()

	cmd := exec.CommandContext(callCtx, command)
	cmd.Dir = e.sourcePath

	// Convert generic context to node-specific ExecuteContext
	execCtx := ExecuteContext{}
	if workflowID, ok := data["workflow_id"].(string); ok {
		execCtx.WorkflowID = workflowID
	}
	if iteration, ok := data["iteration"].(int); ok {
		execCtx.Iteration = iteration
	}
	if variables, ok := data["variables"].(map[string]interface{}); ok {
		execCtx.Variables = variables
	}

	req := ExecuteRequest{
		Kind:    "node",
		TypeID:  nodeType,
		Config:  cloneMap(config),
		Context: execCtx,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return kind.ExecuteResult{}, fmt.Errorf("marshal plugin request: %w", err)
	}

	cmd.Stdin = bytes.NewReader(payload)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if callCtx.Err() == context.Canceled {
			return kind.ExecuteResult{}, context.Canceled
		}
		if callCtx.Err() == context.DeadlineExceeded {
			return kind.ExecuteResult{}, fmt.Errorf("plugin timed out")
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return kind.ExecuteResult{}, fmt.Errorf("plugin failed: %s", msg)
	}

	var resp ExecuteResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return kind.ExecuteResult{}, fmt.Errorf("plugin returned invalid JSON: %v", err)
	}

	if resp.Error != "" {
		return kind.ExecuteResult{}, errors.New(resp.Error)
	}

	handle := resp.Handle
	if handle == "" {
		handle = "success"
	}

	return kind.ExecuteResult{
		Handle:   handle,
		Response: resp.Response,
		Metadata: resp.Metadata,
	}, nil
}

func (e *Executor) resolveCommand() string {
	command := e.command
	if strings.HasPrefix(command, ".") {
		command = filepath.Join(e.sourcePath, command)
	}
	return command
}

func cloneMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// PluginProcessExecutor adapts the Executor to node.NodeExecutor interface.
type PluginProcessExecutor struct {
	executor *Executor
	nodeType string
	config   map[string]interface{}
}

// Execute implements node.NodeExecutor.
func (p *PluginProcessExecutor) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	result, err := p.executor.ExecuteWithContext(ctx.Context, p.nodeType, p.config, map[string]interface{}{
		"workflow_id": ctx.WorkflowID,
		"iteration":   ctx.Iteration,
		"variables":   ctx.Variables,
	})

	if err != nil {
		return node.NodeExecutionResult{}, err
	}

	// Extract Type from metadata if present
	var responseType *node.NodeResponseType
	if result.Metadata != nil {
		if typeVal, ok := result.Metadata["type"]; ok {
			if typeData, err := json.Marshal(typeVal); err == nil {
				var nt node.NodeResponseType
				if json.Unmarshal(typeData, &nt) == nil {
					responseType = &nt
				}
			}
		}
	}

	return node.NodeExecutionResult{
		Handle:   result.Handle,
		Response: result.Response,
		Type:     responseType,
	}, nil
}
