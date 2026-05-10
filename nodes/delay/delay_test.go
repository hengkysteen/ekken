package delay

import (
	"context"
	"ekken/internal/features/workflow/node"
	"testing"
	"time"
)

func TestDelayNode_Execute(t *testing.T) {
	tests := []struct {
		name        string
		duration    float64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid 1 second delay",
			duration:    1.0,
			expectError: false,
		},
		{
			name:        "valid 0.5 second delay",
			duration:    0.5,
			expectError: false,
		},
		{
			name:        "zero delay",
			duration:    0,
			expectError: false,
		},
		{
			name:        "negative duration",
			duration:    -1,
			expectError: true,
			errorMsg:    "duration cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &DelayNode{
				Config: map[string]any{
					"duration": tt.duration,
				},
			}

			ctx := &node.NodeContext{
				Stop:    make(chan struct{}),
				Context: context.Background(),
			}

			start := time.Now()
			result, err := n.Execute(ctx)
			elapsed := time.Since(start)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Handle != "success" {
				t.Errorf("expected handle 'success', got '%s'", result.Handle)
			}

			expectedDuration := time.Duration(tt.duration * float64(time.Second))
			if elapsed < expectedDuration {
				t.Errorf("delay too short: expected ~%v, got %v", expectedDuration, elapsed)
			}
		})
	}
}

func TestDelayNode_InvalidDuration(t *testing.T) {
	n := &DelayNode{
		Config: map[string]any{
			"duration": "invalid",
		},
	}

	ctx := &node.NodeContext{
		Stop:    make(chan struct{}),
		Context: context.Background(),
	}

	_, err := n.Execute(ctx)
	if err == nil {
		t.Error("expected error for invalid duration format")
	}
	if err.Error() != "invalid duration format" {
		t.Errorf("expected 'invalid duration format', got '%s'", err.Error())
	}
}

func TestDelayNode_Cancellation(t *testing.T) {
	n := &DelayNode{
		Config: map[string]any{
			"duration": 5.0, // 5 seconds
		},
	}

	stopChan := make(chan struct{})
	ctx := &node.NodeContext{
		Stop:    stopChan,
		Context: context.Background(),
	}

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(stopChan)
	}()

	start := time.Now()
	_, err := n.Execute(ctx)
	elapsed := time.Since(start)

	if err != node.ErrNodeStopped {
		t.Errorf("expected ErrNodeStopped, got %v", err)
	}

	if elapsed > 1*time.Second {
		t.Errorf("cancellation took too long: %v", elapsed)
	}
}
