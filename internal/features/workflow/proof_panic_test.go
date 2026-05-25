package workflow

import (
	"context"
	"ekken/internal/features/workflow/node"
	"testing"
)

// TestRunner_NoPanicRecovery verifies that node panics are recovered and
// returned as regular execution errors.
func TestRunner_NoPanicRecovery(t *testing.T) {
	// 1. Create a node executor that deliberately panics.
	panicExecutor := &MockExecutor{
		ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
			// Deliberately trigger a panic, such as a simulated nil pointer in chromedp.
			panic("simulated fatal error from external node")
		},
	}

	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"panic_node": panicExecutor,
		},
	}
	obs := &MockObserver{}
	eng := New(obs, reg)

	wf := Workflow{
		ID: "wf-panic",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "panic_node"}, ID: "n1"},
		},
	}

	// 2. Run the workflow.
	// With panic recovery, this returns an error instead of crashing the test runner.
	err := eng.Run(context.Background(), wf)

	// Reaching this line means the panic was recovered.
	if err == nil {
		t.Fatal("Expected an error from panic recovery, got nil")
	}
	t.Log("Panic recovered successfully:", err)
}
