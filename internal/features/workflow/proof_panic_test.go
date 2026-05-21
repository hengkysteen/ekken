package workflow

import (
	"context"
	"ekken/internal/features/workflow/node"
	"testing"
)

// TestRunner_NoPanicRecovery membuktikan bahwa jika ada node yang panic,
// seluruh runtime akan crash (test akan fail dengan exit code non-zero).
func TestRunner_NoPanicRecovery(t *testing.T) {
	// 1. Buat node executor yang sengaja memicu panic
	panicExecutor := &MockExecutor{
		ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
			// Sengaja trigger panic (misal: simulasi nil pointer di chromedp)
			panic("Simulasi fatal error dari node eksternal!")
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

	// 2. Jalankan workflow.
	// Jika sistem memiliki Panic Recovery (defer recover), fungsi ini akan return error.
	// Jika tidak, test runner Go ini akan CRASH dan berhenti total.
	err := eng.Run(context.Background(), wf)

	// Jika kode sampai ke baris ini, berarti panic berhasil ditangani (recovered).
	if err == nil {
		t.Fatal("Expected an error from panic recovery, got nil")
	}
	t.Log("Panic berhasil di-recover, test lewat:", err)
}
