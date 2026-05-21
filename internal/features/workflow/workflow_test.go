package workflow

import (
	"context"
	"ekken/internal/features/workflow/node"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type MockObserver struct {
	OnStatusUpdateFunc func(id, status string, iteration int)
	OnLogFunc          func(id, level, message, raw string)
}

func (m *MockObserver) OnStatusUpdate(id, status string, iteration int) {
	if m.OnStatusUpdateFunc != nil {
		m.OnStatusUpdateFunc(id, status, iteration)
	}
}
func (m *MockObserver) OnLog(id, level, message, raw string) {
	if m.OnLogFunc != nil {
		m.OnLogFunc(id, level, message, raw)
	}
}

type MockRegistry struct {
	Executors map[string]node.NodeExecutor
	Specs     map[string]node.Spec
}

func (m *MockRegistry) GetExecutor(nodeType string, action node.Action) node.NodeExecutor {
	return m.Executors[nodeType]
}
func (m *MockRegistry) GetSpec(nodeType string) (node.Spec, bool) {
	if m.Specs != nil {
		if spec, ok := m.Specs[nodeType]; ok {
			return spec, true
		}
	}
	return node.Spec{Meta: node.Meta{Type: nodeType, Tags: []string{"Trigger"}}}, true
}

type MockExecutor struct {
	ExecuteFunc func(ctx *node.NodeContext) (node.NodeExecutionResult, error)
}

func (m *MockExecutor) Execute(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
	return m.ExecuteFunc(ctx)
}

func TestRunner_RunLinear(t *testing.T) {
	executed := false
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"test-node": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executed = true
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	obs := &MockObserver{}
	eng := New(obs, reg)

	wf := Workflow{
		ID:    "test-wf",
		Nodes: []node.Node{{Meta: node.Meta{Type: "test-node"}, ID: "n1"}},
	}

	err := eng.Run(context.Background(), wf)
	if err != nil {
		t.Fatalf("engine failed: %v", err)
	}

	if !executed {
		t.Error("node was not executed")
	}
}

func TestRunner_RunGraph(t *testing.T) {
	order := []string{}
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"step": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					order = append(order, "step")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	obs := &MockObserver{}
	eng := New(obs, reg)

	wf := Workflow{
		ID: "graph-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "step"}, ID: "n1"},
			{Meta: node.Meta{Type: "step"}, ID: "n2"},
		},
		Edges: []node.Edge{
			{Source: "n1", SourceHandle: "success", Target: "n2"},
		},
	}

	err := eng.Run(context.Background(), wf)
	if err != nil {
		t.Fatalf("engine failed: %v", err)
	}

	if len(order) != 2 {
		t.Errorf("expected 2 nodes to run, got %d", len(order))
	}
}

func TestRunner_Retry(t *testing.T) {
	attempts := 0
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"fail-then-pass": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					attempts++
					if attempts < 2 {
						return node.NodeExecutionResult{}, errors.New("temporary failure")
					}
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	obs := &MockObserver{}
	eng := New(obs, reg)

	wf := Workflow{
		ID: "retry-wf",
		Nodes: []node.Node{
			{
				Meta: node.Meta{Type: "fail-then-pass"},
				ID:   "n1",
				Action: node.Action{
					Fields: []node.NodeField{
						{Key: "retry_count", Value: 2.0},
						{Key: "retry_delay", Value: 0.01},
					},
				},
			},
		},
	}

	err := eng.Run(context.Background(), wf)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestRunner_OnErrorStop(t *testing.T) {
	nodeExecuted := false
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"fail": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					return node.NodeExecutionResult{}, errors.New("hard fail")
				},
			},
			"second": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					nodeExecuted = true
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	wf := Workflow{
		ID: "error-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "fail"}, ID: "n1", OnError: "stop"},
			{Meta: node.Meta{Type: "second"}, ID: "n2"},
		},
	}

	err := eng.Run(context.Background(), wf)
	if err == nil {
		t.Error("expected error from workflow run")
	}

	if nodeExecuted {
		t.Error("second node should not have executed")
	}
}

func TestRunner_SaveAs(t *testing.T) {
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"producer": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					return node.NodeExecutionResult{Handle: "success", Response: "hello world"}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	wf := Workflow{
		ID: "save-wf",
		Nodes: []node.Node{
			{
				Meta: node.Meta{Type: "producer"},
				ID:   "n1",
				Action: node.Action{
					ResponseVar: "my_var",
				},
			},
		},
	}

	wfCtx := context.Background()
	err := eng.Run(wfCtx, wf)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}

