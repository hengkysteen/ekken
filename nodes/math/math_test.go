package mathnode

import (
	"context"
	"math"
	"testing"

	"ekken/internal/features/workflow/node"
)

func TestMathNode_Calculate(t *testing.T) {
	tests := []struct {
		name       string
		op1        string
		operator   string
		op2        string
		wantResult float64
		wantErr    bool
	}{
		{"addition", "5.5", "+", "4.5", 10.0, false},
		{"subtraction", "10", "-", "4.5", 5.5, false},
		{"multiplication", "3", "*", "1.5", 4.5, false},
		{"division", "10", "/", "4", 2.5, false},
		{"modulo", "10", "%", "3", 1.0, false},
		{"power", "2", "^", "3", 8.0, false},
		{"division by zero", "5", "/", "0", 0.0, true},
		{"modulo by zero", "5", "%", "0", 0.0, true},
		{"invalid op1", "abc", "+", "2", 0.0, true},
		{"invalid op2", "2", "+", "xyz", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &MathNode{
				Action: node.Action{
					Type: "calculate",
					Fields: []node.NodeField{
						{Key: "operand_1", Value: tt.op1},
						{Key: "operator", Value: tt.operator},
						{Key: "operand_2", Value: tt.op2},
					},
				},
			}

			ctx := &node.NodeContext{
				Context:   context.Background(),
				Variables: make(map[string]any),
				Stop:      make(chan struct{}),
			}

			res, err := n.Execute(ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if res.Handle != "success" {
					t.Fatalf("expected handle success, got %s", res.Handle)
				}
				gotVal, ok := res.Response.(float64)
				if !ok {
					t.Fatalf("expected float64 response, got %T", res.Response)
				}
				if math.Abs(gotVal-tt.wantResult) > 1e-9 {
					t.Errorf("got %v, want %v", gotVal, tt.wantResult)
				}
			}
		})
	}
}

func TestMathNode_Round(t *testing.T) {
	tests := []struct {
		name       string
		val        string
		decimals   float64
		method     string
		wantResult float64
	}{
		{"round to nearest default", "3.4", 0, "round", 3.0},
		{"round to nearest up", "3.6", 0, "round", 4.0},
		{"round 2 decimals", "3.14159", 2, "round", 3.14},
		{"floor", "3.9", 0, "floor", 3.0},
		{"floor 2 decimals", "3.149", 2, "floor", 3.14},
		{"ceil", "3.1", 0, "ceil", 4.0},
		{"ceil 2 decimals", "3.141", 2, "ceil", 3.15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &MathNode{
				Action: node.Action{
					Type: "round",
					Fields: []node.NodeField{
						{Key: "value", Value: tt.val},
						{Key: "decimals", Value: tt.decimals},
						{Key: "method", Value: tt.method},
					},
				},
			}

			ctx := &node.NodeContext{
				Context:   context.Background(),
				Variables: make(map[string]any),
				Stop:      make(chan struct{}),
			}

			res, err := n.Execute(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotVal, ok := res.Response.(float64)
			if !ok {
				t.Fatalf("expected float64 response, got %T", res.Response)
			}
			if math.Abs(gotVal-tt.wantResult) > 1e-9 {
				t.Errorf("got %v, want %v", gotVal, tt.wantResult)
			}
		})
	}
}

func TestMathNode_Random(t *testing.T) {
	n := &MathNode{
		Action: node.Action{
			Type: "random",
			Fields: []node.NodeField{
				{Key: "min", Value: 5.0},
				{Key: "max", Value: 10.0},
				{Key: "type", Value: "integer"},
			},
		},
	}

	ctx := &node.NodeContext{
		Context:   context.Background(),
		Variables: make(map[string]any),
		Stop:      make(chan struct{}),
	}

	for i := 0; i < 100; i++ {
		res, err := n.Execute(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		gotVal, ok := res.Response.(float64)
		if !ok {
			t.Fatalf("expected float64 response, got %T", res.Response)
		}
		if gotVal != math.Floor(gotVal) {
			t.Fatalf("expected integer float, got %v", gotVal)
		}
		if gotVal < 5.0 || gotVal > 10.0 {
			t.Fatalf("random value out of range [5, 10]: %v", gotVal)
		}
	}
}
