package webhook

import (
	"context"
	"testing"
	"time"

	"ekken/internal/features/workflow/node"
)

func TestWebhookSpecRegistered(t *testing.T) {
	spec, ok := node.GlobalRegistry.GetSpec("webhook")
	if !ok {
		t.Fatal("webhook spec is not registered")
	}
	if spec.DefaultAction != "on_request" {
		t.Fatalf("DefaultAction = %q, want on_request", spec.DefaultAction)
	}
	if len(spec.Tags) != 1 || spec.Tags[0] != "Trigger" {
		t.Fatalf("Tags = %#v, want [Trigger]", spec.Tags)
	}
	if len(spec.Actions) != 1 {
		t.Fatalf("len(Actions) = %d, want 1", len(spec.Actions))
	}
	if !spec.Actions[0].HasResponse {
		t.Fatal("webhook on_request must have response")
	}
}

func TestWebhookExecuteReturnsTriggerPayload(t *testing.T) {
	executor := (&WebhookNode{action: node.Action{
		Type: "on_request",
		Fields: []node.NodeField{
			{Key: "webhook_id", Value: "hook-execute"},
		},
	}})
	payload := map[string]any{"event": "ping"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := make(chan node.NodeExecutionResult, 1)
	errCh := make(chan error, 1)
	go func() {
		result, err := executor.Execute(&node.NodeContext{
			Context:    ctx,
			WorkflowID: "wf-1",
			Metadata:   map[string]any{},
			Stop:       ctx.Done(),
		})
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	if err := waitForRegisteredListener("hook-execute"); err != nil {
		t.Fatal(err)
	}
	if err := node.GlobalEventListeners.Dispatch("hook-execute", payload); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}

	var result node.NodeExecutionResult
	select {
	case err := <-errCh:
		t.Fatalf("Execute returned error: %v", err)
	case result = <-resultCh:
	case <-time.After(time.Second):
		t.Fatal("Execute did not receive webhook payload")
	}

	if result.Handle != "success" {
		t.Fatalf("Handle = %q, want success", result.Handle)
	}
	response, ok := result.Response.(map[string]any)
	if !ok {
		t.Fatalf("Response type = %T, want map[string]any", result.Response)
	}
	if response["event"] != "ping" {
		t.Fatalf("Response = %#v, want trigger payload", result.Response)
	}
}

func TestWebhookExecuteStopsWhileWaiting(t *testing.T) {
	executor := (&WebhookNode{action: node.Action{
		Type: "on_request",
		Fields: []node.NodeField{
			{Key: "webhook_id", Value: "hook-stop"},
		},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		_, err := executor.Execute(&node.NodeContext{
			Context:    ctx,
			WorkflowID: "wf-1",
			Metadata:   map[string]any{},
			Stop:       ctx.Done(),
		})
		errCh <- err
	}()

	if err := waitForRegisteredListener("hook-stop"); err != nil {
		t.Fatal(err)
	}
	cancel()

	select {
	case err := <-errCh:
		if err != node.ErrNodeStopped {
			t.Fatalf("Execute error = %v, want ErrNodeStopped", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Execute did not stop")
	}
}

func waitForRegisteredListener(id string) error {
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, ok := node.GlobalEventListeners.Get(id); ok {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return &timeoutError{id: id}
}

type timeoutError struct {
	id string
}

func (e *timeoutError) Error() string {
	return "listener not active: " + e.id
}