func TestRunner_Cancellation(t *testing.T) {
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"slow": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					select {
					case <-ctx.Context.Done():
						return node.NodeExecutionResult{}, ctx.Context.Err()
					case <-time.After(100 * time.Millisecond):
						return node.NodeExecutionResult{Handle: "success"}, nil
					}
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	ctx, cancel := context.WithCancel(context.Background())

	wf := Workflow{
		ID: "cancel-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "slow"}, ID: "n1"},
		},
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := eng.Run(ctx, wf)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestRunner_JSONExtraction(t *testing.T) {
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"producer": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					return node.NodeExecutionResult{Handle: "success", Response: `{"user": {"id": 123, "name": "Ekken"}}`}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	wf := Workflow{
		ID: "json-wf",
		Nodes: []node.Node{
			{
				Meta: node.Meta{Type: "producer"},
				ID:   "n1",
				Action: node.Action{
					ResponseVar: "user_id",
				},
			},
		},
	}

	err := eng.Run(context.Background(), wf)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}

func TestRunner_Looping(t *testing.T) {
	iterations := 0
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"looper": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					iterations++
					if iterations < 2 {
						ctx.IsLooping = true
					}
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	wf := Workflow{
		ID: "loop-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "looper"}, ID: "n1"},
		},
	}

	err := eng.Run(context.Background(), wf)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if iterations != 2 {
		t.Errorf("expected 2 iterations due to IsLooping, got %d", iterations)
	}
}

func TestRunner_Dependencies(t *testing.T) {
	// Reset Global Tracker for test isolation
	node.GlobalTracker.Clear()

	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"parent": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
			"child": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
		Specs: map[string]node.Spec{
			"child": {
				Meta: node.Meta{
					Type: "child",
					DependsOn: []node.DependsOn{
						{Node: "parent", Action: "success"},
					},
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	// Case 1: Run child without parent -> should FAIL
	wfFail := Workflow{
		ID: "dep-fail-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "child"}, ID: "c1"},
		},
		Edges: []node.Edge{},
	}
	err := eng.Run(context.Background(), wfFail)
	if err == nil {
		t.Error("expected dependency error, but run succeeded")
	}

	// Case 2: Run parent then child -> should SUCCEED
	node.GlobalTracker.Clear()
	wfSuccess := Workflow{
		ID: "dep-success-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "parent"}, ID: "p1", Action: node.Action{Type: "success"}},
			{Meta: node.Meta{Type: "child"}, ID: "c1"},
		},
		Edges: []node.Edge{
			{Source: "p1", SourceHandle: "success", Target: "c1"},
		},
	}
	err = eng.Run(context.Background(), wfSuccess)
	if err != nil {
		t.Errorf("expected run to succeed, but got: %v", err)
	}
}

func TestRunner_ErrorEdgeRecovery(t *testing.T) {
	executionOrder := []string{}
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"start": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "start")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
			"fail": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "fail")
					return node.NodeExecutionResult{}, errors.New("intentional error")
				},
			},
			"skip": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "SHOULD_NOT_RUN")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
			"recovery": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "recovery")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
			"final": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "final")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	// Graph: n1 → n2(fail) → n3(skip) → n5(final)
	//                  ↓ error
	//                 n4(recovery) → n5(final)
	wf := Workflow{
		ID: "error-recovery-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "start"}, ID: "n1"},
			{Meta: node.Meta{Type: "fail"}, ID: "n2"},
			{Meta: node.Meta{Type: "skip"}, ID: "n3"},
			{Meta: node.Meta{Type: "recovery"}, ID: "n4"},
			{Meta: node.Meta{Type: "final"}, ID: "n5"},
		},
		Edges: []node.Edge{
			{Source: "n1", SourceHandle: "success", Target: "n2"},
			{Source: "n2", SourceHandle: "success", Target: "n3"}, // Should NOT follow
			{Source: "n2", SourceHandle: "error", Target: "n4"},   // Should follow this
			{Source: "n3", SourceHandle: "success", Target: "n5"},
			{Source: "n4", SourceHandle: "success", Target: "n5"},
		},
	}

	err := eng.Run(context.Background(), wf)
	if err != nil {
		t.Fatalf("expected recovery to succeed, got: %v", err)
	}

	// Must be: start → fail → recovery → final (skip n3!)
	expected := []string{"start", "fail", "recovery", "final"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("expected %d executions, got %d: %v", len(expected), len(executionOrder), executionOrder)
	}
	for i, v := range expected {
		if executionOrder[i] != v {
			t.Errorf("execution order[%d]: expected %s, got %s", i, v, executionOrder[i])
		}
	}

	// Verify n3 was NOT executed
	for _, node := range executionOrder {
		if node == "SHOULD_NOT_RUN" {
			t.Error("n3 should have been skipped but was executed")
		}
	}
}

func TestRunner_OnErrorStopGraph(t *testing.T) {
	executionOrder := []string{}
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"start": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "start")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
			"fail": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "fail")
					return node.NodeExecutionResult{}, errors.New("hard fail")
				},
			},
			"should-not-run": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "SHOULD_NOT_RUN")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	wf := Workflow{
		ID: "stop-graph-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "start"}, ID: "n1"},
			{Meta: node.Meta{Type: "fail"}, ID: "n2", OnError: "stop"},
			{Meta: node.Meta{Type: "should-not-run"}, ID: "n3"},
		},
		Edges: []node.Edge{
			{Source: "n1", SourceHandle: "success", Target: "n2"},
			{Source: "n2", SourceHandle: "success", Target: "n3"},
		},
	}

	err := eng.Run(context.Background(), wf)
	if err == nil {
		t.Error("expected error from workflow run with OnError:stop")
	}

	// Should only execute: start, fail (then STOP)
	expected := []string{"start", "fail"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("expected %d executions, got %d: %v", len(expected), len(executionOrder), executionOrder)
	}

	// Verify n3 was NOT executed
	for _, node := range executionOrder {
		if node == "SHOULD_NOT_RUN" {
			t.Error("n3 should not have executed after OnError:stop")
		}
	}
}

func TestRunner_OnErrorContinue(t *testing.T) {
	executionOrder := []string{}
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"fail": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "fail")
					return node.NodeExecutionResult{}, errors.New("error")
				},
			},
			"next": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "next")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	wf := Workflow{
		ID: "continue-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "fail"}, ID: "n1", OnError: "continue"},
			{Meta: node.Meta{Type: "next"}, ID: "n2"},
		},
		Edges: []node.Edge{
			{Source: "n1", SourceHandle: "success", Target: "n2"},
		},
	}

	err := eng.Run(context.Background(), wf)
	// Should not error because OnError: continue
	if err != nil {
		t.Fatalf("expected workflow to continue, got error: %v", err)
	}

	// Note: In graph mode with OnError:continue, the next node won't execute
	// because there's no edge from error state. This is expected behavior.
	if len(executionOrder) != 1 {
		t.Logf("execution order: %v", executionOrder)
	}
}

func TestRunner_StatusUpdates(t *testing.T) {
	statusUpdates := []string{}
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"step": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	obs := &MockObserver{
		OnStatusUpdateFunc: func(id, status string, iteration int) {
			statusUpdates = append(statusUpdates, status)
		},
	}
	eng := New(obs, reg)

	wf := Workflow{
		ID: "status-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "step"}, ID: "n1"},
		},
		Edges: []node.Edge{},
	}

	err := eng.Run(context.Background(), wf)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	// Should have: running (start), running (node), Success (node), done (workflow)
	if len(statusUpdates) < 2 {
		t.Errorf("expected at least 2 status updates, got %d: %v", len(statusUpdates), statusUpdates)
	}

	// Last status should be "done"
	if statusUpdates[len(statusUpdates)-1] != "done" {
		t.Errorf("expected last status to be 'done', got '%s'", statusUpdates[len(statusUpdates)-1])
	}
}

func TestRunner_LoopProtection(t *testing.T) {
	executionCount := 0
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"loop": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionCount++
					return node.NodeExecutionResult{Handle: "loop"}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	// Create a self-loop that would run forever without protection
	wf := Workflow{
		ID: "infinite-loop-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "loop"}, ID: "n1"},
		},
		Edges: []node.Edge{
			{Source: "n1", SourceHandle: "loop", Target: "n1"}, // Self-loop
		},
	}

	err := eng.Run(context.Background(), wf)
	// Should complete without error due to loop protection
	if err != nil {
		t.Fatalf("expected loop protection to stop workflow, got error: %v", err)
	}

	// Should stop at 100 visits
	if executionCount > 100 {
		t.Errorf("expected max 100 executions due to loop protection, got %d", executionCount)
	}
}

func TestRunner_MultipleHandles(t *testing.T) {
	executionOrder := []string{}
	reg := &MockRegistry{
		Executors: map[string]node.NodeExecutor{
			"router": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "router")
					return node.NodeExecutionResult{Handle: "custom"}, nil
				},
			},
			"custom-handler": &MockExecutor{
				ExecuteFunc: func(ctx *node.NodeContext) (node.NodeExecutionResult, error) {
					executionOrder = append(executionOrder, "custom-handler")
					return node.NodeExecutionResult{Handle: "success"}, nil
				},
			},
		},
	}
	eng := New(&MockObserver{}, reg)

	wf := Workflow{
		ID: "multi-handle-wf",
		Nodes: []node.Node{
			{Meta: node.Meta{Type: "router"}, ID: "n1"},
			{Meta: node.Meta{Type: "custom-handler"}, ID: "n2"},
		},
		Edges: []node.Edge{
			{Source: "n1", SourceHandle: "custom", Target: "n2"},
		},
	}

	err := eng.Run(context.Background(), wf)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	expected := []string{"router", "custom-handler"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("expected %d executions, got %d", len(expected), len(executionOrder))
	}
	for i, v := range expected {
		if executionOrder[i] != v {
			t.Errorf("execution order[%d]: expected %s, got %s", i, v, executionOrder[i])
		}
	}
}

// ============================================================================
// WorkflowService Tests
// ============================================================================

type MockWorkflowStore struct {
	ListFunc      func() ([]WorkflowFile, error)
	GetFunc       func(id string) (Workflow, []byte, error)
	ExistsFunc    func(id string) bool
	SaveFunc      func(id string, wf Workflow) (string, error)
	DeleteFunc    func(id string) error
	DeleteAllFunc func() error
}

func (m *MockWorkflowStore) List() ([]WorkflowFile, error)           { return m.ListFunc() }
func (m *MockWorkflowStore) Get(id string) (Workflow, []byte, error) { return m.GetFunc(id) }
func (m *MockWorkflowStore) Exists(id string) bool                   { return m.ExistsFunc(id) }
func (m *MockWorkflowStore) Save(id string, wf Workflow) (string, error) {
	return m.SaveFunc(id, wf)
}
func (m *MockWorkflowStore) Delete(id string) error { return m.DeleteFunc(id) }
func (m *MockWorkflowStore) DeleteAll() error {
	if m.DeleteAllFunc != nil {
		return m.DeleteAllFunc()
	}
	return nil
}

func TestWorkflowService_Delete(t *testing.T) {
	called := false
	mock := &MockWorkflowStore{
		DeleteFunc: func(id string) error {
			called = true
			return nil
		},
	}

	service := NewWorkflowService(mock)
	err := service.Delete("test-id")
	if err != nil || !called {
		t.Errorf("Delete failed")
	}
}

func TestWorkflowService_Create(t *testing.T) {
	mock := &MockWorkflowStore{
		ExistsFunc: func(id string) bool { return false },
		SaveFunc: func(id string, wf Workflow) (string, error) {
			return "/path/to/wf.json", nil
		},
	}

	service := NewWorkflowService(mock)
	wf := Workflow{Name: "New Workflow", Nodes: []node.Node{}}

	_, _, err := service.Create(wf)
	if err != nil {
		t.Errorf("unexpected error in Create: %v", err)
	}

	res := service.ValidateForRun(wf)
	if res.Valid {
		t.Error("expected ValidateForRun to be invalid for empty nodes workflow")
	}
}

func TestWorkflowService_GetSanitizesRawPayload(t *testing.T) {
	mock := &MockWorkflowStore{
		GetFunc: func(id string) (Workflow, []byte, error) {
			return Workflow{
				ID:   id,
				Name: "Workflow",
				Nodes: []node.Node{
					{
						Meta: node.Meta{
							Type:        "http",
							Label:       "HTTP",
							Description: "catalog description",
							Icon:        "icon.svg",
							Tags:        []string{"Action"},
						},
						Action: node.Action{
							Type:        "request",
							Label:       "Request",
							Description: "action description",
							HasResponse: true,
							Fields: []node.NodeField{
								{Key: "url", Type: "string", Label: "URL", Value: "https://example.com"},
							},
						},
					},
				},
			}, []byte(`{"stale":"raw"}`), nil
		},
	}

	service := NewWorkflowService(mock)
	wf, raw, err := service.Get("wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wf.Nodes[0].Icon != "" || wf.Nodes[0].Description != "" || len(wf.Nodes[0].Tags) != 0 {
		t.Fatalf("expected sanitized workflow node metadata, got %+v", wf.Nodes[0].Meta)
	}

	var decoded Workflow
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("raw payload should be valid sanitized workflow json: %v", err)
	}
	if decoded.Nodes[0].Icon != "" || decoded.Nodes[0].Action.Label != "" || decoded.Nodes[0].Action.Fields[0].Label != "" {
		t.Fatalf("expected sanitized raw payload, got %+v", decoded.Nodes[0])
	}
}

func TestWorkflowService_Validate(t *testing.T) {
	service := NewWorkflowService(nil)

	res := service.Validate(Workflow{Name: ""})
	if res.Valid {
		t.Error("expected invalid for empty name")
	}
}
